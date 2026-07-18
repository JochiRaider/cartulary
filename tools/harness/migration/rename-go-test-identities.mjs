#!/usr/bin/env node

import { existsSync, readFileSync, readdirSync, renameSync, statSync, writeFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
const apply = process.argv.includes("--apply");

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function walk(relativeRoot) {
  const result = [];
  const visit = (relativePath) => {
    const absolutePath = path.join(root, relativePath);
    for (const entry of readdirSync(absolutePath, { withFileTypes: true })) {
      const child = path.join(relativePath, entry.name);
      if (entry.isDirectory()) visit(child);
      else if (entry.isFile()) result.push(child);
    }
  };
  visit(relativeRoot);
  return result.sort(asciiCompare);
}

function semanticFilename(filename) {
  if (filename === "phase6_ws_test.go") return "presence_transport_test.go";
  if (filename === "phase2_smoke_test.go") return "incident_membership_process_test.go";
  if (filename === "phase5_smoke_test.go") return "evidence_process_test.go";
  if (filename === "phase3_integration_test.go") return "timeline_event_integration_test.go";
  if (filename === "phase4_integration_test.go") return "resolution_integration_test.go";
  if (filename === "config_phase0_test.go") return "core_config_test.go";
  if (filename === "config_phase10_test.go") return "backup_config_test.go";
  let next = filename
    .replace(/(?:^|_)phase\d+(?=_|\.)/giu, "_")
    .replace(/(?:^|_)sprint\d+(?=_|\.)/giu, "_")
    .replace(/^_+/u, "")
    .replace(/_{2,}/gu, "_");
  if (next === "sentinel_test.go") next = "catalog_evidence_sentinel_test.go";
  return next;
}

function semanticSymbol(symbol) {
  const prefix = /^(Test|Benchmark|Fuzz)/u.exec(symbol)?.[1];
  if (!prefix) return symbol;
  const evidenceClass = /_(U|I|E|V)_\d+(?:_|$)/u.exec(symbol)?.[1];
  let body = symbol.slice(prefix.length);
  body = body
    .replace(/^SupportPhase\d+(?:Integration|SharedHarness)?_?/u, "")
    .replace(/^Phase\d+_?/u, "")
    .replace(/(?:^|_)Phase\d+(?=_|$)/gu, "_")
    .replace(/Phase\d+/gu, "")
    .replace(/phase\d+/gu, "")
    .replace(/Sprint\d+/gu, "")
    .replace(/(?:^|_)(?:U|I|E|V)_\d+(?:_[A-Z0-9]+)*_[0-9]+[A-Z0-9]*(?:_AC\d+)?(?=_|$)/gu, "_")
    .replace(/(?:^|_)AC\d+(?=_|$)/gu, "_")
    .replace(/_ProcessSmoke$/u, "")
    .replace(/^Integration_/u, "")
    .replace(/^Support_/u, "")
    .replace(/_{2,}/gu, "_")
    .replace(/^_+|_+$/gu, "");
  const evidenceSuffix = {
    U: "Unit",
    I: "Integration",
    E: "Process",
    V: "Visual",
  }[evidenceClass];
  if (evidenceSuffix && !body.endsWith(evidenceSuffix)) body = `${body}_${evidenceSuffix}`;
  return `${prefix}${body}`;
}

const identifierOverrides = new Map([
  ["seedPhase5EvidenceRecord", "seedEvidenceAttachmentRecord"],
  ["phase3TimelineCommands", "eventTimelineCommands"],
  ["phase4TimelineCommands", "resolutionTimelineCommands"],
  ["newPhase3TimelineCommands", "newEventTimelineCommands"],
  ["newPhase3TimelineCommandsWithOptions", "newEventTimelineCommandsWithOptions"],
  ["newPhase4TimelineCommands", "newResolutionTimelineCommands"],
  ["requirePhase3MutationRecorded", "requireTimelineMutationRecorded"],
  ["requireSprint7CellValue", "requireCoordinationCellValue"],
  ["requireSprint7CollectionItemCount", "requireCoordinationCollectionItemCount"],
  ["addSprint6RecordRef", "addOptionalSurfaceRecordRef"],
  ["phase0DeploymentProfileConfig", "bootstrapDeploymentProfileConfig"],
  ["phase10DeploymentProfileConfig", "backupDeploymentProfileConfig"],
  ["phase0SupportedDeploymentProfiles", "bootstrapSupportedDeploymentProfiles"],
  ["phase10SupportedDeploymentProfiles", "backupSupportedDeploymentProfiles"],
  ["requireU909OptionalSurfaceResources", "requireOptionalSurfaceResources"],
  ["requireU909BandQuery", "requireOptionalSurfaceBandQuery"],
  ["requireU911CellValue", "requirePartyReferenceCellValue"],
]);

function semanticIdentifier(identifier) {
  return identifierOverrides.get(identifier)
    ?? identifier
      .replace(/Phase\d+/gu, "")
      .replace(/phase\d+/gu, "")
      .replace(/Sprint\d+/gu, "")
      .replace(/sprint\d+/gu, "")
      .replace(/[UIEV]\d{3,}/gu, "");
}

const goFiles = [...walk("internal"), ...walk("cmd")].filter((entry) => entry.endsWith("_test.go"));
const symbolMap = new Map();
const declarationPattern = /^func\s+((?:Test|Benchmark|Fuzz)[A-Za-z0-9_]*(?:(?:Phase|phase|Sprint|sprint)\d+|_(?:U|I|E|V)_\d+_)[A-Za-z0-9_]*)\s*\(/gmu;
for (const relativePath of goFiles) {
  const source = readFileSync(path.join(root, relativePath), "utf8");
  for (const match of source.matchAll(declarationPattern)) {
    const previous = match[1];
    const next = semanticSymbol(previous);
    if (next === previous) throw new Error(`no semantic rewrite for ${relativePath}:${previous}`);
    symbolMap.set(previous, next);
  }
}

const identifierMap = new Map();
const declarationNamePattern = /^(?:func|type|const|var)\s+([A-Za-z_][A-Za-z0-9_]*)/gmu;
const helperDeliveryIdentityPattern = /(?:(?:Phase|phase|Sprint|sprint)\d+|[UIEV]\d{3,})/u;
for (const relativePath of goFiles) {
  const source = readFileSync(path.join(root, relativePath), "utf8");
  for (const match of source.matchAll(declarationNamePattern)) {
    const previous = match[1];
    if (/^(?:Test|Benchmark|Fuzz)/u.test(previous) || !helperDeliveryIdentityPattern.test(previous)) continue;
    const next = semanticIdentifier(previous);
    if (next === previous) throw new Error(`no semantic helper rewrite for ${relativePath}:${previous}`);
    const prior = identifierMap.get(previous);
    if (prior && prior !== next) throw new Error(`inconsistent helper rewrite for ${previous}`);
    identifierMap.set(previous, next);
  }
}

const packageSymbols = new Map();
for (const relativePath of goFiles) {
  const packagePath = path.dirname(relativePath);
  const source = readFileSync(path.join(root, relativePath), "utf8");
  const declarations = [...source.matchAll(/^func\s+((?:Test|Benchmark|Fuzz)[A-Za-z0-9_]*)\s*\(/gmu)].map((match) => match[1]);
  const resolved = declarations.map((symbol) => symbolMap.get(symbol) ?? symbol);
  const existing = packageSymbols.get(packagePath) ?? new Map();
  for (const symbol of resolved) {
    const prior = existing.get(symbol);
    if (prior) throw new Error(`symbol collision in ${packagePath}: ${symbol} (${prior}, ${relativePath})`);
    existing.set(symbol, relativePath);
  }
  packageSymbols.set(packagePath, existing);
}

const packageIdentifiers = new Map();
for (const relativePath of goFiles) {
  const packagePath = path.dirname(relativePath);
  const source = readFileSync(path.join(root, relativePath), "utf8");
  const packageName = /^package\s+([A-Za-z_][A-Za-z0-9_]*)/mu.exec(source)?.[1] ?? "unknown";
  const packageKey = `${packagePath}:${packageName}`;
  const declarations = [...source.matchAll(/^(?:func|type|const|var)\s+([A-Za-z_][A-Za-z0-9_]*)/gmu)].map((match) => match[1]);
  const resolved = declarations.map((identifier) => identifierMap.get(identifier) ?? identifier);
  const existing = packageIdentifiers.get(packageKey) ?? new Map();
  for (const identifier of resolved) {
    if (identifier === "_") continue;
    const prior = existing.get(identifier);
    if (prior) throw new Error(`identifier collision in ${packageKey}: ${identifier} (${prior}, ${relativePath})`);
    existing.set(identifier, relativePath);
  }
  packageIdentifiers.set(packageKey, existing);
}

const fileMap = new Map();
for (const relativePath of goFiles) {
  const filename = path.basename(relativePath);
  const nextFilename = semanticFilename(filename);
  if (filename === nextFilename) continue;
  const nextPath = path.join(path.dirname(relativePath), nextFilename);
  if (existsSync(path.join(root, nextPath)) && nextPath !== relativePath) {
    throw new Error(`file collision: ${relativePath} -> ${nextPath}`);
  }
  fileMap.set(relativePath, nextPath);
}

const inverseFiles = new Map();
for (const [previous, next] of fileMap) {
  const conflict = inverseFiles.get(next);
  if (conflict) throw new Error(`file collision: ${conflict} and ${previous} -> ${next}`);
  inverseFiles.set(next, previous);
}

const replacementRoots = ["internal", "cmd", "tools", "docs/testing"];
const frozenPaths = new Set([
  "tools/test_migration_baseline.json",
  "tools/test_migration_crosswalk.json",
  "tools/harness/migration/rename-go-test-identities.mjs",
]);
const replacements = [
  ...[...symbolMap].sort(([left], [right]) => right.length - left.length || asciiCompare(left, right)),
  ...[...identifierMap].sort(([left], [right]) => right.length - left.length || asciiCompare(left, right)),
  ...[...fileMap].sort(([left], [right]) => right.length - left.length || asciiCompare(left, right)),
];
const changedTextFiles = [];
for (const replacementRoot of replacementRoots) {
  for (const relativePath of walk(replacementRoot)) {
    if (frozenPaths.has(relativePath)) continue;
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

let supportSelectorRewrites = 0;
for (let phase = 0; phase <= 12; phase += 1) {
  const relativePath = `tools/phase${phase}_test_map.json`;
  const absolutePath = path.join(root, relativePath);
  const manifest = JSON.parse(readFileSync(absolutePath, "utf8"));
  let changed = false;
  for (const entry of manifest.support_go_targets ?? []) {
    if (!/(?:phase|sprint)[ _.-]?\d+/iu.test(entry.selection_pattern)) continue;
    const selectors = [...new Set([...(entry.symbols ?? []), ...(entry.symbol ? [entry.symbol] : [])])]
      .sort(asciiCompare);
    if (selectors.length === 0) throw new Error(`${relativePath} support selector has no exact symbols`);
    const previous = entry.selection_pattern;
    entry.selection_pattern = selectors.length === 1
      ? selectors[0]
      : `^(?:${selectors.join("|")})$`;
    entry.primary_evidence_owner = entry.primary_evidence_owner?.replace(previous, entry.selection_pattern);
    entry.execution_label = entry.execution_label?.replace(/\s+(?:phase|sprint)\d+\b/giu, "");
    supportSelectorRewrites += 1;
    changed = true;
  }
  if (apply && changed) writeFileSync(absolutePath, `${JSON.stringify(manifest, null, 2)}\n`);
}

const remainingDeliverySymbols = [...symbolMap.values()].filter((symbol) => /(?:Phase|phase|Sprint|sprint)\d+/u.test(symbol));
if (remainingDeliverySymbols.length > 0) {
  throw new Error(`semantic names still encode delivery identities: ${remainingDeliverySymbols.join(", ")}`);
}

process.stdout.write(`${JSON.stringify({
  mode: apply ? "apply" : "check",
  symbol_renames: symbolMap.size,
  helper_renames: identifierMap.size,
  file_renames: fileMap.size,
  support_selector_rewrites: supportSelectorRewrites,
  text_files_changed: changedTextFiles.length,
  changed_text_files: changedTextFiles,
}, null, 2)}\n`);
