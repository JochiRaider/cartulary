#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
SCRIPT="${ROOT_DIR}/scripts/run-phase-slice.mjs"
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

assert_file_absent() {
  local path="$1"
  local label="$2"

  if [[ -e "$path" ]]; then
    fail "$label: expected $path to be absent"
  fi
}

json_field() {
  local file="$1"
  local path="$2"

  "$NODE_BIN" -e '
const fs = require("node:fs");
const [file, path] = process.argv.slice(1);
const value = path.split(".").reduce((current, key) => current?.[key], JSON.parse(fs.readFileSync(file, "utf8")));
if (value === undefined || value === null) {
  process.exit(1);
}
process.stdout.write(String(value));
' "$file" "$path"
}

write_fake_make() {
  local dir="$1"

  cat >"${dir}/fake-make" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

echo "$*" >>"${FAKE_MAKE_LOG}"

target="${@: -1}"
if [[ -n "${FAKE_MAKE_ENV_LOG:-}" ]]; then
  printf 'target=%s test_target=%s\n' "$target" "${CARTULARY_TEST_TARGET:-}" >>"${FAKE_MAKE_ENV_LOG}"
fi
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

phase4_dir="$(mktemp -d "${ROOT_DIR}/tmp/phase-slice-phase4.XXXXXX")"
cleanup_paths+=("$phase4_dir")
write_fake_make "$phase4_dir"
phase4_results="${phase4_dir}/results"
phase4_output="$(
  MAKE="${phase4_dir}/fake-make" \
  FAKE_MAKE_LOG="${phase4_dir}/make.log" \
  FAKE_MAKE_ENV_LOG="${phase4_dir}/make-env.log" \
  CARTULARY_TEST_RESULTS_DIR="$phase4_results" \
  CARTULARY_TEST_RUN_ID="phase4" \
    "$NODE_BIN" "$SCRIPT" --phase phase4 --mode phase \
    2>&1
)"
assert_contains "$phase4_output" "[PASS] phase-slice kind=aggregate children=5/5" "phase4 phase-slice summary"
assert_equals "$(cat "${phase4_dir}/make.log")" "$(printf '%s\n%s\n%s\n%s\n%s' \
  "--no-print-directory backend-unit" \
  "--no-print-directory backend-store" \
  "--no-print-directory backend-integration" \
  "--no-print-directory backend-integration-support" \
  "--no-print-directory browser-e2e-webserver-backed")" "phase4 phase-slice child order"
assert_contains "$(cat "${phase4_dir}/make-env.log")" "target=backend-integration-support test_target=backend-integration-support" "phase-slice internal support target ownership"
phase4_summary="${phase4_results}/phase4/phase-slice/target-summary.json"
assert_equals "$(json_field "$phase4_summary" "status")" "pass" "phase4 phase-slice status"
assert_equals "$(json_field "$phase4_summary" "children.expected.3")" "backend-integration-support" "phase4 phase-slice support child"

service_dir="$(mktemp -d "${ROOT_DIR}/tmp/phase-slice-service.XXXXXX")"
cleanup_paths+=("$service_dir")
write_fake_make "$service_dir"
service_results="${service_dir}/results"
service_output="$(
  MAKE="${service_dir}/fake-make" \
  FAKE_MAKE_LOG="${service_dir}/make.log" \
  FAKE_MAKE_ENV_LOG="${service_dir}/make-env.log" \
  CARTULARY_TEST_RESULTS_DIR="$service_results" \
  CARTULARY_TEST_RUN_ID="phase4-service" \
    "$NODE_BIN" "$SCRIPT" --phase phase4 --mode service-backed \
    2>&1
)"
assert_contains "$service_output" "[PASS] service-backed-slice kind=aggregate children=4/4" "phase4 service-backed-slice summary"
assert_equals "$(cat "${service_dir}/make.log")" "$(printf '%s\n%s\n%s\n%s' \
  "--no-print-directory backend-store" \
  "--no-print-directory backend-integration" \
  "--no-print-directory backend-integration-support" \
  "--no-print-directory browser-e2e-webserver-backed")" "phase4 service-backed-slice child order"
