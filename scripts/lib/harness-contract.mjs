import { randomBytes, timingSafeEqual } from "node:crypto";
import {
  existsSync,
  lstatSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  rmSync,
  unlinkSync,
} from "node:fs";
import { createRequire } from "node:module";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const requireFromHarness = createRequire(import.meta.url);
export const repoRoot = path.resolve(scriptDir, "..", "..");
export const schemasDir = path.join(repoRoot, "tools", "schemas");
export const redactionManifestPath = path.join(
  repoRoot,
  "tools",
  "harness_redaction_manifest.json",
);

export const outputModes = Object.freeze([
  "quiet",
  "summary",
  "ci",
  "verbose",
  "debug",
  "machine",
]);
const outputModeSet = new Set(outputModes);
const machineAcceptedOutputClasses = new Set([
  "aggregate_summary_with_artifacts",
  "machine_stdout",
  "scheduler_summary_with_artifacts",
  "service_summary",
  "service_summary_with_artifacts",
  "summary_with_artifacts",
]);
const machineRejectedOutputClasses = new Set([
  "destructive_human",
  "human_summary",
  "interactive_raw",
]);
const schedulerResourceRegistryPath = path.join(
  repoRoot,
  "tools",
  "scheduler_resource_registry.json",
);
const defaultHarnessConfig = Object.freeze({
  CARTULARY_TEST_RESULTS_DIR: ".cartulary/test-results",
  CARTULARY_TEST_RUN_ID: null,
});
const exactOneBooleans = Object.freeze([
  "VERBOSE",
  "CI_VERBOSE",
  "CI",
  "CARTULARY_TEST_SERVICES_ACTIVE",
  "CARTULARY_ENABLE_TEST_ROUTES",
  "CARTULARY_CLEANUP_DRY_RUN",
  "LINT_SHELL_STRICT",
]);
const optionalPathVariables = Object.freeze([
  "GO",
  "GO_CACHE_DIR",
  "GO_MOD_CACHE_DIR",
  "GOCACHE",
  "GOMODCACHE",
  "NODE_RUNTIME_DIR",
  "NODE_BIN",
  "PNPM",
  "COREPACK_HOME",
  "CONFIG_FILE",
  "CARTULARY_CONFIG_FILE",
  "TEST_SERVICES_BIN",
  "CARTULARY_TEST_SERVICES_BIN",
  "CARTULARY_HARNESS_REPO_ROOT",
  "CARTULARY_HARNESS_SCRATCH_ROOT",
  "TMPDIR",
  "CARTULARY_COMPOSE_FILE",
  "CARTULARY_SERVER_BIN",
  "CARTULARY_MIGRATE_BIN",
]);
const versionVariables = Object.freeze([
  "NODE_VERSION",
  "PNPM_VERSION",
  "SHELLCHECK_VERSION",
]);
const positiveIntegerVariables = Object.freeze([
  "BACKEND_STORE_GO_TEST_P",
  "BACKEND_INTEGRATION_GO_TEST_P",
  "GO_TEST_SERVICE_PACKAGE_PARALLELISM",
  "BACKEND_INTEGRATION_SHARD_JOBS",
  "HARNESS_SMOKE_JOBS",
  "PLAYWRIGHT_WORKERS",
  "VITEST_MAX_WORKERS",
  "FIXTURE_TOP",
]);
const boundedPositiveIntegerVariables = Object.freeze({
  CARTULARY_TEST_SERVICES_WEB_E2E_CLEANUP_WORKERS: { min: 1, max: 16 },
  CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET: { min: 0, max: 1024 },
  CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS: { min: 1, max: 999999999 },
});
const autoOrPositiveIntegerVariables = Object.freeze([
  "BROWSER_E2E_FUNCTIONAL_SHARDS",
]);
const serviceAttachGroups = Object.freeze([
  {
    label: "Postgres attach set",
    names: [
      "CARTULARY_PGTEST_ADMIN_DSN",
      "CARTULARY_PGTEST_DSN_TEMPLATE",
      "CARTULARY_PGTEST_TEMPLATE_DB",
    ],
    validate(values) {
      if (!values.CARTULARY_PGTEST_DSN_TEMPLATE.includes("{database}")) {
        throw new HarnessConfigError(
          "CARTULARY_PGTEST_DSN_TEMPLATE must contain {database}",
        );
      }
    },
  },
  {
    label: "MinIO attach set",
    names: [
      "CARTULARY_S3TEST_ENDPOINT",
      "CARTULARY_S3TEST_ACCESS_KEY_ID",
      "CARTULARY_S3TEST_SECRET_ACCESS_KEY",
      "CARTULARY_S3TEST_SECURE",
    ],
    validate(values) {
      validateBooleanToken("CARTULARY_S3TEST_SECURE", values.CARTULARY_S3TEST_SECURE);
    },
  },
]);

