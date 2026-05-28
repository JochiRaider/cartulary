#!/usr/bin/env bash
# Single-quoted literals below intentionally assert Make/shell text without expansion.
# shellcheck disable=SC2016
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=scripts/lib/task-surface-check-common.sh
# shellcheck disable=SC1091
source "$repo_root/scripts/lib/task-surface-check-common.sh"

makefile="$repo_root/Makefile"
generated_make="$repo_root/tools/task_surface.generated.mk"
check_schedule_manifest="$repo_root/tools/scheduler_manifest.json"
execution_topology_manifest="$repo_root/tools/execution_topology_manifest.json"
runner_script="$repo_root/scripts/run-frontend-unit.sh"
frontend_biome_script="$repo_root/scripts/run-frontend-biome.sh"
frontend_import_boundary_script="$repo_root/scripts/check-frontend-import-boundaries.mjs"
frontend_import_boundary_config="$repo_root/tools/frontend_import_boundaries.json"
scripts_biome_script="$repo_root/scripts/run-scripts-biome.sh"
node_bin="${NODE_BIN:-node}"

assert_target_recipe_invokes() {
  local target="$1"
  local invoked_target="$2"
  local message="$3"
  local block

  block="$(extract_target_block "$target")"
  if [[ -z "$block" ]]; then
    fail "Makefile must define a non-empty $target block"
  fi
  assert_text_has_token "$target recipe" "$block" "$invoked_target" "$message"
}

assert_text_order() {
  local label="$1"
  local text="$2"
  local first="$3"
  local second="$4"
  local message="$5"

  if ! printf '%s\n' "$text" | awk -v first="$first" -v second="$second" '
    index($0, first) && first_line == 0 { first_line = NR }
    index($0, second) && second_line == 0 { second_line = NR }
    END { exit first_line > 0 && second_line > first_line ? 0 : 1 }
  '; then
    fail "$message (${label} must order ${first} before ${second})"
  fi
}

assert_target_exists frontend-import-boundary-check "Makefile must define frontend-import-boundary-check"
if [[ ! -f "$frontend_import_boundary_script" ]]; then
  fail "missing scripts/check-frontend-import-boundaries.mjs"
fi

frontend_unit_block="$(extract_target_block frontend-unit)"
if [[ -z "$frontend_unit_block" ]]; then
  fail "Makefile must define a non-empty frontend-unit block"
fi
if ! text_contains "$frontend_unit_block" './scripts/run-frontend-unit.sh'; then
  fail "frontend-unit must delegate to scripts/run-frontend-unit.sh"
fi

assert_target_exists toolchain-drift "Makefile must define toolchain-drift"
assert_target_prereq codegen-toolchain '$(SQLC_BIN)' "codegen-toolchain must own pinned SQLC_BIN readiness"
assert_target_prereq go-lint-toolchain '$(STATICCHECK_BIN)' "go-lint-toolchain must own pinned Staticcheck readiness"
assert_target_prereq govulncheck-toolchain '$(GOVULNCHECK_BIN)' "govulncheck-toolchain must own pinned Govulncheck readiness"
assert_target_prereq gosec-toolchain '$(GOSEC_BIN)' "gosec-toolchain must own pinned Gosec readiness"
assert_target_prereq shell-lint-toolchain '$(SHELLCHECK_BIN)' "shell-lint-toolchain must own pinned ShellCheck readiness"
generate_block="$(extract_target_block generate)"
assert_text_contains "generate recipe" "$generate_block" "codegen-toolchain" "generate must prepare the codegen toolchain before generating artifacts"
assert_text_order "generate recipe" "$generate_block" "codegen-toolchain" "generate-artifacts" "generate must prepare the codegen toolchain before generating artifacts"
assert_target_recipe_invokes generate generate-artifacts "generate must delegate to generate-artifacts"
generate_artifacts_block="$(extract_target_block generate-artifacts)"
if [[ -z "$generate_artifacts_block" ]]; then
  fail "Makefile must define a non-empty generate-artifacts block"
