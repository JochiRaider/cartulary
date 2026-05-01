#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
SCRIPT="$ROOT_DIR/scripts/check-generate-drift.sh"
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

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: expected output not to contain [$needle]"
  fi
}

missing_sqlc="$ROOT_DIR/tmp/generate-drift-missing-sqlc"
rm -f "$missing_sqlc"

set +e
output="$(SQLC_BIN="$missing_sqlc" "$SCRIPT" 2>&1)"
status=$?
set -e

if [[ "$status" -eq 0 ]]; then
  fail "missing SQLC_BIN: expected generate-drift failure"
fi

assert_contains "$output" "generate-drift requires an executable SQLC_BIN at $missing_sqlc" "missing SQLC_BIN diagnostic"
assert_contains "$output" "run make codegen-toolchain before generate-drift" "missing SQLC_BIN setup guidance"
assert_not_contains "$output" "bootstrap sqlc tool" "missing SQLC_BIN bootstrap avoidance"
assert_not_contains "$output" "generate sqlc" "missing SQLC_BIN generation avoidance"

scratch_tool_dir="$(mktemp -d "$ROOT_DIR/tmp/generate-drift-smoke.XXXXXX")"
cleanup_paths+=("$scratch_tool_dir")
fake_sqlc="$scratch_tool_dir/sqlc"
fake_go="$scratch_tool_dir/go"

cat >"$fake_sqlc" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" != "generate" ]]; then
  echo "unexpected sqlc invocation: $*" >&2
  exit 2
fi
EOF

cat >"$fake_go" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" == "run" && "${2:-}" == "./tools/contractgen" ]]; then
  exit 0
fi

echo "unexpected go invocation: $*" >&2
exit 2
EOF

chmod +x "$fake_sqlc" "$fake_go"

set +e
output="$(
  SQLC_BIN="$fake_sqlc" \
  GO="$fake_go" \
  GO_CACHE_DIR="$scratch_tool_dir/go-cache" \
  GO_MOD_CACHE_DIR="$scratch_tool_dir/go-mod" \
    "$SCRIPT" 2>&1
)"
status=$?
set -e

if [[ "$status" -ne 0 ]]; then
  fail "scratch Make include smoke: expected generate-drift success with fake generators, got output: $output"
fi

assert_not_contains \
  "$output" \
  "No rule to make target 'tools/task_surface.generated.mk'" \
  "scratch Make include smoke"
