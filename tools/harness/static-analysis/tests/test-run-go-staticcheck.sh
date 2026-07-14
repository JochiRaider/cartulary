#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
SCRIPT="$ROOT_DIR/tools/harness/static-analysis/go-staticcheck.sh"
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

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: expected output not to contain [$needle]"
  fi
}

scratch="$(cartulary_harness_mktemp_dir go-staticcheck.XXXXXX)"
cleanup_paths+=("$scratch")

fake_go="$scratch/go"
cat >"$fake_go" <<'EOF'
#!/usr/bin/env bash
if [[ "$1" != "list" ]]; then
  echo "unexpected go invocation: $*" >&2
  exit 97
fi
cat <<'PACKAGES'
github.com/JochiRaider/cartulary/cmd/server
github.com/JochiRaider/cartulary/cmd/operator
github.com/JochiRaider/cartulary/internal/app/server
github.com/JochiRaider/cartulary/internal/gen/contracts
github.com/JochiRaider/cartulary/internal/gen/sql
github.com/JochiRaider/cartulary/internal/modules/auth
github.com/JochiRaider/cartulary/tools/testservices
PACKAGES
EOF
chmod +x "$fake_go"

missing_staticcheck="$scratch/missing-staticcheck"
status=0
output="$(GO="$fake_go" STATICCHECK_BIN="$missing_staticcheck" "$SCRIPT" 2>&1)" || status=$?
if [[ "$status" -eq 0 ]]; then
  fail "missing STATICCHECK_BIN: expected wrapper failure"
fi
assert_contains "$output" "lint-go-staticcheck requires an executable STATICCHECK_BIN at $missing_staticcheck" "missing STATICCHECK_BIN diagnostic"
assert_contains "$output" "run make go-lint-toolchain before lint-go-staticcheck" "missing STATICCHECK_BIN setup guidance"

fake_staticcheck="$scratch/staticcheck"
args_log="$scratch/staticcheck-args.log"
env_log="$scratch/staticcheck-env.log"
cat >"$fake_staticcheck" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$@" >"${FAKE_STATICCHECK_ARGS_LOG:?}"
printf 'GOCACHE=%s\nGOMODCACHE=%s\n' "${GOCACHE:-}" "${GOMODCACHE:-}" >"${FAKE_STATICCHECK_ENV_LOG:?}"
EOF
chmod +x "$fake_staticcheck"

GO="$fake_go" \
  GO_CACHE_DIR="$scratch/go-cache" \
  GO_MOD_CACHE_DIR="$scratch/go-mod-cache" \
  STATICCHECK_BIN="$fake_staticcheck" \
  FAKE_STATICCHECK_ARGS_LOG="$args_log" \
  FAKE_STATICCHECK_ENV_LOG="$env_log" \
  "$SCRIPT"

args="$(cat "$args_log")"
assert_contains "$args" "github.com/JochiRaider/cartulary/cmd/server" "staticcheck cmd package"
assert_contains "$args" "github.com/JochiRaider/cartulary/cmd/operator" "staticcheck operator cmd package"
assert_contains "$args" "github.com/JochiRaider/cartulary/internal/app/server" "staticcheck application facade package"
assert_contains "$args" "github.com/JochiRaider/cartulary/internal/modules/auth" "staticcheck module package"
assert_contains "$args" "github.com/JochiRaider/cartulary/tools/testservices" "staticcheck tools package"
assert_not_contains "$args" "github.com/JochiRaider/cartulary/internal/gen/contracts" "staticcheck generated contracts exclusion"
assert_not_contains "$args" "github.com/JochiRaider/cartulary/internal/gen/sql" "staticcheck generated sql exclusion"
assert_not_contains "$args" "-checks=" "staticcheck default inherits root config"

env_output="$(cat "$env_log")"
assert_contains "$env_output" "GOCACHE=$scratch/go-cache" "staticcheck GOCACHE"
assert_contains "$env_output" "GOMODCACHE=$scratch/go-mod-cache" "staticcheck GOMODCACHE"

GO="$fake_go" \
  GO_CACHE_DIR="$scratch/go-cache" \
  GO_MOD_CACHE_DIR="$scratch/go-mod-cache" \
  STATICCHECK_BIN="$fake_staticcheck" \
  STATICCHECK_CHECKS="SA*" \
  FAKE_STATICCHECK_ARGS_LOG="$args_log" \
  FAKE_STATICCHECK_ENV_LOG="$env_log" \
  "$SCRIPT"

override_args="$(cat "$args_log")"
assert_contains "$override_args" "-checks=SA*" "staticcheck explicit check override"
