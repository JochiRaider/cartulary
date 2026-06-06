import { readFileSync } from "node:fs";
import path from "node:path";

const defaultResultsRoot = ".cartulary/test-results";
const defaultTaskSurfaceManifest = "tools/task_surface_manifest.json";
const retiredPublicPassthroughEnvNames = Object.freeze([
  "CARTULARY_EXECUTION_TOPOLOGY_MANIFEST",
  "CARTULARY_FIXTURE_THRESHOLD_MS",
  "CARTULARY_FIXTURE_TOP",
  "CARTULARY_TASK_SURFACE_MANIFEST",
  "EXECUTION_TOPOLOGY_MANIFEST",
  "SCHEDULER_MANIFEST",
  "TASK_SURFACE_MANIFEST",
]);

function value(env, name) {
  return env[name] ?? "";
}

function hasValue(env, name) {
  return value(env, name) !== "";
}

function optionalFlag(args, env, name, flag) {
  if (hasValue(env, name)) {
    args.push(flag, value(env, name));
  }
}

function jsonFlag(args, env) {
  if (value(env, "JSON") === "1") {
    args.push("--json");
  }
}

function resultDirForMode(env, mode) {
  if (!mode) {
    return "";
  }
  if (hasValue(env, "RESULTS_DIR")) {
    return value(env, "RESULTS_DIR");
  }
  if (mode === "required") {
    return "";
  }
  if (mode === "resultsRootDefault") {
    return value(env, "CARTULARY_TEST_RESULTS_DIR") || defaultResultsRoot;
  }
  if (mode === "currentRunDefault") {
    const root = value(env, "CARTULARY_TEST_RESULTS_DIR") || defaultResultsRoot;
    const runID = value(env, "CARTULARY_TEST_RUN_ID");
    return runID ? path.join(root, runID) : "";
  }
  throw new Error(`unsupported result-dir mode ${mode}`);
}

