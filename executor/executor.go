// File: executor.go
// Directory Path: /EagleDeploy_CLI/executor

package executor

import (
	"EagleDeploy_CLI/sshutils"
	"EagleDeploy_CLI/tasks"
	"fmt"
	"log"
	"os"
)

// Function: ExecuteRemote
// Purpose: Executes a task on a remote machine using SSH.
// Parameters:
// - task: The task to execute.
// - port: The port number for the SSH connection.
// Returns: An error if the task execution fails.
func ExecuteRemote(task tasks.Task, port int) error {
	// Fetching SSH credentials securely from the environment
	username := os.Getenv("USER_1_USERNAME")
	password := os.Getenv("USER_1_PASSWORD")

	if username == "" || password == "" {
		return fmt.Errorf("SSH username or password not set in environment variables")
	}

	// Connecting to the remote host
	client, err := sshutils.ConnectSSH(task.Host, username, password, port)
	if err != nil {
		log.Printf("Failed to connect to host %s on port %d: %v", task.Host, port, err)
		return err
	}
	defer client.Close()

	// Executing the task command
	output, err := sshutils.RunSSHCommand(client, task.Command)
	if err != nil {
		log.Printf("Failed to execute task '%s' on host %s: %v", task.Name, task.Host, err)
		return err
	}

	fmt.Printf("Output of task '%s' on host %s:\n%s", task.Name, task.Host, output)
	return nil
}

// Function: ExecuteLocal
// Purpose: Executes a command locally on the machine.
// Parameters:
// - command: The shell command to execute.
// Returns: An error if the command execution fails.
func ExecuteLocal(command string) error {
	log.Printf("Executing local command: %s", command)
	output, err := sshutils.RunLocalCommand(command)
	if err != nil {
		log.Printf("Failed to execute local command '%s': %v", command, err)
		return err
	}

	log.Printf("Local command '%s' executed successfully: %s", command, output)
	return nil
}
