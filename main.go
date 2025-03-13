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

	// Remove the redundant declaration and use assignment instead
	if port == 0 {
		log.Fatalf("Port is not specified or invalid in the playbook settings.")
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
// Called By: main() in menu loop
// Dependencies: None
func displayMenu() int {
	return menu.RunMainMenu()
}

// Add this function to toggle Telemetry levels
func toggleTelemetryLevel() {
	t := telemetry.GetInstance()

	fmt.Println("\nSelect Telemetry Level:")
	fmt.Println("1. Error only")
	fmt.Println("2. Error + Warning")
	fmt.Println("3. Error + Warning + Info")
	fmt.Println("4. All (Debug mode)")

	var choice int
	fmt.Print("\nEnter choice: ")
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		t.SetLevel(telemetry.LevelError)
		fmt.Println("Telemetry set to Error level")
	case 2:
		t.SetLevel(telemetry.LevelWarning)
		fmt.Println("Telemetry set to Warning level")
	case 3:
		t.SetLevel(telemetry.LevelInfo)
		fmt.Println("Telemetry set to Info level")
	case 4:
		t.SetLevel(telemetry.LevelDebug)
		fmt.Println("Telemetry set to Debug level")
	default:
		fmt.Println("Invalid choice. Telemetry level unchanged.")
	}
}

// Function: viewLogs
// Purpose: Display and filter logs from Telemetry
// Parameters: None
// Returns: None
// Called By: main() when user selects the "View Logs" option
func viewLogs() {
	t := telemetry.GetInstance()

	// Create a submenu for log viewing options
	for {
		fmt.Println("\nLog Management:")
		fmt.Println("1. View Logs")
		fmt.Println("2. Configure Telemetry Level")
		fmt.Println("3. Clear Logs")
		fmt.Println("0. Return to Main Menu")

		var choice int
		fmt.Print("\nEnter choice: ")
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			filterAndViewLogs(t)
		case 2:
			configureTelemetryLevel() // Use the configureTelemetryLevel function here
		case 3:
			confirmAndClearLogs() // Remove the unused parameter t
		case 0:
			return
		default:
			fmt.Println("Invalid choice.")
		}
	}
}

// Function: filterAndViewLogs
// Purpose: Allow filtering and viewing log entries
// Parameters:
//   - t: *telemetry.Telemetry - Telemetry instance
//
// Returns: None
func filterAndViewLogs(t *telemetry.Telemetry) {
	var levelFilter, categoryFilter, messageFilter string
	var limit int = 50

	fmt.Println("\nLog Filtering Options:")

	fmt.Print("Level filter (ERROR, WARNING, INFO, DEBUG) or empty for all: ")
	fmt.Scanln(&levelFilter)
	levelFilter = strings.ToUpper(levelFilter)

	fmt.Print("Category filter or empty for all: ")
	fmt.Scanln(&categoryFilter)

	fmt.Print("Message contains (or empty for all): ")
	fmt.Scanln(&messageFilter)

	fmt.Print("Number of entries to show (default 50): ")
	fmt.Scanln(&limit)
	if limit <= 0 {
		limit = 50
	}

	// Get filtered entries
	entries := t.FilterLogs(levelFilter, categoryFilter, messageFilter, limit)

	if len(entries) == 0 {
		fmt.Println("No log entries match your filters.")
		return
	}

	// Display logs with paging
	pageSize := 10
	totalPages := (len(entries) + pageSize - 1) / pageSize
	currentPage := 1

	for {
		// Calculate page boundaries
		start := (currentPage - 1) * pageSize
		end := start + pageSize
		if end > len(entries) {
			end = len(entries)
		}

		// Clear screen
		fmt.Print("\033[H\033[2J")

		// Show filter info
		fmt.Println("Applied filters:")
		if levelFilter != "" {
			fmt.Printf("- Level: %s\n", levelFilter)
		}
		if categoryFilter != "" {
			fmt.Printf("- Category: %s\n", categoryFilter)
		}
		if messageFilter != "" {
			fmt.Printf("- Message contains: %s\n", messageFilter)
		}

		fmt.Printf("\nShowing entries %d-%d of %d (Page %d/%d)\n\n",
			start+1, end, len(entries), currentPage, totalPages)

		// Display entries
		for i := start; i < end; i++ {
			entry := entries[i]
			event := entry.Event

			// Format level with color
			levelColor := "\033[0m" // Reset
			switch event.Level {
			case "ERROR":
				levelColor = "\033[31m" // Red
			case "WARNING":
				levelColor = "\033[33m" // Yellow
			case "INFO":
				levelColor = "\033[36m" // Cyan
			case "DEBUG":
				levelColor = "\033[35m" // Magenta
			}

			// Display entry
			fmt.Printf("[%s] %s%s\033[0m: %s (%s)\n",
				event.Timestamp.Format("2006-01-02 15:04:05"),
				levelColor,
				event.Level,
				event.Message,
				event.Category,
			)

			// Show additional data if present
			if len(event.Data) > 0 {
				fmt.Println("  Data:")
				for k, v := range event.Data {
					fmt.Printf("    %s: %v\n", k, v)
				}
			}

			// Add separator between entries
			fmt.Println(strings.Repeat("-", 80))
		}

		// Navigation instructions
		fmt.Println("\nNavigation: [n]ext page, [p]revious page, [f]ilter again, [q]uit")
		var input string
		fmt.Scanln(&input)

		switch strings.ToLower(input) {
		case "n":
			if currentPage < totalPages {
				currentPage++
			}
		case "p":
			if currentPage > 1 {
				currentPage--
			}
		case "f":
			return
		case "q":
			return
		}
	}
}

