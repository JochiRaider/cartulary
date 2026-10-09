import { createHash } from "node:crypto";
import path from "node:path";
import { createAtomicNDJSONWriter } from "../scheduler/work-graph/atomic-ndjson.mjs";
import { atomicLocalFile, privateDirectory } from "../runtime/secure-local-files.mjs";
import { validateSchemaSync } from "../contract/index.mjs";

export function createResourceStore({ policy, runRoot, createWriter = createAtomicNDJSONWriter, publish = atomicLocalFile }) {
  const directory = privateDirectory(path.join(runRoot, "diagnostics"));
  const writer = createWriter(path.join(directory, "resource-samples.ndjson"), JSON.stringify);
  const hash = createHash("sha256");
  let samples = 0, bytes = 0, omitted = 0, truncated = false, failed = false;
  return {
    get samples() { return samples; }, get bytes() { return bytes; }, get truncated() { return truncated || failed; },
    async append(sample) {
      if (failed) throw new Error("resource writer failed");
      validateSchemaSync(sample.schema_id, sample);
      const serialized = `${JSON.stringify(sample)}\n`, size = Buffer.byteLength(serialized);
      if (size > policy.maximum_record_bytes || bytes + size > policy.maximum_sample_bytes) { omitted += 1; truncated = true; return; }
      try { await writer.write(sample); } catch (error) { failed = true; throw error; }
      hash.update(serialized); bytes += size; samples += 1;
    },
    async finish(facts, shutdownStart, now) {
      let sampleDigest = null;
      try { if (failed) throw new Error("resource writer failed"); await writer.close(); sampleDigest = `sha256:${hash.digest("hex")}`; }
      catch { failed = true; await writer.abort(); }
      const index = { ...facts, samples, sample_bytes: bytes, sample_digest: sampleDigest,
        omitted_observations: facts.omitted_observations + omitted,
        status: failed ? "collector_failed" : truncated ? "truncated" : facts.status,
        observer: { ...facts.observer, write_bytes: bytes, shutdown_ms: now() - shutdownStart } };
      validateSchemaSync(index.schema_id, index);
      const serialized = `${JSON.stringify(index)}\n`;
      if (Buffer.byteLength(serialized) > policy.maximum_metadata_bytes) throw new Error("resource index limit");
      publish(path.join(directory, "resource-index.json"), serialized);
    },
  };
}
