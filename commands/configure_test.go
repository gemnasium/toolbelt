package commands

import (
	"bytes"
	"os"
	"testing"

	"github.com/gemnasium/toolbelt/config"
	"github.com/gemnasium/toolbelt/models"
)

func TestConfigue(t *testing.T) {
	config.APIKey = "abcdef123"
	config.ProjectSlug = "projectSlug"

	var output bytes.Buffer
	confFunc = func(project *models.Project) error {
		return project.Configure(project.Slug, os.Stdin, &output)
	}
	app, err := App()
	if err != nil {
		t.Fatal(err)
	}

	// Call autoupdate command
	os.Args = []string{"gemnasium", "configure", "myProject"}
	app.Run(os.Args)
	if output.String() != "project_slug: myProject\n" {
		t.Errorf("Config file should contain: 'project_slug: MyProject', got: '%s'", output.String())
	}
}
