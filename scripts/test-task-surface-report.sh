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

write_phase_registry() {
  local root="$1"
  local phase="$2"
  local phase_number="${phase#phase}"

  mkdir -p "$root/tools"
  cat >"$root/tools/phase_registry.json" <<JSON
{
  "schema_id": "cartulary.phase_registry.v1",
  "phases": [
    {
      "phase": "$phase",
      "order": $phase_number,
      "status": "active",
      "label": "Phase $phase_number",
      "manifest_path": "tools/${phase}_test_map.json",
      "ledger_path": "docs/testing/${phase}_coverage_ledger.md",
      "scope": "synthetic $phase scope.",
      "normative_owners": "Synthetic owner."
    }
  ]
}
JSON
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
assert_not_contains "$(cat "$ROOT_DIR/Makefile")" "TARGET_OWNED_PHASE_TARGETS" "Makefile must not keep target ownership list"
assert_not_contains "$(cat "$ROOT_DIR/Makefile")" "export CARTULARY_TEST_TARGET" "Makefile must not keep target-specific test-target exports"

valid_all_output="$(assert_passes "current exhaustive task-surface report" "$NODE_BIN" "$REPORTER" --check --all)"
assert_contains "$valid_all_output" "public Make targets:" "current exhaustive report public target section"
assert_contains "$valid_all_output" "browser-e2e-measurement" "current exhaustive report measurement target"
assert_contains "$valid_all_output" "task classifications:" "current exhaustive report target section"
assert_contains "$valid_all_output" "logical harness checks:" "current exhaustive report harness section"
assert_contains "$valid_all_output" "phase-map execution dependencies:" "current exhaustive report phase dependency section"

"$NODE_BIN" --input-type=module - "$ROOT_DIR" <<'EOF'
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { pathToFileURL } from "node:url";

const [root] = process.argv.slice(2);
const manifest = JSON.parse(readFileSync(path.join(root, "tools/task_surface_manifest.json"), "utf8"));
const { helpAllLines, renderTaskSurfaceMake } = await import(pathToFileURL(path.join(root, "scripts/lib/task-surface.mjs")));
assert.equal(manifest.schema_id, "cartulary.task_surface_manifest.v11", "task surface schema must be v11");
assert.deepEqual(
  manifest.targets.map((target) => target.name).filter((target) => !manifest.make_recipes[target]),
  [],
  "every Make target must have a generated recipe",
);
assert.equal(manifest.make_recipes.help.type, "print_help", "help must be generated as print_help");
assert.equal(manifest.make_recipes["help-all"].scope, "all", "help-all must print exhaustive help");
assert.equal(manifest.make_recipes["frontend-install"].type, "alias", "frontend-install must be a generated alias");
assert.equal(
  manifest.make_recipes["frontend-install"].test_target,
  "self",
  "target-owned exports must use test_target self",
);
assert.equal(
  manifest.make_recipes["frontend-typecheck"].success_summary,
  true,
  "frontend-typecheck must keep its explicit pass summary",
);
const renderedMake = renderTaskSurfaceMake(manifest);
assert.match(
  helpAllLines(manifest).join("\n"),
  /phase -> target -> scheduler work unit -> artifact/,
  "help-all must explain the task evidence concept hierarchy",
);
assert.match(renderedMake, /^help:\n\t\$\([Q]\)printf '%s\\n' \$\(TASK_SURFACE_HELP_LINES\)$/m, "help recipe must be generated");
assert.match(
  renderedMake,
  /frontend-install: export CARTULARY_TEST_TARGET \?= frontend-install\nfrontend-install:\n\t\$\(Q\)env CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \$\(MAKE\) --silent --no-print-directory \$\(FRONTEND_INSTALL_STAMP\)/,
  "test_target self must render target-specific export and centralized prerequisite prelude",
);
assert.match(
  renderedMake,
  /\$\(call RUN_TARGET_SUMMARY,frontend-typecheck,pass\)/,
  "success_summary must render an explicit pass summary",
);
const longTarget = "synthetic-help-target-name-longer-than-command-column";
const usageTarget = "synthetic-help-usage";
manifest.targets.push(
  {
    name: longTarget,
    classification: "public",
    included_in: ["helper_only"],
  },
  {
    name: usageTarget,
    classification: "public",
    included_in: ["helper_only"],
  },
);
manifest.help_tiers[0].entries.push(
  {
    target: longTarget,
    description: "synthetic long target description",
  },
  {
    target: usageTarget,
    usage: "RESULTS_DIR=<dir>",
    description: "synthetic usage description",
  },
);

const lines = helpAllLines(manifest);
const continuationIndent = " ".repeat(38);
const longIndex = lines.indexOf(`  make ${longTarget}`);
assert.notEqual(longIndex, -1, "long target must render command on its own line");
assert.equal(
  lines[longIndex + 1],
  `${continuationIndent}synthetic long target description`,
  "long target description must render on an aligned continuation line",
);
const usageIndex = lines.indexOf(`  make ${usageTarget}`);
assert.notEqual(usageIndex, -1, "usage target must render command on its own line");
assert.equal(
  lines[usageIndex + 1],
  `${continuationIndent}RESULTS_DIR=<dir> synthetic usage description`,
  "usage text and description must render on an aligned continuation line",
);
assert.ok(
  !lines.some((line) => line.includes(`${longTarget} synthetic long target description`)),
  "long target help must not concatenate target and description",
);
EOF

phase_root="$(mktemp -d "$ROOT_DIR/tmp/task-surface-phase-root.XXXXXX")"
cleanup_paths+=("$phase_root")
mkdir -p "$phase_root/tools"
write_phase_registry "$phase_root" phase99
cat >"$phase_root/tools/phase99_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v1",
  "phase": "phase99",
  "note": "Synthetic task-surface report fixture.",
  "ledger": {
    "title": "Phase 99 Coverage Ledger",
    "notes": "Synthetic task-surface report fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase99",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-99-01"],
  "support_go_targets": [],
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
  ],
  "integration": [],
  "e2e": []
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
printf '\nTARGET_OWNED_PHASE_TARGETS := frontend-install\n' >>"$makefile_copy"
target_owned_output="$(assert_fails "target ownership list drift" run_report_copy)"
assert_contains "$target_owned_output" "Makefile must not define TARGET_OWNED_PHASE_TARGETS" "target ownership list output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
printf '\nfrontend-install: export CARTULARY_TEST_TARGET ?= frontend-install\n' >>"$makefile_copy"
target_export_output="$(assert_fails "target-specific test target export drift" run_report_copy)"
assert_contains "$target_export_output" "Makefile must not define target-specific CARTULARY_TEST_TARGET for frontend-install" "target-specific test target output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
"$NODE_BIN" - "$manifest_copy" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const manifestPath = process.argv[2];
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
delete manifest.make_recipes.help;
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
missing_recipe_output="$(assert_fails "missing generated recipe" run_report_copy)"
assert_contains "$missing_recipe_output" "make_recipes is missing target help" "missing generated recipe output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
printf '\n# stale generated task surface\n' >>"$generated_make_copy"
stale_generated_output="$(assert_fails "stale generated task surface" run_report_copy)"
assert_contains "$stale_generated_output" "tools/task_surface.generated.mk is stale" "stale generated task surface output"
