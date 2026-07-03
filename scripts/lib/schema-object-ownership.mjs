import { readFileSync, readdirSync } from "node:fs";
import path from "node:path";

import {
  assertObjectKeys,
  assertRequiredKeys,
  assertUnique,
  readJsonObject,
  requireObjectArray,
  requireRepoRelativePath,
  requireSchemaID,
  requireString,
  requireStringArray,
} from "../../tools/harness/generated-artifacts/json-shape.mjs";

export const schemaObjectOwnershipSchemaID =
  "cartulary.schema_object_ownership_manifest.v1";

const manifestKeys = new Set([
  "schema_id",
  "migration_root",
  "synthetic_objects",
  "allowed_owners",
  "allowed_profiles",
  "entries",
]);
const syntheticObjectKeys = new Set([
  "kind",
  "name",
  "source",
  "owner",
  "profiles",
  "downstream_surfaces",
  "notes",
]);
const entryKeys = new Set([
  "id",
  "owner",
  "profiles",
  "object_patterns",
  "downstream_surfaces",
  "notes",
]);
const objectPatternKeys = new Set(["kind", "name_pattern"]);
const objectKinds = new Set([
  "any",
  "bookkeeping",
  "constraint",
  "data_backfill",
  "extension",
  "function",
  "generated_column",
  "index",
  "policy",
  "sequence",
  "table",
  "trigger",
  "type",
  "view",
]);
const ownerPattern = /^[a-z][a-z0-9_]*$/u;
const profilePattern = /^[a-z][a-z0-9_]*$/u;
const migrationFilenamePattern = /^\d{5}_.+\.sql$/u;
const sqlIdentifierPattern = String.raw`(?:"[^"]+"|[A-Za-z_][A-Za-z0-9_]*)(?:\s*\.\s*(?:"[^"]+"|[A-Za-z_][A-Za-z0-9_]*))*`;

export function validateSchemaObjectOwnershipManifestShape(file) {
  const manifest = readJsonObject(file, file);
  validateManifestShape(manifest, file);
  return manifest;
}

export function validateSchemaObjectOwnership(root) {
  const manifestFile = path.join(
    root,
    "tools/schema_object_ownership_manifest.json",
  );
  const manifest = validateSchemaObjectOwnershipManifestShape(manifestFile);
  const objects = collectSchemaObjects(root, manifest);
  const uncovered = [];
  const ambiguous = [];
  const compiledEntries = compileEntries(manifest.entries, manifestFile);

  for (const object of objects) {
    const matches = compiledEntries.filter((entry) =>
      entry.patterns.some((pattern) => patternMatches(pattern, object)),
    );
    if (matches.length === 0) {
      uncovered.push(object);
      continue;
    }
    if (matches.length > 1) {
      ambiguous.push({ object, entries: matches.map((entry) => entry.id) });
    }
  }

  if (uncovered.length > 0) {
    const details = uncovered
      .map((object) => `${object.kind}:${object.name} (${object.sources.join(", ")})`)
      .sort()
      .join("; ");
    throw new Error(
      `${manifestFile} does not assign owners for schema objects: ${details}`,
    );
  }
  if (ambiguous.length > 0) {
    const details = ambiguous
      .map(
        ({ object, entries }) =>
          `${object.kind}:${object.name} (${object.sources.join(", ")}) matches ${entries.join(", ")}`,
      )
      .sort()
      .join("; ");
    throw new Error(
      `${manifestFile} assigns multiple owners for schema objects: ${details}`,
    );
  }

  return {
    manifestFile,
    objectCount: objects.length,
    entryCount: manifest.entries.length,
  };
}

function validateManifestShape(manifest, label) {
  assertObjectKeys(manifest, manifestKeys, label);
  assertRequiredKeys(manifest, manifestKeys, label);
  requireSchemaID(manifest, schemaObjectOwnershipSchemaID, label);
  requireRepoRelativePath(manifest.migration_root, `${label}.migration_root`);
  const owners = requireStringArray(manifest.allowed_owners, `${label}.allowed_owners`, {
    nonEmpty: true,
    pattern: ownerPattern,
  });
  const profiles = requireStringArray(
    manifest.allowed_profiles,
    `${label}.allowed_profiles`,
    {
      nonEmpty: true,
      pattern: profilePattern,
    },
  );
  const ownerSet = new Set(owners);
  const profileSet = new Set(profiles);

  const syntheticObjects = requireObjectArray(
    manifest.synthetic_objects,
    `${label}.synthetic_objects`,
  );
  for (const [index, object] of syntheticObjects.entries()) {
    validateSyntheticObjectShape(
      object,
      `${label}.synthetic_objects[${index + 1}]`,
      ownerSet,
      profileSet,
    );
  }

  const entries = requireObjectArray(manifest.entries, `${label}.entries`, {
    nonEmpty: true,
  });
  const ids = [];
  for (const [index, entry] of entries.entries()) {
    ids.push(
      validateEntryShape(
        entry,
        `${label}.entries[${index + 1}]`,
        ownerSet,
        profileSet,
      ),
    );
  }
  assertUnique(ids, `${label}.entries.id`);
}

