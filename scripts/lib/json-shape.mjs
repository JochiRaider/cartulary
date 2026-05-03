import { readFileSync } from "node:fs";
import path from "node:path";

export function readJsonObject(file, label = file) {
  const value = JSON.parse(readFileSync(file, "utf8"));
  return requireObject(value, label);
}

export function requireObject(value, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  return value;
}

export function requireArray(value, label, { nonEmpty = false } = {}) {
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  if (nonEmpty && value.length === 0) {
    throw new Error(`${label} must be a non-empty array`);
  }
  return value;
}

export function requireString(value, label, { pattern = null } = {}) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  if (pattern && !pattern.test(value)) {
    throw new Error(`${label} has invalid value ${JSON.stringify(value)}`);
  }
  return value;
}

export function requireBoolean(value, label) {
  if (typeof value !== "boolean") {
    throw new Error(`${label} must be a boolean`);
  }
  return value;
}

export function requireInteger(value, label, { min = null } = {}) {
  if (!Number.isInteger(value)) {
    throw new Error(`${label} must be an integer`);
  }
  if (min !== null && value < min) {
    throw new Error(`${label} must be >= ${min}`);
  }
  return value;
}

export function requirePositiveInteger(value, label) {
  return requireInteger(value, label, { min: 1 });
}

export function requireEnum(value, label, allowed) {
  const string = requireString(value, label);
  if (!allowed.has(string)) {
    throw new Error(`${label} must be one of ${[...allowed].join("|")}`);
  }
  return string;
}

export function optionalStringArray(value, label) {
  if (value === undefined) {
    return [];
  }
  return requireStringArray(value, label);
}

export function requireStringArray(value, label, { nonEmpty = false, pattern = null } = {}) {
  const array = requireArray(value, label, { nonEmpty });
  const seen = new Set();
  const result = [];
  for (const [index, entry] of array.entries()) {
    const string = requireString(entry, `${label}[${index + 1}]`, { pattern });
    if (seen.has(string)) {
      throw new Error(`${label} contains duplicate ${string}`);
    }
    seen.add(string);
    result.push(string);
  }
  return result;
}

export function requireSchemaID(object, expected, label) {
  if (object.schema_id !== expected) {
    throw new Error(`${label} must declare schema_id ${expected}`);
  }
}

export function assertObjectKeys(object, allowedKeys, label) {
  requireObject(object, label);
  for (const key of Object.keys(object)) {
    if (!allowedKeys.has(key)) {
      throw new Error(`${label} has unknown key ${key}`);
    }
  }
}

export function assertRequiredKeys(object, requiredKeys, label) {
  requireObject(object, label);
  for (const key of requiredKeys) {
    if (!Object.hasOwn(object, key)) {
      throw new Error(`${label}.${key} is required`);
    }
  }
}

export function validateObjectShape(
  value,
  label,
  { keys = null, requiredKeys = null } = {},
) {
  const object = requireObject(value, label);
  if (keys) {
    assertObjectKeys(object, keys, label);
  }
  if (requiredKeys) {
    assertRequiredKeys(object, requiredKeys, label);
  }
  return object;
}

export function assertUnique(values, label) {
  const seen = new Set();
  for (const value of values) {
    if (seen.has(value)) {
      throw new Error(`${label} contains duplicate ${value}`);
    }
    seen.add(value);
  }
}

export function requireRepoRelativePath(value, label, { extension = "" } = {}) {
  const relative = requireString(value, label);
  if (relative.includes("\0") || relative.includes("\\") || path.posix.isAbsolute(relative)) {
    throw new Error(`${label} must be a normalized repo-relative path`);
  }
  const normalized = path.posix.normalize(relative);
  if (normalized === "." || normalized.startsWith("../") || normalized.includes("/../")) {
    throw new Error(`${label} must stay under the repository root`);
  }
  if (normalized !== relative) {
    throw new Error(`${label} must be normalized`);
  }
  if (extension && !relative.endsWith(extension)) {
    throw new Error(`${label} must end with ${extension}`);
  }
  return relative;
}

export function objectEntries(value, label) {
  return Object.entries(requireObject(value, label));
}

export function requireObjectArray(value, label, { nonEmpty = false } = {}) {
  return requireArray(value, label, { nonEmpty }).map((entry, index) =>
    requireObject(entry, `${label}[${index + 1}]`),
  );
}

export function validateObjectArray(
  value,
  label,
  { nonEmpty = false, keys = null, requiredKeys = null } = {},
  visit = null,
) {
  const entries = requireObjectArray(value, label, { nonEmpty });
  for (const [index, entry] of entries.entries()) {
    const entryLabel = `${label}[${index + 1}]`;
    validateObjectShape(entry, entryLabel, { keys, requiredKeys });
    if (visit) {
      visit(entry, entryLabel, index);
    }
  }
  return entries;
}
