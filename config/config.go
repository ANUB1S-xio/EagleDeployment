package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func LoadConfig(filePath string, target interface{}) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, target)
}
