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

target_plan_support_patterns() {
  local target="$1"
  local shared_report="$2"

  "$node_bin" - "$target_plan_file" "$target" "$shared_report" <<'EOF'
const fs = require("node:fs");

const [planFile, target, sharedReport] = process.argv.slice(2);
const rows = JSON.parse(fs.readFileSync(planFile, "utf8"));
const patterns = new Set();
for (const row of rows) {
  if (
    row.target === target &&
    row.shared_report === sharedReport &&
    row.support_only === true &&
    typeof row.support_selector === "string" &&
    row.support_selector !== ""
  ) {
    patterns.add(row.support_selector);
  }
}

const values = Array.from(patterns).sort();
if (values.length > 0) {
  process.stdout.write(`${values.join("\n")}\n`);
}
EOF
}

target_plan_boolean_targets() {
  local field="$1"
  local expected="$2"

  "$node_bin" - "$target_plan_file" "$field" "$expected" <<'EOF'
const fs = require("node:fs");

const [planFile, field, expectedRaw] = process.argv.slice(2);
const expected = expectedRaw === "true";
const rows = JSON.parse(fs.readFileSync(planFile, "utf8"));
const targets = new Set();
for (const row of rows) {
  if (row[field] === expected) {
    targets.add(row.target);
  }
}
const values = Array.from(targets).sort();
if (values.length > 0) {
  process.stdout.write(`${values.join("\n")}\n`);
}
EOF
}

list_target_plan_service_backed_unsafe_targets() {
  "$node_bin" - "$target_plan_file" <<'EOF'
const fs = require("node:fs");

const [planFile] = process.argv.slice(2);
const rows = JSON.parse(fs.readFileSync(planFile, "utf8"));
const targets = new Set();
for (const row of rows) {
  if (row.service_backed === true && row.check_service_backed_safe !== true) {
    targets.add(row.target);
  }
}
const values = Array.from(targets).sort();
if (values.length > 0) {
  process.stdout.write(`${values.join("\n")}\n`);
}
EOF
}

assert_text_contains_targets() {
  local label="$1"
  local text="$2"
  shift 2

  local target
  for target in "$@"; do
    if ! printf '%s\n' "$text" | rg -q "(^|[[:space:]])$target($|[[:space:]])"; then
      fail "$label must include target-plan target $target"
    fi
  done
}

assert_text_excludes_targets() {
  local label="$1"
  local text="$2"
  shift 2

  local target
  for target in "$@"; do
    if printf '%s\n' "$text" | rg -q "(^|[[:space:]])$target($|[[:space:]])"; then
      fail "$label must not include target-plan target $target"
    fi
  done
}

target_plan_file="$(mktemp)"
trap 'rm -f "$target_plan_file"' EXIT
"$node_bin" "$repo_root/scripts/print-target-plan.mjs" --json >"$target_plan_file"
mapfile -t target_plan_check_heavy_targets < <(target_plan_boolean_targets check_heavy_safe true)
mapfile -t target_plan_service_backed_safe_targets < <(target_plan_boolean_targets check_service_backed_safe true)
mapfile -t target_plan_service_backed_unsafe_targets < <(list_target_plan_service_backed_unsafe_targets)

check_heavy_line="$(sed -n 's/^check-heavy:[[:space:]]*//p' "$makefile" | head -n 1)"
if [[ -z "$check_heavy_line" ]]; then
  fail "Makefile must define check-heavy prerequisites"
fi
check_parallel_line="$(sed -n 's/^check-parallel:[[:space:]]*//p' "$makefile" | head -n 1)"
if [[ -z "$check_parallel_line" ]]; then
  fail "Makefile must define check-parallel prerequisites"
fi
if ! printf '%s\n' "$check_parallel_line" | rg -q '(^|[[:space:]])check-heavy($|[[:space:]])'; then
  fail "check-parallel must include check-heavy"
fi
if ! printf '%s\n' "$check_parallel_line" | rg -q '(^|[[:space:]])check-static-validation($|[[:space:]])'; then
  fail "check-parallel must include static validation alongside backend product checks"
