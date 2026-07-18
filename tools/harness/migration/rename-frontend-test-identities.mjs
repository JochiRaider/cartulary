#!/usr/bin/env node

import { createHash } from "node:crypto";
import { existsSync, readFileSync, readdirSync, renameSync, statSync, writeFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { frontendEvidenceFreshnessDigest } from "../phase-accounting/index.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
const apply = process.argv.includes("--apply");

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function walk(relativeRoot) {
  const absoluteRoot = path.join(root, relativeRoot);
  if (!existsSync(absoluteRoot)) return [];
  const result = [];
  for (const entry of readdirSync(absoluteRoot, { withFileTypes: true })) {
    if (entry.isDirectory() && new Set([".cartulary", "node_modules", "dist", "coverage"]).has(entry.name)) continue;
    const child = path.posix.join(relativeRoot, entry.name);
    if (entry.isDirectory()) result.push(...walk(child));
    else if (entry.isFile()) result.push(child);
  }
  return result.sort(asciiCompare);
}

const filenameOverrides = new Map([
  ["apps/web/e2e/frontend.phase4.public-route.spec.ts", "apps/web/e2e/timeline-public-route.spec.ts"],
  ["apps/web/e2e/frontend.phase10.public-route.spec.ts", "apps/web/e2e/coordination-public-route.spec.ts"],
  ["apps/web/e2e/phase1.spec.ts", "apps/web/e2e/auth-and-incident-directory.spec.ts"],
  ["apps/web/e2e/phase2.spec.ts", "apps/web/e2e/incident-administration.spec.ts"],
  ["apps/web/e2e/phase3.spec.ts", "apps/web/e2e/timeline-workbook.spec.ts"],
  ["apps/web/e2e/phase1.support.spec.ts", "apps/web/e2e/auth.support.spec.ts"],
  ["apps/web/e2e/phase2.support.spec.ts", "apps/web/e2e/incident.support.spec.ts"],
  ["apps/web/e2e/phase3.support.spec.ts", "apps/web/e2e/timeline.support.spec.ts"],
  ["apps/web/e2e/auth-support.spec.ts", "apps/web/e2e/auth.support.spec.ts"],
  ["apps/web/e2e/incident-support.spec.ts", "apps/web/e2e/incident.support.spec.ts"],
  ["apps/web/e2e/timeline-support.spec.ts", "apps/web/e2e/timeline.support.spec.ts"],
  ["apps/web/e2e/measurement/phase3_measurement.spec.ts", "apps/web/e2e/measurement/timeline-grid.spec.ts"],
  ["apps/web/e2e/measurement/phase12_network_flow_grid.spec.ts", "apps/web/e2e/measurement/network-flow-grid.spec.ts"],
  ["apps/web/src/app/App.phase1.test.tsx", "apps/web/src/app/App.auth.test.tsx"],
  ["apps/web/src/app/App.phase1.support.test.tsx", "apps/web/src/app/App.auth.support.test.tsx"],
  ["apps/web/src/app/App.test.tsx", "apps/web/src/app/App.timeline-invalidation.support.test.tsx"],
  ["apps/web/src/workbook/WorkbookShell.phase5.test.tsx", "apps/web/src/workbook/WorkbookShell.evidence.test.tsx"],
  ["apps/web/src/workbook/WorkbookShell.phase6.test.tsx", "apps/web/src/workbook/WorkbookShell.collaboration.test.tsx"],
  ["apps/web/src/workbook/WorkbookShell.phase7.test.tsx", "apps/web/src/workbook/WorkbookShell.history.test.tsx"],
]);

function semanticFilename(relativePath) {
  const override = filenameOverrides.get(relativePath);
  if (override) return override;
  const dirname = path.posix.dirname(relativePath);
  let filename = path.posix.basename(relativePath)
    .replace(/frontend\.phase\d+\./giu, "")
    .replace(/(?:^|[._-])phase\d+(?=[._-])/giu, ".")
    .replace(/(?:^|[._-])sprint\d+(?=[._-])/giu, ".")
    .replace(/\.{2,}/gu, ".")
    .replace(/^\./u, "");
  return path.posix.join(dirname, filename);
}

function semanticTitle(title) {
  const overrides = new Map([
    [
      "Phase 3 support suppresses self-originated websocket invalidations without refocusing the draft row",
      "suppresses self-originated websocket invalidations without refocusing the draft row",
    ],
    [
      "support suppresses self-originated websocket invalidations without refocusing the draft row",
      "suppresses self-originated websocket invalidations without refocusing the draft row",
    ],
    [
      "Timeline workbook support suppresses self-originated websocket invalidations without refocusing the draft row",
      "Timeline workbook suppresses self-originated websocket invalidations without refocusing the draft row",
    ],
  ]);
  if (overrides.has(title)) return overrides.get(title);
  let result = title
    .replace(/^support\s+Phase\s+\d+\s+/iu, "support ")
    .replace(/^Phase\s+\d+\s+/iu, "")
    .replace(/^Sprint\s+\d+\s+/iu, "")
    .replace(/^(?:FE-[A-Z]+-P\d+(?:-[A-Z0-9]+)*|[UIEV]-\d+(?:-[A-Z0-9]+)+)\s+/iu, "");
  const semanticPhases = new Map([
    ["Phase 0", "bootstrap"],
    ["Phase 1", "authentication"],
    ["Phase 2", "incident administration"],
    ["Phase 3", "Timeline"],
    ["Phase 4", "record relationship"],
    ["Phase 5", "evidence"],
    ["Phase 6", "collaboration"],
    ["Phase 7", "history"],
    ["Phase 8", "workbook query"],
    ["Phase 9", "extended workbook"],
    ["Phase 10", "recovery and coordination"],
    ["Phase 11", "enterprise integration"],
    ["Phase 12 Network Flow", "Network Flow"],
    ["Phase 12", "Network Flow"],
  ]);
  for (const [delivery, semantic] of [...semanticPhases].sort(([left], [right]) => right.length - left.length)) {
    result = result.replaceAll(delivery, semantic);
  }
  return result.replace(/\bSprint\s+\d+\b/giu, "").replace(/\s{2,}/gu, " ").trim();
}

function semanticIdentifier(identifier) {
  const overrides = new Map([
    ["Phase1A11yAppLocalTestId", "AuthA11yAppLocalTestId"],
    ["phase1A11yAppLocalTestId", "authA11yAppLocalTestId"],
    ["openPhase12Incident", "openNetworkFlowIncident"],
    ["importPhase12NetworkFlowCSV", "importNetworkFlowCSV"],
    ["expectPhase12RuntimeProfile", "expectNetworkFlowRuntimeProfile"],
    ["Phase6RecoveryScenario", "SessionRecoveryScenario"],
    ["phase12NetworkFlowMinimalCSV", "networkFlowMinimalCSV"],
    ["editPhase9GenericCell", "editExtendedSurfaceCell"],
  ]);
  if (overrides.has(identifier)) return overrides.get(identifier);
  return identifier
    .replace(/Phase\d+/gu, "")
    .replace(/phase\d+/gu, "")
    .replace(/Sprint\d+/gu, "")
    .replace(/sprint\d+/gu, "");
}

const testFilePattern = /(?:test|spec)\.[cm]?[jt]sx?$/u;
const testFiles = [...walk("apps"), ...walk("packages")].filter((entry) => testFilePattern.test(entry));
const fileMap = new Map();
for (const relativePath of testFiles) {
  const next = semanticFilename(relativePath);
  if (next === relativePath) continue;
  if (existsSync(path.join(root, next))) throw new Error(`file collision: ${relativePath} -> ${next}`);
  fileMap.set(relativePath, next);
}
const inverseFiles = new Map();
for (const [previous, next] of fileMap) {
  const conflict = inverseFiles.get(next);
  if (conflict) throw new Error(`file collision: ${conflict} and ${previous} -> ${next}`);
  inverseFiles.set(next, previous);
}

const migrationBaseline = JSON.parse(readFileSync(path.join(root, "tools/test_migration_baseline.json"), "utf8"));
const knownFileMap = new Map(fileMap);
for (const relativePath of migrationBaseline.inventories?.phase_identity_paths ?? []) {
  if (!relativePath.startsWith("apps/") && !relativePath.startsWith("packages/")) continue;
  if (!testFilePattern.test(relativePath)) continue;
  knownFileMap.set(relativePath, semanticFilename(relativePath));
}
const basenameMap = new Map();
for (const [previous, next] of knownFileMap) {
  const previousBasename = path.posix.basename(previous);
  const nextBasename = path.posix.basename(next);
  const conflict = basenameMap.get(previousBasename);
  if (conflict && conflict !== nextBasename) {
    throw new Error(`ambiguous basename rewrite: ${previousBasename} -> ${conflict} or ${nextBasename}`);
  }
  basenameMap.set(previousBasename, nextBasename);
}

const titleMap = new Map();
for (const manifestPath of walk("tools/test_families").filter((entry) => entry.endsWith(".json"))) {
  const manifest = JSON.parse(readFileSync(path.join(root, manifestPath), "utf8"));
  for (const row of manifest.rows ?? []) {
    if (!new Set(["vitest", "playwright"]).has(row.runner)) continue;
    for (const title of row.selector?.titles ?? []) {
      const next = semanticTitle(title);
      if (next !== title) titleMap.set(title, next);
    }
  }
}
const directCallTitlePattern = /\b(?:describe|it|test)(?:\.[A-Za-z]+)?\s*\(\s*[`"']([^`"']*)/gu;
for (const relativePath of testFiles) {
  const source = readFileSync(path.join(root, relativePath), "utf8");
  for (const match of source.matchAll(directCallTitlePattern)) {
    const next = semanticTitle(match[1]);
    if (next !== match[1]) titleMap.set(match[1], next);
  }
}

const identifierMap = new Map();
const declarationPattern = /\b(?:function|class|interface|type|const|let|var)\s+([A-Za-z_$][A-Za-z0-9_$]*)/gu;
const deliveryIdentifierPattern = /(?:Phase|phase|Sprint|sprint)\d+/u;
const helperFiles = [...new Set([...testFiles, ...walk("apps/web/e2e/support")])].sort(asciiCompare);
for (const relativePath of helperFiles) {
  const source = readFileSync(path.join(root, relativePath), "utf8");
  const resolved = new Map();
  for (const match of source.matchAll(declarationPattern)) {
    const previous = match[1];
    const next = deliveryIdentifierPattern.test(previous) ? semanticIdentifier(previous) : previous;
    const conflict = resolved.get(next);
    if (conflict && conflict !== previous) {
      throw new Error(`identifier collision in ${relativePath}: ${conflict} and ${previous} -> ${next}`);
    }
    resolved.set(next, previous);
    if (next !== previous) identifierMap.set(previous, next);
  }
}

const frozenPaths = new Set([
  "tools/test_migration_baseline.json",
  "tools/test_migration_crosswalk.json",
  "tools/harness/migration/rename-frontend-test-identities.mjs",
]);
const replacements = [
  ...[...titleMap].sort(([left], [right]) => right.length - left.length || asciiCompare(left, right)),
  ...[...identifierMap].sort(([left], [right]) => right.length - left.length || asciiCompare(left, right)),
  ...[...knownFileMap].sort(([left], [right]) => right.length - left.length || asciiCompare(left, right)),
  ...[...basenameMap].sort(([left], [right]) => right.length - left.length || asciiCompare(left, right)),
];
const changedTextFiles = [];
for (const replacementRoot of ["apps", "packages", "tools", "docs/testing"]) {
  for (const relativePath of walk(replacementRoot)) {
    if (frozenPaths.has(relativePath)) continue;
    if (relativePath.endsWith(".tsbuildinfo")) continue;
    const absolutePath = path.join(root, relativePath);
    if (statSync(absolutePath).size > 5_000_000) continue;
    let source;
    try {
      source = readFileSync(absolutePath, "utf8");
    } catch {
      continue;
    }
    let next = source;
    for (const [previous, replacement] of replacements) next = next.replaceAll(previous, replacement);
    if (next === source) continue;
    changedTextFiles.push(relativePath);
    if (apply) writeFileSync(absolutePath, next);
  }
}

if (apply) {
  for (const [previous, next] of [...fileMap].sort(([left], [right]) => asciiCompare(left, right))) {
    renameSync(path.join(root, previous), path.join(root, next));
  }
}

let catalogOrderRewrites = 0;
for (const relativePath of walk("tools/test_families").filter((entry) => entry.endsWith(".json"))) {
  const absolutePath = path.join(root, relativePath);
  const manifest = JSON.parse(readFileSync(absolutePath, "utf8"));
  let changed = false;
  for (const row of manifest.rows ?? []) {
    if (!Array.isArray(row.selector?.titles)) continue;
    const sorted = [...row.selector.titles].sort(asciiCompare);
    if (new Set(sorted).size !== sorted.length) throw new Error(`${relativePath} ${row.row_id} has duplicate titles`);
    if (JSON.stringify(sorted) === JSON.stringify(row.selector.titles)) continue;
    row.selector.titles = sorted;
    changed = true;
    catalogOrderRewrites += 1;
  }
  if (apply && changed) writeFileSync(absolutePath, `${JSON.stringify(manifest, null, 2)}\n`);
}

const frontendRegistryPath = path.join(root, "tools/frontend_phase_registry.json");
const frontendRegistry = JSON.parse(readFileSync(frontendRegistryPath, "utf8"));
let registryDigestRewrites = 0;
for (const entry of frontendRegistry.phases ?? []) {
  const manifestBody = readFileSync(path.join(root, entry.manifest_path));
  const manifestDigest = createHash("sha256").update(manifestBody).digest("hex");
  const ledgerBody = readFileSync(path.join(root, entry.ledger_path));
  const ledgerDigest = createHash("sha256").update(ledgerBody).digest("hex");
  if (entry.manifest_digest !== manifestDigest) {
    entry.manifest_digest = manifestDigest;
    registryDigestRewrites += 1;
  }
  if (entry.ledger_digest !== ledgerDigest) {
    entry.ledger_digest = ledgerDigest;
    registryDigestRewrites += 1;
  }
}
for (const entry of frontendRegistry.phases ?? []) {
  const freshnessDigest = frontendEvidenceFreshnessDigest(root, frontendRegistry, entry);
  if (entry.evidence_freshness_digest === freshnessDigest) continue;
  entry.evidence_freshness_digest = freshnessDigest;
  registryDigestRewrites += 1;
}
if (apply && registryDigestRewrites > 0) {
  writeFileSync(frontendRegistryPath, `${JSON.stringify(frontendRegistry, null, 2)}\n`);
}

process.stdout.write(`${JSON.stringify({
  mode: apply ? "apply" : "check",
  file_renames: fileMap.size,
  title_renames: titleMap.size,
  helper_renames: identifierMap.size,
  helper_rename_map: Object.fromEntries([...identifierMap].sort(([left], [right]) => asciiCompare(left, right))),
  catalog_order_rewrites: catalogOrderRewrites,
  registry_digest_rewrites: registryDigestRewrites,
  text_files_changed: changedTextFiles.length,
  changed_text_files: changedTextFiles,
}, null, 2)}\n`);
