#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
PLANNER="$ROOT_DIR/scripts/lib/browser-shard-plan.mjs"
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
  "schema_id": "cartulary.phase_test_map.v1",
  "phase": "phase1",
  "expected_ids": ["E-1-01", "E-1-02", "E-1-03"],
  "e2e": [
    {
      "id": "E-1-01",
      "coverage": "authoritative",
      "runner": "playwright",
      "file": "apps/web/e2e/alpha.spec.ts",
      "title": "E-1-01 alpha one",
      "execution_dependency": "browser_functional",
      "evidence_layer": "browser",
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
      "claim": "alpha duplicate",
      "out_of_scope": "none"
    },
    {
      "id": "E-1-03",
      "coverage": "authoritative",
      "runner": "playwright",
      "file": "apps/web/e2e/beta.spec.ts",
      "title": "E-1-03 beta",
      "execution_dependency": "browser_functional",
      "evidence_layer": "browser",
      "claim": "beta",
      "out_of_scope": "none"
    }
  ]
}
JSON

cat >"$tmp_dir/manifests/tools/phase2_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v1",
  "phase": "phase2",
  "expected_ids": ["E-2-01", "E-2-02"],
  "e2e": [
    {
      "id": "E-2-01",
      "coverage": "authoritative",
      "runner": "playwright",
      "file": "apps/web/e2e/gamma.spec.ts",
      "title": "E-2-01 gamma",
      "execution_dependency": "browser_functional",
      "evidence_layer": "browser",
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
      "claim": "ignored",
      "out_of_scope": "none"
    }
  ]
}
JSON

cat >"$tmp_dir/manifests/tools/phase12_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v1",
  "phase": "phase12",
  "expected_ids": ["E-12-01"],
  "e2e": [
    {
      "id": "E-12-01",
      "coverage": "authoritative",
      "runner": "playwright",
      "file": "apps/web/e2e/future.spec.ts",
      "title": "E-12-01 future phase functional browser row",
      "execution_dependency": "browser_functional",
      "evidence_layer": "browser",
      "claim": "future",
      "out_of_scope": "none"
    }
  ]
}
JSON

cat >"$tmp_dir/baseline.json" <<'JSON'
{
  "schema_id": "cartulary.browser_e2e_duration_baselines.v1",
  "default_spec_weight_ms": 7000,
  "shard_target_ms": 8000,
  "specs": {
    "apps/web/e2e/alpha.spec.ts": 30000,
    "apps/web/e2e/beta.spec.ts": 5000
  }
}
JSON

node_cmd="${NODE:-node}"

CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
  "$node_cmd" "$PLANNER" plan --baseline-file "$tmp_dir/baseline.json" --max-shards 3 >"$tmp_dir/plan.json"

assert_equals "$(json_field "$tmp_dir/plan.json" "spec_count")" "4" "spec count collapses duplicate manifest rows"
assert_equals "$(json_field "$tmp_dir/plan.json" "shard_count")" "3" "shard count respects max and target weight"
assert_equals "$(json_field "$tmp_dir/plan.json" "specs.0.file")" "apps/web/e2e/alpha.spec.ts" "deterministic spec ordering"
assert_equals "$(json_field "$tmp_dir/plan.json" "specs.2.weight_ms")" "7000" "missing baseline uses default weight"
assert_equals "$(json_field "$tmp_dir/plan.json" "specs.3.file")" "apps/web/e2e/gamma.spec.ts" "numeric future phase discovery keeps deterministic files"
assert_equals "$(json_field "$tmp_dir/plan.json" "shards.0.files")" "apps/web/e2e/alpha.spec.ts" "long spec gets first stable shard"
assert_equals "$(json_field "$tmp_dir/plan.json" "shards.0.entries.0.id")" "E-1-01" "duplicate spec keeps first row"
assert_equals "$(json_field "$tmp_dir/plan.json" "shards.0.entries.1.id")" "E-1-02" "duplicate spec keeps second row"

CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
  "$node_cmd" "$PLANNER" plan --phase phase2 --baseline-file "$tmp_dir/baseline.json" --max-shards 3 >"$tmp_dir/phase2-plan.json"
assert_equals "$(json_field "$tmp_dir/phase2-plan.json" "phase")" "phase2" "phase-filtered plan records selected phase"
assert_equals "$(json_field "$tmp_dir/phase2-plan.json" "spec_count")" "1" "phase-filtered plan keeps only selected functional specs"
assert_equals "$(json_field "$tmp_dir/phase2-plan.json" "specs.0.file")" "apps/web/e2e/gamma.spec.ts" "phase-filtered plan selects phase2 functional file"
assert_equals "$(json_field "$tmp_dir/phase2-plan.json" "shards.0.entries.0.id")" "E-2-01" "phase-filtered plan selects phase2 row"

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
    "$node_cmd" "$ROOT_DIR/scripts/lib/phase-manifest.mjs" playwright-phases authoritative browser_functional
)"
case "$future_phases" in
  *phase12*) ;;
  *) fail "future phase browser rows must be selected from registry phase manifests, got [$future_phases]" ;;
esac

future_files="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
    "$node_cmd" "$ROOT_DIR/scripts/lib/phase-manifest.mjs" playwright-files-all authoritative browser_functional
)"
assert_contains "$future_files" "e2e/gamma.spec.ts" "future phase browser file discovery"

