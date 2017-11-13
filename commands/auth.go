package commands

import (
	"github.com/gemnasium/toolbelt/auth"
	"github.com/urfave/cli"
)

var login = func() error {
	return auth.Login()
}

var login_with_api_token = func() error {
	return auth.LoginWithAPIToken()
}

// auth.Login wrapper with a cli.Content
func Login(ctx *cli.Context) error {
	err := login()
	return err
}

func LoginWithAPIToken(ctx *cli.Context) error {
	err := login_with_api_token()
	return err
}

var logout = func() error {
	return auth.Logout()
}

// auth.Logout wrapper with a cli.Content
func Logout(ctx *cli.Context) error {
	err := logout()
	return err
}
