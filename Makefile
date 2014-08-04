GOOS="linux"

TEMPDIR := $(shell mktemp -d)
BUILDDIR := ${shell pwd}/build
TOOLBELTDIR := ${BUILDDIR}/src/github.com/gemnasium/toolbelt

all: build/gemnasium

build/gemnasium:
	cp -r `pwd` ${TEMPDIR}/
	mkdir -p ${BUILDDIR}/src/github.com/gemnasium
	rm -rf ${TOOLBELTDIR}
	cp -r $(TEMPDIR)/toolbelt ${BUILDDIR}/src/github.com/gemnasium/
	cd ${TOOLBELTDIR} && GOPATH=${BUILDDIR} go get -t
	cd ${TOOLBELTDIR} && GOPATH=${BUILDDIR} GOOS=${GOOS} go build -o ${BUILDDIR}/gemnasium
	rm -rf ${TEMPDIR}	
