#!/usr/bin/env node
import path from "node:path";
import { fileURLToPath } from "node:url";

import { validateMigrationHistory } from "../tools/harness/backend/migration-history.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");

function parseArgs(argv) {
  const options = { root: repoRoot };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--root") {
      options.root = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    throw new Error("usage: check-migration-history.mjs [--root <path>]");
  }
  if (!options.root) {
    throw new Error("usage: check-migration-history.mjs [--root <path>]");
  }
  options.root = path.resolve(options.root);
  return options;
}

try {
  const options = parseArgs(process.argv.slice(2));
  const result = validateMigrationHistory(options.root);
  console.log(
    `migration history check passed: ${result.migrationCount} migrations, immutable through ${result.immutableThroughVersion}`,
  );
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  console.error(`migration history check failed: ${message}`);
  process.exit(1);
}
