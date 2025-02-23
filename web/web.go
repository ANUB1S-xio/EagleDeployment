package web

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"EagleDeploy_CLI/executor" // executor package
)

// find open port
func findPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// StartWebServer initializes server with dynamic port signing
func StartWebServer() {
	port := 8742 // Default port

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		fmt.Println("Port 8742 occupied, dynamically assigning port...")
		port, err = findPort()
		if err != nil {
			fmt.Printf("Failed to assign port: %v\n", err)
			return
		}
	} else {
		listener.Close()
	}

	fmt.Printf("EagleDeployment GUI running at http://127.0.0.1:%d\n", port)

	// Serve static assets (CSS, JS, Images)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Login Page (default)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/login.html")
	})

	// Dashboard Page
	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/dashboard.html")
	})

	// Execute Playbook Page
	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/execute.html")
	})

	// List YAML Playbooks Page
	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/list.html")
	})

	// API Endpoint to Execute YAML Playbooks (Backend Integration)
	http.HandleFunc("/api/execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		r.ParseForm()
		playbookName := r.FormValue("playbook")
		if playbookName == "" {
			http.Error(w, "No playbook selected", http.StatusBadRequest)
			return
		}

		playbookPath := fmt.Sprintf("./playbooks/%s", playbookName)

		// Call executor package's ExecuteYAML (integrated properly)
		go executor.ExecuteYAML(playbookPath, nil)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Executing playbook: %s", playbookName)
	})

	// API Endpoint to List YAML Playbooks
	http.HandleFunc("/api/list_playbooks", func(w http.ResponseWriter, r *http.Request) {
		playbooks := listPlaybooks()
		if playbooks == nil {
			http.Error(w, "No playbooks found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(playbooks)
	})

	// Start HTTP server
	err = http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
	if err != nil {
		fmt.Printf("Web Interface failed to start: %v\n", err)
	}
}

// Helper Function to List YAML Playbooks
func listPlaybooks() []string {
	playbooksDir := "./playbooks"
	if _, err := os.Stat(playbooksDir); os.IsNotExist(err) {
		return nil
	}

	files, err := os.ReadDir(playbooksDir)
	if err != nil {
		return nil
	}

	var playbooks []string
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			playbooks = append(playbooks, file.Name())
		}
	}
	return playbooks
}
