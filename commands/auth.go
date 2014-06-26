package commands

import (
	"github.com/codegangsta/cli"
	"github.com/gemnasium/toolbelt/auth"
	"github.com/gemnasium/toolbelt/utils"
)

// auth.Login wrapper with a cli.Content
func Login(ctx *cli.Context) {
	err := auth.Login()
	utils.ExitWithError(err)
}

// auth.Logout wrapper with a cli.Content
func Logout(ctx *cli.Context) {
	err := auth.Logout()
	utils.ExitWithError(err)
}
