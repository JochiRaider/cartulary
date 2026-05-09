#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
CHECKER="$ROOT_DIR/scripts/check-json-shapes.mjs"
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
    fail "$label: expected output to contain [$needle], got [$haystack]"
  fi
}

assert_passes() {
  local label="$1"
  shift

  local output
  if ! output="$("$@" 2>&1)"; then
    fail "$label: expected success, got output: $output"
  fi
  printf '%s' "$output"
}

assert_fails() {
  local label="$1"
  shift

  local output
  local status
  set +e
  output="$("$@" 2>&1)"
  status=$?
  set -e

  if [[ "$status" -eq 0 ]]; then
    fail "$label: expected failure"
  fi
  printf '%s' "$output"
}

run_shape_check() {
  local kind="$1"
  local file="$2"

  "$NODE_BIN" "$CHECKER" --kind "$kind" --file "$file" --root "$tmp_dir"
}

write_valid_phase_map() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v1",
  "phase": "phase9",
  "note": "Synthetic shape fixture.",
  "ledger": {
    "title": "Phase 9 Coverage Ledger",
    "notes": "Synthetic shape fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase9",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-P9-001"],
  "forbidden_id_files": [],
  "support_go_targets": [],
  "unit": [],
  "integration": [],
  "e2e": []
}
JSON
}

write_valid_phase_registry() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.phase_registry.v1",
  "phases": [
    {
      "phase": "phase9",
      "order": 9,
      "status": "planned",
      "label": "Phase 9",
      "manifest_path": "tools/phase9_test_map.json",
      "ledger_path": "docs/testing/phase9_coverage_ledger.md",
      "scope": "Synthetic shape fixture.",
      "normative_owners": "docs/spec/00_document_set_status_and_precedence.md"
    }
  ]
}
JSON
}

write_valid_execution_topology() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.execution_topology.v3",
  "execution_dependencies": [
    {
      "id": "backend_unit",
      "target": "backend-unit",
      "category": "backend",
      "order": 0,
      "service_backed": false
    }
  ],
  "task_surface": {
    "targets": [
      {
        "name": "backend-unit"
      }
    ]
  }
}
JSON
}

write_valid_check_schedule() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.check_schedule.v12",
  "schedules": [
    {
      "target": "check",
      "scheduler_kind": "check",
      "capacity_profile": "check_default",
      "resource_limits": {
        "host_cpu": 1
      },
      "summary_groups": [],
      "work_units": [
        {
          "target": "json-shape-check",
          "weight_ms": 1,
          "needs": [],
          "resource_claims": {
            "host_cpu": 1
          },
          "make_jobs": "host_cpu"
        }
      ]
    }
  ]
}
JSON
}

write_valid_service_backed_schedule() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.service_backed_schedule.v11",
  "generated": {
    "generator": "synthetic",
    "topology": "tools/execution_topology_manifest.json",
    "browser_batch_manifest": "tools/browser_e2e_batch_manifest.json",
    "make_target_duration_baseline": "tools/service_backed_make_target_duration_baselines.json"
  },
  "schedules": [
    {
      "target": "test",
      "scheduler_kind": "service_backed",
      "capacity_profile": "service_backed_default",
      "resource_limits": {
        "go_cpu": 1
      },
      "work_unit_sources": [
        {
          "type": "make_target",
          "class": "backend",
          "target": "backend-store",
          "needs": [],
          "weight_ms": 1,
          "resource_claims": {}
        }
      ]
    }
  ]
}
JSON
}

write_valid_browser_batch() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.browser_e2e_batch_manifest.v5",
  "stages": [
    {
      "name": "functional",
      "target": "browser-e2e-functional",
      "schedule_tags": ["browser"],
      "scheduler_needs": [],
      "summary_children": [],
      "groups": [
        {
          "name": "functional-core",
          "kind": "functional",
          "target": "browser-e2e-functional"
        }
      ]
    }
  ]
}
JSON
}