fi
if ! printf '%s\n' "$check_parallel_line" | rg -q '(^|[[:space:]])check-harness-smoke($|[[:space:]])'; then
  fail "check-parallel must include harness smoke alongside backend product checks"
fi

read -r -a heavy_prereqs <<<"$check_heavy_line"
assert_text_contains_targets "check-heavy prerequisites" "$check_heavy_line" "${target_plan_check_heavy_targets[@]}"
assert_text_excludes_targets "check-heavy prerequisites" "$check_heavy_line" "${target_plan_service_backed_safe_targets[@]}" "${target_plan_service_backed_unsafe_targets[@]}"

if grep -Fq 'rg --files' "$makefile"; then
  fail "Makefile must not use parse-time rg --files for build input discovery"
fi
if ! grep -Fq 'scripts/list-build-inputs.sh' "$makefile"; then
  fail "Makefile must use scripts/list-build-inputs.sh for build input discovery"
fi

for target in services-up services-wait postgres-wait minio-wait minio-init; do
  if ! rg -q "^${target}:" "$makefile"; then
    fail "Makefile must define $target"
  fi
done

services_wait_prereqs="$(extract_target_prereqs services-wait)"
for target in postgres-wait minio-wait; do
  if ! printf '%s\n' "$services_wait_prereqs" | rg -q "(^|[[:space:]])$target($|[[:space:]])"; then
    fail "services-wait must depend on $target"
  fi
done

services_up_block="$(extract_target_block services-up)"
if ! printf '%s\n' "$services_up_block" | grep -Fq 'up -d postgres minio'; then
  fail "services-up must start postgres and minio"
fi
if ! printf '%s\n' "$services_up_block" | grep -Fq 'services-wait'; then
  fail "services-up must wait for service readiness"
fi

db_up_block="$(extract_target_block db-up)"
if ! printf '%s\n' "$db_up_block" | grep -Fq 'services-up'; then
  fail "db-up must delegate service startup to services-up"
fi
if ! printf '%s\n' "$db_up_block" | grep -Fq 'minio-init'; then
  fail "db-up must initialize the default MinIO bucket"
fi

db_reset_block="$(extract_target_block db-reset)"
if ! printf '%s\n' "$db_reset_block" | grep -Fq 'postgres-wait'; then
  fail "db-reset must wait for postgres before resetting the database"
fi
if ! printf '%s\n' "$db_reset_block" | grep -Fq 'MinIO/object storage is not reset'; then
  fail "db-reset must explicitly report that object storage is not reset"
fi

minio_init_block="$(extract_target_block minio-init)"
if ! printf '%s\n' "$minio_init_block" | grep -Fq 'MINIO_BUCKET="$(MINIO_BUCKET)"'; then
  fail "minio-init must pass the configured MINIO_BUCKET"
fi
if ! printf '%s\n' "$minio_init_block" | grep -Fq 'init-minio'; then
  fail "minio-init must delegate bucket creation to dev-services.sh"
fi

help_block="$(extract_target_block help)"
if ! printf '%s\n' "$help_block" | grep -Fq 'make services-up'; then
  fail "help must document services-up"
fi
if ! printf '%s\n' "$help_block" | grep -Fq 'does not reset object storage'; then
  fail "help must document db-reset object-storage scope"
fi

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

if rg -q 'TestPhase4_.*_U_4_' "$go_runner_script"; then
  fail "scripts/run-go-target.sh must not use raw authoritative Phase 4 U-4-* Go selectors"
fi

if ! "$node_bin" - "$repo_root" "$target_plan_file" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");

const [root, planFile] = process.argv.slice(2);
const rows = JSON.parse(fs.readFileSync(planFile, "utf8"));
const backendDependencies = new Set([
  "backend_unit",
  "backend_store",
  "backend_integration",
  "backend_process",
]);

function supportSymbols(entry) {
  if (entry.symbol !== undefined && entry.symbols !== undefined) {
    throw new Error(`${entry.file} must not declare both symbol and symbols`);
  }
  if (entry.symbols !== undefined) {
    return entry.symbols;
  }
  return [entry.symbol];
}

