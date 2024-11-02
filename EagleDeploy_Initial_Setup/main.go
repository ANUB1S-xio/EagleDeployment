package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: eagle <playbook.yaml>")
		os.Exit(1)
	}

	playbookFile := os.Args[1]
	tasks, err := ParsePlaybook(playbookFile)
	if err != nil {
		log.Fatalf("Failed to parse playbook: %v", err)
	}

	scheduler := NewScheduler(tasks)
	scheduler.RunTasks()
}
