import { randomBytes, timingSafeEqual } from "node:crypto";
import {
  existsSync,
  lstatSync,
  readdirSync,
  realpathSync,
  readFileSync,
  rmSync,
  statSync,
  unlinkSync,
} from "node:fs";
import { createRequire } from "node:module";
import path from "node:path";
import { fileURLToPath } from "node:url";
export {
  secureMkdir,
  secureWriteFile,
} from "./artifact-writer.mjs";
import { secureDirMode, secureMkdir } from "./artifact-writer.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const requireFromHarness = createRequire(import.meta.url);
export const repoRoot = path.resolve(scriptDir, "..", "..", "..");
const schemasDir = path.join(repoRoot, "tools", "schemas");
const redactionManifestPath = path.join(
  repoRoot,
  "tools",
  "harness_redaction_manifest.json",
);

const outputModes = Object.freeze([
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
  "machine_stdout_json",
  "scheduler_summary_with_artifacts",
  "service_summary",
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
const defaultTaskSurfaceManifestPath = "tools/task_surface_manifest.json";
const protectedCleanupIdentities = new Set([
  ".git",
  "apps",
  "cmd",
  "configs",
  "contracts",
  "db/migrations",
  "db/queries",
  "docs",
  "go.mod",
  "go.sum",
  "internal",
  "package.json",
  "packages",
  "pnpm-lock.yaml",
  "pnpm-workspace.yaml",
  "scripts",
  "tools",
]);
const structuredSecretKeyTokens = new Set([
  "PASSWORD",
  "PASS",
  "PWD",
  "TOKEN",
  "JWT",
  "BEARER",
  "API_KEY",
  "ACCESS_KEY",
  "SECRET_KEY",
  "AUTHORIZATION",
  "COOKIE",
  "SET_COOKIE",
  "X_CARTULARY_TEST_ROUTE_TOKEN",
]);
const weakTestRouteTokens = new Set([
  "test",
  "token",
  "secret",
  "password",
  "changeme",
]);
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
  "CARTULARY_SERVER_HARNESS_BIN",
  "CARTULARY_MIGRATE_BIN",
  "CARTULARY_OPERATOR_BIN",
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
    label: "Object-store S3 attach set",
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
const restrictedInternalMakeVariables = Object.freeze([
  "CARTULARY_OPERATOR_BIN",
  "CARTULARY_EXECUTION_TOPOLOGY_MANIFEST",
  "CARTULARY_TASK_SURFACE_MANIFEST",
  "EXECUTION_TOPOLOGY_MANIFEST",
  "SCHEDULER_MANIFEST",
  "TASK_SURFACE_MANIFEST",
]);
const nonCanonicalPublicMakeVariables = Object.freeze([
  "GOVULNCHECK_FLAGS",
  "GOVULNCHECK_PATTERNS",
  "GOSEC_AUDIT_RUNTIME_FLAGS",
  "GOSEC_AUDIT_RUNTIME_PATTERNS",
  "GOSEC_AUDIT_RUNTIME_RULES",
  "GOSEC_AUDIT_SUPPORT_FLAGS",
  "GOSEC_AUDIT_SUPPORT_PATTERNS",
  "GOSEC_AUDIT_SUPPORT_RULES",
  "GOSEC_FLAGS",
  "GOSEC_PATTERNS",
  "GOSEC_RULES",
  "GOSEC_TARGETED_RUNTIME_FLAGS",
  "GOSEC_TARGETED_RUNTIME_PATTERNS",
  "GOSEC_TARGETED_RUNTIME_RULES",
  "STATICCHECK_CHECKS",
  "VITEST_FLAGS",
]);
const retiredPublicMakeVariables = Object.freeze([
  "BROWSER_A11Y_RESULTS_DIR",
  "BROWSER_MEASUREMENT_RESULTS_DIR",
  "BROWSER_SUPPORT_RESULTS_DIR",
  "BROWSER_VISUAL_RESULTS_DIR",
  "CHECK_RESULTS_DIR",
]);
const makeInputSourcesEnv = "CARTULARY_MAKE_INPUT_SOURCES";
const makeCommandLineOrigins = new Set(["cli"]);
const makeEnvironmentOrigins = new Set(["env"]);
const makeDefaultOrigins = new Set(["file", "unset"]);
const makeInputSourcePattern = /^([A-Z][A-Z0-9_]*)=(cli|env|file|unset)$/u;
const makeInputSourceCache = new WeakMap();

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

function defaultResultsRoot(root = repoRoot) {
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

function pathSegments(value) {
  return value.split(/[\\/]+/u).filter((segment) => segment.length > 0);
}

function isProtectedRepoPath(resolved) {
  const normalized = path.resolve(resolved);
  for (const identity of protectedCleanupIdentities) {
    const protectedPath = path.resolve(repoRoot, identity);
    if (normalized === protectedPath || normalized.startsWith(`${protectedPath}${path.sep}`)) {
      return identity;
    }
  }
  return "";
}

function assertNoSymlinkAncestor(resolved, name) {
  let current = path.resolve(resolved);
  while (current !== path.dirname(current)) {
    if (existsSync(current)) {
      const info = lstatSync(current);
      if (info.isSymbolicLink()) {
        throw new HarnessConfigError(`${name} must not resolve through a symlink`);
      }
    }
    current = path.dirname(current);
  }
}

function validateBuildOutputPathToken(name, value) {
  const normalized = validatePathToken(name, value);
  if (path.sep === "/" && normalized.includes("\\")) {
    throw new HarnessConfigError(`${name} must not contain backslash on POSIX`);
  }
  const segments = pathSegments(normalized);
  if (
    normalized === "/" ||
    normalized === "." ||
    normalized === ".." ||
    segments.includes(".") ||
    segments.includes("..")
  ) {
    throw new HarnessConfigError(`${name} must be a concrete build output path`);
  }
  const resolved = path.isAbsolute(normalized)
    ? path.resolve(normalized)
    : path.resolve(repoRoot, normalized);
  const protectedIdentity = isProtectedRepoPath(resolved);
  if (protectedIdentity) {
    throw new HarnessConfigError(`${name} must not write under protected repo root ${protectedIdentity}`);
  }
  if (existsSync(resolved)) {
    const info = lstatSync(resolved);
    if (info.isSymbolicLink()) {
      throw new HarnessConfigError(`${name} must not be a symlink`);
    }
    if (!info.isFile()) {
      throw new HarnessConfigError(`${name} must be absent or a regular file`);
    }
  }
  assertNoSymlinkAncestor(path.dirname(resolved), name);
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
  validateResultRootSecurity(resolved, { custom: path.isAbsolute(configured) });
  if (create) {
    secureMkdir(resolved);
  }
  return resolved;
}

export function validatePreparedArtifactIdentity(target, env = process.env, options = {}) {
  const marker = env.CARTULARY_HARNESS_IDENTITY_PREPARED;
  if (marker === undefined || marker === null || marker === "") {
    return false;
  }
  if (marker !== "1") {
    throw new HarnessConfigError(
      "CARTULARY_HARNESS_IDENTITY_PREPARED must be exactly 1 when set",
    );
  }
  if (target === "adhoc") {
    throw new HarnessConfigError(
      "prepared harness artifact identity requires a declared target",
    );
  }
  const required = [
    "CARTULARY_TEST_RESULTS_DIR",
    "CARTULARY_TEST_RUN_ID",
    "CARTULARY_TEST_TARGET",
  ];
  const missing = required.filter(
    (name) => !Object.hasOwn(env, name) || String(env[name] ?? "").trim() === "",
  );
  if (missing.length > 0) {
    throw new HarnessConfigError(
      `prepared harness artifact identity is incomplete; missing ${missing.join(", ")}`,
    );
  }
  if (env.CARTULARY_TEST_TARGET !== target) {
    throw new HarnessConfigError(
      `prepared harness artifact target ${JSON.stringify(env.CARTULARY_TEST_TARGET)} does not match ${JSON.stringify(target)}`,
    );
  }
  validateResultRoot(env.CARTULARY_TEST_RESULTS_DIR, {
    root: options.root ?? repoRoot,
    create: false,
  });
  validateRunId(env.CARTULARY_TEST_RUN_ID);
  return true;
}

function isWorldWritableWithoutSticky(stat) {
  return (stat.mode & 0o002) !== 0 && (stat.mode & 0o1000) === 0;
}

function validateResultRootSecurity(resolved, { custom = false } = {}) {
  const parent = path.dirname(resolved);
  const existing = existsSync(resolved) ? resolved : parent;
  if (!existsSync(existing)) {
    return;
  }
  const stat = statSync(existing);
  if (!stat.isDirectory()) {
    throw new HarnessConfigError(`CARTULARY_TEST_RESULTS_DIR parent is not a directory: ${existing}`);
  }
  if (custom && isWorldWritableWithoutSticky(stat)) {
    throw new HarnessConfigError(
      `CARTULARY_TEST_RESULTS_DIR must not use a world-writable directory without sticky bit: ${existing}`,
    );
  }
  if (custom) {
    return;
  }
  if (existsSync(resolved) && statSync(resolved).isDirectory()) {
    secureMkdir(resolved, secureDirMode);
  }
}

export function generateRunId(now = new Date(), pid = process.pid) {
  return `${now.toISOString().replace(/[-:]/gu, "").replace(/\..+$/u, "Z")}-p${pid}`;
}

function targetEmitsRetainedArtifacts(entry) {
  return (entry?.output_policy?.artifact_policy ?? "none") !== "none";
}

function runRootFor(resultRoot, runId) {
  return path.join(resultRoot, runId);
}

function pathExists(value) {
  return existsSync(value);
}

function assertReusableCallerRunRoot(
  runRoot,
  runId,
  { allowExistingRunRoot = false } = {},
) {
  if (!pathExists(runRoot)) {
    secureMkdir(runRoot);
    return;
  }
  const stat = lstatSync(runRoot);
  if (!stat.isDirectory()) {
    throw new HarnessConfigError(
      `CARTULARY_TEST_RUN_ID ${JSON.stringify(runId)} resolves to a non-directory run root`,
    );
  }
  const entries = readdirSync(runRoot);
  if (entries.length > 0) {
    if (allowExistingRunRoot) {
      return;
    }
    throw new HarnessConfigError(
      `CARTULARY_TEST_RUN_ID ${JSON.stringify(runId)} resolves to a non-empty run root`,
    );
  }
}

function firstAvailableGeneratedRunId(resultRoot, baseRunId) {
  if (!pathExists(runRootFor(resultRoot, baseRunId))) {
    return baseRunId;
  }
  for (let suffix = 1; suffix < 1_000_000; suffix += 1) {
    const candidate = `${baseRunId}-n${suffix}`;
    validateRunId(candidate);
    if (!pathExists(runRootFor(resultRoot, candidate))) {
      return candidate;
    }
  }
  throw new HarnessConfigError(
    `could not allocate a non-colliding generated CARTULARY_TEST_RUN_ID for ${baseRunId}`,
  );
}

function prepareRetainedArtifactRunRoot(
  resolved,
  { allowExistingRunRoot = false, materializeGeneratedRunId = false } = {},
) {
  if (resolved.generated_run_id && !materializeGeneratedRunId) {
    return resolved;
  }
  const runId = resolved.generated_run_id
    ? firstAvailableGeneratedRunId(resolved.result_root, resolved.run_id)
    : resolved.run_id;
  const runRoot = runRootFor(resolved.result_root, runId);
  if (resolved.generated_run_id) {
    secureMkdir(runRoot);
  } else {
    assertReusableCallerRunRoot(runRoot, runId, { allowExistingRunRoot });
    secureMkdir(runRoot);
  }
  return {
    ...resolved,
    run_id: runId,
    run_root: runRoot,
  };
}

function loadTaskSurfaceManifest(manifestPath = process.env.TASK_SURFACE_MANIFEST) {
  const file = path.resolve(repoRoot, manifestPath || defaultTaskSurfaceManifestPath);
  return JSON.parse(readFileSync(file, "utf8"));
}

function isMakePreflightEnv(env) {
  return Object.hasOwn(env, makeInputSourcesEnv);
}

function loadResolverTaskSurfaceManifest(env) {
  if (isMakePreflightEnv(env)) {
    return loadTaskSurfaceManifest(defaultTaskSurfaceManifestPath);
  }
  return loadTaskSurfaceManifest();
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

function inputContract(entry) {
  return entry?.input_contract ?? {
    undeclared_make_command_line: "usage_error",
    undeclared_inherited_env: "ignore",
    inputs: [],
  };
}

function inputRows(entry) {
  return inputContract(entry).inputs ?? [];
}

function inputRowMap(entry) {
  return new Map(inputRows(entry).map((input) => [input.name, input]));
}

function publicInputNames(manifest) {
  const names = new Set([
    ...restrictedInternalMakeVariables,
    ...nonCanonicalPublicMakeVariables,
    ...retiredPublicMakeVariables,
  ]);
  for (const entry of manifest.targets ?? []) {
    if (entry?.target_class !== "public") {
      continue;
    }
    for (const input of inputRows(entry)) {
      names.add(input.name);
    }
  }
  return Array.from(names).sort((left, right) => left.localeCompare(right));
}

function makeInputSources(env) {
  const cached = makeInputSourceCache.get(env);
  if (cached) {
    return cached;
  }
  const sources = new Map();
  const raw = String(env[makeInputSourcesEnv] ?? "").trim();
  for (const token of raw === "" ? [] : raw.split(/\s+/u)) {
    const match = makeInputSourcePattern.exec(token);
    if (!match) {
      throw new HarnessConfigError(
        `${makeInputSourcesEnv} contains invalid source token ${JSON.stringify(token)}`,
      );
    }
    const [, name, source] = match;
    if (sources.has(name)) {
      throw new HarnessConfigError(`${makeInputSourcesEnv} contains duplicate ${name}`);
    }
    sources.set(name, source);
  }
  makeInputSourceCache.set(env, sources);
  return sources;
}

function makeOrigin(env, name) {
  return makeInputSources(env).get(name) ?? "unset";
}

function isMakeCommandLineOrigin(origin) {
  return makeCommandLineOrigins.has(origin);
}

function isMakeEnvironmentOrigin(origin) {
  return makeEnvironmentOrigins.has(origin);
}

function isMakeDefaultOrigin(origin) {
  return origin === "" || makeDefaultOrigins.has(origin);
}

function normalizeInputValue(input, raw) {
  let value = String(raw ?? "");
  if (input.normalization === "trim" || input.normalization === "path_token") {
    value = value.trim();
  } else if (input.normalization === "trim_lowercase") {
    value = value.trim().toLowerCase();
  }
  return value;
}

function validatePositiveDecimalInput(name, value, input) {
  if (!/^(?:[0-9]+(?:\.[0-9]+)?|\.[0-9]+)$/u.test(value)) {
    throw new HarnessConfigError(`${name} must be a positive decimal`, {
      reason: input.invalid_reason,
    });
  }
  const parsed = Number(value);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    throw new HarnessConfigError(`${name} must be a positive decimal`, {
      reason: input.invalid_reason,
    });
  }
  if (Number.isFinite(input.min) && parsed < input.min) {
    throw new HarnessConfigError(`${name} must be >= ${input.min}`, {
      reason: input.invalid_reason,
    });
  }
  if (Number.isFinite(input.max) && parsed > input.max) {
    throw new HarnessConfigError(`${name} must be <= ${input.max}`, {
      reason: input.invalid_reason,
    });
  }
  return value;
}

function validateTargetInputValue(name, value, input, manifest) {
  if (input.type === "exact_1_bool") {
    if (value !== "1") {
      throw new HarnessConfigError(`${name} must be exactly 1 when set`, {
        reason: input.invalid_reason,
      });
    }
    return value;
  }
  if (input.type === "enum") {
    if (!(input.values ?? []).includes(value)) {
      throw new HarnessConfigError(
        `${name} must be one of ${(input.values ?? []).join(", ")}`,
        { reason: input.invalid_reason },
      );
    }
    return value;
  }
  if (input.type === "owner_id") {
    if (!/^(?:module|platform|app|web|package|harness)\.[a-z][a-z0-9_]{0,62}$/u.test(value)) {
      throw new HarnessConfigError(`${name} must be an owner ID`, {
        reason: input.invalid_reason,
      });
    }
    return value;
  }
  if (input.type === "row_ids") {
    const rowIDPattern = /^(?:module|platform|app|web|package|harness)\.[a-z][a-z0-9_]{0,62}\.[a-z][a-z0-9_]{0,62}\.[a-z][a-z0-9_]{0,127}_[0-9a-f]{10}$/u;
    const tokens = value.split(",").map((token) => token.trim());
    if (
      tokens.some((token) => token === "" || !rowIDPattern.test(token)) ||
      new Set(tokens).size !== tokens.length
    ) {
      throw new HarnessConfigError(`${name} must contain unique owner row IDs`, {
        reason: input.invalid_reason,
      });
    }
    return tokens.join(",");
  }
  if (input.type === "target_name") {
    const knownTargets = new Set((manifest.targets ?? []).map((entry) => entry.name));
    if (!knownTargets.has(value)) {
      throw new HarnessConfigError(`${name} must name a declared target`, {
        reason: input.invalid_reason,
      });
    }
    return value;
  }
  if (input.type === "run_id") {
    try {
      return validateRunId(value);
    } catch (error) {
      if (error instanceof HarnessConfigError) {
        throw new HarnessConfigError(`${name}: ${error.message}`, {
          reason: input.invalid_reason,
        });
      }
      throw error;
    }
  }
  if (input.type === "result_selector" || input.type === "path") {
    try {
      if (name === "OPERATOR_BIN") {
        return validateBuildOutputPathToken(name, value);
      }
      return validatePathToken(name, value);
    } catch (error) {
      if (error instanceof HarnessConfigError) {
        throw new HarnessConfigError(error.message, { reason: input.invalid_reason });
      }
      throw error;
    }
  }
  if (input.type === "positive_integer") {
    try {
      return String(
        validatePositiveInteger(name, value, {
          min: Number.isInteger(input.min) ? input.min : 1,
          max: Number.isInteger(input.max) ? input.max : 999999999,
        }),
      );
    } catch (error) {
      if (error instanceof HarnessConfigError) {
        throw new HarnessConfigError(error.message, { reason: input.invalid_reason });
      }
      throw error;
    }
  }
  if (input.type === "positive_decimal") {
    return validatePositiveDecimalInput(name, value, input);
  }
  if (input.type === "task_surface_report_args") {
    const allowed = new Set(["--all", "--check", "--check --all", "--all --check"]);
    if (!allowed.has(value)) {
      throw new HarnessConfigError(
        `${name} accepts only --all, --check, or --check --all`,
        { reason: input.invalid_reason },
      );
    }
    return value;
  }
  return value;
}

function resolveDeclaredTargetInputs(target, entry, manifest, env) {
  const resolved = {};
  for (const input of inputRows(entry)) {
    const origin = makeOrigin(env, input.name);
    const hasEnvValue = Object.hasOwn(env, input.name);
    const raw = env[input.name] ?? "";
    if (!hasEnvValue || isMakeDefaultOrigin(origin)) {
      if (input.required) {
        throw new HarnessConfigError(`${input.name} is required for target ${target}`, {
          reason: input.invalid_reason,
        });
      }
      if (input.default !== null && input.default !== undefined) {
        resolved[input.name] = {
          value: String(input.default),
          source: "default",
          summary_emission: input.summary_emission,
        };
      }
      continue;
    }
    if (isMakeEnvironmentOrigin(origin) && !(input.allowed_sources ?? []).includes("environment")) {
      continue;
    }
    const normalized = normalizeInputValue(input, raw);
    if (normalized === "") {
      if (input.empty_string === "invalid" || input.required) {
        throw new HarnessConfigError(`${input.name} must not be empty`, {
          reason: input.invalid_reason,
        });
      }
      if (input.empty_string === "false") {
        resolved[input.name] = {
          value: "false",
          source: origin || "omitted",
          summary_emission: input.summary_emission,
        };
      }
      if (
        input.empty_string === "omitted" &&
        input.default !== null &&
        input.default !== undefined
      ) {
        resolved[input.name] = {
          value: String(input.default),
          source: "default",
          summary_emission: input.summary_emission,
        };
      }
      continue;
    }
    resolved[input.name] = {
      value: validateTargetInputValue(input.name, normalized, input, manifest),
      source: isMakeCommandLineOrigin(origin)
        ? "make_command_line"
        : isMakeEnvironmentOrigin(origin)
          ? "environment"
          : origin || "environment",
      summary_emission: input.summary_emission,
    };
  }
  return resolved;
}

function rejectUndeclaredPublicInputs(target, entry, manifest, env) {
  const declared = inputRowMap(entry);
  const knownNames = new Set(publicInputNames(manifest));
  for (const name of makeInputSources(env).keys()) {
    if (!knownNames.has(name)) {
      throw new HarnessConfigError(`${makeInputSourcesEnv} contains unknown input ${name}`);
    }
  }
  for (const name of knownNames) {
    const origin = makeOrigin(env, name);
    if (!isMakeCommandLineOrigin(origin)) {
      continue;
    }
    if (restrictedInternalMakeVariables.includes(name)) {
      throw new HarnessConfigError(
        `${name} is an internal harness input and cannot be overridden by make ${target}`,
        { reason: "configuration_error" },
      );
    }
    if (!declared.has(name)) {
      throw new HarnessConfigError(
        `${name} is not declared for target ${target}`,
        { reason: "usage_error" },
      );
    }
  }
}

export function resolveHarnessConfig(target, env = process.env, options = {}) {
  const manifest = options.manifest ?? loadResolverTaskSurfaceManifest(env);
  const entry = targetPolicy(target, manifest);
  if (!entry || entry.target_class !== "public") {
    throw new HarnessConfigError(`unknown public target ${JSON.stringify(target)}`);
  }
  const preparedArtifactIdentity = validatePreparedArtifactIdentity(target, env, options);
  rejectUndeclaredPublicInputs(target, entry, manifest, env);
  const targetInputs = resolveDeclaredTargetInputs(target, entry, manifest, env);
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
  const runId = selectedSource(
    env,
    "CARTULARY_TEST_RUN_ID",
    generateRunId(options.now ?? new Date(), options.pid ?? process.pid),
  );
  const emitsRetainedArtifacts = targetEmitsRetainedArtifacts(entry);
  const shouldPrepareRetainedArtifacts =
    options.prepareRetainedArtifacts === true && emitsRetainedArtifacts;
  const resolved = {
    target,
    target_policy: {
      target_class: entry.target_class,
    },
    output_class: outputClass,
    artifact_policy: entry.output_policy?.artifact_policy ?? "none",
    output_mode: outputMode.value,
    output_mode_source: outputMode.source,
    result_root: validateResultRoot(resultRoot.value, {
      root: options.root ?? repoRoot,
      create: options.createResultRoot === true || shouldPrepareRetainedArtifacts,
    }),
    result_root_source: resultRoot.source,
    run_id: validateRunId(runId.value),
    run_id_source: runId.source,
    generated_run_id: runId.source === "default",
    variables: {
      target_inputs: targetInputs,
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
  resolved.run_root = runRootFor(resolved.result_root, resolved.run_id);
  if (!shouldPrepareRetainedArtifacts) {
    return resolved;
  }
  return prepareRetainedArtifactRunRoot(resolved, {
    allowExistingRunRoot:
      options.allowExistingRunRoot === true ||
      preparedArtifactIdentity ||
      env.CARTULARY_SUPPRESS_CHILD_SUCCESS === "1" ||
      env.CARTULARY_ALLOW_EXISTING_RUN_ROOT === "1",
    materializeGeneratedRunId: options.materializeGeneratedRunId === true,
  });
}

export function preflightPublicTarget(target, env = process.env) {
  return resolveHarnessConfig(target, env, {
    prepareRetainedArtifacts: true,
    materializeGeneratedRunId: false,
  });
}

export function resolveRetainedArtifactIdentity(target, env = process.env, options = {}) {
  return resolveHarnessConfig(target, env, {
    ...options,
    prepareRetainedArtifacts: true,
    materializeGeneratedRunId: true,
  });
}

export function resolveArtifactIdentityForTarget(target, env = process.env, options = {}) {
  const manifest = options.manifest ?? loadTaskSurfaceManifest();
  const entry = targetPolicy(target, manifest);
  if (entry?.target_class === "public") {
    return resolveRetainedArtifactIdentity(target, env, { ...options, manifest });
  }
  validatePreparedArtifactIdentity(target, env, options);
  const resultRoot = validateResultRoot(env.CARTULARY_TEST_RESULTS_DIR, {
    root: options.root ?? repoRoot,
    create: true,
  });
  const runId = Object.hasOwn(env, "CARTULARY_TEST_RUN_ID")
    ? validateRunId(env.CARTULARY_TEST_RUN_ID)
    : generateRunId(options.now ?? new Date(), options.pid ?? process.pid);
  const runRoot = runRootFor(resultRoot, runId);
  secureMkdir(runRoot);
  return {
    target,
    result_root: resultRoot,
    result_root_source: Object.hasOwn(env, "CARTULARY_TEST_RESULTS_DIR") ? "env" : "default",
    run_id: runId,
    run_id_source: Object.hasOwn(env, "CARTULARY_TEST_RUN_ID") ? "env" : "default",
    generated_run_id: !Object.hasOwn(env, "CARTULARY_TEST_RUN_ID"),
    run_root: runRoot,
  };
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
  const tokens = manifest.redaction_tokens ?? {};
  redactionRules = {
    replacement,
    tokens,
    valuePatterns: (manifest.value_patterns ?? []).map((rule) => ({
      name: rule.name,
      regex: new RegExp(String(rule.pattern).replace(/^\(\?i\)/u, ""), "giu"),
      replacement: rule.replacement ?? replacement,
    })),
  };
  return redactionRules;
}

function isSensitiveCLIFlag(value) {
  return /^--(?:password|passwd|pwd|secret|token|jwt|api[_-]?key|access[_-]?key|secret[_-]?key|private[_-]?key|client[_-]?secret|dsn)$/iu.test(
    String(value ?? ""),
  );
}

function canonicalStructuredKey(value) {
  return String(value ?? "")
    .toUpperCase()
    .replace(/[^A-Z0-9]+/gu, "_")
    .replace(/^_+|_+$/gu, "");
}

function isStructuredSecretKey(key) {
  const canonical = canonicalStructuredKey(key);
  if (!canonical) {
    return false;
  }
  if (structuredSecretKeyTokens.has(canonical)) {
    return true;
  }
  for (const token of structuredSecretKeyTokens) {
    if (canonical.startsWith(`${token}_`) || canonical.endsWith(`_${token}`)) {
      return true;
    }
  }
  return false;
}

export function redactString(value) {
  const rules = compiledRedactionRules();
  let text = String(value);
  for (const rule of rules.valuePatterns) {
    text = text.replace(rule.regex, rule.replacement);
  }
  return text;
}

function redactStructuredText(value) {
  const text = String(value ?? "");
  try {
    return compactJSONString(JSON.parse(text));
  } catch {
    return redactString(text);
  }
}

export function redactValue(value, key = "") {
  const rules = compiledRedactionRules();
  if (typeof key === "string" && isStructuredSecretKey(key)) {
    return rules.replacement;
  }
  if (typeof value === "string") {
    return redactString(value);
  }
  if (Array.isArray(value)) {
    const redacted = [];
    let redactNext = false;
    for (const entry of value) {
      if (redactNext) {
        redacted.push(rules.replacement);
        redactNext = false;
        continue;
      }
      redacted.push(redactValue(entry));
      if (typeof entry === "string" && isSensitiveCLIFlag(entry)) {
        redactNext = true;
      }
    }
    return redacted;
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
  if (protectedCleanupIdentities.has(identity)) {
    return { status: "reject", identity, reason: "protected_root" };
  }
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
  assertCleanupTraversalSafe(entry);
  const stat = lstatSync(entry.path);
  if (stat.isSymbolicLink()) {
    unlinkSync(entry.path);
  } else {
    rmSync(entry.path, { recursive: true, force: true });
  }
  return true;
}

function assertCleanupTraversalSafe(entry) {
  const candidateRoot = entry.path;
  const stat = lstatSync(candidateRoot);
  if (!stat.isDirectory() || stat.isSymbolicLink()) {
    return;
  }
  const rootRealPath = realpathSync(candidateRoot);
  const visit = (current) => {
    const currentStat = lstatSync(current);
    const relative = path.relative(candidateRoot, current);
    if (relative.startsWith("..") || path.isAbsolute(relative)) {
      throw new HarnessConfigError(`cleanup path escapes candidate root: ${current}`);
    }
    if (currentStat.isSymbolicLink() || !currentStat.isDirectory()) {
      return;
    }
    const currentRealPath = realpathSync(current);
    if (!isUnderPath(rootRealPath, currentRealPath)) {
      throw new HarnessConfigError(`cleanup traversal escapes candidate root: ${current}`);
    }
    for (const name of readdirNames(current)) {
      visit(path.join(current, name));
    }
  };
  visit(candidateRoot);
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
  assertCleanupTraversalSafe(entry);
  for (const name of readdirNames(entry.path)) {
    if (name === entry.preserve) {
      continue;
    }
    rmSync(path.join(entry.path, name), { recursive: true, force: true });
  }
}

export function testRouteTokenValid(token) {
  const value = String(token ?? "");
  if (value.length < 43 || value.length > 512) {
    return false;
  }
  if (!/^[!-~]+$/u.test(value)) {
    return false;
  }
  const lower = value.toLowerCase();
  if (weakTestRouteTokens.has(lower)) {
    return false;
  }
  if (/^(.)\1+$/u.test(value)) {
    return false;
  }
  return true;
}

export function generateTestRouteToken() {
  return randomBytes(32).toString("base64url");
}

function timingSafeTokenEqual(left, right) {
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
