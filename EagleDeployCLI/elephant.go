package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("Hello, world!")

	fmt.Println("I am going to create a yaml file parser")
	fmt.Println("Welcome to the Eagle Deployment shell! Type 'exit' to quit")

	//hard coding the directory for the playbooks
	playbookDir := "./Playbooks/"

	scanner := bufio.NewScanner(os.Stdin) //Reads input from the user

	for {
		fmt.Print("EagleDeployment>")

		//Getting user input

		if !scanner.Scan() {
			//handle input error
			break
		}

		input := scanner.Text()
		if input == "" {
			continue //skip empty input
		}

		parts := strings.Fields(input)
		command := parts[0]
		args := parts[1:]

		switch command {
		case "exit":
			fmt.Println("Goodbye!")
			return

		case "eagle":
			if len(args) == 0 {
				fmt.Println("Error: No playbook specified. Usage: eagle <playbook>")
				continue
			}

			playbook := args[0]
			err := printPlaybookContents(playbookDir, playbook)
			if err != nil {
				fmt.Printf("Error reading playbook: %v\n", err)

			}
		case "echo":
			fmt.Println(strings.Join(args, " "))

		case "help":
			fmt.Println("Available commands:")
			fmt.Println("- exit: Quit the shell")
			fmt.Println("- eagle <playbook>: Runs playbook")
			fmt.Println("-  echo [text]: Print text to the terminal")
			fmt.Println("- help: Show this help message")

		default:
			fmt.Printf("Unknown command: %s\n", command)

		}
	}

	//handle end of input
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}
}

func printPlaybookContents(playbookDir, playbook string) error {

	filePath := playbookDir + playbook
	fmt.Printf("Attempting to open file: %s\n", filePath)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Printf("Contents of %s: \n", playbook)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	return scanner.Err()
}
