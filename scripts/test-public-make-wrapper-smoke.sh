#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
cleanup_paths=()

unset VERBOSE CI_VERBOSE CARTULARY_OUTPUT_MODE CARTULARY_SUPPRESS_CHILD_SUCCESS

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

  "${NODE_BIN}" -e '
const fs = require("node:fs");
const [file, path] = process.argv.slice(1);
const value = path.split(".").reduce((current, key) => current?.[key], JSON.parse(fs.readFileSync(file, "utf8")));
if (value === undefined || value === null) {
  process.exit(1);
}
process.stdout.write(String(value));
' "${file}" "${path}"
}

write_check() {
  local path="$1"
  local status="$2"
  local label="$3"

  cat >"${path}" <<EOF
#!/usr/bin/env bash
set -euo pipefail
printf '%s\\n' "${label}" >>"\${FAKE_CHECK_LOG:?}"
exit ${status}
EOF
  chmod +x "${path}"
}

write_manifest() {
  local source="$1"
  local destination="$2"
  local script_dir="$3"
  local first_status="$4"

  write_check "${script_dir}/public-wrapper.sh" "${first_status}" "fixture-public-wrapper"
  write_check "${script_dir}/check-scheduler.sh" 0 "fixture-check-scheduler"
  write_check "${script_dir}/service-backed-scheduler.sh" 0 "fixture-service-backed-scheduler"

  "${NODE_BIN}" - "${source}" "${destination}" "${script_dir#"${ROOT_DIR}"/}" <<'EOF'
const fs = require("node:fs");
const [source, destination, scriptDir] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(source, "utf8"));
for (const check of manifest.harness_checks) {
  delete check.gate_smoke_role;
}
const checks = [
  {
    name: "fixture-public-wrapper",
    gate_smoke_role: "public_make_wrapper",
    backing_scripts: [`${scriptDir}/public-wrapper.sh`],
  },
  {
    name: "fixture-check-scheduler",
    gate_smoke_role: "check_scheduler_semantic",
    backing_scripts: [`${scriptDir}/check-scheduler.sh`],
  },
  {
    name: "fixture-service-backed-scheduler",
    gate_smoke_role: "service_backed_scheduler_semantic",
    backing_scripts: [`${scriptDir}/service-backed-scheduler.sh`],
  },
];
manifest.harness_tiers.fast = { checks: checks.map((check) => check.name) };
manifest.harness_checks = manifest.harness_checks.filter(
  (check) => !checks.some((fixture) => fixture.name === check.name),
);
manifest.harness_checks.push(...checks);
fs.writeFileSync(destination, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
}

run_make_smoke() {
  local name="$1"
  local first_status="$2"
  local expected_status="$3"
  local dir="$4"
  local output
  local status

  mkdir -p "${dir}/scripts"
  touch "${dir}/frontend-install.stamp"
  write_manifest "${ROOT_DIR}/tools/task_surface_manifest.json" "${dir}/task_surface_manifest.json" "${dir}/scripts" "${first_status}"

  set +e
  output="$(
    FAKE_CHECK_LOG="${dir}/checks.log" \
    CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
    NODE_BIN="${NODE_BIN}" \
    FRONTEND_INSTALL_STAMP="${dir}/frontend-install.stamp" \
    TASK_SURFACE_MANIFEST="${dir}/task_surface_manifest.json" \
    CARTULARY_TEST_RESULTS_DIR="${dir}/results" \
    CARTULARY_TEST_RUN_ID="${name}" \
    HARNESS_SMOKE_JOBS=1 \
      make --no-print-directory check-harness-smoke \
      2>&1
  )"
  status=$?
  set -e

  if [[ "${status}" != "${expected_status}" ]]; then
    fail "${name} exit status: expected [${expected_status}], got [${status}], output: ${output}"
  fi
  assert_file_present "${dir}/results/${name}/run-harness-smoke-fast/target-summary.json" "${name} child target summary"
  assert_file_present "${dir}/results/${name}/check-harness-smoke/target-summary.json" "${name} projected target summary"
  assert_contains "$(cat "${dir}/checks.log")" "fixture-public-wrapper" "${name} invoked fast harness child through Make"
  printf '%s' "${output}"
}

pass_dir="$(mktemp -d "${ROOT_DIR}/tmp/public-make-wrapper-pass.XXXXXX")"
cleanup_paths+=("${pass_dir}")
pass_output="$(run_make_smoke pass 0 0 "${pass_dir}")"
assert_contains "${pass_output}" "[RESULT] target=check-harness-smoke status=pass" "pass projection result"
assert_contains "${pass_output}" "[ARTIFACTS] target=check-harness-smoke" "pass projection artifacts"
assert_not_contains "${pass_output}" "[CHILD-MISSING]" "pass projection child accounting"
assert_equals "$(json_field "${pass_dir}/results/pass/check-harness-smoke/target-summary.json" "target")" "check-harness-smoke" "pass projected target identity"
assert_equals "$(json_field "${pass_dir}/results/pass/check-harness-smoke/target-summary.json" "status")" "pass" "pass projected target status"
assert_equals "$(json_field "${pass_dir}/results/pass/run-harness-smoke-fast/target-summary.json" "status")" "pass" "pass child target status"

failure_dir="$(mktemp -d "${ROOT_DIR}/tmp/public-make-wrapper-failure.XXXXXX")"
cleanup_paths+=("${failure_dir}")
failure_output="$(run_make_smoke failure 7 2 "${failure_dir}")"
assert_contains "${failure_output}" "[FAIL] target=check-harness-smoke" "failure projection result"
assert_contains "${failure_output}" "[ARTIFACTS] target=check-harness-smoke" "failure projection artifacts"
assert_not_contains "${failure_output}" "[CHILD-MISSING] check-harness-smoke fixture-check-scheduler" "failure skipped child accounting"
assert_equals "$(json_field "${failure_dir}/results/failure/check-harness-smoke/tool-run-summary.json" "exit_code")" "1" "failure projected public exit code"
assert_equals "$(json_field "${failure_dir}/results/failure/check-harness-smoke/target-summary.json" "status")" "fail" "failure projected target status"
assert_equals "$(json_field "${failure_dir}/results/failure/check-harness-smoke/target-summary.json" "children.skipped.0.target")" "fixture-check-scheduler" "failure projected skipped child"
assert_equals "$(json_field "${failure_dir}/results/failure/run-harness-smoke-fast/target-summary.json" "children.failed_targets.0")" "fixture-public-wrapper" "failure child failed target"
