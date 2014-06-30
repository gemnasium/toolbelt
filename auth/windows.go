// +build windows

package auth

import "os"

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
