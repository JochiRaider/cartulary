#!/usr/bin/env bash
set -euo pipefail

go_bin="${GO:?GO is required}"
go_toolchain="${GO_TOOLCHAIN:?GO_TOOLCHAIN is required}"
toolbin_dir="${TOOLBIN_DIR:?TOOLBIN_DIR is required}"
output="${TOOL_OUTPUT:?TOOL_OUTPUT is required}"
module="${TOOL_MODULE:?TOOL_MODULE is required}"
binary_name="${TOOL_BINARY_NAME:?TOOL_BINARY_NAME is required}"
go_cache_dir="${GO_CACHE_DIR:?GO_CACHE_DIR is required}"
go_mod_cache_dir="${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}"
run_step="${RUN_STEP_SCRIPT:?RUN_STEP_SCRIPT is required}"
label="${TOOL_LABEL:-bootstrap ${binary_name} tool}"
readiness_script="${GO_TOOLCHAIN_READINESS_SCRIPT:-$(unset CDPATH && cd -- "$(dirname "$0")" && pwd)/go-toolchain-readiness.sh}"

if [[ "${GO_TOOLCHAIN_READY:-}" != "1" ]]; then
  GO="$go_bin" \
  GO_TOOLCHAIN="$go_toolchain" \
  GO_CACHE_DIR="$go_cache_dir" \
  GO_MOD_CACHE_DIR="$go_mod_cache_dir" \
    "$readiness_script" ensure
fi

output_dir="$(dirname -- "$output")"
mkdir -p "$toolbin_dir" "$output_dir"
staging_dir="$(mktemp -d "${output_dir}/.${binary_name}.install.XXXXXX")"
cleanup() {
  rm -rf -- "$staging_dir"
}
trap cleanup EXIT

"$run_step" "$label" -- \
  env GOBIN="$staging_dir" GOTOOLCHAIN="$go_toolchain" GOTELEMETRY=off \
  GOCACHE="$go_cache_dir" GOMODCACHE="$go_mod_cache_dir" \
  "$go_bin" install "$module"

staged_output="${staging_dir}/${binary_name}"
if [[ ! -f "$staged_output" || ! -x "$staged_output" ]]; then
  printf 'Go tool installation did not produce executable %s\n' "$staged_output" >&2
  exit 1
fi
mv -f -- "$staged_output" "$output"
