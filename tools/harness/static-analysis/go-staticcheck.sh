#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=tools/workspace_layout.generated.sh
source "$(dirname -- "${BASH_SOURCE[0]}")/../../workspace_layout.generated.sh"

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
STATICCHECK_BIN="${STATICCHECK_BIN:-$ROOT_DIR/${CARTULARY_LAYOUT_TOOLBIN}/staticcheck-v0.8.1}"
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
  echo "run make bootstrap before make lint CHECK=go or set STATICCHECK_BIN to a ready staticcheck binary" >&2
  exit 1
fi

# shellcheck source=tools/harness/static-analysis/go-lint-modules.sh
# shellcheck disable=SC1091
source "$ROOT_DIR/tools/harness/static-analysis/go-lint-modules.sh"

staticcheck_args=()
if [[ -n "$STATICCHECK_CHECKS" ]]; then
  staticcheck_args+=("-checks=$STATICCHECK_CHECKS")
fi
cartulary_lint_go_modules "$ROOT_DIR" cartulary_filter_authored_go_packages "$STATICCHECK_BIN" "${staticcheck_args[@]}"
