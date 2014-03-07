package main

import (
	"fmt"
	"github.com/bgentry/go-netrc/netrc"
	"github.com/mgutz/ansi"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var nrc *netrc.Netrc

const (
	netrcFilename           = ".netrc"
	acceptPasswordFromStdin = true
)

func netrcPath() string {
	if s := os.Getenv("NETRC_PATH"); s != "" {
		return s
	}

	return filepath.Join(os.Getenv("HOME"), netrcFilename)
}

var loadNetrc = func() {
	if nrc == nil {
		var err error
		if nrc, err = netrc.ParseFile(netrcPath()); err != nil {
			if os.IsNotExist(err) {
				nrc = &netrc.Netrc{}
				return
			}
			printFatal("loading netrc: " + err.Error())
		}
	}
}

func saveCreds(host, user, pass string) error {
	loadNetrc()
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
	loadNetrc()
	nrc.RemoveMachine(host)

	body, err := nrc.MarshalText()
	if err != nil {
		return err
	}
	return writeNetrcFile(body)
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
