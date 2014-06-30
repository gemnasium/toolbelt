package commands

import (
	"github.com/codegangsta/cli"
	"github.com/gemnasium/toolbelt/auth"
	"github.com/gemnasium/toolbelt/autoupdate"
	"github.com/gemnasium/toolbelt/models"
	"github.com/gemnasium/toolbelt/utils"
)

func AutoUpdate(ctx *cli.Context) {
	auth.AttemptLogin(ctx)
	project, err := models.GetProject(ctx.String("project"))
	utils.ExitIfErr(err)
	err = autoupdate.Run(project.Slug, ctx.Args())
	utils.ExitIfErr(err)
}
