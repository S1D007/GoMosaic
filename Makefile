GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

GOFLAGS ?= -ldflags="-s -w" 

BUILD_LINUX = CGO_ENABLED=0 GOOS=linux GOARCH=amd64
BUILD_WINDOWS = CGO_ENABLED=0 GOOS=windows GOARCH=amd64
BUILD_MAC = CGO_ENABLED=0 GOOS=darwin GOARCH=amd64

BINARY_NAME = mosaic
VERSION ?= latest

.PHONY: all build build-linux build-windows build-mac clean test deps help

all: clean test build

build:
	$(GOBUILD) $(GOFLAGS) -o $(BINARY_NAME) -v -ldflags="-X main.Version=$(VERSION)" ./...

build-linux:
	$(BUILD_LINUX) $(GOBUILD) $(GOFLAGS) -o $(BINARY_NAME) -v -ldflags="-X main.Version=$(VERSION)" ./...

build-windows:
	$(BUILD_WINDOWS) $(GOBUILD) $(GOFLAGS) -o $(BINARY_NAME).exe -v -ldflags="-X main.Version=$(VERSION)" ./...

build-mac:
	$(BUILD_MAC) $(GOBUILD) $(GOFLAGS) -o $(BINARY_NAME) -v -ldflags="-X main.Version=$(VERSION)" ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe

test:
	$(GOTEST) -v ./...

deps:
	$(GOGET) -u ./...

help:
	@echo "Available targets:"
	@echo "  make          - Default target (clean, test, build)"
	@echo "  make clean    - Clean the workspace"
	@echo "  make build    - Build the binary (with default options)"
	@echo "  make build-linux  - Build for Linux"
	@echo "  make build-windows - Build for Windows"
	@echo "  make build-mac - Build for macOS"
	@echo "  make test     - Run tests"
	@echo "  make deps     - Install/update dependencies"
	@echo "  make help     - Show this help message"