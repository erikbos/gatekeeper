VERSION := $(shell git describe --tags --always --dirty)
BUILDTAG := $(shell git rev-parse HEAD)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDER := $(shell echo "`git config user.name` <`git config user.email`>")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILDTIME)"
BIN = bin

all: dbadmin envoyauth envoycp testbackend oauthserver

dbadmin:
	mkdir -p $(BIN)
	go build -o $(BIN)/dbadmin $(LDFLAGS) cmd/dbadmin/*.go

envoyauth:
	mkdir -p $(BIN)
	go build -o $(BIN)/envoyauth $(LDFLAGS) cmd/envoyauth/*.go

envoycp:
	mkdir -p $(BIN)
	go build -o $(BIN)/envoycp $(LDFLAGS) cmd/envoycp/*.go

testbackend:
	mkdir -p $(BIN)
	go build -o $(BIN)/testbackend $(LDFLAGS) cmd/testbackend/*.go

oauthserver:
	mkdir -p $(BIN)
	go build -o $(BIN)/oauthserver $(LDFLAGS) cmd/oauthserver/*.go


docker-images: docker-dbadmin docker-envoyauth docker-envoycp docker-testbackend docker-oauthserver

docker-dbadmin:
	 docker build -t gatekeeper/dbadmin:$(VERSION) . -f Dockerfile.dbadmin

docker-envoyauth:
	 docker build -t gatekeeper/envoyauth:$(VERSION) . -f Dockerfile.envoyauth

docker-envoycp:
	 docker build -t gatekeeper/envoycp:$(VERSION) . -f Dockerfile.envoycp

docker-testbackend:
	 docker build -t gatekeeper/testbackend:$(VERSION) . -f Dockerfile.testbackend

docker-oauthserver:
	 docker build -t gatekeeper/oauthserver:$(VERSION) . -f Dockerfile.oauthserver

clean:
	rm -f $(BIN)/dbadmin $(BIN)/envoyauth $(BIN)/envoycp $(BIN)/testbackend $(BIN)/oauthserver
