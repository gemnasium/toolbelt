package commands

import (
	"github.com/gemnasium/toolbelt/auth"
	"github.com/gemnasium/toolbelt/config"
	"github.com/urfave/cli"
	"github.com/gemnasium/toolbelt/api"
	"fmt"
	"os"
)

func App() *cli.App {
	app := cli.NewApp()
	app.Name = "gemnasium"
	app.Usage = "Gemnasium toolbelt"
	app.Version = config.VERSION
	app.Author = "Gemnasium"
	app.Email = "support@gemnasium.com"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "token, t",
			Usage: "Your api token (available in your account page)",
		},
		cli.BoolFlag{
			Name:  "raw, r",
			Usage: "Raw format output",
		},
		cli.IntFlag{
			Name:  "api-version",
			Usage: "API version to use (default: autodetected)",
		},
	}
	app.Before = func(c *cli.Context) error {
		config.RawFormat = c.Bool("raw")
		config.APIVersion = c.Int("api-version")
		if config.APIVersion == 0 {
			// Set API version if it was not set by parameters
			if config.APIEndpoint == config.DEFAULT_API_ENDPOINT {
				config.APIVersion = 1
			} else {
				config.APIVersion = 2
				if !config.RawFormat {
					fmt.Printf("Using API v2 for endpoint %s.\n", config.APIEndpoint)
				}
			}
		}
		switch config.APIVersion {
		case 1:
			api.APIImpl = api.NewAPIv1(config.APIEndpoint, config.APIKey)
		case 2:
			api.APIImpl = &api.V2ToV1{api.NewAPIv2(config.APIEndpoint, config.APIKey)}
		default:
			fmt.Fprintf(os.Stderr, "Unknown API version: %d", config.APIVersion)
		}

		return nil
	}
	app.Commands = []cli.Command{
		{
			Name:  "auth",
			Usage: "Authentication",
			Subcommands: []cli.Command{
				{
					Name:   "login",
					Usage:  "Login",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "with-api-token",
							Usage: "Log in with your API token (API key in the user profile)",
						},
					},
					Action: Login,
				},
				{
					Name:   "logout",
					Usage:  "Logout",
					Action: Logout,
				},
			},
		},
		{
			Name:        "configure",
			Usage:       "Install configuration for an existing project",
			Description: "Will create a .gemnasium.yml file in the current directory. This file will be parse if present.\n   Warning: this command will overwrite existing .gemnasium.yml file.\n\n   Arguments: project_slug (the identifier of the project).",
			Action:      Configure,
		},
		{
			Name:      "projects",
			ShortName: "p",
			Usage:     "Manage current project",
			Before:    auth.ConfigureAPIToken,
			Subcommands: []cli.Command{
				{
					Name:      "list",
					ShortName: "l",
					Usage:     "List projects on Gemnasium",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "private, p",
							Usage: "Display only private projects",
						},
					},
					Action: ProjectsList,
				},
				{
					Name:      "show",
					ShortName: "s",
					Usage:     "Show projet detail",
					Action:    ProjectsShow,
				},
				{
					Name:      "update",
					ShortName: "u",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name, n",
							Usage: "Project name",
						},
						cli.StringFlag{
							Name:  "desc, d",
							Usage: "A short description",
						},
						cli.BoolFlag{
							Name:  "monitored, m",
							Usage: "Whether the project is watched by the user.",
						},
					},
					Usage:  "Edit project details",
					Action: ProjectsUpdate,
				},
				{
					Name:      "create",
					ShortName: "c",
					Usage:     "Create a new project on Gemnasium",
					Action:    ProjectsCreate,
				},
				{
					Name:   "sync",
					Usage:  "Start project synchronization",
					Action: ProjectsSync,
				},
			},
		},
		{
			Name:      "dependencies",
			ShortName: "d",
			Usage:     "Dependencies",
			Before: func(ctx *cli.Context) error {
				auth.ConfigureAPIToken(ctx)
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:      "list",
					ShortName: "l",
					Usage:     "List the first level dependencies of the requested project. Usage: gemnasium dependencies list [project_slug]",
					Action:    DependenciesList,
				},
			},
		},
		{
			Name:      "dependency_files",
			ShortName: "df",
			Usage:     "Dependency files",
			Before: func(ctx *cli.Context) error {
				auth.ConfigureAPIToken(ctx)
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:      "list",
					ShortName: "l",
					Usage:     "List dependency files for project",
					Action:    DependencyFilesList,
				},
				{
					Name:      "push",
					ShortName: "p",
					Usage:     "Push dependency files on Gemnasium",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "files, f",
							Usage: "list of files to send, separated with a comma.",
						},
					},
					Description: "Send files to Gemnasium. If --files is not set, all dependency files supported by Gemnasium found in the current path will be sent to Gemnasium API. You can ignore paths with GEMNASIUM_IGNORED_PATHS",
					Action:      DependenciesPush,
				},
			},
		},
		{
			Name:      "alerts",
			ShortName: "a",
			Usage:     "Dependency Alerts",
			Before: func(ctx *cli.Context) error {
				auth.ConfigureAPIToken(ctx)
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:      "list",
					ShortName: "l",
					Usage:     "List the dependency alerts the given project is affected by",
					Action:    DependencyAlertsList,
				},
			},
		},
		{
			Name:      "eval",
			ShortName: "e",
			Usage:     "Live deps evaluation",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "files, f",
					Usage: "list of files to evaluate, separated with a comma.",
				},
			},
			Action: LiveEvaluation,
		},
		{
			Name:      "autoupdate",
			ShortName: "au",
			Usage:     "Auto-update the dependency files of the project",
			Before:    auth.ConfigureAPIToken,
			Subcommands: []cli.Command{
				{
					Name:      "run",
					ShortName: "r",
					Usage:     "Run the auto-update",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "project, p",
							Usage: "Project slug (identifier on Gemnasium)",
						},
					},
					Description: `Auto-Update will fetch update sets from Gemnasium and run your test suite against them.
   The test suite can be passed as arguments, or through the env var GEMNASIUM_TESTSUITE.

   Arguments:

   - [test suite commands] (string): Commands to run your test suite (ex: "./test.sh")

   Env Vars:

   - GEMNASIUM_PROJECT_SLUG: override --project flag and project_slug in .gemnasium.yyml.
   - GEMNASIUM_TESTSUITE: will be run for each iteration over update sets. This is typically your test suite script.
   - GEMNASIUM_BUNDLE_INSTALL_CMD: [Ruby Only] during each iteration, the new bundle will be installed. Default: "bundle install"
   - GEMNASIUM_BUNDLE_UPDATE_CMD: [Ruby Only] during each iteration, some gems might be updated. This command will be used. Default: "bundle update"
   - BRANCH: Current branch can be specified with this var, if the git command fails to run (git rev-parse --abbrev-ref HEAD).
   - REVISION: Current revision can be specified with this var, if the git command fails to run (git rev-parse --abbrev-ref HEAD)

   Examples:

   - GEMNASIUM_TESTSUITE="bundle exec rake" GEMNASIUM_PROJECT_SLUG=a907c0f9b8e0b89f23f0042d76ae0358 gemnasium autoupdate
   - cat script.sh | gemnasium autoupdate -p=your_project_slug
   - gemnasium autoupdate my_project_slug bundle exec rake
  `,
					Action: AutoUpdateRun,
				},
				{
					Name:      "apply",
					ShortName: "a",
					Usage:     "Apply the best update",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "project, p",
							Usage: "Project slug (identifier on Gemnasium)",
						},
					},
					Description: `Update the dependency files to match the best update that has been found so far.`,
					Action:      AutoUpdateApply,
				},
			},
		},
		{
			Name:   "env",
			Usage:  "Display ENV vars used by gemnasium",
			Action: DisplayEnvVars,
		},
	}
	return app
}
