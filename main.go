// File: main.go
// Directory Path: /EagleDeploy_CLI/

package main

import (
	"EagleDeploy_CLI/config"
	"EagleDeploy_CLI/executor"
	"EagleDeploy_CLI/tasks"
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Function: init
// Purpose: Initializes the environment by loading .env file and setting up debugging for environment variables.
func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Print the environment variables
	fmt.Printf("USER_1_USERNAME: %s\n", os.Getenv("USER_1_USERNAME"))
	fmt.Printf("USER_1_PASSWORD: %s\n", os.Getenv("USER_1_PASSWORD"))

	// Debugging environment variables
	log.Printf("USER_1_USERNAME: %s", os.Getenv("USER_1_USERNAME"))
	log.Printf("USER_1_PASSWORD: %s", os.Getenv("USER_1_PASSWORD"))
}

// Function: listPlaybooks
// Purpose: Lists all YAML playbooks in the 'playbooks' directory.
// Returns: A slice of strings containing the names of the playbooks.
func listPlaybooks() []string {
	playbooksDir := "./playbooks"

	// Ensure the playbooks directory exists
	if _, err := os.Stat(playbooksDir); os.IsNotExist(err) {
		log.Printf("Playbooks directory not found: %s", playbooksDir)
		return nil
	}

	files, err := ioutil.ReadDir(playbooksDir)
	if err != nil {
		log.Printf("Failed to read playbooks directory: %v", err)
		return nil
	}

	var playbooks []string
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			playbooks = append(playbooks, file.Name())
		}
	}
	return playbooks
}

// Function: executeYAML
// Purpose: Executes the tasks defined in a YAML playbook on specified target hosts.
// Parameters:
// - playbookPath: The file path to the playbook.
// - targetHosts: A slice of strings containing target hostnames or IPs.
func executeYAML(playbookPath string, targetHosts []string) {
	playbook := &tasks.Playbook{}
	err := config.LoadConfig(playbookPath, playbook)
	if err != nil {
		log.Fatalf("Failed to load playbook: %v", err)
	}

	if len(playbook.Tasks) == 0 {
		log.Fatalf("No tasks found in the playbook.")
	}

	hosts := playbook.Hosts
	if len(targetHosts) > 0 {
		hosts = targetHosts
	}

	port := playbook.Settings["port"]
	if port == 0 {
		log.Fatalf("Port is not specified in the playbook settings.")
	}

	fmt.Printf("Executing Playbook: %s (Version: %s) on Hosts: %v\n", playbook.Name, playbook.Version, hosts)
	for _, task := range playbook.Tasks {
		for _, host := range hosts {
			task.Host = host // Assign the current host to the task
			if task.SSHUser != "" {
				log.Printf("Executing remote task: %s on host %s", task.Name, host)
				err = executor.ExecuteRemote(task, port)
			} else {
				log.Printf("Executing local task: %s", task.Name)
				err = executor.ExecuteLocal(task.Command)
			}
			if err != nil {
				log.Printf("Task '%s' failed on host %s: %v", task.Name, host, err)
			}
		}
	}
}

// Function: displayMenu
// Purpose: Displays the interactive menu for the EagleDeploy CLI.
// Returns: The user's menu choice as an integer.
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
// Purpose: The main entry point of the application, handling the interactive menu and user actions.
func main() {
	reader := bufio.NewReader(os.Stdin)
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

		case 3: // Manage Inventory
			fmt.Println("Managing inventory (not yet implemented).")

		case 4: // Enable/Disable Detailed Logging
			fmt.Print("Enable detailed logging? (y/n): ")
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(answer)
			if strings.ToLower(answer) == "y" {
				fmt.Println("Detailed logging enabled.")
			} else if strings.ToLower(answer) == "n" {
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
