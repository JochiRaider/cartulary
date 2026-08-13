import { readFileSync, readdirSync } from "node:fs";
import path from "node:path";

import {
  assertObjectKeys,
  assertRequiredKeys,
  assertUnique,
  readJsonObject,
  requireArray,
  requireBoolean,
  requireEnum,
  requireInteger,
  requireObject,
  requireRepoRelativePath,
  requireSchemaID,
  requireString,
  requireStringArray,
} from "../../contract/json-shape.mjs";

const schemaID = "cartulary.schema_object_ownership_manifest.v2";
const ownerByVersion = new Map([
  [1, "database_migrations"], [2, "auth"], [3, "incidents"], [4, "recovery"],
  [5, "deployment_admin"], [6, "platform_jobs"], [7, "records"], [8, "revisions"],
  [9, "parties"], [10, "timeline"], [11, "entities"], [12, "indicators"],
  [13, "assessments"], [14, "links"], [15, "tasksdecisions"], [16, "artifacts"],
  [17, "evidence"], [18, "savedviews"], [19, "imports"], [20, "networkflow"],
  [21, "projections"], [22, "graphprojection"], [23, "reporting"],
  [24, "reportcomposition"], [25, "incidentbundles"], [26, "reference_data"],
  [27, "extensions"], [28, "audit"], [29, "collaboration"], [30, "evidence"],
  [31, "evidence"],
]);
const manifestKeys = new Set([
  "schema_id", "migration_root", "supported_postgres_major", "application_schemas",
  "goose_ledger", "lineage_relation", "allowed_owners", "entries",
]);
const entryKeys = new Set([
  "object_id", "object_kind", "qualified_name", "management_class", "source_owner",
  "migration_version", "migration_file", "dependency_object_ids", "runtime_access_class",
  "recovery_access_class", "extension_profile_id", "recovery_classification", "sqlc_input",
  "foreign_key_index_status", "approval",
]);
const approvalKeys = new Set([
  "source_owner", "rationale", "expected_access_pattern", "approved_by_requirement",
]);
const objectKinds = new Set([
  "schema", "extension", "table", "view", "sequence", "type", "domain", "routine",
  "trigger", "constraint", "index", "operator", "operator_class", "operator_family",
  "cast", "collation", "migration_metadata",
]);
const managementClasses = new Set([
  "cartulary_authored", "goose_managed", "extension_managed", "administrator_managed",
]);
const runtimeClasses = new Set([
  "schema_usage", "table_read_write", "table_append_only", "table_read_only",
  "table_no_access", "migration_ledger_read", "sequence_use", "sequence_no_access",
  "view_read_only", "routine_application", "routine_private", "type_use", "type_no_access",
  "not_applicable",
]);
const recoveryAccessClasses = new Set([
  "schema_usage", "table_restore", "table_read_only", "table_no_access",
  "migration_ledger_read", "sequence_restore", "sequence_no_access", "view_read_only",
  "routine_recovery", "routine_private", "type_use", "type_no_access", "not_applicable",
]);
const profileIDs = new Set([
  "enterprise_authentication", "import", "incident_portability", "network_flow_activity",
  "reference_pack", "snapshot_reporting",
]);
const recoveryClasses = new Set([
  "authoritative_required", "excluded_rebuildable", "excluded_security_state",
  "excluded_recovery_metadata", "excluded_transient", "not_applicable",
]);
const fkStatuses = new Set(["covered", "intentionally_unindexed", "not_applicable"]);
const ownerPattern = /^[a-z][a-z0-9_]*$/u;
const objectIDPattern = /^[a-z][a-z0-9_.-]*$/u;
const requirementPattern = /^[A-Z][A-Z0-9-]*-[0-9]+$/u;

export function validateSchemaObjectOwnershipManifestShape(file) {
  const manifest = readJsonObject(file, file);
  validateManifestShape(manifest, file, canonicalMigrationAllocation(file));
  return manifest;
}

