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
schedule_topology_helper="$repo_root/scripts/lib/service-backed-schedule-topology.mjs"
webserver_batch_script="$repo_root/scripts/lib/run-playwright-webserver-batch.sh"
browser_shard_plan_script="$repo_root/scripts/lib/browser-shard-plan.mjs"
browser_duration_baselines="$repo_root/tools/browser_e2e_duration_baselines.json"
webserver_batch_config="$repo_root/apps/web/playwright.webserver-backed.config.ts"
shared_playwright_config="$repo_root/apps/web/playwright.shared.config.ts"
web_package_json="$repo_root/apps/web/package.json"
stateful_script="$repo_root/scripts/run-browser-e2e-stateful.sh"
measurement_script="$repo_root/scripts/run-browser-e2e-measurement.sh"
resettable_script="$repo_root/scripts/run-browser-e2e-resettable.sh"
reset_script="$repo_root/scripts/reset-web-e2e-stack.sh"
webserver_backed_script="$repo_root/scripts/run-browser-e2e-webserver-backed.sh"
start_web_e2e_script="$repo_root/scripts/start-web-e2e.sh"
schedule_manifest="$repo_root/tools/service_backed_schedule_manifest.json"
check_schedule_manifest="$repo_root/tools/check_schedule_manifest.json"
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
    dry_workers="${PLAYWRIGHT_WORKERS:-2}"
  fi
  if [[ "$dry_run_output" != *"PLAYWRIGHT_WORKERS=$dry_workers"* ]]; then
    fail "$target dry-run must set PLAYWRIGHT_WORKERS=$dry_workers"
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

