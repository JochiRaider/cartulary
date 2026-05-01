#!/usr/bin/env bash
# Single-quoted literals below intentionally assert Make/shell text without expansion.
# shellcheck disable=SC2016
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
makefile="$repo_root/Makefile"
generated_make="$repo_root/tools/task_surface.generated.mk"
check_schedule_manifest="$repo_root/tools/check_schedule_manifest.json"
runner_script="$repo_root/scripts/run-frontend-unit.sh"
frontend_biome_script="$repo_root/scripts/run-frontend-biome.sh"
frontend_import_boundary_script="$repo_root/scripts/check-frontend-import-boundaries.mjs"
frontend_import_boundary_config="$repo_root/tools/frontend_import_boundaries.json"
scripts_biome_script="$repo_root/scripts/run-scripts-biome.sh"
node_bin="${NODE_BIN:-node}"

fail() {
  echo "$*" >&2
  exit 1
}

extract_target_block() {
  local target="$1"
  awk -v target="$target" '
    $0 ~ "^" target ":[[:space:]]+export[[:space:]]" { next }
    $0 ~ "^" target ":" { in_block=1; next }
    in_block && /^[^[:space:]].*:/ { exit }
    in_block { print }
  ' "$generated_make" "$makefile"
}

extract_target_prereqs() {
  local target="$1"
  awk -v target="$target" '
    $0 ~ "^" target ":" && $0 !~ "^" target ":[[:space:]]+export[[:space:]]" {
      sub("^" target ":[[:space:]]*", "", $0)
      print
      exit
    }
  ' "$generated_make" "$makefile"
}

assert_text_has_token() {
  local label="$1"
  local text="$2"
  local token="$3"
  local message="$4"

  if ! printf '%s\n' "$text" | awk -v token="$token" '
    {
      for (i = 1; i <= NF; i++) {
        if ($i == token) {
          found = 1
        }
      }
    }
    END { exit found ? 0 : 1 }
  '; then
    fail "${message} (${label} missing ${token})"
  fi
}

assert_target_prereq() {
  local target="$1"
  local prereq="$2"
  local message="$3"
  local prereqs

  prereqs="$(extract_target_prereqs "$target")"
  if [[ -z "$prereqs" ]]; then
    fail "Makefile must define non-empty $target prerequisites"
  fi
  assert_text_has_token "$target prerequisites" "$prereqs" "$prereq" "$message"
}

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

if ! rg -q '^frontend-task-surface-check:' "$generated_make" "$makefile"; then
  fail "Makefile must define frontend-task-surface-check"
fi
if ! rg -q '^frontend-import-boundary-check:' "$generated_make" "$makefile"; then
  fail "Makefile must define frontend-import-boundary-check"
fi
if [[ ! -f "$frontend_import_boundary_script" ]]; then
  fail "missing scripts/check-frontend-import-boundaries.mjs"
fi

frontend_unit_block="$(extract_target_block frontend-unit)"
if [[ -z "$frontend_unit_block" ]]; then
  fail "Makefile must define a non-empty frontend-unit block"
fi
if ! printf '%s\n' "$frontend_unit_block" | grep -Fq './scripts/run-frontend-unit.sh'; then
  fail "frontend-unit must delegate to scripts/run-frontend-unit.sh"
fi

if ! rg -q '^toolchain-drift:' "$makefile"; then
  fail "Makefile must define toolchain-drift"
fi
assert_target_prereq codegen-toolchain '$(SQLC_BIN)' "codegen-toolchain must own pinned SQLC_BIN readiness"
assert_target_prereq go-lint-toolchain '$(STATICCHECK_BIN)' "go-lint-toolchain must own pinned Staticcheck readiness"
assert_target_prereq shell-lint-toolchain '$(SHELLCHECK_BIN)' "shell-lint-toolchain must own pinned ShellCheck readiness"
assert_target_prereq generate codegen-toolchain "generate must prepare the codegen toolchain before generating artifacts"
assert_target_recipe_invokes generate generate-artifacts "generate must delegate to generate-artifacts"
generate_artifacts_block="$(extract_target_block generate-artifacts)"
if [[ -z "$generate_artifacts_block" ]]; then
  fail "Makefile must define a non-empty generate-artifacts block"
fi
if ! printf '%s\n' "$generate_artifacts_block" | rg -q 'generate sqlc'; then
  fail "generate-artifacts must run sqlc generation"
fi
assert_target_prereq generate-drift codegen-toolchain "generate-drift must prepare the codegen toolchain outside the drift body"
if rg -q '^check-preflight:' "$makefile"; then
  fail "check-preflight must not remain as a legacy alias; use check-setup-blockers"
