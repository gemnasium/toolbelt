package commands

import (
	"github.com/urfave/cli"
	"github.com/gemnasium/toolbelt/project"
	"github.com/gemnasium/toolbelt/dependency"
)

func DependencyAlertsList(ctx *cli.Context) error {
	p, err := project.GetProject(ctx.Args().First())
	if err != nil {
		return err
	}

	err = dependency.ListDependencyAlerts(p)
	return err
}
