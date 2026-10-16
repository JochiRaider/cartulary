import { canonicalJSONString, parseStrictJSON, readRetainedArtifactBytes, validateSchemaSync } from "../contract/index.mjs";

const same = (a, b) => canonicalJSONString(a) === canonicalJSONString(b);
const sorted = (values) => [...values].sort();
const read = (root, relative) => parseStrictJSON(readRetainedArtifactBytes(root, relative, 16 * 1024 ** 2).toString("utf8"));

// Selection fidelity is checked against a compiled owner contract, independently
// of the retained summary's internally consistent counts.
export function validateRetainedGraph(run, plan) {
  if (plan.graph.graph_digest !== run.manifest.graph_digest) throw new Error("retained graph differs from the current selection and profiles");
  if (!same(sorted(run.registrations.keys()), plan.graph.units.map((unit) => unit.unit_id))) throw new Error("retained registration roster differs from the current graph");
  for (const unit of plan.graph.units) {
    const registration = run.registrations.get(unit.unit_id);
    for (const key of ["kind", "needs", "resource_claims", "service_dependencies"]) {
      if (!same(registration[key], unit[key])) throw new Error(`retained registration ${key} differs from the current graph`);
    }
  }
}

export function validateRetainedCleanup(run, cleanup) {
  validateSchemaSync("cartulary.harness_cleanup_results.v1", cleanup);
  if (cleanup.run_id !== run.manifest.run_id || cleanup.results.some((result) => result.outcome !== "completed")) throw new Error("retained cleanup did not complete");
  if (cleanup.results.some((result, index) => result.sequence !== index + 1 || (result.unit_id !== null && !run.registrations.has(result.unit_id)))) throw new Error("retained cleanup roster is inconsistent");
  for (const id of run.started.keys()) {
    if (!cleanup.results.some((result) => result.operation === "host_release" && result.unit_id === id)) throw new Error("started unit has no completed host release");
  }
}

// Read every declared unit and row result in the selected projection. This is
// shared by full-check maintenance and owner audits, not a new execution catalog.
export function validateRetainedSelection(runRoot, run, units, projection, catalogRows) {
  const byRow = new Map(catalogRows.map((row) => [row.row_id, row]));
  const rowUnits = new Map();
  if (projection.status !== "pass" || projection.failure_class !== null || projection.failure_reason !== null) throw new Error("retained target projection did not pass");
  for (const id of projection.unit_ids) {
    const expected = units.get(id), terminal = run.terminal.get(id);
    if (!expected || terminal?.status !== "passed") throw new Error("selected unit has no successful terminal event");
    const reference = `unit-results/${id.replaceAll(/[^A-Za-z0-9_.-]+/gu, "-")}.json`;
    const result = read(runRoot, reference);
    validateSchemaSync("cartulary.harness_unit_result.v1", result);
    if (result.unit_id !== id || result.semantic_digest !== expected.semantic_digest || result.status !== "passed" ||
        (result.exit_code !== 0 && !(result.exit_code === null && terminal.cache_status === "hit")) ||
        result.signal !== null || result.missing_outputs.length || result.failure_class !== null || result.failure_reason !== null) throw new Error("selected unit result is unsuccessful or incompatible");
    if (expected.current_run_evidence_outputs.some((output) => !result.evidence_outputs.includes(output) || !projection.evidence_refs.includes(output))) throw new Error("selected unit omits declared evidence");
    const selectedRows = expected.current_run_evidence_outputs.filter((value) => value.startsWith("rows/") && value.endsWith(".json")).map((value) => value.slice(5, -5));
    if (!same(sorted(run.registrations.get(id).row_ids), sorted(selectedRows))) throw new Error("registered row selection differs from current unit selection");
    for (const rowID of selectedRows) {
      if (rowUnits.has(rowID)) throw new Error("selected row appears in multiple units");
      rowUnits.set(rowID, id);
      const row = byRow.get(rowID);
      if (!row) throw new Error("selected row is absent from the current catalog");
      const result = read(runRoot, `rows/${rowID}.json`);
      validateSchemaSync("cartulary.harness_row_result.v3", result);
      if (result.row_id !== rowID || result.runner !== row.runner || result.terminal_state !== "passed" ||
          result.exit_code !== 0 || result.failure_class !== null || result.failure_reason !== null) throw new Error("required row has no compatible successful terminal result");
    }
  }
  return rowUnits;
}
