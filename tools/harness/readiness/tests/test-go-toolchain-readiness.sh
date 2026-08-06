#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
READINESS_SCRIPT="${ROOT_DIR}/tools/harness/readiness/go-toolchain-readiness.sh"
BOOTSTRAP_SCRIPT="${ROOT_DIR}/tools/harness/readiness/bootstrap-go-tool.sh"
# shellcheck source=tools/harness/test-support/harness-scratch.sh
source "${ROOT_DIR}/tools/harness/test-support/harness-scratch.sh"

scratch="$(cartulary_harness_mktemp_dir "go-toolchain-readiness.XXXXXX")"
trap 'rm -rf "$scratch"' EXIT

fail() {
  echo "$*" >&2
  exit 1
}

assert_contains() {
  local actual="$1"
  local expected="$2"
  local label="$3"
  [[ "$actual" == *"$expected"* ]] || fail "$label: expected output to contain [$expected], got [$actual]"
}

assert_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"
  [[ "$actual" == "$expected" ]] || fail "$label: expected [$expected], got [$actual]"
}

assert_file_contents() {
  local file="$1"
  local expected="$2"
  local label="$3"
  [[ -f "$file" ]] || fail "$label: missing $file"
  assert_equals "$(cat "$file")" "$expected" "$label"
}

make_fake_go() {
  local dir="$1"
  local local_version="$2"
  local selected_version="$3"
  local selected_status="$4"
  local install_status="${5:-$selected_status}"

  mkdir -p "$dir"
  cat >"$dir/go" <<EOF
#!/usr/bin/env bash
set -euo pipefail
if [[ "\${1:-}" == "env" && "\${2:-}" == "GOOS" && "\${3:-}" == "GOARCH" ]]; then
  printf 'linux\namd64\n'
  exit 0
fi
if [[ "\${1:-}" == "version" ]]; then
  if [[ "\${GOTOOLCHAIN:-}" == "local" ]]; then
    printf 'go version ${local_version} linux/amd64\n'
    exit 0
  fi
  printf 'selected:%s\n' "\${GOTOOLCHAIN:-}" >>"${dir}/selected.log"
  printf 'GOCACHE=%s\nGOMODCACHE=%s\nGOTMPDIR=%s\n' \
    "\${GOCACHE:-}" "\${GOMODCACHE:-}" "\${GOTMPDIR:-}" >>"${dir}/machine-state.log"
  if [[ "${selected_status}" != "0" ]]; then
    printf 'simulated download failure\n' >&2
    exit "${selected_status}"
  fi
  printf 'go version ${selected_version} linux/amd64\n'
  exit 0
fi
if [[ "\${1:-}" == "install" ]]; then
  printf 'GOCACHE=%s\nGOMODCACHE=%s\nGOTMPDIR=%s\n' \
    "\${GOCACHE:-}" "\${GOMODCACHE:-}" "\${GOTMPDIR:-}" >>"${dir}/machine-state.log"
  if [[ "${install_status}" != "0" ]]; then
    printf 'simulated install failure\n' >&2
    exit "${install_status}"
  fi
  mkdir -p "\${GOBIN:?}"
  printf '%s\n' "\${FAKE_TOOL_CONTENT:-${selected_version}}" >"\${GOBIN}/fake-tool"
  chmod +x "\${GOBIN}/fake-tool"
  exit 0
fi
printf 'unsupported fake go invocation: %s\n' "\$*" >&2
exit 99
EOF
  chmod +x "$dir/go"
}

run_readiness() {
  local mode="$1"
  local fake_go="$2"
  local cache_root="$3"
  GO="$fake_go" \
  GO_TOOLCHAIN=go1.26.5 \
  GOTOOLCHAIN=go1.99.9 \
  GO_CACHE_DIR="$cache_root/build" \
  GO_MOD_CACHE_DIR="$cache_root/mod" \
  GO_TMP_DIR="$cache_root/tmp" \
    "$READINESS_SCRIPT" "$mode"
}

exact_dir="$scratch/exact"
make_fake_go "$exact_dir/bin" go1.26.5 go1.26.5 0
exact_output="$(run_readiness diagnose "$exact_dir/bin/go" "$exact_dir/cache")"
assert_contains "$exact_output" "effective=go1.26.5 source=local" "exact local toolchain"
[[ ! -e "$exact_dir/cache" ]] || fail "diagnose created cache state for exact local toolchain"

