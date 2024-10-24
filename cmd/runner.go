package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/google/shlex"
)

func RunShell(commands []string, parallel bool) {
	runCommand := func(command string) {
		Log("Shell", fmt.Sprintf("sh -c '%s'", color.YellowString(command)))
		cmd := exec.Command("sh", "-c", command)

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error running command %q: %v\n", command, err)
		}
	}

	if parallel {
		var wg sync.WaitGroup
		for _, command := range commands {
			wg.Add(1)
			go func(c string) {
				defer wg.Done()
				runCommand(c)
			}(command)
		}
		wg.Wait()
	} else {
		for _, cmd := range commands {
			runCommand(cmd)
		}
	}
}

func RunCommand(commands []string, parallel bool) {
	runSingleCommand := func(command string) {
		parts, err := shlex.Split(command)
		if err != nil {
			fmt.Printf("Error splitting command %q: %v\n", command, err)
			return
		}

		qoutedPart := []string{}
		for idx, part := range parts {
			switch {
			case idx == 0:
				part = color.GreenString(part)
			case strings.HasPrefix(part, "-"):
				part = color.CyanString(part)
			default:
				part = color.RedString(part)
			}

			qoutedPart = append(qoutedPart, part)
		}
		Log("Command", qoutedPart)

		if len(parts) == 0 {
			fmt.Println("No command found to run")
			return
		}

		cmd := exec.Command(parts[0], parts[1:]...)

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error running command %q: %v\n", command, err)
		}
	}

	if parallel {
		var wg sync.WaitGroup
		for _, cmd := range commands {
			wg.Add(1)
			go func(c string) {
				defer wg.Done()
				runSingleCommand(c)
			}(cmd)
		}
		wg.Wait()
	} else {
		for _, cmd := range commands {
			runSingleCommand(cmd)
		}
	}
}
