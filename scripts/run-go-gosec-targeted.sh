#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
GO_BIN="${GO:-go}"
GO_CACHE_DIR="${GO_CACHE_DIR:-/tmp/cartulary-go-build}"
GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:-/tmp/cartulary-go-mod}"
GOSEC_BIN="${GOSEC_BIN:-$ROOT_DIR/tmp/toolbin/gosec-v2.26.1}"
GOSEC_RULES="${GOSEC_RULES:-G602,G124,G112,G114}"
GOSEC_FLAGS="${GOSEC_FLAGS:--exclude-generated}"
GOSEC_PATTERNS="${GOSEC_PATTERNS:-./cmd/... ./internal/... ./db/... ./tools/...}"

if [[ "$GO_BIN" != */* ]] && command -v "$GO_BIN" >/dev/null 2>&1; then
  GO_BIN="$(command -v "$GO_BIN")"
elif [[ "$GO_BIN" != /* ]]; then
  GO_BIN="$ROOT_DIR/$GO_BIN"
fi

if [[ ! -x "$GO_BIN" ]]; then
  echo "go-gosec-targeted requires an executable GO at $GO_BIN" >&2
  exit 1
fi

if [[ "$GOSEC_BIN" != */* ]] && command -v "$GOSEC_BIN" >/dev/null 2>&1; then
  GOSEC_BIN="$(command -v "$GOSEC_BIN")"
elif [[ "$GOSEC_BIN" != /* ]]; then
  GOSEC_BIN="$ROOT_DIR/$GOSEC_BIN"
fi

if [[ ! -x "$GOSEC_BIN" ]]; then
  echo "go-gosec-targeted requires an executable GOSEC_BIN at $GOSEC_BIN" >&2
  echo "run make go-security-toolchain before go-gosec-targeted or set GOSEC_BIN to a ready gosec binary" >&2
  exit 1
fi

if [[ -z "$GOSEC_RULES" ]]; then
  echo "go-gosec-targeted requires at least one GOSEC_RULES entry" >&2
  exit 1
fi
if [[ -z "$GOSEC_PATTERNS" ]]; then
  echo "go-gosec-targeted requires at least one GOSEC_PATTERNS entry" >&2
  exit 1
fi

cd "$ROOT_DIR"

args=("-include=$GOSEC_RULES")
if [[ -n "$GOSEC_FLAGS" ]]; then
  read -r -a flag_args <<<"$GOSEC_FLAGS"
  args+=("${flag_args[@]}")
fi
read -r -a patterns <<<"$GOSEC_PATTERNS"

env GOCACHE="$GO_CACHE_DIR" \
  GOMODCACHE="$GO_MOD_CACHE_DIR" \
  PATH="$(dirname "$GO_BIN"):$PATH" \
  "$GOSEC_BIN" "${args[@]}" "${patterns[@]}"
