FROM docker:5000/go:1.2.1

ADD . /go/src/git.tech-angels.net/gemnasium/toolbelt
# Fetch deps
# Make toolbelt source the default working directory
WORKDIR /go/src/git.tech-angels.net/gemnasium/toolbelt
RUN go get
VOLUME /go/src/git.tech-angels.net/gemnasium/toolbelt
