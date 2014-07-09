package commands

import (
	"os"
	"testing"

	"github.com/gemnasium/toolbelt/config"
)

func TestAutoUpdate(t *testing.T) {
	config.APIKey = "abcdef123"
	config.ProjectSlug = "projectSlug"

	var project string
	auFunc = func(slug string, args []string) error {
		project = slug
		return nil
	}
	app, err := App()
	if err != nil {
		t.Fatal(err)
	}

	// Call autoupdate command
	os.Args = []string{"gemnasium", "autoupdate"}
	app.Run(os.Args)
	if project != "projectSlug" {
		t.Errorf("Should have called autoupdate func\n")
	}

	// With alias
	project = ""
	os.Args = []string{"gemnasium", "au"}
	app.Run(os.Args)
	if project != "projectSlug" {
		t.Errorf("Should have called autoupdate func\n")
	}

	// with Flag
	project = ""
	os.Args = []string{"gemnasium", "au", "-p=slug"}
	app.Run(os.Args)
	if project != "slug" {
		t.Errorf("Should have called autoupdate func\n")
	}
}
