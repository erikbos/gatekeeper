VERSION := $(shell git describe --tags --always --dirty)
BUILDTAG := $(shell git rev-parse HEAD)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDER := $(shell echo "`git config user.name` <`git config user.email`>")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILDTIME)"

all: dbadmin envoyauth envoycp

dbadmin:
	mkdir -p foo
	go build -o dbadmin $(LDFLAGS) cmd/dbadmin/*.go

envoyauth:
	mkdir -p foo
	go build -o envoycp $(LDFLAGS) cmd/envoyauth/*.go

envoycp:
	mkdir -p foo
	go build -o envoycp $(LDFLAGS) cmd/envoycp/*.go

