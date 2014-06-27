# Gemnasium Toolbelt

[![Build Status](https://travis-ci.org/gemnasium/toolbelt.svg?branch=master)](https://travis-ci.org/gemnasium/toolbelt)

Gemnasium toolbelt is a CLI for the Gemnasium API.

## How to install it ?

### On Mac OS X

A Homebrew formula is available for you to use. Just run

    brew tap gemnasium/gemnasium

And then

    brew install gemnasium-toolbelt

If you don't want to use Homebrew, the executable is available under the directory `builds/macosx`.  
Or you can build it yourself by running:

    go build -o gemnasium

### On Linux

We built the package for Linux. It is available under the directory `builds/linux`.

Alternatively, you can build it yourself by running:

    go build -o gemnasium

### Binaries

Binaries are available in the [releases](https://github.com/gemnasium/toolbelt/releases) page.

## How to use it ?

### Authentication

Gemnasium Toolbelt stores your Gemnasium API key into your .netrc file.

To be logged in to Gemnasium, you need to run `gemnasium auth login` and enter your Gemnasium credentials.

Alternatively, you can pass directly your API token to all commands with the option `--token` or the env var API_KEY.
Your API token is available in your settings page (https://gemnasium.com/settings).

### Create a new project

To create a new project on Gemnasium, you need to `cd` into your project directory and run

    gemnasium projects create

### Configure an existing project

If your project is already on Gemnasium, you need to `cd` into your project directory and run

    gemnasium projects configure

You will need your project's Slug (available in your project page settings).
A sample configuration file is available here: https://github.com/gemnasium/toolbelt/blob/master/config/gemnasium.yml.example 

### Live Evaluation

If you want to evaluate your project without pushing files or pulling info from Gemnasium, you may use the ```eval``` command:

    gemnasium eval -f=Gemfile,Gemfile.lock

The command will exit with a code 1 if the project global status is "red".

(Needs a paid plan)

### Auto Update

Auto-Update will fetch update sets from Gemnasium and run your test suite against them.
The test suite can be passed as arguments, or through the env var GEMNASIUM_TESTSUITE.

 Examples:

    GEMNASIUM_TESTSUITE="bundle exec rake" PROJECT_SLUG=a907c0f9b8e0b89f23f0042d76ae0358 gemnasium autoupdate

    cat script.sh | gemnasium autoupdate -p=your_project_slug

    gemnasium autoupdate my_project_slug bundle exec rake

Typically, this command is to be used with a CI server, along with nightly builds. 
Although Gemnasium will optimize as much as possible the number of combinasions, the number of iterations isn't predictable, and your test suite might be running for a long time.
To avoid looping to death, the command will stop looping after 1 hour and exit.
As soon as a valid update set is found, the loop will stop, and Gemnasium is notified. A patch will be available to download a few seconds later.
We will propose soon an option to open Pull Requests directly on GitHub.

Currently, only Ruby projects are supported. Follow us to get the latest updates: https://twitter.com/gemnasiumapp

(Needs a paid plan)

## Configuration

The configuration can be saved in ```.gemnasium.yml``` files in the project directory.
Options set in ```.gemnasium.yml``` are overriden by env vars:


 * **PROJECT_SLUG**: override -project flag and project_slug in .gemnasium.yml.
 * **GEMNASIUM_TESTSUITE**: will be run for each iteration over update sets. This is typically your test suite script.
 * **GEMNASIUM_BUNDLE_INSTALL_CMD**: [Ruby Only] during each iteration, the new bundle will be installed. Default: "bundle install"
 * **GEMNASIUM_BUNDLE_UPDATE_CMD**: [Ruby Only] during each iteration, some gems might be updated. This command will be used. Default: "bundle update"
 * **BRANCH**: Current branch can be specified with this var, if the git command fails to run (git rev-parse --abbrev-ref HEAD).
 * **REVISION**: Current revision can be specified with this var, if the git command fails to run (git rev-parse --abbrev-ref HEAD)
 * **API_KEY**: Your API private key (available in your account settings https://gemnasium.com/settings)
 * **IGNORED_PATHS**: A list of paths separated by "," where dependency files are ignored.
 * **RAW_FORMAT**: Display API raw json output (for debug)

 and env vars are overriden by command line options.
 Ex: 

    echo 'project_slug: tic' > .gemnasium.yml ; PROJECT_SLUG="tac" gemnasium projects show toe
    => [toe project details]

### Need further help ?

A full commands documentation is available by running

    gemnasium [command] --help