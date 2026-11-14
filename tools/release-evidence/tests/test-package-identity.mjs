import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import test from "node:test";
import childProcess from "node:child_process";
import { syncBuiltinESMExports } from "node:module";
import { createScratch } from "../../harness/test-support/scratch.mjs";
import { buildReaderEvidence, writeEvidenceFile } from "../../harness/observability/tests/retained-fixture.mjs";
import { packageRunContext, publishPackageEvidence, readPackageJSON, readPackageReceipt, validatePackageReceipt, verifyPackageIdentity } from "../package-identity.mjs";
import { validateLiveReleaseReadiness } from "../release-readiness-evidence.mjs";
import { unpack } from "../package-release.mjs";

const receipt = () => ({ schema_id: "cartulary.local_release_receipt.v1", manifest_sha256: "a".repeat(64),
 archive_sha256: "b".repeat(64), archive: `${"a".repeat(64)}.tar`, operator_sha256: "c".repeat(64),
 images: ["application", "postgres", "seaweedfs"].map(role => ({ role, id: `sha256:${"d".repeat(64)}` })),
 inventory_disposition: "content_complete_review_unestablished" });
const manifest = runID => buildReaderEvidence({ runID, target: "package-inspect", commandID: "cartulary.harness.command.package_inspect.v1" }).manifest;

test("cryptographic build receipt admission has a complete independent current specimen", t => {
 const root = createScratch(t, "cartulary-package-crypto-receipt-");
 const value = { schema_id: "cartulary.cryptographic_build_receipt.v1", toolchain: "go1.27.3", selector: "v1.0.0-c2097c7c",
  service_version: "v1.0.0", archive_sha256: "a".repeat(64), binary_sha256: "b".repeat(64),
  build_info: { GoVersion: "go1.27.3", Path: "example.test/application", Main: { Path: "example.test/module", Version: "(devel)" },
   Deps: [{ Path: "example.test/library", Version: "v1.0.0", Sum: "h1:fixture", Replace: { Path: "example.test/replacement", Version: "v1.1.0" } }],
   Settings: [{ Key: "GOFIPS140", Value: "v1.0.0-c2097c7c" }, { Key: "empty_setting" }] } };
 const admit = specimen => {
  writeEvidenceFile(root, "crypto.json", specimen);
  return readPackageJSON(root, "crypto.json", "cartulary.cryptographic_build_receipt.v1");
 };
 assert.deepEqual(admit(value), value);
 const noDependencies = structuredClone(value); delete noDependencies.build_info.Deps;
 assert.deepEqual(admit(noDependencies), noDependencies);
 for (const mutate of [v => delete v.toolchain, v => delete v.build_info, v => { v.binary_sha256 = "wrong"; },
  v => { v.extra = true; }, v => { v.build_info.extra = true; }, v => { v.build_info.Main.extra = true; },
  v => { v.build_info.Deps[0].extra = true; }, v => { v.build_info.Deps[0].Replace.extra = true; },
  v => { v.build_info.Deps[0].Replace.Replace = { Path: "nested" }; }, v => { v.build_info.Settings[0].extra = true; },
  v => delete v.build_info.Settings[0].Key, v => { v.build_info.Settings = []; },
  v => { v.toolchain = "x".repeat(4097); }, v => { v.build_info.Deps = Array(16385).fill({ Path: "example.test/library" }); }]) {
  const changed = structuredClone(value); mutate(changed);
  assert.throws(() => admit(changed), /^Error: package evidence admission failed$/u);
 }
});

test("package receipts require closed shape and internally consistent identities before comparison", () => {
 assert.deepEqual(validatePackageReceipt(receipt()), receipt());
 assert.deepEqual(verifyPackageIdentity(receipt(), [["consumer", receipt()]]), { manifest_sha256: "a".repeat(64), archive_sha256: "b".repeat(64) });
 for (const mutate of [v => { v.archive = `${"f".repeat(64)}.tar`; }, v => v.images.reverse(),
  v => { v.images[1] = v.images[0]; }, v => v.images.pop(), v => { v.extra = true; },
  v => { v.images[0].extra = true; }, v => { v.schema_id = "legacy"; }, v => delete v.operator_sha256]) {
  const value = receipt(); mutate(value);
  assert.throws(() => validatePackageReceipt(value));
  assert.throws(() => verifyPackageIdentity(value, [["equally malformed", value]]));
  assert.throws(() => verifyPackageIdentity(receipt(), [["malformed consumer", value]]));
 }
});

