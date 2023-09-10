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

rosetta-cli:
	echo "Installing rosetta linter"
	curl -sSfL https://raw.githubusercontent.com/coinbase/rosetta-cli/master/scripts/install.sh | sh -s
	./bin/rosetta-cli --configuration-file configs/rosetta-config-cli.json check:data