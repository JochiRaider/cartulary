#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
TMP_DIR="$(mktemp -d "$ROOT_DIR/tmp/frontend-evidence-audit-test.XXXXXX")"

cleanup() {
  rm -rf "$TMP_DIR"
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
    fail "$label: expected [$needle] in [$haystack]"
  fi
}

write_fixture_roots() {
  local output_dir="$1"
  "$NODE_BIN" --input-type=module - "$ROOT_DIR" "$output_dir" <<'JS'
import { createHash } from "node:crypto";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";

const [repoRoot, outputDir] = process.argv.slice(2);
const phaseID = "FE-P8";
const registryRef = "tools/frontend_phase_registry.json";
const guideRef = "docs/guides/cartulary_frontend_implementation_testing_guide.md";
const mapRef = "tools/frontend_phase_maps/fe_p8_test_map.json";

function readJSON(relativePath) {
  return JSON.parse(readFileSync(path.join(repoRoot, relativePath), "utf8"));
}

function digest(relativePath) {
  return createHash("sha256")
    .update(readFileSync(path.join(repoRoot, relativePath)))
    .digest("hex");
}

function commandID(targetName) {
  return `cartulary.harness.command.${targetName.replaceAll("-", "_")}.v1`;
}

function writeJSON(file, value) {
  mkdirSync(path.dirname(file), { recursive: true });
  writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
}

function targetRootName(targetName) {
  switch (targetName) {
    case "frontend-unit":
    case "browser-e2e-webserver-backed":
    case "browser-e2e-stateful":
      return "check";
    case "browser-e2e-support":
      return "support";
    case "browser-e2e-visual":
      return "visual";
    case "browser-e2e-a11y":
      return "a11y";
    default:
      throw new Error(`unexpected FE-P8 target ${targetName}`);
  }
}

const registry = readJSON(registryRef);
const map = readJSON(mapRef);
const phase = registry.phases.find((entry) => entry.phase_id === phaseID);
const digests = {
  registry: digest(registryRef),
  guide: digest(guideRef),
  map: digest(mapRef),
};
const rowsByTarget = new Map();
for (const row of map.rows.filter((entry) => entry.claim_status === "implemented")) {
  for (const target of row.targets.filter((entry) => entry.required_for_closure)) {
    const rows = rowsByTarget.get(target.target_name) ?? [];
    rows.push(row);
    rowsByTarget.set(target.target_name, rows);
  }
}

const inputRoots = {};
for (const [targetName, rows] of rowsByTarget.entries()) {
  const rootName = targetRootName(targetName);
  const root = path.join(outputDir, rootName);
  const targetDir = path.join(root, targetName);
  inputRoots[rootName] = root;
  writeJSON(path.join(targetDir, "tool-run-summary.json"), {
    schema_id: "cartulary.tool_run_summary.v3",
    target: targetName,
    status: "pass",
  });
  writeJSON(path.join(targetDir, "target-summary.json"), {
    schema_id: "cartulary.test_target_summary.v4",
    target: targetName,
    status: "pass",
  });
  const scenarioResults = rows.flatMap((row) =>
    row.scenario_titles.map((title) => ({
      scenario_title: title,
      status: "passed",
      row_ids: [row.id],
      artifact_refs: [`fixtures/${targetName}.json`],
    })),
  );
  const rowResults = rows.map((row) => ({
    row_id: row.id,
    phase_id: row.phase_id,
    evidence_class: row.evidence_class,
    claim_status_at_run: row.claim_status,
    target_mapping_status: "mapped",
    closure_status: "closed",
    closing_scenario_titles: row.scenario_titles,
    failure_reason: "",
  }));
  const compatRows = rows.map((row) => ({
    phase_id: row.phase_id,
    phase_status: phase.status,
    row_rollup_state: phase.row_rollup_state,
    row_id: row.id,
    layer: row.layer,
    evidence_class: row.evidence_class,
    claim_status: row.claim_status,
    claim: row.claim,
    blockers: row.blockers,
    required_for_closure: true,
    scenario_titles: row.scenario_titles,
    target: targetName,
    target_status: "pass",
    scenarios: row.scenario_titles.map((title) => ({
      title,
      status: "passed",
      files: [`fixtures/${targetName}.json`],
    })),
    closure_status: "closed",
  }));
  writeJSON(path.join(targetDir, "frontend-row-accounting.json"), {
    schema_id: "cartulary.frontend_row_accounting.v3",
    target_name: targetName,
    command_id: commandID(targetName),
    phase_namespace: "frontend",
    accounting_scope: {
      mode: "selected_rows",
      invocation_kind: "frontend_phase_slice",
      phase_namespace: "frontend",
      phase: phaseID,
      selection_policy: "frontend_rows_through_selected_phase",
      selected_row_ids: rows.map((row) => row.id),
    },
    registry_ref: registryRef,
    registry_digest: digests.registry,
    guide_ref: guideRef,
    guide_digest: digests.guide,
    phase_map_refs: [mapRef],
    phase_map_digests: [digests.map],
    run_root: path.relative(repoRoot, targetDir).replaceAll("\\", "/"),
    target_status: "pass",
    scenario_results: scenarioResults,
    row_results: rowResults,
    rollup: {
      implemented: rows.length,
      blocked: 0,
      missing: 0,
      stale: 0,
      not_applicable: 0,
      closed: rows.length,
      failed: 0,
    },
    target: targetName,
    rows: compatRows,
    counts: {
      rows: rows.length,
      scenarios: scenarioResults.length,
      closed_rows: rows.length,
      blocked_by_target_rows: 0,
      failed_rows: 0,
      missing_rows: 0,
      not_evaluable_rows: 0,
      passed_scenarios: scenarioResults.length,
      failed_scenarios: 0,
      missing_scenarios: 0,
      skipped_scenarios: 0,
      unknown_scenarios: 0,
    },
  });
}

writeJSON(path.join(outputDir, "inputs.json"), {
  check: inputRoots.check,
  support: inputRoots.support,
  visual: inputRoots.visual,
  a11y: inputRoots.a11y,
});
JS
}

