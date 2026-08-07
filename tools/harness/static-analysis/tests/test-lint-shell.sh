#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
SCRIPT="${ROOT_DIR}/tools/harness/static-analysis/shellcheck.sh"
cleanup_paths=()

# shellcheck source=tools/harness/test-support/harness-scratch.sh
# shellcheck disable=SC1091
source "$ROOT_DIR/tools/harness/test-support/harness-scratch.sh"

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -rf "$path"
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

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: expected output not to contain [$needle]"
  fi
}

assert_equals() {
  local got="$1"
  local want="$2"
  local label="$3"

  if [[ "$got" != "$want" ]]; then
    fail "$label: got [$got], want [$want]"
  fi
}

make_fixture_repo() {
  local dir
  dir="$(cartulary_harness_mktemp_dir lint-shell-fixture.XXXXXX)"
  cleanup_paths+=("$dir")
  git -C "$dir" init -q
  printf '%s\n' "$dir"
}

track_all() {
  local repo="$1"
  git -C "$repo" add -A
}

make_fake_shellcheck() {
  local dir="$1"
  local fake="$dir/tmp/fake-shellcheck"

  mkdir -p "$dir/tmp"

cat >"$fake" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$@" >"${FAKE_SHELLCHECK_ARGS_LOG:?}"
if [[ "${1:-}" == "--version" ]]; then
  printf '%s\n' "ShellCheck - shell script analysis tool" "version: 0.11.0"
  exit 0
fi
if [[ -n "${FAKE_SHELLCHECK_OUTPUT:-}" ]]; then
  printf '%s\n' "$FAKE_SHELLCHECK_OUTPUT" >&2
fi
exit "${FAKE_SHELLCHECK_STATUS:-0}"
EOF
  chmod +x "$fake"
  printf '%s\n' "$fake"
}

make_counting_shellcheck() {
  local dir="$1"
  local fake="$dir/tmp/counting-shellcheck"

  mkdir -p "$dir/tmp"

cat >"$fake" <<'EOF'
#!/usr/bin/env bash
if [[ "${1:-}" == "--version" ]]; then
  printf '%s\n' "ShellCheck - shell script analysis tool" "version: 0.11.0"
  exit 0
fi
printf 'RUN\n' >>"${FAKE_SHELLCHECK_RUN_LOG:?}"
printf '%s\n' "$@" >"${FAKE_SHELLCHECK_ARGS_LOG:?}"
if [[ -n "${FAKE_SHELLCHECK_OUTPUT:-}" ]]; then
  printf '%s\n' "$FAKE_SHELLCHECK_OUTPUT" >&2
fi
exit "${FAKE_SHELLCHECK_STATUS:-0}"
EOF
  chmod +x "$fake"
  printf '%s\n' "$fake"
}

count_shellcheck_runs() {
  local log_file="$1"

  if [[ ! -f "$log_file" ]]; then
    printf '0\n'
    return 0
  fi
  grep -c '^RUN$' "$log_file" || true
}

inventory_repo="$(make_fixture_repo)"
mkdir -p \
  "$inventory_repo/bin" \
  "$inventory_repo/scripts" \
  "$inventory_repo/generated" \
  "$inventory_repo/vendor" \
  "$inventory_repo/node_modules/pkg" \
  "$inventory_repo/tmp" \
  "$inventory_repo/.cartulary/test-results" \
  "$inventory_repo/internal/gen" \
  "$inventory_repo/packages/protocol-ts/src/generated" \
  "$inventory_repo/apps/web/dist" \
  "$inventory_repo/build" \
  "$inventory_repo/reports"
