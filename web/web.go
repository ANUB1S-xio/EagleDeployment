package web

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	telemetry "EagleDeployment/Telemetry"
	"EagleDeployment/executor" // executor package
)

// Function: findPort
// Purpose: Dynamically finds an available TCP port on localhost
// Parameters: None
// Returns:
//   - int: Available port number
//   - error: Any error encountered during port discovery
//
// Called By: StartWebServer when default port is unavailable
// Dependencies:
//   - net.Listen for TCP port binding
func findPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// StartWebServer with telemetry
func StartWebServer() {
	t := telemetry.GetInstance()
	port := 8742 // Default port

	// Check if templates directory exists
	templatesDir := "web/templates"
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		absPath, _ := filepath.Abs(templatesDir)
		t.LogError("Web", "Templates directory not found", map[string]interface{}{
			"path":  absPath,
			"error": err.Error(),
		})
		fmt.Printf("ERROR: Templates directory not found: %s\n", absPath)
		return
	} else {
		// List template files for debugging
		files, _ := os.ReadDir(templatesDir)
		fileNames := []string{}
		for _, file := range files {
			fileNames = append(fileNames, file.Name())
		}
		t.LogInfo("Web", "Found template files", map[string]interface{}{
			"files": fileNames,
		})
		fmt.Printf("Found template files: %v\n", fileNames)
	}

	// Check if static directory exists
	staticDir := "web/static"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		absPath, _ := filepath.Abs(staticDir)
		t.LogError("Web", "Static directory not found", map[string]interface{}{
			"path":  absPath,
			"error": err.Error(),
		})
		fmt.Printf("ERROR: Static directory not found: %s\n", absPath)
		return
	} else {
		fmt.Printf("Found static directory: %s\n", staticDir)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.LogWarning("Web", "Default port unavailable, trying dynamic port", map[string]interface{}{
			"default_port": port,
			"error":        err.Error(),
		})

		port, err = findPort() // Get available port
		if err != nil {
			t.LogError("Web", "Failed to find available port", map[string]interface{}{
				"error": err.Error(),
			})
			fmt.Printf("Failed to find an available port: %v\n", err)
			return
		}

		t.LogInfo("Web", "Using dynamic port", map[string]interface{}{
			"port": port,
		})
	} else {
		listener.Close() // Close it since it was just a check
		t.LogInfo("Web", "Using default port", map[string]interface{}{
			"port": port,
		})
	}
	fmt.Printf("EagleDeployment GUI running at http://127.0.0.1:%d\n", port)

	// Configure HTTP handlers with logging middleware
	logRequest := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t.LogInfo("Web", "HTTP request", map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"remote": r.RemoteAddr,
			})
			fmt.Printf("HTTP Request: %s %s from %s\n", r.Method, r.URL.Path, r.RemoteAddr)
			handler(w, r)
		}
	}

	// Root handler - Redirect to dashboard
	http.HandleFunc("/", logRequest(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
		} else {
			http.NotFound(w, r)
		}
	}))

	// Serve static assets (CSS, JS, Images)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Dashboard Page
	http.HandleFunc("/dashboard", logRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/dashboard.html")
	}))

	// Execute Playbook Page
	http.HandleFunc("/execute", logRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/execute.html")
	}))

	// List YAML Playbooks Page
	http.HandleFunc("/list", logRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/list.html")
	}))

	// Login Page
	http.HandleFunc("/login", logRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/login.html")
	}))

	// API Endpoint to Execute YAML Playbooks (Backend Integration)
	http.HandleFunc("/api/execute", logRequest(func(w http.ResponseWriter, r *http.Request) {
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
	}))

	// API Endpoint to List YAML Playbooks
	http.HandleFunc("/api/list_playbooks", logRequest(func(w http.ResponseWriter, r *http.Request) {
		playbooks := listPlaybooks()
		if playbooks == nil {
			http.Error(w, "No playbooks found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(playbooks)
	}))

	// Start HTTP server
	t.LogInfo("Web", "Starting web server", map[string]interface{}{
		"address": fmt.Sprintf("127.0.0.1:%d", port),
	})
	fmt.Printf("Starting HTTP server on http://127.0.0.1:%d\n", port)
	err = http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
	if err != nil {
		t.LogError("Web", "Web server failed", map[string]interface{}{
			"error": err.Error(),
		})
		fmt.Printf("ERROR: Web Interface failed to start: %v\n", err)
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
