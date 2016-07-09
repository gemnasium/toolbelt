package commands

import (
	"os"

	"github.com/gemnasium/toolbelt/models"
	"github.com/urfave/cli"
)

var confFunc = func(project *models.Project) error {
	f, err := os.Create(".gemnasium.yml")
	if err != nil {
		return err
	}
	defer f.Close()
	err = project.Configure(project.Slug, os.Stdin, f)
	return err
}

func Configure(ctx *cli.Context) error {
	slug := ctx.Args().First()
	project, err := models.GetProject(slug)
	if err != nil {
		return err
	}

	err = confFunc(project)
	return err
}
