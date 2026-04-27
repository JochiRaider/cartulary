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

assert_json_field_absent() {
  local file="$1"
  local path="$2"
  local label="$3"

  if "${NODE:-node}" -e '
const fs = require("node:fs");
const [file, path] = process.argv.slice(1);
const value = path.split(".").reduce((current, key) => current?.[key], JSON.parse(fs.readFileSync(file, "utf8")));
process.exit(value === undefined ? 0 : 1);
' "${file}" "${path}"; then
    return 0
  fi
  fail "${label}: expected JSON field [${path}] to be absent"
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

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "${haystack}" == *"${needle}"* ]]; then
    fail "${label}: expected output not to contain [${needle}]"
  fi
}

assert_file_absent() {
  local path="$1"
  local label="$2"

  if [[ -e "${path}" ]]; then
    fail "${label}: expected ${path} to be absent"
  fi
}

assert_count() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  assert_equals "${actual}" "${expected}" "${label}"
}

make_target_block() {
  local target="$1"

  awk -v target="${target}" '
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
  ' "${ROOT_DIR}/Makefile"
}

line_count() {
  local pattern="$1"

  grep -Ec "${pattern}" "${ROOT_DIR}/Makefile"
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
  "executed_duration_ms": 1,
  "logical_duration_ms": 1,
  "reused_duration_ms": 0,
  "derived_duration_ms": 0,
  "wall_duration_ms": 1,
  "critical_path_wall_duration_ms": 1,
  "teardown_duration_ms": 0,
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
    "${SCRIPT}" --label smoke --summary-targets " alpha, beta " --summary-groups "alpha-group=alpha;beta-group=beta" --step alpha --parallel-step beta:3 \
    2>&1
)"
assert_contains "${success_output}" "[RUN] smoke steps=2 targets=2 jobs=3 run_id=success" "success run start output"
assert_contains "${success_output}" "[STEP] smoke 1/2 alpha mode=serial jobs=1" "success serial step output"
assert_contains "${success_output}" "[STEP] smoke 2/2 beta mode=parallel jobs=3" "success parallel step output"
assert_contains "${success_output}" "[PASS] smoke" "success run summary output"
assert_contains "${success_output}" "[GROUP] smoke alpha-group targets=alpha status=pass" "success alpha group output"
assert_contains "${success_output}" "[GROUP] smoke beta-group targets=beta status=pass" "success beta group output"
success_summary="${success_results}/success/run-summary.json"
assert_equals "$(json_field "${success_summary}" "status")" "pass" "success status"
assert_equals "$(json_field "${success_summary}" "completed_targets")" "2/2" "success completed"
assert_equals "$(json_field "${success_summary}" "targets.0")" "alpha" "success target 0"
assert_equals "$(json_field "${success_summary}" "targets.1")" "beta" "success target 1"
assert_equals "$(json_field "${success_summary}" "summary_groups.0.name")" "alpha-group" "success group 0"
assert_equals "$(json_field "${success_summary}" "summary_groups.0.targets.0")" "alpha" "success group target 0"
assert_equals "$(json_field "${success_summary}" "summary_groups.0.wall_duration_ms")" "1000" "success group wall duration"
assert_equals "$(json_field "${success_summary}" "summary_groups.0.critical_path_wall_duration_ms")" "1000" "success group critical path duration"
assert_equals "$(json_field "${success_summary}" "summary_groups.0.teardown_duration_ms")" "0" "success group teardown duration"
assert_json_field_absent "${success_summary}" "duration_ms" "success legacy run duration"
assert_json_field_absent "${success_summary}" "summary_groups.0.duration_ms" "success legacy group duration"
assert_contains "$(cat "${success_dir}/make.log")" "--output-sync=target -j3 beta" "parallel make invocation"

