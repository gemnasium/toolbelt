package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestListDependencies(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		jsonOutput :=
			`[
    {
        "package": {
            "name": "gemnasium-gem",
            "slug": "gems/gemnasium-gem",
            "type": "rubygem"
        },
        "requirement": ">=1.0.0",
        "locked": "2.0.0",
        "type": "development",
        "first_level": true,
        "color": "green"
    },
    {
        "package": {
            "name": "rails",
            "slug": "gems/rails",
            "type": "rubygem"
        },
        "requirement": "=3.1.12",
        "locked": "3.1.12",
        "type": "runtime",
        "first_level": true,
        "color": "red"
    }
]`
		fmt.Fprintln(w, jsonOutput)
	}))
	defer ts.Close()
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	config := &Config{APIEndpoint: ts.URL}
	ListDependencies("blah", config)
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = old // restoring the real stdout

	expectedOutput :=
		`+---------------+--------------+--------+--------+------------+
| DEPENDENCIES  | REQUIREMENTS | LOCKED | STATUS | ADVISORIES |
+---------------+--------------+--------+--------+------------+
| gemnasium-gem | >=1.0.0      |        | green  |            |
| rails         | =3.1.12      |        | red    |            |
+---------------+--------------+--------+--------+------------+
`
	if buf.String() != expectedOutput {
		t.Errorf("Expected ouput:\n%s\n\nGot:\n%s", expectedOutput, buf.String())
	}

}
