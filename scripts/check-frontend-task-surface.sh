#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
makefile="$repo_root/Makefile"
runner_script="$repo_root/scripts/run-frontend-unit.sh"
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
  ' "$makefile"
}

extract_target_prereqs() {
  local target="$1"
  awk -v target="$target" '
    $0 ~ "^" target ":" && $0 !~ "^" target ":[[:space:]]+export[[:space:]]" {
      sub("^" target ":[[:space:]]*", "", $0)
      print
      exit
    }
  ' "$makefile"
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

if ! rg -q '^frontend-task-surface-check:' "$makefile"; then
  fail "Makefile must define frontend-task-surface-check"
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
if ! printf '%s\n' "$check_setup_block" | rg -q 'frontend-install'; then
  fail "check-setup-blockers must invoke frontend install after toolchain drift"
fi
assert_text_order "check-setup-blockers recipe" "$check_setup_block" "toolchain-drift" "codegen-toolchain" "check-setup-blockers must prepare codegen after toolchain drift"
assert_text_order "check-setup-blockers recipe" "$check_setup_block" "codegen-toolchain" "frontend-install" "check-setup-blockers must install frontend dependencies after codegen readiness"
if printf '%s\n' "$check_setup_block" | rg -q 'frontend-task-surface-check|phase-ledger-drift|run-phase-smoke|generate-drift|lint-biome'; then
  fail "check-setup-blockers must not include static validation or harness smoke work"
fi
check_prereqs="$(extract_target_prereqs check)"
if printf '%s\n' "$check_prereqs" | rg -q 'FRONTEND_INSTALL_STAMP'; then
  fail "check must not depend directly on FRONTEND_INSTALL_STAMP"
fi
check_meta_validation_prereqs="$(extract_target_prereqs check-meta-validation)"
if ! printf '%s\n' "$check_meta_validation_prereqs" | rg -q '(^|[[:space:]])check-static-validation($|[[:space:]])'; then
  fail "check-meta-validation must include check-static-validation"
fi
if ! printf '%s\n' "$check_meta_validation_prereqs" | rg -q '(^|[[:space:]])check-harness-smoke($|[[:space:]])'; then
  fail "check-meta-validation must include check-harness-smoke"
fi
check_static_block="$(extract_target_block check-static-validation)"
if [[ -z "$check_static_block" ]]; then
  fail "Makefile must define a non-empty check-static-validation block"
fi
if ! printf '%s\n' "$check_static_block" | rg -q 'frontend-task-surface-check'; then
  fail "check-static-validation must invoke frontend-task-surface-check"
fi
if ! rg -q '^phase-ledgers:' "$makefile"; then
  fail "Makefile must define phase-ledgers"
fi
if ! rg -q '^phase-ledger-drift:' "$makefile"; then
  fail "Makefile must define phase-ledger-drift"
fi
if ! printf '%s\n' "$check_static_block" | rg -q 'phase-ledger-drift'; then
  fail "check-static-validation must invoke phase-ledger-drift"
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
if ! printf '%s\n' "$frontend_typecheck_block" | grep -Fq 'tsc --noEmit'; then
  fail "frontend-typecheck must run the frontend TypeScript compiler"
fi
if ! printf '%s\n' "$frontend_typecheck_block" | grep -Fq '$(TARGET_SUMMARY) frontend-typecheck pass'; then
  fail "frontend-typecheck must emit a target summary"
fi

if rg -q '^lint-typecheck:' "$makefile"; then
  fail "lint-typecheck must not remain as a legacy alias; use frontend-typecheck"
fi

if ! rg -q '^format:[[:space:]]+format-frontend$$' "$makefile"; then
  fail "format must delegate to format-frontend"
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
if ! printf '%s\n' "$lint_biome_block" | grep -Fq 'run make format to apply the authoritative frontend Biome scope'; then
  fail "lint-biome must tell developers to run make format"
fi

test_fast_block="$(extract_target_block test-fast)"
if [[ -z "$test_fast_block" ]]; then
  fail "Makefile must define a non-empty test-fast block"
fi
assert_target_recipe_invokes test-fast test-local "test-fast must route local frontend checks through test-local"
assert_target_prereq test-local backend-unit "test-fast must route local checks through test-local, and test-local must include backend-unit"
assert_target_prereq test-local frontend-typecheck "test-fast must route local frontend checks through test-local, and test-local must include frontend-typecheck"
assert_target_prereq test-local frontend-unit "test-fast must route local frontend checks through test-local, and test-local must include frontend-unit"
assert_target_recipe_invokes test-fast test-fast-service-backed "test-fast must invoke test-fast-service-backed"

check_local_product_prereqs="$(extract_target_prereqs check-local-product)"
if [[ -z "$check_local_product_prereqs" ]]; then
  fail "Makefile must define non-empty check-local-product prerequisites"
fi
if ! printf '%s\n' "$check_local_product_prereqs" | rg -q '(^|[[:space:]])frontend-typecheck($|[[:space:]])'; then
  fail "check-local-product must invoke frontend-typecheck"
fi

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
