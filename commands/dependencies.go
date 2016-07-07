package commands

import (
	"github.com/gemnasium/toolbelt/models"
	"github.com/urfave/cli"
)

func DependenciesList(ctx *cli.Context) error {
	project, err := models.GetProject(ctx.Args().First())
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	err = models.ListDependencies(project)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}
