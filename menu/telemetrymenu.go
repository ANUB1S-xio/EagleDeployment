package menu

import (
	telemetry "EagleDeployment/Telemetry"
	"fmt"
	"strings"
)

// ToggleTelemetryLevel allows the user to configure the telemetry logging level
func ToggleTelemetryLevel() {
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

// ViewLogs displays and filters logs from Telemetry
func ViewLogs() {
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
			FilterAndViewLogs(t)
		case 2:
			ToggleTelemetryLevel()
		case 3:
			ConfirmAndClearLogs(t)
		case 0:
			return
		default:
			fmt.Println("Invalid choice.")
		}
	}
}

// FilterAndViewLogs allows filtering and viewing log entries
func FilterAndViewLogs(t *telemetry.Telemetry) {
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
			event := entry

			// Extract level, message, category, and data
			level := event.Type
			message := event.Payload["message"]
			category := event.Payload["category"]
			data := event.Payload

			// Format level with color
			levelColor := "\033[0m" // Reset
			switch level {
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
				level,
				message,
				category,
			)

			// Show additional data if present
			if len(data) > 0 {
				fmt.Println("  Data:")
				for k, v := range data {
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

// ConfirmAndClearLogs confirms and clears all logs
func ConfirmAndClearLogs(t *telemetry.Telemetry) {
	fmt.Print("\nAre you sure you want to clear all logs? (y/n): ")
	var confirm string
	fmt.Scanln(&confirm)

	if strings.ToLower(confirm) == "y" {
		err := t.ClearLogs()
		if err != nil {
			fmt.Printf("Failed to clear logs: %v\n", err)
		} else {
			fmt.Println("Logs cleared successfully.")
		}
	}
}
