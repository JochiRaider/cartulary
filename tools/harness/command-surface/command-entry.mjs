#!/usr/bin/env node
// The harness command entry (P5-3e). A public graph target's recipe runs it
// once, after the Node runtime: Make passes only what Make alone knows, each
// public input's value and origin (CARTULARY_MAKE_INPUT_SOURCES), and the
// entry admits the inputs, produces the target's producers and launches its
// graph with the environment the recipe used to assemble in Make.
import { spawn, spawnSync } from "node:child_process";
import { mkdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { HarnessConfigError, preflightPublicTarget } from "../contract/index.mjs";
import { inputSourcesEnvironment } from "../contract/run-context.mjs";
import { producerBindings, producerCommand, readProducerRegistry, splitProducedPrerequisites } from "../execution/producers.mjs";
import { canonicalInputPaths, defaultTaskSurfaceManifestPath, graphForwardedInputNames, machineStateNames, publicMakeInputNames } from "../generated-artifacts/index.mjs";

const repositoryRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");

// The bindings a graph launch keeps after the public inputs are stripped.
const launchBindings = ["MAKE", "NODE_BIN", "TEST_SERVICES_BIN"];

// A graph-entry operation target needs what every graph launch needs.
const graphEntryPrerequisites = Object.freeze(["$(NODE_BIN)", "$(FRONTEND_INSTALL_RECEIPT)", "$(HARNESS_RUNTIME_BIN)"]);

function parseSources(text = "") {
  return new Map(text.split(/\s+/u).filter(Boolean).map((entry) => {
    const index = entry.indexOf("=");
    return [entry.slice(0, index), entry.slice(index + 1)];
  }));
}

function graphRecipe(manifest, target) {
  const recipe = manifest.make_recipes[target];
  const entry = manifest.targets.find((candidate) => candidate.name === target);
  if (!recipe || entry?.target_class !== "public") throw new HarnessConfigError(`${target} is not a public target`, { reason: "usage_error" });
  // A graph-entry operation target enters its graph like any graph target.
  if (recipe.graph_entry === true) return { entry, recipe: { type: "work_graph", selection: "target", prerequisites: graphEntryPrerequisites, preflight_owner: recipe.preflight_owner } };
  if (recipe.type !== "work_graph") throw new HarnessConfigError(`${target} does not launch a graph`, { reason: "usage_error" });
  return { entry, recipe };
}

// graphProducers are the producers a graph target needs besides the Node
// runtime, which its recipe produces before the entry runs.
export function graphProducers(root, recipe) {
  const prerequisites = (recipe.prerequisites ?? []).filter((prerequisite) => prerequisite !== "$(NODE_BIN)");
  const { producers, files } = splitProducedPrerequisites(prerequisites, producerBindings(readProducerRegistry(root)));
  if (files.length > 0) throw new Error(`prerequisites ${files.join(", ")} name no registered producer`);
  return producers;
}

// graphLaunch returns the graph runtime invocation for target: the entry's
// environment without the public inputs, with the machine state, the
// target's forwarded inputs and their origins, and the launch bindings.
export function graphLaunch(target, env, { manifest, root = repositoryRoot }) {
  const { entry, recipe } = graphRecipe(manifest, target);
  const launch = { ...env };
  for (const name of publicMakeInputNames(manifest)) delete launch[name];
  for (const name of [...machineStateNames, ...launchBindings]) {
    if (Object.hasOwn(env, name)) launch[name] = env[name];
  }
  const sources = parseSources(env.CARTULARY_MAKE_INPUT_SOURCES);
  const forwarded = graphForwardedInputNames(entry);
  for (const name of forwarded) launch[name] = canonicalInputPaths[name] ? path.join(root, canonicalInputPaths[name]) : env[name] ?? "";
  Object.assign(launch, inputSourcesEnvironment(forwarded.map((name) => `${name}=${sources.get(name) ?? "unset"}`)));
  let selection;
  if (recipe.selection === "owner") {
    selection = ["--selection", "owner", "--target", target, "--owner", env.OWNER ?? "",
      ...((env.ROWS ?? "").trim() ? ["--rows", env.ROWS] : []), ...(recipe.service_backed_only === true ? ["--service-backed-only"] : [])];
  } else {
    selection = ["--selection", recipe.selection, "--target", target, ...(recipe.selection === "rows" ? ["--rows", recipe.row_ids.join(",")] : [])];
  }
  if (!env.HARNESS_RUNTIME_BIN) throw new HarnessConfigError("HARNESS_RUNTIME_BIN is required");
  return { command: env.HARNESS_RUNTIME_BIN, args: ["graph", ...selection], env: launch, recipe };
}

// produce runs the producers through the artifact store, as the Make
// recipe's producer step did; a graph child that skips prerequisites skips it.
function produce(root, producers, env) {
  if (producers.length === 0 || env.CARTULARY_HARNESS_SKIP_PREREQUISITES === "1") return 0;
  const caches = { GOCACHE: env.GO_CACHE_DIR, GOMODCACHE: env.GO_MOD_CACHE_DIR, GOTMPDIR: env.GO_TMP_DIR };
  for (const directory of Object.values(caches)) mkdirSync(directory, { recursive: true });
  const produceEnv = { ...env, ...caches, CARTULARY_SUPPRESS_CHILD_SUCCESS: "1" };
  delete produceEnv.CARTULARY_TEST_TARGET;
  const result = spawnSync(env.GO || "go", [...producerCommand, ...producers], { cwd: root, env: produceEnv, stdio: "inherit" });
  if (result.error) throw result.error;
  return result.status ?? 1;
}

// run waits for the launched graph, leaving interrupts to it: Make delivers
// them to the whole process group.
function run(command, args, env) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, { cwd: repositoryRoot, env, stdio: "inherit" });
    const ignore = () => {};
    for (const signal of ["SIGINT", "SIGTERM", "SIGHUP"]) process.on(signal, ignore);
    child.on("error", reject);
    child.on("exit", (code, signal) => resolve(signal ? 128 + ({ SIGINT: 2, SIGTERM: 15, SIGHUP: 1, SIGKILL: 9 }[signal] ?? 1) : code ?? 1));
  });
}

export async function runCommandEntry(target, env = process.env, root = repositoryRoot) {
  // Generation and the source checks validate the manifest; the preflight
  // admits the inputs. The entry only reads it.
  const manifest = JSON.parse(readFileSync(defaultTaskSurfaceManifestPath, "utf8"));
  const launch = graphLaunch(target, env, { manifest, root });
  if (launch.recipe.preflight_owner !== "entrypoint") preflightPublicTarget(target, env);
  const produced = produce(root, graphProducers(root, launch.recipe), env);
  if (produced !== 0) return produced;
  return run(launch.command, launch.args, launch.env);
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  const [target, ...extra] = process.argv.slice(2);
  try {
    if (!target || extra.length > 0) throw new HarnessConfigError("usage: command-entry.mjs <target>", { reason: "usage_error" });
    process.exitCode = await runCommandEntry(target);
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    if (error instanceof HarnessConfigError) {
      process.stderr.write(`[FAIL] failure_class=${error.failure_class} failure_reason=${error.failure_reason} exit_code=${error.exit_code} ${message}\n`);
      process.exitCode = error.exit_code;
    } else {
      process.stderr.write(`${message}\n`);
      process.exitCode = 1;
    }
  }
}
