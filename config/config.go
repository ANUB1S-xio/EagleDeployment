// File: config.go
// Directory Path: /EagleDeploy_CLI/config

package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// Function: LoadConfig
// Purpose: Loads a YAML configuration file and unmarshals it into a given Go structure.
// Parameters:
// - filePath: The file path to the YAML configuration file.
// - target: The Go structure to populate with the configuration data.
// Returns: An error if the file cannot be read or unmarshalled.
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
