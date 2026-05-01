#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
REPORTER="$ROOT_DIR/scripts/print-task-surface-report.mjs"
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

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: expected output not to contain [$needle]"
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

valid_output="$(assert_passes "current task-surface report" "$NODE_BIN" "$REPORTER" --check)"
assert_contains "$valid_output" "Cartulary task-surface report" "current report header"
assert_contains "$valid_output" "classification counts:" "current report classification summary"
assert_contains "$valid_output" "compact help count:" "current report compact help summary"
assert_contains "$valid_output" "help tier counts:" "current report help tier summary"
mapfile -t expected_count_lines < <("$NODE_BIN" - "$ROOT_DIR/tools/task_surface_manifest.json" <<'EOF'
const { readFileSync } = require("node:fs");
const manifest = JSON.parse(readFileSync(process.argv[2], "utf8"));
console.log(`compact: ${manifest.compact_help.entries.length}`);
for (const tier of manifest.help_tiers) {
  console.log(`${tier.name}: ${tier.entries.length}`);
}
EOF
)
for expected_count_line in "${expected_count_lines[@]}"; do
  assert_contains "$valid_output" "$expected_count_line" "current report manifest-derived count"
done
assert_not_contains "$valid_output" "public Make targets:" "current report compact output omits public target list"
assert_not_contains "$valid_output" "browser-e2e-measurement" "current report compact output omits target rows"
assert_contains "$valid_output" "use --all to print public targets" "current report compact output hint"

valid_all_output="$(assert_passes "current exhaustive task-surface report" "$NODE_BIN" "$REPORTER" --check --all)"
assert_contains "$valid_all_output" "public Make targets:" "current exhaustive report public target section"
assert_contains "$valid_all_output" "browser-e2e-measurement" "current exhaustive report measurement target"
assert_contains "$valid_all_output" "task classifications:" "current exhaustive report target section"
assert_contains "$valid_all_output" "logical harness checks:" "current exhaustive report harness section"
assert_contains "$valid_all_output" "phase-map execution dependencies:" "current exhaustive report phase dependency section"

phase_root="$(mktemp -d "$ROOT_DIR/tmp/task-surface-phase-root.XXXXXX")"
cleanup_paths+=("$phase_root")
mkdir -p "$phase_root/tools"
cat >"$phase_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v1",
  "phase": "phase99",
  "expected_ids": ["U-99-01"],
  "unit": [
    {
      "id": "U-99-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase1_store_test.go",
      "symbol": "TestPhase1_ConcurrencyLimitRevokesLRUNonCurrent_U_1_05",
      "execution_dependency": "backend_store",
      "execution_family": "backend-store",
      "execution_label": "Backend store",
      "evidence_layer": "store_domain",
      "claim": "task surface report discovers future phase dependencies",
      "out_of_scope": "task surface report discovers future phase dependencies"
    }
  ]
}
JSON
synthetic_report="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$phase_root" "$NODE_BIN" "$REPORTER" --json
)"
assert_contains "$synthetic_report" '"phase": "phase99"' "synthetic task-surface phase dependency"
assert_contains "$synthetic_report" '"execution_dependency": "backend_store"' "synthetic task-surface execution dependency"

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/task-surface-report.XXXXXX")"
cleanup_paths+=("$tmp_dir")
makefile_copy="$tmp_dir/Makefile"
manifest_copy="$tmp_dir/task_surface_manifest.json"
generated_make_copy="$tmp_dir/task_surface.generated.mk"
cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"

run_report_copy() {
  CARTULARY_TASK_SURFACE_MAKEFILE="$makefile_copy" \
  CARTULARY_TASK_SURFACE_MANIFEST="$manifest_copy" \
  CARTULARY_TASK_SURFACE_GENERATED_MAKE="$generated_make_copy" \
    "$NODE_BIN" "$REPORTER" --check
}

