//New file for listing YAML playbooks
// File: list_yaml.go
// Directory Path: /EagleDeploy_CLI/
// Purpose: Standalone function to list YAML playbooks.

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Function: ListAllYAMLFiles
// Purpose: Lists all YAML playbooks in the 'playbooks' directory.
func ListAllYAMLFiles() {
	playbooksDir := "./playbooks"

	// Ensure the directory exists
	if _, err := os.Stat(playbooksDir); os.IsNotExist(err) {
		log.Printf("Playbooks directory not found: %s", playbooksDir)
		return
	}

	// Read directory contents
	files, err := os.ReadDir(playbooksDir)
	if err != nil {
		log.Printf("Failed to read playbooks directory: %v", err)
		return
	}

	// Print playbooks in YAML format
	fmt.Println("\nAvailable YAML Playbooks:")
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			fmt.Printf("- %s\n", file.Name())
		}
	}
}
