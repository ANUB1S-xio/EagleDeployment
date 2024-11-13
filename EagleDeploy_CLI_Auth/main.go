package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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

type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

const userFilePath = "users.json"

// Load users from users.json
func loadUsers() ([]User, error) {
	var users []User
	file, err := ioutil.ReadFile(userFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			err = ioutil.WriteFile(userFilePath, []byte("[]"), 0644)
			if err != nil {
				return nil, err
			}
			return users, nil
		}
		return nil, err
	}
	if len(file) == 0 {
		err = ioutil.WriteFile(userFilePath, []byte("[]"), 0644)
		if err != nil {
			return nil, err
		}
		return users, nil
	}
	err = json.Unmarshal(file, &users)
	return users, err
}

// Save users to users.json
func saveUsers(users []User) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(userFilePath, data, 0644)
}

// Register a new user
func registerUser(username, password string) error {
	users, err := loadUsers()
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Username == username {
			return fmt.Errorf("user already exists")
		}
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	users = append(users, User{Username: username, PasswordHash: string(hashedPassword)})
	return saveUsers(users)
}

// Authenticate a user
func authenticateUser(username, password string) bool {
	users, err := loadUsers()
	if err != nil {
		fmt.Println("Error loading users:", err)
		return false
	}
	for _, user := range users {
		if user.Username == username {
			err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
			if err == nil {
				return true
			}
			break
		}
	}
	return false
}

// Display menu and get user choice
func displayMenu() int {
	fmt.Println("\n-----------------------------------")
	fmt.Println("\n       EagleDeploy Menu:")
	fmt.Println("\n-----------------------------------")
	fmt.Println("\n1. Execute a Playbook")
	fmt.Println("2. List YAML Files")
	fmt.Println("3. Manage Inventory")
	fmt.Println("4. Enable/Disable Detailed Logging")
	fmt.Println("5. Rollback Changes")
	fmt.Println("6. Show Help")
	fmt.Println("0. Logout")
	fmt.Print("\nSelect an option: ")

	var choice int
	fmt.Scanln(&choice)
	return choice
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

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n-----------------------------------")
		fmt.Println("\n Welcome to EagleDeploy CLI Auth!")
		fmt.Println("\n-----------------------------------")
		fmt.Println("1. Register")
		fmt.Println("2. Login")
		fmt.Println("3. Exit")
		fmt.Print("\nChoose an option: ")

		var choice int
		fmt.Scan(&choice)

		var username, password string
		var targetHosts []string

		switch choice {
		case 1:
			fmt.Print("\nEnter username: ")
			fmt.Scan(&username)
			fmt.Print("Enter password: ")
			fmt.Scan(&password)
			if err := registerUser(username, password); err != nil {
				fmt.Println("\nRegistration error:", err)
			} else {
				fmt.Println("\nRegistration successful!")
			}
		case 2:
			fmt.Print("\nEnter username: ")
			fmt.Scan(&username)
			fmt.Print("Enter password: ")
			fmt.Scan(&password)
			if authenticateUser(username, password) {
				fmt.Println("\nLogin successful!")
				// Main menu loop after successful login
				for {
					choice := displayMenu()
					switch choice {
					case 1: // Execute a Playbook
						for {
							fmt.Print("\nEnter the path to the YAML playbook file (or type 'back' to return to the menu): ")
							ymlFilePath, _ := reader.ReadString('\n')
							ymlFilePath = strings.TrimSpace(ymlFilePath)
							if ymlFilePath == "back" {
								break
							}

							fmt.Print("\nEnter comma-separated list of target hosts (leave empty for all in playbook): ")
							hosts, _ := reader.ReadString('\n')
							hosts = strings.TrimSpace(hosts)
							if hosts != "" {
								targetHosts = strings.Split(hosts, ",")
							}

							executeYAML(ymlFilePath, targetHosts)
						}

					case 2: // List YAML Files
						for {
							fmt.Print("\nEnter keyword to filter YAML files (or type 'back' to return to the menu): ")
							keyword, _ := reader.ReadString('\n')
							keyword = strings.TrimSpace(keyword)
							if keyword == "back" {
								break
							}
							listYAMLFiles(keyword)
						}

					case 3: // Manage Inventory
						fmt.Println("\nManaging inventory (not yet implemented).")

					case 4: // Enable/Disable Detailed Logging
						for {
							fmt.Print("\nEnable detailed logging? (y/n, or type 'back' to return to the menu): ")
							answer, _ := reader.ReadString('\n')
							answer = strings.TrimSpace(answer)
							if answer == "back" {
								break
							}
							if answer == "y" {
								fmt.Println("\nDetailed logging enabled.")
								break
							} else if answer == "n" {
								fmt.Println("\nDetailed logging disabled.")
								break
							} else {
								fmt.Println("\nInvalid option. Please enter 'y' or 'n'.")
							}
						}

					case 5: // Rollback Changes
						fmt.Println("Rolling back recent changes (not yet implemented).")

					case 6: // Help
						fmt.Println("Help Page:")
						fmt.Println("-e <yaml-file>: Execute the specified YAML file.")
						fmt.Println("-l <keyword>: List YAML files or related names in the EagleDeployment directory.")
						fmt.Println("-hosts <comma-separated-hosts>: Specify hosts to target (only with -e).")
						fmt.Println("-h: Display this help page.")

					case 0: // Logout
						fmt.Println("\n\nLogging out...")
						break // Exit the main menu loop to return to the login menu

					default:
						fmt.Println("\nInvalid choice. Please try again.")
					}

					// Break out of the for loop when choice is 0 to return to the login menu
					if choice == 0 {
						break
					}
				}
			} else {
				fmt.Println("\nInvalid username or password.")
			}
		case 3:
			fmt.Println("\n-----------------------------------")
			fmt.Println("\n      Exiting EagleDeploy.")
			fmt.Println("\n			   Thank You")
			fmt.Println("\n-----------------------------------")
			return
		default:
			fmt.Println("\nInvalid choice.")
		}
	}
}