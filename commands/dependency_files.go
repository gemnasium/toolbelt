package commands

import (
	"strings"

	"github.com/codegangsta/cli"
	"github.com/gemnasium/toolbelt/models"
	"github.com/gemnasium/toolbelt/utils"
)

func DependencyFilesList(ctx *cli.Context) {
	project, err := models.GetProject(ctx.Args().First())
	utils.ExitIfErr(err)
	err = models.ListDependencyFiles(project)
	utils.ExitIfErr(err)
}

func DependenciesPush(ctx *cli.Context) {
	project, err := models.GetProject(ctx.Args().First())
	utils.ExitIfErr(err)
	files := strings.Split(ctx.String("files"), ",")
	err = models.PushDependencyFiles(project.Slug, files)
	utils.ExitIfErr(err)
}
