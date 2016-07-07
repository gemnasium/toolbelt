package commands

import (
	"github.com/gemnasium/toolbelt/auth"
	"github.com/urfave/cli"
)

var login = func() error {
	return auth.Login()
}

// auth.Login wrapper with a cli.Content
func Login(ctx *cli.Context) error {
	err := login()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

var logout = func() error {
	return auth.Logout()
}

// auth.Logout wrapper with a cli.Content
func Logout(ctx *cli.Context) error {
	err := logout()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}
