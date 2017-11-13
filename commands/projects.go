package commands

import (
	"os"

	"github.com/urfave/cli"
	"github.com/gemnasium/toolbelt/project"
	"github.com/gemnasium/toolbelt/api"
	"errors"
)

func ProjectsList(ctx *cli.Context) error {
	err := project.ListProjects(ctx.Bool("private"))
	return err
}

func ProjectsShow(ctx *cli.Context) error {
	p, err := project.GetProject(ctx.Args().First())
	if err != nil {
		return err
	}

	err = project.ProjectShow(p)
	return err
}

func ProjectsUpdate(ctx *cli.Context) error {
	p, err := project.GetProject(ctx.Args().First())
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
		switch api.APIImpl.(type) {
		case *api.APIv1:
			// Ok
		default:
			// Monitoring is an API v1 feature only
			return errors.New("Setting monitored can only be used on API version 1.")
		}
		mon := ctx.Bool("monitored")
		monitored = &mon
	}
	err = project.ProjectUpdate(p, name, desc, monitored)
	return err
}

func ProjectsCreate(ctx *cli.Context) error {
	projectName := ctx.Args().First()
	// will scan from os.Stding if projectName is empty
	err := project.CreateProject(projectName, os.Stdin)
	return err
}

func ProjectsSync(ctx *cli.Context) error {
	p, err := project.GetProject(ctx.Args().First())
	if err != nil {
		return err
	}

	err = project.ProjectSync(p)
	return err
}
