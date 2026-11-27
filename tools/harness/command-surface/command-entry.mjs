#!/usr/bin/env node
// The harness command entry (P5-3e). A public target's recipe runs it once,
// under workspace admission and after the Node runtime
// (command-entry.sh). Make passes only what Make alone knows: each public
// input's value and origin (CARTULARY_MAKE_INPUT_SOURCES). The entry sets the
// target's own exports, admits the inputs, produces the target's producers
// and runs the target with the process tree and environment its Make recipe
// used to assemble:
//
// - a graph target launches its graph;
// - a step command runs its operation's command, through the step runner
//   when it reports a step;
// - a Node or owner tool runs through the Make tool launcher;
// - bootstrap runs its selected installation operation.
import { spawn, spawnSync } from "node:child_process";
import { mkdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { HarnessConfigError, preflightPublicTarget } from "../contract/index.mjs";
import { inputSourcesEnvironment, runContextEnvironment } from "../contract/run-context.mjs";
import { operationArgv, operationBindingNames, readOperationOwner } from "../execution/operations.mjs";
import { childEnvironment } from "../runtime/child-environment.mjs";
import { producerBindings, producerInvocation, readProducerRegistry, splitProducedPrerequisites } from "../execution/producers.mjs";
import {
  canonicalInputPaths, defaultTaskSurfaceManifestPath, entryRecipeFamily, graphForwardedInputNames, machineStateNames, publicMakeInputNames,
  requiresLiteralMakeTransport, stepForwardedInputNames, toolInputNames,
} from "../generated-artifacts/index.mjs";

const repositoryRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");

// The bindings a graph launch keeps after the public inputs are stripped.
const launchBindings = ["MAKE", "NODE_BIN", "TEST_SERVICES_BIN"];

// A graph-entry operation target needs what every graph launch needs: the
// runtime that plans and runs it. The graph produces everything its units use
// (P6-5).
const graphEntryPrerequisites = Object.freeze(["$(NODE_BIN)", "$(HARNESS_RUNTIME_BIN)"]);

function parseSources(text = "") {
  return new Map(text.split(/\s+/u).filter(Boolean).map((entry) => {
    const index = entry.indexOf("=");
    return [entry.slice(0, index), entry.slice(index + 1)];
  }));
}

function originsOf(names, sources) {
  return inputSourcesEnvironment(names.map((name) => `${name}=${sources.get(name) ?? "unset"}`));
}

// publicRecipe returns target's entry, recipe and how the entry runs it.
export function publicRecipe(manifest, target) {
  const recipe = manifest.make_recipes[target];
  const entry = manifest.targets.find((candidate) => candidate.name === target);
  if (!recipe || entry?.target_class !== "public") throw new HarnessConfigError(`${target} is not a public target`, { reason: "usage_error" });
  const family = entryRecipeFamily(recipe, target);
  if (family === null) throw new HarnessConfigError(`${target} does not run through the command entry`, { reason: "usage_error" });
  // A graph-entry operation target enters its graph like any graph target.
  if (recipe.graph_entry === true) {
    return { entry, family, recipe: { type: "work_graph", selection: "target", prerequisites: graphEntryPrerequisites, preflight_owner: recipe.preflight_owner } };
  }
  return { entry, family, recipe };
}

// recipeProducers are the producers a recipe needs besides the Node runtime,
// which the admitted launcher readies before the entry runs.
export function recipeProducers(root, recipe) {
  const prerequisites = (recipe.prerequisites ?? []).filter((prerequisite) => prerequisite !== "$(NODE_BIN)");
  const { producers, files } = splitProducedPrerequisites(prerequisites, producerBindings(readProducerRegistry(root)));
  if (files.length > 0) throw new Error(`prerequisites ${files.join(", ")} name no registered producer`);
  return producers;
}

// expandMakeValue expands a recipe value's named bindings as Make did: a
// canonical input is the checkout's own file, and an unset name is empty.
function expandMakeValue(value, env, root) {
  return value.replace(/\$\(([A-Z][A-Z0-9_]*)\)/gu, (_, name) => forwardedValue(name, env, root));
}

function forwardedValue(name, env, root) {
  return canonicalInputPaths[name] ? path.join(root, canonicalInputPaths[name]) : env[name] ?? "";
}

// targetEnvironment is env with the target's own exports, which its Make
// rule used to set: the recipe's exports, and the target as its run context
// unless a public target invoked it.
export function targetEnvironment(target, env, { manifest, root = repositoryRoot }) {
  const recipe = manifest.make_recipes[target] ?? {};
  const result = { ...env };
  for (const [name, value] of Object.entries(recipe.exports ?? {})) result[name] = expandMakeValue(value, env, root);
  if (recipe.test_target === "self" && !Object.hasOwn(env, "CARTULARY_TEST_TARGET")) Object.assign(result, runContextEnvironment({ target }));
  return result;
}

// graphLaunch returns the graph runtime invocation for target: the entry's
// environment without the public inputs, with the machine state, the
// target's forwarded inputs and their origins, and the launch bindings.
export function graphLaunch(target, env, { manifest, root = repositoryRoot }) {
  const { entry, family, recipe } = publicRecipe(manifest, target);
  if (family !== "graph") throw new HarnessConfigError(`${target} does not launch a graph`, { reason: "usage_error" });
  const launch = { ...env };
  for (const name of publicMakeInputNames(manifest)) delete launch[name];
  for (const name of [...machineStateNames, ...launchBindings]) {
    if (Object.hasOwn(env, name)) launch[name] = env[name];
  }
  const sources = parseSources(env.CARTULARY_MAKE_INPUT_SOURCES);
  const forwarded = graphForwardedInputNames(entry);
  for (const name of forwarded) launch[name] = forwardedValue(name, env, root);
  Object.assign(launch, originsOf(forwarded, sources));
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

// childEnvironmentArgs are the `env` arguments that give a step's child the
// environment its recipe gave it: no public input, then the machine state,
// the forwarded inputs with their origins and the recipe's own variables.
function childEnvironmentArgs(manifest, entry, recipe, env, root) {
  const sources = parseSources(env.CARTULARY_MAKE_INPUT_SOURCES);
  const forwarded = stepForwardedInputNames(entry);
  return [
    ...publicMakeInputNames(manifest).flatMap((name) => ["-u", name]),
    ...machineStateNames.map((name) => `${name}=${env[name] ?? ""}`),
    ...forwarded.map((name) => `${name}=${forwardedValue(name, env, root)}`),
    ...(forwarded.length > 0 ? Object.entries(originsOf(forwarded, sources)).map(([name, value]) => `${name}=${value}`) : []),
    ...Object.entries(recipe.env ?? {}).map(([name, value]) => `${name}=${expandMakeValue(value, env, root)}`),
  ];
}

// stepInvocation returns a step command's process: its child, wrapped in the
// step runner when the recipe reports a step. The runner sees the forwarded
// and literal inputs, as the recipe passed them.
export function stepInvocation(target, env, { manifest, root = repositoryRoot }) {
  const { entry, family, recipe } = publicRecipe(manifest, target);
  if (family !== "step") throw new HarnessConfigError(`${target} is not a step command`, { reason: "usage_error" });
  const canonical = Object.fromEntries(Object.entries(canonicalInputPaths).map(([name, file]) => [name, path.join(root, file)]));
  const argv = operationArgv(recipe.operation_id ?? target, recipe, { ...env, ...canonical });
  const child = ["env", ...childEnvironmentArgs(manifest, entry, recipe, env, root), ...argv];
  const reportsStep = recipe.mode === "run_step" || entry.output_policy?.summary_schema === "cartulary.tool_run_summary.v5";
  if (!reportsStep) return { command: child[0], args: child.slice(1), env, recipe };
  const sources = parseSources(env.CARTULARY_MAKE_INPUT_SOURCES);
  const forwarded = stepForwardedInputNames(entry);
  const inputs = new Map((entry.input_contract?.inputs ?? []).map((input) => [input.name, input]));
  const literal = [...inputs.keys()].filter((name) => requiresLiteralMakeTransport(inputs.get(name)) && !forwarded.includes(name))
    .sort((left, right) => left.localeCompare(right));
  const runner = { ...env };
  for (const name of forwarded) runner[name] = forwardedValue(name, env, root);
  for (const name of literal) runner[name] = env[name] ?? "";
  // The recipe recorded the forwarded inputs' origins, then the literal
  // inputs' origins in their place.
  const recorded = literal.length > 0 ? literal : forwarded;
  if (recorded.length > 0) Object.assign(runner, originsOf(recorded, sources));
  // A command reports for the public target that invoked it, else itself.
  if (recipe.mode !== "run_step") Object.assign(runner, runContextEnvironment({ target: env.CARTULARY_TEST_TARGET || target }));
  if (!env.RUN_STEP_SCRIPT) throw new HarnessConfigError("RUN_STEP_SCRIPT is required");
  return { command: env.RUN_STEP_SCRIPT, args: [recipe.mode === "run_step" ? recipe.step_label : target, "--", ...child], env: runner, recipe };
}

// Roles whose names a harness tool never receives from its launcher: product
// configuration. A test's own names are not registered at all.
const undeclaredRoles = new Set(["product_derived", "product_interface"]);

// declaredNames are the names a tool may receive from the entry, the rule unit
// environments follow (TH-HARNESS Section 5.5, default-deny): the registered
// harness names outside the product and test roles and the Make variables
// operation executions name, never a public input. The process and tool
// basics pass through childEnvironment itself.
export function declaredNames(manifest, root = repositoryRoot) {
  const registry = JSON.parse(readFileSync(path.join(root, "tools/harness_environment_registry.json"), "utf8"));
  const publicNames = new Set(publicMakeInputNames(manifest));
  const registered = Object.entries(registry.roles).filter(([role]) => !undeclaredRoles.has(role)).flatMap(([, names]) => names);
  return [...new Set([...registered, ...operationBindingNames(readOperationOwner(root))])].filter((name) => !publicNames.has(name));
}

// toolInvocation returns a Node or owner tool's process: the Make tool
// launcher with a default-deny environment of declared names, the tool's
// inputs, every public input's origin and the target as its run context.
export function toolInvocation(target, env, { manifest, root = repositoryRoot }) {
  const { entry, family, recipe } = publicRecipe(manifest, target);
  if (family !== "tool") throw new HarnessConfigError(`${target} is not a Node tool`, { reason: "usage_error" });
  const sources = parseSources(env.CARTULARY_MAKE_INPUT_SOURCES);
  const publicNames = publicMakeInputNames(manifest);
  const tool = childEnvironment(env, { pass: [...declaredNames(manifest, root), ...machineStateNames] });
  Object.assign(tool, { NODE_BIN: env.NODE_BIN ?? "" }, runContextEnvironment({ target, prepared: true }), originsOf(publicNames, sources));
  for (const name of toolInputNames(entry, recipe.tool_id)) tool[name] = forwardedValue(name, env, root);
  return { command: path.join(root, "tools/harness/execution/run-make-node-tool.sh"), args: [recipe.tool_id], env: tool, recipe };
}

// reviewInvocation returns a UI review command's process: its client with a
// default-deny environment holding the command's own inputs and their
// origins, the output mode, a caller's run identity and the names review
// preparation forwards. The client validates its own inputs and identity.
export async function reviewInvocation(target, env, { manifest, root = repositoryRoot }) {
  const { entry, family } = publicRecipe(manifest, target);
  if (family !== "review") throw new HarnessConfigError(`${target} is not a UI review command`, { reason: "usage_error" });
  const { reviewForwarded } = await import("../browser/review-recipes.mjs");
  const inputs = (entry.input_contract?.inputs ?? []).map((input) => input.name);
  const pass = [...inputs, ...reviewForwarded(root), "CARTULARY_MAKE_INPUT_SOURCES", "CARTULARY_OUTPUT_MODE", "CARTULARY_TEST_RESULTS_DIR", "CARTULARY_TEST_RUN_ID"];
  return { command: env.NODE_BIN || process.execPath, args: [path.join(root, "tools/harness/browser/ui-review/cli.mjs"), target], env: childEnvironment(env, { pass }) };
}

// bootstrapOperation is the installation operation bootstrap selects: the
// command-line selector's variant, else the default selection.
export function bootstrapOperation(target, env, { manifest }) {
  const { entry, family } = publicRecipe(manifest, target);
  if (family !== "bootstrap") throw new HarnessConfigError(`${target} is not bootstrap`, { reason: "usage_error" });
  const binding = entry.selection_binding;
  const sources = parseSources(env.CARTULARY_MAKE_INPUT_SOURCES);
  const selector = sources.get(binding.input) === "cli" ? (env[binding.input] ?? "").trim() : "";
  return binding.variants[selector] ?? binding.default_selection;
}

// produce runs the producers through the artifact store, as the Make
// recipe's producer step did. The entry owns the terminal: it keeps the
// producers' output and shows it only when they fail.
function produce(root, producers, env) {
  if (producers.length === 0) return 0;
  const caches = { GOCACHE: env.GO_CACHE_DIR, GOMODCACHE: env.GO_MOD_CACHE_DIR, GOTMPDIR: env.GO_TMP_DIR };
  for (const directory of Object.values(caches)) mkdirSync(directory, { recursive: true });
  const produceEnv = { ...env, ...caches };
  delete produceEnv.CARTULARY_TEST_TARGET;
  const launcher = producerInvocation(root, produceEnv, producers);
  const result = spawnSync(launcher.command, launcher.args, { cwd: root, env: produceEnv, stdio: ["ignore", "pipe", "pipe"], maxBuffer: 64 * 1024 * 1024 });
  if (result.error) throw result.error;
  const status = result.status ?? 1;
  if (status !== 0) {
    process.stderr.write(result.stdout ?? "");
    process.stderr.write(result.stderr ?? "");
  }
  return status;
}

// run waits for a child, leaving interrupts to it: Make delivers them to the
// whole process group.
function run(command, args, env) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, { cwd: repositoryRoot, env, stdio: "inherit" });
    const ignore = () => {};
    for (const signal of ["SIGINT", "SIGTERM", "SIGHUP"]) process.on(signal, ignore);
    child.on("error", reject);
    child.on("exit", (code, signal) => resolve(signal ? 128 + ({ SIGINT: 2, SIGTERM: 15, SIGHUP: 1, SIGKILL: 9 }[signal] ?? 1) : code ?? 1));
  });
}

