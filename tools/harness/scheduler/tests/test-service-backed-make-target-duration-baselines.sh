#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
MAKE_HELPER="${MAKE:-make}"
SCRIPT="$ROOT_DIR/tools/harness/scheduler/service-backed-make-target-durations-cli.mjs"

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

assert_fails_with() {
  local label="$1"
  local needle="$2"
  shift 2

  set +e
  local output
  output="$("$@" 2>&1)"
  local status=$?
  set -e
  if [[ "$status" -eq 0 ]]; then
    fail "$label: expected command to fail"
  fi
  assert_contains "$output" "$needle" "$label"
}

phase_stdout_from_result() {
  local output="$1"
  local root
  root="$(printf '%s\n' "$output" | sed -n 's/.* run_root=\([^ ]*\) .*/\1/p' | head -n 1)"
  if [[ -z "$root" ]]; then
    fail "missing run_root in output: $output"
  fi
  if [[ "$root" = /* ]]; then
    local dir="$root"
  else
    local dir="$ROOT_DIR/$root"
  fi
  local target
  target="$(printf '%s\n' "$output" | sed -n 's/.* target=\([^ ]*\) .*/\1/p' | head -n 1)"
  [[ -f "$dir/stdout.log" ]] && cat "$dir/stdout.log"
  if [[ -n "$target" ]]; then
    [[ -f "$dir/$target/stdout.log" ]] && cat "$dir/$target/stdout.log"
  fi
}

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/service-backed-make-target-duration-baselines.XXXXXX")"
trap 'rm -rf "$tmp_dir"' EXIT

results_dir="$tmp_dir/results"
mkdir -p "$results_dir/run-a/check-service-backed" "$results_dir/run-b/test-service-backed" "$results_dir/run-c/check-service-backed" "$results_dir/run-d/check"

cat >"$results_dir/run-a/check-service-backed/scheduler-summary.json" <<'JSON'
{
  "schema_id": "cartulary.service_backed_scheduler_summary.v10",
  "target": "check-service-backed",
  "status": "pass",
  "scheduler_kind": "service_backed"
}
JSON
cat >"$results_dir/run-a/check-service-backed/scheduler-events.jsonl" <<'JSONL'
{"event":"start","work_unit_id":"backend-process","work_unit":"backend-process","work_unit_type":"make_target","aggregate_target":"backend-process"}
{"event":"finish","work_unit_id":"backend-process","work_unit":"backend-process","status":0,"duration_ms":7000}
{"event":"start","work_unit_id":"browser-e2e-webserver-backed","work_unit":"browser-e2e-webserver-backed","work_unit_type":"make_target","aggregate_target":"browser-e2e-webserver-backed"}
{"event":"finish","work_unit_id":"browser-e2e-webserver-backed","work_unit":"browser-e2e-webserver-backed","status":0,"duration_ms":27600}
{"event":"start","work_unit_id":"backend-store:backend-store-shard-01","work_unit":"backend-store/backend-store-shard-01","work_unit_type":"go_shard","aggregate_target":"backend-store"}
{"event":"finish","work_unit_id":"backend-store:backend-store-shard-01","work_unit":"backend-store/backend-store-shard-01","status":0,"duration_ms":90000}
{"event":"start","work_unit_id":"failed-target","work_unit":"failed-target","work_unit_type":"make_target","aggregate_target":"failed-target"}
{"event":"finish","work_unit_id":"failed-target","work_unit":"failed-target","status":1,"duration_ms":50000}
{"event":"finalize-finish","finalizer":"backend-store","finalizer_id":"finalize:backend-store","status":0,"duration_ms":40000}
JSONL

cat >"$results_dir/run-b/test-service-backed/scheduler-summary.json" <<'JSON'
{
  "schema_id": "cartulary.service_backed_scheduler_summary.v10",
  "target": "test-service-backed",
  "status": "pass",
  "scheduler_kind": "service_backed"
}
JSON
cat >"$results_dir/run-b/test-service-backed/scheduler-events.jsonl" <<'JSONL'
{"event":"start","work_unit_id":"backend-process","work_unit":"backend-process","work_unit_type":"make_target","aggregate_target":"backend-process"}
{"event":"finish","work_unit_id":"backend-process","work_unit":"backend-process","status":0,"duration_ms":9000}
JSONL

cat >"$results_dir/run-c/check-service-backed/scheduler-summary.json" <<'JSON'
{
  "schema_id": "cartulary.service_backed_scheduler_summary.v10",
  "target": "check-service-backed",
  "status": "fail",
  "scheduler_kind": "service_backed"
}
JSON
cat >"$results_dir/run-c/check-service-backed/scheduler-events.jsonl" <<'JSONL'
{"event":"start","work_unit_id":"browser-e2e","work_unit":"browser-e2e","work_unit_type":"make_target","aggregate_target":"browser-e2e"}
{"event":"finish","work_unit_id":"browser-e2e","work_unit":"browser-e2e","status":0,"duration_ms":99999}
JSONL

cat >"$results_dir/run-d/check/scheduler-summary.json" <<'JSON'
{
  "schema_id": "cartulary.check_scheduler_summary.v10",
  "target": "check",
  "status": "pass",
  "scheduler_kind": "check"
}
JSON
cat >"$results_dir/run-d/check/scheduler-events.jsonl" <<'JSONL'
{"event":"start","work_unit_id":"check-service-backed","work_unit":"check-service-backed","work_unit_type":"make_target","aggregate_target":"check-service-backed"}
{"event":"finish","work_unit_id":"check-service-backed","work_unit":"check-service-backed","status":0,"duration_ms":121233}
JSONL

cat >"$tmp_dir/baseline.json" <<'JSON'
{
  "schema_id": "cartulary.scheduler_work_unit_duration_baselines.v2",
  "default_work_unit_weight_ms": 10000,
  "work_units": {}
}
JSON
cp "$ROOT_DIR/tools/execution_topology_manifest.json" "$tmp_dir/topology.json"
cp "$ROOT_DIR/tools/service_backed_make_target_duration_baselines.json" \
  "$tmp_dir/service_backed_make_target_duration_baselines.json"
cat >"$tmp_dir/schedule.json" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check-service-backed",
      "scheduler_kind": "service_backed",
      "work_units": [
        { "id": "backend-process", "kind": "make_target", "target": "backend-process", "aggregate_target": "backend-process" },
        { "id": "browser-e2e-webserver-backed", "kind": "browser_group", "target": "browser-e2e-webserver-backed", "aggregate_target": "browser-e2e-webserver-backed" }
      ]
    }
  ]
}
JSON

