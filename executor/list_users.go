package executor

import (
	"EagleDeployment/inventory"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"gopkg.in/yaml.v2"
)

// Playbook structure
type Playbook struct {
	Name  string   `yaml:"name"`
	Hosts []string `yaml:"hosts"`
	Tasks []Task   `yaml:"tasks"`
}

// Task structure
type Task struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

// ListUsersHandler - Reads List_Users.yaml and returns user data
func ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Inject Inventory into Playbook
	err := inventory.InjectInventoryIntoPlaybook("playbooks/List_Users.yaml", "playbooks/processed_List_Users.yaml")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to inject inventory: %v", err), http.StatusInternalServerError)
		return
	}

	// Step 2: Read the processed YAML file (not the original one)
	data, err := ioutil.ReadFile("playbooks/processed_List_Users.yaml")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read processed YAML: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Println("Raw YAML Content:\n", string(data))

	// Step 3: Parse YAML into struct
	var playbook Playbook
	err = yaml.Unmarshal(data, &playbook)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse YAML: %v", err), http.StatusInternalServerError)
		return
	}

	// Step 4: Extract users from tasks
	var users []string
	for _, task := range playbook.Tasks {
		if strings.Contains(task.Name, "List users") {
			lines := strings.Split(task.Command, "\n")
			for _, line := range lines {
				if line != "" && !strings.Contains(line, "TASK") {
					users = append(users, strings.TrimSpace(line))
				}
			}
		}
	}

	// Step 5: Return as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