export function validateSchemaObjectOwnership(root) {
  const manifestFile = path.join(root, "tools/schema_object_ownership_manifest.json");
  const manifest = validateSchemaObjectOwnershipManifestShape(manifestFile);
  const migrationAllocation = canonicalMigrationAllocation(manifestFile);
  const migrationDir = path.join(root, manifest.migration_root);
  const actualFiles = readdirSync(migrationDir)
    .filter((filename) => /^\d{5}_.+\.sql$/u.test(filename))
    .sort((left, right) => left.localeCompare(right));
  const allocatedFiles = [...migrationAllocation.values()];
  if (JSON.stringify(actualFiles) !== JSON.stringify(allocatedFiles)) {
    throw new Error(`${migrationDir} must match the canonical migration-history allocation`);
  }

  const entryByID = new Map(manifest.entries.map((entry) => [entry.object_id, entry]));
  const authoredByIdentity = new Map(
    manifest.entries
      .filter((entry) => entry.management_class === "cartulary_authored")
      .map((entry) => [`${entry.object_kind}:${entry.qualified_name}`, entry]),
  );
  const observed = collectFinalAuthoredObjects(migrationDir, actualFiles);
  const missing = [...observed.keys()].filter((identity) => !authoredByIdentity.has(identity));
  const stale = [...authoredByIdentity.keys()].filter((identity) => !observed.has(identity));
  if (missing.length > 0 || stale.length > 0) {
    throw new Error(
      `${manifestFile} authored-object parity failed; missing=[${missing.sort().join(", ")}], stale=[${stale.sort().join(", ")}]`,
    );
  }

  for (const [identity, allocation] of observed) {
    const entry = authoredByIdentity.get(identity);
    if (entry.migration_version !== allocation.version || entry.migration_file !== allocation.filename) {
      throw new Error(
        `${manifestFile} allocates ${identity} to ${entry.migration_file}, expected ${allocation.filename}`,
      );
    }
  }
  for (const entry of manifest.entries) {
    for (const dependency of entry.dependency_object_ids) {
      const dependencyEntry = entryByID.get(dependency);
      if (!dependencyEntry) {
        throw new Error(`${manifestFile} entry ${entry.object_id} references unknown dependency ${dependency}`);
      }
      if (dependency === entry.object_id) {
        throw new Error(`${manifestFile} entry ${entry.object_id} depends on itself`);
      }
      if (
        entry.migration_version !== null &&
        dependencyEntry.migration_version !== null &&
        dependencyEntry.migration_version > entry.migration_version
      ) {
        throw new Error(`${manifestFile} entry ${entry.object_id} has forward dependency ${dependency}`);
      }
    }
  }
  for (const version of ownerByVersion.keys()) {
    if (!manifest.entries.some((entry) => entry.migration_version === version)) {
      throw new Error(`${manifestFile} has no object allocated to migration ${version}`);
    }
  }
  return { manifestFile, objectCount: observed.size, entryCount: manifest.entries.length };
}

function validateManifestShape(manifest, label, migrationAllocation) {
  assertObjectKeys(manifest, manifestKeys, label);
  assertRequiredKeys(manifest, manifestKeys, label);
  requireSchemaID(manifest, schemaID, label);
  requireRepoRelativePath(manifest.migration_root, `${label}.migration_root`);
  if (requireInteger(manifest.supported_postgres_major, `${label}.supported_postgres_major`) !== 16) {
    throw new Error(`${label}.supported_postgres_major must be 16`);
  }
  const schemas = requireStringArray(manifest.application_schemas, `${label}.application_schemas`, { nonEmpty: true });
  if (schemas.length !== 1 || schemas[0] !== "public") throw new Error(`${label}.application_schemas must be exactly public`);
  if (requireString(manifest.goose_ledger, `${label}.goose_ledger`) !== "public.goose_db_version") {
    throw new Error(`${label}.goose_ledger must be public.goose_db_version`);
  }
  if (requireString(manifest.lineage_relation, `${label}.lineage_relation`) !== "public.schema_migration_lineage") {
    throw new Error(`${label}.lineage_relation must be public.schema_migration_lineage`);
  }
  const owners = requireStringArray(manifest.allowed_owners, `${label}.allowed_owners`, { nonEmpty: true, pattern: ownerPattern });
  assertUnique(owners, `${label}.allowed_owners`);
  const ownerSet = new Set(owners);
  const entries = requireArray(manifest.entries, `${label}.entries`, { nonEmpty: true });
  const objectIDs = [];
  const identities = [];
  for (const [index, rawEntry] of entries.entries()) {
    const entry = requireObject(rawEntry, `${label}.entries[${index}]`);
    validateEntry(entry, `${label}.entries[${index}]`, ownerSet, migrationAllocation);
    objectIDs.push(entry.object_id);
    identities.push(`${entry.object_kind}:${entry.qualified_name}`);
  }
  assertUnique(objectIDs, `${label}.entries.object_id`);
  assertUnique(identities, `${label}.entries object kind/name identity`);
}

