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
functional_script="$repo_root/scripts/run-browser-e2e-functional.sh"
browser_batch_script="$repo_root/scripts/run-browser-e2e-batch.sh"
browser_target_script="$repo_root/scripts/run-browser-e2e-target.sh"
cartulary_runner_script="$repo_root/scripts/cartulary-runner.mjs"
phase_manifest_helper="$repo_root/scripts/lib/phase-manifest.mjs"
browser_batch_manifest_helper="$repo_root/scripts/lib/browser-batch-manifest.mjs"
browser_batch_manifest="$repo_root/tools/browser_e2e_batch_manifest.json"
execution_topology_manifest="$repo_root/tools/execution_topology_manifest.json"
webserver_batch_script="$repo_root/scripts/lib/run-playwright-webserver-batch.sh"
browser_shard_plan_script="$repo_root/scripts/lib/browser-shard-plan.mjs"
browser_duration_baselines="$repo_root/tools/browser_e2e_duration_baselines.json"
webserver_batch_config="$repo_root/apps/web/playwright.webserver-backed.config.ts"
shared_playwright_config="$repo_root/apps/web/playwright.shared.config.ts"
web_package_json="$repo_root/apps/web/package.json"
stateful_script="$repo_root/scripts/run-browser-e2e-stateful.sh"
measurement_script="$repo_root/scripts/run-browser-e2e-measurement.sh"
visual_script="$repo_root/scripts/run-browser-e2e-visual.sh"
resettable_script="$repo_root/scripts/run-browser-e2e-resettable.sh"
reset_script="$repo_root/scripts/reset-web-e2e-stack.sh"
webserver_backed_script="$repo_root/scripts/run-browser-e2e-webserver-backed.sh"
start_web_e2e_script="$repo_root/scripts/start-web-e2e.sh"
schedule_manifest="$repo_root/tools/scheduler_manifest.json"
check_schedule_manifest="$repo_root/tools/scheduler_manifest.json"
task_surface_manifest="$repo_root/tools/task_surface_manifest.json"
node_bin="${NODE_BIN:-node}"

extract_target_block() {
  local target="$1"
  awk -v target="$target" '
    $0 ~ "^" target ":" { in_block=1; print; next }
    in_block && /^[^[:space:]].*:/ { exit }
    in_block { print }
  ' "$generated_make" "$makefile"
}

require_browser_owned_stack_target_uses_built_binaries() {
  local target="$1"
  local block
  block="$(extract_target_block "$target")"

  if [[ -z "$block" ]]; then
    fail "Makefile must define a non-empty $target block"
  fi
  if ! text_contains "$block" 'build-server'; then
    fail "$target must depend on build-server"
  fi
  if ! text_contains "$block" 'build-web'; then
    fail "$target must depend on build-web"
  fi
  if ! text_contains "$block" 'build-migrate'; then
    fail "$target must depend on build-migrate"
  fi
}

require_browser_batch_target() {
  local target="$1"
  local stage="$2"
  local workers="$3"
  local service_wrapper="$4"
  local dry_workers="$workers"
  local block
  local dry_run_output

  block="$(extract_target_block "$target")"
  if [[ -z "$block" ]]; then
    fail "Makefile must define a non-empty $target block"
  fi
  if ! text_contains "$block" './scripts/run-browser-e2e-target.sh'; then
    fail "$target must delegate through the browser E2E target wrapper"
  fi
  if ! text_contains "$block" "./scripts/run-browser-e2e-target.sh $stage"; then
    fail "$target must pass stage=$stage to the browser E2E target wrapper"
  fi
  if ! text_contains "$block" "PLAYWRIGHT_WORKERS=$workers"; then
    fail "$target must set PLAYWRIGHT_WORKERS=$workers"
  fi
  if ! text_contains "$block" 'BROWSER_E2E_FUNCTIONAL_SHARDS="$(BROWSER_E2E_FUNCTIONAL_SHARDS)"'; then
    fail "$target must forward BROWSER_E2E_FUNCTIONAL_SHARDS"
  fi
  if [[ "$service_wrapper" == "test-services" ]] && ! text_contains "$block" '$(TEST_SERVICES_BIN) run --'; then
    fail "$target must wrap the browser target through test services"
  fi
  if [[ "$service_wrapper" != "test-services" ]] && text_contains "$block" '$(TEST_SERVICES_BIN) run --'; then
    fail "$target must not wrap direct browser target through test services"
  fi
  if text_contains "$block" '$(TEST_OUTPUT_SCRIPT) target-summary'; then
    fail "$target Make target must not emit a duplicate browser target summary"
  fi

  dry_run_output="$(make -n --no-print-directory "$target" 2>&1)"
  if [[ "$dry_run_output" != *"BROWSER_E2E_OWNED_STACK_ENV"* && "$dry_run_output" != *"CARTULARY_SERVER_BIN="* ]]; then
    fail "$target dry-run must expand owned browser stack environment"
  fi
  if [[ "$dry_workers" == '$(PLAYWRIGHT_WORKERS)' ]]; then
    dry_workers="${PLAYWRIGHT_WORKERS:-3}"
  fi
  if [[ "$dry_run_output" != *"PLAYWRIGHT_WORKERS=$dry_workers"* ]]; then
    fail "$target dry-run must set PLAYWRIGHT_WORKERS=$dry_workers"
  fi
  if [[ "$dry_run_output" != *"BROWSER_E2E_FUNCTIONAL_SHARDS="* ]]; then
    fail "$target dry-run must forward BROWSER_E2E_FUNCTIONAL_SHARDS"
  fi
  if [[ "$dry_run_output" != *"./scripts/run-browser-e2e-target.sh $stage"* ]]; then
    fail "$target dry-run must run browser target wrapper stage $stage"
  fi
  if [[ "$service_wrapper" == "test-services" && "$dry_run_output" != *'cartulary-test-services run --'* ]]; then
    fail "$target dry-run must wrap the browser target through test services"
  fi
  if [[ "$service_wrapper" != "test-services" && "$dry_run_output" == *'cartulary-test-services run --'* ]]; then
    fail "$target dry-run must not wrap direct browser target through test services"
  fi
}

require_service_backed_schedule_target() {
  local target="$1"
  local phase_label="$2"
  local require_migrate="$3"
  local block

  block="$(extract_target_block "$target")"
  if [[ -z "$block" ]]; then
    fail "Makefile must define a non-empty $target block"
  fi
  if ! text_contains "$block" 'service-backed-target'; then
    fail "$target must delegate through the service-backed target runner"
  fi
  if ! text_contains "$block" "--target $target"; then
    fail "$target must pass target=$target to the service-backed target runner"
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
  if ! text_contains "$block" 'build-server'; then
    fail "$target must prebuild server before service-backed scheduling"
  fi
  if ! text_contains "$block" 'build-web'; then
    fail "$target must prebuild web assets before service-backed scheduling"
  fi
  if ! text_contains "$block" 'test-service-images'; then
    fail "$target must depend on test-service-images for direct runs"
  fi
  if [[ "$require_migrate" == "1" ]] && ! text_contains "$block" 'build-migrate'; then
    fail "$target must prebuild migrate before service-backed scheduling"
  fi
  if [[ "$require_migrate" == "0" ]] && text_contains "$block" 'build-migrate'; then
    fail "$target must not require build-migrate"
  fi
}

