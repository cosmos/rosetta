#!/usr/bin/env bash

set -e

REPO_ROOT="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )/.." &> /dev/null && pwd )"
export REPO_ROOT

lint_module() {
  local root="$1"
  shift
  if [ -f $root ]; then
    cd "$(dirname "$root")"
  else
    cd "$REPO_ROOT/$root"
  fi
  echo "linting $(grep "^module" go.mod) [$(date -Iseconds -u)]"
  golangci-lint run ./... -c "${REPO_ROOT}/.golangci.yml" "$@"
}
export -f lint_module

if [[ -z ${GIT_DIFF:-} ]]; then
  GIT_DIFF=$(git diff --name-only) || true
fi

if [[ -z "$GIT_DIFF" ]]; then
  echo "no files to lint"
  exit 0
fi

GIT_DIFF=$(echo $GIT_DIFF | tr -d "'" | tr ' ' '\n' | grep '\.go$' | grep -v '\.pb\.go$' | grep -Eo '^[^/]+\/[^/]+' | uniq)

lint_sdk=false
for dir in ${GIT_DIFF[@]}; do
  if [[ ! -f "$REPO_ROOT/$dir/go.mod" ]]; then
    lint_sdk=true
  else
    lint_module $dir "$@"
  fi
done

cd "$REPO_ROOT"
echo "linting github.com/cosmos/rosetta [$(date -Iseconds -u)]"
golangci-lint run ./... -c "${REPO_ROOT}/.golangci.yml" "$@" --build-tags=${LINT_TAGS}
