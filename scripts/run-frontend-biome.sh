#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-${ROOT_DIR}/tmp/node-runtime}"
PNPM_BIN="${PNPM:-}"

if [[ $# -lt 1 ]]; then
  echo "usage: run-frontend-biome.sh <check|write> [biome args...]" >&2
  exit 2
fi

mode="$1"
shift

case "${mode}" in
  check|write) ;;
  *)
    echo "usage: run-frontend-biome.sh <check|write> [biome args...]" >&2
    exit 2
    ;;
esac

if [[ -z "${PNPM_BIN}" ]]; then
  if command -v pnpm >/dev/null 2>&1; then
    PNPM_BIN="$(command -v pnpm)"
  elif [[ -x "${HOME}/.local/share/pnpm/pnpm" ]]; then
    PNPM_BIN="${HOME}/.local/share/pnpm/pnpm"
  else
    echo "pnpm was not provided and could not be discovered" >&2
    exit 1
  fi
fi

path_prefix="${NODE_RUNTIME_DIR}/bin:${PATH}"
scope=(
  "src"
  "e2e"
  "vite.config.ts"
  "playwright.config.ts"
)

if [[ "${mode}" == "write" ]]; then
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

env PATH="${path_prefix}" "${command[@]}"
