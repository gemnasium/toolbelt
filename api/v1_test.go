package api

import (
	"testing"
	"strings"
	"io"
	"github.com/gemnasium/toolbelt/config"
	"os"
	"io/ioutil"
)

func TestNewAPIRequest(t *testing.T) {
	r := strings.NewReader("testing string")
	var tt = []struct {
		Method   string
		UrlStr   string
		Body     io.Reader
		Revision string
		Branch   string
	}{
		{"GET", "http://localhost/api", nil, "abcdef123456", "master"},
		{"POST", "https://localhost/api", r, "abcdef123456", "develop"},
	}
	apiV1 := NewAPIv1(config.DEFAULT_API_ENDPOINT, "secretkey")
	for _, testReq := range tt {
		os.Setenv(config.ENV_REVISION, testReq.Revision)
		os.Setenv(config.ENV_BRANCH, testReq.Branch)
		req, err := apiV1.NewAPIRequest(testReq.Method, testReq.UrlStr, testReq.Body)
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
		if req.Header["X-Gms-Client-Version"][0] != config.VERSION {
			t.Errorf("X-Gms-Client-Version should be \"%s\", got: %s", config.VERSION, req.Header["X-Gms-Client-Version"])
		}
		if req.Header["X-Gms-Revision"][0] != testReq.Revision {
			t.Errorf("X-Gms-Revision should be \"%s\", got: %s", testReq.Revision, req.Header["X-Gms-Revision"])
		}
		if req.Header["X-Gms-Branch"][0] != testReq.Branch {
			t.Errorf("X-Gms-Branch should be \"%s\", got: %s", testReq.Branch, req.Header["X-Gms-Branch"])
		}
	}
}