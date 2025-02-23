// File: executor.go
// Directory Path: /EagleDeploy_CLI/executor
// Purpose: Executes tasks defined in YAML playbooks via SSH concurrently, remotely, or locally.

package executor

import (
	"EagleDeploy_CLI/config"
	"EagleDeploy_CLI/inventory"
	"EagleDeploy_CLI/sshutils"
	"EagleDeploy_CLI/tasks"
	"fmt"
	"log"
	"sync"
)

// ExecuteRemote executes a task on a remote machine using SSH.
func ExecuteRemote(task tasks.Task, port int) error {
	username := task.SSHUser
	password := task.SSHPassword

	if username == "" || password == "" {
		return fmt.Errorf("SSH username or password missing for task '%s'", task.Name)
	}

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
	var wg sync.WaitGroup
	results := make(chan string, len(taskList)*len(hosts))
	errors := make(chan error, len(taskList)*len(hosts))

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

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	for res := range results {
		log.Println(res)
	}
	for err := range errors {
		log.Printf("Error: %v", err)
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
