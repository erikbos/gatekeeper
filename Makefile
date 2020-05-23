VERSION := $(shell git describe --tags --always --dirty)
BUILDTAG := $(shell git rev-parse HEAD)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDER := $(shell echo "`git config user.name` <`git config user.email`>")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILDTIME)"
BIN = bin

all: dbadmin envoyauth envoycp testbackend

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


docker-images: docker-dbadmin docker-envoyauth docker-envoycp docker-testbackend

docker-dbadmin:
	 docker build -t apiedge/dbadmin:$(VERSION) . -f Dockerfile.dbadmin

docker-envoyauth:
	 docker build -t apiedge/envoyauth:$(VERSION) . -f Dockerfile.envoyauth

docker-envoycp:
	 docker build -t apiedge/envoycp:$(VERSION) . -f Dockerfile.envoycp

docker-testbackend:
	 docker build -t apiedge/testbackend:$(VERSION) . -f Dockerfile.testbackend

clean:
	rm -f $(BIN)/dbadmin $(BIN)/envoyauth $(BIN)/envoycp $(BIN)/testbackend
