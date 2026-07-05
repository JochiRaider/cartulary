import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";

import { readJsonObject } from "../../contract/json-shape.mjs";
import { entryTitles, repoPath } from "./common.mjs";

const baseAuthoritativePlaywrightTitlesByRoot = new Map();

export function baseAuthoritativePlaywrightTitleIndex(root) {
  const normalizedRoot = path.resolve(root);
  const cached = baseAuthoritativePlaywrightTitlesByRoot.get(normalizedRoot);
  if (cached) {
    return cached;
  }
  const index = new Map();
  const toolsDir = repoPath(normalizedRoot, "tools");
  for (const filename of readdirSync(toolsDir).filter((name) =>
    /^phase[0-9]+_test_map\.json$/u.test(name),
  )) {
    const file = path.posix.join("tools", filename);
    const manifest = readJsonObject(repoPath(normalizedRoot, file), file);
    for (const value of Object.values(manifest)) {
      if (!Array.isArray(value)) {
        continue;
      }
      for (const entry of value) {
        if (
          entry?.runner !== "playwright" ||
          entry?.coverage !== "authoritative"
        ) {
          continue;
        }
        for (const title of entryTitles(entry)) {
          index.set(title, entry.id ?? file);
        }
      }
    }
  }
  baseAuthoritativePlaywrightTitlesByRoot.set(normalizedRoot, index);
  return index;
}

const playwrightSourceFilesByRoot = new Map();

export function playwrightSourceFiles(root) {
  const normalizedRoot = path.resolve(root);
  const cached = playwrightSourceFilesByRoot.get(normalizedRoot);
  if (cached) {
    return cached;
  }
  const e2eRoot = path.join(normalizedRoot, "apps", "web", "e2e");
  const files = [];
  const stack = [e2eRoot];
  while (stack.length > 0) {
    const current = stack.pop();
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
        continue;
      }
      if (entry.isFile() && entry.name.endsWith(".spec.ts")) {
        files.push(path.relative(normalizedRoot, next).replaceAll("\\", "/"));
      }
    }
  }
  const sorted = files.sort();
  playwrightSourceFilesByRoot.set(normalizedRoot, sorted);
  return sorted;
}

const goModulePathByRoot = new Map();

export function loadGoModulePath(root) {
  const normalizedRoot = path.resolve(root);
  const cached = goModulePathByRoot.get(normalizedRoot);
  if (cached !== undefined) {
    return cached;
  }
  const goMod = readFileSync(path.join(normalizedRoot, "go.mod"), "utf8");
  const match = goMod.match(/^module\s+(\S+)$/m);
  if (!match) {
    throw new Error("unable to determine Go module path from go.mod");
  }
  goModulePathByRoot.set(normalizedRoot, match[1]);
  return match[1];
}
