#!/usr/bin/env node

import { createHash } from "node:crypto";
import {
  closeSync,
  existsSync,
  fsyncSync,
  lstatSync,
  mkdirSync,
  openSync,
  readFileSync,
  readdirSync,
  renameSync,
  statSync,
  writeFileSync,
  writeSync,
} from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { validateSchemaSync } from "../contract/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../..");
const digestPattern = /^sha256:[0-9a-f]{64}$/u;
const identityPattern = /^[a-zA-Z0-9_.-]+$/u;

function usage() {
  return "usage: browser-session-evidence.mjs event <state> <message> [failure-class failure-reason] | terminal <ready|failed> <message> [failure-class failure-reason] | snapshot-service-scope | lease | stack | attach <stack-v4.json> | attach-json <stack-v4.json>";
}

function requiredEnv(name) {
  const value = (process.env[name] ?? "").trim();
  if (!value) throw new Error(`${name} is required`);
  return value;
}

function identityEnv(name) {
  const value = requiredEnv(name);
  if (!identityPattern.test(value)) throw new Error(`${name} has an unsafe identity`);
  return value;
}

function runRoot() {
  const results = requiredEnv("CARTULARY_TEST_RESULTS_DIR");
  const runID = identityEnv("CARTULARY_TEST_RUN_ID");
  return path.resolve(repoRoot, results, runID);
}

function sessionRoot() {
  const configured = path.resolve(requiredEnv("CARTULARY_WEB_E2E_SESSION_ARTIFACT_DIR"));
  const root = runRoot();
  if (configured !== root && !configured.startsWith(`${root}${path.sep}`)) {
    throw new Error("browser session artifact directory must be beneath the current run root");
  }
  return configured;
}

function requireRegularNoSymlink(file, label) {
  const info = lstatSync(file);
  if (!info.isFile() || info.isSymbolicLink()) {
    throw new Error(`${label} must be a non-symlink regular file`);
  }
}

function relativeToRun(file) {
  const root = runRoot();
  const resolved = path.resolve(file);
  const relative = path.relative(root, resolved).replaceAll("\\", "/");
  if (!relative || relative.startsWith("../") || path.isAbsolute(relative)) {
    throw new Error(`artifact path escapes current run root: ${file}`);
  }
  return relative;
}

