import { HarnessConfigError, resolveCommandInputs } from "../../contract/index.mjs";

import { attachedSelectionAllowed } from "../../services/local-session.mjs";

const compare = (left, right) => left < right ? -1 : left > right ? 1 : 0;
const usage = (message) => new HarnessConfigError(message, { reason: "usage_error" });

export function admitCommandSelection(compiler, target, validatedInputs) {
  // Source restrictions were already enforced by the public input resolver.
  // Reapply defaults from the command owner; defaults are not explicit CLI data.
  return resolveCommandSelection(compiler, target, Object.fromEntries(Object.entries(validatedInputs)
    .filter(([, input]) => input.source !== "default").map(([name, input]) => [name, input.value])));
}

function catalogRows(compiler, query) {
  return compiler.catalog.rows.filter((row) => {
    if (row.runner !== query.runner) return false;
    if (query.stages && !query.stages.includes(row.selector.stage)) return false;
    if (query.go_suite) {
      const context = compiler.rowTargets.get(row.row_id);
      return context === `backend-${query.go_suite}` || query.go_suite === "integration" && context === "backend-integration-support";
    }
    return true;
  }).map((row) => row.row_id);
}

// This is selection expansion inside the existing compiler, not execution or a
// second registry. Private unions remain owned by work-graph.target_members.
export function resolvePrivateSelection(compiler, selection, visiting = new Set()) {
  if (typeof selection === "object" && selection !== null) {
    if (selection.row_id) {
      if (!compiler.catalog.rowByID.has(selection.row_id)) throw new Error(`unknown active row ${selection.row_id}`);
      return { rows: [selection.row_id], operations: [] };
    }
    if (selection.catalog) {
      const rows = catalogRows(compiler, selection.catalog);
      if (rows.length === 0) throw new Error("catalog selection has no active rows");
      return { rows, operations: [] };
    }
    if (selection.operation_id) return resolvePrivateSelection(compiler, selection.operation_id, visiting);
    throw new Error("unknown private selection binding");
  }
  if (typeof selection !== "string" || !selection.startsWith("harness.")) throw new Error("private selection requires a canonical identity");
  if (visiting.has(selection)) throw new Error(`private selection cycle at ${selection}`);
  if (Object.hasOwn(compiler.owner.operations, selection)) return { rows: [], operations: [selection] };
  const members = compiler.owner.target_members[selection];
  if (!members) throw new Error(`unknown private selection ${selection}`);
  const next = new Set(visiting).add(selection);
  const resolved = members.map((member) => resolvePrivateSelection(compiler, member, next));
  return { rows: [...new Set(resolved.flatMap((value) => value.rows))].sort(compare),
    operations: [...new Set(resolved.flatMap((value) => value.operations))].sort(compare) };
}

export function resolveCommandSelection(compiler, target, suppliedInputs) {
  const inputs = resolveCommandInputs(target, suppliedInputs, compiler.taskSurface);
  if (inputs.CARTULARY_TEST_SERVICES_MODE === "attach" && !attachedSelectionAllowed(target, inputs)) throw usage("selected command variant is owned-only and rejects local service attachment");
  if (inputs.CARTULARY_TEST_SERVICES_SESSION_FILE && inputs.CARTULARY_TEST_SERVICES_MODE !== "attach") throw usage("CARTULARY_TEST_SERVICES_SESSION_FILE is accepted only in attach mode");
  const entry = compiler.taskSurface.targets.find((candidate) => candidate.name === target);
  const binding = entry.selection_binding;
  if (!binding) throw new Error(`${target} has no command selection binding`);
  let selection;
  if (binding.kind === "owner_or_suite" && inputs.OWNER) {
    const ownerRows = compiler.catalog.rows.filter((row) => row.owner_id === inputs.OWNER).map((row) => row.row_id);
    if (ownerRows.length === 0) throw usage(`unknown active owner ${inputs.OWNER}`);
    const requested = inputs.ROWS ? inputs.ROWS.split(",") : ownerRows;
    for (const row of requested) if (!ownerRows.includes(row)) throw usage(`${row} does not belong to ${inputs.OWNER}`);
    selection = { rows: requested, operations: [] };
  } else {
    const key = inputs[binding.input];
    const name = binding.kind === "fixed" ? binding.selection : key === undefined ? binding.default_selection : binding.variants[key];
    if (!name) throw new Error(`${target} has no selection for ${binding.input}=${key ?? "<omitted>"}`);
    selection = resolvePrivateSelection(compiler, name);
  }
  if (inputs.SCOPE === "services") {
    const serviceRows = selection.rows.filter((id) => compiler.catalog.rowByID.get(id).service_dependencies.length > 0);
    if (inputs.ROWS && serviceRows.length !== selection.rows.length) throw usage("requested row is outside service scope");
    if (serviceRows.length === 0 || selection.operations.length) throw usage("selection has no service-only row scope");
    selection = { ...selection, rows: serviceRows };
  }
  return { inputs, rows: selection.rows.sort(compare), operations: selection.operations };
}
