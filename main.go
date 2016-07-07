package main

import (
	"os"

	"github.com/gemnasium/toolbelt/commands"
)

func main() {
	app := commands.App()
	app.Run(os.Args)
}
