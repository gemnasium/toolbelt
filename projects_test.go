package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
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
	r := strings.NewReader("Project description\n")
	err := CreateProject("test_project", config, r)
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
	r := strings.NewReader("Project description\n")
	err := CreateProject("test_project", config, r)
	if err.Error() != "Server returned non-200 status: 401 Unauthorized\n" {
		t.Error(err)
	}
}

func TestConfigureProject(t *testing.T) {

	config, _ := NewConfig([]byte{})

	// set first param
	tmp, err := ioutil.TempFile("", "gemnasium")
	if err != nil {
		t.Error(err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())
	err = ConfigureProject("my_slug", config, os.Stdin, tmp)
	if err != nil {
		t.Error(err)
	}
}

func TestSyncProject(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		jsonOutput := ""
		fmt.Fprintln(w, jsonOutput)

	}))
	defer ts.Close()
	config := &Config{APIEndpoint: ts.URL}
	// silent stdout
	old := os.Stdout
	_, w, _ := os.Pipe()
	w.Close()
	os.Stdout = w
	err := SyncProject("blah", config)
	os.Stdout = old
	if err != nil {
		t.Errorf("SyncProject failed with err: %s", err)
	}

}