aggregate_missing_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-aggregate-missing.XXXXXX")"
cleanup_paths+=("${aggregate_missing_dir}")
write_fake_make "${aggregate_missing_dir}"
aggregate_missing_results="${aggregate_missing_dir}/results"
set +e
aggregate_missing_output="$(
  MAKE="${aggregate_missing_dir}/fake-make" \
  FAKE_MAKE_LOG="${aggregate_missing_dir}/make.log" \
  CARTULARY_TEST_RESULTS_DIR="${aggregate_missing_results}" \
  CARTULARY_TEST_RUN_ID="aggregate-missing" \
    "${SCRIPT}" --label aggregate-missing --summary-targets alpha,missing-target --step alpha \
    2>&1
)"
aggregate_missing_status=$?
set -e
assert_equals "${aggregate_missing_status}" "1" "aggregate missing target exit status"
assert_contains "${aggregate_missing_output}" "[RUN] aggregate-missing steps=1 targets=2 jobs=1 run_id=aggregate-missing" "aggregate missing run start output"
assert_contains "${aggregate_missing_output}" "[STEP] aggregate-missing 1/1 alpha mode=serial jobs=1" "aggregate missing step output"
assert_contains "${aggregate_missing_output}" "[FAIL] aggregate-missing" "aggregate missing target run summary output"
aggregate_missing_summary="${aggregate_missing_results}/aggregate-missing/run-summary.json"
assert_equals "$(json_field "${aggregate_missing_summary}" "status")" "fail" "aggregate missing target status"
assert_equals "$(json_field "${aggregate_missing_summary}" "missing_target_summaries.0")" "missing-target" "aggregate missing target list"

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
assert_contains "${failure_output}" "[RUN] fail-smoke steps=3 targets=2 jobs=1 run_id=failure" "failure run start output"
assert_contains "${failure_output}" "[STEP] fail-smoke 1/3 alpha mode=serial jobs=1" "failure alpha step output"
assert_contains "${failure_output}" "[STEP] fail-smoke 2/3 fail-step mode=serial jobs=1" "failure failing step output"
assert_contains "${failure_output}" "[FAIL] fail-smoke" "failure run summary output"
failure_summary="${failure_results}/failure/run-summary.json"
assert_equals "$(json_field "${failure_summary}" "status")" "fail" "failure status"
assert_equals "$(json_field "${failure_summary}" "completed_targets")" "1/3" "failure completed"
assert_equals "$(json_field "${failure_summary}" "aborted_after")" "fail-step" "failure aborted_after"
assert_equals "$(json_field "${failure_summary}" "counts.non_test_failed")" "1" "failure non-test count"

dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-dry-run.XXXXXX")"
cleanup_paths+=("${dry_run_dir}")
write_fake_make "${dry_run_dir}"
dry_run_output="$(
  MAKEFLAGS="n" \
  MAKE="${dry_run_dir}/fake-make" \
  FAKE_MAKE_LOG="${dry_run_dir}/make.log" \
  CARTULARY_TEST_RESULTS_DIR="${dry_run_dir}/results" \
  CARTULARY_TEST_RUN_ID="dry-run" \
    "${SCRIPT}" --label dry-run --summary-targets alpha --step alpha \
    2>&1
)"
assert_not_contains "${dry_run_output}" "[RUN]" "script dry-run run start output"
assert_not_contains "${dry_run_output}" "[STEP]" "script dry-run step output"
assert_file_absent "${dry_run_dir}/results/dry-run/run-summary.json" "script dry-run summary"
assert_contains "$(cat "${dry_run_dir}/make.log")" "--no-print-directory alpha" "script dry-run child make"

invalid_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-invalid.XXXXXX")"
cleanup_paths+=("${invalid_dir}")
write_fake_make "${invalid_dir}"
set +e
invalid_output="$(
  MAKE="${invalid_dir}/fake-make" \
  FAKE_MAKE_LOG="${invalid_dir}/make.log" \
    "${SCRIPT}" --label invalid --summary-targets alpha --parallel-step alpha \
    2>&1
)"
invalid_status=$?
set -e
assert_equals "${invalid_status}" "2" "invalid usage status"
assert_contains "${invalid_output}" "--parallel-step requires <target>:<jobs>" "invalid usage output"
assert_file_absent "${invalid_dir}/make.log" "invalid usage child make log"

makefile_content="$(cat "${ROOT_DIR}/Makefile")"
generated_make="$(cat "${ROOT_DIR}/tools/task_surface.generated.mk")"
generated_phony_line="$(printf '%s\n' "${generated_make}" | sed -n 's/^\\.PHONY: //p')"
manifest_content="$(cat "${ROOT_DIR}/tools/task_surface_manifest.json")"
assert_count "$(line_count '^RUN_MAKE_SEQUENCE_SCRIPT :=')" "1" "run sequence helper declaration"
assert_count "$(line_count '^RUN_HARNESS_SMOKE_SCRIPT :=')" "1" "harness smoke helper declaration"
assert_count "$(line_count '^RUN_SERVICE_BACKED_SCHEDULE_SCRIPT :=')" "1" "service-backed scheduler helper declaration"
assert_count "$(line_count '^RUN_CHECK_SCHEDULE_SCRIPT :=')" "1" "check scheduler helper declaration"
assert_contains "${makefile_content}" "include tools/task_surface.generated.mk" "Makefile includes generated task surface"
assert_not_contains "${generated_make}" "TASK_SURFACE_HARNESS_TIER_" "generated Make harness tier variables"
assert_not_contains "${generated_phony_line}" "harness-smoke-toolchain-pins" "generated Make harness leaf targets"
assert_not_contains "${generated_phony_line}" "run-harness-smoke-fast-all" "generated Make fast harness aggregate leaf"
assert_count "$(line_count '^RUN_MAKE_SEQUENCE_SCRIPT :=')" "1" "run sequence helper declaration"
assert_count "$(line_count '^RUN_HARNESS_SMOKE_SCRIPT :=')" "1" "harness smoke helper declaration"
assert_count "$(line_count '^RUN_SERVICE_BACKED_SCHEDULE_SCRIPT :=')" "1" "service-backed scheduler helper declaration"
assert_count "$(line_count '^RUN_CHECK_SCHEDULE_SCRIPT :=')" "1" "check scheduler helper declaration"

