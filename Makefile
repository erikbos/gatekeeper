VERSION := $(shell git describe --tags --always --dirty)
BUILDTAG := $(shell git rev-parse HEAD)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDER := $(shell echo "`git config user.name` <`git config user.email`>")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILDTIME)"
BIN = bin

all: managementserver accesslogserver authserver controlplane testbackend

managementserver:
	mkdir -p $(BIN)
	go build -o $(BIN)/managementserver $(LDFLAGS) cmd/managementserver/*.go

accesslogserver:
	mkdir -p $(BIN)
	go build -o $(BIN)/accesslogserver $(LDFLAGS) cmd/accesslogserver/*.go

authserver:
	mkdir -p $(BIN)
	go build -o $(BIN)/authserver $(LDFLAGS) cmd/authserver/*.go

controlplane:
	mkdir -p $(BIN)
	go build -o $(BIN)/controlplane $(LDFLAGS) cmd/controlplane/*.go

testbackend:
	mkdir -p $(BIN)
	go build -o $(BIN)/testbackend $(LDFLAGS) cmd/testbackend/*.go


docker-images: docker-baseimage \
				docker-managementserver \
				docker-accesslogserver \
				docker-authserver \
				docker-controlplane \
				docker-testbackend

docker-baseimage:
	 docker build -f build/Dockerfile.baseimage . -t gatekeeper/baseimage

docker-managementserver:
	 docker build -f build/Dockerfile.managementserver . -t gatekeeper/managementserver:$(VERSION) -t gatekeeper/managementserver:latest

docker-accesslogserver:
	 docker build -f build/Dockerfile.accesslogserver . -t gatekeeper/accesslogserver:$(VERSION) -t gatekeeper/accesslogserver:latest

docker-authserver:
	 docker build -f build/Dockerfile.authserver . -t gatekeeper/authserver:$(VERSION) -t gatekeeper/authserver:latest

docker-controlplane:
	 docker build -f build/Dockerfile.controlplane . -t gatekeeper/controlplane:$(VERSION) -t gatekeeper/controlplane:latest

docker-testbackend:
	 docker build -f  build/Dockerfile.testbackend . -t gatekeeper/testbackend:$(VERSION) -t gatekeeper/testbackend:latest

.PHONY: test
test:
	mkdir -p tmp
	go test -v -coverpkg=./... -covermode=atomic -coverprofile=tmp/coverage.txt ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: clean
clean:
	rm -f $(BIN)/managementserver $(BIN)/authserver $(BIN)/controlplane $(BIN)/accesslogserver $(BIN)/testbackend
