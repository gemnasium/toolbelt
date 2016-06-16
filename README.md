# Gemnasium Toolbelt

[![Build Status](https://travis-ci.org/gemnasium/toolbelt.svg?branch=master)](https://travis-ci.org/gemnasium/toolbelt)

Gemnasium toolbelt is a CLI for the Gemnasium API.

## How to install it?

### On Mac OS X

A Homebrew formula is available for you to use. Just run

    brew tap gemnasium/gemnasium

And then

    brew install gemnasium-toolbelt

### Ubuntu and Debian

#### Configure Gemnasium repository

    sudo sh -c 'echo "deb http://apt.gemnasium.com stable main" > /etc/apt/sources.list.d/gemnasium.list'
    
#### Trust Gemnasium GPG key

    sudo apt-key adv --recv-keys --keyserver keyserver.ubuntu.com E5CEAB0AC5F1CA2A
    
#### Update package list

    sudo apt-get update
    
#### Install Gemnasium toolbelt

    sudo apt-get install gemnasium-toolbelt

The ```gemnasium``` command will be available in ```/usr/bin/gemnasium```

### From source

    go build -o gemnasium

(or ```gemnasium.exe``` for windows users)

### Binaries

Binaries are available in the [releases](https://github.com/gemnasium/toolbelt/releases) page.

## How to use it?

### Authentication

Gemnasium Toolbelt stores your Gemnasium API key into your .netrc file.

To be logged in to Gemnasium, you need to run `gemnasium auth login` and enter your Gemnasium credentials.

Alternatively, you can pass directly your API token to all commands with the option `--token` or the env var ```GEMNASIUM_TOKEN```.
Your API token is available in your settings page (https://gemnasium.com/settings).

### Create a new project

To create a new project on Gemnasium, you need to `cd` into your project directory and run

    gemnasium projects create

### Configure an existing project

If your project is already on Gemnasium, you need to `cd` into your project directory and run

    gemnasium configure [project_slug]

You will need your project's Slug (available in your project page settings).
A sample configuration file is available here: https://github.com/gemnasium/toolbelt/blob/master/config/gemnasium.yml.example 

### Push dependency files

For projects not automatically synced with Github or Gitlab, you may want to push your files directly to Gemnasium.
The corresponding project will updated soon after the files have been received. To push your files

    gemnasium dependency_files push -f=Gemfile,Gemfile.lock


### Live Evaluation

If you want to evaluate your project without pushing files or pulling info from Gemnasium, you may use the ```eval``` command:

    gemnasium eval -f=Gemfile,Gemfile.lock

The command will exit with a code 1 if the project global status is "red".

(Needs a paid plan)

### Auto Update

Auto-Update will fetch update sets from Gemnasium and run your test suite against them.
The test suite can be passed as arguments, or through the env var GEMNASIUM_TESTSUITE.

 Examples:

    GEMNASIUM_TESTSUITE="bundle exec rake" GEMNASIUM_PROJECT_SLUG=a907c0f9b8e0b89f23f0042d76ae0358 gemnasium autoupdate

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


 * **GEMNASIUM_PROJECT_SLUG**: override -project flag and project_slug in .gemnasium.yml.
 * **GEMNASIUM_TESTSUITE**: will be run for each iteration over update sets. This is typically your test suite script.
 * **GEMNASIUM_BUNDLE_INSTALL_CMD**: [Ruby Only] during each iteration, the new bundle will be installed. Default: "bundle install"
 * **GEMNASIUM_BUNDLE_UPDATE_CMD**: [Ruby Only] during each iteration, some gems might be updated. This command will be used. Default: "bundle update"
 * **BRANCH**: Current branch can be specified with this var, if the git command fails to run (git rev-parse --abbrev-ref HEAD).
 * **REVISION**: Current revision can be specified with this var, if the git command fails to run (git rev-parse --abbrev-ref HEAD)
 * **GEMNASIUM_TOKEN**: Your API private token (available in your account settings https://gemnasium.com/settings)
 * **GEMNASIUM_IGNORED_PATHS**: A list of paths separated by "," where dependency files are ignored.
 * **GEMNASIUM_RAW_FORMAT**: Display API raw json output (for debug)
 * **NETRC_PATH**: Location of your .netrc file (default: ~/.netrc)

 and env vars are overriden by command line options.
 Ex: 

```
echo 'project_slug: tic' > .gemnasium.yml ; GEMNASIUM_PROJECT_SLUG="tac" gemnasium projects show toe
=> [toe project details]
```

To obtain the list of env vars used and set:

    gemnasium env

### Need further help?

A full commands documentation is available by running

    gemnasium [command] --help
