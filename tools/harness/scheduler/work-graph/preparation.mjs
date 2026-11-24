// Defaults are already canonical in retained manifests. Do not reinterpret
// normalized false flags as user-supplied command-line tokens during replay.
export function replayInputs(entry, inputs) {
  const definitions = new Map((entry.input_contract?.inputs ?? []).map((input) => [input.name, input]));
  return Object.fromEntries(Object.entries(inputs).filter(([name, value]) => {
    const input = definitions.get(name);
    return !input || input.default === null || input.default === undefined || String(input.default) !== value;
  }));
}

export function selectionFor(options, compiler) {
  if (options.selection === "command") {
    const compiled = compiler.compileCommand(options.target, compiler.targetInputs[options.target] ?? {});
    return { graph: compiled.graph, projections: compiler.aggregateProjections(options.target, compiled.graph) };
  }
  if (options.selection === "aggregate") {
    const plan = compiler.compileAggregatePlan(options.target);
    return { graph: plan.graph, projections: plan.projections };
  }
  if (options.selection === "target") {
    const graph = compiler.compile({ kind: "target", target: options.target });
    return { graph, projections: { [options.target]: graph.units.map((unit) => unit.unit_id) } };
  }
  if (options.selection === "rows") {
    const graph = compiler.compile({ kind: "rows", row_ids: options.rows });
    return { graph, projections: { [options.target]: graph.units.map((unit) => unit.unit_id) } };
  }
  let rowIDs = options.rows;
  if (options.serviceBackedOnly) {
    const serviceDependencies = new Map(
      compiler.catalog.rows.map((row) => [row.row_id, row.service_dependencies]),
    );
    const ownerRows = compiler.catalog.rows.filter((row) => row.owner_id === options.owner);
    if (rowIDs) {
      if (new Set(rowIDs).size !== rowIDs.length) throw new Error("row selection contains duplicate values");
      for (const rowID of rowIDs) {
        if (!ownerRows.some((row) => row.row_id === rowID)) throw new Error(`${rowID} does not belong to ${options.owner}`);
        if (serviceDependencies.get(rowID).length === 0) throw new Error(`${rowID} is outside service-backed scope`);
      }
    } else rowIDs = ownerRows.filter((row) => row.service_dependencies.length > 0).map((row) => row.row_id);
    if (rowIDs.length === 0) throw new Error(`${options.owner} has no service-backed rows`);
  }
  const graph = compiler.compile({ kind: "owner", owner_id: options.owner, row_ids: rowIDs });
  return { graph, projections: { [options.target]: graph.units.map((unit) => unit.unit_id) } };
}
