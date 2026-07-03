#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
PLANNER="$ROOT_DIR/tools/harness/browser/browser-shard-plan.mjs"
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

json_field() {
  local file="$1"
  local path="$2"

  "${NODE:-node}" -e '
const fs = require("node:fs");
const [file, path] = process.argv.slice(1);
const value = path.split(".").reduce((current, key) => current?.[key], JSON.parse(fs.readFileSync(file, "utf8")));
if (value === undefined || value === null) {
  process.exit(1);
}
process.stdout.write(Array.isArray(value) ? value.join(",") : String(value));
' "$file" "$path"
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
    fail "$label: expected output to contain [$needle], got [$haystack]"
  fi
}

tool_logs_from_result() {
  local output="$1"
  local root
  root="$(printf '%s\n' "$output" | sed -n 's/.* run_root=\([^ ]*\) .*/\1/p' | head -n 1)"
  if [[ -z "$root" ]]; then
    fail "missing run_root in output: $output"
  fi
  local dir
  if [[ "$root" = /* ]]; then
    dir="$root"
  else
    dir="$ROOT_DIR/$root"
  fi
  local target
  target="$(printf '%s\n' "$output" | sed -n 's/.* target=\([^ ]*\) .*/\1/p' | head -n 1)"
  [[ -f "$dir/stdout.log" ]] && cat "$dir/stdout.log"
  [[ -f "$dir/stderr.log" ]] && cat "$dir/stderr.log"
  if [[ -n "$target" ]]; then
    [[ -f "$dir/$target/stdout.log" ]] && cat "$dir/$target/stdout.log"
    [[ -f "$dir/$target/stderr.log" ]] && cat "$dir/$target/stderr.log"
  fi
}

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/browser-shard-plan.XXXXXX")"
cleanup_paths+=("$tmp_dir")
mkdir -p "$tmp_dir/manifests/tools"

cat >"$tmp_dir/manifests/tools/phase_registry.json" <<'JSON'
{
  "schema_id": "cartulary.phase_registry.v1",
  "phases": [
    {
      "phase": "phase1",
      "order": 1,
      "status": "active",
      "label": "Phase 1",
      "manifest_path": "tools/phase1_test_map.json",
      "ledger_path": "docs/testing/phase1_coverage_ledger.md",
      "scope": "synthetic phase1 scope.",
      "normative_owners": "Synthetic owner."
    },
    {
      "phase": "phase2",
      "order": 2,
      "status": "active",
      "label": "Phase 2",
      "manifest_path": "tools/phase2_test_map.json",
      "ledger_path": "docs/testing/phase2_coverage_ledger.md",
      "scope": "synthetic phase2 scope.",
      "normative_owners": "Synthetic owner."
    },
    {
      "phase": "phase12",
      "order": 12,
      "status": "active",
      "label": "Phase 12",
      "manifest_path": "tools/phase12_test_map.json",
      "ledger_path": "docs/testing/phase12_coverage_ledger.md",
      "scope": "synthetic phase12 scope.",
      "normative_owners": "Synthetic owner."
    }
  ]
}
JSON

cat >"$tmp_dir/manifests/tools/phase1_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase1",
  "note": "Synthetic browser shard plan fixture.",
  "ledger": {
    "title": "Phase 1 Coverage Ledger",
    "notes": "Synthetic browser shard plan fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase1",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["E-1-01", "E-1-02", "E-1-03"],
  "support_go_targets": [],
  "unit": [],
  "integration": [],
  "e2e": [
    {
      "id": "E-1-01",
      "coverage": "authoritative",
      "runner": "playwright",
      "file": "apps/web/e2e/alpha.spec.ts",
      "title": "E-1-01 alpha one",
      "execution_dependency": "browser_functional",
      "evidence_layer": "browser",
      "evidence_class": "product_conformance",
      "layer": "browser_functional",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "browser-shard-plan-fixture",
      "duplicate_of": null,
      "evidence_delta": "Synthetic browser shard plan fixture coverage.",
      "warm_local_cost_class": "medium",
      "claim": "alpha",
      "out_of_scope": "none"
    },
    {
      "id": "E-1-02",
      "coverage": "authoritative",
      "runner": "playwright",
      "file": "apps/web/e2e/alpha.spec.ts",
      "title": "E-1-02 alpha two",
      "execution_dependency": "browser_functional",
      "evidence_layer": "browser",
      "evidence_class": "product_conformance",
      "layer": "browser_functional",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "browser-shard-plan-fixture",
      "duplicate_of": null,
      "evidence_delta": "Synthetic browser shard plan fixture coverage.",
      "warm_local_cost_class": "medium",
      "claim": "alpha duplicate",
      "out_of_scope": "none"
    },
    {
      "id": "E-1-03",
      "coverage": "authoritative",
      "runner": "playwright",
      "file": "apps/web/e2e/beta.spec.ts",
      "titles": [
        "E-1-03 beta primary",
        "E-1-03 beta secondary"
      ],
      "execution_dependency": "browser_functional",
      "evidence_layer": "browser",
      "evidence_class": "product_conformance",
      "layer": "browser_functional",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "browser-shard-plan-fixture",
      "duplicate_of": null,
      "evidence_delta": "Synthetic browser shard plan fixture coverage.",
      "warm_local_cost_class": "medium",
      "claim": "beta",
      "out_of_scope": "none"
    }
  ]
}
JSON

cat >"$tmp_dir/manifests/tools/phase2_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase2",
  "note": "Synthetic browser shard plan fixture.",
  "ledger": {
    "title": "Phase 2 Coverage Ledger",
    "notes": "Synthetic browser shard plan fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase2",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["E-2-01", "E-2-02"],
  "support_go_targets": [],
  "unit": [],
  "integration": [],
  "e2e": [
    {
      "id": "E-2-01",
      "coverage": "authoritative",
      "runner": "playwright",
      "file": "apps/web/e2e/gamma.spec.ts",
      "title": "E-2-01 gamma",
      "execution_dependency": "browser_functional",
      "evidence_layer": "browser",
      "evidence_class": "product_conformance",
      "layer": "browser_functional",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "browser-shard-plan-fixture",
      "duplicate_of": null,
      "evidence_delta": "Synthetic browser shard plan fixture coverage.",
      "warm_local_cost_class": "medium",
      "claim": "gamma",
      "out_of_scope": "none"
    },
    {
      "id": "E-2-02",
      "coverage": "authoritative",
      "runner": "playwright",
      "file": "apps/web/e2e/ignored-stateful.spec.ts",
      "title": "E-2-02 ignored stateful",
      "execution_dependency": "browser_stateful",
      "evidence_layer": "browser",
      "evidence_class": "product_conformance",
      "layer": "browser_stateful",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "browser-shard-plan-fixture",
      "duplicate_of": null,
      "evidence_delta": "Synthetic browser shard plan fixture coverage.",
      "warm_local_cost_class": "medium",
      "claim": "ignored",
      "out_of_scope": "none"
    }
  ]
}
JSON

cat >"$tmp_dir/manifests/tools/phase12_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase12",
  "note": "Synthetic browser shard plan fixture.",
  "ledger": {
    "title": "Phase 12 Coverage Ledger",
    "notes": "Synthetic browser shard plan fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase12",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["E-12-01", "E-12-GRID-01"],
  "support_go_targets": [],
  "unit": [],
  "integration": [],
  "e2e": [
    {
      "id": "E-12-01",
      "coverage": "authoritative",
      "runner": "playwright",
      "file": "apps/web/e2e/future.spec.ts",
      "title": "E-12-01 future phase functional browser row",
      "execution_dependency": "browser_functional",
      "evidence_layer": "browser",
      "evidence_class": "product_conformance",
      "layer": "browser_functional",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "browser-shard-plan-fixture",
      "duplicate_of": null,
      "evidence_delta": "Synthetic browser shard plan fixture coverage.",
      "warm_local_cost_class": "medium",
      "claim": "future",
      "out_of_scope": "none"
    },
    {
      "id": "E-12-GRID-01",
      "coverage": "authoritative",
      "runner": "playwright",
      "file": "apps/web/e2e/zz-grid.spec.ts",
      "title": "E-12-GRID-01 future phase named-grid functional browser row",
      "execution_dependency": "browser_functional",
      "evidence_layer": "browser",
      "evidence_class": "product_conformance",
      "layer": "browser_functional",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "browser-shard-plan-fixture",
      "duplicate_of": null,
      "evidence_delta": "Synthetic browser shard plan fixture coverage.",
      "warm_local_cost_class": "medium",
      "claim": "future named row",
      "out_of_scope": "none"
    }
  ]
}
JSON

cat >"$tmp_dir/baseline.json" <<'JSON'
{
  "schema_id": "cartulary.browser_e2e_duration_baselines.v3",
  "default_entry_weight_ms": 7000,
  "shard_target_ms": 8000,
  "entries": {
    "E-1-01": {
      "file": "apps/web/e2e/alpha.spec.ts",
      "title": "E-1-01 alpha one",
      "weight_ms": 30000
    },
    "E-1-02": {
      "file": "apps/web/e2e/alpha.spec.ts",
      "title": "E-1-02 alpha two",
      "weight_ms": 20000
    },
    "E-1-03": {
      "file": "apps/web/e2e/beta.spec.ts",
      "title": "E-1-03 beta primary",
      "weight_ms": 5000
    }
  }
}
JSON

node_cmd="${NODE:-node}"

CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
  "$node_cmd" "$PLANNER" plan --baseline-file "$tmp_dir/baseline.json" --max-shards 3 >"$tmp_dir/plan.json"

assert_equals "$(json_field "$tmp_dir/plan.json" "shard_count")" "3" "shard count respects max and target weight"
assert_equals "$(json_field "$tmp_dir/plan.json" "entries.0.file")" "apps/web/e2e/alpha.spec.ts" "deterministic entry ordering"
assert_equals "$(json_field "$tmp_dir/plan.json" "entries.2.titles.1")" "E-1-03 beta secondary" "multi-title Playwright rows keep all executable scenarios"
assert_equals "$(json_field "$tmp_dir/plan.json" "shards.0.entries.0.id")" "E-1-01" "largest same-file entry gets first stable shard"
assert_equals "$(json_field "$tmp_dir/plan.json" "shards.1.entries.0.id")" "E-1-02" "same-file entries can split across shards"
"$node_cmd" - "$tmp_dir/plan.json" <<'EOF'
const fs = require("node:fs");
const [planFile] = process.argv.slice(2);
const plan = JSON.parse(fs.readFileSync(planFile, "utf8"));
if (plan.entry_count !== plan.entries.length) {
  throw new Error("entry_count must match emitted entries length");
}
for (const id of ["E-1-01", "E-1-02", "E-1-03", "E-2-01", "E-12-01", "E-12-GRID-01"]) {
  if (!plan.entries.some((entry) => entry.id === id)) {
    throw new Error(`synthetic manifest row ${id} missing from shard plan`);
  }
}
const future = plan.entries.find((entry) => entry.id === "E-12-01");
if (!future || future.file !== "apps/web/e2e/future.spec.ts") {
  throw new Error("numeric future phase discovery must keep deterministic files");
}
const unweighted = plan.entries.find((entry) => entry.id === "E-2-01");
if (!unweighted || unweighted.weight_ms !== 7000) {
  throw new Error("missing baseline entries must use the configured default weight");
}
if (!plan.shards.some((shard) => String(shard.grep).includes("E-1-03 beta secondary"))) {
  throw new Error("multi-title Playwright shard grep must include every row title");
}
EOF

CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
  "$node_cmd" "$PLANNER" selected-tests "$tmp_dir/plan.json" phase1 >"$tmp_dir/selected-phase1.json"
assert_equals "$(json_field "$tmp_dir/selected-phase1.json" "schema_id")" "cartulary.playwright_manifest_selection.v1" "selected tests schema"
assert_equals "$(json_field "$tmp_dir/selected-phase1.json" "expected_count")" "4" "selected tests flatten multi-title rows"
"$node_cmd" - "$tmp_dir/selected-phase1.json" <<'EOF'
const fs = require("node:fs");
const [selectionFile] = process.argv.slice(2);
const selection = JSON.parse(fs.readFileSync(selectionFile, "utf8"));
const keys = new Set(
  (selection.selected_tests ?? []).map((entry) => `${entry.id}::${entry.file}::${entry.title}`),
);
for (const key of [
  "E-1-01::e2e/alpha.spec.ts::E-1-01 alpha one",
  "E-1-03::e2e/beta.spec.ts::E-1-03 beta primary",
  "E-1-03::e2e/beta.spec.ts::E-1-03 beta secondary",
]) {
  if (!keys.has(key)) {
    throw new Error(`selected-tests report missing ${key}`);
  }
}
EOF
multi_title_shard="$("$node_cmd" - "$tmp_dir/plan.json" <<'EOF'
const fs = require("node:fs");
const [planFile] = process.argv.slice(2);
const plan = JSON.parse(fs.readFileSync(planFile, "utf8"));
for (const shard of plan.shards ?? []) {
  if ((shard.entries ?? []).some((entry) => entry.id === "E-1-03")) {
    process.stdout.write(shard.name);
    process.exit(0);
  }
}
throw new Error("missing shard for E-1-03");
EOF
)"
CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
  "$node_cmd" "$PLANNER" selected-tests "$tmp_dir/plan.json" phase1 "$multi_title_shard" >"$tmp_dir/selected-phase1-shard.json"
"$node_cmd" - "$tmp_dir/selected-phase1-shard.json" <<'EOF'
const fs = require("node:fs");
const [selectionFile] = process.argv.slice(2);
const selected = JSON.parse(fs.readFileSync(selectionFile, "utf8")).selected_tests ?? [];
const beta = selected.filter((entry) => entry.id === "E-1-03").map((entry) => entry.title).sort();
const expected = ["E-1-03 beta primary", "E-1-03 beta secondary"];
if (JSON.stringify(beta) !== JSON.stringify(expected)) {
  throw new Error(`per-shard selected-tests must keep only selected multi-title scenarios, got ${JSON.stringify(beta)}`);
}
if (selected.some((entry) => entry.phase && entry.phase !== "phase1")) {
  throw new Error("per-shard selected-tests leaked a non-phase1 entry");
}
EOF

CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
  "$node_cmd" "$PLANNER" plan --phase phase2 --baseline-file "$tmp_dir/baseline.json" --max-shards 3 >"$tmp_dir/phase2-plan.json"
assert_equals "$(json_field "$tmp_dir/phase2-plan.json" "phase")" "phase2" "phase-filtered plan records selected phase"

CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
  "$node_cmd" "$PLANNER" plan --frontend-row-ids FE-I-P5-01 --baseline-file "$tmp_dir/baseline.json" --max-shards 3 >"$tmp_dir/frontend-row-plan.json"
assert_equals "$(json_field "$tmp_dir/frontend-row-plan.json" "entry_count")" "1" "frontend selected-row plan keeps only selected row"
assert_equals "$(json_field "$tmp_dir/frontend-row-plan.json" "entries.0.id")" "FE-I-P5-01" "frontend selected-row plan records FE row id"
assert_equals "$(json_field "$tmp_dir/frontend-row-plan.json" "entries.0.file")" "apps/web/e2e/frontend.phase5.grid-provenance.spec.ts" "frontend selected-row plan records FE row file"
assert_equals "$(json_field "$tmp_dir/frontend-row-plan.json" "shards.0.entries.0.id")" "FE-I-P5-01" "frontend selected-row shard records FE row id"
set +e
unknown_frontend_row_output="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
    "$node_cmd" "$PLANNER" plan --frontend-row-ids FE-I-P99-01 --baseline-file "$tmp_dir/baseline.json" --max-shards 3 2>&1
)"
unknown_frontend_row_status=$?
set -e
if [[ "$unknown_frontend_row_status" -eq 0 ]]; then
  fail "frontend selected-row browser planning should reject unknown row ids"
