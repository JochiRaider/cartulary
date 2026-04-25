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
check_preflight_prereqs="$(extract_target_prereqs check-preflight)"
if ! printf '%s\n' "$check_preflight_prereqs" | rg -q '(^|[[:space:]])check-setup-blockers($|[[:space:]])'; then
  fail "check-preflight must remain a compatibility alias to check-setup-blockers"
fi
check_setup_block="$(extract_target_block check-setup-blockers)"
if [[ -z "$check_setup_block" ]]; then
  fail "Makefile must define a non-empty check-setup-blockers block"
fi
if ! printf '%s\n' "$check_setup_block" | rg -q 'toolchain-drift'; then
  fail "check-setup-blockers must invoke toolchain-drift"
fi
if ! printf '%s\n' "$check_setup_block" | rg -q 'frontend-install'; then
  fail "check-setup-blockers must invoke frontend install after toolchain drift"
fi
if printf '%s\n' "$check_setup_block" | rg -q 'frontend-task-surface-check|phase-ledger-drift|run-phase-smoke|generate-drift|lint-biome'; then
  fail "check-setup-blockers must not include static validation or harness smoke work"
fi
check_prereqs="$(extract_target_prereqs check)"
if printf '%s\n' "$check_prereqs" | rg -q 'FRONTEND_INSTALL_STAMP'; then
  fail "check must not depend directly on FRONTEND_INSTALL_STAMP"
fi
check_parallel_prereqs="$(extract_target_prereqs check-parallel)"
if ! printf '%s\n' "$check_parallel_prereqs" | rg -q '(^|[[:space:]])check-static-validation($|[[:space:]])'; then
  fail "check-parallel must include check-static-validation"
fi
if ! printf '%s\n' "$check_parallel_prereqs" | rg -q '(^|[[:space:]])check-harness-smoke($|[[:space:]])'; then
  fail "check-parallel must include check-harness-smoke"
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
if ! printf '%s\n' "$frontend_typecheck_block" | grep -Fq 'tsc --noEmit'; then
  fail "frontend-typecheck must run the frontend TypeScript compiler"
fi

if ! rg -q '^lint-typecheck:[[:space:]]+frontend-typecheck$$' "$makefile"; then
  fail "lint-typecheck must remain a compatibility alias to frontend-typecheck"
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
if ! printf '%s\n' "$test_fast_block" | rg -q '(^|[[:space:]])frontend-typecheck($|[[:space:]])'; then
  fail "test-fast must invoke frontend-typecheck"
fi

check_heavy_prereqs="$(extract_target_prereqs check-heavy)"
if [[ -z "$check_heavy_prereqs" ]]; then
  fail "Makefile must define non-empty check-heavy prerequisites"
fi
if ! printf '%s\n' "$check_heavy_prereqs" | rg -q '(^|[[:space:]])frontend-typecheck($|[[:space:]])'; then
  fail "check-heavy must invoke frontend-typecheck"
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

for phase in "${manifest_phases[@]}"; do
  if ! grep -Fq "frontend-unit ${phase} authoritative" "$runner_script"; then
    fail "scripts/run-frontend-unit.sh must emit frontend-unit ${phase} authoritative"
  fi
done
