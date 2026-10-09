#!/usr/bin/env bash
set -euo pipefail

# Explicit setup owns module acquisition; review only validates this state.
if [[ "${CARTULARY_PREPARATION_POLICY:-}" == installed_only ]]; then
  exit 2
fi
mkdir -p "${GO_CACHE_DIR:?}" "${GO_MOD_CACHE_DIR:?}" "${GO_TMP_DIR:?}"
if [[ "${1:-}" == --analysis ]]; then
  root_dir="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
  cd "$root_dir/tools/go-analysis"
  env GOTOOLCHAIN="${GO_TOOLCHAIN:?}" GOTELEMETRY=off \
    GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" \
    "${GO:?}" mod tidy
elif [[ "$#" -ne 0 ]]; then
  echo 'usage: install-go-dependencies.sh [--analysis]' >&2
  exit 2
fi
env GOTOOLCHAIN="${GO_TOOLCHAIN:?}" GOTELEMETRY=off \
  GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" \
  "${GO:?}" mod download all
