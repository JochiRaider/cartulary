#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
GO_BIN="${GO:-go}"
GO_CACHE_DIR="${GO_CACHE_DIR:-/tmp/cartulary-go-build}"
GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:-/tmp/cartulary-go-mod}"
GOSEC_BIN="${GOSEC_BIN:-$ROOT_DIR/tmp/toolbin/gosec-v2.26.1}"
GOSEC_AUDIT_RUNTIME_RULES="${GOSEC_AUDIT_RUNTIME_RULES:-G118,G304,G122,G301,G302,G306,G307}"
GOSEC_AUDIT_RUNTIME_FLAGS="${GOSEC_AUDIT_RUNTIME_FLAGS:--exclude-generated -no-fail -quiet -exclude-dir=internal/testutil}"
GOSEC_AUDIT_RUNTIME_PATTERNS="${GOSEC_AUDIT_RUNTIME_PATTERNS:-./cmd/... ./internal/...}"
GOSEC_AUDIT_SUPPORT_RULES="${GOSEC_AUDIT_SUPPORT_RULES:-G304,G122,G301,G302,G306,G307}"
GOSEC_AUDIT_SUPPORT_FLAGS="${GOSEC_AUDIT_SUPPORT_FLAGS:--exclude-generated -no-fail -quiet}"
GOSEC_AUDIT_SUPPORT_PATTERNS="${GOSEC_AUDIT_SUPPORT_PATTERNS:-./internal/testutil/... ./tools/...}"

if [[ "$GO_BIN" != */* ]] && command -v "$GO_BIN" >/dev/null 2>&1; then
  GO_BIN="$(command -v "$GO_BIN")"
elif [[ "$GO_BIN" != /* ]]; then
  GO_BIN="$ROOT_DIR/$GO_BIN"
fi

if [[ ! -x "$GO_BIN" ]]; then
  echo "go-gosec-audit requires an executable GO at $GO_BIN" >&2
  exit 1
fi

if [[ "$GOSEC_BIN" != */* ]] && command -v "$GOSEC_BIN" >/dev/null 2>&1; then
  GOSEC_BIN="$(command -v "$GOSEC_BIN")"
elif [[ "$GOSEC_BIN" != /* ]]; then
  GOSEC_BIN="$ROOT_DIR/$GOSEC_BIN"
fi

if [[ ! -x "$GOSEC_BIN" ]]; then
  echo "go-gosec-audit requires an executable GOSEC_BIN at $GOSEC_BIN" >&2
  echo "run make go-security-toolchain before go-gosec-audit or set GOSEC_BIN to a ready gosec binary" >&2
  exit 1
fi

run_profile() {
  local label="$1"
  local rules="$2"
  local flags="$3"
  local patterns_value="$4"

  if [[ -z "$rules" ]]; then
    echo "go-gosec-audit requires at least one ${label} rule entry" >&2
    exit 1
  fi
  if [[ -z "$patterns_value" ]]; then
    echo "go-gosec-audit requires at least one ${label} package pattern" >&2
    exit 1
  fi

  local args=("-include=$rules")
  if [[ -n "$flags" ]]; then
    local flag_args=()
    read -r -a flag_args <<<"$flags"
    args+=("${flag_args[@]}")
  fi

  local patterns=()
  read -r -a patterns <<<"$patterns_value"

  printf 'go-gosec-audit %s profile rules=%s patterns=%s\n' "$label" "$rules" "$patterns_value"
  env GOCACHE="$GO_CACHE_DIR" \
    GOMODCACHE="$GO_MOD_CACHE_DIR" \
    PATH="$(dirname "$GO_BIN"):$PATH" \
    "$GOSEC_BIN" "${args[@]}" "${patterns[@]}"
}

cd "$ROOT_DIR"

run_profile "runtime" "$GOSEC_AUDIT_RUNTIME_RULES" "$GOSEC_AUDIT_RUNTIME_FLAGS" "$GOSEC_AUDIT_RUNTIME_PATTERNS"
run_profile "support" "$GOSEC_AUDIT_SUPPORT_RULES" "$GOSEC_AUDIT_SUPPORT_FLAGS" "$GOSEC_AUDIT_SUPPORT_PATTERNS"
