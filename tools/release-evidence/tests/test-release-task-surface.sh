#!/usr/bin/env bash
# Single-quoted literals below intentionally assert Make/shell text without expansion.
# shellcheck disable=SC2016
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
if command -v "${NODE_BIN}" >/dev/null 2>&1; then
  NODE_BIN="$(command -v "${NODE_BIN}")"
fi
task_surface_makefile="$ROOT_DIR/Makefile"
task_surface_generated_make_file="$ROOT_DIR/tools/task_surface.generated.mk"
# shellcheck source=tools/harness/test-support/harness-scratch.sh
source "$ROOT_DIR/tools/harness/test-support/harness-scratch.sh"
# shellcheck source=tools/harness/test-support/task-surface-check-common.sh
source "$ROOT_DIR/tools/harness/test-support/task-surface-check-common.sh"
cleanup_paths=()

cd "$ROOT_DIR"

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -rf "${path}"
  done
}

trap cleanup EXIT

probe_results_root="$(cartulary_harness_mktemp_dir "release-task-surface-results.XXXXXX")"
cleanup_paths+=("$probe_results_root")

fail() {
  echo "$*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label: expected output to contain [$needle]"
  fi
}

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: expected output not to contain [$needle]"
  fi
}

assert_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  if [[ "$actual" != "$expected" ]]; then
    fail "$label: expected [$expected], got [$actual]"
  fi
}

assert_file_absent() {
  local file="$1"
  local label="$2"

  if [[ -e "$file" ]]; then
    fail "$label: expected file to be absent: $file"
  fi
}

assert_files_equal() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  if ! cmp -s "$actual" "$expected"; then
    fail "$label: expected file bytes to remain unchanged: $actual"
  fi
}

normalize_make_continuations() {
  local text="$1"

  printf '%s\n' "$text" | sed -e ':join' -e '/\\$/ { N; s/[[:space:]]*\\\n[[:space:]]*/ /; b join; }'
}

probe_run_id() {
  local label="$1"
  local slug

  slug="$(printf '%s' "$label" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//; s/-+/-/g')"
  if [[ -z "$slug" ]]; then
    slug="probe"
  fi
  printf 'release-task-surface-%s\n' "$slug"
}

