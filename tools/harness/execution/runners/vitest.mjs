import path from "node:path";
import { resolveVitestTitles } from "../../test-catalog/index.mjs";
import { primaryPublicFailure, publicExitCodeForFailure, validateSchemaSync } from "../../contract/index.mjs";
import {
  diagnosticFailure, failureDetailsSchemaID, readVitestJSON,
  reconcileVitestReport, selectedVitestObservations, vitestObservationKey,
} from "../../diagnostics/vitest-failure-details.mjs";

function regexEscape(value) { return value.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&"); }

export function buildVitestInvocations(root, rows, command, runRoot) {
  if (!rows.length || new Set(rows.map((row) => `${row.selector.project_id}:${row.selector.file}`)).size !== 1 ||
    rows.length > 1 && rows.some((row) => row.process_isolation === "exclusive" || row.fixture_capability === "managed_process")) throw new Error("Vitest unit must be one compatible file selection");
  const exactTitles = resolveVitestTitles(root, rows);
  const fileRows = [...rows].sort((a, b) => a.row_id < b.row_id ? -1 : a.row_id > b.row_id ? 1 : 0);
  const file = fileRows[0].selector.file;
  const directory = path.join(runRoot, "unit-logs", `row-${fileRows[0].row_id}`);
  const runnerJSON = path.join(directory, "runner.json");
  const detailsFile = path.join(directory, "vitest-failure-details.json");
  const titles = fileRows.flatMap((row) => exactTitles.get(row.row_id)).sort();
  if (new Set(titles).size !== titles.length) throw new Error("Vitest rows select overlapping test identities");
  return [{
    root, file, project: fileRows[0].selector.project_id, runnerJSON, detailsFile,
    command: process.execPath,
    args: [path.join(root, "tools/harness/execution/vitest-invocation-cli.mjs"), runnerJSON, detailsFile, "--",
      command, "--dir", "apps/web", "exec", "vitest", "run", path.resolve(root, file),
      "--project", fileRows[0].selector.project_id, "-t", `^(?:${titles.map(regexEscape).join("|")})$`, "--maxWorkers=1"],
    rows: fileRows.map((row) => ({ row_id: row.row_id, selectors: exactTitles.get(row.row_id) })),
  }];
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
    const selectedTitles = new Set(invocation.rows.flatMap((row) => row.selectors));
    if (result.details.observations.some((item) => item.scope === "test" && ["passed", "failed"].includes(item.status) &&
      (item.owner_path !== invocation.file || item.project !== invocation.project || !selectedTitles.has(item.title)))) {
      throw diagnosticFailure("Vitest executed an unselected test", "scheduler_accounting_error");
    }
    return invocation.rows.map((row) => {
      try {
      const suiteFailures = result.details.failures.filter((failure) => result.details.observations.some((item) =>
        item.scope !== "test" && (!item.owner_path || item.owner_path === invocation.file) && vitestObservationKey(item) === vitestObservationKey(failure)));
      if (suiteFailures.length && !result.details.observations.some((item) => item.scope === "test" && item.owner_path === invocation.file)) {
        return rowResult(row.row_id, [], primaryPublicFailure(suiteFailures));
      }
      const observations = selectedVitestObservations(result.details, invocation.file, row.selectors, invocation.project);
      const selected = new Set(observations.map(vitestObservationKey));
      const nonTest = new Set(result.details.observations.filter((item) => item.scope !== "test").map(vitestObservationKey));
      const failures = result.details.failures.filter((failure) => selected.has(vitestObservationKey(failure)) || nonTest.has(vitestObservationKey(failure)));
      if (observations.some((item) => ["pending", "skipped"].includes(item.status)) && !failures.length) {
        failures.push({ failure_class: "harness", failure_reason: "scheduler_accounting_error" });
      }
      if (observations.some((item) => item.status === "failed") && !failures.length) throw diagnosticFailure("Failed Vitest row has no diagnostic cause", "scheduler_accounting_error");
      return rowResult(row.row_id, observations, primaryPublicFailure(failures));
      } catch (error) {
        return rowResult(row.row_id, [], error.failure_reason ? error : diagnosticFailure("Invalid selected Vitest observations"));
      }
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
