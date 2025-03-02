// File: menu/inventory_menu.go

package menu

import (
	"fmt"

	"EagleDeployment/inventory"

	tea "github.com/charmbracelet/bubbletea"
)

// InventoryMenuModel represents the inventory management menu
type InventoryMenuModel struct {
	choices  []string
	cursor   int
	selected int
}

// InitInventoryMenu initializes the inventory menu
func InitInventoryMenu() InventoryMenuModel {
	return InventoryMenuModel{
		choices: []string{
			"Add Hosts",
			"Manage Current Inventory",
			"Show SSH Credentials",
			"List Users",
			"Return to Main Menu",
		},
		cursor:   0,
		selected: -1,
	}
}

// Init is called when the program starts
func (m InventoryMenuModel) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (m InventoryMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.selected = 4 // Return to Main Menu
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
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
func (m InventoryMenuModel) View() string {
	s := "\nInventory Management Menu:\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %d. %s\n", cursor, i+1, choice)
	}

	s += "\nSelect an option using up/down arrows, then press Enter.\n"
	s += "Press Esc to return to the main menu.\n"

	return s
}

// ManageInventoryMenu handles the inventory submenu
func ManageInventoryMenu() InventoryMenuModel {
	return InventoryMenuModel{
		choices: []string{
			"List Hosts",
			"Update Host",
			"Delete Host",
			"Edit SSH Credentials",
			"Return to Inventory Menu",
		},
		cursor:   0,
		selected: -1,
	}
}

// RunInventoryMenu displays inventory menu and handles interactions
func RunInventoryMenu() {
	for {
		p := tea.NewProgram(InitInventoryMenu())
		m, err := p.Run()
		if err != nil {
			fmt.Println("Error running inventory menu:", err)
			return
		}

		model := m.(InventoryMenuModel)

		switch model.selected {
		case 0: // Add Hosts
			var ipRange string
			fmt.Print("Enter IP Address or Range: ")
			fmt.Scanln(&ipRange)
			inventory.AddHost(ipRange)
		case 1: // Manage Current Inventory
			manageCurrentInventory()
		case 2: // Show SSH Credentials
			user, pass := inventory.GetSSHCreds()
			fmt.Printf("\nSSH User: %s\nSSH Password: %s\n", user, pass)
			fmt.Println("\nPress Enter to continue...")
			fmt.Scanln()
		case 3: // List Users
			fmt.Println("\nRegistered Users:")
			for _, user := range inventory.GetUsers() {
				fmt.Printf("- Username: %s, Group: %s\n", user.Username, user.Group)
			}
			fmt.Println("\nPress Enter to continue...")
			fmt.Scanln()
		case 4, -1: // Return to Main Menu or ESC/Ctrl+C
			return
		}
	}
}

// Manage current inventory submenu
func manageCurrentInventory() {
	p := tea.NewProgram(ManageInventoryMenu())
	m, err := p.Run()
	if err != nil {
		fmt.Println("Error running manage inventory menu:", err)
		return
	}

	model := m.(InventoryMenuModel)

	switch model.selected {
	case 0: // List Hosts
		inventory.ListHosts()
		fmt.Println("\nPress Enter to continue...")
		fmt.Scanln()
	case 1: // Update Host
		inventory.ListHosts()
		fmt.Print("Enter the index of the host to update: ")
		var index int
		fmt.Scanln(&index)
		index-- // Convert to zero-based index

		fmt.Print("Enter new IP: ")
		var ip string
		fmt.Scanln(&ip)

		fmt.Print("Enter new Hostname: ")
		var hostname string
		fmt.Scanln(&hostname)

		fmt.Print("Enter new OS: ")
		var os string
		fmt.Scanln(&os)

		newHost := inventory.Host{IP: ip, Hostname: hostname, OS: os}
		inventory.UpdateHost(index, newHost)
	case 2: // Delete Host
		inventory.ListHosts()
		fmt.Print("Enter the index of the host to delete: ")
		var index int
		fmt.Scanln(&index)
		index-- // Convert to zero-based index
		inventory.DeleteHost(index)
	case 3: // Edit SSH Credentials
		inventory.EditSSHCreds()
	case 4, -1: // Return to Inventory Menu
		return
	}
}
