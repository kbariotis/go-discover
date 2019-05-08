MODULE   := github.com/kbariotis/go-discover
LDFLAGS  := -w -s
GOBIN    := $(CURDIR)/bin
PATH     := $(GOBIN):$(PATH)
NAME     := discover
VERSION  := unknown

# Tools (will be installed in GOBIN)
TOOLS += github.com/mattn/goveralls
TOOLS += github.com/maxbrunsfeld/counterfeiter/v6
TOOLS += github.com/golangci/golangci-lint/cmd/golangci-lint

# Verbose output
ifdef VERBOSE
V = -v
endif

# Git Status
GIT_SHA ?= $(shell git rev-parse --short HEAD)

# Default target
.DEFAULT_GOAL := all

# Make All targets
.PHONY: all
all: deps test build

# Download dependencies
.PHONY: deps
deps:
	$(info Installing dependencies)
	@go mod download

# Install tools
.PHONY: tools
tools: deps $(TOOLS)

# Check tools
.PHONY: $(TOOLS)
$(TOOLS): %:
	GOBIN=$(GOBIN) go install $*

# Lint code
.PHONY: lint
lint: tools
	$(info Running linter)
	@golangci-lint -v run

# Builds binaries
.PHONY: build
build: deps
build: LDFLAGS += -X $(MODULE)/internal/version.Timestamp=$(shell date +%s)
build: LDFLAGS += -X $(MODULE)/internal/version.Version=${VERSION}
build: LDFLAGS += -X $(MODULE)/internal/version.GitSHA=${GIT_SHA}
build: LDFLAGS += -X $(MODULE)/internal/version.ServiceName=${NAME}
build:
	$(info building binary to bin/$(NAME))
	@CGO_ENABLED=0 go build -o bin/$(NAME) -installsuffix cgo -ldflags '$(LDFLAGS)' ./cmd/$(NAME)

# Builds and runs the binary with debug logging
.PHONY: run
run: build
	@LOG_LEVEL=debug ./bin/$(NAME)

# Run test suite
.PHONY: test
test: tools
	$(info Running tests)
	go test $(V) -count=1 --race -covermode=atomic ./...
 
# Clean temp things
.PHONY: clean
clean:
	@rm bin/$(NAME)
