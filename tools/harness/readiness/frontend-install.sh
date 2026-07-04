#!/usr/bin/env bash
set -euo pipefail

stamp="${FRONTEND_INSTALL_STAMP:?FRONTEND_INSTALL_STAMP is required}"
run_phase="${RUN_PHASE_SCRIPT:?RUN_PHASE_SCRIPT is required}"
pnpm="${PNPM:?PNPM is required}"

expected_store_dir=".pnpm-store"
expected_confirm_modules_purge="false"
script_path="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/$(basename "${BASH_SOURCE[0]}")"

pnpm_flags=()
if [[ -n "${PNPM_INSTALL_FLAGS:-}" ]]; then
  # The Make-owned flag string is a controlled list of CLI switches.
  # shellcheck disable=SC2206
  pnpm_flags=(${PNPM_INSTALL_FLAGS})
fi

run_install_child() {
  local stdout_log
  local stderr_log
  local status
  local lockfile_failure=0

  stdout_log="$(mktemp)"
  stderr_log="$(mktemp)"

  set +e
  env CI=true "$pnpm" install --frozen-lockfile "${pnpm_flags[@]}" \
    > >(tee "$stdout_log") \
    2> >(tee "$stderr_log" >&2)
  status=$?
  set -e

  if [[ "$status" -ne 0 ]] && grep -Eq "ERR_PNPM_OUTDATED_LOCKFILE|frozen-lockfile" "$stdout_log" "$stderr_log"; then
    lockfile_failure=1
  fi
  rm -f "$stdout_log" "$stderr_log"
  if [[ "$lockfile_failure" -eq 1 ]]; then
    return 2
  fi
  return "$status"
}

if [[ "${1:-}" == "--run-install" ]]; then
  run_install_child
  exit "$?"
fi

store_dir="$("$pnpm" config get store-dir)"
if [[ "$store_dir" != "$expected_store_dir" ]]; then
  echo "pnpm store-dir must be ${expected_store_dir}; got ${store_dir:-<unset>}" >&2
  exit 2
fi

confirm_modules_purge="$("$pnpm" config get confirmModulesPurge)"
if [[ "$confirm_modules_purge" != "$expected_confirm_modules_purge" ]]; then
  echo "pnpm confirmModulesPurge must be ${expected_confirm_modules_purge}; got ${confirm_modules_purge:-<unset>}" >&2
  exit 2
fi

mkdir -p "$(dirname "$stamp")"
CARTULARY_TEST_TARGET="${CARTULARY_TEST_TARGET:-frontend-install}" \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
  "$run_phase" "frontend install" -- \
  bash "$script_path" --run-install
printf 'node_path=%s\nnode_version=v%s\npnpm_path=%s\npnpm_version=%s\npnpm_store_dir=%s\n' \
  "${NODE_BIN:?NODE_BIN is required}" \
  "${NODE_VERSION:?NODE_VERSION is required}" \
  "$pnpm" \
  "${PNPM_VERSION:?PNPM_VERSION is required}" \
  "$store_dir" >"$stamp"
