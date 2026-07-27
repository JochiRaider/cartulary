#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
CHECKER="$ROOT_DIR/tools/harness/generated-artifacts/check-json-shapes.mjs"
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

  "$NODE_BIN" "$ROOT_DIR/tools/harness/contract/harness-contract-cli.mjs" \
    validate-schema "$schema_id" "$file"
}

write_valid_execution_topology() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.execution_topology.v4",
  "execution_dependencies": [
    {
      "id": "backend_unit",
      "target": "backend-unit",
      "category": "backend",
      "order": 0,
      "service_backed": false
    }
  ],
  "task_surface_owner": "tools/task_surface_owner.json"
}
JSON
}

write_valid_check_schedule() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v2",
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
          "make_prerequisite_policy": "skip",
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
  "schema_id": "cartulary.scheduler_manifest.v2",
  "generated": {
    "generator": "synthetic"
  },
  "schedules": [
    {
      "target": "test",
      "scheduler_kind": "service_backed",
      "capacity_profile": "service_backed_full",
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
          "make_prerequisite_policy": "skip",
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
  "schema_id": "cartulary.browser_e2e_batch_manifest.v6",
  "runtime_profiles": [
    {
      "id": "default",
      "kind": "default"
    }
  ],
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
      "display_name": "host CPU",
      "schedulers": ["check", "sequence"],
      "display_order": 10,
      "capacity": {
        "auto_policy": "host_cpu",
        "override_env": "CHECK_HOST_CPU_JOBS",
        "max_limit": 256
      }
    },
    {
      "name": "host_io",
      "display_name": "host IO",
      "schedulers": ["check", "sequence"],
      "display_order": 20,
      "capacity": {
        "auto_policy": "host_io",
        "override_env": "CHECK_HOST_IO_JOBS",
        "max_limit": 256
      }
    },
    {
      "name": "suite_service_stack",
      "display_name": "suite service stack",
      "schedulers": ["check"],
      "display_order": 30,
      "capacity": {
        "default_limit": 1,
        "max_limit": 256
      }
    },
    {
      "name": "migration_scratch_postgres",
      "display_name": "migration scratch Postgres",
      "schedulers": ["check"],
      "display_order": 40,
      "capacity": {
        "default_limit": 1,
        "max_limit": 256
      }
    },
    {
      "name": "go_cpu",
      "display_name": "Go CPU",
      "schedulers": ["service_backed", "test_slice"],
      "display_order": 110,
      "capacity": {
        "auto_policy": "service_backed_go_cpu",
        "override_env": "CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT",
        "max_limit": 256
      }
    },
    {
      "name": "go_io",
      "display_name": "Go IO",
      "schedulers": ["service_backed", "test_slice"],
      "display_order": 120,
      "capacity": {
        "auto_policy": "service_backed_go_io",
        "override_env": "CARTULARY_SERVICE_BACKED_GO_IO_LIMIT",
        "max_limit": 256
      }
    },
    {
      "name": "browser_stack",
      "display_name": "browser stack",
      "schedulers": ["check", "service_backed", "test_slice"],
      "display_order": 130,
      "capacity": {
        "auto_policy": "service_backed_browser_stack",
        "override_env": "CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT",
        "max_limit": 256
      }
    },
    {
      "name": "object_store",
      "display_name": "object store",
      "schedulers": ["check", "service_backed", "test_slice"],
      "display_order": 140,
      "capacity": {
        "default_limit": 32,
        "max_limit": 256
      }
    },
    {
      "name": "postgres",
      "display_name": "Postgres",
      "schedulers": ["check", "service_backed", "test_slice"],
      "display_order": 150,
      "capacity": {
        "default_limit": 32,
        "max_limit": 256
      }
    },
    {
      "name": "process",
      "display_name": "process slots",
      "schedulers": ["check", "sequence", "service_backed", "test_slice"],
      "display_order": 160,
      "capacity": {
        "auto_policy": "host_process_slots",
        "max_limit": 256
      }
    },
    {
      "name": "postgres_reset",
      "display_name": "Postgres reset",
      "schedulers": ["check", "service_backed", "test_slice"],
      "display_order": 170,
      "capacity": {
        "auto_policy": "service_backed_postgres_reset",
        "override_env": "CARTULARY_SERVICE_BACKED_POSTGRES_RESET_LIMIT",
        "max_limit": 8
      }
    },
    {
      "name": "postgres_clone",
      "display_name": "Postgres clone",
      "schedulers": ["check", "service_backed", "test_slice"],
      "display_order": 175,
      "capacity": {
        "auto_policy": "service_backed_postgres_clone",
        "override_env": "CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT",
        "max_limit": 8
      }
    }
  ],
  "templates": [
    {
      "name": "browser_stage",
      "prefix": "browser_stage_",
      "display_name": "browser stage",
      "schedulers": ["check", "service_backed", "test_slice"],
      "display_order": 135,
      "max_limit": 8
    }
  ],
  "capacity_profiles": [
    {
      "name": "check_default",
      "scheduler": "check",
      "resources": [
        "host_cpu",
        "host_io",
        "suite_service_stack",
        "migration_scratch_postgres"
      ]
    },
    {
      "name": "sequence_adaptive",
      "scheduler": "sequence",
      "resources": [
        "host_cpu",
        "host_io",
        "process"
      ]
    },
    {
      "name": "service_backed_full",
      "scheduler": "service_backed",
      "resources": [
        "postgres",
        "object_store",
        "go_cpu",
        "go_io",
        "postgres_reset",
        "postgres_clone",
        "process",
        "browser_stack"
      ]
    },
    {
      "name": "service_backed_backend",
      "scheduler": "service_backed",
      "resources": [
        "postgres",
        "object_store",
        "go_cpu",
        "go_io",
        "postgres_reset",
        "postgres_clone",
        "process"
      ]
    },
    {
      "name": "test_slice_default",
      "scheduler": "test_slice",
      "resources": [
        "postgres",
        "object_store",
        "go_cpu",
        "go_io",
        "postgres_reset",
        "postgres_clone",
        "process",
        "browser_stack"
      ]
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
  "schema_id": "cartulary.tool_run_summary.v5",
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
      "path_kind": "file",
      "format": "json",
      "path": "json-shape-check/tool-run-summary.json"
    }
  ],
  "log_artifacts": [],
  "work_units": [],
  "evidence_targets": [],
  "helper_units": [],
  "counts": {
    "steps": 0,
    "tests": 0,
    "failed": 0,
    "non_test": 0,
    "non_test_failed": 0,
    "packages": 0
  },
  "step_accounting": {
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

write_valid_govulncheck_findings() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.govulncheck_findings.v1",
  "tool": "govulncheck",
  "status": "fail",
  "config": null,
  "counts": {
    "raw_event_count": 2,
    "osv_count": 1,
    "finding_count": 1,
    "blocking_count": 1,
    "reachability": {
      "module": 0,
      "package": 0,
      "symbol": 1
    }
  },
  "vulnerability_ids": ["GO-2099-0001"],
  "blocking_vulnerability_ids": ["GO-2099-0001"],
  "findings": [
    {
      "id": "GO-2099-0001",
      "aliases": ["CVE-2099-0001"],
      "summary": "synthetic reachable vulnerability",
      "fixed_version": "v1.2.3",
      "fixed_versions": ["v1.2.3"],
      "affected_packages": ["example.test/module"],
      "reachability": "symbol",
      "blocking": true,
      "modules": ["example.test/module"],
      "packages": ["example.test/module/pkg"],
      "symbols": [
        {
          "package": "example.test/module/pkg",
          "receiver": "",
          "function": "Vulnerable",
          "position": {
            "filename": "pkg/vulnerable.go",
            "line": 12,
            "column": 3
          }
        }
      ],
      "trace": [
        {
          "module": "example.test/module",
          "version": "v1.0.0",
          "package": "example.test/module/pkg",
          "receiver": "",
          "function": "Vulnerable",
          "position": {
            "filename": "pkg/vulnerable.go",
            "line": 12,
            "column": 3
          }
        }
      ]
    }
  ]
}
JSON
}

