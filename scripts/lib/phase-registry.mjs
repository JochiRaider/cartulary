import { existsSync, readdirSync } from "node:fs";
import path from "node:path";

import { assertObjectKeys, readJsonObject } from "./json-shape.mjs";

export const phaseRegistrySchemaID = "cartulary.phase_registry.v1";
export const activePhaseStatus = "active";
export const plannedPhaseStatus = "planned";
export const retiredPhaseStatus = "retired";

const validPhaseStatuses = new Set([
  activePhaseStatus,
  plannedPhaseStatus,
  retiredPhaseStatus,
]);
const phaseNamePattern = /^phase(?:0|[1-9]\d*)$/;
const phaseManifestFilenamePattern = /^(phase(?:0|[1-9]\d*))_test_map\.json$/;
const phaseLedgerFilenamePattern = /^(phase(?:0|[1-9]\d*))_coverage_ledger\.md$/;
const phaseRegistryKeys = new Set(["schema_id", "phases"]);
const activePhaseEntryKeys = new Set([
  "phase",
  "order",
  "status",
  "label",
  "manifest_path",
  "ledger_path",
  "scope",
  "normative_owners",
]);
const retiredPhaseEntryKeys = new Set([
  ...activePhaseEntryKeys,
  "retired_reason",
  "retained_artifacts",
]);

export function phaseManifestRoot(root) {
  return process.env.CARTULARY_PHASE_MANIFEST_ROOT
    ? path.resolve(process.env.CARTULARY_PHASE_MANIFEST_ROOT)
    : root;
}

function registryPath(root) {
  return path.join(phaseManifestRoot(root), "tools", "phase_registry.json");
}

function compareRegistryEntries(left, right) {
  return left.order - right.order || left.phase.localeCompare(right.phase);
}

function requireObject(value, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  return value;
}

function requireNonEmptyString(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  return value.trim();
}

function requirePhase(value, label) {
  const phase = requireNonEmptyString(value, label);
  if (!phaseNamePattern.test(phase)) {
    throw new Error(`${label} must match phase0 or phase[1-9][0-9]*`);
  }
  return phase;
}

function requireOrder(value, label) {
  if (!Number.isInteger(value) || value < 0) {
    throw new Error(`${label} must be a non-negative integer`);
  }
  return value;
}

function requireRepoRelativePath(value, label) {
  const relative = requireNonEmptyString(value, label);
  if (path.isAbsolute(relative)) {
    throw new Error(`${label} must be repo-relative`);
  }
  const normalized = path.posix.normalize(relative.replaceAll("\\", "/"));
  if (normalized === "." || normalized.startsWith("../") || normalized.includes("/../")) {
    throw new Error(`${label} must not escape the repository root`);
  }
  if (normalized !== relative.replaceAll("\\", "/")) {
    throw new Error(`${label} must be normalized`);
  }
  return normalized;
}

function phaseFromManifestPath(manifestPath, label) {
  const match = phaseManifestFilenamePattern.exec(path.posix.basename(manifestPath));
  if (!match) {
    throw new Error(`${label} must end with phaseN_test_map.json`);
  }
  return match[1];
}

function phaseFromLedgerPath(ledgerPath, label) {
  const match = phaseLedgerFilenamePattern.exec(path.posix.basename(ledgerPath));
  if (!match) {
    throw new Error(`${label} must end with phaseN_coverage_ledger.md`);
  }
  return match[1];
}

function normalizeEntry(rawEntry, index) {
  const label = `phase registry entry ${index}`;
  const entry = requireObject(rawEntry, label);
  assertObjectKeys(entry, retiredPhaseEntryKeys, label);
  const phase = requirePhase(entry.phase, `${label}.phase`);
  const order = requireOrder(entry.order, `${label}.order`);
  const status = requireNonEmptyString(entry.status, `${label}.status`);
  if (!validPhaseStatuses.has(status)) {
    throw new Error(`${label}.status must be active|planned|retired`);
  }
  assertObjectKeys(
    entry,
    status === retiredPhaseStatus ? retiredPhaseEntryKeys : activePhaseEntryKeys,
    label,
  );
  const manifestPath = requireRepoRelativePath(entry.manifest_path, `${label}.manifest_path`);
  const ledgerPath = requireRepoRelativePath(entry.ledger_path, `${label}.ledger_path`);
  const manifestPhase = phaseFromManifestPath(manifestPath, `${label}.manifest_path`);
  const ledgerPhase = phaseFromLedgerPath(ledgerPath, `${label}.ledger_path`);
  if (manifestPhase !== phase) {
    throw new Error(`${label}.manifest_path declares ${manifestPhase} but phase is ${phase}`);
  }
  if (ledgerPhase !== phase) {
    throw new Error(`${label}.ledger_path declares ${ledgerPhase} but phase is ${phase}`);
  }

  const normalized = {
    phase,
    order,
    status,
    label: requireNonEmptyString(entry.label, `${label}.label`),
    manifest_path: manifestPath,
    ledger_path: ledgerPath,
    scope: requireNonEmptyString(entry.scope, `${label}.scope`),
    normative_owners: requireNonEmptyString(
      entry.normative_owners,
      `${label}.normative_owners`,
    ),
  };

  if (status === retiredPhaseStatus) {
    normalized.retired_reason = requireNonEmptyString(
      entry.retired_reason,
      `${label}.retired_reason`,
    );
    if (!Array.isArray(entry.retained_artifacts) || entry.retained_artifacts.length === 0) {
      throw new Error(`${label}.retained_artifacts must be a non-empty array for retired phases`);
    }
    normalized.retained_artifacts = entry.retained_artifacts.map((artifact, artifactIndex) =>
      requireRepoRelativePath(
        artifact,
        `${label}.retained_artifacts[${artifactIndex}]`,
      ),
    );
  } else if (entry.retired_reason !== undefined || entry.retained_artifacts !== undefined) {
    throw new Error(`${label} must not declare retired metadata unless status=retired`);
  }

  return normalized;
}