test_block="$(make_target_block test)"
check_block="$(make_target_block check)"
run_harness_smoke_fast_block="$(make_target_block run-harness-smoke-fast)"
run_harness_smoke_extended_block="$(make_target_block run-harness-smoke-extended)"
run_harness_smoke_full_block="$(make_target_block run-harness-smoke-full)"
check_harness_smoke_block="$(make_target_block check-harness-smoke)"
test_service_backed_block="$(make_target_block test-service-backed)"
test_fast_service_backed_block="$(make_target_block test-fast-service-backed)"
check_service_backed_block="$(make_target_block check-service-backed)"
assert_contains "${test_block}" '$(RUN_MAKE_SEQUENCE_SCRIPT)' "make test helper invocation"
assert_contains "${test_block}" "--summary-profile test" "make test summary profile"
assert_contains "${test_block}" "--parallel-step test-local:3 --step test-service-backed --step browser-e2e" "make test sequence"
assert_not_contains "${test_block}" "--step test-isolated" "make test old split browser sequence"
assert_not_contains "${test_block}" "completed=" "make test inline completed counter"
assert_not_contains "${test_block}" "total=" "make test inline total counter"
assert_contains "${check_block}" '$(RUN_CHECK_SCHEDULE_SCRIPT)' "make check scheduler invocation"
assert_contains "${check_block}" "--summary-profile check" "make check summary profile"
assert_contains "${check_block}" '--resource-limit cpu=$(CHECK_JOBS)' "make check scheduler cpu resource"
assert_not_contains "${check_block}" '$(RUN_MAKE_SEQUENCE_SCRIPT)' "make check no longer uses serial sequence helper"
assert_not_contains "${check_block}" "--step browser-e2e" "make check no final serial browser step"
assert_not_contains "${check_block}" "--step check-isolated" "make check old split browser sequence"
assert_not_contains "${check_block}" "completed=" "make check inline completed counter"
assert_not_contains "${check_block}" "total=" "make check inline total counter"
assert_contains "${run_harness_smoke_fast_block}" '$(RUN_HARNESS_SMOKE_SCRIPT) --tier fast --jobs "$(HARNESS_SMOKE_JOBS)"' "run-harness-smoke-fast manifest runner"
assert_contains "${run_harness_smoke_extended_block}" '$(RUN_HARNESS_SMOKE_SCRIPT) --tier extended --jobs "$(HARNESS_SMOKE_JOBS)"' "run-harness-smoke-extended manifest runner"
assert_contains "${run_harness_smoke_full_block}" '$(RUN_HARNESS_SMOKE_SCRIPT) --tier full --jobs "$(HARNESS_SMOKE_JOBS)"' "run-harness-smoke-full manifest runner"
assert_contains "${check_harness_smoke_block}" "run-harness-smoke-fast" "check-harness-smoke fast tier invocation"
assert_contains "${check_harness_smoke_block}" "--projection check-harness-smoke" "check-harness-smoke summary projection"
assert_contains "${manifest_content}" "\"summary_profiles\"" "manifest summary profiles"
assert_contains "${manifest_content}" "\"summary_projection\"" "manifest summary projections"
NODE_BIN="${NODE_BIN:-node}" "${NODE_BIN:-node}" - "${ROOT_DIR}/tools/task_surface_manifest.json" <<'EOF'
const fs = require("node:fs");

