GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

BINARY_NAME = mosaic

all: clean test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME) -v

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME).exe -v

help:
	@echo "Available targets:"
	@echo "  make          - Default target (clean, test, build)"
	@echo "  make clean    - Clean the workspace"
	@echo "  make build    - Build the binary"
	@echo "  make build-linux  - Build for Linux"
	@echo "  make build-windows - Build for Windows"
	@echo "  make help     - Show this help message"
