import path from "node:path";
import { createHash } from "node:crypto";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { readLocalFile } from "../runtime/secure-local-files.mjs";
import { executionRecordsDigest } from "../runtime/execution-observations.mjs";

export function readExecutionIndex(runRoot, run) {
  let index;
  try { index = parseStrictJSON(readLocalFile(path.join(runRoot, "diagnostics/execution-index.json"), { maximum: 4 * 1024 ** 2 }).toString("utf8")); }
  catch (error) { if (error.code === "ENOENT") return null; throw error; }
  validateSchemaSync("cartulary.harness_execution_index.v1", index);
  if (run.manifest.instrumentation.mode !== "basic") throw new Error("execution evidence requires basic diagnostics");
  for (const key of ["run_id", "source_digest", "graph_digest"]) if (index[key] !== run.manifest[key]) throw new Error("execution identity mismatch");
  if (index.policy_digest !== run.manifest.instrumentation.policy_digest || index.records_digest !== executionRecordsDigest(index.records)) throw new Error("execution digest mismatch");
  const refs = new Set(), launches = new Map();
  for (const record of index.records) {
    if (Buffer.byteLength(JSON.stringify(record)) > 65536) throw new Error("execution record exceeds bound");
    if (refs.has(record.ref)) throw new Error("duplicate execution reference");
    refs.add(record.ref);
    if (record.unit_id !== null && record.unit_id !== undefined && !run.registrations.has(record.unit_id)) throw new Error("unknown execution unit");
    if (record.kind === "launch") {
      if (record.run_id !== index.run_id || record.ref !== `launch:${record.invocation_id}`) throw new Error("launch identity mismatch");
      if (record.outcome === "passed" && !record.launched || record.outcome === "spawn_failed" && record.launched) throw new Error("launch outcome contradicts dispatch");
      launches.set(record.invocation_id, record);
      const rows = new Set(record.unit_id ? run.registrations.get(record.unit_id).row_ids : []);
      if (record.row_ids.some((row) => !rows.has(row))) throw new Error("unknown launch row");
    }
  }
  const omitted = new Set(index.omitted_refs);
  for (const ref of omitted) if (refs.has(ref)) throw new Error("execution reference retained and omitted");
  const allocations = new Set(index.records.filter((record) => record.kind === "allocation").map((record) => record.allocation_ref));
  const leaseIDs = new Set();
  const allocationIDs = new Set();
  for (const record of index.records) {
    if (record.kind !== "allocation" && record.allocation_ref && !allocations.has(record.allocation_ref) && !omitted.has(record.allocation_ref)) throw new Error("unknown execution allocation");
    if (record.kind !== "launch" && record.invocation_id && !launches.has(record.invocation_id) && !omitted.has(`launch:${record.invocation_id}`)) throw new Error("unknown execution invocation");
    if (record.kind === "allocation") {
      if (record.ref !== record.allocation_ref || allocationIDs.has(record.allocation_ref)) throw new Error("invalid allocation identity");
      allocationIDs.add(record.allocation_ref);
      if (record.outcome === "passed" && record.ownership === null) throw new Error("successful allocation has unknown ownership");
    }
    if (record.kind === "activity") {
      if (!record.ref.startsWith("activity:")) throw new Error("invalid activity identity");
      if (record.outcome === "incomplete" && (record.duration_ms !== null || record.availability !== "not_observed")) throw new Error("incomplete activity has a duration");
      if (record.availability === "available" ? record.duration_ms === null || record.clock_identity === null || record.resolution_ms === null || record.clock === "unavailable" : record.duration_ms !== null) throw new Error("inconsistent activity clock");
      if (record.clock === "unavailable" && (record.clock_identity !== null || record.resolution_ms !== null)) throw new Error("unavailable clock has identity");
    }
    if (record.kind === "relationship") {
      const { identity_digest, unit_id, allocation_ref, invocation_id, provenance } = record;
      const key = JSON.stringify({ kind: "relationship", identity_digest, unit_id, allocation_ref, invocation_id, provenance });
      if (record.ref !== `relationship:${createHash("sha256").update(key).digest("hex")}`) throw new Error("invalid relationship identity");
    }
    const launch = launches.get(record.invocation_id);
    if (record.kind !== "launch" && launch && record.unit_id !== null && record.unit_id !== launch.unit_id) throw new Error("execution unit and invocation disagree");
    if (record.kind === "lease") {
      if (record.ref !== `lease:${record.lease_ref}` || !/^[A-Za-z0-9_.-]+$/u.test(record.lease_ref)) throw new Error("invalid lease identity");
      const lease = parseStrictJSON(readLocalFile(path.join(runRoot, "_shared/fixture-leases", `${record.lease_ref}.json`)).toString("utf8"));
      validateSchemaSync("cartulary.harness_fixture_lease.v4", lease);
      if (lease.lease_id !== record.lease_ref || lease.capability !== record.capability || lease.ownership !== record.ownership) throw new Error("canonical lease identity mismatch");
      if (leaseIDs.has(record.lease_ref)) throw new Error("duplicate execution lease");
      leaseIDs.add(record.lease_ref);
      const allocation = index.records.find((candidate) => candidate.kind === "allocation" && candidate.allocation_ref === record.allocation_ref);
      if (allocation && (allocation.outcome !== "passed" || allocation.ownership !== record.ownership || allocation.capability !== record.capability)) throw new Error("lease requires successful allocation");
    }
  }
  if (index.status === "complete" && (omitted.size || index.omitted_records !== 0 || index.records.some((record) => record.outcome === "incomplete"))) throw new Error("complete execution index has omissions");
  for (const record of launches.values()) {
    const seen = new Set([record.invocation_id]);
    let parent = record.parent_invocation_id;
    while (parent !== null) {
      if (seen.has(parent)) throw new Error("cyclic launch parentage");
      seen.add(parent);
      if (!launches.has(parent)) {
        if (!omitted.has(`launch:${parent}`)) throw new Error("unknown launch parent");
        break;
      }
      parent = launches.get(parent).parent_invocation_id;
    }
  }
  return index;
}
