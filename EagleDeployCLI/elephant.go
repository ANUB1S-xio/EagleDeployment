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

		case "echo":
			fmt.Println(strings.Join(args, " "))

		case "help":
			fmt.Println("Available commands:")
			fmt.Println("- exit: Quit the shell")
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
