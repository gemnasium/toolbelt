package commands

import (
	"github.com/codegangsta/cli"
	"github.com/gemnasium/toolbelt/config"
	"github.com/gemnasium/toolbelt/models"
	"github.com/gemnasium/toolbelt/utils"
)

func DependencyFilesList(ctx *cli.Context) {
	project, err := models.GetProject(ctx.Args().First(), config.ProjectSlug)
	utils.ExitIfErr(err)
	err = models.ListDependencyFiles(project)
	utils.ExitIfErr(err)
}

func DependenciesPush(ctx *cli.Context) {
	project, err := models.GetProject(ctx.Args().First(), config.ProjectSlug)
	utils.ExitIfErr(err)
	err = models.PushDependencyFiles(project.Slug)
	utils.ExitIfErr(err)
}
