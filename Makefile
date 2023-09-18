#!/usr/bin/make -f

all: build

rosetta:
	go build -mod=readonly ./cmd/rosetta

build:
	go build -mod=readonly ./cmd/rosetta

plugin:
	cd plugins/cosmos-hub && make plugin

test:
	go test -mod=readonly ./...

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

docker:
	docker build . --tag rosetta

rosetta-cli:
 	rosetta-cli check:data --configuration-file ./configs/rosetta-cli.json
