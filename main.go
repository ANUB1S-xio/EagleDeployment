// File: main.go
// Directory Path: /EagleDeployment/
// Purpose: Main entry point for the EagleDeploy CLI, handling menu navigation and execution of YAML playbooks.

package main

import (
	telemetry "EagleDeployment/Telemetry"
	"EagleDeployment/config"
	"EagleDeployment/executor"
	"EagleDeployment/inventory"
	"EagleDeployment/menu"
	"EagleDeployment/tasks"
	"EagleDeployment/web"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Function: listPlaybooks
// Purpose: Lists all YAML playbooks in the playbooks directory
// Parameters: None
// Returns: []string - Slice of playbook filenames
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
func executeYAML(playbookPath string, targetHosts []string) {
	t := telemetry.GetInstance()

	t.LogInfo("Playbook", "Starting playbook execution", map[string]interface{}{
		"playbook_path":      playbookPath,
		"target_hosts_count": len(targetHosts),
	})

	// Process the playbook template by injecting inventory data
	processedPlaybook := "./playbooks/processed_add_user.yaml"
	err := inventory.InjectInventoryIntoPlaybook(playbookPath, processedPlaybook)
	if err != nil {
		t.LogError("Playbook", "Failed to inject inventory into playbook", map[string]interface{}{
			"playbook_path": playbookPath,
			"error":         err.Error(),
		})
		log.Fatalf("Failed to inject inventory into playbook: %v", err)
	}

	// Now load and execute the processed playbook
	playbook := &tasks.Playbook{}
	err = config.LoadConfig(processedPlaybook, playbook)
	if err != nil {
		t.LogError("Playbook", "Failed to load playbook", map[string]interface{}{
			"processed_playbook": processedPlaybook,
			"error":              err.Error(),
		})
		log.Fatalf("Failed to load playbook: %v", err)
	}

	// Use targetHosts if provided; otherwise, use playbook hosts
	hosts := playbook.Hosts
	if len(targetHosts) > 0 {
		hosts = targetHosts
		t.LogInfo("Playbook", "Using provided target hosts", map[string]interface{}{
			"target_hosts": targetHosts,
		})
	} else {
		t.LogInfo("Playbook", "Using hosts from playbook", map[string]interface{}{
			"hosts": hosts,
		})
	}

	// Get port setting from the playbook settings
	portStr := playbook.Settings["port"]
	if portStr == "" {
		t.LogError("Playbook", "Port not specified in playbook settings", nil)
		log.Fatalf("Port is not specified in the playbook settings.")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.LogError("Playbook", "Invalid port value in settings", map[string]interface{}{
			"port_value": portStr,
			"error":      err.Error(),
		})
		log.Fatalf("Invalid port value: %v", err)
	}

	// Execute tasks concurrently using the executor package
	t.LogInfo("Playbook", "Executing tasks concurrently", map[string]interface{}{
		"tasks_count": len(playbook.Tasks),
		"hosts_count": len(hosts),
		"port":        port,
	})

	executor.ExecuteConcurrently(playbook.Tasks, hosts, port)

	t.LogInfo("Playbook", "Playbook execution completed", map[string]interface{}{
		"playbook_path": playbookPath,
	})
}

// Function: displayMenu
// Purpose: Shows interactive CLI menu and captures user input
// Parameters: None
// Returns: int - User's menu selection
func displayMenu() int {
	return menu.RunMainMenu()
}

// Function: main
// Purpose: Entry point for CLI application
func main() {
	// Initialize Telemetry
	t := telemetry.GetInstance()
	defer t.Close()

	t.LogInfo("App", "EagleDeploy starting", nil)

	var targetHosts []string

	// Channel to monitor server lifecycle
	serverShutdown := make(chan bool, 1)

	go func() {
		t.LogInfo("Web", "Starting web server", nil)
		web.StartWebServer()
		t.LogInfo("Web", "Web server stopped", nil)
		serverShutdown <- true
	}()

	// Signal handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			choice := displayMenu()
			switch choice {
			case 0: // Execute a Playbook
				playbookName, selected := menu.RunPlaybookMenu()
				if selected {
					selectedPlaybook := "./playbooks/" + playbookName
					fmt.Printf("Executing Playbook: %s\n", selectedPlaybook)

					t.LogInfo("Playbook", "Executing playbook", map[string]interface{}{
						"playbook": playbookName,
						"path":     selectedPlaybook,
					})

					executeYAML(selectedPlaybook, targetHosts)
				}

			case 1: // Manage Inventory
				t.LogInfo("Inventory", "Opening inventory menu", nil)
				menu.RunInventoryMenu()

			case 2: // View Logs
				t.LogInfo("Logs", "Viewing logs", nil)
				menu.ViewLogs()

			case -1: // User pressed q or Ctrl+C to exit
				t.LogInfo("App", "User initiated exit", nil)
				fmt.Println("Exiting EagleDeploy.")
				serverShutdown <- true
				return

			default:
				t.LogWarning("Menu", "Invalid menu choice", map[string]interface{}{
					"choice": choice,
				})
				fmt.Println("Invalid choice. Please try again.")
				time.Sleep(1 * time.Second)
			}
		}
	}()

	select {
	case <-serverShutdown:
		t.LogInfo("App", "Server shutdown detected", nil)
		fmt.Println("")
		fmt.Println("Server stopped...shutting down...")
	case <-signalChan:
		t.LogInfo("App", "Termination signal received", nil)
		fmt.Println("Termination signal received...")
	}

	fmt.Println("Closing EagleDeployment...")
	t.LogInfo("App", "EagleDeploy shutting down", nil)
}
