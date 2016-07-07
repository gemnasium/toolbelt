package commands

import (
	"github.com/gemnasium/toolbelt/auth"
	"github.com/gemnasium/toolbelt/autoupdate"
	"github.com/gemnasium/toolbelt/models"
	"github.com/urfave/cli"
)

var auRunFunc = func(projectSlug string, args []string) error {
	return autoupdate.Run(projectSlug, args)
}

var auApplyFunc = func(projectSlug string, args []string) error {
	return autoupdate.Apply(projectSlug, args)
}

func AutoUpdateRun(ctx *cli.Context) error {
	auth.AttemptLogin(ctx)
	project, err := models.GetProject(ctx.String("project"))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	err = auRunFunc(project.Slug, ctx.Args())
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func AutoUpdateApply(ctx *cli.Context) error {
	auth.AttemptLogin(ctx)
	project, err := models.GetProject(ctx.String("project"))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	err = auApplyFunc(project.Slug, ctx.Args())
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}
