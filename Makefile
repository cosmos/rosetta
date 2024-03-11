#!/usr/bin/make -f

all: build plugin

build:
	go build -mod=readonly ./cmd/rosetta

plugin:
	cd plugins/cosmos-hub && make plugin

plugin-debug:
	cd plugins/cosmos-hub && make plugin-debug

docker:
	docker build . --tag rosetta

test:
	go test -mod=readonly -timeout 30m -coverprofile=coverage.out -covermode=atomic ./...

test-rosetta-ci:
	sh ./scripts/simapp-start-node.sh
	make build && make plugin
	./rosetta --blockchain "cosmos" --network "cosmos" --tendermint "tcp://localhost:26657" --addr "localhost:8080" --grpc "localhost:9090" &
	sleep 30
	export SIMD_BIN=./cosmos-sdk/build/simd && sh ./tests/rosetta-cli/rosetta-cli-test.sh

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
