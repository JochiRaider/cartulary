import path from "node:path";
import { createHash } from "node:crypto";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { readLocalFile } from "../runtime/secure-local-files.mjs";
import { executionRecordsDigest } from "../runtime/execution-observations.mjs";

export function readExecutionIndex(runRoot, run) {
  let index;
  try { index = parseStrictJSON(readLocalFile(path.join(runRoot, "diagnostics/execution-index.json"), { maximum: 24 * 1024 ** 2 }).toString("utf8")); }
  catch (error) { if (error.code === "ENOENT") return null; throw error; }
  validateSchemaSync("cartulary.harness_execution_index.v5", index);
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
      validateSchemaSync("cartulary.harness_fixture_lease.v5", lease);
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
  validateAllocationLifecycles(index, run);
  validateHostAdmission(index.host_admission, run);
  return index;
}

function validateAllocationLifecycles(index, run) {
  const section = index.allocation_lifecycle;
  const { allocation_lifecycle: _, host_admission: __, ...ordinary } = index;
  if (Buffer.byteLength(JSON.stringify(ordinary)) > 4 * 1024 ** 2) throw new Error("ordinary execution metadata exceeds bound");
  if (Buffer.byteLength(JSON.stringify(section)) > 4 * 1024 ** 2 || section.records_digest !== executionRecordsDigest(section.records)) throw new Error("allocation lifecycle size or digest mismatch");
  if (section.status === "complete" && (section.omitted_records || section.omitted_refs.length)) throw new Error("complete allocation lifecycle has omissions");
  if (section.omitted_records < section.omitted_refs.length) throw new Error("allocation omissions contradict identities");
  const allocations = new Set(), leases = new Set();
  for (const record of section.records) {
    if (Buffer.byteLength(JSON.stringify(record)) > 65536 || allocations.has(record.allocation_ref) || section.omitted_refs.includes(record.allocation_ref)) throw new Error("duplicate or oversized allocation lifecycle");
    allocations.add(record.allocation_ref);
    if (record.triggering_unit_id !== null && !run.registrations.has(record.triggering_unit_id)) throw new Error("unknown retirement trigger unit");
    for (const interval of [record.acquisition, record.retirement]) {
      if (interval.end_ms < interval.start_ms || Math.abs(interval.duration_ms - (interval.end_ms - interval.start_ms)) > 1e-6) throw new Error("invalid allocation interval");
    }
    if (record.retirement.start_ms < record.acquisition.end_ms) throw new Error("allocation retired before acquisition");
    for (const borrow of record.borrows) {
      if (leases.has(borrow.lease_ref) || !run.registrations.has(borrow.unit_id)) throw new Error("invalid allocation borrow identity");
      leases.add(borrow.lease_ref);
      const observed = index.records.find((r) => r.kind === "lease" && r.lease_ref === borrow.lease_ref);
      if (observed && (observed.allocation_ref !== record.allocation_ref || observed.unit_id !== borrow.unit_id)) throw new Error("allocation borrow contradicts lease observation");
    }
    const observed = index.records.find((r) => r.kind === "allocation" && r.allocation_ref === record.allocation_ref);
    if (observed && observed.capability !== "browser_stack") throw new Error("lifecycle belongs to non-browser allocation");
  }
  if (section.status === "complete") {
    for (const record of index.records) if (record.kind === "allocation" && record.capability === "browser_stack" && !allocations.has(record.allocation_ref)) throw new Error("complete lifecycle misses observed browser allocation");
  }
}

// This validates producer facts and their clock/identity bindings. It deliberately
// does not recompute weighted admission or infer a blocker between observations.
export function validateHostAdmission(section, run) {
  if (Buffer.byteLength(JSON.stringify(section)) > 16 * 1024 ** 2 || section.records_digest !== executionRecordsDigest(section.records)) throw new Error("host admission size or digest mismatch");
  if (section.status === "complete" && section.omitted_records !== 0) throw new Error("complete host admission has omissions");
  const requests = new Map(), previous = new Map(), durations = new Map();
  for (const r of section.records) {
    if (Buffer.byteLength(JSON.stringify(r)) > 65536 || !run.registrations.has(r.unit_id)) throw new Error("unknown or oversized host admission record");
    if (r.end_ms < r.start_ms || r.duration_ms > r.end_ms - r.start_ms + 1e-6) throw new Error("invalid host admission duration");
    const host = [...(run.phases?.values() ?? [])].find((p) => p.unit_id === r.unit_id && p.phase === "host_admission");
    if (!host || r.start_ms < host.start_ms - 1 || r.end_ms > host.end_ms + 1) throw new Error("host admission observation outside canonical phase");
    if (r.kind === "decision") {
      if (r.request_ref !== r.decision.request_ref || r.duration_ms !== 0) throw new Error("invalid admission decision identity or duration");
      const old = requests.get(r.request_ref);
      if (old && (old.unit_id !== r.unit_id || old.sequence !== r.decision.sequence || old.mode !== r.decision.mode)) throw new Error("admission request identity conflict");
      requests.set(r.request_ref, { unit_id:r.unit_id, sequence:r.decision.sequence, mode:r.decision.mode });
      if (r.decision.granted && (r.outcome !== "committed" || Object.keys(r.decision.current_deficits).length || r.decision.priority_blockers.length || r.decision.exclusive_blockers.length)) throw new Error("granted admission has blockers");
      for (const key of ["required_ancestors", "active_holders", "residents", "prospective"]) {
        const refs = r.decision[key].map((h) => h.reference);
        if (new Set(refs).size !== refs.length) throw new Error("duplicate admission holder");
      }
      if (r.decision.residents.some((h) => r.decision.prospective.some((p) => p.reference === h.reference))) throw new Error("resident is also prospective");
    } else if(r.kind==="retirement_evaluation") {
      const e=r.evaluation;
      if(r.request_ref!==e.request_ref || r.duration_ms!==0) throw new Error("invalid retirement evaluation identity or duration");
      const resources=new Set([...Object.keys(e.before_deficits),...Object.keys(e.after_deficits)]);
      const relief=[...resources].some(k=>(e.after_deficits[k]??0)<(e.before_deficits[k]??0));
      const increase=[...resources].some(k=>(e.after_deficits[k]??0)>(e.before_deficits[k]??0));
      if(e.eligible!==(r.outcome==="committed" && relief && !increase)) throw new Error("retirement eligibility contradicts declared deficits");
    } else if (r.request_ref !== null) throw new Error("unassigned admission record has request identity");
    if (r.kind === "phase") {
      durations.set(r.unit_id, (durations.get(r.unit_id) ?? 0) + r.duration_ms);
      if (durations.get(r.unit_id) > host.end_ms - host.start_ms + 1) throw new Error("host phase duration exceeds canonical interval");
    }
    if (r.kind === "retirement" && (new Set(r.resident_refs).size !== r.resident_refs.length || r.observations !== 1 || Math.abs(r.duration_ms - (r.end_ms-r.start_ms)) > 1e-6)) throw new Error("invalid retirement observation");
    const key = `${r.unit_id}:${r.kind}:${r.phase ?? (r.kind==="retirement_evaluation" ? `${r.request_ref}:${r.evaluation.resident_refs.join(",")}` : r.request_ref) ?? r.resident_refs.join(",")}`;
    if (previous.has(key) && r.start_ms < previous.get(key)) throw new Error("overlapping or reordered admission observations");
    previous.set(key, r.end_ms);
  }
}
