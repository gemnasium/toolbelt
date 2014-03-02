package main

import (
	"reflect"
	"testing"
)

func TestNewConfig(t *testing.T) {
	config_data := []byte(`
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
	config, err := NewConfig(config_data)
	if err != nil {
		t.Fatal(err)
	}
	if config.APIKey != "5590c4910af0ee9428a1447f6ef8090a" {
		t.Errorf("config.APIKey should be '5590c4910af0ee9428a1447f6ef8090a', was %s", config.APIKey)
	}
	if config.APIEndpoint != "https://gemnasium.com/api/v3" {
		t.Errorf("config.APIEndpoint should be true, was %s", config.APIEndpoint)
	}
	if config.ProjectBranch != "develop" {
		t.Errorf("config.ProjectBranch should be 'develop', was %s", config.ProjectBranch)
	}
	ignored_paths := []string{"features/", "spec/", "test/", "tmp/", "vendor/"}
	if !reflect.DeepEqual(config.IgnoredPaths, ignored_paths) {
		t.Errorf("config.IgnoredPaths doesn't match. Expected: %v, got %v", ignored_paths, config.IgnoredPaths)
	}

}