function sha256Bytes(bytes) {
  return `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
}

function sha256File(file) {
  requireRegularNoSymlink(file, file);
  return sha256Bytes(readFileSync(file));
}

function normalizeDigest(value, label) {
  const normalized = String(value ?? "").trim().toLowerCase();
  if (digestPattern.test(normalized)) return normalized;
  if (/^[0-9a-f]{64}$/u.test(normalized)) return `sha256:${normalized}`;
  throw new Error(`${label} must be a SHA-256 digest`);
}

function atomicWrite(file, contents) {
  mkdirSync(path.dirname(file), { recursive: true, mode: 0o700 });
  const temporary = `${file}.tmp-${process.pid}-${Date.now()}`;
  const fd = openSync(temporary, "wx", 0o600);
  try {
    writeFileSync(fd, contents);
    fsyncSync(fd);
  } finally {
    closeSync(fd);
  }
  renameSync(temporary, file);
  const parentFD = openSync(path.dirname(file), "r");
  try {
    fsyncSync(parentFD);
  } finally {
    closeSync(parentFD);
  }
}

function artifactRefs() {
  return [
    process.env.CARTULARY_WEB_E2E_SERVER_LOG,
    process.env.CARTULARY_WEB_E2E_WEB_LOG,
  ]
    .filter((value) => typeof value === "string" && value.trim() !== "")
    .map(relativeToRun)
    .sort();
}

function eventPath() {
  return path.join(sessionRoot(), "startup-events.jsonl");
}

function diagnosticPath() {
  return path.join(sessionRoot(), "startup-diagnostics.json");
}

function readEvents(file) {
  if (!existsSync(file)) return [];
  requireRegularNoSymlink(file, "startup event stream");
  return readFileSync(file, "utf8")
    .trim()
    .split(/\r?\n/u)
    .filter(Boolean)
    .map((line, index) => {
      const event = JSON.parse(line);
      validateSchemaSync(event.schema_id, event);
      if (event.sequence !== index + 1) {
        throw new Error("browser startup event sequence is not contiguous");
      }
      return event;
    });
}

function validateStateTransition(events, state) {
  const previous = events.at(-1)?.state;
  const allowed = new Map([
    [undefined, new Set(["initializing", "failed"])],
    ["initializing", new Set(["service_attached", "failed"])],
    ["service_attached", new Set(["fixture_ready", "failed"])],
    ["fixture_ready", new Set(["backend_ready", "failed"])],
    ["backend_ready", new Set(["frontend_ready", "failed"])],
    ["frontend_ready", new Set(["ready", "failed"])],
  ]);
  if (!allowed.get(previous)?.has(state)) {
    throw new Error(
      `invalid browser startup state transition ${previous ?? "<start>"} -> ${state}`,
    );
  }
}

function appendEvent(state, message, failureClass = "", failureReason = "") {
  if (existsSync(diagnosticPath())) {
    throw new Error("browser startup diagnostic is terminal and immutable");
  }
  const file = eventPath();
  mkdirSync(path.dirname(file), { recursive: true, mode: 0o700 });
  const events = readEvents(file);
  validateStateTransition(events, state);
  const payload = {
    schema_id: "cartulary.browser_startup_event.v1",
    suite_id: identityEnv("CARTULARY_TEST_SUITE_ID"),
    browser_session_id: identityEnv("CARTULARY_BROWSER_SESSION_GROUP"),
    runtime_profile_id: identityEnv("CARTULARY_BROWSER_RUNTIME_PROFILE_ID"),
    sequence: events.length + 1,
    emitted_at: new Date().toISOString(),
    monotonic_ms: Math.floor(process.uptime() * 1000),
    state,
    ...(failureClass ? { failure_class: failureClass } : {}),
    ...(failureReason ? { failure_reason: failureReason } : {}),
    message: String(message).slice(0, 2048),
    artifact_refs: artifactRefs(),
  };
  validateSchemaSync(payload.schema_id, payload);
  const fd = openSync(file, "a", 0o600);
  try {
    writeSync(fd, `${JSON.stringify(payload)}\n`);
    fsyncSync(fd);
  } finally {
    closeSync(fd);
  }
  return payload;
}

function terminal(status, message, failureClass = "", failureReason = "") {
  const output = diagnosticPath();
  if (existsSync(output)) {
    const existing = JSON.parse(readFileSync(output, "utf8"));
    validateSchemaSync(existing.schema_id, existing);
    if (existing.status !== status) {
      throw new Error("browser startup diagnostic is terminal and immutable");
    }
    return existing;
  }
  appendEvent(
    status === "ready" ? "ready" : "failed",
    message,
    failureClass,
    failureReason,
  );
  const events = eventPath();
  const parsedEvents = readEvents(events);
  const expectedStates = [
    "initializing",
    "service_attached",
    "fixture_ready",
    "backend_ready",
    "frontend_ready",
    "ready",
  ];
  if (
    status === "ready" &&
    parsedEvents.map((event) => event.state).join("\0") !==
      expectedStates.join("\0")
  ) {
    throw new Error("browser ready diagnostic requires the complete startup state graph");
  }
  const payload = {
    schema_id: "cartulary.browser_startup_diagnostics.v2",
    suite_id: identityEnv("CARTULARY_TEST_SUITE_ID"),
    browser_session_id: identityEnv("CARTULARY_BROWSER_SESSION_GROUP"),
    runtime_profile_id: identityEnv("CARTULARY_BROWSER_RUNTIME_PROFILE_ID"),
    generated_at: new Date().toISOString(),
    status,
    startup_phase: status === "ready" ? "ready" : "failed",
    ...(failureClass ? { failure_class: failureClass } : {}),
    ...(failureReason ? { failure_reason: failureReason } : {}),
    message: String(message).slice(0, 2048),
    events_ref: relativeToRun(events),
    events_sha256: sha256File(events),
    frontend_mode: "preview",
    frontend_command_kind: "vite-preview",
    ...(process.env.CARTULARY_WEB_E2E_API_ORIGIN
      ? { api_origin: process.env.CARTULARY_WEB_E2E_API_ORIGIN }
      : {}),
    ...(process.env.CARTULARY_WEB_E2E_PUBLIC_ORIGIN
      ? { public_origin: process.env.CARTULARY_WEB_E2E_PUBLIC_ORIGIN }
      : {}),
    artifact_refs: artifactRefs(),
  };
  validateSchemaSync(payload.schema_id, payload);
  atomicWrite(output, `${JSON.stringify(payload, null, 2)}\n`);
  return payload;
}

function snapshotServiceScope() {
  const suiteRoot = path.join(
    runRoot(),
    "_shared",
    "test-services",
    identityEnv("CARTULARY_TEST_SUITE_ID"),
  );
  const source = path.join(suiteRoot, "service-scope.json");
  requireRegularNoSymlink(source, "test-services service scope");
  const destination = path.join(sessionRoot(), "service-scope-admission.json");
  atomicWrite(destination, readFileSync(source));
  return destination;
}

function processProof(pidValue, processGroupID) {
  const pid = Number.parseInt(String(pidValue), 10);
  if (!Number.isInteger(pid) || pid < 1) throw new Error(`invalid process id ${pidValue}`);
  const statText = readFileSync(`/proc/${pid}/stat`, "utf8");
  const close = statText.lastIndexOf(")");
  if (close < 0) throw new Error(`invalid /proc/${pid}/stat`);
  const fields = statText.slice(close + 2).trim().split(/\s+/u);
  const startTimeTicks = Number.parseInt(fields[19] ?? "", 10);
  const status = readFileSync(`/proc/${pid}/status`, "utf8");
  const uidLine = status.split(/\r?\n/u).find((line) => line.startsWith("Uid:"));
  const effectiveUID = Number.parseInt(uidLine?.trim().split(/\s+/u)[2] ?? "", 10);
  const executable = `/proc/${pid}/exe`;
  const executableStat = statSync(executable);
  return {
    status: "pass",
    pid,
    process_group_id: Number.parseInt(String(processGroupID), 10),
    boot_id: readFileSync("/proc/sys/kernel/random/boot_id", "utf8").trim(),
    start_time_ticks: startTimeTicks,
    effective_uid: effectiveUID,
    executable_device: executableStat.dev,
    executable_inode: executableStat.ino,
    executable_sha256: sha256Bytes(readFileSync(executable)),
  };
}

function writeLease() {
  const payload = {
    schema_id: "cartulary.web_e2e_stack_lease.v1",
    suite_id: identityEnv("CARTULARY_TEST_SUITE_ID"),
    browser_session_id: identityEnv("CARTULARY_BROWSER_SESSION_GROUP"),
    runtime_profile_id: identityEnv("CARTULARY_BROWSER_RUNTIME_PROFILE_ID"),
    backend_process_group_id: Number.parseInt(requiredEnv("CARTULARY_WEB_E2E_SERVER_PGID"), 10),
    frontend_process_group_id: Number.parseInt(requiredEnv("CARTULARY_WEB_E2E_VITE_PGID"), 10),
    backend_port: Number.parseInt(requiredEnv("CARTULARY_WEB_E2E_BACKEND_PORT"), 10),
    frontend_port: Number.parseInt(requiredEnv("CARTULARY_WEB_E2E_FRONTEND_PORT"), 10),
    runtime_root: requiredEnv("CARTULARY_WEB_E2E_RUNTIME_ROOT"),
    created_at: new Date().toISOString(),
  };
  validateSchemaSync(payload.schema_id, payload);
  const output = path.join(sessionRoot(), "browser-stack-lease.json");
  atomicWrite(output, `${JSON.stringify(payload, null, 2)}\n`);
  return output;
}

function directoryDigest(directory) {
  const hash = createHash("sha256");
  function walk(current) {
    const names = readdirSync(current).sort();
    for (const name of names) {
      const absolute = path.join(current, name);
      const relative = path.relative(directory, absolute).replaceAll("\\", "/");
      const info = lstatSync(absolute);
      if (info.isSymbolicLink()) throw new Error(`build artifact contains symlink ${relative}`);
      if (info.isDirectory()) {
        hash.update(`d\0${relative}\0`);
        walk(absolute);
      } else if (info.isFile()) {
        hash.update(`f\0${relative}\0${info.mode & 0o777}\0`);
        hash.update(readFileSync(absolute));
        hash.update("\0");
      } else {
        throw new Error(`build artifact contains unsupported entry ${relative}`);
      }
    }
  }
  walk(directory);
  return `sha256:${hash.digest("hex")}`;
}

function endpointOrigin(endpoint, secure) {
  const raw = String(endpoint ?? "").trim();
  const candidate = raw.includes("://")
    ? raw
    : `${secure === "true" ? "https" : "http"}://${raw}`;
  const parsed = new URL(candidate);
  if (
    !["http:", "https:"].includes(parsed.protocol) ||
    !["127.0.0.1", "localhost"].includes(parsed.hostname) ||
    parsed.username ||
    parsed.password ||
    parsed.search ||
    parsed.hash ||
    parsed.pathname !== "/"
  ) {
    throw new Error("object-store endpoint must be a credential-free loopback origin");
  }
  return parsed.origin;
}

