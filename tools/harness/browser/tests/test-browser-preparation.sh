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
printf '%s\n' 'browser preparation contract: pass'
