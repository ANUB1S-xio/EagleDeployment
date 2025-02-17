// File: executor.go
// Directory Path: /EagleDeploy_CLI/executor
// Purpose: Executes tasks remotely using SSH and integrates with inventory management.

package executor

import (
	"EagleDeploy_CLI/inventory"
	"EagleDeploy_CLI/sshutils"
	"EagleDeploy_CLI/tasks"
	"fmt"
	"log"
	"sync"
)

// ExecuteRemote runs a task on a remote machine
func ExecuteRemote(task tasks.Task, port int) error {
	// Fetch SSH credentials
	sshUser, sshPass := inventory.GetSSHCreds()

	// Override credentials if playbook provides them
	if task.SSHUser != "" {
		sshUser = task.SSHUser
	}
	if task.SSHPassword != "" {
		sshPass = task.SSHPassword
	}

	// Connect to remote host
	client, err := sshutils.ConnectSSH(task.Host, sshUser, sshPass, port)
	if err != nil {
		return fmt.Errorf("connection failed to %s: %v", task.Host, err)
	}
	defer sshutils.CloseSSHConnection(client)

	// Execute the command (without printing output)
	_, err = sshutils.RunSSHCommand(client, task.Command)
	if err != nil {
		return fmt.Errorf("task '%s' failed: %v", task.Name, err)
	}

	// Only print task status
	log.Printf("Task '%s' executed successfully on host '%s'", task.Name, task.Host)
	return nil
}

// ExecuteConcurrently runs tasks on multiple hosts concurrently using inventory.yaml
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