export function loadPhaseRegistry(root = process.cwd()) {
  const file = registryPath(root);
  if (!existsSync(file)) {
    throw new Error(`${file} must exist and declare schema_id ${phaseRegistrySchemaID}`);
  }
  const registry = readJsonObject(file, file);
  assertObjectKeys(registry, phaseRegistryKeys, file);
  if (registry.schema_id !== phaseRegistrySchemaID) {
    throw new Error(`${file} must declare schema_id ${phaseRegistrySchemaID}`);
  }
  if (!Array.isArray(registry.phases) || registry.phases.length === 0) {
    throw new Error(`${file}.phases must be a non-empty array`);
  }

  const phases = registry.phases.map((entry, index) => normalizeEntry(entry, index));
  const seenPhases = new Map();
  const seenOrders = new Map();
  for (const entry of phases) {
    if (seenPhases.has(entry.phase)) {
      throw new Error(
        `${file} declares duplicate phase ${entry.phase} at orders ${seenPhases.get(entry.phase)} and ${entry.order}`,
      );
    }
    if (seenOrders.has(entry.order)) {
      throw new Error(
        `${file} declares duplicate order ${entry.order} for ${seenOrders.get(entry.order)} and ${entry.phase}`,
      );
    }
    seenPhases.set(entry.phase, entry.order);
    seenOrders.set(entry.order, entry.phase);
  }

  return {
    path: file,
    phases: phases.sort(compareRegistryEntries),
  };
}

export function phaseRegistryEntries(root = process.cwd()) {
  return loadPhaseRegistry(root).phases;
}

export function activePhaseRegistryEntries(root = process.cwd()) {
  return phaseRegistryEntries(root).filter((entry) => entry.status === activePhaseStatus);
}

export function manifestPhaseRegistryEntries(root = process.cwd()) {
  return phaseRegistryEntries(root).filter(
    (entry) => entry.status !== retiredPhaseStatus && existsSync(repoPath(root, entry.manifest_path)),
  );
}

export function phaseRegistryEntry(root, phase) {
  return phaseRegistryEntries(root).find((entry) => entry.phase === phase) ?? null;
}

export function activePhaseRegistryEntry(root, phase) {
  const entry = phaseRegistryEntry(root, phase);
  if (!entry) {
    return null;
  }
  return entry.status === activePhaseStatus ? entry : null;
}

function repoPath(root, relativePath) {
  return path.join(phaseManifestRoot(root), relativePath);
}

function registeredPaths(entries, field) {
  return new Set(entries.map((entry) => entry[field]));
}

function discoveredPhaseMaps(root) {
  const toolsDir = path.join(phaseManifestRoot(root), "tools");
  if (!existsSync(toolsDir)) {
    return [];
  }
  return readdirSync(toolsDir)
    .filter((filename) => /^phase\d+_test_map\.json$/.test(filename))
    .map((filename) => path.posix.join("tools", filename))
    .sort();
}

function discoveredPhaseLedgers(root) {
  const ledgersDir = path.join(phaseManifestRoot(root), "docs", "testing");
  if (!existsSync(ledgersDir)) {
    return [];
  }
  return readdirSync(ledgersDir)
    .filter((filename) => /^phase\d+_coverage_ledger\.md$/.test(filename))
    .map((filename) => path.posix.join("docs", "testing", filename))
    .sort();
}

export function validatePhaseRegistry(root = process.cwd()) {
  const registry = loadPhaseRegistry(root);
  const entries = registry.phases;
  const activeEntries = entries.filter((entry) => entry.status === activePhaseStatus);
  if (activeEntries.length === 0) {
    throw new Error(`${registry.path} must declare at least one active phase`);
  }

  for (const entry of activeEntries) {
    for (const [field, description] of [
      ["manifest_path", "manifest"],
      ["ledger_path", "ledger"],
    ]) {
      const absolutePath = repoPath(root, entry[field]);
      if (!existsSync(absolutePath)) {
        throw new Error(`active ${entry.phase} ${description} missing: ${entry[field]}`);
      }
    }
  }

  const manifestPaths = registeredPaths(entries, "manifest_path");
  for (const manifestPath of discoveredPhaseMaps(root)) {
    if (!manifestPaths.has(manifestPath)) {
      throw new Error(`unregistered phase test map: ${manifestPath}`);
    }
  }

  const ledgerPaths = registeredPaths(entries, "ledger_path");
  for (const ledgerPath of discoveredPhaseLedgers(root)) {
    if (!ledgerPaths.has(ledgerPath)) {
      throw new Error(`unregistered phase coverage ledger: ${ledgerPath}`);
    }
  }

  return registry;
}

function main(argv) {
  const [command] = argv;
  const root = process.cwd();

  switch (command) {
    case "validate":
      validatePhaseRegistry(root);
      console.log("phase registry verified");
      return;
    case "list":
      console.log(phaseRegistryEntries(root).map((entry) => entry.phase).join("\n"));
      return;
    case "list-active":
      console.log(activePhaseRegistryEntries(root).map((entry) => entry.phase).join("\n"));
      return;
    default:
      throw new Error("usage: phase-registry.mjs validate|list|list-active");
  }
}

if (import.meta.url === `file://${process.argv[1]}`) {
  try {
    main(process.argv.slice(2));
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(`${message}\n`);
    process.exit(1);
  }
}
