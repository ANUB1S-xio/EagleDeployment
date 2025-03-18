package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings" // Ensure this is imported

	telemetry "EagleDeployment/Telemetry"
)

// AdminUser struct for login authentication
type AdminUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var adminUsers []AdminUser

// Function: findPort - Finds an available TCP port
func findPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// LoadAdminUsers - Reads users from logs/web_logins.json
func LoadAdminUsers() {
	file, err := os.Open("logs/web_logins.json")
	if err != nil {
		fmt.Println("Error opening web_logins.json:", err)
		return
	}
	defer file.Close()

	bytes, _ := ioutil.ReadAll(file)
	err = json.Unmarshal(bytes, &adminUsers)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	fmt.Println("Loaded Users from JSON:", adminUsers)
}

// LoginHandler - Handles user login requests
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var creds AdminUser
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		fmt.Println("Error decoding request:", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	fmt.Println("Received login attempt:", creds.Username, creds.Password)

	// Debugging: Compare input with stored credentials
	for _, user := range adminUsers {
		fmt.Printf("Checking: [Input: %s | %s] vs [Stored: %s | %s]\n",
			creds.Username, creds.Password, user.Username, user.Password)

		if user.Username == creds.Username && user.Password == creds.Password {
			fmt.Println("Login successful for:", creds.Username)
			response := map[string]string{"status": "success", "redirect": "/menu.html"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	fmt.Println("Login failed for:", creds.Username)
	response := map[string]string{"status": "failed", "message": "Invalid username or password"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartWebServer - Starts HTTP server with telemetry logging
func StartWebServer() {
	t := telemetry.GetInstance()
	port := 8742

	// Load admin users before starting the web server
	fmt.Println("Loading admin users...")
	LoadAdminUsers()
	fmt.Println("Loaded Users from JSON:", adminUsers)

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.LogWarning("Web", "Default port unavailable, trying dynamic port", map[string]interface{}{
			"default_port": port,
			"error":        err.Error(),
		})

		port, err = findPort()
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
		listener.Close()
		t.LogInfo("Web", "Using default port", map[string]interface{}{
			"port": port,
		})
	}

	fmt.Printf("EagleDeployment GUI running at http://127.0.0.1:%d\n", port)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	logRequest := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t.LogInfo("Web", "HTTP request", map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"remote": r.RemoteAddr,
			})
			handler(w, r)
		}
	}

	http.HandleFunc("/", logRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/index.html")
	}))

	http.HandleFunc("/login.html", logRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/login.html")
	}))

	http.HandleFunc("/dashboard.html", logRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/dashboard.html")
	}))

	// Removing duplicate /menu.html registration
	http.HandleFunc("/menu.html", logRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/menu.html")
	}))

	http.HandleFunc("/inventory.html", logRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/inventory.html")
	}))

	http.HandleFunc("/logs.html", logRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/logs.html")
	}))

	http.HandleFunc("/api/login", logRequest(LoginHandler)) // API for login authentication

	// ✅ Adding the API route for listing users
	http.HandleFunc("/api/list_users", logRequest(ListUsersHandler))

	t.LogInfo("Web", "Starting web server", map[string]interface{}{
		"address": fmt.Sprintf("127.0.0.1:%d", port),
	})

	err = http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
	if err != nil {
		t.LogError("Web", "Web server failed", map[string]interface{}{
			"error": err.Error(),
		})
		fmt.Printf("Web Interface failed to start: %v\n", err)
	}
}

// ListUsersHandler - Executes the playbook and returns the list of users
func ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("ansible-playbook", "playbooks/List_Users.yaml")
	output, err := cmd.CombinedOutput()

	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing playbook: %v", err), http.StatusInternalServerError)
		return
	}

	users := parseUsersFromOutput(string(output))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Helper function to extract users from output
func parseUsersFromOutput(output string) []string {
	var users []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line != "" && !strings.Contains(line, "TASK") {
			users = append(users, line)
		}
	}
	return users
}
