#!/usr/bin/env node

import { existsSync, readFileSync, readdirSync } from "node:fs";
import path from "node:path";
import process from "node:process";

const deliveryIdentityPattern = /(?:^|[^a-z0-9])(?:phase|sprint)[ _.-]?\d+(?:[^a-z0-9]|$)/iu;
const goDeliverySymbolPattern = /(?:(?:phase|sprint)[ _.-]?\d+|_(?:u|i|e|v)_\d+_|(?:^|_)[uiev]\d{3,}(?:_|$))/iu;
const declarationPattern = /^(?:func|type|const|var)\s+([A-Za-z_][A-Za-z0-9_]*)/gmu;

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
    const child = path.posix.join(relativeRoot, entry.name);
    if (entry.isDirectory()) result.push(...walk(root, child));
    else if (entry.isFile()) result.push(child);
  }
  return result.sort(asciiCompare);
}

function allowlistEntries(root) {
  const allowlistPath = path.join(root, "tools/delivery_phase_semantic_allowlist.json");
  if (!existsSync(allowlistPath)) return [];
  const document = readJSON(allowlistPath);
  return document.allowlist ?? [];
}

function allowed(entries, violation) {
  return entries.some((entry) => entry.location === violation.location
    && entry.locator_kind === violation.locator_kind
    && entry.locator === violation.locator);
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

export function validateSemanticGoIdentities(root) {
  const entries = allowlistEntries(root);
  const violations = [...collectGoViolations(root), ...collectCatalogGoViolations(root)]
    .filter((violation) => !allowed(entries, violation))
    .sort((left, right) => asciiCompare(
      `${left.location}\0${left.locator_kind}\0${left.locator}`,
      `${right.location}\0${right.locator_kind}\0${right.locator}`,
    ));
  return violations;
}

export function main(root = process.cwd()) {
  const violations = validateSemanticGoIdentities(root);
  if (violations.length === 0) return;
  process.stderr.write("Go test identities must be semantic and must not encode delivery phases or sprints.\n");
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
