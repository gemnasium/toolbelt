#!/bin/sh

exec gpg --digest-algo SHA512 $@
