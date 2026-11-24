import path from "node:path";
import { parseStrictJSON, redactionMarker, resolveCommandInputs } from "../contract/index.mjs";
import { loadExecutionTopology } from "../generated-artifacts/execution-topology.mjs";
import { planCommand, planInvocation } from "../runtime/harness-binary.mjs";
import { loadTestCatalog } from "../test-catalog/index.mjs";
import { readLocalFile } from "../runtime/secure-local-files.mjs";

export function readPlanInputs(root, filename) {
  if (!filename) return {};
  if (path.extname(filename).toLowerCase() !== ".json") throw new Error("INPUTS must name a JSON file");
  const value = parseStrictJSON(readLocalFile(path.resolve(root, filename), { privateFile: false }).toString("utf8"), "target-plan INPUTS");
  if (!value || typeof value !== "object" || Array.isArray(value)) throw new Error("INPUTS must contain an object of declared command inputs");
  return value;
}

export function buildTargetPlan(root, target, suppliedInputs = {}, {
  plan = planInvocation, surface = loadExecutionTopology({ root }).taskSurface, catalog = loadTestCatalog(root),
} = {}) {
  const entry = surface.targets.find((candidate) => candidate.name === target && candidate.target_class === "public");
  if (!entry) throw new Error(`unknown public target ${target}`);
  const inputs = resolveCommandInputs(target, suppliedInputs, surface);
  const planned = planCommand({ root, surface, target, inputs: suppliedInputs, plan });
  const { units } = planned.graph;
  const rowIDs = [...new Set(units.flatMap((unit) => unit.row_bindings.map((binding) => binding.row_id)))].sort();
  return {
    schema_id: "cartulary.harness_target_plan.v7",
    target, command_id: entry.command_id,
    inputs: Object.fromEntries(Object.entries(inputs).map(([key, value]) => [key,
      entry.input_contract.inputs.find((input) => input.name === key).summary_emission === "redacted_value" ? redactionMarker() : value])),
    capability_snapshot: planned.manifest.capability_snapshot, go_worker_budget: planned.manifest.go_worker_budget,
    graph_digest: planned.graph.graph_digest,
    projections: Object.fromEntries(Object.entries(planned.projections).map(([name, projection]) => [name, projection.unit_ids])),
    rows: rowIDs.map((id) => {
      const row = catalog.rowByID.get(id);
      return { row_id: id, owner_id: row.owner_id, runner: row.runner, evidence_class: row.evidence_class,
        aggregate_placement: row.aggregate_placement, fixture_capability: row.fixture_capability, service_dependencies: row.service_dependencies };
    }),
    units: units.map((unit) => ({
      unit_id: unit.unit_id, owner_id: unit.owner_id, kind: unit.kind, needs: unit.needs,
      resource_claims: unit.resource_claims, fixture_capability: unit.fixture_lease,
      ...(unit.fixture_profile_id ? { fixture_profile_id: unit.fixture_profile_id, snapshot_key: unit.snapshot_key } : {}),
      service_dependencies: unit.service_dependencies, cache_policy: unit.cache_policy,
      estimated_work_ms: unit.estimated_work_ms, semantic_digest: unit.semantic_digest, evidence_outputs: unit.current_run_evidence_outputs,
    })),
  };
}

export function renderTargetPlan(plan, detail = "summary") {
  if (detail === "summary") {
    const cached = plan.units.filter((unit) => unit.cache_policy !== "none").length;
    const fixtures = plan.units.filter((unit) => unit.fixture_capability !== "none").length;
    return `target=${plan.target} units=${plan.units.length} cached=${cached} fixtures=${fixtures} digest=${plan.graph_digest}`;
  }
  const lines = [`target ${plan.target}`, `graph ${plan.graph_digest}`,
    `capacity ${plan.capability_snapshot.snapshot_digest} cpu=${plan.capability_snapshot.cpu_tokens} go_workers=${plan.go_worker_budget} postgres=${plan.capability_snapshot.postgres_lanes}`];
  if (detail === "rows") lines.push(...plan.rows.map((row) => `${row.row_id} owner=${row.owner_id} runner=${row.runner} evidence=${row.evidence_class} placement=${row.aggregate_placement} fixture=${row.fixture_capability}`));
  else if (detail === "artifacts") lines.push(...plan.units.flatMap((unit) => unit.evidence_outputs.map((output) => `${unit.unit_id} ${output}`)));
  else if (detail === "graph") lines.push(...plan.units.map((unit) => `${unit.unit_id} kind=${unit.kind} owner=${unit.owner_id} needs=${unit.needs.join(",") || "-"} fixture=${unit.fixture_capability} services=${unit.service_dependencies.join(",") || "-"} cache=${unit.cache_policy} work_ms=${unit.estimated_work_ms}`));
  else throw new Error("DETAIL must be summary, graph, rows, or artifacts");
  return lines.join("\n");
}
