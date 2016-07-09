package commands

import (
	"strings"

	"github.com/gemnasium/toolbelt/auth"
	"github.com/gemnasium/toolbelt/live-eval"
	"github.com/urfave/cli"
)

func LiveEvaluation(ctx *cli.Context) error {
	auth.AttemptLogin(ctx)
	files := strings.Split(ctx.String("files"), ",")
	err := liveeval.LiveEvaluation(files)
	return err
}
