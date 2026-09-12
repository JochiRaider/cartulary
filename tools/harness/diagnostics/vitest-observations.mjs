import { constants, closeSync, fstatSync, openSync, readFileSync } from "node:fs";
import path from "node:path";
import { validateSchemaSync } from "../contract/index.mjs";

export const failureDetailsSchemaID = "cartulary.vitest_failure_details.v2";
const maxReportBytes = 16 * 1024 * 1024;

export function diagnosticFailure(message, reason = "artifact_error") {
  return Object.assign(new Error(message), {
    failure_class: reason === "artifact_error" ? "artifact" : "harness",
    failure_reason: reason,
  });
}

export function readVitestJSON(file) {
  let fd;
  try {
    fd = openSync(file, constants.O_RDONLY | constants.O_NOFOLLOW);
    const stat = fstatSync(fd);
    if (!stat.isFile() || stat.size > maxReportBytes) throw new Error("not a bounded regular file");
    return JSON.parse(readFileSync(fd, "utf8"));
  } catch {
    throw diagnosticFailure("Required Vitest diagnostic input is missing, unsafe, oversized, or malformed");
  } finally {
    if (fd !== undefined) closeSync(fd);
  }
}

export function vitestObservationKey(observation) {
  return JSON.stringify([observation.project, observation.owner_path, observation.title]);
}

function classifyError(error, observation) {
  // This is the pinned runner's original message, not JSON's preferred stack.
  const timeout = error.message.match(/^(Test|Hook) timed out in ([1-9][0-9]*)ms\.\nIf this is a long-running (test|hook),/u);
  if (timeout && timeout[1].toLowerCase() === timeout[3]) {
    return {
      kind: timeout[1] === "Test" ? "test_timeout" : "hook_timeout",
      failure_class: "timing", failure_reason: "timeout_failure",
      timeout_ms: Number(timeout[2]),
    };
  }
  if (!error.message || /^(?:Error: )?STACK_TRACE_ERROR$/u.test(error.message.trim())) {
    return { kind: "unknown", failure_class: "unknown", failure_reason: "unknown_failure", timeout_ms: null };
  }
  if (observation.scope !== "test") {
    return { kind: "suite_setup", failure_class: "infra", failure_reason: "preflight_error", timeout_ms: null };
  }
  return { kind: "assertion", failure_class: "product", failure_reason: "test_assertion_failure", timeout_ms: null };
}

export function normalizeVitestObservations(capture, { invocationID, runID, runnerJSON, stdoutLog, stderrLog }) {
  try { validateSchemaSync(failureDetailsSchemaID, capture); }
  catch { throw diagnosticFailure("Vitest collector output does not satisfy the current diagnostic schema"); }
  if (capture.invocation_id !== invocationID || capture.run_id !== runID) {
    throw diagnosticFailure("Vitest collector output belongs to a different invocation");
  }
  if (capture.failures.length !== 0) throw diagnosticFailure("Vitest collector must not classify parent-owned failures");
  const seen = new Set();
  const failures = [];
  for (const observation of capture.observations) {
    const key = vitestObservationKey(observation);
    if (seen.has(key)) throw diagnosticFailure("Duplicate Vitest terminal observation", "scheduler_accounting_error");
    seen.add(key);
    if (observation.status === "passed" && observation.errors.length > 0) {
      throw diagnosticFailure("Passing Vitest observation contains errors", "scheduler_accounting_error");
    }
    if (observation.status !== "failed") continue;
    const errors = observation.errors.length ? observation.errors : [{ name: "Error", message: "", stack: "" }];
    for (const error of errors) {
      const classification = classifyError(error, observation);
      failures.push({
        project: observation.project, owner_path: observation.owner_path, title: observation.title,
        ...classification, duration_ms: observation.duration_ms,
        message: classification.kind === "unknown" ? "Vitest failed without preserving an actionable cause" : error.message,
        original_error: error,
        diagnostic_tags: error.stack.includes("STACK_TRACE_ERROR") ? ["vitest_stack_trace_error"] : [],
      });
    }
  }
  return {
    ...capture, runner_json: runnerJSON, stdout_log: stdoutLog, stderr_log: stderrLog,
    failures,
  };
}

export function selectedVitestObservations(details, file, titles) {
  return titles.map((title) => {
    const candidates = details.observations.filter((item) => item.scope === "test" && item.owner_path === file);
    const exact = candidates.filter((item) => item.title === title);
    const matches = exact.length ? exact : candidates.filter((item) => item.test_name === title);
    if (matches.length !== 1) throw diagnosticFailure("Vitest selected title has missing or ambiguous observations", "scheduler_accounting_error");
    return matches[0];
  });
}

export function reconcileVitestReport(details, report, root) {
  if (!report || !Array.isArray(report.testResults)) throw diagnosticFailure("Vitest runner JSON has no test results");
  const seen = new Set();
  for (const suite of report.testResults) {
    const file = path.relative(root, path.resolve(root, suite.name)).replaceAll("\\", "/");
    for (const assertion of suite.assertionResults ?? []) {
      const title = assertion.fullName ?? [...(assertion.ancestorTitles ?? []), assertion.title].join(" ");
      const [observation] = selectedVitestObservations(details, file, [title]);
      const key = vitestObservationKey(observation);
      if (seen.has(key) || observation.status !== assertion.status) {
        throw diagnosticFailure("Vitest runner and collector terminal observations contradict", "scheduler_accounting_error");
      }
      seen.add(key);
    }
  }
  if (details.observations.some((item) => item.scope === "test" && !seen.has(vitestObservationKey(item)))) {
    throw diagnosticFailure("Vitest collector has an unaccounted test observation", "scheduler_accounting_error");
  }
}
