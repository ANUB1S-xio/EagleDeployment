// File: sshutils.go
// Directory Path: /EagleDeploy_CLI/sshutils

package sshutils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"

	//"path/filepath"
	//"strings"

	"golang.org/x/crypto/ssh"
)

// Function: ConnectSSH
// Purpose: Establishes an SSH connection to a remote host.
// Parameters:
// - host: The hostname or IP of the remote machine.
// - user: The SSH username.
// - password: The SSH password.
// - port: The SSH port number.
// Returns: An SSH client connection or an error.
func ConnectSSH(host, user, password string, port int) (*ssh.Client, error) {
	log.Printf("Connecting to SSH server: %s on port: %d as user: %s", host, port, user)

	// Fetch password from environment if not provided
	if password == "" {
		password = os.Getenv("SSH_PASSWORD")
		if password == "" {
			return nil, fmt.Errorf("SSH password not set in environment variables")
		}
	}

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
		log.Printf("Failed to connect to SSH server: %s. Error: %v", address, err)
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	log.Printf("Successfully connected to SSH server: %s", address)

	return client, nil
}

// Function: RunSSHCommand
// Purpose: Executes a command on a remote host via SSH.
// Parameters:
// - client: The SSH client connection.
// - command: The command to execute.
// Returns: The output of the command or an error.
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
// Purpose: Executes a shell command locally on the machine.
// Parameters:
// - command: The shell command to execute.
// Returns: The output of the command or an error.
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
