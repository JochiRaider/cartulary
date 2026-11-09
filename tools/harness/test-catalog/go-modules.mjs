import { existsSync, lstatSync, readFileSync } from "node:fs";
import path from "node:path";

// Authored module boundaries are explicit. Discovery never walks upward into
// an arbitrary parent workspace or interprets a local-path replace directive.
export function independentGoModules(root) {
  if (!existsSync(path.join(root, "tools/go_modules.json"))) {
    if (existsSync(path.join(root, "referencepacks"))) throw new Error("missing independent Go module registry");
    return [];
  }
  const registry = JSON.parse(readFileSync(path.join(root, "tools/go_modules.json"), "utf8"));
  if (registry.schema_id !== "cartulary.go_modules.v1" || !Array.isArray(registry.modules)) throw new Error("invalid Go module registry");
  const seen = new Set();
  return registry.modules.map((entry) => {
    if (!/^referencepacks\/[a-z][a-z0-9_-]*$/u.test(entry.root) || seen.has(entry.root) ||
      entry.module_path !== `github.com/JochiRaider/cartulary/${entry.root}` || !/^v[0-9]+\.[0-9]+\.[0-9]+$/u.test(entry.version)) throw new Error("invalid independent Go module boundary");
    seen.add(entry.root);
    return Object.freeze({ ...entry });
  });
}

// A selected module must not acquire a dependency on another producer merely
// because its sibling's directory happens to be present in a developer checkout.
export function validateIndependentGoModule(root, directory) {
  const entry = independentGoModules(root).find(module => module.root === directory);
  if (!entry) throw new Error("unregistered independent Go module selection");
  for (const relative of ["referencepacks", entry.root, `${entry.root}/go.mod`]) {
    const stat = lstatSync(path.join(root, relative));
    if (stat.isSymbolicLink() || relative.endsWith("go.mod") && !stat.isFile()) throw new Error("unsafe independent Go module input");
  }
  const manifest = readFileSync(path.join(root, entry.root, "go.mod"), "utf8");
  if (!manifest.split(/\r?\n/u).includes(`module ${entry.module_path}`) || /^\s*replace\b/mu.test(manifest)) throw new Error("independent module identity or dependency override mismatch");
}

export function goModuleSelection(packageName, modules) {
  const relative = packageName.replace(/^\.\//u, "");
  for (const module of modules) {
    if (relative === module.root || relative.startsWith(`${module.root}/`)) {
      return { directory: module.root, package: relative === module.root ? "." : `.${relative.slice(module.root.length)}` };
    }
  }
  if (relative.startsWith("referencepacks/")) throw new Error("unregistered independent Go module selection");
  return null;
}
