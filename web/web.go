package web

import (
	"fmt"
	"net/http"
)

// StartWebServer starts the web server with TLS on a predefined port
func StartWebServer() {
	//handle request to root URL
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "EagleDeployment Web Interface is running...")

		// Add more routes as needed (api endpoints)

	})

	port := 8742
	fmt.Printf("EagleDeployment GUI running at https://localhost:%d\n", port)

	// Use a self-signed certificate (replace with paths to your cert and key files)
	certFile := "web/cert.pem" // Path to the certificate file
	keyFile := "web/key.pem"   // Path to the private key file

	err := http.ListenAndServeTLS(fmt.Sprintf(":%d", port), certFile, keyFile, nil)
	if err != nil {
		fmt.Printf("Web Interface failed to start: %v\n", err)
	}
}
