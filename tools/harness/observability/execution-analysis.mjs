// Local activities are observations, never additive canonical cost. This bounded
// view keeps full facts in the terminal index and makes every omission explicit.
export function analyzeExecution(index) {
  if (!index) return { availability: "not_observed" };
  const view = { availability: "available", completeness: index.status, omitted_records: index.omitted_records };
  const relations = index.records.filter((record) => record.kind === "relationship");
  const explicit = new Map();
  for (const record of relations.filter((record) => record.provenance === "registered")) {
    const owners = explicit.get(record.identity_digest) ?? new Set();
    owners.add(JSON.stringify([record.unit_id, record.allocation_ref, record.invocation_id]));
    explicit.set(record.identity_digest, owners);
  }
  let bytes = 0;
  for (const [kind, name] of [["launch", "launches"], ["allocation", "allocations"], ["lease", "leases"], ["relationship", "relationships"], ["activity", "activities"]]) {
    const all = index.records.filter((record) => record.kind === kind);
    const records = [];
    for (const record of all) {
      const value = kind === "relationship" ? { fact: record, attribution:
        (explicit.get(record.identity_digest)?.size ?? 0) > 1 ? "conflicted" : record.provenance === "observed_descendant" ? "inferred" : record.allocation_ref ? "allocation_context" : "registered" } : record;
      const size = Buffer.byteLength(JSON.stringify(value));
      if (records.length < 20 && bytes + size <= 128 * 1024) { records.push(value); bytes += size; }
    }
    view[name] = { records, omitted: all.length - records.length };
  }
  return view;
}
