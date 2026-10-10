import { createHash } from "node:crypto";
import path from "node:path";
import { readLocalFile, readLocalChunks } from "../runtime/secure-local-files.mjs";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";

// Callback state is provisional until this function returns after digest closure.
// The facade exposes no partial aggregate when validation fails.
export function readResourceObservations(runRoot, run, consumeObservation, execution = null) {
  let index;
  try { index = parseStrictJSON(readLocalFile(path.join(runRoot, "diagnostics/resource-index.json"), { maximum: 4 * 1024 ** 2 }).toString("utf8")); }
  catch (error) { if (error.code === "ENOENT") return { availability: "not_observed", mode: run.manifest.instrumentation.mode }; throw error; }
  validateSchemaSync("cartulary.harness_resource_index.v2", index);
  for (const key of ["run_id", "source_digest", "graph_digest"]) if (index[key] !== run.manifest[key]) throw new Error("resource identity mismatch");
  if (index.policy_digest !== run.manifest.instrumentation.policy_digest) throw new Error("resource policy mismatch");
  if (index.sample_digest === null) return { availability: "collector_failed", mode: run.manifest.instrumentation.mode };
  const roster = new Map(index.processes.map((entry) => [entry.process_ref, entry]));
  if (roster.size !== index.processes.length || new Set(index.processes.map((entry) => entry.identity_digest)).size !== roster.size) throw new Error("duplicate resource identity");
  const relations = execution?.records.filter((record) => record.kind === "relationship") ?? [];
  const proof = new Set(index.processes.map((entry) => entry.identity_digest));
  if (relations.some((record) => !proof.has(record.identity_digest))) throw new Error("relationship has no physical proof");
  for (const entry of roster.values()) {
    const selected = relations.filter((record) => record.identity_digest === entry.identity_digest && record.provenance === "registered");
    const owners = new Set(selected.map((record) => JSON.stringify([record.unit_id, record.allocation_ref, record.invocation_id])));
    const owner = owners.size === 1 ? selected[0] : null;
    entry.unit_id = owner?.allocation_ref ? null : owner?.unit_id ?? null;
    entry.allocation_ref = owner?.allocation_ref ?? null;
    entry.attribution_start_ms = owner?.observed_elapsed_ms ?? Infinity;
  }
  const hash = createHash("sha256");
  let tail = "", bytes = 0, sequence = 0, previousTime = 0;
  const consume = (line) => {
    if (!line || Buffer.byteLength(line) > 65536) throw new Error("invalid resource record length");
    const record = parseStrictJSON(line);
    validateSchemaSync("cartulary.harness_resource_sample.v1", record);
    if (record.availability !== "available" && Object.keys(record.metrics).length) throw new Error("unavailable scope contains measurements");
    if (record.seq !== ++sequence || record.elapsed_ms < previousTime) throw new Error("invalid resource ordering");
    previousTime = record.elapsed_ms;
    if (record.scope === "process" && !roster.has(record.scope_ref)) throw new Error("unknown sampled process");
    if (record.scope === "process" && record.identity_digest !== roster.get(record.scope_ref).identity_digest) throw new Error("sample start identity mismatch");
    consumeObservation(record, roster);
  };
  for (const chunk of readLocalChunks(path.join(runRoot, "diagnostics/resource-samples.ndjson"), { maximum: 64 * 1024 ** 2 })) {
    hash.update(chunk); bytes += chunk.length; tail += chunk.toString("utf8");
    for (let newline; (newline = tail.indexOf("\n")) >= 0;) { consume(tail.slice(0, newline)); tail = tail.slice(newline + 1); }
    if (Buffer.byteLength(tail) > 65536) throw new Error("resource line too long");
  }
  if (tail || bytes !== index.sample_bytes || sequence !== index.samples || `sha256:${hash.digest("hex")}` !== index.sample_digest) throw new Error("resource artifact does not close");
  return { index, mode: run.manifest.instrumentation.mode };
}
