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

assert_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  if [[ "${actual}" != "${expected}" ]]; then
    fail "${label}: expected [${expected}], got [${actual}]"
  fi
}

assert_file_absent() {
  local path="$1"
  local label="$2"

  if [[ -e "${path}" ]]; then
    fail "${label}: expected ${path} to be absent"
  fi
}

assert_file_present() {
  local path="$1"
  local label="$2"

  if [[ ! -f "${path}" ]]; then
    fail "${label}: expected ${path} to exist"
  fi
}

json_field() {
  local file="$1"
  local path="$2"

  "${NODE_BIN:-node}" -e '
const fs = require("node:fs");
const [file, path] = process.argv.slice(1);
const value = path.split(".").reduce((current, key) => current?.[key], JSON.parse(fs.readFileSync(file, "utf8")));
if (value === undefined || value === null) {
  process.exit(1);
}
process.stdout.write(String(value));
' "${file}" "${path}"
}

make_target_block() {
  local target="$1"

  cat "${ROOT_DIR}/tools/task_surface.generated.mk" "${ROOT_DIR}/Makefile" | awk -v target="${target}" '
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

write_fake_make() {
  local dir="$1"

  cat >"${dir}/fake-make" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

echo "$*" >>"${FAKE_MAKE_LOG}"

target="${@: -1}"
if [[ -n "${CARTULARY_TEST_RESULTS_DIR:-}" && -n "${CARTULARY_TEST_RUN_ID:-}" ]]; then
  write_summary() {
    local summary_target="$1"
    mkdir -p "${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${summary_target}"
    cat >"${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${summary_target}/target-summary.json" <<JSON
{
  "target": "${summary_target}",
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
  }
  case "${target}" in
    test-local)
      write_summary backend-unit
      write_summary frontend-typecheck
      write_summary frontend-unit
      ;;
    test-fast-service-backed)
      write_summary backend-integration
      write_summary backend-integration-support
      write_summary backend-store
      write_summary backend-process
      write_summary test-fast-service-backed
      ;;
    check)
      write_summary check
      ;;
    run-harness-smoke-extended)
      write_summary run-harness-smoke-extended
      ;;
    *)
      write_summary "${target}"
      ;;
  esac
fi
EOF
  chmod +x "${dir}/fake-make"
}

manifest_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-manifest.XXXXXX")"
cleanup_paths+=("${manifest_dir}")
sequence_manifest="${manifest_dir}/task_surface_manifest.json"
"${NODE_BIN:-node}" - "${ROOT_DIR}/tools/task_surface_manifest.json" "${sequence_manifest}" <<'EOF'
const fs = require("node:fs");
const [source, destination] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(source, "utf8"));
for (const name of ["alpha", "beta", "smoke", "dry-run"]) {
  if (!manifest.targets.some((target) => target.name === name)) {
    manifest.targets.push({ name, classification: "helper_only", included_in: ["helper_only"] });
  }
}
manifest.sequences.smoke = {
  summary_groups: [
    { name: "alpha-group", summary_targets: ["alpha"] },
    { name: "beta-group", summary_targets: ["beta"] },
  ],
  steps: [
    { type: "step", target: "alpha", produces_summary_targets: ["alpha"] },
    { type: "parallel", target: "beta", jobs: 3, produces_summary_targets: ["beta"] },
  ],
};
manifest.sequences["dry-run"] = {
  summary_groups: [],
  steps: [{ type: "step", target: "alpha", produces_summary_targets: ["alpha"] }],
};
fs.writeFileSync(destination, `${JSON.stringify(manifest, null, 2)}\n`);
EOF

success_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-success.XXXXXX")"
cleanup_paths+=("${success_dir}")
write_fake_make "${success_dir}"
success_output="$(
  VERBOSE= \
  CI_VERBOSE= \
  CARTULARY_OUTPUT_MODE= \
  MAKE="${success_dir}/fake-make" \
  FAKE_MAKE_LOG="${success_dir}/make.log" \
  CARTULARY_TEST_RESULTS_DIR="${success_dir}/results" \
  CARTULARY_TEST_RUN_ID="success" \
  TASK_SURFACE_MANIFEST="${sequence_manifest}" \
    "${SCRIPT}" --sequence smoke \
    2>&1
)"
assert_contains "${success_output}" "[RUN] smoke work_units=2 summary_targets=2 helper_units=0 jobs=3 run_id=success" "success run start output"
assert_contains "${success_output}" "[STEP] smoke 1/2 alpha mode=serial jobs=1" "success serial step output"
assert_contains "${success_output}" "[STEP] smoke 2/2 beta mode=parallel jobs=3" "success parallel step output"
assert_contains "${success_output}" "[PASS] smoke" "success run summary output"
assert_file_present "${success_dir}/results/success/smoke/target-summary.json" "success target summary"
assert_equals "$(json_field "${success_dir}/results/success/smoke/target-summary.json" "target")" "smoke" "success target summary identity"
assert_contains "$(cat "${success_dir}/make.log")" "--output-sync=target -j3 beta" "parallel make invocation"

