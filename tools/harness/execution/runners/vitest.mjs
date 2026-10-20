import path from "node:path";
import { existsSync } from "node:fs";
import { resolveVitestTitles } from "../../test-catalog/index.mjs";
import { primaryPublicFailure, publicExitCodeForFailure, secureWriteFile, validateSchemaSync } from "../../contract/index.mjs";
import {
  diagnosticFailure, failureDetailsSchemaID, readVitestJSON,
  reconcileVitestReport, selectedVitestObservations, vitestObservationKey,
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

export function validateVitestRequest(request) {
  validateSchemaSync("cartulary.vitest_invocation.v1", request);
  const files = new Set(), rows = new Set();
  for (const file of request.files) {
    if (files.has(file.file)) throw diagnosticFailure("Duplicate Vitest file selection", "scheduler_accounting_error");
    files.add(file.file);
    const titles = new Set();
    for (const row of file.rows) {
      if (rows.has(row.row_id)) throw diagnosticFailure("Duplicate Vitest row selection", "scheduler_accounting_error");
      rows.add(row.row_id);
      for (const title of row.selectors) {
        if (titles.has(title)) throw diagnosticFailure("Vitest rows select overlapping test identities", "scheduler_accounting_error");
        titles.add(title);
      }
    }
  }
}

export function vitestInvocation(root, request) {
  return { root, project: request.project, rows: request.files.flatMap((file) => file.rows.map((row) => ({ ...row, file: file.file, project: request.project }))) };
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
    reconcileVitestReport(result.details, result.report, invocation.root);
    const completed = new Set(result.details.completed_files.map((file) => JSON.stringify([file.project, file.owner_path])));
    const selected = new Set(invocation.rows.flatMap((row) => row.selectors.map((title) => vitestObservationKey({ project: row.project, owner_path: row.file, title }))));
    if (result.details.observations.some((item) => item.scope === "test" && ["passed", "failed"].includes(item.status) && !selected.has(vitestObservationKey(item)))) {
      throw diagnosticFailure("Vitest executed an unselected test", "scheduler_accounting_error");
    }
    const runKeys = new Set(result.details.observations.filter((item) => item.scope === "run").map(vitestObservationKey));
    const runFailures = result.details.failures.filter((failure) => runKeys.has(vitestObservationKey(failure)));
    return invocation.rows.map((row) => {
      try {
        if (runFailures.length) return rowResult(row.row_id, [], primaryPublicFailure(runFailures));
        if (!completed.has(JSON.stringify([row.project, row.file]))) {
          if (interrupted || !result.details.final_complete) return cancelled(row);
          throw diagnosticFailure("Vitest selected file has no completed observation", "scheduler_accounting_error");
        }
        const fileObservations = result.details.observations.filter((item) => item.project === row.project && item.owner_path === row.file);
        const suiteKeys = new Set(fileObservations.filter((item) => item.scope === "suite").map(vitestObservationKey));
        const suiteFailures = result.details.failures.filter((failure) => suiteKeys.has(vitestObservationKey(failure)));
        if (suiteFailures.length) return rowResult(row.row_id, fileObservations.filter((item) => item.scope === "test" && row.selectors.includes(item.title)), primaryPublicFailure(suiteFailures));
        const observations = selectedVitestObservations(result.details, row.file, row.selectors, row.project);
        const keys = new Set(observations.map(vitestObservationKey));
        const failures = result.details.failures.filter((failure) => keys.has(vitestObservationKey(failure)));
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
    const details = existsSync(invocation.detailsFile) ? readVitestJSON(invocation.detailsFile) : undefined;
    let report;
    if (details) {
      const reportPath = path.resolve(invocation.root, details.runner_json);
      if (path.dirname(reportPath) !== path.dirname(invocation.runnerJSON) || !/^runner(?:-checkpoint-[0-9]+)?\.json$/u.test(path.basename(reportPath))) throw diagnosticFailure("Vitest native report reference escapes the invocation");
      report = readVitestJSON(reportPath);
    }
    return adaptVitestInvocation(invocation, { ...result, details, report });
  } catch (error) { return invocation.rows.map((row) => rowResult(row.row_id, [], error)); }
}
