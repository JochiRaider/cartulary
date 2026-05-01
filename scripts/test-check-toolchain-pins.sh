#!/usr/bin/env bash
# Single-quoted literals below intentionally assert text containing shell/Markdown metacharacters.
# shellcheck disable=SC2016
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
if command -v "${NODE_BIN}" >/dev/null 2>&1; then
  NODE_BIN="$(command -v "${NODE_BIN}")"
fi
SCRIPT="${ROOT_DIR}/scripts/check-toolchain-pins.mjs"
source "${ROOT_DIR}/scripts/lib/harness-scratch.sh"
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

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "${haystack}" == *"${needle}"* ]]; then
    fail "${label}: expected output not to contain [${needle}]"
  fi
}

assert_harness_scratch_rejects_repo_tmp() {
  local output
  local status

  set +e
  output="$(
    CARTULARY_HARNESS_SCRATCH_ROOT="${ROOT_DIR}/tmp/toolchain-pins-invalid" \
      cartulary_harness_mktemp_dir "toolchain-pins-invalid.XXXXXX" 2>&1
  )"
  status=$?
  set -e

  if [[ "${status}" -eq 0 ]]; then
    fail "harness scratch root guard: expected repo-local scratch root failure"
  fi
  assert_contains "${output}" "CARTULARY_HARNESS_SCRATCH_ROOT must be outside the repository" "harness scratch root guard"
}

copy_minimal_repo() {
  local dest="$1"

  mkdir -p "${dest}/scripts" "${dest}/tools"
  cp "${ROOT_DIR}/Makefile" "${dest}/Makefile"
  cp "${ROOT_DIR}/package.json" "${dest}/package.json"
  cp "${ROOT_DIR}/go.mod" "${dest}/go.mod"
  cp "${ROOT_DIR}/README.md" "${dest}/README.md"
  cp "${ROOT_DIR}/AGENTS.md" "${dest}/AGENTS.md"
  cp "${ROOT_DIR}/tools/task_surface.generated.mk" "${dest}/tools/task_surface.generated.mk"
  cp "${ROOT_DIR}/tools/task_surface_manifest.json" "${dest}/tools/task_surface_manifest.json"
  cp "${ROOT_DIR}/scripts/list-build-inputs.sh" "${dest}/scripts/list-build-inputs.sh"
  cp "${ROOT_DIR}/scripts/bootstrap-node-runtime.sh" "${dest}/scripts/bootstrap-node-runtime.sh"
  cp "${ROOT_DIR}/scripts/bootstrap-shellcheck.sh" "${dest}/scripts/bootstrap-shellcheck.sh"
  cp "${ROOT_DIR}/scripts/check-toolchain-pins.mjs" "${dest}/scripts/check-toolchain-pins.mjs"
  mkdir -p \
    "${dest}/apps/web" \
    "${dest}/cmd/migrate" \
    "${dest}/cmd/server" \
    "${dest}/contracts" \
    "${dest}/db/migrations" \
    "${dest}/internal/app" \
    "${dest}/internal/modules" \
    "${dest}/internal/platform" \
    "${dest}/internal/platform/postgres" \
    "${dest}/internal/testutil/pgtest" \
    "${dest}/internal/testutil/s3test" \
    "${dest}/internal/testutil/suiteservices" \
    "${dest}/packages" \
    "${dest}/tools/testservices"
  local root
  for root in \
    apps/web \
    cmd/migrate \
    cmd/server \
    contracts \
    db/migrations \
    internal/app \
    internal/modules \
    internal/platform \
    internal/platform/postgres \
    internal/testutil/pgtest \
    internal/testutil/s3test \
    internal/testutil/suiteservices \
    packages \
    tools/testservices; do
    printf 'placeholder\n' >"${dest}/${root}/placeholder.txt"
  done
}

replace_text() {
  local file="$1"
  local search="$2"
  local replacement="$3"

  "$NODE_BIN" - "$file" "$search" "$replacement" <<'EOF'
const fs = require("node:fs");

const [file, search, replacement] = process.argv.slice(2);
const before = fs.readFileSync(file, "utf8");
if (!before.includes(search)) {
  process.stderr.write(`missing search text: ${search}\n`);
  process.exit(2);
}
fs.writeFileSync(file, before.replace(search, replacement));
EOF
}

