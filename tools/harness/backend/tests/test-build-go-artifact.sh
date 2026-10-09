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
# Use the real pinned archive for verification; fake only compilation and metadata.
pin_identity="$("${NODE_BIN:?}" --input-type=module -e '
  import { readFileSync } from "node:fs";
  const pins = JSON.parse(readFileSync(process.argv[1], "utf8"));
  console.log(pins.go_toolchain, pins.cryptographic_module.selector);
' "$ROOT_DIR/tools/toolchain_pins.json")"
read -r FAKE_GO_TOOLCHAIN FAKE_GO_MODULE_SELECTOR <<<"$pin_identity"
export FAKE_GO_ROOT FAKE_GO_TOOLCHAIN FAKE_GO_MODULE_SELECTOR
FAKE_GO_ROOT="$("${GO:-go}" env GOROOT)"

cat >"$fake_go" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
if [[ "${1:-}" == "env" ]]; then
  printf '{"GOROOT":"%s","GOVERSION":"%s","GOFIPS140":"%s"}\n' "$FAKE_GO_ROOT" "$FAKE_GO_TOOLCHAIN" "${GOFIPS140:-}"
  exit 0
fi
if [[ "${1:-}" == "version" ]]; then
  printf '{"GoVersion":"%s","Settings":[{"Key":"GOFIPS140","Value":"%s"}]}\n' "$FAKE_GO_TOOLCHAIN" "$FAKE_GO_MODULE_SELECTOR"
  exit 0
fi
printf '%s\n' "$@" >"${FAKE_GO_ARGS_LOG:?}"
printf 'GOCACHE=%s\nGOMODCACHE=%s\nGOTMPDIR=%s\n' \
  "${GOCACHE:-}" "${GOMODCACHE:-}" "${GOTMPDIR:-}" >"${FAKE_GO_ENV_LOG:?}"
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
FAKE_GO_ENV_LOG="$TMP_DIR/go-env.log" \
GO="$fake_go" \
GOFIPS140="$FAKE_GO_MODULE_SELECTOR" \
GO_BUILD_TAGS="cartulary_harness" \
BUILD_OUTPUT="$output" \
BUILD_PACKAGE="./cmd/server" \
BUILD_LABEL="build deterministic fixture" \
CARTULARY_TEST_TARGET="build-server-harness" \
RUN_STEP_SCRIPT="$fake_run_step" \
GO_CACHE_DIR="$TMP_DIR/go-cache" \
GO_MOD_CACHE_DIR="$TMP_DIR/go-mod-cache" \
GO_TMP_DIR="$TMP_DIR/go-tmp" \
  "$SCRIPT"

[[ -f "$output" ]] || fail "build helper did not create the declared output"
[[ -f "$output.crypto.json" ]] || fail "build helper did not create the build identity receipt"
grep -Fq "\"selector\": \"$FAKE_GO_MODULE_SELECTOR\"" "$output.crypto.json" || fail "pinned receipt omitted identity"
if GO="$fake_go" GOFIPS140=off "${NODE_BIN:?}" tools/harness/readiness/cryptographic-build-cli.mjs verify >"$TMP_DIR/rejected.log" 2>&1; then
  fail "ordinary build was admitted"
fi
grep -Fxq -- "build" "$args_log" || fail "build helper did not invoke go build"
grep -Fxq -- "-buildvcs=false" "$args_log" || fail "cached Go build retained undeclared VCS stamping"
grep -Fxq -- "-tags" "$args_log" || fail "build helper omitted declared build tags"
grep -Fxq -- "cartulary_harness" "$args_log" || fail "build helper omitted the harness tag value"
grep -Fxq -- "./cmd/server" "$args_log" || fail "build helper omitted the declared package"
grep -Fxq -- "GOCACHE=$TMP_DIR/go-cache" "$TMP_DIR/go-env.log" || fail "build helper omitted GOCACHE"
grep -Fxq -- "GOMODCACHE=$TMP_DIR/go-mod-cache" "$TMP_DIR/go-env.log" || fail "build helper omitted GOMODCACHE"
grep -Fxq -- "GOTMPDIR=$TMP_DIR/go-tmp" "$TMP_DIR/go-env.log" || fail "build helper omitted GOTMPDIR"

echo "Go build artifact tests passed"
