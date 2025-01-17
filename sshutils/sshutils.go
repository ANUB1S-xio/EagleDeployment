// File: sshutils.go
// Directory: EagleDeployment\sshutils
// Purpose: Provides utility functions for SSH and local command execution.
package sshutils

import (
	"bytes"
	"fmt"
	"log"

	//"os"
	"os/exec"
	//"path/filepath"
	//"strings"

	"golang.org/x/crypto/ssh"
)

// ConnectSSH establishes an SSH connection to a remote server.
// Parameters:
// - host: The remote server address.
// - user: The SSH username.
// - password: The SSH password.
// - port: The SSH port number.
// Returns:
// - An SSH client instance or an error if the connection fails.
func ConnectSSH(host, user, password string, port int) (*ssh.Client, error) {
	log.Printf("Connecting to SSH server: %s on port: %d as user: %s", host, port, user)

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	address := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		log.Printf("Failed to connect to SSH server: %s. Error: %v", address, err)
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	log.Printf("Successfully connected to SSH server: %s", address)
	return client, nil
}

// RunSSHCommand executes a command on a remote host via SSH.
// Parameters:
// - client: An SSH client instance.
// - command: The command to execute.
// Returns:
// - The command output or an error if execution fails.
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

// RunLocalCommand executes a shell command locally.
// Parameters:
// - command: The shell command to execute.
// Returns:
// - The command output or an error if execution fails.
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

// Function: CloseSSHConnection
// Purpose: Closes an established SSH connection safely.
// Parameters:
// - client: The SSH client connection to close.
// Returns: An error if the connection fails to close.
func CloseSSHConnection(client *ssh.Client) error {
	if client == nil {
		return nil // No connection to close
	}

	log.Println("Closing SSH connection...")
	err := client.Close()
	if err != nil {
		log.Printf("Failed to close SSH connection: %v", err)
		return err
	}

	log.Println("SSH connection closed successfully.")
	return nil
}
