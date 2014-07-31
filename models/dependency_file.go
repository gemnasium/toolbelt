package models

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gemnasium/toolbelt/config"
	"github.com/gemnasium/toolbelt/gemnasium"
	"github.com/olekukonko/tablewriter"
)

const (
	SUPPORTED_DEPENDENCY_FILES = `(Gemfile|Gemfile\.lock|.*\.gemspec|package\.json|npm-shrinkwrap\.json|setup\.py|requirements\.txt|requires\.txt|composer\.json|composer\.lock|bower\.json)$`
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

func ListDependencyFiles(project *Project) error {

	dfiles, err := project.DependencyFiles()
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Path", "SHA"})

	for _, df := range dfiles {
		table.Append([]string{df.Path, df.SHA})
	}
	table.Render() // Send output

	return nil
}

var getLocalDependencyFiles = func() ([]*DependencyFile, error) {
	dfiles := []*DependencyFile{}
	searchDeps := func(path string, info os.FileInfo, err error) error {

		// Skip excluded paths
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		// Skip ignored_pathes
		if len(config.IgnoredPaths) > 0 {
			if matched, _ := regexp.MatchString(os.Getenv(config.ENV_IGNORED_PATHS), info.Name()); matched {
				return filepath.SkipDir
			}
		}

		matched, err := regexp.MatchString(SUPPORTED_DEPENDENCY_FILES, info.Name())
		if err != nil {
			return err
		}

		if matched {
			fmt.Printf("Found: %s\n", path)
			dfiles = append(dfiles, NewDependencyFile(path))
		}
		return nil
	}
	filepath.Walk(".", searchDeps)
	return dfiles, nil
}

// Push project dependencies
// The current path will be scanned for supported dependency files (SUPPORTED_DEPENDENCY_FILES)
func PushDependencyFiles(projectSlug string, files []string) error {
	dfiles := LookupDependencyFiles(files)
	var jsonResp map[string][]DependencyFile

	opts := &gemnasium.APIRequestOptions{
		Method: "POST",
		URI:    fmt.Sprintf("/projects/%s/dependency_files", projectSlug),
		Body:   dfiles,
		Result: &jsonResp,
	}
	err := gemnasium.APIRequest(opts)
	if err != nil {
		return err
	}

	added := []string{}
	for _, df := range jsonResp["added"] {
		added = append(added, df.Path)
	}
	updated := []string{}
	for _, df := range jsonResp["updated"] {
		updated = append(updated, df.Path)
	}
	unchanged := []string{}
	for _, df := range jsonResp["unchanged"] {
		unchanged = append(unchanged, df.Path)
	}
	unsupported := []string{}
	for _, df := range jsonResp["unsupported"] {
		unsupported = append(unsupported, df.Path)
	}
	fmt.Printf("Added: %s\n", strings.Join(added, ", "))
	fmt.Printf("Updated: %s\n", strings.Join(updated, ", "))
	fmt.Printf("Unchanged: %s\n", strings.Join(unchanged, ", "))
	fmt.Printf("Unsupported: %s\n", strings.Join(unsupported, ", "))
	return nil
}

// Load dependency files if files is not empty, otherwise search in the current
// path for files
func LookupDependencyFiles(files []string) []*DependencyFile {
	var dfiles = []*DependencyFile{}

	if len(files) > 0 {
		for _, path := range files {
			dfiles = append(dfiles, NewDependencyFile(path))
		}

	} else {
		files, err := getLocalDependencyFiles()
		if err != nil {
			fmt.Println(err)
		}
		dfiles = files
	}
	return dfiles
}
