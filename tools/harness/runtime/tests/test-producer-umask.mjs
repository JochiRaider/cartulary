import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { lstatSync, mkdirSync, mkdtempSync, rmSync } from "node:fs";
import path from "node:path";
import { tmpdir } from "node:os";
import test from "node:test";
import { pathToFileURL } from "node:url";
import { createContractTestContext } from "../../tests/contract-test-context.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
async function assertProducerUmaskBoundary() {
  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "cartulary-suite-runtime-contract."));
  const repo = path.join(fixtureRoot, "repo");
  const runRoot = path.join(fixtureRoot, "results", "run-runtime-contract");
  mkdirSync(repo, { recursive: true, mode: 0o700 });
  mkdirSync(runRoot, { recursive: true, mode: 0o700 });
  try {
    const descendantRoot = path.join(fixtureRoot, "descendant-output", "nested", "leaf");
    const privateProcessModule = pathToFileURL(path.join(root, "tools/harness/runtime/private-process.mjs")).href;
    execFileSync(process.execPath, [
      "--input-type=module",
      "--eval",
      `
    import { mkdirSync } from "node:fs";
    import { enforcePrivateProcessUmask } from ${JSON.stringify(privateProcessModule)};
    process.umask(0o022);
    enforcePrivateProcessUmask();
    mkdirSync(process.argv[1], { recursive: true });
   `,
      descendantRoot,
    ]);
    for (const directory of [
      path.join(fixtureRoot, "descendant-output"),
      path.join(fixtureRoot, "descendant-output", "nested"),
      descendantRoot,
    ]) {
      assert.equal(lstatSync(directory).mode & 0o777, 0o700);
    }
  }
  finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}
export function registerBoundariesTests() {
  test("producer_umask: suite runtime owns producer permissions and cleanup", async () => {
    await assertProducerUmaskBoundary();
  });
}
