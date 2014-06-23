package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
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
