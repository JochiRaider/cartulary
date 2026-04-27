#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
makefile="$repo_root/Makefile"
functional_script="$repo_root/scripts/run-browser-e2e-functional.sh"
browser_batch_script="$repo_root/scripts/run-browser-e2e-batch.sh"
browser_batch_manifest="$repo_root/tools/browser_e2e_batch_manifest.json"
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

fail() {
  echo "$*" >&2
  exit 1
}

extract_target_block() {
  local target="$1"
  awk -v target="$target" '
    $0 ~ "^" target ":" { in_block=1; print; next }
    in_block && /^[^[:space:]].*:/ { exit }
    in_block { print }
  ' "$makefile"
}

require_browser_owned_stack_target_uses_built_binaries() {
  local target="$1"
  local block
  block="$(extract_target_block "$target")"

  if [[ -z "$block" ]]; then
    fail "Makefile must define a non-empty $target block"
  fi
  if ! printf '%s\n' "$block" | grep -Fq 'build-server'; then
    fail "$target must depend on build-server"
  fi
  if ! printf '%s\n' "$block" | grep -Fq 'build-migrate'; then
    fail "$target must depend on build-migrate"
  fi
  if ! printf '%s\n' "$block" | grep -Fq '$(BROWSER_E2E_OWNED_STACK_ENV)'; then
    fail "$target must use BROWSER_E2E_OWNED_STACK_ENV"
  fi
}

