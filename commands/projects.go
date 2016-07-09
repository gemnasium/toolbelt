package commands

import (
	"os"

	"github.com/gemnasium/toolbelt/models"
	"github.com/urfave/cli"
)

func ProjectsList(ctx *cli.Context) error {
	err := models.ListProjects(ctx.Bool("private"))
	return err
}

func ProjectsShow(ctx *cli.Context) error {
	project, err := models.GetProject(ctx.Args().First())
	if err != nil {
		return err
	}

	err = project.Show()
	return err
}

func ProjectsUpdate(ctx *cli.Context) error {
	project, err := models.GetProject(ctx.Args().First())
	if err != nil {
		return err
	}
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
	return err
}

func ProjectsCreate(ctx *cli.Context) error {
	projectName := ctx.Args().First()
	// will scan from os.Stding if projectName is empty
	err := models.CreateProject(projectName, os.Stdin)
	return err
}

func ProjectsSync(ctx *cli.Context) error {
	project, err := models.GetProject(ctx.Args().First())
	if err != nil {
		return err
	}

	err = project.Sync()
	return err
}
