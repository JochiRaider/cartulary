#!/usr/bin/env bash
set -euo pipefail

stamp="${FRONTEND_INSTALL_STAMP:?FRONTEND_INSTALL_STAMP is required}"
run_phase="${RUN_PHASE_SCRIPT:?RUN_PHASE_SCRIPT is required}"
pnpm="${PNPM:?PNPM is required}"

pnpm_flags=()
if [[ -n "${PNPM_INSTALL_FLAGS:-}" ]]; then
  # The Make-owned flag string is a controlled list of CLI switches.
  # shellcheck disable=SC2206
  pnpm_flags=(${PNPM_INSTALL_FLAGS})
fi

mkdir -p "$(dirname "$stamp")"
CARTULARY_SUPPRESS_CHILD_SUCCESS=1 "$run_phase" "check frontend install" -- \
  "$pnpm" install --frozen-lockfile "${pnpm_flags[@]}"
printf 'node_path=%s\nnode_version=v%s\npnpm_path=%s\npnpm_version=%s\n' \
  "${NODE_BIN:?NODE_BIN is required}" \
  "${NODE_VERSION:?NODE_VERSION is required}" \
  "$pnpm" \
  "${PNPM_VERSION:?PNPM_VERSION is required}" >"$stamp"
