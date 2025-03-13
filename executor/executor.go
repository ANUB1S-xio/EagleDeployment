// File: executor.go
// Directory Path: /EagleDeployment/executor
// Purpose: Executes tasks remotely using SSH and integrates with inventory management.
// Directory Path: /EagleDeployment/executor
// Purpose: Executes tasks defined in YAML playbooks via SSH concurrently, remotely, or locally.

package executor

import (
	"EagleDeployment/Telemetry"
	"EagleDeployment/inventory"
	"EagleDeployment/sshutils"
	"EagleDeployment/tasks"
	"EagleDeployment/config"

	"fmt"
	"log"
	"strconv"
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
	t := telemetry.GetInstance()

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
	output, err := sshutils.RunSSHCommand(client, task.Command)
	if err != nil {
		log.Printf("Failed executing '%s' on host %s: %v", task.Name, task.Host, err)
		return err
	}

	fmt.Printf("Output of task '%s' on host %s:\n%s\n", task.Name, task.Host, output)
	return nil
}

// ExecuteLocal executes a command locally on the machine.
func ExecuteLocal(command string) error {
	log.Printf("Executing local command: %s", command)
	output, err := sshutils.RunLocalCommand(command)
	if err != nil {
		log.Printf("Failed executing local command '%s': %v", command, err)
		return err
	}

	log.Printf("Local command output: %s", output)
	return nil
}

// ExecuteConcurrently executes multiple tasks across multiple hosts concurrently.
func ExecuteConcurrently(taskList []tasks.Task, hosts []string, port int) {
	t := telemetry.GetInstance()

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

// ExecuteYAML processes a YAML playbook by injecting inventory data and executing tasks concurrently.
func ExecuteYAML(playbookPath string, targetHosts []string) {
	processedPlaybook := "./playbooks/processed_playbook.yaml"

	// Inject inventory data into playbook template
	err := inventory.InjectInventoryIntoPlaybook(playbookPath, processedPlaybook)
	if err != nil {
		log.Fatalf("Inventory injection failed: %v", err)
		return
	}

	// Load processed YAML into playbook struct
	playbook := &tasks.Playbook{}
	err = config.LoadConfig(processedPlaybook, playbook)
	if err != nil {
		log.Fatalf("Playbook load failed: %v", err)
		return
	}

	if len(playbook.Tasks) == 0 {
		log.Fatalf("No tasks found in playbook: %s", processedPlaybook)
		return
	}

	// Use provided targetHosts or default from playbook
	hosts := playbook.Hosts
	if len(targetHosts) > 0 {
		hosts = targetHosts
	}
	// Convert port from string to int
	portStr := playbook.Settings["port"]
	port, err := strconv.Atoi(portStr)
	if err != nil || port == 0 {
		log.Fatalf("Playbook settings missing or invalid port: %v", err)
		return
	}

	fmt.Printf("Executing Playbook: '%s' on Hosts: %v\n", playbook.Name, hosts)
	ExecuteConcurrently(playbook.Tasks, hosts, port)
	ExecuteConcurrently(playbook.Tasks, hosts, port)
}
