#!/usr/bin/env bash
set -euo pipefail

: "${ROOT_DIR:?}"
: "${NODE_RUNTIME_DIR:?}"
: "${PNPM_BIN:?}"
: "${CARTULARY_STATIC_CACHE_STAMP:?}"

rm -f "$CARTULARY_STATIC_CACHE_STAMP"

PATH="$NODE_RUNTIME_DIR/bin:$PATH" \
  COREPACK_HOME="${COREPACK_HOME:-$NODE_RUNTIME_DIR/corepack}" \
  "$PNPM_BIN" --dir "$ROOT_DIR" exec biome lint --config-path "$ROOT_DIR/biome.markdown.json" --error-on-warnings "$@" "$ROOT_DIR"

mkdir -p "$(dirname "$CARTULARY_STATIC_CACHE_STAMP")"
printf 'lint-markdown ok\n' >"$CARTULARY_STATIC_CACHE_STAMP"
