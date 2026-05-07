#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../npm"
bun install
bun run build
