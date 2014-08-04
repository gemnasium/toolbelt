package commands

import (
	"strings"

	"github.com/codegangsta/cli"
	"github.com/gemnasium/toolbelt/models"
	"github.com/gemnasium/toolbelt/utils"
)

func DependencyFilesList(ctx *cli.Context) {
	project, err := models.GetProject()
	utils.ExitIfErr(err)
	err = models.ListDependencyFiles(project)
	utils.ExitIfErr(err)
}

func DependenciesPush(ctx *cli.Context) {
	project, err := models.GetProject()
	utils.ExitIfErr(err)
	var files []string
	if ctx.IsSet("files") {
		// Only call strings.Split on non-empty strings, otherwise len(strings) will be 1 instead of 0.
		files = strings.Split(ctx.String("files"), ",")
	}
	err = models.PushDependencyFiles(project.Slug, files)
	utils.ExitIfErr(err)
}
