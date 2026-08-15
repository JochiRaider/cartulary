#!/usr/bin/env node

import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";
import { loadTestCatalog, targetForCatalogRow } from "../test-catalog/index.mjs";
import { validateCanonicalRun } from "./canonical-evidence.mjs";

const root = path.resolve(import.meta.dirname, "../../..");

function usage() {
  throw new Error("usage: canonical-evidence-audit-cli.mjs --owner <owner-id> --evidence-roots-file <path>");
}

function parseArgs(argv) {
  const options = { owner: "", manifest: "" };
  for (let index = 0; index < argv.length; index += 1) {
    if (argv[index] === "--owner") options.owner = argv[++index] ?? "";
    else if (argv[index] === "--evidence-roots-file") options.manifest = argv[++index] ?? "";
    else usage();
  }
  if (!options.owner || !options.manifest) usage();
  return options;
}

try {
  const options = parseArgs(process.argv.slice(2));
  const manifestFile = path.resolve(root, options.manifest);
  const manifest = JSON.parse(readFileSync(manifestFile, "utf8"));
  validateSchemaSync("cartulary.harness_evidence_root_manifest.v1", manifest);
  if (manifest.owner_id !== options.owner) throw new Error("evidence manifest owner does not match OWNER");
  const catalog = loadTestCatalog(root);
  const taskSurface = JSON.parse(readFileSync(path.join(root, "tools/task_surface_owner.json"), "utf8"));
  const targets = new Map(taskSurface.targets.filter((entry) => entry.command_id).map((entry) => [entry.command_id, entry.name]));
  const required = new Map(catalog.rows.filter((row) => row.owner_id === options.owner).map((row) => [row.row_id, targetForCatalogRow(row, { commandTargetByID: targets })]));
  if (required.size === 0) throw new Error(`unknown active owner ${options.owner}`);
  const observed = new Set();
  const entryTargets = new Set();
  for (const entry of manifest.entries) {
    if (entryTargets.has(entry.target)) throw new Error(`duplicate evidence target ${entry.target}`);
    entryTargets.add(entry.target);
    const runRoot = path.resolve(path.dirname(manifestFile), entry.run_root);
    const run = await validateCanonicalRun(runRoot);
    const targetSummary = run.targetSummaries.get(entry.target);
    if (!targetSummary) throw new Error(`${runRoot} has no canonical projection for ${entry.target}`);
    for (const rowID of required.keys()) {
      if (!targetSummary.evidence_refs.includes(`rows/${rowID}.json`)) continue;
      if (!existsSync(path.join(runRoot, "rows", `${rowID}.json`))) throw new Error(`${runRoot} is missing row evidence ${rowID}`);
      if (observed.has(rowID)) throw new Error(`row ${rowID} appears in more than one evidence partition`);
      observed.add(rowID);
    }
  }
  const missing = [...required.keys()].filter((rowID) => !observed.has(rowID));
  if (missing.length > 0) throw new Error(`owner evidence is incomplete: ${missing.join(",")}`);
  process.stdout.write(`[EVIDENCE] status=pass owner=${options.owner} rows=${observed.size} partitions=${manifest.entries.length}\n`);
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
}
