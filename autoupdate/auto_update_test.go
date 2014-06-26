package autoupdate

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gemnasium/toolbelt/config"
	"github.com/gemnasium/toolbelt/models"
)

func TestFetchUpdateSet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		jsonOutput :=
			`{"id":1,"requirement_updates":{"Rubygem": [{"file":{"path":"Gemfile","sha":"dc6bdc865c85a4f5c6ef0f4ba8909d8652fd8cd0"},"patch":"--- Gemfile\n+++ Gemfile\n@@ -5 +5 @@\n-gem \"warden\", \"0.10.3\"\n+gem \"warden\", '~> 1.2.3'\n@@ -4 +4 @@\n-gem \"rails\", \"3.0.0.beta3\"\n+gem \"rails\", '~> 4.0.3'\n@@ -7 +7 @@\n-gem \"webrat\", \"0.7\"\n+gem \"webrat\", '~> 0.7.3'\n"}]},"version_updates":{}}`
		fmt.Fprintln(w, jsonOutput)
	}))
	defer ts.Close()
	config.APIEndpoint = ts.URL
	expectedUpdateSet := &UpdateSet{
		ID: 1,
		RequirementUpdates: map[string][]RequirementUpdate{
			"Rubygem": []RequirementUpdate{
				RequirementUpdate{
					File: models.DependencyFile{
						Path: "Gemfile",
						SHA:  "dc6bdc865c85a4f5c6ef0f4ba8909d8652fd8cd0",
					},
					Patch: "--- Gemfile\n+++ Gemfile\n@@ -5 +5 @@\n-gem \"warden\", \"0.10.3\"\n+gem \"warden\", '~> 1.2.3'\n@@ -4 +4 @@\n-gem \"rails\", \"3.0.0.beta3\"\n+gem \"rails\", '~> 4.0.3'\n@@ -7 +7 @@\n-gem \"webrat\", \"0.7\"\n+gem \"webrat\", '~> 0.7.3'\n",
				},
			},
		},
		VersionUpdates: map[string][]VersionUpdate{},
	}

	resultSet, err := fetchUpdateSet("blah")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(resultSet, expectedUpdateSet) {
		t.Errorf("Expected resultSet to be:\n%#v\nGot:\n%#v\n", expectedUpdateSet, resultSet)
	}
}
