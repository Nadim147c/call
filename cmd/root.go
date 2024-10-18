package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Verbose bool

func Debug(p string, a ...any) {
	if Verbose {
		fmt.Fprintf(os.Stderr, "%s: ", p)
		fmt.Fprintln(os.Stderr, a...)
	}
}

func Log(p string, a ...any) {
	fmt.Fprintf(os.Stderr, "%s: ", p)
	fmt.Fprintln(os.Stderr, a...)
}

func runSection(s Section) {
	if len(s.Shell) > 0 {
		RunShell(s.Shell, false)
	}
	if len(s.Command) > 0 {
		RunCommand(s.Command, false)
	}
}

var rootCmd = &cobra.Command{
	Use:   "call [call-flags] -- [command]",
	Short: "A brief description of your application",
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

		config, err := GetConfig(ast)
		if err != nil {
			panic(err)
		}

		if len(args) == 0 {
			if defaultSection, found := config.Sections["default"]; found {
				runSection(defaultSection)
			} else {
				return fmt.Errorf("Default section doesn't exists")
			}
		}
		for _, arg := range args {
			if section, found := config.Sections[arg]; found {
				runSection(section)
			} else {
				return fmt.Errorf("%s section doesn't exists", arg)
			}
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
	rootCmd.SetErrPrefix("Task Error:")
	rootCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "enable verbose logging")
}
