PKGNAME    := sensors
GOPKG      := github.com/rclancey/sensors
GITHASH    := $(shell git rev-parse --short HEAD)
FULLBRANCH := $(shell git branch --show-current)
TAGNAME    := $(shell git describe --exact-match --tags $(GITHASH) 2>/dev/null)
BRANCHNAME := $(shell basename "$(FULLBRANCH)")
DATE       := $(shell date '+%Y%m%d')
GITVERSION := $(shell sh -c 'if [ "x$(TAGNAME)" = "x" ] ; then echo $(GITHASH) ; else echo $(TAGNAME) ; fi')
VERSION    ?= $(GITVERSION)
BUILDTIME  := $(shell date '+%s')
TARGET     ?= local
BUILDDIR   = build
TARFILE    = $(PKGNAME)-$(VERSION).tar.gz

CGO_ENABLED = 1
ifeq "$(TARGET)" "macos"
GOOS = darwin
GOARCH = amd64
GOARM =
BUILDDIR = build-macos
TARFILE = $(PKGNAME)-$(TARGET)-$(VERSION).tar.gz
endif
ifeq "$(TARGET)" "rpi"
GOOS = linux
GOARCH = arm
GOARM = 6
BUILDDIR = build-rpi
TARFILE = $(PKGNAME)-$(TARGET)-$(VERSION).tar.gz
#CC = x86_64-linux-musl-gcc
#CXX = x86_64-linux-musl-g++
CC=gcc
CXX=g++
CGO_ENABLED = 1
LDFLAGS = -linkmode external -extldflags -static
endif

BUILDFLAGS = -X $(GOPKG)/api.BuildDate=$(BUILDTIME) -X $(GOPKG)/api.Commit=$(GITHASH) -X $(GOPKG)/api.Branch=$(FULLBRANCH)
ifeq "$(TAGNAME)" ""
	ALLBUILDFLAGS = $(BUILDFLAGS)
else
	ALLBUILDFLAGS= $(BUILDFLAGS) -X $(GOPKG)/api.Version=$(TAGNAME)
endif

GOSRC := $(shell find * -type f -name "*.go")

all: compile

$(BUILDDIR)/$(PKGNAME)/bin/%: $(GOSRC) go.mod go.sum
	mkdir -p $(BUILDDIR)/$(PKGNAME)/bin
	env CC=$(CC) CXX=$(CXX) GOARCH=$(GOARCH) GOOS=$(GOOS) CGO_ENABLED=$(CGO_ENABLED) go build -ldflags "$(LDFLAGS) $(ALLBUILDFLAGS)" -o $@ cmd/$*.go

go-compile: $(BUILDDIR)/$(PKGNAME)/bin/sensors

.PHONY: go-compile

compile: go-compile

.PHONY: compile

$(BUILDDIR)/$(TARFILE): compile
	cd $(BUILDDIR) && tar -zcf $(TARFILE) $(PKGNAME)

dist: $(BUILDDIR)/$(TARFILE)

.PHONY: dist

clean:
	rm -rf $(BUILDDIR)
