#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
makefile="$repo_root/Makefile"
generated_make="$repo_root/tools/task_surface.generated.mk"
cartulary_runner_script="$repo_root/scripts/cartulary-runner.mjs"
go_runner_script="$repo_root/scripts/run-go-target.sh"
schedule_manifest="$repo_root/tools/service_backed_schedule_manifest.json"
check_schedule_manifest="$repo_root/tools/check_schedule_manifest.json"
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

require_service_backed_schedule_target() {
  local target="$1"
  local phase_label="$2"
  local require_migrate="$3"
  local block
  local prereqs

  block="$(extract_target_block "$target")"
  if [[ -z "$block" ]]; then
    fail "Makefile must define a non-empty $target block"
  fi
  if ! printf '%s\n' "$block" | grep -Fq 'service-backed-target'; then
    fail "$target must delegate through the canonical service-backed target runner"
  fi
  if ! printf '%s\n' "$block" | grep -Fq -- "--target $target"; then
    fail "$target must pass its target to the service-backed target runner"
  fi
  if ! printf '%s\n' "$block" | grep -Fq -- "--phase-label \"$phase_label\""; then
    fail "$target must pass phase label=$phase_label to the service-backed target runner"
  fi
  if ! printf '%s\n' "$block" | grep -Fq -- "--service-wrapper test-services"; then
    fail "$target must run through the test-services service wrapper"
  fi
  if printf '%s\n' "$block" | grep -Fq -- '--jobs'; then
    fail "$target must not pass a fixed scheduler job cap"
  fi
  if printf '%s\n' "$block" | rg -q "$target-lane-[ab]|(^|[[:space:]])(backend-process-support|phase2-process-smoke)($|[[:space:]])"; then
    fail "$target must not invoke fixed legacy targets or Phase 2 process smoke coverage"
  fi

  prereqs="$(extract_target_prereqs "$target")"
  if ! printf '%s\n' "$prereqs" | rg -q '(^|[[:space:]])test-service-images($|[[:space:]])'; then
    fail "$target must depend on test-service-images for direct runs"
  fi
  if ! printf '%s\n' "$prereqs" | rg -q '(^|[[:space:]])build-server($|[[:space:]])'; then
    fail "$target must prebuild server before service-backed scheduling"
  fi
  if [[ "$require_migrate" == "1" ]] && ! printf '%s\n' "$prereqs" | rg -q '(^|[[:space:]])build-migrate($|[[:space:]])'; then
    fail "$target must prebuild migrate before service-backed scheduling"
  fi
  if [[ "$require_migrate" == "0" ]] && printf '%s\n' "$prereqs" | rg -q '(^|[[:space:]])build-migrate($|[[:space:]])'; then
    fail "$target must not require build-migrate"
  fi
}

