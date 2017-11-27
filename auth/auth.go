package auth

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/bgentry/speakeasy"
	"github.com/gemnasium/toolbelt/config"
	"github.com/gemnasium/toolbelt/utils"
	"github.com/gemnasium/toolbelt/api"
	"github.com/heroku/hk/term"
	"github.com/urfave/cli"

	"fmt"
	"io/ioutil"
	"net/url"
)

const (
	LOGIN_PATH = "/login"
)

// Login with the user email and password
// An entry will be created in ~/.netrc on successful login.
func Login() error {
	// Create a function to be overriden in tests
	email := getEmail()
	password, err := getPassword("Enter password (will be hidden): ")
	if err != nil {
		return err
	}

	err = api.APIImpl.Login(email, password)
	if err != nil {
		return err
	}

	err = saveCreds(api.APIImpl.Host(), email, api.APIImpl.Key())
	if err != nil {
		utils.PrintFatal("saving new token: " + err.Error())
	}
	fmt.Println("Logged in.")

	return nil
}

// Login with the user email and API token
// An entry will be created in ~/.netrc on successful login.
func LoginWithAPIToken(token string) (err error) {
	// Create a function to be overriden in tests
	email := getEmail()
	if err != nil {
		return err
	}

	// Configure API with the token
	api.APIImpl.SetKey(token)
	if err != nil {
		return err
	}

	err = saveCreds(api.APIImpl.Host(), email, api.APIImpl.Key())
	if err != nil {
		utils.PrintFatal("saving new token: " + err.Error())
	}
	fmt.Println("Logged in.")

	return nil
}

// Logout doesn't hit the API of course.
// It simply removes the corresponding entry in ~/.netrc
func Logout() error {
	err := removeCreds(api.APIImpl.Host())
	if err != nil {
		return err
	}
	fmt.Println("Logged out.")
	return nil
}

// Lambda to be overriden in tests
var getEmail = func() (email string) {
	fmt.Printf("Enter your email: ")
	fmt.Scanf("%s", &email)
	return email
}

// Lambda to be overriden in tests
var getPassword = func(prompt string) (password string, err error) {
	// NOTE: gopass doesn't support multi-byte chars on Windows
	password, err = readPassword(prompt)
	if err != nil {
		return "", err
	}
	return password, nil
}

func readPassword(prompt string) (password string, err error) {
	if acceptPasswordFromStdin && !term.IsTerminal(os.Stdin) {
		_, err = fmt.Scanln(&password)
		return
	}
	// NOTE: speakeasy may not support multi-byte chars on Windows
	return speakeasy.Ask(prompt)
}

// Try to get credential from 3 sources (in that exact order):
// - from netrc file
// - local config file (ie: .gemnasium.yml), with a `api_key` yaml key
// - from command line flag `token`
//
// Each source will override previous one (token flag has priority above all).
//
// WARNING: Directly exit the programm in case of error
func ConfigureAPIToken(ctx *cli.Context) error {
	// APIKey has been set localy in APIconfig file
	if config.APIKey == "" {
		_, config.APIKey = getCreds()
	}
	// User can override token
	if ctx.GlobalString("token") != "" {
		// Try to fetch token from command line
		config.APIKey = ctx.GlobalString("token")
	}
	// Configure the API instance with the chosen token
	api.APIImpl.SetKey(config.APIKey)
	if config.APIKey == "" {
		return ErrEmptyToken
	}
	return nil
}

// Error codes returned by auth failures
var (
	ErrEmptyToken = errors.New("auth: You must be logged in. Please use `gemnasium auth login` first, or pass your api token with --token or GEMNASIUM_TOKEN")
)

func netrcPath() string {
	if s := os.Getenv("NETRC_PATH"); s != "" {
		return s
	}

	return filepath.Join(os.Getenv("HOME"), netrcFilename)
}

var loadNetrc = func() *netrc.Netrc {
	nrc, err := netrc.ParseFile(netrcPath())
	if err != nil {
		if os.IsNotExist(err) {
			nrc = &netrc.Netrc{}
		} else {
			utils.PrintFatal("loading netrc: " + err.Error())
		}
	}
	return nrc
}

func saveCreds(host, user, pass string) error {
	nrc := loadNetrc()
	m := nrc.FindMachine(host)
	if m == nil || m.IsDefault() {
		m = nrc.NewMachine(host, user, pass, "")
	}
	m.UpdateLogin(user)
	m.UpdatePassword(pass)

	body, err := nrc.MarshalText()
	if err != nil {
		return err
	}
	return writeNetrcFile(body)
}

var writeNetrcFile = func(body []byte) error {
	return ioutil.WriteFile(netrcPath(), body, 0600)
}

func removeCreds(host string) error {
	nrc := loadNetrc()
	nrc.RemoveMachine(host)

	body, err := nrc.MarshalText()
	if err != nil {
		return err
	}
	return writeNetrcFile(body)
}

func getCreds() (user, pass string) {
	nrc := loadNetrc()
	if nrc == nil {
		return "", ""
	}

	apiURL, err := url.Parse(api.APIImpl.Endpoint())
	if err != nil {
		utils.PrintFatal("invalid API URL: %s", err)
	}
	if apiURL.Host == "" {
		utils.PrintFatal("missing API host: %s", config.APIEndpoint)
	}
	if apiURL.User != nil {
		pw, _ := apiURL.User.Password()
		return apiURL.User.Username(), pw
	}

	m := nrc.FindMachine(apiURL.Host)
	if m == nil {
		return "", ""
	}
	return m.Login, m.Password
}
