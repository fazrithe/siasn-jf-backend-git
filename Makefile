DOCKERCMD=docker

DOCKER_CONTAINER_NAME?=siasn-jf-backend
DOCKER_CONTAINER_IMAGE?=siasn-jf-backend:latest
DOCKER_BUILD_ARGS?=
DOCKER_DEBIAN_MIRROR?=http://deb.debian.org/debian

BUILD_DATE?=$(shell date -u +'%Y-%m-%dT00:00:00Z')
BUILD_VERSION?=0.1.0

TOPDIR=$(PWD)
BINARY=siasn-jf-backend

.FORCE:
.PHONY: build
.PHONY: vet
.PHONY: unit-test
.PHONY: generate
.PHONY: depend
.PHONY: docker-build
.PHONY: solr
.PHONY: clean
.PHONY: install
.PHONY: all
.PHONY: .FORCE

guard-%:
	@if [ -z '${${*}}' ]; then echo 'Environment variable $* not set' && exit 1; fi

build:
	@echo "Executing go build"
	go build -v -buildmode=pie -ldflags "-X main.version=$(BUILD_VERSION)"
	@echo "Binary ready"

vet:
	@echo "Running Go static code analysis with go vet"
	go vet -asmdecl -atomic -bool -buildtags -copylocks -httpresponse -loopclosure -lostcancel -methods -nilfunc -printf -rangeloops -shift -structtag -tests -unreachable -unsafeptr ./...
	@echo "go vet complete"

unit-test:
	@echo "Executing go unit test"
	go test -v -json -count=1 -parallel=4 ./...
	@echo "Unit test done"

generate:
	go generate ./...

depend:
	@echo "Pulling all Go dependencies"
	go mod download
	go mod verify
	go mod tidy
	@echo "You can now run 'make build' to compile all packages"

docker-build:
	$(DOCKERCMD) build -t $(DOCKER_CONTAINER_IMAGE) --build-arg GOPROXY=$(GOPROXY) --build-arg GOSUMDB=$(GOSUMDB) --build-arg BUILD_VERSION=$(BUILD_VERSION) $(DOCKER_BUILD_ARGS) .

default: depend

all: depend generate build unit-test

install: depend build

clean:
	rm -f $(BINARY)
	rm -f $(BINARY).exe

