package cmd

import "fmt"

// HelpCommand prints the available commands
func HelpCommand() {
	fmt.Println("Available commands:")
	fmt.Println("  tide [file]               - Execute yml file")
	fmt.Println("  swim [file/keyword]       - Search yml file")
	fmt.Println("  exit                      - Exit the framework")
}