function validateEntry(entry, label, ownerSet, migrationAllocation) {
  assertObjectKeys(entry, entryKeys, label);
  assertRequiredKeys(entry, entryKeys, label);
  requireString(entry.object_id, `${label}.object_id`, { pattern: objectIDPattern });
  requireEnum(entry.object_kind, `${label}.object_kind`, objectKinds);
  const qualifiedName = requireString(entry.qualified_name, `${label}.qualified_name`);
  if (entry.object_kind !== "schema" && !qualifiedName.includes(".")) {
    throw new Error(`${label}.qualified_name must be schema-qualified`);
  }
  const management = requireEnum(entry.management_class, `${label}.management_class`, managementClasses);
  const sourceOwner = requireString(entry.source_owner, `${label}.source_owner`, { pattern: ownerPattern });
  if (!ownerSet.has(sourceOwner)) throw new Error(`${label}.source_owner is not allowed`);
  if (management === "cartulary_authored") {
    const version = requireInteger(entry.migration_version, `${label}.migration_version`, { min: 1 });
    if (entry.migration_file !== migrationAllocation.get(version)) {
      throw new Error(`${label} must use the canonical migration version/file allocation`);
    }
    const allocatedOwner = ownerByVersion.get(version);
    if (sourceOwner !== allocatedOwner && entry.object_kind !== "table") {
      throw new Error(`${label}.source_owner must be ${allocatedOwner} for migration ${version}`);
    }
  } else if (entry.migration_version !== null || entry.migration_file !== null) {
    throw new Error(`${label} non-authored objects must not have a migration allocation`);
  }
  const dependencies = requireStringArray(entry.dependency_object_ids, `${label}.dependency_object_ids`, { pattern: objectIDPattern });
  assertUnique(dependencies, `${label}.dependency_object_ids`);
  if (JSON.stringify(dependencies) !== JSON.stringify([...dependencies].sort())) {
    throw new Error(`${label}.dependency_object_ids must be sorted`);
  }
  requireEnum(entry.runtime_access_class, `${label}.runtime_access_class`, runtimeClasses);
  requireEnum(entry.recovery_access_class, `${label}.recovery_access_class`, recoveryAccessClasses);
  if (entry.extension_profile_id !== null) requireEnum(entry.extension_profile_id, `${label}.extension_profile_id`, profileIDs);
  requireEnum(entry.recovery_classification, `${label}.recovery_classification`, recoveryClasses);
  requireBoolean(entry.sqlc_input, `${label}.sqlc_input`);
  const fkStatus = requireEnum(entry.foreign_key_index_status, `${label}.foreign_key_index_status`, fkStatuses);
  if (fkStatus === "intentionally_unindexed") validateApproval(entry.approval, `${label}.approval`, sourceOwner);
  else if (entry.approval !== null) throw new Error(`${label}.approval must be null unless the foreign key is intentionally unindexed`);
}

function canonicalMigrationAllocation(ownershipManifestFile) {
  const historyFile = path.join(path.dirname(ownershipManifestFile), "migration_history_manifest.json");
  const history = readJsonObject(historyFile, historyFile);
  const entries = requireArray(history.entries, `${historyFile}.entries`, { nonEmpty: true });
  const allocation = new Map();
  for (const [index, rawEntry] of entries.entries()) {
    const entry = requireObject(rawEntry, `${historyFile}.entries[${index}]`);
    const version = requireInteger(entry.version, `${historyFile}.entries[${index}].version`, { min: 1 });
    const filename = requireString(entry.filename, `${historyFile}.entries[${index}].filename`);
    if (version !== index + 1 || allocation.has(version)) {
      throw new Error(`${historyFile}.entries must be uniquely contiguous by version`);
    }
    allocation.set(version, filename);
  }
  return allocation;
}

function validateApproval(rawApproval, label, sourceOwner) {
  const approval = requireObject(rawApproval, label);
  assertObjectKeys(approval, approvalKeys, label);
  assertRequiredKeys(approval, approvalKeys, label);
  if (requireString(approval.source_owner, `${label}.source_owner`, { pattern: ownerPattern }) !== sourceOwner) {
    throw new Error(`${label}.source_owner must match the entry source owner`);
  }
  requireString(approval.rationale, `${label}.rationale`);
  requireString(approval.expected_access_pattern, `${label}.expected_access_pattern`);
  requireString(approval.approved_by_requirement, `${label}.approved_by_requirement`, { pattern: requirementPattern });
}