function targetSummary(target, status, env, root) {
  return run(env.NODE_BIN || process.execPath, [path.join(root, "tools/harness/execution/cartulary-runner-cli.mjs"), "target-summary", target, status], {
    ...env, TEST_OUTPUT_SCRIPT: path.join(root, "tools/harness/output/test-output.mjs"), TASK_SURFACE_MANIFEST: path.join(root, canonicalInputPaths.TASK_SURFACE_MANIFEST),
  });
}

export async function runCommandEntry(target, inherited = process.env, root = repositoryRoot) {
  // Generation and the source checks validate the manifest; the preflight
  // admits the inputs. The entry only reads it.
  const manifest = JSON.parse(readFileSync(defaultTaskSurfaceManifestPath, "utf8"));
  const { family, recipe } = publicRecipe(manifest, target);
  const env = targetEnvironment(target, inherited, { manifest, root });
  // The UI review client validates its own inputs and identity.
  if (family === "review") {
    const invocation = await reviewInvocation(target, env, { manifest, root });
    return run(invocation.command, invocation.args, invocation.env);
  }
  if (recipe.preflight_owner !== "entrypoint") preflightPublicTarget(target, env, { allowExistingRunRoot: recipe.existing_run_root === true });
  const produced = produce(root, recipeProducers(root, recipe), env);
  if (produced !== 0) return produced;
  if (family === "bootstrap") {
    const operation = bootstrapOperation(target, env, { manifest });
    const status = await run(env.NODE_BIN || process.execPath, [path.join(root, "tools/harness/execution/operation-runner-cli.mjs"), "--report", operation], env);
    return status === 0 ? targetSummary(target, "pass", env, root) : status;
  }
  const invocation = { graph: graphLaunch, step: stepInvocation, tool: toolInvocation }[family](target, env, { manifest, root });
  return run(invocation.command, invocation.args, invocation.env);
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
