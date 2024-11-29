package executor

import (
	"EagleDeploy_CLI/sshutils"
	"EagleDeploy_CLI/tasks"
	"fmt"
	"log"
)

func ExecuteRemote(task tasks.Task, port int) error {
	client, err := sshutils.ConnectSSH(task.Host, task.SSHUser, task.SSHPassword, port)
	if err != nil {
		log.Printf("Failed to connect to host %s on port %d: %v", task.Host, port, err)
		return err
	}
	defer client.Close()

	output, err := sshutils.RunSSHCommand(client, task.Command)
	if err != nil {
		log.Printf("Failed to execute task '%s' on host %s: %v", task.Name, task.Host, err)
		return err
	}

	fmt.Printf("Output of task '%s' on host %s:\n%s", task.Name, task.Host, output)
	return nil
}

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
