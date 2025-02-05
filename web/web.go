package web

import (
	"fmt"
	"net"
	"net/http"
)

// findAvailablePort finds an open port dynamically
func findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0") // Bind to localhost only
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	// Extract the assigned port
	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// StartWebServer initializes the web server with dynamic port signing
func StartWebServer() {
	port := 8742 // Default port

	// Check if port 8742 is available
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		fmt.Println("The chosen port (8742) is taken, attempting dynamic port signing...")
		port, err = findAvailablePort() // Get a new available port
		if err != nil {
			fmt.Printf("Failed to find an available port: %v\n", err)
			return
		}
	} else {
		listener.Close() // Close it since it was just a check
	}

	// Display assigned port
	fmt.Printf("EagleDeployment GUI running at https://127.0.0.1:%d\n", port)

	// Serve React frontend
	fs := http.FileServer(http.Dir("web/frontend/build"))
	http.Handle("/", fs)

	// Start HTTPS server on localhost (not publicly exposed)
	err = http.ListenAndServeTLS(fmt.Sprintf("127.0.0.1:%d", port), "web/cert.pem", "web/key.pem", nil)
	if err != nil {
		fmt.Printf("Web Interface failed to start: %v\n", err)
	}
}