function validateSyntheticObjectShape(object, label, ownerSet, profileSet) {
  assertObjectKeys(object, syntheticObjectKeys, label);
  assertRequiredKeys(object, syntheticObjectKeys, label);
  validateKind(object.kind, `${label}.kind`, false);
  requireString(object.name, `${label}.name`);
  requireString(object.source, `${label}.source`);
  validateOwner(object.owner, `${label}.owner`, ownerSet);
  validateProfiles(object.profiles, `${label}.profiles`, profileSet);
  requireStringArray(object.downstream_surfaces, `${label}.downstream_surfaces`, {
    nonEmpty: true,
  });
  requireString(object.notes, `${label}.notes`);
}

function validateEntryShape(entry, label, ownerSet, profileSet) {
  assertObjectKeys(entry, entryKeys, label);
  assertRequiredKeys(entry, entryKeys, label);
  const id = requireString(entry.id, `${label}.id`, {
    pattern: /^[a-z][a-z0-9_-]*$/u,
  });
  validateOwner(entry.owner, `${label}.owner`, ownerSet);
  validateProfiles(entry.profiles, `${label}.profiles`, profileSet);
  const patterns = requireObjectArray(
    entry.object_patterns,
    `${label}.object_patterns`,
    { nonEmpty: true },
  );
  for (const [index, pattern] of patterns.entries()) {
    validateObjectPatternShape(
      pattern,
      `${label}.object_patterns[${index + 1}]`,
    );
  }
  requireStringArray(entry.downstream_surfaces, `${label}.downstream_surfaces`, {
    nonEmpty: true,
  });
  requireString(entry.notes, `${label}.notes`);
  return id;
}

function validateObjectPatternShape(pattern, label) {
  assertObjectKeys(pattern, objectPatternKeys, label);
  assertRequiredKeys(pattern, objectPatternKeys, label);
  validateKind(pattern.kind, `${label}.kind`, true);
  const expression = requireString(pattern.name_pattern, `${label}.name_pattern`);
  try {
    new RegExp(expression, "u");
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`${label}.name_pattern is not a valid regex: ${message}`);
  }
}

function validateKind(value, label, allowAny) {
  const kind = requireString(value, label);
  if (!objectKinds.has(kind) || (!allowAny && kind === "any")) {
    throw new Error(`${label} must be one of ${[...objectKinds].join("|")}`);
  }
  return kind;
}

function validateOwner(value, label, ownerSet) {
  const owner = requireString(value, label, { pattern: ownerPattern });
  if (!ownerSet.has(owner)) {
    throw new Error(`${label} must be declared in allowed_owners`);
  }
}

function validateProfiles(value, label, profileSet) {
  const profiles = requireStringArray(value, label, {
    nonEmpty: true,
    pattern: profilePattern,
  });
  for (const profile of profiles) {
    if (!profileSet.has(profile)) {
      throw new Error(`${label} contains undeclared profile ${profile}`);
    }
  }
}

function compileEntries(entries, label) {
  return entries.map((entry, entryIndex) => ({
    id: entry.id,
    patterns: entry.object_patterns.map((pattern, patternIndex) => ({
      kind: pattern.kind,
      regex: compileRegex(
        pattern.name_pattern,
        `${label}.entries[${entryIndex + 1}].object_patterns[${patternIndex + 1}].name_pattern`,
      ),
    })),
  }));
}

function compileRegex(expression, label) {
  try {
    return new RegExp(expression, "u");
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`${label} is not a valid regex: ${message}`);
  }
}

function patternMatches(pattern, object) {
  return (
    (pattern.kind === "any" || pattern.kind === object.kind) &&
    pattern.regex.test(object.name)
  );
}

function collectSchemaObjects(root, manifest) {
  const migrationDir = path.join(root, manifest.migration_root);
  const objects = new Map();
  for (const filename of readdirSync(migrationDir).sort((left, right) =>
    left.localeCompare(right),
  )) {
    if (!migrationFilenamePattern.test(filename)) {
      continue;
    }
    const file = path.join(migrationDir, filename);
    const sql = stripSqlComments(readFileSync(file, "utf8"));
    collectObjectsFromSql(sql, filename, objects);
  }
  for (const object of manifest.synthetic_objects) {
    addObject(objects, object.kind, object.name, object.source);
  }
  return [...objects.values()];
}

