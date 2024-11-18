package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

// Structs for the YAML structure
type Task struct {
	Name        string `yaml:"name"`
	Command     string `yaml:"command"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	SSHUser     string `yaml:"ssh_user"`
	SSHPassword string `yaml:"ssh_password"`
}

type Playbook struct {
	Name     string         `yaml:"name"`
	Version  string         `yaml:"version"`
	Tasks    []Task         `yaml:"tasks"`
	Hosts    []string       `yaml:"hosts"`
	Settings map[string]int `yaml:"settings"`
}

// Function to add a user via SSH
func addUserTask(client *ssh.Client, username, password string) error {
	command := fmt.Sprintf("useradd -m %s && echo '%s:%s' | chpasswd", username, username, password)
	output, err := runSSHCommand(client, command)
	if err != nil {
		return fmt.Errorf("failed to add user: %w", err)
	}
	fmt.Println("User added successfully:", output)
	return nil
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

	hosts := playbook.Hosts
	if len(targetHosts) > 0 {
		hosts = targetHosts
	}

	fmt.Printf("Executing Playbook: %s (Version: %s) on Hosts: %v\n", playbook.Name, playbook.Version, hosts)
	for _, task := range playbook.Tasks {
		if task.Command == "add_user" {
			for _, host := range hosts {
				client, err := connectSSH(host, task.SSHUser, task.SSHPassword, playbook.Settings["port"])
				if err != nil {
					fmt.Printf("Error connecting to host %s: %v\n", host, err)
					continue
				}
				defer client.Close()

				err = addUserTask(client, task.Username, task.Password)
				if err != nil {
					fmt.Printf("Error adding user on host %s: %v\n", host, err)
				}
			}
		} else {
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
	fmt.Println("\nEagleDeploy Menu:")
	fmt.Println("1. Execute a Playbook")
	fmt.Println("2. List YAML Files")
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

		case 0: // Exit
			fmt.Println("Exiting EagleDeploy.")
			return

		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
