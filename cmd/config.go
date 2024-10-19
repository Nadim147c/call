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

func expendSubShell(c *Config, s *string, value AstValue) {
	for idx, sh := range value.Shell {
		cmd := exec.Command("sh", "-c", sh)
		cmd.Env = os.Environ()

		for key, value := range c.Properties {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}

		output, err := cmd.Output()
		outString := strings.TrimRightFunc(string(output), unicode.IsSpace)
		if err == nil {
			*s = insertAtIndex(*s, outString, idx)
		} else {
			Debug(fmt.Sprintf("Shell $(%s)", sh), err)
		}
	}
}

func expendVariable(c *Config, s *string, value AstValue) {
	for idx, varName := range value.Variables {
		if varValue, found := c.Properties[varName]; found {
			*s = insertAtIndex(*s, varValue, idx)
		} else {
			Log(fmt.Sprintf("Variable $(%s)", varName), fmt.Errorf("Variable doesn't exists"))
		}
	}
}

func GetConfig(ast AST) (Config, error) {
	config := Config{
		Sections:   make(Sections),
		Properties: make(Properties),
	}

	for key, value := range ast.Properties {
		str := value[0].String
		expendVariable(&config, &str, value[0])
		expendSubShell(&config, &str, value[0])

		config.Properties[key] = str
	}

	for section, properties := range ast.Sections {
		sec := Section{}
		for key, values := range properties {
			switch key {
			case "shell":
				for _, value := range values {
					str := value.String
					expendVariable(&config, &str, value)
					expendSubShell(&config, &str, value)
					sec.Shell = append(sec.Shell, str)
				}
			case "cmd":
				for _, value := range values {
					str := value.String
					expendVariable(&config, &str, value)
					expendSubShell(&config, &str, value)
					sec.Command = append(sec.Command, str)
				}
			}
		}
		config.Sections[section] = sec
	}

	return config, nil
}
