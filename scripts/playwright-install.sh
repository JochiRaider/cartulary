#!/usr/bin/env bash
set -euo pipefail

stamp="${PLAYWRIGHT_INSTALL_STAMP:?PLAYWRIGHT_INSTALL_STAMP is required}"
run_phase="${RUN_PHASE_SCRIPT:?RUN_PHASE_SCRIPT is required}"
node_runtime_dir="${NODE_RUNTIME_DIR:?NODE_RUNTIME_DIR is required}"
node_bin="${NODE_BIN:?NODE_BIN is required}"
pnpm="${PNPM:?PNPM is required}"
node_version="${NODE_VERSION:?NODE_VERSION is required}"
pnpm_version="${PNPM_VERSION:?PNPM_VERSION is required}"

mkdir -p "$(dirname "$stamp")"
"$run_phase" "playwright-install" -- \
  env PATH="${node_runtime_dir}/bin:${PATH}" COREPACK_HOME="${node_runtime_dir}/corepack" \
  "$pnpm" --dir apps/web exec playwright install chromium
printf 'node_path=%s\nnode_version=v%s\npnpm_path=%s\npnpm_version=%s\n' \
  "$node_bin" \
  "$node_version" \
  "$pnpm" \
  "$pnpm_version" >"$stamp"
