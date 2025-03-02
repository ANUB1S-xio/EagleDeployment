package web

import (
	"EagleDeploy_CLI/Telemetry"
	"fmt"
	"net"
	"net/http"
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
	listener, err := net.Listen("tcp", "127.0.0.1:0") // Bind to localhost only
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	// Extract assigned port
	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// StartWebServer with telemetry
func StartWebServer() {
	t := Telemetry.GetInstance()
	port := 8742 // Default port

	// Check default port availability
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

	// Display assigned port
	fmt.Printf("EagleDeployment GUI running at http://127.0.0.1:%d\n", port)

	// Configure HTTP handlers
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Log each HTTP request
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

	// Apply logging middleware
	http.HandleFunc("/", logRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/index.html")
	}))

	http.HandleFunc("/login.html", logRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/login.html")
	}))

	http.HandleFunc("/dashboard.html", logRequest(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/dashboard.html")
	}))

	// Start HTTP server
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
