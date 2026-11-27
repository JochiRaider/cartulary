#!/usr/bin/env bash
# The admitted part of a public Make recipe (P5-3e). The recipe runs this
# under workspace admission with each public input's value and origin. It
# readies the Node runtime, then runs the harness command entry.
#
#   install    produce the pinned Node runtime;
#   installed  require the installed runtime at the pinned version.
#
# An unavailable or incompatible runtime is a configuration failure (exit 2)
# in the target's result envelope: the harness's [FAIL] line, or the UI
# review's command result when the recipe passes ui_review.
set -euo pipefail

root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../../.." && pwd)"
[[ ($# == 2 || ($# == 3 && $3 == ui_review)) && -n "$2" ]] || { echo 'usage: command-entry.sh install|installed <target> [ui_review]' >&2; exit 2; }
envelope="${3:-harness}"

node_unavailable() {
  if [[ "$envelope" == ui_review ]]; then
    printf '{"schema_id":"cartulary.ui_review_command_result.v2","command_id":"cartulary.harness.command.%s.v1","session_id":null,"operation_id":null,"state":null,"epoch":null,"status":"error","exit_code":2,"failures":[{"failure_class":"config","failure_reason":"configuration_error","diagnostic_code":"tool_configuration","phase":"prerequisites","subject_id":"node","condition":"%s","recovery_id":"bootstrap_node"}],"receipt":null,"bundle_id":null,"private_refs":[]}\n' "${2//-/_}" "$1"
  else
    printf '[FAIL] target=%s failure_class=config failure_reason=configuration_error exit_code=2 subject=node condition=%s recovery="make bootstrap PART=node"\n' "$2" "$1" >&2
  fi
  exit 2
}

case "$1" in
  install)
    # The entry owns the terminal: the producer's output is shown only when
    # it fails.
    mkdir -p -- "$GO_CACHE_DIR" "$GO_MOD_CACHE_DIR" "$GO_TMP_DIR"
    produce_log="$(mktemp "$GO_TMP_DIR/command-entry-node.XXXXXX")"
    if ! (cd -- "$root" && env -u CARTULARY_TEST_TARGET \
      GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR" GOTMPDIR="$GO_TMP_DIR" \
      "${GO:-go}" run ./tools/harnessbuild produce node-runtime) >"$produce_log" 2>&1; then
      cat -- "$produce_log" >&2
      rm -f -- "$produce_log"
      exit 2
    fi
    rm -f -- "$produce_log"
    ;;
  installed)
    [[ -x "${NODE_BIN:-}" ]] || node_unavailable missing "$2"
    pinned="$(sed -nE 's/^[[:space:]]*"node_version": *"([0-9]+\.[0-9]+\.[0-9]+)",?$/\1/p' "$root/tools/toolchain_pins.json")"
    [[ -n "$pinned" && "$("$NODE_BIN" --version 2>/dev/null || true)" == "v$pinned" ]] || node_unavailable incompatible "$2"
    ;;
  *) echo "command-entry.sh: unknown Node readiness $1" >&2; exit 2 ;;
esac
exec "$NODE_BIN" "$root/tools/harness/command-surface/command-entry.mjs" "$2"
