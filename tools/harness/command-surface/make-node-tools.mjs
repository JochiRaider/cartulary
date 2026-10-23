import { readFileSync } from "node:fs";
import path from "node:path";
import { resolveHarnessConfig } from "../contract/harness-contract.mjs";

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
  "GOROOT",
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
    inputs: mode === "task-guide" ? ["OWNER", "DETAIL", "JSON"] : ["OWNER", "JSON"],
    script: "./tools/harness/diagnostics/owner-diagnostics-cli.mjs",
    usage: mode === "task-guide"
      ? `usage: make ${target} OWNER=<owner-id> [DETAIL=summary|rows|coverage] [JSON=1]`
      : `usage: make ${target} OWNER=<owner-id> [JSON=1]`,
    buildArgs(env) {
      const args = ["--mode", mode, "--owner", value(env, "OWNER")];
      if (mode === "task-guide") args.push("--detail", value(env, "DETAIL") || "summary");
      if (value(env, "JSON") === "1") args.push("--json");
      else if (value(env, "JSON") !== "") args.push("--json-value", value(env, "JSON"));
      return args;
    },
  };
}

const makeNodeTools = {
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
  "target-plan": {
    inputs: ["TARGET", "INPUTS", "DETAIL", "JSON"],
    script: "./tools/harness/diagnostics/target-plan-cli.mjs",
    usage: "usage: make target-plan TARGET=<command> [INPUTS=<json-file>] [DETAIL=summary|graph|rows|artifacts] [JSON=1]",
    buildArgs(env) {
      const args = ["--target", value(env, "TARGET"), "--view", value(env, "DETAIL") || "summary"];
      optionalFlag(args, env, "INPUTS", "--inputs");
      jsonFlag(args, env);
      return args;
    },
  },
  "explain-run": {
    inputs: ["DETAIL", "RUN_ID", "TARGET", "JSON", "COMPARE_RESULTS_DIR", "COMPARISON", "FIXTURE_THRESHOLD_MS", "FIXTURE_TOP", "VIEW", "UNIT_ID", "OFFSET", "LIMIT"],
    script: "./tools/harness/diagnostics/explain-run-cli.mjs",
    resultDir: { mode: "required", flag: "--results-dir" },
    usage:
      "usage: make explain-run RESULTS_DIR=<root|run-dir> [RUN_ID=<id>] [TARGET=<target>] [DETAIL=summary|children|logs|progress|accounting|performance|fixtures|critical-path] [JSON=1] [COMPARE_RESULTS_DIR=<exact-run-dir> COMPARISON=equivalent|instrumentation]",
    buildArgs(env) {
      const args = ["--detail", value(env, "DETAIL") || "summary"];
      if (value(env, "JSON") === "1") args.push("--json");
      optionalFlag(args, env, "RUN_ID", "--run-id");
      optionalFlag(args, env, "TARGET", "--target");
      optionalFlag(args, env, "COMPARE_RESULTS_DIR", "--compare-results-dir");
      optionalFlag(args, env, "COMPARISON", "--comparison");
      if (value(env, "DETAIL") !== "critical-path") {
        optionalFlag(args, env, "FIXTURE_THRESHOLD_MS", "--fixture-threshold-ms");
        optionalFlag(args, env, "FIXTURE_TOP", "--fixture-top");
      }
      for (const name of ["VIEW", "UNIT_ID", "OFFSET", "LIMIT"]) optionalFlag(args, env, name, `--${name.toLowerCase().replaceAll("_", "-")}`);
      return args;
    },
  },
  "harness-observability-check": {
    inputs: ["RUN_ID", "CHECK"],
    script: "./tools/harness/observability/observability-check-cli.mjs",
    resultDir: { mode: "required", flag: "--results-dir" },
    usage: "usage: make harness-observability-check RESULTS_DIR=<root|run-dir> [RUN_ID=<id>]",
    buildArgs(env) {
      const args = ["--check", value(env, "CHECK") || "all"];
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
  "scheduler-event-order-drift": {
    inputs: ["TARGET"],
    script: "./tools/harness/diagnostics/scheduler-event-order-drift-cli.mjs",
    resultDir: { mode: "currentRunDefault", positional: true },
    usage: "usage: make harness-observability-check RESULTS_DIR=<exact-run-dir> CHECK=events",
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
    usage: "usage: make harness-observability-check RESULTS_DIR=<exact-run-dir> CHECK=timing",
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
  const configured = Object.hasOwn(env, "CARTULARY_MAKE_INPUT_SOURCES")
    ? defaultTaskSurfaceManifest
    : env.TASK_SURFACE_MANIFEST || env.CARTULARY_TASK_SURFACE_MANIFEST || defaultTaskSurfaceManifest;
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

function makeNodeToolMakeEnvVars(name) {
  const tool = makeNodeTool(name);
  return uniqueNames([
    ...contractInputNames(name),
    ...makeNodeToolResultDirMakeEnvVars(name),
    ...(tool.runtimeEnv ?? []),
  ]);
}

function makeNodeToolKnownEnvVars() {
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
  const manifest = loadTaskSurfaceManifest(env);
  const publicEntry = manifest?.targets?.find((entry) => entry.name === name && entry.target_class === "public");
  if (Object.hasOwn(env, "CARTULARY_MAKE_INPUT_SOURCES") && publicEntry) {
    const resolved = resolveHarnessConfig(name, env, { manifest });
    env = { ...env };
    for (const input of declaredInputs) delete env[input];
    const booleans = new Set(publicEntry.input_contract.inputs.filter((input) => input.type === "exact_1_bool").map((input) => input.name));
    for (const [input, binding] of Object.entries(resolved.variables.target_inputs)) {
      env[input] = booleans.has(input) && binding.value === "false" ? "" : binding.value;
    }
  }

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
