package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LocateFiles searches for all files starting from a specified directory
func LocateFiles(args []string) {
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

	fmt.Printf("Searching for files in: %s\n", startDir)

	// Find and display all files in the directory
	err := filepath.Walk(startDir, visitFiles)
	if err != nil {
		fmt.Printf("Error while searching for files: %s\n", err)
	}
}

// visitFiles is a helper function that filters and prints files during directory traversal
func visitFiles(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	// Only process regular yml files
	if !info.IsDir() && strings.HasSuffix(info.Name(), ".yml") {
		fmt.Println(path)
	}
	return nil
}
