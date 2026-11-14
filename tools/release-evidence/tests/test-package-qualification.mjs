import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import test from "node:test";
import childProcess, { spawnSync } from "node:child_process";
import { syncBuiltinESMExports } from "node:module";
import { packageFixtures } from "../reference-pack-fixtures.mjs";
import { createScratch } from "../../harness/test-support/scratch.mjs";
import { buildReaderEvidence, writeEvidenceFile } from "../../harness/observability/tests/retained-fixture.mjs";
import { workspaceLayout } from "../../workspace_layout.generated.mjs";
import { createDeploymentFixture, borrowedDeploymentID } from "./package-deployment-fixture.mjs";
import { assertPackageQualificationSource, packageDestinationProject, packageSourceProjectPattern } from "../package-qualification-project.mjs";
import { qualifyPackage } from "../package-qualification.mjs";
const repoRoot = path.resolve(import.meta.dirname, "../../..");
const source = "cartulary-qualification-11111111-2222-4333-8444-555555555555-source";

test("qualification project grammar has one source and one derived destination, without legacy aliases", async () => {
 assert.equal(packageSourceProjectPattern.test(source), true);
 const env = { CARTULARY_PACKAGE_WORK_DIR: "/private/allocation/installation", CARTULARY_PACKAGE_SOURCE_PROJECT: source, CARTULARY_PACKAGE_DESTINATION_PROJECT: packageDestinationProject(source) };
 assert.equal(assertPackageQualificationSource(env.CARTULARY_PACKAGE_WORK_DIR, source, env), env.CARTULARY_PACKAGE_DESTINATION_PROJECT);
 for (const change of [value => { delete value.CARTULARY_PACKAGE_WORK_DIR; }, value => { value.CARTULARY_PACKAGE_SOURCE_PROJECT = "borrowed"; }, value => { value.CARTULARY_PACKAGE_DESTINATION_PROJECT = source; }]) {
  const changed = { ...env }; change(changed); assert.throws(() => assertPackageQualificationSource(env.CARTULARY_PACKAGE_WORK_DIR, source, changed));
 }
 for (const work of ["relative", "/operator/installation", null]) assert.throws(() => assertPackageQualificationSource(work, source, env));
 assert.equal(packageDestinationProject(source), "cartulary-qualification-11111111-2222-4333-8444-555555555555-destination");
 for (const value of ["cartularymvpsmoke123", "cartularymvprecoverysmk123", source.replace("-4333-", "-3333-"), source.replace("-8444-", "-7444-"), source + "destination", source.replace("-source", "-destination"), "--all", null]) assert.throws(() => packageDestinationProject(value));
 for (const mode of ["unknown", "__proto__", "constructor", ""]) await assert.rejects(qualifyPackage(mode, { environment: {} }), error => error.failure_reason === "usage_error");
});

