package commands

import (
	"bytes"
	"os"
	"testing"

	"github.com/gemnasium/toolbelt/config"
	"github.com/gemnasium/toolbelt/api"
	"github.com/gemnasium/toolbelt/project"
)

func TestConfigue(t *testing.T) {
	config.APIKey = "abcdef123"
	config.ProjectSlug = "projectSlug"

	var output bytes.Buffer
	confFunc = func(p *api.Project) error {
		return project.ProjectConfigure(p, p.Slug, os.Stdin, &output)
	}
	app := App()

	// Call autoupdate command
	os.Args = []string{"gemnasium", "configure", "myProject"}
	app.Run(os.Args)
	if output.String() != "project_slug: myProject\n" {
		t.Errorf("Config file should contain: 'project_slug: MyProject', got: '%s'", output.String())
	}
}
