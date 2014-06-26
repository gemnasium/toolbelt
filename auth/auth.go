package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/bgentry/speakeasy"
	"github.com/codegangsta/cli"
	"github.com/gemnasium/toolbelt/config"
	"github.com/gemnasium/toolbelt/utils"
	"github.com/heroku/hk/term"

	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	LOGIN_PATH = "/login"
)

// Login with the user email and password
// An entry will be created in ~/.netrc on successful login.
func Login() error {
	// Create a function to be overriden in tests
	email, password, err := getCredentials()
	if err != nil {
		return err
	}

	loginAsJson, err := json.Marshal(map[string]string{"email": email, "password": password})
	if err != nil {
		return err
	}
	resp, err := http.Post(config.APIEndpoint+LOGIN_PATH, "application/json", bytes.NewReader(loginAsJson))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Server returned non-200 status: %v\n", resp.Status)
	}

	// Read api token from response
	// body will be of the form:
	// {"api_token": "abcxzy123"}

	var response_body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&response_body)
	if err != nil {
		return err
	}

	api_token := response_body["api_token"]

	err = saveCreds(strings.Split(resp.Request.Host, ":")[0], email, api_token)
	if err != nil {
		utils.PrintFatal("saving new token: " + err.Error())
	}
	fmt.Println("Logged in.")

	return nil
}

// Logout doesn't hit the API of course.
// It simply removes the corresponding entry in ~/.netrc
func Logout() error {
	api_url, err := url.Parse(config.APIEndpoint)
	if err != nil {
		return err
	}
	err = removeCreds(api_url.Host)
	if err != nil {
		return err
	}
	fmt.Println("Logged out.")
	return nil
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
// WARNING: Directly exit the programm in case of error
func AttemptLogin(ctx *cli.Context) error {
	// APIKey has been set localy in config file
	if config.APIKey == "" {
		_, config.APIKey = getCreds()
	}
	// User can override token
	if ctx.GlobalString("token") != "" {
		// Try to fetch token from command line
		config.APIKey = ctx.GlobalString("token")
	}
	if config.APIKey == "" {
		utils.ExitWithError(ErrEmptyToken)
	}
	return nil
}

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

	apiURL, err := url.Parse(config.APIEndpoint)
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