write_valid_scheduler_resource_registry() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.scheduler_resource_registry.v3",
  "resources": [
    {
      "name": "host_cpu",
      "display_name": "Host CPU",
      "schedulers": ["check"],
      "display_order": 10,
      "capacity": {
        "default_limit": 1,
        "override_env": "CHECK_HOST_CPU_JOBS"
      }
    }
  ],
  "templates": [
    {
      "name": "browser_stage",
      "prefix": "browser_stage_",
      "display_name": "Browser stage",
      "schedulers": ["service_backed"],
      "display_order": 100
    }
  ],
  "capacity_profiles": [
    {
      "name": "check_default",
      "scheduler": "check",
      "resources": ["host_cpu"]
    }
  ],
  "forwarding_profiles": []
}
JSON
}

write_valid_bootstrap_admin() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "bootstrap_schema_id": "cartulary.bootstrap_admin.v1",
  "bootstrap_artifact_id": "synthetic-bootstrap-admin",
  "email": "admin@example.test",
  "display_name": "Synthetic Admin",
  "initial_password": "replace-me"
}
JSON
}

write_valid_tool_run_summary() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.tool_run_summary.v2",
  "target": "json-shape-check",
  "command": {
    "cwd": "/repo",
    "argv": ["make", "json-shape-check"],
    "make_target": "json-shape-check",
    "env": {}
  },
  "status": "pass",
  "exit_code": 0,
  "started_at": "2026-01-01T00:00:00Z",
  "completed_at": "2026-01-01T00:00:01Z",
  "duration_ms": 1000,
  "output_mode": "summary",
  "result_root": ".cartulary/test-results",
  "run_id": "run",
  "run_root": ".cartulary/test-results/run",
  "summary_artifacts": [
    {
      "role": "tool_run_summary",
      "kind": "json",
      "path": ".cartulary/test-results/run/json-shape-check/tool-run-summary.json"
    }
  ],
  "log_artifacts": [],
  "work_units": [],
  "evidence_targets": [],
  "helper_units": [],
  "counts": {
    "phases": 0,
    "tests": 0,
    "failed": 0,
    "non_test": 0,
    "non_test_failed": 0,
    "packages": 0
  },
  "phase_accounting": {
    "authoritative": 0,
    "support": 0,
    "raw": 0,
    "tooling_support": 0,
    "unowned_regression": 0,
    "unmapped": 0,
    "authoritative_failed": 0,
    "support_failed": 0,
    "raw_failed": 0,
    "tooling_support_failed": 0,
    "unowned_regression_failed": 0,
    "unmapped_failed": 0,
    "missing": 0
  },
  "failure_class": null,
  "failure_reason": null,
  "failures": [],
  "slowest": [],
  "warnings": [],
  "rerun_commands": ["make json-shape-check"],
  "extensions": {}
}
JSON
}

