import { canonicalJSONString } from "../contract/index.mjs";
import { WorkGraphCompiler } from "../scheduler/work-graph/index.mjs";

// Planning only: fixed inputs make structural measurements independent of the
// invoking machine's available capacity. No compiled work is executed.
export const schedulerGrowthProfile = Object.freeze({
  target: "release-check",
  go_workers: 4,
  postgres_lanes: 4,
  browser_lanes: 4,
});

// Each unit has a fixed contract allowance. Packed row identities/evidence and
// dependency edges have explicit linear allowances; unrelated payload growth
// cannot hide behind a smaller graph or another kind's smaller units.
export const schedulerGrowthPolicy = Object.freeze({
  bytes_per_row_binding: 384,
  unit_budget_by_kind: Object.freeze({
    artifact: Object.freeze({ base_bytes: 4096, bytes_per_dependency: 192 }),
    // Browser finalizers carry exact group/target/unit mappings as well as edges.
    finalizer: Object.freeze({ base_bytes: 4096, bytes_per_dependency: 384 }),
    fixture_builder: Object.freeze({ base_bytes: 4096, bytes_per_dependency: 192 }),
    lifecycle: Object.freeze({ base_bytes: 4096, bytes_per_dependency: 192 }),
    policy: Object.freeze({ base_bytes: 4096, bytes_per_dependency: 192 }),
    readiness: Object.freeze({ base_bytes: 4096, bytes_per_dependency: 192 }),
    runner: Object.freeze({ base_bytes: 4096, bytes_per_dependency: 192 }),
  }),
});

function encoded(value) {
  return `${canonicalJSONString(value)}\n`;
}

function statistics(sizes) {
  const sorted = [...sizes].sort((left, right) => left - right);
  const total = sorted.reduce((sum, value) => sum + value, 0);
  return {
    work_unit_count: sorted.length,
    serialized_unit_bytes: total,
    bytes_per_unit: Math.ceil(total / sorted.length),
    p95_serialized_unit_bytes: sorted[Math.ceil(sorted.length * 0.95) - 1],
    max_serialized_unit_bytes: sorted.at(-1),
  };
}

export function measureSchedulerGrowth(first, second, policy = schedulerGrowthPolicy) {
  if (!Number.isSafeInteger(policy.bytes_per_row_binding) || policy.bytes_per_row_binding < 0) {
    throw new Error("invalid scheduler growth row budget");
  }
  for (const graph of [first, second]) {
    if (!Array.isArray(graph?.units) || graph.units.length === 0) {
      throw new Error("scheduler growth requires a nonempty compiled graph");
    }
    const identities = new Set();
    for (const unit of graph.units) {
      if (typeof unit.unit_id !== "string" || !unit.unit_id || identities.has(unit.unit_id)) {
        throw new Error("scheduler growth requires unique unit identities");
      }
      identities.add(unit.unit_id);
      if (!Object.hasOwn(policy.unit_budget_by_kind, unit.kind)) {
        throw new Error(`scheduler growth has no policy for unit kind ${unit.kind}`);
      }
      const budget = policy.unit_budget_by_kind[unit.kind];
      if (!Number.isSafeInteger(budget.base_bytes) || budget.base_bytes < 1 ||
          !Number.isSafeInteger(budget.bytes_per_dependency) || budget.bytes_per_dependency < 0) {
        throw new Error(`invalid scheduler growth budget for ${unit.kind}`);
      }
      if (!Array.isArray(unit.row_bindings) || !Array.isArray(unit.needs)) {
        throw new Error("scheduler growth requires explicit row bindings and dependencies");
      }
    }
  }
  const firstBytes = encoded(first);
  const secondBytes = encoded(second);
  const errors = [];
  const byKind = new Map();
  let maximum = 0;
  for (const unit of first.units) {
    const bytes = Buffer.byteLength(canonicalJSONString(unit), "utf8");
    const sizes = byKind.get(unit.kind) ?? [];
    sizes.push(bytes);
    byKind.set(unit.kind, sizes);
    maximum = Math.max(maximum, bytes);
    const budget = policy.unit_budget_by_kind[unit.kind];
    const limit = budget.base_bytes +
      unit.row_bindings.length * policy.bytes_per_row_binding +
      unit.needs.length * budget.bytes_per_dependency;
    if (bytes > limit) {
      errors.push(`scheduler ${unit.kind} unit ${unit.unit_id} uses ${bytes} bytes; structural budget is ${limit}`);
    }
  }
  const identical = firstBytes === secondBytes;
  if (!identical) errors.push("two fresh scheduler graph compiles are not byte-identical");
  const graphBytes = Buffer.byteLength(firstBytes, "utf8");
  return {
    graph_bytes: graphBytes,
    work_unit_count: first.units.length,
    graph_bytes_per_work_unit: Math.ceil(graphBytes / first.units.length),
    max_serialized_work_unit_bytes: maximum,
    per_kind: Object.fromEntries([...byKind].sort(([left], [right]) => left < right ? -1 : 1)
      .map(([kind, sizes]) => [kind, statistics(sizes)])),
    scratch_compiles_byte_identical: identical,
    errors,
  };
}

export function collectSchedulerGrowth(root) {
  const compile = () => {
    const compiler = new WorkGraphCompiler(root);
    compiler.availableGoLanes = schedulerGrowthProfile.go_workers;
    compiler.availablePostgresLanes = schedulerGrowthProfile.postgres_lanes;
    compiler.browserLaneLimit = schedulerGrowthProfile.browser_lanes;
    return compiler.compileTarget(schedulerGrowthProfile.target);
  };
  return {
    planning_profile: schedulerGrowthProfile,
    unit_budget_policy: schedulerGrowthPolicy,
    ...measureSchedulerGrowth(compile(), compile()),
  };
}
