// File: executor.go
// Directory Path: /EagleDeployment/executor
// Purpose: Executes tasks remotely using SSH and integrates with inventory management.
// Directory Path: /EagleDeployment/executor
// Purpose: Executes tasks defined in YAML playbooks via SSH concurrently, remotely, or locally.

package executor

import (
	telemetry "EagleDeployment/Telemetry"
	"EagleDeployment/config"
	"EagleDeployment/inventory"
	"EagleDeployment/sshutils"
	"EagleDeployment/tasks"

	"fmt"
	"log"
	"os"
	"reflect"
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

	// // Fetch SSH credentials
	// sshUser, hostMap := inventory.GetHosts()
	// if hostData, ok := hostMap[task.Host]; ok {
	// 	sshUser = hostData.SSHUser
	// 	sshPass = hostData.SSHPass
	// } else {
	// 	sshUser, sshPass = inventory.GetSSHCreds()
	// }

	// os.Setenv("EAGLE_SSH_USER", sshUser)
	// os.Setenv("EAGLE_SSH_PASS", sshPass)



// 	// Fetch SSH credentials
// sshUser, hostMap := inventory.GetHosts()

// // Default to first host if task.Host is not set
// if task.Host == "" && len(inv.Hosts) > 0 {
// 	task.Host = inv.Hosts[0].IP
// }

// if hostData, ok := hostMap[task.Host]; ok {
// 	sshUser = hostData.SSHUser
// 	sshPass = hostData.SSHPass
// } else {
// 	sshUser, sshPass = inventory.GetSSHCreds()
// }



// Load inventory
inv, err := inventory.LoadInventory()
if err != nil {
	t.LogError("Playbook", "Failed to load inventory for playbook", map[string]interface{}{
		"error": err.Error(),
	})
	return fmt.Errorf("failed to load inventory: %v", err)
}

// Set default host if not specified
if task.Host == "" && len(inv.Hosts) > 0 {
	task.Host = inv.Hosts[0].IP
}

// Use MapHostsByIP to resolve credentials
hostMap := inventory.MapHostsByIP(inv)

var sshUser, sshPass string
if hostData, ok := hostMap[task.Host]; ok {
	sshUser = hostData.SSHUser
	sshPass = hostData.SSHPass
} else {
	sshUser, sshPass = inventory.GetSSHCreds()
}

// Export to environment for templating use
os.Setenv("EAGLE_SSH_USER", sshUser)
os.Setenv("EAGLE_SSH_PASS", sshPass)


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
	hostMap := inventory.GetHosts()

	// Debug print to verify host mapping
	log.Printf("[DEBUG] hostMap keys: %v", reflect.ValueOf(hostMap).MapKeys())
	log.Printf("[DEBUG] incoming host list from playbook: %v", hosts)

	for _, task := range taskList {
		for _, host := range hosts {
			var hostData inventory.Host
			found := false

			// Match host either by IP or Hostname
			for _, h := range hostMap {
				if h.IP == host || h.Hostname == host {
					hostData = h
					task.Host = h.IP
					found = true
					break
				}
			}

			if !found {
				log.Printf("[WARN] Host '%s' not found in inventory, skipping task '%s'", host, task.Name)
				continue
			}

			// Inject SSH creds from host or fallback to global
			if hostData.SSHUser != "" {
				task.SSHUser = hostData.SSHUser
			} else {
				task.SSHUser, _ = inventory.GetSSHCreds()
			}

			if hostData.SSHPass != "" {
				task.SSHPassword = hostData.SSHPass
			} else {
				_, task.SSHPassword = inventory.GetSSHCreds()
			}

			// Launch task in goroutine with closure safety
			wg.Add(1)
			go func(t tasks.Task) {
				defer wg.Done()
				err := ExecuteRemote(t, port)
				if err != nil {
					log.Printf("Task '%s' failed on host '%s': %v", t.Name, t.Host, err)
				} else {
					results <- fmt.Sprintf("Task '%s' executed successfully on host '%s'", t.Name, t.Host)
				}
			}(task)
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