mutate_json_fixture() {
  local mutation="$1"
  local file="$2"

  "$NODE_BIN" - "$mutation" "$file" <<'JS'
const fs = require("node:fs");

const mutation = process.argv[2];
const file = process.argv[3];
const value = JSON.parse(fs.readFileSync(file, "utf8"));
const mutations = {
  "bootstrap-admin-bad-email": (fixture) => {
    fixture.email = "not-an-email";
  },
  "tool-run-summary-empty-started-at": (fixture) => {
    fixture.started_at = "";
  },
  "tool-run-summary-null-completed-at": (fixture) => {
    fixture.completed_at = null;
  },
  "tool-run-summary-invalid-started-at": (fixture) => {
    fixture.started_at = "not-a-timestamp";
  },
  "tool-run-summary-unsorted-artifacts": (fixture) => {
    fixture.summary_artifacts = [
      {
        role: "z_artifact",
        kind: "json",
        path: ".cartulary/test-results/run/json-shape-check/z.json",
      },
      {
        role: "a_artifact",
        kind: "json",
        path: ".cartulary/test-results/run/json-shape-check/a.json",
      },
    ];
  },
  "check-schedule-schema-v6": (fixture) => {
    fixture.schema_id = "cartulary.check_schedule.v6";
  },
  "check-schedule-unknown-work-unit-key": (fixture) => {
    fixture.schedules[0].work_units[0].legacy_key = true;
  },
  "check-schedule-empty-schedules": (fixture) => {
    fixture.schedules = [];
  },
  "check-schedule-invalid-env-name": (fixture) => {
    fixture.schedules[0].work_units[0].env = {
      "not-safe": "1",
    };
  },
  "service-backed-unknown-source-key": (fixture) => {
    fixture.schedules[0].work_unit_sources[0].legacy_key = true;
  },
  "service-backed-empty-sources": (fixture) => {
    fixture.schedules[0].work_unit_sources = [];
  },
  "service-backed-missing-generated": (fixture) => {
    delete fixture.generated.topology;
  },
  "browser-batch-unknown-stage-key": (fixture) => {
    fixture.stages[0].legacy_key = true;
  },
  "browser-batch-unknown-group-key": (fixture) => {
    fixture.stages[0].groups[0].legacy_key = true;
  },
  "browser-batch-empty-groups": (fixture) => {
    fixture.stages[0].groups = [];
  },
  "browser-batch-obsolete-scheduler-policy": (fixture) => {
    delete fixture.stages[0].scheduler_needs;
    fixture.stages[0].scheduler_dependency_policy = "parallel";
  },
  "scheduler-registry-bad-capacity-one-of": (fixture) => {
    fixture.resources[0].capacity.auto_policy = "host_cpu_auto";
  },
  "scheduler-registry-unknown-key": (fixture) => {
    fixture.legacy_key = true;
  },
  "execution-topology-duplicate-target": (fixture) => {
    fixture.task_surface.targets.push({ name: "backend-unit" });
  },
  "phase-map-missing-unit": (fixture) => {
    delete fixture.unit;
  },
  "phase-map-unknown-legacy-key": (fixture) => {
    fixture.legacy_manifest_key = true;
  },
  "phase-registry-schema-v0": (fixture) => {
    fixture.schema_id = "cartulary.phase_registry.v0";
  },
};

const applyMutation = mutations[mutation];
if (!applyMutation) {
  throw new Error(`unknown JSON fixture mutation ${mutation}`);
}

applyMutation(value);
fs.writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
JS
}

mkdir -p "$ROOT_DIR/tmp"
tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/json-shapes.XXXXXX")"
cleanup_paths+=("$tmp_dir")

phase_registry="$tmp_dir/phase_registry.json"
write_valid_phase_registry "$phase_registry"
assert_contains "$(assert_passes "valid phase registry" run_shape_check phase-registry "$phase_registry")" \
  "json shape check passed" \
  "valid phase registry"

bad_schema_registry="$tmp_dir/phase_registry_bad_schema.json"
write_valid_phase_registry "$bad_schema_registry"
mutate_json_fixture phase-registry-schema-v0 "$bad_schema_registry"
bad_schema_output="$(assert_fails "malformed registry schema_id" run_shape_check phase-registry "$bad_schema_registry")"
assert_contains "$bad_schema_output" "must declare schema_id cartulary.phase_registry.v1" "malformed schema_id"

duplicate_phase_registry="$tmp_dir/phase_registry_duplicate_phase.json"
cat >"$duplicate_phase_registry" <<'JSON'
{
  "schema_id": "cartulary.phase_registry.v1",
  "phases": [
    {
      "phase": "phase9",
      "order": 9,
      "status": "planned",
      "label": "Phase 9",
      "manifest_path": "tools/phase9_test_map.json",
      "ledger_path": "docs/testing/phase9_coverage_ledger.md",
      "scope": "Synthetic shape fixture.",
      "normative_owners": "docs/spec/00_document_set_status_and_precedence.md"
    },
    {
      "phase": "phase9",
      "order": 10,
      "status": "planned",
      "label": "Phase 9 Duplicate",
      "manifest_path": "tools/phase9_test_map.json",
      "ledger_path": "docs/testing/phase9_coverage_ledger.md",
      "scope": "Synthetic shape fixture.",
      "normative_owners": "docs/spec/00_document_set_status_and_precedence.md"
    }
  ]
}
JSON
duplicate_phase_output="$(assert_fails "duplicate phase identifiers" run_shape_check phase-registry "$duplicate_phase_registry")"
assert_contains "$duplicate_phase_output" "phases.phase contains duplicate phase9" "duplicate phase identifiers"

