import assert from "node:assert/strict";
import test from "node:test";
import { mkdtempSync, readFileSync, rmSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { createResourceCollector } from "../resource-collector.mjs";
import { nativeCollector } from "../native-collector.mjs";

test("collector death prevents proof publication and never starts another observer", async (t) => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-collector-death-"));
  let worker, starts = 0;
  const digest = `sha256:${"1".repeat(64)}`;
  const collector = createResourceCollector({ runRoot: root, capacities: {}, epoch: process.hrtime.bigint(),
    runtime: { leaseID: "runtime", privatePath: (...parts) => path.join(root, ...parts) },
    manifest: { run_id: "fixture", source_digest: digest, graph_digest: digest, instrumentation: { mode: "basic", policy_digest: digest } },
    workerFactory: (...args) => { starts++; worker = nativeCollector(...args); return worker; },
  });
  t.after(async () => { await collector.stop(); rmSync(root, { recursive: true, force: true }); });
  const proof = { lease: {}, ownership: "owned", borrower_file: "", borrower_id: "" };
  collector.registerContainers(proof);
  const file = path.join(root, "container-observation/proof.json");
  assert.equal(JSON.parse(readFileSync(file)).run_id, "fixture");
  collector.registerContainers(null);
  assert.equal(readFileSync(file, "utf8"), "null");
  await worker.terminate();
  await collector.pause();
  collector.registerContainers(proof);
  assert.equal(readFileSync(file, "utf8"), "null");
  assert.equal(starts, 1);
});