fi
assert_contains "$unknown_frontend_row_output" "selected frontend browser row id(s) not found: FE-I-P99-01" "frontend selected-row browser planning unknown id"

stale_metadata_baseline="$tmp_dir/browser-stale-metadata-plan.json"
cp "$tmp_dir/baseline.json" "$stale_metadata_baseline"
"$node_cmd" - "$stale_metadata_baseline" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
baseline.entries["E-1-01"].title = "E-1-01 stale title";
fs.writeFileSync(baselineFile, `${JSON.stringify(baseline, null, 2)}\n`);
EOF
set +e
stale_plan_output="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
    "$node_cmd" "$PLANNER" plan --baseline-file "$stale_metadata_baseline" --max-shards 3 2>&1
)"
stale_plan_status=$?
set -e
if [[ "$stale_plan_status" -eq 0 ]]; then
  fail "browser baseline planning should reject stale title metadata"
fi
assert_contains "$stale_plan_output" "entries.E-1-01 must match active manifest file/title" "browser baseline planning stale metadata"
assert_equals "$(json_field "$tmp_dir/phase2-plan.json" "entries.0.file")" "apps/web/e2e/gamma.spec.ts" "phase-filtered plan selects phase2 functional file"
"$node_cmd" - "$tmp_dir/phase2-plan.json" <<'EOF'
const fs = require("node:fs");
const [planFile] = process.argv.slice(2);
const plan = JSON.parse(fs.readFileSync(planFile, "utf8"));
if (plan.entry_count !== plan.entries.length) {
  throw new Error("phase-filtered entry_count must match emitted entries length");
}
if (!plan.entries.every((entry) => entry.phase === "phase2")) {
  throw new Error("phase-filtered plan must keep only selected phase rows");
}
if (!plan.entries.some((entry) => entry.id === "E-2-01")) {
  throw new Error("phase-filtered plan must include the phase2 manifest row");
}
EOF

