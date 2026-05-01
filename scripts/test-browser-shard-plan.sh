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
  *) fail "future phase browser rows must be discovered from phase manifests, got [$future_phases]" ;;
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
