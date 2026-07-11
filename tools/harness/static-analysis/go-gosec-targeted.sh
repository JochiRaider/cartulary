#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
GO_BIN="${GO:-go}"
GO_CACHE_DIR="${GO_CACHE_DIR:-/tmp/cartulary-go-build}"
GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:-/tmp/cartulary-go-mod}"
GOSEC_BIN="${GOSEC_BIN:-$ROOT_DIR/tmp/toolbin/gosec-v2.26.1}"
NODE_BIN="${NODE_BIN:-node}"
SUPPORT_PROFILE_SCRIPT="$ROOT_DIR/tools/harness/static-analysis/support-inventory-profiles.mjs"
TEST_SUPPORT_INVENTORY="${TEST_SUPPORT_INVENTORY:-${CARTULARY_TEST_SUPPORT_INVENTORY:-$ROOT_DIR/tools/test_support_inventory.json}}"
GOSEC_RULES="${GOSEC_RULES:-G602,G124,G112,G114}"
GOSEC_FLAGS="${GOSEC_FLAGS:--exclude-generated}"
GOSEC_PATTERNS="${GOSEC_PATTERNS:-./cmd/... ./internal/... ./db/... ./tools/...}"
GOSEC_TARGETED_RUNTIME_RULES="${GOSEC_TARGETED_RUNTIME_RULES:-G122,G301,G302,G303,G304,G305,G306,G307}"
GOSEC_TARGETED_RUNTIME_FLAGS="${GOSEC_TARGETED_RUNTIME_FLAGS:-}"
GOSEC_TARGETED_RUNTIME_PATTERNS="${GOSEC_TARGETED_RUNTIME_PATTERNS:-./cmd/... ./internal/...}"
profile_metadata="${CARTULARY_PHASE_ARTIFACT_DIR:+${CARTULARY_PHASE_ARTIFACT_DIR}/security-profiles.jsonl}"

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

if [[ -z "$GOSEC_TARGETED_RUNTIME_FLAGS" ]]; then
  GOSEC_TARGETED_RUNTIME_FLAGS="$("$NODE_BIN" "$SUPPORT_PROFILE_SCRIPT" gosec-runtime-flags --base "-exclude-generated -quiet" --inventory "$TEST_SUPPORT_INVENTORY" --root "$ROOT_DIR")"
fi

run_profile() {
  local label="$1"
  local rules="$2"
  local flags="$3"
  local patterns_value="$4"

  if [[ -z "$rules" ]]; then
    echo "go-gosec-targeted requires at least one ${label} rule entry" >&2
    exit 1
  fi
  if [[ -z "$patterns_value" ]]; then
    echo "go-gosec-targeted requires at least one ${label} package pattern" >&2
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

  if [[ -n "$profile_metadata" ]]; then
    mkdir -p "$(dirname "$profile_metadata")"
    printf '{"tool":"gosec","target":"go-gosec-targeted","profile":"%s","rules":"%s","flags":"%s","patterns":"%s"}\n' \
      "$label" "$rules" "$flags" "$patterns_value" >>"$profile_metadata"
  fi
  printf 'go-gosec-targeted %s profile rules=%s patterns=%s\n' "$label" "$rules" "$patterns_value"
  env GOCACHE="$GO_CACHE_DIR" \
    GOMODCACHE="$GO_MOD_CACHE_DIR" \
    PATH="$(dirname "$GO_BIN"):$PATH" \
    "$GOSEC_BIN" "${args[@]}" "${patterns[@]}"
}

cd "$ROOT_DIR"

run_profile "general" "$GOSEC_RULES" "$GOSEC_FLAGS" "$GOSEC_PATTERNS"
run_profile "runtime" "$GOSEC_TARGETED_RUNTIME_RULES" "$GOSEC_TARGETED_RUNTIME_FLAGS" "$GOSEC_TARGETED_RUNTIME_PATTERNS"