write_valid_fallow_static_summary() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.fallow_static_summary.v2",
  "target": "frontend-fallow-static",
  "generated_at": "2026-01-01T00:00:00Z",
  "mode": "static_reachability_report",
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
      "message": "Current Fallow static profile findings are retained as non-blocking static-analysis evidence."
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
  "observed_failed_work_units": [],
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

write_valid_scheduler_pressure_summary() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.scheduler_pressure_summary.v4",
  "target": "check",
  "scheduler_kind": "check",
  "status": "pass",
  "total_work_units": 2,
  "completed_work_units": 2,
  "scheduler_total_duration_ms": 42,
  "target_counts": {
    "backend-unit": 1,
    "frontend-unit": 1
  },
  "lane_duration_ms": {
    "backend-unit": 25,
    "frontend-unit": 17
  },
  "resource_claim_counts": {
    "host_cpu": 2
  },
  "fixture_class_counts": {
    "none": 2
  },
  "row_fixture_pressure": [
    {
      "target": "backend-store",
      "row_id": "module.workbook.store",
      "execution_family": "graph_projection_coordination",
      "fixture_class": "transaction_or_shared_postgres",
      "work_unit_count": 1,
      "duration_ms": 25
    }
  ],
  "execution_family_fixture_pressure": [
    {
      "target": "backend-store",
      "execution_family": "graph_projection_coordination",
      "fixture_class": "transaction_or_shared_postgres",
      "work_unit_count": 1,
      "duration_ms": 25
    }
  ],
  "fixture_proof_records": [
    {
      "target": "backend-store",
      "row_id": "module.workbook.store",
      "execution_family": "graph_projection_coordination",
      "symbol": "TestCoordinationProjectionSortFilterGroup_Unit",
      "fixture_policy": "transaction",
      "proof_kind": "transaction",
      "proof_status": "accepted",
      "proof_ref": "catalog-row:module.graphprojection.storage.lifecycle",
      "reason": "Store-layer symbol uses rollback-scoped StartStore fixture.",
      "dirty_tables": []
    }
  ],
  "fixture_tier_proofs": [
    {
      "schema_id": "cartulary.fixture_tier_proof.v2",
      "target": "backend-store",
      "owner_id": "module.graphprojection",
      "row_id": "module.workbook.store",
      "execution_family": "graph_projection_coordination",
      "symbol": "TestCoordinationProjectionSortFilterGroup_Unit",
      "effective_fixture_policy": "transaction",
      "proof_kind": "transaction",
      "proof_status": "accepted",
      "proof_ref": "catalog-row:module.graphprojection.storage.lifecycle",
      "reason": "Store-layer symbol uses rollback-scoped StartStore fixture.",
      "execution_boundary": "rollback_transaction",
      "observed_surfaces": {
        "postgres": "observed",
        "auth_session_bootstrap": "not_observed",
        "route_idempotency": "not_observed",
        "jobs": "not_observed",
        "object_store": "not_observed",
        "websocket_observer": "not_observed",
        "cross_connection_observer": "not_observed",
        "process_lifecycle": "not_observed",
        "schema_migration": "not_observed"
      },
      "reset_surface": {
        "postgres_reset": "rollback",
        "postgres_dirty_tables": [],
        "postgres_fk_closure": "not_applicable",
        "goose_metadata": "not_applicable",
        "route_idempotency": "not_applicable",
        "jobs": "not_applicable",
        "object_store": "none"
      },
      "final_verdict": "accepted"
    }
  ],
  "slowest_work_units": [
    {
      "id": "backend-unit",
      "label": "backend-unit",
      "status": 0,
      "duration_ms": 25
    }
  ],
  "reused_accounting_counts": {
    "executed": 2,
    "reused": 0,
    "skipped": 0
  },
  "readiness_attribution_counts": {
    "frontend_install": 1
  },
  "readiness_attribution_duration_ms": {
    "frontend_install": 0
  },
  "readiness_attribution_units": [
    {
      "id": "check-frontend-install",
      "label": "check-frontend-install",
      "timing_role": "readiness",
      "readiness_class": "frontend_install",
      "duration_ms": 0,
      "warm_threshold_ms": 30000,
      "warm_status": "within_threshold",
      "reason": "pnpm-managed workspace dependency readiness"
    }
  ],
  "generated_at": "2026-01-01T00:00:42Z"
}
JSON
}

