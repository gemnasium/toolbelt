package dependency

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gemnasium/toolbelt/api"
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
            "name": "activerecord",
            "slug": "gems/activerecord",
            "type": "rubygem"
        },
        "requirement": "=3.1.12",
        "locked": "3.1.12",
        "type": "development",
        "first_level": false,
        "color": "red",
		"advisories": [ { "id": 1 }, { "id": 2 }]
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
	api.APIImpl = api.NewAPIv1(ts.URL, "")
	ListDependencies(&api.Project{Slug: "blah"})
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = old // restoring the real stdout

	expectedOutput := "+------------------+--------------+--------+--------+------------+\n"
	expectedOutput += "|   DEPENDENCIES   | REQUIREMENTS | LOCKED | STATUS | ADVISORIES |\n"
	expectedOutput += "+------------------+--------------+--------+--------+------------+\n"
	expectedOutput += "| gemnasium-gem    | >=1.0.0      | 2.0.0  | green  |            |\n"
	expectedOutput += "| +-- activerecord | =3.1.12      | 3.1.12 | red    | 1, 2       |\n"
	expectedOutput += "| rails            | =3.1.12      | 3.1.12 | red    |            |\n"
	expectedOutput += "+------------------+--------------+--------+--------+------------+\n"

	if buf.String() != expectedOutput {
		t.Errorf("Expected ouput:\n%s\n\nGot:\n%s", expectedOutput, buf.String())
	}

}
