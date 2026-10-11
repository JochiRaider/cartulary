import path from "node:path";
import { readFileSync } from "node:fs";
import { createHash } from "node:crypto";
import { workspaceLayout } from "../../../workspace_layout.generated.mjs";
import { resolveGraphNodeBinary, assertGraphNodeLaunch } from "./executor.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
function usage() {
  return "usage: runner-cli.mjs --selection target|aggregate|owner|rows --target <target> [--owner <owner-id>] [--rows <row-id,...>] [--service-backed-only]";
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
    else if (arg === "--rows") options.rows = (argv[++index] ?? "").split(",").filter(Boolean);
    else throw new Error(usage());
  }
  if (!new Set(["target", "aggregate", "owner", "rows"]).has(options.selection)) {
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

function readJSON(relative) {
  return JSON.parse(readFileSync(path.join(root, relative), "utf8"));
}

export function resolvedRunRoot() {
  const resultsDir = process.env.CARTULARY_TEST_RESULTS_DIR || ".cartulary/test-results";
  const runID =
    process.env.CARTULARY_TEST_RUN_ID ||
    `${new Date().toISOString().replaceAll(/[-:.]/gu, "").replace("Z", "Z")}-p${process.pid}`;
  if (!/^[A-Za-z0-9_.-]+$/u.test(runID)) throw new Error("invalid graph run ID");
  return { resultsDir, runID, runRoot: path.resolve(root, resultsDir, runID) };
}

export function commandEntry(target) {
  const taskSurface = readJSON("tools/task_surface_owner.json");
  const entry = taskSurface.targets.find((candidate) => candidate.name === target);
  if (!entry?.command_id) throw new Error(`public graph target ${target} has no command ID`);
  return entry;
}

export function declaredInputs(entry) {
  return Object.fromEntries(
    (entry.input_contract?.inputs ?? [])
      .filter((input) => process.env[input.name] !== undefined && process.env[input.name] !== "")
      .map((input) => [
        input.name,
        input.summary_emission === "redacted_value"
          ? "<redacted>"
          : String(process.env[input.name]),
      ])
      .sort(([left], [right]) => left.localeCompare(right)),
  );
}

export function selectionInputs(options) {
  return {
    ...(options.owner ? { OWNER: options.owner } : {}),
    ...(options.rows ? { ROWS: options.rows.join(",") } : {}),
    ...(options.serviceBackedOnly ? { SERVICE_BACKED_ONLY: "1" } : {}),
  };
}

export function fixtureSelectionEnvironment(options) {
  const values = selectionInputs(options);
  const publicSelectionNames = ["OWNER", "ROWS"].filter(
    (name) => values[name] !== undefined,
  );
  const retainedSources = String(
    process.env.CARTULARY_MAKE_INPUT_SOURCES ?? "",
  )
    .split(/\s+/u)
    .filter(Boolean)
    .filter(
      (token) =>
        !publicSelectionNames.some((name) => token.startsWith(`${name}=`)),
    );
  return {
    ...values,
    CARTULARY_MAKE_INPUT_SOURCES: [
      ...retainedSources,
      ...publicSelectionNames.map((name) => `${name}=cli`),
    ].join(" "),
  };
}

export function graphChildEnvironment(options) {
  const environment = { ...process.env };
  for (const name of [
    "CARTULARY_TEST_SERVICES_MODE",
    "CARTULARY_TEST_SERVICES_PERSISTENT_BORROWER",
    "CARTULARY_TEST_SERVICES_SESSION_FILE",
    "HARNESS_DIAGNOSTICS",
    "MAKEFLAGS",
    "MAKEOVERRIDES",
    "MFLAGS",
  ]) {
    delete environment[name];
  }
  if (new Set(["test-slice", "service-backed-test-slice"]).has(options.target)) {
    for (const name of ["JSON", "OWNER", "ROWS", "SERVICE_BACKED_ONLY"]) {
      delete environment[name];
    }
  }
  return environment;
}

export function resolvedRuntimeEnvironment(compiler) {
  const environment = Object.fromEntries(
    compiler.topology.runtime_binaries
      .map((binary) => [
        binary.consumer_env,
        path.resolve(
          root,
          process.env[binary.output_make_variable] || binary.default_output_path,
        ),
      ])
      .sort(([left], [right]) => left.localeCompare(right)),
  );
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
    rowIDs = (rowIDs ?? ownerRows.map((row) => row.row_id)).filter(
      (rowID) => (serviceDependencies.get(rowID)?.length ?? 0) > 0,
    );
    if (rowIDs.length === 0) throw new Error(`${options.owner} has no service-backed rows`);
  }
  const graph = compiler.compile({ kind: "owner", owner_id: options.owner, row_ids: rowIDs });
  return { graph, projections: { [options.target]: graph.units.map((unit) => unit.unit_id) } };
}