"$node_cmd" -e '
const fs = require("node:fs");
const plan = JSON.parse(fs.readFileSync(process.argv[1], "utf8"));
const files = new Set(plan.shards.flatMap((shard) => shard.files));
if (files.has("apps/web/e2e/ignored-stateful.spec.ts")) {
  throw new Error("stateful browser row leaked into functional shard plan");
}
' "$tmp_dir/plan.json"

future_phases="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
    "$node_cmd" "$ROOT_DIR/tools/harness/planning/phase-manifest.mjs" playwright-phases authoritative browser_functional
)"
case "$future_phases" in
  *phase12*) ;;
  *) fail "future phase browser rows must be selected from registry phase manifests, got [$future_phases]" ;;
esac

future_files="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
    "$node_cmd" "$ROOT_DIR/tools/harness/planning/phase-manifest.mjs" playwright-files-all authoritative browser_functional
)"
assert_contains "$future_files" "e2e/gamma.spec.ts" "future phase browser file discovery"
assert_contains "$future_files" "e2e/future.spec.ts" "future phase browser file discovery"

future_count="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
    "$node_cmd" "$ROOT_DIR/tools/harness/planning/phase-manifest.mjs" playwright-count-all authoritative browser_functional
)"
if [[ "$future_count" -lt 7 ]]; then
  fail "future phase browser title count discovery: expected at least [7], got [$future_count]"
