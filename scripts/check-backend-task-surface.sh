#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
makefile="$repo_root/Makefile"
go_runner_script="$repo_root/scripts/run-go-target.sh"
node_bin="${NODE_BIN:-node}"

fail() {
  echo "$*" >&2
  exit 1
}

extract_target_block() {
  local target="$1"
  awk -v target="$target" '
    $0 ~ "^" target ":" { in_block=1; next }
    in_block && /^[^[:space:]].*:/ { exit }
    in_block { print }
  ' "$makefile"
}

check_heavy_line="$(sed -n 's/^check-heavy:[[:space:]]*//p' "$makefile" | head -n 1)"
if [[ -z "$check_heavy_line" ]]; then
  fail "Makefile must define check-heavy prerequisites"
fi

read -r -a heavy_prereqs <<<"$check_heavy_line"
service_backed_targets=(
  backend-store
  backend-integration
  backend-integration-support
  backend-process
)
for target in "${service_backed_targets[@]}"; do
  for prereq in "${heavy_prereqs[@]}"; do
    if [[ "$prereq" == "$target" ]]; then
      fail "check-heavy must not include service-backed target $target"
    fi
  done
done

if ! rg -q '^backend-store:' "$makefile"; then
  fail "Makefile must define backend-store"
fi

if ! rg -q '^phase2-process-smoke:' "$makefile"; then
  fail "Makefile must define phase2-process-smoke"
fi

if rg -q '^backend-process-support:' "$makefile"; then
  fail "Makefile must not define backend-process-support; Phase 2 smoke must stay direct-run support-only via phase2-process-smoke"
fi

if rg -q 'TestPhase0_.*_U_0_|TestPhase0_.*_I_0_|TestPhase0_.*_E_0_' "$makefile"; then
  fail "Makefile must not use regex-based Phase 0 Go selection"
fi

if ! "$node_bin" - "$repo_root" <<'EOF'
const fs = require("fs");
const path = require("path");

const root = process.argv[2];
for (const [phase, sections] of [
  [
    "phase0",
    [
      ["unit", "backend_unit"],
      ["integration", "backend_integration"],
      ["e2e", "backend_process"],
    ],
  ],
  [
    "phase1",
    [
      ["unit", "backend_unit"],
      ["integration", "backend_integration"],
    ],
  ],
]) {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(root, "tools", `${phase}_test_map.json`), "utf8"),
  );
  for (const [section, dependency] of sections) {
    for (const entry of manifest[section] ?? []) {
      if (entry.coverage !== "authoritative" || entry.runner !== "go_test") {
        continue;
      }
      if (entry.execution_dependency !== dependency) {
        console.error(
          `${phase} authoritative ${section} row ${entry.id} must declare execution_dependency=${dependency}`,
        );
        process.exit(1);
      }
    }
  }
}

const phase2 = JSON.parse(
  fs.readFileSync(path.join(root, "tools", "phase2_test_map.json"), "utf8"),
);
const phase2UnitDeps = new Map([
  ["U-2-01", "backend_unit"],
  ["U-2-02", "backend_store"],
  ["U-2-03", "backend_store"],
  ["U-2-04", "backend_store"],
  ["U-2-05", "backend_unit"],
  ["U-2-06", "backend_unit"],
  ["U-2-07", "backend_store"],
  ["U-2-08", "backend_unit"],
  ["U-2-09", "backend_unit"],
  ["U-2-10", "backend_unit"],
]);
for (const entry of phase2.unit ?? []) {
  if (entry.coverage !== "authoritative" || entry.runner !== "go_test") {
    continue;
  }
  const expected = phase2UnitDeps.get(entry.id);
  if (!expected) {
    console.error(`phase2 authoritative unit row ${entry.id} is missing a canonical execution_dependency expectation`);
    process.exit(1);
  }
  if (entry.execution_dependency !== expected) {
    console.error(
      `phase2 authoritative unit row ${entry.id} must declare execution_dependency=${expected}`,
    );
    process.exit(1);
  }
}
for (const entry of phase2.integration ?? []) {
  if (entry.coverage !== "authoritative" || entry.runner !== "go_test") {
    continue;
  }
  if (entry.execution_dependency !== "backend_integration") {
    console.error(
      `phase2 authoritative integration row ${entry.id} must declare execution_dependency=backend_integration`,
    );
    process.exit(1);
  }
}
EOF
then
  fail "Authoritative backend manifests must carry the canonical execution_dependency for their layer"
fi

backend_unit_block="$(extract_target_block backend-unit)"
if ! printf '%s\n' "$backend_unit_block" | grep -Fq 'run-go-target.sh backend-unit'; then
  fail "backend-unit must delegate to scripts/run-go-target.sh backend-unit"