for aggregate_sequence in test-fast ci release-check; do
  aggregate_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-${aggregate_sequence}.XXXXXX")"
  cleanup_paths+=("${aggregate_dir}")
  write_fake_make "${aggregate_dir}"
  aggregate_output="$(
    VERBOSE= \
    CI_VERBOSE= \
    CARTULARY_OUTPUT_MODE= \
    MAKE="${aggregate_dir}/fake-make" \
    FAKE_MAKE_LOG="${aggregate_dir}/make.log" \
    CARTULARY_TEST_RESULTS_DIR="${aggregate_dir}/results" \
    CARTULARY_TEST_RUN_ID="${aggregate_sequence}" \
    TASK_SURFACE_MANIFEST="${ROOT_DIR}/tools/task_surface_manifest.json" \
      "${SCRIPT}" --sequence "${aggregate_sequence}" \
      2>&1
  )"
  assert_contains "${aggregate_output}" "[PASS] ${aggregate_sequence}" "${aggregate_sequence} run summary output"
  assert_file_present "${aggregate_dir}/results/${aggregate_sequence}/${aggregate_sequence}/target-summary.json" "${aggregate_sequence} target summary"
  assert_equals "$(json_field "${aggregate_dir}/results/${aggregate_sequence}/${aggregate_sequence}/target-summary.json" "target")" "${aggregate_sequence}" "${aggregate_sequence} target summary identity"
  assert_equals "$(json_field "${aggregate_dir}/results/${aggregate_sequence}/run-summary.json" "label")" "${aggregate_sequence}" "${aggregate_sequence} run summary identity"
done

dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-dry-run.XXXXXX")"
cleanup_paths+=("${dry_run_dir}")
write_fake_make "${dry_run_dir}"
dry_run_output="$(
  VERBOSE= \
  CI_VERBOSE= \
  CARTULARY_OUTPUT_MODE= \
  MAKEFLAGS="n" \
  MAKE="${dry_run_dir}/fake-make" \
  FAKE_MAKE_LOG="${dry_run_dir}/make.log" \
  CARTULARY_TEST_RESULTS_DIR="${dry_run_dir}/results" \
  CARTULARY_TEST_RUN_ID="dry-run" \
  TASK_SURFACE_MANIFEST="${sequence_manifest}" \
    "${SCRIPT}" --sequence dry-run \
    2>&1
)"
assert_not_contains "${dry_run_output}" "[RUN]" "script dry-run run start output"
assert_not_contains "${dry_run_output}" "[STEP]" "script dry-run step output"
assert_file_absent "${dry_run_dir}/results/dry-run/run-summary.json" "script dry-run summary"
assert_contains "$(cat "${dry_run_dir}/make.log")" "--no-print-directory alpha" "script dry-run child make"

harness_quiet_dir="$(mktemp -d "${ROOT_DIR}/tmp/harness-smoke-quiet.XXXXXX")"
cleanup_paths+=("${harness_quiet_dir}")
mkdir -p "${harness_quiet_dir}/scripts"
cat >"${harness_quiet_dir}/scripts/check-a.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
exit 0
EOF
cat >"${harness_quiet_dir}/scripts/check-b.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
exit 0
EOF
chmod +x "${harness_quiet_dir}/scripts/check-a.sh" "${harness_quiet_dir}/scripts/check-b.sh"
harness_manifest="${harness_quiet_dir}/manifest.json"
"${NODE_BIN:-node}" - "${ROOT_DIR}/tools/task_surface_manifest.json" "${harness_manifest}" "${harness_quiet_dir#${ROOT_DIR}/}/scripts" <<'EOF'
const fs = require("node:fs");
const [source, destination, scriptDir] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(source, "utf8"));
const checks = ["harness-quiet-a", "harness-quiet-b"];
manifest.harness_tiers.fast = { checks };
manifest.harness_checks.push(
  { name: "harness-quiet-a", backing_scripts: [`${scriptDir}/check-a.sh`] },
  { name: "harness-quiet-b", backing_scripts: [`${scriptDir}/check-b.sh`] },
);
fs.writeFileSync(destination, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
harness_quiet_output="$(
  VERBOSE= \
  CI_VERBOSE= \
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="${harness_quiet_dir}/results" \
  CARTULARY_TEST_RUN_ID="quiet" \
  TASK_SURFACE_MANIFEST="${harness_manifest}" \
    "${NODE_BIN:-node}" "${ROOT_DIR}/scripts/run-harness-smoke.mjs" --tier fast --jobs 2 --manifest "${harness_manifest}" \
    2>&1
)"
assert_equals "${harness_quiet_output}" "" "quiet harness internal success output"
check_harness_quiet_output="$(
  VERBOSE= \
  CI_VERBOSE= \
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="${harness_quiet_dir}/results" \
  CARTULARY_TEST_RUN_ID="quiet" \
    "${ROOT_DIR}/scripts/lib/test-output.sh" target-summary check-harness-smoke pass --children harness-quiet-a,harness-quiet-b \
    2>&1
)"
assert_contains "${check_harness_quiet_output}" "[PASS] check-harness-smoke kind=aggregate children=2/2" "quiet check harness aggregate summary"
assert_contains "${check_harness_quiet_output}" "failed_children=none" "quiet check harness failure hint"
assert_not_contains "${check_harness_quiet_output}" "[CHILD]" "quiet check harness hides child detail"

