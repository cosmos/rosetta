#!/bin/bash
go install github.com/coinbase/rosetta-cli@v0.10.0
rosetta-cli check:data --configuration-file ./tests/config/rosetta-cli.json