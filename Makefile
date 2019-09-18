# Default Go binary.
ifndef GOROOT
  GOROOT = /usr/local/go
endif

# Determine the OS.
ifeq ($(OS),)
  ifeq ($(shell  uname -s), Darwin)
    GOOS = darwin
  else
    GOOS = linux
  endif
else
  GOOS = $(OS)
endif

GOCMD = GOOS=$(GOOS) go
GOTEST = $(GOCMD) test -race
GO_PKGS?=$$(go list ./... | grep -v /vendor/)

.PHONY		: test

test		:
		$(GOTEST) -v $(GO_PKGS)

# Install golangci under .bin
tools		:
		curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s v1.18.0

# See .golangci.yml for linters setup
linter		:
		./bin/golangci-lint run -c golangci.yml ./...

integration	:
		$(GOTEST) -count=1 -v -tags integration $(GO_PKGS)

bench		:
		$(GOCMD) test -tags integration -bench=. ./... -benchmem
