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

// Function: ConnectSSH
// Purpose: Establishes an SSH connection to a remote server
// Parameters:
//   - host: string - The remote server address
//   - user: string - The SSH username
//   - password: string - The SSH password
//   - port: int - The SSH port number
//
// Returns:
//   - *ssh.Client - SSH client instance
//   - error - Any connection errors
//
// Called By:
//   - [`executor.ExecuteRemote`](../executor/executor.go)
//
// Dependencies:
//   - golang.org/x/crypto/ssh
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

// Function: RunSSHCommand
// Purpose: Executes a command on a remote host via SSH
// Parameters:
//   - client: *ssh.Client - Active SSH connection
//   - command: string - Command to execute
//
// Returns:
//   - string - Command output
//   - error - Any execution errors
//
// Called By:
//   - [`executor.ExecuteRemote`](../executor/executor.go)
//
// Dependencies:
//   - Active SSH client connection
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

// Function: RunLocalCommand
// Purpose: Executes a shell command on the local system
// Parameters:
//   - command: string - Shell command to execute
//
// Returns:
//   - string - Command output
//   - error - Any execution errors
//
// Called By:
//   - Various playbook tasks requiring local execution
//
// Dependencies:
//   - os/exec for command execution
//   - bash shell availability
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
// Purpose: Safely closes an established SSH connection
// Parameters:
//   - client: *ssh.Client - The SSH connection to close
//
// Returns:
//   - error - Any errors during connection closure
//
// Called By:
//   - [`executor.ExecuteRemote`](../executor/executor.go)
//   - Cleanup routines
//
// Notes:
//   - Handles nil client gracefully
//   - Logs connection closure status
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
