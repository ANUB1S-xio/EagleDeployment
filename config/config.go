// File: config.go
// Directory Path: /EagleDeployment/config
// Purpose: Provides configuration file loading and parsing functionality

package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// Function: LoadConfig
// Purpose: Loads and parses YAML configuration files into Go structures
// Parameters:
//   - filePath: string - Path to YAML configuration file
//   - target: interface{} - Target structure for unmarshaling
//
// Returns:
//   - error - Any loading or parsing errors
//
// Called By:
//   - [`inventory.LoadInventory`](../inventory/inventory.go)
//   - [`tasks.LoadPlaybook`](../tasks/tasks.go)
//
// Dependencies:
//   - gopkg.in/yaml.v2 for YAML parsing
//   - io/ioutil for file operations
//
// Notes:
//   - Handles both inventory and playbook configurations
//   - Logs all operations for debugging
//   - Returns detailed error messages
func LoadConfig(filePath string, target interface{}) error {
	log.Printf("Loading configuration from file: %s", filePath)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read configuration file: %v", err)
		return err
	}

	err = yaml.Unmarshal(data, target)
	if err != nil {
		log.Printf("Failed to unmarshal YAML data: %v", err)
		return err
	}

	log.Printf("Configuration loaded successfully from file: %s", filePath)
	return nil
}
