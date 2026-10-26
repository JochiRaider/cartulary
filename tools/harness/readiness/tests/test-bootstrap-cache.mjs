import assert from "node:assert/strict";
import path from "node:path";
import test from "node:test";
import { spawnSync } from "node:child_process";
test("bootstrap artifact cache preserves content identity and failure handling", () => {
  const result = spawnSync("bash", ["tools/harness/readiness/tests/test-cache-artifact.sh"], {
    cwd: path.resolve(import.meta.dirname, "../../../.."),
    env: { ...process.env, NODE_BIN: process.execPath }, encoding: "utf8",
  });
  assert.equal(result.status, 0, result.stdout + result.stderr);
});
