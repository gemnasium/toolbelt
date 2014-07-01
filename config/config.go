package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v1"
)

var (
	APIEndpoint   = DEFAULT_API_ENDPOINT
	ProjectBranch = DEFAULT_PROJECT_BRANCH
	APIKey,
	ProjectSlug string
	IgnoredPaths []string
	RawFormat    bool
)

const (
	CONFIG_FILE_PATH = ".gemnasium.yml"

	ENV_API_ENDDPOINT  = "GEMNASIUM_API_ENDPOINT"
	ENV_TOKEN          = "GEMNASIUM_TOKEN"
	ENV_PROJECT_SLUG   = "GEMNASIUM_PROJECT_SLUG"
	ENV_PROJECT_BRANCH = "GEMNASIUM_PROJECT_BRANCH"
	ENV_IGNORED_PATHS  = "GEMNASIUM_IGNORED_PATHS"
	ENV_RAW_FORMAT     = "GEMNASIUM_RAW_FORMAT"

	DEFAULT_API_ENDPOINT   = "https://api.gemnasium.com/v1"
	DEFAULT_PROJECT_BRANCH = "master"
)

func init() {
	loadConfig()
	loadEnv() // Env will override config file
}

func getEnvOrElse(name, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return value
}

func loadConfig() {
	dat, err := ioutil.ReadFile(CONFIG_FILE_PATH)
	if err != nil {
		return
	}
	//fmt.Println("Reading config file")
	c := make(map[string]interface{})
	err = yaml.Unmarshal(dat, &c)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if api_endpoint, ok := c["api_endpoint"]; ok {
		APIEndpoint = api_endpoint.(string)
	}
	if api_key, ok := c["api_key"]; ok {
		APIKey = api_key.(string)
	}
	if project_slug, ok := c["project_slug"]; ok {
		ProjectSlug = project_slug.(string)
	}
	if project_branch, ok := c["project_branch"]; ok {
		ProjectBranch = project_branch.(string)
	}
	if ignored_paths, ok := c["ignored_paths"]; ok {
		for _, ip := range ignored_paths.([]interface{}) {
			IgnoredPaths = append(IgnoredPaths, ip.(string))
		}
	}
}

func loadEnv() {
	APIEndpoint = getEnvOrElse(ENV_API_ENDDPOINT, APIEndpoint)
	APIKey = getEnvOrElse(ENV_TOKEN, APIKey)
	ProjectSlug = getEnvOrElse(ENV_PROJECT_SLUG, ProjectSlug)
	ProjectBranch = getEnvOrElse(ENV_PROJECT_BRANCH, ProjectBranch)
	if ip := os.Getenv(ENV_IGNORED_PATHS); ip != "" {
		IgnoredPaths = strings.Split(ip, ",")
	}
	if raw := os.Getenv(ENV_RAW_FORMAT); raw != "" {
		RawFormat = true
	}
}
