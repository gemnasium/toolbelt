package config

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	if APIEndpoint != DEFAULT_API_ENDPOINT {
		t.Errorf("APIEndpoint is invalid. Expected: %s, Got: %s\n", DEFAULT_API_ENDPOINT, APIEndpoint)
	}
	if ProjectBranch != DEFAULT_PROJECT_BRANCH {
		t.Errorf("ProjectBranch is invalid. Expected: %s, Got: %s\n", DEFAULT_PROJECT_BRANCH, ProjectBranch)
	}
}

func TestWithConfigFile(t *testing.T) {
	configData := []byte(`
api_endpoint: "http://localhost/"
api_key: 5590c4910af0ee9428a1447f6ef8090a    # You personal (secret) API key. Get it at https://gemnasium.com/settings/api_access
project_name: project_name    # A name to remember your project.
project_slug: e22c6e1a59e77e595949c936e3e797ea               # Unique slug for this project. Get it on the "project settings" page.
project_branch: develop       # /!\ If you don't use git, remove this line
ignored_paths:                # Paths you want to ignore when searching for dependency files (from app root)
  - features/
  - spec/
  - test/
  - tmp/
  - vendor/
`)
	err := ioutil.WriteFile(CONFIG_FILE_PATH, configData, 0666)
	if err != nil {
		t.Fatal(err)
	}
	// don't forget to remove our config file when test is done
	defer func() {
		err := os.Remove(CONFIG_FILE_PATH)
		if err != nil {
			t.Fatal(err)
		}
	}()

	loadConfig()

	if APIKey != "5590c4910af0ee9428a1447f6ef8090a" {
		t.Errorf("APIKey should be '5590c4910af0ee9428a1447f6ef8090a', was %s", APIKey)
	}
	if ProjectSlug != "e22c6e1a59e77e595949c936e3e797ea" {
		t.Errorf("ProjectSlug should be 'e22c6e1a59e77e595949c936e3e797ea', was %s", ProjectSlug)
	}
	if APIEndpoint != "http://localhost/" {
		t.Errorf("APIEndpoint should be 'https://api.gemnasium.com/v1', was %s", APIEndpoint)
	}
	if ProjectBranch != "develop" {
		t.Errorf("ProjectBranch should be 'develop', was %s", ProjectBranch)
	}
	ignored_paths := []string{"features/", "spec/", "test/", "tmp/", "vendor/"}
	if !reflect.DeepEqual(IgnoredPaths, ignored_paths) {
		t.Errorf("IgnoredPaths doesn't match. Expected: %v, got %v", ignored_paths, IgnoredPaths)
	}
}

func TestWithEnvVars(t *testing.T) {
	os.Setenv(ENV_API_ENDDPOINT, "http://localhost/")
	os.Setenv(ENV_TOKEN, "new-key")
	os.Setenv(ENV_PROJECT_SLUG, "new-slug")
	os.Setenv(ENV_PROJECT_BRANCH, "develop")
	os.Setenv(ENV_IGNORED_PATHS, "/tmp,/foo,/bar")
	os.Setenv(ENV_RAW_FORMAT, "true")

	loadEnv()
	if APIKey != "new-key" {
		t.Errorf("APIKey should be 'new-key', was %s", APIKey)
	}
	if ProjectSlug != "new-slug" {
		t.Errorf("ProjectSlug should be 'new-slug', was %s", ProjectSlug)
	}
	if APIEndpoint != "http://localhost/" {
		t.Errorf("APIEndpoint should be true, was %s", APIEndpoint)
	}
	if ProjectBranch != "develop" {
		t.Errorf("ProjectBranch should be 'develop', was %s", ProjectBranch)
	}
	ignored_paths := []string{"/tmp", "/foo", "/bar"}
	if !reflect.DeepEqual(IgnoredPaths, ignored_paths) {
		t.Errorf("IgnoredPaths doesn't match. Expected: %v, got %v", ignored_paths, IgnoredPaths)
	}
}
