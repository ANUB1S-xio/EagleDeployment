// File: tasks.go
// Directory: EagleDeployment\tasks
// Purpose: Defines the structure and loading functionality for playbooks and tasks.

package tasks

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Task struct {
	Name        string `yaml:"name"`         // The name of the task
	Command     string `yaml:"command"`      // The command to execute
	SSHUser     string `yaml:"ssh_user"`     // The SSH user for the task
	SSHPassword string `yaml:"ssh_password"` // The SSH password for the task
	Host        string `yaml:"host"`         // The host for the task
	Port        int    `yaml:"port"`         // The port for the task
}

type Playbook struct {
	Name     string            `yaml:"name"`     // The name of the playbook
	Version  string            `yaml:"version"`  // The version of the playbook
	Hosts    []string          `yaml:"hosts"`    // List of hosts targeted by the playbook
	Tasks    []Task            `yaml:"tasks"`    // List of tasks in the playbook
	Settings map[string]string `yaml:"settings"` // Additional settings like retries, timeouts
}

// LoadPlaybook loads a playbook from a YAML file.
// Parameters:
// - filePath: The path to the playbook file.
// Returns:
// - A Playbook instance or an error if loading fails.
func LoadPlaybook(filePath string) (*Playbook, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var playbook Playbook
	err = yaml.Unmarshal(data, &playbook)
	if err != nil {
		return nil, err
	}
	return &playbook, nil
}
