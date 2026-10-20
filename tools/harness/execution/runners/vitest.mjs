import path from "node:path";
import { resolveVitestTitles } from "../../test-catalog/index.mjs";
import { primaryPublicFailure, publicExitCodeForFailure, secureWriteFile, validateSchemaSync } from "../../contract/index.mjs";
import {
  diagnosticFailure, failureDetailsSchemaID, readVitestJSON,
  readVitestEvidence, reconcileVitestReport, selectedVitestObservations, validateVitestRequest, vitestFileKey, vitestObservationKey,
} from "../../diagnostics/vitest-failure-details.mjs";

export function buildVitestInvocations(root, rows, runRoot) {
  if (!rows.length || new Set(rows.map((row) => row.selector.project_id)).size !== 1 ||
    rows.length > 1 && rows.some((row) => row.process_isolation === "exclusive" || row.fixture_capability === "managed_process")) throw new Error("Vitest unit must be one compatible project cohort");
  const exactTitles = resolveVitestTitles(root, rows);
  const sortedRows = [...rows].sort((a, b) => a.row_id < b.row_id ? -1 : a.row_id > b.row_id ? 1 : 0);
  const files = new Map();
  for (const row of sortedRows) {
    const file = files.get(row.selector.file) ?? { file: row.selector.file, rows: [] };
    file.rows.push({ row_id: row.row_id, selectors: exactTitles.get(row.row_id) });
    files.set(file.file, file);
  }
  const request = { schema_id: "cartulary.vitest_invocation.v1", project: rows[0].selector.project_id,
    files: [...files.values()].sort((a, b) => a.file < b.file ? -1 : a.file > b.file ? 1 : 0), publish_rows: true };
  validateVitestRequest(request);
  const directory = path.join(runRoot, "unit-logs", `row-${sortedRows[0].row_id}`);
  const runnerJSON = path.join(directory, "runner.json");
  const detailsFile = path.join(directory, "vitest-failure-details.json");
  const requestFile = path.join(directory, "invocation.json");
  secureWriteFile(requestFile, `${JSON.stringify(request)}\n`, { allowedRoot: runRoot });
  return [{ ...vitestInvocation(root, request), requestFile, runnerJSON, detailsFile, command: process.execPath,
    args: [path.join(root, "tools/harness/execution/vitest-invocation-cli.mjs"), runnerJSON, detailsFile, requestFile] }];
}

export { validateVitestRequest };

export function vitestInvocation(root, request) {
  return { root, request, project: request.project, rows: request.files.flatMap((file) => file.rows.map((row) => ({ ...row, file: file.file, project: request.project }))) };
}

function rowResult(rowID, observations, failure = null) {
  return {
    row_id: rowID,
    terminal_state: failure ? failure.failure_class === "interrupted" ? "cancelled" : failure.failure_class === "product" ? "failed" : "infrastructure_failed" : "passed",
    duration_ms: Math.round(observations.reduce((sum, item) => sum + item.duration_ms, 0)),
    exit_code: failure ? publicExitCodeForFailure(failure) : 0,
    failure_class: failure?.failure_class ?? null,
    failure_reason: failure?.failure_reason ?? null,
    failure_diagnostic: null,
  };
}

