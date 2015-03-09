package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gemnasium/toolbelt/config"
	"github.com/mgutz/ansi"
	"github.com/wsxiaoys/terminal/color"
)

func PrintFatal(message string, args ...interface{}) {
	log.Fatal(colorizeMessage("red", "error:", message, args...))
}

func colorizeMessage(color, prefix, message string, args ...interface{}) string {
	prefResult := ""
	if prefix != "" {
		prefResult = ansi.Color(prefix, color+"+b") + " " + ansi.ColorCode("reset")
	}
	return prefResult + ansi.Color(fmt.Sprintf(message, args...), color) + ansi.ColorCode("reset")
}

// Create a new API request, with needed headers for auth and content-type
func NewAPIRequest(method, urlStr, APIKey string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("x", APIKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Gemnasium Toolbelt "+config.VERSION)
	req.Header.Add("X-Gms-Revision", GetCurrentRevision())
	req.Header.Add("X-Gms-Branch", GetCurrentBranch())
	return req, nil
}

// Return unicode colorized text dots for each status
// Status is supposed to red|yellow|green otherwise "none" will be returned
func StatusDots(status string) string {
	var dots string
	switch status {
	case "red":
		dots = "@k\u2B24 @k\u2B24 @r\u2B24  @{|}(red)"
	case "yellow":
		dots = "@k\u2B24 @y\u2B24 @k\u2B24  @{|}(yellow)"
	case "green":
		dots = "@g\u2B24 @k\u2B24 @k\u2B24  @{|}(green)"
	default:
		dots = "@k\u2B24 @k\u2B24 @k\u2B24  @{|}(none)"
	}
	return dots
}

func ExitIfErr(err error) {
	if err != nil {
		ExitWithError(err)
	}
}

func ExitWithError(err error) {
	color.Println("@{r!}" + err.Error())
	os.Exit(1)
}

// return the current commit sha, using git
// If the env var "REVISION" is specified, its value is returned directly
func GetCurrentRevision() string {
	if envRevision := os.Getenv(config.ENV_REVISION); envRevision != "" {
		return envRevision
	}
	out, err := exec.Command(GitPath(), "rev-parse", "--verify", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// Return the current branch name, using git.
// If the env var "BRANCH" is declared, its value is returned diretly
func GetCurrentBranch() string {
	if envBranch := os.Getenv(config.ENV_BRANCH); envBranch != "" {
		return envBranch
	}
	out, err := exec.Command(GitPath(), "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "master"
	}
	return strings.TrimSpace(string(out))
}

// Lookup for "git" in $PATH
func GitPath() string {
	path, _ := exec.LookPath("git")
	return path
}