// Function: configureTelemetryLevel
// Purpose: Configure the Telemetry logging level
// Parameters: None
// Returns: None
func configureTelemetryLevel() {
	t := telemetry.GetInstance()

	fmt.Println("\nSelect Telemetry Level:")
	fmt.Println("1. Error only")
	fmt.Println("2. Error + Warning")
	fmt.Println("3. Error + Warning + Info")
	fmt.Println("4. All (Debug mode)")

	// We don't need to get the current level if GetLevel() isn't available
	fmt.Printf("\nCurrent level: Setting a new level will override the current one")

	var choice int
	fmt.Print("\nEnter choice: ")
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		t.SetLevel(telemetry.LevelError)
		fmt.Println("Telemetry set to Error level")
	case 2:
		t.SetLevel(telemetry.LevelWarning)
		fmt.Println("Telemetry set to Warning level")
	case 3:
		t.SetLevel(telemetry.LevelInfo)
		fmt.Println("Telemetry set to Info level")
	case 4:
		t.SetLevel(telemetry.LevelDebug)
		fmt.Println("Telemetry set to Debug level")
	default:
		fmt.Println("Invalid choice. Telemetry level unchanged.")
	}
}

// Function: confirmAndClearLogs
// Purpose: Confirm and clear all logs
// Parameters: None
// Returns: None
func confirmAndClearLogs() {
	fmt.Print("\nAre you sure you want to clear all logs? (y/n): ")
	var confirm string
	fmt.Scanln(&confirm)

	if strings.ToLower(confirm) == "y" {
		// Since ClearLogs is not implemented, we'll just print a message
		fmt.Println("Log clearing functionality not yet implemented.")
		// TODO: Implement ClearLogs method in the Telemetry package
		/*
		   err := t.ClearLogs()
		   if err != nil {
		       fmt.Printf("Failed to clear logs: %v\n", err)
		   } else {
		       fmt.Println("Logs cleared successfully.")
		   }
		*/
	}
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
	// Initialize Telemetry
	t := telemetry.GetInstance()
	defer t.Close()

	t.LogInfo("App", "EagleDeploy starting", nil)

	fmt.Println()
	var targetHosts []string

	// channel to monitor server lifecycle
	serverShutdown := make(chan bool, 1)
	go func() {
		t.LogInfo("Web", "Starting web server", nil)
		web.StartWebServer()
		t.LogInfo("Web", "Web server stopped", nil)
		serverShutdown <- true
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

			case 1: // Manage Inventory (was option 2)
				t.LogInfo("Inventory", "Opening inventory menu", nil)
				menu.RunInventoryMenu()

			case 2: // View Logs (replacing Enable/Disable Detailed Logging)
				t.LogInfo("Settings", "Opening log viewer", nil)
				viewLogs()

			case 3: // List YAML Playbooks
				playbooks := listPlaybooks()
				if len(playbooks) == 0 {
					fmt.Println("No playbooks found in the 'playbooks' directory.")
				} else {
					fmt.Println("Available Playbooks:")
					for _, playbook := range playbooks {
						fmt.Printf("- %s\n", playbook)
					}
				}

			case 4: // Enable/Disable Detailed Logging
				toggleTelemetryLevel() // Use the toggleTelemetryLevel function here

			case 5: // Rollback Changes
				fmt.Println("Rolling back changes (not yet implemented).")

			case 6: // Show Help
				fmt.Println("Help Page:")
				fmt.Println("-e <yaml-file>: Execute the specified YAML file.")
				fmt.Println("-l <keyword>: List YAML files or related names in the EagleDeployment directory.")
				fmt.Println("-hosts <comma-separated-hosts>: Specify hosts to target (only with -e).")
				fmt.Println("-h: Display this help page.")

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
		fmt.Println("Server stopped...shutting down...")
	case <-signalChan:
		t.LogInfo("App", "Termination signal received", nil)
		fmt.Println("Termination signal received...")
	}

	fmt.Println("Closing EagleDeployment...")
	t.LogInfo("App", "EagleDeploy shutting down", nil)
}
