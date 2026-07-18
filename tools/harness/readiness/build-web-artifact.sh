#!/usr/bin/env bash
set -euo pipefail

run_step="${RUN_STEP_SCRIPT:?RUN_STEP_SCRIPT is required}"
node_runtime_dir="${NODE_RUNTIME_DIR:?NODE_RUNTIME_DIR is required}"
pnpm="${PNPM:?PNPM is required}"

vite_flags=()
if [[ -n "${VITE_BUILD_FLAGS:-}" ]]; then
  # Make owns this controlled list of frontend build flags.
  # shellcheck disable=SC2206
  vite_flags=(${VITE_BUILD_FLAGS})
fi

CARTULARY_TEST_TARGET="${CARTULARY_TEST_TARGET:-build-web}" CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
  "$run_step" "build web" -- \
  env PATH="${node_runtime_dir}/bin:${PATH}" COREPACK_HOME="${node_runtime_dir}/corepack" \
  "$pnpm" --dir apps/web exec vite build "${vite_flags[@]}"