ambient_summary_path() {
  local target="$1"
  local results_root="${CARTULARY_TEST_RESULTS_DIR:-}"

  if [[ -z "$results_root" || -z "${CARTULARY_TEST_RUN_ID:-}" ]]; then
    return 1
  fi
  if [[ "$results_root" != /* ]]; then
    results_root="$ROOT_DIR/$results_root"
  fi
  printf '%s/%s/%s/target-summary.json\n' "$results_root" "$CARTULARY_TEST_RUN_ID" "$target"
}

assert_no_ambient_summary() {
  local target="$1"
  local summary_path

  if ! summary_path="$(ambient_summary_path "$target")"; then
    return 0
  fi
  assert_file_absent "$summary_path" "ambient $target summary"
}

run_isolated_make_probe() {
  local label="$1"
  shift

  local run_id
  run_id="$(probe_run_id "$label")"
  env -u CARTULARY_HARNESS_IDENTITY_PREPARED -u CARTULARY_TEST_TARGET \
    CARTULARY_OUTPUT_MODE=verbose \
    CARTULARY_TEST_RESULTS_DIR="$probe_results_root" \
    CARTULARY_TEST_RUN_ID="$run_id" \
    make --no-print-directory "$@"
}

assert_make_fails() {
  local label="$1"
  shift

  local output
  local status
  set +e
  output="$(run_isolated_make_probe "$label" "$@" 2>&1)"
  status=$?
  set -e

  if [[ "$status" -eq 0 ]]; then
    fail "$label: expected make to fail"
  fi

  printf '%s' "$output"
}

assert_make_passes() {
  local label="$1"
  shift

  local output
  if ! output="$(run_isolated_make_probe "$label" "$@" 2>&1)"; then
    fail "$label: expected make to pass, got output: $output"
  fi

  printf '%s' "$output"
}

makefile_content="$(cat "$ROOT_DIR/Makefile"; printf '\n'; cat "$ROOT_DIR/tools/task_surface.generated.mk")"
release_check_block="$(extract_target_definition release-check)"
release_check_logical_block="$(normalize_make_continuations "$release_check_block")"
release_readiness_block="$(extract_target_definition release-readiness-evidence)"
license_report_block="$(extract_target_definition license-report)"
sbom_block="$(extract_target_definition sbom)"
release_inventory_block="$(extract_target_definition release-inventory-artifacts)"
help_output="$(env -u CARTULARY_HARNESS_IDENTITY_PREPARED -u CARTULARY_TEST_RESULTS_DIR -u CARTULARY_TEST_RUN_ID -u CARTULARY_TEST_TARGET make --no-print-directory help)"
help_all_output="$(env -u CARTULARY_HARNESS_IDENTITY_PREPARED -u CARTULARY_TEST_RESULTS_DIR -u CARTULARY_TEST_RUN_ID -u CARTULARY_TEST_TARGET make --no-print-directory help-all)"
release_check_explain="$(env -u CARTULARY_HARNESS_IDENTITY_PREPARED -u CARTULARY_TEST_RESULTS_DIR -u CARTULARY_TEST_RUN_ID -u CARTULARY_TEST_TARGET make --no-print-directory explain-target TARGET=release-check DETAIL=summary)"

for public_target in test-fast release-check release-readiness-evidence license-report sbom; do
  assert_contains "$makefile_content" "$public_target:" "release task surface target $public_target"
done
assert_contains "$release_check_logical_block" 'work-graph/runner-cli.mjs --selection aggregate --target release-check' "release-check graph runner"
assert_contains "$release_readiness_block" './tools/release-evidence/release-readiness-evidence.mjs' "release readiness evidence command"
assert_contains "$makefile_content" '$(SBOM_ARTIFACT) $(LICENSE_REPORT_ARTIFACT):' "SBOM/license artifact generation rule"
assert_contains "$makefile_content" './tools/release-evidence/generate-sbom-license-evidence.mjs' "SBOM/license generator command"
assert_contains "$license_report_block" 'license-report: release-inventory-artifacts' "license-report producer prerequisite"
assert_contains "$license_report_block" 'ifeq ($(CARTULARY_HARNESS_GRAPH_CHILD),1)' "license-report graph-child prerequisite cutover"
assert_contains "$license_report_block" './tools/release-evidence/check-release-artifact.sh "license report" "$(LICENSE_REPORT_ARTIFACT)"' "license-report validation command"
assert_contains "$sbom_block" 'sbom: release-inventory-artifacts' "sbom producer prerequisite"
assert_contains "$sbom_block" 'ifeq ($(CARTULARY_HARNESS_GRAPH_CHILD),1)' "sbom graph-child prerequisite cutover"
assert_contains "$sbom_block" './tools/release-evidence/check-release-artifact.sh "SBOM" "$(SBOM_ARTIFACT)"' "sbom validation command"
assert_contains "$release_inventory_block" 'work-graph/runner-cli.mjs --selection target --target release-inventory-artifacts' "release inventory graph producer"
assert_not_contains "$help_output" "make release-check" "compact help omits release-check documentation"
assert_contains "$help_all_output" "target -> owner partition -> semantic row -> artifact" "help-all concept hierarchy"
assert_contains "$help_all_output" "make release-check" "help-all release-check documentation"
assert_contains "$help_all_output" "make release-readiness-evidence" "help-all release readiness documentation"
assert_contains "$help_all_output" "extended harness" "help-all release-check extended harness documentation"
assert_contains "$release_check_explain" "browser-e2e-support" "release-check explains support readiness child"
assert_contains "$release_check_explain" "browser-e2e-visual" "release-check explains visual readiness child"
assert_contains "$release_check_explain" "browser-e2e-a11y" "release-check explains accessibility readiness child"
assert_contains "$release_check_explain" "release-readiness-evidence" "release-check explains release readiness aggregation child"

tmp_dir="$(cartulary_harness_mktemp_dir "release-task-surface-artifacts.XXXXXX")"
cleanup_paths+=("$tmp_dir")
empty_license="$tmp_dir/empty-license.json"
valid_license="$tmp_dir/license-report.json"
legacy_license="$tmp_dir/license-report-v1.json"
empty_sbom="$tmp_dir/empty-sbom.cyclonedx.json"
valid_sbom="$tmp_dir/sbom.cyclonedx.json"

touch "$empty_license"
cp "$empty_license" "$empty_license.before"
empty_license_output="$(assert_make_fails "empty license report" CARTULARY_HARNESS_GRAPH_CHILD=1 --old-file="$empty_license" NODE_BIN="$NODE_BIN" LICENSE_REPORT_ARTIFACT="$empty_license" license-report)"
assert_contains "$empty_license_output" "license report artifact is empty" "empty license report failure"
assert_files_equal "$empty_license" "$empty_license.before" "empty license report probe"

printf '%s\n' '{"schema_id":"cartulary.license_report.v2","semantic_input_digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","entries":[]}' >"$valid_license"
cp "$valid_license" "$valid_license.before"
assert_make_passes "valid license report" CARTULARY_HARNESS_GRAPH_CHILD=1 --old-file="$valid_license" NODE_BIN="$NODE_BIN" LICENSE_REPORT_ARTIFACT="$valid_license" license-report >/dev/null
assert_files_equal "$valid_license" "$valid_license.before" "valid license report probe"

printf '%s\n' '{"schema_id":"cartulary.license_report.v1","entries":[]}' >"$legacy_license"
legacy_license_output="$(assert_make_fails "legacy license report" CARTULARY_HARNESS_GRAPH_CHILD=1 --old-file="$legacy_license" NODE_BIN="$NODE_BIN" LICENSE_REPORT_ARTIFACT="$legacy_license" license-report)"
assert_contains "$legacy_license_output" "unsupported license report schema cartulary.license_report.v1" "legacy license report hard cutover"

graph_missing_license="$tmp_dir/graph-child-missing-license.json"
graph_missing_output="$(assert_make_fails "graph child missing license" CARTULARY_HARNESS_GRAPH_CHILD=1 NODE_BIN="$NODE_BIN" LICENSE_REPORT_ARTIFACT="$graph_missing_license" license-report)"
assert_contains "$graph_missing_output" "license report artifact missing" "graph child validates without regeneration"
assert_file_absent "$graph_missing_license" "graph child producer bypass"

touch "$empty_sbom"
cp "$empty_sbom" "$empty_sbom.before"
empty_sbom_output="$(assert_make_fails "empty SBOM" CARTULARY_HARNESS_GRAPH_CHILD=1 --old-file="$empty_sbom" NODE_BIN="$NODE_BIN" SBOM_ARTIFACT="$empty_sbom" sbom)"
assert_contains "$empty_sbom_output" "SBOM artifact is empty" "empty SBOM failure"
assert_files_equal "$empty_sbom" "$empty_sbom.before" "empty SBOM probe"

printf '%s\n' '{"bomFormat":"CycloneDX","specVersion":"1.7","serialNumber":"urn:uuid:aaaaaaaa-aaaa-5aaa-8aaa-aaaaaaaaaaaa","version":1,"metadata":{"component":{"type":"application","name":"fixture","properties":[{"name":"cartulary:semantic_input_digest","value":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}]}}}' >"$valid_sbom"
cp "$valid_sbom" "$valid_sbom.before"
assert_make_passes "valid SBOM" CARTULARY_HARNESS_GRAPH_CHILD=1 --old-file="$valid_sbom" NODE_BIN="$NODE_BIN" SBOM_ARTIFACT="$valid_sbom" sbom >/dev/null
assert_files_equal "$valid_sbom" "$valid_sbom.before" "valid SBOM probe"

assert_no_ambient_summary "license-report"
assert_no_ambient_summary "sbom"

assert_equals "$(extract_target_definition release-check | grep -c '^release-check:[[:space:]]*$')" "1" "release-check target declaration"
