package models

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

	if err := df.CheckFileSHA1(); err != nil {
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

	patch := fmt.Sprintf("--- %s\n+++ titi\n@@ -1,3 +1,3 @@\n source 'https://rubygems.org'\n\n-gem 'rails', '3.2.18'\n+gem 'rails', '3.2.21'\n")
	expected := "source 'https://rubygems.org'\n\ngem 'rails', '3.2.21\n'"

	df.Patch(patch)

	if bytes.Equal(df.Content, []byte(expected)) {
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

	df.Update()

	if bytes.Equal(df.Content, []byte(newContent)) {
		t.Errorf("DependencyFile content is incorrect (Exp: '%s', Got: '%s')\n", newContent, df.Content)
	}
	sha := "956b10afe3cf5511c4aa42ee93f208d8cd707ce3"
	if df.SHA == sha {
		t.Error("DependencyFile SHA is incorrect (Exp: %s, Got: %s)", "s", df.SHA, sha)
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
	config.APIEndpoint = ts.URL
	err := ListDependencyFiles(&Project{Slug: "blah"})
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
	config.APIEndpoint = ts.URL

	getLocalDependencyFiles = func() ([]*DependencyFile, error) {
		return []*DependencyFile{
			&DependencyFile{Path: "Gemfile", SHA: "Gemfile SHA-1", Content: []byte("Gemfile.lock base64 encoded content")},
			&DependencyFile{Path: "Gemfile.lock", SHA: "Gemfile.lock SHA-1", Content: []byte("Gemfile base64 encoded content")},
			&DependencyFile{Path: "js/package.json", SHA: "package.json SHA-1", Content: []byte("package.json content")},
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
