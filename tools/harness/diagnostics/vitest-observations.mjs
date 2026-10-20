import path from "node:path";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { readLocalFile } from "../runtime/secure-local-files.mjs";

export const failureDetailsSchemaID = "cartulary.vitest_failure_details.v3";
const maxReportBytes = 16 * 1024 * 1024;

export function diagnosticFailure(message, reason = "artifact_error") {
  return Object.assign(new Error(message), {
    failure_class: reason === "artifact_error" ? "artifact" : "harness",
    failure_reason: reason,
  });
}

export function readVitestJSON(file) {
  try {
    return parseStrictJSON(readLocalFile(file, { maximum: maxReportBytes, privateFile: false, allowReplacement: true }).toString("utf8"));
  } catch {
    throw diagnosticFailure("Required Vitest diagnostic input is missing, unsafe, oversized, or malformed");
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

export function selectedVitestObservations(details, file, titles, project = null) {
  return titles.map((title) => {
    const candidates = details.observations.filter((item) => item.scope === "test" && item.owner_path === file && (project === null || item.project === project));
    const matches = candidates.filter((item) => item.title === title);
    if (matches.length !== 1) throw diagnosticFailure("Vitest selected title has missing or ambiguous observations", "scheduler_accounting_error");
    return matches[0];
  });
}

export function reconcileVitestReport(details, envelope, root) {
  try { validateSchemaSync("cartulary.vitest_runner_reports.v1", envelope); }
  catch { throw diagnosticFailure("Vitest native reports do not satisfy the current schema"); }
  if (envelope.invocation_id !== details.invocation_id || envelope.run_id !== details.run_id) throw diagnosticFailure("Vitest native reports belong to a different invocation");
  const seen = new Set(), files = new Set();
  for (const entry of envelope.reports) {
    const fileKey = JSON.stringify([entry.project, entry.file]);
    if (files.has(fileKey) || !Array.isArray(entry.report.testResults)) throw diagnosticFailure("Duplicate or invalid Vitest native file report", "scheduler_accounting_error");
    files.add(fileKey);
    if (entry.report.testResults.length !== 1) throw diagnosticFailure("Vitest native report must describe exactly one selected file", "scheduler_accounting_error");
    for (const suite of entry.report.testResults) {
      const file = path.relative(root, path.resolve(root, suite.name)).replaceAll("\\", "/");
      if (file !== entry.file) throw diagnosticFailure("Vitest native report file identity mismatch", "scheduler_accounting_error");
      for (const assertion of suite.assertionResults ?? []) {
        const title = assertion.fullName ?? [...(assertion.ancestorTitles ?? []), assertion.title].join(" ");
        const [observation] = selectedVitestObservations(details, file, [title], entry.project);
        const key = vitestObservationKey(observation);
        if (seen.has(key) || observation.status !== assertion.status) throw diagnosticFailure("Vitest runner and collector terminal observations contradict", "scheduler_accounting_error");
        seen.add(key);
      }
    }
  }
  const completed = new Set(details.completed_files.map((file) => JSON.stringify([file.project, file.owner_path])));
  if (files.size !== completed.size || [...files].some((key) => !completed.has(key))) throw diagnosticFailure("Vitest completed files disagree with native reports", "scheduler_accounting_error");
  if (details.observations.some((item) => item.scope === "test" && !seen.has(vitestObservationKey(item)))) throw diagnosticFailure("Vitest collector has an unaccounted test observation", "scheduler_accounting_error");
}

// Presentation-only aggregation; retained evidence keeps each native JSON object.
export function vitestReportProjection(envelope) {
  validateSchemaSync("cartulary.vitest_runner_reports.v1", envelope);
  const result = { success: envelope.reports.length > 0, testResults: [] };
  for (const { report } of envelope.reports) {
    result.success &&= report.success === true;
    result.testResults.push(...report.testResults);
    for (const [key, value] of Object.entries(report)) if (key.startsWith("num") && Number.isInteger(value)) result[key] = (result[key] ?? 0) + value;
  }
  return result;
}
