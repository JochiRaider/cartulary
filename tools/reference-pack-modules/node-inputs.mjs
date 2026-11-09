// Copy only the installed dependency closure of the declared generator tools.
// Resolution is local and pinned by the workspace lock; no install or fallback.
import { createRequire } from "node:module";
import { cpSync, lstatSync, mkdirSync, readFileSync, realpathSync, symlinkSync } from "node:fs";
import path from "node:path";

export function nodeInputs(root, destination, names) {
  const selected = new Map();
  mkdirSync(path.join(destination, "store"), {recursive: true});
  function resolvePackage(name, parent) {
    const require = createRequire(path.join(parent, "package.json"));
    let file;
    try { file = require.resolve(`${name}/package.json`); }
    catch { file = require.resolve(name); }
    let directory = path.dirname(realpathSync(file));
    for (;;) {
      try {
        const manifest = JSON.parse(readFileSync(path.join(directory, "package.json")));
        if (manifest.name === name) return {directory, manifest};
      } catch (error) { if (error.code !== "ENOENT") throw error; }
      const parent = path.dirname(directory);
      if (parent === directory) throw new Error(`missing dependency manifest: ${name}`);
      directory = parent;
    }
  }
  function copy(name, parent) {
    const {directory, manifest} = resolvePackage(name, parent);
    if (selected.has(directory)) return selected.get(directory).target;
    const target = path.join(destination, "store", String(selected.size));
    selected.set(directory, {name, version: manifest.version, target});
    cpSync(directory, target, {recursive: true, filter: source => { if (lstatSync(source).isSymbolicLink()) throw new Error(`symlinked dependency content: ${source}`); return path.relative(directory, source).split(path.sep)[0] !== "node_modules"; }});
    for (const dependency of Object.keys(manifest.dependencies ?? {}).sort()) {
      const child = copy(dependency, directory), link = path.join(target, "node_modules", dependency);
      mkdirSync(path.dirname(link), {recursive: true});
      symlinkSync(path.relative(path.dirname(link), child), link);
    }
    return target;
  }
  const manifest = JSON.parse(readFileSync(path.join(root, "package.json")));
  for (const name of names) {
    const target = copy(name, root), actual = JSON.parse(readFileSync(path.join(target, "package.json")));
    if (manifest.devDependencies?.[name] !== actual.version) throw new Error(`generator dependency pin mismatch: ${name}`);
    const link = path.join(destination, name);
    mkdirSync(path.dirname(link), {recursive: true});
    symlinkSync(path.relative(path.dirname(link), target), link);
  }
  return [...selected.values()].map(({name, version}) => ({name, version}));
}
