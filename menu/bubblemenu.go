// File: menu/bubblemenu.go

package menu

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define some styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00ff99")).
			Padding(1, 0, 0, 0)

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
)

// Menu model
type MainMenuModel struct {
	choices  []string
	cursor   int
	selected int
}

// Initial menu model
func InitialMainMenu() MainMenuModel {
	return MainMenuModel{
		choices: []string{
			"Execute a Playbook",
			"List YAML Playbooks",
			"Manage Inventory",
			"Enable/Disable Detailed Logging",
			"Rollback Changes",
			"Show Help",
			"Exit",
		},
		cursor:   0,
		selected: -1,
	}
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
		}
	}

	return m, nil
}

// View renders the UI
func (m MainMenuModel) View() string {
	s := titleStyle.Render("EagleDeploy Menu")
	s += "\n\n"

	for i, choice := range m.choices {
		if m.cursor == i {
			s += selectedItemStyle.Render(fmt.Sprintf("> %d. %s", i+1, choice)) + "\n"
		} else {
			s += itemStyle.Render(fmt.Sprintf("  %d. %s", i+1, choice)) + "\n"
		}
	}

	s += "\n" + infoStyle.Render("Select an option using up/down arrows, then press Enter.")
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
