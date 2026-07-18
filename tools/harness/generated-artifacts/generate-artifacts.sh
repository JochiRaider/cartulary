#!/usr/bin/env bash
set -euo pipefail

mkdir -p "${GO_CACHE_DIR:?GO_CACHE_DIR is required}" "${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}" internal/gen/sql
find internal/gen/sql -maxdepth 1 -type f -name '*.go' -delete
"${RUN_PHASE_SCRIPT:?RUN_PHASE_SCRIPT is required}" "generate sqlc" -- \
  "${SQLC_BIN:?SQLC_BIN is required}" generate
"$RUN_PHASE_SCRIPT" "generate contracts" -- \
  env GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" "${GO:?GO is required}" run ./tools/contractgen
"$RUN_PHASE_SCRIPT" "generate frontend protocol types and decoders" -- \
  "${NODE_BIN:?NODE_BIN is required}" ./tools/protocol-ts/generate-protocol-types.mjs
"$RUN_PHASE_SCRIPT" "generate Network Flow tzdb" -- \
  env GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" "${GO:?GO is required}" run ./tools/networkflow-tzdb
"$RUN_PHASE_SCRIPT" "generate Network Flow Unicode 17 NFC" -- \
  env GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" "${GO:?GO is required}" run ./tools/networkflow-unicode17
"$RUN_PHASE_SCRIPT" "generate otel contracts" -- \
  "${NODE_BIN:?NODE_BIN is required}" ./tools/otel/generate-otel-contracts.mjs --write
"$RUN_PHASE_SCRIPT" "generate design tokens" -- \
  "${NODE_BIN:?NODE_BIN is required}" ./tools/harness/generated-artifacts/design-tokens/design-token-cli.mjs
"$RUN_PHASE_SCRIPT" "generate task surface and execution topology" -- \
  "$NODE_BIN" ./tools/harness/generated-artifacts/render-execution-topology-artifacts.mjs \
  --topology tools/execution_topology_manifest.json
