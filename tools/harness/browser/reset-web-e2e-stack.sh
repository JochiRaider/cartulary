#!/usr/bin/env bash
set -euo pipefail
umask 077

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
# shellcheck source=tools/harness/browser/browser-lifecycle-adapter.sh
source "$ROOT_DIR/tools/harness/browser/browser-lifecycle-adapter.sh"

usage() {
  echo "usage: reset-web-e2e-stack.sh [--label <label>]" >&2
}

label="reset"
while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --label)
      if [[ -z "${2:-}" ]]; then
        usage
        exit 2
      fi
      label="$2"
      shift 2
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

web_e2e_reset_stack "$ROOT_DIR" "$label"
