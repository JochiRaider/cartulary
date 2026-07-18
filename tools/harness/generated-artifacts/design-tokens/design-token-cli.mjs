#!/usr/bin/env node
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  DesignTokenValidationError,
  loadDesignTokenDocument,
  renderDesignTokenTypeScript,
} from "./design-tokens.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../../..");
const defaultRegistryPath = path.join(
  repoRoot,
  "contracts",
  "design",
  "tokens.v1.json",
);
const defaultOutputPath = path.join(
  repoRoot,
  "packages",
  "ui-contracts",
  "src",
  "generated",
  "design-tokens.ts",
);

function usage() {
  throw new Error(
    "usage: generate-design-tokens.mjs [--check] [--registry <path>] [--output <path>]",
  );
}

function parseArgs(argv) {
  const options = {
    check: false,
    registry: defaultRegistryPath,
    output: defaultOutputPath,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--check") {
      options.check = true;
      continue;
    }
    if (arg === "--registry") {
      options.registry = path.resolve(argv[index + 1] ?? "");
      index += 1;
      continue;
    }
    if (arg === "--output") {
      options.output = path.resolve(argv[index + 1] ?? "");
      index += 1;
      continue;
    }
    usage();
  }
  if (!options.registry || !options.output) {
    usage();
  }
  return options;
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const document = loadDesignTokenDocument(options.registry);
  const output = renderDesignTokenTypeScript(document);
  if (options.check) {
    const current = readFileSync(options.output, "utf8");
    if (current !== output) {
      throw new Error(`${path.relative(repoRoot, options.output)} is stale; run make generate`);
    }
    console.log("design token validation passed: generated artifact is current");
    return;
  }
  mkdirSync(path.dirname(options.output), { recursive: true });
  writeFileSync(options.output, output, "utf8");
  console.log(`generated ${path.relative(repoRoot, options.output)}`);
}

try {
  main();
} catch (error) {
  if (error instanceof DesignTokenValidationError) {
    console.error("design token validation failed:");
    for (const failure of error.failures) {
      console.error(`  - ${failure.class}: ${failure.message}`);
    }
    process.exit(1);
  }
  const message = error instanceof Error ? error.message : String(error);
  console.error(`design token generation failed: ${message}`);
  process.exit(1);
}
