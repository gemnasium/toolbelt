package utils

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/gemnasium/toolbelt/config"
)

func TestNewAPIRequest(t *testing.T) {
	r := strings.NewReader("testing string")
	var tt = []struct {
		Method   string
		UrlStr   string
		APIKey   string
		Body     io.Reader
		Revision string
		Branch   string
	}{
		{"GET", "http://localhost/api", "secretkey", nil, "abcdef123456", "master"},
		{"POST", "https://localhost/api", "secretkey", r, "abcdef123456", "develop"},
	}
	for _, testReq := range tt {
		os.Setenv(config.ENV_REVISION, testReq.Revision)
		os.Setenv(config.ENV_BRANCH, testReq.Branch)
		req, err := NewAPIRequest(testReq.Method, testReq.UrlStr, testReq.APIKey, testReq.Body)
		if err != nil {
			t.Error(err)
		}
		if len(req.Header["Content-Type"]) > 1 {
			t.Error("Content-Type defined more than once")
		}
		if req.Header["Content-Type"][0] != "application/json" {
			t.Errorf("Content-Type should be \"application/json\", got: %s", req.Header["Content-Type"])
		}
		if req.Header["Authorization"][0] != "Basic eDpzZWNyZXRrZXk=" {
			t.Errorf("Authorization should be \"eDpzZWNyZXRrZXk=\", got: %s", req.Header["Authorization"])
		}
		if req.Method == "POST" {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err)
			}
			req.Body.Close()

			if string(body) != "testing string" {
				t.Errorf("Body should be \"testing string\", got: %s", body)
			}
		}
		if req.Header["User-Agent"][0] != "Gemnasium Toolbelt "+config.VERSION {
			t.Errorf("User-Agent should be \"%s\", got: %s", config.VERSION, req.Header["User-Agent"])
		}
		if req.Header["X-Gms-Revision"][0] != testReq.Revision {
			t.Errorf("X-Gms-Revision should be \"%s\", got: %s", testReq.Revision, req.Header["X-Gms-Revision"])
		}
		if req.Header["X-Gms-Branch"][0] != testReq.Branch {
			t.Errorf("X-Gms-Branch should be \"%s\", got: %s", testReq.Branch, req.Header["X-Gms-Branch"])
		}
	}
}

func testStatusDot(t *testing.T) {
	var tt = []struct {
		Color    string
		Expected string
	}{
		{"red", "@k\u2B24 @k\u2B24 @r\u2B24  @{|}(red)"},
		{"yellow", "@k\u2B24 @y\u2B24 @k\u2B24  @{|}(yellow)"},
		{"green", "@g\u2B24 @k\u2B24 @k\u2B24  @{|}(green)"},
		{"purple", "@k\u2B24 @k\u2B24 @k\u2B24  @{|}(none)"},
	}
	for _, test := range tt {
		dots := StatusDots(test.Color)
		if dots != test.Expected {
			t.Errorf("%s expected, got: %s", test.Expected, dots)
		}
	}

}
