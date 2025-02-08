// File: inventory.go
// Directory: EagleDeploy_CLI/inventory
// Purpose: Manages inventory data retrieval and provides an inventory management menu.

package inventory

import (
	"EagleDeploy_CLI/config"
	"fmt"
	"log"
)

// Inventory represents the structure of inventory.yaml
type Inventory struct {
	Hosts   []Host  `yaml:"Hosts"`
	SSHCred SSHCred `yaml:"SSH_CRED"`
	Users   []User  `yaml:"Users"`
}

type Host struct {
	IP       string `yaml:"IP"`
	Hostname string `yaml:"Hostname"`
	OS       string `yaml:"OS"`
}

type SSHCred struct {
	SSHUser string `yaml:"ssh_user"`
	SSHPass string `yaml:"ssh_pass"`
}

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Group    string `yaml:"group"`
}

// LoadInventory loads inventory.yaml and unmarshals it into the Inventory struct
func LoadInventory() (*Inventory, error) {
	var inventory Inventory
	err := config.LoadConfig("./inventory/inventory.yaml", &inventory)
	if err != nil {
		log.Printf("Failed to load inventory.yaml: %v", err)
		return nil, err
	}
	return &inventory, nil
}

// GetHosts returns the list of hosts from the inventory
func GetHosts() map[string]Host {
	inv, err := LoadInventory()
	if err != nil {
		log.Printf("Error retrieving hosts: %v", err)
		return nil
	}
	hostsMap := make(map[string]Host)
	for _, host := range inv.Hosts {
		hostsMap[host.IP] = host
		hostsMap[host.Hostname] = host
	}
	return hostsMap
}

// GetSSHCreds returns SSH credentials from inventory.yaml
func GetSSHCreds() (string, string) {
	inv, err := LoadInventory()
	if err != nil {
		log.Printf("Error retrieving SSH credentials: %v", err)
		return "", ""
	}
	return inv.SSHCred.SSHUser, inv.SSHCred.SSHPass
}

// GetUsers returns the list of users from inventory.yaml
func GetUsers() []User {
	inv, err := LoadInventory()
	if err != nil {
		log.Printf("Error retrieving users: %v", err)
		return nil
	}
	return inv.Users
}

// DisplayInventoryMenu provides an interactive menu for inventory management
func DisplayInventoryMenu() {
	for {
		fmt.Println("\nInventory Management Menu:")
		fmt.Println("1. List Hosts")
		fmt.Println("2. Show SSH Credentials")
		fmt.Println("3. List Users")
		fmt.Println("0. Return to Main Menu")
		fmt.Print("Select an option: ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			fmt.Println("\nRegistered Hosts:")
			for _, host := range GetHosts() {
				fmt.Printf("- IP: %s, Hostname: %s, OS: %s\n", host.IP, host.Hostname, host.OS)
			}
		case 2:
			user, pass := GetSSHCreds()
			fmt.Printf("\nSSH User: %s\nSSH Password: %s\n", user, pass)
		case 3:
			fmt.Println("\nRegistered Users:")
			for _, user := range GetUsers() {
				fmt.Printf("- Username: %s, Group: %s\n", user.Username, user.Group)
			}
		case 0:
			return
		default:
			fmt.Println("Invalid choice, please try again.")
		}
	}
}