let schedulerResourceRegistry = null;

export class HarnessConfigError extends Error {
  constructor(message, { reason = "configuration_error", exitCode = 2 } = {}) {
    super(message);
    this.name = "HarnessConfigError";
    this.failure_class = "config";
    this.failure_reason = reason;
    this.exit_code = exitCode;
  }
}

export function defaultResultsRoot(root = repoRoot) {
  return path.join(root, ".cartulary", "test-results");
}

function isPresent(value) {
  return value !== undefined && value !== null && String(value) !== "";
}

function selectedSource(env, name, defaultValue = undefined) {
  if (Object.hasOwn(env, name)) {
    return { value: String(env[name] ?? ""), source: "env" };
  }
  if (defaultValue !== undefined && defaultValue !== null) {
    return { value: String(defaultValue), source: "default" };
  }
  return { value: "", source: "omitted" };
}

function requireNonEmptyString(name, value) {
  if (!isPresent(value)) {
    throw new HarnessConfigError(`${name} must not be empty`);
  }
  const normalized = String(value).trim();
  if (normalized === "") {
    throw new HarnessConfigError(`${name} must not be empty`);
  }
  return normalized;
}

function validatePathToken(name, value) {
  const normalized = requireNonEmptyString(name, value);
  if (normalized.includes("\0")) {
    throw new HarnessConfigError(`${name} must not contain NUL`);
  }
  return normalized;
}

function validateVersionToken(name, value) {
  const normalized = requireNonEmptyString(name, value);
  if (!/^[A-Za-z0-9._+-]+$/u.test(normalized)) {
    throw new HarnessConfigError(`${name} must be a version token`);
  }
  return normalized;
}

function validatePositiveInteger(name, value, { min = 1, max = 9999 } = {}) {
  const raw = requireNonEmptyString(name, value);
  if (!/^[0-9]+$/u.test(raw)) {
    throw new HarnessConfigError(`${name} must be a positive integer`);
  }
  const parsed = Number.parseInt(raw, 10);
  if (!Number.isSafeInteger(parsed) || parsed < min || parsed > max) {
    if (min === 1) {
      throw new HarnessConfigError(`${name} must be a positive integer <= ${max}`);
    }
    throw new HarnessConfigError(`${name} must be an integer between ${min} and ${max}`);
  }
  return parsed;
}

function validateAutoOrPositiveInteger(name, value, { max = 9999 } = {}) {
  const raw = requireNonEmptyString(name, value);
  if (raw === "auto") {
    return raw;
  }
  return validatePositiveInteger(name, raw, { max });
}

function validateBooleanToken(name, value) {
  const raw = requireNonEmptyString(name, value);
  if (!["true", "false", "1", "0"].includes(raw)) {
    throw new HarnessConfigError(`${name} must be true, false, 1, or 0`);
  }
  return raw;
}

function resolveOutputModeRecord(env = process.env, target = "") {
  if (isPresent(env.CARTULARY_OUTPUT_MODE)) {
    const raw = String(env.CARTULARY_OUTPUT_MODE).trim();
    const mode = raw.toLowerCase();
    if (raw !== mode || !outputModeSet.has(mode)) {
      throw new HarnessConfigError(
        `invalid CARTULARY_OUTPUT_MODE ${JSON.stringify(raw)}; expected one of ${outputModes.join(", ")}`,
      );
    }
    return { value: mode, source: "env:CARTULARY_OUTPUT_MODE" };
  }
  if (env.VERBOSE === "1") {
    return { value: "verbose", source: "env:VERBOSE" };
  }
  if (env.CI_VERBOSE === "1") {
    return { value: "ci", source: "env:CI_VERBOSE" };
  }
  if (target === "ci" || env.CI === "1") {
    return { value: "ci", source: target === "ci" ? "target" : "env:CI" };
  }
  return { value: "summary", source: "default" };
}

