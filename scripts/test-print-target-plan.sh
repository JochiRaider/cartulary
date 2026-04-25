#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
NODE_HELPER="${NODE_BIN:-node}"
MAKE_HELPER="${MAKE:-make}"
PLAN_SCRIPT="$ROOT_DIR/scripts/print-target-plan.mjs"
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
    fail "$label: expected output to contain [$needle]"
  fi
}

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/target-plan-smoke.XXXXXX")"
cleanup_paths+=("$tmp_dir")

json_a="$tmp_dir/target-plan-a.json"
json_b="$tmp_dir/target-plan-b.json"
"$NODE_HELPER" "$PLAN_SCRIPT" --json >"$json_a"
"$NODE_HELPER" "$PLAN_SCRIPT" --json >"$json_b"

"$NODE_HELPER" -e 'JSON.parse(require("node:fs").readFileSync(process.argv[1], "utf8"))' "$json_a"
cmp -s "$json_a" "$json_b" || fail "target-plan JSON must be deterministic across invocations"

backend_store_output="$("$NODE_HELPER" "$PLAN_SCRIPT" --target backend-store)"
assert_contains "$backend_store_output" "backend-store service_backed=1" "backend-store compact target plan"
assert_contains "$backend_store_output" "rows=" "backend-store compact row count"
assert_contains "$backend_store_output" "shared_reports=" "backend-store compact shared report count"

default_output="$("$NODE_HELPER" "$PLAN_SCRIPT")"
detail_output="$("$NODE_HELPER" "$PLAN_SCRIPT" --detail)"
default_lines="$(printf '%s\n' "$default_output" | wc -l | tr -d '[:space:]')"
detail_lines="$(printf '%s\n' "$detail_output" | wc -l | tr -d '[:space:]')"
if (( default_lines >= detail_lines )); then
  fail "default target-plan output must be more compact than --detail"
fi

backend_store_detail="$("$NODE_HELPER" "$PLAN_SCRIPT" --detail --target backend-store)"
for phase in phase1 phase2 phase3 phase4; do
  assert_contains "$backend_store_detail" "$phase unit authoritative" "backend-store detailed target plan"
done
assert_contains "$backend_store_detail" "packages:" "backend-store detail packages"

results_dir="$tmp_dir/results"
make_output="$(
  CARTULARY_TEST_RESULTS_DIR="$results_dir" \
    "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" explain-target TARGET=backend-store
)"
assert_contains "$make_output" "backend-store" "make explain-target"
assert_contains "$make_output" "phase1 unit authoritative" "make explain-target detailed default"
if [[ -d "$results_dir" ]] && [[ -n "$(find "$results_dir" -mindepth 1 -print -quit)" ]]; then
  fail "make explain-target must not create test report artifacts"
fi

make_compact_output="$("$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" explain-target TARGET=backend-store DETAIL=0)"
assert_contains "$make_compact_output" "backend-store service_backed=1" "make explain-target compact mode"

phase_root="$tmp_dir/phase-root"
mkdir -p "$phase_root/tools"
cp "$ROOT_DIR"/tools/phase*_test_map.json "$phase_root/tools/"
cat >"$phase_root/tools/phase5_test_map.json" <<'JSON'
{
  "expected_ids": ["U-5-01"],
  "unit": [],
  "integration": [],
  "e2e": [],
  "support_go_targets": [
    {
      "target": "backend_unit",
      "section": "unit",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_support_test.go",
      "selection_pattern": "TestSupportPhase5_",
      "symbol": "TestSupportPhase5_Discovered"
    }
  ]
}
JSON

discovered_phases="$(CARTULARY_PHASE_MANIFEST_ROOT="$phase_root" "$NODE_HELPER" "$ROOT_DIR/scripts/lib/phase-manifest.mjs" list-phases)"
assert_contains "$discovered_phases" "phase5" "phase manifest discovery includes phase5"

phase5_plan="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$phase_root" \
    "$NODE_HELPER" "$PLAN_SCRIPT" --json
)"
assert_contains "$phase5_plan" '"manifest_phase": "phase5"' "target-plan support rows include discovered phase"

phase5_shared_command="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$phase_root" \
  NODE_BIN="$NODE_HELPER" \
    "$ROOT_DIR/scripts/run-go-target.sh" inspect-shared-command backend-unit backend-unit-auth
)"
assert_contains "$phase5_shared_command" "TestSupportPhase5_Discovered" "run-go-target support selection includes discovered phase"
