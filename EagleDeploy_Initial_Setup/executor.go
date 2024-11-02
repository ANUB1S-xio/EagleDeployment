package main

import "fmt"

type Executor struct {
	task Task
}

func NewExecutor(task Task) *Executor {
	return &Executor{task: task}
}

func (e *Executor) Execute() error {
	communicator := NewCommunicator(e.task)
	err := communicator.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer communicator.Disconnect()

	output, err := communicator.RunCommand(e.task.Command)
	if err != nil {
		return fmt.Errorf("command execution failed: %v", err)
	}
	fmt.Printf("Output of task %s: %s\n", e.task.Name, output)
	return nil
}

