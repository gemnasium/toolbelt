# Gemnasium Toolbelt

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

## How to use it ?

### Authentication

Gemnasium Toolbelt stores your Gemnasium API key into your .netrc file.

To be logged in to Gemnasium, you need to run `gemnasium auth login` and enter your Gemnasium credentials.

Alternatively, you can pass directly your API token to all commands with the option `--token`

### Create a new project

To create a new project on Gemnasium, you need to `cd` into your project directory and run

    gemnasium projects create

### Configure an existing project

If your project is already on Gemnasium, you need to `cd` into your project directory and run

    gemnasium projects configure

You will need your project's Slug (available in your project page settings).

### Need further help ?

A full commands documentation is available by running

    gemnasium [command] --help
