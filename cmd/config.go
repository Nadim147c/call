package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

type (
	Config struct {
		Properties Properties
		Sections   Sections
	}

	Properties map[string]string

	Sections map[string]Section

	Section struct {
		Child   []string
		Shell   []string
		Command []string
	}
)

func insertAtIndex(s string, insert string, idx int) string {
	if idx > len(s) {
		idx = len(s)
	}
	return s[:idx] + insert + s[idx:]
}

func expendValue(c *Config, value AstValue, input string) string {
	positionMap := make(map[int]string)

	for idx, varName := range value.Variables {
		if varValue, found := c.Properties[varName]; found {
			positionMap[idx] = varValue
		} else {
			Log(fmt.Sprintf("Variable ${%s}", varName), fmt.Errorf("Variable doesn't exists"))
		}
	}

	for idx, sh := range value.Shell {
		cmd := exec.Command("sh", "-c", sh)
		cmd.Env = os.Environ()

		for key, value := range c.Properties {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}

		output, err := cmd.Output()
		if err != nil {
			Log(fmt.Sprintf("Shell $(%s)", sh), err)
			continue
		}

		outputString := strings.TrimRightFunc(string(output), unicode.IsSpace)
		positionMap[idx] = outputString
	}

	var outString strings.Builder

	for idx, char := range []byte(input) {
		if varValue, found := positionMap[idx]; found {
			delete(positionMap, idx)
			outString.WriteString(varValue)
		}
		outString.WriteByte(char)
	}

	if len(positionMap) != 0 {
		var extra strings.Builder

		for _, varValue := range positionMap {
			extra.WriteByte(' ')
			extra.WriteString(varValue)
		}
		if outString.Len() == 0 {
			return extra.String()[1:]
		}

		outString.WriteString(extra.String())
	}

	return outString.String()
}

func detectCycle(s Sections, key string, visited map[string]bool, recStack map[string]bool) bool {
	if recStack[key] {
		return true
	}
	if visited[key] {
		return false
	}

	visited[key] = true
	recStack[key] = true

	for _, childKey := range s[key].Child {
		if detectCycle(s, childKey, visited, recStack) {
			return true
		}
	}

	// Remove the key from the recursion stack.
	recStack[key] = false
	return false
}

func GetConfig(ast AST, args []string) (Config, error) {
	config := Config{
		Sections:   make(Sections),
		Properties: make(Properties),
	}

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			config.Properties[key] = value
		}
	}

	for key, values := range ast.Properties {
		for _, value := range values {
			if _, keyExists := config.Properties[key]; keyExists && value.Optional {
				continue
			}

			str := expendValue(&config, value, value.String)

			if str != "" {
				config.Properties[key] = str
			}
		}
	}

	for section, properties := range ast.Sections {
		sec := Section{}
		for key, values := range properties {
			switch key {
			case "child":
				for _, value := range values {
					sec.Child = append(sec.Child, value.String)
				}
			case "shell":
				for _, value := range values {
					str := expendValue(&config, value, value.String)
					sec.Shell = append(sec.Shell, str)
				}
			case "cmd":
				for _, value := range values {
					str := expendValue(&config, value, value.String)
					sec.Command = append(sec.Command, str)
				}
			}
		}
		config.Sections[section] = sec
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for key := range config.Sections {
		if detectCycle(config.Sections, key, visited, recStack) {
			return config, fmt.Errorf("Cycle detected in the map.")
		}
	}

	return config, nil
}