export function resolveOutputMode(env = process.env, target = "") {
  return resolveOutputModeRecord(env, target).value;
}

export function validateRunId(value) {
  const runId = String(value ?? "").trim();
  if (!/^[A-Za-z0-9_.-]{1,96}$/u.test(runId) || runId === "." || runId === "..") {
    throw new HarnessConfigError(
      "CARTULARY_TEST_RUN_ID must be 1-96 characters of A-Z a-z 0-9 _ . -",
    );
  }
  return runId;
}

export function validateResultRoot(value, { root = repoRoot, create = false } = {}) {
  const configured = value === undefined || value === null
    ? ".cartulary/test-results"
    : String(value);
  if (
    configured === "" ||
    configured === "/" ||
    configured === "." ||
    configured === ".." ||
    configured.includes("\0") ||
    configured.includes("\\") ||
    configured.split("/").includes("..")
  ) {
    throw new HarnessConfigError(
      `invalid CARTULARY_TEST_RESULTS_DIR ${JSON.stringify(configured)}`,
    );
  }
  const resolved = path.resolve(root, configured);
  if (create) {
    mkdirSync(resolved, { recursive: true });
  }
  return resolved;
}

export function generateRunId(now = new Date()) {
  return `${now.toISOString().replace(/[-:]/gu, "").replace(/\..+$/u, "Z")}-r${randomBytes(6).toString("hex")}`;
}

function loadTaskSurfaceManifest(manifestPath = process.env.TASK_SURFACE_MANIFEST) {
  const file = path.resolve(repoRoot, manifestPath || "tools/task_surface_manifest.json");
  return JSON.parse(readFileSync(file, "utf8"));
}

export function targetPolicy(target, manifest = loadTaskSurfaceManifest()) {
  return manifest.targets?.find((entry) => entry.name === target) ?? null;
}

function loadSchedulerResourceRegistry() {
  if (!schedulerResourceRegistry) {
    schedulerResourceRegistry = JSON.parse(readFileSync(schedulerResourceRegistryPath, "utf8"));
  }
  return schedulerResourceRegistry;
}

function declaredSchedulerOverrideEnvNames() {
  const registry = loadSchedulerResourceRegistry();
  return new Set(
    (registry.resources ?? [])
      .map((resource) => resource.capacity?.override_env)
      .filter((name) => typeof name === "string" && name !== ""),
  );
}

function validateDeclaredResourceLimits(env) {
  const declaredNames = declaredSchedulerOverrideEnvNames();
  const resolved = {};
  for (const name of Array.from(declaredNames).sort()) {
    if (!isPresent(env[name])) {
      continue;
    }
    resolved[name] = validatePositiveInteger(name, env[name], { max: 256 });
  }
  return resolved;
}

function validateExactOneBooleans(env) {
  const resolved = {};
  for (const name of exactOneBooleans) {
    if (!isPresent(env[name])) {
      continue;
    }
    resolved[name] = env[name] === "1";
  }
  return resolved;
}

function validateOptionalPaths(env) {
  const resolved = {};
  for (const name of optionalPathVariables) {
    if (!isPresent(env[name])) {
      continue;
    }
    const normalized = validatePathToken(name, env[name]);
    if (name === "CARTULARY_HARNESS_SCRATCH_ROOT") {
      const scratchRoot = path.resolve(repoRoot, normalized);
      const relative = path.relative(repoRoot, scratchRoot);
      if (relative === "" || (!relative.startsWith("..") && !path.isAbsolute(relative))) {
        throw new HarnessConfigError(
          `CARTULARY_HARNESS_SCRATCH_ROOT must be outside the repository: ${scratchRoot}`,
        );
      }
      resolved[name] = scratchRoot;
      continue;
    }
    resolved[name] = normalized;
  }
  for (const [name, value] of Object.entries(env)) {
    if (!/^CARTULARY__ROOTS__[A-Z0-9_]+__PATH$/u.test(name) || !isPresent(value)) {
      continue;
    }
    resolved[name] = validatePathToken(name, value);
  }
  return resolved;
}

function validateVersionTokens(env) {
  const resolved = {};
  for (const name of versionVariables) {
    if (!isPresent(env[name])) {
      continue;
    }
    resolved[name] = validateVersionToken(name, env[name]);
  }
  return resolved;
}

