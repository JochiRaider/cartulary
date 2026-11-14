import { isDeepStrictEqual } from "node:util";
import { browserCompatibility } from "./browser-compatibility.mjs";
import { semanticJSONDigest } from "../contract/index.mjs";
import { packInvocationCohorts } from "../scheduler/work-graph/cohorts.mjs";

// The generated group is a compatibility partition. Only selected catalog rows
// become executable cohorts, so selection never inherits an obsolete file split.
export function browserCohort(partition, rows, policy) {
  const ordered = [...rows].sort((a, b) => a.row_id < b.row_id ? -1 : a.row_id > b.row_id ? 1 : 0);
  const available = new Set(partition.selectedRowIDs);
  const compatibility = browserCompatibility(partition, policy);
  if (ordered.length === 0 || new Set(ordered.map((row) => row.row_id)).size !== ordered.length ||
      (partition.stateScope !== "test_owned" && ordered.length !== 1) ||
      ordered.some((row) => !available.has(row.row_id) || row.runner !== "playwright" ||
        !isDeepStrictEqual(browserCompatibility(partition, policy, row), compatibility) ||
        !partition.specs.includes(row.selector.file))) {
    throw new Error(`browser cohort violates compatibility partition ${partition.name}`);
  }
  const digest = semanticJSONDigest({
    compatibility,
    selections: ordered.map((row) => ({
      row_id: row.row_id, file: row.selector.file,
      cases: row.selector.titles.map((title, index) => ({
        title, scenario_id: row.selector.scenario_ids?.[index] ?? null,
      })).sort((a, b) => a.title < b.title ? -1 : a.title > b.title ? 1 : 0),
    })),
  }).replace(/^sha256:/u, "");
  return {
    ...partition,
    partitionName: partition.name,
    name: `cohort-${digest}`,
    selectedRowIDs: ordered.map((row) => row.row_id),
    specs: [...new Set(ordered.map((row) => row.selector.file))].sort(),
  };
}

const compareItems = (a, b) => b.cost - a.cost || (a.id < b.id ? -1 : a.id > b.id ? 1 : 0);

// Balanced packing for ordinary browser cohorts. The cohort count is the larger
// of the first-fit count and the smallest of the lane limit, the selected row
// count, and the number of cohorts whose selected work can each cover one
// allocation acquisition. Rows are assigned in descending cost to the cohort
// with the lowest projected cost; a multi-row cohort beyond the timeout
// envelope adds one cohort and repeats. Estimates never grant capacity.
export function packBalancedBrowserCohorts(items, { startup_ms, cleanup_ms }, timeoutMS, { laneLimit, minimumSelectedWorkMs }) {
  if (!Number.isSafeInteger(laneLimit) || laneLimit < 1) throw new Error("balanced browser packing requires a positive lane limit");
  if (!Number.isSafeInteger(minimumSelectedWorkMs) || minimumSelectedWorkMs < 1) {
    throw new Error("balanced browser packing requires a positive minimum selected work estimate");
  }
  if (!items.length) return [];
  const ordered = [...items].sort(compareItems);
  const overhead = startup_ms + cleanup_ms;
  const selectedWork = ordered.reduce((total, item) => total + item.cost, 0);
  const useful = Math.max(1, Math.floor(selectedWork / minimumSelectedWorkMs));
  const firstFit = packInvocationCohorts(ordered, { startup_ms, cleanup_ms }, timeoutMS).length;
  for (let count = Math.max(firstFit, Math.min(laneLimit, ordered.length, useful)); count <= ordered.length; count += 1) {
    const cohorts = Array.from({ length: count }, () => ({ cost: overhead, items: [] }));
    for (const item of ordered) {
      const cohort = cohorts.reduce((lowest, candidate) => candidate.cost < lowest.cost ? candidate : lowest);
      cohort.items.push(item);
      cohort.cost += item.cost;
    }
    if (cohorts.every((cohort) => cohort.items.length === 1 || cohort.cost <= timeoutMS)) return cohorts;
  }
  throw new Error("balanced browser packing could not place the selection");
}

export function packBrowserCohorts(partitions, catalog, policy, owner, timeoutFor, { laneLimit = 1 } = {}) {
  return partitions.flatMap((partition) => {
    const items = partition.selectedRowIDs.map((id) => {
      const row = catalog.rowByID.get(id);
      const cost = owner.evidence_estimates_ms[row?.evidence_class];
      if (!row || !Number.isSafeInteger(cost) || cost < 1) throw new Error(`browser row ${id} has no selected execution estimate`);
      return { id, cost: cost * row.selector.titles.length, row };
    });
    // Placement is known only after packing. Reserve both allocation boundaries
    // for every candidate; the compiler charges them only at their actual ends.
    const headroom = Math.max(owner.lifecycle_estimates_ms.browser_allocation_acquire, owner.lifecycle_estimates_ms.browser_reset) +
      owner.lifecycle_estimates_ms.browser_allocation_release;
    const envelope = timeoutFor(partition) - headroom;
    const packed = partition.resourceProfileID === "browser_ordinary"
      ? packBalancedBrowserCohorts(items, policy, envelope, {
        laneLimit, minimumSelectedWorkMs: owner.lifecycle_estimates_ms.browser_allocation_acquire,
      })
      : packInvocationCohorts(items, policy, envelope);
    return packed.map((cohort) => ({
      ...browserCohort(partition, cohort.items.map((item) => item.row), policy),
      estimatedWorkMs: cohort.cost,
    }));
  });
}
