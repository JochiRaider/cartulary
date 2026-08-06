#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
RUN_STEP="${ROOT_DIR}/tools/harness/execution/run-step.sh"
TEST_OUTPUT="${ROOT_DIR}/tools/harness/output/test-output.sh"
NODE_BIN="${NODE_BIN:-${ROOT_DIR}/tmp/node-runtime/bin/node}"
# shellcheck source=tools/harness/test-support/harness-scratch.sh
source "${ROOT_DIR}/tools/harness/test-support/harness-scratch.sh"

scratch="$(cartulary_harness_mktemp_dir "capacity-failure-contract.XXXXXX")"
trap 'rm -rf "$scratch"' EXIT
affected_path="${scratch}/go/tmp/go-build-fixture"
mkdir -p "$(dirname "$affected_path")"

set +e
output="$({
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$scratch" \
  CARTULARY_TEST_RUN_ID=capacity-failure \
  CARTULARY_TEST_TARGET=build-server \
  GO_TMP_DIR="$(dirname "$affected_path")" \
    "$RUN_STEP" "build server" -- \
      bash -lc "echo 'mkdir $affected_path: no space left on device (ENOSPC)' >&2; exit 1"
} 2>&1)"
status=$?
set -e

[[ "$status" -eq 1 ]] || {
  printf 'capacity child status: expected 1, got %s\n%s\n' "$status" "$output" >&2
  exit 1
}
[[ "$output" == *"failure_class=infra"* && "$output" == *"reason=resource_conflict"* ]] || {
  printf 'capacity failure was not normalized as infra/resource_conflict:\n%s\n' "$output" >&2
  exit 1
}

step_summary="${scratch}/capacity-failure/build-server/build-server/step-summary.json"
"$NODE_BIN" - "$step_summary" "$(dirname "$affected_path")" <<'EOF'
const fs = require("node:fs");
const [file, expectedPath] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(file, "utf8"));
if (summary.failure_class !== "infra" || summary.failure_reason !== "resource_conflict") {
  throw new Error(`unexpected failure normalization: ${summary.failure_class}/${summary.failure_reason}`);
}
const message = summary.failures?.[0]?.message ?? "";
for (const expected of [
  `filesystem_capacity path=${expectedPath}`,
  "filesystem_dev=",
  "available_bytes=",
]) {
  if (!message.includes(expected)) throw new Error(`capacity diagnostic omitted ${expected}`);
}
EOF

CARTULARY_OUTPUT_MODE=quiet \
CARTULARY_TEST_RESULTS_DIR="$scratch" \
CARTULARY_TEST_RUN_ID=capacity-failure \
  "$TEST_OUTPUT" target-summary build-server fail >/dev/null

"$NODE_BIN" - "${scratch}/capacity-failure/build-server/tool-run-summary.json" <<'EOF'
const fs = require("node:fs");
const summary = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
if (summary.exit_code !== 4) {
  throw new Error(`resource_conflict exit code: expected 4, got ${summary.exit_code}`);
}
EOF
