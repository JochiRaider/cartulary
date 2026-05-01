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
cartulary_runner_script="$repo_root/scripts/cartulary-runner.mjs"
go_runner_module="$repo_root/scripts/lib/go-target-runner.mjs"
go_target_plan_coverage_helper="$repo_root/scripts/check-go-target-plan-coverage.mjs"
schedule_manifest="$repo_root/tools/service_backed_schedule_manifest.json"
execution_topology_manifest="$repo_root/tools/execution_topology_manifest.json"
schedule_topology_helper="$repo_root/scripts/lib/service-backed-schedule-topology.mjs"
check_schedule_manifest="$repo_root/tools/check_schedule_manifest.json"
node_bin="${NODE_BIN:-node}"

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
  if ! text_contains "$block" 'service-backed-target'; then
    fail "$target must delegate through the canonical service-backed target runner"
  fi
  if ! text_contains "$block" "--target $target"; then
    fail "$target must pass its target to the service-backed target runner"
  fi
  if ! text_contains "$block" "--phase-label \"$phase_label\""; then
    fail "$target must pass phase label=$phase_label to the service-backed target runner"
  fi
  if ! text_contains "$block" "--service-wrapper test-services"; then
    fail "$target must run through the test-services service wrapper"
  fi
  if text_contains "$block" '--jobs'; then
    fail "$target must not pass a fixed scheduler job cap"
  fi
  prereqs="$(extract_target_prereqs "$target")"
  if ! text_matches "$prereqs" '(^|[[:space:]])test-service-images($|[[:space:]])'; then
    fail "$target must depend on test-service-images for direct runs"
  fi
  if ! text_matches "$prereqs" '(^|[[:space:]])build-server($|[[:space:]])'; then
    fail "$target must prebuild server before service-backed scheduling"
  fi
  if [[ "$require_migrate" == "1" ]] && ! text_matches "$prereqs" '(^|[[:space:]])build-migrate($|[[:space:]])'; then
    fail "$target must prebuild migrate before service-backed scheduling"
  fi
  if [[ "$require_migrate" == "0" ]] && text_matches "$prereqs" '(^|[[:space:]])build-migrate($|[[:space:]])'; then
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

