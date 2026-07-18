#!/usr/bin/env node

import { createHash } from "node:crypto";
import {
  existsSync,
  readFileSync,
  readdirSync,
  renameSync,
  statSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { frontendEvidenceFreshnessDigest } from "../phase-accounting/index.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
const apply = process.argv.includes("--apply");

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function digestFile(relativePath) {
  return `sha256:${createHash("sha256").update(readFileSync(path.join(root, relativePath))).digest("hex")}`;
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

function semanticSlug(value) {
  return value
    .normalize("NFKD")
    .replace(/[^A-Za-z0-9]+/gu, "_")
    .replace(/^_+|_+$/gu, "")
    .toLowerCase();
}

const baseline = JSON.parse(readFileSync(path.join(root, "tools/test_migration_baseline.json"), "utf8"));
const snapshotRoot = "apps/web/e2e/workbook.visual.spec.ts-snapshots";
const frontendPrefixes = new Map([
  [1, ""],
  [2, "incident-directory-"],
  [3, "timeline-"],
  [4, "timeline-mutation-"],
  [5, "entity-"],
  [6, "evidence-"],
  [7, "collaboration-"],
  [8, "workbook-query-"],
  [9, "workbook-inspector-"],
  [11, "design-"],
  [12, "network-flow-"],
]);
const gridPrefixes = new Map([
  [3, "timeline-grid-"],
  [4, "record-relationships-"],
  [5, "evidence-grid-"],
  [6, "collaboration-grid-"],
]);

function semanticSnapshotBasename(basename) {
  const frontend = /^fe-v-p(\d+)-\d+-(.+)$/u.exec(basename);
  if (frontend) {
    const phase = Number(frontend[1]);
    let suffix = frontend[2];
    if (phase === 6) suffix = suffix.replace(/^evidence-/u, "");
    if (phase === 9) suffix = suffix.replace(/^inspector-/u, "");
    if (phase === 12) suffix = suffix.replace(/^network-analysis-/u, "analysis-");
    const prefix = frontendPrefixes.get(phase);
    if (prefix === undefined) throw new Error(`no semantic visual prefix for ${basename}`);
    return `${prefix}${suffix}`;
  }
  const grid = /^v-(\d+)-grid-\d+-(.+)$/u.exec(basename);
  if (grid) {
    const prefix = gridPrefixes.get(Number(grid[1]));
    if (prefix === undefined) throw new Error(`no semantic grid visual prefix for ${basename}`);
    return `${prefix}${grid[2]}`;
  }
  throw new Error(`unrecognized frozen visual golden ${basename}`);
}

const snapshotMap = new Map();
for (const golden of baseline.inventories.visual_goldens ?? []) {
  const basename = path.posix.basename(golden.path);
  const next = path.posix.join(snapshotRoot, semanticSnapshotBasename(basename));
  snapshotMap.set(golden.path, next);
}
if (snapshotMap.size !== 56) throw new Error(`expected 56 frozen visual goldens, found ${snapshotMap.size}`);
if (new Set(snapshotMap.values()).size !== snapshotMap.size) throw new Error("semantic visual golden paths collide");

const registryPath = path.join(root, "tools/frontend_visual_fixture_registry.json");
const originalRegistry = JSON.parse(readFileSync(registryPath, "utf8"));
const retainedFixtures = originalRegistry.fixtures.filter((fixture) => fixture.status === "current");
const deletedFixtures = originalRegistry.fixtures.filter((fixture) => fixture.status !== "current");
const semanticFixtureState = retainedFixtures.every((fixture) => fixture.fixture_id.startsWith("visual.fixture."));
if (
  retainedFixtures.length !== 16 ||
  (semanticFixtureState && deletedFixtures.length !== 0) ||
  (!semanticFixtureState && deletedFixtures.length !== 6)
) {
  throw new Error(`expected 16 current fixtures and either 6 pre-migration or zero post-migration placeholders, found ${retainedFixtures.length} and ${deletedFixtures.length}`);
}

const fixtureIDMap = new Map();
const seedIDMap = new Map();
for (const fixture of retainedFixtures) {
  const slug = semanticSlug(fixture.fixture_title);
  fixtureIDMap.set(fixture.fixture_id, `visual.fixture.${slug}`);
  seedIDMap.set(fixture.seed_id, `visual.seed.${slug}`);
}
if (new Set(fixtureIDMap.values()).size !== fixtureIDMap.size) throw new Error("semantic visual fixture IDs collide");

const internalGoldenMap = new Map();
for (const relativePath of baseline.inventories.phase_identity_paths ?? []) {
  if (!relativePath.startsWith("internal/testutil/golden/phase0/")) continue;
  internalGoldenMap.set(
    relativePath,
    relativePath.replace("internal/testutil/golden/phase0/", "internal/testutil/golden/bootstrap/"),
  );
}
if (internalGoldenMap.size !== 15) throw new Error(`expected 15 bootstrap diagnostic goldens, found ${internalGoldenMap.size}`);

const semanticPhaseNames = new Map([
  ["Phase 0", "Bootstrap"],
  ["Phase 1", "Authentication"],
  ["Phase 2", "Incident administration"],
  ["Phase 3", "Timeline"],
  ["Phase 4", "Record relationships"],
  ["Phase 5", "Evidence"],
  ["Phase 6", "Collaboration"],
  ["Phase 7", "History"],
  ["Phase 8", "Workbook query"],
  ["Phase 9", "Workbook inspector"],
  ["Phase 10", "Recovery and coordination"],
  ["Phase 11", "Enterprise integration"],
  ["Phase 12", "Network Flow"],
]);

const explicitLabels = new Map([
  ["\"phase0\", \"diagnostics\"", "\"bootstrap\", \"diagnostics\""],
  ["phase1-user", "deployment-user"],
  ["phase6-timeline-row", "collaboration-timeline-row"],
  ["phase6-e2e-", "collaboration-e2e-"],
  ["phase8-savedviews-system-fixture-unregistered", "savedviews-system-fixture-unregistered"],
  ["phase8-savedviews-system-fixture", "savedviews-system-fixture"],
  ["txn-phase8-system-fixture-incident", "txn-savedviews-system-fixture-incident"],
  ["phase10-u-10-01-artifacts", "recovery-metadata-artifacts"],
  ["phase11 missing blob fixture", "incident-bundle missing blob fixture"],
  ["phase 3 autosave unit fixture", "Timeline autosave unit fixture"],
  ["phase2 mutation artifact", "incident mutation artifact"],
  ["phase2 mutation artifacts", "incident mutation artifacts"],
  ["Phase6", "Collaboration"],
  ["phase6-", "collaboration-"],
  ["fe-v-p6-preview", "evidence-preview"],
  ["fe-v-p6-download-handle", "evidence-download-handle"],
  ["fe-v-p6-failed-handle", "evidence-failed-handle"],
  ["fe-v-p6-inconsistent-handle", "evidence-inconsistent-handle"],
  ["fe-v-p6-timeline-evidence", "timeline-evidence"],
  ["fe-v-p7-remote", "collaboration-remote"],
  ["fe-v-p10-01-fixture-matrix", "coordination-visual-fixture-matrix"],
  ["fe-v-p10-01-deterministic-seed", "coordination-visual-deterministic-seed"],
  ["fe-v-p11-01-owned-stack-visual-suite", "owned-stack-visual-suite"],
  ["fe-v-p11-02-visual-fixture-matrix", "visual-fixture-matrix"],
]);

const replacements = [];
for (const [previous, next] of snapshotMap) {
  replacements.push([previous, next]);
  replacements.push([
    path.posix.basename(previous).replace(/-linux\.png$/u, ""),
    path.posix.basename(next).replace(/-linux\.png$/u, ""),
  ]);
}
for (const entry of fixtureIDMap) replacements.push(entry);
for (const entry of seedIDMap) replacements.push(entry);
for (const entry of internalGoldenMap) replacements.push(entry);
replacements.push([
  "internal/testutil/golden/phase0/diagnostics",
  "internal/testutil/golden/bootstrap/diagnostics",
]);
replacements.sort(([left], [right]) => right.length - left.length || asciiCompare(left, right));
const explicitTestLabelReplacements = [...explicitLabels]
  .sort(([left], [right]) => right.length - left.length || asciiCompare(left, right));
const semanticPhaseLabelReplacements = [...semanticPhaseNames]
  .sort(([left], [right]) => right.length - left.length || asciiCompare(left, right));

function isTestOrFixtureSource(relativePath) {
  return relativePath.startsWith("apps/web/e2e/") ||
    /(?:^|\/)(?:testdata|testsupport|test-utils)(?:\/|$)/u.test(relativePath) ||
    /(?:^|[._-])(?:test|spec)\.[cm]?[jt]sx?$/u.test(relativePath) ||
    /_test\.go$/u.test(relativePath);
}

const frozenPaths = new Set([
  "tools/test_migration_baseline.json",
  "tools/test_migration_crosswalk.json",
  "tools/harness/migration/rename-fixture-scenario-identities.mjs",
]);
const changedTextFiles = [];
for (const replacementRoot of ["apps", "packages", "internal", "tools", "docs/testing"]) {
  for (const relativePath of walk(replacementRoot)) {
    if (frozenPaths.has(relativePath) || relativePath.endsWith(".tsbuildinfo")) continue;
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
    if (isTestOrFixtureSource(relativePath)) {
      for (const [previous, replacement] of explicitTestLabelReplacements) next = next.replaceAll(previous, replacement);
      if (!new Set([
        "apps/web/e2e/workbook.visual.spec.ts",
        "apps/web/e2e/support/extensions/network_flow_activity/workspace.ts",
      ]).has(relativePath)) {
        for (const [previous, replacement] of semanticPhaseLabelReplacements) next = next.replaceAll(previous, replacement);
      }
    }
    if (next === source) continue;
    changedTextFiles.push(relativePath);
    if (apply) writeFileSync(absolutePath, next);
  }
}

if (apply) {
  const registry = JSON.parse(readFileSync(registryPath, "utf8"));
  registry.fixtures = registry.fixtures
    .filter((fixture) => fixture.status === "current")
    .map((fixture) => {
      if (fixture.fixture_id === "visual.fixture.task_requests_or_decisions") {
        fixture.playwright_scenario_title = "Capture Task Requests or Decisions, Parties link state, Communications Log, Handoff, Status Review, Lesson, keyboard focus, frozen column, resize handle, and fill-down fixtures.";
      }
      return fixture;
    })
    .sort((left, right) => asciiCompare(left.fixture_id, right.fixture_id));
  writeFileSync(registryPath, `${JSON.stringify(registry, null, 2)}\n`);

  for (const [previous, next] of snapshotMap) {
    if (!existsSync(path.join(root, previous))) continue;
    if (existsSync(path.join(root, next))) throw new Error(`visual golden collision: ${next}`);
    renameSync(path.join(root, previous), path.join(root, next));
  }
  const previousInternalRoot = path.join(root, "internal/testutil/golden/phase0");
  const nextInternalRoot = path.join(root, "internal/testutil/golden/bootstrap");
  if (existsSync(previousInternalRoot)) {
    if (existsSync(nextInternalRoot)) throw new Error("bootstrap diagnostic golden destination already exists");
    renameSync(previousInternalRoot, nextInternalRoot);
  }

  const frontendRegistryPath = path.join(root, "tools/frontend_phase_registry.json");
  const frontendRegistry = JSON.parse(readFileSync(frontendRegistryPath, "utf8"));
  for (const entry of frontendRegistry.phases ?? []) {
    entry.manifest_digest = createHash("sha256")
      .update(readFileSync(path.join(root, entry.manifest_path)))
      .digest("hex");
    entry.ledger_digest = createHash("sha256")
      .update(readFileSync(path.join(root, entry.ledger_path)))
      .digest("hex");
  }
  for (const entry of frontendRegistry.phases ?? []) {
    entry.evidence_freshness_digest = frontendEvidenceFreshnessDigest(root, frontendRegistry, entry);
  }
  writeFileSync(frontendRegistryPath, `${JSON.stringify(frontendRegistry, null, 2)}\n`);
}

const digestReport = [];
for (const golden of baseline.inventories.visual_goldens ?? []) {
  const next = snapshotMap.get(golden.path);
  const resolved = existsSync(path.join(root, next)) ? next : golden.path;
  const actual = digestFile(resolved);
  if (actual !== golden.digest) throw new Error(`${resolved} digest ${actual} differs from frozen ${golden.digest}`);
  digestReport.push({ previous_path: golden.path, semantic_path: next, digest: actual });
}
for (const [previous, next] of internalGoldenMap) {
  const resolved = existsSync(path.join(root, next)) ? next : previous;
  digestReport.push({ previous_path: previous, semantic_path: next, digest: digestFile(resolved) });
}

process.stdout.write(`${JSON.stringify({
  schema_id: "cartulary.test_fixture_identity_migration_summary.v1",
  mode: apply ? "apply" : "check",
  status: "pass",
  retained_visual_fixtures: retainedFixtures.length,
  deleted_unowned_visual_fixtures: 6,
  frozen_visual_goldens: snapshotMap.size,
  bootstrap_diagnostic_goldens: internalGoldenMap.size,
  changed_text_files: changedTextFiles.length,
  digest_report: digestReport,
}, null, 2)}\n`);