expect_drift() {
  local label="$1"
  local expected_output="$2"
  local mutate="$3"

  local repo_dir
  repo_dir="$(cartulary_harness_mktemp_dir "toolchain-pins-${label}.XXXXXX")"
  cleanup_paths+=("${repo_dir}")
  copy_minimal_repo "${repo_dir}"
  "${mutate}" "${repo_dir}"

  set +e
  local output
  output="$("$NODE_BIN" "$SCRIPT" --root "${repo_dir}" 2>&1)"
  local status=$?
  set -e

  if [[ "${status}" -eq 0 ]]; then
    fail "${label}: expected drift failure"
  fi
  assert_contains "${output}" "${expected_output}" "${label} diagnostic"
}

mutate_package_node() {
  replace_text "$1/package.json" '"node": "24.15.0"' '"node": "24.16.0"'
}

mutate_package_manager() {
  replace_text "$1/package.json" '"packageManager": "pnpm@10.33.0"' '"packageManager": "pnpm@10.34.0"'
}

mutate_go_toolchain() {
  replace_text "$1/go.mod" 'toolchain go1.26.2' 'toolchain go1.26.3'
}

mutate_go_testcontainers() {
  replace_text "$1/go.mod" 'github.com/testcontainers/testcontainers-go v0.42.0' 'github.com/testcontainers/testcontainers-go v0.43.0'
}

mutate_staticcheck_tool() {
  replace_text "$1/Makefile" 'STATICCHECK_TOOL := honnef.co/go/tools/cmd/staticcheck@v0.7.0' 'STATICCHECK_TOOL := honnef.co/go/tools/cmd/staticcheck@v0.7.1'
}

mutate_govulncheck_tool() {
  replace_text "$1/Makefile" 'GOVULNCHECK_TOOL := golang.org/x/vuln/cmd/govulncheck@v1.3.0' 'GOVULNCHECK_TOOL := golang.org/x/vuln/cmd/govulncheck@v1.3.1'
}

mutate_gosec_tool() {
  replace_text "$1/Makefile" 'GOSEC_TOOL := github.com/securego/gosec/v2/cmd/gosec@v2.26.1' 'GOSEC_TOOL := github.com/securego/gosec/v2/cmd/gosec@v2.26.2'
}

mutate_shellcheck_version() {
  replace_text "$1/Makefile" 'SHELLCHECK_VERSION ?= 0.11.0' 'SHELLCHECK_VERSION ?= 0.11.1'
}

mutate_bootstrap_shellcheck_version() {
  replace_text "$1/scripts/bootstrap-shellcheck.sh" 'SHELLCHECK_VERSION="${SHELLCHECK_VERSION:-0.11.0}"' 'SHELLCHECK_VERSION="${SHELLCHECK_VERSION:-0.11.1}"'
}

mutate_readme_node() {
  replace_text "$1/README.md" '- Node.js `24.15.0`' '- Node.js `24.16.0`'
}

mutate_readme_staticcheck() {
  replace_text "$1/README.md" '- Staticcheck `v0.7.0`' '- Staticcheck `v0.7.1`'
}

mutate_readme_govulncheck() {
  replace_text "$1/README.md" '- Govulncheck `v1.3.0`' '- Govulncheck `v1.3.1`'
}

mutate_readme_gosec() {
  replace_text "$1/README.md" '- Gosec `v2.26.1`' '- Gosec `v2.26.2`'
}

mutate_readme_shellcheck() {
  replace_text "$1/README.md" '- ShellCheck `0.11.0`' '- ShellCheck `0.11.1`'
}

mutate_agents_govulncheck() {
  replace_text "$1/AGENTS.md" 'golang.org/x/vuln/cmd/govulncheck@v1.3.0' 'golang.org/x/vuln/cmd/govulncheck@v1.3.1'
}

mutate_agents_gosec() {
  replace_text "$1/AGENTS.md" 'github.com/securego/gosec/v2/cmd/gosec@v2.26.1' 'github.com/securego/gosec/v2/cmd/gosec@v2.26.2'
}

mutate_agents_shellcheck() {
  replace_text "$1/AGENTS.md" 'ShellCheck `0.11.0`' 'ShellCheck `0.11.1`'
}

"$NODE_BIN" "$SCRIPT" --root "${ROOT_DIR}" >/dev/null
assert_harness_scratch_rejects_repo_tmp

expect_drift "node-engine" \
  "package.json: engines.node mismatch: expected 24.15.0, got 24.16.0" \
  mutate_package_node

