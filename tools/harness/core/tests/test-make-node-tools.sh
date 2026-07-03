#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
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
} = await import(pathToFileURL(path.join(root, "tools/harness/core/make-node-tools.mjs")).href);

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
  "browser-e2e-duration-baselines": [
    "RESULTS_DIR",
    "BROWSER_E2E_DURATION_BASELINE",
  ],
  "browser-e2e-duration-baseline-drift": [
    "RESULTS_DIR",
    "BROWSER_E2E_DURATION_BASELINE",
    "CARTULARY_TEST_RESULTS_DIR",
    "CARTULARY_TEST_RUN_ID",
  ],
  "explain-phase": ["PHASE", "PHASE_NAMESPACE", "JSON"],
  "explain-run": ["RESULTS_DIR", "RUN_ID", "TARGET", "DETAIL"],
  "explain-target": ["TARGET", "DETAIL", "JSON"],
  "fixture-report": [
    "RESULTS_DIR",
    "RUN_ID",
    "TARGET",
    "JSON",
    "FIXTURE_THRESHOLD_MS",
    "FIXTURE_TOP",
    "CARTULARY_TEST_RESULTS_DIR",
  ],
  "frontend-fallow-static": [
    "CARTULARY_TEST_RESULTS_DIR",
    "CARTULARY_TEST_RUN_ID",
    "NODE_BIN",
    "NODE_RUNTIME_DIR",
    "PNPM",
  ],
  "go-test-duration-baseline-coverage": ["GO_TEST_DURATION_BASELINE"],
  "go-test-duration-baseline-drift": [
    "RESULTS_DIR",
    "GO_TEST_DURATION_BASELINE",
    "CARTULARY_TEST_RESULTS_DIR",
    "CARTULARY_TEST_RUN_ID",
  ],
  "go-test-duration-baselines": [
    "RESULTS_DIR",
    "PRUNE_OBSERVED_PACKAGES",
    "ALLOW_COMMAND_OVERHEAD_DECREASE",
    "GO_TEST_DURATION_BASELINE",
  ],
  "harness-smoke-duration-baseline-drift": [
    "RESULTS_DIR",
    "HARNESS_SMOKE_DURATION_BASELINE",
    "CARTULARY_TEST_RESULTS_DIR",
    "CARTULARY_TEST_RUN_ID",
  ],
  "harness-smoke-duration-baselines": [
    "RESULTS_DIR",
    "HARNESS_SMOKE_DURATION_BASELINE",
  ],
  "phase-slice": [
    "PHASE",
    "PHASE_NAMESPACE",
    "ROWS",
    "JSON",
    "MAKE",
    "TEST_OUTPUT_SCRIPT",
    "CARTULARY_TEST_RESULTS_DIR",
    "CARTULARY_TEST_RUN_ID",
    "TEST_SERVICES_BIN",
    "NODE_BIN",
    "NODE_RUNTIME_DIR",
    "PNPM",
    "SERVER_BIN",
    "MIGRATE_BIN",
    "GO",
    "GO_CACHE_DIR",
    "GO_MOD_CACHE_DIR",
    "GO_TEST_SERVICE_PACKAGE_PARALLELISM",
    "BACKEND_STORE_GO_TEST_P",
    "BACKEND_INTEGRATION_GO_TEST_P",
    "BACKEND_INTEGRATION_SHARD_JOBS",
    "PLAYWRIGHT_WORKERS",
    "BROWSER_E2E_FUNCTIONAL_SHARDS",
    "VITEST_FLAGS",
    "VITEST_MAX_WORKERS",
    "TASK_SURFACE_MANIFEST",
    "SCHEDULER_MANIFEST",
    "BROWSER_E2E_BATCH_MANIFEST",
    "CARTULARY_RUNNER_SCRIPT",
    "RUN_PHASE_SCRIPT",
    "RUN_GO_TARGET_SCRIPT",
    "RUN_SERVICE_BACKED_SCHEDULE_SCRIPT",
  ],
  "scheduler-event-order-drift": [
    "RESULTS_DIR",
    "TARGET",
    "CARTULARY_TEST_RESULTS_DIR",
    "CARTULARY_TEST_RUN_ID",
  ],
  "scheduler-summary-timing-drift": [
    "RESULTS_DIR",
    "TARGET",
    "SCHEDULER_WARM_CHECK_BUDGET_MS",
    "SCHEDULER_WARM_CHECK_BALANCE_RATIO",
    "CARTULARY_TEST_RESULTS_DIR",
    "CARTULARY_TEST_RUN_ID",
  ],
  "service-backed-make-target-duration-baseline-drift": [
    "RESULTS_DIR",
    "SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE",
    "CARTULARY_TEST_RESULTS_DIR",
    "CARTULARY_TEST_RUN_ID",
  ],
  "service-backed-make-target-duration-baselines": [
    "RESULTS_DIR",
    "SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE",
  ],
  "service-backed-slice": [
    "PHASE",
    "PHASE_NAMESPACE",
    "ROWS",
    "JSON",
    "MAKE",
    "TEST_OUTPUT_SCRIPT",
    "CARTULARY_TEST_RESULTS_DIR",
    "CARTULARY_TEST_RUN_ID",
    "TEST_SERVICES_BIN",
    "NODE_BIN",
    "NODE_RUNTIME_DIR",
    "PNPM",
    "SERVER_BIN",
    "MIGRATE_BIN",
    "GO",
    "GO_CACHE_DIR",
    "GO_MOD_CACHE_DIR",
    "GO_TEST_SERVICE_PACKAGE_PARALLELISM",
    "BACKEND_STORE_GO_TEST_P",
    "BACKEND_INTEGRATION_GO_TEST_P",
    "BACKEND_INTEGRATION_SHARD_JOBS",
    "PLAYWRIGHT_WORKERS",
    "BROWSER_E2E_FUNCTIONAL_SHARDS",
    "VITEST_FLAGS",
    "VITEST_MAX_WORKERS",
    "TASK_SURFACE_MANIFEST",
    "SCHEDULER_MANIFEST",
    "BROWSER_E2E_BATCH_MANIFEST",
    "CARTULARY_RUNNER_SCRIPT",
    "RUN_PHASE_SCRIPT",
    "RUN_GO_TARGET_SCRIPT",
    "RUN_SERVICE_BACKED_SCHEDULE_SCRIPT",
  ],
  "target-plan": ["TARGET"],
  "target-plan-json": ["TARGET"],
  "task-guide": ["ROLE", "PHASE", "PHASE_NAMESPACE", "JSON", "CARTULARY_TEST_RESULTS_DIR"],
  "task-surface-report": ["TASK_SURFACE_REPORT_ARGS"],
};
assertList("registered make node tools", makeNodeToolNames(), Object.keys(expectedMakeEnvVars).sort());
for (const [name, expected] of Object.entries(expectedMakeEnvVars)) {
  assertList(`${name} Make env vars`, makeNodeToolMakeEnvVars(name), expected);
}
assertList("task-guide runtime env", makeNodeToolRuntimeEnvVars("task-guide"), [
  "CARTULARY_TEST_RESULTS_DIR",
]);
assertList("frontend-fallow-static runtime env", makeNodeToolRuntimeEnvVars("frontend-fallow-static"), [
  "CARTULARY_TEST_RESULTS_DIR",
  "CARTULARY_TEST_RUN_ID",
  "NODE_BIN",
  "NODE_RUNTIME_DIR",
  "PNPM",
]);
assertList("phase-slice runtime env", makeNodeToolRuntimeEnvVars("phase-slice"), [
  "MAKE",
  "TEST_OUTPUT_SCRIPT",
  "CARTULARY_TEST_RESULTS_DIR",
  "CARTULARY_TEST_RUN_ID",
  "TEST_SERVICES_BIN",
  "NODE_BIN",
  "NODE_RUNTIME_DIR",
  "PNPM",
  "SERVER_BIN",
  "MIGRATE_BIN",
  "GO",
  "GO_CACHE_DIR",
  "GO_MOD_CACHE_DIR",
  "GO_TEST_SERVICE_PACKAGE_PARALLELISM",
  "BACKEND_STORE_GO_TEST_P",
  "BACKEND_INTEGRATION_GO_TEST_P",
  "BACKEND_INTEGRATION_SHARD_JOBS",
  "PLAYWRIGHT_WORKERS",
  "BROWSER_E2E_FUNCTIONAL_SHARDS",
  "VITEST_FLAGS",
  "VITEST_MAX_WORKERS",
  "TASK_SURFACE_MANIFEST",
  "SCHEDULER_MANIFEST",
  "BROWSER_E2E_BATCH_MANIFEST",
  "CARTULARY_RUNNER_SCRIPT",
  "RUN_PHASE_SCRIPT",
  "RUN_GO_TARGET_SCRIPT",
  "RUN_SERVICE_BACKED_SCHEDULE_SCRIPT",
]);
assertArgs("task-guide", { ROLE: "feature-dev", PHASE: "phase4", JSON: "1" }, [
  "--role",
  "feature-dev",
  "--phase",
  "phase4",
  "--json",
]);
assertArgs("task-guide", { ROLE: "feature-dev", PHASE: "FE-P3", PHASE_NAMESPACE: "frontend", JSON: "1" }, [
  "--role",
  "feature-dev",
  "--phase",
  "FE-P3",
  "--phase-namespace",
  "frontend",
  "--json",
]);
assertArgs("phase-slice", { PHASE: "phase4" }, [
  "--phase",
  "phase4",
  "--mode",
  "phase",
]);
assertArgs("phase-slice", { PHASE: "phase4", JSON: "1" }, [
  "--phase",
  "phase4",
  "--mode",
  "phase",
  "--json",
]);
assertArgs("phase-slice", { PHASE: "FE-P3", PHASE_NAMESPACE: "frontend" }, [
  "--phase",
  "FE-P3",
  "--mode",
  "phase",
  "--phase-namespace",
  "frontend",
]);
assertArgs("phase-slice", { PHASE: "FE-P5", PHASE_NAMESPACE: "frontend", ROWS: "FE-I-P5-01" }, [
  "--phase",
  "FE-P5",
  "--mode",
  "phase",
  "--phase-namespace",
  "frontend",
  "--rows",
  "FE-I-P5-01",
]);
assertArgs("service-backed-slice", { PHASE: "phase4" }, [
  "--phase",
  "phase4",
  "--mode",
  "service-backed",
]);
assertArgs("service-backed-slice", { PHASE: "FE-P5", PHASE_NAMESPACE: "frontend", ROWS: "FE-I-P5-01" }, [
  "--phase",
  "FE-P5",
  "--mode",
  "service-backed",
  "--phase-namespace",
  "frontend",
  "--rows",
  "FE-I-P5-01",
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
assertArgs("explain-run", { RESULTS_DIR: "/tmp/cartulary-results/run-a", TARGET: "check", DETAIL: "progress" }, [
  "--results-dir",
  "/tmp/cartulary-results/run-a",
  "--detail",
  "progress",
  "--target",
  "check",
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
    GO_TEST_DURATION_BASELINE: "/tmp/baseline.json",
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
    GO_TEST_DURATION_BASELINE: "/tmp/baseline.json",
    CARTULARY_TEST_RESULTS_DIR: "/tmp/cartulary-results",
    CARTULARY_TEST_RUN_ID: "run-a",
  },
  ["--baseline-file", "/tmp/baseline.json", "/tmp/cartulary-results/run-a"],
);
assertArgs(
  "browser-e2e-duration-baselines",
  {
    BROWSER_E2E_DURATION_BASELINE: "/tmp/browser-baseline.json",
    RESULTS_DIR: "/tmp/cartulary-results/run-a",
  },
  [
    "update-baselines",
    "--baseline-file",
    "/tmp/browser-baseline.json",
    "/tmp/cartulary-results/run-a",
  ],
);
assertArgs(
  "browser-e2e-duration-baseline-drift",
  {
    BROWSER_E2E_DURATION_BASELINE: "/tmp/browser-baseline.json",
    CARTULARY_TEST_RESULTS_DIR: "/tmp/cartulary-results",
    CARTULARY_TEST_RUN_ID: "run-a",
  },
  [
    "check-baseline-drift",
    "--baseline-file",
    "/tmp/browser-baseline.json",
    "/tmp/cartulary-results/run-a",
  ],
);
assertArgs(
  "service-backed-make-target-duration-baseline-drift",
  {
    CARTULARY_TEST_RESULTS_DIR: "/tmp/cartulary-results",
    CARTULARY_TEST_RUN_ID: "run-a",
    SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE: "/tmp/service-baseline.json",
  },
  [
    "check-drift",
    "--baseline-file",
    "/tmp/service-baseline.json",
    "/tmp/cartulary-results/run-a",
  ],
);
assertArgs(
  "harness-smoke-duration-baselines",
  {
    HARNESS_SMOKE_DURATION_BASELINE: "/tmp/harness-baseline.json",
    RESULTS_DIR: "/tmp/cartulary-results/run-a",
  },
  [
    "update",
    "--baseline-file",
    "/tmp/harness-baseline.json",
    "/tmp/cartulary-results/run-a",
  ],
);
assertArgs(
  "harness-smoke-duration-baseline-drift",
  {
    CARTULARY_TEST_RESULTS_DIR: "/tmp/cartulary-results",
    CARTULARY_TEST_RUN_ID: "run-a",
    HARNESS_SMOKE_DURATION_BASELINE: "/tmp/harness-baseline.json",
  },
  [
    "check-drift",
    "--baseline-file",
    "/tmp/harness-baseline.json",
    "/tmp/cartulary-results/run-a",
  ],
);
assertArgs("task-surface-report", { TASK_SURFACE_REPORT_ARGS: "--check --all" }, [
  "--check",
  "--all",
]);
assertArgs("target-plan", { PHASE: "phase4", TARGET: "backend-store", RESULTS_DIR: "/tmp/results" }, [
  "--target",
  "backend-store",
]);
assertUsage("explain-run", {}, "make explain-run RESULTS_DIR=<root|run-dir>");
assertUsage("phase-slice", {}, "make phase-slice PHASE=<phaseN|FE-PN>");
assertUsage("service-backed-slice", {}, "make service-backed-slice PHASE=<phaseN|FE-PN>");
assertUsage("go-test-duration-baselines", {}, "make go-test-duration-baselines RESULTS_DIR=<successful results dir> [PRUNE_OBSERVED_PACKAGES=1 requires full service-backed]");
assertUsage("browser-e2e-duration-baselines", {}, "make browser-e2e-duration-baselines RESULTS_DIR=<successful browser results dir>");
assertUsage("harness-smoke-duration-baselines", {}, "make harness-smoke-duration-baselines RESULTS_DIR=<successful harness results dir>");
assertUsage("scheduler-event-order-drift", { CARTULARY_TEST_RESULTS_DIR: "/tmp/results" }, "make scheduler-event-order-drift");
assertUsage("scheduler-summary-timing-drift", { CARTULARY_TEST_RESULTS_DIR: "/tmp/results" }, "make scheduler-summary-timing-drift");

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
  TASK_SURFACE_MANIFEST: "/tmp/task-surface.json",
  TARGET: "backend-store",
});
assert(targetPlanChildEnv.PATH === "/bin", "child env must preserve unrelated runtime env");
assert(!("PHASE" in targetPlanChildEnv), "target-plan child env must not expose undeclared PHASE");
assert(!("RESULTS_DIR" in targetPlanChildEnv), "target-plan child env must not expose undeclared RESULTS_DIR");
assert(!("TASK_SURFACE_MANIFEST" in targetPlanChildEnv), "target-plan child env must not expose internal manifest override");
assert(!("TARGET" in targetPlanChildEnv), "target-plan child env must not expose TARGET after args are built");

