package internal

import (
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
)

type Application struct {
    Name    string `json:"name"`
    Version string `json:"version"`
}

// This could be populated dynamically from a real data source or web scraping
var applications = []Application{
    {"nginx", "1.21.6"},
    {"redis", "7.0.3"},
}

// GetApplications serves the list of applications and versions
func GetApplications(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(applications)
}

// GetApplicationByName serves a single application and version by name
func GetApplicationByName(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    for _, app := range applications {
        if app.Name == params["name"] {
            json.NewEncoder(w).Encode(app)
            return
        }
    }
    http.Error(w, "Application not found", http.StatusNotFound)
}