phase_map="$tmp_dir/phase9_test_map.json"
write_valid_phase_map "$phase_map"
assert_contains "$(assert_passes "valid phase map" run_shape_check phase-map "$phase_map")" \
  "json shape check passed" \
  "valid phase map"

unknown_phase_key="$tmp_dir/phase9_test_map_unknown_key.json"
write_valid_phase_map "$unknown_phase_key"
mutate_json_fixture phase-map-unknown-legacy-key "$unknown_phase_key"
unknown_key_output="$(assert_fails "unknown phase manifest key" run_shape_check phase-map "$unknown_phase_key")"
assert_contains "$unknown_key_output" "unknown key legacy_manifest_key" "unknown manifest key"

missing_section="$tmp_dir/phase9_test_map_missing_unit.json"
write_valid_phase_map "$missing_section"
mutate_json_fixture phase-map-missing-unit "$missing_section"
missing_section_output="$(assert_fails "missing phase manifest section" run_shape_check phase-map "$missing_section")"
assert_contains "$missing_section_output" ".unit is required" "missing manifest section"

duplicate_target_topology="$tmp_dir/execution_topology_duplicate_target.json"
write_valid_execution_topology "$duplicate_target_topology"
mutate_json_fixture execution-topology-duplicate-target "$duplicate_target_topology"
duplicate_target_output="$(assert_fails "duplicate target identifiers" run_shape_check execution-topology "$duplicate_target_topology")"
assert_contains "$duplicate_target_output" "task_surface.targets.name contains duplicate backend-unit" "duplicate target identifiers"

stale_schedule="$tmp_dir/check_schedule_stale.json"
write_valid_check_schedule "$stale_schedule"
mutate_json_fixture check-schedule-schema-v6 "$stale_schedule"
stale_schedule_output="$(assert_fails "stale generated schedule shape" run_shape_check check-schedule "$stale_schedule")"
assert_contains "$stale_schedule_output" "must declare schema_id cartulary.check_schedule.v12" "stale generated schedule shape"

unknown_work_unit_key="$tmp_dir/check_schedule_unknown_work_unit_key.json"
write_valid_check_schedule "$unknown_work_unit_key"
mutate_json_fixture check-schedule-unknown-work-unit-key "$unknown_work_unit_key"
unknown_work_unit_output="$(assert_fails "unknown check schedule work unit key" run_shape_check check-schedule "$unknown_work_unit_key")"
assert_contains "$unknown_work_unit_output" "unknown key legacy_key" "unknown check schedule work unit key"

empty_check_schedules="$tmp_dir/check_schedule_empty_schedules.json"
write_valid_check_schedule "$empty_check_schedules"
mutate_json_fixture check-schedule-empty-schedules "$empty_check_schedules"
empty_check_schedules_output="$(assert_fails "empty check schedule schedules" run_shape_check check-schedule "$empty_check_schedules")"
assert_contains "$empty_check_schedules_output" "schedules must be a non-empty array" "empty check schedule schedules"

invalid_check_env="$tmp_dir/check_schedule_invalid_env.json"
write_valid_check_schedule "$invalid_check_env"
mutate_json_fixture check-schedule-invalid-env-name "$invalid_check_env"
invalid_check_env_output="$(assert_fails "invalid check schedule env name" run_shape_check check-schedule "$invalid_check_env")"
assert_contains "$invalid_check_env_output" "env key has invalid value" "invalid check schedule env name"

service_backed_schedule="$tmp_dir/service_backed_schedule.json"
write_valid_service_backed_schedule "$service_backed_schedule"
assert_contains "$(assert_passes "valid service-backed schedule" run_shape_check service-backed-schedule "$service_backed_schedule")" \
  "json shape check passed" \
  "valid service-backed schedule"