# shellcheck disable=SC2016
printf '\nRUN_GO_PHASE = @./scripts/lib/run-go-phase.sh\nRUN_PLAYWRIGHT_MANIFEST_PHASE_SCRIPT := $(CURDIR)/scripts/lib/run-playwright-manifest-phase.sh\n' >>"$makefile_copy"
retired_helper_output="$(assert_fails "retired runner helper drift" run_report_copy)"
assert_contains "$retired_helper_output" "retired runner-specific helper RUN_GO_PHASE" "retired runner helper output"
assert_contains "$retired_helper_output" "retired runner-specific helper RUN_PLAYWRIGHT_MANIFEST_PHASE_SCRIPT" "retired runner script helper output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
printf '\n.PHONY: unclassified-target\nunclassified-target:\n\t@true\n' >>"$makefile_copy"
unclassified_output="$(assert_fails "unclassified target drift" run_report_copy)"
assert_contains "$unclassified_output" "unclassified-target is missing task-surface classification" "unclassified target output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
printf '\n.PHONY: undocumented-public\nundocumented-public:\n\t@true\n' >>"$makefile_copy"
"$NODE_BIN" - "$manifest_copy" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const manifestPath = process.argv[2];
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
manifest.targets.push({
  name: "undocumented-public",
  classification: "public",
  included_in: ["helper_only"]
});
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
undocumented_output="$(assert_fails "public target without help tier" run_report_copy)"
assert_contains "$undocumented_output" "public target undocumented-public is missing help tier placement" "undocumented public output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
"$NODE_BIN" - "$manifest_copy" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const manifestPath = process.argv[2];
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
const investigation = manifest.help_tiers.find((tier) => tier.name === "investigate a run");
investigation.entries.push({
  target: "help",
  description: "duplicate target for synthetic coverage"
});
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
duplicate_help_tier_output="$(assert_fails "duplicate help tier target" run_report_copy)"
assert_contains "$duplicate_help_tier_output" "public target help appears in multiple help tiers" "duplicate help tier output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
"$NODE_BIN" - "$manifest_copy" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const manifestPath = process.argv[2];
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
const localDev = manifest.help_tiers.find((tier) => tier.name === "local dev");
localDev.entries = localDev.entries.filter((entry) => entry.target !== "clean");
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
missing_help_tier_output="$(assert_fails "missing help tier target" run_report_copy)"
assert_contains "$missing_help_tier_output" "public target clean is missing help tier placement" "missing help tier output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
"$NODE_BIN" - "$manifest_copy" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const manifestPath = process.argv[2];
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
const compact = manifest.compact_help;
compact.entries.push({
  target: "task-surface-check",
  description: "synthetic non-public help tier entry"
});
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
non_public_help_tier_output="$(assert_fails "non-public compact help target" run_report_copy)"
assert_contains "$non_public_help_tier_output" "task-surface-check appears in compact_help but is not classified public" "non-public compact help output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
"$NODE_BIN" - "$manifest_copy" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const manifestPath = process.argv[2];
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
manifest.compact_help.entries[1].target = "help";
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
duplicate_compact_help_output="$(assert_fails "duplicate compact help target" run_report_copy)"
assert_contains "$duplicate_compact_help_output" "compact_help.entries[2] contains duplicate target help" "duplicate compact help output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
printf '\n.PHONY: cap-target-1\ncap-target-1:\n\t@true\n' >>"$makefile_copy"
"$NODE_BIN" - "$manifest_copy" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const manifestPath = process.argv[2];
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
const localDev = manifest.help_tiers.find((tier) => tier.name === "local dev");
const compact = manifest.compact_help;
for (let index = 1; index <= 1; index += 1) {
  const target = `cap-target-${index}`;
  manifest.targets.push({
    name: target,
    classification: "public",
    included_in: ["helper_only"]
  });
  localDev.entries.push({
    target,
    description: "synthetic default help cap target"
  });
  compact.entries.push({
    target,
    description: "synthetic default help cap target"
  });
}
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
default_help_cap_output="$(assert_fails "compact help cap" run_report_copy)"
assert_contains "$default_help_cap_output" "compact_help.entries must not exceed 12 entries" "compact help cap output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
"$NODE_BIN" - "$manifest_copy" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const manifestPath = process.argv[2];
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
const target = manifest.targets.find((entry) => entry.name === "task-surface-report");
target.backing_scripts = ["scripts/missing-task-surface-helper.mjs"];
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
missing_script_output="$(assert_fails "missing backing script" run_report_copy)"
assert_contains "$missing_script_output" "backing script missing: scripts/missing-task-surface-helper.mjs" "missing backing script output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
printf '\n.PHONY: undeclared-script-ref\nundeclared-script-ref:\n\t@node ./scripts/check-toolchain-pins.mjs\n' >>"$makefile_copy"
"$NODE_BIN" - "$manifest_copy" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const manifestPath = process.argv[2];
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
manifest.targets.push({
  name: "undeclared-script-ref",
  classification: "helper_only",
  included_in: ["helper_only"],
  backing_scripts: []
});
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
undeclared_script_output="$(assert_fails "undeclared script reference" run_report_copy)"
assert_contains "$undeclared_script_output" "references scripts/check-toolchain-pins.mjs" "undeclared script reference output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
printf '\n# stale generated task surface\n' >>"$generated_make_copy"
stale_generated_output="$(assert_fails "stale generated task surface" run_report_copy)"
assert_contains "$stale_generated_output" "tools/task_surface.generated.mk is stale" "stale generated task surface output"
