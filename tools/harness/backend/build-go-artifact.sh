#!/usr/bin/env bash
set -euo pipefail

report_build_failure() {
  local status=$?
  if [[ -n "${CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT:-}" ]]; then
    "${NODE_BIN:?}" tools/harness/runtime/command-failure-cli.mjs harness tool_diagnostic_failure || true
  fi
  return "$status"
}
trap report_build_failure ERR

go_bin="${GO:?GO is required}"
output="${BUILD_OUTPUT:?BUILD_OUTPUT is required}"
package="${BUILD_PACKAGE:?BUILD_PACKAGE is required}"
label="${BUILD_LABEL:?BUILD_LABEL is required}"
: "${CARTULARY_TEST_TARGET:?CARTULARY_TEST_TARGET is required}"
run_step="${RUN_STEP_SCRIPT:?RUN_STEP_SCRIPT is required}"
go_cache_dir="${GO_CACHE_DIR:?GO_CACHE_DIR is required}"
go_mod_cache_dir="${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}"
go_tmp_dir="${GO_TMP_DIR:?GO_TMP_DIR is required}"
go_build_tags="${GO_BUILD_TAGS:-}"

"${NODE_BIN:?}" tools/harness/readiness/cryptographic-build-cli.mjs verify

go_build_args=(-buildvcs=false)
if [[ -n "$go_build_tags" ]]; then
  go_build_args+=(-tags "$go_build_tags")
fi

mkdir -p "$(dirname "$output")" "$go_cache_dir" "$go_mod_cache_dir" "$go_tmp_dir"
# Units execute these binaries while a nested command may rebuild them, and Go
# removes and rewrites its output when GOTMPDIR is on another filesystem (a
# unit's temporary root is memory-backed). Build beside the output and rename it
# into place: an exec sees the previous or the new binary, never a partial one.
staging="$(mktemp "$(dirname "$output")/.$(basename "$output").build.XXXXXX")"
trap 'rm -f -- "$staging"' EXIT
CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
  "$run_step" "$label" -- \
  env GOCACHE="$go_cache_dir" GOMODCACHE="$go_mod_cache_dir" GOTMPDIR="$go_tmp_dir" \
  "$go_bin" build "${go_build_args[@]}" -o "$staging" "$package"
mv -f -- "$staging" "$output"

"${NODE_BIN:?}" tools/harness/readiness/cryptographic-build-cli.mjs receipt "$output"
