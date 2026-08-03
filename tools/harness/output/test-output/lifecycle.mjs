import { targetStartStats } from "./target-start-stats.mjs";
import { repoRoot } from "../../contract/index.mjs";
import { normalizeOutputMode, verboseOutput } from "../tool-output.mjs";

function parseLifecycleOptions(args) {
  const options = { positional: [], force: false };
  for (let index = 0; index < args.length; index += 1) {
    const arg = args[index];
    if (arg === "--force") {
      options.force = true;
      continue;
    }
    if (arg.startsWith("--")) {
      const name = arg.slice(2).replaceAll("-", "_");
      const value = args[index + 1];
      if (value === undefined) {
        throw new Error(`${arg} requires a value`);
      }
      options[name] = value;
      index += 1;
      continue;
    }
    options.positional.push(arg);
  }
  return options;
}

function shouldEmitLifecycle(options) {
  const mode = normalizeOutputMode();
  return options.force || verboseOutput() || mode === "ci";
}

function parseNonNegativeInteger(value, label) {
  const parsed = Number.parseInt(value ?? "", 10);
  if (!Number.isFinite(parsed) || parsed < 0) {
    throw new Error(`${label} must be a non-negative integer`);
  }
  return parsed;
}

function parseTargetListValue(value) {
  if (!value) {
    return [];
  }
  return value
    .split(",")
    .map((entry) => entry.trim())
    .filter((entry) => entry.length > 0);
}

export function handleRunStart(args) {
  const options = parseLifecycleOptions(args);
  const [label] = options.positional;
  if (
    !label ||
    options.steps === undefined ||
    options.summary_targets === undefined ||
    options.jobs === undefined
  ) {
    throw new Error(
      "usage: test-output.mjs run-start <label> --steps <n> --summary-targets <n> [--helper-units <n>] --jobs <n> [--force]",
    );
  }
  const workUnits = parseNonNegativeInteger(options.steps, "steps");
  const summaryTargets = parseNonNegativeInteger(
    options.summary_targets,
    "summary-targets",
  );
  const helperUnits = parseNonNegativeInteger(
    options.helper_units ?? "0",
    "helper-units",
  );
  const jobs = parseNonNegativeInteger(options.jobs, "jobs");
  if (!shouldEmitLifecycle(options)) {
    return 0;
  }
  process.stdout.write(
    `[RUN] ${label} work_units=${workUnits} summary_targets=${summaryTargets} helper_units=${helperUnits} jobs=${jobs}\n`,
  );
  return 0;
}

export function handleStepStart(args) {
  const options = parseLifecycleOptions(args);
  const [label, indexText, totalText, target] = options.positional;
  const mode = options.mode ?? options.positional[4] ?? "serial";
  const jobsText = options.jobs ?? options.positional[5] ?? "1";
  if (!label || !indexText || !totalText || !target) {
    throw new Error(
      "usage: test-output.mjs step-start <label> <index> <total> <target> [--mode <mode>] [--jobs <n>] [--force]",
    );
  }
  const index = parseNonNegativeInteger(indexText, "index");
  const total = parseNonNegativeInteger(totalText, "total");
  const jobs = parseNonNegativeInteger(jobsText, "jobs");
  if (!shouldEmitLifecycle(options)) {
    return 0;
  }
  process.stdout.write(
    `[STEP] ${label} ${index}/${total} ${target} mode=${mode} jobs=${jobs}\n`,
  );
  return 0;
}

export function handleStepEligible(args) {
  const options = parseLifecycleOptions(args);
  const [label, indexText, target] = options.positional;
  if (!label || !indexText || !target) {
    throw new Error("usage: test-output.mjs step-eligible <label> <index> <target> [--needs <a,b>] [--resource-class <class>]");
  }
  parseNonNegativeInteger(indexText, "index");
  return 0;
}

export function handleStepFinish(args) {
  const options = parseLifecycleOptions(args);
  const [label, indexText, target, status, exitCodeText] = options.positional;
  if (!label || !indexText || !target || !status || exitCodeText === undefined) {
    throw new Error("usage: test-output.mjs step-finish <label> <index> <target> <pass|fail|interrupted> <exit-code>");
  }
  const index = parseNonNegativeInteger(indexText, "index");
  const exitCode = parseNonNegativeInteger(exitCodeText, "exit-code");
  if (shouldEmitLifecycle(options)) {
    process.stdout.write(`[STEP] ${label} ${index} ${target} status=${status} exit_code=${exitCode}\n`);
  }
  return 0;
}

export function handleStepCancelled(args) {
  const options = parseLifecycleOptions(args);
  const [label, indexText, target] = options.positional;
  if (!label || !indexText || !target) {
    throw new Error("usage: test-output.mjs step-cancelled <label> <index> <target>");
  }
  parseNonNegativeInteger(indexText, "index");
  return 0;
}

export function handleRunFinish(args) {
  const options = parseLifecycleOptions(args);
  const [label, status, exitCodeText] = options.positional;
  if (!label || !status || exitCodeText === undefined) {
    throw new Error("usage: test-output.mjs run-finish <label> <pass|fail|interrupted> <exit-code>");
  }
  const exitCode = parseNonNegativeInteger(exitCodeText, "exit-code");
  if (shouldEmitLifecycle(options)) {
    process.stdout.write(`[RUN] ${label} status=${status} exit_code=${exitCode}\n`);
  }
  return 0;
}

export function handleTargetStart(args) {
  const options = parseLifecycleOptions(args);
  const [target] = options.positional;
  if (!target) {
    throw new Error(
      "usage: test-output.mjs target-start <target> [--children <a,b>] [--service-backed <0|1>] [--expected-steps <n>] [--expected-tests <n>] [--force]",
    );
  }
  if (!shouldEmitLifecycle(options)) {
    return 0;
  }
  const children = parseTargetListValue(options.children);
  const stats = targetStartStats(repoRoot, target, children);
  const serviceBacked =
    options.service_backed === undefined
      ? stats.serviceBacked
      : options.service_backed === "1" || options.service_backed === "true";
  const expectedSteps =
    options.expected_steps === undefined
      ? stats.expectedSteps
      : parseNonNegativeInteger(options.expected_steps, "expected-steps");
  const expectedTests =
    options.expected_tests === undefined
      ? stats.expectedTests
      : parseNonNegativeInteger(options.expected_tests, "expected-tests");
  const childField =
    children.length > 0 ? ` children=${children.join(",")}` : "";
  process.stdout.write(
    `[TARGET] start ${target} service_backed=${serviceBacked ? 1 : 0} expected_steps=${expectedSteps} expected_tests=${expectedTests}${childField}\n`,
  );
  return 0;
}
