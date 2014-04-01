package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "gemnasium"
	app.Usage = "Interact with your gemnasium account"
	app.Version = "0.1.0"
	app.Author = "Gemnasium"
	app.Email = "support@gemnasium.com"
	app.Flags = []cli.Flag{
		cli.StringFlag{"config, c", ".gemnasium.yml", "Path to config file"},
		cli.StringFlag{"token, t", "", "Pass your api token in command line"},
	}

	config, _ := NewConfig([]byte{})
	app.Before = func(c *cli.Context) error {
		config, _ = LoadConfigFile(c.String("config"))
		return nil // Don't fail if config file isn't found
	}
	app.Commands = []cli.Command{
		{
			Name:      "login",
			ShortName: "l",
			Usage:     "Login",
			Action: func(ctx *cli.Context) {
				err := Login(ctx, config)
				if err != nil {
					printFatal(err.Error())
					os.Exit(1)
				}
			},
		},
		{
			Name:  "logout",
			Usage: "Logout",
			Action: func(ctx *cli.Context) {
				err := Logout(ctx, config)
				if err != nil {
					printFatal(err.Error())
					os.Exit(1)
				}
			},
		},
		{
			Name:      "create",
			ShortName: "c",
			Usage:     "Create a new project on Gemnasium",
			Action: func(ctx *cli.Context) {
				err := CreateProject(ctx, config)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			},
		},
		{
			Name:      "configure",
			ShortName: "i",
			Usage:     "Install configuration for an existing project",
			Action: func(ctx *cli.Context) {
				println("Project configured!")
			},
		},
		{
			Name:      "push",
			ShortName: "p",
			Usage:     "Push dependencies files on Gemnasium",
			Action: func(c *cli.Context) {
				println("Files pushed!")
			},
		},
		{
			Name:      "changelog",
			ShortName: "cl",
			Usage:     "Check changelog for a given package",
			Action: func(c *cli.Context) {
				package_name := c.Args().First()
				if package_name == "" {
					fmt.Println("Error: You must specify a package name")
					os.Exit(1)
				}
				changelog, err := Changelog(package_name)
				if err != nil {
					fmt.Println("Error: You must specify a package name")
					os.Exit(1)
				}

				println(changelog)
			},
		},
	}

	app.Run(os.Args)
}
