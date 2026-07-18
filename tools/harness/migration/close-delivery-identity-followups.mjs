#!/usr/bin/env node

import { execFileSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import { createRequire } from "node:module";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
const requireFromRoot = createRequire(path.join(root, "package.json"));
const ledgerPath = "tools/delivery_identity_followup_ledger.json";
const schemaPath = "tools/harness/migration/schemas/cartulary.delivery_identity_followup_ledger.v1.schema.json";
const phasePathPattern = /(?:phase|sprint)[ _.-]?\d+/iu;
const retiredV1PathPatterns = [
  /^docs\/archive\//u,
  /^docs\/testing\/(?:frontend_phase_coverage_ledgers\/fe_p\d+|phase\d+)_coverage_ledger\.md$/u,
  /^tools\/frontend_phase_maps\/fe_p\d+_test_map\.json$/u,
  /^tools\/phase\d+_test_map\.json$/u,
];

function readJSON(relativePath) {
  return JSON.parse(readFileSync(path.join(root, relativePath), "utf8"));
}

function fail(message) {
  throw new Error(`delivery identity follow-up closure failed: ${message}`);
}

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function assertSortedUnique(values, label) {
  const sorted = [...new Set(values)].sort(asciiCompare);
  if (values.length !== sorted.length || values.some((value, index) => value !== sorted[index])) {
    fail(`${label} must be ASCII-sorted and duplicate-free`);
  }
}

function repositoryPaths() {
  return execFileSync("git", ["ls-files", "--cached", "--others", "--exclude-standard", "-z"], {
    cwd: root,
    encoding: "utf8",
  }).split("\0").filter(Boolean).filter((entry) => existsSync(path.join(root, entry))).sort(asciiCompare);
}

function validateSchema(value) {
  const Ajv2020 = requireFromRoot("ajv/dist/2020").default;
  const validator = new Ajv2020({ allErrors: true, strict: false }).compile(readJSON(schemaPath));
  if (!validator(value)) fail(`ledger schema validation failed: ${JSON.stringify(validator.errors)}`);
}

function main() {
  const ledger = readJSON(ledgerPath);
  const baseline = readJSON("tools/test_migration_baseline.json");
  const crosswalk = readJSON("tools/test_migration_crosswalk.json");
  const ownerRegistry = readJSON("tools/test_catalog_owner.json");
  validateSchema(ledger);

  if (ledger.baseline_digest !== crosswalk.baseline_digest) fail("ledger baseline digest differs from the migration crosswalk");
  const owners = new Set(ownerRegistry.owners.map((entry) => entry.owner_id));
  for (const collection of [ledger.path_dispositions, ledger.symbol_dispositions, ledger.retained_protocol_identifiers]) {
    for (const entry of collection) {
      if (!owners.has(entry.owner_id)) fail(`unknown owner ${entry.owner_id}`);
    }
  }

  assertSortedUnique(ledger.path_dispositions.map((entry) => entry.legacy_path), "path dispositions");
  assertSortedUnique(
    ledger.symbol_dispositions.map((entry) => `${entry.location}\0${entry.legacy_locator}`),
    "symbol dispositions",
  );
  assertSortedUnique(
    ledger.retained_protocol_identifiers.map((entry) => `${entry.location}\0${entry.locator}`),
    "retained protocol identifiers",
  );

  const pathDispositions = new Map(ledger.path_dispositions.map((entry) => [entry.legacy_path, entry]));
  const frozenFollowups = baseline.inventories.production_followups;
  if (frozenFollowups.length !== 13) fail(`expected 13 frozen production follow-ups, found ${frozenFollowups.length}`);
  for (const followup of frozenFollowups) {
    const disposition = pathDispositions.get(followup.path);
    if (!disposition) fail(`frozen follow-up has no path disposition: ${followup.path}`);
    if (disposition.owner_id !== followup.owner_id) fail(`owner changed for frozen follow-up ${followup.path}`);
  }
  for (const entry of ledger.path_dispositions) {
    if (existsSync(path.join(root, entry.legacy_path))) fail(`legacy path still exists: ${entry.legacy_path}`);
    if (!existsSync(path.join(root, entry.semantic_path))) fail(`semantic path is missing: ${entry.semantic_path}`);
    if (phasePathPattern.test(entry.semantic_path)) fail(`semantic path remains delivery-shaped: ${entry.semantic_path}`);
  }

  const catalogText = repositoryPaths()
    .filter((entry) => entry.startsWith("tools/test_families/") && entry.endsWith(".json"))
    .map((entry) => readFileSync(path.join(root, entry), "utf8"))
    .join("\n");
  for (const entry of ledger.retained_protocol_identifiers) {
    const source = readFileSync(path.join(root, entry.location), "utf8");
    const count = source.split(entry.locator).length - 1;
    if (count !== 1) fail(`retained protocol locator must resolve exactly once: ${entry.location}::${entry.locator}`);
    if (catalogText.includes(entry.locator)) fail(`retained product protocol leaks into a catalog selector: ${entry.locator}`);
  }

  const unclassifiedPhasePaths = repositoryPaths()
    .filter((entry) => phasePathPattern.test(entry))
    .filter((entry) => !retiredV1PathPatterns.some((pattern) => pattern.test(entry)));
  if (unclassifiedPhasePaths.length > 0) fail(`unclassified delivery-shaped path: ${unclassifiedPhasePaths[0]}`);

  for (const relativePath of [
    "apps/web/src/app/debug/AuthenticationDebugHarness.tsx",
    "apps/web/src/app/debug/IncidentDirectoryDebugHarness.tsx",
    "packages/ui-contracts/src/index.ts",
    "tools/recoverybrowserrestore/main.go",
  ]) {
    const source = readFileSync(path.join(root, relativePath), "utf8");
    if (phasePathPattern.test(source)) fail(`live selector/helper source remains delivery-shaped: ${relativePath}`);
  }

  process.stdout.write(`${JSON.stringify({
    schema_id: "cartulary.delivery_identity_followup_closure_summary.v1",
    status: "pass",
    frozen_followups: frozenFollowups.length,
    renamed_paths: ledger.path_dispositions.length,
    symbol_dispositions: ledger.symbol_dispositions.length,
    retained_product_protocol_identifiers: ledger.retained_protocol_identifiers.length,
    semantic_allowlist_entries: readJSON("tools/delivery_phase_semantic_allowlist.json").allowlist.length,
    unclassified_phase_paths: unclassifiedPhasePaths.length,
  })}\n`);
}

try {
  main();
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
}
