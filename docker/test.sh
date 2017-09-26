#!/bin/sh

set -ex

cp -r /src/* .
go get -t
go get github.com/BurntSushi/toml
go get gopkg.in/urfave/cli.v1
go get gopkg.in/yaml.v2
go test ./...

