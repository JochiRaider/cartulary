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

inspect_shared_command() {
  local target="$1"
  local shared_name="$2"
  NODE_BIN="$node_bin" "$go_runner_script" inspect-shared-command "$target" "$shared_name"
}

require_shared_command_match() {
  local shared_name="$1"
  shift

  if [[ "$#" -lt 2 ]]; then
    fail "require_shared_command_match needs <shared-name> <target>..."
  fi

  local baseline_target="$1"
  shift
  local baseline_command
  baseline_command="$(inspect_shared_command "$baseline_target" "$shared_name")" || fail "failed to inspect $shared_name for $baseline_target"

  local target
  local current_command
  for target in "$@"; do
    current_command="$(inspect_shared_command "$target" "$shared_name")" || fail "failed to inspect $shared_name for $target"
    if [[ "$baseline_command" != "$current_command" ]]; then
      fail "shared report $shared_name must resolve to the same go test command for $baseline_target and $target"
    fi
  done
}

require_shared_command_contains() {
  local target="$1"
  local shared_name="$2"
  local needle="$3"
  local command_text

  command_text="$(inspect_shared_command "$target" "$shared_name")" || fail "failed to inspect $shared_name for $target"
  if [[ "$command_text" != *"$needle"* ]]; then
    fail "shared report $shared_name for $target must include $needle"
  fi
}