const worker = `
import fs from "node:fs";
import {createRuntime} from ${JSON.stringify(new URL("../../harness/runtime/session-client.mjs", import.meta.url).href)};
import {qualifyPackage} from ${JSON.stringify(new URL("../package-qualification.mjs", import.meta.url).href)};
import {createLaunchContext} from ${JSON.stringify(new URL("../../harness/runtime/launch-context.mjs", import.meta.url).href)};
const options=JSON.parse(process.env.PACKAGE_QUALIFICATION_OPTIONS),runtime=createRuntime(options.runtime);
fs.writeFileSync(options.identity,JSON.stringify({root:runtime.root,leaseID:runtime.leaseID}));
Object.assign(process.env,runtime.producerEnvironment());
Object.assign(process.env,createLaunchContext({repoRoot:options.runtime.repoRoot,environment:process.env,commandID:"cartulary.harness.command.package_inspect.v1",unitID:"fixture"}).environment);
process.env.PACKAGE_QUALIFICATION_OWNER_PID=String(process.pid);
let failure=null;
try {await qualifyPackage(options.mode,{workflowRoot:options.workflowRoot});} catch(error) {failure={reason:error.failure_reason??"unclassified",diagnostic:error.failure_reason?error.message:"unclassified",cleanup:(error.cleanupFailures??[]).map(value=>value.cleanup_step??"workload")};}
let closed=true;try{runtime.close();}catch{closed=false;}
fs.writeFileSync(options.result,JSON.stringify({failure,closed}));process.exitCode=failure||!closed?1:0;
`;
const payload = `
import fs from "node:fs";import path from "node:path";
import {populateDeploymentFixture} from ${JSON.stringify(new URL("./package-deployment-fixture.mjs", import.meta.url).href)};
import {publishCommandFailure} from ${JSON.stringify(new URL("../../harness/runtime/command-failure.mjs", import.meta.url).href)};
const options=JSON.parse(process.env.PACKAGE_QUALIFICATION_OPTIONS),allocation={workspace:path.dirname(process.env.CARTULARY_PACKAGE_WORK_DIR),sourceProject:process.env.CARTULARY_PACKAGE_SOURCE_PROJECT,destinationProject:process.env.CARTULARY_PACKAGE_DESTINATION_PROJECT};
populateDeploymentFixture(process.env.PACKAGE_DEPLOYMENT_STATE,allocation);
fs.writeFileSync(options.capture,JSON.stringify({...allocation,temporary:process.env.CARTULARY_PACKAGE_TEMP_DIR,installation:process.env.CARTULARY_PACKAGE_WORK_DIR,mode:process.env.CARTULARY_PACKAGE_QUALIFICATION,pid:process.pid}));
process.stdout.write("private fixture workflow output");
if(["classified","contradictory"].includes(options.scenario))publishCommandFailure(options.runtime.repoRoot,{failure_class:"artifact",failure_reason:"artifact_error"});
if(options.scenario==="death"){process.kill(Number(process.env.PACKAGE_QUALIFICATION_OWNER_PID),"SIGKILL");setInterval(()=>{},1000);}
else if(options.scenario==="interrupt"){process.kill(Number(process.env.PACKAGE_QUALIFICATION_OWNER_PID),"SIGTERM");setInterval(()=>{},1000);}
else process.exitCode=options.scenario==="failure"||options.scenario==="both"||options.scenario==="classified"?7:options.scenario==="configuration"?2:0;
`;