schedule_targets() {
  local schedule_target="$1"
  local kind="${2:-}"

  "$node_bin" - "$schedule_manifest" "$schedule_target" "$kind" <<'EOF'
const fs = require("node:fs");

const [manifestFile, scheduleTarget, kind] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
if (manifest.schema_id !== "cartulary.scheduler_manifest.v1") {
  throw new Error("scheduler manifest must declare schema_id=cartulary.scheduler_manifest.v1");
}
const schedules = manifest.schedules.filter((entry) => entry.target === scheduleTarget && entry.scheduler_kind === "service_backed");
if (schedules.length !== 1) {
  throw new Error(`expected exactly one schedule for ${scheduleTarget}, found ${schedules.length}`);
}
const seen = new Set();
const targets = [];
for (const entry of schedules[0].work_units ?? []) {
  if (entry.count_in_total === false || (kind !== "" && entry.class !== kind)) {
    continue;
  }
  const target = entry.aggregate_target ?? entry.target;
  if (!seen.has(target)) {
    seen.add(target);
    targets.push(target);
  }
}
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
if (manifest.schema_id !== "cartulary.scheduler_manifest.v1") {
  throw new Error("scheduler manifest must declare schema_id=cartulary.scheduler_manifest.v1");
}
const schedules = manifest.schedules.filter((entry) => entry.target === scheduleTarget && entry.scheduler_kind === "service_backed");
if (schedules.length !== 1) {
  throw new Error(`expected exactly one schedule for ${scheduleTarget}, found ${schedules.length}`);
}
const children = (schedules[0].work_units ?? []).filter((entry) => (entry.aggregate_target ?? entry.target) === childTarget && entry.count_in_total !== false);
if (children.length === 0) {
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

schedule_child_weight() {
  local schedule_target="$1"
  local child_target="$2"

  "$node_bin" - "$schedule_manifest" "$schedule_target" "$child_target" <<'EOF'
const fs = require("node:fs");

const [manifestFile, scheduleTarget, childTarget] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
const schedule = manifest.schedules.find((entry) => entry.target === scheduleTarget && entry.scheduler_kind === "service_backed");
if (!schedule) {
  throw new Error(`missing schedule ${scheduleTarget}`);
}
const child = (schedule.work_units ?? []).find((entry) => (entry.aggregate_target ?? entry.target) === childTarget && entry.count_in_total !== false);
if (!child) {
  throw new Error(`missing child ${childTarget} in ${scheduleTarget}`);
}
process.stdout.write(`${child.weight_ms ?? 0}\n`);
EOF
}

check_schedule_targets() {
  "$node_bin" - "$check_schedule_manifest" <<'EOF'
const fs = require("node:fs");

const [manifestFile] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
if (manifest.schema_id !== "cartulary.scheduler_manifest.v1") {
  throw new Error("scheduler manifest must declare schema_id=cartulary.scheduler_manifest.v1");
}
const schedules = manifest.schedules.filter((entry) => entry.target === "check" && entry.scheduler_kind === "check");
if (schedules.length !== 1) {
  throw new Error(`expected exactly one check schedule, found ${schedules.length}`);
}
const targets = schedules[0].work_units.map((entry) => entry.target);
process.stdout.write(`${targets.join("\n")}\n`);
EOF
}

check_schedule_field() {
  local work_unit="$1"
  local field="$2"

  "$node_bin" - "$check_schedule_manifest" "$work_unit" "$field" <<'EOF'
const fs = require("node:fs");

const [manifestFile, workUnit, field] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
if (manifest.schema_id !== "cartulary.scheduler_manifest.v1") {
  throw new Error("scheduler manifest must declare schema_id=cartulary.scheduler_manifest.v1");
}
const schedules = manifest.schedules.filter((entry) => entry.target === "check" && entry.scheduler_kind === "check");
if (schedules.length !== 1) {
  throw new Error(`expected exactly one check schedule, found ${schedules.length}`);
}
const unit = (schedules[0].work_units ?? []).find((entry) => entry.target === workUnit);
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

manifest_playwright_files_all() {
  local coverage="$1"
  local execution_dependency="$2"
  "$node_bin" "$phase_manifest_helper" playwright-files-all "$coverage" "$execution_dependency"
}

assert_manifest_owned_files_not_raw_selected() {
  local label="$1"
  local coverage="$2"
  local execution_dependency="$3"
  shift 3

  local manifest_files=()
  mapfile -t manifest_files < <(manifest_playwright_files_all "$coverage" "$execution_dependency")

  local manifest_file
  local web_relative
  local e2e_relative
  local candidate_file
  for manifest_file in "${manifest_files[@]}"; do
    web_relative="${manifest_file#apps/web/}"
    e2e_relative="${web_relative#e2e/}"
    for candidate_file in "$@"; do
      if grep -Fq "$manifest_file" "$candidate_file" ||
        grep -Fq "$web_relative" "$candidate_file" ||
        grep -Fq "$e2e_relative" "$candidate_file"; then
        fail "$label must not raw-select manifest-owned Playwright file $manifest_file in $candidate_file"
      fi
    done
  done
}

browser_e2e_owned_stack_env="$(sed -n 's/^BROWSER_E2E_OWNED_STACK_ENV[[:space:]]*:=//p' "$makefile" | head -n 1)"
if [[ -z "$browser_e2e_owned_stack_env" ]]; then
  fail "Makefile must define BROWSER_E2E_OWNED_STACK_ENV"
fi
if ! text_contains "$browser_e2e_owned_stack_env" 'CARTULARY_SERVER_BIN=$(SERVER_BIN)'; then
  fail "BROWSER_E2E_OWNED_STACK_ENV must export CARTULARY_SERVER_BIN=$(SERVER_BIN)"
fi
if ! text_contains "$browser_e2e_owned_stack_env" 'CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN)'; then
  fail "BROWSER_E2E_OWNED_STACK_ENV must export CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN)"
fi
if ! text_contains "$browser_e2e_owned_stack_env" 'CARTULARY_TEST_SERVICES_BIN=$(TEST_SERVICES_BIN)'; then
  fail "BROWSER_E2E_OWNED_STACK_ENV must export CARTULARY_TEST_SERVICES_BIN=$(TEST_SERVICES_BIN)"
fi
if ! text_contains "$browser_e2e_owned_stack_env" 'CARTULARY_WEB_E2E_USE_REPO_ROOT_BINARIES=1'; then
  fail "BROWSER_E2E_OWNED_STACK_ENV must opt Makefile-owned browser E2E into built repo-root binaries"
fi
for browser_owned_stack_target in \
  browser-e2e-webserver-backed \
  browser-e2e-functional \
  browser-e2e-support \
  browser-e2e-stateful \
  browser-e2e-resettable \
  browser-e2e-measurement \
  browser-e2e-visual
do
  require_browser_owned_stack_target_uses_built_binaries "$browser_owned_stack_target"
done

for removed_check_bundle in check-static-validation check-local-product check-meta-validation; do
  if rg -q "^${removed_check_bundle}:" "$generated_make" "$makefile"; then
    fail "$removed_check_bundle must not remain as a scheduled check bundle target"
  fi
done

browser_e2e_block="$(extract_target_block browser-e2e)"
if [[ -z "$browser_e2e_block" ]]; then
  fail "Makefile must define a non-empty browser-e2e block"
fi
require_browser_batch_target browser-e2e isolated 1 test-services
if text_contains "$browser_e2e_block" '-j$(BROWSER_E2E_JOBS)'; then
  fail 'browser-e2e must not fan out aggregate browser children with -j$(BROWSER_E2E_JOBS)'
fi
require_service_backed_schedule_target test-service-backed "test service-backed" 1
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

"$node_bin" --input-type=module - "$task_surface_manifest" "$check_schedule_manifest" "$browser_batch_manifest" <<'EOF'
import fs from "node:fs";
import {
  loadSummaryTopologyContext,
  resolveSummaryGroups,
  serviceBackedScheduleChildren,
} from "./scripts/lib/summary-topology.mjs";
import { loadBrowserBatchStages } from "./scripts/lib/browser-batch-manifest.mjs";

const [manifestFile, checkScheduleFile, browserBatchManifestFile] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
const checkSchedule = JSON.parse(fs.readFileSync(checkScheduleFile, "utf8"));
if (manifest.summary_profiles !== undefined) {
  throw new Error("task-surface manifest must not hard-code copied summary_profiles");
}
for (const entry of manifest.targets ?? []) {
  if (entry.summary_projection !== undefined) {
    throw new Error(`${entry.name} must not hard-code derived summary_projection children`);
  }
}
const context = loadSummaryTopologyContext({ taskSurfaceManifest: manifest });
const browserStages = loadBrowserBatchStages(browserBatchManifestFile);
const webserverBackedStage = browserStages.get("webserver-backed");
const isolatedStage = browserStages.get("isolated");
if (!webserverBackedStage || !isolatedStage) {
  throw new Error("browser batch manifest must declare webserver-backed and isolated stages for service-backed summaries");
}
const expectedBrowser = [
  webserverBackedStage.target,
  ...(isolatedStage.summaryChildren.length > 0 ? isolatedStage.summaryChildren : [isolatedStage.target]),
].sort();
const expectedCheckBrowser = expectedBrowser
  .filter((target) => !["browser-e2e-measurement", "browser-e2e-visual", "browser-e2e-a11y"].includes(target))
  .sort();
const expectedBackend = ["backend-integration", "backend-integration-support", "backend-process", "backend-store"];
const expectedCheckBackend = expectedBackend.filter((target) => target !== "backend-integration-support");
const testGroups = resolveSummaryGroups(context, manifest.sequences?.test?.summary_groups ?? []);
const checkGroups = resolveSummaryGroups(context, (checkSchedule.schedules ?? []).find((entry) => entry.target === "check")?.summary_groups ?? []);
for (const [label, groups, browserTargets, backendTargets] of [
  ["test", testGroups, expectedBrowser, expectedBackend],
  ["check", checkGroups, expectedCheckBrowser, expectedCheckBackend],
]) {
  const browser = groups.find((group) => group.name === "browser");
  const backend = groups.find((group) => group.name === "backend-service-backed");
  if (JSON.stringify([...(browser?.summaryTargets ?? [])].sort()) !== JSON.stringify(browserTargets)) {
    throw new Error(`${label} browser summary group must derive service-backed browser leaves`);
  }
  if (JSON.stringify([...(backend?.summaryTargets ?? [])].sort()) !== JSON.stringify(backendTargets)) {
    throw new Error(`${label} backend-service-backed summary group must derive backend service targets`);
  }
}
for (const [target, browserTargets] of [
  ["test-service-backed", expectedBrowser],
  ["check-service-backed", expectedCheckBrowser],
]) {
  const children = serviceBackedScheduleChildren(context, target);
  for (const child of browserTargets) {
    if (!children.includes(child)) {
      throw new Error(`${target} service-backed schedule must include ${child}`);
    }
  }
  if (
    target === "check-service-backed" &&
    ["browser-e2e-measurement", "browser-e2e-visual", "browser-e2e-a11y"].some((excluded) => children.includes(excluded))
  ) {
    throw new Error("check-service-backed service-backed schedule must exclude ordinary measurement, visual, and accessibility browser work");
  }
  if (children.includes(isolatedStage.target)) {
    throw new Error(`${target} service-backed schedule must not include aggregate ${isolatedStage.target}`);
  }
}
EOF

test_block="$(extract_target_block test)"
if [[ -z "$test_block" ]]; then
  fail "Makefile must define a non-empty test block"
fi
if ! text_contains "$test_block" '--sequence test'; then
  fail "test must delegate to the manifest-owned test sequence"
fi
if text_contains "$test_block" '--step browser-e2e'; then
  fail "test must not run browser-e2e as a final serial step"
fi
if text_contains "$test_block" 'test-isolated'; then
  fail "test must not route browser evidence through test-isolated"
fi
"$node_bin" - "$task_surface_manifest" <<'EOF'
const fs = require("node:fs");

const [manifestFile] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
const steps = manifest.sequences?.test?.steps ?? [];
if (!steps.some((step) => step.target === "test-service-backed")) {
  throw new Error("manifest test sequence must run service-backed scheduler work");
}
EOF

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
if text_contains "$check_block" '--step browser-e2e'; then
  fail "check must not run browser-e2e as a final serial step"
fi
if text_contains "$check_block" 'check-isolated'; then
  fail "check must not route browser evidence through check-isolated"
fi
if ! [[ -f "$check_schedule_manifest" ]]; then
  fail "missing tools/scheduler_manifest.json"
fi
check_schedule_text="$(check_schedule_targets)"
for scheduled_target in \
  toolchain-drift \
  codegen-toolchain \
  go-lint-toolchain \
  govulncheck-toolchain \
  gosec-toolchain \
  shell-lint-toolchain \
  check-frontend-install \
  build-web \
	  build-server \
	  build-migrate \
	  testservices-build \
  test-service-images \
  check-service-backed \
  migration-input-drift \
  migration-scratch-apply \
  backend-unit \
  frontend-typecheck \
  lint-go \
  go-vulncheck \
  go-gosec-targeted \
  frontend-unit \
  check-harness-smoke \
  lint-biome \
  frontend-import-boundary-check \
  lint-scripts \
  lint-shell \
  phase-test-name-check \
  phase-map-check \
  go-test-duration-baseline-coverage \
  phase-ledger-drift \
  phase-schedule-drift \
  service-backed-unit-check \
  generated-artifact-policy-check \
  generate-drift
do
  if ! text_has_token "$check_schedule_text" "$scheduled_target"; then
    fail "check schedule must include $scheduled_target"
  fi
done
for removed_check_bundle in check-static-validation check-local-product check-meta-validation; do
  if text_has_token "$check_schedule_text" "$removed_check_bundle"; then
    fail "check schedule must not include removed bundle $removed_check_bundle"
  fi
done
if text_has_token "$check_schedule_text" browser-e2e; then
  fail "browser-e2e must be service-backed scheduler work, not a top-level check work unit"
fi
for explicit_only_target in \
  build-operator \
  deployable-shape \
  go-gosec-audit \
  license-report \
  sbom \
  browser-e2e-measurement \
  browser-e2e-visual \
  browser-e2e-a11y
do
  if text_has_token "$check_schedule_text" "$explicit_only_target"; then
    fail "check schedule must not include explicit-only target $explicit_only_target"
  fi
done
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
for (const [target, profile] of [
  ["check-frontend-install", "setup_cpu_io"],
  ["check-service-backed", "service_session_start"],
]) {
  const metadata = topologyTargets.get(target)?.check_schedule;
  if (!metadata?.schedules?.includes("check") || metadata.profile !== profile) {
    throw new Error(`execution topology must schedule ${target} through ${profile} profile metadata`);
  }
}
const schedule = (manifest.schedules ?? []).find((entry) => entry.target === "check");
if (!schedule) {
  throw new Error("missing check schedule");
}
if (schedule.capacity_profile !== "check_default") {
  throw new Error("check schedule must resolve capacity through check_default");
}
const limits = schedule.resource_limits ?? {};
if (
  limits.host_cpu !== "auto" ||
  limits.host_io !== "auto" ||
  limits.suite_service_stack !== 1 ||
  limits.migration_scratch_postgres !== 1
) {
  throw new Error("check schedule must declare auto host_cpu, auto host_io, suite_service_stack, and migration_scratch_postgres limits");
}
const service = (schedule.work_units ?? []).find((entry) => entry.kind === "service_session" && entry.target === "check-service-backed");
if (!service) {
  throw new Error("missing check-service-backed work unit");
}
const claims = service.resource_claims ?? {};
if (JSON.stringify(claims) !== JSON.stringify({ host_cpu: 1, host_io: 1, suite_service_stack: 1 })) {
  throw new Error(`check-service-backed service session has unexpected claims ${JSON.stringify(claims)}`);
}
if (service.retained_resource_claims?.suite_service_stack !== 1) {
  throw new Error("check-service-backed service session must retain suite_service_stack");
}
if ((schedule.work_units ?? []).some((entry) => entry.nested_scheduler)) {
  throw new Error("check schedule must not use nested service-backed scheduler metadata");
}
function browserWorkerSlotCount(unit) {
  const group = unit.browser_group ?? {};
  if (group.kind === "functional_shard" || group.kind === "support") {
    return 1;
  }
  const workers = group.workers ?? "1";
  if (workers === "default") {
    return 1;
  }
  const parsed = Number.parseInt(String(workers), 10);
  if (!Number.isInteger(parsed) || parsed < 1 || String(parsed) !== String(workers)) {
    throw new Error(`${unit.label} workers must be a positive integer or default`);
  }
  return parsed;
}
function assertBrowserWorkerSlots(units, label) {
  const expectedTotal = units.reduce((sum, unit) => sum + browserWorkerSlotCount(unit), 0);
  const occupied = new Set();
  for (const unit of units) {
    const env = unit.env ?? {};
    if (env.CARTULARY_PLAYWRIGHT_WORKER_COUNT !== String(expectedTotal)) {
      throw new Error(`${unit.label} must use service-session worker count ${expectedTotal}`);
    }
    if (!/^(0|[1-9][0-9]*)$/.test(env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET ?? "")) {
      throw new Error(`${unit.label} must declare an explicit worker offset`);
    }
    const offset = Number.parseInt(env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET, 10);
    for (let slot = offset; slot < offset + browserWorkerSlotCount(unit); slot += 1) {
      if (occupied.has(slot)) {
        throw new Error(`${label} has overlapping worker-admin slot ${slot}`);
      }
      occupied.add(slot);
    }
    if (unit.browser_group?.kind === "support" && env.PLAYWRIGHT_WORKERS !== "1") {
      throw new Error(`${unit.label} support group must run with one Playwright worker`);
    }
  }
  const actualSlots = [...occupied].sort((left, right) => left - right);
  const expectedSlots = Array.from({ length: expectedTotal }, (_value, index) => index);
  if (JSON.stringify(actualSlots) !== JSON.stringify(expectedSlots)) {
    throw new Error(`${label} worker-admin slots must be contiguous`);
  }
  return expectedTotal;
}
const expectedBrowserCompletions = [
  "browser-e2e-webserver-backed",
  "browser-e2e-stateful",
];
const browserSessions = (schedule.work_units ?? []).filter((entry) =>
  entry.kind === "browser_stage_session" &&
  entry.service_session?.target === "check-service-backed"
);
const sessionByGroup = new Map(browserSessions.map((entry) => [entry.browser_session_group, entry]));
if (browserSessions.length !== 2 || !sessionByGroup.has("default-check-browser-shared") || !sessionByGroup.has("default-check-stateful-isolated")) {
  throw new Error(`check schedule must expose the shared default browser session and isolated stateful session, got ${browserSessions.map((entry) => entry.browser_session_group).join(",")}`);
}
const sharedSession = sessionByGroup.get("default-check-browser-shared");
const statefulSession = sessionByGroup.get("default-check-stateful-isolated");
for (const session of [sharedSession, statefulSession]) {
  if (
    !session.needs?.includes("service_session:check-service-backed") ||
    !session.needs?.includes("build-web") ||
    !session.needs?.includes("build-server") ||
    !session.needs?.includes("build-migrate")
  ) {
    throw new Error(`${session.label} browser stage session must depend on the service session, build-web, build-server, and build-migrate`);
  }
  if (
    Object.prototype.hasOwnProperty.call(session.retained_resource_claims ?? {}, "postgres") ||
    Object.prototype.hasOwnProperty.call(session.retained_resource_claims ?? {}, "object_store")
  ) {
    throw new Error(`${session.label} browser stage session must not retain broad Postgres or object-store claims`);
  }
}
for (const resource of ["browser_stage_webserver_backed"]) {
  if (sharedSession.retained_resource_claims?.[resource] !== 1) {
    throw new Error(`shared default browser session must retain ${resource}`);
  }
}
if (statefulSession.retained_resource_claims?.browser_stage_stateful !== 1 || !statefulSession.browser_session_isolation_reason) {
  throw new Error("stateful browser session must be isolated and declare an isolation reason");
}
for (const expectedLeaf of expectedBrowserCompletions) {
  const complete = (schedule.work_units ?? []).find((entry) =>
    entry.target === expectedLeaf &&
    entry.kind === "browser_stage_complete" &&
    entry.service_session?.target === "check-service-backed"
  );
  if (!complete) {
    throw new Error(`check schedule must expose ${expectedLeaf} as browser stage completion work`);
  }
}
const webserverGroups = (schedule.work_units ?? []).filter((entry) =>
  entry.target === "browser-e2e-webserver-backed" && entry.kind === "browser_group"
);
if (webserverGroups.length < 3 || !webserverGroups.some((entry) => entry.browser_group?.kind === "functional_shard")) {
  throw new Error("check schedule must split browser-e2e-webserver-backed into functional shard and support browser groups");
}
const webserverGroupKeys = new Set(webserverGroups.map((entry) => `browser_group:${entry.browser_group?.id}`));
const sharedBrowserGroupKeys = new Set(
  (schedule.work_units ?? [])
    .filter((entry) => entry.kind === "browser_group" && entry.browser_session_group === "default-check-browser-shared")
    .map((entry) => `browser_group:${entry.browser_group?.id}`),
);
const webserverStageSessionKey = "browser_stage_session:default-check-browser-shared";
const functionalShardCount = Math.max(
  ...webserverGroups
    .filter((entry) => entry.browser_group?.kind === "functional_shard")
    .map((entry) => entry.browser_group?.shard_count ?? 0),
);
if (functionalShardCount !== 6) {
  throw new Error(`check schedule must render 6 duration-balanced functional shards, got ${functionalShardCount}`);
}
const serviceBrowserWorkerCount = assertBrowserWorkerSlots(
  (schedule.work_units ?? []).filter((entry) =>
    entry.kind === "browser_group" &&
    entry.service_session?.target === "check-service-backed"
  ),
  "check-service-backed browser groups",
);
for (const group of webserverGroups) {
  if (JSON.stringify(group.needs ?? []) !== JSON.stringify([webserverStageSessionKey])) {
    throw new Error(`${group.label} must depend only on the webserver-backed browser stage session`);
  }
  if ((group.needs ?? []).some((need) => webserverGroupKeys.has(need))) {
    throw new Error(`${group.label} must not depend on another browser group`);
  }
  if (group.browser_group?.kind === "functional_shard") {
    if (
      group.env?.CARTULARY_PLAYWRIGHT_WORKER_COUNT !== String(serviceBrowserWorkerCount) ||
      group.env?.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET !== String(group.browser_group.shard_index)
    ) {
      throw new Error(`${group.label} must use its functional shard worker slot`);
    }
  }
  if (group.browser_group?.kind === "support") {
    if (
      group.env?.CARTULARY_PLAYWRIGHT_WORKER_COUNT !== String(serviceBrowserWorkerCount) ||
      group.env?.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET !== String(functionalShardCount) ||
      group.env?.PLAYWRIGHT_WORKERS !== "1"
    ) {
      throw new Error(`${group.label} must use a worker slot outside the functional shard range`);
    }
  }
}
const measurementSession = (schedule.work_units ?? []).find((entry) =>
  entry.target === "browser-e2e-measurement" && entry.kind === "browser_stage_session"
);
if (measurementSession) {
  throw new Error("default local check must not include browser-e2e-measurement stage work");
}
for (const excludedTarget of ["browser-e2e-visual", "browser-e2e-a11y"]) {
  if ((schedule.work_units ?? []).some((entry) => entry.target === excludedTarget && entry.service_session?.target === "check-service-backed")) {
    throw new Error(`default local check must not include ${excludedTarget} stage work`);
  }
}
const webserverComplete = (schedule.work_units ?? []).find((entry) =>
  entry.target === "browser-e2e-webserver-backed" && entry.kind === "browser_stage_complete"
);
const expectedWebserverCompleteNeeds = [...webserverGroupKeys].sort();
const actualWebserverCompleteNeeds = [...(webserverComplete?.needs ?? [])].sort();
if (JSON.stringify(actualWebserverCompleteNeeds) !== JSON.stringify(expectedWebserverCompleteNeeds)) {
  throw new Error("browser-e2e-webserver-backed completion must depend only on webserver-backed browser groups");
}
const sharedSessionFinalizer = (schedule.work_units ?? []).find((entry) =>
  entry.kind === "browser_session_finalizer" &&
  entry.browser_session_group === "default-check-browser-shared"
);
if (
  sharedSessionFinalizer &&
  JSON.stringify([...(sharedSessionFinalizer.needs ?? [])].sort()) !== JSON.stringify([...sharedBrowserGroupKeys].sort())
) {
  throw new Error("shared default browser session finalizer must wait for every shared browser group");
}
if ((schedule.work_units ?? []).some((entry) =>
  entry.target === "browser-e2e-webserver-backed" && entry.kind === "service_make_target"
)) {
  throw new Error("check schedule must not keep browser-e2e-webserver-backed as one service_make_target leaf");
}
EOF

if ! rg -q '^browser-e2e-webserver-backed:' "$generated_make" "$makefile"; then
  fail "Makefile must define browser-e2e-webserver-backed"
fi
browser_webserver_backed_block="$(extract_target_block browser-e2e-webserver-backed)"
if [[ -z "$browser_webserver_backed_block" ]]; then
  fail "Makefile must define a non-empty browser-e2e-webserver-backed block"
fi
require_browser_batch_target browser-e2e-webserver-backed webserver-backed '$(PLAYWRIGHT_WORKERS)' direct
require_browser_batch_target browser-e2e-functional functional '$(PLAYWRIGHT_WORKERS)' direct
require_browser_batch_target browser-e2e-support support '$(PLAYWRIGHT_WORKERS)' direct
require_browser_batch_target browser-e2e-stateful stateful 1 direct
require_browser_batch_target browser-e2e-resettable resettable 1 direct
require_browser_batch_target browser-e2e-measurement measurement 1 test-services
require_browser_batch_target browser-e2e-visual visual 1 direct
browser_measurement_block="$(extract_target_block browser-e2e-measurement)"
if text_contains "$browser_measurement_block" 'Core 05-bound timing evidence'; then
  fail "browser-e2e-measurement must be labeled ordinary measurement, not Core 05-bound claim evidence"
fi

if ! [[ -f "$functional_script" ]]; then
  fail "missing scripts/run-browser-e2e-functional.sh"
fi
if ! [[ -x "$browser_batch_script" ]]; then
  fail "missing executable scripts/run-browser-e2e-batch.sh"
fi
if ! [[ -f "$browser_batch_manifest" ]]; then
  fail "missing tools/browser_e2e_batch_manifest.json"
fi
if ! [[ -f "$browser_batch_manifest_helper" ]]; then
  fail "missing scripts/lib/browser-batch-manifest.mjs"
fi
"$node_bin" "$browser_batch_manifest_helper" validate "$browser_batch_manifest"
while IFS=$'\t' read -r stage_name group_name group_target _group_kind group_coverage group_execution_dependency _stage_schedule_tags _stage_dependency_policy; do
  if [[ "$group_coverage" == "raw" || -z "$group_coverage" ]]; then
    continue
  fi
  if [[ -z "$group_execution_dependency" ]]; then
    fail "browser batch group $stage_name/$group_name must declare execution_dependency for $group_coverage coverage"
  fi
  if [[ "$group_execution_dependency" == "browser_a11y" || "$group_execution_dependency" == "browser_a11y_preflight" ]]; then
    continue
  fi
  group_count="$("$node_bin" "$phase_manifest_helper" playwright-count-all "$group_coverage" "$group_execution_dependency")"
  if [[ "$group_count" == "0" ]]; then
    fail "browser batch group $stage_name/$group_name target=$group_target must match manifest Playwright rows for $group_coverage $group_execution_dependency"
  fi
done < <("$node_bin" "$browser_batch_manifest_helper" group-selections "$browser_batch_manifest")
if ! [[ -f "$webserver_batch_script" ]]; then
  fail "missing scripts/lib/run-playwright-webserver-batch.sh"
fi
if ! [[ -f "$browser_shard_plan_script" ]]; then
  fail "missing scripts/lib/browser-shard-plan.mjs"
fi
if ! [[ -f "$browser_duration_baselines" ]]; then
  fail "missing tools/browser_e2e_duration_baselines.json"
fi
if ! [[ -f "$webserver_batch_config" ]]; then
  fail "missing apps/web/playwright.webserver-backed.config.ts"
fi
if ! [[ -f "$shared_playwright_config" ]]; then
  fail "missing apps/web/playwright.shared.config.ts"
fi
if ! [[ -f "$stateful_script" ]]; then
  fail "missing scripts/run-browser-e2e-stateful.sh"
fi
if ! [[ -f "$measurement_script" ]]; then
  fail "missing scripts/run-browser-e2e-measurement.sh"
fi
if ! [[ -f "$visual_script" ]]; then
  fail "missing scripts/run-browser-e2e-visual.sh"
fi
if ! [[ -f "$resettable_script" ]]; then
  fail "missing scripts/run-browser-e2e-resettable.sh"
fi
if ! [[ -f "$reset_script" ]]; then
  fail "missing scripts/reset-web-e2e-stack.sh"
fi
if ! [[ -f "$start_web_e2e_script" ]]; then
  fail "missing scripts/start-web-e2e.sh"
fi
if ! grep -Fq 'DEV_SERVICES_SCRIPT=' "$start_web_e2e_script"; then
  fail "scripts/start-web-e2e.sh must configure the shared dev service helper"
fi
if ! grep -Fq 'CARTULARY_TEST_SERVICES_ACTIVE' "$start_web_e2e_script"; then
  fail "scripts/start-web-e2e.sh must detect active test-service suites"
fi
if ! grep -Fq 'prepare-web-e2e --env-file' "$start_web_e2e_script"; then
  fail "scripts/start-web-e2e.sh must prepare browser fixtures through cartulary-test-services when active"
fi
if ! grep -Fq 'cleanup-web-e2e --metadata-file' "$start_web_e2e_script"; then
  fail "scripts/start-web-e2e.sh must clean browser fixtures through cartulary-test-services when active"
fi
if ! grep -Fq '"${DEV_SERVICES_SCRIPT}" wait' "$start_web_e2e_script"; then
  fail "scripts/start-web-e2e.sh must wait for Postgres and object-store through dev-services.sh"
fi
if ! grep -Fq 'docker compose -f "${COMPOSE_FILE}" up -d postgres seaweedfs-s3' "$start_web_e2e_script"; then
  fail "scripts/start-web-e2e.sh must keep Compose-backed startup for standalone browser E2E"
fi

if ! grep -Fq 'cartulary.browser_e2e_batch_manifest.v5' "$browser_batch_manifest"; then
  fail "browser E2E batch manifest must declare its schema"
fi
"$node_bin" - "$browser_batch_manifest" <<'EOF' || fail "browser E2E batch manifest must tag service-backed scheduler stages"
const fs = require("node:fs");
const [manifestFile] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
if (!(manifest.stages ?? []).some((stage) => (stage.schedule_tags ?? []).includes("service_backed_full"))) {
  throw new Error("missing service_backed_full schedule tag");
}
EOF
"$node_bin" - "$browser_batch_manifest" <<'EOF' || fail "browser E2E batch manifest must schedule isolated browser leaves independently"
const fs = require("node:fs");
const [manifestFile] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
const byName = new Map((manifest.stages ?? []).map((stage) => [stage.name, stage]));
if ((manifest.stages ?? []).some((stage) => stage.scheduler_dependency_policy !== undefined)) {
  throw new Error("browser stages must not use obsolete scheduler_dependency_policy");
}
for (const name of ["webserver-backed", "stateful"]) {
  const stage = byName.get(name);
  if (!stage) {
    throw new Error(`missing ${name} stage`);
  }
  if (!(stage.schedule_tags ?? []).includes("service_backed_full")) {
    throw new Error(`${name} must be tagged service_backed_full`);
  }
  if (!(stage.schedule_tags ?? []).includes("service_backed_check")) {
    throw new Error(`${name} must be tagged service_backed_check`);
  }
  if ((stage.scheduler_needs ?? []).length !== 0) {
    throw new Error(`${name} must not declare broad service-backed scheduler needs`);
  }
}
const visual = byName.get("visual");
if (!visual) {
  throw new Error("missing visual stage");
}
if (!(visual.schedule_tags ?? []).includes("service_backed_full")) {
  throw new Error("visual must be tagged service_backed_full");
}
if ((visual.schedule_tags ?? []).includes("service_backed_check")) {
  throw new Error("visual full stage must not be tagged service_backed_check");
}
const measurement = byName.get("measurement");
if (!measurement) {
  throw new Error("missing measurement stage");
}
if (!(measurement.schedule_tags ?? []).includes("service_backed_full")) {
  throw new Error("measurement must be tagged service_backed_full");
}
if ((measurement.schedule_tags ?? []).includes("service_backed_check")) {
  throw new Error("measurement must not be tagged service_backed_check");
}
if ((measurement.scheduler_needs ?? []).length !== 0) {
  throw new Error("measurement must not declare broad service-backed scheduler needs");
}
const isolated = byName.get("isolated");
if (!isolated || (isolated.schedule_tags ?? []).length > 0 || (isolated.scheduler_needs ?? []).length > 0) {
  throw new Error("isolated aggregate must remain unscheduled direct-run aggregation");
}
EOF
"$node_bin" - "$browser_batch_manifest" "$browser_batch_script" <<'EOF' || fail "browser E2E batch runner must route every generated group kind"
const fs = require("node:fs");
const [manifestFile, batchScript] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
const script = fs.readFileSync(batchScript, "utf8");
const routedKinds = new Set([...script.matchAll(/^\s{4}([A-Za-z0-9_-]+)\)/gm)].map((match) => match[1]));
for (const stage of manifest.stages ?? []) {
  for (const group of stage.groups ?? []) {
    if (!routedKinds.has(group.kind)) {
      throw new Error(`browser group kind ${group.kind} from stage ${stage.name} has no batch runner route`);
    }
  }
}
EOF
if ! grep -Fq '"kind": "duration_balanced_specs"' "$browser_batch_manifest"; then
  fail "browser E2E batch manifest must route functional browser work through duration-balanced specs"
fi
if ! grep -Fq '"execution_dependency": "browser_functional"' "$browser_batch_manifest"; then
  fail "browser E2E batch manifest must declare manifest execution dependencies for browser groups"
fi
if ! grep -Fq '"execution_dependency": "browser_support"' "$browser_batch_manifest"; then
  fail "browser E2E batch manifest must declare browser_support manifest selection for support groups"
fi
if ! grep -Fq '"execution_dependency": "browser_visual"' "$browser_batch_manifest"; then
  fail "browser E2E batch manifest must declare browser_visual manifest selection for visual groups"
fi
if ! grep -Fq 'cartulary.browser_e2e_duration_baselines.v2' "$browser_duration_baselines"; then
  fail "browser E2E duration baselines must declare their schema"
fi
if ! grep -Fq 'browser_functional' "$browser_shard_plan_script"; then
  fail "browser shard planner must select authoritative browser_functional rows"
fi
if ! grep -Fq 'shard_target_ms' "$browser_duration_baselines"; then
  fail "browser E2E duration baselines must declare shard_target_ms"
fi
if ! grep -Fq 'default_entry_weight_ms' "$browser_duration_baselines"; then
  fail "browser E2E duration baselines must declare default_entry_weight_ms"
fi
if ! grep -Fq '"name": "isolated"' "$browser_batch_manifest"; then
  fail "browser E2E batch manifest must define the isolated stage"
fi
if ! grep -Fq '"reset_before": "stateful-to-measurement"' "$browser_batch_manifest"; then
  fail "browser E2E batch manifest must reset between stateful and measurement groups"
fi
if ! grep -Fq '"reset_before": "measurement-to-visual"' "$browser_batch_manifest"; then
  fail "browser E2E batch manifest must reset between measurement and visual groups"
fi
if ! grep -Fq 'reset-web-e2e-stack.sh' "$browser_batch_script"; then
  fail "browser E2E batch runner must enforce reset boundaries"
fi
if ! grep -Fq 'run-browser-e2e-webserver-backed.sh' "$browser_batch_script"; then
  fail "browser E2E batch runner must route the webserver-backed group"
fi
if ! grep -Fq 'run-browser-e2e-stateful.sh' "$browser_batch_script"; then
  fail "browser E2E batch runner must route the stateful group"
fi
if ! grep -Fq 'run-browser-e2e-measurement.sh' "$browser_batch_script"; then
  fail "browser E2E batch runner must route the measurement group"
fi
if ! grep -Fq 'run-browser-e2e-visual.sh' "$browser_batch_script"; then
  fail "browser E2E batch runner must route the visual group"
fi
if ! grep -Fq -- '--defer-summary' "$browser_batch_script"; then
  fail "browser E2E batch runner must support deferred stage summaries"
fi
if ! grep -Fq 'target-summary "$target"' "$browser_batch_script"; then
  fail "browser E2E batch runner must emit child target summaries"
fi
if ! [[ -x "$browser_target_script" ]]; then
  fail "missing executable scripts/run-browser-e2e-target.sh"
fi
if ! grep -Fq 'run-browser-e2e-batch.sh" "$stage" --defer-summary' "$browser_target_script"; then
  fail "browser E2E target wrapper must defer batch stage summary ownership"
fi
if ! grep -Fq 'start-web-e2e.sh' "$browser_target_script"; then
  fail "browser E2E target wrapper must give every scheduled browser leaf an owned stack"
fi
if ! grep -Fq 'target-summary "$target" "$requested"' "$browser_target_script"; then
  fail "browser E2E target wrapper must emit the authoritative stage target summary"
fi

if ! grep -Fq 'run-playwright-webserver-batch.sh' "$functional_script"; then
  fail "scripts/run-browser-e2e-functional.sh must delegate to the Playwright webserver batch runner"
fi
if ! grep -Fq 'functional' "$functional_script"; then
  fail "scripts/run-browser-e2e-functional.sh must run the functional-only batch mode"
fi
if ! grep -Fq 'playwright.webserver-backed.config.ts' "$functional_script"; then
  fail "scripts/run-browser-e2e-functional.sh must use the batched webserver-backed Playwright config"
fi
if grep -Fq 'run-playwright-manifest-phase.sh' "$functional_script"; then
  fail "scripts/run-browser-e2e-functional.sh must not launch one Playwright process per manifest phase"
fi
if ! grep -Fq 'browser-shard-plan.mjs' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must use the browser shard planner"
fi
if ! grep -Fq 'PLAYWRIGHT_WORKERS=1' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must run each functional shard with one Playwright worker"
fi
if ! grep -Fq 'playwright_worker_count="${CARTULARY_PLAYWRIGHT_WORKER_COUNT:-$functional_shard_limit}"' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must allow scheduled browser groups to provision a stage-wide worker-admin range"
fi
if ! grep -Fq 'CARTULARY_PLAYWRIGHT_WORKER_COUNT="$playwright_worker_count"' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must pass the resolved worker-admin count to Playwright"
fi
if ! grep -Fq 'resolve_functional_worker_offset' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must resolve scheduled functional shard worker-admin offsets"
fi
if ! grep -Fq 'CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET="$worker_offset"' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must pass the resolved worker-admin offset to Playwright"
fi
if ! grep -Fq 'merge-reports' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must merge functional shard reports before phase summaries"
fi
if ! grep -Fq 'functional_shard_limit' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must cap browser entry shard parallelism by BROWSER_E2E_FUNCTIONAL_SHARDS"
fi
if ! grep -Fq 'browser_functional' "$browser_shard_plan_script"; then
  fail "scripts/lib/browser-shard-plan.mjs must select authoritative browser_functional manifest rows"
fi
if awk 'index($0, "playwright-grep-many") && index($0, "browser_functional") { found = 1 } END { exit found ? 0 : 1 }' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must not keep the old all-phase functional Playwright grep batch"
fi
browser_functional_phases=()
while IFS= read -r phase; do
  if [[ -z "$phase" ]]; then
    continue
  fi
  count="$("$node_bin" "$repo_root/scripts/lib/phase-manifest.mjs" playwright-count "$phase" authoritative browser_functional)"
  if [[ "$count" == "0" ]]; then
    continue
  fi
  browser_functional_phases+=("$phase")
done < <("$node_bin" "$repo_root/scripts/lib/phase-manifest.mjs" list-phases)

mapfile -t expected_browser_functional_phases < <("$node_bin" "$repo_root/scripts/lib/phase-manifest.mjs" playwright-phases authoritative browser_functional)
if [[ "$(printf '%s\n' "${browser_functional_phases[@]}")" != "$(printf '%s\n' "${expected_browser_functional_phases[@]}")" ]]; then
  fail "authoritative browser_functional phases must be manifest-derived, found: ${browser_functional_phases[*]:-none}; expected: ${expected_browser_functional_phases[*]:-none}"
fi
for browser_functional_phase in "${browser_functional_phases[@]}"; do
  if [[ "$browser_functional_phase" == "phase0" ]]; then
    fail "authoritative browser_functional manifest phases must skip phase0, found: ${browser_functional_phases[*]}"
  fi
done
if ! grep -Fq 'CARTULARY_REPORT_SLICE=1' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must emit sliced Playwright summaries"
fi
if ! grep -Fq 'accounting_mode=actual' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must account browser_functional phase slices as actual timings"
fi
if grep -Fq 'if [[ "$index" == "0" ]]' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must not charge the full batch only to the first functional phase"
fi
if grep -Eq 'playwright test e2e/phase[0-9]' "$web_package_json"; then
  fail "apps/web/package.json must not hardcode browser phase spec lists"
fi
if "$node_bin" - "$web_package_json" <<'EOF'
const fs = require("node:fs");
const [packageFile] = process.argv.slice(2);
const packageJSON = JSON.parse(fs.readFileSync(packageFile, "utf8"));
const phaseScripts = Object.keys(packageJSON.scripts ?? {}).filter((name) =>
  /^test:e2e:phase\d+$/.test(name),
);
if (phaseScripts.length > 0) {
  console.error(`apps/web/package.json must not expose phase-numbered browser E2E aliases: ${phaseScripts.join(",")}`);
  process.exit(1);
}
EOF
then
  :
else
  fail "apps/web/package.json must not expose phase-numbered browser E2E aliases"
fi
if ! grep -Fq 'name: "functional"' "$webserver_batch_config"; then
  fail "apps/web/playwright.webserver-backed.config.ts must define the functional project"
fi
if ! grep -Fq 'name: "support"' "$webserver_batch_config"; then
  fail "apps/web/playwright.webserver-backed.config.ts must define the support project"
fi
if ! grep -Fq 'CARTULARY_PLAYWRIGHT_FUNCTIONAL_GREP' "$webserver_batch_config"; then
  fail "apps/web/playwright.webserver-backed.config.ts must scope manifest grep to the functional project"
fi
if ! grep -Fq 'CARTULARY_PLAYWRIGHT_FUNCTIONAL_FILES' "$webserver_batch_config"; then
  fail "apps/web/playwright.webserver-backed.config.ts must scope functional files from the Playwright manifest"
fi
if ! grep -Fq 'CARTULARY_PLAYWRIGHT_SUPPORT_GREP' "$webserver_batch_config"; then
  fail "apps/web/playwright.webserver-backed.config.ts must scope support grep from the Playwright manifest"
fi
if ! grep -Fq 'CARTULARY_PLAYWRIGHT_SUPPORT_FILES' "$webserver_batch_config"; then
  fail "apps/web/playwright.webserver-backed.config.ts must scope support files from the Playwright manifest"
fi
assert_manifest_owned_files_not_raw_selected \
  "browser functional execution" \
  authoritative \
  browser_functional \
  "$webserver_batch_config" \
  "$web_package_json" \
  "$functional_script" \
  "$webserver_backed_script" \
  "$browser_batch_script" \
  "$webserver_batch_script"
assert_manifest_owned_files_not_raw_selected \
  "browser support execution" \
  supplemental \
  browser_support \
  "$webserver_batch_config" \
  "$web_package_json" \
  "$browser_batch_script" \
  "$webserver_batch_script"
if ! grep -Fq 'webE2EBaseConfig' "$webserver_batch_config"; then
  fail "apps/web/playwright.webserver-backed.config.ts must use the shared Playwright web E2E config"
fi
if ! grep -Fq 'webE2EBaseConfig' "$shared_playwright_config"; then
  fail "apps/web/playwright.shared.config.ts must expose the shared Playwright web E2E config"
fi

if ! grep -Fq 'browser_stateful' "$stateful_script"; then
  fail "scripts/run-browser-e2e-stateful.sh must execute browser_stateful rows through the manifest"
fi
if ! grep -Fq 'run-browser-e2e-manifest-dependency.sh' "$stateful_script"; then
  fail "scripts/run-browser-e2e-stateful.sh must use manifest-derived phase discovery"
fi
assert_manifest_owned_files_not_raw_selected \
  "browser stateful execution" \
  authoritative \
  browser_stateful \
  "$stateful_script" \
  "$web_package_json"
if ! grep -Fq 'claim_bearing": false' "$measurement_script"; then
  fail "scripts/run-browser-e2e-measurement.sh must emit claim_bearing=false ordinary measurement metadata"
fi
if ! grep -Fq 'evidence_kind": "ordinary_measurement"' "$measurement_script"; then
  fail "scripts/run-browser-e2e-measurement.sh must emit ordinary_measurement evidence_kind metadata"
fi
if ! grep -Fq '"sample_count_per_predicate": 25' "$measurement_script"; then
  fail "scripts/run-browser-e2e-measurement.sh must emit the ordinary 25-sample p95 measurement policy"
fi
if ! grep -Fq 'browser_measurement' "$measurement_script"; then
  fail "scripts/run-browser-e2e-measurement.sh must execute browser_measurement rows through the manifest"
fi
if ! grep -Fq 'run-browser-e2e-manifest-dependency.sh' "$measurement_script"; then
  fail "scripts/run-browser-e2e-measurement.sh must use manifest-derived phase discovery"
fi
assert_manifest_owned_files_not_raw_selected \
  "browser measurement execution" \
  authoritative \
  browser_measurement \
  "$measurement_script" \
  "$web_package_json"
if ! grep -Fq 'browser_visual' "$visual_script"; then
  fail "scripts/run-browser-e2e-visual.sh must execute browser_visual rows through the manifest"
fi
if ! grep -Fq 'run-browser-e2e-manifest-dependency.sh' "$visual_script"; then
  fail "scripts/run-browser-e2e-visual.sh must use manifest-derived phase discovery"
fi
if ! grep -Fq 'CARTULARY_FRONTEND_ROW_ACCOUNTING_SCOPE' "$visual_script"; then
  fail "scripts/run-browser-e2e-visual.sh must honor frontend row-accounting scope"
fi
if ! grep -Fq -- '--row-ids' "$visual_script"; then
  fail "scripts/run-browser-e2e-visual.sh must constrain frontend visual readiness to selected rows"
fi
if ! grep -Fq 'CARTULARY_WEB_E2E_PORT_LEASE_ROOT' "$start_web_e2e_script"; then
  fail "scripts/start-web-e2e.sh must expose a repo-local browser E2E port lease root"
fi
if ! grep -Fq 'reserve_port_lease' "$start_web_e2e_script"; then
  fail "scripts/start-web-e2e.sh must reserve ports during browser E2E allocation"
fi
if ! grep -Fq 'release_port_leases' "$start_web_e2e_script"; then
  fail "scripts/start-web-e2e.sh must release browser E2E port reservations during cleanup"
fi
selected_visual_grep="$("$node_bin" "$repo_root/scripts/lib/frontend-phase-manifest.mjs" playwright-grep browser-e2e-visual visual --row-ids FE-V-P5-01)"
if [[ "$selected_visual_grep" != *"FE-V-P5-01"* || "$selected_visual_grep" == *"FE-V-P3-01"* ]]; then
  fail "frontend visual readiness grep must filter by selected row IDs"
fi
assert_manifest_owned_files_not_raw_selected \
  "browser visual execution" \
  authoritative \
  browser_visual \
  "$web_package_json"
if ! grep -Fq 'run-browser-e2e-batch.sh" resettable' "$resettable_script"; then
  fail "scripts/run-browser-e2e-resettable.sh must delegate resettable sequencing to the browser batch runner"
fi
if ! grep -Fq '/api/v1/test/runtime/reset' "$reset_script"; then
  fail "scripts/reset-web-e2e-stack.sh must call the test runtime reset route"
fi
if ! grep -Fq 'X-Cartulary-Test-Route-Token' "$reset_script"; then
  fail "scripts/reset-web-e2e-stack.sh must authorize reset calls with the test route token header"
fi
if ! grep -Fq 'CARTULARY_PLAYWRIGHT_STATE_DIR' "$reset_script"; then
  fail "scripts/reset-web-e2e-stack.sh must clear shared Playwright state after backend reset"
fi

if ! grep -Fq 'run-playwright-webserver-batch.sh' "$webserver_backed_script"; then
  fail "scripts/run-browser-e2e-webserver-backed.sh must delegate to the Playwright webserver batch runner"
fi
if ! grep -Fq 'webserver-backed' "$webserver_backed_script"; then
  fail "scripts/run-browser-e2e-webserver-backed.sh must run the functional-plus-support batch mode"
fi
if ! grep -Fq 'playwright.webserver-backed.config.ts' "$webserver_backed_script"; then
  fail "scripts/run-browser-e2e-webserver-backed.sh must use the batched webserver-backed Playwright config"
fi
if grep -Fq 'run-browser-e2e-functional.sh' "$webserver_backed_script"; then
  fail "scripts/run-browser-e2e-webserver-backed.sh must not serialize through scripts/run-browser-e2e-functional.sh"
fi

if ! "$node_bin" - "$repo_root" <<'EOF'
const fs = require("fs");
const path = require("path");

const root = process.argv[2];
const phaseSchemaID = "cartulary.phase_test_map.v2";
for (const phaseEntry of manifestPhases()) {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(root, phaseEntry.manifest_path), "utf8"),
  );
  for (const entry of manifest.e2e ?? []) {
    if (entry.coverage !== "authoritative" || entry.runner !== "playwright") {
      continue;
    }
    if (!["browser_functional", "browser_stateful", "browser_measurement"].includes(entry.execution_dependency)) {
      console.error(
        `${phaseEntry.phase} authoritative e2e row ${entry.id} must declare a canonical browser execution_dependency`,
      );
      process.exit(1);
    }
  }
}
function manifestPhases() {
  const registrySchemaID = "cartulary.phase_registry.v1";
  const registry = JSON.parse(fs.readFileSync(path.join(root, "tools", "phase_registry.json"), "utf8"));
  if (registry.schema_id !== registrySchemaID) {
    throw new Error(`tools/phase_registry.json must declare schema_id ${registrySchemaID}`);
  }
  return (registry.phases ?? [])
    .filter((entry) => entry.status === "active")
    .sort((left, right) => left.order - right.order || left.phase.localeCompare(right.phase))
    .map((entry) => {
    const manifest = JSON.parse(fs.readFileSync(path.join(root, entry.manifest_path), "utf8"));
    if (manifest.schema_id !== phaseSchemaID) {
      throw new Error(`${entry.manifest_path} must declare schema_id ${phaseSchemaID}`);
    }
    if (manifest.phase !== entry.phase) {
      throw new Error(`${entry.manifest_path} must declare phase ${entry.phase}`);
    }
    return entry;
  });
}
EOF
then
  fail "Authoritative browser manifest rows must carry canonical execution_dependency values"
fi
