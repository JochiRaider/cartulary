import assert from "node:assert/strict";
import { mkdtempSync, readFileSync, rmSync, statSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { holdReviewSession, reviewProfile, withReviewResources } from "../design-review.mjs";
import { reviewTotp } from "../design-review-seed.mjs";
import { writeReviewSamples } from "../design-review-samples.mjs";

assert.equal(reviewProfile(), "network_flow_claimed");
for (const profile of ["default", "network_flow_claimed"]) assert.equal(reviewProfile(profile), profile);
assert.equal(reviewProfile(" default "), "default");
for (const invalid of ["", "base", "../default", "enterprise"]) assert.throws(() => reviewProfile(invalid));
assert.equal(reviewTotp("GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", 59000), "287082");

await assert.rejects(holdReviewSession({ check: async () => { throw new Error("stale attachment"); } }), /stale attachment/u);
for (const checking of [true, false]) {
  const controller = new AbortController();
  let checks = 0;
  await holdReviewSession({ signal: controller.signal, check: async () => {
    checks++;
    if (checking) {
      controller.abort();
      controller.abort();
      throw new Error("interrupted attachment check");
    }
    setTimeout(() => controller.abort(), 1);
  } });
  assert.equal(checks, 1, "interruption must stop without another attachment check");
}

for (const failure of [null, "acquire", "prepare", "hold", "close"]) {
  const events = [];
  const step = (name) => async () => { events.push(name); if (failure === name) throw new Error(name); return "lease"; };
  const promise = withReviewResources({ acquire: step("acquire"), prepare: step("prepare"), hold: step("hold"), close: step("close"), finish: step("finish") });
  if (failure) await assert.rejects(promise, new RegExp(failure)); else await promise;
  assert.deepEqual(events.slice(-2), ["close", "finish"]);
  if (failure === "acquire") assert.ok(!events.includes("prepare"));
  if (failure === "prepare") assert.ok(!events.includes("hold"));
}
await assert.rejects(withReviewResources({
  acquire: async () => { throw new Error("primary"); }, prepare: async () => {}, hold: async () => {},
  close: async () => { throw new Error("secondary"); }, finish: async () => {},
}), /primary/u);

const root = path.resolve(import.meta.dirname, "../../../..");
const directory = mkdtempSync(path.join(os.tmpdir(), "cartulary-review-samples-"));
try {
  const first = writeReviewSamples(root, path.join(directory, "first"));
  const second = writeReviewSamples(root, path.join(directory, "second"));
  assert.deepEqual(readFileSync(path.join(first, "network-flow.csv")), readFileSync(path.join(second, "network-flow.csv")));
  assert.notDeepEqual(readFileSync(path.join(first, "incident.tar")), readFileSync(path.join(second, "incident.tar")));
  assert.equal(statSync(path.join(first, "incident.tar")).mode & 0o777, 0o600);
  for (const sample of ["network-flow.csv", "workbook-partial.xlsx", "evidence.txt", "timeline.csv", "reference-pack.tar"]) {
    assert.equal(statSync(path.join(first, sample)).mode & 0o777, 0o600);
  }
  const archive = readFileSync(path.join(first, "reference-pack.tar"));
  assert.equal(archive.subarray(257, 263).toString(), "ustar\0");
  assert.ok(archive.includes(Buffer.from("manifest_sha256_v1")));
} finally { rmSync(directory, { recursive: true, force: true }); }
process.stdout.write("design review lifecycle and sample preparation: pass\n");
