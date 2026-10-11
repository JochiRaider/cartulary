import assert from "node:assert/strict";
import path from "node:path";
import { createHash } from "node:crypto";
import { canonicalJSONString, parseStrictJSON, validateSchemaSync } from "../../contract/index.mjs";
import { readLocalChunks } from "../../runtime/secure-local-files.mjs";

const schema = "cartulary.harness_measurement_trace.v1";
const maximumBytes = 67108864, maximumRecord = 65536;
const header = ({ attempt_id, run_id, source_digest, plan_digest }) => ({ schema_id: schema, kind: "header", attempt_id, run_id, source_digest, plan_digest });
export function validateMemoryRecord(sample, sequence) {
  validateSchemaSync(schema, { schema_id: schema, kind: "memory", sequence, sample });
  assert.equal(sample.rss === null, sample.availability === "unavailable", "memory availability mismatch");
}
function freeze(value) {
  for (const child of Object.values(value)) if (child && typeof child === "object") freeze(child);
  return Object.freeze(value);
}

// Exactly one anchored, bounded traversal. Nothing escapes until the terminal
// record, identities, byte/record counts and digest have all closed.
export function loadMeasurementTrace(root, identity, reference) {
  assert.equal(reference?.path, `measurements/${identity.attempt_id}.ndjson`, "measurement trace reference mismatch");
  assert.match(reference.path, /^measurements\/[A-Za-z0-9_.-]+\.ndjson$/u);
  const hash = createHash("sha256"), samples = [];
  let pending = Buffer.alloc(0), bytes = 0, records = 0, raw = null;
  for (const chunk of readLocalChunks(path.join(root, reference.path), { maximum: maximumBytes })) {
    hash.update(chunk); bytes += chunk.length;
    pending = Buffer.concat([pending, chunk]);
    let end;
    while ((end = pending.indexOf(10)) !== -1) {
      assert.ok(end + 1 <= maximumRecord, "measurement record bound");
      const line = pending.subarray(0, end); pending = pending.subarray(end + 1);
      const value = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(line));
      validateSchemaSync(schema, value);
      assert.equal(raw, null, "measurement write after terminal");
      if (records === 0) assert.equal(canonicalJSONString(value), canonicalJSONString(header(identity)), "measurement trace identity mismatch");
      else if (value.kind === "memory") {
        assert.equal(value.sequence, samples.length, "measurement sample sequence mismatch");
        validateMemoryRecord(value.sample, value.sequence); samples.push(value.sample);
      } else { assert.equal(value.kind, "terminal", "measurement transition mismatch"); raw = value.raw; }
      records++;
    }
    assert.ok(pending.length < maximumRecord, "measurement record bound");
  }
  assert.ok(pending.length === 0 && raw !== null, "incomplete measurement trace");
  assert.equal(bytes, reference.bytes, "measurement trace byte mismatch");
  assert.equal(records, reference.records, "measurement trace record mismatch");
  assert.equal(`sha256:${hash.digest("hex")}`, reference.digest, "measurement trace digest mismatch");
  assert.equal(raw.exit_code, identity.exit_code, "measurement exit mismatch");
  assert.equal(raw.reader_starts.length, identity.readers, "measurement reader count mismatch");
  assert.equal(raw.cpu_error_us, raw.reaped_processes * 2, "measurement accounting bound mismatch");
  assert.ok(raw.adopted_descendants <= raw.reaped_processes && (!raw.waited_descendants || raw.losses.length === 0), "measurement accounting coverage mismatch");
  assert.ok(raw.support_cpu_us + raw.support_cpu_error_us >= 0, "support accounting lost");
  return freeze({ ...raw, samples });
}
