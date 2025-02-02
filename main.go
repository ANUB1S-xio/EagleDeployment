// File: main.go
// Directory Path: /EagleDeploy_CLI/
// Purpose: Main entry point for the EagleDeploy CLI, handling menu navigation and execution of YAML playbooks.

package main

import (
	"EagleDeploy_CLI/config"
	"EagleDeploy_CLI/executor"
	"EagleDeploy_CLI/tasks"
	"fmt"
	"log"
	"os"
	"strings"
)

// Function: listPlaybooks
// Purpose: Lists all YAML playbooks in the 'playbooks' directory.
// Returns: A slice of strings containing the names of the playbooks.
// Precedes: executeYAML (called when a user selects a playbook to execute).
func listPlaybooks() []string {
	playbooksDir := "./playbooks" // Default directory for playbooks

	// Ensure the playbooks directory exists
	if _, err := os.Stat(playbooksDir); os.IsNotExist(err) {
		log.Printf("Playbooks directory not found: %s", playbooksDir)
		return nil
	}

	// Read the playbooks directory
	files, err := os.ReadDir(playbooksDir)
	if err != nil {
		log.Printf("Failed to read playbooks directory: %v", err)
		return nil
	}

	// Filter files to include only YAML playbooks
	var playbooks []string
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			playbooks = append(playbooks, file.Name())
		}
	}
	return playbooks
}

// Function: listYAMLFiles
// Purpose: Lists all YAML playbooks in the 'playbooks' directory.
func listYAMLFiles() {
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
        // Print all YAML files (not just playbooks)
	fmt.Println("\n🔍 Checking for available YAML files...")
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
  			fmt.Printf("- %s\n", file.Name())
		}
	}
}

// Function: executeYAML
// Purpose: Executes tasks in a playbook and lists available YAML files afterward.
func executeYAML(playbookPath string, targetHosts []string) {
	playbook := &tasks.Playbook{}
	err := config.LoadConfig(playbookPath, playbook)
	if err != nil {
		log.Fatalf("Failed to load playbook: %v", err)
	}

	// Validate playbook tasks
	if len(playbook.Tasks) == 0 {
		log.Fatalf("No tasks found in the playbook.")
	}

	// Use target hosts if provided
	hosts := playbook.Hosts
	if len(targetHosts) > 0 {
		hosts = targetHosts
	}

	// Get port setting
	port := playbook.Settings["port"]
	if port == 0 {
		log.Fatalf("Port is not specified in the playbook settings.")
	}

	fmt.Printf("Executing Playbook: %s (Version: %s) on Hosts: %v\n", playbook.Name, playbook.Version, hosts)

	// Execute tasks concurrently
	executor.ExecuteConcurrently(playbook.Tasks, hosts, port)

	// **NEW: Automatically List YAML Playbooks After Execution**
	fmt.Println("\n✅ Playbook Execution Completed! Listing Available Playbooks:")
	listYAMLFiles()
}

// Function: displayMenu
// Purpose: Displays an interactive menu for the EagleDeploy CLI.
// Returns: The user's menu choice as an integer.
// Precedes: main (used for interactive user input).
func displayMenu() int {
	fmt.Println()
	fmt.Println("EagleDeploy Menu:")
	fmt.Println("1. Execute a Playbook")
	fmt.Println("2. List YAML Playbooks")
	fmt.Println("3. List YAML Files")
	fmt.Println("4. Manage Inventory")
	fmt.Println("5. Enable/Disable Detailed Logging")
	fmt.Println("6. Rollback Changes")
	fmt.Println("7. Show Help")
	fmt.Println("0. Exit")
	fmt.Print("Select an option: ")

	var choice int
	fmt.Scanln(&choice)
	return choice
}

// Function: main
// Purpose: The main entry point for the EagleDeploy CLI, handling menu navigation and user actions.
// References: listPlaybooks, executeYAML, and displayMenu.
func main() {
	var targetHosts []string

	for {
		choice := displayMenu()
		switch choice {
		case 1: // Execute a Playbook
			playbooks := listPlaybooks()
			if len(playbooks) == 0 {
				fmt.Println("No playbooks found in the 'playbooks' directory.")
				break
			}

			fmt.Println("Available Playbooks:")
			for i, playbook := range playbooks {
				fmt.Printf("%d. %s\n", i+1, playbook)
			}

			fmt.Print("Select a playbook to execute by number: ")
			var choice int
			fmt.Scanln(&choice)

			if choice < 1 || choice > len(playbooks) {
				fmt.Println("Invalid choice. Returning to the menu.")
				break
			}

			selectedPlaybook := "./playbooks/" + playbooks[choice-1]
			fmt.Printf("Executing Playbook: %s\n", selectedPlaybook)
			executeYAML(selectedPlaybook, targetHosts)

		case 2: // List YAML Playbooks
			playbooks := listPlaybooks()
			if len(playbooks) == 0 {
				fmt.Println("No playbooks found in the 'playbooks' directory.")
			} else {
				fmt.Println("Available Playbooks:")
				for _, playbook := range playbooks {
					fmt.Printf("- %s\n", playbook)
				}
			}
		
                case 3: // List YAML Files
			fmt.Println("\n🔍 Checking for available YAML files...")
                         listYAMLFiles()


		case 4: // Manage Inventory
			fmt.Println("Managing inventory (not yet implemented).")

		case 5: // Enable/Disable Detailed Logging
			fmt.Print("Enable detailed logging? (y/n): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) == "y" {
				fmt.Println("Detailed logging enabled.")
			} else if strings.ToLower(response) == "n" {
				fmt.Println("Detailed logging disabled.")
			} else {
				fmt.Println("Invalid input. Logging state unchanged.")
			}

		case 6: // Rollback Changes
			fmt.Println("Rolling back changes (not yet implemented).")

		case 7: // Show Help
			fmt.Println("Help Page:")
			fmt.Println("-e <yaml-file>: Execute the specified YAML file.")
			fmt.Println("-l <keyword>: List YAML files or related names in the EagleDeployment directory.")
			fmt.Println("-hosts <comma-separated-hosts>: Specify hosts to target (only with -e).")
			fmt.Println("-h: Display this help page.")

		case 0: // Exit
			fmt.Println("Exiting EagleDeploy.")
			return

		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