function validateIntegerVariables(env) {
  const resolved = {};
  for (const name of positiveIntegerVariables) {
    if (!isPresent(env[name])) {
      continue;
    }
    resolved[name] = validatePositiveInteger(name, env[name]);
  }
  for (const [name, bounds] of Object.entries(boundedPositiveIntegerVariables)) {
    if (!isPresent(env[name])) {
      continue;
    }
    resolved[name] = validatePositiveInteger(name, env[name], bounds);
  }
  for (const name of autoOrPositiveIntegerVariables) {
    if (!isPresent(env[name])) {
      continue;
    }
    resolved[name] = validateAutoOrPositiveInteger(name, env[name]);
  }
  return resolved;
}

function validateServiceAttachGroups(env) {
  const resolved = {};
  for (const group of serviceAttachGroups) {
    const present = group.names.filter((name) => isPresent(env[name]));
    if (present.length === 0) {
      continue;
    }
    if (present.length !== group.names.length) {
      const missing = group.names.filter((name) => !isPresent(env[name]));
      throw new HarnessConfigError(`${group.label} is incomplete; missing ${missing.join(", ")}`);
    }
    const values = {};
    for (const name of group.names) {
      values[name] = requireNonEmptyString(name, env[name]);
    }
    group.validate(values);
    resolved[group.label] = Object.fromEntries(
      Object.entries(values).map(([name, value]) => [
        name,
        /SECRET|KEY|DSN/u.test(name) ? "[REDACTED]" : value,
      ]),
    );
  }
  return resolved;
}

export function resolveHarnessConfig(target, env = process.env, options = {}) {
  const manifest = options.manifest ?? loadTaskSurfaceManifest();
  const entry = targetPolicy(target, manifest);
  if (!entry || entry.classification !== "public") {
    throw new HarnessConfigError(`unknown public target ${JSON.stringify(target)}`);
  }
  const outputMode = resolveOutputModeRecord(env, target);
  const outputClass = entry.output_policy?.output_class ?? "";
  if (outputMode.value === "machine" && machineRejectedOutputClasses.has(outputClass)) {
    throw new HarnessConfigError(
      `target ${target} does not accept CARTULARY_OUTPUT_MODE=machine`,
      { reason: "usage_error" },
    );
  }
  if (outputMode.value === "machine" && !machineAcceptedOutputClasses.has(outputClass)) {
    throw new HarnessConfigError(
      `target ${target} has no machine output contract`,
      { reason: "usage_error" },
    );
  }
  const resultRoot = selectedSource(
    env,
    "CARTULARY_TEST_RESULTS_DIR",
    defaultHarnessConfig.CARTULARY_TEST_RESULTS_DIR,
  );
  const runId = selectedSource(env, "CARTULARY_TEST_RUN_ID", generateRunId());
  const resolved = {
    target,
    target_policy: {
      classification: entry.classification,
    },
    output_class: outputClass,
    output_mode: outputMode.value,
    output_mode_source: outputMode.source,
    result_root: validateResultRoot(resultRoot.value, {
      root: options.root ?? repoRoot,
      create: options.createResultRoot === true,
    }),
    result_root_source: resultRoot.source,
    run_id: validateRunId(runId.value),
    run_id_source: runId.source,
    generated_run_id: runId.source === "default",
    variables: {
      booleans: validateExactOneBooleans(env),
      paths: validateOptionalPaths(env),
      versions: validateVersionTokens(env),
      integers: validateIntegerVariables(env),
      service_attach: validateServiceAttachGroups(env),
    },
    resource_limits: {
      scheduler_overrides: validateDeclaredResourceLimits(env),
    },
    warnings: [],
  };
  return resolved;
}

export function preflightPublicTarget(target, env = process.env) {
  return resolveHarnessConfig(target, env);
}

function loadRedactionManifest() {
  return JSON.parse(readFileSync(redactionManifestPath, "utf8"));
}

let redactionRules = null;
const syncValidatorCache = new Map();