const manifestRows = [];
const supportRows = [];
for (const file of fs.readdirSync(path.join(root, "tools")).sort()) {
  if (!/^phase\d+_test_map\.json$/.test(file)) {
    continue;
  }
  const phase = file.replace(/_test_map\.json$/, "");
  const manifest = JSON.parse(fs.readFileSync(path.join(root, "tools", file), "utf8"));
  for (const section of ["unit", "integration", "e2e"]) {
    for (const entry of manifest[section] ?? []) {
      if (
        entry.coverage === "authoritative" &&
        entry.runner === "go_test" &&
        backendDependencies.has(entry.execution_dependency)
      ) {
        manifestRows.push({ ...entry, phase, section });
      }
    }
  }
  for (const entry of manifest.support_go_targets ?? []) {
    for (const symbol of supportSymbols(entry)) {
      supportRows.push({ ...entry, phase, symbol });
    }
  }
}

for (const entry of manifestRows) {
  const matches = rows.filter(
    (row) =>
      row.canonical_authoritative === true &&
      row.support_only === false &&
      row.coverage === "authoritative" &&
      row.id === entry.id &&
      row.manifest_phase === entry.phase,
  );
  if (matches.length !== 1) {
    console.error(
      `${entry.phase} authoritative ${entry.section} row ${entry.id} must appear in exactly one canonical target-plan row, found ${matches.length}`,
    );
    process.exit(1);
  }
  const row = matches[0];
  if (row.execution_dependency !== entry.execution_dependency || row.section !== entry.section) {
    console.error(
      `${entry.phase} authoritative row ${entry.id} target-plan mismatch: expected ${entry.section}/${entry.execution_dependency}, found ${row.section}/${row.execution_dependency}`,
    );
    process.exit(1);
  }
}

const manifestKeys = new Set(manifestRows.map((entry) => `${entry.phase}:${entry.id}`));
for (const row of rows) {
  if (
    row.canonical_authoritative === true &&
    row.support_only === false &&
    row.coverage === "authoritative" &&
    !manifestKeys.has(`${row.manifest_phase}:${row.id}`)
  ) {
    console.error(`target-plan row ${row.target} ${row.manifest_phase} ${row.id} is not backed by an authoritative backend manifest row`);
    process.exit(1);
  }
}

for (const entry of supportRows) {
  const matches = rows.filter(
    (row) =>
      row.support_only === true &&
      row.manifest_phase === entry.phase &&
      row.execution_dependency === entry.target &&
      row.file === entry.file &&
      row.support_selector === entry.selection_pattern &&
      Array.isArray(row.symbols) &&
      row.symbols.includes(entry.symbol),
  );
  if (matches.length !== 1) {
    console.error(
      `${entry.phase} support row ${entry.file}::${entry.symbol} must appear in exactly one target-plan support row, found ${matches.length}`,
    );
    process.exit(1);
  }
}
EOF
then
  fail "Backend target plan must match authoritative manifests and support selectors"
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
  'manifest_go_regex phase4 unit authoritative backend_unit ./internal/app ./internal/modules/incidents ./internal/modules/entities ./internal/modules/timeline' \
  'emit_go_manifest_phase "backend-unit phase2 authoritative"' \
  'emit_go_manifest_phase "backend-unit phase3 authoritative"' \
  'emit_go_manifest_phase "backend-unit phase4 authoritative"' \
  'emit_declared_support_phase "backend-unit support phase0"' \
  'emit_declared_support_phase "backend-unit support phase1"' \
  'emit_declared_support_phase "backend-unit support phase2"' \
  'emit_declared_support_phase "backend-unit support phase3"'
do
  if ! grep -Fq "$expected" "$go_runner_script"; then
    fail "scripts/run-go-target.sh must preserve backend-unit selection surface: missing $expected"
  fi
done
mapfile -t backend_unit_core_support_patterns < <(target_plan_support_patterns backend-unit backend-unit-core)
if [[ "${#backend_unit_core_support_patterns[@]}" -eq 0 ]]; then
  fail "backend-unit core packages must have declared support selectors"
