#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-${ROOT_DIR}/tmp/node-runtime}"
PNPM_BIN="${PNPM:-${NODE_RUNTIME_DIR}/bin/pnpm}"

if [[ $# -lt 1 ]]; then
  echo "usage: run-frontend-biome.sh <check|format|write> [biome args...]" >&2
  exit 2
fi

mode="$1"
shift

case "${mode}" in
  check|format|write) ;;
  *)
    echo "usage: run-frontend-biome.sh <check|format|write> [biome args...]" >&2
    exit 2
    ;;
esac

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
scope=(
  "apps/web/src"
  "apps/web/e2e"
  "apps/web/vite.config.ts"
  "apps/web/playwright.config.ts"
  "apps/web/playwright.shared.config.ts"
  "apps/web/playwright.webserver-backed.config.ts"
  "packages/grid-adapter/src"
  "packages/ui-contracts/src"
  "packages/view-contracts/src"
  "packages/test-utils/src"
  "packages/protocol-ts/src"
)

if [[ "${mode}" == "format" || "${mode}" == "write" ]]; then
  command=("${PNPM_BIN}" --dir "${ROOT_DIR}" exec biome check --write "${biome_root_flags[@]}")
else
  command=("${PNPM_BIN}" --dir "${ROOT_DIR}" exec biome check --error-on-warnings "${biome_root_flags[@]}")
fi

if [[ "${mode}" == "check" && "$#" -eq 0 ]]; then
  command+=(--reporter=summary --diagnostic-level=warn)
fi

if [[ "$#" -gt 0 ]]; then
  command+=("$@")
fi

command+=("${scope[@]}")

status=0
env PATH="${path_prefix}" COREPACK_HOME="${corepack_home}" "${command[@]}" || status=$?

if [[ "$status" -eq 0 && "${CARTULARY_TEST_TARGET:-}" == "lint-biome" ]]; then
  node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi
  test_output_helper="${TEST_OUTPUT_SCRIPT:-${ROOT_DIR}/tools/harness/output/test-output.mjs}"
  "${node_bin}" "${test_output_helper}" target-summary lint-biome pass \
    --quiet-success \
    --suppress-machine-output || status=$?
fi

exit "$status"
