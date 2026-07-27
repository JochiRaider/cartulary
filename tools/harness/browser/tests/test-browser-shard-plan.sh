#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
NODE_BIN="${NODE:-node}"
PLANNER="$ROOT_DIR/tools/harness/browser/browser-shard-plan.mjs"
tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/browser-shard-plan.XXXXXX")"

cleanup() {
  rm -rf "$tmp_dir"
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

baseline="$tmp_dir/baseline.json"
cat >"$baseline" <<'JSON'
{
  "schema_id": "cartulary.browser_e2e_duration_baselines.v3",
  "default_entry_weight_ms": 7000,
  "file_overhead_ms": 500,
  "shard_target_ms": 8000,
  "entries": {
    "module.fixture.browser.alpha_one": {
      "file": "apps/web/e2e/alpha.spec.ts",
      "title": "alpha one",
      "weight_ms": 30000
    },
    "module.fixture.browser.alpha_two": {
      "file": "apps/web/e2e/alpha.spec.ts",
      "title": "alpha two",
      "weight_ms": 20000
    },
    "module.fixture.browser.beta": {
      "file": "apps/web/e2e/beta.spec.ts",
      "title": "beta primary",
      "weight_ms": 5000
    }
  }
}
JSON

BASELINE_FILE="$baseline" "$NODE_BIN" --input-type=module - <<'JS' >"$tmp_dir/plan.json"
import { createPlanFromEntries } from "./tools/harness/browser/browser-shard-plan.mjs";

const entries = [
  {
    id: "module.fixture.browser.alpha_one",
    stage: "webserver_backed",
    file: "apps/web/e2e/alpha.spec.ts",
    title: "alpha one",
    titles: ["alpha one"],
    runtime_profile_id: "default",
  },
  {
    id: "module.fixture.browser.alpha_two",
    stage: "webserver_backed",
    file: "apps/web/e2e/alpha.spec.ts",
    title: "alpha two",
    titles: ["alpha two"],
    runtime_profile_id: "default",
  },
  {
    id: "module.fixture.browser.beta",
    stage: "webserver_backed",
    file: "apps/web/e2e/beta.spec.ts",
    title: "beta primary",
    titles: ["beta primary", "beta secondary"],
    runtime_profile_id: "default",
  },
];
const plan = createPlanFromEntries({
  baselineFile: process.env.BASELINE_FILE,
  minShards: 1,
  maxShards: 3,
  baselineEntries: entries,
  selectedEntries: entries,
});
process.stdout.write(`${JSON.stringify(plan, null, 2)}\n`);
JS

"$NODE_BIN" - "$tmp_dir/plan.json" <<'JS'
const fs = require("node:fs");
const plan = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
if (plan.entry_count !== 3 || plan.shard_count !== 3) {
  throw new Error(`unexpected semantic shard plan shape ${plan.entry_count}/${plan.shard_count}`);
}
if (!plan.entries.every((entry) => entry.id.startsWith("module.fixture.browser."))) {
  throw new Error("semantic shard plan retained a legacy row identity");
}
if (!plan.shards.some((shard) => shard.grep.includes("beta secondary"))) {
  throw new Error("semantic shard plan lost an exact multi-title selector");
}
JS

legacy_baseline="$tmp_dir/legacy-baseline.json"
cat >"$legacy_baseline" <<'JSON'
{
  "schema_id": "cartulary.browser_e2e_duration_baselines.v3",
  "default_entry_weight_ms": 7000,
  "file_overhead_ms": 500,
  "shard_target_ms": 8000,
  "entries": {
    "integration.entity-linking.row-01": {
      "file": "apps/web/e2e/alpha.spec.ts",
      "title": "legacy",
      "weight_ms": 1
    }
  }
}
JSON
set +e
legacy_output="$({ BASELINE_FILE="$legacy_baseline" "$NODE_BIN" --input-type=module - <<'JS'; } 2>&1
import { createPlanFromEntries } from "./tools/harness/browser/browser-shard-plan.mjs";
const entries = [{
  id: "module.fixture.browser.alpha_one",
  stage: "webserver_backed",
  file: "apps/web/e2e/alpha.spec.ts",
  title: "alpha one",
  titles: ["alpha one"],
  runtime_profile_id: "default",
}];
createPlanFromEntries({
  baselineFile: process.env.BASELINE_FILE,
  minShards: 1,
  maxShards: 1,
  baselineEntries: entries,
  selectedEntries: entries,
});
JS
)"
legacy_status=$?
set -e
if [[ "$legacy_status" -eq 0 ]]; then
  fail "legacy browser baseline identity unexpectedly passed"
fi
assert_contains "$legacy_output" "must be a semantic catalog row ID" "legacy browser baseline rejection"

results_dir="$tmp_dir/results"
refresh_baseline="$tmp_dir/refresh-baseline.json"
cat >"$refresh_baseline" <<'JSON'
{
  "schema_id": "cartulary.browser_e2e_duration_baselines.v3",
  "default_entry_weight_ms": 10000,
  "file_overhead_ms": 2500,
  "shard_target_ms": 12000,
  "entries": {}
}
JSON
mkdir -p "$results_dir/check"
cat >"$results_dir/check/tool-run-summary.json" <<'JSON'
{
  "schema_id": "cartulary.tool_run_summary.v5",
  "target": "check",
  "status": "pass"
}
JSON
RESULTS_ROOT="$results_dir" "$NODE_BIN" --input-type=module - <<'JS'
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import {
  accountingRowsForTarget,
  buildTestEvidenceAccounting,
} from "./tools/harness/evidence-accounting/index.mjs";
import { buildSourceSnapshot } from "./tools/harness/owner-slice/source-snapshot.mjs";
import { loadTestCatalog } from "./tools/harness/test-catalog/index.mjs";

