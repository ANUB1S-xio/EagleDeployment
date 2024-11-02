package main

import (
	"flag"
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

	// Validate required fields
	if len(playbook.Tasks) == 0 {
		log.Fatalf("Error: No tasks found in the playbook.")
	}

	// Determine target hosts
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
		// Execute the command (here just local example, can add SSH for remote)
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

func main() {
	// Parse command-line arguments
	var hostsFlag string
	flag.StringVar(&hostsFlag, "hosts", "", "Comma-separated list of hosts to target")
	flag.Parse()

	// Split the hostsFlag into a slice if provided
	var targetHosts []string
	if hostsFlag != "" {
		targetHosts = strings.Split(hostsFlag, ",")
	}

	// Get the base name of the tool to avoid path issues
	toolName := filepath.Base(os.Args[0])
	fmt.Println("Tool name is:", toolName) // Debug line to print tool name

	// Check if at least two arguments are provided (tool name and command)
	if len(flag.Args()) < 1 {
		fmt.Println("Usage: eagle <command> [options]")
		os.Exit(1)
	}

	// Process command based on the first argument (command)
	command := flag.Args()[0]
	switch command {
	case "-e": // Execute YAML file
		if len(flag.Args()) < 2 {
			fmt.Println("Error: '-e' requires a YAML file path as an additional argument.")
			os.Exit(1)
		}
		ymlFilePath := flag.Args()[1]
		fmt.Printf("Executing YAML file: %s\n", ymlFilePath)
		executeYAML(ymlFilePath, targetHosts)

	case "-l": // List YAML files or related names
		if len(flag.Args()) < 2 {
			fmt.Println("Error: '-l' requires a keyword or filename to list matching YAML files.")
			os.Exit(1)
		}
		listKeyword := flag.Args()[1]
		fmt.Printf("Listing YAML files related to: %s\n", listKeyword)
		listYAMLFiles(listKeyword)

	case "-h": // Help
		fmt.Println("Help Page:")
		fmt.Println("Commands:")
		fmt.Println("-e <yaml-file>: Execute the specified YAML file.")
		fmt.Println("-l <keyword>: List YAML files or related names in the EagleDeployment directory.")
		fmt.Println("-hosts <comma-separated-hosts>: Specify hosts to target (only with -e).")
		fmt.Println("-h: Display this help page.")

	default:
		fmt.Printf("Error: Unknown command '%s'. Use '-h' for help.\n", command)
		os.Exit(1)
	}
}
