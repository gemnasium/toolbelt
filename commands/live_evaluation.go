package commands

import (
	"strings"

	"github.com/codegangsta/cli"
	"github.com/gemnasium/toolbelt/auth"
	"github.com/gemnasium/toolbelt/live-eval"
	"github.com/gemnasium/toolbelt/utils"
)

func LiveEvaluation(ctx *cli.Context) {
	auth.AttemptLogin(ctx)
	if ctx.String("files") == "" {
		cli.ShowCommandHelp(ctx, "eval")
	}
	files := strings.Split(ctx.String("files"), ",")
	err := liveeval.LiveEvaluation(files)
	utils.ExitIfErr(err)
}
