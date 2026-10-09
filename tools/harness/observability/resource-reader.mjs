import { createHash } from "node:crypto";
import path from "node:path";
import { readLocalFile, readLocalChunks } from "../runtime/secure-local-files.mjs";
import { validateSchemaSync } from "../contract/index.mjs";

// Callback state is provisional until this function returns after digest closure.
// The facade exposes no partial aggregate when validation fails.
export function readResourceObservations(runRoot, run, consumeObservation) {
  let index;
  try { index = JSON.parse(readLocalFile(path.join(runRoot, "diagnostics/resource-index.json"), { maximum: 4 * 1024 ** 2 })); }
  catch (error) { if (error.code === "ENOENT") return { availability: "not_observed", mode: run.manifest.instrumentation.mode }; throw error; }
  validateSchemaSync("cartulary.harness_resource_index.v1", index);
  for (const key of ["run_id", "source_digest", "graph_digest"]) if (index[key] !== run.manifest[key]) throw new Error("resource identity mismatch");
  if (index.policy_digest !== run.manifest.instrumentation.policy_digest) throw new Error("resource policy mismatch");
  if (index.sample_digest === null) return { availability: "collector_failed", mode: run.manifest.instrumentation.mode };
  const roster = new Map(index.processes.map((entry) => [entry.process_ref, entry]));
  if (roster.size !== index.processes.length || new Set(index.processes.map((entry) => entry.identity_digest)).size !== roster.size) throw new Error("duplicate resource identity");
  for (const entry of roster.values()) if (entry.unit_id !== null && !run.registrations.has(entry.unit_id)) throw new Error("unknown resource unit");
  if (new Set(index.leases.map((lease) => lease.lease_ref)).size !== index.leases.length) throw new Error("duplicate lease correlation");
  for (const lease of index.leases) if (!run.registrations.has(lease.unit_id)) throw new Error("unknown lease unit");
  const hash = createHash("sha256");
  let tail = "", bytes = 0, sequence = 0, previousTime = 0;
  const consume = (line) => {
    if (!line || Buffer.byteLength(line) > 65536) throw new Error("invalid resource record length");
    const record = JSON.parse(line);
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
