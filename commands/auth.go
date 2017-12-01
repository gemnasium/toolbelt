package commands

import (
	"github.com/gemnasium/toolbelt/auth"
	"github.com/urfave/cli"
)

var login = func() error {
	return auth.Login()
}

var login_with_api_token = func(api_token string) error {
	return auth.LoginWithAPIToken(api_token)
}

// auth.Login wrapper with a cli.Content
func Login(ctx *cli.Context) (err error) {
	if ctx.IsSet("with-api-token") {
		// log in with the provided token
		api_token := ctx.String("with-api-token")
		err = login_with_api_token(api_token)
	} else {
		// log in with the user and password
		err = login()
	}
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
