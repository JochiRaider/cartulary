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

run_schema_validation() {
  local schema_id="$1"
  local file="$2"

  "$ROOT_DIR/scripts/harness-contract.sh" validate-schema "$schema_id" "$file"
}

run_accessibility_summary_writer() {
  (
    cd "$ROOT_DIR"
    "$NODE_BIN" scripts/write-frontend-accessibility-summary.mjs "$@"
  )
}

run_frontend_phase_map_validation() {
  local file="$1"
  local phase="$2"

  (
    cd "$ROOT_DIR"
    "$NODE_BIN" --input-type=module - "$file" "$phase" <<'JS'
import { readFileSync } from "node:fs";
import { validateFrontendPhaseMap } from "./scripts/lib/frontend-phase-manifest.mjs";

const [file, phase] = process.argv.slice(2);
const manifest = JSON.parse(readFileSync(file, "utf8"));
validateFrontendPhaseMap(manifest, file, phase);
console.log("frontend phase map validated");
JS
  )
}

write_valid_phase_map() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
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
  "schema_id": "cartulary.scheduler_manifest.v1",
  "generated": {
    "generator": "synthetic"
  },
  "schedules": [
    {
      "target": "check",
      "scheduler_kind": "check",
      "capacity_profile": "check_default",
      "resource_limits": {
        "host_cpu": 1
      },
      "stop_on_first_failure": true,
      "progress_tick_seconds": 30,
      "validate_timing": true,
      "summary_groups": [],
      "work_units": [
        {
          "target": "json-shape-check",
          "weight_ms": 1,
          "needs": [],
          "resource_claims": {
            "host_cpu": 1
          },
          "make_jobs": "host_cpu",
          "command": {
            "type": "make_target",
            "target": "json-shape-check"
          }
        }
      ],
      "finalizers": []
    }
  ]
}
JSON
}

write_valid_service_backed_schedule() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "generated": {
    "generator": "synthetic"
  },
  "schedules": [
    {
      "target": "test",
      "scheduler_kind": "service_backed",
      "capacity_profile": "service_backed_default",
      "resource_limits": {
        "go_cpu": 1
      },
      "stop_on_first_failure": false,
      "progress_tick_seconds": 30,
      "validate_timing": true,
      "summary_groups": [],
      "work_units": [
        {
          "kind": "make_target",
          "class": "backend",
          "target": "backend-store",
          "needs": [],
          "weight_ms": 1,
          "resource_claims": {},
          "command": {
            "type": "make_target",
            "target": "backend-store"
          }
        }
      ],
      "finalizers": []
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
  "schema_id": "cartulary.scheduler_resource_registry.v4",
  "resources": [
    {
      "name": "host_cpu",
      "display_name": "Host CPU",
      "schedulers": ["check"],
      "display_order": 10,
      "capacity": {
        "default_limit": 1,
        "override_env": "CHECK_HOST_CPU_JOBS",
        "max_limit": 256
      }
    }
  ],
  "templates": [
    {
      "name": "browser_stage",
      "prefix": "browser_stage_",
      "display_name": "Browser stage",
      "schedulers": ["service_backed"],
      "display_order": 100,
      "max_limit": 8
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
  "schema_id": "cartulary.tool_run_summary.v3",
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
  "scheduler_timing": null,
  "extensions": {}
}
JSON
}

write_valid_fallow_static_summary() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.fallow_static_summary.v1",
  "target": "frontend-fallow-static",
  "generated_at": "2026-01-01T00:00:00Z",
  "mode": "phase_a_report",
  "config": {
    "path": ".fallowrc.json",
    "static_layer": "open",
    "runtime_enabled": false
  },
  "reports": [
    {
      "name": "dead-code",
      "command": ["fallow", "dead-code", "--format", "json"],
      "status": "pass",
      "exit_code": 0,
      "artifact": {
        "role": "dead_code_json",
        "kind": "json",
        "path": ".cartulary/test-results/run/frontend-fallow-static/fallow/dead-code.json"
      },
      "issue_count": 1,
      "by_rule": {
        "apps-web-no-react-data-grid": 1
      },
      "by_severity": {
        "warn": 1
      },
      "raw_count_fields": {
        "policy_violations": 1
      }
    }
  ],
  "totals": {
    "reports": 1,
    "issue_count": 1,
    "by_rule": {
      "apps-web-no-react-data-grid": 1
    },
    "by_severity": {
      "warn": 1
    }
  },
  "baseline": {
    "mode": "not_configured",
    "artifacts": []
  },
  "enforcement": {
    "blocking": false,
    "failure_on_issues": false
  },
  "artifacts": [
    {
      "role": "dead_code_json",
      "kind": "json",
      "path": ".cartulary/test-results/run/frontend-fallow-static/fallow/dead-code.json"
    }
  ],
  "warnings": [
    {
      "kind": "fallow_findings",
      "issue_count": 1,
      "message": "Fallow Phase A findings are retained as non-blocking static-analysis evidence."
    }
  ],
  "extensions": {}
}
JSON
}

write_minimal_scheduler_summary() {
  local file="$1"
  local schema_id="$2"

  cat >"$file" <<JSON
{
  "schema_id": "$schema_id",
  "target": "check",
  "status": "pass",
  "failure_class": null,
  "failure_reason": null,
  "failure_classes": {},
  "failure_reasons": {},
  "failures": [],
  "failure_headline": "",
  "scheduler_kind": "check",
  "total_work_units": 0,
  "completed_work_units": 0,
  "scheduler_started_monotonic_ms": 0,
  "scheduler_completed_monotonic_ms": 0,
  "scheduler_total_duration_ms": 0,
  "scheduler_started_at": "2026-01-01T00:00:00Z",
  "scheduler_completed_at": "2026-01-01T00:00:00Z",
  "critical_path_wall_duration_ms": 0,
  "critical_path_units": [],
  "critical_path_blockers": [],
  "critical_path_terminal_unit": null,
  "skipped_work_units": [],
  "failed_work_unit": null,
  "failed_work_unit_detail": null,
  "resource_limits": null,
  "resource_limit_sources": null,
  "max_running_work_units": 0,
  "max_running_groups": 0,
  "max_active_resource_claims": null,
  "blocked_reasons_seen": [],
  "blocked_resources_seen": [],
  "blocked_explanations_seen": [],
  "waiting_on_seen": [],
  "top_blockers": [],
  "slowest_work_units": [],
  "nested_scheduler_limits": [],
  "nested_scheduler_observations": [],
  "finalizer_count": 0,
  "finalizer_failures": 0,
  "finalizer_timings": [],
  "artifacts": null,
  "extensions": {}
}
JSON
}

