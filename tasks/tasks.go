// File: tasks.go
// Directory: EagleDeployment/tasks
// Purpose: Adds YAML file detection and proper struct handling.

package tasks

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// Task struct defining a single playbook task
type Task struct {
	Name        string `yaml:"name"`
	Command     string `yaml:"command"`
	SSHUser     string `yaml:"ssh_user"`
	SSHPassword string `yaml:"ssh_password"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
}

// Playbook struct defining the overall playbook structure
type Playbook struct {
	Name     string         `yaml:"name"`
	Version  string         `yaml:"version"`
	Hosts    []string       `yaml:"hosts"`
	Tasks    []Task         `yaml:"tasks"`
	Settings map[string]int `yaml:"settings"`
}

// Function: LoadPlaybook
// Purpose: Loads a YAML playbook from a file.
func LoadPlaybook(filePath string) (*Playbook, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read YAML file: %v", err)
		return nil, err
	}

	var playbook Playbook
	err = yaml.Unmarshal(data, &playbook)
	if err != nil {
		log.Printf("Failed to parse YAML file: %v", err)
		return nil, err
	}

	return &playbook, nil
}