function canonicalDigest(value) {
  return sha256Bytes(Buffer.from(JSON.stringify(value)));
}

function writeStack() {
  const diagnostic = diagnosticPath();
  const lease = path.join(sessionRoot(), "browser-stack-lease.json");
  const serviceScope = path.join(sessionRoot(), "service-scope-admission.json");
  const metadataFile = requiredEnv("CARTULARY_WEB_E2E_TEST_SERVICES_METADATA_FILE");
  for (const [file, label] of [
    [diagnostic, "startup diagnostic"],
    [lease, "browser stack lease"],
    [serviceScope, "service scope snapshot"],
    [metadataFile, "browser fixture metadata"],
  ]) {
    requireRegularNoSymlink(file, label);
  }
  const terminalDiagnostic = JSON.parse(readFileSync(diagnostic, "utf8"));
  if (
    terminalDiagnostic.schema_id !== "cartulary.browser_startup_diagnostics.v2" ||
    terminalDiagnostic.status !== "ready"
  ) {
    throw new Error("v4 stack publication requires a terminal ready diagnostic");
  }
  const fixture = JSON.parse(readFileSync(metadataFile, "utf8"));
  const buildDirectory = path.join(repoRoot, "apps", "web", "dist");
  const runtimeProfileFingerprint = normalizeDigest(
    requiredEnv("CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT"),
    "runtime profile fingerprint",
  );
  const configurationFingerprint = canonicalDigest({
    runtime_profile_fingerprint: runtimeProfileFingerprint,
    api_origin: requiredEnv("CARTULARY_WEB_E2E_API_ORIGIN"),
    public_origin: requiredEnv("CARTULARY_WEB_E2E_PUBLIC_ORIGIN"),
    database_name: fixture.database_name,
    bucket: fixture.bucket,
  });
  const payload = {
    schema_id: "cartulary.web_e2e_stack.v4",
    suite_id: identityEnv("CARTULARY_TEST_SUITE_ID"),
    browser_session_id: identityEnv("CARTULARY_BROWSER_SESSION_GROUP"),
    service_mode: requiredEnv("CARTULARY_TEST_SERVICES_CALL_MODE"),
    runtime_profile_id: identityEnv("CARTULARY_BROWSER_RUNTIME_PROFILE_ID"),
    configuration_fingerprint: configurationFingerprint,
    service_scope_ref: relativeToRun(serviceScope),
    service_scope_sha256: sha256File(serviceScope),
    postgres_identity: {
      database_name: String(fixture.database_name),
      template_database: requiredEnv("CARTULARY_PGTEST_TEMPLATE_DB"),
      schema_hash: normalizeDigest(
        requiredEnv("CARTULARY_PGTEST_SCHEMA_HASH"),
        "PostgreSQL schema hash",
      ),
      fixture_capability: "postgres_dedicated",
    },
    object_store_identity: {
      endpoint_origin: endpointOrigin(
        requiredEnv("CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT"),
        requiredEnv("CARTULARY_S3_OBJECT_PRIMARY_SECURE"),
      ),
      bucket: String(fixture.bucket),
      fixture_generation: sha256File(metadataFile),
    },
    backend: {
      ...processProof(
        requiredEnv("CARTULARY_WEB_E2E_BACKEND_IDENTITY_SERVER_PID"),
        requiredEnv("CARTULARY_WEB_E2E_SERVER_PGID"),
      ),
      origin: requiredEnv("CARTULARY_WEB_E2E_API_ORIGIN"),
      port: Number.parseInt(requiredEnv("CARTULARY_WEB_E2E_BACKEND_PORT"), 10),
      ready_at: requiredEnv("CARTULARY_WEB_E2E_BACKEND_READY_AT"),
    },
    frontend: {
      ...processProof(
        requiredEnv("CARTULARY_WEB_E2E_VITE_PGID"),
        requiredEnv("CARTULARY_WEB_E2E_VITE_PGID"),
      ),
      origin: requiredEnv("CARTULARY_WEB_E2E_PUBLIC_ORIGIN"),
      port: Number.parseInt(requiredEnv("CARTULARY_WEB_E2E_FRONTEND_PORT"), 10),
      frontend_mode: "preview",
      frontend_command_kind: "vite-preview",
      build_artifact_ref: "apps/web/dist",
      build_artifact_sha256: directoryDigest(buildDirectory),
      ready_at: requiredEnv("CARTULARY_WEB_E2E_FRONTEND_READY_AT"),
    },
    fixture_identity: {
      fixture_capability: "browser_stack",
      fixture_id: sha256File(metadataFile),
      scenario_id: identityEnv("CARTULARY_BROWSER_SESSION_GROUP"),
    },
    startup_diagnostics_ref: relativeToRun(diagnostic),
    startup_diagnostics_sha256: sha256File(diagnostic),
    lease_ref: relativeToRun(lease),
    lease_sha256: sha256File(lease),
    ready_at: new Date().toISOString(),
  };
  validateSchemaSync(payload.schema_id, payload);
  const output = path.join(sessionRoot(), "stack-v4.json");
  if (existsSync(output)) throw new Error("v4 browser stack evidence is immutable");
  atomicWrite(output, `${JSON.stringify(payload, null, 2)}\n`);
  return output;
}