fi
for expected in \
  'manifest_go_regex phase0 unit authoritative backend_unit ./internal/platform/...' \
  'manifest_go_regex phase0 unit authoritative backend_unit ./internal/app' \
  'manifest_go_count phase1 unit authoritative backend_unit ./internal/platform/...' \
  'manifest_go_regex phase1 unit authoritative backend_unit ./internal/modules/auth' \
  'emit_go_manifest_phase "backend-unit phase2 authoritative"' \
  'emit_go_manifest_phase "backend-unit phase3 authoritative"' \
  'TestSupportPhase1_'
do
  if ! grep -Fq "$expected" "$go_runner_script"; then
    fail "scripts/run-go-target.sh must preserve backend-unit selection surface: missing $expected"
  fi
done

backend_store_block="$(extract_target_block backend-store)"
if ! printf '%s\n' "$backend_store_block" | grep -Fq 'run-go-target.sh backend-store'; then
  fail "backend-store must delegate to scripts/run-go-target.sh backend-store"
fi
for expected in \
  'emit_go_manifest_phase "backend-store phase2 authoritative"' \
  'emit_go_manifest_phase "backend-store phase3 authoritative"'
do
  if ! grep -Fq "$expected" "$go_runner_script"; then
    fail "scripts/run-go-target.sh must preserve backend-store selection surface: missing $expected"
  fi
done

backend_integration_block="$(extract_target_block backend-integration)"
if ! printf '%s\n' "$backend_integration_block" | grep -Fq 'run-go-target.sh backend-integration'; then
  fail "backend-integration must delegate to scripts/run-go-target.sh backend-integration"
fi
for expected in \
  'emit_go_manifest_phase "backend-integration phase0 authoritative"' \
  'emit_go_manifest_phase "backend-integration phase1 authoritative"' \
  'emit_go_manifest_phase "backend-integration phase2 authoritative"' \
  'emit_go_raw_phase "backend-integration support phase1"' \
  'emit_go_raw_phase "backend-integration support phase2"' \
  'emit_go_raw_phase "backend-integration support phase3"'
do
  if ! grep -Fq "$expected" "$go_runner_script"; then
    fail "scripts/run-go-target.sh must preserve backend-integration selection surface: missing $expected"
  fi
done

backend_process_block="$(extract_target_block backend-process)"
if ! printf '%s\n' "$backend_process_block" | grep -Fq 'run-go-target.sh backend-process'; then
  fail "backend-process must delegate to scripts/run-go-target.sh backend-process"
fi
if ! grep -Fq 'emit_go_manifest_phase "backend-process phase0 authoritative"' "$go_runner_script"; then
  fail "scripts/run-go-target.sh must preserve backend-process Phase 0 manifest selection"
fi

phase0_process_block="$(extract_target_block phase0-process-e2e)"
if ! printf '%s\n' "$phase0_process_block" | grep -Fq 'run-go-target.sh phase0-process-e2e'; then
  fail "phase0-process-e2e must delegate to scripts/run-go-target.sh phase0-process-e2e"
fi
if ! grep -Fq 'emit_go_manifest_phase "phase0-process-e2e"' "$go_runner_script"; then
  fail "scripts/run-go-target.sh must preserve phase0-process-e2e Phase 0 manifest selection"
fi

check_service_block="$(extract_target_block check-service-backed)"
if [[ -z "$check_service_block" ]]; then
  fail "Makefile must define a non-empty check-service-backed block"
fi

for target in "${service_backed_targets[@]}"; do
  if ! printf '%s\n' "$check_service_block" | rg -q "(^|[[:space:]])$target($|[[:space:]])"; then
    fail "check-service-backed must invoke $target"
  fi
done

if printf '%s\n' "$check_service_block" | rg -q '(^|[[:space:]])(backend-process-support|phase2-process-smoke)($|[[:space:]])'; then
  fail "check-service-backed must not invoke Phase 2 process smoke coverage"
fi

backend_integration_support_block="$(extract_target_block backend-integration-support)"
if ! printf '%s\n' "$backend_integration_support_block" | grep -Fq 'run-go-target.sh backend-integration-support'; then
  fail "backend-integration-support must delegate to scripts/run-go-target.sh backend-integration-support"
fi
if ! grep -Fq "TestSupportPhase1_" "$go_runner_script"; then
  fail "scripts/run-go-target.sh must run Phase 1 support coverage through TestSupportPhase1_"
fi

test_fast_block="$(extract_target_block test-fast)"
if [[ -z "$test_fast_block" ]]; then
  fail "Makefile must define a non-empty test-fast block"
fi
if ! printf '%s\n' "$test_fast_block" | rg -q '(^|[[:space:]])backend-store($|[[:space:]])'; then
  fail "test-fast must invoke backend-store"
fi
if printf '%s\n' "$test_fast_block" | rg -q '(^|[[:space:]])(backend-process-support|phase2-process-smoke)($|[[:space:]])'; then
  fail "test-fast must not invoke Phase 2 process smoke coverage"
fi
