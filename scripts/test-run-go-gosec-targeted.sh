#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
SCRIPT="$ROOT_DIR/scripts/run-go-gosec-targeted.sh"
cleanup_paths=()

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

scratch="$(mktemp -d "$ROOT_DIR/tmp/go-gosec-targeted.XXXXXX")"
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
assert_contains "$output" "go-gosec-targeted requires an executable GOSEC_BIN at $missing_gosec" "missing GOSEC_BIN diagnostic"
assert_contains "$output" "run make go-security-toolchain before go-gosec-targeted" "missing GOSEC_BIN setup guidance"

fake_gosec="$scratch/gosec"
args_log="$scratch/gosec-args.log"
env_log="$scratch/gosec-env.log"
cat >"$fake_gosec" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "--call--" "$@" >>"${FAKE_GOSEC_ARGS_LOG:?}"
printf 'GOCACHE=%s\nGOMODCACHE=%s\nPATH=%s\n' "${GOCACHE:-}" "${GOMODCACHE:-}" "${PATH:-}" >>"${FAKE_GOSEC_ENV_LOG:?}"
if [[ " $* " == *" -no-fail "* ]]; then
  echo "targeted gosec profile must remain blocking" >&2
  exit 96
fi
EOF
chmod +x "$fake_gosec"

output="$(
  GO="$fake_go" \
    GO_CACHE_DIR="$scratch/go-cache" \
    GO_MOD_CACHE_DIR="$scratch/go-mod-cache" \
    GOSEC_BIN="$fake_gosec" \
    GOSEC_RULES="G602,G124,G112,G114" \
    GOSEC_FLAGS="-exclude-generated -quiet" \
    GOSEC_PATTERNS="./cmd/... ./internal/..." \
    GOSEC_TARGETED_RUNTIME_RULES="G122,G301,G302,G303,G304,G305,G306,G307" \
    GOSEC_TARGETED_RUNTIME_FLAGS="-exclude-generated -quiet -exclude-dir=internal/testutil" \
    GOSEC_TARGETED_RUNTIME_PATTERNS="./cmd/... ./internal/..." \
    FAKE_GOSEC_ARGS_LOG="$args_log" \
    FAKE_GOSEC_ENV_LOG="$env_log" \
    "$SCRIPT" 2>&1
)"

assert_contains "$output" "go-gosec-targeted general profile rules=G602,G124,G112,G114 patterns=./cmd/... ./internal/..." "general targeted profile label"
assert_contains "$output" "go-gosec-targeted runtime profile rules=G122,G301,G302,G303,G304,G305,G306,G307 patterns=./cmd/... ./internal/..." "runtime targeted profile label"

args="$(cat "$args_log")"
assert_equals "$(grep -c '^--call--$' "$args_log")" "2" "gosec invocation count"
assert_contains "$args" "-include=G602,G124,G112,G114" "gosec include rules"
assert_contains "$args" "-exclude-generated" "gosec generated exclusion"
assert_contains "$args" "-quiet" "gosec passthrough flags"
assert_contains "$args" "./cmd/..." "gosec cmd package pattern"
assert_contains "$args" "./internal/..." "gosec internal package pattern"
assert_contains "$args" "-include=G122,G301,G302,G303,G304,G305,G306,G307" "runtime gosec include rules"
assert_contains "$args" "-exclude-dir=internal/testutil" "runtime gosec excludes internal test helpers"
if grep -q '^-no-fail$' "$args_log"; then
  fail "targeted gosec wrapper must not pass -no-fail"
fi

env_output="$(cat "$env_log")"
assert_contains "$env_output" "GOCACHE=$scratch/go-cache" "gosec GOCACHE"
assert_contains "$env_output" "GOMODCACHE=$scratch/go-mod-cache" "gosec GOMODCACHE"
assert_contains "$env_output" "PATH=$scratch:" "gosec PATH includes GO directory"

failing_gosec="$scratch/failing-gosec"
cat >"$failing_gosec" <<'EOF'
#!/usr/bin/env bash
if [[ " $* " == *" -include=G122,G301,G302,G303,G304,G305,G306,G307 "* ]]; then
  echo "simulated runtime gosec finding" >&2
  exit 42
fi
EOF
chmod +x "$failing_gosec"

status=0
output="$(
  GO="$fake_go" \
    GOSEC_BIN="$failing_gosec" \
    GOSEC_RULES="G602,G124,G112,G114" \
    GOSEC_TARGETED_RUNTIME_RULES="G122,G301,G302,G303,G304,G305,G306,G307" \
    "$SCRIPT" 2>&1
)" || status=$?
if [[ "$status" -eq 0 ]]; then
  fail "runtime targeted findings must fail the wrapper"
fi
assert_contains "$output" "simulated runtime gosec finding" "runtime targeted failure propagation"
