// File: utils.go
// Package: websrv
// Directory: EagleDeployment/websrv
// Purpose: Provides utility functions for managing NGINX configurations and commands.

package websrv

import (
	"fmt"
	"os"
	"os/exec"
)

// Function: DefaultNginxConfig
// Purpose: Returns a default NGINX configuration as a string.
// Parameters:
// - rootPath: The root directory for the web server.
// Returns: The NGINX configuration content.
func DefaultNginxConfig(rootPath string) string {
	return fmt.Sprintf(`
server {
    listen 80;
    server_name localhost;

    root %s;
    index index.html;

    location / {
        try_files $uri $uri/ =404;
    }
}`, rootPath)
}

// Function: WriteToFile
// Purpose: Writes content to a specified file path.
// Parameters:
// - filePath: The full path to the file.
// - content: The content to write into the file.
// Returns: An error if the write operation fails.
func WriteToFile(filePath, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filePath, err)
	}

	return nil
}

// Function: RunCommand
// Purpose: Executes a shell command locally and returns the output or an error.
// Parameters:
// - command: The command to execute.
// Returns: The command output or an error if execution fails.
func RunCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Function: FileExists
// Purpose: Checks if a file exists at the specified path.
// Parameters:
// - filePath: The full path to the file.
// Returns: True if the file exists, false otherwise.
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
