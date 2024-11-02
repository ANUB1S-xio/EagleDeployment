package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Task struct {
	Name     string `yaml:"name"`
	Command  string `yaml:"command"`
	Target   string `yaml:"target"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func ParsePlaybook(filename string) ([]Task, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %v", err)
	}

	var tasks []Task
	err = yaml.Unmarshal(data, &tasks)
	if err != nil {
		return nil, fmt.Errorf("unable to parse YAML: %v", err)
	}

	return tasks, nil
}
