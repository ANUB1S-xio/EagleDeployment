//New file for listing YAML playbooks

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func listYAMLFiles() {
	playbooksDir := "./playbooks"

	files, err := os.ReadDir(playbooksDir)
	if err != nil {
		log.Printf("Failed to read playbooks directory: %v", err)
		return
	}

	fmt.Println("\nAvailable YAML Playbooks:")
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml") {
			fmt.Printf("- %s\n", file.Name())
		}
	}
}
