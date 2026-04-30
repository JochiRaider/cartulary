#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"

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

"$NODE_BIN" --input-type=module - "$ROOT_DIR" <<'EOF'
import path from "node:path";
import { pathToFileURL } from "node:url";

const [root] = process.argv.slice(2);
const {
  buildMakeNodeToolInvocation,
  makeNodeToolNames,
  UsageError,
} = await import(pathToFileURL(path.join(root, "scripts/lib/make-node-tools.mjs")).href);

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

function assertArgs(name, env, expected) {
  const actual = buildMakeNodeToolInvocation(name, env).args;
  assert(
    JSON.stringify(actual) === JSON.stringify(expected),
    `${name} args mismatch\nexpected=${JSON.stringify(expected)}\nactual=${JSON.stringify(actual)}`,
  );
}

function assertUsage(name, env, usagePart) {
  try {
    buildMakeNodeToolInvocation(name, env);
  } catch (error) {
    assert(error instanceof UsageError, `${name} must throw UsageError`);
    assert(String(error.usage).includes(usagePart), `${name} usage must include ${usagePart}`);
    return;
  }
  throw new Error(`${name} should fail usage validation`);
}

assert(makeNodeToolNames().includes("task-guide"), "registry must list task-guide");
assertArgs("task-guide", { ROLE: "feature-dev", PHASE: "phase4", JSON: "1" }, [
  "--role",
  "feature-dev",
  "--phase",
  "phase4",
  "--json",
]);
assertArgs("explain-target", { TARGET: "backend-store" }, [
  "--target",
  "backend-store",
  "--detail",
  "summary",
]);
assertArgs("explain-target", { TARGET: "backend-store", DETAIL: "rows", JSON: "1" }, [
  "--target",
  "backend-store",
  "--detail",
  "rows",
  "--json",
]);
assertArgs(
  "fixture-report",
  { CARTULARY_TEST_RESULTS_DIR: "/tmp/cartulary-results", FIXTURE_THRESHOLD_MS: "4000", FIXTURE_TOP: "9" },
  ["--results-dir", "/tmp/cartulary-results", "--threshold-ms", "4000", "--top", "9"],
);
assertArgs(
  "fixture-report",
  { RESULTS_DIR: "/tmp/explicit-results", FIXTURE_THRESHOLD_MS: "4000", FIXTURE_TOP: "9", RUN_ID: "run-a", TARGET: "backend-store", JSON: "1" },
  [
    "--results-dir",
    "/tmp/explicit-results",
    "--threshold-ms",
    "4000",
    "--top",
    "9",
    "--run-id",
    "run-a",
    "--target",
    "backend-store",
    "--json",
  ],
);
assertArgs(
  "go-test-duration-baseline-drift",
  {
    BASELINE_FILE: "/tmp/baseline.json",
    CARTULARY_TEST_RESULTS_DIR: "/tmp/cartulary-results",
    CARTULARY_TEST_RUN_ID: "run-a",
  },
  ["--baseline-file", "/tmp/baseline.json", "/tmp/cartulary-results/run-a"],
);
assertArgs(
  "service-backed-make-target-duration-baseline-drift",
  {
    CARTULARY_TEST_RESULTS_DIR: "/tmp/cartulary-results",
    CARTULARY_TEST_RUN_ID: "run-a",
    EXECUTION_TOPOLOGY_MANIFEST: "/tmp/topology.json",
    SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE: "/tmp/service-baseline.json",
    SERVICE_BACKED_SCHEDULE_MANIFEST: "/tmp/schedule.json",
  },
  [
    "check-drift",
    "--baseline-file",
    "/tmp/service-baseline.json",
    "--topology",
    "/tmp/topology.json",
    "--schedule-manifest",
    "/tmp/schedule.json",
    "/tmp/cartulary-results/run-a",
  ],
);
assertArgs("task-surface-report", { TASK_SURFACE_REPORT_ARGS: "--check --all" }, [
  "--check",
  "--all",
]);
assertUsage("explain-run", {}, "make explain-run RESULTS_DIR=<root|run-dir>");
assertUsage("go-test-duration-baselines", {}, "make go-test-duration-baselines RESULTS_DIR=<successful test results dir>");
assertUsage("scheduler-event-order-drift", { CARTULARY_TEST_RESULTS_DIR: "/tmp/results" }, "make scheduler-event-order-drift");

try {
  buildMakeNodeToolInvocation("task-surface-report", { TASK_SURFACE_REPORT_ARGS: "'--check'" });
  throw new Error("quoted task-surface args should fail");
} catch (error) {
  assert(error instanceof UsageError, "quoted task-surface args must throw UsageError");
}
EOF

set +e
missing_phase_output="$("$NODE_BIN" "$ROOT_DIR/scripts/run-make-node-tool.mjs" explain-phase 2>&1)"
missing_phase_status=$?
set -e
if [[ "$missing_phase_status" -ne 2 ]]; then
  fail "missing phase launcher validation should exit 2"
fi
assert_contains "$missing_phase_output" "usage: make explain-phase PHASE=<phaseN>" "missing phase launcher usage"