printf '%s\n' 'echo z' >"$inventory_repo/scripts/z.sh"
printf '%s\n' 'echo a' >"$inventory_repo/scripts/a.sh"
printf '%s\n' '#!/usr/bin/env bash' 'echo shebang' >"$inventory_repo/bin/shebang-runner"
printf '%s\n' '#!/usr/bin/env python3' 'print("no")' >"$inventory_repo/scripts/not-shell"
printf '%s\n' 'echo generated' >"$inventory_repo/generated/skip.sh"
printf '%s\n' 'echo vendor' >"$inventory_repo/vendor/skip.sh"
printf '%s\n' 'echo node_modules' >"$inventory_repo/node_modules/pkg/skip.sh"
printf '%s\n' 'echo tmp' >"$inventory_repo/tmp/skip.sh"
printf '%s\n' 'echo cartulary report' >"$inventory_repo/.cartulary/test-results/skip.sh"
printf '%s\n' 'echo internal gen' >"$inventory_repo/internal/gen/skip.sh"
printf '%s\n' 'echo protocol generated' >"$inventory_repo/packages/protocol-ts/src/generated/skip.sh"
printf '%s\n' 'echo dist' >"$inventory_repo/apps/web/dist/skip.sh"
printf '%s\n' 'echo build' >"$inventory_repo/build/skip.sh"
printf '%s\n' 'echo report' >"$inventory_repo/reports/skip.sh"
track_all "$inventory_repo"

fake_shellcheck="$(make_fake_shellcheck "$inventory_repo")"
args_log="$inventory_repo/shellcheck-args.log"
inventory_output="$(
  CARTULARY_SHELLCHECK_ROOT="$inventory_repo" \
  CARTULARY_STEP_ARTIFACT_DIR="" \
  SHELLCHECK_BIN="$fake_shellcheck" \
  FAKE_SHELLCHECK_ARGS_LOG="$args_log" \
    "$SCRIPT" 2>&1
)"
expected_inventory_output=$'bin/shebang-runner\ngenerated/skip.sh\nscripts/a.sh\nscripts/z.sh\n4 files checked'
assert_equals "$inventory_output" "$expected_inventory_output" "deterministic lint-shell output"
expected_args=$'bin/shebang-runner\ngenerated/skip.sh\nscripts/a.sh\nscripts/z.sh'
assert_equals "$(cat "$args_log")" "$expected_args" "deterministic shellcheck argv"

artifact_dir="$(cartulary_harness_mktemp_dir lint-shell-artifacts.XXXXXX)"
cleanup_paths+=("$artifact_dir")
artifact_output="$(
  CARTULARY_SHELLCHECK_ROOT="$inventory_repo" \
  CARTULARY_STEP_ARTIFACT_DIR="$artifact_dir" \
  SHELLCHECK_BIN="$fake_shellcheck" \
  FAKE_SHELLCHECK_ARGS_LOG="$args_log" \
    "$SCRIPT" 2>&1
)"
assert_equals "$artifact_output" "4 files checked" "artifact-mode lint-shell output"
assert_equals "$(cat "$artifact_dir/shellcheck-inventory.txt")" "$expected_args" "artifact-mode lint-shell inventory"

empty_repo="$(make_fixture_repo)"
mkdir -p "$empty_repo/scripts"
printf '%s\n' '#!/usr/bin/env python3' 'print("not shell")' >"$empty_repo/scripts/not-shell"
track_all "$empty_repo"
empty_output="$(
  CARTULARY_SHELLCHECK_ROOT="$empty_repo" \
  CARTULARY_STEP_ARTIFACT_DIR="" \
  SHELLCHECK_BIN="$empty_repo/missing-shellcheck" \
    "$SCRIPT" 2>&1
)"
assert_equals "$empty_output" "0 files checked" "empty lint-shell output"

warning_repo="$(make_fixture_repo)"
mkdir -p "$warning_repo/scripts"
printf '%s\n' "echo \"\$HOME\"" >"$warning_repo/scripts/warn.sh"
track_all "$warning_repo"
warning_fake_shellcheck="$(make_fake_shellcheck "$warning_repo")"
warning_args_log="$warning_repo/shellcheck-args.log"
set +e
warning_output="$(
  CARTULARY_SHELLCHECK_ROOT="$warning_repo" \
  SHELLCHECK_BIN="$warning_fake_shellcheck" \
  FAKE_SHELLCHECK_ARGS_LOG="$warning_args_log" \
  FAKE_SHELLCHECK_OUTPUT="SC2086 simulated finding" \
  FAKE_SHELLCHECK_STATUS=1 \
    "$SCRIPT" 2>&1
)"
warning_status=$?
set -e
if [[ "$warning_status" -ne 0 ]]; then
  fail "warning mode: expected success, got status $warning_status output: $warning_output"
