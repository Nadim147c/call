package cmd

import (
	"encoding/json"
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

var rootCmd = &cobra.Command{
	Use:   "call [call-flags] -- [command]",
	Short: "A brief description of your application",
	RunE: func(cmd *cobra.Command, args []string) error {
		Debug("Parsing Taskfile...")

		taskFile, err := os.ReadFile("Taskfile")
		if err != nil {
			panic(err)
		}

		config, err := GetAst(string(taskFile))
		if err != nil {
			panic(err)
		}

		json.NewEncoder(os.Stdout).Encode(config)

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
