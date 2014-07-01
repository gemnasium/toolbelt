package commands

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/gemnasium/toolbelt/models"
	"github.com/gemnasium/toolbelt/utils"
)

func Configure(ctx *cli.Context) {
	f, err := os.Create(".gemnasium.yml")
	utils.ExitIfErr(err)
	defer f.Close()

	slug := ctx.Args().First()
	project, err := models.GetProject(slug)
	utils.ExitIfErr(err)

	// TODO: slug can be empty
	err = project.Configure(project.Slug, os.Stdin, f)
	utils.ExitIfErr(err)
}
