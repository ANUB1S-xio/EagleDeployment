package sshutils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

// Function to establish an SSH connection
func ConnectSSH(host, user, password string, port int) (*ssh.Client, error) {
	// Logging input parameters for verbose mode
	log.Printf("Connecting to SSH server: %s on port: %d as user: %s", host, port, user)

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // WARNING: Not recommended for production
	}

	address := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Attempting to dial %s...", address)

	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		// Logging detailed error information
		log.Printf("Failed to connect to SSH server: %s. Error: %v", address, err)
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	// Logging successful connection
	log.Printf("Successfully connected to SSH server: %s", address)

	return client, nil
}

// Function to run an SSH command
func RunSSHCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %w\n%s", err, output)
	}
	return string(output), nil
}

// Function to list YAML files based on a keyword
func ListYAMLFiles(keyword string) {
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			if strings.Contains(path, keyword) {
				fmt.Println("Found YAML file:", path)
			}
		}
		return nil
	})
}

// RunLocalCommand executes a shell command locally and returns the output or an error.
func RunLocalCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to execute local command: %w\nstderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