fi
check_setup_block="$(extract_target_block check-setup-blockers)"
if [[ -z "$check_setup_block" ]]; then
  fail "Makefile must define a non-empty check-setup-blockers block"
fi
if ! printf '%s\n' "$check_setup_block" | rg -q 'toolchain-drift'; then
  fail "check-setup-blockers must invoke toolchain-drift"
fi
if ! printf '%s\n' "$check_setup_block" | rg -q 'codegen-toolchain'; then
  fail "check-setup-blockers must prepare the codegen toolchain after toolchain drift"
fi
if ! printf '%s\n' "$check_setup_block" | rg -q 'go-lint-toolchain'; then
  fail "check-setup-blockers must prepare the Go lint toolchain after codegen readiness"
fi
if ! printf '%s\n' "$check_setup_block" | rg -q 'shell-lint-toolchain'; then
  fail "check-setup-blockers must prepare ShellCheck after Go lint tooling"
fi
if ! printf '%s\n' "$check_setup_block" | rg -q 'frontend-install'; then
  fail "check-setup-blockers must invoke frontend install after toolchain drift"
fi
assert_text_order "check-setup-blockers recipe" "$check_setup_block" "toolchain-drift" "codegen-toolchain" "check-setup-blockers must prepare codegen after toolchain drift"
assert_text_order "check-setup-blockers recipe" "$check_setup_block" "codegen-toolchain" "go-lint-toolchain" "check-setup-blockers must prepare Go lint tooling after codegen readiness"
assert_text_order "check-setup-blockers recipe" "$check_setup_block" "go-lint-toolchain" "shell-lint-toolchain" "check-setup-blockers must prepare ShellCheck after Go lint tooling"
assert_text_order "check-setup-blockers recipe" "$check_setup_block" "go-lint-toolchain" "frontend-install" "check-setup-blockers must install frontend dependencies after Go lint tooling readiness"
if printf '%s\n' "$check_setup_block" | rg -q 'frontend-task-surface-check|frontend-import-boundary-check|phase-ledger-drift|run-phase-smoke|generate-drift|lint-biome|lint-scripts'; then
  fail "check-setup-blockers must not include static validation or harness smoke work"
fi
check_prereqs="$(extract_target_prereqs check)"
if printf '%s\n' "$check_prereqs" | rg -q 'FRONTEND_INSTALL_STAMP'; then
  fail "check must not depend directly on FRONTEND_INSTALL_STAMP"
fi
"$node_bin" - "$check_schedule_manifest" <<'EOF'
const fs = require("node:fs");
const [manifestFile] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
const schedule = (manifest.schedules ?? []).find((entry) => entry.target === "check");
if (!schedule) {
  throw new Error("missing check schedule");
}
const targets = new Set((schedule.work_units ?? []).map((entry) => entry.target));
for (const removed of ["check-static-validation", "check-local-product", "check-meta-validation"]) {
  if (targets.has(removed)) {
    throw new Error(`${removed} must not remain scheduled after leaf check expansion`);
  }
}
for (const required of ["frontend-typecheck", "frontend-task-surface-check", "frontend-import-boundary-check", "lint-biome", "lint-scripts", "lint-shell", "check-harness-smoke"]) {
  if (!targets.has(required)) {
    throw new Error(`check schedule must include ${required}`);
  }
}
EOF
if ! rg -q '^phase-ledgers:' "$makefile"; then
  fail "Makefile must define phase-ledgers"
fi
if ! rg -q '^phase-ledger-drift:' "$makefile"; then
  fail "Makefile must define phase-ledger-drift"
fi

if ! [[ -f "$runner_script" ]]; then
  fail "missing scripts/run-frontend-unit.sh"
fi

frontend_typecheck_block="$(extract_target_block frontend-typecheck)"
if [[ -z "$frontend_typecheck_block" ]]; then
  fail "Makefile must define a non-empty frontend-typecheck block"
fi
if ! rg -q '^frontend-typecheck:[[:space:]]+export CARTULARY_TEST_TARGET := frontend-typecheck$' "$makefile"; then
  fail "frontend-typecheck must export CARTULARY_TEST_TARGET"
fi
if ! printf '%s\n' "$frontend_typecheck_block" | grep -Fq '$(PNPM) typecheck'; then
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
if printf '%s\n' "$frontend_typecheck_block" | grep -Fq -- '--dir apps/web exec tsc'; then
  fail "frontend-typecheck must not remain app-only"
fi
if ! printf '%s\n' "$frontend_typecheck_block" | grep -Fq '$(TARGET_SUMMARY) frontend-typecheck pass'; then
  fail "frontend-typecheck must emit a target summary"