missing_dir="$scratch/missing"
make_fake_go "$missing_dir/bin" go1.26.4 go1.26.5 0
set +e
missing_output="$(run_readiness diagnose "$missing_dir/bin/go" "$missing_dir/cache" 2>&1)"
missing_status=$?
set -e
assert_equals "$missing_status" "2" "missing cached toolchain status"
assert_contains "$missing_output" "run make bootstrap" "missing cached toolchain diagnostic"
[[ ! -e "$missing_dir/cache" ]] || fail "diagnose created missing cache state"

set +e
doctor_output="$({
  GO="$missing_dir/bin/go" \
  GO_TOOLCHAIN=go1.26.5 \
  GO_CACHE_DIR="$missing_dir/cache/build" \
  GO_MOD_CACHE_DIR="$missing_dir/cache/mod" \
  GO_TMP_DIR="$missing_dir/cache/tmp" \
  NODE_BIN="$missing_dir/no-node" \
  PNPM="$missing_dir/no-pnpm" \
  SHELLCHECK_BIN="$missing_dir/no-shellcheck" \
    "$ROOT_DIR/tools/harness/readiness/check-doctor.sh"
} 2>&1)"
doctor_status=$?
set -e
assert_equals "$doctor_status" "2" "doctor aggregate failure status"
assert_contains "$doctor_output" "pinned effective toolchain go1.26.5 is not installed" "doctor Go diagnostic"
assert_contains "$doctor_output" "missing node" "doctor continues after Go diagnostic"
[[ ! -e "$missing_dir/cache" ]] || fail "doctor created cache state"
[[ ! -f "$missing_dir/bin/selected.log" ]] || fail "doctor invoked Go toolchain selection"

cached_dir="$scratch/cached"
make_fake_go "$cached_dir/bin" go1.26.4 go1.26.5 0
module_version="v0.0.1-go1.26.5.linux-amd64"
toolchain_dir="$cached_dir/cache/mod/golang.org/toolchain@$module_version"
ziphash_file="$cached_dir/cache/mod/cache/download/golang.org/toolchain/@v/$module_version.ziphash"
mkdir -p "$toolchain_dir/src" "$toolchain_dir/bin" "$(dirname "$ziphash_file")"
printf 'module std\n' >"$toolchain_dir/src/_go.mod"
cp "$toolchain_dir/src/_go.mod" "$toolchain_dir/src/go.mod"
make_fake_go "$toolchain_dir/bin" go1.26.5 go1.26.5 0
printf 'h1:fixture\n' >"$ziphash_file"
cached_output="$(run_readiness diagnose "$cached_dir/bin/go" "$cached_dir/cache")"
assert_contains "$cached_output" "source=automatic-cache" "valid cached toolchain"

corrupt_dir="$scratch/corrupt"
make_fake_go "$corrupt_dir/bin" go1.26.4 go1.26.5 0
corrupt_toolchain="$corrupt_dir/cache/mod/golang.org/toolchain@$module_version"
corrupt_ziphash="$corrupt_dir/cache/mod/cache/download/golang.org/toolchain/@v/$module_version.ziphash"
mkdir -p "$corrupt_toolchain/src" "$(dirname "$corrupt_ziphash")"
printf 'h1:fixture\n' >"$corrupt_ziphash"
before_tree="$(find "$corrupt_dir/cache" -printf '%P %y %s\n' | LC_ALL=C sort)"
set +e
corrupt_output="$(run_readiness ensure "$corrupt_dir/bin/go" "$corrupt_dir/cache" 2>&1)"
corrupt_status=$?
set -e
after_tree="$(find "$corrupt_dir/cache" -printf '%P %y %s\n' | LC_ALL=C sort)"
assert_equals "$corrupt_status" "2" "corrupt cache status"
assert_contains "$corrupt_output" "corrupt Go automatic-toolchain cache" "corrupt cache diagnostic"
assert_contains "$corrupt_output" "$corrupt_toolchain" "corrupt cache exact directory"
assert_contains "$corrupt_output" "$corrupt_ziphash" "corrupt cache exact ziphash"
assert_equals "$after_tree" "$before_tree" "corrupt cache remains unchanged"
[[ ! -f "$corrupt_dir/bin/selected.log" ]] || fail "corrupt cache invoked Go selection before failing"

