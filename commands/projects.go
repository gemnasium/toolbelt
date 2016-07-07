package commands

import (
	"os"

	"github.com/gemnasium/toolbelt/models"
	"github.com/urfave/cli"
)

func ProjectsList(ctx *cli.Context) error {
	err := models.ListProjects(ctx.Bool("private"))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func ProjectsShow(ctx *cli.Context) error {
	project, err := models.GetProject(ctx.Args().First())
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	err = project.Show()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func ProjectsUpdate(ctx *cli.Context) error {
	project, err := models.GetProject(ctx.Args().First())
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
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
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func ProjectsCreate(ctx *cli.Context) error {
	projectName := ctx.Args().First()
	// will scan from os.Stding if projectName is empty
	err := models.CreateProject(projectName, os.Stdin)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func ProjectsSync(ctx *cli.Context) error {
	project, err := models.GetProject(ctx.Args().First())
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	err = project.Sync()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}
