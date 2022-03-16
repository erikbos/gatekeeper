VERSION := $(shell git describe --tags --always --dirty)
BUILDTAG := $(shell git rev-parse HEAD)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDER := $(shell echo "`git config user.name` <`git config user.email`>")
BIN = bin

all: managementserver accesslogserver authserver controlplane testbackend

managementserver:
	mkdir -p $(BIN)
	go build -o $(BIN)/managementserver cmd/managementserver/*.go

accesslogserver:
	mkdir -p $(BIN)
	go build -o $(BIN)/accesslogserver cmd/accesslogserver/*.go

authserver:
	mkdir -p $(BIN)
	go build -o $(BIN)/authserver cmd/authserver/*.go

controlplane:
	mkdir -p $(BIN)
	go build -o $(BIN)/controlplane cmd/controlplane/*.go

testbackend:
	mkdir -p $(BIN)
	go build -o $(BIN)/testbackend cmd/testbackend/*.go

docker-images: docker-baseimage \
				docker-managementserver \
				docker-accesslogserver \
				docker-authserver \
				docker-controlplane \
				docker-testbackend \
				docker-dbadmin-test

docker-baseimage:
	 docker build -f build/Dockerfile.baseimage . -t gatekeeper/baseimage:latest

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

docker-dbadmin-test:
	cd tests/dbadmin && \
		docker build . -t gatekeeper/dbadmin-test:$(VERSION)

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

.PHONY: minikube-helm-install
minikube-helm-install: docker-images minikube-load-images
	kubectl create namespace gatekeeper
	helm dep up ./helm
	helm install gatekeeper ./helm --wait --namespace gatekeeper -f helm/values.yaml \
		--set global.e2e=true \
		--set global.images.tag=$(VERSION) \
		--set cassandra.persistance='{"enabled":"false"}'

.PHONY: minikube-inplace-docker-upgrade
minikube-inplace-docker-upgrade: test docker-images minikube-load-images minikube-helm-upgrade

.PHONY: minikube-helm-upgrade
minikube-helm-upgrade:
	@helm upgrade gatekeeper helm --wait -f helm/values.yaml -n gatekeeper \
		--set global.e2e=true \
		--set cassandra.dbUser.password=`kubectl get secret -n gatekeeper gatekeeper-cassandra -o=jsonpath='{.data.cassandra-password}' | base64 -D` \
		--set global.images.tag=$(VERSION) \
		--set cassandra.persistance='{"enabled":"false"}'

.PHONY: minikube-helm-diff
minikube-helm-diff:
	@helm diff gatekeeper helm -f helm/values.yaml -n gatekeeper \
		--set global.e2e=true \
		--set cassandra.dbUser.password=`kubectl get secret -n gatekeeper gatekeeper-cassandra -o=jsonpath='{.data.cassandra-password}' | base64 -D` \
		--set global.images.tag=$(VERSION) \
		--set cassandra.persistance='{"enabled":"false"}'

.PHONY: minikube-helm-test
minikube-helm-test:
	helm test -n gatekeeper gatekeeper

.PHONY: e2e-local
e2e-local: test docker-images minikube-start minikube-helm-install minikube-helm-test minikube-stop

.PHONY: e2e-actions
e2e-actions: test docker-images minikube-helm-install minikube-helm-test

.PHONY: minikube-start
minikube-start:
	minikube start --cpus='4'

.PHONY: minikube-stop
minikube-stop:
	minikube delete

.PHONY: minikube-load-images
minikube-load-images:
	@for image in gatekeeper/testbackend gatekeeper/controlplane gatekeeper/authserver gatekeeper/accesslogserver gatekeeper/managementserver gatekeeper/dbadmin-test; do \
		echo Loading $$image:$(VERSION) to Minikube ; \
		minikube image load $$image:$(VERSION) ; \
	done