check_schedule_targets() {
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
const targets = (schedules[0].work_units ?? []).map((entry) => entry.target);
if (targets.length > 0) {
  process.stdout.write(`${targets.join("\n")}\n`);
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
    if ! text_matches "$text" "(^|[[:space:]])$target($|[[:space:]])"; then
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
    if text_matches "$text" "(^|[[:space:]])$target($|[[:space:]])"; then
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

"$node_bin" "$schedule_topology_helper" validate "$schedule_manifest" "$execution_topology_manifest"

check_build_prereqs_line="$(extract_target_prereqs check-build-prereqs)"
if [[ -z "$check_build_prereqs_line" ]]; then
  fail "Makefile must define check-build-prereqs prerequisites"
fi
for removed_check_bundle in check-static-validation check-local-product check-meta-validation; do
  if target_exists "$removed_check_bundle"; then
    fail "$removed_check_bundle must not remain as a scheduled check bundle target"
  fi
done

if ! [[ -f "$check_schedule_manifest" ]]; then
  fail "missing tools/check_schedule_manifest.json"
fi
check_block="$(extract_target_block check)"
if [[ -z "$check_block" ]]; then
  fail "Makefile must define a non-empty check block"
fi
if ! text_contains "$check_block" '$(RUN_CHECK_SCHEDULE_SCRIPT)'; then
  fail "check must delegate to the check scheduler"
fi
if ! text_contains "$check_block" '$(TASK_SURFACE_CHECK_SCHEDULER_OVERRIDE_ENV)'; then
  fail "check must forward registry-declared scheduler override env when explicitly set"
fi
if text_contains "$check_block" '--resource-limit host_cpu=' || text_contains "$check_block" '--resource-limit host_io='; then
  fail "check must not pass default host scheduler capacity as CLI resource-limit overrides"
fi
if text_contains "$check_block" '$(RUN_MAKE_SEQUENCE_SCRIPT)'; then
  fail "check must not use the serial make sequence runner"
fi
if text_contains "$check_block" '--step browser-e2e'; then
  fail "check must not run browser-e2e as a final serial step"
fi
lint_go_prereqs="$(extract_target_prereqs lint-go)"
if ! text_matches "$lint_go_prereqs" '(^|[[:space:]])lint-go-format($|[[:space:]])'; then
  fail "lint-go must delegate to lint-go-format"
fi
if ! text_matches "$lint_go_prereqs" '(^|[[:space:]])lint-go-vet($|[[:space:]])'; then
  fail "lint-go must delegate to lint-go-vet"
fi
if ! text_matches "$lint_go_prereqs" '(^|[[:space:]])lint-go-staticcheck($|[[:space:]])'; then
  fail "lint-go must delegate to lint-go-staticcheck"
fi
lint_go_format_block="$(extract_target_block lint-go-format)"
if ! text_contains "$lint_go_format_block" 'scripts/run-go-format.sh --check'; then
  fail "lint-go-format must run the curated Go formatter wrapper in check mode"
fi
if ! text_contains "$lint_go_format_block" 'run make format to apply authored Go formatting'; then
  fail "lint-go-format must tell developers to run make format"
fi
lint_go_vet_block="$(extract_target_block lint-go-vet)"
if ! text_contains "$lint_go_vet_block" 'scripts/run-go-vet.sh'; then
  fail "lint-go-vet must run the curated Go vet wrapper"
fi
lint_go_staticcheck_prereqs="$(extract_target_prereqs lint-go-staticcheck)"
if ! text_matches "$lint_go_staticcheck_prereqs" '(^|[[:space:]])go-lint-toolchain($|[[:space:]])'; then
  fail "lint-go-staticcheck must prepare the pinned Go lint toolchain"
fi
lint_go_staticcheck_block="$(extract_target_block lint-go-staticcheck)"
if ! text_contains "$lint_go_staticcheck_block" 'scripts/run-go-staticcheck.sh'; then
  fail "lint-go-staticcheck must run the curated Staticcheck wrapper"
fi
if ! text_contains "$lint_go_staticcheck_block" 'STATICCHECK_CHECKS="$(STATICCHECK_CHECKS)"'; then
  fail "lint-go-staticcheck must pass the configured Staticcheck check set"
fi
go_security_prereqs="$(extract_target_prereqs go-security-toolchain)"
if ! text_matches "$go_security_prereqs" '(^|[[:space:]])\$\((GOVULNCHECK_BIN)\)($|[[:space:]])'; then
  fail "go-security-toolchain must prepare the pinned Govulncheck binary"
fi
if ! text_matches "$go_security_prereqs" '(^|[[:space:]])\$\((GOSEC_BIN)\)($|[[:space:]])'; then
  fail "go-security-toolchain must prepare the pinned Gosec binary"
fi
go_gosec_prereqs="$(extract_target_prereqs go-gosec-targeted)"
if ! text_matches "$go_gosec_prereqs" '(^|[[:space:]])go-security-toolchain($|[[:space:]])'; then
  fail "go-gosec-targeted must prepare the pinned Go security toolchain"
fi
go_gosec_block="$(extract_target_block go-gosec-targeted)"
if ! text_contains "$go_gosec_block" 'scripts/run-go-gosec-targeted.sh'; then
  fail "go-gosec-targeted must run the curated targeted Gosec wrapper"
fi
if ! text_contains "$go_gosec_block" 'GOSEC_RULES="$(GOSEC_RULES)"'; then
  fail "go-gosec-targeted must pass the configured Gosec rule set"
fi
if ! text_contains "$go_gosec_block" 'GOSEC_PATTERNS="$(GOSEC_PATTERNS)"'; then
  fail "go-gosec-targeted must pass the configured Gosec package patterns"
fi
go_gosec_audit_prereqs="$(extract_target_prereqs go-gosec-audit)"
if ! text_matches "$go_gosec_audit_prereqs" '(^|[[:space:]])go-security-toolchain($|[[:space:]])'; then
  fail "go-gosec-audit must prepare the pinned Go security toolchain"
fi
go_gosec_audit_block="$(extract_target_block go-gosec-audit)"
if ! text_contains "$go_gosec_audit_block" 'scripts/run-go-gosec-audit.sh'; then
  fail "go-gosec-audit must run the curated warning-only Gosec audit wrapper"
fi
if ! text_contains "$go_gosec_audit_block" 'GOSEC_AUDIT_RUNTIME_RULES="$(GOSEC_AUDIT_RUNTIME_RULES)"'; then
  fail "go-gosec-audit must pass the configured runtime Gosec audit rule set"
fi
if ! text_contains "$go_gosec_audit_block" 'GOSEC_AUDIT_SUPPORT_PATTERNS="$(GOSEC_AUDIT_SUPPORT_PATTERNS)"'; then
  fail "go-gosec-audit must pass the configured support Gosec audit package patterns"
fi
shell_lint_prereqs="$(extract_target_prereqs lint-shell)"
if ! text_matches "$shell_lint_prereqs" '(^|[[:space:]])shell-lint-toolchain($|[[:space:]])'; then
  fail "lint-shell must prepare the pinned ShellCheck toolchain"
fi
lint_shell_block="$(extract_target_block lint-shell)"
if ! text_contains "$lint_shell_block" 'scripts/run-shellcheck.sh'; then
  fail "lint-shell must run the curated warning-only ShellCheck wrapper"
fi
if ! text_contains "$lint_shell_block" 'LINT_SHELL_STRICT="$(LINT_SHELL_STRICT)"'; then
  fail "lint-shell must expose LINT_SHELL_STRICT strict-mode passthrough"
fi
for scheduled_target in \
  check-setup-blockers \
  check-build-prereqs \
  check-service-backed \
  check-go-test-duration-baseline-drift \
  check-browser-e2e-duration-baseline-drift \
  check-service-backed-make-target-duration-baseline-drift \
  migration-drift \
  deployable-shape \
  backend-unit \
  frontend-typecheck \
  lint-go \
  go-vulncheck \
  go-gosec-targeted \
  go-gosec-audit \
  frontend-unit \
  check-harness-smoke \
  lint-biome \
  frontend-import-boundary-check \
  lint-scripts \
  lint-shell \
  phase-test-name-check \
  task-surface-check \
  browser-e2e-task-surface-check \
  frontend-task-surface-check \
  backend-task-surface-check \
  phase-map-check \
  go-test-duration-baseline-coverage \
  phase-ledger-drift \
  phase-schedule-drift \
  service-backed-unit-check \
  generate-drift
do
  check_schedule_field "$scheduled_target" target >/dev/null
done
check_schedule_targets_text="$(check_schedule_targets)"
for unscheduled_lint_leaf in lint-go-format lint-go-vet lint-go-staticcheck; do
  if text_has_token "$check_schedule_targets_text" "$unscheduled_lint_leaf"; then
    fail "check schedule must keep lint-go as the scheduler-visible Go lint work unit"
  fi
done
"$node_bin" - "$check_schedule_manifest" "$execution_topology_manifest" <<'EOF'
const fs = require("node:fs");

const [manifestFile, topologyFile] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
const topology = JSON.parse(fs.readFileSync(topologyFile, "utf8"));
if (topology.schema_id !== "cartulary.execution_topology.v2") {
  throw new Error("execution topology must declare schema_id=cartulary.execution_topology.v2");
}
if (Array.isArray(topology.check_schedules)) {
  throw new Error("execution topology must own check schedule profiles, not flat schedules");
}
const profiles = topology.check_schedules?.defaults?.resource_profiles ?? {};
for (const requiredProfile of ["setup_blocker", "build_readiness_gate", "nested_service_backed_scheduler", "post_build_service_stack", "after_setup_cpu", "after_setup_cpu_io"]) {
  if (!profiles[requiredProfile]) {
    throw new Error(`execution topology must declare ${requiredProfile} check schedule profile`);
  }
}
const topologyTargets = new Map((topology.task_surface?.targets ?? []).map((entry) => [entry.name, entry]));
const assertCheckMetadata = (target, profile) => {
  const metadata = topologyTargets.get(target)?.check_schedule;
  if (!metadata?.schedules?.includes("check") || metadata.profile !== profile) {
    throw new Error(`execution topology must schedule ${target} through ${profile} profile metadata`);
  }
};
assertCheckMetadata("check-setup-blockers", "setup_blocker");
assertCheckMetadata("check-build-prereqs", "build_readiness_gate");
assertCheckMetadata("check-service-backed", "nested_service_backed_scheduler");
assertCheckMetadata("migration-drift", "post_build_service_stack");
assertCheckMetadata("backend-unit", "after_setup_cpu");
assertCheckMetadata("go-vulncheck", "after_setup_cpu_io");
if (manifest.schema_id !== "cartulary.check_schedule.v6") {
  throw new Error("check schedule manifest must declare schema_id=cartulary.check_schedule.v6");
}
const schedules = manifest.schedules.filter((entry) => entry.target === "check");
if (schedules.length !== 1) {
  throw new Error(`expected exactly one check schedule, found ${schedules.length}`);
}
const schedule = schedules[0];
if (schedule.capacity_profile !== "check_default") {
  throw new Error("check schedule must resolve capacity through check_default");
}
if ((schedule.work_units ?? []).some((entry) => entry.target === "browser-e2e")) {
  throw new Error("browser-e2e must be service-backed scheduler work, not a top-level check work unit");
}
for (const removed of ["check-static-validation", "check-local-product", "check-meta-validation"]) {
  if ((schedule.work_units ?? []).some((entry) => entry.target === removed)) {
    throw new Error(`${removed} must not remain in the check schedule`);
  }
}
const limits = schedule.resource_limits ?? {};
if (limits.host_cpu !== 12 || limits.host_io !== 12 || limits.service_stack !== 1) {
  throw new Error("check schedule must declare host_cpu, host_io, and service_stack limits");
}
const build = (schedule.work_units ?? []).find((entry) => entry.target === "check-build-prereqs");
if (!build) {
  throw new Error("missing check-build-prereqs work unit");
}
const buildCpu = build.resource_claims?.host_cpu;
if (
  !buildCpu ||
  typeof buildCpu !== "object" ||
  Array.isArray(buildCpu) ||
  buildCpu.mode !== "bounded_limit" ||
  buildCpu.reserve !== 3 ||
  buildCpu.min !== 1 ||
  buildCpu.max !== 8
) {
  throw new Error("check-build-prereqs must claim bounded host_cpu with reserve=3 min=1 max=8");
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
for scheduled_target in check-service-backed migration-drift deployable-shape; do
  if [[ "$(check_schedule_field "$scheduled_target" needs)" != "check-build-prereqs" ]]; then
    fail "$scheduled_target must depend on check-build-prereqs in the check schedule"
  fi
done
for scheduled_target in \
  backend-unit \
  frontend-typecheck \
  lint-go \
  go-vulncheck \
  go-gosec-targeted \
  go-gosec-audit \
  frontend-unit \
  check-harness-smoke \
  lint-biome \
  frontend-import-boundary-check \
  lint-scripts \
  lint-shell \
  phase-test-name-check \
  task-surface-check \
  browser-e2e-task-surface-check \
  frontend-task-surface-check \
  backend-task-surface-check \
  phase-map-check \
  go-test-duration-baseline-coverage \
  phase-ledger-drift \
  phase-schedule-drift \
  service-backed-unit-check \
  generate-drift
do
  if [[ "$(check_schedule_field "$scheduled_target" needs)" != "check-setup-blockers" ]]; then
    fail "$scheduled_target must depend on check-setup-blockers in the check schedule"
  fi
done
if [[ "$(check_schedule_field check-go-test-duration-baseline-drift needs)" != "check-service-backed" ]]; then
  fail "check-go-test-duration-baseline-drift must depend on check-service-backed in the check schedule"
fi
if [[ "$(check_schedule_field check-browser-e2e-duration-baseline-drift needs)" != "check-service-backed" ]]; then
  fail "check-browser-e2e-duration-baseline-drift must depend on check-service-backed in the check schedule"
fi
if [[ "$(check_schedule_field check-service-backed-make-target-duration-baseline-drift needs)" != "check-service-backed" ]]; then
  fail "check-service-backed-make-target-duration-baseline-drift must depend on check-service-backed in the check schedule"
fi
if [[ "$(check_schedule_field check-service-backed resource_claims)" != "host_cpu,host_io,service_stack" ]]; then
  fail "check-service-backed must claim host_cpu, host_io, and service_stack resources in the check schedule"
fi
if [[ "$(check_schedule_field migration-drift resource_claims)" != "host_cpu,host_io,service_stack" ]]; then
  fail "migration-drift must claim host_cpu, host_io, and service_stack resources in the check schedule"
fi

if grep -Fq 'rg --files' "$makefile"; then
  fail "Makefile must not use parse-time rg --files for build input discovery"
fi
if ! grep -Fq 'scripts/list-build-inputs.sh' "$makefile"; then
  fail "Makefile must use scripts/list-build-inputs.sh for build input discovery"
fi

for target in services-up services-wait postgres-wait minio-wait minio-init; do
  if ! target_exists "$target"; then
    fail "Makefile must define $target"
  fi
done

services_wait_prereqs="$(extract_target_prereqs services-wait)"
for target in postgres-wait minio-wait; do
  if ! text_matches "$services_wait_prereqs" "(^|[[:space:]])$target($|[[:space:]])"; then
    fail "services-wait must depend on $target"
  fi
done

dev_services_body="$(cat "$repo_root/scripts/dev-services.sh")"
services_up_block="$(extract_target_block services-up)"
if ! text_contains "$services_up_block" './scripts/dev-services.sh up'; then
  fail "services-up must delegate to scripts/dev-services.sh"
fi
if ! text_contains "$dev_services_body" 'compose up -d postgres minio'; then
  fail "services-up must start postgres and minio"
fi
if ! text_contains "$dev_services_body" 'wait_postgres' || ! text_contains "$dev_services_body" 'wait_minio'; then
  fail "services-up must wait for service readiness"
fi

db_up_block="$(extract_target_block db-up)"
if ! text_contains "$db_up_block" './scripts/dev-services.sh db-up'; then
  fail "db-up must delegate to scripts/dev-services.sh"
fi
if ! text_contains "$dev_services_body" 'services_up'; then
  fail "db-up must delegate service startup to services-up"
fi
if ! text_contains "$dev_services_body" 'init_minio'; then
  fail "db-up must initialize the default MinIO bucket"
fi

db_reset_block="$(extract_target_block db-reset)"
if ! text_contains "$db_reset_block" './scripts/dev-services.sh db-reset'; then
  fail "db-reset must delegate to scripts/dev-services.sh"
fi
if ! text_contains "$dev_services_body" 'wait_postgres'; then
  fail "db-reset must wait for postgres before resetting the database"
fi
if ! text_contains "$dev_services_body" 'MinIO/object storage is not reset'; then
  fail "db-reset must explicitly report that object storage is not reset"
fi

minio_init_block="$(extract_target_block minio-init)"
if ! text_contains "$minio_init_block" 'MINIO_BUCKET="$(MINIO_BUCKET)"'; then
  fail "minio-init must pass the configured MINIO_BUCKET"
fi
if ! text_contains "$minio_init_block" 'init-minio'; then
  fail "minio-init must delegate bucket creation to dev-services.sh"
fi

help_text="$(cat "$generated_make")"
if ! text_contains "$help_text" 'make services-up'; then
  fail "help must document services-up"
fi
if ! text_contains "$help_text" 'does not reset object storage'; then
  fail "help must document db-reset object-storage scope"
fi

if ! rg -q '^backend-store:' "$generated_make" "$makefile"; then
  fail "Makefile must define backend-store"
fi

if ! "$node_bin" "$go_target_plan_coverage_helper" --root "$repo_root" --quiet
then
  fail "Backend target plan must match authoritative manifests and support selectors"
fi

backend_unit_block="$(extract_target_block backend-unit)"
if ! text_contains "$backend_unit_block" '$(CARTULARY_RUNNER_SCRIPT) go-target backend-unit'; then
  fail "backend-unit must delegate to cartulary-runner.mjs go-target backend-unit"
fi
for expected in \
  'runUnshardedTarget' \
  'runShardedTarget' \
  'emitExecutionFamily' \
  'assignExecutionFamily' \
  'inspectAggregateCommand'
do
  if ! grep -Fq "$expected" "$go_runner_module"; then
    fail "scripts/lib/go-target-runner.mjs must preserve generic target-plan-driven execution surface: missing $expected"
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
if ! text_contains "$backend_store_block" '$(CARTULARY_RUNNER_SCRIPT) go-target backend-store'; then
  fail "backend-store must delegate to cartulary-runner.mjs go-target backend-store"
fi
if ! text_contains "$backend_store_block" '$(TEST_SERVICES_BIN) run --'; then
  fail "backend-store must run through $(TEST_SERVICES_BIN) for suite-scoped services"
fi
require_shared_command_contains backend-store backend-store "TestPhase"
require_shared_command_contains backend-store backend-store "CARTULARY_POSTGRES_FIXTURE_POLICY_TESTS"

backend_integration_block="$(extract_target_block backend-integration)"
if ! text_contains "$backend_integration_block" '$(CARTULARY_RUNNER_SCRIPT) go-target backend-integration'; then
  fail "backend-integration must delegate to cartulary-runner.mjs go-target backend-integration"
fi
if ! text_contains "$backend_integration_block" '$(TEST_SERVICES_BIN) run --'; then
  fail "backend-integration must run through $(TEST_SERVICES_BIN) for suite-scoped services"
fi
for expected in \
  'targetShards' \
  'targetAggregates' \
  'captureNamedSharedReportsParallel' \
  'finalizeScheduledShards'
do
  if ! grep -Fq "$expected" "$go_runner_module"; then
    fail "scripts/lib/go-target-runner.mjs must preserve planned backend-integration selection surface: missing $expected"
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
if ! text_contains "$backend_process_block" '$(CARTULARY_RUNNER_SCRIPT) go-target backend-process'; then
  fail "backend-process must delegate to cartulary-runner.mjs go-target backend-process"
fi
if ! text_contains "$backend_process_block" '$(TEST_SERVICES_BIN) run --'; then
  fail "backend-process must run through $(TEST_SERVICES_BIN) for suite-scoped services"
fi
if ! text_contains "$backend_process_block" 'CARTULARY_SERVER_BIN="$(SERVER_BIN)"'; then
  fail 'backend-process must export CARTULARY_SERVER_BIN="$(SERVER_BIN)"'
fi
backend_process_command="$(inspect_shared_command backend-process backend-process)"
for expected in TestPhase0_ TestPhase1_ TestPhase2_; do
  if [[ "$backend_process_command" != *"$expected"* ]]; then
    fail "backend-process aggregate must include manifest-owned process coverage matching $expected"
  fi
done

target="backend-process"
prereqs="$(extract_target_prereqs "$target")"
if [[ -z "$prereqs" || "$prereqs" != *build-server* ]]; then
  fail "$target must depend on build-server"
fi

test_service_images_block="$(extract_target_block test-service-images)"
if [[ -z "$test_service_images_block" ]]; then
  fail "Makefile must define a non-empty test-service-images block"
fi
if ! text_contains "$test_service_images_block" '$(TEST_SERVICES_BIN) warm-images'; then
  fail "test-service-images must warm pinned service images through $(TEST_SERVICES_BIN)"
fi

check_build_prereqs="$(extract_target_prereqs check-build-prereqs)"
if ! text_matches "$check_build_prereqs" '(^|[[:space:]])test-service-images($|[[:space:]])'; then
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

mapfile -t check_service_backend_schedule_targets < <(schedule_targets check-service-backed backend)
check_service_backend_schedule_text="$(printf '%s\n' "${check_service_backend_schedule_targets[@]}")"
assert_text_contains_targets "check-service-backed schedule" "$check_service_backend_schedule_text" "${target_plan_service_backed_safe_targets[@]}"
assert_text_excludes_targets "check-service-backed schedule" "$check_service_backend_schedule_text" "${target_plan_check_heavy_targets[@]}" "${target_plan_service_backed_unsafe_targets[@]}"

backend_integration_support_block="$(extract_target_block backend-integration-support)"
if ! text_contains "$backend_integration_support_block" '$(CARTULARY_RUNNER_SCRIPT) go-target backend-integration-support'; then
  fail "backend-integration-support must delegate to cartulary-runner.mjs go-target backend-integration-support"
fi
if ! text_contains "$backend_integration_support_block" '$(TEST_SERVICES_BIN) run --'; then
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
if ! text_contains "$test_fast_block" '$(RUN_MAKE_SEQUENCE_SCRIPT) --sequence test-fast'; then
  fail "test-fast must use the manifest-backed sequence runner"
fi
"$node_bin" - "$repo_root/tools/task_surface_manifest.json" <<'EOF'
const fs = require("node:fs");
const [manifestPath] = process.argv.slice(2);
const sequence = JSON.parse(fs.readFileSync(manifestPath, "utf8")).sequences?.["test-fast"];
const stepTargets = (sequence?.steps ?? []).map((step) => step.target);
if (!stepTargets.includes("test-fast-service-backed")) {
  console.error("test-fast sequence must invoke test-fast-service-backed");
  process.exit(1);
}
EOF

require_service_backed_schedule_target test-fast-service-backed "test-fast service-backed" 0

mapfile -t test_fast_backend_schedule_targets < <(schedule_targets test-fast-service-backed backend)
test_fast_backend_schedule_text="$(printf '%s\n' "${test_fast_backend_schedule_targets[@]}")"
assert_text_contains_targets "test-fast-service-backed schedule" "$test_fast_backend_schedule_text" "${target_plan_service_backed_safe_targets[@]}"
assert_text_excludes_targets "test-fast-service-backed schedule" "$test_fast_backend_schedule_text" "${target_plan_check_heavy_targets[@]}" "${target_plan_service_backed_unsafe_targets[@]}"
