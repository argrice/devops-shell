package main

import (
    "bufio"
    "fmt"
    "os"
)


func loadHistory(path string) []string {
    file, err := os.Open(path)
    if err != nil {
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
