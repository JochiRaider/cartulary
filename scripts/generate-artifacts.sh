#!/usr/bin/env bash
set -euo pipefail

mkdir -p "${GO_CACHE_DIR:?GO_CACHE_DIR is required}" "${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}"
"${RUN_PHASE_SCRIPT:?RUN_PHASE_SCRIPT is required}" "generate sqlc" -- \
  "${SQLC_BIN:?SQLC_BIN is required}" generate
"$RUN_PHASE_SCRIPT" "generate contracts" -- \
  env GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" "${GO:?GO is required}" run ./tools/contractgen
"$RUN_PHASE_SCRIPT" "generate otel contracts" -- \
  "${NODE_BIN:?NODE_BIN is required}" ./scripts/generate-otel-contracts.mjs --write
"$RUN_PHASE_SCRIPT" "generate design tokens" -- \
  "${NODE_BIN:?NODE_BIN is required}" ./scripts/generate-design-tokens.mjs
