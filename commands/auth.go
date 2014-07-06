package commands

import (
	"github.com/codegangsta/cli"
	"github.com/gemnasium/toolbelt/auth"
	"github.com/gemnasium/toolbelt/utils"
)

var login = func() error {
	return auth.Login()
}

// auth.Login wrapper with a cli.Content
func Login(ctx *cli.Context) {
	err := login()
	utils.ExitIfErr(err)
}

var logout = func() error {
	return auth.Logout()
}

// auth.Logout wrapper with a cli.Content
func Logout(ctx *cli.Context) {
	err := logout()
	utils.ExitIfErr(err)
}
