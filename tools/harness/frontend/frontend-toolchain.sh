#!/usr/bin/env bash
set -euo pipefail

stamp="${FRONTEND_TOOLCHAIN_STAMP:?FRONTEND_TOOLCHAIN_STAMP is required}"
node_runtime_dir="${NODE_RUNTIME_DIR:?NODE_RUNTIME_DIR is required}"
node_bin="${NODE_BIN:?NODE_BIN is required}"
pnpm="${PNPM:?PNPM is required}"
node_version_expected="${NODE_VERSION:?NODE_VERSION is required}"
pnpm_version_expected="${PNPM_VERSION:?PNPM_VERSION is required}"

mkdir -p "$(dirname "$stamp")"
env PATH="${node_runtime_dir}/bin:${PATH}" COREPACK_HOME="${node_runtime_dir}/corepack" \
  "${node_runtime_dir}/bin/corepack" enable --install-directory "${node_runtime_dir}/bin" pnpm
env PATH="${node_runtime_dir}/bin:${PATH}" COREPACK_HOME="${node_runtime_dir}/corepack" \
  "${node_runtime_dir}/bin/corepack" prepare "pnpm@${pnpm_version_expected}" --activate >/dev/null

node_version="$("$node_bin" --version)"
if [[ "$node_version" != "v${node_version_expected}" ]]; then
  echo "node version mismatch: expected v${node_version_expected}, got ${node_version} at ${node_bin}" >&2
  exit 1
fi

pnpm_version="$(env PATH="${node_runtime_dir}/bin:${PATH}" COREPACK_HOME="${node_runtime_dir}/corepack" "$pnpm" --version)"
if [[ "$pnpm_version" != "$pnpm_version_expected" ]]; then
  echo "pnpm version mismatch: expected ${pnpm_version_expected}, got ${pnpm_version} at ${pnpm}" >&2
  exit 1
fi

printf 'node_path=%s\nnode_version=%s\npnpm_path=%s\npnpm_version=%s\n' \
  "$node_bin" \
  "$node_version" \
  "$pnpm" \
  "$pnpm_version" >"$stamp"
