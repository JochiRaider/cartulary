#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
REPORTER="$ROOT_DIR/tools/harness/generated-artifacts/task-surface-report-cli.mjs"
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
assert_contains "$valid_output" "target_class counts:" "current report target_class summary"
assert_contains "$valid_output" "compact help count:" "current report compact help summary"
assert_contains "$valid_output" "scheduler growth:" "current report scheduler growth summary"
assert_contains "$valid_output" "scratch_generations_byte_identical: true" "scheduler generation determinism gate"
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
assert_contains "$valid_all_output" "side-effect counts:" "current report side-effect count section"
assert_contains "$valid_all_output" "runtime_reset:" "current report runtime reset side-effect count"
assert_contains "$valid_all_output" "side_effects=" "current exhaustive report public target side effects"
assert_contains "$valid_all_output" "task target classes:" "current exhaustive report target section"
assert_contains "$valid_all_output" "logical harness checks:" "current exhaustive report harness section"
assert_contains "$valid_all_output" "catalog target partitions:" "current exhaustive report catalog partition section"

"$NODE_BIN" --input-type=module - "$ROOT_DIR" <<'EOF'
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { pathToFileURL } from "node:url";

const [root] = process.argv.slice(2);
const manifest = JSON.parse(readFileSync(path.join(root, "tools/task_surface_manifest.json"), "utf8"));
const {
  collectTaskSurfaceManifestErrors,
  helpAllLines,
  makeRecipeEntries,
  renderTaskSurfaceMake,
  renderTaskSurfaceMakeRuntime,
  requiredRecipeBackingScripts,
  workGraphRuntimeSource,
} = await import(pathToFileURL(path.join(root, "tools/harness/generated-artifacts/task-surface/index.mjs")));
assert.equal(manifest.schema_id, "cartulary.task_surface_manifest.v16", "task surface schema must be v16");
assert.deepEqual(
  manifest.global_inputs.map((input) => input.name),
  ["CARTULARY_MACHINE_CACHE_DIR", "GO_CACHE_DIR", "GO_MOD_CACHE_DIR", "GO_TMP_DIR", "CARTULARY_HARNESS_SCRATCH_ROOT"],
  "machine-state overrides must be declared once as global task-surface inputs",
);
for (const target of manifest.targets.filter((entry) => entry.target_class === "public")) {
  assert.ok(target.input_contract, `${target.name} must declare an input contract`);
  assert.equal(
    target.input_contract.undeclared_make_command_line,
    "usage_error",
    `${target.name} command-line undeclared input policy`,
  );
  assert.equal(
    target.input_contract.undeclared_inherited_env,
    "ignore",
    `${target.name} inherited env undeclared input policy`,
  );
  assert.ok(Array.isArray(target.input_contract.inputs), `${target.name} inputs must be an array`);
}
const planningInputs = manifest.targets.find((target) => target.name === "target-plan").input_contract.inputs;
assert.deepEqual(planningInputs.map((input) => input.name), ["TARGET", "INPUTS", "DETAIL", "JSON"]);
assert.equal(planningInputs.find((input) => input.name === "TARGET").required, true);
assert.deepEqual(
  manifest.targets.find((target) => target.name === "task-guide").input_contract.inputs.map((input) => input.name),
  ["OWNER", "JSON", "DETAIL"],
  "task-guide input contract must include the owner selector",
);
assert.deepEqual(
  manifest.targets.map((target) => target.name).filter((target) => !manifest.make_recipes[target]),
  [],
  "every Make target must have a generated recipe",
);
const targetByName = new Map(manifest.targets.map((target) => [target.name, target]));
for (const recipe of makeRecipeEntries(manifest)) {
  for (const script of requiredRecipeBackingScripts(recipe)) {
    assert.ok(
      targetByName.get(recipe.target)?.backing_scripts?.includes(script),
      `${recipe.target} must declare recipe-required backing script ${script}`,
    );
  }
}
const missingGraphRunnerManifest = structuredClone(manifest);
const releaseInventoryTarget = missingGraphRunnerManifest.targets.find(
  (target) => target.name === "package-release",
);
releaseInventoryTarget.backing_scripts = releaseInventoryTarget.backing_scripts.filter(
  (script) => script !== workGraphRuntimeSource,
);
assert.ok(
  collectTaskSurfaceManifestErrors(missingGraphRunnerManifest).includes(
    `package-release.backing_scripts must declare recipe-required script ${workGraphRuntimeSource}`,
  ),
  "graph-entry recipes must fail semantic validation when their dispatcher is undeclared",
);
assert.equal(manifest.make_recipes.help.type, "print_help", "help must be generated as print_help");
assert.equal(manifest.make_recipes.lint.type, "work_graph", "lint must use the unified graph");
assert.equal(manifest.make_recipes.bootstrap.type, "bootstrap", "bootstrap must dispatch its declared installation selection");
const { WorkGraphCompiler } = await import(pathToFileURL(path.join(root, "tools/harness/scheduler/work-graph/index.mjs")));
const compiler = new WorkGraphCompiler(root);
const migration = compiler.compileTarget("migration-drift");
const selectedRows = migration.units.flatMap((unit) => unit.current_run_evidence_outputs.filter((file) => file.startsWith("rows/")).map((file) => file.slice(5, -5)));
assert.deepEqual(selectedRows.sort(), ["harness.generated_artifacts.support.migration_input_integrity", "module.database_migrations.integration.schema_bootstrap"], "migration drift retains its exact input and TLS fixture rows");
const operationOwner = JSON.parse(readFileSync(path.join(root, "tools/harness_work_graph_owner.json"), "utf8"));
const bootstrap = operationOwner.operations["harness.bootstrap_full"].execution;
assert.equal(bootstrap.type, "installation");
for (const file of ["$(FRONTEND_INSTALL_STAMP)", "$(HARNESS_RUNTIME_BIN)", "$(PLAYWRIGHT_INSTALL_STAMP)"]) assert.ok(bootstrap.prerequisites.includes(file), `${file} remains an admitted bootstrap producer`);
assert.ok(bootstrap.command.includes("tools/harness/readiness/install-go-dependencies.sh"));
const renderedMake = renderTaskSurfaceMake(manifest);
const renderedRuntime = renderTaskSurfaceMakeRuntime(manifest);
assert.match(
  renderedRuntime,
  /^TASK_SURFACE_NODE_READINESS = if \[ "\$\$\{CARTULARY_HARNESS_SKIP_PREREQUISITES:-0\}" != "1" \]; then env -u CARTULARY_TEST_TARGET CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \$\(MAKE\) --silent --no-print-directory \$\(NODE_BIN\); fi$/m,
  "shared Node readiness must retain prerequisite admission and child target isolation",
);
assert.match(
  renderedRuntime,
  /TASK_SURFACE_MACHINE_STATE_ENV = CARTULARY_MACHINE_CACHE_DIR="\$\(CARTULARY_MACHINE_CACHE_DIR\)" GO_CACHE_DIR="\$\(GO_CACHE_DIR\)" GO_MOD_CACHE_DIR="\$\(GO_MOD_CACHE_DIR\)" GO_TMP_DIR="\$\(GO_TMP_DIR\)" CARTULARY_HARNESS_SCRATCH_ROOT="\$\(CARTULARY_HARNESS_SCRATCH_ROOT\)"/,
  "generated runtime must forward the resolved machine-state configuration after stripping public inputs",
);
assert.match(
  renderedMake,
  /test-slice:[\s\S]*?env \$\(TASK_SURFACE_PUBLIC_INPUT_STRIP_ENV\) \$\(TASK_SURFACE_MACHINE_STATE_ENV\)[\s\S]*?\$\(TASK_SURFACE_GRAPH_RUNTIME\) --selection command --target test-slice/,
  "work-graph recipes must restore resolved machine-state paths after stripping public inputs",
);
for (const target of manifest.targets.filter((entry) =>
  entry.target_class === "public" && !["none", "ui_review_receipts"].includes(entry.output_policy.artifact_policy))) {
  const escapedTarget = target.name.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&");
  assert.match(
    renderedMake,
    new RegExp(`^${escapedTarget}: export CARTULARY_TEST_RUN_ID := \\$\\(CARTULARY_TEST_RUN_ID\\)$`, "mu"),
    `${target.name} must freeze its public retained run identity at target scope`,
  );
}
for (const target of manifest.targets.filter((entry) => entry.output_policy?.artifact_policy === "ui_review_receipts")) {
  const recipe = manifest.make_recipes[target.name];
  assert.equal(recipe.type, "ui_review", "private review uses its explicit output projection");
  if (target.name === "ui-review") {
    assert.equal(target.input_contract.inputs.some((input) => input.name === "UI_SESSION"), false, "the launcher creates a locator");
  } else {
    assert.ok(target.input_contract.inputs.some((input) => input.name === "UI_SESSION" && input.required), "finite review commands bind an exact existing locator");
  }
}
assert.match(renderedMake, /CARTULARY_TEST_TARGET="\$\$\{CARTULARY_TEST_TARGET:-services-up\}" \$\(RUN_STEP_SCRIPT\) "services-up"/, "lifecycle capture preserves inherited public identity");
assert.match(
  helpAllLines(manifest).join("\n"),
  /target -> owner partition -> semantic row -> artifact/,
  "help-all must explain the task evidence concept hierarchy",
);
assert.match(
  helpAllLines(manifest).join("\n"),
  /CARTULARY_CLEANUP_DRY_RUN=1/,
  "help-all must advertise cleanup dry-run usage",
);
const helpRecipe = renderedMake.slice(renderedMake.indexOf("\nhelp:"), renderedMake.indexOf("\ntask-guide:"));
assert.doesNotMatch(helpRecipe, /TASK_SURFACE_NODE_READINESS|RUN_PUBLIC_PREFLIGHT|RUN_MAKE_PREREQUISITES/, "help must perform no installation or build preparation");
assert.match(helpRecipe, /TASK_SURFACE_HELP_ALL_SELECTED/, "help uses its declared ALL selector");
assert.match(renderedMake, /--selection command --target migration-drift/, "migration drift resolves its exact command binding");
for (const retired of ["help-all", "frontend-install", "frontend-typecheck", "bootstrap-tool-installations", "deployable-shape"]) assert.doesNotMatch(renderedMake, new RegExp(`^${retired}:`, "m"), "private producers cannot remain public Make wrappers");
const longTarget = "synthetic-help-target-name-longer-than-command-column";
const usageTarget = "synthetic-help-usage";
manifest.targets.push(
  {
    name: longTarget,
    target_class: "public",
    default_inclusion_sets: ["helper_only"],
  },
  {
    name: usageTarget,
    target_class: "public",
    default_inclusion_sets: ["helper_only"],
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

printf '\n.PHONY: unclassified-target\nunclassified-target:\n\t@true\n' >>"$makefile_copy"
unclassified_output="$(assert_fails "unclassified target drift" run_report_copy)"
assert_contains "$unclassified_output" "unclassified-target is missing task-surface target_class" "unclassified target output"

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
  target_class: "public",
  default_inclusion_sets: ["helper_only"]
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
manifest.targets.push({name:"synthetic-private-helper", target_class:"internal_helper", lifecycle_state:"candidate_child", default_inclusion_sets:[]});
compact.entries.push({
  target: "synthetic-private-helper",
  description: "synthetic non-public help tier entry"
});
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
non_public_help_tier_output="$(assert_fails "non-public compact help target" run_report_copy)"
assert_contains "$non_public_help_tier_output" "synthetic-private-helper appears in compact_help but is not target_class public" "non-public compact help output"

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
    target_class: "public",
    default_inclusion_sets: ["helper_only"]
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
const target = manifest.targets.find((entry) => entry.name === "source-check");
target.backing_scripts = ["tools/harness/generated-artifacts/tests/missing-task-surface-helper.mjs"];
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
missing_script_output="$(assert_fails "missing backing script" run_report_copy)"
assert_contains "$missing_script_output" "backing script missing: tools/harness/generated-artifacts/tests/missing-task-surface-helper.mjs" "missing backing script output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
"$NODE_BIN" - "$manifest_copy" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const manifestPath = process.argv[2];
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
const target = manifest.targets.find((entry) => entry.name === "package-release");
target.backing_scripts = target.backing_scripts.filter(
  (script) => script !== "tools/harnessruntime/main.go",
);
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
missing_graph_runner_output="$(assert_fails "missing graph-entry dispatcher" run_report_copy)"
assert_contains "$missing_graph_runner_output" "package-release.backing_scripts must declare recipe-required script tools/harnessruntime/main.go" "missing graph-entry dispatcher semantic output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
"$NODE_BIN" - "$manifest_copy" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const manifestPath = process.argv[2];
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
manifest.harness_checks.push({
  name: "unreachable-harness-check",
  backing_scripts: ["tools/harness/generated-artifacts/tests/test-task-surface-report.sh"]
});
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
unreachable_check_output="$(assert_fails "unreachable harness check" run_report_copy)"
assert_contains "$unreachable_check_output" "harness check unreachable-harness-check is not reachable from any harness tier" "unreachable harness check output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
"$NODE_BIN" - "$manifest_copy" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const manifestPath = process.argv[2];
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
const target = manifest.targets.find((entry) => entry.name === "source-check");
target.backing_scripts = ["scripts/root-task-surface-helper.mjs"];
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
root_script_output="$(assert_fails "root script backing path" run_report_copy)"
assert_contains "$root_script_output" "must not reference retired root scripts/ path scripts/root-task-surface-helper.mjs" "root script backing path output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
cp "$ROOT_DIR/tools/task_surface.generated.mk" "$generated_make_copy"
printf '\n.PHONY: undeclared-script-ref\nundeclared-script-ref:\n\t@node ./tools/harness/readiness/toolchain-pin-check-cli.mjs\n' >>"$makefile_copy"
"$NODE_BIN" - "$manifest_copy" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const manifestPath = process.argv[2];
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
manifest.targets.push({
  name: "undeclared-script-ref",
  target_class: "internal_helper",
  default_inclusion_sets: [],
  lifecycle_state: "candidate_child",
  backing_scripts: []
});
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
undeclared_script_output="$(assert_fails "undeclared script reference" run_report_copy)"
assert_contains "$undeclared_script_output" "references tools/harness/readiness/toolchain-pin-check-cli.mjs" "undeclared script reference output"

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
