package main

import (
	"bufio"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Structs for the YAML structure
type Task struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}
type Playbook struct {
	Name     string   `yaml:"name"`
	Version  string   `yaml:"version"`
	Tasks    []Task   `yaml:"tasks"`
	Hosts    []string `yaml:"hosts"`
	Settings map[string]int `yaml:"settings"`
}

// Function to execute the YAML file by parsing its content
func executeYAML(ymlFilePath string, targetHosts []string) {
	data, err := ioutil.ReadFile(ymlFilePath)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

	var playbook Playbook
	err = yaml.Unmarshal(data, &playbook)
	if err != nil {
		log.Fatalf("Error parsing YAML file: %v", err)
	}

	if len(playbook.Tasks) == 0 {
		log.Fatalf("Error: No tasks found in the playbook.")
	}

	var hosts []string
	if len(targetHosts) > 0 {
		for _, host := range playbook.Hosts {
			if contains(targetHosts, host) {
				hosts = append(hosts, host)
			}
		}
		if len(hosts) == 0 {
			log.Fatalf("Error: No matching hosts found in the playbook for the provided targets.")
		}
	} else {
		hosts = playbook.Hosts
	}

	fmt.Printf("Executing Playbook: %s (Version: %s) on Hosts: %v\n", playbook.Name, playbook.Version, hosts)
	for _, task := range playbook.Tasks {
		if task.Command == "" {
			log.Fatalf("Error: Task '%s' has no command to execute.", task.Name)
		}

		fmt.Printf("Executing Task: %s\n", task.Name)
		cmd := exec.Command("bash", "-c", task.Command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error executing task '%s': %v\n", task.Name, err)
		} else {
			fmt.Printf("Output of '%s':\n%s\n", task.Name, string(output))
		}
	}
}

// Helper function to check if a slice contains a specific element
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Function to list YAML files based on a keyword in the current directory
func listYAMLFiles(keyword string) {
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			if strings.Contains(path, keyword) {
				fmt.Println("Found YAML file:", path)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error listing YAML files: %v", err)
	}
}

// Display menu and get user choice
func displayMenu() int {
	fmt.Println() // Adds a blank line for spacing
	fmt.Println("EagleDeploy Menu:")
	fmt.Println("1. Execute a Playbook")
	fmt.Println("2. List YAML Files")
	fmt.Println("3. Manage Inventory")
	fmt.Println("4. Enable/Disable Detailed Logging")
	fmt.Println("5. Rollback Changes")
	fmt.Println("6. Show Help")
	fmt.Println("0. Exit")
	fmt.Print("Select an option: ")

	var choice int
	fmt.Scanln(&choice)
	return choice
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	var targetHosts []string

	for {
		choice := displayMenu()
		switch choice {
		case 1: // Execute a Playbook
			for {
				fmt.Print("Enter the path to the YAML playbook file (or type 'back' to return to the menu): ")
				ymlFilePath, _ := reader.ReadString('\n')
				ymlFilePath = strings.TrimSpace(ymlFilePath)
				if ymlFilePath == "back" {
					break
				}
				
				fmt.Print("Enter comma-separated list of target hosts (leave empty for all in playbook): ")
				hosts, _ := reader.ReadString('\n')
				hosts = strings.TrimSpace(hosts)
				if hosts != "" {
					targetHosts = strings.Split(hosts, ",")
				}
				
				executeYAML(ymlFilePath, targetHosts)
			}

		case 2: // List YAML Files
			for {
				fmt.Print("Enter keyword to filter YAML files (or type 'back' to return to the menu): ")
				keyword, _ := reader.ReadString('\n')
				keyword = strings.TrimSpace(keyword)
				if keyword == "back" {
					break
				}
				listYAMLFiles(keyword)
			}

		case 3: // Manage Inventory
			fmt.Println("Managing inventory (not yet implemented).")
			// Add implementation for inventory management here

		case 4: // Enable/Disable Detailed Logging
			for {
				fmt.Print("Enable detailed logging? (y/n, or type 'back' to return to the menu): ")
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(answer)
				if answer == "back" {
					break
				}
				if answer == "y" {
					fmt.Println("Detailed logging enabled.")
					break
				} else if answer == "n" {
					fmt.Println("Detailed logging disabled.")
					break
				} else {
					fmt.Println("Invalid option. Please enter 'y' or 'n'.")
				}
			}

		case 5: // Rollback Changes
			fmt.Println("Rolling back recent changes (not yet implemented).")
			// Add rollback implementation here

		case 6: // Help
			fmt.Println("Help Page:")
			fmt.Println("-e <yaml-file>: Execute the specified YAML file.")
			fmt.Println("-l <keyword>: List YAML files or related names in the EagleDeployment directory.")
			fmt.Println("-hosts <comma-separated-hosts>: Specify hosts to target (only with -e).")
			fmt.Println("-h: Display this help page.")

		case 0: // Exit
			fmt.Println("Exiting EagleDeploy.")
			return

		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
