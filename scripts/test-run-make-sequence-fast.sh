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

assert_file_absent() {
  local path="$1"
  local label="$2"

  if [[ -e "${path}" ]]; then
    fail "${label}: expected ${path} to be absent"
  fi
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

write_fake_make() {
  local dir="$1"

  cat >"${dir}/fake-make" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

echo "$*" >>"${FAKE_MAKE_LOG}"

target="${@: -1}"
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

success_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-success.XXXXXX")"
cleanup_paths+=("${success_dir}")
write_fake_make "${success_dir}"
success_output="$(
  MAKE="${success_dir}/fake-make" \
  FAKE_MAKE_LOG="${success_dir}/make.log" \
  CARTULARY_TEST_RESULTS_DIR="${success_dir}/results" \
  CARTULARY_TEST_RUN_ID="success" \
    "${SCRIPT}" --label smoke --summary-targets " alpha, beta " --summary-groups "alpha-group=alpha;beta-group=beta" --step alpha --parallel-step beta:3 \
    2>&1
)"
assert_contains "${success_output}" "[RUN] smoke steps=2 targets=2 jobs=3 run_id=success" "success run start output"
assert_contains "${success_output}" "[STEP] smoke 1/2 alpha mode=serial jobs=1" "success serial step output"
assert_contains "${success_output}" "[STEP] smoke 2/2 beta mode=parallel jobs=3" "success parallel step output"
assert_contains "${success_output}" "[PASS] smoke" "success run summary output"
assert_contains "$(cat "${success_dir}/make.log")" "--output-sync=target -j3 beta" "parallel make invocation"

dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-dry-run.XXXXXX")"
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

invalid_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-invalid.XXXXXX")"
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
if [[ "${invalid_status}" != "2" ]]; then
  fail "invalid usage status: expected [2], got [${invalid_status}]"
fi
assert_contains "${invalid_output}" "--parallel-step requires <target>:<jobs>" "invalid usage output"
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

const expectedFull = [...fast.targets, ...extended.targets, ...lifecycle.targets];
if (JSON.stringify(full.targets) !== JSON.stringify(expectedFull)) {
  fail("full harness tier must equal fast + extended + lifecycle tiers");
}

const tierMembership = new Map();
for (const [tier, targets] of [["fast", fast.targets], ["extended", extended.targets], ["lifecycle", lifecycle.targets]]) {
  for (const target of targets) {
    if (tierMembership.has(target)) {
      fail(`${target} is present in both ${tierMembership.get(target)} and ${tier}`);
    }
    tierMembership.set(target, tier);
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

assert_contains "${run_fast_block}" "--summary-profile run-harness-smoke-fast" "fast harness summary profile"
assert_contains "${run_fast_block}" "--parallel-step run-harness-smoke-fast-all:\$(HARNESS_SMOKE_JOBS)" "fast harness parallel aggregate step"
assert_contains "${run_extended_block}" "--summary-profile run-harness-smoke-extended" "extended harness summary profile"
assert_contains "${run_extended_block}" "--parallel-step run-harness-smoke-extended-all:\$(HARNESS_SMOKE_JOBS)" "extended harness parallel aggregate step"
assert_contains "${run_full_block}" "--summary-profile run-harness-smoke-full" "full harness summary profile"
assert_contains "${check_harness_smoke_block}" "run-harness-smoke-fast" "check harness fast tier"
assert_contains "${check_harness_smoke_block}" "--projection check-harness-smoke" "check harness summary projection"
assert_contains "${ci_script}" "make --no-print-directory check" "CI check invocation"
assert_contains "${ci_script}" "make --no-print-directory run-harness-smoke-extended" "CI extended harness invocation"
assert_contains "${release_check_block}" "release-check: check run-harness-smoke-extended license-report sbom build" "release-check extended harness dependency"
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
  assert_contains "${make_dry_run_output}" "scripts/run-make-sequence.sh --label ${target}" "make -n ${target} helper command"
  assert_file_absent "${make_dry_run_dir}/results/make-n-${target}/run-summary.json" "make -n ${target} summary"
done
