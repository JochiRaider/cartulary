#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"

SHELLCHECK_VERSION="${SHELLCHECK_VERSION:-0.11.0}"
export SHELLCHECK_VERSION

exec "${ROOT_DIR}/tools/harness/readiness/bootstrap-shellcheck.sh" "$@"
