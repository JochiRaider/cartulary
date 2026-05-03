#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
SCRIPT="$ROOT_DIR/scripts/run-go-govulncheck.sh"
cleanup_paths=()

# shellcheck source=scripts/lib/harness-scratch.sh
# shellcheck disable=SC1091
source "$ROOT_DIR/scripts/lib/harness-scratch.sh"

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

scratch="$(cartulary_harness_mktemp_dir go-govulncheck.XXXXXX)"
cleanup_paths+=("$scratch")

fake_go="$scratch/go"
go_list_args_log="$scratch/go-list-args.log"
cat >"$fake_go" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" != "list" ]]; then
  echo "unexpected go invocation: $*" >&2
  exit 97
fi

shift
printf '%s\n' "$@" >"${FAKE_GO_LIST_ARGS_LOG:?}"
for pattern in "$@"; do
  case "$pattern" in
    ./cmd/... | github.com/JochiRaider/cartulary/cmd/...)
      printf '%s\n' github.com/JochiRaider/cartulary/cmd/server
      ;;
    ./internal/... | github.com/JochiRaider/cartulary/internal/...)
      printf '%s\n' \
        github.com/JochiRaider/cartulary/internal/app \
        github.com/JochiRaider/cartulary/internal/gen/contracts \
        github.com/JochiRaider/cartulary/internal/gen/sql \
        github.com/JochiRaider/cartulary/internal/modules/auth
      ;;
    ./db/... | github.com/JochiRaider/cartulary/db/...)
      printf '%s\n' github.com/JochiRaider/cartulary/db/migrations
      ;;
    ./tools/... | github.com/JochiRaider/cartulary/tools/...)
      printf '%s\n' github.com/JochiRaider/cartulary/tools/testservices
      ;;
    ./...)
      printf '%s\n' github.com/JochiRaider/cartulary/tmp/generated-policy.fake/internal/gen/contracts
      ;;
  esac
done
EOF
chmod +x "$fake_go"

missing_govulncheck="$scratch/missing-govulncheck"
status=0
output="$(GO="$fake_go" GOVULNCHECK_BIN="$missing_govulncheck" "$SCRIPT" 2>&1)" || status=$?
if [[ "$status" -eq 0 ]]; then
  fail "missing GOVULNCHECK_BIN: expected wrapper failure"
fi
assert_contains "$output" "go-vulncheck requires an executable GOVULNCHECK_BIN at $missing_govulncheck" "missing GOVULNCHECK_BIN diagnostic"
assert_contains "$output" "run make go-security-toolchain before go-vulncheck" "missing GOVULNCHECK_BIN setup guidance"

fake_govulncheck="$scratch/govulncheck"
govulncheck_args_log="$scratch/govulncheck-args.log"
govulncheck_env_log="$scratch/govulncheck-env.log"
cat >"$fake_govulncheck" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$@" >"${FAKE_GOVULNCHECK_ARGS_LOG:?}"
printf 'GOCACHE=%s\nGOMODCACHE=%s\nPATH=%s\n' "${GOCACHE:-}" "${GOMODCACHE:-}" "${PATH:-}" >"${FAKE_GOVULNCHECK_ENV_LOG:?}"
EOF
chmod +x "$fake_govulncheck"

GO="$fake_go" \
  GO_CACHE_DIR="$scratch/go-cache" \
  GO_MOD_CACHE_DIR="$scratch/go-mod-cache" \
  GOVULNCHECK_BIN="$fake_govulncheck" \
  GOVULNCHECK_FLAGS="-test -json" \
  GOVULNCHECK_DB="file://$scratch/vulndb" \
  FAKE_GO_LIST_ARGS_LOG="$go_list_args_log" \
  FAKE_GOVULNCHECK_ARGS_LOG="$govulncheck_args_log" \
  FAKE_GOVULNCHECK_ENV_LOG="$govulncheck_env_log" \
  "$SCRIPT"

list_args="$(cat "$go_list_args_log")"
assert_contains "$list_args" "./cmd/..." "govulncheck package discovery cmd root"
assert_contains "$list_args" "./internal/..." "govulncheck package discovery internal root"
assert_contains "$list_args" "./db/..." "govulncheck package discovery db root"
assert_contains "$list_args" "./tools/..." "govulncheck package discovery tools root"
assert_not_contains "$list_args" "./..." "govulncheck package discovery must not use repo-wide pattern"

args="$(cat "$govulncheck_args_log")"
assert_contains "$args" "-db" "govulncheck DB flag"
assert_contains "$args" "file://$scratch/vulndb" "govulncheck DB value"
assert_contains "$args" "-test" "govulncheck test flag"
assert_contains "$args" "-json" "govulncheck passthrough flag"
assert_contains "$args" "github.com/JochiRaider/cartulary/cmd/server" "govulncheck cmd package"
assert_contains "$args" "github.com/JochiRaider/cartulary/internal/app" "govulncheck internal authored package"
assert_contains "$args" "github.com/JochiRaider/cartulary/internal/modules/auth" "govulncheck module package"
assert_contains "$args" "github.com/JochiRaider/cartulary/db/migrations" "govulncheck db package"
assert_contains "$args" "github.com/JochiRaider/cartulary/tools/testservices" "govulncheck tools package"
assert_not_contains "$args" "github.com/JochiRaider/cartulary/internal/gen/contracts" "govulncheck generated contracts exclusion"
assert_not_contains "$args" "github.com/JochiRaider/cartulary/internal/gen/sql" "govulncheck generated sql exclusion"
assert_not_contains "$args" "github.com/JochiRaider/cartulary/tmp/" "govulncheck repo-local tmp exclusion"

env_output="$(cat "$govulncheck_env_log")"
assert_contains "$env_output" "GOCACHE=$scratch/go-cache" "govulncheck GOCACHE"
assert_contains "$env_output" "GOMODCACHE=$scratch/go-mod-cache" "govulncheck GOMODCACHE"
assert_contains "$env_output" "PATH=$scratch:" "govulncheck PATH includes GO directory"
