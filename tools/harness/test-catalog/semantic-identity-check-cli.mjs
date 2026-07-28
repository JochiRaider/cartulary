#!/usr/bin/env node

import { existsSync, readFileSync, readdirSync } from "node:fs";
import path from "node:path";
import process from "node:process";

const deliveryIdentityPattern = /(?:^|[^a-z0-9])(?:phase|sprint)[ _.-]?\d+(?:[^a-z0-9]|$)/iu;
const goDeliverySymbolPattern = /(?:(?:phase|sprint)[ _.-]?\d+|_(?:u|i|e|v)_\d+_|(?:^|_)[uiev]\d{3,}(?:_|$))/iu;
const declarationPattern = /^(?:func|type|const|var)\s+([A-Za-z_][A-Za-z0-9_]*)/gmu;
const frontendDeclarationPattern = /\b(?:function|class|interface|type|const|let|var)\s+([A-Za-z_$][A-Za-z0-9_$]*)/gu;
const frontendCallTitlePattern = /\b(?:describe|it|test)(?:\.[A-Za-z]+)?\s*\(\s*[`"']([^`"']*)/gu;
const frontendTitlePattern = /(?:(?:Phase|Sprint)\s+\d+|FE-[A-Z]+-P\d+(?:-[A-Z0-9]+)*|(?:^|\s)[UIEV]-\d+(?:-[A-Z0-9]+)+)/iu;
const catalogIdentityPattern = /(?:^|[^a-z0-9])(?:module|platform|app|web|package|harness)\.[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)+(?:[^a-z0-9]|$)/iu;
const compactDeliveryIdentityPattern = /(?:Phase|phase|Sprint|sprint)\d+/u;
const goFixtureIdentityContextPattern = /\b(?:StartStore|StartServer|SeedLocalUserFlags|CreateIncidentInStore|ClientTxnID|ClientInstanceID|clientTxnID|client_txn_id|incident_key)\b/u;
const goFixtureIdentityValuePattern = /(?:^|[^a-z0-9])(?:txn|req|run|fixture|scenario|artifact|isolation)[_.-]|@example\.test|(?:^|[^a-z0-9])IR-/iu;
const goStringLiteralPattern = /`([^`]*)`|"((?:\\.|[^"\\])*)"/gu;
const frontendFixtureIdentityCallPattern = /\b(uniqueIncidentKey|uniqueTxn|uniqueEmail)\s*\(\s*[`"']([^`"']*)/gu;
const frontendFixtureIdentityMetadataPattern = /\b(createdBy|purpose|scenario|incidentKeyPrefix|txnPrefix|displayPrefix|hostnamePrefix|rawTextPrefix|client_txn_id|incident_key|fixture_id|seed_id|artifact_name)\s*:\s*[`"']([^`"']*)/gu;

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function readJSON(absolutePath) {
  return JSON.parse(readFileSync(absolutePath, "utf8"));
}

function walk(root, relativeRoot) {
  const absoluteRoot = path.join(root, relativeRoot);
  if (!existsSync(absoluteRoot)) return [];
  const result = [];
  for (const entry of readdirSync(absoluteRoot, { withFileTypes: true })) {
    if (entry.isDirectory() && new Set([".cartulary", "node_modules", "dist", "coverage"]).has(entry.name)) continue;
    const child = path.posix.join(relativeRoot, entry.name);
    if (entry.isDirectory()) result.push(...walk(root, child));
    else if (entry.isFile()) result.push(child);
  }
  return result.sort(asciiCompare);
}

function collectGoViolations(root) {
  const violations = [];
  const testFiles = ["internal", "cmd"]
    .flatMap((relativeRoot) => walk(root, relativeRoot))
    .filter((relativePath) => relativePath.endsWith("_test.go"));
  for (const relativePath of testFiles) {
    if (deliveryIdentityPattern.test(relativePath)) {
      violations.push({
        location: relativePath,
        locator_kind: "filename",
        locator: path.basename(relativePath),
        reason: "test filename encodes a delivery phase or sprint",
      });
    }
    const source = readFileSync(path.join(root, relativePath), "utf8");
    for (const match of source.matchAll(declarationPattern)) {
      if (!goDeliverySymbolPattern.test(match[1])) continue;
      violations.push({
        location: relativePath,
        locator_kind: "symbol",
        locator: match[1],
        reason: "test declaration encodes a delivery phase or sprint",
      });
    }
    for (const [lineIndex, line] of source.split("\n").entries()) {
      const identityContext = goFixtureIdentityContextPattern.test(line);
      for (const match of line.matchAll(goStringLiteralPattern)) {
        const value = match[1] ?? match[2] ?? "";
        const catalogIdentity = catalogIdentityPattern.test(value);
        const deliveryIdentity = deliveryIdentityPattern.test(value) || compactDeliveryIdentityPattern.test(value);
        if (!catalogIdentity && !deliveryIdentity) continue;
        if (catalogIdentity && !identityContext) continue;
        if (deliveryIdentity && !identityContext && !goFixtureIdentityValuePattern.test(value)) continue;
        violations.push({
          location: relativePath,
          locator_kind: "fixture_identity",
          locator: `line ${lineIndex + 1}:${value}`,
          reason: catalogIdentity
            ? "identity-bearing Go fixture metadata embeds a catalog identity"
            : "identity-bearing Go fixture metadata encodes a delivery phase or sprint",
        });
      }
    }
  }
  return violations;
}

function collectCatalogGoViolations(root) {
  const catalogRoot = path.join(root, "tools/test_families");
  if (!existsSync(catalogRoot)) return [];
  const violations = [];
  for (const relativePath of walk(root, "tools/test_families").filter((entry) => entry.endsWith(".json"))) {
    const manifest = readJSON(path.join(root, relativePath));
    for (const row of manifest.rows ?? []) {
      if (row.runner !== "go") continue;
      for (const selector of row.selector?.tests ?? []) {
        if (!goDeliverySymbolPattern.test(selector)) continue;
        violations.push({
          location: relativePath,
          locator_kind: "selector",
          locator: selector,
          reason: `Go selector for ${row.row_id} encodes a delivery phase or sprint`,
        });
      }
    }
  }
  return violations;
}

function frontendTestFiles(root) {
  return ["apps", "packages"]
    .flatMap((relativeRoot) => walk(root, relativeRoot))
    .filter((relativePath) => /(?:test|spec)\.[cm]?[jt]sx?$/u.test(relativePath));
}

function collectFrontendViolations(root) {
  const violations = [];
  for (const relativePath of frontendTestFiles(root)) {
    if (deliveryIdentityPattern.test(relativePath) || /(?:^|\/)fe[._-]?p\d+/iu.test(relativePath)) {
      violations.push({
        location: relativePath,
        locator_kind: "filename",
        locator: path.basename(relativePath),
        reason: "frontend test filename encodes a delivery phase or sprint",
      });
    }
    const source = readFileSync(path.join(root, relativePath), "utf8");
    for (const match of source.matchAll(frontendDeclarationPattern)) {
      if (!/(?:Phase|phase|Sprint|sprint)\d+/u.test(match[1])) continue;
      violations.push({
        location: relativePath,
        locator_kind: "symbol",
        locator: match[1],
        reason: "frontend test declaration encodes a delivery phase or sprint",
      });
    }
    for (const match of source.matchAll(frontendCallTitlePattern)) {
      if (!frontendTitlePattern.test(match[1])) continue;
      violations.push({
        location: relativePath,
        locator_kind: "title",
        locator: match[1],
        reason: "frontend suite or test title encodes a delivery phase, sprint, or legacy row",
      });
    }
    for (const pattern of [frontendFixtureIdentityCallPattern, frontendFixtureIdentityMetadataPattern]) {
      for (const match of source.matchAll(pattern)) {
        const value = match[2];
        if (
          !deliveryIdentityPattern.test(value) &&
          !frontendTitlePattern.test(value) &&
          !catalogIdentityPattern.test(value)
        ) continue;
        violations.push({
          location: relativePath,
          locator_kind: "fixture_identity",
          locator: `${match[1]}=${value}`,
          reason: catalogIdentityPattern.test(value)
            ? "identity-bearing fixture metadata embeds a catalog identity"
            : "identity-bearing fixture metadata encodes a delivery phase, sprint, or legacy row",
        });
      }
    }
  }
  for (const relativePath of walk(root, "apps/web/e2e/support").filter((entry) => /\.[cm]?[jt]sx?$/u.test(entry))) {
    const source = readFileSync(path.join(root, relativePath), "utf8");
    for (const match of source.matchAll(frontendDeclarationPattern)) {
      if (!/(?:Phase|phase|Sprint|sprint)\d+/u.test(match[1])) continue;
      violations.push({
        location: relativePath,
        locator_kind: "symbol",
        locator: match[1],
        reason: "frontend test-support declaration encodes a delivery phase or sprint",
      });
    }
  }
  return violations;
}

function collectCatalogFrontendViolations(root) {
  const catalogRoot = path.join(root, "tools/test_families");
  if (!existsSync(catalogRoot)) return [];
  const violations = [];
  for (const relativePath of walk(root, "tools/test_families").filter((entry) => entry.endsWith(".json"))) {
    const manifest = readJSON(path.join(root, relativePath));
    for (const row of manifest.rows ?? []) {
      if (!new Set(["vitest", "playwright"]).has(row.runner)) continue;
      if (deliveryIdentityPattern.test(row.selector?.file ?? "")) {
        violations.push({
          location: relativePath,
          locator_kind: "selector",
          locator: row.selector.file,
          reason: `${row.runner} selector for ${row.row_id} uses a delivery-shaped file`,
        });
      }
      for (const title of row.selector?.titles ?? []) {
        if (!frontendTitlePattern.test(title)) continue;
        violations.push({
          location: relativePath,
          locator_kind: "selector",
          locator: title,
          reason: `${row.runner} selector for ${row.row_id} uses a delivery-shaped title`,
        });
      }
    }
  }
  return violations;
}

function collectFixtureViolations(root) {
  const violations = [];
  const designContractIDs = Array.from(
    { length: 12 },
    (_, index) => `D-VFIX-${String(index + 1).padStart(3, "0")}`,
  );
  for (const relativePath of walk(root, "apps/web/e2e/workbook.visual.spec.ts-snapshots")) {
    if (!/\/(?:fe-v-p\d+|v-\d+-grid-\d+)-/iu.test(relativePath)) continue;
    violations.push({
      location: relativePath,
      locator_kind: "fixture_path",
      locator: path.basename(relativePath),
      reason: "visual golden path encodes a delivery phase or legacy row",
    });
  }
  for (const relativePath of walk(root, "internal/testutil/golden")) {
    if (!deliveryIdentityPattern.test(relativePath)) continue;
    violations.push({
      location: relativePath,
      locator_kind: "fixture_path",
      locator: relativePath,
      reason: "diagnostic golden path encodes a delivery phase or sprint",
    });
  }
  const registryPath = path.join(root, "tools/frontend_visual_fixture_registry.json");
  if (existsSync(registryPath)) {
    const registry = readJSON(registryPath);
    const designContractCounts = new Map(
      designContractIDs.map((designContractID) => [designContractID, 0]),
    );
    for (const fixture of registry.fixtures ?? []) {
      if (!/^visual\.fixture\.[a-z][a-z0-9_]*$/u.test(fixture.fixture_id ?? "")) {
        violations.push({
          location: "tools/frontend_visual_fixture_registry.json",
          locator_kind: "fixture_id",
          locator: fixture.fixture_id ?? "",
          reason: "visual fixture ID is not semantic and owner-neutral",
        });
      }
      if (!/^visual\.seed\.[a-z][a-z0-9_]*$/u.test(fixture.seed_id ?? "")) {
        violations.push({
          location: "tools/frontend_visual_fixture_registry.json",
          locator_kind: "fixture_id",
          locator: fixture.seed_id ?? "",
          reason: "visual seed ID is not semantic and owner-neutral",
        });
      }
      if (fixture.status !== "current") {
        violations.push({
          location: "tools/frontend_visual_fixture_registry.json",
          locator_kind: "fixture_id",
          locator: fixture.fixture_id ?? "",
          reason: "missing, retired, and placeholder fixtures must not remain active migration inputs",
        });
      }
      if (fixture.design_contract_id !== undefined) {
        designContractCounts.set(
          fixture.design_contract_id,
          (designContractCounts.get(fixture.design_contract_id) ?? 0) + 1,
        );
      }
      for (const golden of [fixture.golden_filename, ...(fixture.golden_artifacts ?? [])]) {
        if (!/(?:fe-v-p\d+|v-\d+-grid-\d+)-/iu.test(golden ?? "")) continue;
        violations.push({
          location: "tools/frontend_visual_fixture_registry.json",
          locator_kind: "fixture_path",
          locator: golden,
          reason: "visual fixture metadata references a delivery-shaped golden path",
        });
      }
    }
    for (const designContractID of designContractIDs) {
      const count = designContractCounts.get(designContractID) ?? 0;
      if (count === 1) continue;
      violations.push({
        location: "tools/frontend_visual_fixture_registry.json",
        locator_kind: "fixture_id",
        locator: designContractID,
        reason: `design visual fixture contract must resolve exactly once; found ${count}`,
      });
    }
  }
  return violations;
}

function collectLiveHelperViolations(root) {
  const violations = [];
  const roots = [
    "apps/web/src/app/debug",
    "db/queries",
    "internal/modules",
    "packages/ui-contracts/src",
    "tools/recoverybrowserrestore",
  ];
  for (const relativePath of roots.flatMap((relativeRoot) => walk(root, relativeRoot))) {
    if (
      relativePath.startsWith("internal/modules/") && (relativePath.endsWith("_test.go") || relativePath.includes("/testdata/") || relativePath.includes("/testsupport/"))
    ) continue;
    if (relativePath.startsWith("packages/ui-contracts/src/") && relativePath.endsWith(".test.ts")) continue;
    if (relativePath.includes("/node_modules/") || relativePath.endsWith(".tsbuildinfo")) continue;
    if (deliveryIdentityPattern.test(relativePath) || /(?:Phase|phase|Sprint|sprint)\d+/u.test(relativePath)) {
      violations.push({
        location: relativePath,
        locator_kind: "filename",
        locator: path.basename(relativePath),
        reason: "live production, selector, or harness-helper path encodes a delivery phase or sprint",
      });
    }
    if (!/\.(?:go|[cm]?[jt]sx?)$/u.test(relativePath)) continue;
    const source = readFileSync(path.join(root, relativePath), "utf8");
    const patterns = relativePath.endsWith(".go")
      ? [declarationPattern]
      : [frontendDeclarationPattern];
    for (const pattern of patterns) {
      for (const match of source.matchAll(pattern)) {
        if (!deliveryIdentityPattern.test(match[1]) && !/(?:Phase|phase|Sprint|sprint)\d+/u.test(match[1])) continue;
        violations.push({
          location: relativePath,
          locator_kind: "symbol",
          locator: match[1],
          reason: "live production, selector, or harness-helper declaration encodes a delivery phase or sprint",
        });
      }
    }
  }
  return violations;
}

function collectAllCandidates(root) {
  return [
    ...collectGoViolations(root),
    ...collectCatalogGoViolations(root),
    ...collectFrontendViolations(root),
    ...collectCatalogFrontendViolations(root),
    ...collectFixtureViolations(root),
    ...collectLiveHelperViolations(root),
  ];
}

export function validateSemanticGoIdentities(root) {
  const candidates = [...collectGoViolations(root), ...collectCatalogGoViolations(root)];
  return candidates.sort((left, right) => asciiCompare(
      `${left.location}\0${left.locator_kind}\0${left.locator}`,
      `${right.location}\0${right.locator_kind}\0${right.locator}`,
    ));
}

export function validateSemanticIdentities(root) {
  return collectAllCandidates(root).sort((left, right) => asciiCompare(
      `${left.location}\0${left.locator_kind}\0${left.locator}`,
      `${right.location}\0${right.locator_kind}\0${right.locator}`,
    ));
}

export function main(root = process.cwd()) {
  const violations = validateSemanticIdentities(root);
  if (violations.length === 0) return;
  process.stderr.write("Test identities must be semantic and must not encode delivery phases, sprints, or legacy rows.\n");
  process.stderr.write("Invalid identities:\n");
  for (const violation of violations) {
    process.stderr.write(`  ${violation.location}::${violation.locator_kind}:${violation.locator} (${violation.reason})\n`);
  }
  process.exitCode = 1;
}

if (process.argv[1] && path.resolve(process.argv[1]) === path.resolve(new URL(import.meta.url).pathname)) {
  try {
    main();
  } catch (error) {
    process.stderr.write(`semantic identity check failed: ${error.message}\n`);
    process.exitCode = 1;
  }
}
