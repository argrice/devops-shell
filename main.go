package main

import (
    "errors"
    "fmt"
    "os"
    "os/user"
    "path/filepath"
    "strings"
    "github.com/peterh/liner"
)

var historyFile string

// ErrExit is returned when loop needs to exit
var ErrExit = errors.New("exit")

func main() {
    // User shouldn't be changing. This will change the shell itself and cause errors
    currUser, err := user.Current()    
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error getting current user:", err)
        return
    }

    line := liner.NewLiner()
    defer line.Close()

    // Create input history file in user's home dir
    historyFile = filepath.Join(currUser.HomeDir, ".dv_shell_history.txt")
    
    history := loadHistory(historyFile)
    for _, cmd := range history {
        line.AppendHistory(cmd)
    }

    for {
        // Pull system's hostname
        hostname, err := os.Hostname()
        if err != nil {
            fmt.Fprintln(os.Stderr, "Error getting hostname:", err)
            continue
        }

        // Pull current working directory
        cwd, err := os.Getwd()
        if err != nil {
            fmt.Fprintln(os.Stderr, "Error getting current working directory:", err)
            continue
        }

        // Generate prompt for input
        prompt := fmt.Sprintf("%s@%s %s > ", currUser.Username, hostname, cwd)
        input, err := line.Prompt(prompt)
        if err != nil {
            if err == liner.ErrPromptAborted {
                fmt.Println("Exiting shell...")
                saveHistory(history, historyFile)
                break
            }
            fmt.Fprintln(os.Stderr, err)
            continue
        }

        input = strings.TrimSpace(input)
        if input == "" {
            continue
        }

        // Add command to history file and line for in-memory caching of history
        line.AppendHistory(input)
        history = append(history, input)

        // For custom command runparallel, see taskManager.go
        if strings.HasPrefix(input, "runparallel ") {
            commands := strings.Split(input[len("runparallel "):], ";")
            tm := NewTaskManager()
            tm.RunParallel(commands)
        } else if err = execInput(input); err != nil {
            if errors.Is(err, ErrExit) {
                saveHistory(history, historyFile)
                fmt.Println("Exiting shell...")
                break
            }
            fmt.Fprintf(os.Stderr, "%s is not valid\n", input)
        }
    }
}
