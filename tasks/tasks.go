// File: tasks.go
// Directory Path: /EagleDeploy_CLI/tasks

package tasks

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// Struct: Task
// Purpose: Represents a single task to be executed, including command, user, and host information.
type Task struct {
	Name        string `yaml:"name"`
	Command     string `yaml:"command"`
	SSHUser     string `yaml:"ssh_user"`
	SSHPassword string `yaml:"ssh_password"`
	Host        string `yaml:"host"` // Host for the task execution
	Port        int    `yaml:"port"` // Port for the task execution
}

// Struct: Playbook
// Purpose: Represents a collection of tasks to execute, with associated metadata and settings.
type Playbook struct {
	Name     string         `yaml:"name"`
	Version  string         `yaml:"version"`
	Hosts    []string       `yaml:"hosts"`
	Tasks    []Task         `yaml:"tasks"`
	Settings map[string]int `yaml:"settings"` // General settings like retries, timeouts, etc.
}

// Function: LoadPlaybook
// Purpose: Loads a YAML playbook file and unmarshals it into a Playbook struct.
// Parameters:
// - filePath: The file path to the YAML playbook.
// Returns: A pointer to the Playbook struct and an error if loading fails.
func LoadPlaybook(filePath string) (*Playbook, error) {
	log.Printf("Loading playbook from file: %s", filePath)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read playbook file: %v", err)
		return nil, err
	}

	var playbook Playbook
	err = yaml.Unmarshal(data, &playbook)
	if err != nil {
		log.Printf("Failed to unmarshal YAML data: %v", err)
		return nil, err
	}

	log.Printf("Playbook loaded successfully from file: %s", filePath)
	return &playbook, nil
}
