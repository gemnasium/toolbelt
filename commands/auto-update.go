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
		return err
	}
	err = auRunFunc(project.Slug, ctx.Args())
	return err
}

func AutoUpdateApply(ctx *cli.Context) error {
	auth.AttemptLogin(ctx)
	project, err := models.GetProject(ctx.String("project"))
	if err != nil {
		return err
	}
	err = auApplyFunc(project.Slug, ctx.Args())
	return err
}