assert_contains "$(cat "${service_dir}/make-env.log")" "target=browser-e2e-webserver-backed test_target=browser-e2e-webserver-backed" "service-backed-slice browser target ownership"

phase_root="$(mktemp -d "${ROOT_DIR}/tmp/phase-slice-root.XXXXXX")"
cleanup_paths+=("$phase_root")
mkdir -p "${phase_root}/tools"
cat >"${phase_root}/tools/phase99_test_map.json" <<'JSON'
{
  "expected_ids": ["U-99-01"],
  "support_go_targets": [
    {
      "target": "backend_integration_support",
      "section": "integration",
      "package": "./internal/modules/entities",
      "file": "internal/modules/entities/phase99_support_integration_test.go",
      "symbol": "TestSupportPhase99Integration_FutureSupport",
      "selection_pattern": "TestSupportPhase99Integration_",
      "execution_family": "backend-integration-entities",
      "execution_label": "backend-integration support phase99 entities"
    }
  ]
}
JSON
cat >"${phase_root}/tools/phase100_test_map.json" <<'JSON'
{
  "expected_ids": ["U-100-01"],
  "unit": [
    {
      "id": "U-100-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "./internal/modules/auth",
      "file": "internal/modules/auth/phase100_unit_test.go",
      "symbol": "TestPhase100_UnitOnly_U_100_01",
      "execution_dependency": "backend_unit",
      "execution_family": "backend-unit",
      "execution_label": "backend-unit phase100 authoritative"
    }
  ]
}
JSON

future_dir="$(mktemp -d "${ROOT_DIR}/tmp/phase-slice-future.XXXXXX")"
cleanup_paths+=("$future_dir")
write_fake_make "$future_dir"
future_output="$(
  MAKE="${future_dir}/fake-make" \
  FAKE_MAKE_LOG="${future_dir}/make.log" \
  FAKE_MAKE_ENV_LOG="${future_dir}/make-env.log" \
  CARTULARY_PHASE_MANIFEST_ROOT="$phase_root" \
  CARTULARY_TEST_RESULTS_DIR="${future_dir}/results" \
  CARTULARY_TEST_RUN_ID="future-support" \
    "$NODE_BIN" "$SCRIPT" --phase phase99 --mode service-backed \
    2>&1
)"
assert_contains "$future_output" "[PASS] service-backed-slice kind=aggregate children=1/1" "future support-only service-backed summary"
assert_equals "$(cat "${future_dir}/make.log")" "--no-print-directory backend-integration-support" "future support-only child target"
assert_contains "$(cat "${future_dir}/make-env.log")" "target=backend-integration-support test_target=backend-integration-support" "future support-only target ownership"

noop_dir="$(mktemp -d "${ROOT_DIR}/tmp/phase-slice-noop.XXXXXX")"
cleanup_paths+=("$noop_dir")
write_fake_make "$noop_dir"
noop_output="$(
  MAKE="${noop_dir}/fake-make" \
  FAKE_MAKE_LOG="${noop_dir}/make.log" \
  FAKE_MAKE_ENV_LOG="${noop_dir}/make-env.log" \
  CARTULARY_PHASE_MANIFEST_ROOT="$phase_root" \
  CARTULARY_TEST_RESULTS_DIR="${noop_dir}/results" \
  CARTULARY_TEST_RUN_ID="noop" \
    "$NODE_BIN" "$SCRIPT" --phase phase100 --mode service-backed \
    2>&1
)"
assert_contains "$noop_output" "[NOOP] service-backed-slice phase=phase100 mode=service-backed children=0" "unit-only service-backed no-op"
assert_contains "$noop_output" "[PASS] service-backed-slice kind=leaf" "unit-only service-backed no-op summary"
assert_file_absent "${noop_dir}/make.log" "unit-only service-backed no-op child make log"

set +e
unknown_output="$("$NODE_BIN" "$SCRIPT" --phase phase404 --mode phase 2>&1)"
unknown_status=$?
set -e
if [[ "$unknown_status" -eq 0 ]]; then
  fail "unknown phase must fail"
fi
assert_contains "$unknown_output" "unknown phase phase404" "unknown phase output"