future_count="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
    "$node_cmd" "$ROOT_DIR/scripts/lib/phase-manifest.mjs" playwright-count-all authoritative browser_functional
)"
assert_equals "$future_count" "5" "future phase browser row count discovery"

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
  "files": [
    {
      "file": "apps/web/e2e/alpha.spec.ts",
      "wall_duration_ms": 32000
    },
    {
      "file": "e2e/beta.spec.ts",
      "wall_duration_ms": 9000
    },
    {
      "file": "apps/web/e2e/future.spec.ts",
      "wall_duration_ms": 11000
    },
    {
      "file": "apps/web/e2e/gamma.spec.ts",
      "wall_duration_ms": 7000
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
  "files": [
    {
      "file": "apps/web/e2e/alpha.spec.ts",
      "wall_duration_ms": 99999
    }
  ]
}
JSON

cat >"$tmp_dir/browser-refresh-baseline.json" <<'JSON'
{
  "schema_id": "cartulary.browser_e2e_duration_baselines.v1",
  "note": "old note",
  "default_spec_weight_ms": 7000,
  "shard_target_ms": 8000,
  "retained_metadata": {
    "owner": "browser"
  },
  "specs": {
    "apps/web/e2e/retired.spec.ts": 1234
  }
}
JSON
refresh_output="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
    "$node_cmd" "$PLANNER" update-baselines --baseline-file "$tmp_dir/browser-refresh-baseline.json" "$browser_results"
)"
assert_contains "$refresh_output" "updated 4 browser E2E duration baselines" "browser baseline refresh output"
"$node_cmd" - "$tmp_dir/browser-refresh-baseline.json" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
const specKeys = Object.keys(baseline.specs);
const expected = [
  "apps/web/e2e/alpha.spec.ts",
  "apps/web/e2e/beta.spec.ts",
  "apps/web/e2e/future.spec.ts",
  "apps/web/e2e/gamma.spec.ts",
];
if (JSON.stringify(specKeys) !== JSON.stringify(expected)) {
  throw new Error(`expected sorted refreshed specs ${JSON.stringify(expected)}, got ${JSON.stringify(specKeys)}`);
}
if (baseline.specs["apps/web/e2e/alpha.spec.ts"] !== 32000) {
  throw new Error(`failed timing artifact leaked into refresh, got alpha=${baseline.specs["apps/web/e2e/alpha.spec.ts"]}`);
}
if (baseline.default_spec_weight_ms !== 7000 || baseline.shard_target_ms !== 8000) {
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
  "schema_id": "cartulary.browser_e2e_duration_baselines.v1",
  "default_spec_weight_ms": 7000,
  "shard_target_ms": 8000,
  "specs": {}
}
JSON
make_refresh_output="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
  BROWSER_E2E_DURATION_BASELINE="$tmp_dir/browser-make-baseline.json" \
  RESULTS_DIR="$browser_results" \
    "${MAKE:-make}" --no-print-directory -C "$ROOT_DIR" browser-e2e-duration-baselines 2>&1
)"
assert_contains "$make_refresh_output" "updated 4 browser E2E duration baselines" "make browser baseline refresh output"
CARTULARY_PHASE_MANIFEST_ROOT="$tmp_dir/manifests" \
  "$node_cmd" "$PLANNER" check-baseline-drift --baseline-file "$tmp_dir/browser-make-baseline.json" "$browser_results" >/dev/null

tolerated_baseline="$tmp_dir/browser-tolerated-drift.json"
cp "$tmp_dir/browser-make-baseline.json" "$tolerated_baseline"
"$node_cmd" - "$tolerated_baseline" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
baseline.specs["apps/web/e2e/alpha.spec.ts"] = 13000;
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
baseline.specs["apps/web/e2e/alpha.spec.ts"] = 1000;
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
assert_contains "$underplanned_output" "underplanned file=apps/web/e2e/alpha.spec.ts" "browser underplanned drift"
assert_contains "$underplanned_output" "make browser-e2e-duration-baselines RESULTS_DIR=" "browser drift refresh guidance"

overplanned_baseline="$tmp_dir/browser-overplanned.json"
cp "$tmp_dir/browser-make-baseline.json" "$overplanned_baseline"
"$node_cmd" - "$overplanned_baseline" <<'EOF'
const fs = require("node:fs");
const [baselineFile] = process.argv.slice(2);
const baseline = JSON.parse(fs.readFileSync(baselineFile, "utf8"));
baseline.specs["apps/web/e2e/gamma.spec.ts"] = 50000;
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
assert_contains "$overplanned_output" "overplanned file=apps/web/e2e/gamma.spec.ts" "browser overplanned drift"

missing_results="$tmp_dir/browser-missing-results"
missing_timing_dir="$missing_results/browser-e2e-webserver-backed/browser-e2e-functional-authoritative"
mkdir -p "$missing_timing_dir"
cp "$timing_dir/phase-summary.json" "$missing_timing_dir/phase-summary.json"
cat >"$missing_timing_dir/playwright-timing.json" <<'JSON'
{
  "files": [
    {
      "file": "apps/web/e2e/alpha.spec.ts",
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
assert_contains "$missing_refresh_output" "missing observed browser spec timings:" "browser missing observed refresh output"
