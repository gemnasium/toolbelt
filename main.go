package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/wsxiaoys/terminal/color"
)

type Project struct {
	Name              string `json:"name,omitempty"`
	Slug              string `json:"slug,omitempty"`
	Description       string `json:"description,omitempty"`
	Origin            string `json:"origin,omitempty"`
	Private           bool   `json:"private,omitempty"`
	Status            string `json:"status,omitempty"`
	Monitored         bool   `json:"monitored,omitempty"`
	UnmonitoredReason string `json:"unmonitored_reason,omitempty"`
}

type Package struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	Type string `json:"type"`
}

type Advisory struct {
	ID               int      `json:"id"`
	Title            string   `json:"title"`
	Identifier       string   `json:"identifier"`
	Description      string   `json:"description"`
	Solution         string   `json:"solution"`
	AffectedVersions string   `json:"affected_versions"`
	Package          Package  `json:"package"`
	CuredVersions    string   `json:"cured_versions"`
	Credits          string   `json:"credits"`
	Links            []string `json:"links"`
}

type Alert struct {
	ID       int       `json:"id"`
	Advisory Advisory  `json:"advisory"`
	OpenAt   time.Time `json:"open_at"`
	Status   string    `json:"status"`
}

type Dependency struct {
	Requirement   string  `json:"requirement"`
	LockedVersion string  `json:"locked_version"`
	Package       Package `json:"package"`
	Type          string  `json:"type"`
	FirstLevel    bool    `json:"first_level"`
	Color         string  `json:"color"`
	Advisories    []Advisory
}

type DependencyFile struct {
	Name    string `json:"name"`
	SHA     string `json:"sha,omitempty"`
	Content []byte `json:"content"`
}

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
					Flags: []cli.Flag{
						cli.BoolFlag{"private, p", "Display only private projects"},
					},
					Action: func(ctx *cli.Context) {
						AttemptLogin(ctx, config)
						err := ListProjects(config, ctx.Bool("private"))
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
			Name:      "dependencies",
			ShortName: "d",
			Usage:     "Dependencies",
			Subcommands: []cli.Command{
				{
					Name:      "list",
					ShortName: "l",
					Usage:     "List the first level dependencies of the requested project. Usage: gemnasium deps list [project_slug]",
					Action: func(ctx *cli.Context) {
						AttemptLogin(ctx, config)
						projectSlug := ctx.Args().First()
						err := ListDependencies(projectSlug, config)
						if err != nil {
							printFatal(err.Error())
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
			Name:      "alerts",
			ShortName: "a",
			Usage:     "Dependency Alerts",
			Subcommands: []cli.Command{
				{
					Name:      "list",
					ShortName: "l",
					Usage:     "List the dependency alerts the given project is affected by",
					Action: func(ctx *cli.Context) {
						AttemptLogin(ctx, config)
						projectSlug := ctx.Args().First()
						err := ListDependencyAlerts(projectSlug, config)
						if err != nil {
							printFatal(err.Error())
							os.Exit(1)
						}
					},
				},
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
