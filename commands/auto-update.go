package commands

import (
	"github.com/codegangsta/cli"
	"github.com/gemnasium/toolbelt/auth"
	"github.com/gemnasium/toolbelt/autoupdate"
	"github.com/gemnasium/toolbelt/models"
	"github.com/gemnasium/toolbelt/utils"
)

var auRunFunc = func(projectSlug string, args []string) error {
	return autoupdate.Run(projectSlug, args)
}

var auApplyFunc = func(projectSlug string, args []string) error {
	return autoupdate.Apply(projectSlug, args)
}

func AutoUpdateRun(ctx *cli.Context) {
	auth.AttemptLogin(ctx)
	project, err := models.GetProject(ctx.String("project"))
	utils.ExitIfErr(err)
	err = auRunFunc(project.Slug, ctx.Args())
	utils.ExitIfErr(err)
}

func AutoUpdateApply(ctx *cli.Context) {
	auth.AttemptLogin(ctx)
	project, err := models.GetProject(ctx.String("project"))
	utils.ExitIfErr(err)
	err = auApplyFunc(project.Slug, ctx.Args())
	utils.ExitIfErr(err)
}
