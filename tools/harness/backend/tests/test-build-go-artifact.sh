#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
SCRIPT="$ROOT_DIR/tools/harness/backend/build-go-artifact.sh"
TMP_DIR="$(mktemp -d)"

cleanup() {
  rm -rf "$TMP_DIR"
}

trap cleanup EXIT

fail() {
  echo "$*" >&2
  exit 1
}

fake_go="$TMP_DIR/go"
fake_run_step="$TMP_DIR/run-step.sh"
output="$TMP_DIR/server-harness"
args_log="$TMP_DIR/go-args.log"

cat >"$fake_go" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$@" >"${FAKE_GO_ARGS_LOG:?}"
output=""
while [[ "$#" -gt 0 ]]; do
  if [[ "$1" == "-o" ]]; then
    output="${2:-}"
    break
  fi
  shift
done
[[ -n "$output" ]] || exit 2
printf 'fake Go artifact\n' >"$output"
SH

cat >"$fake_run_step" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
shift
[[ "${1:-}" == "--" ]] || exit 2
shift
exec "$@"
SH

chmod +x "$fake_go" "$fake_run_step"

FAKE_GO_ARGS_LOG="$args_log" \
GO="$fake_go" \
GO_BUILD_TAGS="cartulary_harness" \
BUILD_OUTPUT="$output" \
BUILD_PACKAGE="./cmd/server" \
BUILD_LABEL="build deterministic fixture" \
CARTULARY_TEST_TARGET="build-server-harness" \
RUN_STEP_SCRIPT="$fake_run_step" \
GO_CACHE_DIR="$TMP_DIR/go-cache" \
GO_MOD_CACHE_DIR="$TMP_DIR/go-mod-cache" \
  "$SCRIPT"

[[ -f "$output" ]] || fail "build helper did not create the declared output"
grep -Fxq -- "build" "$args_log" || fail "build helper did not invoke go build"
grep -Fxq -- "-buildvcs=false" "$args_log" || fail "cached Go build retained undeclared VCS stamping"
grep -Fxq -- "-tags" "$args_log" || fail "build helper omitted declared build tags"
grep -Fxq -- "cartulary_harness" "$args_log" || fail "build helper omitted the harness tag value"
grep -Fxq -- "./cmd/server" "$args_log" || fail "build helper omitted the declared package"

echo "Go build artifact tests passed"
