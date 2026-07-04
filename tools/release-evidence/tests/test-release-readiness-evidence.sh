#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
if command -v "${NODE_BIN}" >/dev/null 2>&1; then
  NODE_BIN="$(command -v "${NODE_BIN}")"
fi

# shellcheck source=tools/harness/test-support/harness-scratch.sh
source "$ROOT_DIR/tools/harness/test-support/harness-scratch.sh"

cleanup_paths=()

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -rf "${path}"
  done
}

trap cleanup EXIT

fail() {
  echo "$*" >&2
  exit 1
}

assert_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  if [[ "$actual" != "$expected" ]]; then
    fail "$label: expected [$expected], got [$actual]"
  fi
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label: expected output to contain [$needle]"
  fi
}

json_field() {
  local file="$1"
  local expression="$2"

  "$NODE_BIN" - "$file" "$expression" <<'JS'
const fs = require("node:fs");
const [file, expression] = process.argv.slice(2);
const value = JSON.parse(fs.readFileSync(file, "utf8"));
const result = Function("value", `return (${expression});`)(value);
if (Array.isArray(result)) {
  process.stdout.write(`${result.join("\n")}\n`);
} else if (result && typeof result === "object") {
  process.stdout.write(`${JSON.stringify(result)}\n`);
} else {
  process.stdout.write(`${String(result)}\n`);
}
JS
}

write_target_summary() {
  local run_root="$1"
  local target="$2"

  mkdir -p "$run_root/$target"
  cat >"$run_root/$target/target-summary.json" <<JSON
{
  "schema_id": "cartulary.test_target_summary.v4",
  "target": "$target",
  "status": "pass",
  "artifacts": {
    "dir": "$run_root/$target"
  },
  "own": {
    "artifacts": {
      "dir": "$run_root/$target"
    }
  },
  "totals": {
    "artifacts": {
      "dir": "$run_root/$target"
    }
  }
}
JSON
}

write_required_target_summaries() {
  local run_root="$1"
  local target

  for target in \
    check \
    harness-contract \
    go-gosec-audit \
    license-report \
    sbom \
    seaweedfs-release-gate \
    build-web \
    build-server \
    build-migrate \
    build-operator \
    deployable-shape \
    browser-e2e-support \
    browser-e2e-visual \
    browser-e2e-a11y \
    browser-e2e-a11y-preflight; do
    write_target_summary "$run_root" "$target"
  done
}

