import { readFileSync } from "node:fs";
import path from "node:path";

const rfc3339TimestampPattern =
  /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/u;

export function readJsonObject(file, label = file) {
  const text = readFileSync(file, "utf8");
  assertNoDuplicateObjectMembers(text, label);
  const value = JSON.parse(text);
  return requireObject(value, label);
}

export function parseJsonObject(text, label = "json") {
  assertNoDuplicateObjectMembers(text, label);
  return requireObject(JSON.parse(text), label);
}

function assertNoDuplicateObjectMembers(text, label) {
  let index = skipWhitespace(text, 0);

  function fail(message) {
    throw new Error(`${label} ${message}`);
  }

  function skipWhitespaceAt(position) {
    return skipWhitespace(text, position);
  }

  function parseString(position) {
    if (text[position] !== '"') {
      fail(`expected JSON string at byte ${position}`);
    }
    let cursor = position + 1;
    while (cursor < text.length) {
      const char = text[cursor];
      if (char === "\\") {
        cursor += 2;
        continue;
      }
      if (char === '"') {
        return [JSON.parse(text.slice(position, cursor + 1)), cursor + 1];
      }
      cursor += 1;
    }
    fail(`unterminated JSON string at byte ${position}`);
  }

  function parsePrimitive(position) {
    let cursor = position;
    while (cursor < text.length && !/[,\]}]/.test(text[cursor])) {
      cursor += 1;
    }
    return cursor;
  }

  function parseArray(position) {
    let cursor = skipWhitespaceAt(position + 1);
    if (text[cursor] === "]") {
      return cursor + 1;
    }
    while (cursor < text.length) {
      cursor = skipWhitespaceAt(parseValue(cursor));
      if (text[cursor] === ",") {
        cursor = skipWhitespaceAt(cursor + 1);
        continue;
      }
      if (text[cursor] === "]") {
        return cursor + 1;
      }
      fail(`expected array separator at byte ${cursor}`);
    }
    fail(`unterminated JSON array at byte ${position}`);
  }

  function parseObject(position) {
    const keys = new Set();
    let cursor = skipWhitespaceAt(position + 1);
    if (text[cursor] === "}") {
      return cursor + 1;
    }
    while (cursor < text.length) {
      const [key, afterKey] = parseString(cursor);
      if (keys.has(key)) {
        fail(`has duplicate object member ${JSON.stringify(key)}`);
      }
      keys.add(key);
      cursor = skipWhitespaceAt(afterKey);
      if (text[cursor] !== ":") {
        fail(`expected object member separator at byte ${cursor}`);
      }
      cursor = skipWhitespaceAt(parseValue(skipWhitespaceAt(cursor + 1)));
      if (text[cursor] === ",") {
        cursor = skipWhitespaceAt(cursor + 1);
        continue;
      }
      if (text[cursor] === "}") {
        return cursor + 1;
      }
      fail(`expected object separator at byte ${cursor}`);
    }
    fail(`unterminated JSON object at byte ${position}`);
  }

  function parseValue(position) {
    const cursor = skipWhitespaceAt(position);
    const char = text[cursor];
    if (char === "{") {
      return parseObject(cursor);
    }
    if (char === "[") {
      return parseArray(cursor);
    }
    if (char === '"') {
      return parseString(cursor)[1];
    }
    return parsePrimitive(cursor);
  }

  index = parseValue(index);
  index = skipWhitespaceAt(index);
  if (index !== text.length) {
    fail(`has trailing content at byte ${index}`);
  }
}

function skipWhitespace(text, position) {
  let cursor = position;
  while (cursor < text.length && /\s/.test(text[cursor])) {
    cursor += 1;
  }
  return cursor;
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

export function requireNullableEnum(value, label, allowed) {
  if (value === null) {
    return null;
  }
  return requireEnum(value, label, allowed);
}

export function requireRFC3339Timestamp(value, label) {
  const timestamp = requireString(value, label);
  if (
    !rfc3339TimestampPattern.test(timestamp) ||
    Number.isNaN(Date.parse(timestamp))
  ) {
    throw new Error(`${label} must be an RFC3339 timestamp`);
  }
  return timestamp;
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

export function requireSorted(
  values,
  label,
  keyFn,
  orderLabel = "stable key",
) {
  let previous = null;
  for (const value of values) {
    const key = keyFn(value);
    if (previous !== null && key < previous) {
      throw new Error(`${label} must be sorted by ${orderLabel}`);
    }
    previous = key;
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
