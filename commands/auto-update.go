package commands

import (
	"github.com/codegangsta/cli"
	"github.com/gemnasium/toolbelt/auth"
	"github.com/gemnasium/toolbelt/autoupdate"
	"github.com/gemnasium/toolbelt/models"
	"github.com/gemnasium/toolbelt/utils"
)

var auFunc = func(projectSlug string, args []string) error {
	return autoupdate.Run(projectSlug, args)
}

func AutoUpdateRun(ctx *cli.Context) {
	auth.AttemptLogin(ctx)
	project, err := models.GetProject(ctx.String("project"))
	utils.ExitIfErr(err)
	err = auFunc(project.Slug, ctx.Args())
	utils.ExitIfErr(err)
}
