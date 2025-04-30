package web

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	telemetry "EagleDeployment/Telemetry"
	"EagleDeployment/executor" // executor package
)

// Store the current port for other parts of the application to access
var (
	currentPort   int
	portMutex     sync.RWMutex
	serverRunning bool
)

// GetServerPort returns the current server port
func GetServerPort() int {
	portMutex.RLock()
	defer portMutex.RUnlock()
	return currentPort
}

// IsServerRunning returns whether the server is currently running
func IsServerRunning() bool {
	return serverRunning
}

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
	// Try the default port first (8742)
	listener, err := net.Listen("tcp", "127.0.0.1:8742")
	if err == nil {
		listener.Close()
		return 8742, nil
	}

	// If default port is unavailable, get a random one
	listener, err = net.Listen("tcp", "127.0.0.1:0")
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

	// Find an available port dynamically
	var err error
	port, err := findPort()
	if err != nil {
		t.LogError("Web", "Failed to find available port", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Update the current port for external access
	portMutex.Lock()
	currentPort = port
	portMutex.Unlock()

	// Create a file indicating which port we're using for other processes
	// Write this EARLY so other processes can find it
	portFile := "web/port.txt"
	os.MkdirAll(filepath.Dir(portFile), 0755) // Ensure the directory exists
	err = os.WriteFile(portFile, []byte(fmt.Sprintf("%d", port)), 0644)
	if err != nil {
		t.LogError("Web", "Failed to write port file", map[string]interface{}{
			"error": err.Error(),
		})
		// Continue anyway, it's not critical
	}

	// Check if templates directory exists
	templatesDir := "web/templates"
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		absPath, _ := filepath.Abs(templatesDir)
		t.LogError("Web", "Templates directory not found", map[string]interface{}{
			"path":  absPath,
			"error": err.Error(),
		})
		return
	} else {
		_, err = os.ReadDir(templatesDir)
		if err != nil {
			t.LogError("Web", "Failed to read templates directory", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
	}

	// Check if static directory exists
	staticDir := "web/static"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		absPath, _ := filepath.Abs(staticDir)
		t.LogError("Web", "Static directory not found", map[string]interface{}{
			"path":  absPath,
			"error": err.Error(),
		})
		return
	}

	fmt.Printf("\n====================================\n")
	fmt.Printf("EagleDeployment GUI running at http://127.0.0.1:%d\n", port)
	fmt.Printf("====================================\n\n")

	// Configure HTTP handlers with logging middleware
	logRequest := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t.LogInfo("Web", "HTTP request", map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
			})
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

	// Health check endpoint for server status verification - make this very simple and robust
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Skip logging to make health checks lighter
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		response := map[string]interface{}{
			"status": "ok",
			"port":   port,
			"time":   time.Now().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

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
	// API Endpoint to Create a New Playbook
	http.HandleFunc("/api/create_playbook", logRequest(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		var data struct {
			Filename string `json:"filename"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		// Sanitize filename
		if data.Filename == "" || strings.Contains(data.Filename, "..") {
			http.Error(w, "Invalid filename", http.StatusBadRequest)
			return
		}
		path := fmt.Sprintf("./playbooks/%s", data.Filename)
		if _, err := os.Stat(path); err == nil {
			http.Error(w, "File already exists", http.StatusConflict)
			return
		}
		err := os.WriteFile(path, []byte("# New playbook\n\n"), 0644)
		if err != nil {
			http.Error(w, "Failed to create file", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Playbook '%s' created successfully!", data.Filename)
	}))


	// Serve raw YAML playbooks for viewing/editing in list.html
	http.Handle("/playbooks/", http.StripPrefix("/playbooks/", http.FileServer(http.Dir("playbooks"))))

	// API Endpoint to Save YAML Playbooks 
	http.HandleFunc("/api/save_playbook", logRequest(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Decode JSON from request
		var payload struct {
			Filename string `json:"filename"`
			Content  string `json:"content"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Failed to decode request body", http.StatusBadRequest)
			return
		}

		// Write to file
		path := fmt.Sprintf("./playbooks/%s", payload.Filename)
		if err := os.WriteFile(path, []byte(payload.Content), 0644); err != nil {
			http.Error(w, "Failed to save playbook", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Playbook saved successfully.")
	}))

	// Mark server as running before starting
	serverRunning = true

	// Start HTTP server
	t.LogInfo("Web", "Starting web server", map[string]interface{}{
		"address": fmt.Sprintf("127.0.0.1:%d", port),
	})
	err = http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
	if err != nil {
		t.LogError("Web", "Web server failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Mark server as stopped when it exits
	serverRunning = false
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
