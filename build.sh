#! /usr/bin/env bash
set -euo pipefail
pushd "$(dirname "$0")"

go build
go test -v ./...
