// File: websrv.go
// Package: websrv
// Directory: EagleDeployment/websrv
// Purpose: Handles NGINX web server setup and execution based on the host OS.

package websrv

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

// Function: StartWebServer
// Purpose: Starts the webserver by detecting the OS and running the correct commands.
// Behavior:
// - Checks the OS type.
// - Runs appropriate commands for installing, configuring, and starting NGINX.
// Returns: An error if any step fails.
func StartWebServer() error {
	osType := runtime.GOOS
	log.Printf("Detected operating system: %s", osType)

	switch osType {
	case "linux":
		return setupLinux()
	case "windows":
		return setupWindows()
	default:
		return fmt.Errorf("unsupported operating system: %s", osType)
	}
}

// Function: setupLinux
// Purpose: Sets up NGINX on Linux-based systems.
func setupLinux() error {
	// Check if NGINX is installed
	_, err := exec.LookPath("nginx")
	if err != nil {
		log.Println("NGINX not found. Installing NGINX...")
		// Install NGINX (for Debian-based systems; can add checks for others)
		cmd := "apt-get update && apt-get install -y nginx"
		output, err := runCommand(cmd)
		if err != nil {
			return fmt.Errorf("failed to install NGINX: %w\n%s", err, output)
		}
	}

	// Start NGINX
	log.Println("Starting NGINX...")
	cmd := "systemctl start nginx && systemctl enable nginx"
	output, err := runCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to start NGINX: %w\n%s", err, output)
	}

	// Optional: Check if NGINX is running
	output, err = runCommand("systemctl status nginx")
	if err != nil {
		return fmt.Errorf("NGINX failed to start: %w\n%s", err, output)
	}

	log.Println("NGINX is running.")
	return nil
}

// Function: setupWindows
// Purpose: Sets up NGINX on Windows-based systems.
func setupWindows() error {
	log.Println("Setting up NGINX on Windows...")
	// Assume NGINX is downloaded as a zip and available in the current directory
	nginxDir := "nginx"
	startScript := fmt.Sprintf("%s\\nginx.exe -c %s\\conf\\nginx.conf", nginxDir, nginxDir)

	// Start NGINX
	output, err := runCommand(startScript)
	if err != nil {
		return fmt.Errorf("failed to start NGINX on Windows: %w\n%s", err, output)
	}

	log.Println("NGINX is running on Windows.")
	return nil
}

// Function: runCommand
// Purpose: Executes a shell command and returns the output or an error.
func runCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
