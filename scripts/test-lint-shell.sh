#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
SCRIPT="${ROOT_DIR}/scripts/run-shellcheck.sh"
cleanup_paths=()

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
  dir="$(mktemp -d "$ROOT_DIR/tmp/lint-shell-fixture.XXXXXX")"
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
  local fake="$dir/fake-shellcheck"

  cat >"$fake" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$@" >"${FAKE_SHELLCHECK_ARGS_LOG:?}"
if [[ -n "${FAKE_SHELLCHECK_OUTPUT:-}" ]]; then
  printf '%s\n' "$FAKE_SHELLCHECK_OUTPUT" >&2
fi
exit "${FAKE_SHELLCHECK_STATUS:-0}"
EOF
  chmod +x "$fake"
  printf '%s\n' "$fake"
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
  SHELLCHECK_BIN="$fake_shellcheck" \
  FAKE_SHELLCHECK_ARGS_LOG="$args_log" \
    "$SCRIPT" 2>&1
)"
expected_inventory_output=$'bin/shebang-runner\ngenerated/skip.sh\nscripts/a.sh\nscripts/z.sh\n4 files checked'
assert_equals "$inventory_output" "$expected_inventory_output" "deterministic lint-shell output"
expected_args=$'bin/shebang-runner\ngenerated/skip.sh\nscripts/a.sh\nscripts/z.sh'
assert_equals "$(cat "$args_log")" "$expected_args" "deterministic shellcheck argv"

empty_repo="$(make_fixture_repo)"
mkdir -p "$empty_repo/scripts"
printf '%s\n' '#!/usr/bin/env python3' 'print("not shell")' >"$empty_repo/scripts/not-shell"
track_all "$empty_repo"
empty_output="$(
  CARTULARY_SHELLCHECK_ROOT="$empty_repo" \
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