harness_failure_dir="$(mktemp -d "${ROOT_DIR}/tmp/harness-smoke-failure.XXXXXX")"
cleanup_paths+=("${harness_failure_dir}")
mkdir -p "${harness_failure_dir}/scripts"
cat >"${harness_failure_dir}/scripts/check-fail.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
exit 7
EOF
cat >"${harness_failure_dir}/scripts/check-skipped.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
exit 0
EOF
chmod +x "${harness_failure_dir}/scripts/check-fail.sh" "${harness_failure_dir}/scripts/check-skipped.sh"
harness_failure_manifest="${harness_failure_dir}/manifest.json"
"${NODE_BIN:-node}" - "${ROOT_DIR}/tools/task_surface_manifest.json" "${harness_failure_manifest}" "${harness_failure_dir#${ROOT_DIR}/}/scripts" <<'EOF'
const fs = require("node:fs");
const [source, destination, scriptDir] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(source, "utf8"));
const checks = ["harness-fail-a", "harness-skipped-b"];
manifest.harness_tiers.fast = { checks };
manifest.harness_checks.push(
  { name: "harness-fail-a", backing_scripts: [`${scriptDir}/check-fail.sh`] },
  { name: "harness-skipped-b", backing_scripts: [`${scriptDir}/check-skipped.sh`] },
);
fs.writeFileSync(destination, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
set +e
harness_failure_output="$(
  VERBOSE= \
  CI_VERBOSE= \
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="${harness_failure_dir}/results" \
  CARTULARY_TEST_RUN_ID="failure" \
  TASK_SURFACE_MANIFEST="${harness_failure_manifest}" \
    "${NODE_BIN:-node}" "${ROOT_DIR}/scripts/run-harness-smoke.mjs" --tier fast --jobs 1 --manifest "${harness_failure_manifest}" \
    2>&1
)"
harness_failure_status=$?
set -e
assert_equals "${harness_failure_status}" "7" "failing harness preserves child status"
assert_contains "${harness_failure_output}" "[CHILD-SKIPPED] run-harness-smoke-fast harness-skipped-b reason=schedule_stopped_after_failure failed_dependency=harness-fail-a" "failing harness reports skipped child"
assert_not_contains "${harness_failure_output}" "[CHILD-MISSING] run-harness-smoke-fast harness-skipped-b" "failing harness does not report skipped child missing"
assert_not_contains "${harness_failure_output}" "missing child target summary: harness-skipped-b" "failing harness does not create missing child artifact failure"
harness_failure_summary="${harness_failure_dir}/results/failure/run-harness-smoke-fast/target-summary.json"
assert_equals "$(json_field "${harness_failure_summary}" "children.missing.length")" "0" "failing harness skipped child missing list"
assert_equals "$(json_field "${harness_failure_summary}" "children.skipped.0.target")" "harness-skipped-b" "failing harness skipped child target"
assert_equals "$(json_field "${harness_failure_summary}" "children.skipped.0.reason")" "schedule_stopped_after_failure" "failing harness skipped child reason"
assert_equals "$(json_field "${harness_failure_summary}" "children.skipped.0.failed_dependency")" "harness-fail-a" "failing harness skipped child dependency"
assert_equals "$(json_field "${harness_failure_summary}" "children.failed_targets.0")" "harness-fail-a" "failing harness failed child"

invalid_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-invalid.XXXXXX")"
cleanup_paths+=("${invalid_dir}")
write_fake_make "${invalid_dir}"
set +e
invalid_output="$(
  MAKE="${invalid_dir}/fake-make" \
  FAKE_MAKE_LOG="${invalid_dir}/make.log" \
    "${SCRIPT}" --summary-targets alpha \
    2>&1
)"
invalid_status=$?
set -e
if [[ "${invalid_status}" != "2" ]]; then
  fail "invalid usage status: expected [2], got [${invalid_status}]"
