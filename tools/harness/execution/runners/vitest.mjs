import path from "node:path";
import { primaryPublicFailure, publicExitCodeForFailure, validateSchemaSync } from "../../contract/index.mjs";
import {
  diagnosticFailure, failureDetailsSchemaID, readVitestJSON,
  reconcileVitestReport, selectedVitestObservations, vitestObservationKey,
} from "../../diagnostics/vitest-failure-details.mjs";

export const runnerContract = Object.freeze({ runner: "vitest", selector_kind: "vitest_exact_titles" });

function regexEscape(value) { return value.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&"); }

export function buildVitestInvocations(root, rows, workers, command, runRoot) {
  const byFile = new Map();
  for (const row of rows) {
    const values = byFile.get(row.selector.file) ?? [];
    values.push(row);
    byFile.set(row.selector.file, values);
  }
  return [...byFile.entries()].sort(([a], [b]) => a < b ? -1 : a > b ? 1 : 0).map(([file, fileRows]) => {
    const directory = path.join(runRoot, "unit-logs", `row-${fileRows[0].row_id}`);
    const runnerJSON = path.join(directory, "runner.json");
    const detailsFile = path.join(directory, "vitest-failure-details.json");
    const titles = [...new Set(fileRows.flatMap((row) => row.selector.titles))].sort();
    return {
      root, file, runnerJSON, detailsFile,
      command: process.execPath,
      args: [path.join(root, "tools/harness/execution/vitest-invocation-cli.mjs"), runnerJSON, detailsFile, "--",
        command, "--dir", "apps/web", "exec", "vitest", "run", path.resolve(root, file),
        "-t", `(?:^| )(?:${titles.map(regexEscape).join("|")})$`, `--maxWorkers=${workers}`],
      rows: fileRows.map((row) => ({ row_id: row.row_id, selectors: [...row.selector.titles] })),
    };
  });
}

function rowResult(rowID, observations, failure = null) {
  return {
    row_id: rowID,
    terminal_state: failure ? failure.failure_class === "product" ? "failed" : "infrastructure_failed" : "passed",
    duration_ms: Math.round(observations.reduce((sum, item) => sum + item.duration_ms, 0)),
    exit_code: failure ? publicExitCodeForFailure(failure) : 0,
    failure_class: failure?.failure_class ?? null,
    failure_reason: failure?.failure_reason ?? null,
    failure_diagnostic: null,
  };
}

export function adaptVitestInvocation(invocation, result) {
  try {
    if (result.commandFailure) throw Object.assign(new Error("Vitest invocation failed"), result.commandFailure);
    if (result.signal || [130, 143].includes(result.status)) {
      return invocation.rows.map((row) => ({
        ...rowResult(row.row_id, [], { failure_class: "interrupted", failure_reason: "cancelled_or_interrupted" }),
        terminal_state: "cancelled", exit_code: result.signal === "SIGTERM" || result.status === 143 ? 143 : 130,
      }));
    }
    validateSchemaSync(failureDetailsSchemaID, result.details);
    reconcileVitestReport(result.details, result.report, invocation.root);
    return invocation.rows.map((row) => {
      const suiteFailures = result.details.failures.filter((failure) => result.details.observations.some((item) =>
        item.scope !== "test" && (!item.owner_path || item.owner_path === invocation.file) && vitestObservationKey(item) === vitestObservationKey(failure)));
      if (suiteFailures.length && !result.details.observations.some((item) => item.scope === "test" && item.owner_path === invocation.file)) {
        return rowResult(row.row_id, [], primaryPublicFailure(suiteFailures));
      }
      const observations = selectedVitestObservations(result.details, invocation.file, row.selectors);
      const selected = new Set(observations.map(vitestObservationKey));
      const nonTest = new Set(result.details.observations.filter((item) => item.scope !== "test").map(vitestObservationKey));
      const failures = result.details.failures.filter((failure) => selected.has(vitestObservationKey(failure)) || nonTest.has(vitestObservationKey(failure)));
      if (observations.some((item) => ["pending", "skipped"].includes(item.status)) && !failures.length) {
        failures.push({ failure_class: "harness", failure_reason: "scheduler_accounting_error" });
      }
      if (observations.some((item) => item.status === "failed") && !failures.length) throw diagnosticFailure("Failed Vitest row has no diagnostic cause", "scheduler_accounting_error");
      return rowResult(row.row_id, observations, primaryPublicFailure(failures));
    });
  } catch (error) {
    const failure = error.failure_reason ? error : diagnosticFailure("Invalid Vitest diagnostic evidence");
    return invocation.rows.map((row) => rowResult(row.row_id, [], failure));
  }
}

export function adaptVitestInvocationFile(invocation, result) {
  try {
    if (result.commandFailure || result.signal || [130, 143].includes(result.status)) return adaptVitestInvocation(invocation, result);
    return adaptVitestInvocation(invocation, { ...result, details: readVitestJSON(invocation.detailsFile), report: readVitestJSON(invocation.runnerJSON) });
  } catch (error) {
    return invocation.rows.map((row) => rowResult(row.row_id, [], error));
  }
}
