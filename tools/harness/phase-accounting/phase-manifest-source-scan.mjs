import { readFileSync, readdirSync } from "node:fs";
import path from "node:path";

function collectGoTestFiles(root, relativeRoot) {
  const absoluteRoot = path.join(root, relativeRoot);
  const files = [];
  for (const entry of readdirSync(absoluteRoot, { withFileTypes: true })) {
    const relativePath = path.posix.join(relativeRoot, entry.name);
    if (entry.isDirectory()) {
      files.push(...collectGoTestFiles(root, relativePath));
      continue;
    }
    if (entry.isFile() && entry.name.endsWith("_test.go")) {
      files.push(relativePath);
    }
  }
  return files;
}

export function collectAuthoritativeGoTestFunctions(root, phaseNumber) {
  const testFunctionPattern = new RegExp(
    String.raw`\bfunc\s+(TestPhase${phaseNumber}[A-Za-z0-9_]*_[UIE]_${phaseNumber}_(?:[A-Z0-9]+_)*\d{2})\s*\(`,
    "g",
  );
  const functions = [];
  for (const searchRoot of ["internal", path.posix.join("cmd", "server")]) {
    for (const file of collectGoTestFiles(root, searchRoot)) {
      const source = readFileSync(path.join(root, file), "utf8");
      for (const match of source.matchAll(testFunctionPattern)) {
        functions.push({ file, symbol: match[1] });
      }
    }
  }
  return functions.sort((left, right) => {
    if (left.symbol !== right.symbol) {
      return left.symbol.localeCompare(right.symbol);
    }
    return left.file.localeCompare(right.file);
  });
}