fi
if ! text_contains "$generate_artifacts_block" './scripts/generate-artifacts.sh'; then
  fail "generate-artifacts must delegate to scripts/generate-artifacts.sh"
fi
if ! grep -Fq '"generate sqlc"' "$repo_root/scripts/generate-artifacts.sh"; then
  fail "scripts/generate-artifacts.sh must run sqlc generation"
fi
generate_drift_block="$(extract_target_block generate-drift)"
assert_text_contains "generate-drift recipe" "$generate_drift_block" "codegen-toolchain" "generate-drift must prepare the codegen toolchain outside the drift body"
assert_text_order "generate-drift recipe" "$generate_drift_block" "codegen-toolchain" "./scripts/check-generate-drift.sh" "generate-drift must prepare the codegen toolchain outside the drift body"
assert_target_absent check-preflight "check-preflight must not remain as a legacy alias; use scheduler-visible readiness targets"
assert_target_absent check-setup-blockers "check-setup-blockers must not remain after setup readiness fanout"
if [[ -e "$repo_root/scripts/check-setup-blockers.sh" ]]; then
  fail "scripts/check-setup-blockers.sh must not remain as a serial setup wrapper"
fi
assert_target_exists check-frontend-install "Makefile must define check-frontend-install"
assert_target_prereq check-frontend-install frontend-install "check-frontend-install must use the unified frontend install target"
check_prereqs="$(extract_target_prereqs check)"
if text_matches "$check_prereqs" 'FRONTEND_INSTALL_STAMP'; then
  fail "check must not depend directly on FRONTEND_INSTALL_STAMP"
fi
"$node_bin" - "$check_schedule_manifest" "$execution_topology_manifest" <<'EOF'
const fs = require("node:fs");
const [manifestFile, topologyFile] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
const topology = JSON.parse(fs.readFileSync(topologyFile, "utf8"));
if (topology.schema_id !== "cartulary.execution_topology.v3") {
  throw new Error("execution topology must declare schema_id=cartulary.execution_topology.v3");
}
if (Array.isArray(topology.check_schedules)) {
  throw new Error("execution topology must own check schedule profiles, not flat schedules");
}
const topologyTargets = new Map((topology.task_surface?.targets ?? []).map((entry) => [entry.name, entry]));
for (const required of ["frontend-typecheck", "frontend-unit", "frontend-import-boundary-check", "lint-biome", "lint-scripts", "lint-shell"]) {
  const metadata = topologyTargets.get(required)?.check_schedule;
  if (!metadata?.schedules?.includes("check")) {
    throw new Error(`execution topology must schedule ${required} through check_schedule metadata`);
  }
}
const schedule = (manifest.schedules ?? []).find((entry) => entry.target === "check");
if (!schedule) {
  throw new Error("missing check schedule");
}
const targets = new Set((schedule.work_units ?? []).map((entry) => entry.target));
const unitByTarget = new Map((schedule.work_units ?? []).map((entry) => [entry.target, entry]));
for (const removed of ["check-setup-blockers"]) {
  if (targets.has(removed)) {
    throw new Error(`${removed} must not remain scheduled after setup readiness fanout`);
  }
}
for (const removed of ["check-static-validation", "check-local-product", "check-meta-validation"]) {
  if (targets.has(removed)) {
    throw new Error(`${removed} must not remain scheduled after leaf check expansion`);
  }
}
for (const required of ["toolchain-drift", "check-frontend-install", "frontend-typecheck", "frontend-unit", "frontend-import-boundary-check", "lint-biome", "harness-contract-tests", "lint-scripts", "lint-shell", "check-harness-smoke"]) {
  if (!targets.has(required)) {
    throw new Error(`check schedule must include ${required}`);
  }
}
for (const target of ["frontend-typecheck", "frontend-unit", "frontend-import-boundary-check", "lint-biome", "lint-scripts"]) {
  const needs = unitByTarget.get(target)?.needs ?? [];
  if (needs.join(",") !== "check-frontend-install") {
    throw new Error(`${target} must depend on check-frontend-install`);
  }
}
const frontendUnit = unitByTarget.get("frontend-unit");
if (JSON.stringify(frontendUnit?.env ?? {}) !== JSON.stringify({ VITEST_MAX_WORKERS: "2" })) {
  throw new Error("scheduled frontend-unit must pin VITEST_MAX_WORKERS=2");
}
if (JSON.stringify(frontendUnit?.resource_claims ?? {}) !== JSON.stringify({ host_cpu: 2 })) {
  throw new Error("scheduled frontend-unit must claim exactly host_cpu=2");
}
if ((unitByTarget.get("check-frontend-install")?.needs ?? []).join(",") !== "toolchain-drift") {
  throw new Error("check-frontend-install must depend on toolchain-drift");
}
EOF
assert_target_exists phase-ledgers "Makefile must define phase-ledgers"
assert_target_exists phase-ledger-drift "Makefile must define phase-ledger-drift"

