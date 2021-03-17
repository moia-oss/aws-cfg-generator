GO_PREFIX    := CGO_ENGABLED=0
GO           := $(GO_PREFIX) go
LINT_TARGETS ?= $(shell $(GO) list -f '{{.Dir}}' ./... | sed -e"s|${CURDIR}/\(.*\)\$$|\1/...|g" )
SYSTEM       := $(shell uname -s | tr A-Z a-z)_$(shell uname -m | sed "s/x86_64/amd64/")

# The current version of golangci-lint.
# See: https://github.com/golangci/golangci-lint/releases
GOLANGCI_LINT_VERSION := 1.38.0

# Executes the linter on all our go files inside of the project
.PHONY: lint
lint: bin/golangci-lint-$(GOLANGCI_LINT_VERSION)
	$(GO_PREFIX) ./bin/golangci-lint-$(GOLANGCI_LINT_VERSION) run $(LINT_TARGETS)

# Format all code
.PHONY: format
format:
	gofmt -s -w ./aws-cfg-generator/

# Check formatting
.PHONY: format-check
format-check:
	if [ "$$(gofmt -s -l ./aws-cfg-generator/ | wc -l)" -gt 0 ]; then exit 1; fi;

# Run all tests
.PHONY: test
test:
	cd aws-cfg-generator; go test -test.v

.PHONY: create-golint-config
create-golint-config: .golangci.yml

# Downloads the current golangci-lint executable into the bin directory and
# makes it executable
bin/golangci-lint-$(GOLANGCI_LINT_VERSION):
	mkdir -p bin
	curl -sSLf \
		https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/golangci-lint-$(GOLANGCI_LINT_VERSION)-$(shell echo $(SYSTEM) | tr '_' '-').tar.gz \
		| tar xzOf - golangci-lint-$(GOLANGCI_LINT_VERSION)-$(shell echo $(SYSTEM) | tr '_' '-')/golangci-lint > bin/golangci-lint-$(GOLANGCI_LINT_VERSION) && chmod +x bin/golangci-lint-$(GOLANGCI_LINT_VERSION)

# Builds binaries for windows, macOS, and linux
.PHONY: build
build:
	GOOS=darwin $(GO) build -o bin/darwin-aws-cfg-generator aws-cfg-generator/main.go
	GOOS=linux $(GO) build -o bin/linux-aws-cfg-generator aws-cfg-generator/main.go
	GOOS=windows $(GO) build -o bin/windows-aws-cfg-generator aws-cfg-generator/main.go