test("package evidence is a strict bounded no-link snapshot and publication preserves prior bytes", t => {
 const root = createScratch(t, "cartulary-package-evidence-");
 const relative = "package-release/artifacts/package.json";
 publishPackageEvidence(root, "package", receipt());
 assert.deepEqual(readPackageReceipt(root), receipt());
 const file = path.join(root, relative), before = fs.readFileSync(file);
 assert.equal(fs.statSync(file).mode & 0o777, 0o600);
 assert.equal(fs.statSync(path.dirname(file)).mode & 0o777, 0o700);
 assert.throws(() => publishPackageEvidence(root, "package", receipt()));
 assert.deepEqual(fs.readFileSync(file), before);
 assert.deepEqual(fs.readdirSync(path.dirname(file)), ["package.json"]);
 for (const bytes of [before.toString().replace('{', '{"archive":"wrong",'), '{"unfinished":', Buffer.from([0xff]), '{}tail', JSON.stringify({ ...receipt(), extra: true })]) {
  fs.writeFileSync(file, bytes);
  assert.throws(() => readPackageReceipt(root), /^Error: package evidence admission failed$/u);
 }
 fs.truncateSync(file, 16 * 1024 * 1024 + 1);
 assert.throws(() => readPackageReceipt(root), /^Error: package evidence admission failed$/u);
 fs.writeFileSync(file, before);
 fs.symlinkSync(relative, path.join(root, "alias.json"));
 fs.symlinkSync(path.dirname(file), path.join(root, "linked"));
 for (const candidate of ["alias.json", "linked/package.json", "../outside.json", "/outside.json", "missing.json", "package-release"]) {
  assert.throws(() => readPackageJSON(root, candidate, "cartulary.local_release_receipt.v1"), /^Error: package evidence admission failed$/u);
 }
 const open = fs.openSync; let probes = 0;
 try {
  fs.openSync = () => { probes++; throw new Error("forbidden probe"); }; syncBuiltinESMExports();
  assert.throws(() => readPackageJSON(root, "README", "cartulary.local_release_receipt.v1"), /^Error: package evidence admission failed$/u);
  assert.equal(probes, 0);
 } finally { fs.openSync = open; syncBuiltinESMExports(); }
 const inspected = { ...receipt(), schema_id: "cartulary.local_release_inspection.v1", installed_without_build_toolchain: true, occupied_destination_rejected: true };
 publishPackageEvidence(root, "inspection", inspected);
 assert.deepEqual(readPackageJSON(root, "package-release/artifacts/inspection.json", inspected.schema_id), inspected);
 assert.throws(() => publishPackageEvidence(root, "unknown", inspected));
 assert.deepEqual(fs.readFileSync(file), before);
});

test("package run context and live readiness require exact current typed evidence", t => {
 const parent = createScratch(t, "cartulary-package-current-");
 const env = { CARTULARY_TEST_RESULTS_DIR: parent, CARTULARY_TEST_RUN_ID: "package-fixture", CARTULARY_HARNESS_GRAPH_CHILD: "1" };
 const runRoot = path.join(parent, env.CARTULARY_TEST_RUN_ID), m = manifest(env.CARTULARY_TEST_RUN_ID);
 assert.throws(() => packageRunContext(env), /admission failed/u);
 writeEvidenceFile(runRoot, "run-manifest.json", m);
 assert.deepEqual(packageRunContext(env), { runRoot, manifest: m });
 for (const mutate of [v => delete v.CARTULARY_TEST_RUN_ID, v => { v.CARTULARY_TEST_RUN_ID = "../escape"; },
  v => { v.CARTULARY_TEST_RESULTS_DIR = "relative"; }, v => delete v.CARTULARY_TEST_RESULTS_DIR]) {
  const value = { ...env }; mutate(value); assert.throws(() => packageRunContext(value), /explicit current run context/u);
 }
 writeEvidenceFile(runRoot, "run-manifest.json", { ...m, run_id: "previous-run" });
 assert.throws(() => packageRunContext(env), /^Error: package run identity mismatch$/u);
 writeEvidenceFile(runRoot, "run-manifest.json", m);
 publishPackageEvidence(runRoot, "package", receipt());
 const consumers = ["standup-package-smoke", "standup-reference-pack-smoke", "standup-operational-recovery-smoke", "credential-capacity-assessment"];
 for (const name of consumers) writeEvidenceFile(runRoot, `${name}/artifacts/package.json`, receipt());
 validateLiveReleaseReadiness(env);
 assert.throws(() => validateLiveReleaseReadiness({ ...env, CARTULARY_HARNESS_GRAPH_CHILD: "0" }), /private graph operation/u);
 for (const name of consumers) {
  const relative = `${name}/artifacts/package.json`;
  fs.unlinkSync(path.join(runRoot, relative));
  writeEvidenceFile(parent, `historical/${relative}`, receipt());
  assert.throws(() => validateLiveReleaseReadiness(env), /admission failed/u);
  writeEvidenceFile(runRoot, relative, { ...receipt(), archive_sha256: "f".repeat(64) });
  assert.throws(() => validateLiveReleaseReadiness(env), /package identity mismatch/u);
  writeEvidenceFile(runRoot, relative, receipt());
 }
 validateLiveReleaseReadiness(env);
});

test("unpack rejects malformed receipt before any extractor executes", t => {
 const parent = createScratch(t, "cartulary-package-unpack-admission-");
 const runID = "unpack-fixture", runRoot = path.join(parent, runID);
 writeEvidenceFile(runRoot, "run-manifest.json", manifest(runID));
 writeEvidenceFile(runRoot, "package-release/artifacts/package.json", { ...receipt(), archive: "../other.tar" });
 const previous = { root: process.env.CARTULARY_TEST_RESULTS_DIR, run: process.env.CARTULARY_TEST_RUN_ID, spawn: childProcess.spawnSync };
 let executions = 0;
 try {
  process.env.CARTULARY_TEST_RESULTS_DIR = parent; process.env.CARTULARY_TEST_RUN_ID = runID;
  childProcess.spawnSync = () => { executions++; throw new Error("extractor must not run"); }; syncBuiltinESMExports();
  assert.throws(() => unpack(path.join(parent, "destination")), /admission failed/u);
  assert.equal(executions, 0); assert.equal(fs.existsSync(path.join(parent, "destination")), false);
 } finally {
  childProcess.spawnSync = previous.spawn; syncBuiltinESMExports();
  if (previous.root === undefined) delete process.env.CARTULARY_TEST_RESULTS_DIR; else process.env.CARTULARY_TEST_RESULTS_DIR = previous.root;
  if (previous.run === undefined) delete process.env.CARTULARY_TEST_RUN_ID; else process.env.CARTULARY_TEST_RUN_ID = previous.run;
 }
});