fi
for pattern in "${backend_unit_core_support_patterns[@]}"; do
  [[ -z "$pattern" ]] && continue
  require_shared_command_contains backend-unit backend-unit-core "$pattern"
done
mapfile -t backend_unit_auth_support_patterns < <(target_plan_support_patterns backend-unit backend-unit-auth)
if [[ "${#backend_unit_auth_support_patterns[@]}" -eq 0 ]]; then
  fail "backend-unit auth packages must have declared support selectors"
fi
for pattern in "${backend_unit_auth_support_patterns[@]}"; do
  [[ -z "$pattern" ]] && continue
  require_shared_command_contains backend-unit backend-unit-auth "$pattern"
done

backend_store_block="$(extract_target_block backend-store)"
if ! printf '%s\n' "$backend_store_block" | grep -Fq 'run-go-target.sh backend-store'; then
  fail "backend-store must delegate to scripts/run-go-target.sh backend-store"
fi
if ! printf '%s\n' "$backend_store_block" | grep -Fq '$(TEST_SERVICES_BIN) run --'; then
  fail "backend-store must run through $(TEST_SERVICES_BIN) for suite-scoped services"
fi
for expected in \
  'manifest_go_regex phase4 unit authoritative backend_store ./internal/modules/entities ./internal/modules/timeline' \
  'emit_go_manifest_phase "backend-store phase4 authoritative"' \
  'emit_go_manifest_phase "backend-store phase2 authoritative"' \
  'emit_go_manifest_phase "backend-store phase1 authoritative"' \
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
if ! printf '%s\n' "$backend_integration_block" | grep -Fq '$(TEST_SERVICES_BIN) run --'; then
  fail "backend-integration must run through $(TEST_SERVICES_BIN) for suite-scoped services"
fi
for expected in \
  'emit_go_manifest_phase "backend-integration phase0 authoritative platform"' \
  'emit_go_manifest_phase "backend-integration phase0 authoritative app"' \
  'emit_go_manifest_phase "backend-integration phase1 authoritative"' \
  'emit_go_manifest_phase "backend-integration phase4 authoritative entities"' \
  'emit_go_manifest_phase "backend-integration phase4 authoritative timeline"' \
  'emit_go_manifest_phase "backend-integration phase2 authoritative"' \
  'emit_declared_support_phase "backend-integration support phase0 platform"' \
  'emit_declared_support_phase "backend-integration support phase1"' \
  'emit_declared_support_phase "backend-integration support phase2"' \
  'emit_declared_support_phase "backend-integration support phase3"' \
  'emit_declared_support_phase "backend-integration support phase4 entities"'
do
  if ! grep -Fq "$expected" "$go_runner_script"; then
    fail "scripts/run-go-target.sh must preserve backend-integration selection surface: missing $expected"
  fi
done

backend_process_block="$(extract_target_block backend-process)"
if ! printf '%s\n' "$backend_process_block" | grep -Fq 'run-go-target.sh backend-process'; then
  fail "backend-process must delegate to scripts/run-go-target.sh backend-process"
fi
if ! printf '%s\n' "$backend_process_block" | grep -Fq '$(TEST_SERVICES_BIN) run --'; then
  fail "backend-process must run through $(TEST_SERVICES_BIN) for suite-scoped services"
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
if ! printf '%s\n' "$phase0_process_block" | grep -Fq '$(TEST_SERVICES_BIN) run --'; then
  fail "phase0-process-e2e must run through $(TEST_SERVICES_BIN) for suite-scoped services"
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
if ! printf '%s\n' "$phase1_process_block" | grep -Fq '$(TEST_SERVICES_BIN) run --'; then
  fail "phase1-process-smoke must run through $(TEST_SERVICES_BIN) for suite-scoped services"
fi

phase2_process_block="$(extract_target_block phase2-process-smoke)"
if ! printf '%s\n' "$phase2_process_block" | grep -Fq 'CARTULARY_SERVER_BIN=$(SERVER_BIN)'; then
  fail "phase2-process-smoke must export CARTULARY_SERVER_BIN=$(SERVER_BIN)"