ensure_dir="$scratch/ensure"
make_fake_go "$ensure_dir/bin" go1.26.4 go1.26.5 0
ensure_output="$(run_readiness ensure "$ensure_dir/bin/go" "$ensure_dir/cache")"
assert_contains "$ensure_output" "effective=go1.26.5 source=selected" "ensure selects exact toolchain"
assert_file_contents "$ensure_dir/bin/selected.log" "selected:go1.26.5" "ensure GOTOOLCHAIN pin"
ensure_machine_state="$(cat "$ensure_dir/bin/machine-state.log")"
assert_contains "$ensure_machine_state" "GOCACHE=$ensure_dir/cache/build" "ensure GOCACHE"
assert_contains "$ensure_machine_state" "GOMODCACHE=$ensure_dir/cache/mod" "ensure GOMODCACHE"
assert_contains "$ensure_machine_state" "GOTMPDIR=$ensure_dir/cache/tmp" "ensure GOTMPDIR"

mismatch_dir="$scratch/mismatch"
make_fake_go "$mismatch_dir/bin" go1.26.4 go1.26.6 0
set +e
mismatch_output="$(run_readiness ensure "$mismatch_dir/bin/go" "$mismatch_dir/cache" 2>&1)"
mismatch_status=$?
set -e
assert_equals "$mismatch_status" "2" "effective mismatch status"
assert_contains "$mismatch_output" "effective Go toolchain mismatch" "effective mismatch diagnostic"

download_dir="$scratch/download-failure"
make_fake_go "$download_dir/bin" go1.26.4 go1.26.5 41
set +e
download_output="$(run_readiness ensure "$download_dir/bin/go" "$download_dir/cache" 2>&1)"
download_status=$?
set -e
assert_equals "$download_status" "2" "download failure status"
assert_contains "$download_output" "Go toolchain readiness failed" "download failure diagnostic"

bootstrap_readiness_dir="$scratch/bootstrap-readiness-failure"
make_fake_go "$bootstrap_readiness_dir/bin" go1.26.4 go1.26.5 41
mkdir -p "$bootstrap_readiness_dir/toolbin"
printf 'old-tool\n' >"$bootstrap_readiness_dir/toolbin/fake-tool-v1"
set +e
bootstrap_readiness_output="$({
  GO="$bootstrap_readiness_dir/bin/go" \
  GO_TOOLCHAIN=go1.26.5 \
  TOOLBIN_DIR="$bootstrap_readiness_dir/toolbin" \
  TOOL_OUTPUT="$bootstrap_readiness_dir/toolbin/fake-tool-v1" \
  TOOL_MODULE=example.invalid/fake-tool@v1 \
  TOOL_BINARY_NAME=fake-tool \
  GO_CACHE_DIR="$bootstrap_readiness_dir/cache/build" \
  GO_MOD_CACHE_DIR="$bootstrap_readiness_dir/cache/mod" \
  GO_TMP_DIR="$bootstrap_readiness_dir/cache/tmp" \
  RUN_STEP_SCRIPT=/bin/false \
    "$BOOTSTRAP_SCRIPT"
} 2>&1)"
bootstrap_readiness_status=$?
set -e
assert_equals "$bootstrap_readiness_status" "2" "bootstrap readiness failure status"
assert_contains "$bootstrap_readiness_output" "Go toolchain readiness failed" "bootstrap readiness failure diagnostic"
assert_file_contents "$bootstrap_readiness_dir/toolbin/fake-tool-v1" "old-tool" "failed readiness preserves old output"

run_step="$scratch/run-step"
cat >"$run_step" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
shift
[[ "${1:-}" == "--" ]] || exit 99
shift
exec "$@"
EOF
chmod +x "$run_step"

bootstrap_failure_dir="$scratch/bootstrap-failure"
make_fake_go "$bootstrap_failure_dir/bin" go1.26.5 go1.26.5 0 41
mkdir -p "$bootstrap_failure_dir/toolbin"
printf 'old-tool\n' >"$bootstrap_failure_dir/toolbin/fake-tool-v1"
set +e
bootstrap_failure_output="$({
  GO="$bootstrap_failure_dir/bin/go" \
  GO_TOOLCHAIN=go1.26.5 \
  TOOLBIN_DIR="$bootstrap_failure_dir/toolbin" \
  TOOL_OUTPUT="$bootstrap_failure_dir/toolbin/fake-tool-v1" \
  TOOL_MODULE=example.invalid/fake-tool@v1 \
  TOOL_BINARY_NAME=fake-tool \
  TOOL_LABEL='bootstrap fake tool' \
  GO_CACHE_DIR="$bootstrap_failure_dir/cache/build" \
  GO_MOD_CACHE_DIR="$bootstrap_failure_dir/cache/mod" \
  GO_TMP_DIR="$bootstrap_failure_dir/cache/tmp" \
  RUN_STEP_SCRIPT="$run_step" \
    "$BOOTSTRAP_SCRIPT"
} 2>&1)"
bootstrap_failure_status=$?
set -e
assert_equals "$bootstrap_failure_status" "41" "bootstrap install failure status"
assert_contains "$bootstrap_failure_output" "simulated install failure" "bootstrap install failure diagnostic"
assert_file_contents "$bootstrap_failure_dir/toolbin/fake-tool-v1" "old-tool" "failed bootstrap preserves old output"
if find "$bootstrap_failure_dir/toolbin" -mindepth 1 -maxdepth 1 -type d -name '.fake-tool.install.*' | grep -q .; then
  fail "failed bootstrap left a staging directory"
