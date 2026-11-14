import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import test from "node:test";
import { spawnSync } from "node:child_process";
import { createScratch } from "../../harness/test-support/scratch.mjs";
import { buildReaderEvidence, writeEvidenceFile } from "../../harness/observability/tests/retained-fixture.mjs";
import { workspaceLayout } from "../../workspace_layout.generated.mjs";
import { hashPackageFile } from "../package-archive.mjs";
import { createPackageProducerFixture, packageWorkerScript } from "./package-producer-fixture.mjs";
const repoRoot = path.resolve(import.meta.dirname, "../../..");

test("unpack and inspection use recoverable snapshots while preserving borrowed destinations and archives", t => {
 const scratch = createScratch(t, "cartulary-package-consumers-");
 const fixture = createPackageProducerFixture(path.join(scratch, "package"), "success");
 const resultParent = path.join(scratch, "results");
 function launch(operation, scenario, receipt) {
  const runID = `${operation}-${scenario}`, directory = path.join(scratch, runID), runRoot = path.join(resultParent, runID);
  fs.mkdirSync(directory, { mode: 0o700 });
  writeEvidenceFile(runRoot, "run-manifest.json", buildReaderEvidence({ runID, target: operation === "produce" ? "package-release" : "package-inspect", commandID: `cartulary.harness.command.${operation === "produce" ? "package_release" : "package_inspect"}.v1` }).manifest);
  if (receipt) writeEvidenceFile(runRoot, "package-release/artifacts/package.json", receipt);
  const destination = path.join(directory, "borrowed"); fs.mkdirSync(destination); fs.writeFileSync(path.join(destination, "sentinel"), "borrowed bytes");
  const identityFile = path.join(directory, "identity.json"), resultFile = path.join(directory, "result.json"), capture = path.join(directory, "capture.json");
  const env = { ...process.env, ...fixture.environment, CARTULARY_TEST_RESULTS_DIR: resultParent, CARTULARY_TEST_RUN_ID: runID,
   PACKAGE_CONSUMER_SCENARIO: scenario, PACKAGE_CONSUMER_CAPTURE: capture,
   PACKAGE_FIXTURE_OPTIONS: JSON.stringify({ runtime: { repoRoot, runRoot, runID, scratchRoot: path.join(directory, "private") }, identity: identityFile, result: resultFile,
    operation, destination, producer: { sourceRoot: fixture.source, archiveRoot: fixture.archiveRoot } }) };
  const result = spawnSync(process.execPath, ["--input-type=module", "-e", packageWorkerScript], { cwd: repoRoot, env, encoding: "utf8", timeout: 30000 });
  assert.ifError(result.error);
  return { result, env, runRoot, destination, capture, identity: JSON.parse(fs.readFileSync(identityFile)) };
 }
 const producer = launch("produce", "success");
 assert.equal(producer.result.status, 0, producer.result.stderr); assert.equal(fs.existsSync(producer.identity.root), false);
 const receipt = JSON.parse(fs.readFileSync(path.join(producer.runRoot, "package-release/artifacts/package.json")));
 const archive = path.join(fixture.archiveRoot, receipt.archive), originalDigest = hashPackageFile(archive);
 const state = fs.readFileSync(fixture.state);
 const tarPath = path.join(path.dirname(fixture.environment.SYFT_BIN), "tar");
 // The extraction child observes the real private snapshot before its ordinary
 // tar call. A forced death exits that child too, leaving no orphan fixture loop.
 fs.writeFileSync(tarPath, `#!${process.execPath}
import fs from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
const args=process.argv.slice(2);
if(args.includes("--no-same-owner")) {
 const snapshot=args[args.indexOf("-xf")+1], destination=args[args.indexOf("-C")+1];
 fs.writeFileSync(process.env.PACKAGE_CONSUMER_CAPTURE,JSON.stringify({snapshot,destination,mode:fs.statSync(snapshot).mode&511,parentMode:fs.statSync(path.dirname(snapshot)).mode&511}));
 if(process.env.PACKAGE_CONSUMER_SCENARIO==="death") {process.kill(process.ppid,"SIGKILL");process.exit(0);}
 if(process.env.PACKAGE_CONSUMER_SCENARIO==="failure") process.exit(1);
}
const result=spawnSync("/usr/bin/tar",args,{stdio:"inherit",timeout:10000});
process.exitCode=result.status??1;
`, { mode: 0o700 });
 for (const operation of ["unpack", "inspect"]) for (const scenario of ["success", "failure", "death"]) {
  const selected = launch(operation, scenario, receipt), { result, identity, env, runRoot, destination } = selected;
  const recover = () => spawnSync(process.env.CARTULARY_HARNESS_RUNTIME_BIN || path.join(repoRoot, workspaceLayout.toolbin, "cartulary-harness"), ["recover"], {
   cwd: repoRoot, env: { ...env, RESULTS_DIR: runRoot, RUNTIME_ROOT: identity.root, RUNTIME_LEASE_ID: identity.leaseID }, encoding: "utf8", timeout: 30000,
  });
  try {
   assert.equal(result.signal, scenario === "death" ? "SIGKILL" : null, `${operation}/${scenario}: ${result.stderr}`);
   assert.equal(result.status, scenario === "death" ? null : scenario === "success" ? 0 : 1, `${operation}/${scenario}: ${result.stderr}`);
   assert.equal(result.stdout + result.stderr, "");
   const captured = JSON.parse(fs.readFileSync(selected.capture));
   assert.ok(captured.snapshot.startsWith(identity.root + path.sep));
   assert.equal(captured.mode, 0o600); assert.equal(captured.parentMode, 0o700);
   assert.equal(fs.existsSync(identity.root), scenario === "death");
   if (operation === "inspect") assert.ok(captured.destination.startsWith(identity.root + path.sep));
   else assert.equal(captured.destination, destination);
   const originalManifest = fs.readFileSync(path.join(runRoot, "run-manifest.json"));
   if (scenario === "death") {
    assert.equal(fs.existsSync(captured.snapshot), true, "snapshot outlives the dead JS producer until exact recovery");
    const recovered = recover(); assert.ifError(recovered.error); assert.equal(recovered.status, 0, recovered.stdout + recovered.stderr);
    assert.match(recovered.stdout, /recovery complete; original execution result unchanged/u);
    assert.deepEqual(JSON.parse(recovered.stdout.trim().split("\n").at(-1)), {}, "private detail requires no external-directory resource kind");
   }
   assert.equal(fs.existsSync(identity.root), false); assert.equal(fs.existsSync(captured.snapshot), false);
   assert.equal(fs.readFileSync(path.join(destination, "sentinel"), "utf8"), "borrowed bytes");
   assert.equal(fs.existsSync(path.join(destination, "release")), operation === "unpack" && scenario === "success");
   const inspection = path.join(runRoot, "package-release/artifacts/inspection.json");
   assert.equal(fs.existsSync(inspection), operation === "inspect" && scenario === "success");
   if (operation === "inspect" && scenario === "success") assert.deepEqual(JSON.parse(fs.readFileSync(inspection)), {
    ...receipt, schema_id: "cartulary.local_release_inspection.v1", installed_without_build_toolchain: true, occupied_destination_rejected: true,
   });
   assert.equal(hashPackageFile(archive), originalDigest); assert.deepEqual(fs.readdirSync(fixture.archiveRoot), [receipt.archive]);
   assert.deepEqual(fs.readFileSync(fixture.state), state, "consumers and recovery cannot dispose immutable or borrowed images");
   assert.deepEqual(fs.readFileSync(path.join(runRoot, "run-manifest.json")), originalManifest);
  } finally {
   if (fs.existsSync(identity.root)) { const recovered = recover(); assert.ifError(recovered.error); assert.equal(recovered.status, 0, recovered.stdout + recovered.stderr); }
  }
 }
});