const taskGuideChildEnv = buildMakeNodeToolChildEnv("task-guide", {
  CARTULARY_TEST_RESULTS_DIR: "/tmp/results",
  JSON: "1",
  PHASE: "phase4",
  PHASE_NAMESPACE: "frontend",
  ROLE: "feature-dev",
});
assert(
  taskGuideChildEnv.CARTULARY_TEST_RESULTS_DIR === "/tmp/results",
  "task-guide child env must keep result-root runtime env",
);
assert(!("JSON" in taskGuideChildEnv), "task-guide child env must not expose JSON after args are built");
assert(!("PHASE" in taskGuideChildEnv), "task-guide child env must not expose PHASE after args are built");
assert(!("PHASE_NAMESPACE" in taskGuideChildEnv), "task-guide child env must not expose PHASE_NAMESPACE after args are built");
assert(!("ROLE" in taskGuideChildEnv), "task-guide child env must not expose ROLE after args are built");

const fallowChildEnv = buildMakeNodeToolChildEnv("frontend-fallow-static", {
  CARTULARY_TEST_RESULTS_DIR: "/tmp/results",
  CARTULARY_TEST_RUN_ID: "run-fallow",
  JSON: "1",
  NODE_BIN: "/tmp/node",
  NODE_RUNTIME_DIR: "/tmp/node-runtime",
  PHASE: "phase4",
  PNPM: "/tmp/pnpm",
});
assert(
  fallowChildEnv.CARTULARY_TEST_RESULTS_DIR === "/tmp/results",
  "frontend-fallow-static child env must keep result-root runtime env",
);
assert(
  fallowChildEnv.CARTULARY_TEST_RUN_ID === "run-fallow",
  "frontend-fallow-static child env must keep run-id runtime env",
);
assert(fallowChildEnv.NODE_BIN === "/tmp/node", "frontend-fallow-static child env must keep NODE_BIN");
assert(fallowChildEnv.PNPM === "/tmp/pnpm", "frontend-fallow-static child env must keep PNPM");
assert(!("JSON" in fallowChildEnv), "frontend-fallow-static child env must not expose JSON");
assert(!("PHASE" in fallowChildEnv), "frontend-fallow-static child env must not expose PHASE");

const phaseSliceChildEnv = buildMakeNodeToolChildEnv("phase-slice", {
  CARTULARY_TEST_RESULTS_DIR: "/tmp/results",
  CARTULARY_TEST_RUN_ID: "run-a",
  MAKE: "make",
  PHASE: "phase4",
  ROWS: "FE-I-P5-01",
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
assert(!("ROWS" in phaseSliceChildEnv), "phase-slice child env must not expose ROWS after args are built");
assert(!("TARGET" in phaseSliceChildEnv), "phase-slice child env must not expose unrelated TARGET");
EOF

set +e
missing_phase_output="$("$NODE_BIN" "$ROOT_DIR/tools/harness/core/run-make-node-tool-cli.mjs" explain-phase 2>&1)"
missing_phase_status=$?
set -e
if [[ "$missing_phase_status" -ne 2 ]]; then
  fail "missing phase launcher validation should exit 2"
fi
assert_contains "$missing_phase_output" "usage: make explain-phase PHASE=<phaseN|FE-PN>" "missing phase launcher usage"
