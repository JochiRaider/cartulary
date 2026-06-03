#!/usr/bin/env bash
set -euo pipefail

go_bin="${GO:?GO is required}"
output="${BUILD_OUTPUT:?BUILD_OUTPUT is required}"
package="${BUILD_PACKAGE:?BUILD_PACKAGE is required}"
label="${BUILD_LABEL:?BUILD_LABEL is required}"
target="${CARTULARY_TEST_TARGET:?CARTULARY_TEST_TARGET is required}"
run_phase="${RUN_PHASE_SCRIPT:?RUN_PHASE_SCRIPT is required}"
go_cache_dir="${GO_CACHE_DIR:?GO_CACHE_DIR is required}"
go_mod_cache_dir="${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}"

mkdir -p "$(dirname "$output")" "$go_cache_dir" "$go_mod_cache_dir"
CARTULARY_TEST_TARGET="$target" CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
  "$run_phase" "$label" -- \
  env GOCACHE="$go_cache_dir" GOMODCACHE="$go_mod_cache_dir" \
  "$go_bin" build -o "$output" "$package"
