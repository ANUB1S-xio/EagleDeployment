package main

import (
	"log"
	"sync"
)

type Scheduler struct {
	tasks []Task
}

func NewScheduler(tasks []Task) *Scheduler {
	return &Scheduler{tasks: tasks}
}

func (s *Scheduler) RunTasks() {
	var wg sync.WaitGroup
	for _, task := range s.tasks {
		wg.Add(1)
		go func(task Task) {
			defer wg.Done()
			executor := NewExecutor(task)
			err := executor.Execute()
			if err != nil {
				log.Printf("Task %s failed: %v", task.Name, err)
			} else {
				log.Printf("Task %s completed successfully", task.Name)
			}
		}(task)
	}
	wg.Wait()
}
