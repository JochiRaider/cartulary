import { readFileSync, readdirSync } from "node:fs";
import path from "node:path";

const importPattern = /(?:^|\n)\s*(?:import|export)\s+(?:[^"'\n]*?\s+from\s+)?["']([^"']+)["']/gu;

export function collectTestCatalogImportViolations(root) {
  const catalogRoot = path.join(root, "tools/harness/test-catalog");
  const violations = [];
  for (const entry of readdirSync(catalogRoot, { withFileTypes: true })) {
    if (!entry.isFile() || !entry.name.endsWith(".mjs")) continue;
    const file = path.join(catalogRoot, entry.name);
    const source = readFileSync(file, "utf8");
    for (const match of source.matchAll(importPattern)) {
      const specifier = match[1];
      if (!specifier.startsWith(".")) continue;
      const resolved = path.resolve(catalogRoot, specifier);
      const contractRoot = path.join(root, "tools/harness/contract");
      if (
        resolved !== catalogRoot &&
        !resolved.startsWith(`${catalogRoot}${path.sep}`) &&
        resolved !== contractRoot &&
        !resolved.startsWith(`${contractRoot}${path.sep}`)
      ) {
        violations.push({
          file: path.relative(root, file).split(path.sep).join("/"),
          specifier,
        });
      }
    }
  }
  return violations;
}

export function validateTestCatalogImportBoundary(root) {
  const violations = collectTestCatalogImportViolations(root);
  if (violations.length > 0) {
    throw new Error(
      `test catalog import boundary violation: ${violations.map((entry) => `${entry.file} -> ${entry.specifier}`).join(", ")}`,
    );
  }
  return violations;
}
