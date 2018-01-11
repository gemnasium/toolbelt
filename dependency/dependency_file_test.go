package dependency

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gemnasium/toolbelt/api"
	"path/filepath"
	"reflect"
	"encoding/json"
	"github.com/gemnasium/toolbelt/config"
)

type TestFile struct {
	Content []byte
	SHA     string
}

func testFile() TestFile {
	// echo "source 'https://rubygems.org'\n\ngem 'rails', '3.2.18'" | git hash-object --stdin
	// => 752751b8bf149dfea0d8b8b45cec23e6ac30b4a1
	return TestFile{
		Content: []byte("source 'https://rubygems.org'\n\ngem 'rails', '3.2.18'\n"),
		SHA:     "752751b8bf149dfea0d8b8b45cec23e6ac30b4a1",
	}
}

func TestNewDependencyFile(t *testing.T) {
	tmp, err := ioutil.TempFile("", "gemnasium-df")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	df := NewDependencyFile(tmp.Name())
	if df == nil {
		t.Errorf("NewDependencyFile returned nil")
	}
}

func TestCheckFileSHA1(t *testing.T) {
	tf := testFile()
	tmp, err := ioutil.TempFile("", "gemnasium-df")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	_, err = tmp.Write(tf.Content)
	if err != nil {
		t.Error(err)
	}

	df := NewDependencyFile(tmp.Name())

	if df.SHA != tf.SHA {
		t.Errorf("DependencyFile has an invalid SHA (Exp: '%s', Got: '%s')\n", tf.SHA, df.SHA)
	}

	if err := DependencyFileCheckFileSHA1(df); err != nil {
		t.Fatal(err)
	}
}

func TestPatch(t *testing.T) {
	tf := testFile()
	tmp, err := ioutil.TempFile("", "gemnasium-df")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	_, err = tmp.Write(tf.Content)
	if err != nil {
		t.Error(err)
	}
	df := NewDependencyFile(tmp.Name())

	patch := "--- %s\n+++ titi\n@@ -1,3 +1,3 @@\n source 'https://rubygems.org'\n\n-gem 'rails', '3.2.18'\n+gem 'rails', '3.2.21'\n"
	expected := "source 'https://rubygems.org'\n\ngem 'rails', '3.2.21'\n"

	DependencyFilePatch(df, patch)

	var bytesExpected = []byte(expected)
	if !bytes.Equal(df.Content, bytesExpected) {
		t.Errorf("DependencyFile content is incorrect (Exp: '%s', Got: '%s')\n", expected, df.Content)
	}
	if df.SHA == tf.SHA {
		t.Error("DependencyFile SHA is incorrect (should have changed)")
	}
}

func TestUpdate(t *testing.T) {
	tf := testFile()
	tmp, err := ioutil.TempFile("", "gemnasium-df")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	_, err = tmp.Write(tf.Content)
	if err != nil {
		t.Error(err)
	}
	df := NewDependencyFile(tmp.Name())

	newContent := "source 'https://rubygems.org'\n\ngem 'rails', '4.1.1'\n"
	_, err = tmp.WriteAt([]byte(newContent), 0)
	if err != nil {
		t.Error(err)
	}

	DependencyFileUpdate(df)

	if bytes.Equal(df.Content, []byte(newContent)) {
		t.Errorf("DependencyFile content is incorrect (Exp: '%s', Got: '%s')\n", newContent, df.Content)
	}
	sha := "956b10afe3cf5511c4aa42ee93f208d8cd707ce3"
	if df.SHA == sha {
		t.Errorf("DependencyFile SHA is incorrect (Exp: %s, Got: %s)", df.SHA, sha)
	}
}

func TestGetFileSHA1(t *testing.T) {
	tf := testFile()
	tmp, err := ioutil.TempFile("", "gemnasium-df")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	_, err = tmp.Write(tf.Content)
	if err != nil {
		t.Error(err)
	}

	sha, err := GetFileSHA1(tmp.Name())
	if err != nil {
		t.Error(err)
	}

	if tf.SHA != sha {
		t.Errorf("GetFileSHA1 error. Exp: %s, Got: %s", tf.SHA, sha)
	}
}

