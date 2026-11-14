import path from "node:path";
import { parseStrictJSON, readRetainedArtifactBytes, validateSchemaSync } from "../harness/contract/index.mjs";
import { validateCanonicalRun } from "../harness/observability/canonical-evidence.mjs";
import { validateExecutionContext } from "../harness/observability/canonical-context.mjs";
import { validateRetainedCleanup, validateRetainedGraph, validateRetainedSelection } from "../harness/observability/retained-selection.mjs";
import { verifyPackageIdentity } from "./package-identity.mjs";

const read = (root, name) => parseStrictJSON(readRetainedArtifactBytes(root, name, 16 * 1024 ** 2).toString("utf8"));
export const packageConsumerNames = Object.freeze(["standup-package-smoke", "standup-reference-pack-smoke", "standup-operational-recovery-smoke", "credential-capacity-assessment"]);

/** Admit authored obligations against current executable owners, without running them. */
export function qualificationRequirements(policy, catalog, compileTarget) {
  validateSchemaSync("cartulary.release_qualification.v1", policy);
  const rows = new Map(catalog.rows.map(row => [row.row_id, row]));
  const obligations = new Set(), rowIDs = new Set(), operationIDs = new Set();
  for (const obligation of policy.obligations) {
    if (obligations.has(obligation.obligation_id)) throw new Error(`duplicate release obligation: ${obligation.obligation_id}`);
    obligations.add(obligation.obligation_id);
    const owners = new Set();
    for (const id of obligation.row_ids) {
      const row = rows.get(id);
      if (!row) throw new Error(`unknown release row: ${id}`);
      if (!obligation.owner_ids.includes(row.owner_id)) throw new Error(`release row owner mismatch: ${id}`);
      owners.add(row.owner_id); rowIDs.add(id);
    }
    if (obligation.owner_ids.some(owner => !owners.has(owner))) throw new Error(`release owner has no selected evidence: ${obligation.obligation_id}`);
  }
  for (const target of policy.required_operations) {
    const graph = compileTarget(target);
    if (!graph.units.length) throw new Error(`empty release operation: ${target}`);
    for (const unit of graph.units) {
      for (const binding of unit.row_bindings) {
        if (!rows.has(binding.row_id)) throw new Error(`unknown release operation row: ${binding.row_id}`);
        rowIDs.add(binding.row_id);
      }
      if (unit.unit_id.startsWith("operation:")) operationIDs.add(unit.unit_id.slice("operation:".length));
    }
  }
  return {row_ids: [...rowIDs].sort(), operation_ids: [...operationIDs].sort()};
}

/** Pure evidence consumption. No runner, command invocation or repair capability. */
export async function verifySelectedEvidence({ roots, sourceDigest, catalog, required, compileRun, loadRun = validateCanonicalRun, readArtifact = read }) {
  if (!Array.isArray(roots) || roots.length === 0 || roots.length > 512 || new Set(roots).size !== roots.length) throw new Error("explicit unique evidence roots are required");
  const observedRows = new Set(), observedOperations = new Set(), packages = [], consumers = new Map(), accepted = [];
  let archiveInspection = false;
  for (const root of roots) {
    const run = await loadRun(root);
    if (run.manifest.source_digest !== sourceDigest) throw new Error("retained source is incompatible");
    if (run.summary.status !== "pass" || run.summary.failure_class !== null || run.summary.failure_reason !== null) throw new Error("retained invocation did not pass");
    const plan = compileRun(run.manifest);
    validateRetainedGraph(run, plan);
    const cleanup = readArtifact(root, "cleanup-results.json");
    validateRetainedCleanup(run, cleanup);
    validateExecutionContext(readArtifact(root, "execution-context.json"), run, cleanup);
    const scan = readArtifact(root, "retained-secret-scan.json");
    // This receipt schema admits only a passing scan. Failure or absence cannot
    // supply a receipt; schema admission is the acceptance boundary.
    validateSchemaSync("cartulary.harness_retained_secret_scan.v1", scan);
    const units = new Map(plan.graph.units.map(unit => [unit.unit_id, unit]));
    const projection = run.targetSummaries.get(run.manifest.target);
    if (!projection) throw new Error("missing public target projection");
    const rows = validateRetainedSelection(root, run, units, projection, catalog.rows);
    for (const row of rows.keys()) observedRows.add(row);
    for (const id of projection.unit_ids) {
      if (id.startsWith("operation:")) observedOperations.add(id.slice("operation:".length));
    }
    if (units.has("operation:harness.package_release")) packages.push(readArtifact(root, "package-release/artifacts/package.json"));
    if (units.has("operation:harness.package_inspect")) {
      const inspection = readArtifact(root, "package-release/artifacts/inspection.json");
      validateSchemaSync("cartulary.local_release_inspection.v1", inspection);
      if (!inspection.installed_without_build_toolchain || !inspection.occupied_destination_rejected) throw new Error("archive inspection incomplete");
      const {installed_without_build_toolchain: _installed, occupied_destination_rejected: _occupied, ...receipt} = inspection;
      packages.push({...receipt, schema_id: "cartulary.local_release_receipt.v1"});
      archiveInspection = true;
    }
    for (const name of packageConsumerNames) {
      const operation = `harness.${name.replaceAll("-", "_")}`;
      const report = [...run.targetSummaries.values()].find(value => value.target === name || value.target === operation);
      if (!report) continue;
      const receipt = readArtifact(root, `${name}/artifacts/package.json`);
      packages.push(receipt); consumers.set(name, receipt);
    }
    accepted.push({run_id: run.manifest.run_id, run_root: root, graph_digest: run.manifest.graph_digest, row_ids: [...rows.keys()].sort()});
  }
  const missingRows = required.row_ids.filter(id => !observedRows.has(id));
  const missingOperations = required.operation_ids.filter(id => !observedOperations.has(id));
  if (missingRows.length || missingOperations.length) throw new Error(`mandatory evidence missing: ${[...missingRows, ...missingOperations].join(",")}`);
  if (!archiveInspection || packages.length === 0 || packageConsumerNames.some(name => !consumers.has(name))) throw new Error("mandatory package producer, archive inspection or consumers missing");
  const identity = verifyPackageIdentity(packages[0], packages.slice(1).map((receipt, index) => [`package receipt ${index + 1}`, receipt]));
  return {schema_id: "cartulary.release_evidence_closure.v1", status: "pass", source_digest: sourceDigest, ...identity, accepted};
}

export function resolveEvidenceRoots(manifest, file) {
  validateSchemaSync("cartulary.release_evidence_roots.v1", manifest);
  return manifest.run_roots.map(value => path.resolve(path.dirname(file), value));
}
