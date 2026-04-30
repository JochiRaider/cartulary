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
  buildMakeNodeToolChildEnv,
  buildMakeNodeToolInvocation,
  makeNodeToolMakeEnvVars,
  makeNodeToolNames,
  makeNodeToolRuntimeEnvVars,
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

function assertList(label, actual, expected) {
  assert(
    JSON.stringify(actual) === JSON.stringify(expected),
    `${label} mismatch\nexpected=${JSON.stringify(expected)}\nactual=${JSON.stringify(actual)}`,
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
const expectedMakeEnvVars = {
  "browser-e2e-duration-baseline-drift": [
    "BASELINE_FILE",
    "RESULTS_DIR",
    "CARTULARY_TEST_RESULTS_DIR",
    "CARTULARY_TEST_RUN_ID",
  ],
  "explain-phase": ["PHASE", "JSON"],
  "explain-run": ["DETAIL", "RUN_ID", "TARGET", "RESULTS_DIR"],
  "explain-target": ["TARGET", "DETAIL", "JSON"],
  "fixture-report": [
    "FIXTURE_THRESHOLD_MS",
    "FIXTURE_TOP",
    "RUN_ID",
    "TARGET",
    "JSON",
    "RESULTS_DIR",
    "CARTULARY_TEST_RESULTS_DIR",
  ],
  "go-test-duration-baseline-coverage": ["BASELINE_FILE"],
  "go-test-duration-baseline-drift": [
    "BASELINE_FILE",
    "RESULTS_DIR",
    "CARTULARY_TEST_RESULTS_DIR",
    "CARTULARY_TEST_RUN_ID",
  ],
  "go-test-duration-baselines": [
    "PRUNE_OBSERVED_PACKAGES",
    "ALLOW_COMMAND_OVERHEAD_DECREASE",
    "BASELINE_FILE",
    "RESULTS_DIR",
  ],
  "phase-slice": [
    "PHASE",
    "MAKE",
    "TEST_OUTPUT_SCRIPT",
    "CARTULARY_TEST_RESULTS_DIR",
    "CARTULARY_TEST_RUN_ID",
  ],
  "scheduler-event-order-drift": [
    "TARGET",
    "RESULTS_DIR",
    "CARTULARY_TEST_RESULTS_DIR",
    "CARTULARY_TEST_RUN_ID",
  ],
  "service-backed-make-target-duration-baseline-drift": [
    "SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE",
    "EXECUTION_TOPOLOGY_MANIFEST",
    "SERVICE_BACKED_SCHEDULE_MANIFEST",
    "RESULTS_DIR",
    "CARTULARY_TEST_RESULTS_DIR",
    "CARTULARY_TEST_RUN_ID",
  ],
  "service-backed-make-target-duration-baselines": [
    "SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE",
    "RESULTS_DIR",
  ],
  "service-backed-slice": [
    "PHASE",
    "MAKE",
    "TEST_OUTPUT_SCRIPT",
    "CARTULARY_TEST_RESULTS_DIR",
    "CARTULARY_TEST_RUN_ID",
  ],
  "target-plan": [],
  "target-plan-json": [],
  "task-guide": ["ROLE", "PHASE", "JSON", "CARTULARY_TEST_RESULTS_DIR"],
  "task-surface-report": ["TASK_SURFACE_REPORT_ARGS"],
};
assertList("registered make node tools", makeNodeToolNames(), Object.keys(expectedMakeEnvVars).sort());
for (const [name, expected] of Object.entries(expectedMakeEnvVars)) {
  assertList(`${name} Make env vars`, makeNodeToolMakeEnvVars(name), expected);
}
assertList("task-guide runtime env", makeNodeToolRuntimeEnvVars("task-guide"), [
  "CARTULARY_TEST_RESULTS_DIR",
]);
assertList("phase-slice runtime env", makeNodeToolRuntimeEnvVars("phase-slice"), [
  "MAKE",
  "TEST_OUTPUT_SCRIPT",
  "CARTULARY_TEST_RESULTS_DIR",
  "CARTULARY_TEST_RUN_ID",
]);
assertArgs("task-guide", { ROLE: "feature-dev", PHASE: "phase4", JSON: "1" }, [
  "--role",
  "feature-dev",
  "--phase",
  "phase4",
  "--json",
]);
assertArgs("phase-slice", { PHASE: "phase4" }, [
  "--phase",
  "phase4",
  "--mode",
  "phase",
]);
assertArgs("service-backed-slice", { PHASE: "phase4" }, [
  "--phase",
  "phase4",
  "--mode",
  "service-backed",
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
  "go-test-duration-baselines",
  {
    ALLOW_COMMAND_OVERHEAD_DECREASE: "1",
    BASELINE_FILE: "/tmp/baseline.json",
    PRUNE_OBSERVED_PACKAGES: "1",
    RESULTS_DIR: "/tmp/cartulary-results/run-a",
  },
  [
    "--prune-observed-packages",
    "--allow-command-overhead-decrease",
    "--baseline-file",
    "/tmp/baseline.json",
    "/tmp/cartulary-results/run-a",
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
assertArgs("target-plan", { PHASE: "phase4", TARGET: "backend-store", RESULTS_DIR: "/tmp/results" }, []);
assertUsage("explain-run", {}, "make explain-run RESULTS_DIR=<root|run-dir>");
assertUsage("phase-slice", {}, "make phase-slice PHASE=<phaseN>");
assertUsage("service-backed-slice", {}, "make service-backed-slice PHASE=<phaseN>");
assertUsage("go-test-duration-baselines", {}, "make go-test-duration-baselines RESULTS_DIR=<successful test results dir>");
assertUsage("scheduler-event-order-drift", { CARTULARY_TEST_RESULTS_DIR: "/tmp/results" }, "make scheduler-event-order-drift");

try {
  buildMakeNodeToolInvocation("task-surface-report", { TASK_SURFACE_REPORT_ARGS: "'--check'" });
  throw new Error("quoted task-surface args should fail");
} catch (error) {
  assert(error instanceof UsageError, "quoted task-surface args must throw UsageError");
}

const targetPlanChildEnv = buildMakeNodeToolChildEnv("target-plan", {
  PATH: "/bin",
  PHASE: "phase4",
  RESULTS_DIR: "/tmp/results",
  TARGET: "backend-store",
});
assert(targetPlanChildEnv.PATH === "/bin", "child env must preserve unrelated runtime env");
assert(!("PHASE" in targetPlanChildEnv), "target-plan child env must not expose undeclared PHASE");
assert(!("RESULTS_DIR" in targetPlanChildEnv), "target-plan child env must not expose undeclared RESULTS_DIR");
assert(!("TARGET" in targetPlanChildEnv), "target-plan child env must not expose undeclared TARGET");

const taskGuideChildEnv = buildMakeNodeToolChildEnv("task-guide", {
  CARTULARY_TEST_RESULTS_DIR: "/tmp/results",
  JSON: "1",
  PHASE: "phase4",
  ROLE: "feature-dev",
});
assert(
  taskGuideChildEnv.CARTULARY_TEST_RESULTS_DIR === "/tmp/results",
  "task-guide child env must keep result-root runtime env",
);
assert(!("JSON" in taskGuideChildEnv), "task-guide child env must not expose JSON after args are built");
assert(!("PHASE" in taskGuideChildEnv), "task-guide child env must not expose PHASE after args are built");
assert(!("ROLE" in taskGuideChildEnv), "task-guide child env must not expose ROLE after args are built");

const phaseSliceChildEnv = buildMakeNodeToolChildEnv("phase-slice", {
  CARTULARY_TEST_RESULTS_DIR: "/tmp/results",
  CARTULARY_TEST_RUN_ID: "run-a",
  MAKE: "make",
  PHASE: "phase4",
  TARGET: "backend-store",
  TEST_OUTPUT_SCRIPT: "/tmp/test-output.mjs",
});
assert(phaseSliceChildEnv.MAKE === "make", "phase-slice child env must keep MAKE");
assert(
  phaseSliceChildEnv.TEST_OUTPUT_SCRIPT === "/tmp/test-output.mjs",
  "phase-slice child env must keep TEST_OUTPUT_SCRIPT",
);
assert(
  phaseSliceChildEnv.CARTULARY_TEST_RUN_ID === "run-a",
  "phase-slice child env must keep CARTULARY_TEST_RUN_ID",
);
assert(!("PHASE" in phaseSliceChildEnv), "phase-slice child env must not expose PHASE after args are built");
assert(!("TARGET" in phaseSliceChildEnv), "phase-slice child env must not expose unrelated TARGET");
EOF

set +e
missing_phase_output="$("$NODE_BIN" "$ROOT_DIR/scripts/run-make-node-tool.mjs" explain-phase 2>&1)"
missing_phase_status=$?
set -e
if [[ "$missing_phase_status" -ne 2 ]]; then
  fail "missing phase launcher validation should exit 2"
fi
assert_contains "$missing_phase_output" "usage: make explain-phase PHASE=<phaseN>" "missing phase launcher usage"
