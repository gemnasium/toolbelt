package models

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gemnasium/toolbelt/config"
	"github.com/wsxiaoys/terminal/color"
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

	config.APIEndpoint = ts.URL
	config.APIKey = apiKey
	r := strings.NewReader("Project description\n")
	err := CreateProject("test_project", r)
	if err != nil {
		t.Error(err)
	}
}

func TestCreateProjectWithWrongToken(t *testing.T) {
	apiKey := "abcxyz123"

	// Fake gemnasium api
	ts := CreateProjectTestServer(t, apiKey)
	defer ts.Close()

	config.APIEndpoint = ts.URL
	config.APIKey = "invalid_key"
	r := strings.NewReader("Project description\n")
	err := CreateProject("test_project", r)
	if err.Error() != "Server returned non-200 status: 401 Unauthorized\n" {
		t.Error(err)
	}
}

func TestConfigureProject(t *testing.T) {
	// set first param
	tmp, err := ioutil.TempFile("", "gemnasium")
	if err != nil {
		t.Error(err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())
	p := &Project{Slug: "blah"}
	err = p.Configure("my_slug", os.Stdin, tmp)
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
	config.APIEndpoint = ts.URL
	// silent stdout
	old := os.Stdout
	_, w, _ := os.Pipe()
	w.Close()
	os.Stdout = w
	p := &Project{Slug: "blah"}
	err := p.Sync()
	os.Stdout = old
	if err != nil {
		t.Errorf("SyncProject failed with err: %s", err)
	}
}

func TestUpdateProject(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		jsonOutput :=
			`{
    "slug": "gemnasium/API_project",
    "name: "API_project",
    "description": "This is a brief description of a project on Gemnasium"
    "origin": "github",
    "private": false,
    "color": "green",
    "monitored": true,
    "unmonitored_reason": ""
}`
		fmt.Fprintln(w, jsonOutput)
	}))
	defer ts.Close()
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	config.APIEndpoint = ts.URL
	var name, desc *string
	var monitored *bool
	nameStr := "API_project"
	name = &nameStr
	descStr := "A desc"
	desc = &descStr
	monitoredBool := false
	monitored = &monitoredBool
	p := &Project{Slug: "blah"}
	err := p.Update(name, desc, monitored)
	if err != nil {
		t.Fatal(err)
	}
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = old // restoring the real stdout

	expectedOutput := color.Sprintf("@gProject %s updated succesfully\n", "blah")
	if buf.String() != expectedOutput {
		t.Errorf("Expected ouput:\n%s\n\nGot:\n%s", expectedOutput, buf.String())
	}
}

func TestUpdateProjectWithNoParams(t *testing.T) {
	p := &Project{Slug: "blah"}
	err := p.Update(nil, nil, nil)
	if err.Error() != "Please specify at least one thing to update (name, desc, or monitored" {
		t.Errorf("Expected error to be 'Please specify at least one thing to update (name, desc, or monitored', got %s\n", err)
	}

}