unknown_service_source_key="$tmp_dir/service_backed_schedule_unknown_source_key.json"
write_valid_service_backed_schedule "$unknown_service_source_key"
mutate_json_fixture service-backed-unknown-source-key "$unknown_service_source_key"
unknown_service_source_output="$(assert_fails "unknown service-backed source key" run_shape_check service-backed-schedule "$unknown_service_source_key")"
assert_contains "$unknown_service_source_output" "unknown key legacy_key" "unknown service-backed source key"

empty_service_sources="$tmp_dir/service_backed_schedule_empty_sources.json"
write_valid_service_backed_schedule "$empty_service_sources"
mutate_json_fixture service-backed-empty-sources "$empty_service_sources"
empty_service_sources_output="$(assert_fails "empty service-backed sources" run_shape_check service-backed-schedule "$empty_service_sources")"
assert_contains "$empty_service_sources_output" "work_unit_sources must be a non-empty array" "empty service-backed sources"

missing_service_generated="$tmp_dir/service_backed_schedule_missing_generated.json"
write_valid_service_backed_schedule "$missing_service_generated"
mutate_json_fixture service-backed-missing-generated "$missing_service_generated"
missing_service_generated_output="$(assert_fails "missing service-backed generated metadata" run_shape_check service-backed-schedule "$missing_service_generated")"
assert_contains "$missing_service_generated_output" "generated.topology must be a non-empty string" "missing service-backed generated metadata"

browser_batch="$tmp_dir/browser_batch.json"
write_valid_browser_batch "$browser_batch"
assert_contains "$(assert_passes "valid browser batch" run_shape_check browser-batch "$browser_batch")" \
  "json shape check passed" \
  "valid browser batch"

unknown_browser_stage_key="$tmp_dir/browser_batch_unknown_stage_key.json"
write_valid_browser_batch "$unknown_browser_stage_key"
mutate_json_fixture browser-batch-unknown-stage-key "$unknown_browser_stage_key"
unknown_browser_stage_output="$(assert_fails "unknown browser stage key" run_shape_check browser-batch "$unknown_browser_stage_key")"
assert_contains "$unknown_browser_stage_output" "unknown key legacy_key" "unknown browser stage key"

unknown_browser_group_key="$tmp_dir/browser_batch_unknown_group_key.json"
write_valid_browser_batch "$unknown_browser_group_key"
mutate_json_fixture browser-batch-unknown-group-key "$unknown_browser_group_key"
unknown_browser_group_output="$(assert_fails "unknown browser group key" run_shape_check browser-batch "$unknown_browser_group_key")"
assert_contains "$unknown_browser_group_output" "unknown key legacy_key" "unknown browser group key"

empty_browser_groups="$tmp_dir/browser_batch_empty_groups.json"
write_valid_browser_batch "$empty_browser_groups"
mutate_json_fixture browser-batch-empty-groups "$empty_browser_groups"
empty_browser_groups_output="$(assert_fails "empty browser groups" run_shape_check browser-batch "$empty_browser_groups")"
assert_contains "$empty_browser_groups_output" "groups must be a non-empty array" "empty browser groups"

obsolete_browser_policy="$tmp_dir/browser_batch_obsolete_scheduler_policy.json"
write_valid_browser_batch "$obsolete_browser_policy"
mutate_json_fixture browser-batch-obsolete-scheduler-policy "$obsolete_browser_policy"
obsolete_browser_policy_output="$(assert_fails "obsolete browser scheduler policy" run_shape_check browser-batch "$obsolete_browser_policy")"
assert_contains "$obsolete_browser_policy_output" "scheduler_dependency_policy is obsolete; use scheduler_needs[]" "obsolete browser scheduler policy"

scheduler_registry="$tmp_dir/scheduler_resource_registry.json"
write_valid_scheduler_resource_registry "$scheduler_registry"
assert_contains "$(assert_passes "valid scheduler resource registry" run_shape_check scheduler-resource-registry "$scheduler_registry")" \
  "json shape check passed" \
  "valid scheduler resource registry"

