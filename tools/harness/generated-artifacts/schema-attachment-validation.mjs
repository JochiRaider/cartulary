import { existsSync, readFileSync, readdirSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";

const sharedExtensionsRef = "cartulary.harness.defs.v1#/$defs/extensions";
const supportSchemaIDs = new Set([
  "cartulary.harness.defs.v1",
  "cartulary.harness_artifact_ref.v1",
]);

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function assertUnique(values, label) {
  const seen = new Set();
  for (const value of values) {
    if (seen.has(value)) {
      throw new Error(`${label} contains duplicate ${value}`);
    }
    seen.add(value);
  }
}

function assertSorted(values, label) {
  for (let index = 1; index < values.length; index += 1) {
    if (values[index] < values[index - 1]) {
      throw new Error(`${label} must be sorted`);
    }
  }
}

function schemaIDFromFile(file) {
  const base = path.basename(file);
  if (!base.endsWith(".schema.json")) {
    throw new Error(`${file} must end with .schema.json`);
  }
  return base.slice(0, -".schema.json".length);
}

function schemaDeclaresSchemaID(schema, schemaID) {
  return (
    schema?.properties?.schema_id?.const === schemaID ||
    (schema?.allOf ?? []).some(
      (entry) => entry?.properties?.schema_id?.const === schemaID,
    )
  );
}

function schemaRequiresSchemaID(schema) {
  return (
    (schema?.required ?? []).includes("schema_id") ||
    (schema?.allOf ?? []).some((entry) =>
      (entry?.required ?? []).includes("schema_id"),
    )
  );
}

function schemaIsClosed(schema) {
  return schema?.additionalProperties === false;
}

function schemaIsAliasOnly(schema) {
  return JSON.stringify(Object.keys(schema).sort()) ===
    JSON.stringify(["$id", "$ref", "$schema"]);
}

function validateExtensionProperties(schema, label) {
  if (!schema || typeof schema !== "object" || Array.isArray(schema)) {
    return;
  }
  if (
    schema.properties?.extensions !== undefined &&
    !(
      schema.properties.extensions?.$ref === sharedExtensionsRef &&
      Object.keys(schema.properties.extensions).length === 1
    )
  ) {
    throw new Error(
      `${label}.properties.extensions must reference ${sharedExtensionsRef}`,
    );
  }
  for (const [key, value] of Object.entries(schema)) {
    if (key === "properties" && value && typeof value === "object") {
      for (const [propertyName, propertySchema] of Object.entries(value)) {
        validateExtensionProperties(propertySchema, `${label}.properties.${propertyName}`);
      }
    } else if (Array.isArray(value)) {
      value.forEach((entry, index) =>
        validateExtensionProperties(entry, `${label}.${key}[${index + 1}]`),
      );
    } else {
      validateExtensionProperties(value, `${label}.${key}`);
    }
  }
}

export function validateSchemaAttachmentPolicy(root) {
  const registryFile = path.join(root, "tools/harness_schema_attachments.json");
  const registry = readJSON(registryFile);
  validateSchemaSync("cartulary.harness_schema_attachments.v1", registry);
  const schemaIDs = registry.attachments.map((entry) => entry.schema_id);
  const registeredPaths = registry.attachments.map((entry) => entry.path);
  assertSorted(schemaIDs, `${registryFile}.attachments`);
  assertUnique(schemaIDs, `${registryFile}.attachments.schema_id`);
  assertUnique(registeredPaths, `${registryFile}.attachments.path`);

  const schemaDir = path.join(root, "tools/schemas");
  const discoveredPaths = [];
  for (const name of readdirSync(schemaDir).sort((left, right) =>
    left.localeCompare(right),
  )) {
    if (!name.endsWith(".schema.json")) {
      continue;
    }
    const file = path.join(schemaDir, name);
    const relativeFile = path.relative(root, file).replaceAll("\\", "/");
    discoveredPaths.push(relativeFile);
    if (!registeredPaths.includes(relativeFile)) {
      throw new Error(`${relativeFile} is not registered by ${registryFile}`);
    }
    const schema = readJSON(file);
    const schemaID = schemaIDFromFile(file);
    if (schema.$id !== schemaID) {
      throw new Error(`${file} $id must match ${schemaID}`);
    }
    validateExtensionProperties(schema, file);
    if (supportSchemaIDs.has(schemaID)) {
      continue;
    }
    if (schemaIsAliasOnly(schema)) {
      throw new Error(`${file} must not be an alias-only public schema`);
    }
    if (!schemaRequiresSchemaID(schema)) {
      throw new Error(`${file} must require schema_id`);
    }
    if (!schemaDeclaresSchemaID(schema, schemaID)) {
      throw new Error(`${file} must constrain schema_id to ${schemaID}`);
    }
    if (!schemaIsClosed(schema)) {
      throw new Error(`${file} must be closed at the top level`);
    }
  }
  for (const registeredPath of registeredPaths) {
    if (!discoveredPaths.includes(registeredPath)) {
      throw new Error(`${registryFile} references missing schema ${registeredPath}`);
    }
  }
}

export function validateHarnessHelperOwnership(root) {
  const ownerFile = path.join(root, "tools/harness_helper_ownership.json");
  const owner = readJSON(ownerFile);
  validateSchemaSync("cartulary.harness_helper_ownership.v1", owner);
  const authoredKeys = owner.facades.map((entry) => entry.key);
  assertSorted(authoredKeys, `${ownerFile}.facades`);
  assertUnique(authoredKeys, `${ownerFile}.facades.key`);
  const facadePaths = owner.facades.flatMap((entry) => entry.paths);
  assertUnique(facadePaths, `${ownerFile}.facades.paths`);
  for (const facadePath of facadePaths) {
    if (!existsSync(path.join(root, facadePath))) {
      throw new Error(`${ownerFile} references missing current facade ${facadePath}`);
    }
  }

}
