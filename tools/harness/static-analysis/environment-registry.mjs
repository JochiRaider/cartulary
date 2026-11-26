import { readFileSync } from "node:fs";
import path from "node:path";

import { executableRepositoryPaths } from "../test-catalog/restricted-input-boundary.mjs";

// The harness environment registry (TH-HARNESS, declared environment, P5-4a).
// Every CARTULARY_* variable name that harness code mentions is declared here
// with one role, or is a public input the task surface already declares. The
// product owns CARTULARY__* configuration and names that appear only in
// product code; a name both sides use is a product interface.

export const environmentRegistryPath = "tools/harness_environment_registry.json";
export const environmentRegistrySchemaID = "cartulary.harness_environment_registry.v1";
export const taskSurfaceManifestPath = "tools/task_surface_manifest.json";

export const environmentRoles = Object.freeze([
  "fixture_binding", // a prepared service, fixture or browser stack handed to a test
  "invocation_context", // run identity, results root, lease, budget and failure context
  "location", // a machine, cache, scratch or layout path
  "mode_switch", // a flag that changes what a command or recipe does
  "product_interface", // a name product code reads that the harness sets
  "step_runtime", // the run-step wrapper's protocol with its child
  "test_internal", // a name a test sets for its own child process or fixture
  "tool_option", // an operator knob of a harness tool, read with a default
]);

const namePattern = /(?<![A-Za-z0-9_])CARTULARY_(?!_)[A-Z0-9]+(?:_[A-Z0-9]+)*(?![A-Za-z0-9_])/gu;
const exactName = /^CARTULARY_(?!_)[A-Z0-9]+(?:_[A-Z0-9]+)*$/u;
const binaryFile = /\.(?:png|jpe?g|gif|webp|ico|pdf|woff2?|ttf|otf|gz|tgz|zip|tar|pack|wasm|sum|der|p12)$/iu;
const asciiCompare = (a, b) => (a < b ? -1 : a > b ? 1 : 0);

// globPattern converts a repository glob (`*` within a segment, `**` across
// segments) to an anchored expression.
export function globPattern(glob) {
  let source = "";
  for (let index = 0; index < glob.length; index += 1) {
    const char = glob[index];
    if (char === "*" && glob[index + 1] === "*") {
      index += 1;
      if (glob[index + 1] === "/") {
        index += 1;
        source += "(?:.*/)?";
      } else source += ".*";
    } else if (char === "*") source += "[^/]*";
    else source += char.replace(/[.+?^${}()|[\]\\]/gu, "\\$&");
  }
  return new RegExp(`^${source}$`, "u");
}

function requireStringList(value, label, { sorted = false, pattern } = {}) {
  if (!Array.isArray(value) || value.some((entry) => typeof entry !== "string" || entry.length === 0)) {
    throw new Error(`${label} must be an array of non-empty strings`);
  }
  if (new Set(value).size !== value.length) throw new Error(`${label} must not contain duplicates`);
  if (sorted && JSON.stringify(value) !== JSON.stringify([...value].sort(asciiCompare))) throw new Error(`${label} must be ASCII-sorted`);
  if (pattern) for (const entry of value) if (!pattern.test(entry)) throw new Error(`${label} entry ${entry} is not a valid name`);
  return value;
}

// readEnvironmentRegistry validates the registry's closed shape.
export function readEnvironmentRegistry(root) {
  const registry = JSON.parse(readFileSync(path.join(root, environmentRegistryPath), "utf8"));
  const keys = Object.keys(registry ?? {}).sort(asciiCompare);
  if (JSON.stringify(keys) !== JSON.stringify(["roles", "schema_id", "scopes"])) throw new Error(`${environmentRegistryPath} must hold exactly schema_id, scopes and roles`);
  if (registry.schema_id !== environmentRegistrySchemaID) throw new Error(`${environmentRegistryPath} schema_id must be ${environmentRegistrySchemaID}`);
  const scopeKeys = Object.keys(registry.scopes ?? {}).sort(asciiCompare);
  if (JSON.stringify(scopeKeys) !== JSON.stringify(["excluded", "harness"])) throw new Error(`${environmentRegistryPath} scopes must hold exactly excluded and harness`);
  requireStringList(registry.scopes.excluded, "scopes.excluded");
  requireStringList(registry.scopes.harness, "scopes.harness");
  const roleKeys = Object.keys(registry.roles ?? {});
  if (JSON.stringify(roleKeys) !== JSON.stringify([...environmentRoles])) throw new Error(`${environmentRegistryPath} roles must be exactly ${environmentRoles.join(", ")} in that order`);
  for (const role of environmentRoles) requireStringList(registry.roles[role], `roles.${role}`, { sorted: true, pattern: exactName });
  return registry;
}