if ! [[ -f "$runner_script" ]]; then
  fail "missing scripts/run-frontend-unit.sh"
fi

frontend_typecheck_block="$(extract_target_block frontend-typecheck)"
if [[ -z "$frontend_typecheck_block" ]]; then
  fail "Makefile must define a non-empty frontend-typecheck block"
fi
assert_target_exports_self frontend-typecheck "frontend-typecheck must export CARTULARY_TEST_TARGET"
if ! text_contains "$frontend_typecheck_block" '$(PNPM) typecheck'; then
  fail "frontend-typecheck must run the root workspace TypeScript typecheck script"
fi
if ! [[ -f "$repo_root/tsconfig.json" ]]; then
  fail "missing root TypeScript solution config"
fi
if ! [[ -f "$repo_root/tsconfig.base.json" ]]; then
  fail "missing root TypeScript base config"
fi
if ! [[ -f "$repo_root/apps/web/tsconfig.e2e.json" ]]; then
  fail "missing apps/web e2e TypeScript config"
fi
"$node_bin" - "$repo_root/package.json" "$repo_root/tsconfig.json" "$repo_root/tsconfig.base.json" "$repo_root/apps/web/tsconfig.e2e.json" <<'EOF'
const fs = require("node:fs");
const [packagePath, rootConfigPath, baseConfigPath, e2eConfigPath] = process.argv.slice(2);
const packageJson = JSON.parse(fs.readFileSync(packagePath, "utf8"));
if (packageJson.scripts?.typecheck !== "tsc -b tsconfig.json --noEmit --incremental false --pretty false") {
  throw new Error("root package.json must expose the canonical workspace typecheck script");
}
const rootConfig = JSON.parse(fs.readFileSync(rootConfigPath, "utf8"));
const references = new Set((rootConfig.references ?? []).map((entry) => entry.path));
for (const required of [
  "./apps/web",
  "./apps/web/tsconfig.e2e.json",
  "./packages/grid-adapter",
  "./packages/protocol-ts",
  "./packages/test-utils",
  "./packages/ui-contracts",
  "./packages/view-contracts",
]) {
  if (!references.has(required)) {
    throw new Error(`root tsconfig.json references are missing ${required}`);
  }
}
const baseConfig = JSON.parse(fs.readFileSync(baseConfigPath, "utf8"));
for (const required of [
  "noImplicitReturns",
  "noUnusedLocals",
  "noUnusedParameters",
  "verbatimModuleSyntax",
]) {
  if (baseConfig.compilerOptions?.[required] !== true) {
    throw new Error(`tsconfig.base.json compilerOptions.${required} must be true`);
  }
}
const e2eConfig = JSON.parse(fs.readFileSync(e2eConfigPath, "utf8"));
const includes = new Set(e2eConfig.include ?? []);
for (const required of [
  "playwright.config.ts",
  "playwright.shared.config.ts",
  "playwright.webserver-backed.config.ts",
  "e2e/**/*.ts",
]) {
  if (!includes.has(required)) {
    throw new Error(`apps/web/tsconfig.e2e.json include is missing ${required}`);
  }
}
const types = new Set(e2eConfig.compilerOptions?.types ?? []);
for (const required of ["node", "@playwright/test"]) {
  if (!types.has(required)) {
    throw new Error(`apps/web/tsconfig.e2e.json compilerOptions.types is missing ${required}`);
  }
}
EOF
if text_contains "$frontend_typecheck_block" '--dir apps/web exec tsc'; then
  fail "frontend-typecheck must not remain app-only"
