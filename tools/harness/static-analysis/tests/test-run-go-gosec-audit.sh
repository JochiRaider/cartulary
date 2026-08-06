#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
SCRIPT="$ROOT_DIR/tools/harness/static-analysis/go-gosec-audit.sh"
cleanup_paths=()

# shellcheck source=tools/harness/test-support/harness-scratch.sh
# shellcheck disable=SC1091
source "$ROOT_DIR/tools/harness/test-support/harness-scratch.sh"

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -rf "$path"
  done
}

trap cleanup EXIT

fail() {
  echo "$*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label: expected output to contain [$needle]"
  fi
}

assert_equals() {
  local got="$1"
  local want="$2"
  local label="$3"

  if [[ "$got" != "$want" ]]; then
    fail "$label: got [$got], want [$want]"
  fi
}

scratch="$(cartulary_harness_mktemp_dir go-gosec-audit.XXXXXX)"
export GO_CACHE_DIR="$scratch/go-cache"
export GO_MOD_CACHE_DIR="$scratch/go-mod-cache"
export GO_TMP_DIR="$scratch/go-tmp"
cleanup_paths+=("$scratch")

fake_go="$scratch/go"
cat >"$fake_go" <<'EOF'
#!/usr/bin/env bash
echo "unexpected go invocation: $*" >&2
exit 97
EOF
chmod +x "$fake_go"

missing_gosec="$scratch/missing-gosec"
status=0
output="$(GO="$fake_go" GOSEC_BIN="$missing_gosec" "$SCRIPT" 2>&1)" || status=$?
if [[ "$status" -eq 0 ]]; then
  fail "missing GOSEC_BIN: expected wrapper failure"
fi
assert_contains "$output" "go-gosec-audit requires an executable GOSEC_BIN at $missing_gosec" "missing GOSEC_BIN diagnostic"
assert_contains "$output" "run make go-security-toolchain before go-gosec-audit" "missing GOSEC_BIN setup guidance"

fake_gosec="$scratch/gosec"
args_log="$scratch/gosec-args.log"
env_log="$scratch/gosec-env.log"
cat >"$fake_gosec" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "--call--" "$@" >>"${FAKE_GOSEC_ARGS_LOG:?}"
printf 'GOMAXPROCS=%s\nGOCACHE=%s\nGOMODCACHE=%s\nGOTMPDIR=%s\nPATH=%s\n' "${GOMAXPROCS:-}" "${GOCACHE:-}" "${GOMODCACHE:-}" "${GOTMPDIR:-}" "${PATH:-}" >>"${FAKE_GOSEC_ENV_LOG:?}"
if [[ " $* " != *" -no-fail "* ]]; then
  echo "missing -no-fail" >&2
  exit 96
fi
echo "simulated gosec finding"
EOF
chmod +x "$fake_gosec"

status=0
output="$(
  GO="$fake_go" \
    GO_CACHE_DIR="$scratch/go-cache" \
    GO_MOD_CACHE_DIR="$scratch/go-mod-cache" \
  GO_TMP_DIR="$scratch/go-tmp" \
    GOSEC_BIN="$fake_gosec" \
    GOSEC_AUDIT_RUNTIME_RULES="G118,G122,G301,G302,G303,G304,G305,G306,G307" \
    GOSEC_AUDIT_RUNTIME_PATTERNS="./cmd/... ./internal/..." \
    GOSEC_AUDIT_SUPPORT_RULES="G122,G301,G302,G303,G304,G305,G306,G307" \
    GOSEC_AUDIT_SUPPORT_FLAGS="-exclude-generated -no-fail -terse" \
    CARTULARY_SEQUENCE_HOST_CPU_LIMIT=4 \
    FAKE_GOSEC_ARGS_LOG="$args_log" \
    FAKE_GOSEC_ENV_LOG="$env_log" \
    "$SCRIPT" 2>&1
)" || status=$?
if [[ "$status" -ne 0 ]]; then
  fail "advisory audit findings must not fail the target, got status $status: $output"
fi

assert_contains "$output" "go-gosec-audit advisory repository profile rules=G118,G122,G301,G302,G303,G304,G305,G306,G307 patterns=./cmd/... ./internal/... ./tools/..." "repository advisory profile label"
assert_contains "$output" "simulated gosec finding" "advisory finding output"

args="$(cat "$args_log")"
assert_equals "$(grep -c '^--call--$' "$args_log")" "1" "gosec invocation count"
assert_contains "$args" "-include=G118,G122,G301,G302,G303,G304,G305,G306,G307" "runtime audit include rules"
assert_contains "$args" "./cmd/..." "runtime audit cmd package pattern"
assert_contains "$args" "./internal/..." "runtime audit internal package pattern"
assert_contains "$args" "-terse" "support audit passthrough flags"
assert_contains "$args" "./tools/..." "support audit tools package pattern"

env_output="$(cat "$env_log")"
assert_equals "$(grep -c '^GOMAXPROCS=4$' "$env_log")" "1" "gosec audit bounded worker CPUs"
assert_contains "$env_output" "GOCACHE=$scratch/go-cache" "gosec audit GOCACHE"
assert_contains "$env_output" "GOMODCACHE=$scratch/go-mod-cache" "gosec audit GOMODCACHE"
assert_contains "$env_output" "GOTMPDIR=$scratch/go-tmp" "gosec audit GOTMPDIR"
assert_contains "$env_output" "PATH=$scratch:" "gosec audit PATH includes GO directory"
