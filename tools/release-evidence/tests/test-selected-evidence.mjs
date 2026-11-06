import assert from "node:assert/strict";
import test from "node:test";
import { mkdtempSync, readFileSync, rmSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { buildCompleteEvidence, buildEvidenceUnit, loadFixtureIdentity, persistEvidence, writeEvidenceFile } from "../../harness/observability/tests/retained-fixture.mjs";
import { verifySelectedEvidence, packageConsumerNames, qualificationRequirements, resolveEvidenceRoots } from "../verify-selected-evidence.mjs";
import { verifyPackageIdentity } from "../package-identity.mjs";
import { WorkGraphCompiler } from "../../harness/scheduler/work-graph/index.mjs";
import { loadTestCatalog } from "../../harness/test-catalog/index.mjs";

test("release obligations resolve current owners and reject duplicate, unknown or ownerless claims", () => {
  const root = path.resolve(import.meta.dirname, "../../..");
  const policy = JSON.parse(readFileSync(path.join(root, "contracts/verification/release-qualification.v1.json"), "utf8"));
  const catalog = loadTestCatalog(root), compiler = new WorkGraphCompiler(root);
  const compile = target => compiler.compileTarget(target);
  const required = qualificationRequirements(policy, catalog, compile);
  assert.ok(required.row_ids.includes("module.database_migrations.integration.production_preflight_state_matrix"));
  assert.ok(required.operation_ids.includes("harness.package_release"));
  for (const [change, error] of [
    [value => value.obligations.push(structuredClone(value.obligations[0])), /duplicate release obligation/],
    [value => value.obligations[0].row_ids.push("module.absent.unit.missing"), /unknown release row/],
    [value => { value.obligations[0].owner_ids = ["module.absent"]; }, /owner mismatch/],
    [value => value.obligations[0].owner_ids.push("module.absent"), /owner has no selected evidence/],
    [value => value.required_operations.push("harness.absent"), /unknown/i],
  ]) {
    const changed = structuredClone(policy);
    change(changed);
    assert.throws(() => qualificationRequirements(changed, catalog, compile), error);
  }
});

const receipt = {
  schema_id: "cartulary.local_release_receipt.v1", manifest_sha256: "a".repeat(64), archive_sha256: "b".repeat(64), archive: `${"a".repeat(64)}.tar`,
  images: ["application", "postgres", "seaweedfs"].map(role => ({role, id: `sha256:${"c".repeat(64)}`})),
  operator_sha256: "d".repeat(64), inventory_disposition: "content_complete_review_unestablished",
};

function fixture(t) {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-selected-release-"));
  t.after(() => rmSync(root, {recursive: true, force: true}));
  const identity = loadFixtureIdentity("package-inspect");
  const operations = ["package-release", "package-inspect", ...packageConsumerNames];
  const runs = operations.map(name => {
    const directory = path.join(root, name);
    const unit = buildEvidenceUnit(`operation:harness.${name.replaceAll("-", "_")}`);
    unit.current_run_evidence_outputs = [`unit-results/${unit.unit_id.replace(":", "-")}.json`];
    const projections = {"package-inspect": {command_id: identity.commandID, unit_ids: [unit.unit_id]}};
    if (packageConsumerNames.includes(name)) projections[name] = {command_id: identity.commandID, unit_ids: [unit.unit_id]};
    const evidence = buildCompleteEvidence({runID: `independent-${name}`, target: "package-inspect", identity, units: [unit], projections});
    const original = `unit-results/${unit.unit_id}.json`;
    evidence.files[unit.current_run_evidence_outputs[0]] = evidence.files[original];
    delete evidence.files[original];
    evidence.summary.artifact_refs = Object.keys(evidence.files).sort();
    persistEvidence(directory, evidence);
    if (name === "package-release") writeEvidenceFile(directory, "package-release/artifacts/package.json", receipt);
    else if (name === "package-inspect") writeEvidenceFile(directory, "package-release/artifacts/inspection.json", {...receipt, schema_id: "cartulary.local_release_inspection.v1", installed_without_build_toolchain: true, occupied_destination_rejected: true});
    else writeEvidenceFile(directory, `${name}/artifacts/package.json`, receipt);
    return {directory, name, unit, evidence, projections};
  });
  const options = {roots: runs.map(run => run.directory), sourceDigest: runs[0].evidence.manifest.source_digest, catalog: {rows: []}, required: {row_ids: [], operation_ids: runs.map(run => run.unit.unit_id.slice("operation:".length))},
    compileRun: manifest => {
      const run = runs.find(run => run.evidence.manifest.run_id === manifest.run_id);
      return {graph: run.evidence.context.graph, projections: Object.fromEntries(Object.entries(run.projections).map(([name, value]) => [name, value.unit_ids]))};
    }};
  const save = (name, value) => {
    const owner = name.split("/")[0];
    const run = runs.find(run => run.name === owner) ?? runs[0];
    writeEvidenceFile(run.directory, name, value);
  };
  return {root, options, evidence: runs[0].evidence, save};
}

test("selected release evidence accepts one complete package and rejects missing, stale, failed or mixed evidence", async t => {
  const f = fixture(t);
  const report = await verifySelectedEvidence(f.options);
  assert.equal(report.archive_sha256, receipt.archive_sha256);
  await assert.rejects(verifySelectedEvidence({...f.options, sourceDigest: `sha256:${"f".repeat(64)}`}), /source is incompatible/);
  await assert.rejects(verifySelectedEvidence({...f.options, required: {row_ids: ["module.absent.unit.required"], operation_ids: []}}), /mandatory evidence missing/);
  f.save(`${packageConsumerNames[0]}/artifacts/package.json`, {...receipt, archive_sha256: "f".repeat(64)});
  await assert.rejects(verifySelectedEvidence(f.options), /package identity mismatch/);
  f.save(`${packageConsumerNames[0]}/artifacts/package.json`, receipt);
  const cleanup = structuredClone(f.evidence.files["cleanup-results.json"]);
  cleanup.results[0].outcome = "failed";
  f.save("cleanup-results.json", cleanup);
  await assert.rejects(verifySelectedEvidence(f.options));
  f.save("cleanup-results.json", f.evidence.files["cleanup-results.json"]);
  f.save("retained-secret-scan.json", {...f.evidence.files["retained-secret-scan.json"], status: "fail"});
  await assert.rejects(verifySelectedEvidence(f.options));
  f.save("retained-secret-scan.json", f.evidence.files["retained-secret-scan.json"]);
  rmSync(path.join(f.root, packageConsumerNames[1], `${packageConsumerNames[1]}/artifacts/package.json`));
  await assert.rejects(verifySelectedEvidence(f.options));
});

test("release closure requires explicit roots and its Make selection launches no test or service work", () => {
  assert.throws(() => resolveEvidenceRoots({schema_id: "cartulary.release_evidence_roots.v1", run_roots: [], newest: true}, "/tmp/roots.json"));
  assert.throws(() => verifyPackageIdentity(receipt, [["different image", {...receipt, images: receipt.images.map(image => ({...image, id: `sha256:${"f".repeat(64)}`}))}]]), /image identity mismatch/);
  const root = path.resolve(import.meta.dirname, "../../..");
  const compiler = new WorkGraphCompiler(root, {targetInputs: {"release-evidence-verify": {EVIDENCE_ROOTS_FILE: ".cartulary/selected-evidence.json"}}});
  const graph = compiler.compileTarget("release-evidence-verify");
  assert.deepEqual(graph.units.map(unit => unit.unit_id), ["operation:harness.release_evidence_verify"]);
  assert.deepEqual(graph.units[0].service_dependencies, []);
  assert.deepEqual(graph.units[0].row_bindings, []);
  const qualification = compiler.compileCommand("package-inspect", {CHECK: "all"}).graph;
  assert.equal(qualification.units.filter(unit => unit.unit_id === "operation:harness.package_release").length, 1);
  assert.ok(qualification.units.some(unit => unit.unit_id === "operation:harness.credential_capacity_assessment"));
});