json_field() {
  local file="$1"
  local expr="$2"
  "$NODE_BIN" - "$file" "$expr" <<'JS'
const fs = require("node:fs");
const [file, expr] = process.argv.slice(2);
const value = JSON.parse(fs.readFileSync(file, "utf8"));
const result = Function("value", `return ${expr}`)(value);
if (Array.isArray(result)) {
  process.stdout.write(result.join("\n"));
} else if (result === null || result === undefined) {
  process.stdout.write("");
} else {
  process.stdout.write(String(result));
}
JS
}

inputs_file="$TMP_DIR/fixtures/inputs.json"
write_fixture_roots "$TMP_DIR/fixtures"
check_root="$(json_field "$inputs_file" "value.check")"
support_root="$(json_field "$inputs_file" "value.support")"
visual_root="$(json_field "$inputs_file" "value.visual")"
a11y_root="$(json_field "$inputs_file" "value.a11y")"

summary_dir="$TMP_DIR/pass-summary"
PHASE_NAMESPACE=frontend \
PHASE=FE-P8 \
CHECK_RESULTS_DIR="$check_root" \
BROWSER_SUPPORT_RESULTS_DIR="$support_root" \
BROWSER_VISUAL_RESULTS_DIR="$visual_root" \
BROWSER_A11Y_RESULTS_DIR="$a11y_root" \
CARTULARY_PHASE_ARTIFACT_DIR="$summary_dir" \
  "$NODE_BIN" "$ROOT_DIR/tools/harness/frontend/frontend-evidence-audit-cli.mjs"
pass_summary="$summary_dir/frontend-evidence-audit-summary.json"
if [[ ! -f "$pass_summary" ]]; then
  fail "passing audit did not write summary"
fi
if [[ "$(json_field "$pass_summary" "value.status")" != "pass" ]]; then
  fail "passing audit summary did not record pass"
fi

missing_summary_dir="$TMP_DIR/missing-summary"
set +e
PHASE_NAMESPACE=frontend \
PHASE=FE-P8 \
CHECK_RESULTS_DIR="$check_root" \
BROWSER_SUPPORT_RESULTS_DIR="$support_root" \
BROWSER_VISUAL_RESULTS_DIR="" \
BROWSER_A11Y_RESULTS_DIR="$a11y_root" \
CARTULARY_PHASE_ARTIFACT_DIR="$missing_summary_dir" \
  "$NODE_BIN" "$ROOT_DIR/tools/harness/frontend/frontend-evidence-audit-cli.mjs" \
  >"$TMP_DIR/missing.stdout" 2>"$TMP_DIR/missing.stderr"
status=$?
set -e
if [[ "$status" -eq 0 ]]; then
  fail "missing visual root audit unexpectedly passed"
fi
assert_contains "$(cat "$TMP_DIR/missing.stderr")" "BROWSER_VISUAL_RESULTS_DIR is required" "missing explicit root failure"

stale_summary_dir="$TMP_DIR/stale-summary"
"$NODE_BIN" --input-type=module - "$visual_root/browser-e2e-visual/frontend-row-accounting.json" <<'JS'
import { readFileSync, writeFileSync } from "node:fs";
const [file] = process.argv.slice(2);
const value = JSON.parse(readFileSync(file, "utf8"));
value.guide_digest = "0".repeat(64);
writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
JS
set +e
PHASE_NAMESPACE=frontend \
PHASE=FE-P8 \
CHECK_RESULTS_DIR="$check_root" \
BROWSER_SUPPORT_RESULTS_DIR="$support_root" \
BROWSER_VISUAL_RESULTS_DIR="$visual_root" \
BROWSER_A11Y_RESULTS_DIR="$a11y_root" \
CARTULARY_PHASE_ARTIFACT_DIR="$stale_summary_dir" \
  "$NODE_BIN" "$ROOT_DIR/tools/harness/frontend/frontend-evidence-audit-cli.mjs" \
  >"$TMP_DIR/stale.stdout" 2>"$TMP_DIR/stale.stderr"
status=$?
set -e
if [[ "$status" -eq 0 ]]; then
  fail "stale visual root audit unexpectedly passed"
fi
assert_contains "$(cat "$TMP_DIR/stale.stderr")" "guide_digest is stale" "stale digest failure"

echo "frontend evidence audit fixture tests passed"