write_valid_fixture_tier_proof() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.fixture_tier_proof.v2",
  "target": "backend-integration",
  "owner_id": "module.recovery",
  "row_id": "module.imports.integration",
  "execution_family": "backend_integration_import_boundary_negative",
  "symbol": "TestImportInvalidRequestHasNoDurableRows",
  "effective_fixture_policy": "template_clone",
  "proof_kind": "template_clone",
  "proof_status": "retained",
  "reason": "Route-side auth, session, bootstrap, idempotency, and job surfaces are not yet reset-proofed.",
  "execution_boundary": "isolated_template_clone",
  "observed_surfaces": {
    "postgres": "observed",
    "auth_session_bootstrap": "observed",
    "route_idempotency": "observed",
    "jobs": "observed",
    "object_store": "not_observed",
    "websocket_observer": "not_observed",
    "cross_connection_observer": "not_observed",
    "process_lifecycle": "not_observed",
    "schema_migration": "not_observed"
  },
  "reset_surface": {
    "postgres_reset": "clone_isolation",
    "postgres_dirty_tables": [],
    "postgres_fk_closure": "not_applicable",
    "goose_metadata": "not_applicable",
    "route_idempotency": "not_applicable",
    "jobs": "not_applicable",
    "object_store": "none"
  },
  "final_verdict": "retained"
}
JSON
}

