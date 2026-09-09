import assert from "node:assert/strict";
import { chmodSync, lstatSync, linkSync, mkdtempSync, mkdirSync, readFileSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { createSuiteRuntime } from "../../runtime/suite-runtime.mjs";
import { publishFrontendOutput, resolveFrontendArtifact, sealFrontendArtifact } from "../../readiness/frontend-artifact.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const scratch = mkdtempSync(path.join(os.tmpdir(), "frontend-artifact-test-"));
const runtimes = [];
const profile = { id: "production", producer_target: "build-web", entries: ["index.html"] };
function fixture(runID, version) {
  const runRoot = path.join(scratch, runID);
  mkdirSync(runRoot, { mode: 0o700 });
  writeFileSync(path.join(runRoot, "run-manifest.json"), JSON.stringify({
    source_digest: `sha256:${"1".repeat(64)}`, toolchain_digest: `sha256:${"2".repeat(64)}`,
  }), { mode: 0o600 });
  const runtime = createSuiteRuntime({ repoRoot: root, runRoot, runID, scratchRoot: path.join(scratch, "private") });
  runtimes.push(runtime);
  const staging = runtime.privatePath("staging");
  mkdirSync(staging, { mode: 0o700 });
  writeFileSync(path.join(staging, "index.html"), `<script src="/lazy.js"></script>${version}`);
  writeFileSync(path.join(staging, "lazy.js"), `export default ${JSON.stringify(version)}`);
  const environment = {
    CARTULARY_TEST_RESULTS_DIR: scratch, CARTULARY_TEST_RUN_ID: runID,
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root,
    CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runID,
  };
  const directory = sealFrontendArtifact({ repoRoot: root, runRoot, runtime, profile, staging });
  return { directory, environment, runtime, runRoot };
}
try {
  const first = fixture("first", "original");
  const second = fixture("second", "replacement");
  const destination = path.join(scratch, "published");
  const dangling = path.join(scratch, "dangling-publication");
  symlinkSync(path.join(scratch, "missing"), dangling);
  assert.throws(() => publishFrontendOutput(first.directory, dangling), /symlink/u);
  publishFrontendOutput(first.directory, destination);
  const before = resolveFrontendArtifact(root, "build-web", first.environment);
  publishFrontendOutput(second.directory, destination);
  for (let reload = 0; reload < 5; reload++) {
    assert.match(readFileSync(path.join(first.directory, "index.html"), "utf8"), /original/u);
    assert.match(readFileSync(path.join(first.directory, "lazy.js"), "utf8"), /original/u);
    assert.equal(resolveFrontendArtifact(root, "build-web", first.environment).receiptDigest, before.receiptDigest);
  }
  assert.match(readFileSync(path.join(destination, "index.html"), "utf8"), /replacement/u);
  assert.throws(() => resolveFrontendArtifact(root, "build-web-measurement", first.environment));
  const receiptFile = path.join(first.runRoot, before.receiptRef);
  const bytes = readFileSync(receiptFile);
  const bad = JSON.parse(bytes);
  bad.source_digest = `sha256:${"3".repeat(64)}`;
  writeFileSync(receiptFile, JSON.stringify(bad));
  assert.throws(() => resolveFrontendArtifact(root, "build-web", first.environment), /source_digest mismatch/u);
  writeFileSync(receiptFile, bytes);
  writeFileSync(path.join(first.directory, "lazy.js"), "tampered");
  assert.throws(() => resolveFrontendArtifact(root, "build-web", first.environment), /digest mismatch/u);
  const partial = first.runtime.privatePath("partial");
  mkdirSync(partial, { mode: 0o700 });
  assert.throws(() => sealFrontendArtifact({ repoRoot: root, runRoot: first.runRoot, runtime: first.runtime, profile, staging: partial }));
  symlinkSync(path.join(second.directory, "index.html"), path.join(partial, "index.html"));
  assert.throws(() => sealFrontendArtifact({ repoRoot: root, runRoot: first.runRoot, runtime: first.runtime, profile, staging: partial }), /invalid/u);
  rmSync(path.join(partial, "index.html"));
  chmodSync(path.join(second.directory, "index.html"), 0o644);
  linkSync(path.join(second.directory, "index.html"), path.join(partial, "index.html"));
  assert.throws(() => sealFrontendArtifact({ repoRoot: root, runRoot: first.runRoot, runtime: first.runtime, profile, staging: partial }), /share a file inode/u);
  assert.equal(lstatSync(path.join(second.directory, "index.html")).mode & 0o777, 0o644);
  chmodSync(path.join(second.directory, "index.html"), 0o600);
  rmSync(path.join(partial, "index.html"));
  assert.throws(() => sealFrontendArtifact({ repoRoot: root, runRoot: first.runRoot, runtime: first.runtime, profile, staging: scratch }), /private suite root/u);
  second.runtime.close();
  second.runtime.close();
  assert.throws(() => resolveFrontendArtifact(root, "build-web", second.environment));
} finally {
  for (const runtime of runtimes) runtime.close();
  rmSync(scratch, { recursive: true, force: true });
}
