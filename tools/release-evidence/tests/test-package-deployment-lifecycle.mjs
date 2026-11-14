import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import test from "node:test";
import { spawnSync } from "node:child_process";
import { createScratch } from "../../harness/test-support/scratch.mjs";
import { buildReaderEvidence, writeEvidenceFile } from "../../harness/observability/tests/retained-fixture.mjs";
import { workspaceLayout } from "../../workspace_layout.generated.mjs";
import { createDeploymentFixture, deploymentWorker, borrowedDeploymentID } from "./package-deployment-fixture.mjs";
const repoRoot = path.resolve(import.meta.dirname, "../../..");

test("deployment allocation and native recovery agree on independent resources and blocked parents", t => {
 const scratch = createScratch(t, "cartulary-deployment-lifecycle-");
 for (const scenario of ["success", "collision", "admission_reply", "source_failure", "lost_helper_reply", "repair_failure", "death", "repair_death"]) {
  const directory = path.join(scratch, scenario), fixture = createDeploymentFixture(directory, scenario);
  const runID = `deployment-${scenario}`, resultParent = path.join(directory, "results"), runRoot = path.join(resultParent, runID);
  writeEvidenceFile(runRoot, "run-manifest.json", buildReaderEvidence({ runID, target: "package-inspect", commandID: "cartulary.harness.command.package_inspect.v1" }).manifest);
  const manifest = fs.readFileSync(path.join(runRoot, "run-manifest.json"));
  const identityFile = path.join(directory, "identity.json"), resultFile = path.join(directory, "result.json"), allocationFile = path.join(directory, "allocation.json");
  const env = { ...process.env, ...fixture.env, CARTULARY_TEST_RESULTS_DIR: resultParent, CARTULARY_TEST_RUN_ID: runID, CARTULARY_HARNESS_EXECUTION_ENABLED: "0",
   PACKAGE_DEPLOYMENT_OPTIONS: JSON.stringify({ runtime: { repoRoot, runRoot, runID, scratchRoot: path.join(directory, "private") }, identity: identityFile, result: resultFile, allocation: allocationFile, scenario }) };
  const result = spawnSync(process.execPath, ["--input-type=module", "-e", deploymentWorker], { cwd: repoRoot, env, encoding: "utf8", timeout: 60000 });
  assert.ifError(result.error);
  const identity = JSON.parse(fs.readFileSync(identityFile));
  const recover = () => spawnSync(process.env.CARTULARY_HARNESS_RUNTIME_BIN || path.join(repoRoot, workspaceLayout.toolbin, "cartulary-harness"), ["recover"], {
   cwd: repoRoot, env: { ...env, RESULTS_DIR: runRoot, RUNTIME_ROOT: identity.root, RUNTIME_LEASE_ID: identity.leaseID }, encoding: "utf8", timeout: 60000,
  });
  try {
   const death = ["death", "repair_death"].includes(scenario);
   assert.equal(result.signal, death ? "SIGKILL" : null, `${scenario}: ${result.stderr}`);
   assert.equal(result.status, death ? null : scenario === "success" ? 0 : 1, `${scenario}: ${result.stderr}`);
   assert.equal(result.stdout + result.stderr, "", "provider diagnostics stay private");
   const hasAllocation = !["collision", "admission_reply"].includes(scenario);
   assert.equal(fs.existsSync(allocationFile), hasAllocation);
   if (hasAllocation) {
    const allocation = JSON.parse(fs.readFileSync(allocationFile));
    assert.ok(allocation.workspace.startsWith(identity.root + path.sep));
    assert.match(allocation.sourceProject, /^cartulary-qualification-[a-f0-9-]{36}-source$/u);
    assert.equal(allocation.destinationProject, allocation.sourceProject.replace(/-source$/u, "-destination"));
    assert.equal(allocation.dependencies.length, 2);
    const calls = fs.readFileSync(fixture.log, "utf8").trim().split("\n").map(value => JSON.parse(value));
    assert.equal(calls.slice(1, 14).filter(value => value[1] === "ls").length, 13, "all resource and name observations precede acquisition");
   }
   if (!death) {
    const record = JSON.parse(fs.readFileSync(resultFile));
    assert.equal(record.failed, scenario !== "success");
    const expected = scenario === "source_failure" ? ["source.container", "source.volume", "source.network", "shared.workspace"]
     : ["lost_helper_reply", "repair_failure"].includes(scenario) ? ["shared.workspace"] : [];
    assert.deepEqual(record.failures, expected);
   }
   if (["success", "collision"].includes(scenario)) assert.equal(fs.existsSync(identity.root), false);
   else {
    assert.equal(fs.existsSync(identity.root), true, "unresolved intent survives for exact native recovery");
    if (scenario === "source_failure") {
     const state = JSON.parse(fs.readFileSync(fixture.state));
     assert.equal(Object.keys(state.container).length, 2); assert.equal(Object.keys(state.volume).length, 2); assert.equal(Object.keys(state.network).length, 2, "destination siblings settled while source parents remained");
     const failed = recover(); assert.ifError(failed.error); assert.equal(failed.status, 12, failed.stdout + failed.stderr);
     const outcomes = JSON.parse(failed.stdout.trim().split("\n").at(-1)).package_deployment;
     assert.equal(outcomes.failed, 1); assert.equal(outcomes.blocked, 3);
    }
    fs.writeFileSync(fixture.control, JSON.stringify({ scenario: "success" }));
    const recovered = recover(); assert.ifError(recovered.error); assert.equal(recovered.status, 0, recovered.stdout + recovered.stderr);
    assert.match(recovered.stdout, /recovery complete; original execution result unchanged/u);
    const outcomes = JSON.parse(recovered.stdout.trim().split("\n").at(-1));
    const expected = scenario === "admission_reply" ? 2 : scenario === "source_failure" ? 4 : scenario === "death" ? 7 : 1;
    assert.equal(Object.values(outcomes.package_deployment).reduce((sum, count) => sum + count, 0), expected);
    assert.equal(fs.existsSync(identity.root), false);
   }
   const state = JSON.parse(fs.readFileSync(fixture.state));
   assert.deepEqual(Object.keys(state.container), [borrowedDeploymentID]);
   assert.deepEqual(Object.keys(state.volume), ["borrowed_volume"]);
   assert.deepEqual(Object.keys(state.network), [borrowedDeploymentID]);
   assert.deepEqual(fs.readFileSync(path.join(runRoot, "run-manifest.json")), manifest);
  } finally {
   if (fs.existsSync(identity.root)) {
    fs.writeFileSync(fixture.control, JSON.stringify({ scenario: "success" }));
    const recovered = recover(); assert.ifError(recovered.error); assert.equal(recovered.status, 0, recovered.stdout + recovered.stderr);
   }
  }
 }
});
