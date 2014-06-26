package commands

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/gemnasium/toolbelt/config"
	"github.com/gemnasium/toolbelt/models"
	"github.com/gemnasium/toolbelt/utils"
)

func ProjectsList(ctx *cli.Context) {
	err := models.ListProjects(ctx.Bool("private"))
	utils.ExitIfErr(err)
}

func ProjectsShow(ctx *cli.Context) {
	project, err := models.GetProject(ctx.Args().First(), config.ProjectSlug)
	utils.ExitIfErr(err)
	err = project.Show()
	utils.ExitIfErr(err)
}

func ProjectsUpdate(ctx *cli.Context) {
	project, err := models.GetProject(ctx.Args().First(), config.ProjectSlug)
	utils.ExitIfErr(err)
	var name, desc *string
	var monitored *bool
	if ctx.IsSet("name") {
		nameString := ctx.String("name")
		name = &nameString
	}
	if ctx.IsSet("desc") {
		descString := ctx.String("desc")
		desc = &descString
	}
	if ctx.IsSet("monitored") {
		mon := ctx.Bool("monitored")
		monitored = &mon
	}
	err = project.Update(name, desc, monitored)
	utils.ExitIfErr(err)
}

func ProjectsCreate(ctx *cli.Context) {
	projectName := ctx.Args().First()
	// will scan from os.Stding if projectName is empty
	err := models.CreateProject(projectName, os.Stdin)
	utils.ExitIfErr(err)
}

func ProjectsSync(ctx *cli.Context) {
	project, err := models.GetProject(ctx.Args().First(), config.ProjectSlug)
	utils.ExitIfErr(err)
	err = project.Sync()
	utils.ExitIfErr(err)
}
