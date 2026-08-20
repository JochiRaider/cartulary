import { createHash } from "node:crypto";
import { readFileSync, readdirSync, writeFileSync } from "node:fs";
import path from "node:path";

const root = process.cwd();
const migrationRoot = path.join(root, "db/migrations");
const migrationFiles = readdirSync(migrationRoot)
  .filter((name) => /^\d{5}_.+\.sql$/u.test(name))
  .sort((left, right) => left.localeCompare(right));

const immutableBaselineVersion = 29;
if (migrationFiles.length < immutableBaselineVersion) {
  throw new Error(
    `Production DDL Rebaseline v2 requires at least ${immutableBaselineVersion} migrations; found ${migrationFiles.length}`,
  );
}
for (const [index, filename] of migrationFiles.entries()) {
  const version = Number.parseInt(filename.slice(0, 5), 10);
  if (version !== index + 1) {
    throw new Error(`migration catalog must be contiguous at version ${index + 1}; found ${filename}`);
  }
  validateCatalogPolicy(filename, readFileSync(path.join(migrationRoot, filename), "utf8"));
}

const history = {
  schema_id: "cartulary.migration_history_manifest.v1",
  migration_root: "db/migrations",
  immutable_through_version: immutableBaselineVersion,
  entries: migrationFiles.map((filename, index) => {
    const body = readFileSync(path.join(migrationRoot, filename));
    return {
      version: index + 1,
      filename,
      sha256: createHash("sha256").update(body).digest("hex"),
      historical_phase_shaped: false,
    };
  }),
};
writeJSON("tools/migration_history_manifest.json", history);

const ownersByVersion = new Map([
  [1, "database_migrations"], [2, "auth"], [3, "incidents"], [4, "recovery"],
  [5, "deployment_admin"], [6, "platform_jobs"], [7, "records"], [8, "revisions"],
  [9, "parties"], [10, "timeline"], [11, "entities"], [12, "indicators"],
  [13, "assessments"], [14, "links"], [15, "tasksdecisions"], [16, "artifacts"],
  [17, "evidence"], [18, "savedviews"], [19, "imports"], [20, "networkflow"],
  [21, "projections"], [22, "graphprojection"], [23, "reporting"],
  [24, "reportcomposition"], [25, "incidentbundles"], [26, "reference_data"],
  [27, "extensions"], [28, "audit"], [29, "collaboration"],
  [30, "evidence"], [31, "evidence"], [32, "networkflow"],
  [33, "networkflow"], [34, "graphprojection"], [35, "assessments"],
]);
for (const [index, filename] of migrationFiles.entries()) {
  const version = index + 1;
  if (!ownersByVersion.has(version)) {
    throw new Error(`migration ${filename} has no source-owner assignment`);
  }
}
const profileByVersion = new Map([
  [2, "enterprise_authentication"],
  [19, "import"],
  [20, "network_flow_activity"],
  [23, "snapshot_reporting"],
  [24, "snapshot_reporting"],
  [25, "incident_portability"],
  [26, "reference_pack"],
  [32, "network_flow_activity"],
  [33, "network_flow_activity"],
]);
const recoveryCatalog = JSON.parse(
  readFileSync(path.join(root, "contracts/recovery/fixtures/recovery-state-catalog.v1.json"), "utf8"),
);
const recoveryByTable = new Map(
  recoveryCatalog.tables.map((entry) => [entry.table_name, entry]),
);
const queryText = readdirSync(path.join(root, "db/queries"), { recursive: true })
  .filter((entry) => typeof entry === "string" && entry.endsWith(".sql"))
  .map((entry) => readFileSync(path.join(root, "db/queries", entry), "utf8"))
  .join("\n");

const objects = new Map();
for (const [index, filename] of migrationFiles.entries()) {
  const version = index + 1;
  const source = readFileSync(path.join(migrationRoot, filename), "utf8").split(/^-- \+goose Down\s*$/mu, 1)[0];
  collectRelationEvents(source, version, filename);
  collectIndexEvents(source, version, filename);
  collectRoutineEvents(source, version, filename);
  collectTriggerEvents(source, version, filename);
  collectConstraintEvents(source, version, filename);
}

