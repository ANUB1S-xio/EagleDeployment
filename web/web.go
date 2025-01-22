package web

import (
	"fmt"
	"net/http"
)

// StartWebServer starts the web server with TLS on a predefined port
func StartWebServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "EagleDeployment Web Interface is running securely with TLS!")
		// Add more routes as needed
	})

	port := 8443
	fmt.Printf("TLS web server running at https://localhost:%d\n", port)

	// Use a self-signed certificate (replace with paths to your cert and key files)
	certFile := "cert.pem" // Path to the certificate file
	keyFile := "key.pem"   // Path to the private key file

	err := http.ListenAndServeTLS(fmt.Sprintf(":%d", port), certFile, keyFile, nil)
	if err != nil {
		fmt.Printf("Failed to start web server with TLS: %v\n", err)
	}
}
