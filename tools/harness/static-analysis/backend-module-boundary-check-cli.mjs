#!/usr/bin/env node
import { createHash } from "node:crypto";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { inspectBackendBoundaries } from "./backend-module-boundary-check.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const defaultRepoRoot = path.resolve(scriptDir, "../../..");
const summarySchemaID = "cartulary.backend_module_boundary_summary.v1";

function usage() {
  throw new Error(
    "usage: check-backend-module-boundaries.mjs [--manifest <path>] [--support-inventory <path>] [--root <path>]",
  );
}

function parseArgs(argv) {
  const options = {
    manifest:
      process.env.BACKEND_MODULE_BOUNDARIES_MANIFEST ??
      "tools/backend_module_boundaries.json",
    supportInventory:
      process.env.TEST_SUPPORT_INVENTORY ??
      process.env.CARTULARY_TEST_SUPPORT_INVENTORY ??
      path.join(defaultRepoRoot, "tools/test_support_inventory.json"),
    root: process.env.CARTULARY_REPO_ROOT ?? defaultRepoRoot,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--manifest") {
      options.manifest = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--root") {
      options.root = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--support-inventory") {
      options.supportInventory = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    usage();
  }
  if (!options.manifest || !options.root || !options.supportInventory) {
    usage();
  }
  options.root = path.resolve(options.root);
  options.manifest = resolvePath(options.root, options.manifest);
  options.supportInventory = resolvePath(options.root, options.supportInventory);
  return options;
}

function resolvePath(root, value) {
  return path.isAbsolute(value) ? value : path.join(root, value);
}

function sha256(value) {
  return `sha256:${createHash("sha256").update(value).digest("hex")}`;
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const inspection = inspectBackendBoundaries(options);
  const summary = {
    schema_id: summarySchemaID,
    checked_at: new Date().toISOString(),
    repo_root: options.root,
    manifest_digest: sha256(inspection.manifestRawContent),
    support_inventory_digest: sha256(inspection.supportInventoryRaw),
    effective_scan_excludes: inspection.effectiveScanExcludes,
    dynamic_relation_sources: inspection.dynamicRelationSources,
    violations: inspection.violations,
    result: inspection.violations.length === 0 ? "pass" : "fail",
  };
  console.log(JSON.stringify(summary));
  if (inspection.violations.length > 0) process.exitCode = 1;
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  try {
    main();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.error(`backend module boundary check failed: ${message}`);
    process.exitCode = 1;
  }
}
