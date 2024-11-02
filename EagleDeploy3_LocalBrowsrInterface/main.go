package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
    // Define the routes
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/upload-playbook", uploadPlaybookHandler)
    http.HandleFunc("/execute-playbook", executePlaybookHandler)
    http.HandleFunc("/hosts", hostsHandler)
    http.HandleFunc("/task-status", taskStatusHandler)
    http.HandleFunc("/logs", logsHandler)

    // Start the server on port 8080
    fmt.Println("Starting server on :8080...")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "index.html")
}

func uploadPlaybookHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Playbook upload handler (placeholder)")
}

func executePlaybookHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Playbook execute handler (placeholder)")
}

func hostsHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hosts handler (placeholder)")
}

func taskStatusHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Task status handler (placeholder)")
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Logs handler (placeholder)")
}
