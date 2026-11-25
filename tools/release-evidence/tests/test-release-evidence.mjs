import assert from "node:assert/strict";
import test from "node:test";
import { mkdtempSync, readFileSync, rmSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { buildCompleteEvidence, buildEvidenceUnit, loadFixtureIdentity, persistEvidence, writeEvidenceFile } from "../../harness/observability/tests/retained-fixture.mjs";
import { packageConsumers, releaseObligations, releaseRoots, resolveReleaseRoots, verifyReleaseEvidence } from "../verify-release-evidence.mjs";
import { verifyPackageIdentity } from "../package-identity.mjs";
import { composeRelease } from "../release-cli.mjs";
import { loadExecutionTopology } from "../../harness/generated-artifacts/execution-topology.mjs";
import { loadTestCatalog } from "../../harness/test-catalog/index.mjs";

const repoRoot = path.resolve(import.meta.dirname, "../../..");

test("release obligations resolve current owners and reject duplicate, unknown or ownerless claims", () => {
  const policy = JSON.parse(readFileSync(path.join(repoRoot, "contracts/verification/release-qualification.v2.json"), "utf8"));
  const catalog = loadTestCatalog(repoRoot);
  const { obligation_ids, row_ids } = releaseObligations(policy, catalog);
  assert.equal(obligation_ids.length, policy.obligations.length);
  assert.ok(row_ids.includes("module.database_migrations.integration.production_preflight_state_matrix"));
  for (const [change, error] of [
    [value => value.obligations.push(structuredClone(value.obligations[0])), /duplicate release obligation/],
    [value => value.obligations[0].row_ids.push("module.absent.unit.missing"), /unknown release row/],
    [value => { value.obligations[0].owner_ids = ["module.absent"]; }, /owner mismatch/],
    [value => value.obligations[0].owner_ids.push("module.absent"), /owner has no selected evidence/],
    [value => { value.required_operations = []; }, /release_qualification\.v2/],
  ]) {
    const changed = structuredClone(policy);
    change(changed);
    assert.throws(() => releaseObligations(changed, catalog), error);
  }
});

const receipt = {
  schema_id: "cartulary.local_release_receipt.v1", manifest_sha256: "a".repeat(64), archive_sha256: "b".repeat(64), archive: `${"a".repeat(64)}.tar`,
  images: ["application", "postgres", "seaweedfs"].map(role => ({role, id: `sha256:${"c".repeat(64)}`})),
  operator_sha256: "d".repeat(64), inventory_disposition: "content_complete_review_unestablished",
};

// Three complete, passing roots, one per release command, each proving one
// obligation row; release-measure names the release-check root.
function fixture(t, { failed = null } = {}) {
  const parent = mkdtempSync(path.join(os.tmpdir(), "cartulary-release-roots-"));
  t.after(() => rmSync(parent, {recursive: true, force: true}));
  const roots = Object.fromEntries(Object.keys(releaseRoots).map(key => [key, path.join(parent, key)]));
  const catalogRows = [];
  const evidence = {};
  for (const [key, target] of Object.entries(releaseRoots)) {
    const row = { row_id: `harness.release.unit.${key}_fixture`, runner: "go" };
    catalogRows.push({ ...row, owner_id: "harness.release", status: "active" });
    const unit = buildEvidenceUnit(`${key}-fixture`, [row]);
    evidence[key] = buildCompleteEvidence({ runID: `${key}-run`, target, identity: loadFixtureIdentity(target), rows: [row], units: [unit],
      declaredInputs: key === "release_measure" ? { RELEASE_RESULTS_DIR: roots.release_check } : {},
      outcomes: failed === key ? { [unit.unit_id]: "failed" } : {} });
    persistEvidence(roots[key], evidence[key]);
  }
  writeEvidenceFile(roots.release_check, "package-release/artifacts/package.json", receipt);
  writeEvidenceFile(roots.release_check, "package-release/artifacts/inspection.json", {...receipt, schema_id: "cartulary.local_release_inspection.v1", installed_without_build_toolchain: true, occupied_destination_rejected: true});
  for (const [key, names] of Object.entries(packageConsumers)) for (const name of names) writeEvidenceFile(roots[key], `${name}/artifacts/package.json`, receipt);
  const policy = { schema_id: "cartulary.release_qualification.v2", obligations: [{
    obligation_id: "release.fixture", normative_owners: ["TestingHarness"], capability: "Fixture", expectation: "Every root proves its row",
    owner_ids: ["harness.release"], row_ids: catalogRows.map(row => row.row_id) }] };
  const options = { roots, sourceDigest: evidence.ci.manifest.source_digest, catalog: { rows: catalogRows },
    surface: loadExecutionTopology({ root: repoRoot }).taskSurface, policy, repoRoot };
  return { roots, options, evidence };
}

test("one complete passing root per release command composes a release", async t => {
  const f = fixture(t);
  const closure = await verifyReleaseEvidence(f.options);
  assert.equal(closure.schema_id, "cartulary.release_evidence_closure.v2");
  assert.equal(closure.archive_sha256, receipt.archive_sha256);
  assert.deepEqual(Object.fromEntries(Object.entries(closure.roots).map(([key, value]) => [key, [value.target, value.row_count]])),
    { ci: ["ci", 1], release_check: ["release-check", 1], release_measure: ["release-measure", 1] });
  assert.deepEqual(closure.obligation_ids, ["release.fixture"]);
});

test("the release decision rejects foreign, incomplete, failed or mismatched roots", async t => {
  const f = fixture(t);
  await assert.rejects(verifyReleaseEvidence({ ...f.options, sourceDigest: `sha256:${"f".repeat(64)}` }), /source is incompatible/u);
  await assert.rejects(verifyReleaseEvidence({ ...f.options, roots: { ...f.roots, ci: f.roots.release_check } }), /release-check does not match ci|is not a ci invocation/u);
  await assert.rejects(verifyReleaseEvidence({ ...f.options, policy: { ...f.options.policy, obligations: [{ ...f.options.policy.obligations[0],
    row_ids: [...f.options.policy.obligations[0].row_ids, "harness.release.unit.unrun_fixture"] }] },
    catalog: { rows: [...f.options.catalog.rows, { row_id: "harness.release.unit.unrun_fixture", owner_id: "harness.release", runner: "go", status: "active" }] } }),
  /lack passing rows: harness\.release\.unit\.unrun_fixture/u);
  // release-measure must have consumed this release-check root's package.
  const elsewhere = { ...f.options, repoRoot: path.join(f.roots.ci, "elsewhere") };
  const measure = structuredClone(f.evidence.release_measure);
  measure.manifest.declared_inputs.RELEASE_RESULTS_DIR = "relative/release-check";
  measure.plan.manifest = structuredClone(measure.manifest);
  persistEvidence(f.roots.release_measure, measure);
  await assert.rejects(verifyReleaseEvidence(elsewhere), /did not consume the release-check root's package/u);
  persistEvidence(f.roots.release_measure, f.evidence.release_measure);
  // One package identity across release-check and release-measure.
  writeEvidenceFile(f.roots.release_measure, "credential-capacity-assessment/artifacts/package.json", { ...receipt, archive_sha256: "f".repeat(64) });
  await assert.rejects(verifyReleaseEvidence(f.options), /credential-capacity-assessment: package identity mismatch/u);
  writeEvidenceFile(f.roots.release_measure, "credential-capacity-assessment/artifacts/package.json", receipt);
  writeEvidenceFile(f.roots.release_check, "retained-secret-scan.json", { ...f.evidence.release_check.files["retained-secret-scan.json"], status: "fail" });
  await assert.rejects(verifyReleaseEvidence(f.options), /cartulary\.harness_retained_secret_scan\.v1 validation failed/u);
  writeEvidenceFile(f.roots.release_check, "retained-secret-scan.json", f.evidence.release_check.files["retained-secret-scan.json"]);
  const cleanup = structuredClone(f.evidence.ci.files["cleanup-results.json"]);
  cleanup.results[0].outcome = "blocked";
  writeEvidenceFile(f.roots.ci, "cleanup-results.json", cleanup);
  await assert.rejects(verifyReleaseEvidence(f.options), /^Error: retained cleanup did not complete$/u);
  writeEvidenceFile(f.roots.ci, "cleanup-results.json", f.evidence.ci.files["cleanup-results.json"]);
  // A release ci executes every row: a root that reused cached results is refused (D12).
  const cached = structuredClone(f.evidence.ci);
  cached.manifest.cache_mode = "normal";
  cached.plan.manifest = structuredClone(cached.manifest);
  persistEvidence(f.roots.ci, cached);
  await assert.rejects(verifyReleaseEvidence(f.options), /ci root reused cached results/u);
  persistEvidence(f.roots.ci, f.evidence.ci);
  assert.equal((await verifyReleaseEvidence(f.options)).status, "pass");
});

test("a failed release command fails the release", async t => {
  for (const key of Object.keys(releaseRoots)) {
    const f = fixture(t, { failed: key });
    await assert.rejects(verifyReleaseEvidence(f.options), new RegExp(`^Error: ${key} root did not pass$`, "u"));
  }
});

test("a roots file names exactly the three release roots and identical package images", () => {
  assert.deepEqual(resolveReleaseRoots({ schema_id: "cartulary.release_evidence_roots.v2", ci: "a", release_check: "b", release_measure: "c" }, "/tmp/r/roots.json"),
    { ci: "/tmp/r/a", release_check: "/tmp/r/b", release_measure: "/tmp/r/c" });
  for (const manifest of [{ schema_id: "cartulary.release_evidence_roots.v2", ci: "a", release_check: "b" },
    { schema_id: "cartulary.release_evidence_roots.v1", run_roots: ["a"] },
    { schema_id: "cartulary.release_evidence_roots.v2", ci: "a", release_check: "b", release_measure: "c", newest: true }]) {
    assert.throws(() => resolveReleaseRoots(manifest, "/tmp/roots.json"));
  }
  assert.throws(() => verifyPackageIdentity(receipt, [["different image", {...receipt, images: receipt.images.map(image => ({...image, id: `sha256:${"f".repeat(64)}`}))}]]), /image identity mismatch/);
});

test("make release runs each missing stage once, in order, and verifies their roots", t => {
  const parent = mkdtempSync(path.join(os.tmpdir(), "cartulary-release-compose-"));
  t.after(() => rmSync(parent, {recursive: true, force: true}));
  const calls = [];
  const runMake = (target, { runID, inputs }) => { calls.push({ target, runID, inputs }); return 0; };
  const outcome = composeRelease({ supplied: { ci: path.join(parent, "earlier-ci") }, runID: "rel", resultsDir: parent, releaseRoot: path.join(parent, "rel"), runMake });
  assert.equal(outcome.status, 0);
  assert.deepEqual(calls.map(call => [call.target, call.runID]), [
    ["release-check", "rel.release-check"], ["release-measure", "rel.release-measure"], ["release-evidence-verify", "rel.release-evidence-verify"]]);
  // release-measure consumes this release's release-check package.
  assert.deepEqual(calls[1].inputs, { RELEASE_RESULTS_DIR: path.join(parent, "rel.release-check") });
  assert.deepEqual(calls[2].inputs, { EVIDENCE_ROOTS_FILE: outcome.rootsFile });
  assert.deepEqual(resolveReleaseRoots(JSON.parse(readFileSync(outcome.rootsFile, "utf8")), outcome.rootsFile), {
    ci: path.join(parent, "earlier-ci"), release_check: path.join(parent, "rel.release-check"), release_measure: path.join(parent, "rel.release-measure") });
  const stopped = [];
  const failed = composeRelease({ supplied: {}, runID: "again", resultsDir: parent, releaseRoot: path.join(parent, "again"),
    runMake: (target) => { stopped.push(target); return target === "release-check" ? 2 : 0; } });
  assert.deepEqual([failed.status, failed.failed, stopped], [2, "release-check", ["ci", "release-check"]]);
  // The release's own ci runs with the cache off (D12).
  const inputs = [];
  composeRelease({ supplied: {}, runID: "fresh", resultsDir: parent, releaseRoot: path.join(parent, "fresh"),
    runMake: (target, call) => { inputs.push([target, call.inputs]); return 1; } });
  assert.deepEqual(inputs, [["ci", { CARTULARY_HARNESS_CACHE_MODE: "off" }]]);
});

