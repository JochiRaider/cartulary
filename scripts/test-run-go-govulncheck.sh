#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
SCRIPT="$ROOT_DIR/scripts/run-go-govulncheck.sh"
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
      printf '%s\n' \
        github.com/JochiRaider/cartulary/cmd/operator \
        github.com/JochiRaider/cartulary/cmd/server
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
set -euo pipefail
printf '%s\n' "$@" >"${FAKE_GOVULNCHECK_ARGS_LOG:?}"
printf 'GOCACHE=%s\nGOMODCACHE=%s\nPATH=%s\n' "${GOCACHE:-}" "${GOMODCACHE:-}" "${PATH:-}" >"${FAKE_GOVULNCHECK_ENV_LOG:?}"
cat <<'JSON'
{
  "config": {
    "protocol_version": "v1.0.0",
    "scanner_name": "govulncheck",
    "scanner_version": "v1.3.0",
    "scan_level": "symbol",
    "scan_mode": "source"
  }
}
JSON
if [[ "${FAKE_GOVULNCHECK_MODE:-pass}" == "malformed" ]]; then
  printf '%s\n' 'not-json'
  exit 0
fi
if [[ "${FAKE_GOVULNCHECK_MODE:-pass}" == "blocking" ]]; then
  cat <<'JSON'
{
  "osv": {
    "id": "GO-2099-0001",
    "aliases": ["CVE-2099-0001"],
    "summary": "synthetic reachable vulnerability",
    "affected": [
      {
        "package": {
          "name": "example.com/vulnerable",
          "ecosystem": "Go"
        },
        "ranges": [
          {
            "type": "SEMVER",
            "events": [
              {
                "introduced": "0"
              },
              {
                "fixed": "1.2.3"
              }
            ]
          }
        ]
      }
    ]
  }
}
{
  "finding": {
    "osv": "GO-2099-0001",
    "fixed_version": "v1.2.3",
    "trace": [
      {
        "module": "example.com/vulnerable",
        "version": "v1.2.2",
        "package": "example.com/vulnerable",
        "function": "Explode",
        "position": {
          "filename": "vulnerable.go",
          "line": 7,
          "column": 11
        }
      }
    ]
  }
}
JSON
fi
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
assert_contains "$args" "github.com/JochiRaider/cartulary/cmd/operator" "govulncheck operator cmd package"
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

phase_artifact_dir="$scratch/phase-artifacts"
status=0
output="$(
  GO="$fake_go" \
    GO_CACHE_DIR="$scratch/go-cache" \
    GO_MOD_CACHE_DIR="$scratch/go-mod-cache" \
    GOVULNCHECK_BIN="$fake_govulncheck" \
    GOVULNCHECK_FLAGS="-test -json" \
    CARTULARY_PHASE_ARTIFACT_DIR="$phase_artifact_dir" \
    FAKE_GO_LIST_ARGS_LOG="$go_list_args_log" \
    FAKE_GOVULNCHECK_ARGS_LOG="$govulncheck_args_log" \
    FAKE_GOVULNCHECK_ENV_LOG="$govulncheck_env_log" \
    FAKE_GOVULNCHECK_MODE="blocking" \
    "$SCRIPT" 2>&1
)" || status=$?
if [[ "$status" -eq 0 ]]; then
  fail "blocking Govulncheck finding: expected wrapper failure"
fi
assert_contains "$output" "GO-2099-0001" "blocking Govulncheck raw output"
findings_json="$(cat "$phase_artifact_dir/govulncheck-findings.json")"
assert_contains "$findings_json" '"schema_id": "cartulary.govulncheck_findings.v1"' "Govulncheck findings schema"
assert_contains "$findings_json" '"blocking_count": 1' "Govulncheck blocking count"
assert_contains "$findings_json" '"reachability": "symbol"' "Govulncheck symbol reachability"

malformed_artifact_dir="$scratch/malformed-phase-artifacts"
status=0
output="$(
  GO="$fake_go" \
    GO_CACHE_DIR="$scratch/go-cache" \
    GO_MOD_CACHE_DIR="$scratch/go-mod-cache" \
    GOVULNCHECK_BIN="$fake_govulncheck" \
    GOVULNCHECK_FLAGS="-test -json" \
    CARTULARY_PHASE_ARTIFACT_DIR="$malformed_artifact_dir" \
    FAKE_GO_LIST_ARGS_LOG="$go_list_args_log" \
    FAKE_GOVULNCHECK_ARGS_LOG="$govulncheck_args_log" \
    FAKE_GOVULNCHECK_ENV_LOG="$govulncheck_env_log" \
    FAKE_GOVULNCHECK_MODE="malformed" \
    "$SCRIPT" 2>&1
)" || status=$?
if [[ "$status" -eq 0 ]]; then
  fail "malformed Govulncheck JSON: expected wrapper failure"
fi
assert_contains "$output" "govulncheck JSON parse failed" "malformed Govulncheck diagnostic"
