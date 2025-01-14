// File: executor.go
// Directory Path: /EagleDeploy_CLI/executor

package executor

import (
	"EagleDeploy_CLI/sshutils"
	"EagleDeploy_CLI/tasks"
	"fmt"
	"log"
	"sync"
)

// Function: ExecuteRemote
// Purpose: Executes a task on a remote machine using SSH.
// Parameters:
// - task: The task to execute.
// - port: The port number for the SSH connection.
// Returns: An error if the task execution fails.
func ExecuteRemote(task tasks.Task, port int) error {
	// Hardcoded SSH credentials for now
	username := "hunter"
	password := "What a nice day"

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

// Function: ExecuteConcurrently
// Purpose: Executes tasks concurrently across multiple hosts.
// Parameters:
// - tasks: A slice of tasks.Task objects to execute.
// - hosts: A slice of host addresses to execute tasks on.
// - port: The port number for the SSH connection.
// Behavior:
// - Uses goroutines to execute each task on each host.
// - Collects results and errors via channels for logging.
// Returns: None.
func ExecuteConcurrently(taskList []tasks.Task, hosts []string, port int) {
	var wg sync.WaitGroup
	results := make(chan string, len(taskList)*len(hosts))
	errors := make(chan error, len(taskList)*len(hosts))

	// Iterate over tasks and hosts to execute them concurrently
	for _, task := range taskList {
		for _, host := range hosts {
			wg.Add(1)
			go func(t tasks.Task, h string) {
				defer wg.Done()
				t.Host = h
				err := ExecuteRemote(t, port)
				if err != nil {
					errors <- fmt.Errorf("task '%s' on host '%s' failed: %v", t.Name, h, err)
				} else {
					results <- fmt.Sprintf("Task '%s' executed successfully on host '%s'", t.Name, h)
				}
			}(task, host)
		}
	}

	// Close channels after all goroutines complete
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Log results and errors
	for res := range results {
		log.Println(res)
	}
	for err := range errors {
		log.Printf("Error: %v", err)
	}
}