function collectObjectsFromSql(sql, source, objects) {
  scan(sql, sqlRegex(String.raw`\bCREATE\s+EXTENSION\s+(?:IF\s+NOT\s+EXISTS\s+)?(${sqlIdentifierPattern})`), "extension", source, objects);
  scan(sql, sqlRegex(String.raw`\bCREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(${sqlIdentifierPattern})`), "table", source, objects);
  scan(sql, sqlRegex(String.raw`\bCREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?(${sqlIdentifierPattern})`), "index", source, objects);
  scan(sql, sqlRegex(String.raw`\bCREATE\s+(?:MATERIALIZED\s+)?(?:OR\s+REPLACE\s+)?VIEW\s+(${sqlIdentifierPattern})`), "view", source, objects);
  scan(sql, sqlRegex(String.raw`\bCREATE\s+(?:OR\s+REPLACE\s+)?FUNCTION\s+(${sqlIdentifierPattern})\s*\(`), "function", source, objects);
  scan(sql, sqlRegex(String.raw`\bCREATE\s+TRIGGER\s+(${sqlIdentifierPattern})`), "trigger", source, objects);
  scan(sql, sqlRegex(String.raw`\bCREATE\s+TYPE\s+(${sqlIdentifierPattern})`), "type", source, objects);
  scan(sql, sqlRegex(String.raw`\bCREATE\s+SEQUENCE\s+(?:IF\s+NOT\s+EXISTS\s+)?(${sqlIdentifierPattern})`), "sequence", source, objects);
  scan(sql, sqlRegex(String.raw`\bCREATE\s+POLICY\s+(${sqlIdentifierPattern})\s+ON\s+${sqlIdentifierPattern}`), "policy", source, objects);
  scan(sql, sqlRegex(String.raw`\bALTER\s+TABLE\s+(?:IF\s+EXISTS\s+)?(?:ONLY\s+)?(${sqlIdentifierPattern})`), "table", source, objects);
  scan(sql, sqlRegex(String.raw`\bALTER\s+INDEX\s+(?:IF\s+EXISTS\s+)?(${sqlIdentifierPattern})`), "index", source, objects);
  scan(sql, sqlRegex(String.raw`\bALTER\s+(?:MATERIALIZED\s+)?VIEW\s+(?:IF\s+EXISTS\s+)?(${sqlIdentifierPattern})`), "view", source, objects);
  scan(sql, sqlRegex(String.raw`\bDROP\s+TABLE\s+(?:IF\s+EXISTS\s+)?(${sqlIdentifierPattern})`), "table", source, objects);
  scan(sql, sqlRegex(String.raw`\bDROP\s+INDEX\s+(?:IF\s+EXISTS\s+)?(${sqlIdentifierPattern})`), "index", source, objects);
  scan(sql, sqlRegex(String.raw`\bDROP\s+(?:MATERIALIZED\s+)?VIEW\s+(?:IF\s+EXISTS\s+)?(${sqlIdentifierPattern})`), "view", source, objects);
  scan(sql, sqlRegex(String.raw`\bDROP\s+FUNCTION\s+(?:IF\s+EXISTS\s+)?(${sqlIdentifierPattern})\s*(?:\(|;)`), "function", source, objects);
  scan(sql, sqlRegex(String.raw`\bDROP\s+TYPE\s+(?:IF\s+EXISTS\s+)?(${sqlIdentifierPattern})`), "type", source, objects);
  scan(sql, sqlRegex(String.raw`\bDROP\s+SEQUENCE\s+(?:IF\s+EXISTS\s+)?(${sqlIdentifierPattern})`), "sequence", source, objects);
  scan(sql, sqlRegex(String.raw`\bUPDATE\s+(?:ONLY\s+)?(${sqlIdentifierPattern})`), "table", source, objects);
  scan(sql, sqlRegex(String.raw`\bINSERT\s+INTO\s+(${sqlIdentifierPattern})`), "data_backfill", source, objects);
  scan(sql, /\b(?:ADD\s+)?CONSTRAINT\s+([A-Za-z_][A-Za-z0-9_]*)/giu, "constraint", source, objects, {
    skipNames: new Set(["if"]),
  });
  if (/\bGENERATED\b[\s\S]*?\bAS\b/iu.test(sql)) {
    addObject(objects, "generated_column", "generated_column", source);
  }
}

function sqlRegex(expression) {
  return new RegExp(expression, "giu");
}

function scan(sql, regex, kind, source, objects, { skipNames = new Set() } = {}) {
  for (const match of sql.matchAll(regex)) {
    const name = normalizeName(match[1]);
    if (!name || skipNames.has(name)) {
      continue;
    }
    addObject(objects, kind, name, source);
  }
}

function addObject(objects, kind, name, source) {
  const normalizedName = normalizeName(name);
  const key = `${kind}\u0000${normalizedName}`;
  const existing = objects.get(key);
  if (existing) {
    existing.sources.push(source);
    return;
  }
  objects.set(key, {
    kind,
    name: normalizedName,
    sources: [source],
  });
}

function normalizeName(name) {
  return name
    .trim()
    .replace(/[";]/gu, "")
    .replace(/\s*\.\s*/gu, ".")
    .replace(/\(.*/u, "")
    .split(".")
    .pop()
    .toLowerCase();
}

function stripSqlComments(sql) {
  return sql
    .replace(/\/\*[\s\S]*?\*\//gu, " ")
    .split("\n")
    .map((line) => line.replace(/--.*$/u, ""))
    .join("\n");
}
