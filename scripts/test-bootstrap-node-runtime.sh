#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
SCRIPT="${ROOT_DIR}/tools/harness/readiness/bootstrap-node-runtime.sh"
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

assert_file_exists() {
  local file="$1"
  local label="$2"

  if [[ ! -e "${file}" ]]; then
    fail "${label}: expected ${file} to exist"
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

make_fake_sha256sum() {
  local dir="$1"

  cat >"${dir}/sha256sum" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '472655581fb851559730c48763e0c9d3bc25975c59d518003fc0849d3e4ba0f6  %s\n' "$1"
EOF
  chmod +x "${dir}/sha256sum"
}

make_fake_node_archive() {
  local dir="$1"
  local archive="$2"
  local root="${dir}/node-v24.15.0-linux-x64"

  mkdir -p "${root}/bin"
  cat >"${root}/bin/node" <<'EOF'
#!/usr/bin/env bash
if [[ "${1:-}" == "-v" || "${1:-}" == "--version" ]]; then
  echo "v24.15.0"
else
  exit 0
fi
EOF
  chmod +x "${root}/bin/node"
  tar -cJf "${archive}" -C "${dir}" "node-v24.15.0-linux-x64"
}

make_fake_curl() {
  local dir="$1"
  local mode="$2"

  cat >"${dir}/curl" <<EOF
#!/usr/bin/env bash
set -euo pipefail
echo "curl invoked" >>"${dir}/curl.log"
count_file="${dir}/curl.count"
count=0
if [[ -f "\${count_file}" ]]; then
  count="\$(cat "\${count_file}")"
fi
count=\$((count + 1))
printf '%s\n' "\${count}" >"\${count_file}"
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
case "${mode}" in
  fail)
    echo "curl should not have been called" >&2
    exit 99
    ;;
  always-fail-download)
    echo "fake transient download failure" >&2
    exit 56
    ;;
  bad-archive)
    printf 'not a node archive' >"\${out}"
    ;;
  flaky-archive)
    if [[ "\${count}" -lt 2 ]]; then
      echo "fake transient download failure" >&2
      exit 56
    fi
    cp "\${FAKE_NODE_ARCHIVE:?FAKE_NODE_ARCHIVE is required}" "\${out}"
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

retry_dir="$(mktemp -d "${ROOT_DIR}/tmp/bootstrap-node-retry.XXXXXX")"
cleanup_paths+=("${retry_dir}")
valid_archive="${retry_dir}/valid-node.tar.xz"
make_fake_node_archive "${retry_dir}" "${valid_archive}"
make_fake_sha256sum "${retry_dir}"
make_fake_curl "${retry_dir}" flaky-archive
PATH="${retry_dir}:${PATH}" \
NODE_RUNTIME_DIR="${retry_dir}/runtime" \
CARTULARY_NODE_ARCHIVE_DIR="${retry_dir}/archives" \
CARTULARY_BOOTSTRAP_NODE_PLATFORM="linux-x64" \
CARTULARY_BOOTSTRAP_DOWNLOAD_RETRIES=3 \
CARTULARY_BOOTSTRAP_DOWNLOAD_RETRY_DELAY_SECONDS=0 \
FAKE_NODE_ARCHIVE="${valid_archive}" \
  "${SCRIPT}"
assert_file_exists "${retry_dir}/runtime/bin/node" "retry success installs runtime"
assert_equals "$("${retry_dir}/runtime/bin/node" -v)" "v24.15.0" "retry success node version"
assert_equals "$(cat "${retry_dir}/curl.count")" "2" "retry success curl attempts"

download_fail_dir="$(mktemp -d "${ROOT_DIR}/tmp/bootstrap-node-download-fail.XXXXXX")"
cleanup_paths+=("${download_fail_dir}")
make_fake_sha256sum "${download_fail_dir}"
make_fake_curl "${download_fail_dir}" always-fail-download
set +e
download_fail_output="$(
  PATH="${download_fail_dir}:${PATH}" \
  NODE_RUNTIME_DIR="${download_fail_dir}/runtime" \
  CARTULARY_NODE_ARCHIVE_DIR="${download_fail_dir}/archives" \
  CARTULARY_BOOTSTRAP_NODE_PLATFORM="linux-x64" \
  CARTULARY_BOOTSTRAP_DOWNLOAD_RETRIES=2 \
  CARTULARY_BOOTSTRAP_DOWNLOAD_RETRY_DELAY_SECONDS=0 \
    "${SCRIPT}" 2>&1
)"
download_fail_status=$?
set -e
if [[ "${download_fail_status}" -eq 0 ]]; then
  fail "download retry exhaustion: expected failure"
fi
assert_contains "${download_fail_output}" "Node runtime bootstrap failed: unable to download" "download failure diagnostic"
assert_equals "$(cat "${download_fail_dir}/curl.count")" "2" "download failure retry count"
assert_file_absent "${download_fail_dir}/archives/node-v24.15.0-linux-x64.tar.xz" "download failure archive cleanup"
assert_file_absent "${download_fail_dir}/runtime/bin/node" "download failure extraction guard"

concurrent_dir="$(mktemp -d "${ROOT_DIR}/tmp/bootstrap-node-concurrent.XXXXXX")"
cleanup_paths+=("${concurrent_dir}")
concurrent_archive="${concurrent_dir}/valid-node.tar.xz"
make_fake_node_archive "${concurrent_dir}" "${concurrent_archive}"
make_fake_sha256sum "${concurrent_dir}"
make_fake_curl "${concurrent_dir}" flaky-archive
PATH="${concurrent_dir}:${PATH}" \
NODE_RUNTIME_DIR="${concurrent_dir}/runtime" \
CARTULARY_NODE_ARCHIVE_DIR="${concurrent_dir}/archives" \
CARTULARY_BOOTSTRAP_NODE_PLATFORM="linux-x64" \
CARTULARY_BOOTSTRAP_DOWNLOAD_RETRIES=3 \
CARTULARY_BOOTSTRAP_DOWNLOAD_RETRY_DELAY_SECONDS=0 \
FAKE_NODE_ARCHIVE="${concurrent_archive}" \
  "${SCRIPT}" &
first_pid=$!
PATH="${concurrent_dir}:${PATH}" \
NODE_RUNTIME_DIR="${concurrent_dir}/runtime" \
CARTULARY_NODE_ARCHIVE_DIR="${concurrent_dir}/archives" \
CARTULARY_BOOTSTRAP_NODE_PLATFORM="linux-x64" \
CARTULARY_BOOTSTRAP_DOWNLOAD_RETRIES=3 \
CARTULARY_BOOTSTRAP_DOWNLOAD_RETRY_DELAY_SECONDS=0 \
FAKE_NODE_ARCHIVE="${concurrent_archive}" \
  "${SCRIPT}" &
second_pid=$!
wait "${first_pid}"
wait "${second_pid}"
assert_file_exists "${concurrent_dir}/runtime/bin/node" "concurrent install runtime"
assert_equals "$("${concurrent_dir}/runtime/bin/node" -v)" "v24.15.0" "concurrent install node version"
assert_file_absent "${concurrent_dir}/runtime.lock" "concurrent install lock cleanup"

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
