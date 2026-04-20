#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/lib/run-phase-common.sh"

PNPM_BIN="${PNPM:-}"
NODE_BIN_PATH="${NODE_BIN:-}"
NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-$ROOT_DIR/tmp/node-runtime}"

if [[ -z "$PNPM_BIN" ]]; then
  if command -v pnpm >/dev/null 2>&1; then
    PNPM_BIN="$(command -v pnpm)"
  elif [[ -x "$HOME/.local/share/pnpm/pnpm" ]]; then
    PNPM_BIN="$HOME/.local/share/pnpm/pnpm"
  else
    echo "pnpm was not provided and could not be discovered" >&2
    exit 1
  fi
fi

if [[ -z "$NODE_BIN_PATH" ]]; then
  if [[ -x "$NODE_RUNTIME_DIR/bin/node" ]]; then
    NODE_BIN_PATH="$NODE_RUNTIME_DIR/bin/node"
  else
    NODE_BIN_PATH="node"
  fi
fi

common_env=(
  env
  CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER=1
  PLAYWRIGHT_WORKERS="${PLAYWRIGHT_WORKERS:-2}"
  PATH="${NODE_RUNTIME_DIR}/bin:${PATH}"
  CARTULARY_SERVER_BIN="${CARTULARY_SERVER_BIN:-$ROOT_DIR/server}"
  CARTULARY_MIGRATE_BIN="${CARTULARY_MIGRATE_BIN:-$ROOT_DIR/migrate}"
)

run_phase_command \
  "browser-e2e-functional other phases" \
  "${common_env[@]}" \
  "$PNPM_BIN" --dir apps/web exec playwright test \
  e2e/phase1.spec.ts e2e/phase3.spec.ts e2e/phase4.spec.ts

env \
  CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER=1 \
  PLAYWRIGHT_WORKERS="${PLAYWRIGHT_WORKERS:-2}" \
  PATH="${NODE_RUNTIME_DIR}/bin:${PATH}" \
  CARTULARY_SERVER_BIN="${CARTULARY_SERVER_BIN:-$ROOT_DIR/server}" \
  CARTULARY_MIGRATE_BIN="${CARTULARY_MIGRATE_BIN:-$ROOT_DIR/migrate}" \
  NODE_BIN="$NODE_BIN_PATH" \
  "$ROOT_DIR/scripts/lib/run-playwright-manifest-phase.sh" \
  "browser-e2e-functional phase2 authoritative" \
  phase2 authoritative -- \
  "$PNPM_BIN" --dir apps/web exec playwright test

run_phase_command \
  "browser-e2e-support phase2" \
  "${common_env[@]}" \
  "$PNPM_BIN" --dir apps/web exec playwright test \
  e2e/phase2.support.spec.ts
