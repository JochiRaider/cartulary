#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
cleanup_paths=()

cd "$ROOT_DIR"

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -rf "${path}"
  done
}

trap cleanup EXIT

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

assert_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  if [[ "$actual" != "$expected" ]]; then
    fail "$label: expected [$expected], got [$actual]"
  fi
}

make_target_block() {
  local target="$1"

  awk -v target="$target" '
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
  ' "$ROOT_DIR/Makefile"
}

assert_make_fails() {
  local label="$1"
  shift

  local output
  local status
  set +e
  output="$(make --no-print-directory "$@" 2>&1)"
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
  if ! output="$(make --no-print-directory "$@" 2>&1)"; then
    fail "$label: expected make to pass, got output: $output"
  fi

  printf '%s' "$output"
}

makefile_content="$(cat "$ROOT_DIR/Makefile")"
release_check_block="$(make_target_block release-check)"
license_report_block="$(make_target_block license-report)"
sbom_block="$(make_target_block sbom)"
help_output="$(make --no-print-directory help)"

assert_contains "$makefile_content" ".PHONY: test-fast" "release phony target group"
assert_contains "$makefile_content" " release-check license-report sbom" "release phony targets"
assert_contains "$release_check_block" "release-check: check license-report sbom build" "release-check dependencies"
assert_contains "$license_report_block" './scripts/check-release-artifact.sh "license report" "$(LICENSE_REPORT_ARTIFACT)"' "license-report validation command"
assert_contains "$sbom_block" './scripts/check-release-artifact.sh "SBOM" "$(SBOM_ARTIFACT)"' "sbom validation command"
assert_contains "$help_output" "make release-check" "help release-check documentation"

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

assert_equals "$(make_target_block release-check | grep -c '^release-check:')" "1" "release-check target declaration"