support_selection_patterns() {
  local target="$1"
  shift

  "$node_bin" - "$repo_root" "$target" "$@" <<'EOF'
const fs = require("fs");
const path = require("path");

const [root, target, ...packagePatterns] = process.argv.slice(2);

function packageMatchesPattern(pkg, pattern) {
  if (pattern.endsWith("/...")) {
    const prefix = pattern.slice(0, -4);
    return pkg === prefix || pkg.startsWith(`${prefix}/`);
  }
  return pkg === pattern;
}

const patterns = new Set();
for (const entry of fs.readdirSync(path.join(root, "tools")).sort()) {
  if (!/^phase\d+_test_map\.json$/.test(entry)) {
    continue;
  }
  const manifest = JSON.parse(
    fs.readFileSync(path.join(root, "tools", entry), "utf8"),
  );
  for (const supportEntry of manifest.support_go_targets ?? []) {
    if (supportEntry.target !== target) {
      continue;
    }
    if (!packagePatterns.some((pattern) => packageMatchesPattern(supportEntry.package, pattern))) {
      continue;
    }
    patterns.add(supportEntry.selection_pattern);
  }
}

const values = Array.from(patterns).sort();
if (values.length > 0) {
  process.stdout.write(`${values.join("\n")}\n`);
}
EOF
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
  [
    "phase4",
    [
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
  'emit_declared_support_phase "backend-unit support phase1"' \
  'emit_declared_support_phase "backend-unit support phase3"'
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
  'emit_go_manifest_phase "backend-integration phase4 authoritative"' \
  'emit_go_manifest_phase "backend-integration phase2 authoritative"' \
  'emit_declared_support_phase "backend-integration support phase1"' \
  'emit_declared_support_phase "backend-integration support phase2"' \
  'emit_declared_support_phase "backend-integration support phase3"' \
  'emit_declared_support_phase "backend-integration support phase4"'
do
  if ! grep -Fq "$expected" "$go_runner_script"; then
    fail "scripts/run-go-target.sh must preserve backend-integration selection surface: missing $expected"
  fi
done

backend_process_block="$(extract_target_block backend-process)"
if ! printf '%s\n' "$backend_process_block" | grep -Fq 'run-go-target.sh backend-process'; then
  fail "backend-process must delegate to scripts/run-go-target.sh backend-process"
fi
if ! printf '%s\n' "$backend_process_block" | grep -Fq 'CARTULARY_SERVER_BIN=$(SERVER_BIN)'; then
  fail "backend-process must export CARTULARY_SERVER_BIN=$(SERVER_BIN)"
fi
if ! grep -Fq 'emit_go_manifest_phase "backend-process phase0 authoritative"' "$go_runner_script"; then
  fail "scripts/run-go-target.sh must preserve backend-process Phase 0 manifest selection"
fi

phase0_process_block="$(extract_target_block phase0-process-e2e)"
if ! printf '%s\n' "$phase0_process_block" | grep -Fq 'run-go-target.sh phase0-process-e2e'; then
  fail "phase0-process-e2e must delegate to scripts/run-go-target.sh phase0-process-e2e"
fi
if ! printf '%s\n' "$phase0_process_block" | grep -Fq 'CARTULARY_SERVER_BIN=$(SERVER_BIN)'; then
  fail "phase0-process-e2e must export CARTULARY_SERVER_BIN=$(SERVER_BIN)"
fi
if ! grep -Fq 'emit_go_manifest_phase "phase0-process-e2e"' "$go_runner_script"; then
  fail "scripts/run-go-target.sh must preserve phase0-process-e2e Phase 0 manifest selection"
fi

phase1_process_block="$(extract_target_block phase1-process-smoke)"
if ! printf '%s\n' "$phase1_process_block" | grep -Fq 'CARTULARY_SERVER_BIN=$(SERVER_BIN)'; then
  fail "phase1-process-smoke must export CARTULARY_SERVER_BIN=$(SERVER_BIN)"
fi

phase2_process_block="$(extract_target_block phase2-process-smoke)"
if ! printf '%s\n' "$phase2_process_block" | grep -Fq 'CARTULARY_SERVER_BIN=$(SERVER_BIN)'; then
  fail "phase2-process-smoke must export CARTULARY_SERVER_BIN=$(SERVER_BIN)"
fi

for target in backend-process phase0-process-e2e phase1-process-smoke phase2-process-smoke; do
  prereqs="$(extract_target_prereqs "$target")"
  if [[ -z "$prereqs" || "$prereqs" != *build-server* ]]; then
    fail "$target must depend on build-server"
  fi
done

check_service_block="$(extract_target_block check-service-backed)"
if [[ -z "$check_service_block" ]]; then
  fail "Makefile must define a non-empty check-service-backed block"
fi

for lane in check-service-backed-lane-a check-service-backed-lane-b; do
  if ! printf '%s\n' "$check_service_block" | rg -q "(^|[[:space:]])$lane($|[[:space:]])"; then
    fail "check-service-backed must invoke $lane"
  fi
done

if printf '%s\n' "$check_service_block" | rg -q '(^|[[:space:]])(backend-process-support|phase2-process-smoke)($|[[:space:]])'; then
  fail "check-service-backed must not invoke Phase 2 process smoke coverage"
fi

check_service_lane_a_block="$(extract_target_block check-service-backed-lane-a)"
for target in backend-integration backend-integration-support; do
  if ! printf '%s\n' "$check_service_lane_a_block" | rg -q "(^|[[:space:]])$target($|[[:space:]])"; then
    fail "check-service-backed-lane-a must invoke $target"
  fi
done

check_service_lane_b_prereqs="$(extract_target_prereqs check-service-backed-lane-b)"
check_service_lane_b_block="$(extract_target_block check-service-backed-lane-b)"
for target in backend-store backend-process; do
  if ! printf '%s\n' "$check_service_lane_b_prereqs" | rg -q "(^|[[:space:]])$target($|[[:space:]])"; then
    fail "check-service-backed-lane-b must invoke $target"
  fi
done
if printf '%s\n' "${check_service_lane_b_prereqs} ${check_service_lane_b_block}" | rg -q '(^|[[:space:]])phase2-process-smoke($|[[:space:]])'; then
  fail "check-service-backed-lane-b must not invoke Phase 2 process smoke coverage"
fi

backend_integration_support_block="$(extract_target_block backend-integration-support)"
if ! printf '%s\n' "$backend_integration_support_block" | grep -Fq 'run-go-target.sh backend-integration-support'; then
  fail "backend-integration-support must delegate to scripts/run-go-target.sh backend-integration-support"
fi
require_shared_command_match backend-integration-core backend-integration backend-integration-support
require_shared_command_match backend-integration-auth backend-integration backend-integration-support
mapfile -t backend_integration_core_support_patterns < <(support_selection_patterns backend_integration_support ./internal/platform/... ./internal/app ./internal/modules/incidents ./internal/modules/entities ./internal/modules/timeline)
if [[ "${#backend_integration_core_support_patterns[@]}" -eq 0 ]]; then
  fail "backend-integration-core must have declared support selectors"
fi
for pattern in "${backend_integration_core_support_patterns[@]}"; do
  [[ -z "$pattern" ]] && continue
  require_shared_command_contains backend-integration backend-integration-core "$pattern"
done
mapfile -t backend_integration_auth_support_patterns < <(support_selection_patterns backend_integration_support ./internal/modules/auth)
if [[ "${#backend_integration_auth_support_patterns[@]}" -eq 0 ]]; then
  fail "backend-integration-auth must have declared support selectors"
fi
for pattern in "${backend_integration_auth_support_patterns[@]}"; do
  [[ -z "$pattern" ]] && continue
  require_shared_command_contains backend-integration backend-integration-auth "$pattern"
done
require_shared_command_match backend-process-shared backend-process phase0-process-e2e phase1-process-smoke phase2-process-smoke

test_fast_block="$(extract_target_block test-fast)"
if [[ -z "$test_fast_block" ]]; then
  fail "Makefile must define a non-empty test-fast block"
fi
for lane in test-fast-service-backed-lane-a test-fast-service-backed-lane-b; do
  if ! printf '%s\n' "$test_fast_block" | rg -q "(^|[[:space:]])$lane($|[[:space:]])"; then
    fail "test-fast must invoke $lane"
  fi
done
if printf '%s\n' "$test_fast_block" | rg -q '(^|[[:space:]])(backend-process-support|phase2-process-smoke)($|[[:space:]])'; then
  fail "test-fast must not invoke Phase 2 process smoke coverage"
fi

test_fast_lane_a_block="$(extract_target_block test-fast-service-backed-lane-a)"
for target in backend-integration backend-integration-support; do
  if ! printf '%s\n' "$test_fast_lane_a_block" | rg -q "(^|[[:space:]])$target($|[[:space:]])"; then
    fail "test-fast-service-backed-lane-a must invoke $target"
  fi
done

test_fast_lane_b_prereqs="$(extract_target_prereqs test-fast-service-backed-lane-b)"
test_fast_lane_b_block="$(extract_target_block test-fast-service-backed-lane-b)"
for target in backend-store backend-process; do
  if ! printf '%s\n' "$test_fast_lane_b_prereqs" | rg -q "(^|[[:space:]])$target($|[[:space:]])"; then
    fail "test-fast-service-backed-lane-b must invoke $target"
  fi
done
if printf '%s\n' "${test_fast_lane_b_prereqs} ${test_fast_lane_b_block}" | rg -q '(^|[[:space:]])phase2-process-smoke($|[[:space:]])'; then
  fail "test-fast-service-backed-lane-b must not invoke Phase 2 process smoke coverage"
fi
