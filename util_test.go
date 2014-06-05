package main

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

func TestNewAPIRequest(t *testing.T) {
	r := strings.NewReader("testing string")
	var tt = []struct {
		Method string
		UrlStr string
		APIKey string
		Body   io.Reader
	}{
		{"GET", "http://localhost/api", "secretkey", nil},
		{"POST", "https://localhost/api", "secretkey", r},
	}
	for _, testReq := range tt {
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
	}
}
