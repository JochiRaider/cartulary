#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
UPDATE_SCRIPT="$ROOT_DIR/scripts/update-go-test-durations.mjs"
DRIFT_SCRIPT="$ROOT_DIR/scripts/check-go-test-duration-baseline-drift.mjs"

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

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/go-test-duration-baselines.XXXXXX")"
trap 'rm -rf "$tmp_dir"' EXIT

results_dir="$tmp_dir/results"
shared_dir="$results_dir/_shared"
mkdir -p "$shared_dir/backend-integration-auth-shard-01" "$shared_dir/backend-integration-testutil-shard-01"

cat >"$tmp_dir/baseline.json" <<'JSON'
{
  "schema_id": "cartulary.go_test_duration_baselines.v3",
  "default_shard_target_ms": 30000,
  "shard_target_ms_by_target": {
    "backend-integration": 18000,
    "backend-integration-support": 18000,
    "backend-store": 30000
  },
  "default_integration_weight_ms": 10000,
  "raw_aggregates": {},
  "tests": {}
}
JSON

cat >"$shared_dir/backend-integration-auth-shard-01/runner.jsonl" <<'JSONL'
{"Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/modules/auth","Test":"TestPhase1_LoginSessionLifecycle_I_1_01","Elapsed":1}
{"Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/modules/auth","Test":"TestPhase1_UserAdminAudit_I_1_03","Elapsed":1}
JSONL
printf '4000\n' >"$shared_dir/backend-integration-auth-shard-01/duration_ms.txt"
printf '0\n' >"$shared_dir/backend-integration-auth-shard-01/exit_status.txt"

: >"$shared_dir/backend-integration-testutil-shard-01/runner.jsonl"
printf '6000\n' >"$shared_dir/backend-integration-testutil-shard-01/duration_ms.txt"
printf '0\n' >"$shared_dir/backend-integration-testutil-shard-01/exit_status.txt"

"$NODE_BIN" "$UPDATE_SCRIPT" --baseline-file "$tmp_dir/baseline.json" "$results_dir" >/dev/null

"$NODE_BIN" - "$tmp_dir/baseline.json" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
const login = baseline.tests["github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_LoginSessionLifecycle_I_1_01"];
const audit = baseline.tests["github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_UserAdminAudit_I_1_03"];
const raw = baseline.raw_aggregates["backend-integration::backend-integration-testutil"];
if (login !== 2000 || audit !== 2000) {
  throw new Error(`expected allocated manifest shard wall duration of 2000/2000, got ${login}/${audit}`);
}
if (raw !== 6000) {
  throw new Error(`expected raw aggregate wall duration 6000, got ${raw}`);
}
EOF

"$NODE_BIN" "$DRIFT_SCRIPT" --baseline-file "$tmp_dir/baseline.json" "$results_dir" >/dev/null

cat >"$tmp_dir/underplanned.json" <<'JSON'
{
  "schema_id": "cartulary.go_test_duration_baselines.v3",
  "default_shard_target_ms": 30000,
  "shard_target_ms_by_target": {
    "backend-integration": 18000,
    "backend-integration-support": 18000,
    "backend-store": 30000
  },
  "default_integration_weight_ms": 10000,
  "raw_aggregates": {
    "backend-integration::backend-integration-testutil": 100
  },
  "tests": {
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_LoginSessionLifecycle_I_1_01": 100,
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_UserAdminAudit_I_1_03": 100
  }
}
JSON

set +e
underplanned_output="$("$NODE_BIN" "$DRIFT_SCRIPT" --baseline-file "$tmp_dir/underplanned.json" "$results_dir" 2>&1)"
underplanned_status=$?
set -e
if [[ "$underplanned_status" -eq 0 ]]; then
  fail "underplanned drift should fail"
fi
assert_contains "$underplanned_output" "underplanned shard=backend-integration-testutil-shard-01" "underplanned raw drift"

cat >"$tmp_dir/overplanned.json" <<'JSON'
{
  "schema_id": "cartulary.go_test_duration_baselines.v3",
  "default_shard_target_ms": 30000,
  "shard_target_ms_by_target": {
    "backend-integration": 18000,
    "backend-integration-support": 18000,
    "backend-store": 30000
  },
  "default_integration_weight_ms": 10000,
  "raw_aggregates": {
    "backend-integration::backend-integration-testutil": 30000
  },
  "tests": {
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_LoginSessionLifecycle_I_1_01": 20000,
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_UserAdminAudit_I_1_03": 20000
  }
}
JSON

set +e
overplanned_output="$("$NODE_BIN" "$DRIFT_SCRIPT" --baseline-file "$tmp_dir/overplanned.json" "$results_dir" 2>&1)"
overplanned_status=$?
set -e
if [[ "$overplanned_status" -eq 0 ]]; then
  fail "overplanned drift should fail"
fi
assert_contains "$overplanned_output" "overplanned shard=backend-integration-auth-shard-01" "overplanned manifest drift"

cat >"$tmp_dir/missing.json" <<'JSON'
{
  "schema_id": "cartulary.go_test_duration_baselines.v3",
  "default_shard_target_ms": 30000,
  "shard_target_ms_by_target": {
    "backend-integration": 18000,
    "backend-integration-support": 18000,
    "backend-store": 30000
  },
  "default_integration_weight_ms": 10000,
  "raw_aggregates": {},
  "tests": {}
}
JSON

set +e
missing_output="$("$NODE_BIN" "$DRIFT_SCRIPT" --baseline-file "$tmp_dir/missing.json" "$results_dir" 2>&1)"
missing_status=$?
set -e
if [[ "$missing_status" -eq 0 ]]; then
  fail "missing baselines should fail"
fi
assert_contains "$missing_output" "missing test baseline" "missing test drift"
assert_contains "$missing_output" "missing raw aggregate baseline" "missing raw drift"
