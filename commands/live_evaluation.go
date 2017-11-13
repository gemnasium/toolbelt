package commands

import (
	"strings"

	"github.com/gemnasium/toolbelt/auth"
	"github.com/gemnasium/toolbelt/live-eval"
	"github.com/urfave/cli"
	"github.com/gemnasium/toolbelt/api"
	"errors"
)

func LiveEvaluation(ctx *cli.Context) error {
	// Live evaluation is not available on API v2
	switch api.APIImpl.(type) {
	case *api.V2ToV1:
		return errors.New("Live dependencies evaluation is not available on API version 2.")
	}
	auth.ConfigureAPIToken(ctx)
	files := strings.Split(ctx.String("files"), ",")
	err := liveeval.LiveEvaluation(files)
	return err
}
