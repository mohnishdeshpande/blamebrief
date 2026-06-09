# Makefile for BlameBrief CLI

BINARY_NAME=blamebrief
GO_FILES=$(shell find . -type f -name '*.go')

.PHONY: all build test install clean fmt vet help

all: build

help:
	@echo "BlameBrief CLI - Available commands:"
	@echo "  make build      - Build the blamebrief binary"
	@echo "  make test       - Run all unit tests"
	@echo "  make install    - Install the binary to Go's bin directory"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make fmt        - Format all Go source files"
	@echo "  make vet        - Run go vet static analysis"

build: $(BINARY_NAME)

$(BINARY_NAME): $(GO_FILES)
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME)

test:
	@echo "Running tests..."
	go test -v ./...

install: build
	@echo "Installing $(BINARY_NAME) via go install..."
	go install

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)

fmt:
	@echo "Formatting source files..."
	go fmt ./...

vet:
	@echo "Running go vet..."
	go vet ./...