write_valid_release_readiness_evidence() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.release_readiness_evidence.v2",
  "status": "pass",
  "generated_at": "2026-01-01T00:00:00.000Z",
  "run_root": ".cartulary/test-results/run",
  "evidence_records": [
    {
      "evidence_id": "target:check",
      "source_target": "check",
      "schema_id": "cartulary.test_target_summary.v4",
      "owner_refs": [
        "harness.release.verification.current_owner_evidence_only"
      ],
      "evidence_class": "product_conformance",
      "conformance_effect": "product_conformance",
      "claim_publication_effect": "not_claim_bearing",
      "release_gate_effect": "required",
      "run_root": ".cartulary/test-results/run",
      "artifact_refs": [
        {
          "role": "target_summary",
          "path_kind": "file",
          "format": "json",
          "path": "check/target-summary.json"
        }
      ],
      "source_refs": [],
      "status": "passed"
    },
    {
      "evidence_id": "owner-partition:browser-e2e-visual:web.workbook",
      "source_target": "browser-e2e-visual",
      "schema_id": "cartulary.test_owner_summary.v2",
      "owner_refs": [
        "harness.evidence_accounting.verification.semantic_evidence_identity"
      ],
      "evidence_class": "design_direction",
      "conformance_effect": "no_product_conformance",
      "claim_publication_effect": "not_claim_bearing",
      "release_gate_effect": "required",
      "run_root": ".cartulary/test-results/run",
      "artifact_refs": [
        {
          "role": "test_owner_summary",
          "path_kind": "file",
          "format": "json",
          "path": "browser-e2e-visual/owners/web.workbook/test-owner-summary.json"
        }
      ],
      "source_refs": [],
      "status": "passed"
    }
  ],
  "rollup": {
    "total": 2,
    "passed": 2,
    "failed": 0,
    "missing": 0,
    "blocked": 0,
    "stale": 0,
    "diagnostic_only": 0,
    "required_total": 2,
    "required_passed": 2,
    "required_failed": 0
  },
  "failures": []
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
      "action_id": "generated_structure_refresh",
      "description": "Refresh generated task and topology artifacts, then verify no unsupported drift remains.",
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
        "input_profile_id": "agent_finalize.generated_structure_refresh.v2",
        "input_digest_sha256": "sha256:1111111111111111111111111111111111111111111111111111111111111111",
        "output_digest_sha256": "sha256:2222222222222222222222222222222222222222222222222222222222222222",
        "record_path": ".cache/cartulary/agent-finalize-action-cache/generated_structure_refresh/record.json",
        "reason_code": "cache_record_missing"
      },
      "substeps": [
        {
          "id": "generate",
          "target": "generate",
          "command_kind": "make_target",
          "requires_results_dir": false,
          "mutates_repo": true,
          "status": "pass",
          "started_at": "2026-01-01T00:00:00Z",
          "completed_at": "2026-01-01T00:00:01Z",
          "duration_ms": 1000,
          "exit_code": 0,
          "summary_json": ".cartulary/test-results/run/generate/tool-run-summary.json",
          "stdout_log": ".cartulary/test-results/run/generate/generate/stdout.log",
          "stderr_log": ".cartulary/test-results/run/generate/generate/stderr.log",
          "skipped_reason": null
        }
      ]
    }
  ],
  "failures": [],
  "child_artifacts": [
    {
      "role": "generated_structure_refresh_generate_summary",
      "kind": "json",
      "path": ".cartulary/test-results/run/generate/tool-run-summary.json"
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

write_valid_same_run_helper_artifact_ref() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.same_run_helper_artifact_ref.v2",
  "run_id": "run",
  "run_root": ".cartulary/test-results/run",
  "helper_target": "helper-target",
  "producer_work_unit_id": "helper-target",
  "reuse_scope": "same_run_only",
  "accounting_mode": "helper_reused",
  "scheduler_reused": false,
  "declared_inputs": [
    {
      "role": "step_summary",
      "path_kind": "file",
      "format": "json",
      "path": "helper-target/helper-target/step-summary.json",
      "sha256": "sha256:0000000000000000000000000000000000000000000000000000000000000000"
    }
  ],
  "producer_artifacts": [
    {
      "role": "step_summary",
      "path_kind": "file",
      "format": "json",
      "path": "helper-target/helper-target/step-summary.json",
      "sha256": "sha256:0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "role": "stdout_log",
      "path_kind": "file",
      "format": "log",
      "path": "helper-target/helper-target/stdout.log",
      "sha256": "sha256:1111111111111111111111111111111111111111111111111111111111111111"
    }
  ],
  "consumer_refs": [
    {
      "consumer_target": "check",
      "consumer_work_unit_id": "check",
      "accounting_mode": "helper_reused"
    }
  ],
  "input_digest_sha256": "sha256:2222222222222222222222222222222222222222222222222222222222222222",
  "output_digest_sha256": "sha256:3333333333333333333333333333333333333333333333333333333333333333",
  "failure_behavior": "fail_closed_on_missing_or_digest_mismatch",
  "created_at": "2026-01-01T00:00:00Z"
}
JSON
}

