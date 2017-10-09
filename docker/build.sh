#!/bin/sh

set -ex

# Import GPG signing key from environment
echo "$GPG_KEY" | base64 -di > /tmp/key.asc
cat /tmp/key.asc
gpg --import /tmp/key.asc

# Build binary and debian package
cp -r /src/* .
go get -t
go build -o /artifacts/gemnasium
dpkg-buildpackage -tc -k689FC23B -p/bin/gpg_wrapper.sh
mv ../gemnasium* /artifacts