export function adaptVitestInvocation(invocation, result) {
  const interrupted = Boolean(result.signal || [130, 143].includes(result.status) || result.details?.run_status === "interrupted");
  const cancelled = (row) => ({ ...rowResult(row.row_id, [], { failure_class: "interrupted", failure_reason: "cancelled_or_interrupted" }),
    exit_code: result.signal === "SIGTERM" || result.status === 143 ? 143 : 130 });
  try {
    if (result.commandFailure && result.commandFailure.failure_class !== "interrupted") throw Object.assign(new Error("Vitest invocation failed"), result.commandFailure);
    if (!result.details && interrupted) return invocation.rows.map(cancelled);
    validateSchemaSync(failureDetailsSchemaID, result.details);
    const index = reconcileVitestReport(result.details, result.report, invocation.root);
    const completed = new Set(result.details.completed_files.map((file) => JSON.stringify([file.project, file.owner_path])));
    const selected = new Set(invocation.rows.flatMap((row) => row.selectors.map((title) => vitestObservationKey({ project: row.project, owner_path: row.file, title }))));
    if (result.details.observations.some((item) => item.scope === "test" && ["passed", "failed"].includes(item.status) && !selected.has(vitestObservationKey(item)))) {
      throw diagnosticFailure("Vitest executed an unselected test", "scheduler_accounting_error");
    }
    const runFailures = [];
    const suiteFailures = new Map();
    for (const observation of index.observations.values()) {
      const failures = index.failures.get(vitestObservationKey(observation)) ?? [];
      if (observation.scope === "run") runFailures.push(...failures);
      if (observation.scope === "suite") {
        const key = vitestFileKey(observation.project, observation.owner_path);
        const existing = suiteFailures.get(key) ?? [];
        existing.push(...failures); suiteFailures.set(key, existing);
      }
    }
    const runFailure = primaryPublicFailure(runFailures);
    const fileFailure = new Map([...suiteFailures].map(([key, failures]) => [key, primaryPublicFailure(failures)]));
    return invocation.rows.map((row) => {
      try {
        if (runFailure) return rowResult(row.row_id, [], runFailure);
        if (!completed.has(JSON.stringify([row.project, row.file]))) {
          if (interrupted || !result.details.final_complete) return cancelled(row);
          throw diagnosticFailure("Vitest selected file has no completed observation", "scheduler_accounting_error");
        }
        const failure = fileFailure.get(vitestFileKey(row.project, row.file));
        if (failure) return rowResult(row.row_id, row.selectors.map((title) => index.observations.get(vitestObservationKey({ project: row.project, owner_path: row.file, title }))).filter(Boolean), failure);
        const observations = selectedVitestObservations(result.details, row.file, row.selectors, row.project, index);
        const failures = observations.flatMap((observation) => index.failures.get(vitestObservationKey(observation)) ?? []);
        if (observations.some((item) => ["pending", "skipped"].includes(item.status)) && !failures.length) {
          failures.push({ failure_class: "harness", failure_reason: "scheduler_accounting_error" });
        }
        if (observations.some((item) => item.status === "failed") && !failures.length) throw diagnosticFailure("Failed Vitest row has no diagnostic cause", "scheduler_accounting_error");
        return rowResult(row.row_id, observations, primaryPublicFailure(failures));
      } catch (error) { return rowResult(row.row_id, [], error.failure_reason ? error : diagnosticFailure("Invalid selected Vitest observations")); }
    });
  } catch (error) {
    const failure = error.failure_reason ? error : diagnosticFailure("Invalid Vitest diagnostic evidence");
    return invocation.rows.map((row) => rowResult(row.row_id, [], failure));
  }
}

export function adaptVitestInvocationFile(invocation, result) {
  try {
    const request = invocation.requestFile ? readVitestJSON(invocation.requestFile) : invocation.request;
    const evidence = readVitestEvidence({ root: invocation.root, request, reportFile: invocation.runnerJSON,
      detailsFile: invocation.detailsFile, runID: process.env.CARTULARY_TEST_RUN_ID });
    const commandFailure = result.commandFailure ?? evidence?.commandFailure;
    if (result.status === 0 && evidence && !evidence.details.final_complete && !commandFailure) {
      throw diagnosticFailure("Successful Vitest invocation lacks final settlement", "scheduler_accounting_error");
    }
    return adaptVitestInvocation(invocation, { ...result, ...evidence, commandFailure });
  } catch (error) { return invocation.rows.map((row) => rowResult(row.row_id, [], error.failure_reason ? error : diagnosticFailure("Invalid Vitest diagnostic evidence"))); }
}

// Independently settled rows may survive owner loss after the final file. That
// does not make an interrupted or failed enclosing invocation successful.
export function vitestInvocationFailure(execution, rows) {
  if (rows.some((row) => row.terminal_state !== "passed")) return null;
  if (execution.commandFailure) return execution.commandFailure;
  if (execution.signal || [130, 143].includes(execution.status)) {
    return { failure_class: "interrupted", failure_reason: "cancelled_or_interrupted" };
  }
  if (execution.status !== 0) return diagnosticFailure("Vitest invocation failed after file settlement", "scheduler_accounting_error");
  return null;
}
