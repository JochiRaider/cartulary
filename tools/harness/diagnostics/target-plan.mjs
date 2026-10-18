import path from "node:path";
import { parseStrictJSON, resolveCommandInputs } from "../contract/index.mjs";
import { WorkGraphCompiler, buildWorkGraph, selectionFor } from "../scheduler/work-graph/index.mjs";
import { readLocalFile } from "../runtime/secure-local-files.mjs";

export function readPlanInputs(root, filename) {
  if (!filename) return {};
  if (path.extname(filename).toLowerCase() !== ".json") throw new Error("INPUTS must name a JSON file");
  const value = parseStrictJSON(readLocalFile(path.resolve(root, filename), { privateFile: false }).toString("utf8"), "target-plan INPUTS");
  if (!value || typeof value !== "object" || Array.isArray(value)) throw new Error("INPUTS must contain an object of declared command inputs");
  return value;
}

// Compilation is shared with execution. This function never invokes preparation,
// starts a process, installs tools, or samples historical evidence.
export function buildTargetPlan(root, target, suppliedInputs = {}, compiler = new WorkGraphCompiler(root)) {
  const entry = compiler.taskSurface.targets.find((candidate) => candidate.name === target && candidate.target_class === "public");
  if (!entry) throw new Error(`unknown public target ${target}`);
  const inputs = resolveCommandInputs(target, suppliedInputs, compiler.taskSurface);
  compiler.targetInputs = { ...compiler.targetInputs, [target]: inputs };
  const recipe = compiler.taskSurface.make_recipes[target];
  let compiled;
  if (entry.selection_binding) compiled = compiler.compileCommand(target, suppliedInputs);
  else if (Object.hasOwn(compiler.owner.aggregate_tiers, target)) compiled = compiler.compileAggregatePlan(target);
  else if (recipe?.type === "work_graph") compiled = selectionFor({
    selection: recipe.selection, target, owner: inputs.OWNER,
    rows: inputs.ROWS?.split(","), serviceBackedOnly: recipe.service_backed_only === true,
  }, compiler);
  else if (recipe?.type === "print_help") compiled = { graph: buildWorkGraph([]) };
  else compiled = { graph: compiler.compile({ kind: "target", target }) };
  const rowIDs = [...new Set(compiled.graph.units.flatMap((unit) => unit.current_run_evidence_outputs)
    .filter((output) => output.startsWith("rows/") && output.endsWith(".json"))
    .map((output) => output.slice(5, -5)))].sort();
  return {
    schema_id: "cartulary.harness_target_plan.v6",
    target, command_id: entry.command_id,
    inputs: Object.fromEntries(Object.entries(inputs).map(([key, value]) => [key,
      entry.input_contract.inputs.find((input) => input.name === key).summary_emission === "redacted_value" ? "<redacted>" : value])),
    graph_digest: compiled.graph.graph_digest,
    projections: compiled.projections ?? { [target]: compiled.graph.units.map((unit) => unit.unit_id) },
    rows: rowIDs.map((id) => {
      const row = compiler.catalog.rowByID.get(id);
      return { row_id: id, owner_id: row.owner_id, runner: row.runner, evidence_class: row.evidence_class,
        aggregate_placement: row.aggregate_placement, fixture_capability: row.fixture_capability, service_dependencies: row.service_dependencies };
    }),
    units: compiled.graph.units.map((unit) => ({
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
  const lines = [`target ${plan.target}`, `graph ${plan.graph_digest}`];
  if (detail === "rows") lines.push(...plan.rows.map((row) => `${row.row_id} owner=${row.owner_id} runner=${row.runner} evidence=${row.evidence_class} placement=${row.aggregate_placement} fixture=${row.fixture_capability}`));
  else if (detail === "artifacts") lines.push(...plan.units.flatMap((unit) => unit.evidence_outputs.map((output) => `${unit.unit_id} ${output}`)));
  else if (detail === "graph") lines.push(...plan.units.map((unit) => `${unit.unit_id} kind=${unit.kind} owner=${unit.owner_id} needs=${unit.needs.join(",") || "-"} fixture=${unit.fixture_capability} services=${unit.service_dependencies.join(",") || "-"} cache=${unit.cache_policy} work_ms=${unit.estimated_work_ms}`));
  else throw new Error("DETAIL must be summary, graph, rows, or artifacts");
  return lines.join("\n");
}
