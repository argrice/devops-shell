package devops-shell

import (
    "bufio"
    "errors"
    "fmt"
    "io"
    "os"
    "os/exec"
    "os/user"
    "path/filepath"
    "strings"
    "github.com/peterh/liner"
)

func execInput(input string) error {
	// Split the input into commands based on pipe (|)
	cmds := strings.Split(input, "|")

	// Create a wait group to wait for all commands to finish
	var wg sync.WaitGroup
	var prevCmd *exec.Cmd

	for i, cmdStr := range cmds {
		// Prepare the command to execute
		args := strings.Fields(strings.TrimSpace(cmdStr))
		if len(args) == 0 {
			continue
		}

		cmd := exec.Command(args[0], args[1:]...)

		if prevCmd != nil {
			// If this is not the first command, set the previous command's stdout as this command's stdin
			stdout, err := prevCmd.StdoutPipe()
			if err != nil {
				return fmt.Errorf("error getting stdout pipe for command %s: %w", prevCmd.Args, err)
			}
			cmd.Stdin = stdout
		}

		// Capture the output and errors
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf

		// Increment the wait group counter
		wg.Add(1)

		go func(cmd *exec.Cmd, index int) {
			defer wg.Done()
			if err := cmd.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "error starting command %s: %v\n", cmd.Args, err)
				return
			}

			if err := cmd.Wait(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					if exitErr.ExitCode() == 1 {
						fmt.Printf("No matches found for command: %s\n", cmd.Args)
					} else {
						fmt.Fprintf(os.Stderr, "Command %s exited with error: %v\n", cmd.Args, err)
					}
				} else {
					fmt.Fprintf(os.Stderr, "Error waiting for command %s: %v\n", cmd.Args, err)
				}
			}

			// Print the output of the command
			if index == len(cmds)-1 { // Only print for the last command
				if outBuf.Len() > 0 {
					fmt.Print(outBuf.String())
				}
				if errBuf.Len() > 0 {
					fmt.Fprint(os.Stderr, errBuf.String())
				}
			}
		}(cmd, i)

		// Keep track of the previous command
		prevCmd = cmd
	}

	// Close the last command's stdin
	if prevCmd != nil {
		if err := prevCmd.Stdout.Close(); err != nil {
			return fmt.Errorf("error closing stdout for command %s: %w", prevCmd.Args, err)
		}
	}

	// Wait for all commands to finish
	wg.Wait()
	return nil
}