const root = process.cwd();
const resultsRoot = path.resolve(process.env.RESULTS_ROOT);
const row = loadTestCatalog(root).rows.find(
  (entry) => entry.runner === "playwright" && entry.selector.stage === "webserver_backed",
);
const selection = accountingRowsForTarget(root, {
  ownerID: row.owner_id,
  targetName: "browser-e2e-webserver-backed",
});
const plan = {
  evidence_epoch: selection.evidence_epoch,
  command_id: "cartulary.harness.command.browser_duration_fixture.v1",
  run_id: "compatible",
  owner_id: selection.owner_id,
  selected_rows: selection.selected_rows,
  source_snapshot_digest: buildSourceSnapshot(root).digest,
  test_catalog_digest: selection.test_catalog_digest,
  verification_routing_digest: selection.verification_routing_digest,
  runtime_profile_digest: selection.runtime_profile_digest,
  resource_profile_digest: selection.resource_profile_digest,
  fixture_profile_digest: selection.fixture_profile_digest,
  target: "browser-e2e-webserver-backed",
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
  duration_ms: selection.selected_rows.length * 25,
  row_results: selection.selected_rows.map((rowID) => ({
    row_id: rowID,
    terminal_state: "passed",
    duration_ms: 25,
    exit_code: 0,
    attempt: 1,
  })),
};
const timestamp = "2026-01-01T00:00:00.000Z";
const accounting = buildTestEvidenceAccounting(plan, execution, [], timestamp, timestamp);
const directory = path.join(
  resultsRoot,
  "browser-e2e-webserver-backed",
  "owners",
  selection.owner_id,
);
mkdirSync(directory, { recursive: true });
writeFileSync(
  path.join(directory, "test-evidence-accounting.json"),
  `${JSON.stringify(accounting, null, 2)}\n`,
);
JS

"$NODE_BIN" "$PLANNER" update-baselines --baseline-file "$refresh_baseline" "$results_dir" >/dev/null
"$NODE_BIN" - "$refresh_baseline" <<'JS'
const fs = require("node:fs");
const baseline = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const ids = Object.keys(baseline.entries);
if (ids.length === 0 || ids.some((id) => id.startsWith("E-") || id.startsWith("FE-"))) {
  throw new Error(`browser refresh did not emit semantic catalog identities: ${ids.join(",")}`);
}
if (!Object.values(baseline.entries).every((entry) => entry.weight_ms === 25)) {
  throw new Error("browser refresh did not consume owner-accounting durations");
}
JS

retained_source_digest="sha256:1111111111111111111111111111111111111111111111111111111111111111"
RESULTS_ROOT="$results_dir" RETAINED_SOURCE_DIGEST="$retained_source_digest" "$NODE_BIN" --input-type=module - <<'JS'
import { readdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";

const root = path.resolve(process.env.RESULTS_ROOT);
const stack = [root];
let accountingFile = "";
while (stack.length > 0 && !accountingFile) {
  const current = stack.pop();
  for (const entry of readdirSync(current, { withFileTypes: true })) {
    const next = path.join(current, entry.name);
    if (entry.isDirectory()) stack.push(next);
    else if (entry.isFile() && entry.name === "test-evidence-accounting.json") {
      accountingFile = next;
      break;
    }
  }
}
if (!accountingFile) throw new Error("browser accounting fixture is unavailable");
const accounting = JSON.parse(readFileSync(accountingFile, "utf8"));
accounting.source_snapshot_digest = process.env.RETAINED_SOURCE_DIGEST;
writeFileSync(accountingFile, `${JSON.stringify(accounting, null, 2)}\n`);
JS

set +e
stale_output="$($NODE_BIN "$PLANNER" check-baseline-drift --baseline-file "$refresh_baseline" "$results_dir" 2>&1)"
stale_status=$?
set -e
if [[ "$stale_status" -eq 0 ]]; then
  fail "standalone browser drift unexpectedly accepted a stale source identity"
fi
assert_contains "$stale_output" "no observed Playwright timings" "standalone stale source rejection"

CARTULARY_RETAINED_RESULTS_DIR="$results_dir" \
CARTULARY_RETAINED_SOURCE_SNAPSHOT_DIGEST="$retained_source_digest" \
  "$NODE_BIN" "$PLANNER" check-baseline-drift --baseline-file "$refresh_baseline" "$results_dir" >/dev/null

set +e
mismatched_root_output="$(
  CARTULARY_RETAINED_RESULTS_DIR="$tmp_dir/not-the-retained-root" \
  CARTULARY_RETAINED_SOURCE_SNAPSHOT_DIGEST="$retained_source_digest" \
    "$NODE_BIN" "$PLANNER" check-baseline-drift --baseline-file "$refresh_baseline" "$results_dir" 2>&1
)"
mismatched_root_status=$?
set -e
if [[ "$mismatched_root_status" -eq 0 ]]; then
  fail "retained source identity unexpectedly accepted a different results root"
fi
assert_contains "$mismatched_root_output" "requires the matching agent-finalize results root" "retained root binding"

set +e
malformed_identity_output="$(
  CARTULARY_RETAINED_RESULTS_DIR="$results_dir" \
  CARTULARY_RETAINED_SOURCE_SNAPSHOT_DIGEST="not-a-digest" \
    "$NODE_BIN" "$PLANNER" check-baseline-drift --baseline-file "$refresh_baseline" "$results_dir" 2>&1
)"
malformed_identity_status=$?
set -e
if [[ "$malformed_identity_status" -eq 0 ]]; then
  fail "malformed retained source identity unexpectedly passed"
fi
assert_contains "$malformed_identity_output" "invalid retained source-snapshot digest" "malformed retained identity"

echo "browser shard plan tests passed"
