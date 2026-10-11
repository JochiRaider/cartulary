import path from "node:path";
import { createHash } from "node:crypto";
import { atomicLocalFile } from "../../runtime/secure-local-files.mjs";

// Test data only: deliberately does not share validation or serialization logic
// with the reader. Fault fixtures may publish malformed records for rejection.
export function publishMeasurementFixture(root, identity, measurement) {
  const schema_id = "cartulary.harness_measurement_trace.v1";
  const { samples, ...raw } = measurement;
  const { attempt_id, run_id, source_digest, plan_digest } = identity;
  const records = [{ schema_id, kind: "header", attempt_id, run_id, source_digest, plan_digest },
    ...samples.map((sample, sequence) => ({ schema_id, kind: "memory", sequence, sample })),
    { schema_id, kind: "terminal", raw }];
  const bytes = Buffer.from(records.map((record) => JSON.stringify(record)).join("\n") + "\n");
  const relative = `measurements/${attempt_id}.ndjson`;
  atomicLocalFile(path.join(root, relative), bytes);
  return { path: relative, digest: `sha256:${createHash("sha256").update(bytes).digest("hex")}`, bytes: bytes.length, records: records.length };
}
