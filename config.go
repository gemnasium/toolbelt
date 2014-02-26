package main

import (
	"io/ioutil"
	"launchpad.net/goyaml"
)

type Config struct {
	Api_key        string
	Site           string
	Use_ssl        bool
	Api_version    string
	Project_branch string
	Ignored_paths  []string
}

// Load config from config file
//
func NewConfig(config_data []byte) (*Config, error) {
	config := &Config{
		Site:           "gemnasium.com",
		Use_ssl:        true,
		Api_version:    "v3",
		Project_branch: "master",
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
