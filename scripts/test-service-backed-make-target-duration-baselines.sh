#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
MAKE_HELPER="${MAKE:-make}"
SCRIPT="$ROOT_DIR/scripts/service-backed-make-target-durations.mjs"

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

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/service-backed-make-target-duration-baselines.XXXXXX")"
trap 'rm -rf "$tmp_dir"' EXIT

results_dir="$tmp_dir/results"
mkdir -p "$results_dir/run-a/check-service-backed" "$results_dir/run-b/test-service-backed" "$results_dir/run-c/check-service-backed"

cat >"$results_dir/run-a/check-service-backed/scheduler-summary.json" <<'JSON'
{
  "schema_id": "cartulary.service_backed_scheduler_summary.v5",
  "target": "check-service-backed",
  "status": "pass",
  "scheduler_kind": "service-backed"
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
  "schema_id": "cartulary.service_backed_scheduler_summary.v5",
  "target": "test-service-backed",
  "status": "pass",
  "scheduler_kind": "service-backed"
}
JSON
cat >"$results_dir/run-b/test-service-backed/scheduler-events.jsonl" <<'JSONL'
{"event":"start","work_unit_id":"backend-process","work_unit":"backend-process","work_unit_type":"make_target","aggregate_target":"backend-process"}
{"event":"finish","work_unit_id":"backend-process","work_unit":"backend-process","status":0,"duration_ms":9000}
JSONL

cat >"$results_dir/run-c/check-service-backed/scheduler-summary.json" <<'JSON'
{
  "schema_id": "cartulary.service_backed_scheduler_summary.v5",
  "target": "check-service-backed",
  "status": "fail",
  "scheduler_kind": "service-backed"
}
JSON
cat >"$results_dir/run-c/check-service-backed/scheduler-events.jsonl" <<'JSONL'
{"event":"start","work_unit_id":"browser-e2e","work_unit":"browser-e2e","work_unit_type":"make_target","aggregate_target":"browser-e2e"}
{"event":"finish","work_unit_id":"browser-e2e","work_unit":"browser-e2e","status":0,"duration_ms":99999}
JSONL

cat >"$tmp_dir/baseline.json" <<'JSON'
{
  "schema_id": "cartulary.service_backed_make_target_duration_baselines.v1",
  "default_make_target_weight_ms": 10000,
  "targets": {}
}
JSON
cat >"$tmp_dir/profile.json" <<'JSON'
{
  "schema_id": "cartulary.service_backed_schedule_profiles.v3",
  "defaults": {
    "make_target_duration_baseline": "baseline.json",
    "make_target_weight_overrides": {}
  },
  "schedules": []
}
JSON
cat >"$tmp_dir/schedule.json" <<'JSON'
{
  "schema_id": "cartulary.service_backed_schedule.v8",
  "schedules": [
    {
      "target": "check-service-backed",
      "work_unit_sources": [
        { "type": "make_target", "target": "backend-process" },
        { "type": "make_target", "target": "browser-e2e-webserver-backed" }
      ]
    }
  ]
}
JSON

update_output="$("$NODE_BIN" "$SCRIPT" update --baseline-file "$tmp_dir/baseline.json" "$results_dir" 2>&1)"
assert_contains "$update_output" "updated 2 service-backed make-target duration baselines from 2 successful scheduler artifact(s)" "baseline update output"

"$NODE_BIN" - "$tmp_dir/baseline.json" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
if (baseline.schema_id !== "cartulary.service_backed_make_target_duration_baselines.v1") {
  throw new Error(`unexpected schema ${baseline.schema_id}`);
}
if (baseline.targets["backend-process"] !== 9000) {
  throw new Error(`expected max backend-process duration 9000, got ${baseline.targets["backend-process"]}`);
}
if (baseline.targets["browser-e2e-webserver-backed"] !== 27600) {
  throw new Error(`expected browser duration 27600, got ${baseline.targets["browser-e2e-webserver-backed"]}`);
}
for (const ignored of ["backend-store", "failed-target", "browser-e2e"]) {
  if (Object.hasOwn(baseline.targets, ignored)) {
    throw new Error(`unexpected ignored target baseline ${ignored}`);
  }
}
EOF

"$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/baseline.json" --profile "$tmp_dir/profile.json" --schedule-manifest "$tmp_dir/schedule.json" "$results_dir" >/dev/null

