package commands

import (
	"strings"

	"github.com/gemnasium/toolbelt/models"
	"github.com/urfave/cli"
)

func DependencyFilesList(ctx *cli.Context) error {
	project, err := models.GetProject()
	if err != nil {
		return err
	}
	err = models.ListDependencyFiles(project)
	return err
}

func DependenciesPush(ctx *cli.Context) error {
	project, err := models.GetProject()
	if err != nil {
		return err
	}
	var files []string
	if ctx.IsSet("files") {
		// Only call strings.Split on non-empty strings, otherwise len(strings) will be 1 instead of 0.
		files = strings.Split(ctx.String("files"), ",")
	}
	err = models.PushDependencyFiles(project.Slug, files)
	return err
}
