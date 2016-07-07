package commands

import (
	"os"
	"testing"

	"github.com/gemnasium/toolbelt/config"
)

func TestAutoUpdateRun(t *testing.T) {
	config.APIKey = "abcdef123"
	config.ProjectSlug = "projectSlug"

	var project string
	auRunFunc = func(slug string, args []string) error {
		project = slug
		return nil
	}
	app := App()

	// Call autoupdate command
	os.Args = []string{"gemnasium", "autoupdate", "run"}
	app.Run(os.Args)
	if project != "projectSlug" {
		t.Errorf("Should have called autoupdate func\n")
	}

	// With alias
	project = ""
	os.Args = []string{"gemnasium", "au", "r"}
	app.Run(os.Args)
	if project != "projectSlug" {
		t.Errorf("Should have called autoupdate func\n")
	}

	// with Flag
	project = ""
	os.Args = []string{"gemnasium", "au", "r", "-p=slug"}
	app.Run(os.Args)
	if project != "slug" {
		t.Errorf("Should have called autoupdate func\n")
	}
}
