#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
SCRIPT="$ROOT_DIR/tools/harness/readiness/list-build-inputs.sh"
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

mkdir -p "$ROOT_DIR/tmp"
tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/build-input-discovery.XXXXXX")"
cleanup_paths+=("$tmp_dir")

cd "$ROOT_DIR"

normal_output="$("$SCRIPT" cmd/server internal/app)"
assert_contains "$normal_output" "cmd/server/main.go" "server discovery"
assert_contains "$normal_output" "internal/app/runtime.go" "app discovery"

if "$SCRIPT" cmd/server "$tmp_dir/missing-root" >"$tmp_dir/missing.stdout" 2>"$tmp_dir/missing.stderr"; then
  fail "missing root discovery unexpectedly succeeded"
fi
assert_contains "$(cat "$tmp_dir/missing.stderr")" "missing build input root: $tmp_dir/missing-root" "missing root diagnostic"

fake_bin="$tmp_dir/bin"
mkdir -p "$fake_bin"
cat >"$fake_bin/rg" <<'EOF'
#!/usr/bin/env bash
echo "fake rg failed" >&2
exit 97
EOF
chmod +x "$fake_bin/rg"
touch "$tmp_dir/stale-server"

if PATH="$fake_bin:$PATH" make --no-print-directory -n build-server SERVER_BIN="$tmp_dir/stale-server" >"$tmp_dir/make.stdout" 2>"$tmp_dir/make.stderr"; then
  fail "make build-server dry-run unexpectedly succeeded with broken rg"
fi
make_failure="$(cat "$tmp_dir/make.stdout" "$tmp_dir/make.stderr")"
assert_contains "$make_failure" "fake rg failed" "broken rg diagnostic"
assert_contains "$make_failure" "build input discovery failed" "make discovery failure"
assert_not_contains "$make_failure" "up to date" "stale binary reuse"