write_valid_frontend_accessibility_summary_v2() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.frontend_accessibility_summary.v2",
  "status": "pass",
  "phase_rows": [
    {
      "row_id": "FE-A11Y-P1-01",
      "phase_id": "FE-P1",
      "evidence_class": "design_direction",
      "claim_status": "implemented",
      "targets": [
        {
          "target_name": "browser-e2e-a11y",
          "command_id": "cartulary.harness.command.browser_e2e_a11y.v1",
          "evidence_role": "primary",
          "required_for_closure": true,
          "frontend_row_accounting_required": true,
          "scenario_title_required": true
        }
      ]
    }
  ],
  "scenarios": [
    {
      "row_id": "FE-A11Y-P1-01",
      "title": "implemented scenario passes",
      "status": "pass"
    },
    {
      "row_id": "FE-A11Y-P1-01",
      "title": "implemented scenario can fail",
      "status": "fail"
    },
    {
      "row_id": "FE-A11Y-P1-01",
      "title": "implemented scenario can be missing",
      "status": "missing"
    },
    {
      "row_id": "FE-A11Y-P1-01",
      "title": "implemented scenario can be skipped",
      "status": "skipped"
    }
  ],
  "keyboard_matrix": [
    {
      "row_id": "FE-A11Y-P1-01",
      "title": "implemented scenario passes",
      "result": "pass",
      "coverage": "keyboard"
    }
  ],
  "state_communication_checks": [
    {
      "row_id": "FE-A11Y-P1-01",
      "title": "implemented scenario passes",
      "result": "pass",
      "coverage": "state"
    }
  ],
  "contrast_checks": [
    {
      "row_id": "FE-A11Y-P1-01",
      "title": "implemented scenario passes",
      "result": "pass",
      "coverage": "contrast",
      "target": "auth-login-submit",
      "ratio": 7.2,
      "threshold": 4.5,
      "foreground": "rgb(248, 250, 252)",
      "background": "rgb(15, 23, 42)"
    }
  ],
  "violations": [],
  "artifact_refs": [
    {
      "kind": "playwright_phase",
      "path": ".cartulary/test-results/run/browser-e2e-a11y/browser-e2e-a11y-accessibility"
    }
  ]
}
JSON
}

write_valid_frontend_accessibility_preflight_summary() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.frontend_accessibility_preflight_summary.v1",
  "status": "pass",
  "phase_rows": [
    {
      "row_id": "FE-A11Y-P2-01",
      "phase_id": "FE-P2",
      "evidence_class": "design_direction",
      "claim_status": "blocked",
      "targets": [
        {
          "target_name": "browser-e2e-a11y",
          "command_id": "cartulary.harness.command.browser_e2e_a11y.v1",
          "evidence_role": "diagnostic_only",
          "required_for_closure": false,
          "frontend_row_accounting_required": false,
          "scenario_title_required": true
        }
      ]
    }
  ],
  "scenarios": [
    {
      "row_id": "FE-A11Y-P2-01",
      "title": "blocked scenario preflight",
      "status": "pass"
    }
  ],
  "violations": [],
  "artifact_refs": [
    {
      "kind": "playwright_phase",
      "path": ".cartulary/test-results/run/browser-e2e-a11y-preflight/browser-e2e-a11y-preflight-accessibility-preflight"
    }
  ]
}
JSON
}

write_valid_frontend_row_accounting() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.frontend_row_accounting.v3",
  "target_name": "browser-e2e-webserver-backed",
  "command_id": "cartulary.harness.command.browser_e2e_webserver_backed.v1",
  "phase_namespace": "frontend",
  "accounting_scope": {
    "mode": "selected_rows",
    "invocation_kind": "frontend_phase_slice",
    "phase_namespace": "frontend",
    "phase": "FE-P2",
    "selection_policy": "frontend_rows_through_selected_phase",
    "selected_row_ids": [
      "FE-B-P2-02"
    ]
  },
  "registry_ref": "tools/frontend_phase_registry.json",
  "registry_digest": "0000000000000000000000000000000000000000000000000000000000000000",
  "guide_ref": "docs/guides/cartulary_frontend_implementation_testing_guide.md",
  "guide_digest": "1111111111111111111111111111111111111111111111111111111111111111",
  "phase_map_refs": [
    "tools/frontend_phase_maps/fe_p2_test_map.json"
  ],
  "phase_map_digests": [
    "2222222222222222222222222222222222222222222222222222222222222222"
  ],
  "run_root": ".cartulary/test-results/run",
  "target_status": "pass",
  "scenario_results": [
    {
      "scenario_title": "FE-B-P2-02 Verify System views switcher keyboard entry, roving focus, selection, dismissal, and focus restoration.",
      "status": "passed",
      "row_ids": [
        "FE-B-P2-02"
      ],
      "artifact_refs": [
        "apps/web/e2e/phase2.spec.ts"
      ]
    }
  ],
  "row_results": [
    {
      "row_id": "FE-B-P2-02",
      "phase_id": "FE-P2",
      "evidence_class": "product_conformance",
      "claim_status_at_run": "implemented",
      "target_mapping_status": "mapped",
      "closure_status": "closed",
      "closing_scenario_titles": [
        "FE-B-P2-02 Verify System views switcher keyboard entry, roving focus, selection, dismissal, and focus restoration."
      ],
      "failure_reason": ""
    }
  ],
  "rollup": {
    "implemented": 1,
    "blocked": 0,
    "missing": 0,
    "stale": 0,
    "not_applicable": 0,
    "closed": 1,
    "failed": 0
  },
  "target": "browser-e2e-webserver-backed",
  "rows": [
    {
      "phase_id": "FE-P2",
      "phase_status": "active",
      "row_rollup_state": "active_green",
      "row_id": "FE-B-P2-02",
      "layer": "browser_integration",
      "evidence_class": "product_conformance",
      "claim_status": "implemented",
      "claim": {
        "statement": "fixture claim",
        "claim_publication_intent": "none",
        "closure_scope": "scenario"
      },
      "blockers": [],
      "required_for_closure": true,
      "scenario_titles": [
        "FE-B-P2-02 Verify System views switcher keyboard entry, roving focus, selection, dismissal, and focus restoration."
      ],
      "target": "browser-e2e-webserver-backed",
      "target_status": "pass",
      "scenarios": [
        {
          "title": "FE-B-P2-02 Verify System views switcher keyboard entry, roving focus, selection, dismissal, and focus restoration.",
          "status": "passed",
          "files": ["apps/web/e2e/phase2.spec.ts"]
        }
      ],
      "closure_status": "closed"
    }
  ],
  "counts": {
    "rows": 1,
    "scenarios": 1,
    "closed_rows": 1,
    "blocked_by_target_rows": 0,
    "failed_rows": 0,
    "missing_rows": 0,
    "not_evaluable_rows": 0,
    "passed_scenarios": 1,
    "failed_scenarios": 0,
    "missing_scenarios": 0,
    "skipped_scenarios": 0,
    "unknown_scenarios": 0
  }
}
JSON
}

