#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
# shellcheck source=tools/harness/browser/browser-lifecycle-adapter.sh
source "$ROOT_DIR/tools/harness/browser/browser-lifecycle-adapter.sh"

usage() {
  echo "usage: reset-web-e2e-stack.sh [--label <label>]" >&2
}

label="reset"
if [[ "${1:-}" == "--label" ]]; then
  if [[ -z "${2:-}" ]]; then
    usage
    exit 2
  fi
  label="$2"
  shift 2
fi
if [[ "$#" -ne 0 ]]; then
  usage
  exit 2
fi

web_e2e_reset_stack "$ROOT_DIR" "$label"
