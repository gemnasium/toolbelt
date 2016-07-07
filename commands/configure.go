package commands

import (
	"os"

	"github.com/gemnasium/toolbelt/models"
	"github.com/urfave/cli"
)

var confFunc = func(project *models.Project) error {
	f, err := os.Create(".gemnasium.yml")
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	defer f.Close()
	err = project.Configure(project.Slug, os.Stdin, f)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func Configure(ctx *cli.Context) error {
	slug := ctx.Args().First()
	project, err := models.GetProject(slug)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	err = confFunc(project)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}
