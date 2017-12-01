package commands

import (
	"os"

	"github.com/urfave/cli"
	"github.com/gemnasium/toolbelt/api"
	"github.com/gemnasium/toolbelt/project"
)

var confFunc = func(p *api.Project) error {
	f, err := os.Create(".gemnasium.yml")
	if err != nil {
		return err
	}
	defer f.Close()
	err = project.ProjectConfigure(p, p.Slug, os.Stdin, f)
	return err
}

func Configure(ctx *cli.Context) error {
	slug := ctx.Args().First()
	p, err := project.GetProject(slug)
	if err != nil {
		return err
	}

	err = confFunc(p)
	return err
}