function compiledRedactionRules() {
  if (redactionRules) {
    return redactionRules;
  }
  const manifest = loadRedactionManifest();
  const replacement = manifest.replacement || "[REDACTED]";
  redactionRules = {
    replacement,
    keyPatterns: (manifest.sensitive_key_patterns ?? []).map((pattern) =>
      new RegExp(pattern, "iu"),
    ),
    valuePatterns: (manifest.value_patterns ?? []).map((rule) => ({
      name: rule.name,
      regex: new RegExp(String(rule.pattern).replace(/^\(\?i\)/u, ""), "giu"),
      replacement: rule.replacement ?? replacement,
    })),
  };
  return redactionRules;
}

export function redactString(value) {
  const rules = compiledRedactionRules();
  let text = String(value);
  for (const rule of rules.valuePatterns) {
    text = text.replace(rule.regex, rule.replacement);
  }
  return text;
}

export function redactValue(value, key = "") {
  const rules = compiledRedactionRules();
  if (typeof key === "string" && rules.keyPatterns.some((pattern) => pattern.test(key))) {
    return rules.replacement;
  }
  if (typeof value === "string") {
    return redactString(value);
  }
  if (Array.isArray(value)) {
    return value.map((entry) => redactValue(entry));
  }
  if (value && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value).map(([entryKey, entryValue]) => [
        entryKey,
        redactValue(entryValue, entryKey),
      ]),
    );
  }
  return value;
}

export function compactJSONString(value) {
  return `${JSON.stringify(redactValue(value))}\n`;
}

export function prettyJSONString(value) {
  return `${JSON.stringify(redactValue(value), null, 2)}\n`;
}

function isUnderPath(parent, child) {
  const relative = path.relative(parent, child);
  return relative === "" || (!relative.startsWith("..") && !path.isAbsolute(relative));
}

function normalizeCleanupCandidate(candidate, { root = repoRoot } = {}) {
  const raw = String(candidate ?? "");
  if (
    raw === "" ||
    raw === "/" ||
    raw === "." ||
    raw === ".." ||
    raw.includes("\0") ||
    raw.includes("\\") ||
    raw.split("/").includes("..")
  ) {
    return { status: "reject", identity: raw || "<empty>", reason: "invalid_path" };
  }
  const resolved = path.resolve(root, raw);
  if (resolved === root || !isUnderPath(root, resolved)) {
    return { status: "reject", identity: raw, reason: "outside_repo" };
  }
  const identity = path.relative(root, resolved).replaceAll("\\", "/");
  return {
    status: "candidate",
    action: "remove",
    identity,
    path: resolved,
    proof: "registered_repo_local_path",
  };
}

function cleanupTmpPlan({ root = repoRoot } = {}) {
  const tmpDir = path.join(root, "tmp");
  const preserve = new Set([
    "node-runtime",
    "node-archives",
    "toolbin",
    "shellcheck-archives",
    "frontend-install",
    "frontend-toolchain",
    "playwright",
    "frontend-embed",
  ]);
  if (!existsSync(tmpDir)) {
    return [];
  }
  return readdirNames(tmpDir)
    .filter((name) => !preserve.has(name))
    .map((name) =>
      normalizeCleanupCandidate(path.join("tmp", name), { root }),
    );
}

function readdirNames(dir) {
  return (existsSync(dir) ? readdirSync(dir, { withFileTypes: true }) : []).map(
    (entry) => entry.name,
  );
}

function removeCandidate(entry) {
  if (!entry.path || (!existsSync(entry.path) && !lstatExists(entry.path))) {
    return false;
  }
  const stat = lstatSync(entry.path);
  if (stat.isSymbolicLink()) {
    unlinkSync(entry.path);
  } else {
    rmSync(entry.path, { recursive: true, force: true });
  }
  return true;
}

function lstatExists(file) {
  try {
    lstatSync(file);
    return true;
  } catch {
    return false;
  }
}