fi

bootstrap_success_dir="$scratch/bootstrap-success"
make_fake_go "$bootstrap_success_dir/bin" go1.26.5 go1.26.5 0
mkdir -p "$bootstrap_success_dir/toolbin"
printf 'old-tool\n' >"$bootstrap_success_dir/toolbin/fake-tool-v1"
GO="$bootstrap_success_dir/bin/go" \
GO_TOOLCHAIN=go1.26.5 \
TOOLBIN_DIR="$bootstrap_success_dir/toolbin" \
TOOL_OUTPUT="$bootstrap_success_dir/toolbin/fake-tool-v1" \
TOOL_MODULE=example.invalid/fake-tool@v1 \
TOOL_BINARY_NAME=fake-tool \
TOOL_LABEL='bootstrap fake tool' \
GO_CACHE_DIR="$bootstrap_success_dir/cache/build" \
GO_MOD_CACHE_DIR="$bootstrap_success_dir/cache/mod" \
  GO_TMP_DIR="$bootstrap_success_dir/cache/tmp" \
RUN_STEP_SCRIPT="$run_step" \
  "$BOOTSTRAP_SCRIPT"
assert_file_contents "$bootstrap_success_dir/toolbin/fake-tool-v1" "go1.26.5" "successful bootstrap replaces output"
if find "$bootstrap_success_dir/toolbin" -mindepth 1 -maxdepth 1 -type d -name '.fake-tool.install.*' | grep -q .; then
  fail "successful bootstrap left a staging directory"
fi

concurrent_dir="$scratch/bootstrap-concurrent"
make_fake_go "$concurrent_dir/bin" go1.26.5 go1.26.5 0
mkdir -p "$concurrent_dir/toolbin"
run_concurrent_bootstrap() {
  local output="$1"
  local content="$2"
  GO="$concurrent_dir/bin/go" \
  GO_TOOLCHAIN=go1.26.5 \
  TOOLBIN_DIR="$concurrent_dir/toolbin" \
  TOOL_OUTPUT="$output" \
  TOOL_MODULE=example.invalid/fake-tool@v1 \
  TOOL_BINARY_NAME=fake-tool \
  GO_CACHE_DIR="$concurrent_dir/cache/build" \
  GO_MOD_CACHE_DIR="$concurrent_dir/cache/mod" \
  GO_TMP_DIR="$concurrent_dir/cache/tmp" \
  RUN_STEP_SCRIPT="$run_step" \
  FAKE_TOOL_CONTENT="$content" \
    "$BOOTSTRAP_SCRIPT" >/dev/null
}
run_concurrent_bootstrap "$concurrent_dir/toolbin/fake-tool-a-v1" "tool-a" &
concurrent_a_pid=$!
run_concurrent_bootstrap "$concurrent_dir/toolbin/fake-tool-b-v1" "tool-b" &
concurrent_b_pid=$!
wait "$concurrent_a_pid"
wait "$concurrent_b_pid"
assert_file_contents "$concurrent_dir/toolbin/fake-tool-a-v1" "tool-a" "concurrent bootstrap output A"
assert_file_contents "$concurrent_dir/toolbin/fake-tool-b-v1" "tool-b" "concurrent bootstrap output B"
if find "$concurrent_dir/toolbin" -mindepth 1 -maxdepth 1 -type d -name '.fake-tool.install.*' | grep -q .; then
  fail "concurrent bootstrap left a staging directory"
fi

printf 'go toolchain readiness checks passed\n'
"${ROOT_DIR}/tmp/node-runtime/bin/node" \
  "${ROOT_DIR}/tools/harness/contract/tests/test-machine-state-config.mjs"
