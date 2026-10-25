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

export function packBrowserCohorts(partitions, catalog, policy, owner, timeoutFor) {
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
    return packInvocationCohorts(items, policy, timeoutFor(partition) - headroom).map((packed) => ({
      ...browserCohort(partition, packed.items.map((item) => item.row), policy),
      estimatedWorkMs: packed.cost,
    }));
  });
}
