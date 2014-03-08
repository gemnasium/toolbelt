// +build windows

package main

import (
	"os"
	"os/exec"
)

const (
	netrcFilename           = "_netrc"
	acceptPasswordFromStdin = false
)

func homePath() string {
	home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
	return home
}
