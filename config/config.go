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
	APIKey       string
	APIVersion   int = 1
	ProjectSlug  string
	IgnoredPaths []string
	RawFormat    bool
)

const (
	VERSION          = "1.0.3"
	CONFIG_FILE_PATH = ".gemnasium.yml"

	// Don't forget to update DisplayEnvVars func bellow when updating vars
	ENV_API_ENDDPOINT                = "GEMNASIUM_API_ENDPOINT"
	ENV_TOKEN                        = "GEMNASIUM_TOKEN"
	ENV_PROJECT_SLUG                 = "GEMNASIUM_PROJECT_SLUG"
	ENV_BRANCH                       = "BRANCH"
	ENV_REVISION                     = "REVISION"
	ENV_IGNORED_PATHS                = "GEMNASIUM_IGNORED_PATHS"
	ENV_RAW_FORMAT                   = "GEMNASIUM_RAW_FORMAT"
	ENV_GEMNASIUM_TESTSUITE          = "GEMNASIUM_TESTSUITE"
	ENV_GEMNASIUM_BUNDLE_INSTALL_CMD = "GEMNASIUM_BUNDLE_INSTALL_CMD"
	ENV_GEMNASIUM_BUNDLE_UPDATE_CMD  = "GEMNASIUM_BUNDLE_UPDATE_CMD"

	DEFAULT_API_ENDPOINT = "https://api.gemnasium.com/v1"
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
	if ip := os.Getenv(ENV_IGNORED_PATHS); ip != "" {
		IgnoredPaths = strings.Split(ip, ",")
	}
	if raw := os.Getenv(ENV_RAW_FORMAT); raw != "" {
		RawFormat = true
	}
}

func DisplayEnvVars() {
	vars := map[string]string{
		ENV_API_ENDDPOINT:                "API URL (only used for debugging).",
		ENV_TOKEN:                        "Your private API token.",
		ENV_PROJECT_SLUG:                 "The project slug (unique identifier). Use `gemnasium projects list`, or the project settings page to get it.",
		ENV_BRANCH:                       "Current branch.",
		ENV_REVISION:                     "Current revision.",
		ENV_IGNORED_PATHS:                "When using the 'eval' or 'df push' commands, if --files is empty, gemnasium will look for files locally. Paths to be ignored can be set with this var, separated with a comma.",
		ENV_RAW_FORMAT:                   "Display raw json response from API server.",
		ENV_GEMNASIUM_TESTSUITE:          "Used for auto-update command, to set the testsuite to run.",
		ENV_GEMNASIUM_BUNDLE_INSTALL_CMD: "[auto-update] Override command used with ruby sets. default: 'bundle install'",
		ENV_GEMNASIUM_BUNDLE_UPDATE_CMD:  "[auto-update] Override command used with ruby sets. default: 'bundle update'",
	}
	for k, _ := range vars {
		fmt.Printf("%s=%s\n", k, os.Getenv(k))
	}
}