fi
if ! text_contains "$frontend_typecheck_block" '$(call RUN_TARGET_SUMMARY,frontend-typecheck,pass)'; then
  fail "frontend-typecheck must emit a target summary"
fi

assert_target_absent lint-typecheck "lint-typecheck must not remain as a legacy alias; use frontend-typecheck"

format_block="$(extract_target_block format)"
assert_text_contains "format recipe" "$format_block" "format-go" "format must delegate to format-go and format-frontend"
assert_text_contains "format recipe" "$format_block" "format-frontend" "format must delegate to format-go and format-frontend"
format_go_block="$(extract_target_block format-go)"
if ! text_contains "$format_go_block" 'scripts/run-go-format.sh --write'; then
  fail "format-go must run the curated Go formatter wrapper in write mode"
fi
assert_target_prereq format-frontend '$(NODE_BIN)' "format-frontend must depend on NODE_BIN"
assert_target_prereq format-frontend '$(FRONTEND_INSTALL_STAMP)' "format-frontend must depend on FRONTEND_INSTALL_STAMP"
format_frontend_block="$(extract_target_block format-frontend)"
if ! text_contains "$format_frontend_block" './scripts/run-frontend-biome.sh format'; then
  fail "format-frontend must run the curated frontend Biome formatter"
fi
lint_biome_block="$(extract_target_block lint-biome)"
if ! text_contains "$lint_biome_block" 'inspect Biome diagnostics; run make format only for formatting/style diagnostics'; then
  fail "lint-biome must tell developers to inspect diagnostics before using make format"
fi
frontend_import_boundary_block="$(extract_target_block frontend-import-boundary-check)"
if [[ -z "$frontend_import_boundary_block" ]]; then
  fail "Makefile must define a non-empty frontend-import-boundary-check block"
fi
if ! text_contains "$frontend_import_boundary_block" './scripts/check-frontend-import-boundaries.mjs'; then
  fail "frontend-import-boundary-check must run the repo-local import boundary checker"
fi
assert_text_contains "frontend-import-boundary-check recipe" "$frontend_import_boundary_block" '$(NODE_BIN)' "frontend-import-boundary-check must depend on NODE_BIN"
assert_text_contains "frontend-import-boundary-check recipe" "$frontend_import_boundary_block" '$(FRONTEND_INSTALL_STAMP)' "frontend-import-boundary-check must depend on frontend install"
lint_block="$(extract_target_block lint)"
assert_text_contains "lint recipe" "$lint_block" "frontend-import-boundary-check" "lint must include frontend-import-boundary-check"
assert_text_contains "lint recipe" "$lint_block" "lint-shell" "lint must include lint-shell"
if ! grep -Fq 'exec biome check --error-on-warnings' "$frontend_biome_script"; then
  fail "frontend Biome check mode must fail on warnings"
fi
if ! grep -Fq -- '--config-path "${ROOT_DIR}/biome.json"' "$frontend_biome_script"; then
  fail "frontend Biome wrapper must use the repo root Biome config explicitly"
fi
if ! grep -Fq -- '--vcs-root "${ROOT_DIR}"' "$frontend_biome_script"; then
  fail "frontend Biome wrapper must set the repo VCS root explicitly"
fi
lint_scripts_block="$(extract_target_block lint-scripts)"
if ! text_contains "$lint_scripts_block" './scripts/run-scripts-biome.sh'; then
  fail "lint-scripts must run the curated scripts Biome wrapper"
