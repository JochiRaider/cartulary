#!/usr/bin/env bash
set -euo pipefail

go_bin="${GO:?GO is required}"
toolbin_dir="${TOOLBIN_DIR:?TOOLBIN_DIR is required}"
output="${TOOL_OUTPUT:?TOOL_OUTPUT is required}"
module="${TOOL_MODULE:?TOOL_MODULE is required}"
binary_name="${TOOL_BINARY_NAME:?TOOL_BINARY_NAME is required}"
go_cache_dir="${GO_CACHE_DIR:?GO_CACHE_DIR is required}"
go_mod_cache_dir="${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}"
run_step="${RUN_STEP_SCRIPT:?RUN_STEP_SCRIPT is required}"
label="${TOOL_LABEL:-bootstrap ${binary_name} tool}"

mkdir -p "$toolbin_dir" "$go_cache_dir" "$go_mod_cache_dir"
rm -f "${toolbin_dir}/${binary_name}" "$output"
"$run_step" "$label" -- \
  env GOBIN="$toolbin_dir" GOCACHE="$go_cache_dir" GOMODCACHE="$go_mod_cache_dir" \
  "$go_bin" install "$module"
mv "${toolbin_dir}/${binary_name}" "$output"