fi

browser_results="$tmp_dir/browser-results"
timing_dir="$browser_results/browser-e2e-webserver-backed/browser-e2e-functional-authoritative"
failed_timing_dir="$browser_results/browser-e2e-webserver-backed/browser-e2e-functional-failed"
mkdir -p "$timing_dir" "$failed_timing_dir"
cat >"$timing_dir/phase-summary.json" <<'JSON'
{
  "target": "browser-e2e-webserver-backed",
  "runner": "playwright",
  "status": "pass"
}
JSON
cat >"$timing_dir/playwright-timing.json" <<'JSON'
{
  "entries": [
    {
      "id": "E-1-01",
      "phase": "phase1",
      "file": "apps/web/e2e/alpha.spec.ts",
      "title": "E-1-01 alpha one",
      "wall_duration_ms": 32000
    },
    {
      "id": "E-1-02",
      "phase": "phase1",
      "file": "apps/web/e2e/alpha.spec.ts",
      "title": "E-1-02 alpha two",
      "wall_duration_ms": 21000
    },
    {
      "id": "E-1-03",
      "phase": "phase1",
      "file": "e2e/beta.spec.ts",
      "title": "E-1-03 beta",
      "wall_duration_ms": 9000
    },
    {
      "id": "E-2-01",
      "phase": "phase2",
      "file": "apps/web/e2e/gamma.spec.ts",
      "title": "E-2-01 gamma",
      "wall_duration_ms": 7000
    },
    {
      "id": "E-12-01",
      "phase": "phase12",
      "file": "apps/web/e2e/future.spec.ts",
      "title": "E-12-01 future phase functional browser row",
      "wall_duration_ms": 11000
    },
    {
      "id": "E-12-GRID-01",
      "phase": "phase12",
      "file": "apps/web/e2e/zz-grid.spec.ts",
      "title": "E-12-GRID-01 future phase named-grid functional browser row",
      "wall_duration_ms": 12000
    }
  ]
}
JSON
cat >"$failed_timing_dir/phase-summary.json" <<'JSON'
{
  "target": "browser-e2e-webserver-backed",
  "runner": "playwright",
  "status": "fail"
}
JSON
cat >"$failed_timing_dir/playwright-timing.json" <<'JSON'
{
  "entries": [
    {
      "id": "E-1-01",
      "phase": "phase1",
      "file": "apps/web/e2e/alpha.spec.ts",
      "title": "E-1-01 alpha one",
      "wall_duration_ms": 99999
    }
  ]
}
JSON

