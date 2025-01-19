// File: inventory.go
// Directory Path: /EagleDeploy_CLI/inventory

package inventory

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// Host represents a single host's configuration.
type Host struct {
	Address     string `yaml:"address"`      // IP or hostname
	SSHUser     string `yaml:"ssh_user"`     // SSH username
	SSHPassword string `yaml:"ssh_password"` // SSH password
	Port        int    `yaml:"port"`         // SSH port
	Group       string `yaml:"group"`        // Group name (e.g., dev, prod)
	Role        string `yaml:"role"`         // Host role (e.g., web, db)
}

// Inventory holds the list of all hosts and their configurations.
type Inventory struct {
	Hosts []Host `yaml:"hosts"`
}

// LoadInventory loads inventory data from a YAML file.
func LoadInventory(filePath string) (*Inventory, error) {
	log.Printf("Loading inventory from file: %s", filePath)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read inventory file: %v", err)
		return nil, err
	}

	var inventory Inventory
	err = yaml.Unmarshal(data, &inventory)
	if err != nil {
		log.Printf("Failed to unmarshal YAML data: %v", err)
		return nil, err
	}

	log.Printf("Inventory loaded successfully from file: %s", filePath)
	return &inventory, nil
}

// ValidateInventory ensures no duplicate or missing entries in the inventory.
func (inv *Inventory) ValidateInventory() error {
	addresses := make(map[string]bool)
	for _, host := range inv.Hosts {
		if host.Address == "" || host.SSHUser == "" || host.Port == 0 {
			return fmt.Errorf("host configuration incomplete: %+v", host)
		}
		if addresses[host.Address] {
			return fmt.Errorf("duplicate host entry: %s", host.Address)
		}
		addresses[host.Address] = true
	}
	log.Println("Inventory validation successful.")
	return nil
}

// MatchPlaybookHosts checks if the playbook's hosts exist in the inventory.
func (inv *Inventory) MatchPlaybookHosts(playbookHosts []string) ([]Host, error) {
	var matchedHosts []Host
	hostMap := make(map[string]Host)

	for _, host := range inv.Hosts {
		hostMap[host.Address] = host
	}

	for _, pbHost := range playbookHosts {
		host, exists := hostMap[pbHost]
		if !exists {
			return nil, fmt.Errorf("host %s not found in inventory", pbHost)
		}
		matchedHosts = append(matchedHosts, host)
	}

	log.Println("Playbook hosts matched successfully with inventory.")
	return matchedHosts, nil
}

// SaveInventory saves the current inventory to a YAML file.
func (inv *Inventory) SaveInventory(filePath string) error {
	data, err := yaml.Marshal(inv)
	if err != nil {
		log.Printf("Failed to marshal inventory to YAML: %v", err)
		return err
	}

	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		log.Printf("Failed to write inventory to file: %v", err)
		return err
	}

	log.Printf("Inventory saved successfully to file: %s", filePath)
	return nil
}
