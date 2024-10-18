package cmd

import (
	"fmt"
	"os/exec"
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

func expendShell(s *string, value AstValue) {
	for idx, sh := range value.Shell {
		cmd := exec.Command("sh", "-c", sh)

		output, err := cmd.Output()
		if err == nil {
			*s = insertAtIndex(*s, string(output), idx)
		} else {
			Log(fmt.Sprintf("Shell $(%s)", sh), err)
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
		expendShell(&str, value[0])
		expendVariable(&config, &str, value[0])

		config.Properties[key] = str
	}

	for section, properties := range ast.Sections {
		sec := Section{}
		for key, value := range properties {
			switch key {
			case "shell":
				str := value[0].String
				expendShell(&str, value[0])
				expendVariable(&config, &str, value[0])
				sec.Shell = append(sec.Shell, str)
			case "cmd":
				str := value[0].String
				expendShell(&str, value[0])
				expendVariable(&config, &str, value[0])
				sec.Command = append(sec.Command, str)
			}
		}
		config.Sections[section] = sec
	}

	return config, nil
}
