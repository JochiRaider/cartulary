#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
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
scope=(
  "src"
  "e2e"
  "vite.config.ts"
  "playwright.config.ts"
)

if [[ "${mode}" == "format" || "${mode}" == "write" ]]; then
  command=("${PNPM_BIN}" --dir "${ROOT_DIR}/apps/web" exec biome check --write)
else
  command=("${PNPM_BIN}" --dir "${ROOT_DIR}/apps/web" exec biome check)
fi

if [[ "${mode}" == "check" && "$#" -eq 0 ]]; then
  command+=(--reporter=summary --diagnostic-level=warn)
fi

if [[ "$#" -gt 0 ]]; then
  command+=("$@")
fi

command+=("${scope[@]}")

env PATH="${path_prefix}" COREPACK_HOME="${corepack_home}" "${command[@]}"
