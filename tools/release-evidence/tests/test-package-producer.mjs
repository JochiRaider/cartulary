import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import test from "node:test";
import { spawnSync } from "node:child_process";
import { createScratch } from "../../harness/test-support/scratch.mjs";
import { buildReaderEvidence, writeEvidenceFile } from "../../harness/observability/tests/retained-fixture.mjs";
import { workspaceLayout } from "../../workspace_layout.generated.mjs";
import { createPackageProducerFixture, imageIDs, borrowedContainer, borrowedTags, packageWorkerScript } from "./package-producer-fixture.mjs";
const repoRoot = path.resolve(import.meta.dirname, "../../..");

test("actual package production settles partial effects and survives exact native recovery after owner death", t => {
 const scratch = createScratch(t, "cartulary-package-producer-");
 for (const scenario of ["success", "lost_build", "lost_create", "malformed_create", "cleanup_fail", "all_cleanup_fail", "death_pending", "death_build", "death_create"]) {
  const directory = path.join(scratch, scenario); fs.mkdirSync(directory, { mode: 0o700 });
  const fixture = createPackageProducerFixture(directory, scenario);
  const resultParent = path.join(directory, "results"), runID = `package-${scenario}`, runRoot = path.join(resultParent, runID);
  writeEvidenceFile(runRoot, "run-manifest.json", buildReaderEvidence({ runID, target: "package-release", commandID: "cartulary.harness.command.package_release.v1" }).manifest);
  const originalManifest = fs.readFileSync(path.join(runRoot, "run-manifest.json"));
  const identityFile = path.join(directory, "identity.json"), resultFile = path.join(directory, "result.json");
  const environment = { ...process.env, ...fixture.environment, CARTULARY_TEST_RESULTS_DIR: resultParent, CARTULARY_TEST_RUN_ID: runID,
   PACKAGE_FIXTURE_OPTIONS: JSON.stringify({ runtime: { repoRoot, runRoot, runID, scratchRoot: path.join(directory, "private") }, identity: identityFile, result: resultFile, producer: { sourceRoot: fixture.source, archiveRoot: fixture.archiveRoot } }) };
  const invocation = spawnSync(process.execPath, ["--input-type=module", "-e", packageWorkerScript], { cwd: repoRoot, env: environment, encoding: "utf8", timeout: 30000 });
  assert.ifError(invocation.error);
  const identity = JSON.parse(fs.readFileSync(identityFile));
  const recover = () => spawnSync(process.env.CARTULARY_HARNESS_RUNTIME_BIN || path.join(repoRoot, workspaceLayout.toolbin, "cartulary-harness"), ["recover"], {
   cwd: repoRoot, env: { ...environment, RESULTS_DIR: runRoot, RUNTIME_ROOT: identity.root, RUNTIME_LEASE_ID: identity.leaseID }, encoding: "utf8", timeout: 30000,
  });
  try {
   const death = scenario.startsWith("death_"), cleanupFailure = scenario.includes("cleanup_fail");
   assert.equal(invocation.signal, death ? "SIGKILL" : null, `${scenario}: ${invocation.stderr}`);
   assert.equal(invocation.status, death ? null : scenario === "success" ? 0 : 1, `${scenario}: ${invocation.stderr}`);
   assert.equal(invocation.stdout + invocation.stderr, "", "private boundary diagnostics remain private");
   let state = JSON.parse(fs.readFileSync(fixture.state));
   assert.equal(state.stages.length, 1, `${scenario}: producer reached admitted build`);
   assert.ok(state.stages[0].startsWith(identity.root + path.sep), "stage belongs to the existing private runtime");
   assert.equal(fs.existsSync(state.stages[0]), death, "ordinary finalization removes staging even when Docker cleanup fails");
   assert.equal(fs.existsSync(identity.root), death || cleanupFailure);
   if (!death) {
    const result = JSON.parse(fs.readFileSync(resultFile));
    assert.equal(result.failed, scenario !== "success");
    assert.deepEqual(result.cleanup, scenario === "cleanup_fail" ? ["application"] : scenario === "all_cleanup_fail" ? ["application", "build_tag"] : []);
    if (scenario === "cleanup_fail") assert.equal(result.reason, "cleanup_error");
   }
   const calls = fs.readFileSync(fixture.log, "utf8").trim().split("\n").map(line => JSON.parse(line));
   const build = calls.findIndex(call => call[1] === "build");
   assert.equal(calls.slice(0, build).filter(call => call[1] === "ps" || call[1] === "image" && call[2] === "ls").length, 4, "all names observed absent before effects");
   const created = calls.filter(call => call[1] === "create");
   for (const call of created) {
    const name = call[call.indexOf("--name") + 1], labels = Object.fromEntries(call.flatMap((arg, index) => arg === "--label" ? [call[index + 1].split("=")] : []));
    assert.match(name, /^cartulary-package-[a-f0-9-]{36}-(application|postgres|seaweedfs)$/u);
    assert.deepEqual(Object.keys(labels).sort(), ["cartulary.package.allocation", "cartulary.package.lease", "cartulary.package.role", "cartulary.package.run"]);
    assert.equal(labels["cartulary.package.run"], runID); assert.equal(labels["cartulary.package.lease"], identity.leaseID);
    assert.equal(name, `cartulary-package-${labels["cartulary.package.allocation"]}-${labels["cartulary.package.role"]}`);
   }
   if (death || cleanupFailure) {
    if (cleanupFailure) {
     const failed = recover(); assert.ifError(failed.error); assert.equal(failed.status, 12, failed.stdout + failed.stderr);
     assert.ok(fs.existsSync(identity.root), "failed recovery preserves exact proof for a subsequent attempt");
    }
    fs.writeFileSync(fixture.control, JSON.stringify({ scenario: "success" }));
    const recovered = recover(); assert.ifError(recovered.error); assert.equal(recovered.status, 0, recovered.stdout + recovered.stderr);
    assert.match(recovered.stdout, /recovery complete; original execution result unchanged/u);
    const outcomes = JSON.parse(recovered.stdout.trim().split("\n").at(-1));
    const counts = outcomes.package_resource;
    assert.equal(Object.values(counts).reduce((sum, value) => sum + value, 0), death ? 4 : scenario === "all_cleanup_fail" ? 2 : 1, "native recovery observes exactly the still-pending resources");
   }
   assert.equal(fs.existsSync(identity.root), false, "successful closure purges owned private detail");
   state = JSON.parse(fs.readFileSync(fixture.state));
   assert.deepEqual(Object.keys(state.containers), [borrowedContainer]);
   for (const [name, id] of Object.entries(borrowedTags)) assert.equal(state.tags[name], id);
   assert.equal(Object.keys(state.tags).some(name => name.startsWith("cartulary/package-build:")), false);
   if (!["lost_build", "death_pending", "death_build"].includes(scenario)) assert.equal(state.tags[`cartulary/local:sha256-${imageIDs[0].slice(7)}`], imageIDs[0], "immutable output survives temporary cleanup");
   assert.deepEqual(fs.readFileSync(path.join(runRoot, "run-manifest.json")), originalManifest);
   const receiptFile = path.join(runRoot, "package-release/artifacts/package.json");
   assert.equal(fs.existsSync(receiptFile), scenario === "success");
   if (scenario === "success") {
    const receipt = JSON.parse(fs.readFileSync(receiptFile));
    assert.deepEqual(receipt.images, ["application", "postgres", "seaweedfs"].map((role, index) => ({ role, id: imageIDs[index] })));
    assert.deepEqual(fs.readdirSync(fixture.archiveRoot), [receipt.archive], "only immutable archive survives production");
   }
  } finally {
   // Even a failed assertion must settle the synthetic external state through
   // the same native owner before the fixture's private directory is removed.
   if (fs.existsSync(identity.root)) {
    fs.writeFileSync(fixture.control, JSON.stringify({ scenario: "success" }));
    const result = recover(); assert.ifError(result.error); assert.equal(result.status, 0, result.stdout + result.stderr);
   }
  }
 }
});