const entries = [
  externalObject("schema.public", "schema", "public", "administrator_managed", "schema_usage", "schema_usage"),
  externalObject("extension.pgcrypto", "extension", "public.pgcrypto", "administrator_managed", "type_use", "type_use"),
  externalObject("extension.citext", "extension", "public.citext", "administrator_managed", "type_use", "type_use"),
  {
    object_id: "migration-metadata.public.goose-db-version",
    object_kind: "migration_metadata",
    qualified_name: "public.goose_db_version",
    management_class: "goose_managed",
    source_owner: "database_migrations",
    migration_version: null,
    migration_file: null,
    dependency_object_ids: ["schema.public"],
    runtime_access_class: "migration_ledger_read",
    recovery_access_class: "migration_ledger_read",
    extension_profile_id: null,
    recovery_classification: "excluded_recovery_metadata",
    sqlc_input: false,
    foreign_key_index_status: "not_applicable",
    approval: null,
  },
  ...[...objects.values()]
    .sort((left, right) => left.object_id.localeCompare(right.object_id))
    .map(toManifestEntry),
];

const manifest = {
  schema_id: "cartulary.schema_object_ownership_manifest.v2",
  migration_root: "db/migrations",
  supported_postgres_major: 16,
  application_schemas: ["public"],
  goose_ledger: "public.goose_db_version",
  lineage_relation: "public.schema_migration_lineage",
  allowed_owners: [...new Set([
    ...ownersByVersion.values(),
    ...entries.map((entry) => entry.source_owner),
  ])].sort(),
  entries,
};
writeJSON("tools/schema_object_ownership_manifest.json", manifest);

function collectRelationEvents(source, version, filename) {
  const patterns = [
    ["table", /\b(CREATE|DROP)\s+TABLE\s+(?:IF\s+(?:NOT\s+)?EXISTS\s+)?(?:ONLY\s+)?([A-Za-z_][A-Za-z0-9_.]*)/giu],
    ["view", /\b(CREATE(?:\s+OR\s+REPLACE)?|DROP)\s+(?:MATERIALIZED\s+)?VIEW\s+(?:IF\s+EXISTS\s+)?([A-Za-z_][A-Za-z0-9_.]*)/giu],
    ["sequence", /\b(CREATE|DROP)\s+SEQUENCE\s+(?:IF\s+(?:NOT\s+)?EXISTS\s+)?([A-Za-z_][A-Za-z0-9_.]*)/giu],
  ];
  for (const [kind, regex] of patterns) {
    for (const match of source.matchAll(regex)) {
      const operation = match[1].toUpperCase();
      const name = qualify(match[2]);
      const key = `${kind}:${name}`;
      if (operation === "DROP") removeObjectAndDependents(key);
      else setObject(key, kind, name, version, filename, match.index ?? 0, source);
    }
  }
}

function removeObjectAndDependents(key) {
  const removed = new Set([key]);
  let changed = true;
  while (changed) {
    changed = false;
    for (const [candidateKey, candidate] of objects) {
      if (!removed.has(candidateKey) && candidate.dependencies.some((dependency) => removed.has(dependency))) {
        removed.add(candidateKey);
        changed = true;
      }
    }
  }
  for (const removedKey of removed) objects.delete(removedKey);
}

function collectIndexEvents(source, version, filename) {
  const regex = /\b(CREATE(?:\s+UNIQUE)?|DROP)\s+INDEX\s+(?:CONCURRENTLY\s+)?(?:IF\s+(?:NOT\s+)?EXISTS\s+)?([A-Za-z_][A-Za-z0-9_.]*)(?:\s+ON\s+(?:ONLY\s+)?([A-Za-z_][A-Za-z0-9_.]*))?/giu;
  for (const match of source.matchAll(regex)) {
    const name = qualify(match[2]);
    const key = `index:${name}`;
    if (match[1].toUpperCase() === "DROP") {
      objects.delete(key);
      continue;
    }
    const dependency = match[3] ? `table:${qualify(match[3])}` : null;
    setObject(key, "index", name, version, filename, match.index ?? 0, source, dependency ? [dependency] : []);
  }
}

