// File: main.go
// Directory Path: /EagleDeploy_CLI/
// Purpose: Entry point for EagleDeploy CLI with web integration and YAML playbook execution.

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
	"strconv"
	"strings"
	"syscall"
)

// listPlaybooks lists all YAML playbooks in the 'playbooks' directory.
func listPlaybooks() []string {
	playbooksDir := "./playbooks"

	if _, err := os.Stat(playbooksDir); os.IsNotExist(err) {
		log.Printf("Playbooks directory not found: %s", playbooksDir)
		return nil
	}

	files, err := os.ReadDir(playbooksDir)
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

// executeYAML handles executing tasks defined in a YAML playbook.
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

	portStr := playbook.Settings["port"]
	if portStr == "" {
		log.Fatalf("Port is not specified in playbook settings.")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid port specified: %v", err)
	}

	fmt.Printf("Executing Playbook: %s (Version: %s) on Hosts: %v\n", playbook.Name, playbook.Version, hosts)

	executor.ExecuteConcurrently(playbook.Tasks, hosts, port)
}

// displayMenu shows the CLI menu and gets user input.
func displayMenu() int {
	fmt.Println("\nEagleDeploy Menu:")
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

// main entry point starts web server and handles CLI interactions.
func main() {
	serverShutdown := make(chan bool, 1)

	go func() {
		web.StartWebServer()   // Web server starts here
		serverShutdown <- true // Notify when web server stops
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			switch displayMenu() {
			case 1:
				playbooks := listPlaybooks()
				if len(playbooks) == 0 {
					fmt.Println("No playbooks found.")
					break
				}

				fmt.Println("Available Playbooks:")
				for i, playbook := range playbooks {
					fmt.Printf("%d. %s\n", i+1, playbook)
				}

				fmt.Print("Select playbook by number: ")
				var choice int
				fmt.Scanln(&choice)

				if choice < 1 || choice > len(playbooks) {
					fmt.Println("Invalid selection.")
					break
				}

				selectedPlaybook := "./playbooks/" + playbooks[choice-1]
				executeYAML(selectedPlaybook, nil)

			case 2:
				playbooks := listPlaybooks()
				if len(playbooks) == 0 {
					fmt.Println("No playbooks available.")
				} else {
					fmt.Println("Available Playbooks:")
					for _, playbook := range playbooks {
						fmt.Printf("- %s\n", playbook)
					}
				}

			case 3:
				inventory.DisplayInventoryMenu()

			case 4:
				fmt.Print("Enable detailed logging? (y/n): ")
				var response string
				fmt.Scanln(&response)
				switch strings.ToLower(response) {
				case "y":
					fmt.Println("Detailed logging enabled.")
				case "n":
					fmt.Println("Detailed logging disabled.")
				default:
					fmt.Println("Invalid input. Logging state unchanged.")
				}

			case 5:
				fmt.Println("Rollback not implemented yet.")

			case 6:
				fmt.Println("Help:")
				fmt.Println("-e <yaml-file>: Execute YAML file.")
				fmt.Println("-l <keyword>: List YAML files or names.")
				fmt.Println("-hosts <comma-separated-hosts>: Override hosts (with -e).")
				fmt.Println("-h: Display help.")

			case 0:
				fmt.Println("Exiting EagleDeploy.")
				serverShutdown <- true
				return

			default:
				fmt.Println("Invalid choice.")
			}
		}
	}()

	select {
	case <-serverShutdown:
		fmt.Println("\nWeb server stopped, shutting down.")
	case <-signalChan:
		fmt.Println("\nTermination signal received, shutting down.")
	}

	fmt.Println("EagleDeployment terminated gracefully.")
}
