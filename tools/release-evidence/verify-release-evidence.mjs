import path from "node:path";
import { parseStrictJSON, readRetainedArtifactBytes, validateSchemaSync } from "../harness/contract/index.mjs";
import { validateCanonicalRun } from "../harness/observability/canonical-evidence.mjs";
import { loadRetainedPlan } from "../harness/observability/retained-plan.mjs";
import { validateRetainedCleanup, validateRetainedSelection } from "../harness/observability/retained-selection.mjs";
import { verifyPackageIdentity } from "./package-identity.mjs";

const read = (root, name) => parseStrictJSON(readRetainedArtifactBytes(root, name, 16 * 1024 ** 2).toString("utf8"));

// A release is composed from one complete, passing root of each command
// (R1): correctness in `ci`, release-only work in `release-check` and the
// quiet campaign in `release-measure`.
export const releaseRoots = Object.freeze({ ci: "ci", release_check: "release-check", release_measure: "release-measure" });

// Every consumer of the release-check package, by the root that runs it.
export const packageConsumers = Object.freeze({
  release_check: ["standup-package-smoke", "standup-reference-pack-smoke", "standup-operational-recovery-smoke"],
  release_measure: ["credential-capacity-assessment"],
});

/**
 * Admit the authored obligations against current owners: every obligation
 * names existing rows of its owners, and every owner has a row. Obligations
 * are the capability map; the roots prove them.
 */
export function releaseObligations(policy, catalog) {
  validateSchemaSync("cartulary.release_qualification.v2", policy);
  const rows = new Map(catalog.rows.map(row => [row.row_id, row]));
  const seen = new Set(), rowIDs = new Set();
  for (const obligation of policy.obligations) {
    if (seen.has(obligation.obligation_id)) throw new Error(`duplicate release obligation: ${obligation.obligation_id}`);
    seen.add(obligation.obligation_id);
    const owners = new Set();
    for (const id of obligation.row_ids) {
      const row = rows.get(id);
      if (!row) throw new Error(`unknown release row: ${id}`);
      if (!obligation.owner_ids.includes(row.owner_id)) throw new Error(`release row owner mismatch: ${id}`);
      owners.add(row.owner_id); rowIDs.add(id);
    }
    if (obligation.owner_ids.some(owner => !owners.has(owner))) throw new Error(`release owner has no selected evidence: ${obligation.obligation_id}`);
  }
  return { obligation_ids: [...seen].sort(), row_ids: [...rowIDs].sort() };
}

// A complete, passing invocation of exactly its command for the current
// source: plan, cleanup, secret scan and every selected unit and row.
async function admitRoot(key, root, { sourceDigest, catalog, commands, loadRun, loadPlan, readArtifact }) {
  const target = releaseRoots[key];
  const run = await loadRun(root, target);
  if (run.manifest.target !== target || run.manifest.command_id !== commands.get(target)) throw new Error(`${key} root is not a ${target} invocation`);
  if (run.manifest.source_digest !== sourceDigest) throw new Error(`${key} root source is incompatible`);
  if (run.summary.status !== "pass" || run.summary.failure_class !== null || run.summary.failure_reason !== null) throw new Error(`${key} root did not pass`);
  const cleanup = readArtifact(root, "cleanup-results.json");
  validateRetainedCleanup(run, cleanup);
  const { units } = loadPlan(root, run, cleanup);
  // This receipt schema admits only a passing scan; absence cannot pass.
  validateSchemaSync("cartulary.harness_retained_secret_scan.v1", readArtifact(root, "retained-secret-scan.json"));
  const projection = run.targetSummaries.get(target);
  if (!projection) throw new Error(`${key} root has no ${target} projection`);
  const rows = validateRetainedSelection(root, run, units, projection, catalog.rows);
  return { run, units, rows, record: { target, run_id: run.manifest.run_id, run_root: root, graph_digest: run.manifest.graph_digest, row_count: rows.size } };
}

/** The release decision. Pure evidence consumption: it launches, builds and repairs nothing. */
export async function verifyReleaseEvidence({ roots, sourceDigest, catalog, surface, policy, repoRoot,
  loadRun = validateCanonicalRun, loadPlan = loadRetainedPlan, readArtifact = read }) {
  const { obligation_ids, row_ids } = releaseObligations(policy, catalog);
  const commands = new Map(surface.targets.map(entry => [entry.name, entry.command_id]));
  const admitted = {};
  for (const key of Object.keys(releaseRoots)) {
    admitted[key] = await admitRoot(key, roots[key], { sourceDigest, catalog, commands, loadRun, loadPlan, readArtifact });
  }
  const toolchains = new Set(Object.values(admitted).map(({ run }) => run.manifest.toolchain_digest));
  if (toolchains.size !== 1) throw new Error("release roots disagree on toolchain");
  // release-measure consumed the package of exactly this release-check root.
  const named = admitted.release_measure.run.manifest.declared_inputs.RELEASE_RESULTS_DIR;
  if (!named || path.resolve(repoRoot, named) !== path.resolve(roots.release_check)) {
    throw new Error("release-measure did not consume the release-check root's package");
  }
  const releaseCheck = roots.release_check;
  const producer = readArtifact(releaseCheck, "package-release/artifacts/package.json");
  const inspection = readArtifact(releaseCheck, "package-release/artifacts/inspection.json");
  validateSchemaSync("cartulary.local_release_inspection.v1", inspection);
  if (!inspection.installed_without_build_toolchain || !inspection.occupied_destination_rejected) throw new Error("archive inspection incomplete");
  const { installed_without_build_toolchain: _installed, occupied_destination_rejected: _occupied, ...inspected } = inspection;
  const consumers = [["package-inspect", { ...inspected, schema_id: "cartulary.local_release_receipt.v1" }]];
  for (const [key, names] of Object.entries(packageConsumers)) {
    for (const name of names) consumers.push([name, readArtifact(roots[key], `${name}/artifacts/package.json`)]);
  }
  const identity = verifyPackageIdentity(producer, consumers);
  const passed = new Set(Object.values(admitted).flatMap(({ rows }) => [...rows.keys()]));
  const missing = row_ids.filter(id => !passed.has(id));
  if (missing.length) throw new Error(`release obligations lack passing rows: ${missing.join(",")}`);
  return {
    schema_id: "cartulary.release_evidence_closure.v2", status: "pass", source_digest: sourceDigest,
    toolchain_digest: [...toolchains][0], ...identity,
    roots: Object.fromEntries(Object.entries(admitted).map(([key, { record }]) => [key, record])),
    obligation_ids,
  };
}

/** The three roots a roots file names, relative to the file. */
export function resolveReleaseRoots(manifest, file) {
  validateSchemaSync("cartulary.release_evidence_roots.v2", manifest);
  return Object.fromEntries(Object.keys(releaseRoots).map(key => [key, path.resolve(path.dirname(file), manifest[key])]));
}