function collectFinalAuthoredObjects(migrationDir, filenames) {
  const objects = new Map();
  for (const [index, filename] of filenames.entries()) {
    const source = readFileSync(path.join(migrationDir, filename), "utf8").split(/^-- \+goose Down\s*$/mu, 1)[0];
    collectEvents(source, index + 1, filename, objects);
  }
  return objects;
}

function collectEvents(source, version, filename, objects) {
  scanEvents(source, /\b(CREATE|DROP)\s+TABLE\s+(?:IF\s+(?:NOT\s+)?EXISTS\s+)?(?:ONLY\s+)?([A-Za-z_][A-Za-z0-9_.]*)/giu, "table", version, filename, objects);
  scanEvents(source, /\b(CREATE(?:\s+OR\s+REPLACE)?|DROP)\s+(?:MATERIALIZED\s+)?VIEW\s+(?:IF\s+EXISTS\s+)?([A-Za-z_][A-Za-z0-9_.]*)/giu, "view", version, filename, objects);
  scanEvents(source, /\b(CREATE|DROP)\s+SEQUENCE\s+(?:IF\s+(?:NOT\s+)?EXISTS\s+)?([A-Za-z_][A-Za-z0-9_.]*)/giu, "sequence", version, filename, objects);
  scanEvents(source, /\b(CREATE(?:\s+UNIQUE)?|DROP)\s+INDEX\s+(?:CONCURRENTLY\s+)?(?:IF\s+(?:NOT\s+)?EXISTS\s+)?([A-Za-z_][A-Za-z0-9_.]*)/giu, "index", version, filename, objects);
  scanEvents(source, /\b(CREATE(?:\s+OR\s+REPLACE)?|DROP)\s+FUNCTION\s+(?:IF\s+EXISTS\s+)?([A-Za-z_][A-Za-z0-9_.]*)\s*\(/giu, "routine", version, filename, objects);
  scanTriggerEvents(source, version, filename, objects);
  scanConstraintEvents(source, version, filename, objects);
}

function scanEvents(source, regex, kind, version, filename, objects) {
  for (const match of source.matchAll(regex)) {
    const name = qualify(match[2]);
    const identity = `${kind}:${name}`;
    if (match[1].toUpperCase() === "DROP") objects.delete(identity);
    else objects.set(identity, { version, filename });
  }
}

function scanTriggerEvents(source, version, filename, objects) {
  const regex = /\b(CREATE|DROP)\s+TRIGGER\s+(?:IF\s+EXISTS\s+)?([A-Za-z_][A-Za-z0-9_]*)[\s\S]{0,500}?\s+ON\s+([A-Za-z_][A-Za-z0-9_.]*)/giu;
  for (const match of source.matchAll(regex)) {
    const identity = `trigger:${qualify(match[3])}.${match[2]}`;
    if (match[1].toUpperCase() === "DROP") objects.delete(identity);
    else objects.set(identity, { version, filename });
  }
}

function scanConstraintEvents(source, version, filename, objects) {
  const regex = /\b(ADD|DROP)?\s*CONSTRAINT\s+(?:IF\s+EXISTS\s+)?([A-Za-z_][A-Za-z0-9_]*)(?:\s+(PRIMARY\s+KEY|UNIQUE))?/giu;
  for (const match of source.matchAll(regex)) {
    const prefix = source.slice(0, match.index ?? 0);
    const tableMatches = [...prefix.matchAll(/\b(?:ALTER\s+TABLE(?:\s+ONLY)?|CREATE\s+TABLE)\s+([A-Za-z_][A-Za-z0-9_.]*)/giu)];
    const table = tableMatches.length > 0 ? qualify(tableMatches.at(-1)[1]) : "public.unknown";
    const identity = `constraint:${table}.${match[2]}`;
    const indexIdentity = `index:${qualify(match[2])}`;
    if ((match[1] ?? "").toUpperCase() === "DROP") {
      objects.delete(identity);
      objects.delete(indexIdentity);
    } else {
      objects.set(identity, { version, filename });
      if (match[3]) objects.set(indexIdentity, { version, filename });
    }
  }
}

function qualify(value) {
  const normalized = value.replaceAll('"', "").trim().toLowerCase();
  return normalized.includes(".") ? normalized : `public.${normalized}`;
}
