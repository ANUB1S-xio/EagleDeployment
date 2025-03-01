// File: menu/playbook_menu.go

package menu

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// PlaybookModel represents the playbook selection menu
type PlaybookModel struct {
	playbooks []string
	cursor    int
	selected  int
}

// InitPlaybookMenu initializes the playbook selection menu
func InitPlaybookMenu(playbooks []string) PlaybookModel {
	return PlaybookModel{
		playbooks: playbooks,
		cursor:    0,
		selected:  -1,
	}
}

// Init is called when the program starts
func (m PlaybookModel) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (m PlaybookModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.selected = -1
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.playbooks)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.cursor
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the UI
func (m PlaybookModel) View() string {
	if len(m.playbooks) == 0 {
		return "\nNo playbooks found in the 'playbooks' directory.\n\nPress any key to return to menu.\n"
	}

	s := "\nAvailable Playbooks:\n\n"

	for i, playbook := range m.playbooks {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %d. %s\n", cursor, i+1, playbook)
	}

	s += "\nSelect a playbook using up/down arrows, then press Enter.\n"
	s += "Press Esc to return to the main menu.\n"

	return s
}

// Function to get available playbooks
func ListPlaybooks() []string {
	playbooksDir := "./playbooks" // Default directory for playbooks

	// Ensure the playbooks directory exists
	if _, err := os.Stat(playbooksDir); os.IsNotExist(err) {
		return nil
	}

	// Read the playbooks directory
	files, err := os.ReadDir(playbooksDir)
	if err != nil {
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

// RunPlaybookMenu displays the playbook selection menu and returns the selected playbook
func RunPlaybookMenu() (string, bool) {
	playbooks := ListPlaybooks()

	p := tea.NewProgram(InitPlaybookMenu(playbooks))
	m, err := p.Run()
	if err != nil {
		fmt.Println("Error running playbook menu:", err)
		return "", false
	}

	model := m.(PlaybookModel)

	if model.selected == -1 || model.selected >= len(playbooks) {
		return "", false
	}

	return playbooks[model.selected], true
}
