#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
cd "$ROOT_DIR"

target="incident-bundle-v1-retirement-attestation-check"
fixture="tools/release-evidence/fixtures/incident-bundle-v1-retirement-attestation.valid.json"
scratch="$(mktemp -d "${TMPDIR:-/tmp}/cartulary-retirement-command-surface.XXXXXX")"
trap 'rm -rf "$scratch"' EXIT

fail() {
  printf 'incident bundle retirement command-surface check failed: %s\n' "$1" >&2
  exit 1
}

assert_contains() {
  local value="$1"
  local expected="$2"
  local label="$3"
  [[ "$value" == *"$expected"* ]] || fail "$label"
}

assert_not_contains() {
  local value="$1"
  local unexpected="$2"
  local label="$3"
  [[ "$value" != *"$unexpected"* ]] || fail "$label"
}

run_child_make() {
  env \
    -u CARTULARY_HARNESS_IDENTITY_PREPARED \
    -u CARTULARY_MAKE_INPUT_SOURCES \
    -u CARTULARY_STEP_ARTIFACT_DIR \
    -u CARTULARY_SUPPRESS_CHILD_SUCCESS \
    -u CARTULARY_TEST_RUN_ID \
    -u CARTULARY_TEST_TARGET \
    CARTULARY_TEST_RESULTS_DIR="$scratch/results" \
    make --no-print-directory "$@"
}

set +e
missing_output="$(run_child_make "$target" 2>&1)"
missing_status=$?
set -e
[[ "$missing_status" -ne 0 ]] || fail "missing ATTESTATION unexpectedly passed"
assert_contains "$missing_output" "failure_reason=usage_error" "missing ATTESTATION classification"

set +e
environment_output="$(ATTESTATION="$fixture" run_child_make "$target" 2>&1)"
environment_status=$?
set -e
[[ "$environment_status" -ne 0 ]] || fail "environment-only ATTESTATION unexpectedly passed"
assert_contains "$environment_output" "failure_reason=usage_error" "environment-only source classification"
assert_not_contains "$environment_output" "$fixture" "environment-only path disclosure"

injection_marker="$scratch/make-expansion-marker"
injection_value="\$(shell touch $injection_marker)"
set +e
injection_output="$(run_child_make "$target" "ATTESTATION=$injection_value" 2>&1)"
injection_status=$?
set -e
[[ "$injection_status" -ne 0 ]] || fail "Make-expansion-shaped ATTESTATION unexpectedly passed"
[[ ! -e "$injection_marker" ]] || fail "ATTESTATION triggered Make or shell evaluation"
assert_contains "$injection_output" "failure_reason=artifact_error" "literal input failure classification"
assert_not_contains "$injection_output" "$injection_marker" "literal input path disclosure"

set +e
command_output="$(run_child_make "$target" ATTESTATION="$fixture" 2>&1)"
command_status=$?
set -e
[[ "$command_status" -ne 0 ]] || fail "synthetic operational attestation unexpectedly passed"
assert_contains "$command_output" "failure_reason=artifact_error" "invalid operational evidence classification"
assert_not_contains "$command_output" "$fixture" "stdout or stderr path disclosure"

if rg -F --quiet "$fixture" "$scratch/results"; then
  fail "retained artifact disclosed ATTESTATION path"
fi

help_output="$(run_child_make help-all)"
assert_contains "$help_output" "make $target" "help-all target registration"
assert_contains "$help_output" "ATTESTATION=<path>" "help-all target usage"

target_count="$(rg -c "^${target}:$" tools/task_surface.generated.mk)"
[[ "$target_count" == "1" ]] || fail "generated Make target declaration count"

node -e '
  const manifest = require("./tools/task_surface_manifest.json");
  const target = manifest.targets.find((entry) => entry.name === process.argv[1]);
  if (!target || target.command_id !== process.argv[2]) process.exit(1);
  const input = target.input_contract?.inputs?.find((entry) => entry.name === "ATTESTATION");
  if (
    !input ||
    input.required !== true ||
    input.summary_emission !== "redacted_value" ||
    input.child_forwarding !== "argv"
  ) process.exit(1);
' "$target" "cartulary.harness.command.incident_bundle_v1_retirement_attestation_check.v1" || \
  fail "generated task-surface manifest contract"

printf 'incident bundle v1 retirement attestation command-surface checks passed\n'