fi
lint_shell_block="$(extract_target_block lint-shell)"
if ! text_contains "$lint_shell_block" 'scripts/run-shellcheck.sh'; then
  fail "lint-shell must run the curated ShellCheck wrapper"
fi
if ! text_contains "$lint_shell_block" 'LINT_SHELL_STRICT="$(LINT_SHELL_STRICT)"'; then
  fail "lint-shell must expose strict-mode passthrough"
fi
if ! grep -Fq -- '--error-on-warnings' "$scripts_biome_script"; then
  fail "scripts Biome check mode must fail on warnings"
fi
if ! grep -Fq -- '--config-path "${ROOT_DIR}/biome.json"' "$scripts_biome_script"; then
  fail "scripts Biome wrapper must use the repo root Biome config explicitly"
fi
if ! grep -Fq -- '--vcs-root "${ROOT_DIR}"' "$scripts_biome_script"; then
  fail "scripts Biome wrapper must set the repo VCS root explicitly"
fi
if ! grep -Fq -- '"packages/ui-contracts/src"' "$frontend_biome_script"; then
  fail "frontend Biome wrapper must include packages/ui-contracts/src"
fi
"$node_bin" - "$frontend_import_boundary_config" <<'EOF'
const fs = require("node:fs");
const [configPath] = process.argv.slice(2);
const config = JSON.parse(fs.readFileSync(configPath, "utf8"));
if (config.schema_id !== "cartulary.frontend_import_boundaries.v2") {
  throw new Error("frontend import boundary config must declare schema_id=cartulary.frontend_import_boundaries.v2");
}
const scanRoots = new Set(config.scan_roots ?? []);
for (const required of ["apps/web/src", "apps/web/e2e", "packages/grid-adapter/src", "packages/protocol-ts/src", "packages/ui-contracts/src"]) {
  if (!scanRoots.has(required)) {
    throw new Error(`frontend import boundary scan roots missing ${required}`);
  }
}
const rules = new Map((config.rules ?? []).map((rule) => [rule.id, rule]));
const gridRule = rules.get("frontend-grid-vendor-boundary");
if (!gridRule || gridRule.level !== "error") {
  throw new Error("frontend-grid-vendor-boundary must be enforced as an error");
}
if (!(gridRule.allowed_importers ?? []).includes("packages/grid-adapter/src/**")) {
  throw new Error("frontend-grid-vendor-boundary must allow only packages/grid-adapter/src/**");
}
if (!JSON.stringify(gridRule.restricted_imports ?? []).includes('"react-data-grid"')) {
  throw new Error("frontend-grid-vendor-boundary must restrict react-data-grid");
}
const generatedRule = rules.get("frontend-generated-protocol-boundary");
if (!generatedRule || generatedRule.level !== "error") {
  throw new Error("frontend-generated-protocol-boundary must be enforced as an error");
}
if (!(generatedRule.allowed_importers ?? []).includes("packages/protocol-ts/src/index.ts")) {
  throw new Error("frontend-generated-protocol-boundary must allow the protocol-ts facade");
}
const generatedRestrictions = JSON.stringify(generatedRule.restricted_imports ?? []);
for (const required of ["@cartulary/protocol-ts/generated", "packages/protocol-ts/src/generated"]) {
  if (!generatedRestrictions.includes(required)) {
    throw new Error(`frontend-generated-protocol-boundary must restrict ${required}`);
  }
}
const nodeRule = rules.get("frontend-runtime-node-boundary");
if (!nodeRule || nodeRule.level !== "error") {
  throw new Error("frontend-runtime-node-boundary must be enforced as an error");
}
if (!JSON.stringify(nodeRule.restricted_imports ?? []).includes('"node_builtin"')) {
  throw new Error("frontend-runtime-node-boundary must restrict Node builtins");
}
const workspaceRule = rules.get("frontend-workspace-package-facade-boundary");
if (!workspaceRule || workspaceRule.level !== "error") {
  throw new Error("frontend-workspace-package-facade-boundary must be enforced as an error");
}
if (!JSON.stringify(workspaceRule.restricted_imports ?? []).includes('"workspace_package_facade"')) {
  throw new Error("frontend-workspace-package-facade-boundary must restrict workspace package facade bypasses");
}
const testHelperRule = rules.get("frontend-runtime-test-helper-boundary");
if (!testHelperRule || testHelperRule.level !== "error") {
  throw new Error("frontend-runtime-test-helper-boundary must be enforced as an error");
}
const testHelperRestrictions = JSON.stringify(testHelperRule.restricted_imports ?? []);
for (const required of ["@cartulary/test-utils", "@cartulary/grid-adapter/test-support", "vitest", "@testing-library/react", "@playwright/test"]) {
  if (!testHelperRestrictions.includes(required)) {
    throw new Error(`frontend-runtime-test-helper-boundary must restrict ${required}`);
  }
}
EOF
"$node_bin" - "$repo_root/biome.json" <<'EOF'
const fs = require("node:fs");
const [configPath] = process.argv.slice(2);
const config = JSON.parse(fs.readFileSync(configPath, "utf8"));
const includes = config.files?.includes;
if (!Array.isArray(includes)) {
  throw new Error("biome.json must define files.includes for the curated Biome ownership scope");
}
if (
  config.vcs?.enabled !== true ||
  config.vcs?.clientKind !== "git" ||
  config.vcs?.useIgnoreFile !== true ||
  config.vcs?.root !== "."
) {
  throw new Error("biome.json must enable Git ignore handling from the repo root");
}
const requiredIncludes = [
  "apps/web/src/**",
  "apps/web/e2e/**",
  "apps/web/vite.config.ts",
  "apps/web/playwright.config.ts",
  "apps/web/playwright.shared.config.ts",
  "apps/web/playwright.webserver-backed.config.ts",
  "packages/grid-adapter/src/**",
  "packages/ui-contracts/src/**",
  "packages/view-contracts/src/**",
  "packages/test-utils/src/**",
  "packages/protocol-ts/src/**",
  "scripts/**/*.mjs",
];
for (const required of requiredIncludes) {
  if (!includes.includes(required)) {
    throw new Error(`biome.json curated scope is missing ${required}`);
  }
}
if (!includes.includes("!packages/protocol-ts/src/generated/**")) {
  throw new Error("biome.json must exclude generated protocol TypeScript from the curated scope");
}
const requiredExcludes = [
  "!tmp/**",
  "!.cache/**",
  "!.pnpm-store/**",
  "!node_modules/**",
  "!apps/web/dist/**",
  "!apps/web/test-results/**",
  "!apps/web/playwright-report/**",
  "!apps/web/coverage/**",
  "!coverage/**",
  "!playwright-report/**",
  "!test-results/**",
  "!dist/**",
  "!build/**",
  "!out/**",
];
for (const excluded of requiredExcludes) {
  if (!includes.includes(excluded)) {
    throw new Error(`biome.json curated scope must exclude ${excluded}`);
  }
}
const rules = config.linter?.rules;
if (rules?.recommended !== true) {
  throw new Error("biome.json must keep recommended Biome rules enabled");
}
const expectedRuleLevels = [
  ["correctness", "noUndeclaredDependencies", "error"],
  ["nursery", "noFloatingPromises", "error"],
  ["nursery", "noMisusedPromises", "error"],
  ["suspicious", "noImportCycles", "error"],
];
for (const [group, name, level] of expectedRuleLevels) {
  const rule = rules?.[group]?.[name];
  const actualLevel = typeof rule === "string" ? rule : rule?.level;
  if (actualLevel !== level) {
    throw new Error(`biome.json must enable ${group}.${name} at ${level} level`);
  }
}
if (rules?.suspicious?.noImportCycles?.options?.ignoreTypes !== true) {
  throw new Error("biome.json noImportCycles must ignore type-only imports");
}
EOF

