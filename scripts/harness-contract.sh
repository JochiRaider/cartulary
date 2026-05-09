#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
script="${ROOT_DIR}/scripts/harness-contract.mjs"

find_node() {
  if [[ -n "${NODE_BIN:-}" && -x "${NODE_BIN}" ]]; then
    printf '%s\n' "${NODE_BIN}"
    return 0
  fi
  if [[ -x "${ROOT_DIR}/tmp/node-runtime/bin/node" ]]; then
    printf '%s\n' "${ROOT_DIR}/tmp/node-runtime/bin/node"
    return 0
  fi
  if command -v node >/dev/null 2>&1; then
    command -v node
    return 0
  fi
  return 1
}

fallback_preflight() {
  local target="${1:-}"
  local raw="${CARTULARY_OUTPUT_MODE:-}"
  local mode=""

  if [[ -n "${raw}" ]]; then
    case "${raw}" in
      quiet|summary|ci|verbose|debug|machine) mode="${raw}" ;;
      *)
        echo "[FAIL] failure_class=config failure_reason=configuration_error exit_code=2 invalid CARTULARY_OUTPUT_MODE ${raw}" >&2
        return 2
        ;;
    esac
  elif [[ "${VERBOSE:-}" == "1" ]]; then
    mode="verbose"
  elif [[ "${CI_VERBOSE:-}" == "1" || "${CI:-}" == "1" || "${target}" == "ci" ]]; then
    mode="ci"
  else
    mode="summary"
  fi

  if [[ "${mode}" == "machine" ]]; then
    case "${target}" in
      help|help-all|dev|clean|distclean|task-surface-report|task-guide|target-plan|fixture-report|explain-run|explain-phase|explain-target)
        echo "[FAIL] failure_class=config failure_reason=usage_error exit_code=2 target ${target} does not accept CARTULARY_OUTPUT_MODE=machine" >&2
        return 2
        ;;
    esac
  fi
}

if node_bin="$(find_node)"; then
  exec "${node_bin}" "${script}" "$@"
fi

if [[ "${1:-}" == "preflight" ]]; then
  fallback_preflight "${2:-}"
  exit $?
fi

echo "node is required for harness contract command ${1:-<missing>}" >&2
exit 1
