package main

import (
	"fmt"
	"os"
)

func main() {
	// Check if at least two arguments are provided (tool name and command)
	if len(os.Args) < 2 {
		fmt.Println("Usage: eagle <command> [options]")
		os.Exit(1)
	}

	// Assign arguments for better readability
	toolName := os.Args[0]
	command := os.Args[1]

	// Basic tool check
	if toolName != "eagle" {
		fmt.Println("Error: Tool must be called as 'eagle'")
		os.Exit(1)
	}

	// Process command based on the second argument (command)
	switch command {
	case "-e": // Execute YAML file
		if len(os.Args) < 3 {
			fmt.Println("Error: '-e' requires a YAML file path as an additional argument.")
			os.Exit(1)
		}
		ymlFilePath := os.Args[2]
		fmt.Printf("Executing YAML file: %s\n", ymlFilePath)
		// Here you would add the logic to execute the YAML file

	case "-l": // List YAML files or related names
		if len(os.Args) < 3 {
			fmt.Println("Error: '-ll=' requires a keyword or filename to list matching YAML files.")
			os.Exit(1)
		}
		listKeyword := os.Args[2]
		fmt.Printf("Listing YAML files related to: %s\n", listKeyword)
		// Here you would add the logic to list YAML files in the stored directory

	case "-h": // Help
		fmt.Println("Help Page:")
		fmt.Println("Commands:")
		fmt.Println("-e <yaml-file>: Execute the specified YAML file.")
		fmt.Println("-l <keyword>: List YAML files or related names in the EagleDeployment directory.")
		fmt.Println("-h: Display this help page.")
	
	default:
		fmt.Printf("Error: Unknown command '%s'. Use '-h' for help.\n", command)
		os.Exit(1)
	}
}

