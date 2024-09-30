package internal

import (
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "log"
)

type Config struct {
    Applications []Application `yaml:"applications"`
}

// Fetch applications from a YAML file (mock for fetching from the web)
func LoadApplicationsFromYAML(path string) ([]Application, error) {
    var config Config
    data, err := ioutil.ReadFile(path)
    if err != nil {
        log.Printf("Error reading YAML file: %v", err)
        return nil, err
    }

    err = yaml.Unmarshal(data, &config)
    if err != nil {
        log.Printf("Error unmarshalling YAML: %v", err)
        return nil, err
    }

    return config.Applications, nil
}
