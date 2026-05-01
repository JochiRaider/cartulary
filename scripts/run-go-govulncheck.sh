#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
GO_BIN="${GO:-go}"
GO_CACHE_DIR="${GO_CACHE_DIR:-/tmp/cartulary-go-build}"
GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:-/tmp/cartulary-go-mod}"
GOVULNCHECK_BIN="${GOVULNCHECK_BIN:-$ROOT_DIR/tmp/toolbin/govulncheck-v1.3.0}"
GOVULNCHECK_FLAGS="${GOVULNCHECK_FLAGS:--test}"
GOVULNCHECK_PATTERNS="${GOVULNCHECK_PATTERNS:-./...}"
GOVULNCHECK_DB="${GOVULNCHECK_DB:-}"

if [[ "$GO_BIN" != */* ]] && command -v "$GO_BIN" >/dev/null 2>&1; then
  GO_BIN="$(command -v "$GO_BIN")"
elif [[ "$GO_BIN" != /* ]]; then
  GO_BIN="$ROOT_DIR/$GO_BIN"
fi

if [[ ! -x "$GO_BIN" ]]; then
  echo "go-vulncheck requires an executable GO at $GO_BIN" >&2
  exit 1
fi

if [[ "$GOVULNCHECK_BIN" != */* ]] && command -v "$GOVULNCHECK_BIN" >/dev/null 2>&1; then
  GOVULNCHECK_BIN="$(command -v "$GOVULNCHECK_BIN")"
elif [[ "$GOVULNCHECK_BIN" != /* ]]; then
  GOVULNCHECK_BIN="$ROOT_DIR/$GOVULNCHECK_BIN"
fi

if [[ ! -x "$GOVULNCHECK_BIN" ]]; then
  echo "go-vulncheck requires an executable GOVULNCHECK_BIN at $GOVULNCHECK_BIN" >&2
  echo "run make go-security-toolchain before go-vulncheck or set GOVULNCHECK_BIN to a ready govulncheck binary" >&2
  exit 1
fi

cd "$ROOT_DIR"

args=()
if [[ -n "$GOVULNCHECK_DB" ]]; then
  args+=("-db" "$GOVULNCHECK_DB")
fi
if [[ -n "$GOVULNCHECK_FLAGS" ]]; then
  read -r -a flag_args <<<"$GOVULNCHECK_FLAGS"
  args+=("${flag_args[@]}")
fi
if [[ -z "$GOVULNCHECK_PATTERNS" ]]; then
  echo "go-vulncheck requires at least one GOVULNCHECK_PATTERNS entry" >&2
  exit 1
fi
read -r -a patterns <<<"$GOVULNCHECK_PATTERNS"

env GOCACHE="$GO_CACHE_DIR" \
  GOMODCACHE="$GO_MOD_CACHE_DIR" \
  PATH="$(dirname "$GO_BIN"):$PATH" \
  "$GOVULNCHECK_BIN" "${args[@]}" "${patterns[@]}"
