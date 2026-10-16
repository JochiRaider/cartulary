import path from "node:path";
import { canonicalJSONString, parseStrictJSON, readRetainedArtifactBytes, validateSchemaSync } from "../contract/index.mjs";
import { targetForCatalogRow } from "../test-catalog/index.mjs";
import { validateCanonicalRun } from "./canonical-evidence.mjs";

const same = (a, b) => canonicalJSONString(a) === canonicalJSONString(b);
const sorted = (values) => [...values].sort();
const usage = (message) => Object.assign(new Error(message), { exit_code: 2, failure_class: "config", failure_reason: "usage_error" });
const artifact = (message) => Object.assign(new Error(message), { exit_code: 11, failure_class: "artifact", failure_reason: "artifact_error" });
const read = (root, relative) => parseStrictJSON(readRetainedArtifactBytes(root, relative, 16 * 1024 ** 2).toString("utf8"));

export function ownerEvidencePartitions(catalog, surface, owner) {
  const rows = catalog.rows.filter((row) => row.owner_id === owner);
  if (rows.length === 0) throw usage(`unknown active owner ${owner}`);
  const commands = new Map(surface.targets.filter((entry) => entry.command_id).map((entry) => [entry.command_id, entry.name]));
  const partitions = new Map();
  for (const row of rows) {
    const target = targetForCatalogRow(row, { commandTargetByID: commands });
    const selected = partitions.get(target) ?? [];
    selected.push(row.row_id); partitions.set(target, selected);
  }
  return { rows, partitions: new Map([...partitions].sort(([a], [b]) => a < b ? -1 : a > b ? 1 : 0).map(([target, ids]) => [target, sorted(ids)])) };
}

