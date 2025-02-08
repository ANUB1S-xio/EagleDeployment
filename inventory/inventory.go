// File: inventory.go
// Directory: EagleDeploy_CLI/inventory
// Purpose: Manages inventory data retrieval, adding hosts, managing inventory, and provides an inventory management menu.

package inventory

import (
	"EagleDeploy_CLI/config"
	"EagleDeploy_CLI/osdetect"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"text/template"

	"golang.org/x/sync/errgroup"
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
	err = os.WriteFile("./inventory/inventory.yaml", data, 0644)
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

// checkHostAlive uses the system's ping command to check if the host is alive
func checkHostAlive(ip string) bool {
	cmd := exec.Command("ping", "-n", "3", "-w", "5000", ip)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to ping host %s: %v\n", ip, err)
		return false
	}
	return strings.Contains(string(output), "TTL=")
}

// parseIPRange parses an IP range string and returns a slice of IP addresses
func parseIPRange(ipRange string) ([]string, error) {
	var ips []string
	parts := strings.Split(ipRange, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid IP range format")
	}

	startIP := net.ParseIP(parts[0])
	if startIP == nil {
		return nil, fmt.Errorf("invalid start IP address")
	}

	var endIP net.IP
	if strings.Contains(parts[1], ".") {
		endIP = net.ParseIP(parts[1])
		if endIP == nil {
			return nil, fmt.Errorf("invalid end IP address")
		}
	} else {
		// Handle the shorter format (e.g., 10.42.56.1-254)
		startIPParts := strings.Split(parts[0], ".")
		if len(startIPParts) != 4 {
			return nil, fmt.Errorf("invalid start IP address format")
		}
		endIP = net.ParseIP(fmt.Sprintf("%s.%s.%s.%s", startIPParts[0], startIPParts[1], startIPParts[2], parts[1]))
		if endIP == nil {
			return nil, fmt.Errorf("invalid end IP address")
		}
	}

	for ip := startIP; !ip.Equal(endIP); ip = nextIP(ip) {
		ips = append(ips, ip.String())
	}
	ips = append(ips, endIP.String()) // Include the end IP

	return ips, nil
}

// nextIP returns the next IP address in sequence
func nextIP(ip net.IP) net.IP {
	ip = ip.To4()
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
	return ip
}

// AddHost prompts for IP input, detects hostname and OS,
// and appends to inventory.yaml if the host is alive.
func AddHost(ipRange string) {
	ips, err := parseIPRange(ipRange)
	if err != nil {
		fmt.Printf("Error parsing IP range: %v\n", err)
		return
	}

	inv, err := LoadInventory()
	if err != nil {
		inv = &Inventory{}
	}

	existingHosts := make(map[string]bool)
	for _, host := range inv.Hosts {
		existingHosts[host.IP] = true
	}

	var mu sync.Mutex
	var g errgroup.Group
	var aliveHosts []Host

	sshUser, sshPass := GetSSHCreds()

	for _, ip := range ips {
		ip := ip // capture range variable
		g.Go(func() error {
			if checkHostAlive(ip) {
				hostname := detectHostname(ip)
				// Call osdetect.DetectOS to get the OS type
				osType, err := osdetect.DetectOS(ip, sshUser, sshPass, 22)
				if err != nil {
					log.Printf("Error detecting OS for %s: %v", ip, err)
					osType = "Unknown"
				}
				newHost := Host{IP: ip, Hostname: hostname, OS: osType}
				mu.Lock()
				aliveHosts = append(aliveHosts, newHost)
				mu.Unlock()
				fmt.Printf("Host %s is alive. Detected OS: %s\n", ip, osType)
			} else {
				fmt.Printf("Host %s is not alive. Not adding to inventory.\n", ip)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Printf("Error checking hosts: %v\n", err)
	}

	// Add alive hosts to inventory if they are not duplicates
	for _, host := range aliveHosts {
		if !existingHosts[host.IP] {
			inv.Hosts = append(inv.Hosts, host)
			existingHosts[host.IP] = true
			fmt.Printf("Added host: %s (Hostname: %s)\n", host.IP, host.Hostname)
		} else {
			fmt.Printf("Host %s already exists in the inventory. Skipping.\n", host.IP)
		}
	}

	SaveInventory(inv)
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

// EditSSHCreds allows updating the SSH credentials in the inventory.
func EditSSHCreds() {
	inv, err := LoadInventory()
	if err != nil {
		fmt.Println("Error loading inventory:", err)
		return
	}
	var newUser, newPass string
	fmt.Print("Enter new SSH username: ")
	fmt.Scanln(&newUser)
	fmt.Print("Enter new SSH password: ")
	fmt.Scanln(&newPass)

	inv.SSHCred = SSHCred{
		SSHUser: newUser,
		SSHPass: newPass,
	}
	SaveInventory(inv)
	fmt.Println("SSH credentials updated successfully.")
}

// ManageInventory allows modifying hosts, users, and SSH credentials
func ManageInventory() {
	for {
		fmt.Println("\nManage Current Inventory:")
		fmt.Println("1. List Hosts")
		fmt.Println("2. Update Host")
		fmt.Println("3. Delete Host")
		fmt.Println("4. Edit SSH Credentials")
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
			EditSSHCreds()
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

// InjectInventoryIntoPlaybook loads inventory.yaml, injects the inventory data (hosts including OS) and SSH credentials,
// and writes the rendered output to outputPath.
func InjectInventoryIntoPlaybook(templatePath, outputPath string) error {
	inv, err := LoadInventory()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %v", err)
	}

	// Prepare data structure containing all hosts (with OS), SSH credentials, and user details
	data := struct {
		Hosts        []Host
		SSHCred      SSHCred
		UserName     string
		UserPassword string
	}{
		Hosts:        inv.Hosts,
		SSHCred:      inv.SSHCred,
		UserName:     "steve",            // Set the user name here
		UserPassword: "ComplexP@ssw0rd!", // Set the user password here
	}

	tmplBytes, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read playbook template: %v", err)
	}

	tmpl, err := template.New("playbook").Parse(string(tmplBytes))
	if err != nil {
		return fmt.Errorf("failed to parse playbook template: %v", err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	if err := os.WriteFile(outputPath, rendered.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write rendered playbook: %v", err)
	}

	return nil
}
