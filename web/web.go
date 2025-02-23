package web

import (
	"fmt"
	"net"
	"net/http"
)

// find open port
func findPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0") // Bind to localhost only
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	// Extract assigned port
	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// StartWebServer initializes server with dynamic port signing
func StartWebServer() {
	port := 8742 // Default port

	// Check default port availability
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		fmt.Println("The chosen port (8742) is taken, attempting dynamic port signing...")
		port, err = findPort() // Get available port
		if err != nil {
			fmt.Printf("Failed to find an available port: %v\n", err)
			return
		}
	} else {
		listener.Close() // Close it since it was just a check
	}

	// Display assigned port
	fmt.Printf("EagleDeployment GUI running at http://127.0.0.1:%d\n", port)

	// Serve React frontend
	//fs := http.FileServer(http.Dir("web/frontend/build"))
	//http.Handle("/", fs)
	// Serve static files
	// Serve static files (CSS, JS, images)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Default to Login Page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/login.html")
	})

	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
    	// Authentication check logic goes here
    		http.ServeFile(w, r, "web/templates/dashboard.html")
	})
	// Execute playbook page
	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/execute.html")
	})
	// list playbook page
	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/list.html")
	})

	// API endpoint for executing YAML playbooks
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
		go executeYAML(playbookPath, nil)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Executing playbook: %s", playbookName)
	})

	// API endpoint to list available YAML playbooks
	http.HandleFunc("/api/list_playbooks", func(w http.ResponseWriter, r *http.Request) {
		playbooks := listPlaybooks()
		if playbooks == nil {
			http.Error(w, "No playbooks found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(playbooks)
	})



	// Start HTTP server on localhost (ED internally used (by admin), no need for secure http or CA Certificates)
	err = http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
	if err != nil {
		fmt.Printf("Web Interface failed to start: %v\n", err)
	}
}


// Helper function: listPlaybooks (copy from main.go)
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

// Placeholder for executeYAML (move or call from main.go appropriately later)
func executeYAML(playbookPath string, targetHosts []string) {
	fmt.Printf("Executing playbook: %s\n", playbookPath)
	// Future integration with your main.go executeYAML logic here
}
