#!/usr/bin/env bash
# Single-quoted literals below intentionally assert Make/shell text without expansion.
# shellcheck disable=SC2016
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
# shellcheck source=scripts/lib/harness-scratch.sh
source "$ROOT_DIR/scripts/lib/harness-scratch.sh"
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

make_target_block() {
  local target="$1"

  cat "$ROOT_DIR/tools/task_surface.generated.mk" "$ROOT_DIR/Makefile" | awk -v target="$target" '
    $0 ~ "^" target ":" {
      in_target = 1
      print
      next
    }
    in_target && /^[^[:space:]#][^:]*:/ {
      exit
    }
    in_target {
      print
    }
  '
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
release_check_block="$(make_target_block release-check)"
license_report_block="$(make_target_block license-report)"
sbom_block="$(make_target_block sbom)"
help_output="$(make --no-print-directory help)"
help_all_output="$(make --no-print-directory help-all)"

assert_contains "$makefile_content" " test-fast " "release phony target group"
assert_contains "$makefile_content" " release-check license-report sbom" "release phony targets"
assert_contains "$release_check_block" '$(RUN_MAKE_SEQUENCE_SCRIPT) --sequence release-check' "release-check sequence runner"
assert_contains "$license_report_block" './scripts/check-release-artifact.sh "license report" "$(LICENSE_REPORT_ARTIFACT)"' "license-report validation command"
assert_contains "$sbom_block" './scripts/check-release-artifact.sh "SBOM" "$(SBOM_ARTIFACT)"' "sbom validation command"
assert_not_contains "$help_output" "make release-check" "compact help omits release-check documentation"
assert_contains "$help_all_output" "make release-check" "help-all release-check documentation"
assert_contains "$help_all_output" "extended harness" "help-all release-check extended harness documentation"

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/release-task-surface.XXXXXX")"
cleanup_paths+=("$tmp_dir")
missing_license="$tmp_dir/missing-license.json"
empty_license="$tmp_dir/empty-license.json"
valid_license="$tmp_dir/license-report.json"
missing_sbom="$tmp_dir/missing-sbom.cdx.json"
empty_sbom="$tmp_dir/empty-sbom.cdx.json"
valid_sbom="$tmp_dir/sbom.cdx.json"

license_missing_output="$(assert_make_fails "missing license report" LICENSE_REPORT_ARTIFACT="$missing_license" license-report)"
assert_contains "$license_missing_output" "license report artifact missing" "missing license report failure"

touch "$empty_license"
license_empty_output="$(assert_make_fails "empty license report" LICENSE_REPORT_ARTIFACT="$empty_license" license-report)"
assert_contains "$license_empty_output" "license report artifact is empty" "empty license report failure"

printf '%s\n' '{"licenses":[]}' >"$valid_license"
assert_make_passes "valid license report" LICENSE_REPORT_ARTIFACT="$valid_license" license-report >/dev/null

sbom_missing_output="$(assert_make_fails "missing SBOM" SBOM_ARTIFACT="$missing_sbom" sbom)"
assert_contains "$sbom_missing_output" "SBOM artifact missing" "missing SBOM failure"

touch "$empty_sbom"
sbom_empty_output="$(assert_make_fails "empty SBOM" SBOM_ARTIFACT="$empty_sbom" sbom)"
assert_contains "$sbom_empty_output" "SBOM artifact is empty" "empty SBOM failure"

printf '%s\n' '{"bomFormat":"CycloneDX"}' >"$valid_sbom"
assert_make_passes "valid SBOM" SBOM_ARTIFACT="$valid_sbom" sbom >/dev/null

assert_no_ambient_summary "license-report"
assert_no_ambient_summary "sbom"

assert_equals "$(make_target_block release-check | grep -c '^release-check:')" "1" "release-check target declaration"
