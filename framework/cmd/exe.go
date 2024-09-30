package cmd

import "fmt"

func executeCommand(input string) {
    // Example: handle file execution logic
    fmt.Println("Executing file:", input)
}

func searchCommand(input string) {
    
    fmt.Println("Searching file:", input)
}

func catCommand(input string) {
    // Example: handle file display logic
    fmt.Println("Displaying file contents for:", input)
}

// HelpCommand prints the available commands
func HelpCommand() {
	fmt.Println("Available commands:")
	fmt.Println("  tide [file]               - Execute file")
	fmt.Println("  swim [file/keyword]       - Search file")
	fmt.Println("  exit                      - Exit the framework")
}