write_valid_agent_finalize_action_cache_record() {
  local file="$1"

  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.agent_finalize_action_cache_record.v1",
  "action_id": "generated_structure_refresh",
  "command_id": "cartulary.harness.command.agent_finalize.v1",
  "action_contract_version": "v1",
  "input_profile_id": "agent_finalize.generated_structure_refresh.v2",
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
  "generator": "tools/harness/generated-artifacts/render-execution-topology-artifacts.mjs",
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
        path_kind: "file",
        format: "json",
        path: "json-shape-check/z.json",
      },
      {
        role: "a_artifact",
        path_kind: "file",
        format: "json",
        path: "json-shape-check/a.json",
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
  "govulncheck-findings-unknown-key": (fixture) => {
    fixture.legacy_key = true;
  },
  "govulncheck-findings-missing-count": (fixture) => {
    delete fixture.counts.raw_event_count;
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
  "same-run-helper-ref-scheduler-reused": (fixture) => {
    fixture.scheduler_reused = true;
  },
  "same-run-helper-ref-missing-digest": (fixture) => {
    delete fixture.producer_artifacts[0].sha256;
  },
  "same-run-helper-ref-old-run-scope": (fixture) => {
    fixture.reuse_scope = "retained_run";
  },
  "scheduler-pressure-summary-unknown-key": (fixture) => {
    fixture.legacy_key = true;
  },
  "scheduler-pressure-summary-missing-required": (fixture) => {
    delete fixture.resource_claim_counts;
  },
  "fixture-tier-proof-unknown-key": (fixture) => {
    fixture.legacy_key = true;
  },
  "fixture-tier-proof-missing-reset-surface": (fixture) => {
    delete fixture.reset_surface;
  },













  "release-readiness-missing-effect": (fixture) => {
    delete fixture.evidence_records[0].conformance_effect;
  },
  "release-readiness-empty-owner-refs": (fixture) => {
    fixture.evidence_records[0].owner_refs = [];
  },
  "release-readiness-ambiguous-visual-conformance": (fixture) => {
    fixture.evidence_records[1].conformance_effect = "maybe_product_conformance";
  },
  "scheduler-manifest-stale-schema": (fixture) => {
    fixture.schema_id = "cartulary.check_schedule.v12";
  },
  "scheduler-manifest-unsupported-capacity-profile": (fixture) => {
    fixture.schedules[0].capacity_profile = "service_backed_default";
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
  "check-schedule-missing-make-prerequisite-policy": (fixture) => {
    delete fixture.schedules[0].work_units[0].make_prerequisite_policy;
  },
  "check-schedule-invalid-timeout-seconds": (fixture) => {
    fixture.schedules[0].work_units[0].timeout_seconds = 0;
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
    fixture.resources[2].capacity.auto_policy = "host_cpu";
  },
  "scheduler-registry-unknown-auto-policy": (fixture) => {
    fixture.resources[0].capacity.auto_policy = "host_cpu_auto";
  },
  "scheduler-registry-unknown-key": (fixture) => {
    fixture.legacy_key = true;
  },
  "execution-topology-duplicate-target": (fixture) => {
    fixture.targets.push({ ...fixture.targets.find((target) => target.name === "backend-unit") });
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

service_scheduler_summary="$tmp_dir/service-scheduler-summary.json"
write_minimal_scheduler_summary "$service_scheduler_summary" "cartulary.service_backed_scheduler_summary.v10"
assert_passes "service scheduler summary validates exact schema" \
  run_schema_validation cartulary.service_backed_scheduler_summary.v10 "$service_scheduler_summary" >/dev/null
mismatched_scheduler_output="$(assert_fails "scheduler summary rejects mismatched schema_id" \
  run_schema_validation cartulary.check_scheduler_summary.v10 "$service_scheduler_summary")"
assert_contains "$mismatched_scheduler_output" "must be equal to constant" "scheduler summary mismatched schema_id"

scheduler_pressure_summary="$tmp_dir/scheduler-pressure-summary.json"
write_valid_scheduler_pressure_summary "$scheduler_pressure_summary"
assert_passes "scheduler pressure summary validates exact schema" \
  run_schema_validation cartulary.scheduler_pressure_summary.v4 "$scheduler_pressure_summary" >/dev/null

scheduler_pressure_unknown_key="$tmp_dir/scheduler-pressure-summary-unknown-key.json"
write_valid_scheduler_pressure_summary "$scheduler_pressure_unknown_key"
mutate_json_fixture scheduler-pressure-summary-unknown-key "$scheduler_pressure_unknown_key"
scheduler_pressure_unknown_key_output="$(assert_fails "scheduler pressure summary rejects unknown keys" \
  run_schema_validation cartulary.scheduler_pressure_summary.v4 "$scheduler_pressure_unknown_key")"
assert_contains "$scheduler_pressure_unknown_key_output" "must NOT have additional properties" "scheduler pressure summary unknown key"

scheduler_pressure_missing_required="$tmp_dir/scheduler-pressure-summary-missing-required.json"
write_valid_scheduler_pressure_summary "$scheduler_pressure_missing_required"
mutate_json_fixture scheduler-pressure-summary-missing-required "$scheduler_pressure_missing_required"
scheduler_pressure_missing_required_output="$(assert_fails "scheduler pressure summary rejects missing required fields" \
  run_schema_validation cartulary.scheduler_pressure_summary.v4 "$scheduler_pressure_missing_required")"
assert_contains "$scheduler_pressure_missing_required_output" "must have required property 'resource_claim_counts'" "scheduler pressure summary missing required"

fixture_tier_proof="$tmp_dir/fixture-tier-proof.json"
write_valid_fixture_tier_proof "$fixture_tier_proof"
assert_passes "fixture tier proof validates exact schema" \
  run_schema_validation cartulary.fixture_tier_proof.v2 "$fixture_tier_proof" >/dev/null

fixture_tier_proof_unknown_key="$tmp_dir/fixture-tier-proof-unknown-key.json"
write_valid_fixture_tier_proof "$fixture_tier_proof_unknown_key"
mutate_json_fixture fixture-tier-proof-unknown-key "$fixture_tier_proof_unknown_key"
fixture_tier_proof_unknown_key_output="$(assert_fails "fixture tier proof rejects unknown keys" \
  run_schema_validation cartulary.fixture_tier_proof.v2 "$fixture_tier_proof_unknown_key")"
assert_contains "$fixture_tier_proof_unknown_key_output" "must NOT have additional properties" "fixture tier proof unknown key"

fixture_tier_proof_missing_reset="$tmp_dir/fixture-tier-proof-missing-reset-surface.json"
write_valid_fixture_tier_proof "$fixture_tier_proof_missing_reset"
mutate_json_fixture fixture-tier-proof-missing-reset-surface "$fixture_tier_proof_missing_reset"
fixture_tier_proof_missing_reset_output="$(assert_fails "fixture tier proof rejects missing reset surface" \
  run_schema_validation cartulary.fixture_tier_proof.v2 "$fixture_tier_proof_missing_reset")"
assert_contains "$fixture_tier_proof_missing_reset_output" "must have required property 'reset_surface'" "fixture tier proof missing reset surface"

release_readiness_evidence="$tmp_dir/release-readiness-evidence.json"
write_valid_release_readiness_evidence "$release_readiness_evidence"
assert_passes "release readiness evidence validates exact schema" \
  run_schema_validation cartulary.release_readiness_evidence.v2 "$release_readiness_evidence" >/dev/null

release_readiness_missing_effect="$tmp_dir/release-readiness-missing-effect.json"
write_valid_release_readiness_evidence "$release_readiness_missing_effect"
mutate_json_fixture release-readiness-missing-effect "$release_readiness_missing_effect"
release_readiness_missing_effect_output="$(assert_fails "release readiness evidence requires semantic effects" \
  run_schema_validation cartulary.release_readiness_evidence.v2 "$release_readiness_missing_effect")"
assert_contains "$release_readiness_missing_effect_output" "must have required property 'conformance_effect'" "release readiness missing effect"

release_readiness_empty_owner_refs="$tmp_dir/release-readiness-empty-owner-refs.json"
write_valid_release_readiness_evidence "$release_readiness_empty_owner_refs"
mutate_json_fixture release-readiness-empty-owner-refs "$release_readiness_empty_owner_refs"
release_readiness_empty_owner_refs_output="$(assert_fails "release readiness evidence rejects empty owner refs" \
  run_schema_validation cartulary.release_readiness_evidence.v2 "$release_readiness_empty_owner_refs")"
assert_contains "$release_readiness_empty_owner_refs_output" "must NOT have fewer than 1 items" "release readiness empty owner refs"

release_readiness_ambiguous_visual="$tmp_dir/release-readiness-ambiguous-visual.json"
write_valid_release_readiness_evidence "$release_readiness_ambiguous_visual"
mutate_json_fixture release-readiness-ambiguous-visual-conformance "$release_readiness_ambiguous_visual"
release_readiness_ambiguous_visual_output="$(assert_fails "release readiness evidence rejects ambiguous visual conformance effect" \
  run_schema_validation cartulary.release_readiness_evidence.v2 "$release_readiness_ambiguous_visual")"
assert_contains "$release_readiness_ambiguous_visual_output" "must be equal to one of the allowed values" "release readiness ambiguous visual conformance"

duplicate_target_owner="$tmp_dir/task_surface_owner_duplicate_target.json"
cp "$ROOT_DIR/tools/task_surface_owner.json" "$duplicate_target_owner"
mutate_json_fixture execution-topology-duplicate-target "$duplicate_target_owner"
duplicate_target_output="$(assert_fails "duplicate target identifiers" run_shape_check task-surface-owner "$duplicate_target_owner")"
assert_contains "$duplicate_target_output" ".targets.name contains duplicate backend-unit" "duplicate target identifiers"

stale_schedule="$tmp_dir/check_schedule_stale.json"
write_valid_check_schedule "$stale_schedule"
mutate_json_fixture scheduler-manifest-stale-schema "$stale_schedule"
stale_schedule_output="$(assert_fails "stale generated schedule shape" run_shape_check scheduler-manifest "$stale_schedule")"
assert_contains "$stale_schedule_output" "must declare schema_id cartulary.scheduler_manifest.v2" "stale generated schedule shape"

unsupported_capacity_profile="$tmp_dir/scheduler_unsupported_capacity_profile.json"
write_valid_service_backed_schedule "$unsupported_capacity_profile"
mutate_json_fixture scheduler-manifest-unsupported-capacity-profile "$unsupported_capacity_profile"
unsupported_capacity_profile_output="$(assert_fails "unsupported scheduler capacity profile" run_shape_check scheduler-manifest "$unsupported_capacity_profile")"
assert_contains "$unsupported_capacity_profile_output" "capacity_profile must be one of" "unsupported scheduler capacity profile"

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

missing_check_make_prerequisite_policy="$tmp_dir/check_schedule_missing_make_prerequisite_policy.json"
write_valid_check_schedule "$missing_check_make_prerequisite_policy"
mutate_json_fixture check-schedule-missing-make-prerequisite-policy "$missing_check_make_prerequisite_policy"
missing_check_make_prerequisite_policy_output="$(assert_fails "missing scheduler make prerequisite policy" run_shape_check scheduler-manifest "$missing_check_make_prerequisite_policy")"
assert_contains "$missing_check_make_prerequisite_policy_output" "make_prerequisite_policy is required for make_target work units" "missing scheduler make prerequisite policy"

invalid_timeout_seconds="$tmp_dir/check_schedule_invalid_timeout_seconds.json"
write_valid_check_schedule "$invalid_timeout_seconds"
mutate_json_fixture check-schedule-invalid-timeout-seconds "$invalid_timeout_seconds"
invalid_timeout_seconds_output="$(assert_fails "invalid scheduler timeout seconds" run_shape_check scheduler-manifest "$invalid_timeout_seconds")"
assert_contains "$invalid_timeout_seconds_output" "timeout_seconds must be >= 1" "invalid scheduler timeout seconds"

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

unknown_auto_policy="$tmp_dir/scheduler_resource_registry_unknown_auto_policy.json"
write_valid_scheduler_resource_registry "$unknown_auto_policy"
mutate_json_fixture scheduler-registry-unknown-auto-policy "$unknown_auto_policy"
unknown_auto_policy_output="$(assert_fails "unknown scheduler auto policy" run_shape_check scheduler-resource-registry "$unknown_auto_policy")"
assert_contains "$unknown_auto_policy_output" "capacity.auto_policy must be one of" "unknown scheduler auto policy"

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

govulncheck_findings="$tmp_dir/govulncheck-findings.json"
write_valid_govulncheck_findings "$govulncheck_findings"
assert_passes "valid Govulncheck findings schema" run_schema_validation cartulary.govulncheck_findings.v1 "$govulncheck_findings" >/dev/null

govulncheck_findings_unknown_key="$tmp_dir/govulncheck-findings-unknown-key.json"
write_valid_govulncheck_findings "$govulncheck_findings_unknown_key"
mutate_json_fixture govulncheck-findings-unknown-key "$govulncheck_findings_unknown_key"
govulncheck_findings_unknown_output="$(assert_fails "Govulncheck findings reject unknown key" run_schema_validation cartulary.govulncheck_findings.v1 "$govulncheck_findings_unknown_key")"
assert_contains "$govulncheck_findings_unknown_output" "must NOT have additional properties" "Govulncheck findings unknown key"

govulncheck_findings_missing_count="$tmp_dir/govulncheck-findings-missing-count.json"
write_valid_govulncheck_findings "$govulncheck_findings_missing_count"
mutate_json_fixture govulncheck-findings-missing-count "$govulncheck_findings_missing_count"
govulncheck_findings_missing_output="$(assert_fails "Govulncheck findings require raw count" run_schema_validation cartulary.govulncheck_findings.v1 "$govulncheck_findings_missing_count")"
assert_contains "$govulncheck_findings_missing_output" "must have required property 'raw_event_count'" "Govulncheck findings missing count"

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

static_analysis_cache_record="$tmp_dir/static-analysis-cache-record.json"
write_valid_cache_record cartulary.cache.static_analysis.v1 static-analysis "$static_analysis_cache_record"
run_schema_validation cartulary.cache.static_analysis.v1 "$static_analysis_cache_record" >/dev/null

same_run_helper_ref="$tmp_dir/same-run-helper-artifact-ref.json"
write_valid_same_run_helper_artifact_ref "$same_run_helper_ref"
run_schema_validation cartulary.same_run_helper_artifact_ref.v2 "$same_run_helper_ref" >/dev/null

same_run_helper_ref_scheduler_reused="$tmp_dir/same-run-helper-artifact-ref-scheduler-reused.json"
write_valid_same_run_helper_artifact_ref "$same_run_helper_ref_scheduler_reused"
mutate_json_fixture same-run-helper-ref-scheduler-reused "$same_run_helper_ref_scheduler_reused"
same_run_helper_ref_scheduler_reused_output="$(assert_fails "same-run helper ref rejects scheduler reused" run_schema_validation cartulary.same_run_helper_artifact_ref.v2 "$same_run_helper_ref_scheduler_reused")"
assert_contains "$same_run_helper_ref_scheduler_reused_output" "must be equal to constant" "same-run helper ref scheduler reused"

same_run_helper_ref_missing_digest="$tmp_dir/same-run-helper-artifact-ref-missing-digest.json"
write_valid_same_run_helper_artifact_ref "$same_run_helper_ref_missing_digest"
mutate_json_fixture same-run-helper-ref-missing-digest "$same_run_helper_ref_missing_digest"
same_run_helper_ref_missing_digest_output="$(assert_fails "same-run helper ref requires artifact digest" run_schema_validation cartulary.same_run_helper_artifact_ref.v2 "$same_run_helper_ref_missing_digest")"
assert_contains "$same_run_helper_ref_missing_digest_output" "must have required property 'sha256'" "same-run helper ref missing digest"

same_run_helper_ref_old_run_scope="$tmp_dir/same-run-helper-artifact-ref-old-run-scope.json"
write_valid_same_run_helper_artifact_ref "$same_run_helper_ref_old_run_scope"
mutate_json_fixture same-run-helper-ref-old-run-scope "$same_run_helper_ref_old_run_scope"
same_run_helper_ref_old_run_scope_output="$(assert_fails "same-run helper ref rejects retained-run scope" run_schema_validation cartulary.same_run_helper_artifact_ref.v2 "$same_run_helper_ref_old_run_scope")"
assert_contains "$same_run_helper_ref_old_run_scope_output" "must be equal to constant" "same-run helper ref retained-run scope"

write_valid_test_support_inventory_fixture() {
  local file="$1"

  mkdir -p \
    "$tmp_dir/internal/platform/contracttest" \
    "$tmp_dir/internal/testutil/fixtures/config" \
    "$tmp_dir/internal/testutil/golden/otel" \
    "$tmp_dir/tools"
  : >"$tmp_dir/internal/platform/contracttest/contracttest.go"
  : >"$tmp_dir/internal/testutil/fixtures/config/valid.toml"
  : >"$tmp_dir/internal/testutil/golden/otel/.gitkeep"
  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.test_support_inventory.v1",
  "go_support_roots": [
    {
      "path": "internal/platform/contracttest",
      "owner": "platform_contracts",
      "posture": "platform_facade",
      "runtime_scan": "included",
      "support_scan": "included",
      "service_starting": false,
      "rationale": "Synthetic generated contract facade."
    },
    {
      "path": "internal/testutil",
      "owner": "shared_test_infrastructure",
      "posture": "shared",
      "runtime_scan": "excluded",
      "support_scan": "included",
      "service_starting": false,
      "rationale": "Synthetic shared test infrastructure."
    },
    {
      "path": "tools",
      "owner": "harness_tooling",
      "posture": "shared",
      "runtime_scan": "included",
      "support_scan": "included",
      "service_starting": false,
      "rationale": "Synthetic repo-local harness tooling root."
    }
  ],
  "shared_data_roots": [
    {
      "path": "internal/testutil/fixtures/config",
      "owner": "platform_config",
      "posture": "shared",
      "data_kind": "platform_config",
      "file_roles": ["fixture"],
      "owner_semantic_data_policy": "reject_unclassified",
      "retained_path_policy": "stable",
      "rationale": "Synthetic config fixture root."
    },
    {
      "path": "internal/testutil/golden/otel",
      "owner": "platform_otel",
      "posture": "shared",
      "data_kind": "otel_evidence",
      "file_roles": ["fixture", "golden", "manifest", "placeholder"],
      "owner_semantic_data_policy": "adopted_external_evidence",
      "retained_path_policy": "stable",
      "rationale": "Synthetic adopted OTel root."
    }
  ]
}
JSON
}

test_support_inventory="$tmp_dir/test-support-inventory.json"
write_valid_test_support_inventory_fixture "$test_support_inventory"
assert_contains "$(assert_passes "valid test support inventory" run_shape_check test-support-inventory "$test_support_inventory")" \
  "json shape check passed" \
  "valid test support inventory"

mkdir -p "$tmp_dir/internal/testutil/fixtures/product-records"
: >"$tmp_dir/internal/testutil/fixtures/product-records/case.json"
test_support_inventory_unknown_output="$(assert_fails "test support inventory rejects unclassified shared fixture root" run_shape_check test-support-inventory "$test_support_inventory")"
assert_contains "$test_support_inventory_unknown_output" "must classify internal/testutil/fixtures/product-records/case.json exactly once" "test support inventory unknown shared fixture"

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
