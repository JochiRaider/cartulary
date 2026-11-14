#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
SCRIPT="${ROOT_DIR}/tools/harness/readiness/bootstrap-shellcheck.sh"
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

assert_file_absent() {
  local file="$1"
  local label="$2"

  if [[ -e "$file" ]]; then
    fail "$label: expected $file to be absent"
  fi
}

make_fake_curl() {
  local dir="$1"
  local mode="$2"

  cat >"$dir/curl" <<EOF
#!/usr/bin/env bash
set -euo pipefail
echo "curl invoked" >>"$dir/curl.log"
case "$mode" in
  fail)
    echo "curl should not have been called" >&2
    exit 99
    ;;
  bad-archive)
    out=""
    while [[ "\$#" -gt 0 ]]; do
      case "\$1" in
        -o) out="\$2"; shift 2 ;;
        -*) shift ;;
        *) shift ;;
      esac
    done
    if [[ -z "\${out}" ]]; then
      echo "fake curl did not receive -o" >&2
      exit 2
    fi
    printf 'not a shellcheck archive' >"\${out}"
    ;;
esac
EOF
  chmod +x "$dir/curl"
}

unsupported_dir="$(mktemp -d "$ROOT_DIR/tmp/bootstrap-shellcheck-unsupported.XXXXXX")"
cleanup_paths+=("$unsupported_dir")
make_fake_curl "$unsupported_dir" fail
set +e
unsupported_output="$(
  PATH="$unsupported_dir:$PATH" \
  TOOLBIN_DIR="$unsupported_dir/toolbin" \
  SHELLCHECK_BIN="$unsupported_dir/toolbin/shellcheck-v0.11.0" \
  CARTULARY_SHELLCHECK_ARCHIVE_DIR="$unsupported_dir/archives" \
  CARTULARY_BOOTSTRAP_SHELLCHECK_PLATFORM="sunos-sparc" \
    "$SCRIPT" 2>&1
)"
unsupported_status=$?
set -e
if [[ "$unsupported_status" -eq 0 ]]; then
  fail "unsupported platform: expected failure"
fi
assert_contains "$unsupported_output" "unsupported ShellCheck bootstrap platform" "unsupported platform diagnostic"
assert_file_absent "$unsupported_dir/curl.log" "unsupported platform curl guard"

mismatch_dir="$(mktemp -d "$ROOT_DIR/tmp/bootstrap-shellcheck-mismatch.XXXXXX")"
cleanup_paths+=("$mismatch_dir")
make_fake_curl "$mismatch_dir" bad-archive
set +e
mismatch_output="$(
  PATH="$mismatch_dir:$PATH" \
  TOOLBIN_DIR="$mismatch_dir/toolbin" \
  SHELLCHECK_BIN="$mismatch_dir/toolbin/shellcheck-v0.11.0" \
  CARTULARY_SHELLCHECK_ARCHIVE_DIR="$mismatch_dir/archives" \
  CARTULARY_BOOTSTRAP_SHELLCHECK_PLATFORM="linux-x64" \
    "$SCRIPT" 2>&1
)"
mismatch_status=$?
set -e
if [[ "$mismatch_status" -eq 0 ]]; then
  fail "checksum mismatch: expected failure"
fi
assert_contains "$mismatch_output" "checksum mismatch" "checksum mismatch diagnostic"
assert_file_absent "$mismatch_dir/archives/shellcheck-v0.11.0.linux.x86_64.tar.xz" "checksum mismatch archive cleanup"
assert_file_absent "$mismatch_dir/toolbin/shellcheck-v0.11.0" "checksum mismatch extraction guard"

reuse_dir="$(mktemp -d "$ROOT_DIR/tmp/bootstrap-shellcheck-reuse.XXXXXX")"
cleanup_paths+=("$reuse_dir")
mkdir -p "$reuse_dir/toolbin"
cat >"$reuse_dir/toolbin/shellcheck-v0.11.0" <<'EOF'
#!/usr/bin/env bash
if [[ "${1:-}" == "--version" ]]; then
  printf '%s\n' \
    "ShellCheck - shell script analysis tool" \
    "version: 0.11.0" \
    "license: GNU General Public License, version 3"
else
  exit 0
fi
EOF
chmod +x "$reuse_dir/toolbin/shellcheck-v0.11.0"
make_fake_curl "$reuse_dir" fail
PATH="$reuse_dir:$PATH" \
TOOLBIN_DIR="$reuse_dir/toolbin" \
SHELLCHECK_BIN="$reuse_dir/toolbin/shellcheck-v0.11.0" \
CARTULARY_SHELLCHECK_ARCHIVE_DIR="$reuse_dir/archives" \
CARTULARY_BOOTSTRAP_SHELLCHECK_PLATFORM="linux-x64" \
  "$SCRIPT"
assert_file_absent "$reuse_dir/curl.log" "existing shellcheck reuse curl guard"

install_dir="$(mktemp -d "$ROOT_DIR/tmp/bootstrap-shellcheck-install.XXXXXX")"
cleanup_paths+=("$install_dir")
mkdir -p "$install_dir/archive-source/shellcheck-v0.11.0" "$install_dir/archives" "$install_dir/toolbin"
cp "$reuse_dir/toolbin/shellcheck-v0.11.0" "$install_dir/archive-source/shellcheck-v0.11.0/shellcheck"
tar -cJf "$install_dir/archives/shellcheck-v0.11.0.linux.x86_64.tar.xz" -C "$install_dir/archive-source" shellcheck-v0.11.0
cat >"$install_dir/sha256sum" <<'EOF'
#!/usr/bin/env bash
printf '8c3be12b05d5c177a04c29e3c78ce89ac86f1595681cab149b65b97c4e227198  %s\n' "$1"
EOF
chmod +x "$install_dir/sha256sum"
make_fake_curl "$install_dir" fail
run_install() {
  PATH="$install_dir:$PATH" \
  TOOLBIN_DIR="$install_dir/toolbin" \
  SHELLCHECK_BIN="$install_dir/toolbin/shellcheck-v0.11.0" \
  CARTULARY_SHELLCHECK_ARCHIVE_DIR="$install_dir/archives" \
  CARTULARY_BOOTSTRAP_SHELLCHECK_PLATFORM="linux-x64" \
  CARTULARY_FORCE_REINSTALL=1 \
    "$SCRIPT"
}
run_install
cmp "$reuse_dir/toolbin/shellcheck-v0.11.0" "$install_dir/toolbin/shellcheck-v0.11.0"
[[ -x "$install_dir/toolbin/shellcheck-v0.11.0" ]] || fail "installed tool must be executable"
for candidate in "$install_dir"/toolbin/shellcheck-install.*; do
  assert_file_absent "$candidate" "successful extraction staging cleanup"
done

cat >"$install_dir/tar" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
while [[ "$#" -gt 0 ]]; do
  if [[ "$1" == "-C" ]]; then
    printf 'partial extraction' >"$2/partial"
    exit 42
  fi
  shift
done
exit 99
EOF
chmod +x "$install_dir/tar"
extraction_status=0
run_install || extraction_status=$?
[[ "$extraction_status" == 42 ]] || fail "failed extraction: expected exit 42, got $extraction_status"
cmp "$reuse_dir/toolbin/shellcheck-v0.11.0" "$install_dir/toolbin/shellcheck-v0.11.0"
assert_file_absent "$install_dir/curl.log" "verified archive reuse"
for candidate in "$install_dir"/toolbin/shellcheck-install.*; do
  assert_file_absent "$candidate" "failed extraction staging cleanup"
done
