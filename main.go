package devops-shell

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
        "github.com/peterh/liner"
)

var historyFile string // Declare a variable for the history file

// ErrExit is returned when the shell should exit.
var ErrExit = errors.New("exit")

func main() {
	currUser, err := user.Current()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting current user:", err)
		return
	}

	// Set the history file path to the user's home directory
	historyFile = filepath.Join(currUser.HomeDir, ".ash_history.txt") // Hidden file in the home directory

	line := liner.NewLiner()
	defer line.Close()

	// Load previous history
	history := loadHistory(historyFile)
	for _, cmd := range history {
		line.AppendHistory(cmd)
	}

	for {
		hostname, err := os.Hostname()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error getting hostname:", err)
			continue
		}

		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error getting current working directory:", err)
			continue
		}

		// Construct the prompt
		prompt := fmt.Sprintf("%s@%s %s > ", currUser.Username, hostname, cwd)

		// Read the input with history support
		input, err := line.Prompt(prompt)
		if err != nil {
			if err == liner.ErrPromptAborted {
				fmt.Println("Exiting shell...")
				saveHistory(history, historyFile) // Save history before exiting
				break
			}
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		// Trim whitespace and skip empty commands
		input = strings.TrimSpace(input)
		if input == "" {
			// If the input is empty, just continue to the next prompt
			continue
		}

		line.AppendHistory(input)
		history = append(history, input) // Maintain your own history slice

		// Handle the execution of the input
		if err = execInput(input); err != nil {
			if errors.Is(err, ErrExit) {
				saveHistory(history, historyFile) // Save history before exiting
				fmt.Println("Exiting shell...")
				break // Exit the main loop
			}
			// Print custom error message for invalid commands
			fmt.Fprintf(os.Stderr, "%s is not valid\n", input)
		}
	}
}

// Load history from a file
func loadHistory(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		// If the file does not exist, we can safely ignore the error
		return []string{}
	}
	defer file.Close()

	var history []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		history = append(history, scanner.Text())
	}

	return history
}

// Save history to a file
func saveHistory(history []string, path string) {
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error saving history:", err)
		return
	}
	defer file.Close()

	for _, item := range history {
		file.WriteString(item + "\n")
	}
}

// ErrNoPath is returned when 'cd' was called without a second argument.
var ErrNoPath = errors.New("path required")