test_fast_block="$(extract_target_block test-fast)"
if [[ -z "$test_fast_block" ]]; then
  fail "Makefile must define a non-empty test-fast block"
fi
if ! text_contains "$test_fast_block" '$(RUN_MAKE_SEQUENCE_SCRIPT) --sequence test-fast'; then
  fail "test-fast must use the manifest-backed sequence runner"
fi
assert_target_prereq test-local backend-unit "test-fast must route local checks through test-local, and test-local must include backend-unit"
assert_target_prereq test-local frontend-typecheck "test-fast must route local frontend checks through test-local, and test-local must include frontend-typecheck"
assert_target_prereq test-local frontend-unit "test-fast must route local frontend checks through test-local, and test-local must include frontend-unit"
"$node_bin" - "$repo_root/tools/task_surface_manifest.json" <<'EOF'
const fs = require("node:fs");
const [manifestPath] = process.argv.slice(2);
const sequence = JSON.parse(fs.readFileSync(manifestPath, "utf8")).sequences?.["test-fast"];
const stepTargets = (sequence?.steps ?? []).map((step) => step.target);
for (const target of ["test-local", "test-fast-service-backed"]) {
  if (!stepTargets.includes(target)) {
    console.error(`test-fast sequence must invoke ${target}`);
    process.exit(1);
  }
}
EOF

