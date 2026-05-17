#!/usr/bin/env node

import { existsSync, readFileSync, readdirSync } from "node:fs";
import path from "node:path";

import {
  collectEntries,
  goEntrySymbols,
  loadManifest,
  phaseManifestNames,
  vitestEntryTitles,
} from "./lib/phase-manifest.mjs";

const testFunctionPattern = /\bfunc\s+(TestPhase(\d+)[A-Za-z0-9_]*)\s*\(/g;

function collectGoTestFiles(root, relativeRoot) {
  const absoluteRoot = path.join(root, relativeRoot);
  if (!existsSync(absoluteRoot)) {
    return [];
  }

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

function collectPhaseTests(root) {
  const tests = [];
  for (const searchRoot of ["internal", path.posix.join("cmd", "server")]) {
    for (const file of collectGoTestFiles(root, searchRoot)) {
      const source = readFileSync(path.join(root, file), "utf8");
      for (const match of source.matchAll(testFunctionPattern)) {
        tests.push({ file, phase: `phase${match[2]}`, symbol: match[1] });
      }
    }
  }
  return tests.sort((left, right) => {
    if (left.symbol !== right.symbol) {
      return left.symbol.localeCompare(right.symbol);
    }
    return left.file.localeCompare(right.file);
  });
}

function manifestIDFragments(root) {
  const fragmentsByPhase = new Map();
  for (const phase of phaseManifestNames(root, { includePlanned: true })) {
    const { manifest } = loadManifest(root, phase, { allowPlanned: true });
    fragmentsByPhase.set(
      phase,
      collectEntries(manifest).map((entry) => entry.id.replaceAll("-", "_")),
    );
  }
  return fragmentsByPhase;
}

function rowIDFragments(id) {
  return [id, id.replaceAll("-", "_")];
}

function entryEvidenceNames(entry) {
  if (entry.runner === "go_test") {
    return goEntrySymbols(entry);
  }
  if (entry.runner === "vitest") {
    return vitestEntryTitles(entry);
  }
  if (entry.runner === "playwright") {
    return [entry.title];
  }
  return [];
}

function validateAuthoritativeManifestNames(root) {
  const invalid = [];

  for (const phase of phaseManifestNames(root, { includePlanned: true })) {
    const { manifest } = loadManifest(root, phase, { allowPlanned: true });
    for (const entry of collectEntries(manifest)) {
      if (entry.coverage !== "authoritative") {
        continue;
      }
      const fragments = rowIDFragments(entry.id);
      for (const name of entryEvidenceNames(entry)) {
        if (!fragments.some((fragment) => name.includes(fragment))) {
          invalid.push({
            file: entry.file,
            phase,
            symbol: name,
            reason: `authoritative evidence for ${entry.id} must include ${fragments.join(" or ")}`,
          });
        }
      }
    }
  }

  return invalid;
}

function validatePhaseTestNames(root) {
  const fragmentsByPhase = manifestIDFragments(root);
  const invalid = validateAuthoritativeManifestNames(root);

  for (const test of collectPhaseTests(root)) {
    const fragments = fragmentsByPhase.get(test.phase);
    if (!fragments) {
      invalid.push({ ...test, reason: `no active phase registry entry exists for ${test.phase}` });
      continue;
    }
    const matched = fragments.some((fragment) => test.symbol.includes(`_${fragment}`));
    if (!matched) {
      invalid.push({
        ...test,
        reason: `name must include one of: ${fragments.join(", ")}`,
      });
    }
  }

  return invalid;
}

function main() {
  const root = process.cwd();
  const invalid = validatePhaseTestNames(root);
  if (invalid.length === 0) {
    return;
  }

  process.stderr.write(
    "Phase test names must carry a manifest-owned ID fragment so the repo keeps a stable naming contract alongside the executable phase manifests.\n",
  );
  process.stderr.write("Invalid names:\n");
  for (const test of invalid) {
    process.stderr.write(`  ${test.file}::${test.symbol} (${test.reason})\n`);
  }
  process.exit(1);
}

try {
  main();
} catch (error) {
  process.stderr.write(`phase test name check failed: ${error.message}\n`);
  process.exit(1);
}
