package commands

import (
	"github.com/codegangsta/cli"
	"github.com/gemnasium/toolbelt/config"
)

func DisplayEnvVars(ctx *cli.Context) {
	config.DisplayEnvVars()
}
