// File: executor.go
// Directory Path: /EagleDeploy_CLI/executor
// Purpose: Executes tasks remotely using SSH and integrates with inventory management.
// Purpose: Executes tasks defined in YAML playbooks via SSH concurrently, remotely, or locally.

package executor

import (
	"EagleDeploy_CLI/inventory"
	"EagleDeploy_CLI/config"
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
// ExecuteRemote executes a task on a remote machine using SSH.
func ExecuteRemote(task tasks.Task, port int) error {
	// Fetch SSH credentials
	sshUser, sshPass := inventory.GetSSHCreds()

	// Override credentials if playbook provides them
	if task.SSHUser != "" {
		sshUser = task.SSHUser
	}
	if task.SSHPassword != "" {
		sshPass = task.SSHPassword
	username := task.SSHUser
	password := task.SSHPassword

	if username == "" || password == "" {
		return fmt.Errorf("SSH username or password missing for task '%s'", task.Name)
	}

	// Connect to remote host
	client, err := sshutils.ConnectSSH(task.Host, sshUser, sshPass, port)
	client, err := sshutils.ConnectSSH(task.Host, username, password, port)
	if err != nil {
		log.Printf("Failed SSH connection to %s:%d: %v", task.Host, port, err)
		return err
	}
	defer func() {
		if cerr := sshutils.CloseSSHConnection(client); cerr != nil {
			log.Printf("Error closing SSH connection: %v", cerr)
		}
	}()

	output, err := sshutils.RunSSHCommand(client, task.Command)
	if err != nil {
		return fmt.Errorf("connection failed to %s: %v", task.Host, err)
	}
	defer sshutils.CloseSSHConnection(client)

	// Execute the command (without printing output)
	_, err = sshutils.RunSSHCommand(client, task.Command)
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
		return fmt.Errorf("task '%s' failed: %v", task.Name, err)
		log.Printf("Failed executing local command '%s': %v", command, err)
		return err
	}

	// Only print task status
	log.Printf("Task '%s' executed successfully on host '%s'", task.Name, task.Host)
	log.Printf("Local command output: %s", output)
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
// ExecuteConcurrently executes multiple tasks across multiple hosts concurrently.
func ExecuteConcurrently(taskList []tasks.Task, hosts []string, port int) {
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
	for res := range results {
		log.Println(res)
	}
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

	// port is already an int, no need for strconv
	port := playbook.Settings["port"]
	if port == 0 {
		log.Fatalf("Playbook settings missing or invalid port")
		return
	}

	fmt.Printf("Executing Playbook: '%s' on Hosts: %v\n", playbook.Name, hosts)
	ExecuteConcurrently(playbook.Tasks, hosts, port)
}
