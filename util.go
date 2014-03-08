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
