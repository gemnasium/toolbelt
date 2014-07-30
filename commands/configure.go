package commands

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/gemnasium/toolbelt/models"
	"github.com/gemnasium/toolbelt/utils"
)

var confFunc = func(project *models.Project) error {
	f, err := os.Create(".gemnasium.yml")
	utils.ExitIfErr(err)
	defer f.Close()
	return project.Configure(project.Slug, os.Stdin, f)
}

func Configure(ctx *cli.Context) {
	slug := ctx.Args().First()
	project, err := models.GetProject(slug)
	utils.ExitIfErr(err)

	err = confFunc(project)
	utils.ExitIfErr(err)
}
