package main

import (
    "context"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "os/user"
    "path/filepath"
    "strings"
    "sync"
    "time"

    "github.com/peterh/liner"
)

var historyFile string

// ErrExit is returned when the shell should exit.
var ErrExit = errors.New("exit")

// TaskResult holds the result of a task execution.
type TaskResult struct {
    Command string
    Output  string
    Err     error
    Status  string
}

// TaskManager manages running tasks and provides real-time feedback.
type TaskManager struct {
    tasks     map[int]*TaskResult
    taskMutex sync.Mutex
}

// NewTaskManager initializes a TaskManager.
func NewTaskManager() *TaskManager {
    return &TaskManager{
        tasks: make(map[int]*TaskResult),
    }
}

// AddTask adds a task to the manager.
func (tm *TaskManager) AddTask(id int, command string) {
    tm.taskMutex.Lock()
    tm.tasks[id] = &TaskResult{
        Command: command,
        Status:  "running",
    }
    tm.taskMutex.Unlock()
}

// UpdateTaskStatus updates the status of a task.
func (tm *TaskManager) UpdateTaskStatus(id int, status string) {
    tm.taskMutex.Lock()
    if task, exists := tm.tasks[id]; exists {
        task.Status = status
    }
    tm.taskMutex.Unlock()
}

// RunParallel executes commands in parallel and provides real-time feedback.
func (tm *TaskManager) RunParallel(commands []string) {
    var wg sync.WaitGroup
    results := make(chan TaskResult, len(commands))

    for id, cmd := range commands {
        tm.AddTask(id, cmd)
        wg.Add(1)

        go func(id int, cmd string) {
            defer wg.Done()

            ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            defer cancel()

            output, err := executeCommandWithContext(ctx, cmd)
            status := "success"
            if err != nil {
                status = "failed"
            }

            tm.UpdateTaskStatus(id, status)
            results <- TaskResult{Command: cmd, Output: output, Err: err, Status: status}
        }(id, cmd)
    }

    go tm.printTaskStatus()

    wg.Wait()
    close(results)

    fmt.Println("\n--- Parallel Task Results ---")
    for result := range results {
        fmt.Printf("Command: %s\nStatus: %s\n", result.Command, result.Status)
        if result.Err != nil {
            fmt.Printf("Error: %v\n", result.Err)
        } else {
            fmt.Printf("Output: %s\n", result.Output)
        }
        fmt.Println()
    }
}

// printTaskStatus displays the current status of each task.
func (tm *TaskManager) printTaskStatus() {
    for {
        time.Sleep(1 * time.Second)
        fmt.Println("\n--- Task Status ---")
        tm.taskMutex.Lock()
        for id, task := range tm.tasks {
            fmt.Printf("Task %d (%s): %s\n", id, task.Command, task.Status)
        }
        tm.taskMutex.Unlock()

        allDone := true
        tm.taskMutex.Lock()
        for _, task := range tm.tasks {
            if task.Status == "running" {
                allDone = false
                break
            }
        }
        tm.taskMutex.Unlock()
        if allDone {
            break
        }
    }
}

// executeCommandWithContext runs a command with a given context.
func executeCommandWithContext(ctx context.Context, cmd string) (string, error) {
    args := strings.Fields(cmd) // Split command into fields for execution
    command := exec.CommandContext(ctx, args[0], args[1:]...) // Execute command directly

    output, err := command.CombinedOutput() // Combine standard output and error
    return string(output), err
}

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
