import path from "node:path";

const defaultResultsRoot = ".cartulary/test-results";

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

export const makeNodeToolEnvVars = [
  "BASELINE_FILE",
  "CARTULARY_TEST_RESULTS_DIR",
  "CARTULARY_TEST_RUN_ID",
  "DETAIL",
  "FIXTURE_THRESHOLD_MS",
  "FIXTURE_TOP",
  "JSON",
  "PHASE",
  "PRUNE_OBSERVED_PACKAGES",
  "RESULTS_DIR",
  "ROLE",
  "RUN_ID",
  "SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE",
  "SERVICE_BACKED_SCHEDULE_MANIFEST",
  "SERVICE_BACKED_SCHEDULE_PROFILE",
  "TARGET",
  "TASK_SURFACE_REPORT_ARGS",
];

export const makeNodeTools = {
  "task-surface-report": {
    script: "./scripts/print-task-surface-report.mjs",
    usage: "usage: make task-surface-report [TASK_SURFACE_REPORT_ARGS=--all]",
    buildArgs(env) {
      return splitPassthrough(value(env, "TASK_SURFACE_REPORT_ARGS"), "TASK_SURFACE_REPORT_ARGS");
    },
  },
  "task-guide": {
    script: "./scripts/print-task-guide.mjs",
    usage: "usage: make task-guide [ROLE=<role>] [PHASE=phaseN] [JSON=1]",
    buildArgs(env) {
      const args = [];
      optionalFlag(args, env, "ROLE", "--role");
      optionalFlag(args, env, "PHASE", "--phase");
      jsonFlag(args, env);
      return args;
    },
  },
  "target-plan": {
    script: "./scripts/print-target-plan.mjs",
    usage: "usage: make target-plan",
    buildArgs() {
      return [];
    },
  },
  "target-plan-json": {
    script: "./scripts/print-target-plan.mjs",
    usage: "usage: make target-plan-json",
    buildArgs() {
      return ["--json"];
    },
  },
  "fixture-report": {
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
    script: "./scripts/print-explain-run.mjs",
    resultDir: { mode: "required", flag: "--results-dir" },
    usage:
      "usage: make explain-run RESULTS_DIR=<root|run-dir> [RUN_ID=<id>] [TARGET=<target>] [DETAIL=summary|children|logs]",
    buildArgs(env) {
      const args = ["--detail", value(env, "DETAIL") || "summary"];
      optionalFlag(args, env, "RUN_ID", "--run-id");
      optionalFlag(args, env, "TARGET", "--target");
      return args;
    },
  },
  "explain-phase": {
    script: "./scripts/print-explain-phase.mjs",
    usage: "usage: make explain-phase PHASE=<phaseN>",
    buildArgs(env) {
      if (!hasValue(env, "PHASE")) {
        throw new UsageError("PHASE is required", "usage: make explain-phase PHASE=<phaseN>");
      }
      const args = ["--phase", value(env, "PHASE")];
      jsonFlag(args, env);
      return args;
    },
  },
  "explain-target": {
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
    script: "./scripts/update-go-test-durations.mjs",
    resultDir: { mode: "required", positional: true },
    usage: "usage: make go-test-duration-baselines RESULTS_DIR=<successful test results dir>",
    buildArgs(env) {
      const args = [];
      if (value(env, "PRUNE_OBSERVED_PACKAGES") === "1") {
        args.push("--prune-observed-packages");
      }
      optionalFlag(args, env, "BASELINE_FILE", "--baseline-file");
      return args;
    },
  },
  "go-test-duration-baseline-coverage": {
    script: "./scripts/check-go-test-duration-baseline-coverage.mjs",
    usage: "usage: make go-test-duration-baseline-coverage [BASELINE_FILE=<path>]",
    buildArgs(env) {
      const args = [];
      optionalFlag(args, env, "BASELINE_FILE", "--baseline-file");
      return args;
    },
  },
  "go-test-duration-baseline-drift": {
    script: "./scripts/check-go-test-duration-baseline-drift.mjs",
    resultDir: { mode: "currentRunDefault", positional: true },
    usage: "usage: make go-test-duration-baseline-drift [RESULTS_DIR=<dir>] [BASELINE_FILE=<path>]",
    buildArgs(env) {
      const args = [];
      optionalFlag(args, env, "BASELINE_FILE", "--baseline-file");
      return args;
    },
  },
  "browser-e2e-duration-baseline-drift": {
    script: "./scripts/lib/browser-shard-plan.mjs",
    resultDir: { mode: "currentRunDefault", positional: true },
    usage:
      "usage: make browser-e2e-duration-baseline-drift [RESULTS_DIR=<dir>] [BASELINE_FILE=<path>]",
    buildArgs(env) {
      const args = ["check-baseline-drift"];
      optionalFlag(args, env, "BASELINE_FILE", "--baseline-file");
      return args;
    },
  },
  "service-backed-make-target-duration-baselines": {
    script: "./scripts/service-backed-make-target-durations.mjs",
    resultDir: { mode: "required", positional: true },
    usage:
      "usage: make service-backed-make-target-duration-baselines RESULTS_DIR=<successful test results dir>",
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
      optionalFlag(args, env, "SERVICE_BACKED_SCHEDULE_PROFILE", "--profile");
      optionalFlag(args, env, "SERVICE_BACKED_SCHEDULE_MANIFEST", "--schedule-manifest");
      return args;
    },
  },
  "scheduler-event-order-drift": {
    script: "./scripts/check-scheduler-event-order-drift.mjs",
    resultDir: { mode: "currentRunDefault", positional: true },
    usage: "usage: make scheduler-event-order-drift [RESULTS_DIR=<dir>] [TARGET=<target>]",
    buildArgs(env) {
      const args = [];
      optionalFlag(args, env, "TARGET", "--target");
      return args;
    },
  },
};

export function makeNodeToolNames() {
  return Object.keys(makeNodeTools).sort();
}

export function hasMakeNodeTool(name) {
  return Object.hasOwn(makeNodeTools, name);
}

export function buildMakeNodeToolInvocation(name, env = process.env) {
  const tool = makeNodeTools[name];
  if (!tool) {
    throw new UsageError(
      `unknown make node tool ${name}`,
      `usage: run-make-node-tool.mjs <${makeNodeToolNames().join("|")}>`,
    );
  }

  const args = tool.buildArgs(env);
  if (tool.resultDir) {
    const resultsDir = resultDirForMode(env, tool.resultDir.mode);
    if (!resultsDir) {
      throw new UsageError("RESULTS_DIR is required", tool.usage);
    }
    if (tool.resultDir.positional) {
      args.push(resultsDir);
    } else {
      args.unshift(tool.resultDir.flag, resultsDir);
    }
  }
  return { args, script: tool.script, usage: tool.usage };
}