write_valid_agent_finalize_summary() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.agent_finalize_summary.v3",
  "target": "agent-finalize",
  "status": "pass",
  "result_root": ".cartulary/test-results",
  "run_id": "run",
  "run_root": ".cartulary/test-results/run",
  "output_mode": "summary",
  "results_dir": null,
  "results_dir_status": "skipped",
  "retained_run_selection": {
    "status": "skipped",
    "supplied_results_dir": null,
    "latest_results_dir": null,
    "supplied_is_latest": null,
    "allow_older_results_dir": false
  },
  "started_at": "2026-01-01T00:00:00Z",
  "completed_at": "2026-01-01T00:00:01Z",
  "duration_ms": 1000,
  "generated": {
    "status": "unchanged",
    "updated_file_count": 0
  },
  "mutation_rollback": {
    "status": "not_needed",
    "restored_file_count": 0,
    "restored_files": [],
    "error": null
  },
  "duration": {
    "status": "skipped",
    "refreshed": false,
    "checked": false
  },
  "run_checks": {
    "status": "skipped",
    "checked": false
  },
  "updated_files": [],
  "actions": [
    {
      "action_id": "structure_ledger_refresh",
      "description": "Refresh phase-ledger and phase-schedule generated artifacts, then verify no unsupported drift remains.",
      "requires_results_dir": false,
      "mutating": true,
      "status": "pass",
      "execution_state": "executed",
      "started_at": "2026-01-01T00:00:00Z",
      "completed_at": "2026-01-01T00:00:01Z",
      "duration_ms": 1000,
      "skipped_reason": null,
      "cache": {
        "enabled": true,
        "state": "miss",
        "cache_schema_id": "cartulary.agent_finalize_action_cache_record.v1",
        "action_contract_version": "v1",
        "key_sha256": "sha256:0000000000000000000000000000000000000000000000000000000000000000",
        "input_profile_id": "agent_finalize.structure_ledger_refresh.v1",
        "input_digest_sha256": "sha256:1111111111111111111111111111111111111111111111111111111111111111",
        "output_digest_sha256": "sha256:2222222222222222222222222222222222222222222222222222222222222222",
        "record_path": ".cache/cartulary/agent-finalize-action-cache/structure_ledger_refresh/record.json",
        "reason_code": "cache_record_missing"
      },
      "substeps": [
        {
          "id": "phase-ledgers",
          "target": "phase-ledgers",
          "command_kind": "make_target",
          "requires_results_dir": false,
          "mutates_repo": true,
          "status": "pass",
          "started_at": "2026-01-01T00:00:00Z",
          "completed_at": "2026-01-01T00:00:01Z",
          "duration_ms": 1000,
          "exit_code": 0,
          "summary_json": ".cartulary/test-results/run/phase-ledgers/tool-run-summary.json",
          "stdout_log": ".cartulary/test-results/run/phase-ledgers/phase-ledgers/stdout.log",
          "stderr_log": ".cartulary/test-results/run/phase-ledgers/phase-ledgers/stderr.log",
          "skipped_reason": null
        }
      ]
    }
  ],
  "failures": [],
  "child_artifacts": [
    {
      "role": "structure_ledger_refresh_phase-ledgers_summary",
      "kind": "json",
      "path": ".cartulary/test-results/run/phase-ledgers/tool-run-summary.json"
    }
  ]
}
JSON
}

write_valid_cache_record() {
  local schema_id="$1"
  local scope="$2"
  local file="$3"

  cat >"$file" <<JSON
{
  "schema_id": "$schema_id",
  "cache_scope": "$scope",
  "profile_id": "fixture-profile",
  "state": "stored",
  "reason_code": "cache_record_missing",
  "key_sha256": "sha256:0000000000000000000000000000000000000000000000000000000000000000",
  "input_digest_sha256": "sha256:1111111111111111111111111111111111111111111111111111111111111111",
  "output_digest_sha256": "sha256:2222222222222222222222222222222222222222222222222222222222222222",
  "record_path": ".cache/cartulary/${scope}/fixture-profile/record.json",
  "cacheable_outputs": ["tmp/fixture-output"],
  "non_cacheable_side_effects": ["public_target_summary"],
  "created_at": "2026-01-01T00:00:00Z"
}
JSON
}

write_valid_agent_finalize_action_cache_record() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.agent_finalize_action_cache_record.v1",
  "action_id": "structure_ledger_refresh",
  "command_id": "cartulary.harness.command.agent_finalize.v1",
  "action_contract_version": "v1",
  "input_profile_id": "agent_finalize.structure_ledger_refresh.v1",
  "key_sha256": "sha256:0000000000000000000000000000000000000000000000000000000000000000",
  "cache_schema_id": "cartulary.agent_finalize_action_cache_record.v1",
  "digests": {
    "repo_input_digest": "sha256:1111111111111111111111111111111111111111111111111111111111111111",
    "implementation_digest": "sha256:2222222222222222222222222222222222222222222222222222222222222222",
    "toolchain_digest": "sha256:3333333333333333333333333333333333333333333333333333333333333333",
    "environment_digest": "sha256:4444444444444444444444444444444444444444444444444444444444444444",
    "retained_run_digest": null,
    "substep_digest": "sha256:5555555555555555555555555555555555555555555555555555555555555555",
    "input_digest_sha256": "sha256:6666666666666666666666666666666666666666666666666666666666666666",
    "output_digest_sha256": "sha256:7777777777777777777777777777777777777777777777777777777777777777"
  },
  "output_paths": ["tools/task_surface_manifest.json"],
  "updated_at": "2026-01-01T00:00:00Z"
}
JSON
}

write_valid_execution_topology_render_cache_record() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.execution_topology_render_cache.v1",
  "generator": "scripts/render-execution-topology-artifacts.mjs",
  "generator_version": 1,
  "node_version": "v24.15.0",
  "input_digest": "sha256:0000000000000000000000000000000000000000000000000000000000000000",
  "artifacts": [
    {
      "file": "tools/task_surface_manifest.json",
      "hash": "sha256:1111111111111111111111111111111111111111111111111111111111111111",
      "content": "{}\n"
    }
  ],
  "written_at": "2026-01-01T00:00:00Z"
}
JSON
}