function collectRoutineEvents(source, version, filename) {
  const regex = /\b(CREATE(?:\s+OR\s+REPLACE)?|DROP)\s+FUNCTION\s+(?:IF\s+EXISTS\s+)?([A-Za-z_][A-Za-z0-9_.]*)\s*\(/giu;
  for (const match of source.matchAll(regex)) {
    const name = qualify(match[2]);
    const key = `routine:${name}`;
    if (match[1].toUpperCase() === "DROP") objects.delete(key);
    else setObject(key, "routine", name, version, filename, match.index ?? 0, source);
  }
}

function collectTriggerEvents(source, version, filename) {
  const regex = /\b(CREATE|DROP)\s+TRIGGER\s+(?:IF\s+EXISTS\s+)?([A-Za-z_][A-Za-z0-9_]*)[\s\S]{0,500}?\s+ON\s+([A-Za-z_][A-Za-z0-9_.]*)/giu;
  for (const match of source.matchAll(regex)) {
    const table = match[3] ? qualify(match[3]) : "public.unknown";
    const name = `${table}.${match[2]}`;
    const key = `trigger:${name}`;
    if (match[1].toUpperCase() === "DROP") objects.delete(key);
    else setObject(key, "trigger", name, version, filename, match.index ?? 0, source, [`table:${table}`]);
  }
}

function collectConstraintEvents(source, version, filename) {
  const regex = /\b(ADD|DROP)?\s*CONSTRAINT\s+(?:IF\s+EXISTS\s+)?([A-Za-z_][A-Za-z0-9_]*)/giu;
  for (const match of source.matchAll(regex)) {
    const prefix = source.slice(0, match.index ?? 0);
    const tableMatches = [...prefix.matchAll(/\b(?:ALTER\s+TABLE(?:\s+ONLY)?|CREATE\s+TABLE)\s+([A-Za-z_][A-Za-z0-9_.]*)/giu)];
    const table = tableMatches.length > 0 ? qualify(tableMatches.at(-1)[1]) : "public.unknown";
    const name = `${table}.${match[2]}`;
    const key = `constraint:${name}`;
    if ((match[1] ?? "").toUpperCase() === "DROP") {
      objects.delete(key);
      continue;
    }
    const definition = source.slice((match.index ?? 0) + match[0].length, (match.index ?? 0) + match[0].length + 1200);
    const tableForeignKey = /^\s+FOREIGN\s+KEY\s*\(([^)]+)\)\s+REFERENCES\s+([A-Za-z_][A-Za-z0-9_.]*)/iu.exec(definition);
    const columnForeignKey = /^\s+REFERENCES\s+([A-Za-z_][A-Za-z0-9_.]*)/iu.exec(definition);
    const referencedTable = tableForeignKey?.[2] ?? columnForeignKey?.[1];
    const foreignKey = Boolean(referencedTable);
    const dependencies = [`table:${table}`];
    if (referencedTable) dependencies.push(`table:${qualify(referencedTable)}`);
    setObject(key, "constraint", name, version, filename, match.index ?? 0, source, dependencies, foreignKey);
    if (/^\s+(?:PRIMARY\s+KEY|UNIQUE)\b/iu.test(definition)) {
      setObject(
        `index:${qualify(match[2])}`,
        "index",
        qualify(match[2]),
        version,
        filename,
        match.index ?? 0,
        source,
        [`table:${table}`, key],
      );
    }
  }
}

