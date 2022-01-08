EXECUTABLE ?= k8s-leader-election
IMAGE ?= quay.io/cbrgm/$(EXECUTABLE)
GO := CGO_ENABLED=0 go
GOOS := linux
GOARCH := amd64
DATE := $(shell date -u '+%FT%T%z')

PACKAGES = $(shell go list ./...)

.PHONY: all
all: build

.PHONY: clean
clean:
	$(GO) clean -i ./...
	rm -rf dist/

.PHONY: fmt
fmt:
	$(GO) fmt $(PACKAGES)

.PHONY: test
test:
	@for PKG in $(PACKAGES); do $(GO) test -cover $$PKG || exit 1; done;

$(EXECUTABLE): $(wildcard *.go)
	env GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -v -o k8s-leader-election

.PHONY: build
build: $(EXECUTABLE)
