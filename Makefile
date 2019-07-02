MODULE			:= github.com/kbariotis/go-discover
LDFLAGS		:= -w -s
GOBIN			:= $(CURDIR)/bin
PATH			:= $(GOBIN):$(PATH)
CRAWLER_NAME		:= crawler
API_NAME		:= api
EXTRACTION_NAME	:= extraction
VERSION		:= unknown

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
all: deps test build-api build-crawler

# Download dependencies
.PHONY: deps
deps:
	$(info Installing dependencies)
	@go mod download

# Generate mocks
.PHONY: mocks
mocks:
	$(info Generating mocks)
	@go generate ./...

# Vendor dependencies
.PHONY: vendor
vendor: deps
	$(info Vendoring dependencies)
	@go mod vendor

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
.PHONY: build-crawler
build-crawler: deps
build-crawler: LDFLAGS += -X $(MODULE)/internal/version.Timestamp=$(shell date +%s)
build-crawler: LDFLAGS += -X $(MODULE)/internal/version.Version=${VERSION}
build-crawler: LDFLAGS += -X $(MODULE)/internal/version.GitSHA=${GIT_SHA}
build-crawler: LDFLAGS += -X $(MODULE)/internal/version.ServiceName=${CRAWLER_NAME}
build-crawler:
	$(info building binary to bin/$(CRAWLER_NAME))
	@CGO_ENABLED=0 go build -o bin/$(CRAWLER_NAME) -installsuffix cgo -ldflags '$(LDFLAGS)' ./cmd/$(CRAWLER_NAME)

.PHONY: build-extraction
build-extraction: deps
build-extraction: LDFLAGS += -X $(MODULE)/internal/version.Timestamp=$(shell date +%s)
build-extraction: LDFLAGS += -X $(MODULE)/internal/version.Version=${VERSION}
build-extraction: LDFLAGS += -X $(MODULE)/internal/version.GitSHA=${GIT_SHA}
build-extraction: LDFLAGS += -X $(MODULE)/internal/version.ServiceName=${EXTRACTION_NAME}
build-extraction:
	$(info building binary to bin/$(EXTRACTION_NAME))
	@CGO_ENABLED=0 go build -o bin/$(EXTRACTION_NAME) -installsuffix cgo -ldflags '$(LDFLAGS)' ./cmd/$(EXTRACTION_NAME)

# Builds binaries
.PHONY: build-api
build-api: deps
build-api: LDFLAGS += -X $(MODULE)/internal/version.Timestamp=$(shell date +%s)
build-api: LDFLAGS += -X $(MODULE)/internal/version.Version=${VERSION}
build-api: LDFLAGS += -X $(MODULE)/internal/version.GitSHA=${GIT_SHA}
build-api: LDFLAGS += -X $(MODULE)/internal/version.ServiceName=${API_NAME}
build-api:
	$(info building binary to bin/$(API_NAME))
	@CGO_ENABLED=0 go build -o bin/$(API_NAME) -installsuffix cgo -ldflags '$(LDFLAGS)' ./cmd/$(API_NAME)

# Builds and runs the binary with debug logging
.PHONY: run-crawler
run-crawler: build-crawler
	@LOG_LEVEL=debug ./bin/$(CRAWLER_NAME)

.PHONY: run-extraction
run-extraction: build-extraction
	@LOG_LEVEL=debug ./bin/$(EXTRACTION_NAME)

.PHONY: run-api
run-api: build-api
	@LOG_LEVEL=debug ./bin/$(API_NAME)

# Build and runs docker-compose
.PHONY: docker-compose
docker-compose: vendor
	$(info Running docker-compose)
	docker-compose stop
	docker-compose rm -f
	docker-compose down
	docker-compose build
	docker-compose run --service-ports --rm api extraction crawler

# Run test suite
.PHONY: test
test: tools
	$(info Running tests)
	go test $(V) -count=1 --race -covermode=atomic ./...

# Clean temp things
.PHONY: clean-api
clean-api:
	@rm bin/$(API_NAME)

.PHONY: clean-crawler
clean-crawler:
	@rm bin/$(CRAWLER_NAME)

.PHONY: clean-extraction
clean-extraction:
	@rm bin/$(EXTRACTION_NAME)
