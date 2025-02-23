// File: main.go
// Directory Path: /EagleDeploy_CLI/
// Purpose: Main entry point for the EagleDeploy CLI, handling menu navigation and execution of YAML playbooks.

package main

import (
	"EagleDeploy_CLI/config"
	"EagleDeploy_CLI/executor"
	"EagleDeploy_CLI/inventory"
	"EagleDeploy_CLI/tasks"
	"EagleDeploy_CLI/web"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Function: listPlaybooks
// Purpose: Lists all YAML playbooks in the playbooks directory
// Parameters: None
// Returns: []string - Slice of playbook filenames
// Called By: main() when user selects option 1 or 2
// Dependencies:
//   - [`os.Stat`](os/os.go) for directory checking
//   - [`os.ReadDir`](os/os.go) for file listing
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

// Function: executeYAML
// Purpose: Processes and executes a YAML playbook with inventory data
// Parameters:
//   - playbookPath: string - Path to the source playbook
//   - targetHosts: []string - Optional list of target hosts
//
// Returns: None
// Called By: main() when user selects option 1
// Dependencies:
//   - [`inventory.InjectInventoryIntoPlaybook`](inventory/inventory.go)
//   - [`config.LoadConfig`](config/config.go)
//   - [`executor.ExecuteConcurrently`](executor/executor.go)
func executeYAML(playbookPath string, targetHosts []string) {
	// Process the playbook template by injecting inventory data
	processedPlaybook := "./playbooks/processed_add_user.yaml"
	err := inventory.InjectInventoryIntoPlaybook(playbookPath, processedPlaybook)
	if err != nil {
		log.Fatalf("Failed to inject inventory into playbook: %v", err)
	}

	// Now load and execute the processed playbook
	playbook := &tasks.Playbook{}
	err = config.LoadConfig(processedPlaybook, playbook)
	if err != nil {
		log.Fatalf("Failed to load playbook: %v", err)
	}

	// Use targetHosts if provided; otherwise, use playbook hosts
	hosts := playbook.Hosts
	if len(targetHosts) > 0 {
		hosts = targetHosts
	}

	// Port is already an integer
	port := playbook.Settings["port"]
	if port == 0 {
		log.Fatalf("Port is not specified or invalid in the playbook settings.")
	}

	// Execute tasks concurrently using the executor package
	executor.ExecuteConcurrently(playbook.Tasks, hosts, port)
}

// Function: displayMenu
// Purpose: Shows interactive CLI menu and captures user input
// Parameters: None
// Returns: int - User's menu selection
// Called By: main() in menu loop
// Dependencies: None
func displayMenu() int {
	fmt.Println()
	fmt.Println("EagleDeploy Menu:")
	fmt.Println("1. Execute a Playbook")
	fmt.Println("2. List YAML Playbooks")
	fmt.Println("3. Manage Inventory")
	fmt.Println("4. Enable/Disable Detailed Logging")
	fmt.Println("5. Rollback Changes")
	fmt.Println("6. Show Help")
	fmt.Println("0. Exit")
	fmt.Print("Select an option: ")

	var choice int
	fmt.Scanln(&choice)
	return choice
}

// Function: main
// Purpose: Entry point for CLI application
// Parameters: None
// Returns: None
// Called By: System startup
// Dependencies:
//   - [`web.StartWebServer`](web/web.go)
//   - All core package functions
func main() {
	var targetHosts []string

	// channel to monitor server lifecycle
	serverShutdown := make(chan bool, 1)

	go func() {
		web.StartWebServer()   // server start
		serverShutdown <- true // notify after server stops
	}()

	// signal handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM) // terminate signal

	//wait for web server to start first
	time.Sleep(1 * time.Second)

	go func() {
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

			case 3: // Manage Inventory
				inventory.DisplayInventoryMenu()

			case 4: // Enable/Disable Detailed Logging
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

			case 5: // Rollback Changes
				fmt.Println("Rolling back changes (not yet implemented).")

			case 6: // Show Help
				fmt.Println("Help Page:")
				fmt.Println("-e <yaml-file>: Execute the specified YAML file.")
				fmt.Println("-l <keyword>: List YAML files or related names in the EagleDeployment directory.")
				fmt.Println("-hosts <comma-separated-hosts>: Specify hosts to target (only with -e).")
				fmt.Println("-h: Display this help page.")

			case 0: // Exit
				fmt.Println("Exiting EagleDeploy.")
				serverShutdown <- true
				return

			default:
				fmt.Println("Invalid choice. Please try again.")
			}
		}
	}()

	select {
	case <-serverShutdown:
		fmt.Println("\nServer stopped...shutting down...")
	case <-signalChan:
		fmt.Println("Termination signal received...")
	}

	fmt.Println("Closing EagleDeployment...")
}