// declaredInputNames is every public input name the task surface declares.
export function declaredInputNames(root) {
  const manifest = JSON.parse(readFileSync(path.join(root, taskSurfaceManifestPath), "utf8"));
  const names = new Set();
  const walk = (value) => {
    if (Array.isArray(value)) value.forEach(walk);
    else if (value && typeof value === "object") {
      if (typeof value.name === "string" && Array.isArray(value.allowed_sources)) names.add(value.name);
      Object.values(value).forEach(walk);
    }
  };
  walk(manifest);
  return names;
}

// scanEnvironmentNames maps each mentioned name to the harness and product
// files that mention it. Excluded paths and the registry itself are not read.
export function scanEnvironmentNames(root, registry, paths = executableRepositoryPaths(root, { untracked: true })) {
  const excluded = registry.scopes.excluded.map(globPattern);
  const harness = registry.scopes.harness.map(globPattern);
  const names = new Map();
  for (const relative of paths) {
    if (relative === environmentRegistryPath || binaryFile.test(relative) || excluded.some((pattern) => pattern.test(relative))) continue;
    let text;
    try { text = readFileSync(path.join(root, relative), "utf8"); }
    catch (error) { if (error.code === "ENOENT" || error.code === "EISDIR") continue; throw error; }
    const side = harness.some((pattern) => pattern.test(relative)) ? "harness" : "product";
    for (const name of new Set(text.match(namePattern) ?? [])) {
      if (!names.has(name)) names.set(name, { harness: new Set(), product: new Set() });
      names.get(name)[side].add(relative);
    }
  }
  return names;
}

// inspectEnvironmentRegistry returns every finding; an empty list passes.
export function inspectEnvironmentRegistry(root, {
  registry = readEnvironmentRegistry(root),
  declared = declaredInputNames(root),
  paths,
} = {}) {
  const mentions = scanEnvironmentNames(root, registry, paths);
  const roleOf = new Map();
  const findings = [];
  for (const role of environmentRoles) {
    for (const name of registry.roles[role]) {
      if (roleOf.has(name)) findings.push(`${name} is registered as both ${roleOf.get(name)} and ${role}`);
      else roleOf.set(name, role);
      if (declared.has(name)) findings.push(`${name} is a declared public input; remove it from roles.${role}`);
    }
  }
  for (const [name, sides] of [...mentions].sort(([a], [b]) => asciiCompare(a, b))) {
    if (sides.harness.size === 0) continue;
    const role = roleOf.get(name);
    const first = (set) => [...set].sort(asciiCompare)[0];
    if (!role && !declared.has(name)) findings.push(`${name} is unregistered (first harness mention: ${first(sides.harness)})`);
    if (role === "product_interface" && sides.product.size === 0) findings.push(`${name} is a product_interface that no product code mentions`);
    if (role && role !== "product_interface" && sides.product.size > 0) findings.push(`${name} is ${role} but product code mentions it (${first(sides.product)})`);
  }
  for (const [name, role] of roleOf) {
    if ((mentions.get(name)?.harness.size ?? 0) === 0) findings.push(`${name} (${role}) is stale: no harness code mentions it`);
  }
  const summary = {
    registered: roleOf.size,
    declared_inputs: [...declared].filter((name) => exactName.test(name)).length,
    harness_names: [...mentions.values()].filter((sides) => sides.harness.size > 0).length,
    roles: Object.fromEntries(environmentRoles.map((role) => [role, registry.roles[role].length])),
  };
  return { findings: findings.sort(asciiCompare), summary };
}
