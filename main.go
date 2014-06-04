package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/wsxiaoys/terminal/color"
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
		cli.BoolFlag{"raw, r", "Raw format output"},
	}
	config, _ := NewConfig([]byte{})
	app.Before = func(c *cli.Context) error {
		new_config, _ := LoadConfigFile(c.String("config"))
		if new_config != nil {
			config = new_config
		}
		// Pass the raw flag through our app config
		config.RawFormat = c.Bool("raw")
		return nil // Don't fail if config file isn't found
	}
	app.Commands = []cli.Command{
		{
			Name:  "auth",
			Usage: "Authentication",
			Subcommands: []cli.Command{
				{
					Name:  "login",
					Usage: "Login",
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
			},
		},
		{
			Name:      "project",
			ShortName: "p",
			Usage:     "Manage current project",
			Subcommands: []cli.Command{
				{
					Name:      "list",
					ShortName: "l",
					Usage:     "List projects on Gemnasium",
					Action: func(ctx *cli.Context) {
						AttemptLogin(ctx, config)
						err := ListProjects(config)
						if err != nil {
							ExitWithError(err)
						}
					},
				},
				{
					Name:      "show",
					ShortName: "s",
					Usage:     "Show projet detail",
					Flags: []cli.Flag{
						cli.StringFlag{"slug", "", "Project slug (unique identifier on Gemnasium)"},
					},
					Action: func(ctx *cli.Context) {
						AttemptLogin(ctx, config)
						err := GetProject(ctx.Args().First(), config)
						if err != nil {
							ExitWithError(err)
						}
					},
				},
				{
					Name:      "create",
					ShortName: "c",
					Usage:     "Create a new project on Gemnasium",
					Action: func(ctx *cli.Context) {
						AttemptLogin(ctx, config)
						projectName := ctx.Args().First()
						err := CreateProject(projectName, config, os.Stdin)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
					},
				},
				{
					Name:  "configure",
					Usage: "Install configuration for an existing project. Warning: this command will overwrite existing .gemnasium.yml file",
					Action: func(ctx *cli.Context) {
						f, err := os.Create(".gemnasium.yml")
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
						defer f.Close()

						slug := ctx.Args().First()
						err = ConfigureProject(slug, config, os.Stdin, f)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
					},
				},
			},
		},
		{
			Name:      "dependency_files",
			ShortName: "df",
			Usage:     "Dependency files",
			Subcommands: []cli.Command{
				{
					Name:      "push",
					ShortName: "p",
					Usage:     "Push dependencies files on Gemnasium",
					Action: func(ctx *cli.Context) {
						err := PushDependencies(ctx, config)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
					},
				},
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
		{
			Name:      "eval",
			ShortName: "e",
			Usage:     "Live deps evaluation",
			Flags: []cli.Flag{
				cli.StringFlag{"files, f", "", "list of files to evaluate, separated with a comma."},
			},
			Action: func(ctx *cli.Context) {
				AttemptLogin(ctx, config)
				files := strings.Split(ctx.String("files"), ",")
				err := LiveEvaluation(files, config)
				if err != nil {
					ExitWithError(err)
				}
			},
		},
	}

	app.Run(os.Args)
}

func ExitWithError(err error) {
	color.Println("@{r!}" + err.Error())
	os.Exit(1)
}