update_output="$("$NODE_BIN" "$SCRIPT" update --baseline-file "$tmp_dir/baseline.json" "$results_dir" 2>&1)"
assert_contains "$update_output" "updated 4 scheduler work-unit duration baselines from 3 successful scheduler artifact(s)" "baseline update output"

assert_fails_with \
  "update rejects topology flag" \
  "usage:" \
  "$NODE_BIN" "$SCRIPT" update --baseline-file "$tmp_dir/baseline.json" --topology "$tmp_dir/topology.json" "$results_dir"
assert_fails_with \
  "missing schedule manifest flag value shows usage" \
  "usage:" \
  "$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/baseline.json" --topology "$tmp_dir/topology.json" --schedule-manifest
assert_fails_with \
  "multiple service-backed results dirs are rejected" \
  "usage:" \
  "$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/baseline.json" --topology "$tmp_dir/topology.json" --schedule-manifest "$tmp_dir/schedule.json" "$results_dir" "$results_dir"
assert_fails_with \
  "duplicate topology flags are rejected" \
  "usage:" \
  "$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/baseline.json" --topology "$tmp_dir/topology.json" --topology "$tmp_dir/topology.json" --schedule-manifest "$tmp_dir/schedule.json" "$results_dir"

"$NODE_BIN" - "$tmp_dir/baseline.json" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
if (baseline.schema_id !== "cartulary.scheduler_work_unit_duration_baselines.v2") {
  throw new Error(`unexpected schema ${baseline.schema_id}`);
}
const read = (key) => baseline.work_units[key]?.weight_ms;
if (read("service_backed|test-service-backed|backend-process|backend-process") !== 9000) {
  throw new Error(`expected test-service-backed backend-process duration 9000, got ${read("service_backed|test-service-backed|backend-process|backend-process")}`);
}
if (read("service_backed|check-service-backed|backend-process|backend-process") !== 7000) {
  throw new Error(`expected check-service-backed backend-process duration 7000, got ${read("service_backed|check-service-backed|backend-process|backend-process")}`);
}
if (read("service_backed|check-service-backed|browser-e2e-webserver-backed|browser-e2e-webserver-backed") !== 27600) {
  throw new Error(`expected browser duration 27600, got ${read("service_backed|check-service-backed|browser-e2e-webserver-backed|browser-e2e-webserver-backed")}`);
}
if (read("check|check|check-service-backed|check-service-backed") !== 121233) {
  throw new Error(`expected parent check-service-backed duration 121233, got ${read("check|check|check-service-backed|check-service-backed")}`);
}
for (const ignored of ["backend-store", "failed-target", "browser-e2e"]) {
  if (Object.keys(baseline.work_units).some((key) => key.includes(`|${ignored}|`))) {
    throw new Error(`unexpected ignored target baseline ${ignored}`);
  }
}
EOF

