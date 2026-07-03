#!/usr/bin/env node

import { readFileSync, writeFileSync } from "node:fs";
import {
  defaultGeneratedMakePath,
  defaultTaskSurfaceManifestPath,
  loadTaskSurfaceManifest,
  renderTaskSurfaceMake,
} from "./task-surface.mjs";

function parseArgs(argv) {
  const options = {
    check: false,
    manifest: process.env.CARTULARY_TASK_SURFACE_MANIFEST ?? defaultTaskSurfaceManifestPath,
    output: process.env.CARTULARY_TASK_SURFACE_GENERATED_MAKE ?? defaultGeneratedMakePath,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--check") {
      options.check = true;
      continue;
    }
    if (arg === "--manifest") {
      options.manifest = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--output") {
      options.output = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    throw new Error(`unknown option ${arg}`);
  }
  if (!options.manifest || !options.output) {
    throw new Error("usage: render-task-surface-make.mjs [--check] [--manifest <path>] [--output <path>]");
  }
  return options;
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const { manifest } = loadTaskSurfaceManifest(options.manifest);
  const rendered = renderTaskSurfaceMake(manifest);
  if (options.check) {
    const existing = readFileSync(options.output, "utf8");
    if (existing !== rendered) {
      throw new Error(`${options.output} is stale; run tools/harness/generated-artifacts/render-task-surface-make.mjs`);
    }
    return;
  }
  writeFileSync(options.output, rendered);
}

try {
  main();
} catch (error) {
  console.error(`task-surface Make render failed: ${error.message}`);
  process.exit(1);
}
