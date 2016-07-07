package commands

import (
	"os"
	"testing"
)

func TestLogin(t *testing.T) {
	var called bool
	orgLogin := login
	login = func() error {
		called = true
		return nil
	}
	app := App()
	os.Args = []string{"gemnasium", "auth", "login"}
	app.Run(os.Args)
	if called != true {
		t.Errorf("Should have called login func\n")
	}
	// restore login func
	login = orgLogin
}

func TestLogout(t *testing.T) {
	var called bool
	logout = func() error {
		called = true
		return nil
	}
	app := App()
	os.Args = []string{"gemnasium", "auth", "logout"}
	app.Run(os.Args)
	if called != true {
		t.Errorf("Should have called login func\n")
	}
}
