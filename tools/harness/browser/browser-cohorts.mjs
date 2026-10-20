import { semanticJSONDigest } from "../contract/index.mjs";
import { packInvocationCohorts } from "../scheduler/work-graph/cohorts.mjs";

// The generated group is a compatibility partition. Only selected catalog rows
// become executable cohorts, so selection never inherits an obsolete file split.
export function browserCohort(partition, rows, policy) {
  const ordered = [...rows].sort((a, b) => a.row_id < b.row_id ? -1 : a.row_id > b.row_id ? 1 : 0);
  const available = new Set(partition.selectedRowIDs);
  if (ordered.length === 0 || new Set(ordered.map((row) => row.row_id)).size !== ordered.length ||
      (partition.stateScope !== "test_owned" && ordered.length !== 1) ||
      ordered.some((row) => !available.has(row.row_id) || row.runner !== "playwright" ||
        row.owner_id !== partition.ownerID || row.browser_state_scope !== partition.stateScope ||
        row.selector.project_id !== partition.project || row.runtime_profile_id !== partition.runtimeProfileID ||
        row.resource_profile_id !== partition.resourceProfileID ||
        (row.fixture_profile_id ?? "") !== partition.fixtureProfileID ||
        JSON.stringify(row.service_dependencies) !== JSON.stringify(partition.serviceDependencies) ||
        JSON.stringify(row.preparation_operations ?? []) !== JSON.stringify(partition.preparationOperations ?? []) ||
        !partition.specs.includes(row.selector.file))) {
    throw new Error(`browser cohort violates compatibility partition ${partition.name}`);
  }
  const { selectedRowIDs: _available, specs: _specs, ...compatibility } = partition;
  const digest = semanticJSONDigest({
    compatibility, policy, workers: 1, max_failures: 1,
    selections: ordered.map((row) => ({ row_id: row.row_id, selector: row.selector })),
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
    const headroom = owner.lifecycle_estimates_ms.browser_allocation_acquire +
      owner.lifecycle_estimates_ms.browser_allocation_release;
    return packInvocationCohorts(items, policy, timeoutFor(partition) - headroom).map((packed) => ({
      ...browserCohort(partition, packed.items.map((item) => item.row), policy),
      estimatedWorkMs: packed.cost,
    }));
  });
}
