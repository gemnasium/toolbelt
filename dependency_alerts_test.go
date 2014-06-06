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

func TestListDependencyAlerts(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		jsonOutput :=
			`[
    {
        "id": 1,
        "advisory": {
            "id": 1,
            "title": "XSS vulnerability"
        },
        "open_at": "2014-05-07T09:59:53.738404Z",
        "status": "acknowledged"
    },
    {
        "id": 2,
        "advisory": {
            "id": 2,
            "title": "DOS vulnerability"
        },
        "open_at": "2014-05-07T09:59:53.738404Z",
        "status": "closed"
    }
]`
		fmt.Fprintln(w, jsonOutput)
	}))
	defer ts.Close()
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	config := &Config{APIEndpoint: ts.URL}
	ListDependencyAlerts("blah", config)
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = old // restoring the real stdout

	expectedOutput :=
		`+----------+---------------------+--------------+
| ADVISORY |        DATE         |    STATUS    |
+----------+---------------------+--------------+
| 1        | 07 May 14 09:59 UTC | acknowledged |
| 2        | 07 May 14 09:59 UTC | closed       |
+----------+---------------------+--------------+
`
	if buf.String() != expectedOutput {
		t.Errorf("Expected ouput:\n%s\n\nGot:\n%s", expectedOutput, buf.String())
	}

}