cat >"$tmp_dir/make-baseline.json" <<'JSON'
{
  "schema_id": "cartulary.service_backed_make_target_duration_baselines.v1",
  "default_make_target_weight_ms": 10000,
  "targets": {}
}
JSON
make_update_output="$(
  RESULTS_DIR="$results_dir" \
  SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE="$tmp_dir/make-baseline.json" \
    "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" service-backed-make-target-duration-baselines 2>&1
)"
assert_contains "$make_update_output" "updated 2 service-backed make-target duration baselines" "make baseline update output"
RESULTS_DIR="$results_dir" \
SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE="$tmp_dir/make-baseline.json" \
SERVICE_BACKED_SCHEDULE_PROFILE="$tmp_dir/profile.json" \
SERVICE_BACKED_SCHEDULE_MANIFEST="$tmp_dir/schedule.json" \
  "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" service-backed-make-target-duration-baseline-drift >/dev/null

cat >"$tmp_dir/missing.json" <<'JSON'
{
  "schema_id": "cartulary.service_backed_make_target_duration_baselines.v1",
  "default_make_target_weight_ms": 10000,
  "targets": {
    "backend-process": 9000
  }
}
JSON
set +e
missing_output="$("$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/missing.json" --profile "$tmp_dir/profile.json" --schedule-manifest "$tmp_dir/schedule.json" "$results_dir" 2>&1)"
missing_status=$?
set -e
if [[ "$missing_status" -eq 0 ]]; then
  fail "missing baseline drift should fail"
fi
assert_contains "$missing_output" "missing make-target baseline target=browser-e2e-webserver-backed" "missing baseline drift"

cat >"$tmp_dir/underplanned.json" <<'JSON'
{
  "schema_id": "cartulary.service_backed_make_target_duration_baselines.v1",
  "default_make_target_weight_ms": 10000,
  "targets": {
    "backend-process": 100,
    "browser-e2e-webserver-backed": 27600
  }
}
JSON
set +e
underplanned_output="$("$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/underplanned.json" --profile "$tmp_dir/profile.json" --schedule-manifest "$tmp_dir/schedule.json" "$results_dir" 2>&1)"
underplanned_status=$?
set -e
if [[ "$underplanned_status" -eq 0 ]]; then
  fail "underplanned baseline drift should fail"
fi
assert_contains "$underplanned_output" "underplanned target=backend-process" "underplanned baseline drift"

cat >"$tmp_dir/overplanned.json" <<'JSON'
{
  "schema_id": "cartulary.service_backed_make_target_duration_baselines.v1",
  "default_make_target_weight_ms": 10000,
  "targets": {
    "backend-process": 9000,
    "browser-e2e-webserver-backed": 100000
  }
}
JSON
set +e
overplanned_output="$("$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/overplanned.json" --profile "$tmp_dir/profile.json" --schedule-manifest "$tmp_dir/schedule.json" "$results_dir" 2>&1)"
overplanned_status=$?
set -e
if [[ "$overplanned_status" -eq 0 ]]; then
  fail "overplanned baseline drift should fail"
fi
assert_contains "$overplanned_output" "overplanned target=browser-e2e-webserver-backed" "overplanned baseline drift"

cat >"$tmp_dir/expired-profile.json" <<'JSON'
{
  "schema_id": "cartulary.service_backed_schedule_profiles.v3",
  "defaults": {
    "make_target_duration_baseline": "baseline.json",
    "make_target_weight_overrides": {
      "backend-process": {
        "weight_ms": 9000,
        "reason": "temporary timing investigation",
        "expires_at": "2026-01-01T00:00:00.000Z"
      }
    }
  },
  "schedules": []
}
JSON
set +e
expired_output="$("$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/baseline.json" --profile "$tmp_dir/expired-profile.json" --schedule-manifest "$tmp_dir/schedule.json" "$results_dir" 2>&1)"
expired_status=$?
set -e
if [[ "$expired_status" -eq 0 ]]; then
  fail "expired override drift should fail"
fi
assert_contains "$expired_output" "override target=backend-process expired" "expired override drift"

cat >"$tmp_dir/unknown-override-profile.json" <<'JSON'
{
  "schema_id": "cartulary.service_backed_schedule_profiles.v3",
  "defaults": {
    "make_target_duration_baseline": "baseline.json",
    "make_target_weight_overrides": {
      "removed-target": {
        "weight_ms": 9000,
        "reason": "temporary timing investigation",
        "expires_at": "2099-01-01T00:00:00.000Z"
      }
    }
  },
  "schedules": []
}
JSON
set +e
unknown_override_output="$("$NODE_BIN" "$SCRIPT" check-drift --baseline-file "$tmp_dir/baseline.json" --profile "$tmp_dir/unknown-override-profile.json" --schedule-manifest "$tmp_dir/schedule.json" "$results_dir" 2>&1)"
unknown_override_status=$?
set -e
if [[ "$unknown_override_status" -eq 0 ]]; then
  fail "unknown override drift should fail"
fi
assert_contains "$unknown_override_output" "override target=removed-target is not present in generated service-backed schedules" "unknown override drift"