function validateCatalogPolicy(filename, body) {
  const violations = [];
  const upMarkers = body.match(/^-- \+goose Up\s*$/gmu) ?? [];
  const downMarkers = body.match(/^-- \+goose Down\s*$/gmu) ?? [];
  if (upMarkers.length !== 1 || downMarkers.length !== 1 || body.indexOf("-- +goose Up") > body.indexOf("-- +goose Down")) {
    violations.push("marker_pair");
  }
  const upper = body.toUpperCase();
  for (const [code, pattern] of [
    ["no_transaction", /--\s*\+GOOSE\s+NO\s+TRANSACTION/u],
    ["permissive_create", /\bIF\s+NOT\s+EXISTS\b/u],
    ["extension_install", /\bCREATE\s+EXTENSION\b/u],
    ["replacement_ddl", /\bCREATE\s+OR\s+REPLACE\b/u],
    ["unvalidated_constraint", /\bNOT\s+VALID\b/u],
    ["concurrent_index", /\bCREATE\s+(?:UNIQUE\s+)?INDEX\s+CONCURRENTLY\b/u],
    ["cascade_drop", /\bDROP\b[^;]*\bCASCADE\b/u],
    ["historical_lineage", /CARTULARY\.PROD_DDL_REBASELINE\.V1|00062_|MIGRATION\s+62/u],
  ]) {
    if (pattern.test(upper)) violations.push(code);
  }

  const up = body.split(/^-- \+goose Down\s*$/mu, 1)[0];
  for (const match of up.matchAll(/\b(?:CREATE|ALTER|DROP)\s+(?:MATERIALIZED\s+)?(?:TABLE|VIEW|SEQUENCE|FUNCTION)\s+(?:ONLY\s+)?(?:IF\s+EXISTS\s+)?([A-Za-z_][A-Za-z0-9_.]*)/giu)) {
    if (!match[1].includes(".")) violations.push("unqualified_object");
  }
  for (const match of up.matchAll(/\bREFERENCES\s+([A-Za-z_][A-Za-z0-9_.]*)\s*\([^)]*\)([\s\S]{0,160})/giu)) {
    if (!match[1].includes(".")) violations.push("unqualified_reference");
    const actionText = match[2].split(/[,;]/u, 1)[0].toUpperCase();
    if (!actionText.includes("ON UPDATE ") || !actionText.includes("ON DELETE ")) violations.push("foreign_key_actions");
  }
  const lines = up.split(/\r?\n/u);
  for (const [index, line] of lines.entries()) {
    if (line.trimStart().startsWith("--")) continue;
    if (!/\b(?:PRIMARY\s+KEY|UNIQUE)\b/iu.test(line) || /CREATE\s+UNIQUE\s+INDEX/iu.test(line)) continue;
    const previous = index > 0 ? lines[index - 1] : "";
    if (!/\bCONSTRAINT\s+[A-Za-z_][A-Za-z0-9_]*\b/iu.test(`${previous}\n${line}`)) violations.push("unnamed_constraint");
  }

  const routinePattern = /CREATE\s+FUNCTION\s+(public\.([A-Za-z_][A-Za-z0-9_]*))\s*\([\s\S]*?\$\$;/giu;
  for (const routine of up.matchAll(routinePattern)) {
    const definition = routine[0];
    const routineName = routine[2];
    if (!/SET\s+search_path\s*=\s*pg_catalog\s*,\s*public/iu.test(definition)) violations.push("routine_search_path");
    if (!new RegExp(`REVOKE\\s+(?:ALL|EXECUTE)\\s+ON\\s+FUNCTION\\s+public\\.${escapeRegExp(routineName)}\\s*\\(`, "iu").test(up)) {
      violations.push("routine_public_execute");
    }
    if (/\bSECURITY\s+DEFINER\b/iu.test(definition) && !new Set([
      "revisions_incident_bundle_sequence_begin_v1",
      "revisions_incident_bundle_sequence_finish_v1",
    ]).has(routineName)) {
      violations.push("routine_security_class");
    }
    if (/\bEXECUTE\b/iu.test(definition) && (!/\bSECURITY\s+DEFINER\b/iu.test(definition) || !/EXECUTE\s+pg_catalog\.format\s*\(/iu.test(definition))) {
      violations.push("routine_dynamic_sql");
    }
  }
  if (violations.length > 0) {
    throw new Error(`${filename} violates Production DDL Rebaseline v2 policy: ${[...new Set(violations)].sort().join(", ")}`);
  }
}

function setObject(key, kind, name, version, filename, position, source, dependencies = [], foreignKey = false) {
  const bareName = name.split(".").at(-1);
  const owner = version === 32 && bareName.startsWith("graph_projection_")
    ? "graphprojection"
    : ownersByVersion.get(version);
  objects.set(key, {
    object_id: objectID(kind, name),
    object_kind: kind,
    qualified_name: name,
    version,
    filename,
    owner,
    profile: profileByVersion.get(version) ?? null,
    dependencies,
    foreignKey,
    sourceFragment: source.slice(position, position + 1500),
  });
}

function toManifestEntry(object) {
  const bareName = object.qualified_name.split(".").at(-1);
  const tableRecovery = object.object_kind === "table" ? recoveryByTable.get(bareName) : null;
  const recoveryClass = tableRecovery?.backup_inclusion ?? "not_applicable";
  const logicalOwner = tableRecovery?.owner_id?.replace(/^module\./u, "") ?? object.owner;
  const runtimeNoAccessTables = new Set([
    "backup_sets",
    "operator_recovery_journal",
    "restore_verification_runs",
  ]);
  const runtimeAppendOnlyTables = new Set([
    "administrative_audit_projections",
    "deployment_admin_audit_events",
  ]);
  const runtimeApplicationRoutines = new Set([
    "cartulary_confidence_band",
    "change_set_mutations_history_ids_are_canonical",
    "enforce_indicator_support_ref_incident",
    "indicator_support_refs_are_valid",
    "network_flow_reject_immutable_update",
    "reject_administrative_audit_mutation",
    "revisions_incident_bundle_sequence_begin_v1",
    "revisions_incident_bundle_sequence_finish_v1",
    "sync_indicator_active_identity_from_indicator",
    "sync_indicator_active_identity_from_record",
  ]);
  const recoveryRoutines = new Set([
    ...runtimeApplicationRoutines,
    "indicator_active_identities_are_valid",
    "rebuild_indicator_active_identities",
  ]);
  const runtimeClass = {
    table: bareName === "schema_migration_lineage"
      ? "migration_ledger_read"
      : runtimeNoAccessTables.has(bareName)
        ? "table_no_access"
        : runtimeAppendOnlyTables.has(bareName)
          ? "table_append_only"
          : "table_read_write",
    view: "view_read_only",
    sequence: "sequence_use",
    routine: runtimeApplicationRoutines.has(bareName) ? "routine_application" : "routine_private",
  }[object.object_kind] ?? "not_applicable";
  const recoveryAccessClass = {
    table: bareName === "schema_migration_lineage" ? "migration_ledger_read" : "table_restore",
    view: "view_read_only",
    sequence: "sequence_restore",
    routine: recoveryRoutines.has(bareName) ? "routine_recovery" : "routine_private",
  }[object.object_kind] ?? "not_applicable";
  const dependencyIDs = [...new Set(
    object.dependencies
      .map((dependency) => objects.get(dependency)?.object_id)
      .filter((dependency) => dependency && dependency !== object.object_id),
  )].sort();
  return {
    object_id: object.object_id,
    object_kind: object.object_kind,
    qualified_name: object.qualified_name,
    management_class: "cartulary_authored",
    source_owner: logicalOwner,
    migration_version: object.version,
    migration_file: object.filename,
    dependency_object_ids: dependencyIDs,
    runtime_access_class: runtimeClass,
    recovery_access_class: recoveryAccessClass,
    extension_profile_id: object.profile,
    recovery_classification: normalizeRecoveryClass(recoveryClass),
    sqlc_input: new RegExp(`\\b${escapeRegExp(bareName)}\\b`, "u").test(queryText),
    foreign_key_index_status: object.foreignKey ? "covered" : "not_applicable",
    approval: null,
  };
}

function externalObject(id, kind, name, management, runtimeClass, recoveryClass) {
  return {
    object_id: id,
    object_kind: kind,
    qualified_name: name,
    management_class: management,
    source_owner: "database_migrations",
    migration_version: null,
    migration_file: null,
    dependency_object_ids: [],
    runtime_access_class: runtimeClass,
    recovery_access_class: recoveryClass,
    extension_profile_id: null,
    recovery_classification: "not_applicable",
    sqlc_input: false,
    foreign_key_index_status: "not_applicable",
    approval: null,
  };
}

function normalizeRecoveryClass(value) {
  if (["authoritative_required", "excluded_rebuildable", "excluded_security_state", "excluded_recovery_metadata"].includes(value)) return value;
  return "not_applicable";
}

function qualify(value) {
  const normalized = value.replaceAll('"', "").trim();
  return normalized.includes(".") ? normalized : `public.${normalized}`;
}

function objectID(kind, name) {
  return `${kind}.${name}`.toLowerCase().replaceAll("_", "-").replace(/[^a-z0-9.-]+/gu, "-");
}

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&");
}

function writeJSON(relativePath, value) {
  writeFileSync(path.join(root, relativePath), `${JSON.stringify(value, null, 2)}\n`);
}
