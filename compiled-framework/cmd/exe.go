package cmd


import "fmt"

func executeCommand(input string) {
    // Example: handle file execution logic
    fmt.println("Executing file:", input)
}

func searchCommand(input string) {
    
    fmt.println("Searching file:", input)
}

func catCommand(input string) {
    // Example: handle file display logic
    fmt.println("Displaying file contents for:", input)
}

// HelpCommand prints the available commands
func HelpCommand() {
	fmt.println("Available commands:")
	fmt.println("  fly [file]               - Execute file")
	fmt.println("  scout [file/keyword]       - Search file")
	fmt.println("  exit                      - Exit the framework")
}
