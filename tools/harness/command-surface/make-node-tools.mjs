import { readFileSync } from "node:fs";
import path from "node:path";

const defaultResultsRoot = ".cartulary/test-results";
const defaultTaskSurfaceManifest = "tools/task_surface_manifest.json";
const retiredPublicPassthroughEnvNames = Object.freeze([
  "BROWSER_A11Y_RESULTS_DIR",
  "BROWSER_MEASUREMENT_RESULTS_DIR",
  "BROWSER_SUPPORT_RESULTS_DIR",
  "BROWSER_VISUAL_RESULTS_DIR",
  "CARTULARY_EXECUTION_TOPOLOGY_MANIFEST",
  "CARTULARY_FIXTURE_THRESHOLD_MS",
  "CARTULARY_FIXTURE_TOP",
  "CARTULARY_TASK_SURFACE_MANIFEST",
  "CHECK_RESULTS_DIR",
  "EXECUTION_TOPOLOGY_MANIFEST",
  "GOCACHE",
  "GOMODCACHE",
  "GOTMPDIR",
  "SCHEDULER_MANIFEST",
  "TASK_SURFACE_MANIFEST",
  "VITEST_FLAGS",
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

function inputSourceMap(env) {
  return new Map(
    String(env.CARTULARY_MAKE_INPUT_SOURCES ?? "")
      .split(/\s+/u)
      .filter(Boolean)
      .map((entry) => {
        const separator = entry.indexOf("=");
        return separator === -1
          ? [entry, "unset"]
          : [entry.slice(0, separator), entry.slice(separator + 1)];
      }),
  );
}

function inputWasProvided(metadata, name) {
  return new Set(["cli", "env", "file"]).has(metadata.inputSources.get(name));
}

export class UsageError extends Error {
  constructor(message, usage) {
    super(message);
    this.name = "UsageError";
    this.usage = usage;
  }
}

function ownerDiagnosticTool(mode, target) {
  return {
    inputs: mode === "task-guide" ? ["ROLE", "OWNER", "JSON"] : ["OWNER", "JSON"],
    script: "./tools/harness/diagnostics/owner-diagnostics-cli.mjs",
    usage: mode === "task-guide"
      ? `usage: make ${target} ROLE=module-author OWNER=<owner-id> [JSON=1]`
      : `usage: make ${target} OWNER=<owner-id> [JSON=1]`,
    buildArgs(env) {
      const args = ["--mode", mode, "--owner", value(env, "OWNER")];
      if (mode === "task-guide") args.push("--role", value(env, "ROLE"));
      if (value(env, "JSON") === "1") args.push("--json");
      else if (value(env, "JSON") !== "") args.push("--json-value", value(env, "JSON"));
      return args;
    },
  };
}

export const makeNodeTools = {
  "author-test-row-id": {
    inputs: ["FAMILY_ID", "CLAIM", "SELECTOR_KEY"],
    script: "./tools/harness/test-catalog/row-id-authoring-cli.mjs",
    usage:
      "usage: make author-test-row-id FAMILY_ID=<owner.family> CLAIM=<semantic-claim> SELECTOR_KEY=<stable-selector-key>",
    buildArgs(env) {
      return [
        "--family-id",
        value(env, "FAMILY_ID"),
        "--claim",
        value(env, "CLAIM"),
        "--selector-key",
        value(env, "SELECTOR_KEY"),
      ];
    },
  },
  "explain-test-owner": ownerDiagnosticTool("explain", "explain-test-owner"),
  "task-guide": ownerDiagnosticTool("task-guide", "task-guide"),
  "test-evidence-audit": {
    inputs: ["OWNER", "EVIDENCE_ROOTS_FILE"],
    runtimeEnv: ["CARTULARY_TEST_RESULTS_DIR", "CARTULARY_TEST_RUN_ID"],
    script: "./tools/harness/observability/canonical-evidence-audit-cli.mjs",
    usage: "usage: make test-evidence-audit OWNER=<owner-id> EVIDENCE_ROOTS_FILE=<path>",
    buildArgs(env) {
      return [
        "--owner",
        value(env, "OWNER"),
        "--evidence-roots-file",
        value(env, "EVIDENCE_ROOTS_FILE"),
      ];
    },
  },
  "task-surface-report": {
    inputs: ["TASK_SURFACE_REPORT_ARGS"],
    script: "./tools/harness/generated-artifacts/task-surface-report-cli.mjs",
    usage: "usage: make task-surface-report [TASK_SURFACE_REPORT_ARGS='--all|--check|--check --all']",
    buildArgs(env) {
      return splitPassthrough(value(env, "TASK_SURFACE_REPORT_ARGS"), "TASK_SURFACE_REPORT_ARGS");
    },
  },
  "target-plan": {
    inputs: ["TARGET"],
    script: "./tools/harness/diagnostics/target-plan-cli.mjs",
    usage: "usage: make target-plan [TARGET=<backend-go-target>]",
    buildArgs(env) {
      const args = [];
      optionalFlag(args, env, "TARGET", "--target");
      return args;
    },
  },
  "target-plan-json": {
    inputs: ["TARGET"],
    script: "./tools/harness/diagnostics/target-plan-cli.mjs",
    usage: "usage: make target-plan-json [TARGET=<backend-go-target>]",
    buildArgs(env) {
      const args = ["--json"];
      optionalFlag(args, env, "TARGET", "--target");
      return args;
    },
  },
  "fixture-report": {
    inputs: ["FIXTURE_THRESHOLD_MS", "FIXTURE_TOP", "RUN_ID", "TARGET", "JSON"],
    script: "./tools/harness/diagnostics/fixture-report-cli.mjs",
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
    script: "./tools/harness/diagnostics/explain-run-cli.mjs",
    resultDir: { mode: "required", flag: "--results-dir" },
    usage:
      "usage: make explain-run RESULTS_DIR=<root|run-dir> [RUN_ID=<id>] [TARGET=<target>] [DETAIL=summary|children|logs|progress|accounting|performance]",
    buildArgs(env) {
      const args = ["--detail", value(env, "DETAIL") || "summary"];
      optionalFlag(args, env, "RUN_ID", "--run-id");
      optionalFlag(args, env, "TARGET", "--target");
      return args;
    },
  },
  "harness-observability-check": {
    inputs: ["RUN_ID"],
    script: "./tools/harness/observability/observability-check-cli.mjs",
    resultDir: { mode: "required", flag: "--results-dir" },
    usage: "usage: make harness-observability-check RESULTS_DIR=<root|run-dir> [RUN_ID=<id>]",
    buildArgs(env) {
      const args = [];
      optionalFlag(args, env, "RUN_ID", "--run-id");
      return args;
    },
  },
  "harness-otel-export": {
    inputs: ["RUN_ID", "HARNESS_OTLP_ENDPOINT", "HARNESS_OTLP_HEADERS_FILE"],
    script: "./tools/harness/observability/otel-export-cli.mjs",
    resultDir: { mode: "required", flag: "--results-dir" },
    usage: "usage: make harness-otel-export RESULTS_DIR=<root|run-dir> [RUN_ID=<id>] HARNESS_OTLP_ENDPOINT=<url> [HARNESS_OTLP_HEADERS_FILE=<0600-json-file>]",
    buildArgs(env) {
      const args = ["--endpoint", value(env, "HARNESS_OTLP_ENDPOINT")];
      optionalFlag(args, env, "RUN_ID", "--run-id");
      optionalFlag(args, env, "HARNESS_OTLP_HEADERS_FILE", "--headers-file");
      return args;
    },
  },
  "harness-performance-check": {
    inputs: ["EVIDENCE_ROOTS_FILE"],
    script: "./tools/harness/observability/canonical-performance-cli.mjs",
    usage: "usage: make harness-performance-check EVIDENCE_ROOTS_FILE=<manifest>",
    buildArgs(env) {
      return ["check", "--evidence-roots-file", value(env, "EVIDENCE_ROOTS_FILE")];
    },
  },
  "harness-public-target-duration-baselines": {
    inputs: ["EVIDENCE_ROOTS_FILE"],
    script: "./tools/harness/observability/canonical-performance-cli.mjs",
    usage: "usage: make harness-public-target-duration-baselines EVIDENCE_ROOTS_FILE=<baseline-window>",
    buildArgs(env) {
      return ["publish", "--evidence-roots-file", value(env, "EVIDENCE_ROOTS_FILE")];
    },
  },
  "explain-target": {
    inputs: ["TARGET", "DETAIL", "JSON"],
    script: "./tools/harness/diagnostics/explain-target-cli.mjs",
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
  "frontend-fallow-static": {
    inputs: [],
    runtimeEnv: [
      "CARTULARY_TEST_RESULTS_DIR",
      "CARTULARY_TEST_RUN_ID",
      "NODE_BIN",
      "NODE_RUNTIME_DIR",
      "PNPM",
    ],
    script: "./tools/harness/static-analysis/fallow-static-cli.mjs",
    usage: "usage: make frontend-fallow-static",
    buildArgs() {
      return [];
    },
  },
  "go-test-duration-baselines": {
    inputs: [],
    script: "./tools/harness/observability/canonical-performance-cli.mjs",
    resultDir: { mode: "required", flag: "--results-dir" },
    usage: "usage: make go-test-duration-baselines RESULTS_DIR=<successful canonical results dir>",
    buildArgs() {
      return ["observe", "--view", "backend_go"];
    },
  },
  "go-test-duration-baseline-coverage": {
    inputs: [],
    runtimeEnv: ["CARTULARY_TEST_RESULTS_DIR", "CARTULARY_TEST_RUN_ID"],
    script: "./tools/harness/observability/canonical-performance-cli.mjs",
    usage: "usage: make go-test-duration-baseline-coverage",
    buildArgs() {
      return ["coverage"];
    },
  },
  "go-test-duration-baseline-drift": {
    inputs: [],
    script: "./tools/harness/observability/canonical-performance-cli.mjs",
    resultDir: { mode: "currentRunDefault", flag: "--results-dir" },
    usage: "usage: make go-test-duration-baseline-drift [RESULTS_DIR=<dir>]",
    buildArgs() {
      return ["observe", "--view", "backend_go"];
    },
  },
  "browser-e2e-duration-baselines": {
    inputs: [],
    script: "./tools/harness/observability/canonical-performance-cli.mjs",
    resultDir: { mode: "required", flag: "--results-dir" },
    usage: "usage: make browser-e2e-duration-baselines RESULTS_DIR=<successful canonical results dir>",
    buildArgs() {
      return ["observe", "--view", "browser"];
    },
  },
  "browser-e2e-duration-baseline-drift": {
    inputs: [],
    script: "./tools/harness/observability/canonical-performance-cli.mjs",
    resultDir: { mode: "currentRunDefault", flag: "--results-dir" },
    usage: "usage: make browser-e2e-duration-baseline-drift [RESULTS_DIR=<dir>]",
    buildArgs() {
      return ["observe", "--view", "browser"];
    },
  },
  "service-backed-make-target-duration-baselines": {
    inputs: [],
    script: "./tools/harness/observability/canonical-performance-cli.mjs",
    resultDir: { mode: "required", flag: "--results-dir" },
    usage:
      "usage: make service-backed-make-target-duration-baselines RESULTS_DIR=<successful scheduler results dir>",
    buildArgs() {
      return ["observe", "--view", "service_backed"];
    },
  },
  "service-backed-make-target-duration-baseline-drift": {
    inputs: [],
    script: "./tools/harness/observability/canonical-performance-cli.mjs",
    resultDir: { mode: "currentRunDefault", flag: "--results-dir" },
    usage:
      "usage: make service-backed-make-target-duration-baseline-drift [RESULTS_DIR=<dir>]",
    buildArgs() {
      return ["observe", "--view", "service_backed"];
    },
  },
  "harness-smoke-duration-baselines": {
    inputs: [],
    script: "./tools/harness/observability/canonical-performance-cli.mjs",
    resultDir: { mode: "required", flag: "--results-dir" },
    usage:
      "usage: make harness-smoke-duration-baselines RESULTS_DIR=<successful harness results dir>",
    buildArgs() {
      return ["observe", "--view", "harness"];
    },
  },
  "harness-smoke-duration-baseline-drift": {
    inputs: [],
    script: "./tools/harness/observability/canonical-performance-cli.mjs",
    resultDir: { mode: "currentRunDefault", flag: "--results-dir" },
    usage: "usage: make harness-smoke-duration-baseline-drift [RESULTS_DIR=<dir>]",
    buildArgs() {
      return ["observe", "--view", "harness"];
    },
  },
  "scheduler-event-order-drift": {
    inputs: ["TARGET"],
    script: "./tools/harness/diagnostics/scheduler-event-order-drift-cli.mjs",
    resultDir: { mode: "currentRunDefault", positional: true },
    usage: "usage: make scheduler-event-order-drift [RESULTS_DIR=<dir>] [TARGET=<target>]",
    buildArgs(env) {
      const args = [];
      optionalFlag(args, env, "TARGET", "--target");
      return args;
    },
  },
  "scheduler-summary-timing-drift": {
    inputs: ["TARGET"],
    script: "./tools/harness/diagnostics/scheduler-summary-timing-drift-cli.mjs",
    resultDir: { mode: "currentRunDefault", positional: true },
    usage: "usage: make scheduler-summary-timing-drift [RESULTS_DIR=<dir>] [TARGET=<target>]",
    buildArgs(env) {
      const args = [];
      optionalFlag(args, env, "TARGET", "--target");
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
  return [
    ...(manifest.global_inputs ?? []).map((input) => input.name),
    ...manifest.targets.flatMap((entry) =>
      entry?.target_class === "public"
        ? (entry.input_contract?.inputs ?? []).map((input) => input.name)
        : [],
    ),
  ];
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
  delete childEnv.CARTULARY_MAKE_INPUT_SOURCES;
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

  const args = tool.buildArgs(scopedEnv(env, declaredInputs, name), {
    inputSources: inputSourceMap(env),
  });
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
