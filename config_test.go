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
project_branch: master        # /!\ If you don't use git, remove this line
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
	if config.Api_key != "5590c4910af0ee9428a1447f6ef8090a" {
		t.Errorf("config.api_key should be '5590c4910af0ee9428a1447f6ef8090a', was %s", config.Api_key)
	}
	if config.Site != "gemnasium.com" {
		t.Errorf("config.Site should be 'gemnasium.com', was %s", config.Site)
	}
	if config.Use_ssl != true {
		t.Errorf("config.Use_ssl should be true, was %s", config.Use_ssl)
	}
	if config.Api_version != "v3" {
		t.Errorf("config.Api_version should be true, was %s", config.Api_version)
	}
	if config.Project_branch != "master" {
		t.Errorf("config.Project_branch should be 'master', was %s", config.Project_branch)
	}
	ignored_paths := []string{"features/", "spec/", "test/", "tmp/", "vendor/"}
	if !reflect.DeepEqual(config.Ignored_paths, ignored_paths) {
		t.Errorf("config.Ignored_path doesn't match. Expected: %v, got %v", ignored_paths, config.Ignored_paths)
	}

}