test("qualification controller owns all modes before work and joins native recovery across failure and death", t => {
 const scratch = createScratch(t, "cartulary-package-qualification-");
 const cases = [
  ...["package", "reference-pack", "credential-capacity", "recovery"].map(mode => [mode, "success"]),
  ...["failure", "configuration", "classified", "contradictory", "interrupt", "death", "cleanup", "both", "collision", "missing_receipt"].map(scenario => ["package", scenario]),
  ...["package", "recovery"].flatMap(mode => ["setup_failure", "consume_failure"].map(scenario => [mode, scenario])),
 ];
 for (const [mode, scenario] of cases) {
  const runID = `${mode}-${scenario}`, directory = path.join(scratch, runID);
  const fixture = createDeploymentFixture(directory, ["cleanup", "both"].includes(scenario) ? "source_failure" : scenario);
  const resultParent = path.join(directory, "results"), runRoot = path.join(resultParent, runID);
  writeEvidenceFile(runRoot, "run-manifest.json", buildReaderEvidence({ runID, target: "package-inspect", commandID: "cartulary.harness.command.package_inspect.v1" }).manifest);
  const manifest = fs.readFileSync(path.join(runRoot, "run-manifest.json"));
  if (scenario !== "missing_receipt") writeEvidenceFile(runRoot, "package-release/artifacts/package.json", {
   schema_id: "cartulary.local_release_receipt.v1", manifest_sha256: "a".repeat(64), archive_sha256: "b".repeat(64), archive: `${"a".repeat(64)}.tar`, operator_sha256: "c".repeat(64),
   images: ["application", "postgres", "seaweedfs"].map(role => ({ role, id: `sha256:${"c".repeat(64)}` })), inventory_disposition: "content_complete_review_unestablished",
  });
  const workflowRoot = ["setup_failure", "consume_failure"].includes(scenario) ? repoRoot : path.join(directory, "workflow");
  if (workflowRoot !== repoRoot) {
   fs.mkdirSync(path.join(workflowRoot, "tools/release-evidence"), { recursive: true });
   const payloadFile = path.join(directory, "payload.mjs"); fs.writeFileSync(payloadFile, payload);
   for (const name of ["check-standup-package-smoke.sh", "check-standup-operational-recovery-smoke.sh"]) fs.writeFileSync(path.join(workflowRoot, "tools/release-evidence", name), '#!/bin/sh\nexec "$NODE_BIN" "$PACKAGE_QUALIFICATION_PAYLOAD"\n', { mode: 0o700 });
  }
  const identityFile = path.join(directory, "identity.json"), resultFile = path.join(directory, "result.json"), capture = path.join(directory, "capture.json"), temporaryCapture = path.join(directory, "temporary.json");
  fs.writeFileSync(path.join(directory, "bin/mktemp"), `#!${process.execPath}\nimport fs from "node:fs";import {spawnSync} from "node:child_process";fs.writeFileSync(process.env.PACKAGE_TEMPORARY_CAPTURE,JSON.stringify(process.argv.slice(2)));const r=spawnSync("/usr/bin/mktemp",process.argv.slice(2),{stdio:"inherit"});process.exitCode=r.status??1;\n`, { mode: 0o700 });
  const env = { ...process.env, ...fixture.env, NODE_BIN: process.execPath, CARTULARY_TEST_RESULTS_DIR: resultParent, CARTULARY_TEST_RUN_ID: runID, CARTULARY_HARNESS_EXECUTION_ENABLED: "0",
   PACKAGE_QUALIFICATION_PAYLOAD: path.join(directory, "payload.mjs"), PACKAGE_TEMPORARY_CAPTURE: temporaryCapture,
   PACKAGE_QUALIFICATION_OPTIONS: JSON.stringify({ runtime: { repoRoot, runRoot, runID, scratchRoot: path.join(directory, "private") }, identity: identityFile, result: resultFile, capture, mode, scenario, workflowRoot }) };
  const result = spawnSync(process.execPath, ["--input-type=module", "-e", worker], { cwd: repoRoot, env, encoding: "utf8", timeout: 60000 });
  assert.ifError(result.error);
  const identity = JSON.parse(fs.readFileSync(identityFile));
  const recover = () => spawnSync(process.env.CARTULARY_HARNESS_RUNTIME_BIN || path.join(repoRoot, workspaceLayout.toolbin, "cartulary-harness"), ["recover"], {
   cwd: repoRoot, env: { ...env, RESULTS_DIR: runRoot, RUNTIME_ROOT: identity.root, RUNTIME_LEASE_ID: identity.leaseID }, encoding: "utf8", timeout: 60000,
  });
  let primary;
  try {
   assert.equal(result.stdout + result.stderr, "", `${runID}: private output escaped`);
   assert.equal(result.signal, scenario === "death" ? "SIGKILL" : null, runID);
   assert.equal(result.status, scenario === "death" ? null : scenario === "success" ? 0 : 1, `${runID}: ${fs.existsSync(resultFile) ? fs.readFileSync(resultFile,"utf8") : "no result"}`);
   const needsRecovery = ["death", "cleanup", "both"].includes(scenario);
   assert.equal(fs.existsSync(identity.root), needsRecovery, `${runID}: runtime ownership ${fs.existsSync(resultFile) ? fs.readFileSync(resultFile,"utf8") : "no result"}`);
   if (fs.existsSync(capture)) {
    const observed = JSON.parse(fs.readFileSync(capture));
    assert.ok(observed.workspace.startsWith(identity.root + path.sep));
    assert.equal(observed.temporary, path.join(observed.workspace, "tmp"));
    assert.equal(observed.installation, path.join(observed.workspace, "installation"));
    assert.equal(observed.mode, mode); assert.equal(observed.destinationProject, packageDestinationProject(observed.sourceProject));
   }
   if (scenario === "consume_failure") {
    const args = JSON.parse(fs.readFileSync(temporaryCapture));
    assert.equal(args[0], "-d"); assert.ok(args[1].startsWith(identity.root + path.sep)); assert.match(args[1], /\/tmp\/package-extract\.XXXXXXXX$/u);
   }
   if (scenario !== "death") {
    const record = JSON.parse(fs.readFileSync(resultFile));
    const expected = scenario === "success" ? null : scenario === "interrupt" ? "cancelled_or_interrupted" : scenario === "configuration" ? "configuration_error" : scenario === "cleanup" || scenario === "collision" ? "cleanup_error" : scenario === "missing_receipt" ? "unclassified" : scenario === "classified" || scenario === "consume_failure" ? "artifact_error" : scenario === "contradictory" ? "scheduler_accounting_error" : "test_assertion_failure";
    assert.equal(record.failure?.reason ?? null, expected, runID);
    if (scenario === "both") assert.deepEqual(record.failure.cleanup, ["source.container", "source.volume", "source.network", "shared.workspace"]);
    if (!["collision", "missing_receipt"].includes(scenario)) {
     const target = mode === "recovery" ? "standup-operational-recovery-smoke" : mode === "credential-capacity" ? "credential-capacity-assessment" : `standup-${mode}-smoke`;
     const cleanup = JSON.parse(fs.readFileSync(path.join(runRoot, target, "artifacts/cleanup.json")));
     assert.equal(cleanup.cleanup, needsRecovery ? "failed" : "passed");
    }
   }
   if (needsRecovery) {
    fs.writeFileSync(fixture.control, JSON.stringify({ scenario: "success" }));
    const recovered = recover(); assert.ifError(recovered.error); assert.equal(recovered.status, 0, recovered.stdout + recovered.stderr);
    assert.match(recovered.stdout, /recovery complete; original execution result unchanged/u);
    assert.equal(fs.existsSync(identity.root), false);
   }
   const state = JSON.parse(fs.readFileSync(fixture.state));
   assert.deepEqual(Object.keys(state.container), [borrowedDeploymentID]); assert.deepEqual(Object.keys(state.volume), ["borrowed_volume"]); assert.deepEqual(Object.keys(state.network), [borrowedDeploymentID]);
   assert.deepEqual(fs.readFileSync(path.join(runRoot, "run-manifest.json")), manifest);
   if (scenario === "missing_receipt") assert.equal(fs.readFileSync(fixture.log, "utf8"), "");
  } catch (error) { primary = error; }
  finally {
   if (fs.existsSync(identity.root)) {
    fs.writeFileSync(fixture.control, JSON.stringify({ scenario: "success" })); const recovered = recover();
    try { assert.ifError(recovered.error); assert.equal(recovered.status, 0, `${runID}: ${recovered.stdout}${recovered.stderr}`); }
    catch (error) { primary = primary ? new AggregateError([primary, error], `${runID}: primary and cleanup failure`) : error; }
   }
  }
  if (primary) throw primary;
 }
});


