package menu

import (
	telemetry "EagleDeployment/Telemetry"
	"fmt"
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
	return fmt.Sprintf("[%s] %s", i.entry.Type, i.entry.Payload["message"])
}

// Description returns the description of the log item (used in the list)
func (i logItem) Description() string {
	return fmt.Sprintf("Category: %s | Timestamp: %s",
		i.entry.Payload["category"],
		i.entry.Timestamp.Format("2006-01-02 15:04:05"),
	)
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
	const defaultWidth = 50
	l := list.New(items, list.NewDefaultDelegate(), defaultWidth, 15)
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
