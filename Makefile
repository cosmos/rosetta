#!/usr/bin/make -f

all: build plugin

build:
	mkdir -p ./build
	go build -mod=readonly -o ./build/rosetta ./cmd/rosetta

plugin:
	cd plugins/cosmos-hub && make plugin

plugin-debug:
	cd plugins/cosmos-hub && make plugin-debug

docker:
	docker build . --tag rosetta

test:
	go test -mod=readonly -timeout 30m -coverprofile=coverage.out -covermode=atomic ./...

.PHONY: test-system
test-system: build
	mkdir -p ./tests/systemtests/binaries/
	git clone https://github.com/cosmos/cosmos-sdk.git ./build/tmp/cosmos-sdk
	$(eval SDK_VERSION := $(shell grep -m 2 'github.com/cosmos/cosmos-sdk' go.mod | awk 'NR==2 {print $$4; exit} END{if(NR==1) print $$2}'))
	@echo "Checking out cosmos-sdk version: $(SDK_VERSION)"
	cd ./build/tmp/cosmos-sdk && git checkout $(SDK_VERSION)
	$(MAKE) -C ./build/tmp/cosmos-sdk build
	cp ./build/tmp/cosmos-sdk/build/simd$(if $(findstring v2,$(COSMOS_BUILD_OPTIONS)),v2) ./tests/systemtests/binaries/
	cp ./build/rosetta ./tests/systemtests/binaries/
	$(MAKE) -C tests/systemtests test
	rm -rf ./build/tmp

test-rosetta-ci:
	sh ./scripts/simapp-start-node.sh
	make build && make plugin
	./build/rosetta --blockchain "cosmos" --network "cosmos" --tendermint "tcp://localhost:26657" --addr "localhost:8080" --grpc "localhost:9090" &
	sleep 30
	export SIMD_BIN=./cosmos-sdk/build/simd && sh ./tests/rosetta-cli/rosetta-cli-test.sh

###############################################################################
###                                Linting                                  ###
###############################################################################

golangci_version=v1.61.0
golangci_installed_version=$(shell golangci-lint version --format short 2>/dev/null)

#? lint-install: Install golangci-lint
lint-install:
ifneq ($(golangci_installed_version),$(golangci_version))
	@echo "--> Installing golangci-lint $(golangci_version)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
endif

#? lint: Run golangci-lint
lint:
	@echo "--> Running linter"
	$(MAKE) lint-install
	@./scripts/go-lint-all.bash --timeout=15m

#? lint: Run golangci-lint and fix
lint-fix:
	@echo "--> Running linter"
	$(MAKE) lint-install
	@./scripts/go-lint-all.bash --fix

.PHONY: all build rosetta test lint lint-fix