function verifyProcessProof(proof, label) {
  const current = processProof(proof.pid, proof.process_group_id);
  for (const key of [
    "pid",
    "process_group_id",
    "boot_id",
    "start_time_ticks",
    "effective_uid",
    "executable_device",
    "executable_inode",
    "executable_sha256",
  ]) {
    if (current[key] !== proof[key]) {
      throw new Error(`browser v4 attachment ${label} process proof mismatch`);
    }
  }
}

function resolveRunArtifact(reference) {
  const resolved = path.resolve(runRoot(), reference);
  const root = runRoot();
  if (!resolved.startsWith(`${root}${path.sep}`)) {
    throw new Error(`run artifact reference escapes run root: ${reference}`);
  }
  return resolved;
}

function shellQuote(value) {
  return `'${String(value).replaceAll("'", "'\"'\"'")}'`;
}

function attachmentAssignments(stackPath) {
  const resolvedStack = path.resolve(stackPath);
  const expectedRoot = sessionRoot();
  const playwrightStateDir = path.join(
    expectedRoot,
    "runtime-root",
    "playwright-state",
  );
  if (
    resolvedStack !== path.join(expectedRoot, "stack-v4.json") ||
    !resolvedStack.startsWith(`${runRoot()}${path.sep}`)
  ) {
    throw new Error("browser stack path does not identify the current session");
  }
  requireRegularNoSymlink(resolvedStack, "v4 browser stack");
  const stack = JSON.parse(readFileSync(resolvedStack, "utf8"));
  validateSchemaSync(stack.schema_id, stack);
  const expected = {
    suite_id: identityEnv("CARTULARY_TEST_SUITE_ID"),
    browser_session_id: identityEnv("CARTULARY_BROWSER_SESSION_GROUP"),
    runtime_profile_id: identityEnv("CARTULARY_BROWSER_RUNTIME_PROFILE_ID"),
  };
  for (const [key, value] of Object.entries(expected)) {
    if (stack[key] !== value) {
      throw new Error(`browser v4 attachment ${key} mismatch`);
    }
  }
  for (const [referenceKey, digestKey] of [
    ["service_scope_ref", "service_scope_sha256"],
    ["startup_diagnostics_ref", "startup_diagnostics_sha256"],
    ["lease_ref", "lease_sha256"],
  ]) {
    const artifact = resolveRunArtifact(stack[referenceKey]);
    if (sha256File(artifact) !== stack[digestKey]) {
      throw new Error(`browser v4 attachment ${referenceKey} digest mismatch`);
    }
  }
  const serviceScope = JSON.parse(
    readFileSync(resolveRunArtifact(stack.service_scope_ref), "utf8"),
  );
  if (
    serviceScope.suite_id !== stack.suite_id ||
    (serviceScope.schema_hash !== undefined &&
      normalizeDigest(serviceScope.schema_hash, "service scope schema hash") !==
        stack.postgres_identity.schema_hash)
  ) {
    throw new Error("browser v4 attachment service-scope identity mismatch");
  }
  const diagnostic = JSON.parse(
    readFileSync(resolveRunArtifact(stack.startup_diagnostics_ref), "utf8"),
  );
  validateSchemaSync(diagnostic.schema_id, diagnostic);
  if (
    diagnostic.status !== "ready" ||
    diagnostic.suite_id !== stack.suite_id ||
    diagnostic.browser_session_id !== stack.browser_session_id ||
    diagnostic.runtime_profile_id !== stack.runtime_profile_id
  ) {
    throw new Error("browser v4 attachment diagnostic identity mismatch");
  }
  if (
    stack.postgres_identity.schema_hash !==
      normalizeDigest(
        requiredEnv("CARTULARY_PGTEST_SCHEMA_HASH"),
        "active PostgreSQL schema hash",
      ) ||
    stack.postgres_identity.template_database !==
      requiredEnv("CARTULARY_PGTEST_TEMPLATE_DB") ||
    stack.object_store_identity.bucket !==
      requiredEnv("CARTULARY_S3_OBJECT_PRIMARY_BUCKET") ||
    stack.object_store_identity.endpoint_origin !==
      endpointOrigin(
        requiredEnv("CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT"),
        requiredEnv("CARTULARY_S3_OBJECT_PRIMARY_SECURE"),
      )
  ) {
    throw new Error("browser v4 attachment active service identity mismatch");
  }
  if (
    directoryDigest(path.join(repoRoot, stack.frontend.build_artifact_ref)) !==
    stack.frontend.build_artifact_sha256
  ) {
    throw new Error("browser v4 attachment frontend build digest mismatch");
  }
  verifyProcessProof(stack.backend, "backend");
  verifyProcessProof(stack.frontend, "frontend");
  mkdirSync(playwrightStateDir, { recursive: true, mode: 0o700 });
  const stateDirInfo = lstatSync(playwrightStateDir);
  if (!stateDirInfo.isDirectory() || stateDirInfo.isSymbolicLink()) {
    throw new Error(
      "browser v4 attachment Playwright state path must be a non-symlink directory",
    );
  }
  return {
    CARTULARY_WEB_E2E_ATTACHMENT_VALIDATED: "1",
    CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER: "1",
    CARTULARY_PLAYWRIGHT_STATE_DIR: playwrightStateDir,
    CARTULARY_WEB_E2E_STACK_JSON_FILE: resolvedStack,
    CARTULARY_WEB_E2E_API_ORIGIN: stack.backend.origin,
    CARTULARY_WEB_E2E_PUBLIC_ORIGIN: stack.frontend.origin,
    CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID: stack.runtime_profile_id,
    CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT:
      process.env.CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT ?? "",
    CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS: resolveRunArtifact(
      stack.startup_diagnostics_ref,
    ),
    CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS_REF: stack.startup_diagnostics_ref,
    CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS_SHA256:
      stack.startup_diagnostics_sha256,
    CARTULARY_WEB_E2E_STACK_SHA256: sha256File(resolvedStack),
  };
}