fi
assert_contains "$warning_output" "SC2086 simulated finding" "warning mode finding output"
assert_contains "$warning_output" "lint-shell warning-only" "warning mode diagnostic"

set +e
strict_output="$(
  CARTULARY_SHELLCHECK_ROOT="$warning_repo" \
  SHELLCHECK_BIN="$warning_fake_shellcheck" \
  LINT_SHELL_STRICT=1 \
  FAKE_SHELLCHECK_ARGS_LOG="$warning_args_log" \
  FAKE_SHELLCHECK_OUTPUT="SC2086 simulated finding" \
  FAKE_SHELLCHECK_STATUS=1 \
    "$SCRIPT" 2>&1
)"
strict_status=$?
set -e
if [[ "$strict_status" -eq 0 ]]; then
  fail "strict mode: expected failure"
fi
assert_contains "$strict_output" "SC2086 simulated finding" "strict mode finding output"

cache_repo="$(make_fixture_repo)"
mkdir -p "$cache_repo/scripts"
printf '%s\n' '#!/usr/bin/env bash' 'echo ok' >"$cache_repo/scripts/good.sh"
track_all "$cache_repo"
cache_fake_shellcheck="$(make_counting_shellcheck "$cache_repo")"
cache_args_log="$cache_repo/shellcheck-args.log"
cache_run_log="$cache_repo/shellcheck-run.log"
cache_dir="$cache_repo/static-cache"

run_cached_shellcheck() {
  CARTULARY_SHELLCHECK_ROOT="$cache_repo" \
  CARTULARY_STATIC_ANALYSIS_CACHE_DIR="$cache_dir" \
  SHELLCHECK_BIN="$cache_fake_shellcheck" \
  LINT_SHELL_STRICT=1 \
  FAKE_SHELLCHECK_ARGS_LOG="$cache_args_log" \
  FAKE_SHELLCHECK_RUN_LOG="$cache_run_log" \
    "$SCRIPT" 2>&1
}

cache_first_output="$(run_cached_shellcheck)"
assert_contains "$cache_first_output" "1 files checked" "lint-shell cache first run output"
assert_equals "$(count_shellcheck_runs "$cache_run_log")" "1" "lint-shell cache first run executes shellcheck"

cache_second_output="$(run_cached_shellcheck)"
assert_contains "$cache_second_output" "1 files checked" "lint-shell cache hit output"
assert_equals "$(count_shellcheck_runs "$cache_run_log")" "1" "lint-shell cache hit skips shellcheck"

printf '%s\n' 'echo changed' >>"$cache_repo/scripts/good.sh"
cache_input_change_output="$(run_cached_shellcheck)"
assert_contains "$cache_input_change_output" "1 files checked" "lint-shell cache input-change output"
assert_equals "$(count_shellcheck_runs "$cache_run_log")" "2" "lint-shell cache input change executes shellcheck"

mapfile -t cache_records < <(find "$cache_dir" -type f -name '*.json' | sort)
if [[ "${#cache_records[@]}" -eq 0 ]]; then
  fail "lint-shell cache corrupt fixture: expected cache record"
fi
for cache_record in "${cache_records[@]}"; do
  printf '{bad json\n' >"$cache_record"
done
cache_corrupt_output="$(run_cached_shellcheck)"
assert_contains "$cache_corrupt_output" "1 files checked" "lint-shell cache corrupt-record output"
assert_equals "$(count_shellcheck_runs "$cache_run_log")" "3" "lint-shell cache corrupt record executes shellcheck"

rm -f "$cache_dir/outputs/lint-shell.ok"
cache_missing_output="$(run_cached_shellcheck)"
assert_contains "$cache_missing_output" "1 files checked" "lint-shell cache missing-output output"
assert_equals "$(count_shellcheck_runs "$cache_run_log")" "4" "lint-shell cache missing output executes shellcheck"

