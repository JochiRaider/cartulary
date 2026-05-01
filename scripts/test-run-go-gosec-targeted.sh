#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
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
printf '%s\n' "$@" >"${FAKE_GOSEC_ARGS_LOG:?}"
printf 'GOCACHE=%s\nGOMODCACHE=%s\nPATH=%s\n' "${GOCACHE:-}" "${GOMODCACHE:-}" "${PATH:-}" >"${FAKE_GOSEC_ENV_LOG:?}"
EOF
chmod +x "$fake_gosec"

GO="$fake_go" \
  GO_CACHE_DIR="$scratch/go-cache" \
  GO_MOD_CACHE_DIR="$scratch/go-mod-cache" \
  GOSEC_BIN="$fake_gosec" \
  GOSEC_RULES="G602,G124" \
  GOSEC_FLAGS="-exclude-generated -quiet" \
  GOSEC_PATTERNS="./cmd/... ./internal/..." \
  FAKE_GOSEC_ARGS_LOG="$args_log" \
  FAKE_GOSEC_ENV_LOG="$env_log" \
  "$SCRIPT"

args="$(cat "$args_log")"
assert_contains "$args" "-include=G602,G124" "gosec include rules"
assert_contains "$args" "-exclude-generated" "gosec generated exclusion"
assert_contains "$args" "-quiet" "gosec passthrough flags"
assert_contains "$args" "./cmd/..." "gosec cmd package pattern"
assert_contains "$args" "./internal/..." "gosec internal package pattern"

env_output="$(cat "$env_log")"
assert_contains "$env_output" "GOCACHE=$scratch/go-cache" "gosec GOCACHE"
assert_contains "$env_output" "GOMODCACHE=$scratch/go-mod-cache" "gosec GOMODCACHE"
assert_contains "$env_output" "PATH=$scratch:" "gosec PATH includes GO directory"
