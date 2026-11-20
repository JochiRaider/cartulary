#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
# shellcheck source=tools/harness/test-support/harness-scratch.sh
source "$ROOT_DIR/tools/harness/test-support/harness-scratch.sh"
scratch="$(cartulary_harness_mktemp_dir browser-preparation.XXXXXX)"
trap 'rm -rf -- "$scratch"' EXIT
# Sourcing loads the adapter without launching a browser or services.
# shellcheck source=tools/harness/browser/start-web-e2e.sh
source "$ROOT_DIR/tools/harness/browser/start-web-e2e.sh"

# shellcheck disable=SC2329
assert_port_free() { return 0; }
# shellcheck disable=SC2329
using_test_services_stack() { return 0; }
# shellcheck disable=SC2329
require_test_services_bin() { return 0; }
# shellcheck disable=SC2329
write_startup_diagnostics() { printf '%s %s %s %s\n' "$1" "$2" "$3" "$4" >"$scratch/diagnostic"; }
# shellcheck disable=SC2329
snapshot_service_scope() { touch "$scratch/incorrect-admission"; }
# shellcheck disable=SC2329
record_startup_event() { touch "$scratch/incorrect-ready"; }

TEST_SERVICES_BIN="$scratch/preparation"
TEST_SERVICES_ENV_FILE="$scratch/unpublished.env"
TEST_SERVICES_METADATA_FILE="$scratch/unpublished.json"
export CARTULARY_PGTEST_TEMPLATE_DB=template
printf 'touch "%s"\n' "$scratch/incorrect-env-read" >"$TEST_SERVICES_ENV_FILE"

for scenario in '37 37 infra service_start_error' '13 13 timing timeout_failure' '2 2 config configuration_error' '0 1 harness fixture_error'; do
  read -r child_status expected_status expected_class expected_reason <<<"$scenario"
  printf '#!/bin/sh\nexit %s\n' "$child_status" >"$TEST_SERVICES_BIN"
  chmod 700 "$TEST_SERVICES_BIN"
  actual=0
  # Conditional execution proves the adapter fails closed independently of errexit.
  browser_prepare_database || actual=$?
  [[ "$actual" == "$expected_status" ]] || { echo "preparation status changed" >&2; exit 1; }
  [[ "$(cat "$scratch/diagnostic")" == "fail database_preparation $expected_class $expected_reason" ]] || { echo "preparation failure classification changed" >&2; exit 1; }
  [[ ! -e "$scratch/incorrect-env-read" && ! -e "$scratch/incorrect-admission" && ! -e "$scratch/incorrect-ready" ]] || { echo "failed preparation admitted dependent state" >&2; exit 1; }
done
# TH-HARNESS-REQ-861: every backend generation keeps one bounded, redacted log
# labelled with its own generation and the cohort that ran against it.
TARGET_ARTIFACT_DIR="$scratch/results/run/_shared/test-services/suite/browser-sessions/browser-allocation-001"
logs="$scratch/private/session/logs"
mkdir -p "$TARGET_ARTIFACT_DIR" "$logs"
chmod 700 "$logs"
SERVER_LOG="$logs/server.log"
WEB_LOG="$logs/web.log"
install -m 600 /dev/null "$SERVER_LOG"
install -m 600 /dev/null "$WEB_LOG"
mark_backend_generation 1 cohort-a
printf 'generation one Authorization: Bearer sentinel-credential\n' >>"$SERVER_LOG"
printf 'frontend ready\n' >>"$WEB_LOG"
retain_browser_process_diagnostics
mark_backend_generation 2 cohort-b
[[ ! -s "$SERVER_LOG" ]] || { echo "marked generation retained predecessor output" >&2; exit 1; }
printf 'generation two\n' >>"$SERVER_LOG"
# Final settlement uses the same retention path; repetition never replaces evidence.
"$NODE_BIN" "$ROOT_DIR/tools/harness/browser/server-log-retention.mjs" retain "$logs"
retain_browser_process_diagnostics
diagnostics="$TARGET_ARTIFACT_DIR/process-diagnostics"
[[ "$(ls "$diagnostics")" == $'generation-001-cohort-a\ngeneration-002-cohort-b' ]] || { echo "server generations are not labelled by their own cohort" >&2; exit 1; }
grep -q 'generation one' "$diagnostics/generation-001-cohort-a/server.log" || { echo "replaced generation log missing" >&2; exit 1; }
grep -q 'generation two' "$diagnostics/generation-002-cohort-b/server.log" || { echo "final generation log missing" >&2; exit 1; }
! grep -rq 'sentinel-credential\|generation one' "$diagnostics/generation-002-cohort-b" || { echo "retained log crossed generations" >&2; exit 1; }
! grep -rq 'sentinel-credential' "$diagnostics" || { echo "retained log is unredacted" >&2; exit 1; }
grep -q 'frontend ready' "$diagnostics/generation-002-cohort-b/web.log" || { echo "frontend log missing" >&2; exit 1; }
[[ -z "$(find "$diagnostics" \( -type d ! -perm 700 \) -o \( -type f ! -perm 600 \))" ]] || { echo "retained logs are not owner-only" >&2; exit 1; }
if "$NODE_BIN" "$ROOT_DIR/tools/harness/browser/server-log-retention.mjs" mark "$logs" 3 cohort-c "$scratch/elsewhere" 2>/dev/null; then
  echo "server generation marker accepted a foreign artifact directory" >&2; exit 1
fi
printf '%s\n' 'browser preparation contract: pass'
