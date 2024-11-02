package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
)

type Communicator struct {
	task Task
	client *ssh.Client
}

func NewCommunicator(task Task) *Communicator {
	return &Communicator{task: task}
}

func (c *Communicator) Connect() error {
	config := &ssh.ClientConfig{
		User: c.task.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.task.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", c.task.Target+":22", config)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", c.task.Target, err)
	}
	c.client = client
	return nil
}

func (c *Communicator) Disconnect() {
	if c.client != nil {
		c.client.Close()
	}
}

func (c *Communicator) RunCommand(cmd string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("command execution failed: %v", err)
	}
	return string(output), nil
}

