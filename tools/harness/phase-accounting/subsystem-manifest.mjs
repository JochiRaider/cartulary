import path from "node:path";

import { readJsonObject } from "../contract/json-shape.mjs";

const registrySchemaID = "cartulary.subsystem_test_registry.v1";
const manifestSchemaID = "cartulary.subsystem_test_map.v1";
const ownerPattern = /^[a-z][a-z0-9_-]*$/;

function requireString(value, label, pattern = null) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  const normalized = value.trim();
  if (pattern && !pattern.test(normalized)) {
    throw new Error(`${label} has invalid value ${normalized}`);
  }
  return normalized;
}

function requireObject(value, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  return value;
}

function validateEntry(entry, label) {
  requireObject(entry, label);
  for (const field of [
    "id",
    "coverage",
    "runner",
    "package",
    "file",
    "execution_dependency",
    "evidence_class",
    "layer",
    "evidence_layer",
    "execution_family",
    "execution_label",
    "default_check_kind",
    "default_check_reason_code",
    "primary_evidence_owner",
    "evidence_delta",
    "warm_local_cost_class",
  ]) {
    requireString(entry[field], `${label}.${field}`);
  }
  if (entry.coverage !== "authoritative") {
    throw new Error(`${label}.coverage must be authoritative`);
  }
  if (entry.runner !== "go_test") {
    throw new Error(`${label}.runner must be go_test`);
  }
  if (entry.default_check_required !== true && entry.default_check_required !== false) {
    throw new Error(`${label}.default_check_required must be a boolean`);
  }
  if (entry.duplicate_of !== null) {
    requireString(entry.duplicate_of, `${label}.duplicate_of`);
  }
  if (entry.symbol !== undefined && entry.symbols !== undefined) {
    throw new Error(`${label} must declare symbol or symbols, not both`);
  }
  const symbols = entry.symbols ?? (entry.symbol ? [entry.symbol] : []);
  if (!Array.isArray(symbols) || symbols.length === 0) {
    throw new Error(`${label} must declare at least one symbol`);
  }
  for (const [index, symbol] of symbols.entries()) {
    requireString(symbol, `${label}.symbols[${index}]`);
  }
}

function loadRegistry(root) {
  const registryPath = path.join(root, "tools/subsystem_test_registry.json");
  const registry = readJsonObject(registryPath, registryPath);
  if (registry.schema_id !== registrySchemaID || !Array.isArray(registry.subsystems)) {
    throw new Error(`${registryPath} must declare ${registrySchemaID} and subsystems[]`);
  }
  const seen = new Set();
  return registry.subsystems.map((entry, index) => {
    const label = `${registryPath}.subsystems[${index}]`;
    requireObject(entry, label);
    const owner = requireString(entry.owner, `${label}.owner`, ownerPattern);
    if (seen.has(owner)) {
      throw new Error(`${registryPath} contains duplicate subsystem ${owner}`);
    }
    seen.add(owner);
    return {
      owner,
      manifestPath: requireString(entry.manifest_path, `${label}.manifest_path`),
    };
  });
}

export function loadSubsystemManifests(root = process.cwd()) {
  return loadRegistry(root).map((entry) => {
    const manifestPath = path.join(root, entry.manifestPath);
    const manifest = readJsonObject(manifestPath, manifestPath);
    if (manifest.schema_id !== manifestSchemaID || manifest.owner !== entry.owner) {
      throw new Error(`${manifestPath} must declare ${manifestSchemaID} for owner ${entry.owner}`);
    }
    if (!Array.isArray(manifest.unit) || !Array.isArray(manifest.integration)) {
      throw new Error(`${manifestPath} must declare unit[] and integration[]`);
    }
    const ids = new Set();
    for (const [section, entries] of [["unit", manifest.unit], ["integration", manifest.integration]]) {
      for (const [index, item] of entries.entries()) {
        validateEntry(item, `${manifestPath}.${section}[${index}]`);
        if (ids.has(item.id)) {
          throw new Error(`${manifestPath} contains duplicate row id ${item.id}`);
        }
        ids.add(item.id);
      }
    }
    return { owner: entry.owner, manifestPath, manifest };
  });
}

export function subsystemManifestOwner(owner) {
  return `subsystem:${owner}`;
}