fi
if ! printf '%s\n' "$phase2_process_block" | grep -Fq '$(TEST_SERVICES_BIN) run --'; then
  fail "phase2-process-smoke must run through $(TEST_SERVICES_BIN) for suite-scoped services"
fi

for target in backend-process phase0-process-e2e phase1-process-smoke phase2-process-smoke; do
  prereqs="$(extract_target_prereqs "$target")"
  if [[ -z "$prereqs" || "$prereqs" != *build-server* ]]; then
    fail "$target must depend on build-server"
  fi
done

test_service_images_block="$(extract_target_block test-service-images)"
if [[ -z "$test_service_images_block" ]]; then
  fail "Makefile must define a non-empty test-service-images block"
fi
if ! printf '%s\n' "$test_service_images_block" | grep -Fq '$(TEST_SERVICES_BIN) warm-images'; then
  fail "test-service-images must warm pinned service images through $(TEST_SERVICES_BIN)"
fi

check_heavy_prereqs="$(extract_target_prereqs check-heavy)"
if ! printf '%s\n' "$check_heavy_prereqs" | rg -q '(^|[[:space:]])test-service-images($|[[:space:]])'; then
  fail "check-heavy must warm service images during the parallel-safe check stage"
fi

check_service_prereqs="$(extract_target_prereqs check-service-backed)"
if ! printf '%s\n' "$check_service_prereqs" | rg -q '(^|[[:space:]])test-service-images($|[[:space:]])'; then
  fail "check-service-backed must depend on test-service-images for direct runs"
fi

check_service_block="$(extract_target_block check-service-backed)"
if [[ -z "$check_service_block" ]]; then
  fail "Makefile must define a non-empty check-service-backed block"
fi
if ! printf '%s\n' "$check_service_block" | grep -Fq '$(TEST_SERVICES_BIN) run --'; then
  fail "check-service-backed must wrap the shared service-backed lane block through $(TEST_SERVICES_BIN)"
fi
if ! printf '%s\n' "$check_service_block" | grep -Fq '$(RUN_PHASE)'; then
  fail "check-service-backed must report the shared service wrapper through RUN_PHASE"
fi
if ! printf '%s\n' "$check_service_block" | grep -Fq 'target-summary check-service-backed pass --children "$(CHECK_SERVICE_BACKED_CHILD_TARGETS)"'; then
  fail "check-service-backed must emit its success target summary"
fi
if ! printf '%s\n' "$check_service_block" | grep -Fq 'target-summary check-service-backed fail --children "$(CHECK_SERVICE_BACKED_CHILD_TARGETS)"'; then
  fail "check-service-backed must emit its failure target summary"
fi

for lane in check-service-backed-lane-a check-service-backed-lane-b; do
  if ! printf '%s\n' "$check_service_block" | grep -Fq "$lane"; then
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
check_service_backend_lane_text="${check_service_lane_a_block}"$'\n'"${check_service_lane_b_prereqs}"$'\n'"${check_service_lane_b_block}"
assert_text_contains_targets "check-service-backed lanes" "$check_service_backend_lane_text" "${target_plan_service_backed_safe_targets[@]}"
assert_text_excludes_targets "check-service-backed lanes" "$check_service_backend_lane_text" "${target_plan_check_heavy_targets[@]}" "${target_plan_service_backed_unsafe_targets[@]}"
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
if ! printf '%s\n' "$backend_integration_support_block" | grep -Fq '$(TEST_SERVICES_BIN) run --'; then
  fail "backend-integration-support must run through $(TEST_SERVICES_BIN) for suite-scoped services"
fi
for shared_report in \
  backend-integration-phase0-platform \
  backend-integration-phase2-incidents \
  backend-integration-phase3-timeline \
  backend-integration-phase4-entities \
  backend-integration-auth
do
  require_shared_command_match "$shared_report" backend-integration backend-integration-support
  mapfile -t backend_integration_support_patterns < <(target_plan_support_patterns backend-integration-support "$shared_report")
  if [[ "${#backend_integration_support_patterns[@]}" -eq 0 ]]; then
    fail "$shared_report must have declared support selectors"
  fi
  for pattern in "${backend_integration_support_patterns[@]}"; do
    [[ -z "$pattern" ]] && continue
    require_shared_command_contains backend-integration "$shared_report" "$pattern"
  done
