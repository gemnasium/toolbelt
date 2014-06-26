package main

import (
	"os"

	"github.com/gemnasium/toolbelt/commands"
	"github.com/gemnasium/toolbelt/utils"
)

func main() {
	app, err := commands.App()
	utils.ExitIfErr(err)
	app.Run(os.Args)
}
