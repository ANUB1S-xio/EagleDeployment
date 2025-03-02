package menu

import (
	"fmt"
	"os"
	"strings"

	"EagleDeploy_CLI/Telemetry" // Fixed import path

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define listHeight constant
const (
	listHeight = 14
)

type menuItem struct {
	title, desc string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.desc }
func (i menuItem) FilterValue() string { return i.title }

type model struct {
	list      list.Model
	telemetry *Telemetry.Telemetry
}

func newModel() model {
	items := []list.Item{
		menuItem{title: "Set Log Level", desc: "Set the log level (ERROR, WARNING, INFO, DEBUG)"},
		menuItem{title: "Enable/Disable Console Logging", desc: "Enable or disable console logging"},
		menuItem{title: "Log an Error", desc: "Log an error message"},
		menuItem{title: "Log a Warning", desc: "Log a warning message"},
		menuItem{title: "Log an Info", desc: "Log an info message"},
		menuItem{title: "Log a Debug", desc: "Log a debug message"},
		menuItem{title: "Filter Logs", desc: "Filter logs based on criteria"},
		menuItem{title: "Clear Logs", desc: "Clear all logs"},
		menuItem{title: "Exit", desc: "Exit the application"},
	}

	const defaultWidth = 20

	l := list.New(items, list.NewDefaultDelegate(), defaultWidth, listHeight)
	l.Title = "Telemetry Menu"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().MarginLeft(2)
	l.Styles.PaginationStyle = l.Styles.PaginationStyle.MarginLeft(2)
	l.Styles.HelpStyle = l.Styles.HelpStyle.MarginLeft(2).Padding(1, 0, 0, 0)

	return model{list: l, telemetry: Telemetry.GetInstance()}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(menuItem)
			if ok {
				handleMenuSelection(i.title, m.telemetry)
			}
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return "\n" + m.list.View()
}

func handleMenuSelection(selection string, telemetry *Telemetry.Telemetry) {
	switch selection {
	case "Set Log Level":
		level := getInput("Enter log level (ERROR, WARNING, INFO, DEBUG): ")
		switch strings.ToUpper(level) {
		case "ERROR":
			telemetry.SetLevel(Telemetry.LevelError)
		case "WARNING":
			telemetry.SetLevel(Telemetry.LevelWarning)
		case "INFO":
			telemetry.SetLevel(Telemetry.LevelInfo)
		case "DEBUG":
			telemetry.SetLevel(Telemetry.LevelDebug)
		default:
			fmt.Println("Invalid log level")
		}
	case "Enable/Disable Console Logging":
		enable := getInput("Enable console logging? (yes/no): ")
		telemetry.EnableConsole(strings.ToLower(enable) == "yes")
	case "Log an Error":
		category := getInput("Enter category: ")
		message := getInput("Enter message: ")
		telemetry.LogError(category, message, nil)
	case "Log a Warning":
		category := getInput("Enter category: ")
		message := getInput("Enter message: ")
		telemetry.LogWarning(category, message, nil)
	case "Log an Info":
		category := getInput("Enter category: ")
		message := getInput("Enter message: ")
		telemetry.LogInfo(category, message, nil)
	case "Log a Debug":
		category := getInput("Enter category: ")
		message := getInput("Enter message: ")
		telemetry.LogDebug(category, message, nil)
	case "Filter Logs":
		level := getInput("Enter log level to filter (leave blank for all): ")
		category := getInput("Enter category to filter (leave blank for all): ")
		message := getInput("Enter message to filter (leave blank for all): ")
		getInput("Enter limit of entries to display: ")
		entries := telemetry.FilterLogs(level, category, message, 10)
		for _, entry := range entries {
			fmt.Printf("%+v\n", entry)
		}
	case "Clear Logs":
		if err := telemetry.ClearLogs(); err != nil {
			fmt.Printf("Failed to clear logs: %v\n", err)
		} else {
			fmt.Println("Logs cleared successfully")
		}
	case "Exit":
		fmt.Println("Exiting...")
		os.Exit(0)
	default:
		fmt.Println("Invalid selection")
	}
}

func getInput(prompt string) string {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}

func RunTelemetryMenu() {
	p := tea.NewProgram(newModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