function splitPassthrough(valueToSplit, name) {
  const raw = String(valueToSplit ?? "").trim();
  if (!raw) {
    return [];
  }
  if (/["'\\]/.test(raw)) {
    throw new UsageError(`${name} does not support shell quoting; pass plain whitespace-separated flags`);
  }
  return raw.split(/\s+/).filter(Boolean);
}

export class UsageError extends Error {
  constructor(message, usage) {
    super(message);
    this.name = "UsageError";
    this.usage = usage;
  }
}

const phaseSliceRuntimeEnv = [
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
];

export const makeNodeTools = {
  "task-surface-report": {
    inputs: ["TASK_SURFACE_REPORT_ARGS"],
    script: "./scripts/print-task-surface-report.mjs",
    usage: "usage: make task-surface-report [TASK_SURFACE_REPORT_ARGS=--all]",
    buildArgs(env) {
      return splitPassthrough(value(env, "TASK_SURFACE_REPORT_ARGS"), "TASK_SURFACE_REPORT_ARGS");
    },
  },
  "task-guide": {
    inputs: ["ROLE", "PHASE", "PHASE_NAMESPACE", "JSON"],
    runtimeEnv: ["CARTULARY_TEST_RESULTS_DIR"],
    script: "./scripts/print-task-guide.mjs",
    usage: "usage: make task-guide [ROLE=<role>] [PHASE=phaseN] [PHASE_NAMESPACE=base|frontend] [JSON=1]",
    buildArgs(env) {
      const args = [];
      optionalFlag(args, env, "ROLE", "--role");
      optionalFlag(args, env, "PHASE", "--phase");
      optionalFlag(args, env, "PHASE_NAMESPACE", "--phase-namespace");
      jsonFlag(args, env);
      return args;
    },
  },
  "phase-slice": {
    inputs: ["PHASE", "PHASE_NAMESPACE", "ROWS", "JSON"],
    runtimeEnv: phaseSliceRuntimeEnv,
    script: "./scripts/run-phase-slice.mjs",
    usage: "usage: make phase-slice PHASE=<phaseN|FE-PN> [PHASE_NAMESPACE=base|frontend] [ROWS=<frontend-row-id,...>]",
    buildArgs(env) {
      if (!hasValue(env, "PHASE")) {
        throw new UsageError("PHASE is required", "usage: make phase-slice PHASE=<phaseN|FE-PN> [PHASE_NAMESPACE=base|frontend] [ROWS=<frontend-row-id,...>]");
      }
      const args = ["--phase", value(env, "PHASE"), "--mode", "phase"];
      optionalFlag(args, env, "PHASE_NAMESPACE", "--phase-namespace");
      optionalFlag(args, env, "ROWS", "--rows");
      jsonFlag(args, env);
      return args;
    },
  },
  "service-backed-slice": {
    inputs: ["PHASE", "PHASE_NAMESPACE", "ROWS", "JSON"],
    runtimeEnv: phaseSliceRuntimeEnv,
    script: "./scripts/run-phase-slice.mjs",
    usage: "usage: make service-backed-slice PHASE=<phaseN|FE-PN> [PHASE_NAMESPACE=base|frontend] [ROWS=<frontend-row-id,...>]",
    buildArgs(env) {
      if (!hasValue(env, "PHASE")) {
        throw new UsageError("PHASE is required", "usage: make service-backed-slice PHASE=<phaseN|FE-PN> [PHASE_NAMESPACE=base|frontend] [ROWS=<frontend-row-id,...>]");
      }
      const args = ["--phase", value(env, "PHASE"), "--mode", "service-backed"];
      optionalFlag(args, env, "PHASE_NAMESPACE", "--phase-namespace");
      optionalFlag(args, env, "ROWS", "--rows");
      jsonFlag(args, env);
      return args;
    },
  },
  "target-plan": {
    inputs: ["TARGET"],
    script: "./scripts/print-target-plan.mjs",
    usage: "usage: make target-plan [TARGET=<backend-go-target>]",
    buildArgs(env) {
      const args = [];
      optionalFlag(args, env, "TARGET", "--target");
      return args;
    },
  },
  "target-plan-json": {
    inputs: ["TARGET"],
    script: "./scripts/print-target-plan.mjs",
    usage: "usage: make target-plan-json [TARGET=<backend-go-target>]",
    buildArgs(env) {
      const args = ["--json"];
      optionalFlag(args, env, "TARGET", "--target");
      return args;
    },
  },
  "fixture-report": {
    inputs: ["FIXTURE_THRESHOLD_MS", "FIXTURE_TOP", "RUN_ID", "TARGET", "JSON"],
    script: "./scripts/print-fixture-report.mjs",
    resultDir: { mode: "resultsRootDefault", flag: "--results-dir" },
    usage:
      "usage: make fixture-report [RESULTS_DIR=<root|run-dir>] [RUN_ID=<id>] [TARGET=<target>] [JSON=1]",
    buildArgs(env) {
      const args = [
        "--threshold-ms",
        value(env, "FIXTURE_THRESHOLD_MS") || "30000",
        "--top",
        value(env, "FIXTURE_TOP") || "5",
      ];
      optionalFlag(args, env, "RUN_ID", "--run-id");
      optionalFlag(args, env, "TARGET", "--target");
      jsonFlag(args, env);
      return args;
    },
  },
  "explain-run": {
    inputs: ["DETAIL", "RUN_ID", "TARGET"],
    script: "./scripts/print-explain-run.mjs",
    resultDir: { mode: "required", flag: "--results-dir" },
    usage:
      "usage: make explain-run RESULTS_DIR=<root|run-dir> [RUN_ID=<id>] [TARGET=<target>] [DETAIL=summary|children|logs|progress]",
    buildArgs(env) {
      const args = ["--detail", value(env, "DETAIL") || "summary"];
      optionalFlag(args, env, "RUN_ID", "--run-id");
      optionalFlag(args, env, "TARGET", "--target");
      return args;
    },
  },
  "explain-phase": {
    inputs: ["PHASE", "PHASE_NAMESPACE", "JSON"],
    script: "./scripts/print-explain-phase.mjs",
    usage: "usage: make explain-phase PHASE=<phaseN|FE-PN> [PHASE_NAMESPACE=base|frontend]",
    buildArgs(env) {
      if (!hasValue(env, "PHASE")) {
        throw new UsageError("PHASE is required", "usage: make explain-phase PHASE=<phaseN|FE-PN> [PHASE_NAMESPACE=base|frontend]");
      }
      const args = ["--phase", value(env, "PHASE")];
      optionalFlag(args, env, "PHASE_NAMESPACE", "--phase-namespace");
      jsonFlag(args, env);
      return args;
    },
  },
  "explain-target": {
    inputs: ["TARGET", "DETAIL", "JSON"],
    script: "./scripts/print-explain-target.mjs",
    usage: "usage: make explain-target TARGET=<target> [DETAIL=summary|rows|artifacts]",
    buildArgs(env) {
      if (!hasValue(env, "TARGET")) {
        throw new UsageError(
          "TARGET is required",
          "usage: make explain-target TARGET=<target> [DETAIL=summary|rows|artifacts]",
        );
      }
      const args = ["--target", value(env, "TARGET"), "--detail", value(env, "DETAIL") || "summary"];
      jsonFlag(args, env);
      return args;
    },
  },
  "go-test-duration-baselines": {
    inputs: ["PRUNE_OBSERVED_PACKAGES", "ALLOW_COMMAND_OVERHEAD_DECREASE", "GO_TEST_DURATION_BASELINE"],
    script: "./scripts/update-go-test-durations.mjs",
    resultDir: { mode: "required", positional: true },
    usage: "usage: make go-test-duration-baselines RESULTS_DIR=<successful results dir> [PRUNE_OBSERVED_PACKAGES=1 requires full service-backed]",
    buildArgs(env) {
      const args = [];
      if (value(env, "PRUNE_OBSERVED_PACKAGES") === "1") {
        args.push("--prune-observed-packages");
      }
      if (value(env, "ALLOW_COMMAND_OVERHEAD_DECREASE") === "1") {
        args.push("--allow-command-overhead-decrease");
      }
      optionalFlag(args, env, "GO_TEST_DURATION_BASELINE", "--baseline-file");
      return args;
    },
  },
  "go-test-duration-baseline-coverage": {
    inputs: ["GO_TEST_DURATION_BASELINE"],
    script: "./scripts/check-go-test-duration-baseline-coverage.mjs",
    usage: "usage: make go-test-duration-baseline-coverage [GO_TEST_DURATION_BASELINE=<path>]",
    buildArgs(env) {
      const args = [];
      optionalFlag(args, env, "GO_TEST_DURATION_BASELINE", "--baseline-file");
      return args;
    },
  },
  "go-test-duration-baseline-drift": {
    inputs: ["GO_TEST_DURATION_BASELINE"],
    script: "./scripts/check-go-test-duration-baseline-drift.mjs",
    resultDir: { mode: "currentRunDefault", positional: true },
    usage:
      "usage: make go-test-duration-baseline-drift [RESULTS_DIR=<dir>] [GO_TEST_DURATION_BASELINE=<path>]",
    buildArgs(env) {
      const args = [];
      optionalFlag(args, env, "GO_TEST_DURATION_BASELINE", "--baseline-file");
      return args;
    },
  },
  "browser-e2e-duration-baselines": {
    inputs: ["BROWSER_E2E_DURATION_BASELINE"],
    script: "./scripts/lib/browser-shard-plan.mjs",
    resultDir: { mode: "required", positional: true },
    usage: "usage: make browser-e2e-duration-baselines RESULTS_DIR=<successful browser results dir>",
    buildArgs(env) {
      const args = ["update-baselines"];
      optionalFlag(args, env, "BROWSER_E2E_DURATION_BASELINE", "--baseline-file");
      return args;
    },
  },
  "browser-e2e-duration-baseline-drift": {
    inputs: ["BROWSER_E2E_DURATION_BASELINE"],
    script: "./scripts/lib/browser-shard-plan.mjs",
    resultDir: { mode: "currentRunDefault", positional: true },
    usage:
      "usage: make browser-e2e-duration-baseline-drift [RESULTS_DIR=<dir>] [BROWSER_E2E_DURATION_BASELINE=<path>]",
    buildArgs(env) {
      const args = ["check-baseline-drift"];
      optionalFlag(args, env, "BROWSER_E2E_DURATION_BASELINE", "--baseline-file");
      return args;
    },
  },
  "service-backed-make-target-duration-baselines": {
    inputs: ["SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE"],
    script: "./scripts/service-backed-make-target-durations.mjs",
    resultDir: { mode: "required", positional: true },
    usage:
      "usage: make service-backed-make-target-duration-baselines RESULTS_DIR=<successful scheduler results dir>",
    buildArgs(env) {
      const args = ["update"];
      optionalFlag(
        args,
        env,
        "SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE",
        "--baseline-file",
      );
      return args;
    },
  },
  "service-backed-make-target-duration-baseline-drift": {
    inputs: [
      "SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE",
    ],
    script: "./scripts/service-backed-make-target-durations.mjs",
    resultDir: { mode: "currentRunDefault", positional: true },
    usage:
      "usage: make service-backed-make-target-duration-baseline-drift [RESULTS_DIR=<dir>]",
    buildArgs(env) {
      const args = ["check-drift"];
      optionalFlag(
        args,
        env,
        "SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE",
        "--baseline-file",
      );
      return args;
    },
  },
  "harness-smoke-duration-baselines": {
    inputs: ["HARNESS_SMOKE_DURATION_BASELINE"],
    script: "./scripts/harness-smoke-durations.mjs",
    resultDir: { mode: "required", positional: true },
    usage:
      "usage: make harness-smoke-duration-baselines RESULTS_DIR=<successful harness results dir>",
    buildArgs(env) {
      const args = ["update"];
      optionalFlag(args, env, "HARNESS_SMOKE_DURATION_BASELINE", "--baseline-file");
      return args;
    },
  },
  "harness-smoke-duration-baseline-drift": {
    inputs: ["HARNESS_SMOKE_DURATION_BASELINE"],
    script: "./scripts/harness-smoke-durations.mjs",
    resultDir: { mode: "currentRunDefault", positional: true },
    usage: "usage: make harness-smoke-duration-baseline-drift [RESULTS_DIR=<dir>]",
    buildArgs(env) {
      const args = ["check-drift"];
      optionalFlag(args, env, "HARNESS_SMOKE_DURATION_BASELINE", "--baseline-file");
      return args;
    },
  },
  "scheduler-event-order-drift": {
    inputs: ["TARGET"],
    script: "./scripts/check-scheduler-event-order-drift.mjs",
    resultDir: { mode: "currentRunDefault", positional: true },
    usage: "usage: make scheduler-event-order-drift [RESULTS_DIR=<dir>] [TARGET=<target>]",
    buildArgs(env) {
      const args = [];
      optionalFlag(args, env, "TARGET", "--target");
      return args;
    },
  },
  "scheduler-summary-timing-drift": {
    inputs: ["TARGET", "SCHEDULER_WARM_CHECK_BUDGET_MS", "SCHEDULER_WARM_CHECK_BALANCE_RATIO"],
    script: "./scripts/check-scheduler-summary-timing-drift.mjs",
    resultDir: { mode: "currentRunDefault", positional: true },
    usage:
      "usage: make scheduler-summary-timing-drift [RESULTS_DIR=<dir>] [TARGET=<target>] [SCHEDULER_WARM_CHECK_BUDGET_MS=<ms>] [SCHEDULER_WARM_CHECK_BALANCE_RATIO=<ratio>]",
    buildArgs(env) {
      const args = [];
      optionalFlag(args, env, "TARGET", "--target");
      optionalFlag(args, env, "SCHEDULER_WARM_CHECK_BUDGET_MS", "--warm-check-budget-ms");
      optionalFlag(args, env, "SCHEDULER_WARM_CHECK_BALANCE_RATIO", "--warm-check-balance-ratio");
      return args;
    },
  },
};

function uniqueNames(names) {
  return Array.from(new Set(names));
}

function loadTaskSurfaceManifest(env = process.env) {
  const configured =
    env.TASK_SURFACE_MANIFEST ||
    env.CARTULARY_TASK_SURFACE_MANIFEST ||
    defaultTaskSurfaceManifest;
  const file = path.isAbsolute(configured)
    ? configured
    : path.join(process.cwd(), configured);
  try {
    return JSON.parse(readFileSync(file, "utf8"));
  } catch {
    return null;
  }
}

function contractInputNames(name, env = process.env) {
  const manifest = loadTaskSurfaceManifest(env);
  const entry = manifest?.targets?.find((candidate) => candidate.name === name);
  const inputs = entry?.input_contract?.inputs;
  if (Array.isArray(inputs)) {
    return inputs.map((input) => input.name);
  }
  return makeNodeTool(name).inputs ?? [];
}

function publicContractInputNames(env = process.env) {
  const manifest = loadTaskSurfaceManifest(env);
  if (!manifest?.targets) {
    return [];
  }
  return manifest.targets.flatMap((entry) =>
    entry?.target_class === "public"
      ? (entry.input_contract?.inputs ?? []).map((input) => input.name)
      : [],
  );
}

function resultDirMakeEnvVars(resultDir) {
  if (!resultDir) {
    return [];
  }
  if (resultDir.mode === "required") {
    return ["RESULTS_DIR"];
  }
  if (resultDir.mode === "resultsRootDefault") {
    return ["RESULTS_DIR", "CARTULARY_TEST_RESULTS_DIR"];
  }
  if (resultDir.mode === "currentRunDefault") {
    return ["RESULTS_DIR", "CARTULARY_TEST_RESULTS_DIR", "CARTULARY_TEST_RUN_ID"];
  }
  throw new Error(`unsupported result-dir mode ${resultDir.mode}`);
}

function makeNodeTool(name) {
  const tool = makeNodeTools[name];
  if (!tool) {
    throw new UsageError(
      `unknown make node tool ${name}`,
      `usage: run-make-node-tool.mjs <${makeNodeToolNames().join("|")}>`,
    );
  }
  return tool;
}

function scopedEnv(env, allowedNames, label) {
  const allowed = new Set(allowedNames);
  return new Proxy(env, {
    get(target, property) {
      if (typeof property === "string" && !allowed.has(property)) {
        throw new Error(`${label} read undeclared Make env var ${property}`);
      }
      return target[property];
    },
  });
}

export function makeNodeToolNames() {
  return Object.keys(makeNodeTools).sort();
}

export function hasMakeNodeTool(name) {
  return Object.hasOwn(makeNodeTools, name);
}

export function makeNodeToolRuntimeEnvVars(name) {
  return uniqueNames(makeNodeTool(name).runtimeEnv ?? []);
}

export function makeNodeToolResultDirMakeEnvVars(name) {
  return resultDirMakeEnvVars(makeNodeTool(name).resultDir);
}

export function makeNodeToolMakeEnvVars(name) {
  const tool = makeNodeTool(name);
  return uniqueNames([
    ...contractInputNames(name),
    ...makeNodeToolResultDirMakeEnvVars(name),
    ...(tool.runtimeEnv ?? []),
  ]);
}

export function makeNodeToolKnownEnvVars() {
  return uniqueNames([
    ...makeNodeToolNames().flatMap((name) => makeNodeToolMakeEnvVars(name)),
    ...publicContractInputNames(),
    ...retiredPublicPassthroughEnvNames,
  ]);
}

export function buildMakeNodeToolChildEnv(name, env = process.env) {
  const childEnv = { ...env };
  const runtimeEnv = new Set(makeNodeToolRuntimeEnvVars(name));
  for (const envVar of makeNodeToolKnownEnvVars()) {
    if (!runtimeEnv.has(envVar)) {
      delete childEnv[envVar];
    }
  }
  return childEnv;
}

export function buildMakeNodeToolInvocation(name, env = process.env) {
  const tool = makeNodeTool(name);
  const declaredInputs = contractInputNames(name, env);

  const args = tool.buildArgs(scopedEnv(env, declaredInputs, name));
  if (tool.resultDir) {
    const resultsDir = resultDirForMode(
      scopedEnv(env, resultDirMakeEnvVars(tool.resultDir), `${name} result-dir`),
      tool.resultDir.mode,
    );
    if (!resultsDir) {
      throw new UsageError("RESULTS_DIR is required", tool.usage);
    }
    if (tool.resultDir.positional) {
      args.push(resultsDir);
    } else {
      args.unshift(tool.resultDir.flag, resultsDir);
    }
  }
  return {
    args,
    runtimeEnv: makeNodeToolRuntimeEnvVars(name),
    script: tool.script,
    usage: tool.usage,
  };
}
