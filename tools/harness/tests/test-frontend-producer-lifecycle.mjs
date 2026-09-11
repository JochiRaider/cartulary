import assert from "node:assert/strict";
import { fork } from "node:child_process";
import { once } from "node:events";
import { existsSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { createSuiteRuntime } from "../runtime/suite-runtime.mjs";
import { claimFrontendProducer, publishFrontendOutput, resolveFrontendArtifact, sealFrontendArtifact } from "../readiness/frontend-artifact.mjs";
import { executeUnitProcess } from "../scheduler/work-graph/executor.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
const scratch = mkdtempSync(path.join(os.tmpdir(), "frontend-producer-lifecycle-"));
const runRoot = path.join(scratch, "run");
mkdirSync(runRoot, { mode: 0o700 });
writeFileSync(path.join(runRoot, "run-manifest.json"), JSON.stringify({ source_digest: `sha256:${"1".repeat(64)}`, toolchain_digest: `sha256:${"2".repeat(64)}` }), { mode: 0o600 });
const runtime = createSuiteRuntime({ repoRoot: root, runRoot, runID: "run", scratchRoot: path.join(scratch, "private") });
const profile = { id: "production", producer_target: "build-web", entries: ["index.html"] };
try {
  for (const name of ["first", "second"]) {
    const staging = runtime.privatePath(name);
    mkdirSync(staging, { mode: 0o700 });
    writeFileSync(path.join(staging, "index.html"), name, { mode: 0o600 });
    const seal = () => sealFrontendArtifact({ repoRoot: root, runRoot, runtime, profile, staging });
    if (name === "first") seal();
    else assert.throws(seal, (error) => error.failure_class === "harness" && error.failure_reason === "scheduler_accounting_error", "duplicate publication is an attributed producer violation");
  }
} finally { runtime.close(); rmSync(scratch, { recursive: true, force: true }); }

const processRoot = mkdtempSync(path.join(os.tmpdir(), "frontend-producer-process-"));
const runtimes = [];
const children = [];
function fixture(id) {
  const runRoot = path.join(processRoot, id);
  mkdirSync(runRoot, { mode: 0o700 });
  const runtime = createSuiteRuntime({ repoRoot: root, runRoot, runID: id, scratchRoot: path.join(processRoot, "private") });
  runtimes.push(runtime);
  const environment = { CARTULARY_TEST_RESULTS_DIR: processRoot, CARTULARY_TEST_RUN_ID: id, CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: id };
  return { runtime, runRoot, environment };
}
function worker(f, id = "production", mode = "producer", boundary = "", publication = "") {
  const child = fork(path.join(root, "tools/harness/tests/frontend-producer-fixture.mjs"), [mode, id, boundary, publication], { env: { ...process.env, ...f.environment }, stdio: ["ignore", "ignore", "pipe", "ipc", "pipe"] });
  children.push(child);
  return { child, first: once(child, "message"), closed: once(child, "exit") };
}
try {
  const f = fixture("race");
  const contenders = [worker(f, "production", "producer-race"), worker(f, "production", "producer-race")];
  assert.deepEqual(await Promise.all(contenders.map(async (producer) => (await producer.first)[0])), [{ state: "ready" }, { state: "ready" }]);
  const admissions = contenders.map((producer) => once(producer.child, "message"));
  for (const producer of contenders) producer.child.send("claim");
  const outcomes = (await Promise.all(admissions)).map(([message]) => message);
  assert.equal(outcomes.filter((message) => message.state === "admitted").length, 1);
  const first = contenders[outcomes.findIndex((message) => message.state === "admitted")];
  const second = contenders[outcomes.findIndex((message) => message.state === "rejected")];
  assert.throws(() => resolveFrontendArtifact(root, "build-web", f.environment), /ENOENT/u);
  assert.deepEqual(outcomes.find((message) => message.state === "rejected"), { state: "rejected", failure_class: "harness", failure_reason: "scheduler_accounting_error" });
  assert.equal((await second.closed)[0], 11);
  const measurement = worker(f, "measurement");
  assert.equal((await measurement.first)[0].state, "admitted", "profiles do not serialize each other's builds");
  const independent = worker(fixture("independent"));
  assert.equal((await independent.first)[0].state, "admitted", "runs have independent producer claims");
  for (const producer of [first, measurement, independent]) {
    const sealed = once(producer.child, "message");
    producer.child.send("complete");
    assert.equal((await sealed)[0].state, "sealed");
    assert.equal((await producer.closed)[0], 0);
  }
  const before = resolveFrontendArtifact(root, "build-web", f.environment);
  assert.equal(readdirSync(f.runtime.root).filter((name) => name.startsWith("build-start-production-")).length, 1, "only the winning producer enters build work");
  const receiptFile = path.join(f.runRoot, before.receiptRef);
  const receiptBytes = readFileSync(receiptFile);
  for (const [key, value] of Object.entries({ run_id: "another-run", artifact_id: "measurement", producer_unit_id: "target:build-web-measurement", source_digest: `sha256:${"3".repeat(64)}`, toolchain_digest: `sha256:${"4".repeat(64)}`, content_digest: `sha256:${"5".repeat(64)}` })) {
    writeFileSync(receiptFile, JSON.stringify({ ...before.receipt, [key]: value }));
    assert.throws(() => resolveFrontendArtifact(root, "build-web", f.environment), (error) => error.failure_reason === "artifact_error", key);
  }
  writeFileSync(receiptFile, receiptBytes.toString().replace('"schema_id":', '"run_id":"duplicate", "schema_id":'));
  assert.throws(() => resolveFrontendArtifact(root, "build-web", f.environment), /duplicate/u);
  writeFileSync(receiptFile, receiptBytes);
  const duplicate = worker(f);
  assert.equal((await duplicate.first)[0].state, "rejected", "completion does not grant a second producer");
  await duplicate.closed;
  assert.equal(resolveFrontendArtifact(root, "build-web", f.environment).receiptDigest, before.receiptDigest);
  assert.equal(readFileSync(path.join(before.directory, "index.html"), "utf8"), "sealed");
  resolveFrontendArtifact(root, "build-web-measurement", f.environment);

  const aborted = fixture("aborted");
  const pending = worker(aborted);
  await pending.first;
  pending.child.kill("SIGKILL");
  await pending.closed;
  assert.throws(() => resolveFrontendArtifact(root, "build-web", aborted.environment));
  assert.throws(() => claimFrontendProducer(aborted.runtime, profile), /already admitted/u);
  aborted.runtime.close();
  assert.equal(resolveFrontendArtifact(root, "build-web", f.environment).receiptDigest, before.receiptDigest);

  for (const boundary of ["receipt", "conventional"]) {
    const interrupted = fixture(`interrupted-${boundary}`);
    const publication = boundary === "conventional" ? path.join(processRoot, "conventional-output") : "";
    if (publication) publishFrontendOutput(before.directory, publication);
    const producer = worker(interrupted, "production", "producer", boundary, publication);
    assert.equal((await producer.first)[0].state, "admitted");
    const stopped = once(producer.child.stdio[4], "data");
    producer.child.send("complete");
    assert.equal((await stopped)[0].toString(), boundary);
    assert.equal(readFileSync(interrupted.runtime.privatePath("frontend-production", "index.html"), "utf8"), "sealed");
    if (boundary === "receipt") {
      assert.equal(existsSync(path.join(interrupted.runRoot, "build-web/frontend-artifact.json")), false);
      assert.throws(() => resolveFrontendArtifact(root, "build-web", interrupted.environment));
    } else {
      assert.equal(existsSync(publication), false, "conventional publication is interrupted after moving its previous output");
      resolveFrontendArtifact(root, "build-web", interrupted.environment);
    }
    producer.child.kill("SIGKILL");
    assert.deepEqual(await producer.closed, [null, "SIGKILL"], "publication interruption cannot report producer success");
    assert.throws(() => claimFrontendProducer(interrupted.runtime, profile), /already admitted/u);
    if (boundary === "receipt") assert.throws(() => resolveFrontendArtifact(root, "build-web", interrupted.environment));
    else resolveFrontendArtifact(root, "build-web", interrupted.environment);
    interrupted.runtime.close();
    assert.equal(resolveFrontendArtifact(root, "build-web", f.environment).receiptDigest, before.receiptDigest, "another run remains readable after interrupted publication and cleanup");
  }

  const failed = fixture("compiler-failure");
  const result = await executeUnitProcess({ unit_id: "target:build-web", kind: "artifact", command: { executable: process.execPath, args: ["tools/harness/readiness/frontend-artifact.mjs", "build", "build-web", process.execPath, "tools/harness/tests/frontend-producer-fixture.mjs", "compiler-failure"], environment: { CARTULARY_TEST_TARGET: "build-web" } }, timeout_ms: 15000 }, { cwd: root, environment: failed.environment });
  assert.equal(result.failure_reason, "tool_diagnostic_failure", JSON.stringify(result));
  assert.equal(result.exit_code, 1);
  assert.throws(() => resolveFrontendArtifact(root, "build-web", failed.environment));
  assert.throws(() => claimFrontendProducer(failed.runtime, profile), /already admitted/u);
} finally {
  for (const child of children) if (child.exitCode === null && child.signalCode === null) child.kill("SIGKILL");
  for (const runtime of runtimes) runtime.close();
  rmSync(processRoot, { recursive: true, force: true });
}
