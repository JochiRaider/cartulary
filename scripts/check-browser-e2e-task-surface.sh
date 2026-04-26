#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
makefile="$repo_root/Makefile"
functional_script="$repo_root/scripts/run-browser-e2e-functional.sh"
webserver_batch_script="$repo_root/scripts/lib/run-playwright-webserver-batch.sh"
webserver_batch_config="$repo_root/apps/web/playwright.webserver-backed.config.ts"
stateful_script="$repo_root/scripts/run-browser-e2e-stateful.sh"
measurement_script="$repo_root/scripts/run-browser-e2e-measurement.sh"
resettable_script="$repo_root/scripts/run-browser-e2e-resettable.sh"
reset_script="$repo_root/scripts/reset-web-e2e-stack.sh"
webserver_backed_script="$repo_root/scripts/run-browser-e2e-webserver-backed.sh"
start_web_e2e_script="$repo_root/scripts/start-web-e2e.sh"
schedule_manifest="$repo_root/tools/service_backed_schedule_manifest.json"
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
if (manifest.schema_id !== "cartulary.service_backed_schedule.v2") {
  throw new Error("service-backed schedule manifest must declare schema_id=cartulary.service_backed_schedule.v2");
}
const schedules = manifest.schedules.filter((entry) => entry.target === scheduleTarget);
if (schedules.length !== 1) {
  throw new Error(`expected exactly one schedule for ${scheduleTarget}, found ${schedules.length}`);
}
const targets = schedules[0].children
  .filter((entry) => kind === "" || entry.kind === kind)
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
if (manifest.schema_id !== "cartulary.service_backed_schedule.v2") {
  throw new Error("service-backed schedule manifest must declare schema_id=cartulary.service_backed_schedule.v2");
}
const schedules = manifest.schedules.filter((entry) => entry.target === scheduleTarget);
if (schedules.length !== 1) {
  throw new Error(`expected exactly one schedule for ${scheduleTarget}, found ${schedules.length}`);
}
const children = schedules[0].children.filter((entry) => entry.target === childTarget);
if (children.length !== 1) {
  throw new Error(`expected exactly one child ${childTarget} in ${scheduleTarget}, found ${children.length}`);
}
const value = children[0][field] ?? [];
if (!Array.isArray(value)) {
  throw new Error(`${scheduleTarget} child ${childTarget} ${field} must be an array`);
}
if (value.length > 0) {
  process.stdout.write(`${value.join("\n")}\n`);
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
const child = schedule.children.find((entry) => entry.target === childTarget);
if (!child) {
  throw new Error(`missing child ${childTarget} in ${scheduleTarget}`);
}
process.stdout.write(`${child.weight}\n`);
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
  fail "check-parallel must include static validation without moving browser suites into check-heavy"
fi
if ! printf '%s\n' "$check_parallel_line" | rg -q '(^|[[:space:]])check-harness-smoke($|[[:space:]])'; then
  fail "check-parallel must include harness smoke without moving browser suites into check-heavy"
fi

read -r -a heavy_prereqs <<<"$check_heavy_line"
browser_targets=()
for prereq in "${heavy_prereqs[@]}"; do
  if [[ "$prereq" == browser-e2e* ]]; then
    browser_targets+=("$prereq")
  fi
done

if [[ "${#browser_targets[@]}" -ne 0 ]]; then
  fail "check-heavy must not include browser-e2e* prerequisites, found: ${browser_targets[*]}"
fi

browser_e2e_block="$(extract_target_block browser-e2e)"
if [[ -z "$browser_e2e_block" ]]; then
  fail "Makefile must define a non-empty browser-e2e block"
fi
if ! printf '%s\n' "$browser_e2e_block" | grep -Fq '$(TEST_SERVICES_BIN) run --'; then
  fail 'browser-e2e must wrap aggregate browser children through $(TEST_SERVICES_BIN)'
fi
if ! printf '%s\n' "$browser_e2e_block" | grep -Fq -- '-j$(BROWSER_E2E_JOBS)'; then
  fail 'browser-e2e must run aggregate browser children with -j$(BROWSER_E2E_JOBS)'
fi
if ! printf '%s\n' "$browser_e2e_block" | grep -Fq -- '--output-sync=target'; then
  fail "browser-e2e must use output-sync for parallel aggregate browser children"
fi
if ! printf '%s\n' "$browser_e2e_block" | grep -Fq 'browser-e2e-webserver-backed browser-e2e-stateful browser-e2e-resettable'; then
  fail "browser-e2e must run webserver-backed, stateful, and resettable browser children"
fi

test_service_block="$(extract_target_block test-service-backed)"
if [[ -z "$test_service_block" ]]; then
  fail "Makefile must define a non-empty test-service-backed block"
fi

if ! printf '%s\n' "$test_service_block" | grep -Fq '$(RUN_SERVICE_BACKED_SCHEDULE_SCRIPT) --target test-service-backed --jobs $(TEST_SERVICE_BACKED_JOBS)'; then
  fail "test-service-backed must delegate to the service-backed scheduler with TEST_SERVICE_BACKED_JOBS"
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
if [[ "${#test_service_browser_targets[@]}" -ne 1 ]]; then
  fail "test-service-backed schedule must include exactly one browser target, found: ${test_service_browser_targets[*]:-none}"
fi
if [[ "${test_service_browser_targets[0]}" != "browser-e2e-webserver-backed" ]]; then
  fail "test-service-backed schedule must use browser-e2e-webserver-backed as its browser target, found: ${test_service_browser_targets[0]}"
fi
mapfile -t test_service_browser_claims < <(schedule_child_array test-service-backed browser-e2e-webserver-backed resource_claims)
if ! printf '%s\n' "${test_service_browser_claims[@]}" | rg -q '^browser-webserver$'; then
  fail "test-service-backed browser child must claim browser-webserver capacity"
fi
mapfile -t test_service_browser_exclusive < <(schedule_child_array test-service-backed browser-e2e-webserver-backed exclusive_tags)
if printf '%s\n' "${test_service_browser_exclusive[@]}" | rg -q '^(postgres|minio)$'; then
  fail "test-service-backed browser child must not declare broad postgres or minio exclusivity"
fi
if (( $(schedule_child_weight test-service-backed browser-e2e-webserver-backed) < 95 )); then
  fail "test-service-backed browser child weight must keep it in the first scheduling wave"
fi

check_service_block="$(extract_target_block check-service-backed)"
if [[ -z "$check_service_block" ]]; then
  fail "Makefile must define a non-empty check-service-backed block"
fi

if ! printf '%s\n' "$check_service_block" | grep -Fq '$(RUN_SERVICE_BACKED_SCHEDULE_SCRIPT) --target check-service-backed --jobs $(SERVICE_BACKED_JOBS)'; then
  fail "check-service-backed must delegate to the service-backed scheduler with SERVICE_BACKED_JOBS"
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
if [[ "${#check_service_browser_targets[@]}" -ne 1 ]]; then
  fail "check-service-backed schedule must include exactly one browser target, found: ${check_service_browser_targets[*]:-none}"
fi
if [[ "${check_service_browser_targets[0]}" != "browser-e2e-webserver-backed" ]]; then
  fail "check-service-backed schedule must use browser-e2e-webserver-backed as its browser target, found: ${check_service_browser_targets[0]}"
fi
mapfile -t check_service_browser_claims < <(schedule_child_array check-service-backed browser-e2e-webserver-backed resource_claims)
if ! printf '%s\n' "${check_service_browser_claims[@]}" | rg -q '^browser-webserver$'; then
  fail "check-service-backed browser child must claim browser-webserver capacity"
fi
mapfile -t check_service_browser_exclusive < <(schedule_child_array check-service-backed browser-e2e-webserver-backed exclusive_tags)
if printf '%s\n' "${check_service_browser_exclusive[@]}" | rg -q '^(postgres|minio)$'; then
  fail "check-service-backed browser child must not declare broad postgres or minio exclusivity"
fi
if (( $(schedule_child_weight check-service-backed browser-e2e-webserver-backed) < 95 )); then
  fail "check-service-backed browser child weight must keep it in the first scheduling wave"
fi

check_summary_groups_line="$(sed -n 's/^CHECK_SUMMARY_GROUPS[[:space:]]*:=[[:space:]]*//p' "$makefile" | head -n 1)"
if [[ -z "$check_summary_groups_line" ]]; then
  fail "Makefile must define CHECK_SUMMARY_GROUPS"
fi
if ! printf '%s\n' "$check_summary_groups_line" | grep -Fq 'browser-webserver-backed=browser-e2e-webserver-backed'; then
  fail "CHECK_SUMMARY_GROUPS must report browser webserver-backed duration separately"
fi
backend_summary_group="$(printf '%s\n' "$check_summary_groups_line" | tr ';' '\n' | sed -n 's/^backend-service-backed=//p')"
if [[ "$backend_summary_group" != "backend-integration,backend-integration-support,backend-store,backend-process" ]]; then
  fail "backend-service-backed summary group must contain backend service targets, found: ${backend_summary_group:-none}"
fi
browser_summary_group="$(printf '%s\n' "$check_summary_groups_line" | tr ';' '\n' | sed -n 's/^browser-webserver-backed=//p')"
if [[ "$browser_summary_group" != "browser-e2e-webserver-backed" ]]; then
  fail "browser-webserver-backed summary group must contain only browser-e2e-webserver-backed, found: ${browser_summary_group:-none}"
fi

test_summary_groups_line="$(sed -n 's/^TEST_SUMMARY_GROUPS[[:space:]]*:=[[:space:]]*//p' "$makefile" | head -n 1)"
if [[ -z "$test_summary_groups_line" ]]; then
  fail "Makefile must define TEST_SUMMARY_GROUPS"
fi
if [[ "$test_summary_groups_line" != "$check_summary_groups_line" ]]; then
  fail "TEST_SUMMARY_GROUPS must match CHECK_SUMMARY_GROUPS so service-backed and browser webserver-backed durations stay comparable"
fi

check_isolated_block="$(extract_target_block check-isolated)"
if [[ -z "$check_isolated_block" ]]; then
  fail "Makefile must define a non-empty check-isolated block"
fi
if ! printf '%s\n' "$check_isolated_block" | grep -Fq '$(TEST_SERVICES_BIN) run --'; then
  fail 'check-isolated must wrap isolated browser children through $(TEST_SERVICES_BIN)'
fi
if ! printf '%s\n' "$check_isolated_block" | grep -Fq 'browser-e2e-stateful browser-e2e-resettable'; then
  fail "check-isolated must run browser-e2e-stateful and browser-e2e-resettable"
fi
if printf '%s\n' "$check_isolated_block" | grep -Eq 'browser-e2e-measurement|browser-e2e-visual'; then
  fail "check-isolated must not invoke measurement or visual as separate owned-stack targets"
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
  fail "browser-e2e-functional must delegate to scripts/run-browser-e2e-functional.sh"
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
  fail "browser-e2e-stateful must delegate to scripts/run-browser-e2e-stateful.sh"
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
  fail "browser-e2e-measurement must delegate to scripts/run-browser-e2e-measurement.sh"
fi
if printf '%s\n' "$browser_measurement_block" | grep -Fq 'Core 05-bound timing evidence'; then
  fail "browser-e2e-measurement must be labeled ordinary measurement, not Core 05-bound claim evidence"
fi

if ! [[ -f "$functional_script" ]]; then
  fail "missing scripts/run-browser-e2e-functional.sh"
fi
if ! [[ -f "$webserver_batch_script" ]]; then
  fail "missing scripts/lib/run-playwright-webserver-batch.sh"
fi
if ! [[ -f "$webserver_batch_config" ]]; then
  fail "missing apps/web/playwright.webserver-backed.config.ts"
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
if ! grep -Fq 'playwright-grep-many' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must build one multi-phase Playwright grep from manifest titles"
fi
if ! grep -Fq 'list-phases' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must discover browser_functional phases through phase-manifest list-phases"
fi
if ! grep -Fq 'playwright-count "$phase" authoritative browser_functional' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must filter discovered phases by authoritative browser_functional manifest rows"
fi
if ! grep -Fq ':authoritative:browser_functional' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must select authoritative browser_functional manifest rows"
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
if ! grep -Fq 'CARTULARY_PHASE_ACCOUNTING_MODE=derived' "$webserver_batch_script"; then
  fail "scripts/lib/run-playwright-webserver-batch.sh must mark derived slices to avoid duration multiplication"
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
if ! grep -Fq 'CARTULARY_TEST_TARGET="$target"' "$resettable_script"; then
  fail "scripts/run-browser-e2e-resettable.sh must run each child with its own CARTULARY_TEST_TARGET"
fi
if ! grep -Fq 'run-browser-e2e-measurement.sh' "$resettable_script"; then
  fail "scripts/run-browser-e2e-resettable.sh must run browser-e2e-measurement inside the shared stack"
fi
if ! grep -Fq 'reset-web-e2e-stack.sh' "$resettable_script"; then
  fail "scripts/run-browser-e2e-resettable.sh must reset the shared stack between suites"
fi
if ! grep -Fq 'run-browser-e2e-visual.sh' "$resettable_script"; then
  fail "scripts/run-browser-e2e-resettable.sh must run browser-e2e-visual inside the shared stack"
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
    (entry) => (entry.id === "E-1-04" ? "browser_stateful" : "browser_functional"),
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
