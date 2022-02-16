GOPACKAGES=$(shell go list ./... | grep -v /vendor/ | grep -v /tests)
GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./tests/e2e/*")
VERSION := latest

GIT_COMMIT_SHA="$(shell git rev-parse HEAD 2>/dev/null)"
GIT_REMOTE_URL="$(shell git config --get remote.origin.url 2>/dev/null)"
BUILD_DATE="$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")"
ARCH=$(shell docker version -f {{.Client.Arch}})

GO111MODULE_FLAG?=on
export GO111MODULE=$(GO111MODULE_FLAG)

export LINT_VERSION="1.27.0"

COLOR_YELLOW=\033[0;33m
COLOR_RESET=\033[0m

OSS_FILES := go.mod

.PHONY: all
all: deps fmt vet test

.PHONY: deps
deps:
	echo "Installing dependencies ..."
	go mod vendor
	go get github.com/pierrre/gotestcover
	@if ! which golangci-lint >/dev/null || [[ "$$(golangci-lint --version)" != *${LINT_VERSION}* ]]; then \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v${LINT_VERSION}; \
	fi

.PHONY: fmt
fmt: lint
	golangci-lint run --disable-all --enable=gofmt --timeout 600s --skip-dirs=tests

.PHONY: dofmt
dofmt:
	golangci-lint run --disable-all --enable=gofmt --fix --skip-dirs=tests

.PHONY: lint
lint:
	golangci-lint run --timeout 600s --skip-dirs=tests

.PHONY: vet
vet:
	go vet ${GOPACKAGES}

.PHONY: test
test:
	$(GOPATH)/bin/gotestcover -v -race -short -coverprofile=cover.out ${GOPACKAGES}
	go tool cover -html=cover.out -o=cover.html

.PHONY: ut-coverage
ut-coverage: deps fmt vet test
