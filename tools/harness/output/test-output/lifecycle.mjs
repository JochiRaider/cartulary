import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

import { targetStartStats } from "./target-start-stats.mjs";
import { repoRoot, secureWriteFile, validateSchemaSync } from "../../contract/index.mjs";
import { resolveResultsRoot, resolveRunId } from "../../contract/test-output-context.mjs";
import { normalizeOutputMode, verboseOutput } from "../tool-output.mjs";

const runId = resolveRunId();
const resultsRoot = resolveResultsRoot();

function sequenceToken(sequence) {
  const token = String(sequence ?? "")
    .trim()
    .replaceAll(/[^A-Za-z0-9_.-]+/gu, "-")
    .replaceAll(/^-+|-+$/gu, "")
    .slice(0, 128);
  if (!token) throw new Error("sequence label has no safe retained identity");
  return token;
}

function sequenceEventFile(sequence) {
  return path.join(resultsRoot, runId, sequenceToken(sequence), "sequence-events.jsonl");
}

function retainedSequenceEvents(sequence) {
  const file = sequenceEventFile(sequence);
  if (!existsSync(file)) return [];
  return readFileSync(file, "utf8")
    .split(/\r?\n/u)
    .filter(Boolean)
    .map((line) => JSON.parse(line));
}

function appendSequenceEvent(sequence, eventFields) {
  const file = sequenceEventFile(sequence);
  const events = retainedSequenceEvents(sequence);
  const now = new Date();
  const firstMs = events.length > 0 ? Date.parse(events[0].emitted_at) : now.getTime();
  const event = {
    schema_id: "cartulary.harness_sequence_event.v1",
    run_id: runId,
    sequence: sequenceToken(sequence),
    seq: events.length + 1,
    event: eventFields.event,
    emitted_at: now.toISOString(),
    monotonic_ms: Math.max(0, now.getTime() - firstMs),
    ...eventFields,
  };
  validateSchemaSync(event.schema_id, event);
  secureWriteFile(file, `${events.map((item) => JSON.stringify(item)).join("\n")}${events.length > 0 ? "\n" : ""}${JSON.stringify(event)}\n`, {
    allowedRoot: path.join(resultsRoot, runId),
  });
  return event;
}

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
  appendSequenceEvent(label, { event: "sequence_started", status: "running" });
  if (!shouldEmitLifecycle(options)) {
    return 0;
  }
  process.stdout.write(
    `[RUN] ${label} work_units=${workUnits} summary_targets=${summaryTargets} helper_units=${helperUnits} jobs=${jobs} run_id=${runId}\n`,
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
  const needs = parseTargetListValue(options.needs);
  const retainedMode = mode === "parallel" ? "scheduler" : mode;
  if (options.eligibility_retained !== "1") {
    appendSequenceEvent(label, {
      event: "step_eligible",
      step_index: index,
      target,
      mode: retainedMode,
      jobs,
      needs,
      resource_claims: options.resource_class ? { [options.resource_class]: 1 } : {},
      status: "pending",
    });
  }
  appendSequenceEvent(label, {
    event: "step_started",
    step_index: index,
    target,
    mode: retainedMode,
    jobs,
    needs,
    resource_claims: options.resource_class ? { [options.resource_class]: 1 } : {},
    status: "running",
  });
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
  appendSequenceEvent(label, {
    event: "step_eligible",
    step_index: parseNonNegativeInteger(indexText, "index"),
    target,
    mode: "scheduler",
    jobs: 1,
    needs: parseTargetListValue(options.needs),
    resource_claims: options.resource_class ? { [options.resource_class]: 1 } : {},
    status: "pending",
  });
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
  const retained = retainedSequenceEvents(label);
  const prior = retained
    .filter((event) => event.event === "step_started" && event.step_index === index)
    .at(-1);
  const eligible = retained
    .filter((event) => event.event === "step_eligible" && event.step_index === index)
    .at(-1);
  const now = new Date();
  const endedMonotonicMs = retained.length > 0
    ? Math.max(0, now.getTime() - Date.parse(retained[0].emitted_at))
    : 0;
  const event = appendSequenceEvent(label, {
    event: "step_finished",
    step_index: index,
    target,
    needs: [...(prior?.needs ?? eligible?.needs ?? [])],
    resource_claims: { ...(prior?.resource_claims ?? eligible?.resource_claims ?? {}) },
    mode: prior?.mode ?? eligible?.mode ?? "scheduler",
    jobs: prior?.jobs ?? eligible?.jobs ?? 1,
    status,
    exit_code: exitCode,
    eligible_at: eligible?.emitted_at ?? prior?.emitted_at ?? now.toISOString(),
    started_at: prior?.emitted_at ?? now.toISOString(),
    ended_at: now.toISOString(),
    eligible_monotonic_ms: eligible?.monotonic_ms ?? prior?.monotonic_ms ?? endedMonotonicMs,
    started_monotonic_ms: prior?.monotonic_ms ?? endedMonotonicMs,
    ended_monotonic_ms: endedMonotonicMs,
    duration_ms: prior ? Math.max(0, Date.now() - Date.parse(prior.emitted_at)) : 0,
  });
  if (shouldEmitLifecycle(options)) {
    process.stdout.write(`[STEP] ${label} ${index} ${target} status=${status} duration_ms=${event.duration_ms}\n`);
  }
  return 0;
}

export function handleStepCancelled(args) {
  const options = parseLifecycleOptions(args);
  const [label, indexText, target] = options.positional;
  if (!label || !indexText || !target) {
    throw new Error("usage: test-output.mjs step-cancelled <label> <index> <target>");
  }
  const retained = retainedSequenceEvents(label);
  const eligible = retained
    .filter((event) => event.event === "step_eligible" && event.step_index === parseNonNegativeInteger(indexText, "index"))
    .at(-1);
  const now = new Date();
  const endedMonotonicMs = retained.length > 0
    ? Math.max(0, now.getTime() - Date.parse(retained[0].emitted_at))
    : 0;
  appendSequenceEvent(label, {
    event: "step_skipped",
    step_index: parseNonNegativeInteger(indexText, "index"),
    target,
    needs: [...(eligible?.needs ?? [])],
    resource_claims: { ...(eligible?.resource_claims ?? {}) },
    mode: eligible?.mode ?? "scheduler",
    jobs: eligible?.jobs ?? 1,
    status: "interrupted",
    exit_code: 130,
    eligible_at: eligible?.emitted_at ?? now.toISOString(),
    ended_at: now.toISOString(),
    eligible_monotonic_ms: eligible?.monotonic_ms ?? endedMonotonicMs,
    ended_monotonic_ms: endedMonotonicMs,
    skip_reason: "interrupted",
    duration_ms: 0,
  });
  return 0;
}

export function handleRunFinish(args) {
  const options = parseLifecycleOptions(args);
  const [label, status, exitCodeText] = options.positional;
  if (!label || !status || exitCodeText === undefined) {
    throw new Error("usage: test-output.mjs run-finish <label> <pass|fail|interrupted> <exit-code>");
  }
  const exitCode = parseNonNegativeInteger(exitCodeText, "exit-code");
  const prior = retainedSequenceEvents(label).find((event) => event.event === "sequence_started");
  const event = appendSequenceEvent(label, {
    event: status === "interrupted" ? "sequence_interrupted" : "sequence_finished",
    status,
    exit_code: exitCode,
    duration_ms: prior ? Math.max(0, Date.now() - Date.parse(prior.emitted_at)) : 0,
  });
  if (shouldEmitLifecycle(options)) {
    process.stdout.write(`[RUN] ${label} status=${status} duration_ms=${event.duration_ms}\n`);
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
