#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
SCRIPT="${ROOT_DIR}/scripts/bootstrap-node-runtime.sh"
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

assert_file_absent() {
  local file="$1"
  local label="$2"

  if [[ -e "${file}" ]]; then
    fail "${label}: expected ${file} to be absent"
  fi
}

make_fake_curl() {
  local dir="$1"
  local mode="$2"

  cat >"${dir}/curl" <<EOF
#!/usr/bin/env bash
set -euo pipefail
echo "curl invoked" >>"${dir}/curl.log"
case "${mode}" in
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
    printf 'not a node archive' >"\${out}"
    ;;
esac
EOF
  chmod +x "${dir}/curl"
}

unsupported_dir="$(mktemp -d "${ROOT_DIR}/tmp/bootstrap-node-unsupported.XXXXXX")"
cleanup_paths+=("${unsupported_dir}")
make_fake_curl "${unsupported_dir}" fail
set +e
unsupported_output="$(
  PATH="${unsupported_dir}:${PATH}" \
  NODE_RUNTIME_DIR="${unsupported_dir}/runtime" \
  CARTULARY_NODE_ARCHIVE_DIR="${unsupported_dir}/archives" \
  CARTULARY_BOOTSTRAP_NODE_PLATFORM="sunos-sparc" \
    "${SCRIPT}" 2>&1
)"
unsupported_status=$?
set -e
if [[ "${unsupported_status}" -eq 0 ]]; then
  fail "unsupported platform: expected failure"
fi
assert_contains "${unsupported_output}" "unsupported Node bootstrap platform" "unsupported platform diagnostic"
assert_file_absent "${unsupported_dir}/curl.log" "unsupported platform curl guard"

mismatch_dir="$(mktemp -d "${ROOT_DIR}/tmp/bootstrap-node-mismatch.XXXXXX")"
cleanup_paths+=("${mismatch_dir}")
make_fake_curl "${mismatch_dir}" bad-archive
set +e
mismatch_output="$(
  PATH="${mismatch_dir}:${PATH}" \
  NODE_RUNTIME_DIR="${mismatch_dir}/runtime" \
  CARTULARY_NODE_ARCHIVE_DIR="${mismatch_dir}/archives" \
  CARTULARY_BOOTSTRAP_NODE_PLATFORM="linux-x64" \
    "${SCRIPT}" 2>&1
)"
mismatch_status=$?
set -e
if [[ "${mismatch_status}" -eq 0 ]]; then
  fail "checksum mismatch: expected failure"
fi
assert_contains "${mismatch_output}" "checksum mismatch" "checksum mismatch diagnostic"
assert_file_absent "${mismatch_dir}/archives/node-v24.15.0-linux-x64.tar.xz" "checksum mismatch archive cleanup"
assert_file_absent "${mismatch_dir}/runtime/bin/node" "checksum mismatch extraction guard"

reuse_dir="$(mktemp -d "${ROOT_DIR}/tmp/bootstrap-node-reuse.XXXXXX")"
cleanup_paths+=("${reuse_dir}")
mkdir -p "${reuse_dir}/runtime/bin"
cat >"${reuse_dir}/runtime/bin/node" <<'EOF'
#!/usr/bin/env bash
if [[ "${1:-}" == "-v" || "${1:-}" == "--version" ]]; then
  echo "v24.15.0"
else
  exit 0
fi
EOF
chmod +x "${reuse_dir}/runtime/bin/node"
make_fake_curl "${reuse_dir}" fail
PATH="${reuse_dir}:${PATH}" \
NODE_RUNTIME_DIR="${reuse_dir}/runtime" \
CARTULARY_NODE_ARCHIVE_DIR="${reuse_dir}/archives" \
CARTULARY_BOOTSTRAP_NODE_PLATFORM="linux-x64" \
  "${SCRIPT}"
assert_file_absent "${reuse_dir}/curl.log" "existing runtime reuse curl guard"
