// File: inventory.go
// Directory: EagleDeploy_CLI/inventory
// Purpose: Manages inventory data retrieval, adding hosts, managing inventory, and provides an inventory management menu.

package inventory

import (
	"EagleDeploy_CLI/config"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"

	"gopkg.in/yaml.v2"
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

// SaveInventory writes the updated inventory back to inventory.yaml
func SaveInventory(inv *Inventory) {
	data, err := yaml.Marshal(inv)
	if err != nil {
		log.Printf("Failed to marshal inventory: %v", err)
		return
	}
	err = ioutil.WriteFile("./inventory/inventory.yaml", data, 0644)
	if err != nil {
		log.Printf("Failed to write inventory.yaml: %v", err)
	}
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

// detectHostname attempts to resolve the hostname via DNS
func detectHostname(ip string) string {
	hostnames, err := net.LookupAddr(ip)
	if err == nil && len(hostnames) > 0 {
		return strings.TrimSuffix(hostnames[0], ".")
	}
	return ""
}

// AddHost prompts for IP input, detects hostname, and appends to inventory.yaml
func AddHost(ip string) {
	hostname := detectHostname(ip)
	newHost := Host{IP: ip, Hostname: hostname, OS: ""}

	inv, err := LoadInventory()
	if err != nil {
		inv = &Inventory{}
	}
	inv.Hosts = append(inv.Hosts, newHost)
	SaveInventory(inv)
	fmt.Printf("Added host: %s (Hostname: %s)\n", ip, hostname)
}

// ListHosts prints all current hosts in the inventory
func ListHosts() {
	inv, err := LoadInventory()
	if err != nil {
		fmt.Println("Error loading inventory:", err)
		return
	}
	fmt.Println("\nCurrent Hosts:")
	for i, host := range inv.Hosts {
		fmt.Printf("%d. IP: %s, Hostname: %s, OS: %s\n", i+1, host.IP, host.Hostname, host.OS)
	}
}

// UpdateHost updates the details of a host in the inventory
func UpdateHost(index int, newHost Host) {
	inv, err := LoadInventory()
	if err != nil {
		fmt.Println("Error loading inventory:", err)
		return
	}
	if index < 0 || index >= len(inv.Hosts) {
		fmt.Println("Invalid host index")
		return
	}
	inv.Hosts[index] = newHost
	SaveInventory(inv)
	fmt.Println("Host updated successfully")
}

// DeleteHost removes a host from the inventory
func DeleteHost(index int) {
	inv, err := LoadInventory()
	if err != nil {
		fmt.Println("Error loading inventory:", err)
		return
	}
	if index < 0 || index >= len(inv.Hosts) {
		fmt.Println("Invalid host index")
		return
	}
	inv.Hosts = append(inv.Hosts[:index], inv.Hosts[index+1:]...)
	SaveInventory(inv)
	fmt.Println("Host deleted successfully")
}

// ManageInventory allows modifying hosts, users, and SSH credentials
func ManageInventory() {
	for {
		fmt.Println("\nManage Current Inventory:")
		fmt.Println("1. List Hosts")
		fmt.Println("2. Update Host")
		fmt.Println("3. Delete Host")
		fmt.Println("4. Edit Users")
		fmt.Println("5. Edit SSH Credentials")
		fmt.Println("0. Return to Inventory Menu")
		fmt.Print("Select an option: ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			ListHosts()
		case 2:
			ListHosts()
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
			newHost := Host{IP: ip, Hostname: hostname, OS: os}
			UpdateHost(index, newHost)
		case 3:
			ListHosts()
			fmt.Print("Enter the index of the host to delete: ")
			var index int
			fmt.Scanln(&index)
			index-- // Convert to zero-based index
			DeleteHost(index)
		case 4:
			fmt.Println("Editing users...") // Placeholder for future functionality
		case 5:
			fmt.Println("Editing SSH credentials...") // Placeholder for future functionality
		case 0:
			return
		default:
			fmt.Println("Invalid choice, please try again.")
		}
	}
}

// DisplayInventoryMenu provides an interactive menu for inventory management
func DisplayInventoryMenu() {
	for {
		fmt.Println("\nInventory Management Menu:")
		fmt.Println("1. Add Hosts")
		fmt.Println("2. Manage Current Inventory")
		fmt.Println("3. Show SSH Credentials")
		fmt.Println("4. List Users")
		fmt.Println("0. Return to Main Menu")
		fmt.Print("Select an option: ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			fmt.Print("Enter IP Address or Range: ")
			var ip string
			fmt.Scanln(&ip)
			AddHost(ip)
		case 2:
			ManageInventory()
		case 3:
			user, pass := GetSSHCreds()
			fmt.Printf("\nSSH User: %s\nSSH Password: %s\n", user, pass)
		case 4:
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
