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

// ErrExit is returned when the shell should exit.
var ErrExit = errors.New("exit")

func main() {
    currUser, err := user.Current()
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error getting current user:", err)
        return
    }

    historyFile = filepath.Join(currUser.HomeDir, ".ash_history.txt")

    line := liner.NewLiner()
    defer line.Close()

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

        line.AppendHistory(input)
        history = append(history, input)

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