inspect_shared_command() {
  local target="$1"
  local family="$2"
  NODE_BIN="$node_bin" "$node_bin" "$cartulary_runner_script" go-target inspect-aggregate-command "$target" "$family"
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
  local execution_family="$2"

  "$node_bin" - "$target_plan_file" "$target" "$execution_family" <<'EOF'
const fs = require("node:fs");

const [planFile, target, executionFamily] = process.argv.slice(2);
const rows = JSON.parse(fs.readFileSync(planFile, "utf8"));
const patterns = new Set();
for (const row of rows) {
  if (
    row.target === target &&
    row.execution_family === executionFamily &&
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

schedule_targets() {
  local schedule_target="$1"
  local kind="${2:-}"

  "$node_bin" - "$schedule_manifest" "$schedule_target" "$kind" <<'EOF'
const fs = require("node:fs");

const [manifestFile, scheduleTarget, kind] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
if (manifest.schema_id !== "cartulary.service_backed_schedule.v8") {
  throw new Error("service-backed schedule manifest must declare schema_id=cartulary.service_backed_schedule.v8");
}
const schedules = manifest.schedules.filter((entry) => entry.target === scheduleTarget);
if (schedules.length !== 1) {
  throw new Error(`expected exactly one schedule for ${scheduleTarget}, found ${schedules.length}`);
}
const targets = schedules[0].work_unit_sources
  .filter((entry) => kind === "" || entry.class === kind)
  .map((entry) => entry.target);
if (targets.length > 0) {
  process.stdout.write(`${targets.join("\n")}\n`);
}
EOF
}

check_schedule_field() {
  local work_unit="$1"
  local field="$2"

  "$node_bin" - "$check_schedule_manifest" "$work_unit" "$field" <<'EOF'
const fs = require("node:fs");

const [manifestFile, workUnit, field] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
if (manifest.schema_id !== "cartulary.check_schedule.v6") {
  throw new Error("check schedule manifest must declare schema_id=cartulary.check_schedule.v6");
}
const schedules = manifest.schedules.filter((entry) => entry.target === "check");
if (schedules.length !== 1) {
  throw new Error(`expected exactly one check schedule, found ${schedules.length}`);
}
const units = schedules[0].work_units ?? [];
const unit = units.find((entry) => entry.target === workUnit);
if (!unit) {
  throw new Error(`missing check schedule work unit ${workUnit}`);
}
const value = unit[field];
if (Array.isArray(value)) {
  process.stdout.write(value.join(","));
} else if (value && typeof value === "object") {
  process.stdout.write(Object.keys(value).sort().join(","));
} else if (value !== undefined) {
  process.stdout.write(String(value));
}
EOF
}

schedule_child_array() {
  local schedule_target="$1"
  local child_target="$2"
  local field="$3"

  "$node_bin" - "$schedule_manifest" "$schedule_target" "$child_target" "$field" <<'EOF'
const fs = require("node:fs");

const [manifestFile, scheduleTarget, childTarget, field] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
if (manifest.schema_id !== "cartulary.service_backed_schedule.v8") {
  throw new Error("service-backed schedule manifest must declare schema_id=cartulary.service_backed_schedule.v8");
}
const schedules = manifest.schedules.filter((entry) => entry.target === scheduleTarget);
if (schedules.length !== 1) {
  throw new Error(`expected exactly one schedule for ${scheduleTarget}, found ${schedules.length}`);
}
const children = schedules[0].work_unit_sources.filter((entry) => entry.target === childTarget);
if (children.length !== 1) {
  throw new Error(`expected exactly one child ${childTarget} in ${scheduleTarget}, found ${children.length}`);
}
const value = children[0][field] ?? {};
if (Array.isArray(value) || typeof value !== "object") {
  throw new Error(`${scheduleTarget} child ${childTarget} ${field} must be an object`);
}
const values = Object.keys(value).sort();
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

"$node_bin" - "$schedule_manifest" <<'EOF'
const fs = require("node:fs");

const [manifestFile] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
if (manifest.schema_id !== "cartulary.service_backed_schedule.v8") {
  throw new Error("service-backed schedule manifest must declare schema_id=cartulary.service_backed_schedule.v8");
}
for (const schedule of manifest.schedules ?? []) {
  const limits = schedule.resource_limits ?? {};
  if (Object.hasOwn(limits, "backend")) {
    throw new Error(`${schedule.target} must not declare removed generic backend resource limit`);
  }
  if (Object.hasOwn(limits, "browser")) {
    throw new Error(`${schedule.target} must not declare removed generic browser resource limit`);
  }
  for (const resource of ["go_cpu", "go_io"]) {
    if (!Object.hasOwn(limits, resource)) {
      throw new Error(`${schedule.target} must declare ${resource} resource limit`);
    }
  }
  for (const source of schedule.work_unit_sources ?? []) {
    const claims = source.resource_claims ?? {};
    if (Object.hasOwn(claims, "backend")) {
      throw new Error(`${schedule.target} ${source.target} must not claim removed generic backend resource`);
    }
    if (Object.hasOwn(claims, "browser")) {
      throw new Error(`${schedule.target} ${source.target} must not claim removed generic browser resource`);
    }
    if (source.type === "go_shards" && (Object.hasOwn(claims, "go_cpu") || Object.hasOwn(claims, "go_io"))) {
      throw new Error(`${schedule.target} ${source.target} go shard source must leave go_cpu/go_io to per-shard scheduler profiles`);
    }
    if (source.class === "browser") {
      const browserStage = String(source.browser_stage ?? "");
      const laneResource = `browser_stage_${browserStage.replaceAll("-", "_")}`;
      if (!Object.hasOwn(limits, "browser_stack")) {
        throw new Error(`${schedule.target} must declare browser_stack resource limit for browser work`);
      }
      if (!Object.hasOwn(limits, laneResource)) {
        throw new Error(`${schedule.target} must declare ${laneResource} resource limit for ${source.target}`);
      }
      if (!Object.hasOwn(claims, "browser_stack")) {
        throw new Error(`${schedule.target} ${source.target} must claim browser_stack`);
      }
      if (!Object.hasOwn(claims, laneResource)) {
        throw new Error(`${schedule.target} ${source.target} must claim ${laneResource}`);
      }
    }
  }
}
EOF

check_build_prereqs_line="$(sed -n 's/^check-build-prereqs:[[:space:]]*//p' "$makefile" | head -n 1)"
if [[ -z "$check_build_prereqs_line" ]]; then
  fail "Makefile must define check-build-prereqs prerequisites"
fi
check_local_product_line="$(sed -n 's/^check-local-product:[[:space:]]*//p' "$makefile" | head -n 1)"
if [[ -z "$check_local_product_line" ]]; then
  fail "Makefile must define check-local-product prerequisites"
fi
check_meta_validation_line="$(sed -n 's/^check-meta-validation:[[:space:]]*//p' "$makefile" | head -n 1)"
if [[ -z "$check_meta_validation_line" ]]; then
  fail "Makefile must define check-meta-validation prerequisites"
fi
if ! printf '%s\n' "$check_meta_validation_line" | rg -q '(^|[[:space:]])check-static-validation($|[[:space:]])'; then
  fail "check-meta-validation must include static validation"
fi
if ! printf '%s\n' "$check_meta_validation_line" | rg -q '(^|[[:space:]])check-harness-smoke($|[[:space:]])'; then
  fail "check-meta-validation must include harness smoke"
fi

assert_text_contains_targets "check-local-product prerequisites" "$check_local_product_line" "${target_plan_check_heavy_targets[@]}"
assert_text_excludes_targets "check-local-product prerequisites" "$check_local_product_line" "${target_plan_service_backed_safe_targets[@]}" "${target_plan_service_backed_unsafe_targets[@]}"

if ! [[ -f "$check_schedule_manifest" ]]; then
  fail "missing tools/check_schedule_manifest.json"
fi
check_block="$(extract_target_block check)"
if [[ -z "$check_block" ]]; then
  fail "Makefile must define a non-empty check block"
fi
if ! printf '%s\n' "$check_block" | grep -Fq '$(RUN_CHECK_SCHEDULE_SCRIPT)'; then
  fail "check must delegate to the check scheduler"
fi
if ! printf '%s\n' "$check_block" | grep -Fq -- '--resource-limit host_cpu=$(CHECK_HOST_CPU_JOBS)'; then
  fail "check must pass CHECK_HOST_CPU_JOBS as the check scheduler host_cpu resource limit"
fi
if ! printf '%s\n' "$check_block" | grep -Fq -- '--resource-limit host_io=$(CHECK_HOST_IO_JOBS)'; then
  fail "check must pass CHECK_HOST_IO_JOBS as the check scheduler host_io resource limit"
fi
if printf '%s\n' "$check_block" | grep -Fq '$(RUN_MAKE_SEQUENCE_SCRIPT)'; then
  fail "check must not use the serial make sequence runner"
fi
if printf '%s\n' "$check_block" | grep -Fq -- '--step browser-e2e'; then
  fail "check must not run browser-e2e as a final serial step"
fi
if rg -q '^check-pre-browser:' "$makefile"; then
  fail "check-pre-browser must not remain as legacy check orchestration"
fi
for scheduled_target in check-setup-blockers check-build-prereqs check-service-backed check-go-test-duration-baseline-drift check-local-product check-frontend-unit check-meta-validation; do
  check_schedule_field "$scheduled_target" target >/dev/null
done
"$node_bin" - "$check_schedule_manifest" <<'EOF'
const fs = require("node:fs");

const [manifestFile] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
if (manifest.schema_id !== "cartulary.check_schedule.v6") {
  throw new Error("check schedule manifest must declare schema_id=cartulary.check_schedule.v6");
}
const schedules = manifest.schedules.filter((entry) => entry.target === "check");
if (schedules.length !== 1) {
  throw new Error(`expected exactly one check schedule, found ${schedules.length}`);
}
const schedule = schedules[0];
if ((schedule.work_units ?? []).some((entry) => entry.target === "browser-e2e")) {
  throw new Error("browser-e2e must be service-backed scheduler work, not a top-level check work unit");
}
const limits = schedule.resource_limits ?? {};
if (limits.host_cpu !== 12 || limits.host_io !== 12 || limits.service_stack !== 1) {
  throw new Error("check schedule must declare host_cpu, host_io, and service_stack limits");
}
const service = (schedule.work_units ?? []).find((entry) => entry.target === "check-service-backed");
if (!service) {
  throw new Error("missing check-service-backed work unit");
}
const claims = service.resource_claims ?? {};
const assertBoundedClaim = (claim, resource, expected) => {
  if (!claim || typeof claim !== "object" || Array.isArray(claim)) {
    throw new Error(`check-service-backed ${resource} claim must use bounded_limit`);
  }
  const keys = Object.keys(claim).sort().join(",");
  if (keys !== "max,min,mode,reserve") {
    throw new Error(`check-service-backed ${resource} bounded claim has unexpected keys ${keys}`);
  }
  for (const [key, value] of Object.entries(expected)) {
    if (claim[key] !== value) {
      throw new Error(`check-service-backed ${resource}.${key} got ${claim[key]} want ${value}`);
    }
  }
};
assertBoundedClaim(claims.host_cpu, "host_cpu", { mode: "bounded_limit", reserve: 3, min: 1, max: 8 });
assertBoundedClaim(claims.host_io, "host_io", { mode: "bounded_limit", reserve: 4, min: 1, max: 10 });
if (claims.service_stack !== 1) {
  throw new Error("check-service-backed must claim exclusive service_stack");
}
const nested = service.nested_scheduler ?? {};
if (nested.type !== "service_backed" || nested.target !== "check-service-backed") {
  throw new Error("check-service-backed must declare service_backed nested scheduler metadata");
}
if (nested.manifest !== "tools/service_backed_schedule_manifest.json") {
  throw new Error("check-service-backed nested scheduler must point at the service-backed manifest");
}
if (nested.forwarding !== "check_host_to_service_backed_go") {
  throw new Error("check-service-backed nested scheduler must use the host-to-service-backed forwarding profile");
}
EOF
if [[ "$(check_schedule_field check-build-prereqs needs)" != "check-setup-blockers" ]]; then
  fail "check-build-prereqs must depend on check-setup-blockers in the check schedule"
fi
for scheduled_target in check-service-backed check-local-product check-frontend-unit check-meta-validation; do
  if [[ "$(check_schedule_field "$scheduled_target" needs)" != "check-build-prereqs" ]]; then
    fail "$scheduled_target must depend on check-build-prereqs in the check schedule"
  fi
done
if [[ "$(check_schedule_field check-go-test-duration-baseline-drift needs)" != "check-service-backed" ]]; then
  fail "check-go-test-duration-baseline-drift must depend on check-service-backed in the check schedule"
fi
if [[ "$(check_schedule_field check-service-backed resource_claims)" != "host_cpu,host_io,service_stack" ]]; then
  fail "check-service-backed must claim host_cpu, host_io, and service_stack resources in the check schedule"
fi

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

help_text="$(cat "$generated_make")"
if ! printf '%s\n' "$help_text" | grep -Fq 'make services-up'; then
  fail "help must document services-up"
fi
if ! printf '%s\n' "$help_text" | grep -Fq 'does not reset object storage'; then
  fail "help must document db-reset object-storage scope"
fi

if ! rg -q '^backend-store:' "$generated_make" "$makefile"; then
  fail "Makefile must define backend-store"
fi

if rg -q '^backend-process-support:' "$makefile"; then
  fail "Makefile must not define backend-process-support; process support belongs to backend-process manifests"
fi

for removed_target in phase0-process-e2e phase1-process-smoke phase2-process-smoke; do
  if rg -q "^${removed_target}:" "$makefile"; then
    fail "Makefile must not define removed helper target ${removed_target}"
  fi
done

if rg -q 'TestPhase0_.*_U_0_|TestPhase0_.*_I_0_|TestPhase0_.*_E_0_' "$makefile"; then
  fail "Makefile must not use regex-based Phase 0 Go selection"
fi

if rg -q 'CARTULARY_ALLOW_EMPTY_MANIFEST_SELECTION' "$makefile" "$generated_make" "$repo_root/tools/task_surface_manifest.json"; then
  fail "Make and task-surface manifests must not retain CARTULARY_ALLOW_EMPTY_MANIFEST_SELECTION"
fi

if rg -q 'TestPhase4_.*_U_4_' "$go_runner_script"; then
  fail "scripts/run-go-target.sh must not use raw authoritative Phase 4 U-4-* Go selectors"
fi

for removed_runner_fragment in \
  'manifest_go_regex' \
  'manifest_go_count' \
  'support_go_regex' \
  'backend_unit_core_shared_spec' \
  'backend_integration_phase' \
  'backend-store-shared' \
  'backend-process-shared' \
  'phase0-process-e2e' \
  'phase1-process-smoke' \
  'phase2-process-smoke'
do
  if grep -Fq "$removed_runner_fragment" "$go_runner_script"; then
    fail "scripts/run-go-target.sh must not retain phase/package-specific runner fragment: $removed_runner_fragment"
  fi
done

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
  if (
    row.execution_dependency !== entry.execution_dependency ||
    row.section !== entry.section ||
    row.execution_family !== entry.execution_family ||
    row.execution_label !== entry.execution_label
  ) {
    console.error(
      `${entry.phase} authoritative row ${entry.id} target-plan mismatch: expected ${entry.section}/${entry.execution_dependency}/${entry.execution_family}/${entry.execution_label}, found ${row.section}/${row.execution_dependency}/${row.execution_family}/${row.execution_label}`,
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
      row.execution_family === entry.execution_family &&
      row.execution_label === entry.execution_label &&
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
if ! printf '%s\n' "$backend_unit_block" | grep -Fq '$(CARTULARY_RUNNER_SCRIPT) go-target backend-unit'; then
  fail "backend-unit must delegate to cartulary-runner.mjs go-target backend-unit"
fi
for expected in \
  'target_aggregate_names "${target}"' \
  'target_aggregate_spec "${target}" "${family}"' \
  'target_aggregate_emission_count "${target}" "${family}"' \
  'emit_execution_family "${target}" "${aggregate_name}"' \
  'run_unsharded_target backend-unit' \
  'run_sharded_target backend-store' \
  'run_sharded_target backend-integration' \
  'run_sharded_target backend-integration-support' \
  'run_unsharded_target backend-process'
do
  if ! grep -Fq "$expected" "$go_runner_script"; then
    fail "scripts/run-go-target.sh must preserve generic target-plan-driven execution surface: missing $expected"
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
if ! printf '%s\n' "$backend_store_block" | grep -Fq '$(CARTULARY_RUNNER_SCRIPT) go-target backend-store'; then
  fail "backend-store must delegate to cartulary-runner.mjs go-target backend-store"
fi
if ! printf '%s\n' "$backend_store_block" | grep -Fq '$(TEST_SERVICES_BIN) run --'; then
  fail "backend-store must run through $(TEST_SERVICES_BIN) for suite-scoped services"
fi
require_shared_command_contains backend-store backend-store "TestPhase"
require_shared_command_contains backend-store backend-store "CARTULARY_POSTGRES_FIXTURE_POLICY_TESTS"

backend_integration_block="$(extract_target_block backend-integration)"
if ! printf '%s\n' "$backend_integration_block" | grep -Fq '$(CARTULARY_RUNNER_SCRIPT) go-target backend-integration'; then
  fail "backend-integration must delegate to cartulary-runner.mjs go-target backend-integration"
fi
if ! printf '%s\n' "$backend_integration_block" | grep -Fq '$(TEST_SERVICES_BIN) run --'; then
  fail "backend-integration must run through $(TEST_SERVICES_BIN) for suite-scoped services"
fi
for expected in \
  'planned_shard_names()' \
  'planned_aggregate_names()' \
  'mapfile -t shard_names < <(planned_shard_names "${target}")' \
  'mapfile -t aggregate_names < <(planned_aggregate_names "${target}")' \
  'capture_named_shared_reports_parallel "${target}" "${BACKEND_INTEGRATION_SHARD_JOBS}"' \
  'finalize_scheduled_shards "${target}"'
do
  if ! grep -Fq "$expected" "$go_runner_script"; then
    fail "scripts/run-go-target.sh must preserve planned backend-integration selection surface: missing $expected"
  fi
done

if ! "$node_bin" - "$repo_root" "$node_bin" <<'EOF'
const { execFileSync } = require("node:child_process");
const path = require("node:path");
const [root, nodeBin] = process.argv.slice(2);
const plan = JSON.parse(execFileSync(nodeBin, [path.join(root, "scripts/print-go-shard-plan.mjs"), "--json"], { encoding: "utf8", cwd: root }));
const heavyIntegrationAggregates = plan.aggregates.filter(
  (aggregate) => aggregate.target === "backend-integration" && aggregate.weight_ms > 18000,
);
if (
  heavyIntegrationAggregates.length === 0 ||
  heavyIntegrationAggregates.some((aggregate) => aggregate.shards.length < 2)
) {
  process.exit(1);
}
if (!plan.shards.every((shard) => Number.isInteger(shard.shard_target_ms) && shard.shard_target_ms > 0)) {
  process.exit(1);
}
const integrationMultiItemShards = plan.shards.filter((shard) => shard.shard_target_ms === 18000 && (shard.has_authoritative || shard.has_support) && shard.item_count > 1);
if (!integrationMultiItemShards.every((shard) => shard.weight_ms <= 18000 && shard.shard_target_ms === 18000)) {
  process.exit(1);
}
const weights = plan.shards
  .filter((shard) => shard.has_authoritative || shard.has_raw)
  .map((shard) => shard.weight_ms);
for (let index = 1; index < weights.length; index += 1) {
  if (weights[index - 1] < weights[index]) {
    process.exit(1);
  }
}
EOF
then
  fail "backend-integration shard plan must split heavy aggregates, enforce integration shard targets, and schedule longest shards first"
fi

backend_process_block="$(extract_target_block backend-process)"
if ! printf '%s\n' "$backend_process_block" | grep -Fq '$(CARTULARY_RUNNER_SCRIPT) go-target backend-process'; then
  fail "backend-process must delegate to cartulary-runner.mjs go-target backend-process"
fi
if ! printf '%s\n' "$backend_process_block" | grep -Fq '$(TEST_SERVICES_BIN) run --'; then
  fail "backend-process must run through $(TEST_SERVICES_BIN) for suite-scoped services"
fi
if ! printf '%s\n' "$backend_process_block" | grep -Fq 'CARTULARY_SERVER_BIN=$(SERVER_BIN)'; then
  fail "backend-process must export CARTULARY_SERVER_BIN=$(SERVER_BIN)"
fi
backend_process_command="$(inspect_shared_command backend-process backend-process)"
for expected in TestPhase0_ TestPhase1_ TestPhase2_; do
  if [[ "$backend_process_command" != *"$expected"* ]]; then
    fail "backend-process aggregate must include manifest-owned process coverage matching $expected"
  fi
done

for target in backend-process; do
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

check_build_prereqs="$(extract_target_prereqs check-build-prereqs)"
if ! printf '%s\n' "$check_build_prereqs" | rg -q '(^|[[:space:]])test-service-images($|[[:space:]])'; then
  fail "check-build-prereqs must warm service images before pre-browser check work"
fi

if [[ ! -x "$cartulary_runner_script" ]]; then
  fail "missing executable scripts/cartulary-runner.mjs"
fi
service_schedule_target_content="$(cat "$cartulary_runner_script")"
for required_service_schedule_fragment in \
  'TEST_SERVICES_BIN' \
  'service-backed-target' \
  'context.runPhaseScript' \
  'context.serviceBackedScheduleScript' \
  '--defer-summary' \
  '"--children"'
do
  if [[ "$service_schedule_target_content" != *"$required_service_schedule_fragment"* ]]; then
    fail "scripts/cartulary-runner.mjs must contain $required_service_schedule_fragment"
  fi
done
if [[ "$service_schedule_target_content" == *'--jobs'* ]]; then
  fail "scripts/cartulary-runner.mjs must not pass a fixed scheduler job cap"
fi

require_service_backed_schedule_target check-service-backed "check service-backed" 1

mapfile -t check_service_browser_schedule_targets < <(schedule_targets check-service-backed browser)
if [[ "$(printf '%s\n' "${check_service_browser_schedule_targets[@]}")" != $'browser-e2e-webserver-backed\nbrowser-e2e' ]]; then
  fail "check-service-backed schedule must own webserver-backed and isolated browser work, found: ${check_service_browser_schedule_targets[*]:-none}"
fi

mapfile -t check_service_backend_schedule_targets < <(schedule_targets check-service-backed backend)
check_service_backend_schedule_text="$(printf '%s\n' "${check_service_backend_schedule_targets[@]}")"
assert_text_contains_targets "check-service-backed schedule" "$check_service_backend_schedule_text" "${target_plan_service_backed_safe_targets[@]}"
assert_text_excludes_targets "check-service-backed schedule" "$check_service_backend_schedule_text" "${target_plan_check_heavy_targets[@]}" "${target_plan_service_backed_unsafe_targets[@]}"

backend_integration_support_block="$(extract_target_block backend-integration-support)"
if ! printf '%s\n' "$backend_integration_support_block" | grep -Fq '$(CARTULARY_RUNNER_SCRIPT) go-target backend-integration-support'; then
  fail "backend-integration-support must delegate to cartulary-runner.mjs go-target backend-integration-support"
fi
if ! printf '%s\n' "$backend_integration_support_block" | grep -Fq '$(TEST_SERVICES_BIN) run --'; then
  fail "backend-integration-support must run through $(TEST_SERVICES_BIN) for suite-scoped services"
fi
for shared_report in \
  backend-integration-platform \
  backend-integration-incidents \
  backend-integration-timeline \
  backend-integration-entities \
  backend-integration-auth
do
  mapfile -t backend_integration_support_patterns < <(target_plan_support_patterns backend-integration-support "$shared_report")
  if [[ "${#backend_integration_support_patterns[@]}" -eq 0 ]]; then
    fail "$shared_report must have declared support selectors"
  fi
  for pattern in "${backend_integration_support_patterns[@]}"; do
    [[ -z "$pattern" ]] && continue
    require_shared_command_contains backend-integration-support "$shared_report" "$pattern"
  done
done

test_fast_block="$(extract_target_block test-fast)"
if [[ -z "$test_fast_block" ]]; then
  fail "Makefile must define a non-empty test-fast block"
fi
if ! printf '%s\n' "$test_fast_block" | rg -q '(^|[[:space:]])test-fast-service-backed($|[[:space:]])'; then
  fail "test-fast must invoke test-fast-service-backed"
fi

require_service_backed_schedule_target test-fast-service-backed "test-fast service-backed" 0

mapfile -t test_fast_backend_schedule_targets < <(schedule_targets test-fast-service-backed backend)
test_fast_backend_schedule_text="$(printf '%s\n' "${test_fast_backend_schedule_targets[@]}")"
assert_text_contains_targets "test-fast-service-backed schedule" "$test_fast_backend_schedule_text" "${target_plan_service_backed_safe_targets[@]}"
assert_text_excludes_targets "test-fast-service-backed schedule" "$test_fast_backend_schedule_text" "${target_plan_check_heavy_targets[@]}" "${target_plan_service_backed_unsafe_targets[@]}"
