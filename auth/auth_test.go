package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/gemnasium/toolbelt/api"
)

func TestLoadNetRCFileWithNonExistingFile(t *testing.T) {
	f := filepath.Join(os.TempDir(), "netrctmpfile")
	os.Setenv("NETRC_PATH", f)
	if _, err := os.Stat(netrcPath()); err == nil {
		fmt.Printf("%s temp file was not supposed to exist, deleting\n", f)
		os.Remove(f)
	}
	netrc := loadNetrc()
	os.Remove(f)
	if netrc == nil {
		t.Error("loadNetrc should not return nil")
	}
}

func TestLoadNetRCFileWithNonExistingAndInvalidFile(t *testing.T) {
	f := "/nonexistingpath/subpath/file"
	os.Setenv("NETRC_PATH", f)
	netrc := loadNetrc()
	if netrc == nil {
		t.Error("loadNetrc should not fail")
	}
}

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

		var credentials struct {
			Email, Password string
		}
		err := json.NewDecoder(r.Body).Decode(&credentials)
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

	api.APIImpl = api.NewAPIv1(ts.URL, "")
	// don't try to use stdin
	getEmail = func() (email string) {
		return "batman@example.com"
	}
	getPassword = func(prompt string) (password string, err error) {
		return "secret123", nil
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
	err := Login()
	if err != nil {
		t.Error(err)
	}

	expectedNetrcFile :=
		`machine 127.0.0.1
	login batman@example.com
	password abcxyz123`
	if netrcFile.String() != expectedNetrcFile {
		t.Errorf("Expected netrcFile to be:\n%#v\ngot:\n%#v\n", expectedNetrcFile, netrcFile.String())
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

	api.APIImpl = api.NewAPIv1("http://127.0.0.1/api", "")
	loadNetrc = func() *netrc.Netrc {
		nrc, _ := netrc.Parse(netrcFile)
		return nrc
	}
	writeNetrcFile = func(body []byte) error {
		_, err := netrcFile.Write(body)
		return err
	}
	err := Logout()
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
