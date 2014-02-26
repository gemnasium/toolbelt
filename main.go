package main

import (
	"github.com/codegangsta/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "gemnasium"
	app.Usage = "Interact with your gemnasium account"
	app.Version = "0.1.0"
	app.Author = "Gemnasium"
	app.Email = "support@gemnasium.com"

	app.Run(os.Args)
}
