package dependency

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
	"encoding/base64"

	"github.com/gemnasium/toolbelt/config"
	"github.com/olekukonko/tablewriter"
	"github.com/gemnasium/toolbelt/api"
	"github.com/gemnasium/toolbelt/project"
)

const (
	SUPPORTED_DEPENDENCY_FILES = `(Gemfile|Gemfile\.lock|gems\.rb|gems\.locked|.*\.gemspec|package\.json|package-lock\.json|npm-shrinkwrap\.json|yarn\.lock|setup\.py|requirements\.txt|requirements\.pip|requirements.*\.txt|requires\.txt|composer\.json|composer\.lock|bower\.json|pom\.xml)$`
)

func NewDependencyFile(filePath string) *api.DependencyFile {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil
	}
	sha, err := GetFileSHA1(filePath)
	if err != nil {
		return nil
	}
	return &api.DependencyFile{Path: filePath, SHA: sha, Content: content}
}

func DependencyFileCheckFileSHA1(df *api.DependencyFile) error {
	sum, err := GetFileSHA1(df.Path)
	if err != nil {
		return err
	}

	if sum != df.SHA {
		return fmt.Errorf("%s: File signature doesn't match (expected: %s, got: %s)", df.Path, df.SHA, sum)
	}
	return nil
}

func DependencyFileUpdateSHA(df *api.DependencyFile) error {
	sha, err := GetFileSHA1(df.Path)
	if err != nil {
		return err
	}
	df.SHA = sha
	return nil
}

func DependencyFileUpdate(df *api.DependencyFile) error {
	content, err := ioutil.ReadFile(df.Path)
	if err != nil {
		return err
	}
	df.Content = content
	err = DependencyFileUpdateSHA(df)
	if err != nil {
		return err
	}

	return nil
}

// Apply patch to the file referenced by Path
// If Content is empty, the file content is read from the file directly
func DependencyFilePatch(df *api.DependencyFile, patch string) error {
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

	err = DependencyFileUpdate(df)
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

func ListDependencyFiles(p *api.Project) error {

	dfiles, err := project.ProjectDependencyFiles(p)
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

var getLocalDependencyFiles = func() ([]*api.DependencyFile, error) {
	dfiles := []*api.DependencyFile{}
	searchDeps := func(path string, info os.FileInfo, err error) error {

		// Skip excluded paths
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		// Skip ignored_pathes
		if len(config.IgnoredPaths) > 0 {
			for _, path := range config.IgnoredPaths {
				matched, err := filepath.Match(filepath.Clean(path), info.Name())
				if err != nil {
					return err
				}

				if matched {
					fmt.Println("Skipping", info.Name())
					return filepath.SkipDir
				}
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
	err := filepath.Walk(".", searchDeps)
	if err != nil {
		return dfiles, err
	}
	return dfiles, nil
}

// Push project dependencies
// The current path will be scanned for supported dependency files (SUPPORTED_DEPENDENCY_FILES)
func PushDependencyFiles(projectSlug string, files []string) error {
	dfiles, err := LookupDependencyFiles(files)
	if err != nil {
		return err
	}

	fmt.Printf("Sending files to Gemnasium: ")
	// API v1 and v2 returns completelly different informations
	switch a := api.APIImpl.(type) {
	case *api.APIv1:
		jsonResp, err := a.DependencyFilesPush(projectSlug, dfiles)
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
		fmt.Printf("done.\n\n")
		fmt.Printf("Added: %s\n", strings.Join(added, ", "))
		fmt.Printf("Updated: %s\n", strings.Join(updated, ", "))
		fmt.Printf("Unchanged: %s\n", strings.Join(unchanged, ", "))
		fmt.Printf("Unsupported: %s\n", strings.Join(unsupported, ", "))
	case *api.V2ToV1:
		//Converts dfiles to v2
		v2dfiles := []*api.V2DependencyFile{}
		for _, dfile := range dfiles {
			v2dfile := api.V2DependencyFile{}
			api.V1DependencyFileToV2(dfile, &v2dfile)
			// Base64 encode content
			v2dfile.Content = base64.StdEncoding.EncodeToString([]byte(v2dfile.Content))
			v2dfiles = append(v2dfiles, &v2dfile)
		}
		jsonResp, err := a.APIv2.DependencyFilesPush(projectSlug, v2dfiles)
		if err != nil {
			return err
		}

		fmt.Printf("Commit SHA %s in branch %s has been pushed.\n", jsonResp.CommitSHA, jsonResp.Branch)
		fmt.Printf("done.\n\n")
	}

	return nil
}

// Load dependency files if files is not empty, otherwise search in the current
// path for files
func LookupDependencyFiles(files []string) (dfiles []*api.DependencyFile, err error) {
	if len(files) > 0 {
		for _, path := range files {
			df := NewDependencyFile(path)
			if df == nil {
				err = fmt.Errorf("Unable to read file: %s", path)
				return dfiles, err
			}
			dfiles = append(dfiles, df)
		}
	} else {
		fmt.Println("[warning] No files given, scanning current directory instead.")
		files, err := getLocalDependencyFiles()
		if err != nil {
			return nil, err
		}
		dfiles = files
	}
	return dfiles, nil
}
