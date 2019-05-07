MODULE   := github.com/kbariotis/go-discover
GOBIN    := $(CURDIR)/bin
PATH     := $(GOBIN):$(PATH)

# Tools (will be installed in GOBIN)
TOOLS += github.com/mattn/goveralls
TOOLS += github.com/maxbrunsfeld/counterfeiter/v6
TOOLS += github.com/golangci/golangci-lint/cmd/golangci-lint

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
