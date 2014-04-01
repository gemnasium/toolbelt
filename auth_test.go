package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/codegangsta/cli"
)

func TestLogin(t *testing.T) {
	// Fake gemnasium api
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check URI
		if r.RequestURI != LOGIN_PATH {
			t.Errorf("Expected RequestURI to be %s, got: %s", LOGIN_PATH, r.RequestURI)
		}

		// Check request is a POST
		if r.Method != "POST" {
			t.Errorf("Expected a POST, got a %s", r.Method)
		}

		decoder := json.NewDecoder(r.Body)
		var credentials struct {
			Email, Password string
		}
		err := decoder.Decode(&credentials)
		if err != nil {
			t.Error(err)
		}

		w.Header().Set("Content-Type", "application/json")
		if credentials.Email != "" && credentials.Password != "" {
			w.Write([]byte(`{"api_token": "abcxyz123"}`))
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer ts.Close()

	config := &Config{APIEndpoint: ts.URL}
	set := flag.NewFlagSet("test", 0)
	ctx := cli.NewContext(nil, set, set)

	// don't try to use stdin
	getCredentials = func() (email, password string, err error) {
		return "batman@example.com", "secret123", nil
	}

	netrcFile := bytes.NewBufferString("")

	loadNetrc = func() *netrc.Netrc {
		nrc, _ := netrc.Parse(netrcFile)
		return nrc

	}
	// don't actually write to the file
	writeNetrcFile = func(body []byte) error {
		_, err := netrcFile.Write(body)
		return err
	}
	err := Login(ctx, config)
	if err != nil {
		t.Error(err)
	}

	expectedNetrcFile := `
machine 127.0.0.1
	login batman@example.com
	password abcxyz123`
	if netrcFile.String() != expectedNetrcFile {
		t.Errorf("Expected netrcFile to contain 127.0.0.1, got: %s\n", netrcFile.String())
	}
}

func TestLogout(t *testing.T) {
	netrcFile := bytes.NewBufferString(`
	machine github.com
	  login robin@example.com
	  password secretpw
	machine 127.0.0.1
	  login batman@example.com
	  password abcxyz123
	`)

	config := &Config{APIEndpoint: "http://127.0.0.1/api"}
	loadNetrc = func() *netrc.Netrc {
		nrc, _ := netrc.Parse(netrcFile)
		return nrc
	}
	set := flag.NewFlagSet("test", 0)
	ctx := cli.NewContext(nil, set, set)

	writeNetrcFile = func(body []byte) error {
		_, err := netrcFile.Write(body)
		return err
	}
	err := Logout(ctx, config)
	if err != nil {
		t.Error(err)
	}
	expectedNetrcFile := `
	machine github.com
	  login robin@example.com
	  password secretpw
	`
	if netrcFile.String() != expectedNetrcFile {
		t.Error("Expected netrcFile to contain github login")
	}
}