schedule_targets() {
  local schedule_target="$1"
  local kind="${2:-}"

  "$node_bin" - "$schedule_manifest" "$schedule_target" "$kind" <<'EOF'
const fs = require("node:fs");

const [manifestFile, scheduleTarget, kind] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
if (manifest.schema_id !== "cartulary.service_backed_schedule.v6") {
  throw new Error("service-backed schedule manifest must declare schema_id=cartulary.service_backed_schedule.v6");
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
if (manifest.schema_id !== "cartulary.service_backed_schedule.v6") {
  throw new Error("service-backed schedule manifest must declare schema_id=cartulary.service_backed_schedule.v6");
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
if (manifest.schema_id !== "cartulary.check_schedule.v1") {
  throw new Error("check schedule manifest must declare schema_id=cartulary.check_schedule.v1");
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
if (manifest.schema_id !== "cartulary.check_schedule.v1") {
  throw new Error("check schedule manifest must declare schema_id=cartulary.check_schedule.v1");
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

browser_e2e_owned_stack_env="$(sed -n 's/^BROWSER_E2E_OWNED_STACK_ENV[[:space:]]*:=//p' "$makefile" | head -n 1)"
if [[ -z "$browser_e2e_owned_stack_env" ]]; then
  fail "Makefile must define BROWSER_E2E_OWNED_STACK_ENV"
fi
if ! printf '%s\n' "$browser_e2e_owned_stack_env" | grep -Fq 'CARTULARY_SERVER_BIN=$(SERVER_BIN)'; then
  fail "BROWSER_E2E_OWNED_STACK_ENV must export CARTULARY_SERVER_BIN=$(SERVER_BIN)"
fi
if ! printf '%s\n' "$browser_e2e_owned_stack_env" | grep -Fq 'CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN)'; then
  fail "BROWSER_E2E_OWNED_STACK_ENV must export CARTULARY_MIGRATE_BIN=$(MIGRATE_BIN)"
fi
if ! printf '%s\n' "$browser_e2e_owned_stack_env" | grep -Fq 'CARTULARY_TEST_SERVICES_BIN=$(TEST_SERVICES_BIN)'; then
  fail "BROWSER_E2E_OWNED_STACK_ENV must export CARTULARY_TEST_SERVICES_BIN=$(TEST_SERVICES_BIN)"
fi
if ! printf '%s\n' "$browser_e2e_owned_stack_env" | grep -Fq 'CARTULARY_WEB_E2E_USE_REPO_ROOT_BINARIES=1'; then
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

check_local_product_line="$(sed -n 's/^check-local-product:[[:space:]]*//p' "$makefile" | head -n 1)"
if [[ -z "$check_local_product_line" ]]; then
  fail "Makefile must define check-local-product prerequisites"
fi
check_meta_validation_line="$(sed -n 's/^check-meta-validation:[[:space:]]*//p' "$makefile" | head -n 1)"
if [[ -z "$check_meta_validation_line" ]]; then
  fail "Makefile must define check-meta-validation prerequisites"
fi
if ! printf '%s\n' "$check_meta_validation_line" | rg -q '(^|[[:space:]])check-static-validation($|[[:space:]])'; then
  fail "check-meta-validation must include static validation without moving browser suites into local product checks"
fi
if ! printf '%s\n' "$check_meta_validation_line" | rg -q '(^|[[:space:]])check-harness-smoke($|[[:space:]])'; then
  fail "check-meta-validation must include harness smoke without moving browser suites into local product checks"
fi

read -r -a local_product_prereqs <<<"$check_local_product_line"
browser_targets=()
for prereq in "${local_product_prereqs[@]}"; do
  if [[ "$prereq" == browser-e2e* ]]; then
    browser_targets+=("$prereq")
  fi
done

if [[ "${#browser_targets[@]}" -ne 0 ]]; then
  fail "check-local-product must not include browser-e2e* prerequisites, found: ${browser_targets[*]}"
fi

browser_e2e_block="$(extract_target_block browser-e2e)"
if [[ -z "$browser_e2e_block" ]]; then
  fail "Makefile must define a non-empty browser-e2e block"
fi
if ! printf '%s\n' "$browser_e2e_block" | grep -Fq '$(TEST_SERVICES_BIN) run --'; then
  fail 'browser-e2e must wrap aggregate browser children through $(TEST_SERVICES_BIN)'
fi
if ! printf '%s\n' "$browser_e2e_block" | grep -Fq './scripts/start-web-e2e.sh -- ./scripts/run-browser-e2e-batch.sh isolated'; then
  fail "browser-e2e must run the final isolated browser batch inside one owned stack"
fi
if printf '%s\n' "$browser_e2e_block" | grep -Fq -- '-j$(BROWSER_E2E_JOBS)'; then
  fail 'browser-e2e must not fan out aggregate browser children with -j$(BROWSER_E2E_JOBS)'
fi
if printf '%s\n' "$browser_e2e_block" | grep -Fq 'run-browser-e2e-batch.sh all'; then
  fail "browser-e2e must not run the removed all browser batch"
fi
if ! printf '%s\n' "$browser_e2e_block" | grep -Fq -- '--projection browser-e2e'; then
  fail "browser-e2e must summarize the manifest-declared isolated-batch children through task-surface projection"
fi

test_service_block="$(extract_target_block test-service-backed)"
if [[ -z "$test_service_block" ]]; then
  fail "Makefile must define a non-empty test-service-backed block"
fi

if ! printf '%s\n' "$test_service_block" | grep -Fq '$(RUN_SERVICE_BACKED_SCHEDULE_SCRIPT) --target test-service-backed --manifest "$(SERVICE_BACKED_SCHEDULE_MANIFEST)"'; then
  fail "test-service-backed must delegate to the service-backed scheduler manifest"
fi
if printf '%s\n' "$test_service_block" | grep -Fq -- '--jobs'; then
  fail "test-service-backed must not pass a fixed scheduler job cap"
fi
if ! printf '%s\n' "$test_service_block" | grep -Fq -- '--defer-summary'; then
  fail "test-service-backed must defer target summary until after synchronous suite release checks"
fi
if ! printf '%s\n' "$test_service_block" | grep -Fq '$(TEST_OUTPUT_SCRIPT) target-summary test-service-backed $$requested --projection test-service-backed'; then
  fail "test-service-backed must finalize target summary after synchronous suite release checks"
fi
if printf '%s\n' "$test_service_block" | rg -q 'test-service-backed-lane-(a|b|browser)'; then
  fail "test-service-backed must not invoke fixed service-backed lane targets"
fi
if ! printf '%s\n' "$test_service_block" | grep -Fq 'build-server'; then
  fail "test-service-backed must prebuild server before service-backed scheduling"
fi
if ! printf '%s\n' "$test_service_block" | grep -Fq 'build-migrate'; then
  fail "test-service-backed must prebuild migrate before service-backed scheduling"
fi

mapfile -t test_service_browser_targets < <(schedule_targets test-service-backed browser)
if [[ "$(printf '%s\n' "${test_service_browser_targets[@]}")" != "browser-e2e-webserver-backed" ]]; then
  fail "test-service-backed schedule must own only browser-e2e-webserver-backed as browser work, found: ${test_service_browser_targets[*]:-none}"
fi

check_service_block="$(extract_target_block check-service-backed)"
if [[ -z "$check_service_block" ]]; then
  fail "Makefile must define a non-empty check-service-backed block"
fi

if ! printf '%s\n' "$check_service_block" | grep -Fq '$(RUN_SERVICE_BACKED_SCHEDULE_SCRIPT) --target check-service-backed --manifest "$(SERVICE_BACKED_SCHEDULE_MANIFEST)"'; then
  fail "check-service-backed must delegate to the service-backed scheduler manifest"
fi
if printf '%s\n' "$check_service_block" | grep -Fq -- '--jobs'; then
  fail "check-service-backed must not pass a fixed scheduler job cap"
fi
if ! printf '%s\n' "$check_service_block" | grep -Fq -- '--defer-summary'; then
  fail "check-service-backed must defer target summary until after synchronous suite release checks"
fi
if ! printf '%s\n' "$check_service_block" | grep -Fq '$(TEST_OUTPUT_SCRIPT) target-summary check-service-backed $$requested --projection check-service-backed'; then
  fail "check-service-backed must finalize target summary after synchronous suite release checks"
fi
if printf '%s\n' "$check_service_block" | rg -q 'check-service-backed-lane-[ab]'; then
  fail "check-service-backed must not invoke fixed service-backed lane targets"
fi
if ! printf '%s\n' "$check_service_block" | grep -Fq 'build-server'; then
  fail "check-service-backed must prebuild server before service-backed scheduling"
fi
if ! printf '%s\n' "$check_service_block" | grep -Fq 'build-migrate'; then
  fail "check-service-backed must prebuild migrate before service-backed scheduling"
fi

mapfile -t check_service_browser_targets < <(schedule_targets check-service-backed browser)
if [[ "$(printf '%s\n' "${check_service_browser_targets[@]}")" != "browser-e2e-webserver-backed" ]]; then
  fail "check-service-backed schedule must own only webserver-backed browser work, found: ${check_service_browser_targets[*]:-none}"
fi

"$node_bin" - "$schedule_manifest" <<'EOF'
const fs = require("node:fs");

const [manifestFile] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
if (manifest.schema_id !== "cartulary.service_backed_schedule.v6") {
  throw new Error("service-backed schedule manifest must declare schema_id=cartulary.service_backed_schedule.v6");
}
for (const schedule of manifest.schedules ?? []) {
  const limits = schedule.resource_limits ?? {};
  if (Object.hasOwn(limits, "browser")) {
    throw new Error(`${schedule.target} must not declare removed generic browser resource limit`);
  }
  for (const source of schedule.work_unit_sources ?? []) {
    const claims = source.resource_claims ?? {};
    if (Object.hasOwn(claims, "browser")) {
      throw new Error(`${schedule.target} ${source.target} must not claim removed generic browser resource`);
    }
    if (source.class !== "browser") {
      continue;
    }
    const browserStage = String(source.browser_stage ?? "");
    const laneResource = `browser_stage_${browserStage.replaceAll("-", "_")}`;
    if (!Object.hasOwn(limits, "browser_stack")) {
      throw new Error(`${schedule.target} must declare browser_stack resource limit`);
    }
    if (!Object.hasOwn(limits, laneResource)) {
      throw new Error(`${schedule.target} must declare ${laneResource} resource limit`);
    }
    if (!Object.hasOwn(claims, "browser_stack")) {
      throw new Error(`${schedule.target} ${source.target} must claim browser_stack`);
    }
    if (!Object.hasOwn(claims, laneResource)) {
      throw new Error(`${schedule.target} ${source.target} must claim ${laneResource}`);
    }
  }
}
EOF

mapfile -t test_fast_service_browser_targets < <(schedule_targets test-fast-service-backed browser)
if [[ "${#test_fast_service_browser_targets[@]}" -ne 0 ]]; then
  fail "test-fast-service-backed schedule must remain backend-only, found browser targets: ${test_fast_service_browser_targets[*]}"
fi

"$node_bin" - "$task_surface_manifest" <<'EOF'
const fs = require("node:fs");

const [manifestFile] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
for (const profileName of ["test", "check"]) {
  const groups = manifest.summary_profiles?.[profileName]?.groups ?? [];
  const browser = groups.find((group) => group.name === "browser");
  const backend = groups.find((group) => group.name === "backend-service-backed");
  const expectedBrowser = ["browser-e2e-webserver-backed", "browser-e2e-stateful", "browser-e2e-measurement", "browser-e2e-visual"];
  const expectedBackend = ["backend-integration", "backend-integration-support", "backend-store", "backend-process"];
  if (JSON.stringify(browser?.targets) !== JSON.stringify(expectedBrowser)) {
    throw new Error(`${profileName} browser summary group must report service-backed webserver and isolated browser children`);
  }
  if (JSON.stringify(backend?.targets) !== JSON.stringify(expectedBackend)) {
    throw new Error(`${profileName} backend-service-backed summary group must contain backend service targets`);
  }
}
EOF

test_block="$(extract_target_block test)"
if [[ -z "$test_block" ]]; then
  fail "Makefile must define a non-empty test block"
fi
if ! printf '%s\n' "$test_block" | grep -Fq -- '--step test-service-backed --step browser-e2e'; then
  fail "test must run service-backed backend work followed by the aggregate browser-e2e batch"
fi
if printf '%s\n' "$test_block" | grep -Fq 'test-isolated'; then
  fail "test must not route browser evidence through test-isolated"
fi

check_block="$(extract_target_block check)"
if [[ -z "$check_block" ]]; then
  fail "Makefile must define a non-empty check block"
fi
if ! printf '%s\n' "$check_block" | grep -Fq '$(RUN_CHECK_SCHEDULE_SCRIPT)'; then
  fail "check must delegate to the check scheduler"
fi
if ! printf '%s\n' "$check_block" | grep -Fq -- '--resource-limit cpu=$(CHECK_JOBS)'; then
  fail "check must pass CHECK_JOBS as the check scheduler cpu resource limit"
fi
if printf '%s\n' "$check_block" | grep -Fq -- '--step browser-e2e'; then
  fail "check must not run browser-e2e as a final serial step"
fi
if printf '%s\n' "$check_block" | grep -Fq 'check-isolated'; then
  fail "check must not route browser evidence through check-isolated"
fi
if rg -q '^check-pre-browser:' "$makefile"; then
  fail "check-pre-browser must not remain as legacy browser orchestration"
fi
if ! [[ -f "$check_schedule_manifest" ]]; then
  fail "missing tools/check_schedule_manifest.json"
fi
check_schedule_text="$(check_schedule_targets)"
for scheduled_target in check-setup-blockers check-build-prereqs check-service-backed check-local-product check-frontend-unit check-meta-validation browser-e2e; do
  if ! printf '%s\n' "$check_schedule_text" | rg -q "^${scheduled_target}$"; then
    fail "check schedule must include $scheduled_target"
  fi
done
if [[ "$(check_schedule_field browser-e2e needs)" != "check-service-backed,check-go-test-duration-baseline-drift,check-local-product,check-frontend-unit,check-meta-validation" ]]; then
  fail "check schedule must run browser-e2e after service-backed and heavy local check work"
fi
if [[ "$(check_schedule_field browser-e2e resource_claims)" != "cpu,service_stack" ]]; then
  fail "check schedule browser-e2e must claim cpu and service_stack"
fi

if ! rg -q '^browser-e2e-webserver-backed:' "$makefile"; then
  fail "Makefile must define browser-e2e-webserver-backed"
fi

browser_functional_block="$(awk '
  /^browser-e2e-functional:/ { in_block=1; next }
  in_block && /^[^[:space:]].*:/ { exit }
  in_block { print }
' "$makefile")"
if [[ -z "$browser_functional_block" ]]; then
  fail "Makefile must define a non-empty browser-e2e-functional block"
fi
if ! printf '%s\n' "$browser_functional_block" | grep -Fq './scripts/run-browser-e2e-functional.sh'; then
  if ! printf '%s\n' "$browser_functional_block" | grep -Fq './scripts/run-browser-e2e-batch.sh functional'; then
    fail "browser-e2e-functional must delegate to the functional browser batch"
  fi
fi

browser_stateful_block="$(awk '
  /^browser-e2e-stateful:/ { in_block=1; next }
  in_block && /^[^[:space:]].*:/ { exit }
  in_block { print }
' "$makefile")"
if [[ -z "$browser_stateful_block" ]]; then
  fail "Makefile must define a non-empty browser-e2e-stateful block"
fi
if ! printf '%s\n' "$browser_stateful_block" | grep -Fq './scripts/run-browser-e2e-stateful.sh'; then
  if ! printf '%s\n' "$browser_stateful_block" | grep -Fq './scripts/run-browser-e2e-batch.sh stateful'; then
    fail "browser-e2e-stateful must delegate to the stateful browser batch"
  fi
fi

browser_measurement_block="$(awk '
  /^browser-e2e-measurement:/ { in_block=1; next }
  in_block && /^[^[:space:]].*:/ { exit }
  in_block { print }
' "$makefile")"
if [[ -z "$browser_measurement_block" ]]; then
  fail "Makefile must define a non-empty browser-e2e-measurement block"
fi
if ! printf '%s\n' "$browser_measurement_block" | grep -Fq './scripts/run-browser-e2e-measurement.sh'; then
  if ! printf '%s\n' "$browser_measurement_block" | grep -Fq './scripts/run-browser-e2e-batch.sh measurement'; then
    fail "browser-e2e-measurement must delegate to the measurement browser batch"
  fi
fi
if printf '%s\n' "$browser_measurement_block" | grep -Fq 'Core 05-bound timing evidence'; then
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

if ! grep -Fq 'cartulary.browser_e2e_batch_manifest.v2' "$browser_batch_manifest"; then
  fail "browser E2E batch manifest must declare its schema"
fi
if ! grep -Fq '"kind": "duration_balanced_specs"' "$browser_batch_manifest"; then
  fail "browser E2E batch manifest must route functional browser work through duration-balanced specs"
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
if grep -Fq '"name": "all"' "$browser_batch_manifest"; then
  fail "browser E2E batch manifest must not keep the removed all stage"
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
if ! grep -Fq 'target-summary "$target"' "$browser_batch_script"; then
  fail "browser E2E batch runner must emit per-group target summaries"
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
if grep -Fq 'playwright-grep-many' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must not keep the old all-phase Playwright grep batch"
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

for required_browser_functional_phase in phase1 phase2 phase3 phase4; do
  if ! printf '%s\n' "${browser_functional_phases[@]}" | grep -Fxq "$required_browser_functional_phase"; then
    fail "authoritative browser_functional manifest phases must include $required_browser_functional_phase, found: ${browser_functional_phases[*]:-none}"
  fi
done
if printf '%s\n' "${browser_functional_phases[@]}" | grep -Fxq phase0; then
  fail "authoritative browser_functional manifest phases must skip phase0, found: ${browser_functional_phases[*]}"
fi
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
for hardcoded_browser_functional_file in phase1.spec.ts phase2.spec.ts phase3.spec.ts phase4.spec.ts; do
  if grep -Fq "$hardcoded_browser_functional_file" "$webserver_batch_config"; then
    fail "apps/web/playwright.webserver-backed.config.ts must not hardcode browser functional file $hardcoded_browser_functional_file"
  fi
done
if ! grep -Fq 'webE2EBaseConfig' "$webserver_batch_config"; then
  fail "apps/web/playwright.webserver-backed.config.ts must use the shared Playwright web E2E config"
fi
if ! grep -Fq 'webE2EBaseConfig' "$shared_playwright_config"; then
  fail "apps/web/playwright.shared.config.ts must expose the shared Playwright web E2E config"
fi

if ! grep -Fq 'phase1 authoritative browser_stateful' "$stateful_script"; then
  fail "scripts/run-browser-e2e-stateful.sh must execute Phase 1 browser_stateful rows through the manifest"
fi
if grep -Fq 'e2e/phase1.clock.spec.ts' "$stateful_script"; then
  fail "scripts/run-browser-e2e-stateful.sh must not raw-select e2e/phase1.clock.spec.ts"
fi
if ! grep -Fq 'claim_bearing": false' "$measurement_script"; then
  fail "scripts/run-browser-e2e-measurement.sh must emit claim_bearing=false ordinary measurement metadata"
fi
if ! grep -Fq 'evidence_kind": "ordinary_measurement"' "$measurement_script"; then
  fail "scripts/run-browser-e2e-measurement.sh must emit ordinary_measurement evidence_kind metadata"
fi
if ! grep -Fq 'phase3 authoritative browser_measurement' "$measurement_script"; then
  fail "scripts/run-browser-e2e-measurement.sh must execute Phase 3 browser_measurement rows through the manifest"
fi
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
for (const [phase, expectedFor] of [
  [
    "phase1",
    (entry) =>
      new Set(["E-1-04", "E-1-05"]).has(entry.id)
        ? "browser_stateful"
        : "browser_functional",
  ],
  ["phase2", () => "browser_functional"],
  ["phase4", () => "browser_functional"],
]) {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(root, "tools", `${phase}_test_map.json`), "utf8"),
  );
  for (const entry of manifest.e2e ?? []) {
    if (entry.coverage !== "authoritative" || entry.runner !== "playwright") {
      continue;
    }
    const expected = expectedFor(entry);
    if (entry.execution_dependency !== expected) {
      console.error(
        `${phase} authoritative e2e row ${entry.id} must declare execution_dependency=${expected}`,
      );
      process.exit(1);
    }
  }
}
EOF
then
  fail "Phase 1 and Phase 2 authoritative browser manifest rows must carry the canonical execution_dependency for their layer"
fi
