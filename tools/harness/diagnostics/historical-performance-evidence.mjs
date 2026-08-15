import { createHash } from "node:crypto";
import { existsSync, lstatSync, readFileSync, readdirSync, realpathSync } from "node:fs";
import { createRequire } from "node:module";
import path from "node:path";

import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";

const requireFromDiagnostics = createRequire(import.meta.url);
const registrySchemaID = "cartulary.harness_historical_schema_registry.v1";
const registryRelativePath = "tools/historical_performance_evidence_registry.json";
const historicalRootRelativePath = "tools/historical-schemas/performance";
const validatorCache = new Map();
let loadedRegistry = null;

function sha256(bytes) {
  return createHash("sha256").update(bytes).digest("hex");
}

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

export function loadHistoricalPerformanceSchemaRegistry(repoRoot) {
  if (loadedRegistry?.repoRoot === repoRoot) return loadedRegistry;
  const registryFile = path.join(repoRoot, registryRelativePath);
  const registry = parseStrictJSON(readFileSync(registryFile, "utf8"), registryFile);
  validateSchemaSync(registrySchemaID, registry);
  const historicalRoot = realpathSync(
    path.join(repoRoot, historicalRootRelativePath),
  );
  const schemaIDs = registry.schemas.map((entry) => entry.schema_id);
  if (
    new Set(schemaIDs).size !== schemaIDs.length ||
    JSON.stringify(schemaIDs) !== JSON.stringify([...schemaIDs].sort(compareASCII))
  ) {
    throw new Error("historical performance schema registry must be sorted and unique");
  }
  const schemas = new Map();
  for (const entry of registry.schemas) {
    const file = path.resolve(repoRoot, entry.path);
    const relative = path.relative(historicalRoot, file);
    if (!relative || relative.startsWith("..") || path.isAbsolute(relative)) {
      throw new Error(`historical performance schema escapes its registry root: ${entry.path}`);
    }
    const info = lstatSync(file);
    if (!info.isFile() || info.isSymbolicLink() || realpathSync(file) !== file) {
      throw new Error(`historical performance schema is not a regular owned file: ${entry.path}`);
    }
    const bytes = readFileSync(file);
    const schema = parseStrictJSON(bytes.toString("utf8"), file);
    if (schema.$id !== entry.schema_id || sha256(bytes) !== entry.sha256) {
      throw new Error(`historical performance schema identity or digest drifted: ${entry.path}`);
    }
    schemas.set(entry.schema_id, schema);
  }
  const discovered = readdirSync(historicalRoot)
    .filter((name) => name.endsWith(".schema.json"))
    .sort(compareASCII);
  const registered = registry.schemas
    .map((entry) => path.basename(entry.path))
    .sort(compareASCII);
  if (JSON.stringify(discovered) !== JSON.stringify(registered)) {
    throw new Error("historical performance schema registry does not close its directory");
  }
  loadedRegistry = { repoRoot, registry, schemas };
  return loadedRegistry;
}

function validatorFor(repoRoot, schemaID) {
  const cacheKey = `${repoRoot}\0${schemaID}`;
  if (validatorCache.has(cacheKey)) return validatorCache.get(cacheKey);
  const loaded = loadHistoricalPerformanceSchemaRegistry(repoRoot);
  if (!loaded.schemas.has(schemaID)) {
    throw new Error(`schema is not registered for historical inspection: ${schemaID}`);
  }
  const Ajv = requireFromDiagnostics("ajv/dist/2020");
  const ajv = new Ajv({
    allErrors: true,
    strict: false,
    validateFormats: false,
    validateSchema: true,
  });
  const activeSchemaRoot = path.join(repoRoot, "tools/schemas");
  for (const name of readdirSync(activeSchemaRoot)
    .filter((candidate) => candidate.endsWith(".schema.json"))
    .sort(compareASCII)) {
    const schema = JSON.parse(readFileSync(path.join(activeSchemaRoot, name), "utf8"));
    ajv.addSchema(schema, schema.$id);
  }
  for (const schema of loaded.schemas.values()) ajv.addSchema(schema, schema.$id);
  const validate = ajv.getSchema(schemaID);
  if (validate === undefined) {
    throw new Error(`historical performance schema failed to compile: ${schemaID}`);
  }
  validate.ajv = ajv;
  validatorCache.set(cacheKey, validate);
  return validate;
}

export function validateHistoricalPerformanceEvidence(repoRoot, schemaID, value) {
  const validate = validatorFor(repoRoot, schemaID);
  if (!validate(value)) {
    throw new Error(
      `${schemaID} historical validation failed:\n  ${validate.ajv.errorsText(validate.errors, { separator: "\n  " })}`,
    );
  }
}
