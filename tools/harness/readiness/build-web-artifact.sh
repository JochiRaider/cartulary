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

build_target="${CARTULARY_TEST_TARGET:-build-web}"
CARTULARY_TEST_TARGET="${build_target}" CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
  "$run_step" "build web" -- \
  env PATH="${node_runtime_dir}/bin:${PATH}" COREPACK_HOME="${node_runtime_dir}/corepack" \
  "${node_runtime_dir}/bin/node" tools/harness/readiness/frontend-artifact.mjs build "${build_target}" \
  "$pnpm" --dir apps/web exec vite build "${vite_flags[@]}"