"$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/baseline.json" --topology "$tmp_dir/topology.json" --schedule-manifest "$tmp_dir/schedule.json" "$results_dir" >/dev/null

cat >"$tmp_dir/make-baseline.json" <<'JSON'
{
  "schema_id": "cartulary.scheduler_work_unit_duration_baselines.v2",
  "default_work_unit_weight_ms": 10000,
  "work_units": {}
}
JSON
make_update_output="$(
  env -u CARTULARY_TEST_RESULTS_DIR -u CARTULARY_TEST_RUN_ID \
    RESULTS_DIR="$results_dir" \
    SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE="$tmp_dir/make-baseline.json" \
    "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" service-backed-make-target-duration-baselines 2>&1
)"
assert_contains "$make_update_output" "[RESULT] target=service-backed-make-target-duration-baselines status=pass" "make baseline update summary"
assert_contains "$(phase_stdout_from_result "$make_update_output")" "updated 4 scheduler work-unit duration baselines" "make baseline update output"
env -u CARTULARY_TEST_RESULTS_DIR -u CARTULARY_TEST_RUN_ID \
  RESULTS_DIR="$results_dir" \
  SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE="$tmp_dir/make-baseline.json" \
  EXECUTION_TOPOLOGY_MANIFEST="$tmp_dir/topology.json" \
  SCHEDULER_MANIFEST="$tmp_dir/schedule.json" \
  "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" service-backed-make-target-duration-baseline-drift >/dev/null

cat >"$tmp_dir/tolerated-underplanned.json" <<'JSON'
{
  "schema_id": "cartulary.scheduler_work_unit_duration_baselines.v2",
  "default_work_unit_weight_ms": 10000,
  "work_units": {
    "check|check|check-service-backed|check-service-backed": {
      "scheduler_kind": "check",
      "schedule_target": "check",
      "work_unit_id": "check-service-backed",
      "aggregate_target": "check-service-backed",
      "weight_ms": 121233
    },
    "service_backed|check-service-backed|backend-process|backend-process": {
      "scheduler_kind": "service_backed",
      "schedule_target": "check-service-backed",
      "work_unit_id": "backend-process",
      "aggregate_target": "backend-process",
      "weight_ms": 9000
    },
    "service_backed|check-service-backed|browser-e2e-webserver-backed|browser-e2e-webserver-backed": {
      "scheduler_kind": "service_backed",
      "schedule_target": "check-service-backed",
      "work_unit_id": "browser-e2e-webserver-backed",
      "aggregate_target": "browser-e2e-webserver-backed",
      "weight_ms": 12000
    }
  }
}
JSON
"$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/tolerated-underplanned.json" --topology "$tmp_dir/topology.json" --schedule-manifest "$tmp_dir/schedule.json" "$results_dir" >/dev/null

