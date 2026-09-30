#!/usr/bin/env bash
set -euo pipefail

# Explicit setup owns module acquisition; review only validates this state.
if [[ "${CARTULARY_PREPARATION_POLICY:-}" == installed_only ]]; then
  exit 2
fi
mkdir -p "${GO_CACHE_DIR:?}" "${GO_MOD_CACHE_DIR:?}" "${GO_TMP_DIR:?}"
env GOTOOLCHAIN="${GO_TOOLCHAIN:?}" GOTELEMETRY=off \
  GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" \
  "${GO:?}" mod download
