#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
MAKE_HELPER="${MAKE:-make}"
UPDATE_SCRIPT="$ROOT_DIR/scripts/update-go-test-durations.mjs"
DRIFT_SCRIPT="$ROOT_DIR/scripts/check-go-test-duration-baseline-drift.mjs"
COVERAGE_SCRIPT="$ROOT_DIR/scripts/check-go-test-duration-baseline-coverage.mjs"

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

tool_logs_from_result() {
  local output="$1"
  local root
  root="$(printf '%s\n' "$output" | sed -n 's/.* artifact_root=\([^ ]*\) .*/\1/p' | head -n 1)"
  if [[ -z "$root" ]]; then
    fail "missing artifact_root in output: $output"
  fi
  local dir
  if [[ "$root" = /* ]]; then
    dir="$root"
  else
    dir="$ROOT_DIR/$root"
  fi
  [[ -f "$dir/stdout.log" ]] && cat "$dir/stdout.log"
  [[ -f "$dir/stderr.log" ]] && cat "$dir/stderr.log"
}

write_empty_baseline() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.go_test_duration_baselines.v4",
  "default_shard_target_ms": 30000,
  "shard_target_ms_by_target": {
    "backend-integration": 18000,
    "backend-integration-support": 18000,
    "backend-store": 30000
  },
  "default_item_weight_ms": 10000,
  "command_overheads_by_target": {},
  "package_overheads": {},
  "raw_aggregates": {},
  "tests": {}
}
JSON
}

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/go-test-duration-baselines.XXXXXX")"
trap 'rm -rf "$tmp_dir"' EXIT

results_dir="$tmp_dir/results"
shared_dir="$results_dir/_shared"
mkdir -p \
  "$shared_dir/backend-integration-auth-shard-01" \
  "$shared_dir/backend-integration-auth-shard-02" \
  "$shared_dir/backend-integration-testutil-shard-01" \
  "$shared_dir/backend-store-shard-01"

write_empty_baseline "$tmp_dir/baseline.json"

cat >"$shared_dir/backend-integration-auth-shard-01/runner.jsonl" <<'JSONL'
{"Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/modules/auth","Test":"TestPhase1_LoginSessionLifecycle_I_1_01","Elapsed":1}
{"Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/modules/auth","Test":"TestPhase1_UserAdminAudit_I_1_03","Elapsed":1}
{"Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/modules/auth","Elapsed":30}
JSONL
printf '50000\n' >"$shared_dir/backend-integration-auth-shard-01/duration_ms.txt"
printf '0\n' >"$shared_dir/backend-integration-auth-shard-01/exit_status.txt"

cat >"$shared_dir/backend-integration-auth-shard-02/runner.jsonl" <<'JSONL'
{"Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/modules/auth","Test":"TestPhase1_LoginSessionLifecycle_I_1_01","Elapsed":1}
{"Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/modules/auth","Test":"TestPhase1_UserAdminAudit_I_1_03","Elapsed":1}
{"Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/modules/auth","Elapsed":30}
JSONL
cat >"$shared_dir/backend-integration-auth-shard-02/stderr.log" <<'LOG'
go: downloading example.org/slow/module v1.2.3
go: downloading example.org/slow/other v4.5.6
LOG
printf '100000\n' >"$shared_dir/backend-integration-auth-shard-02/duration_ms.txt"
printf '0\n' >"$shared_dir/backend-integration-auth-shard-02/exit_status.txt"

cat >"$shared_dir/backend-integration-testutil-shard-01/runner.jsonl" <<'JSONL'
{"Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/testutil/httptestx","Test":"TestHarnessBootsServerAndAssertsEnvelopes","Elapsed":1}
{"Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/testutil/httptestx","Elapsed":1}
{"Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/testutil/pgtest","Test":"TestPreparePackageDatabaseTReusesAndResetsMutableTables","Elapsed":5}
{"Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/testutil/pgtest","Elapsed":5}
JSONL
printf '6000\n' >"$shared_dir/backend-integration-testutil-shard-01/duration_ms.txt"
printf '0\n' >"$shared_dir/backend-integration-testutil-shard-01/exit_status.txt"

cat >"$shared_dir/backend-store-shard-01/runner.jsonl" <<'JSONL'
{"Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/modules/auth","Test":"TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_1_05","Elapsed":20}
{"Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/modules/auth","Elapsed":21}
JSONL
printf '22000\n' >"$shared_dir/backend-store-shard-01/duration_ms.txt"
printf '0\n' >"$shared_dir/backend-store-shard-01/exit_status.txt"

update_output="$("$NODE_BIN" "$UPDATE_SCRIPT" --baseline-file "$tmp_dir/baseline.json" "$results_dir" 2>&1)"
assert_contains "$update_output" "skipped contaminated Go shard timing artifacts" "contaminated refresh skip output"
assert_contains "$update_output" "shard=backend-integration-auth-shard-02 go_module_downloads=2" "contaminated refresh skip shard"

write_empty_baseline "$tmp_dir/make-baseline.json"
make_update_output="$(
  RESULTS_DIR="$results_dir" \
  GO_TEST_DURATION_BASELINE="$tmp_dir/make-baseline.json" \
  PRUNE_OBSERVED_PACKAGES=1 \
    "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" go-test-duration-baselines 2>&1
)"
assert_contains "$make_update_output" "[RESULT] target=go-test-duration-baselines status=pass" "make baseline update summary"
assert_contains "$(tool_logs_from_result "$make_update_output")" "skipped contaminated Go shard timing artifacts" "make contaminated refresh skip output"
RESULTS_DIR="$results_dir" \
GO_TEST_DURATION_BASELINE="$tmp_dir/make-baseline.json" \
  "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" go-test-duration-baseline-drift >/dev/null 2>/dev/null

"$NODE_BIN" - "$tmp_dir/baseline.json" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
const login = baseline.tests["github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_LoginSessionLifecycle_I_1_01"];
const audit = baseline.tests["github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_UserAdminAudit_I_1_03"];
const store = baseline.tests["github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_1_05"];
const integrationAuthOverhead = baseline.package_overheads["backend-integration::github.com/JochiRaider/cartulary/internal/modules/auth"];
const storeAuthOverhead = baseline.package_overheads["backend-store::github.com/JochiRaider/cartulary/internal/modules/auth"];
const integrationCommand = baseline.command_overheads_by_target["backend-integration"];
const storeCommand = baseline.command_overheads_by_target["backend-store"];
const rawHTTP = baseline.raw_aggregates["backend-integration::backend-integration-testutil::github.com/JochiRaider/cartulary/internal/testutil/httptestx"];
const rawPG = baseline.raw_aggregates["backend-integration::backend-integration-testutil::github.com/JochiRaider/cartulary/internal/testutil/pgtest"];
const legacyRaw = baseline.raw_aggregates["backend-integration::backend-integration-testutil"];
if (baseline.schema_id !== "cartulary.go_test_duration_baselines.v4") {
  throw new Error(`expected v4 schema, got ${baseline.schema_id}`);
}
if (login !== 1000 || audit !== 1000 || store !== 20000) {
  throw new Error(`expected raw test elapsed baselines, got login=${login} audit=${audit} store=${store}`);
}
if (integrationAuthOverhead !== 28000 || storeAuthOverhead !== 1000) {
  throw new Error(`expected package overhead baselines, got integration=${integrationAuthOverhead} store=${storeAuthOverhead}`);
}
if (integrationCommand !== 20000 || storeCommand !== 1000) {
  throw new Error(`expected command overhead baselines, got integration=${integrationCommand} store=${storeCommand}`);
}
if (rawHTTP !== 1000 || rawPG !== 5000 || legacyRaw !== undefined) {
  throw new Error(`expected raw package baselines http=1000 pg=5000 and no legacy aggregate, got http=${rawHTTP} pg=${rawPG} legacy=${legacyRaw}`);
}
EOF

write_empty_baseline "$tmp_dir/guarded-command-overhead.json"
"$NODE_BIN" - "$tmp_dir/guarded-command-overhead.json" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
baseline.command_overheads_by_target["backend-store"] = 4000;
fs.writeFileSync(baselineFile, `${JSON.stringify(baseline, null, 2)}\n`);
EOF

guarded_output="$("$NODE_BIN" "$UPDATE_SCRIPT" --baseline-file "$tmp_dir/guarded-command-overhead.json" "$results_dir" 2>&1)"
assert_contains "$guarded_output" "kept existing Go command overhead baselines after suspicious decreases" "guarded command overhead output"
assert_contains "$guarded_output" "target=backend-store existing_ms=4000 observed_ms=1000" "guarded command overhead target"

"$NODE_BIN" - "$tmp_dir/guarded-command-overhead.json" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
if (baseline.command_overheads_by_target["backend-store"] !== 4000) {
  throw new Error(`expected guarded backend-store command overhead to remain 4000, got ${baseline.command_overheads_by_target["backend-store"]}`);
}
EOF

write_empty_baseline "$tmp_dir/allowed-command-overhead.json"
"$NODE_BIN" - "$tmp_dir/allowed-command-overhead.json" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
baseline.command_overheads_by_target["backend-store"] = 4000;
fs.writeFileSync(baselineFile, `${JSON.stringify(baseline, null, 2)}\n`);
EOF

allowed_output="$(
  RESULTS_DIR="$results_dir" \
  GO_TEST_DURATION_BASELINE="$tmp_dir/allowed-command-overhead.json" \
  ALLOW_COMMAND_OVERHEAD_DECREASE=1 \
    "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" go-test-duration-baselines 2>&1
)"
assert_contains "$allowed_output" "[RESULT] target=go-test-duration-baselines status=pass" "allowed command overhead update summary"
assert_contains "$(tool_logs_from_result "$allowed_output")" "updated 3 Go test baselines" "allowed command overhead update output"

"$NODE_BIN" - "$tmp_dir/allowed-command-overhead.json" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
if (baseline.command_overheads_by_target["backend-store"] !== 1000) {
  throw new Error(`expected allowed backend-store command overhead to decrease to 1000, got ${baseline.command_overheads_by_target["backend-store"]}`);
}
EOF

"$NODE_BIN" "$DRIFT_SCRIPT" --baseline-file "$tmp_dir/baseline.json" "$results_dir" >/dev/null 2>/dev/null

cat >"$tmp_dir/contaminated-underplanned.json" <<'JSON'
{
  "schema_id": "cartulary.go_test_duration_baselines.v4",
  "command_overheads_by_target": {
    "backend-integration": 10000,
    "backend-store": 1000
  },
  "package_overheads": {
    "backend-integration::github.com/JochiRaider/cartulary/internal/modules/auth": 10000,
    "backend-store::github.com/JochiRaider/cartulary/internal/modules/auth": 1000
  },
  "raw_aggregates": {
    "backend-integration::backend-integration-testutil::github.com/JochiRaider/cartulary/internal/testutil/httptestx": 1000,
    "backend-integration::backend-integration-testutil::github.com/JochiRaider/cartulary/internal/testutil/pgtest": 5000
  },
  "tests": {
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_1_05": 20000,
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_LoginSessionLifecycle_I_1_01": 1000,
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_UserAdminAudit_I_1_03": 1000
  }
}
JSON
drift_warning_output="$("$NODE_BIN" "$DRIFT_SCRIPT" --baseline-file "$tmp_dir/contaminated-underplanned.json" "$results_dir" 2>&1 >/dev/null)"
assert_contains "$drift_warning_output" "ignored underplanned contaminated shard=backend-integration-auth-shard-02" "contaminated underplanned drift warning"
assert_contains "$drift_warning_output" "go_module_downloads=2" "contaminated underplanned download count"

cat >"$tmp_dir/tolerated-underplanned.json" <<'JSON'
{
  "schema_id": "cartulary.go_test_duration_baselines.v4",
  "command_overheads_by_target": {
    "backend-integration": 11000,
    "backend-store": 1000
  },
  "package_overheads": {
    "backend-integration::github.com/JochiRaider/cartulary/internal/modules/auth": 13000,
    "backend-store::github.com/JochiRaider/cartulary/internal/modules/auth": 1000
  },
  "raw_aggregates": {
    "backend-integration::backend-integration-testutil::github.com/JochiRaider/cartulary/internal/testutil/httptestx": 1000,
    "backend-integration::backend-integration-testutil::github.com/JochiRaider/cartulary/internal/testutil/pgtest": 5000
  },
  "tests": {
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_1_05": 20000,
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_LoginSessionLifecycle_I_1_01": 1000,
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_UserAdminAudit_I_1_03": 1000
  }
}
JSON
"$NODE_BIN" "$DRIFT_SCRIPT" --baseline-file "$tmp_dir/tolerated-underplanned.json" "$results_dir" >/dev/null 2>/dev/null

underplanned_raw_results="$tmp_dir/underplanned-raw-results"
cp -R "$results_dir" "$underplanned_raw_results"
printf '50000\n' >"$underplanned_raw_results/_shared/backend-integration-testutil-shard-01/duration_ms.txt"

cat >"$tmp_dir/underplanned-raw.json" <<'JSON'
{
  "schema_id": "cartulary.go_test_duration_baselines.v4",
  "command_overheads_by_target": {
    "backend-integration": 20000,
    "backend-store": 1000
  },
  "package_overheads": {
    "backend-integration::github.com/JochiRaider/cartulary/internal/modules/auth": 28000,
    "backend-store::github.com/JochiRaider/cartulary/internal/modules/auth": 1000
  },
  "raw_aggregates": {
    "backend-integration::backend-integration-testutil::github.com/JochiRaider/cartulary/internal/testutil/httptestx": 100,
    "backend-integration::backend-integration-testutil::github.com/JochiRaider/cartulary/internal/testutil/pgtest": 100
  },
  "tests": {
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_1_05": 20000,
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_LoginSessionLifecycle_I_1_01": 1000,
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_UserAdminAudit_I_1_03": 1000
  }
}
JSON

set +e
underplanned_raw_output="$("$NODE_BIN" "$DRIFT_SCRIPT" --baseline-file "$tmp_dir/underplanned-raw.json" "$underplanned_raw_results" 2>&1)"
underplanned_raw_status=$?
set -e
if [[ "$underplanned_raw_status" -eq 0 ]]; then
  fail "underplanned raw drift should fail"
fi
assert_contains "$underplanned_raw_output" "underplanned shard=backend-integration-testutil-shard-01" "underplanned raw drift"
assert_contains "$underplanned_raw_output" "raw_packages=[" "underplanned raw package detail"
assert_contains "$underplanned_raw_output" "package=github.com/JochiRaider/cartulary/internal/testutil/pgtest planned_ms=100 actual_ms=41667 delta_ms=+41567 ratio=416.67" "underplanned raw pgtest detail"

cat >"$tmp_dir/underplanned-components.json" <<'JSON'
{
  "schema_id": "cartulary.go_test_duration_baselines.v4",
  "command_overheads_by_target": {
    "backend-integration": 100,
    "backend-store": 1000
  },
  "package_overheads": {
    "backend-integration::github.com/JochiRaider/cartulary/internal/modules/auth": 100,
    "backend-store::github.com/JochiRaider/cartulary/internal/modules/auth": 1000
  },
  "raw_aggregates": {
    "backend-integration::backend-integration-testutil::github.com/JochiRaider/cartulary/internal/testutil/httptestx": 1000,
    "backend-integration::backend-integration-testutil::github.com/JochiRaider/cartulary/internal/testutil/pgtest": 5000
  },
  "tests": {
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_1_05": 100,
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_LoginSessionLifecycle_I_1_01": 1000,
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_UserAdminAudit_I_1_03": 1000
  }
}
JSON

set +e
underplanned_components_output="$("$NODE_BIN" "$DRIFT_SCRIPT" --baseline-file "$tmp_dir/underplanned-components.json" "$results_dir" 2>&1)"
underplanned_components_status=$?
set -e
if [[ "$underplanned_components_status" -eq 0 ]]; then
  fail "underplanned component drift should fail"
fi
assert_contains "$underplanned_components_output" "underplanned shard=backend-integration-auth-shard-01" "underplanned package and command drift"
assert_contains "$underplanned_components_output" "underplanned shard=backend-store-shard-01" "underplanned test drift"
assert_contains "$underplanned_components_output" "planned_tests_ms=" "underplanned component planned test detail"
assert_contains "$underplanned_components_output" "actual_tests_ms=" "underplanned component actual test detail"
assert_contains "$underplanned_components_output" "planned_package_overhead_ms=" "underplanned component planned package detail"
assert_contains "$underplanned_components_output" "actual_command_overhead_ms=" "underplanned component actual command detail"

service_contaminated_results="$tmp_dir/service-contaminated-results"
cp -R "$results_dir" "$service_contaminated_results"
mkdir -p "$service_contaminated_results/run-a/_shared/test-services/suite-retry"
cat >"$service_contaminated_results/run-a/_shared/test-services/suite-retry/service-scope.json" <<'JSON'
{
  "postgres": {
    "startup": {
      "retry_count": 1,
      "final_status": "pass"
    }
  },
  "minio": {
    "startup": {
      "retry_count": 0,
      "final_status": "pass"
    }
  }
}
JSON
set +e
service_contaminated_update_output="$("$NODE_BIN" "$UPDATE_SCRIPT" --baseline-file "$tmp_dir/service-contaminated-update.json" "$service_contaminated_results" 2>&1)"
service_contaminated_update_status=$?
set -e
if [[ "$service_contaminated_update_status" -eq 0 ]]; then
  fail "contaminated Go duration baseline refresh should fail"
fi
assert_contains "$service_contaminated_update_output" "Refusing to refresh Go test duration baselines from contaminated service timing evidence" "contaminated Go refresh output"
assert_contains "$service_contaminated_update_output" "service_startup_retry service=postgres retries=1" "contaminated Go refresh reason"

service_contaminated_drift_output="$("$NODE_BIN" "$DRIFT_SCRIPT" --baseline-file "$tmp_dir/underplanned-components.json" "$service_contaminated_results" 2>&1 >/dev/null)"
assert_contains "$service_contaminated_drift_output" "ignored underplanned contaminated shard=backend-integration-auth-shard-01" "contaminated service timing underplanned shard warning"
assert_contains "$service_contaminated_drift_output" "service_timing_contamination=[service_startup_retry service=postgres retries=1" "contaminated service timing reason"

cat >"$tmp_dir/overplanned.json" <<'JSON'
{
  "schema_id": "cartulary.go_test_duration_baselines.v4",
  "command_overheads_by_target": {
    "backend-integration": 80000,
    "backend-store": 80000
  },
  "package_overheads": {
    "backend-integration::github.com/JochiRaider/cartulary/internal/modules/auth": 80000,
    "backend-store::github.com/JochiRaider/cartulary/internal/modules/auth": 80000
  },
  "raw_aggregates": {
    "backend-integration::backend-integration-testutil::github.com/JochiRaider/cartulary/internal/testutil/httptestx": 1000,
    "backend-integration::backend-integration-testutil::github.com/JochiRaider/cartulary/internal/testutil/pgtest": 5000
  },
  "tests": {
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_1_05": 80000,
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_LoginSessionLifecycle_I_1_01": 80000,
    "github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_UserAdminAudit_I_1_03": 80000
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

write_empty_baseline "$tmp_dir/missing.json"

set +e
missing_output="$("$NODE_BIN" "$DRIFT_SCRIPT" --baseline-file "$tmp_dir/missing.json" "$results_dir" 2>&1)"
missing_status=$?
set -e
if [[ "$missing_status" -eq 0 ]]; then
  fail "missing baselines should fail"
fi
assert_contains "$missing_output" "missing test baseline" "missing test drift"
assert_contains "$missing_output" "missing package overhead baseline" "missing package overhead drift"
assert_contains "$missing_output" "missing command overhead baseline" "missing command overhead drift"
assert_contains "$missing_output" "missing raw package baseline" "missing raw drift"

cp "$ROOT_DIR/tools/go_test_duration_baselines.json" "$tmp_dir/coverage-complete.json"
"$NODE_BIN" "$COVERAGE_SCRIPT" --baseline-file "$tmp_dir/coverage-complete.json" >/dev/null

"$NODE_BIN" - "$tmp_dir/coverage-complete.json" "$tmp_dir/coverage-missing-test.json" <<'EOF'
const fs = require("node:fs");
const [source, target] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(source, "utf8"));
delete baseline.tests["github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_LoginSessionLifecycle_I_1_01"];
fs.writeFileSync(target, `${JSON.stringify(baseline, null, 2)}\n`);
EOF

set +e
missing_test_coverage_output="$("$NODE_BIN" "$COVERAGE_SCRIPT" --baseline-file "$tmp_dir/coverage-missing-test.json" 2>&1)"
missing_test_coverage_status=$?
set -e
if [[ "$missing_test_coverage_status" -eq 0 ]]; then
  fail "missing test baseline coverage should fail"
fi
assert_contains "$missing_test_coverage_output" "missing test baseline key=github.com/JochiRaider/cartulary/internal/modules/auth::TestPhase1_LoginSessionLifecycle_I_1_01" "missing test baseline coverage"

"$NODE_BIN" - "$tmp_dir/coverage-complete.json" "$tmp_dir/coverage-missing-raw.json" <<'EOF'
const fs = require("node:fs");
const [source, target] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(source, "utf8"));
delete baseline.raw_aggregates["backend-integration::backend-integration-testutil::github.com/JochiRaider/cartulary/internal/testutil/pgtest"];
fs.writeFileSync(target, `${JSON.stringify(baseline, null, 2)}\n`);
EOF

set +e
missing_raw_coverage_output="$("$NODE_BIN" "$COVERAGE_SCRIPT" --baseline-file "$tmp_dir/coverage-missing-raw.json" 2>&1)"
missing_raw_coverage_status=$?
set -e
if [[ "$missing_raw_coverage_status" -eq 0 ]]; then
  fail "missing raw package baseline coverage should fail"
fi
assert_contains "$missing_raw_coverage_output" "missing raw package baseline key=backend-integration::backend-integration-testutil::github.com/JochiRaider/cartulary/internal/testutil/pgtest" "missing raw baseline coverage"

"$NODE_BIN" - "$tmp_dir/coverage-complete.json" "$tmp_dir/coverage-missing-package.json" <<'EOF'
const fs = require("node:fs");
const [source, target] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(source, "utf8"));
delete baseline.package_overheads["backend-integration::github.com/JochiRaider/cartulary/internal/modules/auth"];
fs.writeFileSync(target, `${JSON.stringify(baseline, null, 2)}\n`);
EOF

set +e
missing_package_coverage_output="$("$NODE_BIN" "$COVERAGE_SCRIPT" --baseline-file "$tmp_dir/coverage-missing-package.json" 2>&1)"
missing_package_coverage_status=$?
set -e
if [[ "$missing_package_coverage_status" -eq 0 ]]; then
  fail "missing package overhead baseline coverage should fail"
fi
assert_contains "$missing_package_coverage_output" "missing package overhead baseline key=backend-integration::github.com/JochiRaider/cartulary/internal/modules/auth" "missing package overhead coverage"

"$NODE_BIN" - "$tmp_dir/coverage-complete.json" "$tmp_dir/coverage-missing-command.json" <<'EOF'
const fs = require("node:fs");
const [source, target] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(source, "utf8"));
delete baseline.command_overheads_by_target["backend-store"];
fs.writeFileSync(target, `${JSON.stringify(baseline, null, 2)}\n`);
EOF

set +e
missing_command_coverage_output="$("$NODE_BIN" "$COVERAGE_SCRIPT" --baseline-file "$tmp_dir/coverage-missing-command.json" 2>&1)"
missing_command_coverage_status=$?
set -e
if [[ "$missing_command_coverage_status" -eq 0 ]]; then
  fail "missing command overhead baseline coverage should fail"
fi
assert_contains "$missing_command_coverage_output" "missing command overhead baseline target=backend-store" "missing command overhead coverage"
