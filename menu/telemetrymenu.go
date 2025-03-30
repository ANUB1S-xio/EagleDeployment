package menu

import (
	telemetry "EagleDeployment/Telemetry"
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define the model type to resolve the compile error.
type model struct {
	list list.Model
}

// Init is the initial command for the Bubble Tea program
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model's state
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the UI for the Bubble Tea program
func (m model) View() string {
	return m.list.View()
}

// logItem represents a single log entry in the list
type logItem struct {
	entry telemetry.Event
}

// Title returns the title of the log item (used in the list)
func (i logItem) Title() string {
	return fmt.Sprintf("[%s] %s | %s",
		i.entry.Timestamp.Format("2006-01-02 15:04:05"),
		i.entry.Type,
		i.entry.Payload["message"],
	)
}

// Description returns an empty string since we want single-line logs
func (i logItem) Description() string {
	return ""
}

// FilterValue returns the value used for filtering the log item (not used here)
func (i logItem) FilterValue() string {
	return i.entry.Payload["message"].(string)
}

func newModel(entries []telemetry.Event) model {
	// Convert telemetry events to list items
	items := make([]list.Item, len(entries))
	for i, entry := range entries {
		items[i] = logItem{entry: entry}
	}

	// Create a new list
	const defaultWidth = 80
	const defaultHeight = 25 // 25 logs per page
	l := list.New(items, list.NewDefaultDelegate(), defaultWidth, defaultHeight)
	l.Title = "Filtered Logs"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().MarginLeft(2)
	l.Styles.PaginationStyle = l.Styles.PaginationStyle.MarginLeft(2)
	l.Styles.HelpStyle = l.Styles.HelpStyle.MarginLeft(2).Padding(1, 0, 0, 0)

	return model{list: l}
}

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
	// Define the log file path
	logFilePath := "./logs/eagledeployment.log"

	// Open the log file
	file, err := os.Open(logFilePath)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer file.Close()

	// Read the log file line by line
	var entries []telemetry.Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Parse the JSON log entry
		var event telemetry.Event
		err := json.Unmarshal([]byte(scanner.Text()), &event)
		if err != nil {
			fmt.Printf("Failed to parse log entry: %v\n", err)
			continue
		}
		entries = append(entries, event)
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading log file: %v\n", err)
		return
	}

	if len(entries) == 0 {
		fmt.Println("No log entries found.")
		return
	}

	// Ask the user for sorting preference
	fmt.Println("\nSort logs by:")
	fmt.Println("1. Newest first")
	fmt.Println("2. Oldest first")
	fmt.Print("\nEnter choice: ")
	var sortChoice int
	fmt.Scanln(&sortChoice)

	newestFirst := true
	if sortChoice == 2 {
		newestFirst = false
	}

	// Sort the logs
	entries = sortLogs(entries, newestFirst)

	// Use Bubble Tea to display the logs with pagination
	p := tea.NewProgram(newModel(entries))
	if err := p.Start(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
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

	// Ask the user for sorting preference
	fmt.Println("\nSort logs by:")
	fmt.Println("1. Newest first")
	fmt.Println("2. Oldest first")
	fmt.Print("\nEnter choice: ")
	var sortChoice int
	fmt.Scanln(&sortChoice)

	newestFirst := true
	if sortChoice == 2 {
		newestFirst = false
	}

	// Sort the logs
	entries = sortLogs(entries, newestFirst)

	// Use Bubble Tea to display the logs
	p := tea.NewProgram(newModel(entries))
	if err := p.Start(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
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

// sortLogs sorts the telemetry events based on the timestamp
func sortLogs(entries []telemetry.Event, newestFirst bool) []telemetry.Event {
	if newestFirst {
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timestamp.After(entries[j].Timestamp)
		})
	} else {
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timestamp.Before(entries[j].Timestamp)
		})
	}
	return entries
}

// DisplayLogFileContents reads and displays the contents of the log file
func DisplayLogFileContents() {
	// Define the log file path
	logFilePath := "./logs/eagledeployment.log"

	// Open the log file
	file, err := os.Open(logFilePath)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer file.Close()

	// Read the log file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Parse the JSON log entry
		var event telemetry.Event
		err := json.Unmarshal([]byte(scanner.Text()), &event)
		if err != nil {
			fmt.Printf("Failed to parse log entry: %v\n", err)
			continue
		}

		// Display the log entry
		fmt.Printf("[%s] %s | %s\n",
			event.Timestamp.Format("2006-01-02 15:04:05"),
			event.Type,
			event.Payload["message"],
		)

		// Add a solid line separator
		fmt.Println(strings.Repeat("-", 80))
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading log file: %v\n", err)
	}
}