cat >"$tmp_dir/browser-refresh-baseline.json" <<'JSON'
{
  "schema_id": "cartulary.browser_e2e_duration_baselines.v3",
  "note": "old note",
  "default_entry_weight_ms": 7000,
  "shard_target_ms": 8000,
  "retained_metadata": {
    "owner": "browser"
  },
  "entries": {
    "E-1-01": {
      "file": "apps/web/e2e/stale-alpha.spec.ts",
      "title": "E-1-01 stale alpha title",
      "weight_ms": 1
    },
    "E-99-01": {
      "file": "apps/web/e2e/retired.spec.ts",
      "title": "E-99-01 retired row",
      "weight_ms": 1
    }
  }
}
JSON
refresh_output="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
    "$node_cmd" "$PLANNER" update-baselines --baseline-file "$tmp_dir/browser-refresh-baseline.json" "$browser_results"
)"
assert_contains "$refresh_output" "updated 6 browser E2E row duration baselines" "browser baseline refresh output"
"$node_cmd" - "$tmp_dir/browser-refresh-baseline.json" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
const entryKeys = Object.keys(baseline.entries);
const expected = [
  "E-1-01",
  "E-1-02",
  "E-1-03",
  "E-12-01",
  "E-12-GRID-01",
  "E-2-01",
];
if (JSON.stringify(entryKeys) !== JSON.stringify(expected)) {
  throw new Error(`expected sorted refreshed entries ${JSON.stringify(expected)}, got ${JSON.stringify(entryKeys)}`);
}
if (baseline.entries["E-1-01"].weight_ms !== 32000) {
  throw new Error(`failed timing artifact leaked into refresh, got E-1-01=${baseline.entries["E-1-01"].weight_ms}`);
}
if (baseline.entries["E-1-01"].file !== "apps/web/e2e/alpha.spec.ts" || baseline.entries["E-1-01"].title !== "E-1-01 alpha one") {
  throw new Error("baseline refresh must rewrite stale file/title metadata from the active manifest");
}
if (baseline.default_entry_weight_ms !== 7000 || baseline.shard_target_ms !== 8000) {
  throw new Error("baseline refresh must preserve durable weighting metadata");
}
if (baseline.retained_metadata?.owner !== "browser") {
  throw new Error("baseline refresh must preserve unknown durable metadata");
}
if (!String(baseline.note).includes("make browser-e2e-duration-baselines RESULTS_DIR=<dir>")) {
  throw new Error(`expected public refresh command in note, got ${baseline.note}`);
}
EOF

