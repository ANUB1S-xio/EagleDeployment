package internal

import (
    "gopkg.in/yaml.v2"
    "os"
    "log"
)

type Config struct {
    Applications []Application `yaml:"applications"`
}

// Fetch applications from a YAML file
func LoadApplicationsFromYAML(path string) ([]Application, error) {
    var config Config

    // Use os.ReadFile instead of ioutil.ReadFile
    data, err := os.ReadFile(path)
    if err != nil {
        log.Printf("Error reading YAML file: %v", err)
        return nil, err
    }

    // Unmarshal the YAML data
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        log.Printf("Error unmarshalling YAML: %v", err)
        return nil, err
    }

    return config.Applications, nil
}

