package commands

import (
	"github.com/urfave/cli"
	"github.com/gemnasium/toolbelt/config"
)

func DisplayEnvVars(ctx *cli.Context) {
	config.DisplayEnvVars()
}
