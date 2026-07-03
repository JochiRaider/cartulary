import { readdirSync } from "node:fs";
import path from "node:path";

import { resolveRepoPath } from "./repo-paths.mjs";

export function findFilesNamed(root, fileName, { repoRoot }) {
  const files = [];
  const stack = [resolveRepoPath(repoRoot, root)];
  while (stack.length > 0) {
    const current = stack.pop();
    let entries = [];
    try {
      entries = readdirSync(current, { withFileTypes: true });
    } catch {
      continue;
    }
    for (const entry of entries) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
        continue;
      }
      if (entry.isFile() && entry.name === fileName) {
        files.push(next);
      }
    }
  }
  return files.sort((left, right) => left.localeCompare(right));
}
