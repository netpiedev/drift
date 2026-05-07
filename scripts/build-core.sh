#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../core"
mkdir -p bin
go build -o ./bin/drift ./cmd/drift
