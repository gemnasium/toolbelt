package commands

import (
	"strings"

	"github.com/urfave/cli"
	"github.com/gemnasium/toolbelt/project"
	"github.com/gemnasium/toolbelt/dependency"
)

func DependencyFilesList(ctx *cli.Context) error {
	p, err := project.GetProject()
	if err != nil {
		return err
	}
	err = dependency.ListDependencyFiles(p)
	return err
}

func DependenciesPush(ctx *cli.Context) error {
	p, err := project.GetProject()
	if err != nil {
		return err
	}
	var files []string
	if ctx.IsSet("files") {
		// Only call strings.Split on non-empty strings, otherwise len(strings) will be 1 instead of 0.
		files = strings.Split(ctx.String("files"), ",")
	}
	err = dependency.PushDependencyFiles(p.Slug, files)
	return err
}
