#!/bin/bash

set -o nounset -o pipefail -o errexit
set +u
trap "exit 1" INT

go install github.com/coinbase/rosetta-cli@v0.10.0

printf "### Running rosetta-cli tests \n"

#Add all rosetta-cli checks here
rosetta-cli check:data --configuration-file ./configs/rosetta-config-cli.json

printf "### Tests finished\n"