done
require_shared_command_match backend-process-shared backend-process phase0-process-e2e phase1-process-smoke phase2-process-smoke

test_fast_block="$(extract_target_block test-fast)"
if [[ -z "$test_fast_block" ]]; then
  fail "Makefile must define a non-empty test-fast block"
fi
if ! printf '%s\n' "$test_fast_block" | rg -q '(^|[[:space:]])test-fast-service-backed($|[[:space:]])'; then
  fail "test-fast must invoke test-fast-service-backed"
fi

test_fast_service_block="$(extract_target_block test-fast-service-backed)"
if [[ -z "$test_fast_service_block" ]]; then
  fail "Makefile must define a non-empty test-fast-service-backed block"
fi
if ! printf '%s\n' "$test_fast_service_block" | grep -Fq '$(TEST_SERVICES_BIN) run --'; then
  fail "test-fast-service-backed must wrap the shared service-backed lane block through $(TEST_SERVICES_BIN)"
fi
if ! printf '%s\n' "$test_fast_service_block" | grep -Fq '$(RUN_PHASE)'; then
  fail "test-fast-service-backed must report the shared service wrapper through RUN_PHASE"
fi
if ! printf '%s\n' "$test_fast_service_block" | grep -Fq 'target-summary test-fast-service-backed pass --children "$(TEST_FAST_SERVICE_BACKED_CHILD_TARGETS)"'; then
  fail "test-fast-service-backed must emit its success target summary"
fi
if ! printf '%s\n' "$test_fast_service_block" | grep -Fq 'target-summary test-fast-service-backed fail --children "$(TEST_FAST_SERVICE_BACKED_CHILD_TARGETS)"'; then
  fail "test-fast-service-backed must emit its failure target summary"
fi
for lane in test-fast-service-backed-lane-a test-fast-service-backed-lane-b; do
  if ! printf '%s\n' "$test_fast_service_block" | grep -Fq "$lane"; then
    fail "test-fast-service-backed must invoke $lane"
  fi
done
if printf '%s\n' "$test_fast_service_block" | rg -q '(^|[[:space:]])(backend-process-support|phase2-process-smoke)($|[[:space:]])'; then
  fail "test-fast-service-backed must not invoke Phase 2 process smoke coverage"
fi

test_fast_lane_a_block="$(extract_target_block test-fast-service-backed-lane-a)"
for target in backend-integration backend-integration-support; do
  if ! printf '%s\n' "$test_fast_lane_a_block" | rg -q "(^|[[:space:]])$target($|[[:space:]])"; then
    fail "test-fast-service-backed-lane-a must invoke $target"
  fi
done

test_fast_lane_b_prereqs="$(extract_target_prereqs test-fast-service-backed-lane-b)"
test_fast_lane_b_block="$(extract_target_block test-fast-service-backed-lane-b)"
test_fast_backend_lane_text="${test_fast_lane_a_block}"$'\n'"${test_fast_lane_b_prereqs}"$'\n'"${test_fast_lane_b_block}"
assert_text_contains_targets "test-fast-service-backed lanes" "$test_fast_backend_lane_text" "${target_plan_service_backed_safe_targets[@]}"
assert_text_excludes_targets "test-fast-service-backed lanes" "$test_fast_backend_lane_text" "${target_plan_check_heavy_targets[@]}" "${target_plan_service_backed_unsafe_targets[@]}"
for target in backend-store backend-process; do
  if ! printf '%s\n' "$test_fast_lane_b_prereqs" | rg -q "(^|[[:space:]])$target($|[[:space:]])"; then
    fail "test-fast-service-backed-lane-b must invoke $target"
  fi
done
if printf '%s\n' "${test_fast_lane_b_prereqs} ${test_fast_lane_b_block}" | rg -q '(^|[[:space:]])phase2-process-smoke($|[[:space:]])'; then
  fail "test-fast-service-backed-lane-b must not invoke Phase 2 process smoke coverage"
fi