expect_drift "package-manager" \
  "package.json: packageManager mismatch: expected pnpm@10.33.0, got pnpm@10.34.0" \
  mutate_package_manager

expect_drift "go-toolchain" \
  "go.mod: toolchain mismatch: expected go1.26.2, got go1.26.3" \
  mutate_go_toolchain

expect_drift "go-testcontainers" \
  "go.mod: github.com/testcontainers/testcontainers-go mismatch: expected v0.42.0, got v0.43.0" \
  mutate_go_testcontainers

expect_drift "staticcheck-tool" \
  "Makefile: STATICCHECK_TOOL mismatch: expected honnef.co/go/tools/cmd/staticcheck@v0.7.0, got honnef.co/go/tools/cmd/staticcheck@v0.7.1" \
  mutate_staticcheck_tool

expect_drift "govulncheck-tool" \
  "Makefile: GOVULNCHECK_TOOL mismatch: expected golang.org/x/vuln/cmd/govulncheck@v1.3.0, got golang.org/x/vuln/cmd/govulncheck@v1.3.1" \
  mutate_govulncheck_tool

expect_drift "gosec-tool" \
  "Makefile: GOSEC_TOOL mismatch: expected github.com/securego/gosec/v2/cmd/gosec@v2.26.1, got github.com/securego/gosec/v2/cmd/gosec@v2.26.2" \
  mutate_gosec_tool

expect_drift "shellcheck-version" \
  "Makefile: SHELLCHECK_VERSION mismatch: expected 0.11.0, got 0.11.1" \
  mutate_shellcheck_version

expect_drift "bootstrap-shellcheck-version" \
  "scripts/bootstrap-shellcheck.sh: SHELLCHECK_VERSION default mismatch: expected 0.11.0, got 0.11.1" \
  mutate_bootstrap_shellcheck_version

expect_drift "readme-node" \
  "README.md: Node.js pin line mismatch: expected - Node.js \`24.15.0\`, got - Node.js \`24.16.0\`" \
  mutate_readme_node

expect_drift "readme-staticcheck" \
  "README.md: Staticcheck pin line mismatch: expected - Staticcheck \`v0.7.0\`, got - Staticcheck \`v0.7.1\`" \
  mutate_readme_staticcheck

expect_drift "readme-govulncheck" \
  "README.md: Govulncheck pin line mismatch: expected - Govulncheck \`v1.3.0\`, got - Govulncheck \`v1.3.1\`" \
  mutate_readme_govulncheck

expect_drift "readme-gosec" \
  "README.md: Gosec pin line mismatch: expected - Gosec \`v2.26.1\`, got - Gosec \`v2.26.2\`" \
  mutate_readme_gosec

expect_drift "readme-shellcheck" \
  "README.md: ShellCheck pin line mismatch: expected - ShellCheck \`0.11.0\`, got - ShellCheck \`0.11.1\`" \
  mutate_readme_shellcheck

expect_drift "agents-govulncheck" \
  "AGENTS.md: Pinned bootstrap tools line mismatch" \
  mutate_agents_govulncheck

expect_drift "agents-gosec" \
  "AGENTS.md: Pinned bootstrap tools line mismatch" \
  mutate_agents_gosec

expect_drift "agents-shellcheck" \
  "AGENTS.md: Pinned bootstrap tools line mismatch" \
  mutate_agents_shellcheck

preflight_dir="$(cartulary_harness_mktemp_dir "toolchain-pins-preflight.XXXXXX")"
cleanup_paths+=("${preflight_dir}")
copy_minimal_repo "${preflight_dir}"
replace_text "${preflight_dir}/package.json" '"node": "24.15.0"' '"node": "24.16.0"'

set +e
preflight_output="$(
  make --no-print-directory -C "${preflight_dir}" \
    NODE_BIN="${NODE_BIN}" \
    PNPM="${preflight_dir}/fake-pnpm" \
    NODE_RUNTIME_DIR="${preflight_dir}/node-runtime" \
    check-setup-blockers \
    2>&1
)"
preflight_status=$?
set -e

if [[ "${preflight_status}" -eq 0 ]]; then
  fail "check-setup-blockers mismatch: expected failure"
fi
assert_contains "${preflight_output}" "package.json: engines.node mismatch: expected 24.15.0, got 24.16.0" "check-setup-blockers diagnostic"
assert_not_contains "${preflight_output}" "frontend install" "check-setup-blockers early failure"
