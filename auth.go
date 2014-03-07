package main

import (
	"fmt"
	"github.com/bgentry/speakeasy"
	"github.com/heroku/hk/term"
	"os"
)

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
