import { spawnSync } from "node:child_process";
import { createHash, randomBytes, randomUUID } from "node:crypto";
import {
  chmodSync,
  closeSync,
  existsSync,
  lstatSync,
  mkdirSync,
  openSync,
  readFileSync,
  readdirSync,
  renameSync,
  rmSync,
  statSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";

import { semanticJSONDigest, validateSchemaSync } from "../contract/index.mjs";
import { closeSuiteRuntimeRoot, createSuiteRuntime } from "../runtime/suite-runtime.mjs";

const descriptorSchemaID = "cartulary.test_services.local_session.v1";
const statusSchemaID = "cartulary.test_services.local_session_status.v1";
const sessionDurationMS = 24 * 60 * 60 * 1000;
const allowedAttachedTargets = new Set([
  "browser-e2e-stateful",
  "browser-e2e-webserver-backed",
  "service-backed-test-slice",
]);
const serviceEnvironmentNames = [
  "CARTULARY_PGTEST_ADMIN_DSN",
  "CARTULARY_PGTEST_DSN_TEMPLATE",
  "CARTULARY_PGTEST_SCHEMA_HASH",
  "CARTULARY_PGTEST_TEMPLATE_DB",
  "CARTULARY_S3TEST_ACCESS_KEY_ID",
  "CARTULARY_S3TEST_ENDPOINT",
  "CARTULARY_S3TEST_PROBE_BUCKET",
  "CARTULARY_S3TEST_SECRET_ACCESS_KEY",
  "CARTULARY_S3TEST_SECURE",
];
const configInputs = ["db/migrations", "go.mod", "go.sum", "tools/toolchain_pins.json"];
const identityPattern = /^[A-Za-z0-9_.-]+$/u;

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function sha256(bytes) {
  return `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
}

function normalizeDigest(value) {
  const trimmed = String(value).trim();
  if (/^sha256:[a-f0-9]{64}$/u.test(trimmed)) return trimmed;
  if (/^[a-f0-9]{64}$/u.test(trimmed)) return `sha256:${trimmed}`;
  throw new Error("test-services returned an invalid digest");
}

function contained(parent, child) {
  const relative = path.relative(path.resolve(parent), path.resolve(child));
  return relative === "" || (!relative.startsWith("..") && !path.isAbsolute(relative));
}

function assertNoSymlinkComponents(candidate) {
  const resolved = path.resolve(candidate);
  const parsed = path.parse(resolved);
  let current = parsed.root;
  for (const component of resolved.slice(parsed.root.length).split(path.sep).filter(Boolean)) {
    current = path.join(current, component);
    if (!existsSync(current)) break;
    if (lstatSync(current).isSymbolicLink()) {
      throw new Error("local service session path must not traverse symlinks");
    }
  }
}

function assertPrivateDirectory(directory, label) {
  assertNoSymlinkComponents(directory);
  const info = lstatSync(directory);
  if (!info.isDirectory() || info.isSymbolicLink() || (info.mode & 0o777) !== 0o700) {
    throw new Error(`${label} must be a non-symlink owner-only 0700 directory`);
  }
  if (typeof process.getuid === "function" && info.uid !== process.getuid()) {
    throw new Error(`${label} must be owned by the current user`);
  }
}

function ensurePrivateDirectory(directory, label) {
  assertNoSymlinkComponents(path.dirname(directory));
  if (!existsSync(directory)) mkdirSync(directory, { recursive: true, mode: 0o700 });
  assertPrivateDirectory(directory, label);
}

function assertPrivateFile(file, label) {
  assertNoSymlinkComponents(file);
  const info = lstatSync(file);
  if (!info.isFile() || info.isSymbolicLink() || (info.mode & 0o777) !== 0o600) {
    throw new Error(`${label} must be a non-symlink owner-only 0600 regular file`);
  }
  if (typeof process.getuid === "function" && info.uid !== process.getuid()) {
    throw new Error(`${label} must be owned by the current user`);
  }
}

function writePrivateJSON(file, value) {
  ensurePrivateDirectory(path.dirname(file), "local service session directory");
  const temporary = `${file}.tmp-${process.pid}-${randomUUID()}`;
  const descriptor = openSync(temporary, "wx", 0o600);
  try {
    writeFileSync(descriptor, `${JSON.stringify(value, null, 2)}\n`, "utf8");
  } finally {
    closeSync(descriptor);
  }
  chmodSync(temporary, 0o600);
  renameSync(temporary, file);
  assertPrivateFile(file, "local service session descriptor");
}

function walkConfig(root, relative, entries) {
  const absolute = path.join(root, relative);
  const info = lstatSync(absolute);
  if (info.isSymbolicLink()) throw new Error(`session configuration input is a symlink: ${relative}`);
  if (info.isFile()) {
    entries.push({ path: relative, mode: info.mode & 0o777, digest: sha256(readFileSync(absolute)) });
    return;
  }
  if (!info.isDirectory()) throw new Error(`unsupported session configuration input: ${relative}`);
  for (const name of readdirSync(absolute).sort(compareASCII)) {
    walkConfig(root, path.join(relative, name), entries);
  }
}

export function localSessionConfigDigest(root) {
  const entries = [];
  for (const relative of configInputs) walkConfig(root, relative, entries);
  return semanticJSONDigest(entries);
}

function commandResult(command, args, { cwd, environment = {}, quiet = true } = {}) {
  const result = spawnSync(command, args, {
    cwd,
    env: { ...process.env, ...environment },
    encoding: "utf8",
    stdio: quiet ? ["ignore", "pipe", "pipe"] : "inherit",
  });
  if (result.error) throw result.error;
  if (result.status !== 0) {
    throw new Error(`${path.basename(command)} ${args[0] ?? ""} failed with exit ${result.status}`);
  }
  return String(result.stdout ?? "").trim();
}

function dockerObject(kind, identity, dependencies) {
  if (kind === "container" && dependencies.inspectContainer) {
    return dependencies.inspectContainer(identity);
  }
  if (kind === "image" && dependencies.inspectImage) {
    return dependencies.inspectImage(identity);
  }
  const raw = commandResult("docker", ["inspect", identity], { quiet: true });
  const values = JSON.parse(raw);
  if (!Array.isArray(values) || values.length !== 1) throw new Error(`docker ${kind} proof is invalid`);
  return values[0];
}

function containerProjection(value) {
  return {
    id: value.Id,
    image_id: value.Image,
    image_reference: value.Config?.Image,
    labels: value.Config?.Labels ?? {},
    running: value.State?.Running === true,
  };
}

function imageProjection(value) {
  return { id: value.Id };
}

function testservicesOutput(binary, args, { root, environment = {}, dependencies }) {
  if (dependencies.testservicesOutput) {
    return dependencies.testservicesOutput(args, { root, environment });
  }
  return commandResult(binary, args, { cwd: root, environment, quiet: true });
}

function runTestservices(binary, args, { root, environment = {}, dependencies }) {
  if (dependencies.runTestservices) {
    dependencies.runTestservices(args, { root, environment });
    return;
  }
  commandResult(binary, args, { cwd: root, environment, quiet: true });
}

function currentCompatibility({ root, binary, dependencies }) {
  assertPrivateExecutable(binary);
  const schemaDigest = normalizeDigest(
    testservicesOutput(binary, ["schema-hash"], { root, dependencies }),
  );
  const imageReferences = testservicesOutput(binary, ["images"], { root, dependencies })
    .split("\n")
    .map((value) => value.trim())
    .filter(Boolean)
    .sort(compareASCII);
  const images = imageReferences.map((reference) => ({
    reference,
    image_id: imageProjection(dockerObject("image", reference, dependencies)).id,
  }));
  return {
    schemaDigest,
    toolDigest: sha256(readFileSync(binary)),
    configDigest: localSessionConfigDigest(root),
    imageDigest: semanticJSONDigest(images),
    images,
  };
}

function assertPrivateExecutable(binary) {
  assertNoSymlinkComponents(binary);
  const info = statSync(binary);
  if (!info.isFile() || (info.mode & 0o111) === 0) {
    throw new Error("test-services binary must be a regular executable file");
  }
}

function descriptorEnvironment(environment) {
  const selected = {};
  for (const name of serviceEnvironmentNames) {
    if (Object.hasOwn(environment, name)) selected[name] = environment[name];
  }
  return selected;
}

export function resolveLocalSessionFile(environment = process.env) {
  const explicit = String(environment.CARTULARY_TEST_SERVICES_SESSION_FILE ?? "").trim();
  const machineCache = String(environment.CARTULARY_MACHINE_CACHE_DIR ?? "").trim();
  const file = explicit || (machineCache ? path.join(machineCache, "test-services", "session.json") : "");
  if (!file || !path.isAbsolute(file)) {
    throw new Error("CARTULARY_TEST_SERVICES_SESSION_FILE requires an absolute path or machine-cache default");
  }
  assertNoSymlinkComponents(file);
  return path.resolve(file);
}

function validateDescriptorSecurity(sessionFile, descriptor) {
  assertPrivateDirectory(path.dirname(sessionFile), "local service session directory");
  assertPrivateFile(sessionFile, "local service session descriptor");
  validateSchemaSync(descriptor.schema_id, descriptor);
  if (typeof process.getuid === "function" && descriptor.owner_uid !== process.getuid()) {
    throw new Error("local service session descriptor has the wrong owner UID");
  }
  const expectedStateRoot = path.join(path.dirname(sessionFile), "sessions", descriptor.session_id);
  if (path.resolve(descriptor.state_root) !== path.resolve(expectedStateRoot)) {
    throw new Error("local service session state root does not match its descriptor identity");
  }
  assertPrivateDirectory(descriptor.state_root, "local service session state root");
  if (!contained(descriptor.state_root, descriptor.runtime_root)) {
    throw new Error("local service session runtime root escapes session state");
  }
  if (!contained(descriptor.state_root, descriptor.service_lease.result_root)) {
    throw new Error("local service session lease result root escapes session state");
  }
  if (descriptor.service_lease.suite_id !== descriptor.session_id) {
    throw new Error("local service session lease identity does not match the descriptor");
  }
  if (descriptor.environment_digest !== semanticJSONDigest(descriptor.environment)) {
    throw new Error("local service session environment digest does not match");
  }
  const created = Date.parse(descriptor.created_at);
  const expires = Date.parse(descriptor.expires_at);
  if (!Number.isFinite(created) || !Number.isFinite(expires) || expires <= created) {
    throw new Error("local service session timestamps are invalid");
  }
  const leaseResources = new Map(
    descriptor.service_lease.resources.map((resource) => [resource.service, resource]),
  );
  for (const container of descriptor.containers) {
    const leaseResource = leaseResources.get(container.service);
    if (!leaseResource || leaseResource.container_id !== container.container_id) {
      throw new Error("local service session container does not match its lease");
    }
    if (
      container.labels["cartulary.test-services.session-id"] !== descriptor.session_id ||
      container.labels["cartulary.test-services.session-expires-at"] !== descriptor.expires_at
    ) {
      throw new Error("local service session container labels do not match its descriptor");
    }
  }
  return descriptor;
}

export function readLocalSessionDescriptor(sessionFile) {
  assertPrivateDirectory(path.dirname(sessionFile), "local service session directory");
  assertPrivateFile(sessionFile, "local service session descriptor");
  const descriptor = JSON.parse(readFileSync(sessionFile, "utf8"));
  return validateDescriptorSecurity(sessionFile, descriptor);
}

function borrowerDirectory(descriptor) {
  return path.join(descriptor.state_root, "borrowers");
}

function processStartProof(pid = process.pid) {
  try {
    const raw = readFileSync(`/proc/${pid}/stat`, "utf8");
    const close = raw.lastIndexOf(")");
    const fields = raw.slice(close + 2).trim().split(/\s+/u);
    return `linux-proc-start:${fields[19]}`;
  } catch {
    return `pid-only:${pid}`;
  }
}

function liveBorrower(record, dependencies) {
  if (dependencies.borrowerAlive) return dependencies.borrowerAlive(record);
  try {
    process.kill(record.pid, 0);
  } catch {
    return false;
  }
  return processStartProof(record.pid) === record.process_start_proof;
}

function borrowerRecords(descriptor) {
  const directory = borrowerDirectory(descriptor);
  if (!existsSync(directory)) return [];
  assertPrivateDirectory(directory, "local service session borrower directory");
  return readdirSync(directory)
    .filter((name) => name.endsWith(".json"))
    .sort(compareASCII)
    .map((name) => {
      const file = path.join(directory, name);
      assertPrivateFile(file, "local service session borrower lease");
      const value = JSON.parse(readFileSync(file, "utf8"));
      if (
        value.schema_id !== "cartulary.test_services.local_session_borrower.v1" ||
        value.session_id !== descriptor.session_id ||
        !identityPattern.test(value.borrower_id) ||
        !identityPattern.test(value.run_id) ||
        !Number.isInteger(value.pid) ||
        value.pid < 1 ||
        typeof value.process_start_proof !== "string"
      ) {
        throw new Error("local service session borrower lease is invalid");
      }
      return { file, value };
    });
}

function inspectDescriptor(descriptor, { root, binary, dependencies, strictCompatibility }) {
  const services = [];
  const inspected = [];
  for (const declared of descriptor.containers) {
    let current;
    try {
      current = containerProjection(dockerObject("container", declared.container_id, dependencies));
    } catch {
      services.push({ service: declared.service, running: false });
      return { state: "stale", compatible: false, reason: "container_missing", services };
    }
    const identityMatches =
      current.id === declared.container_id &&
      current.image_id === declared.image_id &&
      current.labels["cartulary.test-services.session-id"] === descriptor.session_id &&
      current.labels["cartulary.test-services.session-expires-at"] === descriptor.expires_at;
    services.push({ service: declared.service, running: current.running && identityMatches });
    if (!identityMatches) {
      return { state: "stale", compatible: false, reason: "container_missing", services };
    }
    if (!current.running) {
      return { state: "stale", compatible: false, reason: "container_not_running", services };
    }
    inspected.push(current);
  }
  if (Date.now() >= Date.parse(descriptor.expires_at)) {
    return { state: "expired", compatible: false, reason: "descriptor_expired", services };
  }
  if (strictCompatibility) {
    const current = currentCompatibility({ root, binary, dependencies });
    if (current.schemaDigest !== descriptor.schema_digest) {
      return { state: "stale", compatible: false, reason: "schema_changed", services };
    }
    if (current.toolDigest !== descriptor.tool_digest) {
      return { state: "stale", compatible: false, reason: "tool_changed", services };
    }
    if (current.configDigest !== descriptor.config_digest) {
      return { state: "stale", compatible: false, reason: "configuration_changed", services };
    }
    if (current.imageDigest !== descriptor.image_digest) {
      return { state: "stale", compatible: false, reason: "image_changed", services };
    }
  }
  return { state: "active", compatible: true, services, inspected };
}

function statusValue(assessment, descriptor, borrowerCount) {
  const value = {
    schema_id: statusSchemaID,
    state: assessment.state,
    compatible: assessment.compatible,
    ...(descriptor
      ? {
          session_id: descriptor.session_id,
          created_at: descriptor.created_at,
          expires_at: descriptor.expires_at,
        }
      : {}),
    borrower_count: borrowerCount,
    ...(assessment.reason ? { reason: assessment.reason } : {}),
    services: assessment.services ?? [],
  };
  validateSchemaSync(value.schema_id, value);
  return value;
}

function withLifecycleLock(sessionFile, operation) {
  ensurePrivateDirectory(path.dirname(sessionFile), "local service session directory");
  const lock = `${sessionFile}.lock`;
  try {
    mkdirSync(lock, { mode: 0o700 });
  } catch (error) {
    if (error?.code === "EEXIST") throw new Error("local service session lifecycle is already active");
    throw error;
  }
  try {
    assertPrivateDirectory(lock, "local service session lifecycle lock");
    return operation();
  } finally {
    rmSync(lock, { recursive: true, force: false });
  }
}

export function localSessionStatus({ root, binary, sessionFile, dependencies = {} }) {
  if (!existsSync(sessionFile)) {
    return statusValue(
      { state: "absent", compatible: false, reason: "descriptor_missing", services: [] },
      null,
      0,
    );
  }
  try {
    const descriptor = readLocalSessionDescriptor(sessionFile);
    const assessment = inspectDescriptor(descriptor, {
      root,
      binary,
      dependencies,
      strictCompatibility: true,
    });
    const live = borrowerRecords(descriptor).filter(({ value }) => liveBorrower(value, dependencies));
    return statusValue(assessment, descriptor, live.length);
  } catch {
    return statusValue(
      { state: "invalid", compatible: false, reason: "descriptor_invalid", services: [] },
      null,
      0,
    );
  }
}

export function createLocalSession({ root, binary, sessionFile, environment = process.env, dependencies = {} }) {
  if (Object.hasOwn(environment, "CARTULARY_TEST_SERVICES_ACTIVE")) {
    throw new Error("CARTULARY_TEST_SERVICES_ACTIVE is internal and must not be supplied by callers");
  }
  return withLifecycleLock(sessionFile, () => {
    if (existsSync(sessionFile)) {
      throw new Error("a local service session descriptor already exists; inspect or stop it first");
    }
    const sessionID = dependencies.sessionID?.() ?? randomBytes(12).toString("hex");
    const now = dependencies.now?.() ?? new Date();
    const expiresAt = new Date(now.getTime() + sessionDurationMS).toISOString();
    const stateRoot = path.join(path.dirname(sessionFile), "sessions", sessionID);
    ensurePrivateDirectory(path.dirname(stateRoot), "local service session state directory");
    mkdirSync(stateRoot, { mode: 0o700 });
    assertPrivateDirectory(stateRoot, "local service session state root");
    const resultsRoot = path.join(stateRoot, "results");
    const runRoot = path.join(resultsRoot, "session-up");
    ensurePrivateDirectory(resultsRoot, "local service session result root");
    ensurePrivateDirectory(runRoot, "local service session run root");
    const suiteRuntime = createSuiteRuntime({
      repoRoot: root,
      runRoot,
      runID: "session-up",
      scratchRoot: path.join(stateRoot, "scratch"),
    });
    const envFile = path.join(stateRoot, "suite-environment.json");
    const leaseFile = path.join(stateRoot, "suite-lease.json");
    const startEnvironment = {
      CARTULARY_TEST_SERVICES_PERSISTENT_SESSION: "1",
      CARTULARY_TEST_SERVICES_SESSION_EXPIRES_AT: expiresAt,
      CARTULARY_TEST_SUITE_ID: sessionID,
      CARTULARY_TEST_TARGET: "test-services-session-up",
      CARTULARY_TEST_RESULTS_DIR: resultsRoot,
      CARTULARY_TEST_RUN_ID: "session-up",
      CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: suiteRuntime.root,
      CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: suiteRuntime.leaseID,
      CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: "session-up",
    };
    let started = false;
    try {
      runTestservices(binary, ["start-suite", "--env-file", envFile, "--lease-file", leaseFile], {
        root,
        environment: startEnvironment,
        dependencies,
      });
      started = true;
      assertPrivateFile(envFile, "local service session environment staging file");
      assertPrivateFile(leaseFile, "local service session lease staging file");
      const suiteEnvironment = JSON.parse(readFileSync(envFile, "utf8"));
      const serviceLease = JSON.parse(readFileSync(leaseFile, "utf8"));
      validateSchemaSync(serviceLease.schema_id, serviceLease);
      const compatibility = currentCompatibility({ root, binary, dependencies });
      const containers = serviceLease.resources
        .map((resource) => {
          const inspected = containerProjection(
            dockerObject("container", resource.container_id, dependencies),
          );
          if (!inspected.running || inspected.id !== resource.container_id) {
            throw new Error(`persistent ${resource.service} container is not running`);
          }
          if (
            inspected.labels["cartulary.test-services.session-id"] !== sessionID ||
            inspected.labels["cartulary.test-services.session-expires-at"] !== expiresAt
          ) {
            throw new Error(`persistent ${resource.service} container proof labels are invalid`);
          }
          return {
            service: resource.service,
            container_id: inspected.id,
            image_reference: resource.image,
            image_id: inspected.image_id,
            labels: inspected.labels,
          };
        })
        .sort((left, right) => compareASCII(left.service, right.service));
      const selectedEnvironment = descriptorEnvironment(suiteEnvironment);
      const descriptor = {
        schema_id: descriptorSchemaID,
        session_id: sessionID,
        owner_uid: typeof process.getuid === "function" ? process.getuid() : 0,
        created_at: now.toISOString(),
        expires_at: expiresAt,
        state_root: stateRoot,
        runtime_root: suiteRuntime.root,
        runtime_lease_id: suiteRuntime.leaseID,
        schema_digest: compatibility.schemaDigest,
        tool_digest: compatibility.toolDigest,
        config_digest: compatibility.configDigest,
        image_digest: compatibility.imageDigest,
        environment_digest: semanticJSONDigest(selectedEnvironment),
        environment: selectedEnvironment,
        containers,
        service_lease: serviceLease,
      };
      validateSchemaSync(descriptor.schema_id, descriptor);
      writePrivateJSON(sessionFile, descriptor);
      rmSync(envFile, { force: true });
      rmSync(leaseFile, { force: true });
      return statusValue(
        { state: "active", compatible: true, services: containers.map((entry) => ({ service: entry.service, running: true })) },
        descriptor,
        0,
      );
    } catch (error) {
      if (started && existsSync(leaseFile)) {
        try {
          runTestservices(binary, ["terminate-suite", "--lease", leaseFile], {
            root,
            environment: startEnvironment,
            dependencies,
          });
        } catch {
          // Preserve the primary creation failure; the state root remains for exact recovery.
        }
      }
      try {
        closeSuiteRuntimeRoot({ root: suiteRuntime.root, expectedLeaseID: suiteRuntime.leaseID });
        rmSync(stateRoot, { recursive: true, force: true });
      } catch {
        // Preserve exact state for manual recovery if ownership validation fails.
      }
      throw error;
    }
  });
}

export function attachLocalSession({
  root,
  binary,
  sessionFile,
  target,
  runID,
  suiteRuntime,
  dependencies = {},
}) {
  if (!allowedAttachedTargets.has(target)) {
    throw new Error(`${target} is owned-only and rejects local service attachment`);
  }
  if (!identityPattern.test(runID)) throw new Error("local service borrower requires a safe run ID");
  return withLifecycleLock(sessionFile, () => {
    const descriptor = readLocalSessionDescriptor(sessionFile);
    const assessment = inspectDescriptor(descriptor, {
      root,
      binary,
      dependencies,
      strictCompatibility: true,
    });
    if (assessment.state !== "active" || !assessment.compatible) {
      throw new Error(`local service session is not attachable: ${assessment.reason ?? assessment.state}`);
    }
    const borrowers = borrowerDirectory(descriptor);
    ensurePrivateDirectory(borrowers, "local service session borrower directory");
    const borrowerID = (dependencies.borrowerID?.() ?? randomUUID()).replaceAll("-", "");
    const borrowerFile = path.join(borrowers, `${borrowerID}.json`);
    const borrower = {
      schema_id: "cartulary.test_services.local_session_borrower.v1",
      session_id: descriptor.session_id,
      borrower_id: borrowerID,
      run_id: runID,
      pid: process.pid,
      process_start_proof: processStartProof(),
      created_at: new Date().toISOString(),
    };
    writePrivateJSON(borrowerFile, borrower);
    const runSuiteID = createHash("sha256")
      .update(`${descriptor.session_id}\0${runID}`)
      .digest("hex")
      .slice(0, 24);
    const environment = {
      ...descriptor.environment,
      CARTULARY_TEST_SERVICES_ACTIVE: "1",
      CARTULARY_TEST_SERVICES_CALL_MODE: "attach",
      CARTULARY_TEST_SERVICES_LIFECYCLE_MODE: "attach",
      CARTULARY_TEST_SERVICES_PERSISTENT_BORROWER: "1",
      CARTULARY_TEST_SUITE_ID: runSuiteID,
      CARTULARY_TEST_TARGET: target,
      CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: suiteRuntime.root,
      CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: suiteRuntime.leaseID,
      CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runID,
    };
    suiteRuntime.registerEnvironment(environment);
    let closed = false;
    return {
      environment,
      session_id: descriptor.session_id,
      borrower_id: borrowerID,
      ownership: "borrowed",
      close() {
        if (closed) return;
        closed = true;
        withLifecycleLock(sessionFile, () => {
          const current = readLocalSessionDescriptor(sessionFile);
          if (current.session_id !== descriptor.session_id) {
            throw new Error("local service session changed before borrower detach");
          }
          if (existsSync(borrowerFile)) {
            assertPrivateFile(borrowerFile, "local service session borrower lease");
            rmSync(borrowerFile, { force: false });
          }
        });
      },
    };
  });
}

export function stopLocalSession({ root, binary, sessionFile, dependencies = {} }) {
  return withLifecycleLock(sessionFile, () => {
    if (!existsSync(sessionFile)) {
      return statusValue(
        { state: "absent", compatible: false, reason: "descriptor_missing", services: [] },
        null,
        0,
      );
    }
    const descriptor = readLocalSessionDescriptor(sessionFile);
    const borrowers = borrowerRecords(descriptor);
    const live = [];
    for (const record of borrowers) {
      if (liveBorrower(record.value, dependencies)) live.push(record);
      else rmSync(record.file, { force: false });
    }
    if (live.length > 0) {
      throw new Error(`local service session has ${live.length} live borrower(s)`);
    }
    const leaseFile = path.join(descriptor.state_root, "down-lease.json");
    writePrivateJSON(leaseFile, descriptor.service_lease);
    const terminateEnvironment = {
      ...descriptor.environment,
      CARTULARY_TEST_SUITE_ID: descriptor.session_id,
      CARTULARY_TEST_SERVICES_ACTIVE: "1",
      CARTULARY_TEST_SERVICES_LIFECYCLE_MODE: "owned",
      CARTULARY_TEST_RESULTS_DIR: descriptor.service_lease.result_root,
      CARTULARY_TEST_RUN_ID: descriptor.service_lease.run_id,
      CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: descriptor.runtime_root,
      CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: descriptor.runtime_lease_id,
      CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: descriptor.service_lease.run_id,
    };
    runTestservices(binary, ["terminate-suite", "--lease", leaseFile], {
      root,
      environment: terminateEnvironment,
      dependencies,
    });
    rmSync(leaseFile, { force: true });
    closeSuiteRuntimeRoot({
      root: descriptor.runtime_root,
      expectedLeaseID: descriptor.runtime_lease_id,
    });
    rmSync(sessionFile, { force: false });
    rmSync(descriptor.state_root, { recursive: true, force: false });
    return statusValue(
      { state: "absent", compatible: false, reason: "descriptor_missing", services: [] },
      null,
      0,
    );
  });
}

export function resolveServiceSessionMode({ target, environment = process.env }) {
  if (Object.hasOwn(environment, "CARTULARY_TEST_SERVICES_ACTIVE")) {
    throw new Error("CARTULARY_TEST_SERVICES_ACTIVE is internal and must not be supplied by callers");
  }
  if (Object.hasOwn(environment, "CARTULARY_TEST_SERVICES_PERSISTENT_BORROWER")) {
    throw new Error(
      "CARTULARY_TEST_SERVICES_PERSISTENT_BORROWER is internal and must not be supplied by callers",
    );
  }
  const mode = String(environment.CARTULARY_TEST_SERVICES_MODE ?? "owned").trim() || "owned";
  if (!new Set(["owned", "attach"]).has(mode)) {
    throw new Error("CARTULARY_TEST_SERVICES_MODE must be owned or attach");
  }
  const explicitFile = String(environment.CARTULARY_TEST_SERVICES_SESSION_FILE ?? "").trim();
  if (mode === "owned" && explicitFile) {
    throw new Error("CARTULARY_TEST_SERVICES_SESSION_FILE is accepted only in attach mode");
  }
  if (mode === "attach" && !allowedAttachedTargets.has(target)) {
    throw new Error(`${target} is owned-only and rejects local service attachment`);
  }
  return {
    mode,
    sessionFile: mode === "attach" ? resolveLocalSessionFile(environment) : null,
  };
}
