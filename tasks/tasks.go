package tasks

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Task struct {
	Name        string `yaml:"name"`
	Command     string `yaml:"command"`
	SSHUser     string `yaml:"ssh_user"`
	SSHPassword string `yaml:"ssh_password"`
	Host        string `yaml:"host"` // Add host field
	Port        int    `yaml:"port"` // Add port field
}
type Playbook struct {
	Name     string         `yaml:"name"`
	Version  string         `yaml:"version"`
	Hosts    []string       `yaml:"hosts"`
	Tasks    []Task         `yaml:"tasks"`
	Settings map[string]int `yaml:"settings"`
}

// Functions
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
