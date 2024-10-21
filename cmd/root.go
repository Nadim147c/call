package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	Version = "GitHub"
	Verbose bool
)

func Debug(p string, a ...any) {
	if Verbose {
		fmt.Fprintf(os.Stderr, "%s: ", color.GreenString(p))
		fmt.Fprintln(os.Stderr, a...)
	}
}

func Log(p string, a ...any) {
	fmt.Fprintf(os.Stderr, "%s: ", color.GreenString(p))
	fmt.Fprintln(os.Stderr, a...)
}

func runTask(config *Config, taskName string) error {
	if sec, found := config.Sections[taskName]; found {
		for _, childTask := range sec.Child {
			err := runTask(config, childTask)
			if err != nil {
				return err
			}
		}

		Log("Task", taskName)

		if len(sec.Shell) > 0 {
			RunShell(sec.Shell, false)
		}
		if len(sec.Command) > 0 {
			RunCommand(sec.Command, false)
		}

		return nil
	} else {
		return fmt.Errorf("%s section doesn't exists", taskName)
	}
}

func filterSlice(args []string) []string {
	var filtered []string
	for _, arg := range args {
		// Check if the string contains "="
		if !strings.Contains(arg, "=") {
			filtered = append(filtered, arg)
		}
	}
	return filtered
}

var rootCmd = &cobra.Command{
	Use:     "call [command]",
	Short:   "A highly experimental make(1) like tool",
	Version: Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		Debug("Parsing Taskfile...")

		taskFile, err := os.ReadFile("Taskfile")
		if err != nil {
			panic(err)
		}

		ast, err := GetAst(string(taskFile))
		if err != nil {
			panic(err)
		}

		astJson, err := json.MarshalIndent(ast, "", " ")
		if err != nil {
			panic(err)
		}
		Debug("AST", string(astJson))

		config, err := GetConfig(ast, args)
		if err != nil {
			panic(err)
		}

		configJson, err := json.MarshalIndent(config, "", " ")
		if err != nil {
			panic(err)
		}
		Debug("Config", string(configJson))

		filteredArgs := filterSlice(args)

		if len(args) == 0 {
			return runTask(&config, "default")
		}

		for _, arg := range filteredArgs {
			return runTask(&config, arg)
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintln("call version:", color.GreenString(Version)))
	rootCmd.SetErrPrefix("Task Error:")
	rootCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "enable verbose logging")
}