fi
assert_contains "${invalid_output}" "usage: run-make-sequence.sh --sequence <name>" "invalid usage output"
assert_file_absent "${invalid_dir}/make.log" "invalid usage child make log"

NODE_BIN="${NODE_BIN:-node}" "${NODE_BIN:-node}" - "${ROOT_DIR}/tools/task_surface_manifest.json" <<'EOF'
const fs = require("node:fs");

const [manifestPath] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestPath, "utf8"));
const { fast, extended, lifecycle, full } = manifest.harness_tiers;

function fail(message) {
  console.error(message);
  process.exit(1);
}

const expectedFull = [...fast.checks, ...extended.checks, ...lifecycle.checks];
if (JSON.stringify(full.checks) !== JSON.stringify(expectedFull)) {
  fail("full harness tier must equal fast + extended + lifecycle tiers");
}

const tierMembership = new Map();
for (const [tier, checks] of [["fast", fast.checks], ["extended", extended.checks], ["lifecycle", lifecycle.checks]]) {
  for (const check of checks) {
    if (tierMembership.has(check)) {
      fail(`${check} is present in both ${tierMembership.get(check)} and ${tier}`);
    }
    tierMembership.set(check, tier);
  }
}

for (const target of ["harness-smoke-run-make-sequence", "harness-smoke-run-go-target"]) {
  if (tierMembership.get(target) !== "extended") {
    fail(`${target} must stay in extended harness smoke`);
  }
}
for (const target of ["harness-smoke-run-make-sequence-fast", "harness-smoke-run-go-target-fast"]) {
  if (tierMembership.get(target) !== "fast") {
    fail(`${target} must stay in fast harness smoke`);
  }
}
EOF

makefile_content="$(cat "${ROOT_DIR}/Makefile")"
run_fast_block="$(make_target_block run-harness-smoke-fast)"
run_extended_block="$(make_target_block run-harness-smoke-extended)"
run_full_block="$(make_target_block run-harness-smoke-full)"
check_harness_smoke_block="$(make_target_block check-harness-smoke)"
release_check_block="$(make_target_block release-check)"
ci_script="$(cat "${ROOT_DIR}/scripts/ci/verify.sh")"
test_fast_block="$(make_target_block test-fast)"

assert_contains "${test_fast_block}" '$(RUN_MAKE_SEQUENCE_SCRIPT) --sequence test-fast' "test-fast sequence runner"
assert_contains "${run_fast_block}" '$(RUN_HARNESS_SMOKE_SCRIPT) --tier fast --jobs "$(HARNESS_SMOKE_JOBS)"' "fast harness manifest runner"
assert_contains "${run_extended_block}" '$(RUN_HARNESS_SMOKE_SCRIPT) --tier extended --jobs "$(HARNESS_SMOKE_JOBS)"' "extended harness manifest runner"
assert_contains "${run_full_block}" '$(RUN_HARNESS_SMOKE_SCRIPT) --tier full --jobs "$(HARNESS_SMOKE_JOBS)"' "full harness manifest runner"
assert_contains "${check_harness_smoke_block}" "run-harness-smoke-fast" "check harness fast tier"
assert_contains "${check_harness_smoke_block}" "--projection check-harness-smoke" "check harness summary projection"
assert_contains "${ci_script}" "exec make --no-print-directory ci" "CI script delegates to canonical target"
assert_contains "${release_check_block}" '$(RUN_MAKE_SEQUENCE_SCRIPT) --sequence release-check' "release-check sequence runner"
assert_contains "$(cat "${ROOT_DIR}/tools/task_surface_manifest.json")" "scripts/test-run-make-sequence-fast.sh" "fast make-sequence smoke backing script"
assert_contains "$(cat "${ROOT_DIR}/tools/task_surface_manifest.json")" "scripts/test-run-go-target-fast.sh" "fast run-go-target smoke backing script"

for target in run-harness-smoke-fast run-harness-smoke-extended run-harness-smoke-full; do
  make_dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-make-n-${target}.XXXXXX")"
  cleanup_paths+=("${make_dry_run_dir}")
  make_dry_run_output="$(
    CARTULARY_TEST_RESULTS_DIR="${make_dry_run_dir}/results" \
    CARTULARY_TEST_RUN_ID="make-n-${target}" \
      make -n --no-print-directory "${target}" \
      2>&1
  )"
  assert_contains "${make_dry_run_output}" "scripts/run-harness-smoke.mjs --tier ${target#run-harness-smoke-}" "make -n ${target} helper command"
  assert_file_absent "${make_dry_run_dir}/results/make-n-${target}/run-summary.json" "make -n ${target} summary"
done