write_valid_web_e2e_stack() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.web_e2e_stack.v3",
  "api_origin": "http://127.0.0.1:8080",
  "public_origin": "http://127.0.0.1:4173",
  "backend_port": 8080,
  "frontend_port": 4173,
  "frontend_mode": "preview",
  "frontend_command_kind": "vite-preview",
  "runtime_root": ".cartulary/test-results/run/browser/runtime-root",
  "server_log": ".cartulary/test-results/run/browser/server.log",
  "web_log": ".cartulary/test-results/run/browser/web.log",
  "startup_diagnostics": ".cartulary/test-results/run/browser/startup-diagnostics.json",
  "backend_process_group_id": 101,
  "frontend_process_group_id": 102,
  "backend_ready_at": "2026-01-01T00:00:00Z",
  "frontend_ready_at": "2026-01-01T00:00:01Z",
  "backend_identity": {
    "schema_id": "cartulary.test.runtime_identity.v1",
    "status": "pass",
    "server_pid": 101
  },
  "frontend_ownership": {
    "status": "pass"
  },
  "database": {
    "logical_id": "cartulary_web_e2e_fixture",
    "source": "standalone"
  }
}
JSON
}

write_valid_browser_startup_diagnostics() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.browser_startup_diagnostics.v1",
  "generated_at": "2026-01-01T00:00:00Z",
  "target": "browser-e2e-webserver-backed",
  "status": "fail",
  "startup_phase": "frontend_readiness",
  "frontend_mode": "preview",
  "frontend_command_kind": "vite-preview",
  "api_origin": "http://127.0.0.1:8080",
  "public_origin": "http://127.0.0.1:4173",
  "backend_port": 8080,
  "frontend_port": 4173,
  "backend_process_group_id": 101,
  "frontend_process_group_id": 102,
  "runtime_root": ".cartulary/test-results/run/browser/runtime-root",
  "failure_class": "infra",
  "failure_reason": "service_start_error",
  "message": "frontend exited before readiness",
  "logs": {
    "backend": ".cartulary/test-results/run/browser/server.log",
    "frontend": ".cartulary/test-results/run/browser/web.log"
  },
  "inotify": {
    "platform": "linux",
    "max_user_watches": 8192,
    "max_user_instances": 128,
    "current_watches": 8192,
    "current_instances": 128,
    "remediation": "raise Linux inotify limits"
  }
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
  "tool-run-summary-invalid-extension-key": (fixture) => {
    fixture.extensions.scheduler_timing = {
      scheduler_total_duration_ms: 1000,
    };
  },
  "tool-run-summary-failed-null-reason": (fixture) => {
    fixture.status = "fail";
    fixture.exit_code = 1;
    fixture.failure_class = "product";
    fixture.failure_reason = null;
  },
  "tool-run-summary-missing-scheduler-timing": (fixture) => {
    delete fixture.scheduler_timing;
  },
  "tool-run-summary-missing-schema-id": (fixture) => {
    delete fixture.schema_id;
  },
  "fallow-static-summary-blocking": (fixture) => {
    fixture.enforcement.blocking = true;
  },
  "fallow-static-summary-unknown-key": (fixture) => {
    fixture.legacy_key = true;
  },
  "web-e2e-stack-wrong-frontend-mode": (fixture) => {
    fixture.frontend_mode = "dev";
  },
  "browser-startup-diagnostics-unknown-key": (fixture) => {
    fixture.legacy_key = true;
  },
  "agent-finalize-summary-bad-action-status": (fixture) => {
    fixture.actions[0].status = "done";
  },
  "agent-finalize-summary-unknown-key": (fixture) => {
    fixture.legacy_key = true;
  },
  "frontend-a11y-summary-invalid-status": (fixture) => {
    fixture.scenarios[0].status = "ok";
  },
  "frontend-a11y-summary-unknown-key": (fixture) => {
    fixture.scenarios[0].legacy_key = true;
  },
  "frontend-a11y-summary-missing-required": (fixture) => {
    delete fixture.contrast_checks[0].ratio;
  },
  "frontend-a11y-summary-blocked-row": (fixture) => {
    fixture.phase_rows[0].claim_status = "blocked";
  },
  "frontend-a11y-preflight-implemented-row": (fixture) => {
    fixture.phase_rows[0].claim_status = "implemented";
  },
  "frontend-row-accounting-unknown-key": (fixture) => {
    fixture.rows[0].legacy_key = true;
  },
  "frontend-row-accounting-invalid-closure": (fixture) => {
    fixture.rows[0].closure_status = "complete";
  },
  "frontend-row-accounting-invalid-scenario-status": (fixture) => {
    fixture.rows[0].scenarios[0].status = "pass";
  },
  "frontend-row-accounting-invalid-scope": (fixture) => {
    fixture.accounting_scope.mode = "phase";
  },
  "scheduler-manifest-stale-schema": (fixture) => {
    fixture.schema_id = "cartulary.check_schedule.v12";
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
  "check-schedule-invalid-make-prerequisite-policy": (fixture) => {
    fixture.schedules[0].work_units[0].make_prerequisite_policy = "maybe";
  },
  "check-schedule-browser-worker-overlap": (fixture) => {
    fixture.schedules[0].work_units = [
      {
        id: "browser-visual",
        kind: "browser_group",
        target: "browser-e2e-visual",
        weight_ms: 1,
        resource_claims: {},
        service_session: { target: "check-service-backed" },
        browser_group: {
          id: "browser-e2e-visual:visual",
          kind: "visual",
          workers: "1",
        },
        env: {
          CARTULARY_PLAYWRIGHT_WORKER_COUNT: "2",
          CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET: "0",
        },
        command: {
          type: "browser_group",
          service_target: "check-service-backed",
          browser_stage: "visual",
          group_id: "browser-e2e-visual:visual",
        },
      },
      {
        id: "browser-a11y",
        kind: "browser_group",
        target: "browser-e2e-a11y",
        weight_ms: 1,
        resource_claims: {},
        service_session: { target: "check-service-backed" },
        browser_group: {
          id: "browser-e2e-a11y:a11y",
          kind: "a11y",
          workers: "1",
        },
        env: {
          CARTULARY_PLAYWRIGHT_WORKER_COUNT: "2",
          CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET: "0",
        },
        command: {
          type: "browser_group",
          service_target: "check-service-backed",
          browser_stage: "a11y",
          group_id: "browser-e2e-a11y:a11y",
        },
      },
    ];
  },
  "check-schedule-command-missing-required": (fixture) => {
    delete fixture.schedules[0].work_units[0].command.target;
  },
  "check-schedule-command-forbidden-field": (fixture) => {
    fixture.schedules[0].work_units[0].command.shard = "backend-store-shard-01";
  },
  "check-schedule-command-unknown-type": (fixture) => {
    fixture.schedules[0].work_units[0].command.type = "legacy_command";
  },
  "check-schedule-command-wrong-field-type": (fixture) => {
    fixture.schedules[0].work_units[0].command.target = 123;
  },
  "check-schedule-finalizer-missing-shard-names": (fixture) => {
    fixture.schedules[0].work_units = [
      {
        id: "backend-store-shard-01",
        kind: "go_shard",
        target: "backend-store",
        shard: "backend-store-shard-01",
        weight_ms: 1,
        needs: [],
        completion_keys: ["go_shard:backend-store-shard-01"],
        resource_claims: {},
        command: {
          type: "go_shard",
          target: "backend-store",
          shard: "backend-store-shard-01",
          service_target: "check-service-backed",
        },
      },
      {
        id: "finalize:backend-store",
        kind: "aggregate_finalize",
        target: "backend-store",
        aggregate_target: "backend-store",
        weight_ms: 1,
        needs: ["go_shard:backend-store-shard-01"],
        resource_claims: {},
        command: {
          type: "go_shard_finalize",
          target: "backend-store",
          service_target: "check-service-backed",
        },
      },
    ];
  },
  "check-schedule-finalizer-shard-name-mismatch": (fixture) => {
    fixture.schedules[0].work_units = [
      {
        id: "backend-store-shard-01",
        kind: "go_shard",
        target: "backend-store",
        shard: "backend-store-shard-01",
        weight_ms: 1,
        needs: [],
        completion_keys: ["go_shard:backend-store-shard-01"],
        resource_claims: {},
        command: {
          type: "go_shard",
          target: "backend-store",
          shard: "backend-store-shard-01",
          service_target: "check-service-backed",
        },
      },
      {
        id: "finalize:backend-store",
        kind: "aggregate_finalize",
        target: "backend-store",
        aggregate_target: "backend-store",
        weight_ms: 1,
        needs: ["go_shard:backend-store-shard-01"],
        shard_names: ["backend-store-shard-02"],
        resource_claims: {},
        command: {
          type: "go_shard_finalize",
          target: "backend-store",
          service_target: "check-service-backed",
        },
      },
    ];
  },
  "service-backed-unknown-source-key": (fixture) => {
    fixture.schedules[0].work_units[0].legacy_key = true;
  },
  "service-backed-empty-sources": (fixture) => {
    fixture.schedules[0].work_units = [];
  },
  "service-backed-missing-generated": (fixture) => {
    delete fixture.generated;
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

service_scheduler_summary="$tmp_dir/service-scheduler-summary.json"
write_minimal_scheduler_summary "$service_scheduler_summary" "cartulary.service_backed_scheduler_summary.v9"
assert_passes "service scheduler summary validates exact schema" \
  run_schema_validation cartulary.service_backed_scheduler_summary.v9 "$service_scheduler_summary" >/dev/null
mismatched_scheduler_output="$(assert_fails "scheduler summary rejects mismatched schema_id" \
  run_schema_validation cartulary.check_scheduler_summary.v9 "$service_scheduler_summary")"
assert_contains "$mismatched_scheduler_output" "must be equal to constant" "scheduler summary mismatched schema_id"

frontend_a11y_summary="$tmp_dir/frontend-accessibility-summary-v2.json"
write_valid_frontend_accessibility_summary_v2 "$frontend_a11y_summary"
assert_passes "frontend accessibility summary v2 validates exact schema" \
  run_schema_validation cartulary.frontend_accessibility_summary.v2 "$frontend_a11y_summary" >/dev/null

frontend_a11y_bad_status="$tmp_dir/frontend-accessibility-summary-v2-bad-status.json"
write_valid_frontend_accessibility_summary_v2 "$frontend_a11y_bad_status"
mutate_json_fixture frontend-a11y-summary-invalid-status "$frontend_a11y_bad_status"
frontend_a11y_bad_status_output="$(assert_fails "frontend accessibility summary rejects invalid scenario status" \
  run_schema_validation cartulary.frontend_accessibility_summary.v2 "$frontend_a11y_bad_status")"
assert_contains "$frontend_a11y_bad_status_output" "must be equal to one of the allowed values" "frontend accessibility invalid scenario status"

frontend_a11y_unknown_key="$tmp_dir/frontend-accessibility-summary-v2-unknown-key.json"
write_valid_frontend_accessibility_summary_v2 "$frontend_a11y_unknown_key"
mutate_json_fixture frontend-a11y-summary-unknown-key "$frontend_a11y_unknown_key"
frontend_a11y_unknown_key_output="$(assert_fails "frontend accessibility summary rejects unknown keys" \
  run_schema_validation cartulary.frontend_accessibility_summary.v2 "$frontend_a11y_unknown_key")"
assert_contains "$frontend_a11y_unknown_key_output" "must NOT have additional properties" "frontend accessibility unknown key"

frontend_a11y_missing_required="$tmp_dir/frontend-accessibility-summary-v2-missing-required.json"
write_valid_frontend_accessibility_summary_v2 "$frontend_a11y_missing_required"
mutate_json_fixture frontend-a11y-summary-missing-required "$frontend_a11y_missing_required"
frontend_a11y_missing_required_output="$(assert_fails "frontend accessibility summary rejects missing required fields" \
  run_schema_validation cartulary.frontend_accessibility_summary.v2 "$frontend_a11y_missing_required")"
assert_contains "$frontend_a11y_missing_required_output" "must have required property 'ratio'" "frontend accessibility missing required"

frontend_a11y_blocked_row="$tmp_dir/frontend-accessibility-summary-v2-blocked-row.json"
write_valid_frontend_accessibility_summary_v2 "$frontend_a11y_blocked_row"
mutate_json_fixture frontend-a11y-summary-blocked-row "$frontend_a11y_blocked_row"
frontend_a11y_blocked_row_output="$(assert_fails "frontend accessibility summary rejects blocked rows" \
  run_schema_validation cartulary.frontend_accessibility_summary.v2 "$frontend_a11y_blocked_row")"
assert_contains "$frontend_a11y_blocked_row_output" "must be equal to constant" "frontend accessibility blocked row"

frontend_a11y_preflight="$tmp_dir/frontend-accessibility-preflight-summary.json"
write_valid_frontend_accessibility_preflight_summary "$frontend_a11y_preflight"
assert_passes "frontend accessibility preflight summary validates exact schema" \
  run_schema_validation cartulary.frontend_accessibility_preflight_summary.v1 "$frontend_a11y_preflight" >/dev/null

frontend_a11y_preflight_implemented="$tmp_dir/frontend-accessibility-preflight-summary-implemented.json"
write_valid_frontend_accessibility_preflight_summary "$frontend_a11y_preflight_implemented"
mutate_json_fixture frontend-a11y-preflight-implemented-row "$frontend_a11y_preflight_implemented"
frontend_a11y_preflight_implemented_output="$(assert_fails "frontend accessibility preflight rejects implemented rows" \
  run_schema_validation cartulary.frontend_accessibility_preflight_summary.v1 "$frontend_a11y_preflight_implemented")"
assert_contains "$frontend_a11y_preflight_implemented_output" "must be equal to constant" "frontend accessibility preflight implemented row"

frontend_row_accounting="$tmp_dir/frontend-row-accounting.json"
write_valid_frontend_row_accounting "$frontend_row_accounting"
assert_passes "frontend row accounting validates exact schema" \
  run_schema_validation cartulary.frontend_row_accounting.v3 "$frontend_row_accounting" >/dev/null

frontend_row_accounting_unknown_key="$tmp_dir/frontend-row-accounting-unknown-key.json"
write_valid_frontend_row_accounting "$frontend_row_accounting_unknown_key"
mutate_json_fixture frontend-row-accounting-unknown-key "$frontend_row_accounting_unknown_key"
frontend_row_accounting_unknown_key_output="$(assert_fails "frontend row accounting rejects unknown row keys" \
  run_schema_validation cartulary.frontend_row_accounting.v3 "$frontend_row_accounting_unknown_key")"
assert_contains "$frontend_row_accounting_unknown_key_output" "must NOT have additional properties" "frontend row accounting unknown key"

frontend_row_accounting_bad_closure="$tmp_dir/frontend-row-accounting-bad-closure.json"
write_valid_frontend_row_accounting "$frontend_row_accounting_bad_closure"
mutate_json_fixture frontend-row-accounting-invalid-closure "$frontend_row_accounting_bad_closure"
frontend_row_accounting_bad_closure_output="$(assert_fails "frontend row accounting rejects invalid closure status" \
  run_schema_validation cartulary.frontend_row_accounting.v3 "$frontend_row_accounting_bad_closure")"
assert_contains "$frontend_row_accounting_bad_closure_output" "must be equal to one of the allowed values" "frontend row accounting invalid closure"

frontend_row_accounting_bad_scenario="$tmp_dir/frontend-row-accounting-bad-scenario.json"
write_valid_frontend_row_accounting "$frontend_row_accounting_bad_scenario"
mutate_json_fixture frontend-row-accounting-invalid-scenario-status "$frontend_row_accounting_bad_scenario"
frontend_row_accounting_bad_scenario_output="$(assert_fails "frontend row accounting rejects invalid scenario status" \
  run_schema_validation cartulary.frontend_row_accounting.v3 "$frontend_row_accounting_bad_scenario")"
assert_contains "$frontend_row_accounting_bad_scenario_output" "must be equal to one of the allowed values" "frontend row accounting invalid scenario"

frontend_row_accounting_bad_scope="$tmp_dir/frontend-row-accounting-bad-scope.json"
write_valid_frontend_row_accounting "$frontend_row_accounting_bad_scope"
mutate_json_fixture frontend-row-accounting-invalid-scope "$frontend_row_accounting_bad_scope"
frontend_row_accounting_bad_scope_output="$(assert_fails "frontend row accounting rejects invalid scope" \
  run_schema_validation cartulary.frontend_row_accounting.v3 "$frontend_row_accounting_bad_scope")"
assert_contains "$frontend_row_accounting_bad_scope_output" "must be equal to one of the allowed values" "frontend row accounting invalid scope"

frontend_visual_product_map="$tmp_dir/fe_p8_visual_product_map.json"
cp "$ROOT_DIR/tools/frontend_phase_maps/fe_p8_test_map.json" "$frontend_visual_product_map"
"$NODE_BIN" - "$frontend_visual_product_map" <<'JS'
const fs = require("node:fs");
const file = process.argv[2];
const manifest = JSON.parse(fs.readFileSync(file, "utf8"));
const row = manifest.rows.find((entry) => entry.id === "FE-V-P8-01");
if (!row) {
  throw new Error("FE-V-P8-01 fixture row not found");
}
row.evidence_class = "product_conformance";
row.core_req_ids = ["REQ-01-035"];
row.core_ac_ids = ["AC-013"];
fs.writeFileSync(file, `${JSON.stringify(manifest, null, 2)}\n`);
JS
frontend_visual_product_output="$(assert_fails "frontend visual rows reject product conformance" \
  run_frontend_phase_map_validation "$frontend_visual_product_map" FE-P8)"
assert_contains "$frontend_visual_product_output" "must not use product_conformance" "frontend visual product conformance rejection"

frontend_a11y_writer_missing="$tmp_dir/frontend-accessibility-summary-writer-missing.json"
frontend_a11y_writer_missing_output="$(assert_fails "frontend accessibility summary writer rejects missing implemented evidence" \
  run_accessibility_summary_writer --output "$frontend_a11y_writer_missing" --status pass --mode evidence)"
assert_contains "$frontend_a11y_writer_missing_output" "frontend accessibility summary failed: frontend accessibility evidence summary status is fail" "frontend accessibility writer missing evidence"

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

duplicate_member_registry="$tmp_dir/phase_registry_duplicate_member.json"
cat >"$duplicate_member_registry" <<'JSON'
{
  "schema_id": "cartulary.phase_registry.v1",
  "schema_id": "cartulary.phase_registry.v1",
  "phases": []
}
JSON
duplicate_member_output="$(assert_fails "duplicate JSON object member" run_shape_check phase-registry "$duplicate_member_registry")"
assert_contains "$duplicate_member_output" "duplicate object member \"schema_id\"" "duplicate JSON object member"

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
mutate_json_fixture scheduler-manifest-stale-schema "$stale_schedule"
stale_schedule_output="$(assert_fails "stale generated schedule shape" run_shape_check scheduler-manifest "$stale_schedule")"
assert_contains "$stale_schedule_output" "must declare schema_id cartulary.scheduler_manifest.v1" "stale generated schedule shape"

unknown_work_unit_key="$tmp_dir/check_schedule_unknown_work_unit_key.json"
write_valid_check_schedule "$unknown_work_unit_key"
mutate_json_fixture check-schedule-unknown-work-unit-key "$unknown_work_unit_key"
unknown_work_unit_output="$(assert_fails "unknown scheduler work unit key" run_shape_check scheduler-manifest "$unknown_work_unit_key")"
assert_contains "$unknown_work_unit_output" "unknown key legacy_key" "unknown check schedule work unit key"

empty_check_schedules="$tmp_dir/check_schedule_empty_schedules.json"
write_valid_check_schedule "$empty_check_schedules"
mutate_json_fixture check-schedule-empty-schedules "$empty_check_schedules"
empty_check_schedules_output="$(assert_fails "empty scheduler schedules" run_shape_check scheduler-manifest "$empty_check_schedules")"
assert_contains "$empty_check_schedules_output" "schedules must be a non-empty array" "empty check schedule schedules"

invalid_check_env="$tmp_dir/check_schedule_invalid_env.json"
write_valid_check_schedule "$invalid_check_env"
mutate_json_fixture check-schedule-invalid-env-name "$invalid_check_env"
invalid_check_env_output="$(assert_fails "invalid scheduler env name" run_shape_check scheduler-manifest "$invalid_check_env")"
assert_contains "$invalid_check_env_output" "env key has invalid value" "invalid check schedule env name"

invalid_check_make_prerequisite_policy="$tmp_dir/check_schedule_invalid_make_prerequisite_policy.json"
write_valid_check_schedule "$invalid_check_make_prerequisite_policy"
mutate_json_fixture check-schedule-invalid-make-prerequisite-policy "$invalid_check_make_prerequisite_policy"
invalid_check_make_prerequisite_policy_output="$(assert_fails "invalid scheduler make prerequisite policy" run_shape_check scheduler-manifest "$invalid_check_make_prerequisite_policy")"
assert_contains "$invalid_check_make_prerequisite_policy_output" "make_prerequisite_policy must be one of" "invalid scheduler make prerequisite policy"

browser_worker_overlap="$tmp_dir/check_schedule_browser_worker_overlap.json"
write_valid_check_schedule "$browser_worker_overlap"
mutate_json_fixture check-schedule-browser-worker-overlap "$browser_worker_overlap"
browser_worker_overlap_output="$(assert_fails "overlapping browser worker slots" run_shape_check scheduler-manifest "$browser_worker_overlap")"
assert_contains "$browser_worker_overlap_output" "overlapping worker-admin slot" "overlapping browser worker slots"

missing_command_required="$tmp_dir/check_schedule_command_missing_required.json"
write_valid_check_schedule "$missing_command_required"
mutate_json_fixture check-schedule-command-missing-required "$missing_command_required"
missing_command_required_output="$(assert_fails "missing scheduler command required field" run_shape_check scheduler-manifest "$missing_command_required")"
assert_contains "$missing_command_required_output" "command.target must be a non-empty string" "missing scheduler command required field"

forbidden_command_field="$tmp_dir/check_schedule_command_forbidden_field.json"
write_valid_check_schedule "$forbidden_command_field"
mutate_json_fixture check-schedule-command-forbidden-field "$forbidden_command_field"
forbidden_command_field_output="$(assert_fails "forbidden scheduler command field" run_shape_check scheduler-manifest "$forbidden_command_field")"
assert_contains "$forbidden_command_field_output" "command has unknown key shard" "forbidden scheduler command field"

unknown_command_type="$tmp_dir/check_schedule_command_unknown_type.json"
write_valid_check_schedule "$unknown_command_type"
mutate_json_fixture check-schedule-command-unknown-type "$unknown_command_type"
unknown_command_type_output="$(assert_fails "unknown scheduler command type" run_shape_check scheduler-manifest "$unknown_command_type")"
assert_contains "$unknown_command_type_output" "command.type must be one of" "unknown scheduler command type"

wrong_command_field_type="$tmp_dir/check_schedule_command_wrong_field_type.json"
write_valid_check_schedule "$wrong_command_field_type"
mutate_json_fixture check-schedule-command-wrong-field-type "$wrong_command_field_type"
wrong_command_field_type_output="$(assert_fails "wrong scheduler command field type" run_shape_check scheduler-manifest "$wrong_command_field_type")"
assert_contains "$wrong_command_field_type_output" "command.target must be a non-empty string" "wrong scheduler command field type"

service_backed_schedule="$tmp_dir/service_backed_schedule.json"
write_valid_service_backed_schedule "$service_backed_schedule"
assert_contains "$(assert_passes "valid service-backed schedule" run_shape_check scheduler-manifest "$service_backed_schedule")" \
  "json shape check passed" \
  "valid service-backed schedule"

unknown_service_source_key="$tmp_dir/service_backed_schedule_unknown_source_key.json"
write_valid_service_backed_schedule "$unknown_service_source_key"
mutate_json_fixture service-backed-unknown-source-key "$unknown_service_source_key"
unknown_service_source_output="$(assert_fails "unknown service-backed work unit key" run_shape_check scheduler-manifest "$unknown_service_source_key")"
assert_contains "$unknown_service_source_output" "unknown key legacy_key" "unknown service-backed source key"

empty_service_sources="$tmp_dir/service_backed_schedule_empty_sources.json"
write_valid_service_backed_schedule "$empty_service_sources"
mutate_json_fixture service-backed-empty-sources "$empty_service_sources"
empty_service_sources_output="$(assert_fails "empty service-backed work units" run_shape_check scheduler-manifest "$empty_service_sources")"
assert_contains "$empty_service_sources_output" "work_units must be a non-empty array" "empty service-backed sources"

missing_service_generated="$tmp_dir/service_backed_schedule_missing_generated.json"
write_valid_service_backed_schedule "$missing_service_generated"
mutate_json_fixture service-backed-missing-generated "$missing_service_generated"
missing_service_generated_output="$(assert_fails "missing scheduler generated metadata" run_shape_check scheduler-manifest "$missing_service_generated")"
assert_contains "$missing_service_generated_output" "generated must be an object" "missing service-backed generated metadata"

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

invalid_extension_key="$tmp_dir/tool-run-summary-invalid-extension-key.json"
write_valid_tool_run_summary "$invalid_extension_key"
mutate_json_fixture tool-run-summary-invalid-extension-key "$invalid_extension_key"
invalid_extension_output="$(assert_fails "invalid tool run extension key" run_shape_check tool-run-summary "$invalid_extension_key")"
assert_contains "$invalid_extension_output" "invalid extension key scheduler_timing" "invalid tool run extension key"

failed_null_reason="$tmp_dir/tool-run-summary-failed-null-reason.json"
write_valid_tool_run_summary "$failed_null_reason"
mutate_json_fixture tool-run-summary-failed-null-reason "$failed_null_reason"
failed_null_reason_output="$(assert_fails "failed tool run null failure reason" run_shape_check tool-run-summary "$failed_null_reason")"
assert_contains "$failed_null_reason_output" "failure_reason must be non-null when status is fail" "failed tool run null failure reason"

missing_scheduler_timing="$tmp_dir/tool-run-summary-missing-scheduler-timing.json"
write_valid_tool_run_summary "$missing_scheduler_timing"
mutate_json_fixture tool-run-summary-missing-scheduler-timing "$missing_scheduler_timing"
missing_scheduler_timing_output="$(assert_fails "missing tool run scheduler_timing" run_shape_check tool-run-summary "$missing_scheduler_timing")"
assert_contains "$missing_scheduler_timing_output" "scheduler_timing is required" "missing tool run scheduler_timing"

missing_schema_id="$tmp_dir/tool-run-summary-missing-schema-id.json"
write_valid_tool_run_summary "$missing_schema_id"
mutate_json_fixture tool-run-summary-missing-schema-id "$missing_schema_id"
missing_schema_id_output="$(assert_fails "missing tool run schema_id" run_shape_check tool-run-summary "$missing_schema_id")"
assert_contains "$missing_schema_id_output" "schema_id is required" "missing tool run schema_id"

fallow_static_summary="$tmp_dir/fallow-static-summary.json"
write_valid_fallow_static_summary "$fallow_static_summary"
assert_contains "$(assert_passes "valid fallow static summary" run_shape_check fallow-static-summary "$fallow_static_summary")" \
  "json shape check passed" \
  "valid fallow static summary"

fallow_static_blocking="$tmp_dir/fallow-static-summary-blocking.json"
write_valid_fallow_static_summary "$fallow_static_blocking"
mutate_json_fixture fallow-static-summary-blocking "$fallow_static_blocking"
fallow_static_blocking_output="$(assert_fails "fallow static summary rejects blocking mode" run_shape_check fallow-static-summary "$fallow_static_blocking")"
assert_contains "$fallow_static_blocking_output" "must be equal to constant" "fallow static summary blocking mode"

fallow_static_unknown_key="$tmp_dir/fallow-static-summary-unknown-key.json"
write_valid_fallow_static_summary "$fallow_static_unknown_key"
mutate_json_fixture fallow-static-summary-unknown-key "$fallow_static_unknown_key"
fallow_static_unknown_key_output="$(assert_fails "fallow static summary rejects unknown key" run_shape_check fallow-static-summary "$fallow_static_unknown_key")"
assert_contains "$fallow_static_unknown_key_output" "must NOT have additional properties" "fallow static summary unknown key"

web_e2e_stack="$tmp_dir/web-e2e-stack.json"
write_valid_web_e2e_stack "$web_e2e_stack"
run_schema_validation cartulary.web_e2e_stack.v3 "$web_e2e_stack" >/dev/null

bad_web_e2e_stack="$tmp_dir/web-e2e-stack-bad-mode.json"
write_valid_web_e2e_stack "$bad_web_e2e_stack"
mutate_json_fixture web-e2e-stack-wrong-frontend-mode "$bad_web_e2e_stack"
bad_web_e2e_stack_output="$(assert_fails "web e2e stack rejects dev frontend mode" run_schema_validation cartulary.web_e2e_stack.v3 "$bad_web_e2e_stack")"
assert_contains "$bad_web_e2e_stack_output" "must be equal to constant" "web e2e stack frontend mode"

browser_startup_diagnostics="$tmp_dir/browser-startup-diagnostics.json"
write_valid_browser_startup_diagnostics "$browser_startup_diagnostics"
run_schema_validation cartulary.browser_startup_diagnostics.v1 "$browser_startup_diagnostics" >/dev/null

browser_startup_resource_conflict="$tmp_dir/browser-startup-diagnostics-resource-conflict.json"
write_valid_browser_startup_diagnostics "$browser_startup_resource_conflict"
"$NODE_BIN" - "$browser_startup_resource_conflict" <<'JS'
const fs = require("node:fs");

const file = process.argv[2];
const payload = JSON.parse(fs.readFileSync(file, "utf8"));
payload.failure_reason = "resource_conflict";
payload.message = "Port 4173 is already in use";
delete payload.inotify;
fs.writeFileSync(file, `${JSON.stringify(payload, null, 2)}\n`);
JS
run_schema_validation cartulary.browser_startup_diagnostics.v1 "$browser_startup_resource_conflict" >/dev/null

bad_browser_startup_diagnostics="$tmp_dir/browser-startup-diagnostics-unknown-key.json"
write_valid_browser_startup_diagnostics "$bad_browser_startup_diagnostics"
mutate_json_fixture browser-startup-diagnostics-unknown-key "$bad_browser_startup_diagnostics"
bad_browser_startup_output="$(assert_fails "browser startup diagnostics reject unknown key" run_schema_validation cartulary.browser_startup_diagnostics.v1 "$bad_browser_startup_diagnostics")"
assert_contains "$bad_browser_startup_output" "must NOT have additional properties" "browser startup diagnostics unknown key"

agent_finalize_summary="$tmp_dir/agent-finalize-summary.json"
write_valid_agent_finalize_summary "$agent_finalize_summary"
assert_contains "$(assert_passes "valid agent finalize summary" run_shape_check agent-finalize-summary "$agent_finalize_summary")" \
  "json shape check passed" \
  "valid agent finalize summary"

readiness_cache_record="$tmp_dir/readiness-cache-record.json"
write_valid_cache_record cartulary.cache.readiness.v1 readiness "$readiness_cache_record"
run_schema_validation cartulary.cache.readiness.v1 "$readiness_cache_record" >/dev/null

build_cache_record="$tmp_dir/build-cache-record.json"
write_valid_cache_record cartulary.cache.build_artifact.v1 build-artifact "$build_cache_record"
run_schema_validation cartulary.cache.build_artifact.v1 "$build_cache_record" >/dev/null

agent_action_cache_record="$tmp_dir/agent-action-cache-record.json"
write_valid_agent_finalize_action_cache_record "$agent_action_cache_record"
run_schema_validation cartulary.agent_finalize_action_cache_record.v1 "$agent_action_cache_record" >/dev/null

render_cache_record="$tmp_dir/render-cache-record.json"
write_valid_execution_topology_render_cache_record "$render_cache_record"
run_schema_validation cartulary.execution_topology_render_cache.v1 "$render_cache_record" >/dev/null

bad_agent_finalize_action="$tmp_dir/agent-finalize-summary-bad-action.json"
write_valid_agent_finalize_summary "$bad_agent_finalize_action"
mutate_json_fixture agent-finalize-summary-bad-action-status "$bad_agent_finalize_action"
bad_agent_finalize_action_output="$(assert_fails "invalid agent finalize action status" run_shape_check agent-finalize-summary "$bad_agent_finalize_action")"
assert_contains "$bad_agent_finalize_action_output" "must be equal to one of the allowed values" "invalid agent finalize action status"

unknown_agent_finalize_key="$tmp_dir/agent-finalize-summary-unknown-key.json"
write_valid_agent_finalize_summary "$unknown_agent_finalize_key"
mutate_json_fixture agent-finalize-summary-unknown-key "$unknown_agent_finalize_key"
unknown_agent_finalize_key_output="$(assert_fails "unknown agent finalize key" run_shape_check agent-finalize-summary "$unknown_agent_finalize_key")"
assert_contains "$unknown_agent_finalize_key_output" "must NOT have additional properties" "unknown agent finalize key"

echo "json shape harness tests passed"
