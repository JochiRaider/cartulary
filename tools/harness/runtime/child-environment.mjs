import { readFileSync } from "node:fs";
import path from "node:path";

// childEnvironment builds the environment a harness component gives a process
// it starts outside the controlled unit path (TH-HARNESS Section 5.5, single
// writer, P5-4c). The child receives only:
// - the parent's process and tool basics: the environment registry's
//   inherited names and prefixes and the task surface's global inputs;
// - the parent's resource budget: the registry's budget names, such as the
//   Go CPU limit the plan gave its unit (TH-HARNESS-REQ-379);
// - the harness names the launcher forwards by name (`pass`);
// - the values the launcher sets (`set`; undefined removes a name).
// A launcher never passes its whole environment on, so a value one component
// set for its own children cannot reach another component's processes.

const root = path.resolve(import.meta.dirname, "../../..");
let basics;

function inheritance() {
  if (!basics) {
    const registry = JSON.parse(readFileSync(path.join(root, "tools/harness_environment_registry.json"), "utf8"));
    const manifest = JSON.parse(readFileSync(path.join(root, "tools/task_surface_manifest.json"), "utf8"));
    basics = {
      names: new Set([...registry.inherited.names, ...registry.budget, ...manifest.global_inputs.map((input) => input.name)]),
      prefixes: [...registry.inherited.prefixes],
    };
  }
  return basics;
}

export function childEnvironment(parent = process.env, { pass = [], set = {} } = {}) {
  const { names, prefixes } = inheritance();
  const child = {};
  for (const [name, value] of Object.entries(parent ?? {})) {
    if (typeof value === "string" && (names.has(name) || prefixes.some((prefix) => name.startsWith(prefix)))) child[name] = value;
  }
  for (const name of pass) {
    if (typeof parent?.[name] === "string") child[name] = parent[name];
  }
  for (const [name, value] of Object.entries(set)) {
    if (value === undefined || value === null) delete child[name];
    else child[name] = String(value);
  }
  return child;
}
