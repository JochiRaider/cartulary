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
assert_contains "$normal_output" "internal/app/server/runtime.go" "app facade discovery"

assert_not_contains "$normal_output" "AGENTS.md" "human guidance exclusion"
assert_not_contains "$normal_output" "README" "human reference exclusion"

if "$SCRIPT" cmd/server tmp/missing-build-root >"$tmp_dir/missing.stdout" 2>"$tmp_dir/missing.stderr"; then
  fail "missing root discovery unexpectedly succeeded"
fi
assert_contains "$(cat "$tmp_dir/missing.stderr")" "build input root is missing, empty or outside executable scope" "missing root diagnostic"
[[ ! -s "$tmp_dir/missing.stdout" ]] || fail "invalid scope emitted a partial input list"

fixture="$tmp_dir/repository"
mkdir -p "$fixture/src"
git -C "$fixture" init --quiet
printf 'package fixture\n' >"$fixture/src/source.go"
ln -s "$tmp_dir/not-admitted" "$fixture/src/README.md"
ln -s "$tmp_dir/not-admitted" "$fixture/src/guide.markdown"
mapfile -t restricted_roots < <("${NODE_BIN:?}" -e 'const fs = require("node:fs"); for (const root of JSON.parse(fs.readFileSync(process.argv[1], "utf8")).restricted_roots) console.log(root)' "$ROOT_DIR/tools/executable_input_policy.json")
[[ "${#restricted_roots[@]}" -gt 0 ]] || fail "restricted-root policy unavailable"
for restricted in "${restricted_roots[@]}"; do
  mkdir -p "$(dirname "$fixture/$restricted")"
  ln -s "$tmp_dir/not-admitted" "$fixture/$restricted"
done

cd "$fixture"
fixture_output="$("$SCRIPT" src)"
[[ "$fixture_output" == "src/source.go" ]] || fail "non-executable input entered the build list"
for restricted in "${restricted_roots[@]}"; do
  if "$SCRIPT" "$restricted" >"$tmp_dir/restricted.stdout" 2>"$tmp_dir/restricted.stderr"; then
    fail "restricted input root was admitted"
  fi
  assert_contains "$(cat "$tmp_dir/restricted.stderr")" "outside executable scope" "restricted root rejected without following its link"
done
ln -s "$tmp_dir/not-admitted" src/escape
if "$SCRIPT" src >"$tmp_dir/link.stdout" 2>"$tmp_dir/link.stderr"; then
  fail "executable symbolic link was admitted"
fi
assert_contains "$(cat "$tmp_dir/link.stderr")" "symbolic link" "link admission"
rm src/escape
printf 'package fixture\n' >"src/has space.go"
if "$SCRIPT" src >"$tmp_dir/space.stdout" 2>"$tmp_dir/space.stderr"; then
  fail "Make-incompatible filename was silently split"
fi
assert_contains "$(cat "$tmp_dir/space.stderr")" "contains whitespace" "path admission"
rm "src/has space.go"

fake_bin="$tmp_dir/bin"
mkdir -p "$fake_bin"
cat >"$fake_bin/git" <<'EOF'
#!/usr/bin/env bash
exit 97
EOF
chmod +x "$fake_bin/git"
if PATH="$fake_bin:$PATH" "$SCRIPT" src >"$tmp_dir/git.stdout" 2>"$tmp_dir/git.stderr"; then
  fail "failed Git discovery was admitted"
fi
assert_contains "$(cat "$tmp_dir/git.stderr")" "cannot enumerate admitted repository inputs" "discovery failure"
[[ ! -s "$tmp_dir/git.stdout" ]] || fail "failed discovery emitted cache inputs"
