#!/usr/bin/env bash
set -euo pipefail

mkdir -p "${GO_CACHE_DIR:?GO_CACHE_DIR is required}" "${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}" "${GO_TMP_DIR:?GO_TMP_DIR is required}" internal/gen/sql
find internal/gen/sql -maxdepth 1 -type f -name '*.go' -delete
"${NODE_BIN:?NODE_BIN is required}" \
  ./tools/harness/generated-artifacts/generate-foundation-schema-validators.mjs
"$RUN_STEP_SCRIPT" "generate migration catalog projections" -- \
  "$NODE_BIN" ./tools/database-migrations/generate-catalog-projections.mjs
"${RUN_STEP_SCRIPT:?RUN_STEP_SCRIPT is required}" "generate sqlc" -- \
  "${SQLC_BIN:?SQLC_BIN is required}" generate
"$RUN_STEP_SCRIPT" "assemble OpenAPI" -- \
  env GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" "${GO:?GO is required}" run ./tools/openapi-assemble --write
"$RUN_STEP_SCRIPT" "verify OpenAPI compatibility" -- \
  env GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" "${GO:?GO is required}" run ./tools/openapi-compatibility
"$RUN_STEP_SCRIPT" "generate OpenAPI operation catalog" -- \
  env GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" "${GO:?GO is required}" run ./tools/openapi-operation-catalog
"$RUN_STEP_SCRIPT" "generate contracts" -- \
  env GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" "${GO:?GO is required}" run ./tools/contractgen
"$RUN_STEP_SCRIPT" "generate frontend protocol types and decoders" -- \
  "${NODE_BIN:?NODE_BIN is required}" ./tools/protocol-ts/generate-protocol-types.mjs
"$RUN_STEP_SCRIPT" "generate Network Flow tzdb" -- \
  env GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" "${GO:?GO is required}" run ./tools/networkflow-tzdb
"$RUN_STEP_SCRIPT" "generate Network Flow Unicode 17 NFC" -- \
  env GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" "${GO:?GO is required}" run ./tools/networkflow-unicode17
"$RUN_STEP_SCRIPT" "generate otel contracts" -- \
  "${NODE_BIN:?NODE_BIN is required}" ./tools/otel/generate-otel-contracts.mjs --write
"$RUN_STEP_SCRIPT" "generate design tokens" -- \
  "${NODE_BIN:?NODE_BIN is required}" ./tools/harness/generated-artifacts/design-tokens/design-token-cli.mjs
"$RUN_STEP_SCRIPT" "generate design presentation" -- \
  "${NODE_BIN:?NODE_BIN is required}" ./tools/harness/generated-artifacts/design-presentation/design-presentation-cli.mjs
"$RUN_STEP_SCRIPT" "generate performance contracts" -- \
  "${NODE_BIN:?NODE_BIN is required}" ./tools/harness/generated-artifacts/performance/performance-contracts-cli.mjs
"$RUN_STEP_SCRIPT" "generate frontend visual golden manifest" -- \
  "${NODE_BIN:?NODE_BIN is required}" ./tools/harness/browser/frontend-visual-golden-manifest.mjs
"$RUN_STEP_SCRIPT" "generate task surface and execution topology" -- \
  "$NODE_BIN" ./tools/harness/generated-artifacts/render-execution-topology-artifacts.mjs \
  --topology tools/execution_topology_manifest.json
