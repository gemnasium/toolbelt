package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/codegangsta/cli"
)

func CreateProjectTestServer(t *testing.T, APIKey string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var siteAuth = "Basic " + base64.StdEncoding.EncodeToString([]byte("x:"+APIKey))
		auth := r.Header.Get("Authorization")
		if siteAuth != auth {
			w.WriteHeader(http.StatusUnauthorized)
		}

		// Check URI
		if r.RequestURI != CREATE_PROJECT_PATH {
			t.Errorf("Expected RequestURI to be %s, got: %s", CREATE_PROJECT_PATH, r.RequestURI)
		}

		// Check request is a POST
		if r.Method != "POST" {
			t.Errorf("Expected a POST, got a %s", r.Method)
		}

		decoder := json.NewDecoder(r.Body)
		var project struct {
			Name, Description string
		}
		err := decoder.Decode(&project)
		if err != nil {
			t.Error(err)
		}

		w.Header().Set("Content-Type", "application/json")
		if project.Name != "" && project.Description != "" {
			w.Write([]byte(`{"name": "my_project", "slug": "my_project_slug", "remaining_slot_count": 1 }`))
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	return ts
}

func TestCreateProject(t *testing.T) {

	apiKey := "abcxyz123"

	// Fake gemnasium api
	ts := CreateProjectTestServer(t, apiKey)
	defer ts.Close()

	config := &Config{APIEndpoint: ts.URL, APIKey: apiKey}
	set := flag.NewFlagSet("test", 0)
	set.Parse([]string{"my_project"})
	ctx := cli.NewContext(nil, set, set)
	r := strings.NewReader("Project description\n")
	err := CreateProject(ctx, config, r)
	if err != nil {
		t.Error(err)
	}
}

func TestCreateProjectWithWrongToken(t *testing.T) {

	apiKey := "abcxyz123"

	// Fake gemnasium api
	ts := CreateProjectTestServer(t, apiKey)
	defer ts.Close()

	config := &Config{APIEndpoint: ts.URL, APIKey: "invalid_key"}
	set := flag.NewFlagSet("test", 0)
	set.Parse([]string{"my_project"})
	ctx := cli.NewContext(nil, set, set)
	r := strings.NewReader("Project description\n")
	err := CreateProject(ctx, config, r)
	if err.Error() != "Server returned non-200 status: 401 Unauthorized\n" {
		t.Error(err)
	}
}

func TestCreateProjectWithoutToken(t *testing.T) {
	config, _ := NewConfig([]byte{})
	set := flag.NewFlagSet("test", 0)
	set.Parse([]string{"my_project"})
	ctx := cli.NewContext(nil, set, set)
	r := strings.NewReader("Project description\n")
	err := CreateProject(ctx, config, r)
	if err != ErrEmptyToken {
		t.Error(err)
	}
}

func TestConfigureProject(t *testing.T) {

	config, _ := NewConfig([]byte{})

	// set first param
	set := flag.NewFlagSet("test", 0)
	set.Parse([]string{"project_slug"})
	ctx := cli.NewContext(nil, set, set)

	tmp, err := ioutil.TempFile("", "gemnasium")
	if err != nil {
		t.Error(err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())
	err = ConfigureProject(ctx, config, os.Stdin, tmp)
	if err != nil {
		t.Error(err)
	}
}
