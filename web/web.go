package web

import (
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

// Function: StartWebServer
// Purpose: Initializes and runs the web interface server
// Parameters: None
// Returns: None
// Called By:
//   - main() during application startup
//
// Dependencies:
//   - findPort() for dynamic port allocation
//   - http package for web server functionality
//   - web/templates/* for HTML content
//   - web/static/* for static assets
//
// Notes:
//   - Default port: 8742
//   - Binds only to localhost for security
//   - Serves static files and HTML templates
//   - No HTTPS as it's for internal admin use only
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

	// Serve the homepage at "/"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/index.html")
	})

	// Serve login and dashboard pages at their respective paths
	http.HandleFunc("/login.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/login.html")
	})

	http.HandleFunc("/dashboard.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/dashboard.html")
	})

	// Start HTTP server on localhost (ED internally used (by admin), no need for secure http or CA Certificates)
	err = http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
	if err != nil {
		fmt.Printf("Web Interface failed to start: %v\n", err)
	}
}
