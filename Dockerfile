FROM docker:5000/go:1.3

ADD . /go/src/github.com/gemnasium/toolbelt
# Fetch deps
# Make toolbelt source the default working directory
WORKDIR /go/src/github.com/gemnasium/toolbelt
RUN go get
VOLUME /go/src/github.com/gemnasium/toolbelt