cat >"$tmp_dir/browser-make-baseline.json" <<'JSON'
{
  "schema_id": "cartulary.browser_e2e_duration_baselines.v3",
  "default_entry_weight_ms": 7000,
  "shard_target_ms": 8000,
  "entries": {}
}
JSON
make_refresh_output="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
  BROWSER_E2E_DURATION_BASELINE="$tmp_dir/browser-make-baseline.json" \
  RESULTS_DIR="$browser_results" \
    env -u CARTULARY_TEST_RESULTS_DIR -u CARTULARY_TEST_RUN_ID \
      "${MAKE:-make}" --no-print-directory -C "$ROOT_DIR" browser-e2e-duration-baselines 2>&1
)"
assert_contains "$make_refresh_output" "[RESULT] target=browser-e2e-duration-baselines status=pass" "make browser baseline refresh summary"
assert_contains "$(tool_logs_from_result "$make_refresh_output")" "updated 6 browser E2E row duration baselines" "make browser baseline refresh output"
CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
  "$node_cmd" "$PLANNER" check-baseline-drift --baseline-file "$tmp_dir/browser-make-baseline.json" "$browser_results" >/dev/null

tolerated_baseline="$tmp_dir/browser-tolerated-drift.json"
cp "$tmp_dir/browser-make-baseline.json" "$tolerated_baseline"
"$node_cmd" - "$tolerated_baseline" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
baseline.entries["E-1-01"].weight_ms = 13000;
fs.writeFileSync(baselineFile, `${JSON.stringify(baseline, null, 2)}\n`);
EOF
CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
  "$node_cmd" "$PLANNER" check-baseline-drift --baseline-file "$tolerated_baseline" "$browser_results" >/dev/null