mapfile -t manifest_phases < <("$node_bin" - "$repo_root" <<'EOF'
const fs = require("fs");
const path = require("path");

const root = process.argv[2];
const sections = ["unit", "integration", "e2e", "visual"];
const phaseSchemaID = "cartulary.phase_test_map.v1";
const registrySchemaID = "cartulary.phase_registry.v1";
const phases = [];
const registry = JSON.parse(fs.readFileSync(path.join(root, "tools", "phase_registry.json"), "utf8"));
if (registry.schema_id !== registrySchemaID) {
  console.error(`tools/phase_registry.json must declare schema_id ${registrySchemaID}`);
  process.exit(1);
}

for (const phaseEntry of (registry.phases ?? [])
  .filter((entry) => entry.status === "active")
  .sort((left, right) => left.order - right.order || left.phase.localeCompare(right.phase))) {
  const entry = phaseEntry.manifest_path;
  const manifest = JSON.parse(
    fs.readFileSync(path.join(root, entry), "utf8"),
  );
  if (manifest.schema_id !== phaseSchemaID) {
    console.error(`${entry} must declare schema_id ${phaseSchemaID}`);
    process.exit(1);
  }
  if (manifest.phase !== phaseEntry.phase) {
    console.error(`${entry} must declare phase ${phaseEntry.phase}`);
    process.exit(1);
  }
  const phase = manifest.phase;
  let ownsFrontendUnit = false;
  for (const section of sections) {
    for (const row of manifest[section] ?? []) {
      if (row.coverage !== "authoritative" || row.runner !== "vitest") {
        continue;
      }
      ownsFrontendUnit = true;
      if (row.execution_dependency !== "frontend_unit") {
        console.error(
          `${phase} authoritative vitest row ${row.id} must declare execution_dependency=frontend_unit`,
        );
        process.exit(1);
      }
    }
  }
  if (ownsFrontendUnit) {
    phases.push(phase);
  }
}

for (const phase of phases) {
  process.stdout.write(`${phase}\n`);
}
EOF
)

if [[ "${#manifest_phases[@]}" -eq 0 ]]; then
  fail "expected at least one phase to own authoritative frontend-unit vitest rows"
fi

if ! grep -Fq 'vitest-phases authoritative frontend_unit' "$runner_script"; then
  fail "scripts/run-frontend-unit.sh must discover frontend-unit Vitest phases from the manifest"
fi
if ! grep -Fq 'frontend-unit ${manifest_phase} authoritative' "$runner_script"; then
  fail "scripts/run-frontend-unit.sh must emit one manifest summary per discovered frontend-unit phase"
fi