function attach(stackPath) {
  const assignments = attachmentAssignments(stackPath);
  process.stdout.write(
    Object.entries(assignments)
      .map(([key, value]) => `${key}=${shellQuote(value)}; export ${key}`)
      .join("\n") + "\n",
  );
}

function main(argv) {
  const [command, ...args] = argv;
  if (command === "event" && (args.length === 2 || args.length === 4)) {
    appendEvent(args[0], args[1], args[2], args[3]);
    return;
  }
  if (command === "terminal" && (args.length === 2 || args.length === 4)) {
    if (!["ready", "failed"].includes(args[0])) throw new Error(usage());
    terminal(args[0], args[1], args[2], args[3]);
    return;
  }
  if (command === "snapshot-service-scope" && args.length === 0) {
    snapshotServiceScope();
    return;
  }
  if (command === "lease" && args.length === 0) {
    writeLease();
    return;
  }
  if (command === "stack" && args.length === 0) {
    process.stdout.write(`${writeStack()}\n`);
    return;
  }
  if (command === "attach" && args.length === 1) {
    attach(args[0]);
    return;
  }
  if (command === "attach-json" && args.length === 1) {
    process.stdout.write(
      `${JSON.stringify(attachmentAssignments(args[0]))}\n`,
    );
    return;
  }
  throw new Error(usage());
}

try {
  main(process.argv.slice(2));
} catch (error) {
  process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
  process.exitCode = error instanceof Error && error.message === usage() ? 2 : 1;
}