make_strict_repo="$(make_fixture_repo)"
mkdir -p "$make_strict_repo/scripts"
printf '%s\n' "echo \"\$HOME\"" >"$make_strict_repo/scripts/warn.sh"
track_all "$make_strict_repo"
make_strict_shellcheck="$(make_fake_shellcheck "$make_strict_repo")"
make_strict_args_log="$make_strict_repo/shellcheck-args.log"
make_strict_results="$(cartulary_harness_mktemp_dir lint-shell-public-results.XXXXXX)"
cleanup_paths+=("$make_strict_results")
make_strict_run_id="lint-shell-public-strict"
set +e
make_strict_output="$(
  env -u CARTULARY_HARNESS_IDENTITY_PREPARED -u CARTULARY_TEST_TARGET \
    CARTULARY_OUTPUT_MODE=verbose \
    CARTULARY_TEST_RESULTS_DIR="$make_strict_results" \
    CARTULARY_TEST_RUN_ID="$make_strict_run_id" \
    CARTULARY_SHELLCHECK_ROOT="$make_strict_repo" \
    SHELLCHECK_BIN="$make_strict_shellcheck" \
    FAKE_SHELLCHECK_ARGS_LOG="$make_strict_args_log" \
    FAKE_SHELLCHECK_OUTPUT="SC2086 simulated finding" \
    FAKE_SHELLCHECK_STATUS=1 \
      make -C "$ROOT_DIR" --no-print-directory CARTULARY_HARNESS_CACHE_MODE=off lint-shell 2>&1
)"
make_strict_status=$?
set -e
if [[ "$make_strict_status" -eq 0 ]]; then
  fail "public Make lint-shell: expected strict ShellCheck failure"
fi
assert_not_contains "$make_strict_output" "lint-shell warning-only" "public Make lint-shell must not use warning-only mode"
if [[ ! -f "$make_strict_args_log" ]]; then
  fail "public Make lint-shell: fake ShellCheck was not invoked"
fi
assert_contains "$(cat "$make_strict_args_log")" "scripts/warn.sh" "public Make lint-shell fake invocation"
make_strict_manifest="$make_strict_results/$make_strict_run_id/run-manifest.json"
if [[ ! -f "$make_strict_manifest" ]]; then
  fail "public Make lint-shell: missing nested run manifest"
fi
assert_contains "$(cat "$make_strict_manifest")" '"cache_mode": "off"' "public Make lint-shell nested cache mode"

real_shellcheck=""
if [[ -n "${SHELLCHECK_BIN:-}" && -x "${SHELLCHECK_BIN}" ]]; then
  real_shellcheck="$SHELLCHECK_BIN"
elif [[ -x "$ROOT_DIR/tmp/toolbin/shellcheck-v0.11.0" ]]; then
  real_shellcheck="$ROOT_DIR/tmp/toolbin/shellcheck-v0.11.0"
elif command -v shellcheck >/dev/null 2>&1; then
  real_shellcheck="$(command -v shellcheck)"
fi

if [[ -n "$real_shellcheck" ]]; then
  real_repo="$(make_fixture_repo)"
  mkdir -p "$real_repo/scripts"
  printf '%s\n' "#!/usr/bin/env bash" "name=\"\${1:-world}\"" "echo \${name}" >"$real_repo/scripts/unsafe.sh"
  track_all "$real_repo"
  set +e
  real_output="$(
    CARTULARY_SHELLCHECK_ROOT="$real_repo" \
    SHELLCHECK_BIN="$real_shellcheck" \
    LINT_SHELL_STRICT=1 \
      "$SCRIPT" 2>&1
  )"
  real_status=$?
  set -e
  if [[ "$real_status" -eq 0 ]]; then
    fail "real ShellCheck fixture: expected strict failure"
  fi
  assert_contains "$real_output" "SC2086" "real ShellCheck unsafe fixture"
fi