cat >"$tmp_dir/missing.json" <<'JSON'
{
  "schema_id": "cartulary.scheduler_work_unit_duration_baselines.v2",
  "default_work_unit_weight_ms": 10000,
  "work_units": {
    "check|check|check-service-backed|check-service-backed": {
      "scheduler_kind": "check",
      "schedule_target": "check",
      "work_unit_id": "check-service-backed",
      "aggregate_target": "check-service-backed",
      "weight_ms": 121233
    },
    "service_backed|check-service-backed|backend-process|backend-process": {
      "scheduler_kind": "service_backed",
      "schedule_target": "check-service-backed",
      "work_unit_id": "backend-process",
      "aggregate_target": "backend-process",
      "weight_ms": 9000
    }
  }
}
JSON
set +e
missing_output="$("$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/missing.json" --topology "$tmp_dir/topology.json" --schedule-manifest "$tmp_dir/schedule.json" "$results_dir" 2>&1)"
missing_status=$?
set -e
if [[ "$missing_status" -eq 0 ]]; then
  fail "missing baseline drift should fail"
fi
assert_contains "$missing_output" "missing scheduler work-unit baseline scheduler=service_backed schedule=check-service-backed work_unit=browser-e2e-webserver-backed aggregate=browser-e2e-webserver-backed" "missing baseline drift"

cat >"$tmp_dir/underplanned.json" <<'JSON'
{
  "schema_id": "cartulary.scheduler_work_unit_duration_baselines.v2",
  "default_work_unit_weight_ms": 10000,
  "work_units": {
    "check|check|check-service-backed|check-service-backed": {
      "scheduler_kind": "check",
      "schedule_target": "check",
      "work_unit_id": "check-service-backed",
      "aggregate_target": "check-service-backed",
      "weight_ms": 121233
    },
    "service_backed|check-service-backed|backend-process|backend-process": {
      "scheduler_kind": "service_backed",
      "schedule_target": "check-service-backed",
      "work_unit_id": "backend-process",
      "aggregate_target": "backend-process",
      "weight_ms": 9000
    },
    "service_backed|check-service-backed|browser-e2e-webserver-backed|browser-e2e-webserver-backed": {
      "scheduler_kind": "service_backed",
      "schedule_target": "check-service-backed",
      "work_unit_id": "browser-e2e-webserver-backed",
      "aggregate_target": "browser-e2e-webserver-backed",
      "weight_ms": 100
    }
  }
}
JSON
set +e
underplanned_output="$("$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/underplanned.json" --topology "$tmp_dir/topology.json" --schedule-manifest "$tmp_dir/schedule.json" "$results_dir" 2>&1)"
underplanned_status=$?
set -e
if [[ "$underplanned_status" -eq 0 ]]; then
  fail "underplanned baseline drift should fail"
fi
assert_contains "$underplanned_output" "underplanned scheduler=service_backed schedule=check-service-backed work_unit=browser-e2e-webserver-backed aggregate=browser-e2e-webserver-backed" "underplanned baseline drift"

contaminated_results_dir="$tmp_dir/contaminated-results"
cp -R "$results_dir" "$contaminated_results_dir"
mkdir -p "$contaminated_results_dir/run-a/_shared/test-services/suite-retry"
cat >"$contaminated_results_dir/run-a/_shared/test-services/suite-retry/service-scope.json" <<'JSON'
{
  "postgres": {
    "startup": {
      "retry_count": 1,
      "final_status": "pass"
    }
  },
  "object_store": {
    "startup": {
      "retry_count": 0,
      "final_status": "pass"
    }
  }
}
JSON
cat >"$tmp_dir/contaminated-update-baseline.json" <<'JSON'
{
  "schema_id": "cartulary.scheduler_work_unit_duration_baselines.v2",
  "default_work_unit_weight_ms": 10000,
  "work_units": {}
}
JSON
set +e
contaminated_update_output="$("$NODE_BIN" "$SCRIPT" update --baseline-file "$tmp_dir/contaminated-update-baseline.json" "$contaminated_results_dir" 2>&1)"
contaminated_update_status=$?
set -e
if [[ "$contaminated_update_status" -eq 0 ]]; then
  fail "contaminated service-backed baseline refresh should fail"
fi
assert_contains "$contaminated_update_output" "Refusing to refresh service-backed make-target duration baselines from contaminated service timing evidence" "contaminated service-backed refresh output"
assert_contains "$contaminated_update_output" "service_startup_retry service=postgres retries=1" "contaminated service-backed refresh reason"

