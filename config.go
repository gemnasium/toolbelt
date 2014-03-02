package main

import (
	"io/ioutil"
	"launchpad.net/goyaml"
)

type Config struct {
	APIEndpoint   string   `yaml:"api_endpoint"`
	APIKey        string   `yaml:"api_key"`
	ProjectBranch string   `yaml:"project_branch"`
	IgnoredPaths  []string `yaml:"ignored_paths"`
}

// Load config from config file
//
func NewConfig(config_data []byte) (*Config, error) {
	config := &Config{
		APIEndpoint:   "https://gemnasium.com/api/v3",
		ProjectBranch: "master",
	}
	goyaml.Unmarshal(config_data, config)
	return config, nil

}

func LoadConfigFile(filepath string) (*Config, error) {

	config_data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return NewConfig(config_data)
}