schedule_child_weight() {
  local schedule_target="$1"
  local child_target="$2"

  "$node_bin" - "$schedule_manifest" "$schedule_target" "$child_target" <<'EOF'
const fs = require("node:fs");

const [manifestFile, scheduleTarget, childTarget] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
const schedule = manifest.schedules.find((entry) => entry.target === scheduleTarget);
if (!schedule) {
  throw new Error(`missing schedule ${scheduleTarget}`);
}
const child = schedule.work_unit_sources.find((entry) => entry.target === childTarget);
if (!child) {
  throw new Error(`missing child ${childTarget} in ${scheduleTarget}`);
}
process.stdout.write(`${child.weight ?? 0}\n`);
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
if (manifest.schema_id !== "cartulary.check_schedule.v6") {
  throw new Error("check schedule manifest must declare schema_id=cartulary.check_schedule.v6");
}
const schedules = manifest.schedules.filter((entry) => entry.target === "check");
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

"$node_bin" "$schedule_topology_helper" validate "$schedule_manifest" "$execution_topology_manifest"

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
];
const expectedBackend = ["backend-store", "backend-integration", "backend-integration-support", "backend-process"];
const testGroups = resolveSummaryGroups(context, manifest.sequences?.test?.summary_groups ?? []);
const checkGroups = resolveSummaryGroups(context, (checkSchedule.schedules ?? []).find((entry) => entry.target === "check")?.summary_groups ?? []);
for (const [label, groups] of [["test", testGroups], ["check", checkGroups]]) {
  const browser = groups.find((group) => group.name === "browser");
  const backend = groups.find((group) => group.name === "backend-service-backed");
  if (JSON.stringify(browser?.summaryTargets) !== JSON.stringify(expectedBrowser)) {
    throw new Error(`${label} browser summary group must derive service-backed browser leaves`);
  }
  if (JSON.stringify(backend?.summaryTargets) !== JSON.stringify(expectedBackend)) {
    throw new Error(`${label} backend-service-backed summary group must derive backend service targets`);
  }
}
for (const target of ["test-service-backed", "check-service-backed"]) {
  const children = serviceBackedScheduleChildren(context, target);
  for (const child of [webserverBackedStage.target, isolatedStage.target]) {
    if (!children.includes(child)) {
      throw new Error(`${target} service-backed schedule must include ${child}`);
    }
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
  fail "missing tools/check_schedule_manifest.json"
fi
check_schedule_text="$(check_schedule_targets)"
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
const topologyTargets = new Map((topology.task_surface?.targets ?? []).map((entry) => [entry.name, entry]));
for (const [target, profile] of [
  ["check-service-backed", "nested_service_backed_scheduler"],
  ["check-browser-e2e-duration-baseline-drift", "post_service_duration_check"],
  ["browser-e2e-task-surface-check", "after_setup_cpu"],
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
if (limits.host_cpu !== 12 || limits.host_io !== 12 || limits.service_stack !== 1) {
  throw new Error("check schedule must declare host_cpu, host_io, and service_stack limits");
}
const service = (schedule.work_units ?? []).find((entry) => entry.target === "check-service-backed");
if (!service) {
  throw new Error("missing check-service-backed work unit");
}
const browserDrift = (schedule.work_units ?? []).find((entry) => entry.target === "check-browser-e2e-duration-baseline-drift");
if (!browserDrift) {
  throw new Error("missing check-browser-e2e-duration-baseline-drift work unit");
}
if (JSON.stringify(browserDrift.needs ?? []) !== JSON.stringify(["check-service-backed"])) {
  throw new Error("check-browser-e2e-duration-baseline-drift must depend on check-service-backed");
}
if (browserDrift.resource_claims?.host_cpu !== 1 || Object.keys(browserDrift.resource_claims ?? {}).length !== 1) {
  throw new Error("check-browser-e2e-duration-baseline-drift must claim only host_cpu=1");
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
require_browser_batch_target browser-e2e-measurement measurement 1 direct
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
  fail "scripts/start-web-e2e.sh must wait for Postgres and MinIO through dev-services.sh"
fi
if ! grep -Fq 'docker compose -f "${COMPOSE_FILE}" up -d postgres minio' "$start_web_e2e_script"; then
  fail "scripts/start-web-e2e.sh must keep Compose-backed startup for standalone browser E2E"
fi

if ! grep -Fq 'cartulary.browser_e2e_batch_manifest.v4' "$browser_batch_manifest"; then
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
if ! grep -Fq '"scheduler_dependency_policy": "after_backend_and_prior_browser"' "$browser_batch_manifest"; then
  fail "browser E2E batch manifest must declare scheduler dependency policy for isolated work"
fi
if ! grep -Fq '"kind": "duration_balanced_specs"' "$browser_batch_manifest"; then
  fail "browser E2E batch manifest must route functional browser work through duration-balanced specs"
fi
if ! grep -Fq '"execution_dependency": "browser_functional"' "$browser_batch_manifest"; then
  fail "browser E2E batch manifest must declare manifest execution dependencies for browser groups"
fi
if ! grep -Fq '"execution_dependency": "browser_support"' "$browser_batch_manifest"; then
  fail "browser E2E batch manifest must declare browser_support manifest selection for support groups"
fi
if ! grep -Fq 'cartulary.browser_e2e_duration_baselines.v1' "$browser_duration_baselines"; then
  fail "browser E2E duration baselines must declare their schema"
fi
if ! grep -Fq 'browser_functional' "$browser_shard_plan_script"; then
  fail "browser shard planner must select authoritative browser_functional rows"
fi
if ! grep -Fq 'shard_target_ms' "$browser_duration_baselines"; then
  fail "browser E2E duration baselines must declare shard_target_ms"
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
if ! grep -Fq 'CARTULARY_PLAYWRIGHT_WORKER_COUNT="${#shard_names[@]}"' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must provision one worker-admin slot per parallel shard"
fi
if ! grep -Fq 'CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET="$shard_index"' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must offset each parallel shard to a distinct worker-admin slot"
fi
if ! grep -Fq 'merge-reports' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must merge functional shard reports before phase summaries"
fi
if ! grep -Fq 'playwright_parallelism' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must cap browser spec shard parallelism by PLAYWRIGHT_WORKERS"
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
if ! grep -Fq 'run-browser-e2e-batch.sh" resettable' "$resettable_script"; then
  fail "scripts/run-browser-e2e-resettable.sh must delegate resettable sequencing to the browser batch runner"
fi
if ! grep -Fq '/api/v1/test/runtime/reset' "$reset_script"; then
  fail "scripts/reset-web-e2e-stack.sh must call the test runtime reset route"
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
const phaseSchemaID = "cartulary.phase_test_map.v1";
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
