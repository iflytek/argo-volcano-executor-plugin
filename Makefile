.PHONY: build clean tool lint help

# These are the values we want to pass for Version and BuildTime
GITTAG=$(tag)
BUILD_TIME=`date +%FT%T%z`
GIT_COMMIT            := $(shell git rev-parse HEAD)
VERSION=v0.0.1
GIT_TAG               := $(shell git describe --exact-match --tags --abbrev=0  2> /dev/null || echo untagged)
GIT_TREE_STATE        := $(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)
OUTPUT_PATH           := bin

BIN_NAME=avp
# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags
ADDITIONAL_GO_LINKER_FLAGS = $(shell GOOS=$(shell go env GOHOSTOS) \
	GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/info-plist.go "$(VERSION)")

override LDFLAGS += "\
  -X argo-volcano-executor-plugin/server/main.BuildDate=${BUILD_TIME} \
  -X argo-volcano-executor-plugin/server/main.gitCommit=${GIT_COMMIT} \
  -X argo-volcano-executor-plugin/server/main.gitTag=${GIT_TAG} \
  -X argo-volcano-executor-plugin/server/main.version=${VERSION} \
  -X argo-volcano-executor-plugin/server/main.gitTreeState=${GIT_TREE_STATE}"


ifeq ($(OS),Windows_NT)
	ifeq ($(PROCESSOR_ARCHITEW6432),AMD64)
	ARCH=amd64
	endif
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Darwin)
	OS=mac
	endif
	ifeq ($(UNAME_S),Linux)
	OS=linux
	endif
endif


all: clean 	build-server-${OS}


build-server-linux:
	mkdir -p bin/linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${SERVER_LDFLAGS} -o bin/linux/$(BIN_NAME)_linux_amd64 -v ./server
#	cp -r config ./bin/linux/config

build-server-mac:
	mkdir -p bin/mac
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ${SERVER_LDFLAGS}  -o bin/mac/$(BIN_NAME)_mac -v ./server


clean:
	rm -rf bin runtime/logs
	go clean -i .