test("Reference Pack fixture staging stays inside its disposable destination across archive failures", t => {
 const root = createScratch(t, "cartulary-package-fixture-staging-");
 const borrowed = path.join(root, "borrowed"); fs.writeFileSync(borrowed, "preserve");
 const original = childProcess.spawnSync;
 try {
  for (const fails of [false, true]) {
   const destination = path.join(root, fails ? "failed" : "complete"); let calls = 0;
   childProcess.spawnSync = (binary, args, options) => {
    assert.equal(binary, "tar"); calls++;
    const staging = args[args.indexOf("-C") + 1];
    assert.equal(path.dirname(staging), destination);
    assert.match(path.basename(staging), /^\.pack-producer-/u);
    assert.equal(fs.statSync(staging).mode & 0o777, 0o700);
    assert.ok(fs.existsSync(path.join(staging, "manifest.json")));
    if (fails && calls === 3) return { status: 1 };
    return original(binary, args, options);
   };
   syncBuiltinESMExports();
   if (fails) assert.throws(() => packageFixtures(repoRoot, destination), /package fixture archive creation failed/u);
   else packageFixtures(repoRoot, destination);
   assert.equal(calls, fails ? 3 : 6);
   assert.equal(fs.readdirSync(destination).some(name => name.startsWith(".pack-producer-")), false);
   assert.deepEqual(fs.readdirSync(destination).sort(), fails ? ["rotated.tar", "trust.json", "valid.tar"] : ["bad-signature.tar", "collectable.tar", "incompatible.tar", "malformed.tar", "rollback.tar", "rotated.tar", "trust.json", "valid.tar"]);
   assert.equal(fs.readFileSync(borrowed, "utf8"), "preserve");
  }
 } finally { childProcess.spawnSync = original; syncBuiltinESMExports(); }
});
