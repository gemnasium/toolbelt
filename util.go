package main

import (
	"errors"
	"fmt"
	"github.com/bgentry/go-netrc/netrc"
	"github.com/bgentry/speakeasy"
	"github.com/codegangsta/cli"
	"github.com/heroku/hk/term"
	"github.com/mgutz/ansi"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
)

// Error codes returned by auth failures
var (
	ErrEmptyToken = errors.New("auth: You must be logged in. Please use `gemnasium login` first, or pass your api token with --token.")
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
		}
		if err != nil {
			printFatal("loading netrc: " + err.Error())
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

func getCreds(config *Config) (user, pass string) {
	nrc := loadNetrc()
	if nrc == nil {
		return "", ""
	}

	apiURL, err := url.Parse(config.APIEndpoint)
	if err != nil {
		printFatal("invalid API URL: %s", err)
	}
	if apiURL.Host == "" {
		printFatal("missing API host: %s", config.APIEndpoint)
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

func printFatal(message string, args ...interface{}) {
	log.Fatal(colorizeMessage("red", "error:", message, args...))
}

func colorizeMessage(color, prefix, message string, args ...interface{}) string {
	prefResult := ""
	if prefix != "" {
		prefResult = ansi.Color(prefix, color+"+b") + " " + ansi.ColorCode("reset")
	}
	return prefResult + ansi.Color(fmt.Sprintf(message, args...), color) + ansi.ColorCode("reset")
}

// Lambda to be overriden in tests
var getCredentials = func() (email, password string, err error) {
	fmt.Printf("Enter your email: ")
	fmt.Scanf("%s", &email)
	// NOTE: gopass doesn't support multi-byte chars on Windows
	password, err = readPassword("Enter password (will be hidden): ")
	if err != nil {
		return "", "", err
	}
	return email, password, nil
}

func readPassword(prompt string) (password string, err error) {
	if acceptPasswordFromStdin && !term.IsTerminal(os.Stdin) {
		_, err = fmt.Scanln(&password)
		return
	}
	// NOTE: speakeasy may not support multi-byte chars on Windows
	return speakeasy.Ask("Enter password: ")
}

// Try to get credential from 3 sources (in that exact order):
// - from netrc file
// - local config file (ie: .gemnasium.yml), with a `api_key` yaml key
// - from command line flag `token`
//
// Each source will override previous one (token flag has priority above all).
//
// Returns a ErrEmptyToken error if all sources failed
func AttemptLogin(ctx *cli.Context, config *Config) error {
	// APIKey has been set localy in config file
	if config.APIKey == "" {
		_, config.APIKey = getCreds(config)
	}
	// User can override token
	if ctx.GlobalString("token") != "" {
		// Try to fetch token from command line
		config.APIKey = ctx.GlobalString("token")
	}
	if config.APIKey == "" {
		return ErrEmptyToken
	}
	return nil

}
