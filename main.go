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
						err := Login(config)
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
						err := Logout(config)
						if err != nil {
							printFatal(err.Error())
							os.Exit(1)
						}
					},
				},
			},
		},
		{
			Name:      "projects",
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
					Action: func(ctx *cli.Context) {
						AttemptLogin(ctx, config)
						err := ShowProject(ctx.Args().First(), config)
						if err != nil {
							ExitWithError(err)
						}
					},
				},
				{
					Name:      "update",
					ShortName: "u",
					Flags: []cli.Flag{
						cli.StringFlag{"name, n", "", "Project name"},
						cli.StringFlag{"desc, d", "", "A short description"},
						cli.BoolFlag{"monitored, m", "Whether the project is watched by the user. "},
					},
					Usage: "Edit project details",
					Action: func(ctx *cli.Context) {
						AttemptLogin(ctx, config)
						var name, desc *string
						var monitored *bool
						if ctx.IsSet("name") {
							nameString := ctx.String("name")
							name = &nameString
						}
						if ctx.IsSet("desc") {
							descString := ctx.String("desc")
							desc = &descString
						}
						if ctx.IsSet("monitored") {
							mon := ctx.Bool("monitored")
							monitored = &mon
						}
						err := UpdateProject(ctx.Args().First(), config, name, desc, monitored)
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
				{
					Name:  "sync",
					Usage: "Start project synchronization",
					Action: func(ctx *cli.Context) {
						AttemptLogin(ctx, config)
						slug := ctx.Args().First()
						err := SyncProject(slug, config)
						if err != nil {
							ExitWithError(err)
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
					Name:      "list",
					ShortName: "l",
					Usage:     "List dependency files for project",
					Action: func(ctx *cli.Context) {
						AttemptLogin(ctx, config)
						projectSlug := ctx.Args().First()
						err := ListDependencyFiles(projectSlug, config)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
					},
				},
				{
					Name:      "push",
					ShortName: "p",
					Usage:     "Push dependency files on Gemnasium",
					Action: func(ctx *cli.Context) {
						AttemptLogin(ctx, config)
						projectSlug := ctx.Args().First()
						err := PushDependencyFiles(projectSlug, config)
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
		{
			Name:      "autoupdate",
			ShortName: "au",
			Usage:     "Auto-update will test updates in your project, and notify Gemnasium of the result",
			Description: `Auto-Update will fetch update sets from Gemnasium and run your test suite against them.
  The test suite can be passed as second argument (first being the project_slug), or through the env var GEMNASIUM_TESTSUITE.

  Arguments:

  - project_slug (string): Project ID on Gemnasium.
  - [test suite commands] (string):

  Env Vars:

  - GEMNASIUM_TESTSUITE: will be run for each iteration over update sets. This is typically your test suite script.
  - GEMNASIUM_BUNDLE_INSTALL_CMD: during each iteration, the new bundle will be installed. Default: "bundle install"
  - GEMNASIUM_BUNDLE_UPDATE_CMD: during each iteration, some gems might be updated. This command will be used. Default: "bundle update"
  - BRANCH: Current branch can be specified with this var, if the git command fails to run (git rev-parse --abbrev-ref HEAD).
  - REVISION: Current revision can be specified with this var, if the git command fails to run (git rev-parse --abbrev-ref HEAD)

  Examples:

  - GEMNASIUM_TESTSUITE="bundle exec rake" gemnasium autoupdate your_project_slug
  - cat script.sh | gemnasium autoupdate your_project_slug
  - gemnasium autoupdate my_project_slug bundle exec rake
  `,
			Action: func(ctx *cli.Context) {
				AttemptLogin(ctx, config)
				if !ctx.Args().Present() {
					cli.ShowCommandHelp(ctx, "autoupdate")
					os.Exit(1)
				}
				projectSlug := ctx.Args().First()
				testSuite := ctx.Args().Tail()
				err := AutoUpdate(projectSlug, testSuite, config)
				if err != nil {
					printFatal(err.Error())
					os.Exit(1)
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
