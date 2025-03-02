// File: menu/bubblemenu.go

package menu

import (
	"fmt"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define some styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00ff99")).
			Padding(1, 0, 0, 0)

	programNameStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00ffff")).
				Background(lipgloss.Color("#333333")).
				PaddingLeft(4).
				PaddingRight(4).
				MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			Padding(0, 0, 0, 2)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00ff99")).
				Bold(true).
				Padding(0, 0, 0, 0)

	infoStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#666666")).
			Padding(1, 0, 0, 2)

	quitStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true).
			Padding(0, 0, 0, 2)

	// Status indicator styles
	activeServerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true)

	inactiveServerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Bold(true)

	serverStatusLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)
)

// ASCII art for the program name
const programNameArt = `
 _____            _      _____             _                                  _   
| ____|__ _  __ _| | ___| ____|_  ___ __ | | ___  _   _ _ __ ___   ___ _ __ | |_ 
|  _| / _` + "`" + ` |/ _` + "`" + ` | |/ _ \  _| \ \/ / '_ \| |/ _ \| | | | '_ ` + "`" + ` _ \ / _ \ '_ \| __|
| |__| (_| | (_| | |  __/ |___ >  <| |_) | | (_) | |_| | | | | | |  __/ | | | |_ 
|_____\__,_|\__, |_|\___|_____/_/\_\ .__/|_|\___/ \__, |_| |_| |_|\___|_| |_|\__|
            |___/                  |_|            |___/                          
`

// Menu model
type MainMenuModel struct {
	choices     []string
	cursor      int
	selected    int
	serverAlive bool
}

// Initial menu model
func InitialMainMenu() MainMenuModel {
	// Check if server is alive
	serverAlive := checkServerStatus()

	return MainMenuModel{
		choices: []string{
			"Execute a Playbook",
			"Manage Inventory",
			"View Logs", // Changed from "Enable/Disable Detailed Logging"
		},
		cursor:      0,
		selected:    -1,
		serverAlive: serverAlive,
	}
}

// Check if the server is responding
func checkServerStatus() bool {
	client := http.Client{
		Timeout: 500 * time.Millisecond, // Short timeout to not block the UI
	}

	// Make a request to the server
	resp, err := client.Get("http://localhost:8742/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// Init is called when the program starts
func (m MainMenuModel) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (m MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = m.cursor
			return m, tea.Quit
		case "r": // Refresh server status
			m.serverAlive = checkServerStatus()
		}
	}

	return m, nil
}

// View renders the UI
func (m MainMenuModel) View() string {
	// Start with the ASCII art program name
	s := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00ffff")).
		Render(programNameArt)

	// Add server status indicator below the program name
	statusLabel := serverStatusLabelStyle.Render("Server Status: ")
	if m.serverAlive {
		statusIndicator := activeServerStyle.Render("● ONLINE")
		s += "\n" + statusLabel + statusIndicator
	} else {
		statusIndicator := inactiveServerStyle.Render("● OFFLINE")
		s += "\n" + statusLabel + statusIndicator
	}

	s += "\n\n" + titleStyle.Render("Main Menu")
	s += "\n\n"

	for i, choice := range m.choices {
		if m.cursor == i {
			s += selectedItemStyle.Render(fmt.Sprintf("> %d. %s", i+1, choice)) + "\n"
		} else {
			s += itemStyle.Render(fmt.Sprintf("  %d. %s", i+1, choice)) + "\n"
		}
	}

	s += "\n" + infoStyle.Render("Select an option using up/down arrows, then press Enter.")
	s += "\n" + infoStyle.Render("Press 'r' to refresh server status.")
	s += "\n" + quitStyle.Render("Press q or Ctrl+C to quit.")
	return s
}

// Function to run the main menu and return the selected option
func RunMainMenu() int {
	p := tea.NewProgram(InitialMainMenu())
	m, err := p.Run()
	if err != nil {
		fmt.Println("Error running menu:", err)
		return -1
	}

	model := m.(MainMenuModel)
	return model.selected
}
