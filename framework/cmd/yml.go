package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ExecuteYAMLCommand reads and executes commands from a YAML file in any directory
func ExecuteYAMLCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: execute <path-to-yaml-file>")
		return
	}

	// Accept the file path from user input
	filePath := args[0]

	// Ensure the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("File '%s' does not exist.\n", filePath)
		return
	}

	// Parse the YAML file
	commands, err := parseYAML(filePath)
	if err != nil {
		fmt.Println("Error parsing YAML file:", err)
		return
	}

	// Execute the commands
	for _, cmd := range commands {
		fmt.Printf("Executing: %s\n", cmd)
		runShellCommand(cmd)
	}
}

// parseYAML is a simple YAML parser to extract commands from a YAML file
func parseYAML(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var commands []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "- ") {
			command := strings.TrimPrefix(line, "- ")
			commands = append(commands, command)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return commands, nil
}

// runShellCommand executes a single shell command
func runShellCommand(command string) {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing command: %s\n", err)
	}
	fmt.Println(string(output))
}
