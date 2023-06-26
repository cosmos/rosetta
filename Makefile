#!/usr/bin/make -f

all: build

rosetta:
	go build -mod=readonly ./cmd/rosetta

build:
	go build ./cmd/rosetta.go

test:
	go test -mod=readonly -race ./...

###############################################################################
###                                Linting                                  ###
###############################################################################

golangci_lint_cmd=golangci-lint
golangci_version=v1.51.2
lint:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@./scripts/go-lint-all.bash --timeout=15m
lint-fix:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@./scripts/go-lint-all.bash --fix

.PHONY: all build rosetta test lint lint-fix