import { existsSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { relToRepo, resolveRepoPath } from "../core/public-contract.mjs";

function findRepoRoot(startDir) {
  let current = startDir;
  for (;;) {
    if (
      existsSync(path.join(current, "Makefile")) &&
      existsSync(path.join(current, "go.mod")) &&
      existsSync(path.join(current, "tools", "task_surface_manifest.json"))
    ) {
      return current;
    }
    const parent = path.dirname(current);
    if (parent === current) {
      throw new Error(`could not locate repository root from ${startDir}`);
    }
    current = parent;
  }
}

export function durationBaselineCliContext(importMetaURL) {
  const scriptDir = path.dirname(fileURLToPath(importMetaURL));
  const repoRoot = findRepoRoot(scriptDir);
  return {
    repoRoot,
    resolvePath(file) {
      return resolveRepoPath(repoRoot, file);
    },
    rel(file) {
      return relToRepo(repoRoot, file);
    },
  };
}

function requireFlagValue(argv, index, usage) {
  const value = argv[index + 1] ?? "";
  if (!value || value.startsWith("--")) {
    usage();
  }
  return value;
}

export function parseDurationBaselineResultsArgs(
  argv,
  { usage, resolvePath, baselineFile, flags = [] },
) {
  const flagDefinitions = new Map([
    ["--baseline-file", { name: "baselineFile", defaultValue: baselineFile }],
  ]);
  for (const flag of flags) {
    if (flagDefinitions.has(flag.flag)) {
      throw new Error(`duplicate duration baseline CLI flag definition ${flag.flag}`);
    }
    flagDefinitions.set(flag.flag, flag);
  }

  const options = { resultsDir: "" };
  for (const { name, defaultValue } of flagDefinitions.values()) {
    options[name] = resolvePath(defaultValue);
  }

  const seenFlags = new Set();
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    const flag = flagDefinitions.get(arg);
    if (flag) {
      if (seenFlags.has(arg)) {
        usage();
      }
      seenFlags.add(arg);
      options[flag.name] = resolvePath(requireFlagValue(argv, index, usage));
      index += 1;
      continue;
    }
    if (arg.startsWith("--") || options.resultsDir) {
      usage();
    }
    options.resultsDir = resolvePath(arg);
  }

  if (!options.resultsDir) {
    usage();
  }
  return options;
}
