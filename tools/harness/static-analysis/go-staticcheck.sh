#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
GO_BIN="${GO:-go}"
GO_CACHE_DIR="${GO_CACHE_DIR:?GO_CACHE_DIR is required}"
GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}"
GO_TMP_DIR="${GO_TMP_DIR:?GO_TMP_DIR is required}"
STATICCHECK_BIN="${STATICCHECK_BIN:-$ROOT_DIR/tmp/toolbin/staticcheck-v0.7.0}"
STATICCHECK_CHECKS="${STATICCHECK_CHECKS:-}"

# shellcheck source=tools/harness/generated-artifacts/generated-artifacts.sh
# shellcheck disable=SC1091
source "$ROOT_DIR/tools/harness/generated-artifacts/generated-artifacts.sh"

if [[ "$STATICCHECK_BIN" != */* ]] && command -v "$STATICCHECK_BIN" >/dev/null 2>&1; then
  STATICCHECK_BIN="$(command -v "$STATICCHECK_BIN")"
elif [[ "$STATICCHECK_BIN" != /* ]]; then
  STATICCHECK_BIN="$ROOT_DIR/$STATICCHECK_BIN"
fi

if [[ ! -x "$STATICCHECK_BIN" ]]; then
  echo "lint-go-staticcheck requires an executable STATICCHECK_BIN at $STATICCHECK_BIN" >&2
  echo "run make go-lint-toolchain before lint-go-staticcheck or set STATICCHECK_BIN to a ready staticcheck binary" >&2
  exit 1
fi

cd "$ROOT_DIR"

mapfile -t packages < <(
  GOCACHE="$GO_CACHE_DIR" \
  GOMODCACHE="$GO_MOD_CACHE_DIR" \
  GOTMPDIR="$GO_TMP_DIR" \
    "$GO_BIN" list ./cmd/... ./internal/... ./db/... ./tools/... |
    cartulary_filter_authored_go_packages
)

if [[ "${#packages[@]}" -eq 0 ]]; then
  echo "staticcheck package discovery returned no packages" >&2
  exit 1
fi

staticcheck_args=()
if [[ -n "$STATICCHECK_CHECKS" ]]; then
  staticcheck_args+=("-checks=$STATICCHECK_CHECKS")
fi

env GOCACHE="$GO_CACHE_DIR" \
  GOMODCACHE="$GO_MOD_CACHE_DIR" \
  GOTMPDIR="$GO_TMP_DIR" \
  "$STATICCHECK_BIN" "${staticcheck_args[@]}" "${packages[@]}"