fi

if rg -q '^lint-typecheck:' "$makefile"; then
  fail "lint-typecheck must not remain as a legacy alias; use frontend-typecheck"
fi

if ! rg -q '^format:[[:space:]]+format-go[[:space:]]+format-frontend$$' "$makefile"; then
  fail "format must delegate to format-go and format-frontend"
fi
format_go_block="$(extract_target_block format-go)"
if ! printf '%s\n' "$format_go_block" | grep -Fq 'scripts/run-go-format.sh --write'; then
  fail "format-go must run the curated Go formatter wrapper in write mode"
fi
format_frontend_prereqs="$(extract_target_prereqs format-frontend)"
if ! printf '%s\n' "$format_frontend_prereqs" | rg -q '(^|[[:space:]])\$\(NODE_BIN\)($|[[:space:]])'; then
  fail "format-frontend must depend on NODE_BIN"
fi
if ! printf '%s\n' "$format_frontend_prereqs" | rg -q '(^|[[:space:]])\$\(FRONTEND_INSTALL_STAMP\)($|[[:space:]])'; then
  fail "format-frontend must depend on FRONTEND_INSTALL_STAMP"
fi
format_frontend_block="$(extract_target_block format-frontend)"
if ! printf '%s\n' "$format_frontend_block" | grep -Fq '$(RUN_FRONTEND_BIOME_SCRIPT) format'; then
  fail "format-frontend must run the curated frontend Biome formatter"
fi
lint_biome_block="$(extract_target_block lint-biome)"
if ! printf '%s\n' "$lint_biome_block" | grep -Fq 'inspect Biome diagnostics; run make format only for formatting/style diagnostics'; then
  fail "lint-biome must tell developers to inspect diagnostics before using make format"
fi
frontend_import_boundary_block="$(extract_target_block frontend-import-boundary-check)"
if [[ -z "$frontend_import_boundary_block" ]]; then
  fail "Makefile must define a non-empty frontend-import-boundary-check block"
fi
if ! printf '%s\n' "$frontend_import_boundary_block" | grep -Fq './scripts/check-frontend-import-boundaries.mjs'; then
  fail "frontend-import-boundary-check must run the repo-local import boundary checker"
fi
assert_target_prereq frontend-import-boundary-check '$(NODE_BIN)' "frontend-import-boundary-check must depend on NODE_BIN"
assert_target_prereq frontend-import-boundary-check '$(FRONTEND_INSTALL_STAMP)' "frontend-import-boundary-check must depend on frontend install"
assert_target_prereq lint frontend-import-boundary-check "lint must include frontend-import-boundary-check"
assert_target_prereq lint lint-shell "lint must include lint-shell"
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
if ! printf '%s\n' "$lint_scripts_block" | grep -Fq '$(RUN_SCRIPTS_BIOME_SCRIPT)'; then
  fail "lint-scripts must run the curated scripts Biome wrapper"
fi
lint_shell_block="$(extract_target_block lint-shell)"
if ! printf '%s\n' "$lint_shell_block" | grep -Fq 'scripts/run-shellcheck.sh'; then
  fail "lint-shell must run the curated ShellCheck wrapper"
fi
if ! printf '%s\n' "$lint_shell_block" | grep -Fq 'LINT_SHELL_STRICT="$(LINT_SHELL_STRICT)"'; then
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
"$node_bin" - "$frontend_import_boundary_config" <<'EOF'
const fs = require("node:fs");
const [configPath] = process.argv.slice(2);
const config = JSON.parse(fs.readFileSync(configPath, "utf8"));
if (config.schema_id !== "cartulary.frontend_import_boundaries.v1") {
  throw new Error("frontend import boundary config must declare schema_id=cartulary.frontend_import_boundaries.v1");
}
const scanRoots = new Set(config.scan_roots ?? []);
for (const required of ["apps/web/src", "apps/web/e2e", "packages/grid-adapter/src", "packages/protocol-ts/src"]) {
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
if (!includes.includes("!packages/protocol-ts/src/generated")) {
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
if ! printf '%s\n' "$test_fast_block" | grep -Fq '$(RUN_MAKE_SEQUENCE_SCRIPT) --sequence test-fast'; then
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
const sections = ["unit", "integration", "e2e"];
const phases = [];

for (const entry of fs.readdirSync(path.join(root, "tools")).sort()) {
  const match = /^(phase\d+)_test_map\.json$/.exec(entry);
  if (!match) {
    continue;
  }
  const phase = match[1];
  const manifest = JSON.parse(
    fs.readFileSync(path.join(root, "tools", entry), "utf8"),
  );
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
