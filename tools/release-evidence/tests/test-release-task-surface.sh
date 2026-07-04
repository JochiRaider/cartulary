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

assert_probe_artifact_contains() {
  local probe_label="$1"
  local target="$2"
  local phase_label="$3"
  local needle="$4"
  local assertion_label="$5"

  "$NODE_BIN" "$ROOT_DIR/tools/harness/test-support/harness-artifact-assert.mjs" \
    --repo-root "$ROOT_DIR" \
    --results-root "$probe_results_root" \
    --run-id "$(probe_run_id "$probe_label")" \
    --target "$target" \
    --phase-label "$phase_label" \
    --needle "$needle" \
    --label "$assertion_label"
}

run_isolated_make_probe() {
  local label="$1"
  shift

  local run_id
  run_id="$(probe_run_id "$label")"
  env -u CARTULARY_TEST_TARGET \
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
release_readiness_block="$(extract_target_definition release-readiness-evidence)"
license_report_block="$(extract_target_definition license-report)"
sbom_block="$(extract_target_definition sbom)"
help_output="$(make --no-print-directory help)"
help_all_output="$(make --no-print-directory help-all)"
release_check_explain="$(make --no-print-directory explain-target TARGET=release-check DETAIL=summary)"

assert_contains "$makefile_content" " test-fast " "release phony target group"
assert_contains "$makefile_content" " release-check release-readiness-evidence license-report sbom" "release phony targets"
assert_contains "$release_check_block" '$(RUN_MAKE_SEQUENCE_SCRIPT) --sequence release-check' "release-check sequence runner"
assert_contains "$release_readiness_block" './tools/release-evidence/release-readiness-evidence.mjs' "release readiness evidence command"
assert_contains "$makefile_content" '$(SBOM_ARTIFACT) $(LICENSE_REPORT_ARTIFACT):' "SBOM/license artifact generation rule"
assert_contains "$makefile_content" './tools/release-evidence/generate-sbom-license-evidence.mjs' "SBOM/license generator command"
assert_contains "$license_report_block" 'license-report: $(LICENSE_REPORT_ARTIFACT)' "license-report generation prerequisite"
assert_contains "$license_report_block" './tools/release-evidence/check-release-artifact.sh "license report" "$(LICENSE_REPORT_ARTIFACT)"' "license-report validation command"
assert_contains "$sbom_block" 'sbom: $(SBOM_ARTIFACT)' "sbom generation prerequisite"
assert_contains "$sbom_block" './tools/release-evidence/check-release-artifact.sh "SBOM" "$(SBOM_ARTIFACT)"' "sbom validation command"
assert_not_contains "$help_output" "make release-check" "compact help omits release-check documentation"
assert_contains "$help_all_output" "phase -> target -> scheduler work unit -> artifact" "help-all concept hierarchy"
assert_contains "$help_all_output" "make release-check" "help-all release-check documentation"
assert_contains "$help_all_output" "make release-readiness-evidence" "help-all release readiness documentation"
assert_contains "$help_all_output" "extended harness" "help-all release-check extended harness documentation"
assert_contains "$release_check_explain" "browser-e2e-support" "release-check explains support readiness child"
assert_contains "$release_check_explain" "browser-e2e-visual" "release-check explains visual readiness child"
assert_contains "$release_check_explain" "browser-e2e-a11y" "release-check explains accessibility readiness child"
assert_contains "$release_check_explain" "release-readiness-evidence" "release-check explains release readiness aggregation child"
assert_contains "$release_check_explain" "frontend-readiness" "release-check explains frontend readiness group"
assert_contains "$release_check_explain" "release-readiness" "release-check explains release readiness group"

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/release-task-surface.XXXXXX")"
cleanup_paths+=("$tmp_dir")
empty_license="$tmp_dir/empty-license.json"
valid_license="$tmp_dir/license-report.json"
empty_sbom="$tmp_dir/empty-sbom.cyclonedx.json"
valid_sbom="$tmp_dir/sbom.cyclonedx.json"
fake_cyclonedx_gomod="$tmp_dir/cyclonedx-gomod"
fake_syft="$tmp_dir/syft"
frontend_install_stamp="$tmp_dir/frontend-install.stamp"

printf '#!/usr/bin/env bash\nexit 0\n' >"$fake_cyclonedx_gomod"
printf '#!/usr/bin/env bash\nexit 0\n' >"$fake_syft"
chmod +x "$fake_cyclonedx_gomod" "$fake_syft"
touch "$frontend_install_stamp"
release_probe_prereqs=(
  --old-file="$NODE_BIN"
  --old-file="$frontend_install_stamp"
  --old-file="$fake_cyclonedx_gomod"
  --old-file="$fake_syft"
  NODE_BIN="$NODE_BIN"
  FRONTEND_INSTALL_STAMP="$frontend_install_stamp"
  CYCLONEDX_GOMOD_BIN="$fake_cyclonedx_gomod"
  SYFT_BIN="$fake_syft"
)

touch "$empty_license"
assert_make_fails "empty license report" "${release_probe_prereqs[@]}" LICENSE_REPORT_ARTIFACT="$empty_license" license-report >/dev/null
assert_probe_artifact_contains "empty license report" "license-report" "license-report" "license report artifact is empty" "empty license report failure"

printf '%s\n' '{"licenses":[]}' >"$valid_license"
assert_make_passes "valid license report" "${release_probe_prereqs[@]}" LICENSE_REPORT_ARTIFACT="$valid_license" license-report >/dev/null

touch "$empty_sbom"
assert_make_fails "empty SBOM" "${release_probe_prereqs[@]}" SBOM_ARTIFACT="$empty_sbom" sbom >/dev/null
assert_probe_artifact_contains "empty SBOM" "sbom" "sbom" "SBOM artifact is empty" "empty SBOM failure"

printf '%s\n' '{"bomFormat":"CycloneDX"}' >"$valid_sbom"
assert_make_passes "valid SBOM" "${release_probe_prereqs[@]}" SBOM_ARTIFACT="$valid_sbom" sbom >/dev/null

assert_no_ambient_summary "license-report"
assert_no_ambient_summary "sbom"

assert_equals "$(extract_target_definition release-check | grep -c '^release-check:')" "1" "release-check target declaration"