underplanned_baseline="$tmp_dir/browser-underplanned.json"
cp "$tmp_dir/browser-make-baseline.json" "$underplanned_baseline"
"$node_cmd" - "$underplanned_baseline" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
baseline.entries["E-1-01"].weight_ms = 1000;
fs.writeFileSync(baselineFile, `${JSON.stringify(baseline, null, 2)}\n`);
EOF
set +e
underplanned_output="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
    "$node_cmd" "$PLANNER" check-baseline-drift --baseline-file "$underplanned_baseline" "$browser_results" 2>&1
)"
underplanned_status=$?
set -e
if [[ "$underplanned_status" -eq 0 ]]; then
  fail "browser underplanned drift should fail"
fi
assert_contains "$underplanned_output" "underplanned id=E-1-01 file=apps/web/e2e/alpha.spec.ts" "browser underplanned drift"
assert_contains "$underplanned_output" "make browser-e2e-duration-baselines RESULTS_DIR=" "browser drift refresh guidance"

overplanned_baseline="$tmp_dir/browser-overplanned.json"
cp "$tmp_dir/browser-make-baseline.json" "$overplanned_baseline"
"$node_cmd" - "$overplanned_baseline" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
baseline.entries["E-2-01"].weight_ms = 50000;
fs.writeFileSync(baselineFile, `${JSON.stringify(baseline, null, 2)}\n`);
EOF
set +e
overplanned_output="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
    "$node_cmd" "$PLANNER" check-baseline-drift --baseline-file "$overplanned_baseline" "$browser_results" 2>&1
)"
overplanned_status=$?
set -e
if [[ "$overplanned_status" -eq 0 ]]; then
  fail "browser overplanned drift should fail"
fi
assert_contains "$overplanned_output" "overplanned id=E-2-01 file=apps/web/e2e/gamma.spec.ts" "browser overplanned drift"

missing_results="$tmp_dir/browser-missing-results"
missing_timing_dir="$missing_results/browser-e2e-webserver-backed/browser-e2e-functional-authoritative"
mkdir -p "$missing_timing_dir"
cp "$timing_dir/phase-summary.json" "$missing_timing_dir/phase-summary.json"
cat >"$missing_timing_dir/playwright-timing.json" <<'JSON'
{
  "entries": [
    {
      "id": "E-1-01",
      "phase": "phase1",
      "file": "apps/web/e2e/alpha.spec.ts",
      "title": "E-1-01 alpha one",
      "wall_duration_ms": 32000
    }
  ]
}
JSON
set +e
missing_refresh_output="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
    "$node_cmd" "$PLANNER" update-baselines --baseline-file "$tmp_dir/browser-missing-refresh.json" "$missing_results" 2>&1
)"
missing_refresh_status=$?
set -e
if [[ "$missing_refresh_status" -eq 0 ]]; then
  fail "browser baseline refresh should require all authoritative functional specs"
fi
assert_contains "$missing_refresh_output" "missing observed browser entry timings:" "browser missing observed refresh output"
