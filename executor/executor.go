// File: executor.go
// Directory Path: /EagleDeploy_CLI/executor
// Purpose: Executes tasks remotely using SSH and integrates with inventory management.

package executor

import (
	"EagleDeploy_CLI/Telemetry"
	"EagleDeploy_CLI/inventory"
	"EagleDeploy_CLI/sshutils"
	"EagleDeploy_CLI/tasks"
	"fmt"
	"log"
	"sync"
)

// Function: ExecuteRemote
// Purpose: Executes a single task on a remote machine via SSH
// Parameters:
//   - task: tasks.Task - Task configuration to execute
//   - port: int - SSH port number
//
// Returns:
//   - error - Any execution errors
//
// Called By:
//   - ExecuteConcurrently during parallel execution
//
// Dependencies:
//   - inventory.GetSSHCreds for authentication
//   - sshutils.ConnectSSH for remote access
//   - sshutils.RunSSHCommand for command execution
func ExecuteRemote(task tasks.Task, port int) error {
	t := Telemetry.GetInstance()

	t.LogInfo("Execution", "Starting task execution", map[string]interface{}{
		"task_name": task.Name,
		"host":      task.Host,
		"port":      port,
	})

	// Fetch SSH credentials
	sshUser, sshPass := inventory.GetSSHCreds()

	// Override credentials if playbook provides them
	if task.SSHUser != "" {
		sshUser = task.SSHUser
		t.LogDebug("Execution", "Using task-specific SSH user", nil)
	}
	if task.SSHPassword != "" {
		sshPass = task.SSHPassword
		t.LogDebug("Execution", "Using task-specific SSH password", nil)
	}

	// Connect to remote host
	client, err := sshutils.ConnectSSH(task.Host, sshUser, sshPass, port)
	if err != nil {
		t.LogError("Execution", "SSH connection failed", map[string]interface{}{
			"host":  task.Host,
			"error": err.Error(),
		})
		return fmt.Errorf("connection failed to %s: %v", task.Host, err)
	}
	defer sshutils.CloseSSHConnection(client)

	t.LogDebug("Execution", "SSH connection established", map[string]interface{}{
		"host": task.Host,
	})

	// Execute the command (without printing output)
	output, err := sshutils.RunSSHCommand(client, task.Command)
	if err != nil {
		t.LogError("Execution", "Task execution failed", map[string]interface{}{
			"task_name": task.Name,
			"host":      task.Host,
			"command":   task.Command,
			"error":     err.Error(),
		})
		return fmt.Errorf("task '%s' failed: %v", task.Name, err)
	}

	// Log successful execution
	t.LogInfo("Execution", "Task executed successfully", map[string]interface{}{
		"task_name":     task.Name,
		"host":          task.Host,
		"command":       task.Command,
		"output_length": len(output),
	})

	// Only print task status
	log.Printf("Task '%s' executed successfully on host '%s'", task.Name, task.Host)
	return nil
}

// Function: ExecuteConcurrently
// Purpose: Manages parallel task execution across multiple hosts
// Parameters:
//   - taskList: []tasks.Task - List of tasks to execute
//   - hosts: []string - Target host addresses
//   - port: int - SSH port number
//
// Returns: None
// Called By:
//   - main.executeYAML during playbook execution
//
// Dependencies:
//   - inventory.GetHosts for host resolution
//   - sync.WaitGroup for concurrency control
//   - ExecuteRemote for individual task execution
//
// Notes:
//   - Uses goroutines for parallel execution
//   - Resolves hostnames to IPs using inventory
//   - Buffers results in channel for ordered logging
func ExecuteConcurrently(taskList []tasks.Task, hosts []string, port int) {
	t := Telemetry.GetInstance()

	t.LogInfo("Execution", "Starting concurrent task execution", map[string]interface{}{
		"tasks_count": len(taskList),
		"hosts_count": len(hosts),
		"port":        port,
	})

	var wg sync.WaitGroup
	results := make(chan string, len(taskList)*len(hosts))
	hostMap := inventory.GetHosts() // Get hosts from inventory.yaml

	// Iterate over tasks and execute them on hosts
	for _, task := range taskList {
		for _, host := range hosts {
			wg.Add(1)
			go func(t tasks.Task, h string) {
				defer wg.Done()

				// Resolve hostname to IP if necessary
				if hostData, exists := hostMap[h]; exists {
					t.Host = hostData.IP
				}

				err := ExecuteRemote(t, port)
				if err != nil {
					log.Printf("Task '%s' failed on host '%s': %v", t.Name, t.Host, err)
				} else {
					results <- fmt.Sprintf("Task '%s' executed successfully on host '%s'", t.Name, t.Host)
				}
			}(task, host)
		}
	}

	// Wait for all tasks to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Log results
	successCount := 0
	for res := range results {
		log.Println(res)
		successCount++
	}

	t.LogInfo("Execution", "Completed concurrent task execution", map[string]interface{}{
		"successful_tasks": successCount,
		"total_tasks":      len(taskList) * len(hosts),
	})
}