// The loader and compiler are explicit ports so adversarial retained evidence
// can be tested without starting work or manufacturing a current checkout.
export async function auditOwnerEvidence({ manifest, manifestFile, owner, catalog, surface, sourceDigest, compileRun, loadRun = validateCanonicalRun }) {
  validateSchemaSync("cartulary.harness_evidence_root_manifest.v1", manifest);
  if (manifest.owner_id !== owner) throw usage("evidence manifest owner does not match OWNER");
  const { rows, partitions } = ownerEvidencePartitions(catalog, surface, owner);
  const byRow = new Map(rows.map((row) => [row.row_id, row]));
  const publicTargets = new Map(surface.targets.filter((entry) => entry.target_class === "public").map((entry) => [entry.name, entry]));
  const supplied = manifest.entries.map((entry) => entry.target);
  if (new Set(supplied).size !== supplied.length || !same(supplied, sorted(supplied))) throw usage("evidence targets must be ASCII-sorted and unique");
  if (supplied.some((target) => !publicTargets.has(target))) throw usage("unknown public evidence target");
  const atomic = supplied.includes("test-slice");
  const required = atomic ? new Map([["test-slice", sorted(byRow.keys())]]) : partitions;
  const missing = [...required.keys()].filter((target) => !supplied.includes(target));
  if (missing.length) throw usage(`missing required evidence targets: ${missing.join(",")}`);
  const report = {
    schema_id: "cartulary.harness_owner_evidence_audit.v1", owner_id: owner,
    status: "pass", selection: atomic ? "complete_owner" : "target_partitions",
    compatibility: { source_digest: sourceDigest, test_catalog_digest: catalog.test_catalog_digest, verification_routing_digest: catalog.verification.routing_digest },
    required_partitions: [...required].map(([target, row_ids]) => ({ target, row_ids })),
    accepted: [], rejected: [], unused_inputs: manifest.entries.filter((entry) => !required.has(entry.target)),
  };
  const loaded = new Map();
  for (const entry of manifest.entries.filter((value) => required.has(value.target))) {
    const runRoot = path.resolve(path.dirname(manifestFile), entry.run_root);
    try {
      let context = loaded.get(runRoot);
      if (!context) {
        const run = await loadRun(runRoot);
        if (run.manifest.source_digest !== sourceDigest) throw artifact("retained source differs from the audited executable snapshot");
        if (run.summary.status !== "pass" || run.summary.failure_class !== null || run.summary.failure_reason !== null) throw artifact("retained invocation did not pass");
        const plan = compileRun(run.manifest);
        if (plan.graph.graph_digest !== run.manifest.graph_digest) throw artifact("retained graph differs from the current selection and profiles");
        if (!same(sorted(run.registrations.keys()), plan.graph.units.map((unit) => unit.unit_id))) throw artifact("retained registration roster differs from the current graph");
        for (const unit of plan.graph.units) {
          const registration = run.registrations.get(unit.unit_id);
          for (const key of ["kind", "needs", "resource_claims", "service_dependencies"]) {
            if (!same(registration[key], unit[key])) throw artifact(`retained registration ${key} differs from the current graph`);
          }
        }
        const cleanup = read(runRoot, "cleanup-results.json");
        validateSchemaSync("cartulary.harness_cleanup_results.v1", cleanup);
        if (cleanup.run_id !== run.manifest.run_id || cleanup.results.some((result) => result.outcome !== "completed")) throw artifact("retained cleanup did not complete");
        if (cleanup.results.some((result, index) => result.sequence !== index + 1 || (result.unit_id !== null && !run.registrations.has(result.unit_id)))) throw artifact("retained cleanup roster is inconsistent");
        for (const id of run.started.keys()) {
          if (!cleanup.results.some((result) => result.operation === "host_release" && result.unit_id === id)) throw artifact("started unit has no completed host release");
        }
        const units = new Map(plan.graph.units.map((unit) => [unit.unit_id, unit]));
        context = { run, plan, units }; loaded.set(runRoot, context);
      }
      const { run, plan, units } = context;
      const projection = run.targetSummaries.get(entry.target);
      const expectedUnits = plan.projections[entry.target];
      if (!projection || !expectedUnits || projection.command_id !== publicTargets.get(entry.target).command_id ||
          projection.status !== "pass" || !same(projection.unit_ids, sorted(expectedUnits))) throw artifact("retained target projection does not close its execution contract");
      if (atomic && (run.manifest.target !== "test-slice" || run.manifest.declared_inputs.OWNER !== owner ||
          run.manifest.declared_inputs.ROWS || run.manifest.declared_inputs.SERVICE_BACKED_ONLY)) throw artifact("full-owner evidence requires an unfiltered owner slice");
      const rowUnits = new Map();
      for (const id of projection.unit_ids) {
        const expected = units.get(id), terminal = run.terminal.get(id);
        if (!expected || terminal?.status !== "passed") throw artifact("selected unit has no successful terminal event");
        const reference = `unit-results/${id.replaceAll(/[^A-Za-z0-9_.-]+/gu, "-")}.json`;
        const result = read(runRoot, reference);
        validateSchemaSync("cartulary.harness_unit_result.v1", result);
        if (result.unit_id !== id || result.semantic_digest !== expected.semantic_digest || result.status !== "passed" ||
            (result.exit_code !== 0 && !(result.exit_code === null && terminal.cache_status === "hit")) ||
            result.signal !== null || result.missing_outputs.length || result.failure_class !== null || result.failure_reason !== null) throw artifact("selected unit result is unsuccessful or incompatible");
        if (expected.current_run_evidence_outputs.some((reference) => !result.evidence_outputs.includes(reference) || !projection.evidence_refs.includes(reference))) throw artifact("selected unit omits declared evidence");
        const selectedRows = expected.current_run_evidence_outputs.filter((value) => value.startsWith("rows/") && value.endsWith(".json")).map((value) => value.slice(5, -5));
        if (!same(sorted(run.registrations.get(id).row_ids), sorted(selectedRows))) throw artifact("registered row selection differs from current unit selection");
        for (const rowID of selectedRows) {
          if (rowUnits.has(rowID)) throw artifact("selected row appears in multiple units");
          rowUnits.set(rowID, id);
        }
      }
      const expectedRows = required.get(entry.target);
      const observedRows = sorted([...rowUnits.keys()].filter((id) => byRow.has(id)));
      if (!same(observedRows, expectedRows)) throw artifact("target owner-row partition is incomplete or contains extra rows");
      for (const rowID of expectedRows) {
        const reference = `rows/${rowID}.json`;
        if (!projection.evidence_refs.includes(reference)) throw artifact("target omits a required row artifact");
        const result = read(runRoot, reference);
        validateSchemaSync("cartulary.harness_row_result.v3", result);
        if (result.row_id !== rowID || result.runner !== byRow.get(rowID).runner || result.terminal_state !== "passed") throw artifact("required row has no compatible successful terminal result");
      }
      report.accepted.push({ target: entry.target, run_root: entry.run_root, run_id: run.manifest.run_id, graph_digest: run.manifest.graph_digest, row_ids: expectedRows });
    } catch (error) {
      report.status = "fail";
      report.rejected.push({ target: entry.target, run_root: entry.run_root, reason: error.message.slice(0, 2000) });
      throw Object.assign(artifact(`${entry.target}: ${error.message}`), { auditReport: report });
    }
  }
  validateSchemaSync(report.schema_id, report);
  return report;
}