const manifest = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const projectedChildren = new Map(
  (manifest.targets ?? []).map((target) => [
    target.name,
    new Set(target.summary_projection?.children ?? []),
  ]),
);
for (const profileName of ["test", "check"]) {
  const roots = new Set(manifest.summary_profiles[profileName].targets);
  for (const root of roots) {
    for (const child of projectedChildren.get(root) ?? []) {
      if (roots.has(child)) {
        throw new Error(`${profileName} summary profile double-counts ${root} and ${child}`);
      }
    }
  }
}
EOF
assert_contains "${manifest_content}" "\"harness_checks\"" "manifest logical harness checks"
assert_contains "${manifest_content}" "harness-smoke-run-make-sequence-fast" "harness smoke fast make sequence check"
assert_contains "${manifest_content}" "harness-smoke-run-go-target-fast" "harness smoke fast go target"
assert_contains "${test_service_backed_block}" '$(RUN_SERVICE_BACKED_SCHEDULE_SCRIPT) --target test-service-backed --manifest "$(SERVICE_BACKED_SCHEDULE_MANIFEST)" --defer-summary' "test service-backed scheduler invocation"
assert_contains "${test_service_backed_block}" '$(TEST_OUTPUT_SCRIPT) target-summary test-service-backed $$requested --projection test-service-backed' "test service-backed post-wrapper summary"
assert_contains "${test_fast_service_backed_block}" '$(RUN_SERVICE_BACKED_SCHEDULE_SCRIPT) --target test-fast-service-backed --manifest "$(SERVICE_BACKED_SCHEDULE_MANIFEST)" --defer-summary' "test-fast service-backed scheduler invocation"
assert_contains "${test_fast_service_backed_block}" '$(TEST_OUTPUT_SCRIPT) target-summary test-fast-service-backed $$requested --projection test-fast-service-backed' "test-fast service-backed post-wrapper summary"
assert_contains "${check_service_backed_block}" '$(RUN_SERVICE_BACKED_SCHEDULE_SCRIPT) --target check-service-backed --manifest "$(SERVICE_BACKED_SCHEDULE_MANIFEST)" --defer-summary' "check service-backed scheduler invocation"
assert_contains "${check_service_backed_block}" '$(TEST_OUTPUT_SCRIPT) target-summary check-service-backed $$requested --projection check-service-backed' "check service-backed post-wrapper summary"
assert_not_contains "${test_service_backed_block}" "--jobs" "test service-backed fixed scheduler jobs"
assert_not_contains "${test_fast_service_backed_block}" "--jobs" "test-fast service-backed fixed scheduler jobs"
assert_not_contains "${check_service_backed_block}" "--jobs" "check service-backed fixed scheduler jobs"
assert_not_contains "${makefile_content}" "test-service-backed-lane-a" "removed fixed test-service-backed lane-a"
assert_not_contains "${makefile_content}" "test-service-backed-lane-b" "removed fixed test-service-backed lane-b"
assert_not_contains "${makefile_content}" "test-service-backed-lane-browser" "removed fixed test-service-backed browser lane"
assert_not_contains "${makefile_content}" "check-service-backed-lane-a" "removed fixed check-service-backed lane-a"
assert_not_contains "${makefile_content}" "check-service-backed-lane-b" "removed fixed check-service-backed lane-b"
assert_not_contains "${makefile_content}" "RUN_SUMMARY =" "unused run summary helper variable"
assert_not_contains "${makefile_content}" "RUN_SUMMARY_CMD =" "unused run summary command variable"
assert_not_contains "${makefile_content}" "bash -lc './scripts/test-check-toolchain-pins.sh &&" "old serialized harness smoke chain"
assert_not_contains "${makefile_content}" "run-phase-smoke:" "removed compatibility phase smoke target"

for target in test run-harness-smoke-fast run-harness-smoke-extended run-harness-smoke-full; do
  make_dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-make-n-${target}.XXXXXX")"
  cleanup_paths+=("${make_dry_run_dir}")
  make_dry_run_output="$(
    CARTULARY_TEST_RESULTS_DIR="${make_dry_run_dir}/results" \
    CARTULARY_TEST_RUN_ID="make-n-${target}" \
      make -n --no-print-directory "${target}" \
      2>&1
  )"
  if [[ "${target}" == run-harness-smoke-* ]]; then
    assert_contains "${make_dry_run_output}" "scripts/run-harness-smoke.mjs --tier ${target#run-harness-smoke-}" "make -n ${target} helper command"
  else
    assert_contains "${make_dry_run_output}" "scripts/run-make-sequence.sh --label ${target}" "make -n ${target} helper command"
  fi
  assert_file_absent "${make_dry_run_dir}/results/make-n-${target}/run-summary.json" "make -n ${target} summary"
done

check_dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-make-n-check.XXXXXX")"
cleanup_paths+=("${check_dry_run_dir}")
check_dry_run_output="$(
  CARTULARY_TEST_RESULTS_DIR="${check_dry_run_dir}/results" \
  CARTULARY_TEST_RUN_ID="make-n-check" \
    make -n --no-print-directory check \
    2>&1
)"
assert_contains "${check_dry_run_output}" "scripts/run-check-schedule.mjs --target check" "make -n check scheduler command"
assert_not_contains "${check_dry_run_output}" "--step browser-e2e" "make -n check no final browser step"
assert_file_absent "${check_dry_run_dir}/results/make-n-check/run-summary.json" "make -n check summary"
