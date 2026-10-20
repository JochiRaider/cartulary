// Deterministic packing within one already-proved compatibility partition.
// Lifecycle costs are paid once per invocation; items remain indivisible.
export function packInvocationCohorts(items, { startup_ms, cleanup_ms }, timeoutMS) {
  const overhead = startup_ms + cleanup_ms;
  const cohorts = [];
  for (const item of [...items].sort((a, b) => b.cost - a.cost || (a.id < b.id ? -1 : a.id > b.id ? 1 : 0))) {
    if (!Number.isFinite(item.cost) || item.cost < 0) throw new Error("invalid cohort item cost");
    let cohort = cohorts.find((candidate) => candidate.cost + item.cost <= timeoutMS);
    if (!cohort) { cohort = { cost: overhead, items: [] }; cohorts.push(cohort); }
    cohort.items.push(item);
    cohort.cost += item.cost;
  }
  return cohorts;
}