contaminated_drift_output="$("$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/underplanned.json" --topology "$tmp_dir/topology.json" --schedule-manifest "$tmp_dir/schedule.json" "$contaminated_results_dir" 2>&1 >/dev/null)"
assert_contains "$contaminated_drift_output" "ignored underplanned contaminated scheduler=service_backed schedule=check-service-backed work_unit=browser-e2e-webserver-backed aggregate=browser-e2e-webserver-backed" "contaminated service-backed underplanned drift warning"
assert_contains "$contaminated_drift_output" "service_timing_contamination=[service_startup_retry service=postgres retries=1" "contaminated service-backed warning reason"

cat >"$tmp_dir/overplanned.json" <<'JSON'
{
  "schema_id": "cartulary.scheduler_work_unit_duration_baselines.v2",
  "default_work_unit_weight_ms": 10000,
  "work_units": {
    "check|check|check-service-backed|check-service-backed": {
      "scheduler_kind": "check",
      "schedule_target": "check",
      "work_unit_id": "check-service-backed",
      "aggregate_target": "check-service-backed",
      "weight_ms": 121233
    },
    "service_backed|check-service-backed|backend-process|backend-process": {
      "scheduler_kind": "service_backed",
      "schedule_target": "check-service-backed",
      "work_unit_id": "backend-process",
      "aggregate_target": "backend-process",
      "weight_ms": 9000
    },
    "service_backed|check-service-backed|browser-e2e-webserver-backed|browser-e2e-webserver-backed": {
      "scheduler_kind": "service_backed",
      "schedule_target": "check-service-backed",
      "work_unit_id": "browser-e2e-webserver-backed",
      "aggregate_target": "browser-e2e-webserver-backed",
      "weight_ms": 150000
    }
  }
}
JSON
set +e
overplanned_output="$("$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/overplanned.json" --topology "$tmp_dir/topology.json" --schedule-manifest "$tmp_dir/schedule.json" "$results_dir" 2>&1)"
overplanned_status=$?
set -e
if [[ "$overplanned_status" -eq 0 ]]; then
  fail "overplanned baseline drift should fail"
fi
assert_contains "$overplanned_output" "overplanned scheduler=service_backed schedule=check-service-backed work_unit=browser-e2e-webserver-backed aggregate=browser-e2e-webserver-backed" "overplanned baseline drift"

"$NODE_BIN" - "$tmp_dir/topology.json" "$tmp_dir/expired-topology.json" <<'EOF'
const fs = require("node:fs");
const [source, output] = process.argv.slice(2);
const topology = JSON.parse(fs.readFileSync(source, "utf8"));
topology.service_backed_schedules.defaults.make_target_weight_overrides = {
  "backend-process": {
    weight_ms: 9000,
    reason: "temporary timing investigation",
    expires_at: "2026-01-01T00:00:00.000Z",
  },
};
fs.writeFileSync(output, `${JSON.stringify(topology, null, 2)}\n`);
EOF
set +e
expired_output="$("$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/baseline.json" --topology "$tmp_dir/expired-topology.json" --schedule-manifest "$tmp_dir/schedule.json" "$results_dir" 2>&1)"
expired_status=$?
set -e
if [[ "$expired_status" -eq 0 ]]; then
  fail "expired override drift should fail"
fi
assert_contains "$expired_output" "override target=backend-process expired" "expired override drift"

"$NODE_BIN" - "$tmp_dir/topology.json" "$tmp_dir/unknown-override-topology.json" <<'EOF'
const fs = require("node:fs");
const [source, output] = process.argv.slice(2);
const topology = JSON.parse(fs.readFileSync(source, "utf8"));
topology.service_backed_schedules.defaults.make_target_weight_overrides = {
  "removed-target": {
    weight_ms: 9000,
    reason: "temporary timing investigation",
    expires_at: "2099-01-01T00:00:00.000Z",
  },
};
fs.writeFileSync(output, `${JSON.stringify(topology, null, 2)}\n`);
EOF
set +e
unknown_override_output="$("$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/baseline.json" --topology "$tmp_dir/unknown-override-topology.json" --schedule-manifest "$tmp_dir/schedule.json" "$results_dir" 2>&1)"
unknown_override_status=$?
set -e
if [[ "$unknown_override_status" -eq 0 ]]; then
  fail "unknown override drift should fail"
fi
assert_contains "$unknown_override_output" "override target=removed-target is not present in generated service-backed schedules" "unknown override drift"
