#!/usr/bin/env bash
# The admitted part of a public Make recipe (P5-3e). The recipe runs this
# under workspace admission with each public input's value and origin. It
# readies the Node runtime, then runs the harness command entry.
#
#   install    produce the pinned Node runtime, unless a graph child skips
#              prerequisites;
#   installed  require the installed runtime.
set -euo pipefail

root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../../.." && pwd)"
[[ $# == 2 && -n "$2" ]] || { echo 'usage: command-entry.sh install|installed <target>' >&2; exit 2; }
case "$1" in
  install)
    if [[ "${CARTULARY_HARNESS_SKIP_PREREQUISITES:-0}" != 1 ]]; then
      mkdir -p -- "$GO_CACHE_DIR" "$GO_MOD_CACHE_DIR" "$GO_TMP_DIR"
      (cd -- "$root" && env -u CARTULARY_TEST_TARGET CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
        GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" \
        "${GO:-go}" run ./tools/harnessbuild produce node-runtime)
    fi
    ;;
  installed)
    [[ -x "$NODE_BIN" ]] || { echo 'installed Node runtime required; run make bootstrap PART=node' >&2; exit 2; }
    ;;
  *) echo "command-entry.sh: unknown Node readiness $1" >&2; exit 2 ;;
esac
exec "$NODE_BIN" "$root/tools/harness/command-surface/command-entry.mjs" "$2"
