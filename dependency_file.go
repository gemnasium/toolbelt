package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
)

type DependencyFile struct {
	Path    string `json:"path"`
	SHA     string `json:"sha,omitempty"`
	Content []byte `json:"content"`
}

func NewDependencyFile(filePath string) *DependencyFile {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil
	}
	sha, err := GetFileSHA1(filePath)
	if err != nil {
		return nil
	}
	return &DependencyFile{Path: filePath, SHA: sha, Content: content}

}

func (df *DependencyFile) CheckFileSHA1() error {
	sum, err := GetFileSHA1(df.Path)
	if err != nil {
		return err
	}

	if sum != df.SHA {
		return fmt.Errorf("%s: File signature doesn't match (expected: %s, got: %s)", df.Path, df.SHA, sum)
	}
	return nil
}

func (df *DependencyFile) UpdateSHA() error {
	sha, err := GetFileSHA1(df.Path)
	if err != nil {
		return err
	}
	df.SHA = sha
	return nil
}

func (df *DependencyFile) Update() error {
	content, err := ioutil.ReadFile(df.Path)
	if err != nil {
		return err
	}
	df.Content = content
	err = df.UpdateSHA()
	if err != nil {
		return err
	}

	return nil
}

// Apply patch to the file referenced by Path
// If Content is empty, the file content is read from the file directly
func (df *DependencyFile) Patch(patch string) error {
	patchPath, err := exec.LookPath("patch")
	if err != nil {
		return err
	}

	cmd := exec.Command(patchPath, df.Path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}

	_, err = io.WriteString(stdin, patch)
	if err != nil {
		return err
	}
	stdin.Close()

	out, err := ioutil.ReadAll(stdout)
	if err != nil {
		return err
	}
	if err = cmd.Wait(); err != nil {
		fmt.Println(string(out))
		return err
	}

	err = df.Update()
	if err != nil {
		return err
	}
	return nil
}

// Return git SHA1 of the given file
// TODO: Make this generic (ie: working with SVN)
func GetFileSHA1(filePath string) (string, error) {
	dat, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	h := sha1.New()
	header := fmt.Sprintf("blob %d\x00", len(dat))
	io.WriteString(h, header)
	io.Copy(h, bytes.NewReader(dat))
	hash := h.Sum(nil)

	return fmt.Sprintf("%x", hash), nil
}