write_valid_frontend_row_accounting() {
  local file="$1"

  mkdir -p "$(dirname "$file")"
  cat >"$file" <<'JSON'
{
  "schema_id": "cartulary.frontend_row_accounting.v3",
  "target_name": "browser-e2e-visual",
  "command_id": "cartulary.harness.command.browser_e2e_visual.v1",
  "phase_namespace": "frontend",
  "accounting_scope": {
    "mode": "active_target",
    "invocation_kind": "standalone_target",
    "phase_namespace": "frontend",
    "phase": "",
    "selection_policy": "all_active_rows_for_target",
    "selected_row_ids": []
  },
  "registry_ref": "tools/frontend_phase_registry.json",
  "registry_digest": "0000000000000000000000000000000000000000000000000000000000000000",
  "guide_ref": "docs/guides/cartulary_frontend_implementation_testing_guide.md",
  "guide_digest": "1111111111111111111111111111111111111111111111111111111111111111",
  "phase_map_refs": [
    "tools/frontend_phase_maps/fe_p8_test_map.json"
  ],
  "phase_map_digests": [
    "2222222222222222222222222222222222222222222222222222222222222222"
  ],
  "run_root": ".cartulary/test-results/run",
  "target_status": "pass",
  "scenario_results": [
    {
      "scenario_title": "FE-V-P8-01 captures schema-driven Timeline workbook shell controls",
      "status": "passed",
      "row_ids": [
        "FE-V-P8-01"
      ],
      "artifact_refs": [
        "apps/web/e2e/visual.spec.ts"
      ]
    }
  ],
  "row_results": [
    {
      "row_id": "FE-V-P8-01",
      "phase_id": "FE-P8",
      "evidence_class": "design_direction",
      "claim_status_at_run": "implemented",
      "target_mapping_status": "mapped",
      "closure_status": "closed",
      "closing_scenario_titles": [
        "FE-V-P8-01 captures schema-driven Timeline workbook shell controls"
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
  "target": "browser-e2e-visual",
  "rows": [
    {
      "phase_id": "FE-P8",
      "phase_status": "active",
      "row_rollup_state": "active_green",
      "row_id": "FE-V-P8-01",
      "layer": "visual",
      "evidence_class": "design_direction",
      "claim_status": "implemented",
      "claim": {
        "statement": "visual readiness fixture",
        "claim_publication_intent": "none",
        "closure_scope": "scenario"
      },
      "blockers": [],
      "required_for_closure": true,
      "scenario_titles": [
        "FE-V-P8-01 captures schema-driven Timeline workbook shell controls"
      ],
      "target": "browser-e2e-visual",
      "target_status": "pass",
      "scenarios": [
        {
          "title": "FE-V-P8-01 captures schema-driven Timeline workbook shell controls",
          "status": "passed",
          "files": [
            "apps/web/e2e/visual.spec.ts"
          ]
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

run_release_readiness() {
  local results_root="$1"
  local run_id="$2"

  CARTULARY_TEST_RESULTS_DIR="$results_root" \
  CARTULARY_TEST_RUN_ID="$run_id" \
    "$NODE_BIN" "$ROOT_DIR/tools/release-evidence/release-readiness-evidence.mjs"
}

tmp_dir="$(cartulary_harness_mktemp_dir "release-readiness-evidence.XXXXXX")"
cleanup_paths+=("$tmp_dir")

pass_results="$tmp_dir/pass-results"
pass_run_id="pass-run"
pass_run_root="$pass_results/$pass_run_id"
write_required_target_summaries "$pass_run_root"
write_valid_frontend_row_accounting "$pass_run_root/browser-e2e-visual/frontend-row-accounting.json"
run_release_readiness "$pass_results" "$pass_run_id" >/dev/null
pass_artifact="$pass_run_root/release-readiness-evidence/release-readiness-evidence.json"
"$NODE_BIN" "$ROOT_DIR/tools/harness/contract/harness-contract-cli.mjs" validate-schema cartulary.release_readiness_evidence.v1 "$pass_artifact" >/dev/null
assert_equals "$(json_field "$pass_artifact" 'value.status')" "pass" "passing release readiness status"
assert_equals "$(json_field "$pass_artifact" 'value.evidence_records.find((record) => record.evidence_id === "frontend-row:FE-V-P8-01:browser-e2e-visual").conformance_effect')" "no_product_conformance" "visual row conformance effect"
assert_equals "$(json_field "$pass_artifact" 'value.evidence_records.some((record) => record.claim_publication_effect === "claim_publication_evidence")')" "false" "no release record is claim publication evidence"

legacy_results="$tmp_dir/legacy-results"
legacy_run_id="legacy-run"
legacy_run_root="$legacy_results/$legacy_run_id"
write_required_target_summaries "$legacy_run_root"
mkdir -p "$legacy_run_root/browser-e2e-visual"
printf '%s\n' '{"schema_id":"cartulary.frontend_row_accounting.v2"}' >"$legacy_run_root/browser-e2e-visual/frontend-row-accounting.json"
set +e
legacy_output="$(run_release_readiness "$legacy_results" "$legacy_run_id" 2>&1)"
legacy_status=$?
set -e
if [[ "$legacy_status" -eq 0 ]]; then
  fail "legacy row-accounting run must fail"
fi
legacy_artifact="$legacy_run_root/release-readiness-evidence/release-readiness-evidence.json"
"$NODE_BIN" "$ROOT_DIR/tools/harness/contract/harness-contract-cli.mjs" validate-schema cartulary.release_readiness_evidence.v1 "$legacy_artifact" >/dev/null
assert_contains "$legacy_output" "frontend-row-accounting:browser-e2e-visual:schema" "legacy row accounting failure output"
assert_equals "$(json_field "$legacy_artifact" 'value.evidence_records.find((record) => record.evidence_id === "frontend-row-accounting:browser-e2e-visual:schema").schema_id')" "cartulary.frontend_row_accounting.v2" "legacy row accounting schema captured"
assert_equals "$(json_field "$legacy_artifact" 'value.status')" "fail" "legacy row accounting fails release readiness"

missing_results="$tmp_dir/missing-results"
missing_run_id="missing-run"
missing_run_root="$missing_results/$missing_run_id"
write_required_target_summaries "$missing_run_root"
rm -rf "$missing_run_root/browser-e2e-a11y"
set +e
missing_output="$(run_release_readiness "$missing_results" "$missing_run_id" 2>&1)"
missing_status=$?
set -e
if [[ "$missing_status" -eq 0 ]]; then
  fail "missing target summary run must fail"
fi
missing_artifact="$missing_run_root/release-readiness-evidence/release-readiness-evidence.json"
assert_contains "$missing_output" "target:browser-e2e-a11y status=missing" "missing target failure output"
assert_equals "$(json_field "$missing_artifact" 'value.rollup.required_failed')" "1" "missing target required failure count"

echo "release readiness evidence tests passed"
