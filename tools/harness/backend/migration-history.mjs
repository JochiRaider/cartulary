import { createHash } from "node:crypto";
import { readFileSync, readdirSync } from "node:fs";
import path from "node:path";

import {
  assertObjectKeys,
  assertRequiredKeys,
  assertUnique,
  readJsonObject,
  requireBoolean,
  requireInteger,
  requireObjectArray,
  requireRepoRelativePath,
  requireSchemaID,
  requireString,
} from "../core/json-shape.mjs";

export const migrationHistorySchemaID = "cartulary.migration_history_manifest.v1";

const manifestKeys = new Set([
  "schema_id",
  "migration_root",
  "immutable_through_version",
  "entries",
]);
const entryKeys = new Set([
  "version",
  "filename",
  "sha256",
  "historical_phase_shaped",
]);
const migrationFilenamePattern = /^(\d{5})_([a-z0-9][a-z0-9_]*)\.sql$/u;
const phaseShapedNamePattern = /^\d{5}_phase[0-9]+_/u;
const sha256Pattern = /^[a-f0-9]{64}$/u;

export function validateMigrationHistoryManifestShape(file) {
  const manifest = readJsonObject(file, file);
  validateManifestShape(manifest, file);
  return manifest;
}

export function validateMigrationHistory(root) {
  const manifestFile = path.join(root, "tools/migration_history_manifest.json");
  const manifest = validateMigrationHistoryManifestShape(manifestFile);
  const migrationDir = path.join(root, manifest.migration_root);
  const files = collectMigrationFiles(migrationDir);
  const entriesByVersion = new Map();

  for (const entry of manifest.entries) {
    entriesByVersion.set(entry.version, entry);
  }

  if (files.length !== manifest.entries.length) {
    throw new Error(
      `${manifestFile} entries must exactly match ${manifest.migration_root}: manifest has ${manifest.entries.length}, directory has ${files.length}`,
    );
  }

  for (const [index, file] of files.entries()) {
    const expectedVersion = index + 1;
    if (file.version !== expectedVersion) {
      throw new Error(
        `${file.filename} creates migration version gap: expected ${formatVersion(expectedVersion)}, got ${formatVersion(file.version)}`,
      );
    }

    validateGooseMarkers(file);

    const entry = entriesByVersion.get(file.version);
    if (!entry) {
      throw new Error(
        `${manifestFile} is missing migration version ${formatVersion(file.version)} (${file.filename})`,
      );
    }
    if (entry.filename !== file.filename) {
      throw new Error(
        `${manifestFile} version ${formatVersion(file.version)} filename drift: got ${file.filename}, manifest has ${entry.filename}`,
      );
    }
    if (entry.sha256 !== file.sha256) {
      throw new Error(
        `${manifestFile} version ${formatVersion(file.version)} hash drift for ${file.filename}: got ${file.sha256}, manifest has ${entry.sha256}`,
      );
    }

    const phaseShaped = phaseShapedNamePattern.test(file.filename);
    if (entry.historical_phase_shaped !== phaseShaped) {
      throw new Error(
        `${manifestFile} version ${formatVersion(file.version)} historical_phase_shaped must match filename ${file.filename}`,
      );
    }
    if (file.version > manifest.immutable_through_version && phaseShaped) {
      throw new Error(
        `${file.filename} uses a phase-shaped name after immutable historical migrations; use an owner- or behavior-shaped migration name`,
      );
    }
  }

  return {
    manifestFile,
    migrationCount: files.length,
    immutableThroughVersion: manifest.immutable_through_version,
  };
}

function validateManifestShape(manifest, label) {
  assertObjectKeys(manifest, manifestKeys, label);
  assertRequiredKeys(manifest, manifestKeys, label);
  requireSchemaID(manifest, migrationHistorySchemaID, label);
  requireRepoRelativePath(manifest.migration_root, `${label}.migration_root`);
  requireInteger(manifest.immutable_through_version, `${label}.immutable_through_version`, {
    min: 0,
  });

  const entries = requireObjectArray(manifest.entries, `${label}.entries`, {
    nonEmpty: true,
  });
  const versions = [];
  const filenames = [];
  let previousVersion = 0;
  for (const [index, entry] of entries.entries()) {
    validateEntryShape(entry, `${label}.entries[${index + 1}]`);
    if (entry.version !== previousVersion + 1) {
      throw new Error(
        `${label}.entries must be contiguous and sorted by version; expected ${previousVersion + 1}, got ${entry.version}`,
      );
    }
    previousVersion = entry.version;
    versions.push(entry.version);
    filenames.push(entry.filename);
  }
  assertUnique(versions, `${label}.entries.version`);
  assertUnique(filenames, `${label}.entries.filename`);
  if (manifest.immutable_through_version > entries.length) {
    throw new Error(
      `${label}.immutable_through_version cannot exceed entries length`,
    );
  }
}

function validateEntryShape(entry, label) {
  assertObjectKeys(entry, entryKeys, label);
  assertRequiredKeys(entry, entryKeys, label);
  const version = requireInteger(entry.version, `${label}.version`, { min: 1 });
  const filename = requireString(entry.filename, `${label}.filename`, {
    pattern: migrationFilenamePattern,
  });
  const parsedVersion = versionFromFilename(filename);
  if (parsedVersion !== version) {
    throw new Error(
      `${label}.filename prefix must match version ${formatVersion(version)}`,
    );
  }
  requireString(entry.sha256, `${label}.sha256`, { pattern: sha256Pattern });
  requireBoolean(entry.historical_phase_shaped, `${label}.historical_phase_shaped`);
}

function collectMigrationFiles(migrationDir) {
  return readdirSync(migrationDir)
    .filter((filename) => filename.endsWith(".sql"))
    .sort((left, right) => left.localeCompare(right))
    .map((filename) => {
      if (!migrationFilenamePattern.test(filename)) {
        throw new Error(
          `${filename} must use exact 5-digit goose prefix and lower_snake_case name`,
        );
      }
      const file = path.join(migrationDir, filename);
      const sql = readFileSync(file, "utf8");
      return {
        filename,
        sql,
        version: versionFromFilename(filename),
        sha256: sha256(sql),
      };
    });
}

function validateGooseMarkers(file) {
  if (!/^-- \+goose Up$/mu.test(file.sql)) {
    throw new Error(`${file.filename} must contain a '-- +goose Up' marker`);
  }
  if (!/^-- \+goose Down$/mu.test(file.sql)) {
    throw new Error(`${file.filename} must contain a '-- +goose Down' marker`);
  }
}

function versionFromFilename(filename) {
  const match = migrationFilenamePattern.exec(filename);
  if (!match) {
    throw new Error(`${filename} is not a migration filename`);
  }
  return Number.parseInt(match[1], 10);
}

function formatVersion(version) {
  return String(version).padStart(5, "0");
}

function sha256(value) {
  return createHash("sha256").update(value, "utf8").digest("hex");
}
