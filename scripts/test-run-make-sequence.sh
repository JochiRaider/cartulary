#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
SCRIPT="${ROOT_DIR}/scripts/run-make-sequence.sh"
cleanup_paths=()

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

json_field() {
  local file="$1"
  local path="$2"

  "${NODE:-node}" -e '
const fs = require("node:fs");
const [file, path] = process.argv.slice(1);
const value = path.split(".").reduce((current, key) => current?.[key], JSON.parse(fs.readFileSync(file, "utf8")));
if (value === undefined || value === null) {
  process.exit(1);
}
process.stdout.write(String(value));
' "${file}" "${path}"
}

assert_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  if [[ "${actual}" != "${expected}" ]]; then
    fail "${label}: expected [${expected}], got [${actual}]"
  fi
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "${haystack}" != *"${needle}"* ]]; then
    fail "${label}: expected output to contain [${needle}]"
  fi
}

write_fake_make() {
  local dir="$1"

  cat >"${dir}/fake-make" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

echo "$*" >>"${FAKE_MAKE_LOG}"

target="${@: -1}"
case "${target}" in
  fail-step)
    exit 7
    ;;
esac

if [[ -n "${CARTULARY_TEST_RESULTS_DIR:-}" && -n "${CARTULARY_TEST_RUN_ID:-}" ]]; then
  mkdir -p "${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${target}"
  cat >"${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${target}/target-summary.json" <<JSON
{
  "target": "${target}",
  "status": "pass",
  "start_time": "2026-01-01T00:00:00Z",
  "end_time": "2026-01-01T00:00:01Z",
  "duration_ms": 1,
  "wall_duration_ms": 1,
  "counts": {
    "phases": 1,
    "tests": 0,
    "failed": 0,
    "authoritative": 0,
    "support": 0,
    "unmapped": 0,
    "non_test": 0,
    "authoritative_failed": 0,
    "support_failed": 0,
    "unmapped_failed": 0,
    "non_test_failed": 0,
    "packages": 0
  }
}
JSON
fi
EOF
  chmod +x "${dir}/fake-make"
}

success_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-success.XXXXXX")"
cleanup_paths+=("${success_dir}")
write_fake_make "${success_dir}"
success_results="${success_dir}/results"
success_output="$(
  MAKE="${success_dir}/fake-make" \
  FAKE_MAKE_LOG="${success_dir}/make.log" \
  CARTULARY_TEST_RESULTS_DIR="${success_results}" \
  CARTULARY_TEST_RUN_ID="success" \
    "${SCRIPT}" --label smoke --summary-targets alpha,beta --step alpha --parallel-step beta:3 \
    2>&1
)"
assert_contains "${success_output}" "[PASS] smoke" "success run summary output"
success_summary="${success_results}/success/run-summary.json"
assert_equals "$(json_field "${success_summary}" "status")" "pass" "success status"
assert_equals "$(json_field "${success_summary}" "completed_targets")" "2/2" "success completed"
assert_equals "$(json_field "${success_summary}" "targets.0")" "alpha" "success target 0"
assert_equals "$(json_field "${success_summary}" "targets.1")" "beta" "success target 1"
assert_contains "$(cat "${success_dir}/make.log")" "--output-sync=target -j3 beta" "parallel make invocation"

failure_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-failure.XXXXXX")"
cleanup_paths+=("${failure_dir}")
write_fake_make "${failure_dir}"
failure_results="${failure_dir}/results"
set +e
failure_output="$(
  MAKE="${failure_dir}/fake-make" \
  FAKE_MAKE_LOG="${failure_dir}/make.log" \
  CARTULARY_TEST_RESULTS_DIR="${failure_results}" \
  CARTULARY_TEST_RUN_ID="failure" \
    "${SCRIPT}" --label fail-smoke --summary-targets alpha,beta --step alpha --step fail-step --step beta \
    2>&1
)"
failure_status=$?
set -e
assert_equals "${failure_status}" "7" "failure child exit status"
assert_contains "${failure_output}" "[FAIL] fail-smoke" "failure run summary output"
failure_summary="${failure_results}/failure/run-summary.json"
assert_equals "$(json_field "${failure_summary}" "status")" "fail" "failure status"
assert_equals "$(json_field "${failure_summary}" "completed_targets")" "1/3" "failure completed"
assert_equals "$(json_field "${failure_summary}" "aborted_after")" "fail-step" "failure aborted_after"
assert_equals "$(json_field "${failure_summary}" "counts.non_test_failed")" "1" "failure non-test count"

dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-dry-run.XXXXXX")"
cleanup_paths+=("${dry_run_dir}")
write_fake_make "${dry_run_dir}"
MAKEFLAGS="n" \
MAKE="${dry_run_dir}/fake-make" \
FAKE_MAKE_LOG="${dry_run_dir}/make.log" \
CARTULARY_TEST_RESULTS_DIR="${dry_run_dir}/results" \
CARTULARY_TEST_RUN_ID="dry-run" \
  "${SCRIPT}" --label dry-run --summary-targets alpha --step alpha
if [[ -e "${dry_run_dir}/results/dry-run/run-summary.json" ]]; then
  fail "dry-run: expected no run-summary JSON"
fi
