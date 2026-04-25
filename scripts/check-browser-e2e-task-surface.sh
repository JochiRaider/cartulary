#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
makefile="$repo_root/Makefile"
functional_script="$repo_root/scripts/run-browser-e2e-functional.sh"
stateful_script="$repo_root/scripts/run-browser-e2e-stateful.sh"
measurement_script="$repo_root/scripts/run-browser-e2e-measurement.sh"
resettable_script="$repo_root/scripts/run-browser-e2e-resettable.sh"
reset_script="$repo_root/scripts/reset-web-e2e-stack.sh"
webserver_backed_script="$repo_root/scripts/run-browser-e2e-webserver-backed.sh"
start_web_e2e_script="$repo_root/scripts/start-web-e2e.sh"
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

check_service_block="$(extract_target_block check-service-backed)"
if [[ -z "$check_service_block" ]]; then
  fail "Makefile must define a non-empty check-service-backed block"
fi

for lane in check-service-backed-lane-a check-service-backed-lane-b; do
  if ! printf '%s\n' "$check_service_block" | rg -q "(^|[[:space:]])$lane($|[[:space:]])"; then
    fail "check-service-backed must invoke $lane"
  fi
done

service_browser_targets=()
while IFS= read -r line; do
  while IFS= read -r target; do
    [[ -n "$target" ]] && service_browser_targets+=("$target")
  done < <(printf '%s\n' "$line" | grep -o 'browser-e2e[^[:space:]]*' || true)
done <<<"$check_service_block"

if [[ "${#service_browser_targets[@]}" -ne 0 ]]; then
  fail "check-service-backed must delegate browser work through its lane targets, found direct browser targets: ${service_browser_targets[*]}"
fi

check_service_lane_b_block="$(extract_target_block check-service-backed-lane-b)"
if [[ -z "$check_service_lane_b_block" ]]; then
  fail "Makefile must define a non-empty check-service-backed-lane-b block"
fi

lane_browser_targets=()
while IFS= read -r line; do
  while IFS= read -r target; do
    [[ -n "$target" ]] && lane_browser_targets+=("$target")
  done < <(printf '%s\n' "$line" | grep -o 'browser-e2e[^[:space:]]*' || true)
done <<<"$check_service_lane_b_block"

if [[ "${#lane_browser_targets[@]}" -ne 1 ]]; then
  fail "check-service-backed-lane-b must invoke exactly one browser-e2e* target, found: ${lane_browser_targets[*]:-none}"
fi

if [[ "${lane_browser_targets[0]}" != "browser-e2e-webserver-backed" ]]; then
  fail "check-service-backed-lane-b must use browser-e2e-webserver-backed as its only browser target, found: ${lane_browser_targets[0]}"
fi

check_isolated_block="$(extract_target_block check-isolated)"
if [[ -z "$check_isolated_block" ]]; then
  fail "Makefile must define a non-empty check-isolated block"
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

if ! grep -Fq 'phase1 authoritative browser_functional' "$functional_script"; then
  fail "scripts/run-browser-e2e-functional.sh must execute Phase 1 browser_functional rows through the manifest"
fi
if grep -Fq 'e2e/phase1.spec.ts' "$functional_script"; then
  fail "scripts/run-browser-e2e-functional.sh must not raw-select e2e/phase1.spec.ts"
fi
if ! grep -Fq 'phase2 authoritative browser_functional' "$functional_script"; then
  fail "scripts/run-browser-e2e-functional.sh must execute Phase 2 browser_functional rows through the manifest"
fi
if ! grep -Fq 'phase4 authoritative browser_functional' "$functional_script"; then
  fail "scripts/run-browser-e2e-functional.sh must execute Phase 4 browser_functional rows through the manifest"
fi
if grep -Fq 'e2e/phase4.spec.ts' "$functional_script"; then
  fail "scripts/run-browser-e2e-functional.sh must not raw-select e2e/phase4.spec.ts"
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

if ! grep -Fq '"$ROOT_DIR/scripts/run-browser-e2e-functional.sh"' "$webserver_backed_script"; then
  fail "scripts/run-browser-e2e-webserver-backed.sh must delegate to scripts/run-browser-e2e-functional.sh"
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
