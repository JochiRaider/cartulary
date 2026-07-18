#!/usr/bin/env bash
# Single-quoted literals below intentionally assert text containing shell/Markdown metacharacters.
# shellcheck disable=SC2016
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
if command -v "${NODE_BIN}" >/dev/null 2>&1; then
  NODE_BIN="$(command -v "${NODE_BIN}")"
fi
SCRIPT="${ROOT_DIR}/tools/harness/readiness/toolchain-pin-check-cli.mjs"
source "${ROOT_DIR}/tools/harness/test-support/harness-scratch.sh"
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

  mkdir -p "${dest}/tools"
  cp "${ROOT_DIR}/Makefile" "${dest}/Makefile"
  cp "${ROOT_DIR}/package.json" "${dest}/package.json"
  if [[ -d "${ROOT_DIR}/node_modules" ]]; then
    ln -s "${ROOT_DIR}/node_modules" "${dest}/node_modules"
  fi
  cp "${ROOT_DIR}/go.mod" "${dest}/go.mod"
  cp "${ROOT_DIR}/README.md" "${dest}/README.md"
  cp "${ROOT_DIR}/AGENTS.md" "${dest}/AGENTS.md"
  cp "${ROOT_DIR}/tools/task_surface.generated.mk" "${dest}/tools/task_surface.generated.mk"
  cp "${ROOT_DIR}/tools/task_surface.runtime.generated.mk" "${dest}/tools/task_surface.runtime.generated.mk"
  cp "${ROOT_DIR}/tools/task_surface_manifest.json" "${dest}/tools/task_surface_manifest.json"
  cp "${ROOT_DIR}/tools/scheduler_manifest.json" "${dest}/tools/scheduler_manifest.json"
  cp "${ROOT_DIR}/tools/browser_e2e_batch_manifest.json" "${dest}/tools/browser_e2e_batch_manifest.json"
  cp "${ROOT_DIR}/tools/execution_topology_manifest.json" "${dest}/tools/execution_topology_manifest.json"
  cp "${ROOT_DIR}/tools/task_surface_owner.json" "${dest}/tools/task_surface_owner.json"
  cp "${ROOT_DIR}/tools/harness_redaction_manifest.json" "${dest}/tools/harness_redaction_manifest.json"
  cp "${ROOT_DIR}/tools/scheduler_resource_registry.json" "${dest}/tools/scheduler_resource_registry.json"
  cp "${ROOT_DIR}/tools/toolchain_pins.json" "${dest}/tools/toolchain_pins.json"
  cp -R "${ROOT_DIR}/tools/harness" "${dest}/tools/harness"
  cp -R "${ROOT_DIR}/tools/schemas" "${dest}/tools/schemas"
  cp "${ROOT_DIR}"/tools/*duration_baselines.json "${dest}/tools/"
  cp "${ROOT_DIR}/tools/harness/readiness/list-build-inputs.sh" "${dest}/tools/harness/readiness/list-build-inputs.sh"
  cp "${ROOT_DIR}/tools/harness/readiness/bootstrap-node-runtime.sh" "${dest}/tools/harness/readiness/bootstrap-node-runtime.sh"
  cp "${ROOT_DIR}/tools/harness/readiness/bootstrap-shellcheck.sh" "${dest}/tools/harness/readiness/bootstrap-shellcheck.sh"
  cp "${ROOT_DIR}/tools/harness/readiness/toolchain-pin-check-cli.mjs" "${dest}/tools/harness/readiness/toolchain-pin-check-cli.mjs"
  cp "${ROOT_DIR}/tools/harness/execution/cartulary-runner-cli.mjs" "${dest}/tools/harness/execution/cartulary-runner-cli.mjs"
  cp "${ROOT_DIR}/tools/harness/contract/harness-contract-cli.mjs" "${dest}/tools/harness/contract/harness-contract-cli.mjs"
  cp "${ROOT_DIR}/tools/harness/scheduler/check-schedule-cli.mjs" "${dest}/tools/harness/scheduler/check-schedule-cli.mjs"
  mkdir -p \
    "${dest}/apps/web" \
    "${dest}/cmd/migrate" \
    "${dest}/cmd/operator" \
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
    cmd/operator \
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

pin_value() {
  local path="$1"

  "$NODE_BIN" - "$ROOT_DIR/tools/toolchain_pins.json" "$path" <<'EOF'
const fs = require("node:fs");
const [file, path] = process.argv.slice(2);
const value = path.split(".").reduce((current, key) => current?.[key], JSON.parse(fs.readFileSync(file, "utf8")));
if (typeof value !== "string" || value.length === 0) {
  process.exit(1);
}
process.stdout.write(value);
EOF
}

bump_dotted_version() {
  local value="$1"
  local prefix=""

  if [[ "$value" == v* || "$value" == go* ]]; then
    prefix="${value%%[0-9]*}"
    value="${value#"$prefix"}"
  fi

  IFS=. read -r major minor patch extra <<<"$value"
  if [[ -n "${extra:-}" || -z "${major:-}" || -z "${minor:-}" || -z "${patch:-}" || ! "$patch" =~ ^[0-9]+$ ]]; then
    fail "cannot bump dotted version [$1]"
  fi
  printf '%s%s.%s.%s\n' "$prefix" "$major" "$minor" "$((patch + 1))"
}

bump_tool_spec() {
  local value="$1"
  local package="${value%@*}"
  local version="${value##*@}"

  if [[ "$package" == "$value" ]]; then
    fail "cannot bump unversioned tool spec [$value]"
  fi
  printf '%s@%s\n' "$package" "$(bump_dotted_version "$version")"
}

node_version="$(pin_value node_version)"
node_version_alt="$(bump_dotted_version "$node_version")"
pnpm_version="$(pin_value pnpm_version)"
pnpm_version_alt="$(bump_dotted_version "$pnpm_version")"
go_toolchain="$(pin_value go_toolchain)"
go_toolchain_alt="$(bump_dotted_version "$go_toolchain")"
testcontainers_go_version="$(pin_value testcontainers_go_version)"
testcontainers_go_version_alt="$(bump_dotted_version "$testcontainers_go_version")"
staticcheck_tool="$(pin_value tools.staticcheck)"
staticcheck_tool_alt="$(bump_tool_spec "$staticcheck_tool")"
staticcheck_version="${staticcheck_tool##*@}"
staticcheck_version_alt="${staticcheck_tool_alt##*@}"
govulncheck_tool="$(pin_value tools.govulncheck)"
govulncheck_tool_alt="$(bump_tool_spec "$govulncheck_tool")"
govulncheck_version="${govulncheck_tool##*@}"
govulncheck_version_alt="${govulncheck_tool_alt##*@}"
gosec_tool="$(pin_value tools.gosec)"
gosec_tool_alt="$(bump_tool_spec "$gosec_tool")"
gosec_version="${gosec_tool##*@}"
gosec_version_alt="${gosec_tool_alt##*@}"
shellcheck_version="$(pin_value shellcheck_version)"
shellcheck_version_alt="$(bump_dotted_version "$shellcheck_version")"

mutate_package_node() {
  replace_text "$1/package.json" '"node": "'"$node_version"'"' '"node": "'"$node_version_alt"'"'
}

mutate_package_manager() {
  replace_text "$1/package.json" '"packageManager": "pnpm@'"$pnpm_version"'"' '"packageManager": "pnpm@'"$pnpm_version_alt"'"'
}

mutate_go_toolchain() {
  replace_text "$1/go.mod" "toolchain $go_toolchain" "toolchain $go_toolchain_alt"
}

mutate_go_testcontainers() {
  replace_text "$1/go.mod" "github.com/testcontainers/testcontainers-go $testcontainers_go_version" "github.com/testcontainers/testcontainers-go $testcontainers_go_version_alt"
}

mutate_staticcheck_tool() {
  replace_text "$1/Makefile" "STATICCHECK_TOOL := $staticcheck_tool" "STATICCHECK_TOOL := $staticcheck_tool_alt"
}

mutate_govulncheck_tool() {
  replace_text "$1/Makefile" "GOVULNCHECK_TOOL := $govulncheck_tool" "GOVULNCHECK_TOOL := $govulncheck_tool_alt"
}

mutate_gosec_tool() {
  replace_text "$1/Makefile" "GOSEC_TOOL := $gosec_tool" "GOSEC_TOOL := $gosec_tool_alt"
}

mutate_shellcheck_version() {
  replace_text "$1/Makefile" "SHELLCHECK_VERSION ?= $shellcheck_version" "SHELLCHECK_VERSION ?= $shellcheck_version_alt"
}

mutate_bootstrap_shellcheck_version() {
  replace_text "$1/tools/harness/readiness/bootstrap-shellcheck.sh" 'SHELLCHECK_VERSION="${SHELLCHECK_VERSION:-'"$shellcheck_version"'}"' 'SHELLCHECK_VERSION="${SHELLCHECK_VERSION:-'"$shellcheck_version_alt"'}"'
}

mutate_readme_node() {
  replace_text "$1/README.md" "- Node.js \`$node_version\`" "- Node.js \`$node_version_alt\`"
}

mutate_readme_staticcheck() {
  replace_text "$1/README.md" "- Staticcheck \`$staticcheck_version\`" "- Staticcheck \`$staticcheck_version_alt\`"
}

mutate_readme_govulncheck() {
  replace_text "$1/README.md" "- Govulncheck \`$govulncheck_version\`" "- Govulncheck \`$govulncheck_version_alt\`"
}

mutate_readme_gosec() {
  replace_text "$1/README.md" "- Gosec \`$gosec_version\`" "- Gosec \`$gosec_version_alt\`"
}

mutate_readme_shellcheck() {
  replace_text "$1/README.md" "- ShellCheck \`$shellcheck_version\`" "- ShellCheck \`$shellcheck_version_alt\`"
}

mutate_agents_govulncheck() {
  replace_text "$1/AGENTS.md" "$govulncheck_tool" "$govulncheck_tool_alt"
}

mutate_agents_gosec() {
  replace_text "$1/AGENTS.md" "$gosec_tool" "$gosec_tool_alt"
}

mutate_agents_shellcheck() {
  replace_text "$1/AGENTS.md" "ShellCheck \`$shellcheck_version\`" "ShellCheck \`$shellcheck_version_alt\`"
}

"$NODE_BIN" "$SCRIPT" --root "${ROOT_DIR}" >/dev/null
assert_harness_scratch_rejects_repo_tmp

expect_drift "node-engine" \
  "package.json: engines.node mismatch: expected $node_version, got $node_version_alt" \
  mutate_package_node

expect_drift "package-manager" \
  "package.json: packageManager mismatch: expected pnpm@$pnpm_version, got pnpm@$pnpm_version_alt" \
  mutate_package_manager

expect_drift "go-toolchain" \
  "go.mod: toolchain mismatch: expected $go_toolchain, got $go_toolchain_alt" \
  mutate_go_toolchain

expect_drift "go-testcontainers" \
  "go.mod: github.com/testcontainers/testcontainers-go mismatch: expected $testcontainers_go_version, got $testcontainers_go_version_alt" \
  mutate_go_testcontainers

expect_drift "staticcheck-tool" \
  "Makefile: STATICCHECK_TOOL mismatch: expected $staticcheck_tool, got $staticcheck_tool_alt" \
  mutate_staticcheck_tool

expect_drift "govulncheck-tool" \
  "Makefile: GOVULNCHECK_TOOL mismatch: expected $govulncheck_tool, got $govulncheck_tool_alt" \
  mutate_govulncheck_tool

expect_drift "gosec-tool" \
  "Makefile: GOSEC_TOOL mismatch: expected $gosec_tool, got $gosec_tool_alt" \
  mutate_gosec_tool

expect_drift "shellcheck-version" \
  "Makefile: SHELLCHECK_VERSION mismatch: expected $shellcheck_version, got $shellcheck_version_alt" \
  mutate_shellcheck_version

expect_drift "bootstrap-shellcheck-version" \
  "tools/harness/readiness/bootstrap-shellcheck.sh: SHELLCHECK_VERSION default mismatch: expected $shellcheck_version, got $shellcheck_version_alt" \
  mutate_bootstrap_shellcheck_version

expect_drift "readme-node" \
  "README.md: Node.js pin line mismatch: expected - Node.js \`$node_version\`, got - Node.js \`$node_version_alt\`" \
  mutate_readme_node

expect_drift "readme-staticcheck" \
  "README.md: Staticcheck pin line mismatch: expected - Staticcheck \`$staticcheck_version\`, got - Staticcheck \`$staticcheck_version_alt\`" \
  mutate_readme_staticcheck

expect_drift "readme-govulncheck" \
  "README.md: Govulncheck pin line mismatch: expected - Govulncheck \`$govulncheck_version\`, got - Govulncheck \`$govulncheck_version_alt\`" \
  mutate_readme_govulncheck

expect_drift "readme-gosec" \
  "README.md: Gosec pin line mismatch: expected - Gosec \`$gosec_version\`, got - Gosec \`$gosec_version_alt\`" \
  mutate_readme_gosec

expect_drift "readme-shellcheck" \
  "README.md: ShellCheck pin line mismatch: expected - ShellCheck \`$shellcheck_version\`, got - ShellCheck \`$shellcheck_version_alt\`" \
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
preflight_results_root="$(cartulary_harness_mktemp_dir "toolchain-pins-preflight-results.XXXXXX")"
preflight_run_id="toolchain-pins-preflight"
cleanup_paths+=("${preflight_dir}")
cleanup_paths+=("${preflight_results_root}")
copy_minimal_repo "${preflight_dir}"
replace_text "${preflight_dir}/package.json" '"node": "'"$node_version"'"' '"node": "'"$node_version_alt"'"'
preflight_log="${preflight_dir}/toolchain-drift-preflight.log"

set +e
make --no-print-directory -C "${preflight_dir}" \
  CARTULARY_TEST_RESULTS_DIR="${preflight_results_root}" \
  CARTULARY_TEST_RUN_ID="${preflight_run_id}" \
  CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES=1 \
  NODE_BIN="${NODE_BIN}" \
  toolchain-drift \
  >"${preflight_log}" 2>&1
preflight_status=$?
set -e

if [[ "${preflight_status}" -eq 0 ]]; then
  fail "check toolchain drift mismatch: expected failure; see ${preflight_log}"
fi
"${NODE_BIN}" "${ROOT_DIR}/tools/harness/test-support/harness-artifact-assert.mjs" \
  --repo-root "${preflight_dir}" \
  --results-root "${preflight_results_root}" \
  --run-id "${preflight_run_id}" \
  --target "toolchain-drift" \
  --step-label "toolchain-drift" \
  --needle "package.json: engines.node mismatch: expected $node_version, got $node_version_alt" \
  --label "Make toolchain-drift diagnostic"
