// File: tasks.go
// Directory: EagleDeployment\tasks
// Purpose: Defines the structure and loading functionality for playbooks and tasks.

package tasks

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Type: Task
// Purpose: Represents a single executable task in a playbook
// Fields:
//   - Name: string - Descriptive name of the task
//   - Command: string - Shell command to execute
//   - SSHUser: string - Username for SSH connection
//   - SSHPassword: string - Password for SSH connection
//   - Host: string - Target host for task execution
//   - Port: int - SSH port number
//
// Used By:
//   - [`executor.ExecuteRemote`](../executor/executor.go)
//   - [`Playbook`](tasks.go) as task collection
type Task struct {
	Name        string `yaml:"name"`         // The name of the task
	Command     string `yaml:"command"`      // The command to execute
	SSHUser     string `yaml:"ssh_user"`     // The SSH user for the task
	SSHPassword string `yaml:"ssh_password"` // The SSH password for the task
	Host        string `yaml:"host"`         // The host for the task
	Port        int    `yaml:"port"`         // The port for the task
}

// Type: Playbook
// Purpose: Represents a complete automation playbook
// Fields:
//   - Name: string - Playbook identifier
//   - Version: string - Playbook version for tracking
//   - Hosts: []string - Target hosts list
//   - Tasks: []Task - List of tasks to execute
//   - Settings: map[string]string - Configuration options
//
// Used By:
//   - [`LoadPlaybook`](tasks.go)
//   - [`executeYAML`](../main.go)
type Playbook struct {
	Name     string            `yaml:"name"`     // The name of the playbook
	Version  string            `yaml:"version"`  // The version of the playbook
	Hosts    []string          `yaml:"hosts"`    // List of hosts targeted by the playbook
	Tasks    []Task            `yaml:"tasks"`    // List of tasks in the playbook
	Settings map[string]string `yaml:"settings"` // Additional settings like retries, timeouts
}

// Function: LoadPlaybook
// Purpose: Loads and parses a YAML playbook file into a Playbook structure
// Parameters:
//   - filePath: string - Path to the YAML playbook file
//
// Returns:
//   - *Playbook - Pointer to parsed playbook structure
//   - error - Any error encountered during loading/parsing
//
// Called By:
//   - [`executeYAML`](../main.go) when executing playbooks
//
// Dependencies:
//   - io/ioutil for file reading
//   - gopkg.in/yaml.v2 for YAML parsing
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