func TestGetLocalDependencyFiles(t *testing.T) {
	wantedResult := []*api.DependencyFile{
		&api.DependencyFile{
			Path: "Gemfile",
			SHA: "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
			Content: []uint8{},
		},
		&api.DependencyFile{
			Path: "subdir/gems.rb",
			SHA: "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
			Content: []uint8{},
		},
	}
	var prettyString = func(v interface{}) string {
		b, _ := json.MarshalIndent(v, "", "  ")
		return string(b)
	}
	config.IgnoredPaths = []string{"sub1/sub2/Gemfile", "sub3/sub4"}
	// Get a list of recognised dependency files from test data
	result, err := getLocalDependencyFiles(filepath.Join("testdata", "test_get_local_dependency_files"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result, wantedResult) {
		t.Errorf("Expected output:\n%s\nGot:\n%s\n", prettyString(wantedResult), prettyString(result))
	}
}

func TestListDependencyFiles(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		jsonOutput :=
			fmt.Sprintf(`[
			{ "id": "1", "path": "Gemfile", "content": "%s", "sha": "Gemfile SHA-1" },
			{ "id": "2", "path": "Gemfile.lock", "content": "%s", "sha": "Gemfile.lock SHA-1" }
			]`,
				base64.StdEncoding.EncodeToString([]byte("Gemfile base64 encoded content")),
				base64.StdEncoding.EncodeToString([]byte("Gemfile.lock base64 encoded content")))
		fmt.Fprintln(w, jsonOutput)
	}))
	defer ts.Close()
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	api.APIImpl = api.NewAPIv1(ts.URL, "")
	err := ListDependencyFiles(&api.Project{Slug: "blah"})
	if err != nil {
		t.Error(err)
	}

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = old // restoring the real stdout

	expectedOutput := "+--------------+--------------------+\n"
	expectedOutput += "|     PATH     |        SHA         |\n"
	expectedOutput += "+--------------+--------------------+\n"
	expectedOutput += "| Gemfile      | Gemfile SHA-1      |\n"
	expectedOutput += "| Gemfile.lock | Gemfile.lock SHA-1 |\n"
	expectedOutput += "+--------------+--------------------+\n"
	if buf.String() != expectedOutput {
		t.Errorf("Expected ouput:\n%s\n\nGot:\n%s", expectedOutput, buf.String())
	}
}

func TestPushDependencyFiles(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		jsonOutput :=
			fmt.Sprintf(`{
					"added": [{ "id": "1", "path": "Gemfile", "content": "%s", "sha": "Gemfile SHA-1"},
						{ "id": "2", "path": "Gemfile.lock", "content": "%s", "sha": "Gemfile.lock SHA-1"}],
					"updated": [],
					"unchanged": [{ "id": "3", "path": "js/package.json", "content": "%s", "sha": "package.sjon SHA-1"}],
					"unsupported": []
			}`,
				base64.StdEncoding.EncodeToString([]byte("Gemfile base64 encoded content")),
				base64.StdEncoding.EncodeToString([]byte("Gemfile.lock base64 encoded content")),
				base64.StdEncoding.EncodeToString([]byte("package.json content")))
		fmt.Fprintln(w, jsonOutput)
	}))
	defer ts.Close()
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	api.APIImpl = api.NewAPIv1(ts.URL, "")

	getLocalDependencyFiles = func(path string) ([]*api.DependencyFile, error) {
		return []*api.DependencyFile{
			&api.DependencyFile{Path: "Gemfile", SHA: "Gemfile SHA-1", Content: []byte("Gemfile.lock base64 encoded content")},
			&api.DependencyFile{Path: "Gemfile.lock", SHA: "Gemfile.lock SHA-1", Content: []byte("Gemfile base64 encoded content")},
			&api.DependencyFile{Path: "js/package.json", SHA: "package.json SHA-1", Content: []byte("package.json content")},
		}, nil
	}

	err := PushDependencyFiles("blah", []string{})
	if err != nil {
		t.Error(err)
	}

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = old // restoring the real stdout

	expectedOutput := "[warning] No files given, scanning current directory instead.\n"
	expectedOutput += "Sending files to Gemnasium: done.\n"
	expectedOutput += "\n"
	expectedOutput += "Added: Gemfile, Gemfile.lock\n"
	expectedOutput += "Updated: \n"
	expectedOutput += "Unchanged: js/package.json\n"
	expectedOutput += "Unsupported: \n"
	if buf.String() != expectedOutput {
		t.Errorf("Expected ouput:\n%s\n\nGot:\n%s", expectedOutput, buf.String())
	}
}