export function runCleanup({
  scope,
  candidates,
  includeTmp = true,
  embeddedWebAssetsDir = path.join(repoRoot, "internal", "platform", "httpapi", "webassets", "dist"),
  dryRun = process.env.CARTULARY_CLEANUP_DRY_RUN === "1",
  stdout = process.stdout,
} = {}) {
  const plan = candidates.map((candidate) => normalizeCleanupCandidate(candidate));
  if (includeTmp) {
    plan.push(...cleanupTmpPlan());
  }
  if (scope === "clean" || scope === "distclean") {
    plan.push({
      ...normalizeCleanupCandidate(embeddedWebAssetsDir),
      action: "remove-children",
      proof: "registered_embedded_web_assets_preserve_keep",
      preserve: ".keep",
    });
  }
  for (const entry of plan) {
    if (dryRun) {
      const action = entry.status === "reject" ? "retain" : entry.action;
      const reason = entry.status === "reject" ? entry.reason : entry.proof;
      stdout.write(`DRY-RUN ${action} ${entry.identity} ${reason}\n`);
      continue;
    }
    if (entry.status === "reject") {
      throw new HarnessConfigError(`refusing cleanup path ${entry.identity}: ${entry.reason}`, {
        reason: "configuration_error",
      });
    }
    if (entry.action === "remove-children") {
      removeChildren(entry);
    } else if (removeCandidate(entry)) {
      stdout.write(`removing ${entry.identity}\n`);
    }
  }
}

function removeChildren(entry) {
  if (!entry.path || !existsSync(entry.path)) {
    return;
  }
  for (const name of readdirNames(entry.path)) {
    if (name === entry.preserve) {
      continue;
    }
    rmSync(path.join(entry.path, name), { recursive: true, force: true });
  }
}

export function testRouteTokenValid(token) {
  const value = String(token ?? "");
  return /^[A-Za-z0-9._~-]{22,512}$/u.test(value);
}

export function generateTestRouteToken() {
  return randomBytes(32).toString("hex");
}

export function timingSafeTokenEqual(left, right) {
  const leftBuffer = Buffer.from(String(left ?? ""));
  const rightBuffer = Buffer.from(String(right ?? ""));
  if (leftBuffer.length !== rightBuffer.length) {
    return false;
  }
  return timingSafeEqual(leftBuffer, rightBuffer);
}

export async function validateSchema(schemaID, value) {
  const { default: Ajv } = await import("ajv/dist/2020.js");
  const schema = loadSchema(schemaID);
  const ajv = new Ajv({
    allErrors: true,
    strict: false,
    validateFormats: false,
    validateSchema: true,
  });
  for (const supportSchema of supportSchemaFiles(schemaID)) {
    const loaded = JSON.parse(readFileSync(supportSchema, "utf8"));
    ajv.addSchema(loaded, loaded.$id);
  }
  const validate = ajv.compile(schema);
  if (!validate(value)) {
    const details = ajv.errorsText(validate.errors, { separator: "\n  " });
    throw new Error(`${schemaID} validation failed:\n  ${details}`);
  }
}

export function validateSchemaSync(schemaID, value) {
  if (syncValidatorCache.has(schemaID)) {
    const cached = syncValidatorCache.get(schemaID);
    if (!cached(value)) {
      const details = cached.ajv.errorsText(cached.errors, { separator: "\n  " });
      throw new Error(`${schemaID} validation failed:\n  ${details}`);
    }
    return;
  }
  const Ajv = requireFromHarness("ajv/dist/2020");
  const schema = loadSchema(schemaID);
  const ajv = new Ajv({
    allErrors: true,
    strict: false,
    validateFormats: false,
    validateSchema: true,
  });
  for (const supportSchema of supportSchemaFiles(schemaID)) {
    const loaded = JSON.parse(readFileSync(supportSchema, "utf8"));
    ajv.addSchema(loaded, loaded.$id);
  }
  const validate = ajv.compile(schema);
  validate.ajv = ajv;
  syncValidatorCache.set(schemaID, validate);
  if (!validate(value)) {
    const details = ajv.errorsText(validate.errors, { separator: "\n  " });
    throw new Error(`${schemaID} validation failed:\n  ${details}`);
  }
}

function schemaFileName(schemaID) {
  return `${schemaID}.schema.json`;
}

function loadSchema(schemaID) {
  const file = path.join(schemasDir, schemaFileName(schemaID));
  if (!existsSync(file)) {
    throw new Error(`missing schema attachment ${file}`);
  }
  return JSON.parse(readFileSync(file, "utf8"));
}

function supportSchemaFiles(schemaID) {
  if (!existsSync(schemasDir)) {
    return [];
  }
  return readdirSync(schemasDir)
    .filter((name) => name.endsWith(".schema.json"))
    .filter((name) => name !== schemaFileName(schemaID))
    .map((name) => path.join(schemasDir, name));
}
