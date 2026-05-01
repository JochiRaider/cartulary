#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
GO_BIN="${GO:-go}"
GO_CACHE_DIR="${GO_CACHE_DIR:-/tmp/cartulary-go-build}"
GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:-/tmp/cartulary-go-mod}"

cd "$ROOT_DIR"

mapfile -t packages < <(
  GOCACHE="$GO_CACHE_DIR" \
  GOMODCACHE="$GO_MOD_CACHE_DIR" \
    "$GO_BIN" list ./cmd/... ./internal/... ./db/... ./tools/...
)

if [[ "${#packages[@]}" -eq 0 ]]; then
  echo "go vet package discovery returned no packages" >&2
  exit 1
fi

env GOCACHE="$GO_CACHE_DIR" \
  GOMODCACHE="$GO_MOD_CACHE_DIR" \
  "$GO_BIN" vet "${packages[@]}"
