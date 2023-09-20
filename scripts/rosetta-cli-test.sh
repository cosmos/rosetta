#!/bin/bash
export PATH=${PATH}:`go env GOPATH`/bin
go install github.com/coinbase/rosetta-cli@v0.10.0
rosetta-cli check:data  --in-memory-mode --configuration-file ./tests/config/rosetta-cli.json