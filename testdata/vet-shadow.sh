#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

go run golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow "$@"
