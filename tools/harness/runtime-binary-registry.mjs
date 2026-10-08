import path from "node:path";

const idPattern = /^[a-z][a-z0-9_-]*$/u;
const makeVariablePattern = /^[A-Z][A-Z0-9_]*$/u;

const runtimeBinaryRecordKeys = Object.freeze([
  "id",
  "producer_target",
  "output_make_variable",
  "consumer_env",
  "default_output_path",
]);

function requireObject(value, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  return value;
}

function requireString(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  return value.trim();
}

function rawField(entry, snakeName, camelName) {
  if (Object.hasOwn(entry, snakeName)) {
    return entry[snakeName];
  }
  return entry[camelName];
}

function normalizeRepoRelativePath(value, label) {
  const normalized = requireString(value, label);
  if (
    normalized.includes("\0") ||
    normalized.includes("\\") ||
    path.isAbsolute(normalized)
  ) {
    throw new Error(`${label} must be a repo-relative path token`);
  }
  const segments = normalized.split("/");
  if (
    segments.some((segment) => segment === "" || segment === "." || segment === "..")
  ) {
    throw new Error(`${label} must not contain empty, . or .. path segments`);
  }
  return normalized;
}

function normalizeRuntimeBinaryEntry(raw, label, { taskTargets = null } = {}) {
  const entry = requireObject(raw, label);
  for (const key of Object.keys(entry)) {
    if (
      !runtimeBinaryRecordKeys.includes(key) &&
      !["producerTarget", "outputMakeVariable", "consumerEnv", "defaultOutputPath"].includes(key)
    ) {
      throw new Error(`${label} has unknown key ${key}`);
    }
  }
  const id = requireString(rawField(entry, "id", "id"), `${label}.id`);
  if (!idPattern.test(id)) {
    throw new Error(`${label}.id must be a lowercase identifier`);
  }
  const producerTarget = requireString(
    rawField(entry, "producer_target", "producerTarget"),
    `${label}.producer_target`,
  );
  if (taskTargets && !taskTargets.has(producerTarget)) {
    throw new Error(
      `${label}.producer_target ${producerTarget} is missing from task_surface.targets`,
    );
  }
  const outputMakeVariable = requireString(
    rawField(entry, "output_make_variable", "outputMakeVariable"),
    `${label}.output_make_variable`,
  );
  if (!makeVariablePattern.test(outputMakeVariable)) {
    throw new Error(`${label}.output_make_variable must be a Make variable name`);
  }
  const consumerEnv = requireString(
    rawField(entry, "consumer_env", "consumerEnv"),
    `${label}.consumer_env`,
  );
  if (!makeVariablePattern.test(consumerEnv)) {
    throw new Error(`${label}.consumer_env must be an environment variable name`);
  }
  const defaultOutputPath = normalizeRepoRelativePath(
    rawField(entry, "default_output_path", "defaultOutputPath"),
    `${label}.default_output_path`,
  );
  return {
    id,
    producerTarget,
    outputMakeVariable,
    consumerEnv,
    defaultOutputPath,
    producer_target: producerTarget,
    output_make_variable: outputMakeVariable,
    consumer_env: consumerEnv,
    default_output_path: defaultOutputPath,
  };
}

export function normalizeRuntimeBinaryEntries(entries = [], options = {}) {
  const records = [];
  const seen = new Set();
  const label = options.label ?? "runtime_binaries";
  for (const [index, raw] of entries.entries()) {
    const record = normalizeRuntimeBinaryEntry(raw, `${label}[${index + 1}]`, options);
    if (seen.has(record.id)) {
      throw new Error(`duplicate runtime binary ${record.id}`);
    }
    seen.add(record.id);
    records.push(record);
  }
  return records;
}
