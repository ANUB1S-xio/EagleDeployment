package main

import (
    "encoding/json"
    "fmt"
    "golang.org/x/crypto/bcrypt"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "log"
    "os"
    "os/exec"
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
			// Create an empty JSON array if the file doesn't exist
			err = ioutil.WriteFile(userFilePath, []byte("[]"), 0644)
			if err != nil {
				return nil, err
			}
			return users, nil
		}
		return nil, err
	}

	// If file is empty, initialize it with an empty array
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

	// Check if username already exists
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

// Main entry point
func main() {
	fmt.Println("Welcome to EagleDeploy CLI Auth!")
	fmt.Println("1. Register")
	fmt.Println("2. Login")
	fmt.Print("Choose an option: ")
	var choice int
	fmt.Scan(&choice)

	var username, password string
	switch choice {
	case 1:
		fmt.Print("Enter username: ")
		fmt.Scan(&username)
		fmt.Print("Enter password: ")
		fmt.Scan(&password)
		if err := registerUser(username, password); err != nil {
			fmt.Println("Registration error:", err)
		} else {
			fmt.Println("Registration successful!")
		}
	case 2:
		fmt.Print("Enter username: ")
		fmt.Scan(&username)
		fmt.Print("Enter password: ")
		fmt.Scan(&password)
		if authenticateUser(username, password) {
			fmt.Println("Login successful!")
			displayMenu() // Proceed to main menu or other functionality
		} else {
			fmt.Println("Invalid username or password.")
		}
	default:
		fmt.Println("Invalid choice.")
	}
}

// Example main menu that could be expanded with other functionalities
func displayMenu() {
	fmt.Println("Main Menu")
	fmt.Println("1. Execute Playbook")
	fmt.Println("2. Exit")
	fmt.Print("Choose an option: ")

	var option int
	fmt.Scan(&option)
	switch option {
	case 1:
		runPlaybook()
	case 2:
		fmt.Println("Exiting.")
		os.Exit(0)
	default:
		fmt.Println("Invalid option.")
	}
}

// Run the playbook from playbook.yaml
func runPlaybook() {
	playbook, err := parsePlaybook("playbook.yaml")
	if err != nil {
		log.Fatalf("Failed to parse playbook: %v", err)
	}
	fmt.Printf("Executing playbook: %s (version %s)\n", playbook.Name, playbook.Version)

	for _, task := range playbook.Tasks {
		fmt.Printf("Running task: %s\n", task.Name)
		runTask(task.Command)
	}
}

// Parse playbook.yaml
func parsePlaybook(filename string) (*Playbook, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}
	var playbook Playbook
	err = yaml.Unmarshal(data, &playbook)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %v", err)
	}
	return &playbook, nil
}

// Run a shell command as part of a task
func runTask(command string) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to run task: %v\n", err)
	}
}
