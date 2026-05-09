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

export function resolveOutputMode(env = process.env, target = "") {
  if (isPresent(env.CARTULARY_OUTPUT_MODE)) {
    const raw = String(env.CARTULARY_OUTPUT_MODE).trim();
    const mode = raw.toLowerCase();
    if (raw !== mode || !outputModeSet.has(mode)) {
      throw new HarnessConfigError(
        `invalid CARTULARY_OUTPUT_MODE ${JSON.stringify(raw)}; expected one of ${outputModes.join(", ")}`,
      );
    }
    return mode;
  }
  if (env.VERBOSE === "1") {
    return "verbose";
  }
  if (env.CI_VERBOSE === "1") {
    return "ci";
  }
  if (target === "ci" || env.CI === "1") {
    return "ci";
  }
  return "summary";
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
  const configured = String(value || ".cartulary/test-results");
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

export function preflightPublicTarget(target, env = process.env) {
  const entry = targetPolicy(target);
  if (!entry || entry.classification !== "public") {
    throw new HarnessConfigError(`unknown public target ${JSON.stringify(target)}`);
  }
  const mode = resolveOutputMode(env, target);
  const outputClass = entry.output_policy?.output_class ?? "";
  if (mode === "machine" && machineRejectedOutputClasses.has(outputClass)) {
    throw new HarnessConfigError(
      `target ${target} does not accept CARTULARY_OUTPUT_MODE=machine`,
      { reason: "usage_error" },
    );
  }
  if (mode === "machine" && !machineAcceptedOutputClasses.has(outputClass)) {
    throw new HarnessConfigError(
      `target ${target} has no machine output contract`,
      { reason: "usage_error" },
    );
  }
  validateResultRoot(env.CARTULARY_TEST_RESULTS_DIR || ".cartulary/test-results");
  validateRunId(env.CARTULARY_TEST_RUN_ID || generateRunId());
  validateResourceLimits(env);
  return { target, output_mode: mode, output_class: outputClass };
}

function validateResourceLimits(env = process.env) {
  for (const [name, value] of Object.entries(env)) {
    if (
      !name.endsWith("_JOBS") &&
      !name.endsWith("_WORKERS") &&
      !name.endsWith("_SHARDS") &&
      ![
        "BACKEND_STORE_GO_TEST_P",
        "BACKEND_INTEGRATION_GO_TEST_P",
        "GO_TEST_SERVICE_PACKAGE_PARALLELISM",
      ].includes(name)
    ) {
      continue;
    }
    if (value === undefined || value === "") {
      continue;
    }
    if (!/^(auto|[1-9][0-9]{0,3})$/u.test(String(value))) {
      throw new HarnessConfigError(`${name} must be a positive integer or auto`);
    }
  }
  const progress = env.CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS;
  if (progress !== undefined && progress !== "" && !/^[1-9][0-9]{0,8}$/u.test(String(progress))) {
    throw new HarnessConfigError(
      "CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS must be a positive integer",
    );
  }
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
