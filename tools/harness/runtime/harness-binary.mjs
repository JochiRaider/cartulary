import { spawnSync } from "node:child_process";
import path from "node:path";

import { lifecyclePolicy } from "../contract/generated/lifecycle-policy.mjs";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { workspaceLayout } from "../../workspace_layout.generated.mjs";

const repoRoot = path.resolve(import.meta.dirname, "../../..");

/** The installed harness runtime binary. */
export function harnessBinary(environment = process.env) {
  return environment.CARTULARY_HARNESS_RUNTIME_BIN || path.resolve(repoRoot, workspaceLayout.toolbin, "cartulary-harness");
}

function invoke(args, { root = repoRoot, environment = process.env } = {}) {
  const result = spawnSync(harnessBinary(environment), args, {
    cwd: root, env: environment, encoding: "utf8", maxBuffer: lifecyclePolicy.limits.plan + 1,
  });
  if (result.error) throw result.error;
  if (result.status !== 0) {
    throw Object.assign(new Error(result.stderr.trim() || `cartulary-harness ${args[0]} failed`), { exit_code: result.status ?? 2 });
  }
  return parseStrictJSON(result.stdout);
}

/**
 * The plan the Go planner prepares for graph arguments under environment, as
 * the runtime would execute it. Nothing is launched or published.
 */
export function planInvocation(args, options = {}) {
  const plan = invoke(["plan", ...args], options);
  validateSchemaSync("cartulary.harness_runtime_plan.v9", plan);
  return plan;
}

/** The host's capability snapshot, the scheduler capacities it maps to and the host profile. */
export function hostCapacity({ root, override, environment } = {}) {
  const capacity = invoke(["capacities", ...(override ? ["--override", override] : [])], { root, environment });
  validateSchemaSync("cartulary.harness_capability_snapshot.v1", capacity.capability_snapshot);
  return capacity;
}

// The selection a public target's Make recipe passes to the runtime. Other
// recipe types run no work graph.
const graphSelections = { work_graph: (recipe) => recipe.selection, operation: () => "target" };

/**
 * The plan of public command target with only inputs supplied, as command-line
 * inputs, under the task surface; every other public input is cleared.
 */
export function planCommand({ root, surface, target, inputs = {}, plan = planInvocation }) {
  const recipe = surface.make_recipes[target];
  const selection = graphSelections[recipe?.type]?.(recipe);
  if (!selection) throw new Error(`${target} runs no work graph`);
  const publicInputs = new Set(surface.targets.flatMap((entry) => entry.input_contract?.inputs ?? []).map((input) => input.name));
  const environment = Object.fromEntries(Object.entries(process.env)
    .filter(([name]) => !publicInputs.has(name) && name !== "CARTULARY_MAKE_INPUT_SOURCES"));
  for (const [name, value] of Object.entries(inputs)) environment[name] = String(value);
  environment.CARTULARY_MAKE_INPUT_SOURCES = Object.keys(inputs).map((name) => `${name}=cli`).join(" ");
  return plan(["--selection", selection, "--target", target], { root, environment });
}
