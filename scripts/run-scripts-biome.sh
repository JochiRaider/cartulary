#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-${ROOT_DIR}/tmp/node-runtime}"
PNPM_BIN="${PNPM:-${NODE_RUNTIME_DIR}/bin/pnpm}"

if [[ ! -x "${PNPM_BIN}" ]]; then
  echo "repo-local pnpm was not found at ${PNPM_BIN}; run make frontend-toolchain" >&2
  exit 1
fi

path_prefix="${NODE_RUNTIME_DIR}/bin:${PATH}"
corepack_home="${NODE_RUNTIME_DIR}/corepack"
biome_root_flags=(
  --config-path "${ROOT_DIR}/biome.json"
  --vcs-root "${ROOT_DIR}"
  --vcs-enabled=true
  --vcs-client-kind=git
  --vcs-use-ignore-file=true
)
scope=("scripts")
command=(
  "${PNPM_BIN}"
  --dir "${ROOT_DIR}"
  exec
  biome
  check
  "${biome_root_flags[@]}"
  --formatter-enabled=false
  --assist-enabled=false
  --reporter=summary
  --diagnostic-level=warn
  --error-on-warnings
)

if [[ "$#" -gt 0 ]]; then
  command+=("$@")
fi

command+=("${scope[@]}")

env PATH="${path_prefix}" COREPACK_HOME="${corepack_home}" "${command[@]}"
