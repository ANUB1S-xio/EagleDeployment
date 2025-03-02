// File: inventory.go
// Directory: EagleDeployment/inventory
// Purpose: Manages inventory data, host discovery, and inventory operations.

package inventory

import (
	"EagleDeployment/config"
	"EagleDeployment/osdetect"
	"EagleDeployment/Telemetry"
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
	"time"

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
	t := telemetry.GetInstance()
	data, err := yaml.Marshal(inv)
	if err != nil {
		log.Printf("Failed to marshal inventory: %v", err)
		t.LogError("Inventory", "Failed to marshal inventory", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	err = os.WriteFile("./inventory/inventory.yaml", data, 0644)
	if err != nil {
		log.Printf("Failed to write inventory.yaml: %v", err)
		t.LogError("Inventory", "Failed to write inventory.yaml", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	t.LogInfo("Inventory", "Saved inventory to disk", map[string]interface{}{
		"hosts_count": len(inv.Hosts),
		"users_count": len(inv.Users),
	})
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

// GetSSHCreds returns SSH credentials from environment variables or inventory.yaml
func GetSSHCreds() (string, string) {
	// Add debug logging
	log.Printf("Reading SSH credentials from environment variables")

	sshUser := os.Getenv("EAGLE_SSH_USER")
	sshPass := os.Getenv("EAGLE_SSH_PASS")

	if sshUser != "" && sshPass != "" {
		log.Printf("Found SSH credentials in environment variables for user: %s", sshUser)
		return sshUser, sshPass
	}

	// Log fallback to inventory.yaml
	log.Printf("No environment variables found, falling back to inventory.yaml")

	inv, err := LoadInventory()
	if err != nil {
		log.Printf("Error retrieving SSH credentials from inventory: %v", err)
		return "", ""
	}

	if inv.SSHCred.SSHUser != "" && inv.SSHCred.SSHPass != "" {
		log.Printf("Found SSH credentials in inventory.yaml for user: %s", inv.SSHCred.SSHUser)
		return inv.SSHCred.SSHUser, inv.SSHCred.SSHPass
	}

	log.Printf("No SSH credentials found in either environment or inventory")
	return "", ""
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
	fmt.Printf("Parsing IP range: %s\n", ipRange) // Debug statement
	var ips []string
	parts := strings.Split(ipRange, "-")
	fmt.Printf("Parts: %v\n", parts) // Debug statement
	if len(parts) == 1 {
		// Single IP address
		ip := net.ParseIP(parts[0])
		if ip == nil {
			return nil, fmt.Errorf("invalid IP address format")
		}
		return []string{ip.String()}, nil
	} else if len(parts) == 2 {
		// IP range
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
	} else {
		return nil, fmt.Errorf("invalid IP range format")
	}
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

// AddHost prompts for IP input, detects hostname and OS, and appends to inventory.yaml if the host is alive.
func AddHost(ipRange string) {
	t := telemetry.GetInstance()
	t.LogInfo("Inventory", "Adding hosts from IP range", map[string]interface{}{
		"ip_range": ipRange,
	})

	ips, err := parseIPRange(ipRange)
	if err != nil {
		fmt.Printf("Error parsing IP range: %v\n", err)
		t.LogError("Inventory", "Failed to parse IP range", map[string]interface{}{
			"ip_range": ipRange,
			"error":    err.Error(),
		})
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

				// Add debug logging for SSH credentials
				log.Printf("Attempting OS detection for %s with credentials - User: %s", ip, sshUser)

				// Add retry logic for OS detection
				var osType string
				for attempts := 1; attempts <= 3; attempts++ {
					osType, err = osdetect.DetectOS(ip, sshUser, sshPass, 22)
					if err == nil {
						break
					}
					log.Printf("Attempt %d: Error detecting OS for %s: %v", attempts, ip, err)
					// Wait briefly before retry
					time.Sleep(2 * time.Second)
				}

				if err != nil {
					log.Printf("All attempts failed to detect OS for %s: %v", ip, err)
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

			t.LogInfo("Inventory", "Added host to inventory", map[string]interface{}{
				"ip":       host.IP,
				"hostname": host.Hostname,
				"os":       host.OS,
			})
		} else {
			fmt.Printf("Host %s already exists in the inventory. Skipping.\n", host.IP)
			t.LogDebug("Inventory", "Skipped existing host", map[string]interface{}{
				"ip": host.IP,
			})
		}
	}

	SaveInventory(inv)
}

// scanAndAddIP prompts for IP input, detects hostname and OS, and appends to inventory.yaml if the host is alive.
func scanAndAddIP() {
	fmt.Print("Enter IP Address or Range: ")
	var ipRange string
	fmt.Scanln(&ipRange)

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
	t := telemetry.GetInstance()
	inv, err := LoadInventory()
	if err != nil {
		fmt.Println("Error loading inventory:", err)
		t.LogError("Inventory", "Error loading inventory for update", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	if index < 0 || index >= len(inv.Hosts) {
		fmt.Println("Invalid host index")
		t.LogWarning("Inventory", "Invalid host index for update", map[string]interface{}{
			"index":       index,
			"hosts_count": len(inv.Hosts),
		})
		return
	}

	oldHost := inv.Hosts[index]
	inv.Hosts[index] = newHost

	t.LogInfo("Inventory", "Updated host in inventory", map[string]interface{}{
		"index":        index,
		"old_ip":       oldHost.IP,
		"old_hostname": oldHost.Hostname,
		"old_os":       oldHost.OS,
		"new_ip":       newHost.IP,
		"new_hostname": newHost.Hostname,
		"new_os":       newHost.OS,
	})

	SaveInventory(inv)
	fmt.Println("Host updated successfully")
}

// DeleteHost removes a host from the inventory
func DeleteHost(index int) {
	t := telemetry.GetInstance()
	inv, err := LoadInventory()
	if err != nil {
		fmt.Println("Error loading inventory:", err)
		t.LogError("Inventory", "Error loading inventory for deletion", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	if index < 0 || index >= len(inv.Hosts) {
		fmt.Println("Invalid host index")
		t.LogWarning("Inventory", "Invalid host index for deletion", map[string]interface{}{
			"index":       index,
			"hosts_count": len(inv.Hosts),
		})
		return
	}

	deletedHost := inv.Hosts[index]
	t.LogInfo("Inventory", "Deleted host from inventory", map[string]interface{}{
		"index":    index,
		"ip":       deletedHost.IP,
		"hostname": deletedHost.Hostname,
		"os":       deletedHost.OS,
	})

	inv.Hosts = append(inv.Hosts[:index], inv.Hosts[index+1:]...)
	SaveInventory(inv)
	fmt.Println("Host deleted successfully")
}

// EditSSHCreds allows updating the SSH credentials in the inventory.
func EditSSHCreds() {
	t := telemetry.GetInstance()
	inv, err := LoadInventory()
	if err != nil {
		fmt.Println("Error loading inventory:", err)
		t.LogError("Inventory", "Error loading inventory for SSH credential update", map[string]interface{}{
			"error": err.Error(),
		})
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

	t.LogInfo("Inventory", "Updated SSH credentials", map[string]interface{}{
		"username": newUser,
	})

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
			scanAndAddIP()
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
	t := telemetry.GetInstance()
	t.LogInfo("Playbook", "Injecting inventory into playbook", map[string]interface{}{
		"template_path": templatePath,
		"output_path":   outputPath,
	})

	inv, err := LoadInventory()
	if err != nil {
		t.LogError("Playbook", "Failed to load inventory for playbook", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to load inventory: %v", err)
	}

	sshUser, sshPass := GetSSHCreds()

	// Initialize user credentials
	var userName, userPass string

	// Check inventory for users first
	if len(inv.Users) > 0 {
		// Use the first user from inventory
		if inv.Users[0].Username != "" && inv.Users[0].Password != "" {
			userName = inv.Users[0].Username
			userPass = inv.Users[0].Password
			log.Printf("Using existing user from inventory: %s", userName)
		} else {
			log.Printf("Invalid user in inventory, prompting for new credentials")
		}
	}

	// Prompt for credentials if we don't have valid ones
	if userName == "" || userPass == "" {
		// Prompt for user credentials
		fmt.Print("Enter username for new user: ")
		fmt.Scanln(&userName)
		fmt.Print("Enter password for new user: ")
		fmt.Scanln(&userPass)

		if userName != "" && userPass != "" {
			// Create new user in inventory
			newUser := User{
				Username: userName,
				Password: userPass,
				Group:    "users",
			}

			// Clear existing users if any
			inv.Users = []User{newUser}
			SaveInventory(inv)
			log.Printf("Created new user: %s", userName)
		} else {
			return fmt.Errorf("username and password cannot be empty")
		}
	}

	// Final verification of credentials
	if userName == "" || userPass == "" {
		return fmt.Errorf("failed to get user credentials - please enter valid credentials")
	}

	// Prepare template data
	data := struct {
		Hosts   []Host
		SSHCred SSHCred
		Vars    struct {
			UserName     string
			UserPassword string
		}
	}{
		Hosts: inv.Hosts,
		SSHCred: SSHCred{
			SSHUser: sshUser,
			SSHPass: sshPass,
		},
		Vars: struct {
			UserName     string
			UserPassword string
		}{
			UserName:     userName,
			UserPassword: userPass,
		},
	}

	// Create template with custom functions
	funcMap := template.FuncMap{
		"env":      os.Getenv,
		"lower":    strings.ToLower,
		"contains": strings.Contains,
	}

	tmplBytes, err := ioutil.ReadFile(templatePath)
	if err != nil {
		t.LogError("Playbook", "Failed to read playbook template", map[string]interface{}{
			"error":         err.Error(),
			"template_path": templatePath,
		})
		return fmt.Errorf("failed to read playbook template: %v", err)
	}

	tmpl, err := template.New("playbook").Funcs(funcMap).Parse(string(tmplBytes))
	if err != nil {
		t.LogError("Playbook", "Failed to parse playbook template", map[string]interface{}{
			"error":         err.Error(),
			"template_path": templatePath,
		})
		return fmt.Errorf("failed to parse playbook template: %v", err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
		t.LogError("Playbook", "Failed to execute template", map[string]interface{}{
			"error":         err.Error(),
			"template_path": templatePath,
		})
		return fmt.Errorf("failed to execute template: %v", err)
	}

	if err := os.WriteFile(outputPath, rendered.Bytes(), 0644); err != nil {
		t.LogError("Playbook", "Failed to write rendered playbook", map[string]interface{}{
			"error":       err.Error(),
			"output_path": outputPath,
		})
		return fmt.Errorf("failed to write rendered playbook: %v", err)
	}

	t.LogInfo("Playbook", "Successfully injected inventory into playbook", map[string]interface{}{
		"template_path": templatePath,
		"output_path":   outputPath,
		"hosts_count":   len(inv.Hosts),
	})
	return nil
}
