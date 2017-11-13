package commands

import (
	"github.com/gemnasium/toolbelt/auth"
	"github.com/gemnasium/toolbelt/autoupdate"
	"github.com/urfave/cli"
	"github.com/gemnasium/toolbelt/api"
	"errors"
	"github.com/gemnasium/toolbelt/project"
)

var auRunFunc = func(projectSlug string, args []string) error {
	return autoupdate.Run(projectSlug, args)
}

var auApplyFunc = func(projectSlug string, args []string) error {
	return autoupdate.Apply(projectSlug, args)
}

func AutoUpdateRun(ctx *cli.Context) error {
	// Auto update is not available on API v2
	switch api.APIImpl.(type) {
	case *api.V2ToV1:
		return errors.New("Auto update is not available on API version 2.")
	}
	auth.ConfigureAPIToken(ctx)
	p, err := project.GetProject(ctx.String("project"))
	if err != nil {
		return err
	}
	err = auRunFunc(p.Slug, ctx.Args())
	return err
}

func AutoUpdateApply(ctx *cli.Context) error {
	// Auto update is not available on API v2
	switch api.APIImpl.(type) {
	case *api.V2ToV1:
		return errors.New("Auto update is not available on API version 2.")
	}
	auth.ConfigureAPIToken(ctx)
	p, err := project.GetProject(ctx.String("project"))
	if err != nil {
		return err
	}
	err = auApplyFunc(p.Slug, ctx.Args())
	return err
}
