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

check_preflight_block="$(extract_target_block check-preflight)"
if [[ -z "$check_preflight_block" ]]; then
  fail "Makefile must define a non-empty check-preflight block"
fi
if ! printf '%s\n' "$check_preflight_block" | rg -q 'frontend-task-surface-check'; then
  fail "check-preflight must invoke frontend-task-surface-check"
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
