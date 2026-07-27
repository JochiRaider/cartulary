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
    browser-e2e-a11y; do
    write_target_summary "$run_root" "$target"
  done
}

write_required_owner_shards() {
  local run_root="$1"
  RELEASE_FIXTURE_ROOT="$run_root" "$NODE_BIN" --input-type=module - <<'JS'
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import {
  accountingRowsForTarget,
  buildTestEvidenceAccounting,
  buildTestOwnerSummary,
} from "./tools/harness/evidence-accounting/index.mjs";
import { loadTestCatalog } from "./tools/harness/test-catalog/index.mjs";

const root = process.cwd();
const runRoot = path.resolve(process.env.RELEASE_FIXTURE_ROOT);
const targets = new Set(["browser-e2e-a11y", "browser-e2e-support", "browser-e2e-visual"]);
const catalog = loadTestCatalog(root);
const ownerTargets = new Map();
for (const row of catalog.rows.filter((entry) => entry.runner === "playwright")) {
  const stageTarget = {
    accessibility: "browser-e2e-a11y",
    support: "browser-e2e-support",
    visual: "browser-e2e-visual",
  }[row.selector.stage];
  if (!targets.has(stageTarget)) continue;
  ownerTargets.set(`${stageTarget}\0${row.owner_id}`, { target: stageTarget, owner: row.owner_id });
}
for (const { target, owner } of [...ownerTargets.values()].sort((left, right) =>
  `${left.target}\0${left.owner}`.localeCompare(`${right.target}\0${right.owner}`),
)) {
  const selection = accountingRowsForTarget(root, { ownerID: owner, targetName: target });
  const plan = {
    evidence_epoch: selection.evidence_epoch,
    command_id: "cartulary.harness.command.release_readiness_fixture.v1",
    run_id: path.basename(runRoot),
    owner_id: owner,
    selected_rows: selection.selected_rows,
    source_snapshot_digest: `sha256:${"0".repeat(64)}`,
    test_catalog_digest: selection.test_catalog_digest,
    verification_routing_digest: selection.verification_routing_digest,
    runtime_profile_digest: selection.runtime_profile_digest,
    resource_profile_digest: selection.resource_profile_digest,
    fixture_profile_digest: selection.fixture_profile_digest,
    target,
    rows: selection.expected_rows,
    work_units: selection.selected_rows.map((rowID) => ({
      work_unit_id: `fixture:${rowID}`,
      row_ids: [rowID],
    })),
    selection: { completion_scope: "target_partition" },
    unused_inputs: [],
  };
  const execution = {
    status: "pass",
    duration_ms: selection.selected_rows.length,
    row_results: selection.selected_rows.map((rowID) => ({
      row_id: rowID,
      terminal_state: "passed",
      duration_ms: 1,
      exit_code: 0,
      attempt: 1,
    })),
  };
  const timestamp = "2026-01-01T00:00:00.000Z";
  const accounting = buildTestEvidenceAccounting(plan, execution, [], timestamp, timestamp);
  const prefix = `${target}/owners/${owner}`;
  const artifacts = {
    evidence_accounting: `${prefix}/test-evidence-accounting.json`,
    owner_summary: `${prefix}/test-owner-summary.json`,
    tool_run_summary: `${target}/tool-run-summary.json`,
  };
  const summary = buildTestOwnerSummary(plan, accounting, artifacts);
  const directory = path.join(runRoot, target, "owners", owner);
  mkdirSync(directory, { recursive: true });
  writeFileSync(path.join(directory, "test-evidence-accounting.json"), `${JSON.stringify(accounting, null, 2)}\n`);
  writeFileSync(path.join(directory, "test-owner-summary.json"), `${JSON.stringify(summary, null, 2)}\n`);
}
JS
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
write_required_owner_shards "$pass_run_root"
run_release_readiness "$pass_results" "$pass_run_id" >/dev/null
pass_artifact="$pass_run_root/release-readiness-evidence/release-readiness-evidence.json"
"$NODE_BIN" "$ROOT_DIR/tools/harness/contract/harness-contract-cli.mjs" validate-schema cartulary.release_readiness_evidence.v2 "$pass_artifact" >/dev/null
assert_equals "$(json_field "$pass_artifact" 'value.status')" "pass" "passing release readiness status"
assert_equals "$(json_field "$pass_artifact" 'value.evidence_records.some((record) => record.evidence_id.startsWith("owner-partition:browser-e2e-visual:") && record.status === "passed")')" "true" "visual owner partitions close from accounting"
assert_equals "$(json_field "$pass_artifact" 'value.evidence_records.some((record) => record.claim_publication_effect === "claim_publication_evidence")')" "false" "no release record is claim publication evidence"

missing_owner_results="$tmp_dir/missing-owner-results"
missing_owner_run_id="missing-owner-run"
missing_owner_run_root="$missing_owner_results/$missing_owner_run_id"
write_required_target_summaries "$missing_owner_run_root"
write_required_owner_shards "$missing_owner_run_root"
missing_owner_dir="$(find "$missing_owner_run_root/browser-e2e-visual/owners" -mindepth 1 -maxdepth 1 -type d | sort | head -n 1)"
rm -rf "$missing_owner_dir"
set +e
missing_owner_output="$(run_release_readiness "$missing_owner_results" "$missing_owner_run_id" 2>&1)"
missing_owner_status=$?
set -e
if [[ "$missing_owner_status" -eq 0 ]]; then
  fail "missing owner accounting partition must fail"
fi
assert_contains "$missing_owner_output" "owner-partition:browser-e2e-visual:" "missing owner partition failure output"

missing_results="$tmp_dir/missing-results"
missing_run_id="missing-run"
missing_run_root="$missing_results/$missing_run_id"
write_required_target_summaries "$missing_run_root"
write_required_owner_shards "$missing_run_root"
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
assert_equals "$(json_field "$missing_artifact" 'value.evidence_records.find((record) => record.evidence_id === "target:browser-e2e-a11y").status')" "missing" "missing target summary status"
assert_equals "$(json_field "$missing_artifact" 'value.evidence_records.some((record) => record.evidence_id.startsWith("owner-partition:browser-e2e-a11y:") && record.status === "missing")')" "true" "missing target owner partitions are explicit"

echo "release readiness evidence tests passed"
