import { makeNodeToolResultDirMakeEnvVars, makeNodeToolRuntimeEnvVars } from "../../command-surface/make-node-tools.mjs";
import { globalInputEntries, nonCanonicalPublicMakeVariables, restrictedInternalMakeVariables, retiredPublicMakeVariables, targetEntries } from "./model.mjs";

// How public inputs travel from Make to the harness (P5-3e). The generated
// Make and the harness command entry share these rules, so the entry hands a
// child exactly the inputs a recipe would have.

// machineStateNames are the resolved machine-state locations every harness
// child receives after the public inputs are stripped.
export const machineStateNames = Object.freeze([
  "CARTULARY_MACHINE_CACHE_DIR", "GO_CACHE_DIR", "GO_MOD_CACHE_DIR", "GO_TMP_DIR", "CARTULARY_HARNESS_SCRATCH_ROOT",
]);

// canonicalInputPaths are the inputs whose forwarded value is always the
// checkout's own file, whatever the caller set.
export const canonicalInputPaths = Object.freeze({
  EXECUTION_TOPOLOGY_MANIFEST: "tools/execution_topology_manifest.json",
  SCHEDULER_MANIFEST: "tools/scheduler_manifest.json",
  TASK_SURFACE_MANIFEST: "tools/task_surface_manifest.json",
});

const byName = (left, right) => left.localeCompare(right);

// publicMakeInputNames are every name a public command may receive: each
// public target's declared inputs, the global inputs and the restricted,
// non-canonical and retired names. Children never inherit them.
export function publicMakeInputNames(manifest) {
  const names = new Set([
    ...restrictedInternalMakeVariables,
    ...nonCanonicalPublicMakeVariables,
    ...retiredPublicMakeVariables,
    ...globalInputEntries(manifest).map((input) => input.name),
  ]);
  for (const entry of targetEntries(manifest ?? { targets: [] })) {
    if (entry.target_class !== "public") continue;
    for (const input of entry.input_contract?.inputs ?? []) names.add(input.name);
  }
  return Array.from(names).sort(byName);
}

// requiresLiteralMakeTransport reports whether an input travels as its raw,
// unexpanded Make value.
export function requiresLiteralMakeTransport(input) {
  if (input?.name?.startsWith("UI_")) return true;
  if (input?.child_forwarding === "argv") return true;
  return (
    input?.binding === "make_variable" &&
    input?.summary_emission === "redacted_value" &&
    Array.isArray(input?.allowed_sources) &&
    input.allowed_sources.length === 1 &&
    input.allowed_sources[0] === "make_command_line"
  );
}

export function literalMakeInputNames(manifest) {
  const names = new Set();
  for (const entry of targetEntries(manifest ?? { targets: [] })) {
    if (entry.target_class !== "public") continue;
    for (const input of entry.input_contract?.inputs ?? []) {
      if (requiresLiteralMakeTransport(input)) names.add(input.name);
    }
  }
  return Array.from(names).sort(byName);
}

export function runtimeForwardedInputNames(entry = null) {
  return new Set(
    (entry?.input_contract?.inputs ?? [])
      .filter((input) => ["runtime_env", "argv_and_runtime_env"].includes(input.child_forwarding))
      .map((input) => input.name),
  );
}

// graphForwardedInputNames are the inputs a graph launch receives: the
// runtime-forwarded inputs and the argument inputs, in name order.
export function graphForwardedInputNames(entry = null) {
  return [...new Set([
    ...runtimeForwardedInputNames(entry),
    ...(entry?.input_contract?.inputs ?? []).filter((input) => input.child_forwarding === "argv").map((input) => input.name),
  ])].sort(byName);
}

// stepForwardedInputNames are the inputs a step command's child receives.
export function stepForwardedInputNames(entry = null) {
  return [...runtimeForwardedInputNames(entry)].sort(byName);
}

// toolInputNames are the inputs a Node or owner tool receives: the target's
// declared inputs and the tool's result-directory and runtime variables.
export function toolInputNames(entry, toolID) {
  return [...new Set([
    ...(entry?.input_contract?.inputs ?? []).map((input) => input.name),
    ...makeNodeToolResultDirMakeEnvVars(toolID),
    ...makeNodeToolRuntimeEnvVars(toolID),
  ])];
}

// entryRecipeFamily names how the command entry runs a public recipe, or
// null for a recipe Make still runs itself: help, the UI review launcher,
// the native status tool, cleanup and the Node-free Markdown lint.
export function entryRecipeFamily(recipe, target) {
  if (recipe?.graph_entry === true || recipe?.type === "work_graph") return "graph";
  if (recipe?.type === "step_command" && ["command", "run_step"].includes(recipe.mode) && target !== "lint-markdown") return "step";
  if (recipe?.type === "node_tool" || recipe?.type === "owner_command") return "tool";
  if (recipe?.type === "bootstrap") return "bootstrap";
  if (recipe?.type === "ui_review") return "review";
  return null;
}

// entryInputNames are the inputs the entry hands a recipe's child, in the
// order Make renders them; Make passes the expanded value of each one that
// does not travel literally.
export function entryInputNames(entry, recipe) {
  switch (entryRecipeFamily(recipe, entry?.name)) {
    case "graph": return graphForwardedInputNames(entry);
    case "step": return stepForwardedInputNames(entry);
    case "tool": return toolInputNames(entry, recipe.tool_id);
    case "review": return (entry?.input_contract?.inputs ?? []).map((input) => input.name);
    default: return [];
  }
}
