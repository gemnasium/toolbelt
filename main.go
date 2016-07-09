package main

import (
	"os"

	"github.com/gemnasium/toolbelt/commands"
	"github.com/wsxiaoys/terminal/color"
)

func main() {
	app := commands.App()
	err := app.Run(os.Args)
	if err != nil {
		color.Printf("@{r!}%s", err.Error())
		os.Exit(1)
	}
}