bad_scheduler_capacity="$tmp_dir/scheduler_resource_registry_bad_capacity.json"
write_valid_scheduler_resource_registry "$bad_scheduler_capacity"
mutate_json_fixture scheduler-registry-bad-capacity-one-of "$bad_scheduler_capacity"
bad_scheduler_capacity_output="$(assert_fails "invalid scheduler capacity one-of" run_shape_check scheduler-resource-registry "$bad_scheduler_capacity")"
assert_contains "$bad_scheduler_capacity_output" "must declare exactly one of default_limit or auto_policy" "invalid scheduler capacity one-of"

unknown_scheduler_key="$tmp_dir/scheduler_resource_registry_unknown_key.json"
write_valid_scheduler_resource_registry "$unknown_scheduler_key"
mutate_json_fixture scheduler-registry-unknown-key "$unknown_scheduler_key"
unknown_scheduler_key_output="$(assert_fails "unknown scheduler registry key" run_shape_check scheduler-resource-registry "$unknown_scheduler_key")"
assert_contains "$unknown_scheduler_key_output" "unknown key legacy_key" "unknown scheduler registry key"

bootstrap_admin="$tmp_dir/bootstrap-admin.json"
write_valid_bootstrap_admin "$bootstrap_admin"
assert_contains "$(assert_passes "valid bootstrap admin" run_shape_check bootstrap-admin "$bootstrap_admin")" \
  "json shape check passed" \
  "valid bootstrap admin"

bad_bootstrap_admin="$tmp_dir/bootstrap-admin-bad-email.json"
write_valid_bootstrap_admin "$bad_bootstrap_admin"
mutate_json_fixture bootstrap-admin-bad-email "$bad_bootstrap_admin"
bad_bootstrap_output="$(assert_fails "invalid bootstrap admin JSON" run_shape_check bootstrap-admin "$bad_bootstrap_admin")"
assert_contains "$bad_bootstrap_output" "email has invalid value" "invalid bootstrap-admin JSON"

tool_run_summary="$tmp_dir/tool-run-summary.json"
write_valid_tool_run_summary "$tool_run_summary"
assert_contains "$(assert_passes "valid tool run summary" run_shape_check tool-run-summary "$tool_run_summary")" \
  "json shape check passed" \
  "valid tool run summary"

empty_started_at="$tmp_dir/tool-run-summary-empty-started-at.json"
write_valid_tool_run_summary "$empty_started_at"
mutate_json_fixture tool-run-summary-empty-started-at "$empty_started_at"
empty_started_output="$(assert_fails "empty tool run started_at" run_shape_check tool-run-summary "$empty_started_at")"
assert_contains "$empty_started_output" "started_at must be a non-empty string" "empty tool run started_at"

null_completed_at="$tmp_dir/tool-run-summary-null-completed-at.json"
write_valid_tool_run_summary "$null_completed_at"
mutate_json_fixture tool-run-summary-null-completed-at "$null_completed_at"
null_completed_output="$(assert_fails "null tool run completed_at" run_shape_check tool-run-summary "$null_completed_at")"
assert_contains "$null_completed_output" "completed_at must be a non-empty string" "null tool run completed_at"

invalid_started_at="$tmp_dir/tool-run-summary-invalid-started-at.json"
write_valid_tool_run_summary "$invalid_started_at"
mutate_json_fixture tool-run-summary-invalid-started-at "$invalid_started_at"
invalid_started_output="$(assert_fails "invalid tool run started_at" run_shape_check tool-run-summary "$invalid_started_at")"
assert_contains "$invalid_started_output" "started_at must be an RFC3339 timestamp" "invalid tool run started_at"

unsorted_artifacts="$tmp_dir/tool-run-summary-unsorted-artifacts.json"
write_valid_tool_run_summary "$unsorted_artifacts"
mutate_json_fixture tool-run-summary-unsorted-artifacts "$unsorted_artifacts"
unsorted_artifacts_output="$(assert_fails "unsorted tool run artifacts" run_shape_check tool-run-summary "$unsorted_artifacts")"
assert_contains "$unsorted_artifacts_output" "summary_artifacts must be sorted" "unsorted tool run artifacts"

echo "json shape harness tests passed"
