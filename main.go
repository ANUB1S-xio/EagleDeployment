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

	"gopkg.in/yaml.v3"
)

// struct for inventory.yml
type Inventory struct {
	Hosts []map[string]string `yaml:"Hosts"`
}

/*All struct {
		Hosts map[string]struct {
			Host string `yaml:"host"`
		} `yaml:"hosts"`
	} `yaml:"all"`
}*/

// Function: parseInventory
// Purpose: Reads and parses the inventory.yml file
func parseInventory(filePath string) (*Inventory, []string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read inventory file: %v", err)
	}

	var inventory Inventory
	err = yaml.Unmarshal(data, &inventory)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse inventory file: %v", err)
	}

	//Extract host IPs from the list
	var hosts []string
	for _, entry := range inventory.Hosts {
		for ip := range entry {
			hosts = append(hosts, ip)
		}
	}
	/*
		hosts := make(map[string]string)
		for name, hostData := range inventory.All.Hosts {
			hosts[name] = hostData.Host
		}*/
	if len(hosts) == 0 {
		return nil, nil, fmt.Errorf("no hosts found in inventory")
	}

	return &inventory, hosts, nil
}

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

// Function: executeYAML
// Purpose: Executes tasks defined in a YAML playbook on specified target hosts using concurrency.
// Parameters:
// - playbookPath: Path to the playbook file.
// - targetHosts: List of target hostnames or IPs to override default playbook hosts.
// Called By: main (when a user selects a playbook to execute).
// Succeeds: Executor functions for task execution.
func executeYAML(playbookPath string, inventoryPath string) {

	//Load inventory
	//inventory, err := parseInventory(inventoryPath)
	_, hosts, err := parseInventory(inventoryPath)
	if err != nil {
		log.Fatalf("Error loading inventory: %v", err)
	}

	//?
	/*var hosts []string
	for _, hostEntry := range inventory.All.Hosts {
		hosts = append(hosts, hostEntry.Host)
	}*/

	//Ensure at least one host exists
	if len(hosts) == 0 {
		log.Fatalf("No hosts found in inventory.")
	}
	//Convert map values to a slice of host IPs
	/*var targetHosts []string
	for _, ip := range hosts {
		targetHosts = append(targetHosts, ip)
	}
	if len(targetHosts) == 0 {
		log.Fatalf("No target hosts found in inventory file.")
	}*/

	// Load playbook into a structured format
	playbook := &tasks.Playbook{}
	err = config.LoadConfig(playbookPath, playbook)
	if err != nil {
		log.Fatalf("Failed to load playbook: %v", err)
	}

	// Validate that the playbook contains tasks
	if len(playbook.Tasks) == 0 {
		log.Fatalf("No tasks found in the playbook.")
	}

	// Use targetHosts if provided; otherwise, use playbook hosts
	/*hosts := playbook.Hosts
	if len(targetHosts) > 0 {
		hosts = targetHosts
	}*/

	// Get port setting from the playbook
	port := playbook.Settings["port"]
	if port == 0 {
		log.Fatalf("Port is not specified in the playbook settings.")
	}

	fmt.Printf("Executing Playbook: %s (Version: %s) on Hosts: %v\n", playbook.Name, playbook.Version, hosts)

	// Execute tasks concurrently using the executor package
	executor.ExecuteConcurrently(playbook.Tasks, hosts, port)
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
// Purpose: The main entry point for the EagleDeploy CLI, handling menu navigation and user actions.
// References: listPlaybooks, executeYAML, and displayMenu.
func main() {
	//var targetHosts []string

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
			//executeYAML(selectedPlaybook, targetHosts)
			executeYAML(selectedPlaybook, "./inventory.yml")

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
			fmt.Println("Managing inventory (not yet implemented).")

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
			return

		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
