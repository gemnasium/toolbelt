FROM debian:wheezy

RUN echo "deb http://ftp.debian.org/debian wheezy-backports main" >> /etc/apt/sources.list.d/backports.list
RUN apt-get update && apt-get install -y debhelper build-essential git
RUN apt-get install -y -t wheezy-backports golang-go
RUN mkdir /go
ENV GOPATH /go
RUN go get github.com/tools/godep

COPY docker/build.sh /bin/build.sh
COPY docker/gpg_wrapper.sh /bin/gpg_wrapper.sh
COPY docker/test.sh /bin/test.sh

WORKDIR /go/src/github.com/gemnasium/toolbelt

