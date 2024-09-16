package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LocateYAMLFiles searches for all YAML files starting from a specified directory
func LocateYAMLFiles(args []string) {
	// If no starting directory is provided, use the current directory
	startDir := "."
	if len(args) > 0 {
		startDir = args[0]
	}

	// Verify the starting directory exists
	if _, err := os.Stat(startDir); os.IsNotExist(err) {
		fmt.Printf("Directory '%s' does not exist.\n", startDir)
		return
	}

	fmt.Printf("Searching for YAML files in: %s\n", startDir)

	// Find and display all YAML files in the directory
	err := filepath.Walk(startDir, visitYAMLFiles)
	if err != nil {
		fmt.Printf("Error while searching for files: %s\n", err)
	}
}

// visitYAMLFiles is a helper function that filters and prints YAML files during directory traversal
func visitYAMLFiles(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	// Only process regular files
	if !info.IsDir() && strings.HasSuffix(info.Name(), ".yml") {
		fmt.Println(path)
	}
	return nil
}
