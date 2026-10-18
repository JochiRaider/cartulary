import { loadExecutionTopology } from "../../generated-artifacts/execution-topology.mjs";
import path from "node:path";
import { createHash } from "node:crypto";
import { resolveDeclaredTargetInputs } from "../../contract/index.mjs";
import { workspaceLayout } from "../../../workspace_layout.generated.mjs";
import { resolveGraphNodeBinary, assertGraphNodeLaunch } from "./launch-binding.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
function usage() {
  return "usage: cartulary-harness graph --selection command|target|aggregate|owner|rows --target <target> [--owner <owner-id>] [--rows <row-id,...>] [--service-backed-only]";
}

export function parseArgs(argv) {
  const options = {
    selection: "",
    target: "",
    owner: "",
    rows: undefined,
    serviceBackedOnly: false,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--service-backed-only") options.serviceBackedOnly = true;
    else if (arg === "--selection") options.selection = argv[++index] ?? "";
    else if (arg === "--target") options.target = argv[++index] ?? "";
    else if (arg === "--owner") options.owner = argv[++index] ?? "";
    else if (arg === "--rows") options.rows = (argv[++index] ?? "").split(",");
    else throw new Error(usage());
  }
  if (!new Set(["command", "target", "aggregate", "owner", "rows"]).has(options.selection)) {
    throw new Error(usage());
  }
  if (!options.target) throw new Error(usage());
  if (options.selection === "owner" && !options.owner) throw new Error(usage());
  if (options.selection === "rows" && (!options.rows || options.rows.length === 0)) {
    throw new Error(usage());
  }
  if (options.serviceBackedOnly && options.selection !== "owner") throw new Error(usage());
  return options;
}

export function sha256(bytes) {
  return `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
}

export function resolvedRunRoot() {
  const resultsDir = process.env.CARTULARY_TEST_RESULTS_DIR || ".cartulary/test-results";
  const runID =
    process.env.CARTULARY_TEST_RUN_ID ||
    `${new Date().toISOString().replaceAll(/[-:.]/gu, "").replace("Z", "Z")}-p${process.pid}`;
  if (!/^[A-Za-z0-9_.-]+$/u.test(runID)) throw new Error("invalid graph run ID");
  return { resultsDir, runID, runRoot: path.resolve(root, resultsDir, runID) };
}

export function commandEntry(target, taskSurface = loadExecutionTopology({ root }).taskSurface) {
  const entry = taskSurface.targets.find((candidate) => candidate.name === target);
  if (!entry?.command_id) throw new Error(`public graph target ${target} has no command ID`);
  return entry;
}

export function commandInputContext(entry, selected = {}, { environment: parent = process.env, surface = loadExecutionTopology({ root }).taskSurface } = {}) {
  // Native graph selectors are admitted arguments, independent of an enclosing
  // Make invocation's artifact identity. Resolve only this command's inputs.
  const sources = (parent.CARTULARY_MAKE_INPUT_SOURCES || "").split(/\s+/u)
    .filter((token) => token && !Object.hasOwn(selected, token.split("=")[0]));
  const environment = { ...parent, ...selected, CARTULARY_MAKE_INPUT_SOURCES:
    [...sources, ...Object.keys(selected).map((name) => `${name}=cli`)].join(" ") };
  const resolved = resolveDeclaredTargetInputs(entry.name, entry, surface, environment);
  const inputs = Object.fromEntries(Object.entries(resolved).sort(([a], [b]) => a.localeCompare(b))
    .map(([name, input]) => [name, input.summary_emission === "redacted_value" ? "<redacted>" : input.value]));
  // Public controls are consumed by selection or explicit private arguments.
  // Clear the complete authored input set, including inputs of other commands,
  // then forward only this command's admitted runtime values. The native owner
  // applies this projection without loading a second command registry.
  const clear = [...new Set(["CARTULARY_MAKE_INPUT_SOURCES", ...surface.targets
    .filter((target) => target.target_class === "public")
    .flatMap((target) => target.input_contract?.inputs ?? []).map((input) => input.name)])].sort();
  const forward = {};
  for (const input of entry.input_contract?.inputs ?? []) {
    if (!["runtime_env", "argv_and_runtime_env"].includes(input.child_forwarding) || !Object.hasOwn(resolved, input.name)) continue;
    if (input.summary_emission === "redacted_value") throw new Error("credential inputs require an owned private runtime binding");
    const value = resolved[input.name].value;
    forward[input.name] = input.type === "exact_1_bool" ? value === "1" || value === "true" ? "1" : "" : value;
  }
  return { inputs, input_environment: { clear, forward } };
}

// Defaults are already canonical in retained manifests. Do not reinterpret
// normalized false flags as user-supplied command-line tokens during replay.
export function replayInputs(entry, inputs) {
  const definitions = new Map((entry.input_contract?.inputs ?? []).map((input) => [input.name, input]));
  return Object.fromEntries(Object.entries(inputs).filter(([name, value]) => {
    const input = definitions.get(name);
    return !input || input.default === null || input.default === undefined || String(input.default) !== value;
  }));
}

export function selectionInputs(options) {
  // Private exact-row subgraphs retain their selection in the manifest; they do
  // not manufacture public forwarding inputs for their enclosing command.
  if (options.selection === "rows") return {};
  return {
    ...(options.owner ? { OWNER: options.owner } : {}),
    ...(options.rows ? { ROWS: [...options.rows].sort().join(",") } : {}),
    ...(options.serviceBackedOnly ? { SCOPE: "services" } : {}),
  };
}

export function resolvedRuntimeEnvironment(compiler) {
  const environment = Object.fromEntries(compiler.topology.runtime_binaries.flatMap((binary) => {
    const output = path.resolve(root, process.env[binary.output_make_variable] || binary.default_output_path);
    return [[binary.consumer_env, output], [binary.output_make_variable, output]];
  }));
  environment.CARTULARY_TEST_SERVICES_BIN = path.resolve(
    root,
    process.env.CARTULARY_TEST_SERVICES_BIN ||
      process.env.TEST_SERVICES_BIN ||
      `${workspaceLayout.toolbin}/cartulary-test-services`,
  );
  environment.CARTULARY_HARNESS_RUNTIME_BIN = path.resolve(root,process.env.CARTULARY_HARNESS_RUNTIME_BIN || process.env.HARNESS_RUNTIME_BIN || `${workspaceLayout.toolbin}/cartulary-harness`);
  environment.NODE_RUNTIME_DIR = path.resolve(
    root,
    process.env.NODE_RUNTIME_DIR || workspaceLayout.node_runtime,
  );
  environment.NODE_BIN = resolveGraphNodeBinary({
    cwd: root,
    nodeBin: process.env.NODE_BIN === undefined
      ? path.join(environment.NODE_RUNTIME_DIR, "bin/node")
      : process.env.NODE_BIN,
  });
  assertGraphNodeLaunch(environment.NODE_BIN);
  environment.PNPM = process.env.PNPM || path.join(environment.NODE_RUNTIME_DIR, "bin/pnpm");
  return environment;
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
