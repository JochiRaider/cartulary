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
// A launcher passes its environment through by spreading or handing on the
// parent environment; childEnvironment is the only permitted route. A spawn of
// the unit's own fixture process (it runs inside the unit) is marked on its
// line with unitFixtureMarker.
// A write of a name in Node or Go code: an assignment, a Setenv, a NAME=value
// literal, or an object or map key (a switch case is a read). Go code also
// names a variable through a constant, so a constant declared for an owned name
// stands for it. In shell and Make recipes a write is an assignment, a command
// prefix, an export, local, readonly or declare, or a ${NAME=...} expansion;
// comment lines and quoted text are not code.
const writePatterns = (name) => [
  new RegExp(`\\[["'\`]${name}["'\`]\\]\\s*=(?!=)`, "u"), new RegExp(`\\.${name}\\s*=(?!=)`, "u"), new RegExp(`Setenv\\(["']${name}["']`, "u"),
  new RegExp(`["'\`]${name}=`, "u"), new RegExp(`(?<!\\bcase)(?:^|[\\s{,(])["']?${name}["']?\\s*:(?!:)`, "u"),
];
const goConstantDeclaration = /^\s*(?:const\s+)?([A-Za-z_]\w*)\s*=\s*"([A-Z][A-Z0-9_]*)"/u;
const goConstantWritePatterns = (constant) => [
  new RegExp(`\\[(?:\\w+\\.)?${constant}\\]\\s*=(?!=)`, "u"), new RegExp(`Setenv\\((?:\\w+\\.)?${constant}\\b`, "u"),
  new RegExp(`(?<!\\bcase)(?:^|[\\s{,(])(?:\\w+\\.)?${constant}\\s*:(?!=)`, "u"),
];
const shellWritePattern = (name) => new RegExp(`(?:^|[\\s;&|(){])(?:(?:export|local|readonly|declare)(?:\\s+-[a-zA-Z]+)*\\s+)?${name}=`, "gu");
const shellCode = /\.(?:sh|bash|mk)$|(?:^|\/)Makefile$/u;
const insideQuotes = (line, index) => ["\"", "'"].some((quote) => (line.slice(0, index).split(quote).length - 1) % 2 === 1);
const shellWrites = (line, name) => !/^\s*#/u.test(line) && [...line.matchAll(shellWritePattern(name))].some((match) => !insideQuotes(line, match.index));
const writableCode = /\.(?:mjs|cjs|js|ts|tsx|go|sh|bash|mk)$|(?:^|\/)Makefile$/u;
const testCode = /(?:_test\.go|\.test\.(?:mjs|ts|tsx)|\.spec\.ts)$|\/tests\//u;
const passThrough = [/\.\.\.\s*(?:process\.env|environment|env)\b(?![.\w])/u, /\benv\s*:\s*process\.env\b(?![.\w])/u,
  // Go: a child's Env built from os.Environ() rather than childenv.Build.
  /append\(\s*os\.Environ\(\)/u, /range\s+os\.Environ\(\)/u, /\bEnv\s*=\s*os\.Environ\(\)/u];
export const unitFixtureMarker = "environment: the unit's own fixture process";

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
  if (JSON.stringify(keys) !== JSON.stringify(["budget", "explicit_launchers", "inherited", "owners", "roles", "schema_id", "scopes"])) throw new Error(`${environmentRegistryPath} must hold exactly schema_id, scopes, inherited, budget, explicit_launchers, owners and roles`);
  // owners name the one component that writes each owned name; everything
  // else calls that component's function or forwards the name unchanged.
  const ownedBy = new Map();
  for (const [component, owner] of Object.entries(registry.owners ?? {})) {
    if (!/^[a-z][a-z0-9_]*$/u.test(component) || JSON.stringify(Object.keys(owner ?? {}).sort(asciiCompare)) !== JSON.stringify(["names", "paths"])) throw new Error(`owners.${component} must hold exactly paths and names`);
    requireStringList(owner.paths, `owners.${component}.paths`, { sorted: true });
    requireStringList(owner.names, `owners.${component}.names`, { sorted: true, pattern: exactName });
    for (const name of owner.names) {
      if (ownedBy.has(name)) throw new Error(`${name} has two owners: ${ownedBy.get(name)} and ${component}`);
      ownedBy.set(name, component);
    }
  }
  // explicit_launchers are the files of components that start processes for
  // other components; they build every child environment with childEnvironment.
  requireStringList(registry.explicit_launchers, "explicit_launchers", { sorted: true });
  if (registry.schema_id !== environmentRegistrySchemaID) throw new Error(`${environmentRegistryPath} schema_id must be ${environmentRegistrySchemaID}`);
  const scopeKeys = Object.keys(registry.scopes ?? {}).sort(asciiCompare);
  if (JSON.stringify(scopeKeys) !== JSON.stringify(["excluded", "harness"])) throw new Error(`${environmentRegistryPath} scopes must hold exactly excluded and harness`);
  requireStringList(registry.scopes.excluded, "scopes.excluded");
  requireStringList(registry.scopes.harness, "scopes.harness");
  // inherited holds the process and tool names (and name prefixes) a unit may
  // take from the environment its graph starts in; harness names belong in
  // roles.
  const inheritedKeys = Object.keys(registry.inherited ?? {}).sort(asciiCompare);
  if (JSON.stringify(inheritedKeys) !== JSON.stringify(["names", "prefixes"])) throw new Error(`${environmentRegistryPath} inherited must hold exactly names and prefixes`);
  requireStringList(registry.inherited.names, "inherited.names", { sorted: true, pattern: /^(?!CARTULARY_)[A-Za-z_][A-Za-z0-9_]*$/u });
  requireStringList(registry.inherited.prefixes, "inherited.prefixes", { sorted: true, pattern: /^(?!CARTULARY_)[A-Za-z][A-Za-z0-9_]*_$/u });
  // budget holds the names the plan binds as a unit's resource budget, such
  // as its Go CPU limit. A unit never takes them from its graph's environment,
  // and every child launcher keeps its parent's value, so a process a unit
  // starts runs within the unit's budget.
  requireStringList(registry.budget, "budget", { sorted: true, pattern: /^(?!CARTULARY_)[A-Za-z_][A-Za-z0-9_]*$/u });
  for (const budgeted of registry.budget) if (registry.inherited.names.includes(budgeted)) throw new Error(`${budgeted} is a budget name; remove it from inherited.names`);
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
  const ownerOf = new Map(Object.entries(registry.owners).flatMap(([component, owner]) => owner.names.map((name) => [name, { component, paths: owner.paths.map(globPattern) }])));
  const outsideOwner = (relative, name) => writableCode.test(relative) && !testCode.test(relative) && !ownerOf.get(name).paths.some((pattern) => pattern.test(relative));
  const writeFinding = (relative, index, name) => `${relative}:${index + 1} writes ${name}, which ${ownerOf.get(name).component} owns; call its function or forward the name`;
  const constants = new Map();
  for (const [name] of ownerOf) {
    const patterns = writePatterns(name);
    for (const relative of [...(mentions.get(name)?.harness ?? [])].sort(asciiCompare)) {
      const lines = readFileSync(path.join(root, relative), "utf8").split("\n");
      if (relative.endsWith(".go")) for (const line of lines) if (goConstantDeclaration.exec(line)?.[2] === name) constants.set(goConstantDeclaration.exec(line)[1], name);
      if (!outsideOwner(relative, name)) continue;
      const writes = shellCode.test(relative) ? (line) => shellWrites(line, name) : (line) => patterns.some((pattern) => pattern.test(line));
      lines.forEach((line, index) => { if (line.includes(name) && writes(line)) findings.push(writeFinding(relative, index, name)); });
    }
  }
  if (constants.size > 0) {
    const constantPatterns = [...constants].map(([constant, name]) => ({ constant, name, patterns: goConstantWritePatterns(constant) }));
    const harness = registry.scopes.harness.map(globPattern);
    for (const relative of paths ?? executableRepositoryPaths(root, { untracked: true })) {
      if (!relative.endsWith(".go") || !harness.some((pattern) => pattern.test(relative))) continue;
      let lines;
      try { lines = readFileSync(path.join(root, relative), "utf8").split("\n"); }
      catch (error) { if (error.code === "ENOENT") continue; throw error; }
      for (const { constant, name, patterns } of constantPatterns) {
        if (!outsideOwner(relative, name)) continue;
        lines.forEach((line, index) => { if (line.includes(constant) && patterns.some((pattern) => pattern.test(line))) findings.push(writeFinding(relative, index, name)); });
      }
    }
  }
  const launchers = registry.explicit_launchers.map(globPattern);
  for (const relative of paths ?? executableRepositoryPaths(root, { untracked: true })) {
    if (!launchers.some((pattern) => pattern.test(relative))) continue;
    let text;
    try { text = readFileSync(path.join(root, relative), "utf8"); }
    catch (error) { if (error.code === "ENOENT") continue; throw error; }
    text.split("\n").forEach((line, index) => {
      if (!line.includes(unitFixtureMarker) && passThrough.some((pattern) => pattern.test(line))) findings.push(`${relative}:${index + 1} passes an environment through; build it with childEnvironment`);
    });
  }
  const summary = {
    registered: roleOf.size,
    declared_inputs: [...declared].filter((name) => exactName.test(name)).length,
    harness_names: [...mentions.values()].filter((sides) => sides.harness.size > 0).length,
    roles: Object.fromEntries(environmentRoles.map((role) => [role, registry.roles[role].length])),
  };
  return { findings: findings.sort(asciiCompare), summary };
}
