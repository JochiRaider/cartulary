import { spawn, spawnSync } from "node:child_process";
import { randomUUID } from "node:crypto";
import { closeSync, existsSync, lstatSync, mkdirSync, openSync, readFileSync, rmSync } from "node:fs";
import { createServer } from "node:net";
import path from "node:path";
import { createCommandFailureContext, CommandFailure } from "../../runtime/command-failure.mjs";
import { ownedProcess, stopOwnedProcess } from "../../runtime/owned-process.mjs";
import { atomicLocalFile, readLocalFile, removePrivateFile } from "../../runtime/secure-local-files.mjs";

import {
  normalizeFailureClass,
  normalizeFailureReason,
  validateSchemaSync,
} from "../../contract/index.mjs";

export class FixtureAcquisitionError extends Error {
  constructor(message, { failureClass, failureReason, artifactRefs = [] }) {
    super(message);
    this.name = "FixtureAcquisitionError";
    this.failure_class = normalizeFailureClass(failureClass, "harness");
    this.failure_reason = normalizeFailureReason(failureReason, "fixture_error");
    this.artifact_refs = [...new Set(artifactRefs)].sort();
  }
}

function acquisitionError(message, failureClass, failureReason, artifactRefs = []) {
  return new FixtureAcquisitionError(message, {
    failureClass,
    failureReason,
    artifactRefs,
  });
}

function contained(parent, child) {
  const relative = path.relative(parent, child);
  return relative === "" || (!relative.startsWith("..") && !path.isAbsolute(relative));
}

function run(command, args, { cwd, environment, timeoutMS }) {
  const result = spawnSync(command, args, {
    cwd,
    env: { ...process.env, ...environment },
    stdio: "ignore",
    timeout: timeoutMS,
    killSignal: "SIGKILL",
  });
  if (result.error) throw result.error;
  if (result.status !== 0) {
    throw new Error(
      `${path.basename(command)} ${args[0] ?? ""} failed with exit ${result.status}`,
    );
  }
}

function acquireProcess(command, args, { cwd, environment, signal, onChildProcess = () => () => {} }) {
  return new Promise((resolve, reject) => {
    signal?.throwIfAborted();
    const diagnostic = ["ensure", "installed_only"].includes(environment.CARTULARY_PREPARATION_POLICY) ? createCommandFailureContext({ repoRoot: cwd, environment, unitID: "review:browser_stack", commandID: environment.CARTULARY_PREPARATION_POLICY === "installed_only" ? "cartulary.harness.command.ui_review.v1" : "cartulary.harness.command.browser_design_review.v1" }) : null;
    const child = spawn(command, args, { cwd, env: { ...process.env, ...environment, ...diagnostic?.environment }, stdio: "ignore", detached: true });
    let released = () => {}, spawnError;
    try { if (child.pid) released = onChildProcess(child.pid); }
    catch (error) { spawnError = error; child.kill("SIGKILL"); }
    const abort = () => { try { process.kill(-child.pid, "SIGTERM"); } catch (error) { if (error.code !== "ESRCH") spawnError ??= error; } };
    signal?.addEventListener("abort", abort, { once: true });
    child.once("error", (error) => { spawnError = error; });
    child.once("close", async (status) => {
      signal?.removeEventListener("abort", abort);
      const failure = diagnostic?.read();
      let primary = signal?.aborted ? signal.reason : failure ? new CommandFailure("browser acquisition failed", status === 0 ? { failure_class: "harness", failure_reason: "scheduler_accounting_error" } : failure) : spawnError ?? (status === 0 ? null : new Error("unclassified browser acquisition failure"));
      for (const cleanup of [released, () => diagnostic?.close()]) {
        try { await cleanup(); } catch (error) { primary ??= new CommandFailure("acquisition cleanup failed", { failure_class: "harness", failure_reason: "cleanup_error" }); (primary.cleanupFailures ??= []).push(error); }
      }
      if (primary) reject(primary); else resolve();
    });
  });
}

function readEnvironmentFile(file) {
  const environment = JSON.parse(readFileSync(file, "utf8"));
  if (!environment || typeof environment !== "object" || Array.isArray(environment)) {
    throw new Error(`invalid environment object in ${file}`);
  }
  for (const [name, value] of Object.entries(environment)) {
    if (!/^[A-Z][A-Z0-9_]*$/u.test(name)) {
      throw new Error(`invalid environment name in ${file}`);
    }
    if (typeof value !== "string") {
      throw new Error(`invalid environment value for ${name} in ${file}`);
    }
  }
  return environment;
}

function requireOwnerOnlyRegularFile(file, label) {
  const stat = lstatSync(file);
  if (!stat.isFile() || stat.isSymbolicLink()) {
    throw new Error(`${label} must be a regular non-symlink file`);
  }
  if ((stat.mode & 0o777) !== 0o600) {
    throw new Error(`${label} must have mode 0600`);
  }
  if (typeof process.getuid === "function" && stat.uid !== process.getuid()) {
    throw new Error(`${label} must be owned by the current user`);
  }
}

// A resource owner's stop protocol may remove its input lease. Give it a
// private snapshot while keeping the registered original through acknowledgement.
function stopWithRetainedProof(leaseFile, stop) {
  const snapshot = `${leaseFile}.cleanup-${randomUUID()}`;
  atomicLocalFile(snapshot, readLocalFile(leaseFile, { maximum: 1048576 }));
  let primary;
  try { stop(snapshot); }
  catch (error) { primary = error; throw error; }
  finally {
    try { removePrivateFile(snapshot); }
    catch (error) {
      if (primary) (primary.cleanupFailures ??= []).push(error);
      else throw error;
    }
  }
}

function acknowledgeRelease(onOwnedResource, kind, target) {
  try { onOwnedResource({ kind, target, state: "released" }); }
  catch (error) {
    throw new CommandFailure("resource ownership publication failed", {
      failure_class: "artifact", failure_reason: "artifact_error",
    }, { cause: error });
  }
  removePrivateFile(target);
}

function readSuiteEnvironmentFile(file) {
  const environment = JSON.parse(readFileSync(file, "utf8"));
  if (!environment || typeof environment !== "object" || Array.isArray(environment)) {
    throw new Error(`invalid environment object in ${file}`);
  }
  const admitted = Object.fromEntries(
    Object.entries(environment).filter(([name]) =>
      name === "CARTULARY_TEST_SUITE_ID" ||
      name === "CARTULARY_HARNESS_SUITE_RUNTIME_ROOT" ||
      name === "CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID" ||
      name === "CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID" ||
      name.startsWith("CARTULARY_TEST_SERVICES_") ||
      name.startsWith("CARTULARY_PGTEST_") ||
      name.startsWith("CARTULARY_S3TEST_"),
    ),
  );
  for (const [name, value] of Object.entries(admitted)) {
    if (typeof value !== "string") {
      throw new Error(`invalid environment value for ${name} in ${file}`);
    }
  }
  return admitted;
}

function availableLoopbackPort() {
  return new Promise((resolve, reject) => {
    const server = createServer();
    server.unref();
    server.once("error", reject);
    server.listen(0, "127.0.0.1", () => {
      const address = server.address();
      server.close((error) => {
        if (error) reject(error);
        else resolve(address.port);
      });
    });
  });
}

function commandSucceeded(command, args, options) {
  const result = spawnSync(command, args, {
    ...options,
    stdio: "ignore",
  });
  return !result.error && result.status === 0;
}

function requiredEnvironment(name) {
  const value = process.env[name];
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${name} is required for Make-owned Go work`);
  }
  return value;
}

function objectStoreNamespaceProvider({ root, suiteController, suiteRuntime, onOwnedResource }) {
  const proxyRoot = suiteRuntime.privatePath("object-store-proxy");
  const proxyBinary = path.join(proxyRoot, "s3corsproxy");
  let binaryReady = false;
  return {
    async acquire() {
      const suite = suiteController.ensure();
      const upstreamEndpoint = suite.environment.CARTULARY_S3TEST_ENDPOINT;
      if (!upstreamEndpoint) throw new Error("test-services suite has no object-store endpoint");
      mkdirSync(proxyRoot, { recursive: true, mode: 0o700 });
      if (!binaryReady) {
        run(process.env.GO || "go", ["build", "-o", proxyBinary, "./tools/s3corsproxy"], {
          cwd: root,
          environment: {
            GOCACHE: requiredEnvironment("GO_CACHE_DIR"),
            GOMODCACHE: requiredEnvironment("GO_MOD_CACHE_DIR"),
            GOTMPDIR: requiredEnvironment("GO_TMP_DIR"),
          },
        });
        binaryReady = true;
      }
      const instanceID = randomUUID().replaceAll("-", "");
      const instanceRoot = path.join(proxyRoot, instanceID);
      mkdirSync(instanceRoot, { recursive: true, mode: 0o700 });
      const attemptFile = path.join(instanceRoot, "start-attempt.json");
      const leaseFile = path.join(instanceRoot, "ready-lease.json");
      const logFile = path.join(instanceRoot, "proxy.log");
      const listen = `127.0.0.1:${await availableLoopbackPort()}`;
      const upstream = `http://${upstreamEndpoint}`;
      const origin = "http://localhost:5173";
      const common = ["--listen", listen, "--upstream", upstream, "--origin", origin];
      run(
        proxyBinary,
        [
          "attempt",
          ...common,
          "--attempt-file",
          attemptFile,
          "--instance-id",
          instanceID,
          "--log-path",
          logFile,
        ],
        { cwd: root, environment: {} },
      );
      const logFD = openSync(logFile, "a", 0o600);
      const child = spawn(
        proxyBinary,
        [
          "serve",
          ...common,
          "--attempt-file",
          attemptFile,
          "--instance-id",
          instanceID,
          "--log-path",
          logFile,
        ],
        { cwd: root, detached: true, stdio: ["ignore", logFD, logFD] },
      );
      child.unref();
      closeSync(logFD);
      const processProof = child.pid ? ownedProcess(child.pid) : null;
      let ready = false;
      const stop = async () => {
        if (ready) {
          const proofFile = existsSync(leaseFile) ? leaseFile : attemptFile;
          requireOwnerOnlyRegularFile(proofFile, "object-store proxy recovery proof");
          stopWithRetainedProof(proofFile, (snapshot) => run(proxyBinary, ["stop", ...common, "--state-file", snapshot], { cwd: root, environment: {}, timeoutMS: 30000 }));
        } else if (processProof) await stopOwnedProcess(processProof);
        else throw new Error("object-store proxy has no process proof");
        acknowledgeRelease(onOwnedResource, "object_store_proxy", existsSync(leaseFile) ? leaseFile : attemptFile);
        if (processProof) onOwnedResource({ kind: "helper_process", target: processProof, state: "released" });
      };
      try {
        if (processProof) onOwnedResource({ kind: "helper_process", target: processProof, state: "acquired" });
        for (let attempt = 0; attempt < 200 && !ready; attempt += 1) {
          ready = commandSucceeded(
            proxyBinary,
            ["status", ...common, "--state-file", attemptFile],
            { cwd: root },
          );
          if (!ready) await new Promise((resolve) => setTimeout(resolve, 25));
        }
        if (!ready) {
          throw new Error(`run-scoped object-store proxy failed to become ready; inspect ${logFile}`);
        }
        run(
          proxyBinary,
          ["promote", ...common, "--attempt-file", attemptFile, "--lease-file", leaseFile],
          { cwd: root, environment: {} },
        );
        onOwnedResource({ kind: "object_store_proxy", target: leaseFile, state: "acquired" });
        if (processProof) onOwnedResource({ kind: "helper_process", target: processProof, state: "released" });
      } catch (error) {
        try { await stop(); } catch (cleanupError) { (error.cleanupFailures ??= []).push(cleanupError); }
        throw error;
      }
      return {
        ownership: "owned",
        resource_ids: [
          "suite:object_store_namespace",
          `object-store-proxy:${instanceID}`,
        ],
        resource: {
          environment: {
            ...suite.environment,
            OBJECT_STORE_ENDPOINT: listen,
            SEAWEEDFS_S3_ACCESS_KEY_ID: suite.environment.CARTULARY_S3TEST_ACCESS_KEY_ID,
            SEAWEEDFS_S3_SECRET_ACCESS_KEY: suite.environment.CARTULARY_S3TEST_SECRET_ACCESS_KEY,
            OBJECT_STORE_SECURE: "false",
          },
        },
        release: stop,
      };
    },
  };
}

export function terminateManagedSuiteLease({ root, executable, leaseFile, environment, retainProof = false }) {
  requireOwnerOnlyRegularFile(leaseFile, "managed suite recovery lease");
  const stop = (proof) => run(executable, ["terminate-suite", "--lease", proof], { cwd: root, environment, timeoutMS: 120000 });
  if (retainProof) stopWithRetainedProof(leaseFile, stop);
  else { stop(leaseFile); removePrivateFile(leaseFile); }
}

export function terminateBrowserStackLease({ root, leaseFile, environment, retainProof = false }) {
  requireOwnerOnlyRegularFile(leaseFile, "browser stack recovery lease");
  run(path.join(root, "tools/harness/browser/start-web-e2e.sh"), ["--session-stop", "--lease-file", leaseFile], {
    cwd: root, environment: { ...environment, CARTULARY_WEB_E2E_RETAIN_SESSION_LEASE: retainProof ? "1" : "0" }, timeoutMS: 120000,
  });
}

export function startManagedSuite({
  root,
  target,
  suiteRuntime,
  environment = {},
  executable: configuredExecutable,
  executableArgs = [],
  onOwnedResource = () => {},
}) {
  const suiteRoot = suiteRuntime.privatePath("test-services");
  mkdirSync(suiteRoot, { recursive: true, mode: 0o700 });
  const envFile = path.join(suiteRoot, "suite-environment.json");
  const leaseFile = path.join(suiteRoot, "suite-lease.json");
  const resultFile = path.join(suiteRoot, "suite-start-result.json");
  const executable = configuredExecutable ||
    process.env.TEST_SERVICES_BIN ||
    process.env.CARTULARY_TEST_SERVICES_BIN ||
    path.join(root, "tmp/toolbin/cartulary-test-services");
  const suiteID = `work-graph-${target}`.replaceAll(/[^A-Za-z0-9_.-]+/gu, "-");
  const startEnvironment = {
    ...environment,
    CARTULARY_TEST_SUITE_ID: suiteID,
    CARTULARY_TEST_TARGET: target,
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: suiteRuntime.root,
    CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: suiteRuntime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: suiteRuntime.runID,
  };
  onOwnedResource({ kind: "managed_suite", target: leaseFile, state: "pending" });
  const start = spawnSync(executable, [
    ...executableArgs,
    "start-suite",
    "--env-file", envFile,
    "--lease-file", leaseFile,
    "--result-file", resultFile,
  ], {
    cwd: root,
    env: { ...process.env, ...startEnvironment },
    stdio: "ignore",
  });
  if (start.error) {
    const missing = start.error.code === "ENOENT";
    const failure = acquisitionError(
      missing ? "test-services helper is unavailable" : "test-services helper failed before publishing start evidence",
      missing ? "config" : "harness",
      missing ? "configuration_error" : "fixture_error",
    );
    try {
      if (existsSync(leaseFile)) {
        terminateManagedSuiteLease({ root, executable, leaseFile, environment: startEnvironment, retainProof: true });
        acknowledgeRelease(onOwnedResource, "managed_suite", leaseFile);
      } else if (!start.pid) {
        // No child was created. This is positive acquisition state, not an
        // inference from absent lease material after a running helper failed.
        onOwnedResource({ kind: "managed_suite", target: leaseFile, state: "released" });
      }
    } catch (error) { failure.cleanupFailures = [error]; }
    throw failure;
  }
  let startResult;
  let serviceScope;
  let serviceScopeRef = "";
  try {
    requireOwnerOnlyRegularFile(resultFile, "test-services start result");
    startResult = JSON.parse(readFileSync(resultFile, "utf8"));
    validateSchemaSync(startResult.schema_id, startResult);
    if (
      startResult.run_id !== suiteRuntime.runID ||
      startResult.target !== target
    ) {
      throw new Error("test-services start result identity does not match the scheduler request");
    }
    const resultsDir = environment.CARTULARY_TEST_RESULTS_DIR;
    if (typeof resultsDir !== "string" || resultsDir.trim() === "") {
      throw new Error("managed suite requires CARTULARY_TEST_RESULTS_DIR");
    }
    const runRoot = path.resolve(root, resultsDir, suiteRuntime.runID);
    const scopePath = path.resolve(runRoot, startResult.service_scope_ref);
    if (!contained(runRoot, scopePath)) {
      throw new Error("test-services start result scope reference escapes the run root");
    }
    serviceScopeRef = path.relative(runRoot, scopePath).split(path.sep).join("/");
    requireOwnerOnlyRegularFile(scopePath, "test-services service scope");
    serviceScope = JSON.parse(readFileSync(scopePath, "utf8"));
    validateSchemaSync(serviceScope.schema_id, serviceScope);
    if (
      serviceScope.run_id !== startResult.run_id ||
      serviceScope.target !== startResult.target ||
      serviceScope.suite_id !== startResult.suite_id ||
      serviceScopeRef !== `_shared/test-services/${startResult.suite_id}/service-scope.json`
    ) {
      throw new Error("test-services start result and service scope identities disagree");
    }
  } catch (error) {
    const cleanupFailures = [];
    if (existsSync(leaseFile)) {
      try {
        terminateManagedSuiteLease({ root, executable, leaseFile, environment: startEnvironment, retainProof: true });
        acknowledgeRelease(onOwnedResource, "managed_suite", leaseFile);
      } catch (cleanupError) {
        cleanupFailures.push(cleanupError);
      }
    }
    if (!existsSync(resultFile) && start.status === 2) {
      throw acquisitionError(
        "test-services helper rejected its configuration before publishing start evidence",
        "config",
        "configuration_error",
      );
    }
    throw Object.assign(acquisitionError(
      `test-services start evidence is invalid after helper exit ${start.status ?? "unknown"}: ${error.message}`,
      "artifact",
      "artifact_error",
    ), { cleanupFailures });
  } finally {
    rmSync(resultFile, { force: true });
  }
  if (start.status !== 0) {
    if (
      startResult.status !== "failed" ||
      !serviceScope.failure ||
      serviceScope.failure.failure_class !== startResult.failure_class ||
      serviceScope.failure.failure_reason !== startResult.failure_reason
    ) {
      throw acquisitionError(
        "test-services failed without matching classified service evidence",
        "artifact",
        "artifact_error",
        [serviceScopeRef],
      );
    }
    const failure = acquisitionError(
      `test-services suite startup failed: ${startResult.failure_class}/${startResult.failure_reason}`,
      startResult.failure_class,
      startResult.failure_reason,
      [serviceScopeRef],
    );
    // The validated owner's terminal cleanup status is positive settlement
    // evidence. Absence of a lease file is never used as this authority.
    if (serviceScope.cleanup.status === "startup_failed") {
      try { onOwnedResource({ kind: "managed_suite", target: leaseFile, state: "released" }); }
      catch (error) { failure.cleanupFailures = [error]; }
    }
    throw failure;
  }
  if (startResult.status !== "ready" || serviceScope.failure) {
    throw acquisitionError(
      "test-services reported success without matching ready evidence",
      "artifact",
      "artifact_error",
      [serviceScopeRef],
    );
  }
  let suiteEnvironment;
  let closed = false;
  try {
    suiteEnvironment = readSuiteEnvironmentFile(envFile);
    suiteRuntime.registerEnvironment(suiteEnvironment);
    rmSync(envFile, { force: true });
    onOwnedResource({ kind: "managed_suite", target: leaseFile, state: "acquired" });
  }
  catch (error) {
    try {
      terminateManagedSuiteLease({ root, executable, leaseFile, environment: { ...environment, ...suiteEnvironment }, retainProof: true });
      acknowledgeRelease(onOwnedResource, "managed_suite", leaseFile);
    } catch (cleanupError) { (error.cleanupFailures ??= []).push(cleanupError); }
    throw error;
  }
  return {
    environment: suiteEnvironment,
    leaseFile,
    executable,
    close() {
      if (closed) return;
      terminateManagedSuiteLease({ root, executable, leaseFile, environment: { ...environment, ...suiteEnvironment }, retainProof: true });
      acknowledgeRelease(onOwnedResource, "managed_suite", leaseFile);
      closed = true;
    },
  };
}

function borrowedProvider(suiteController, resourceID, environmentForSuite = (value) => value) {
  return {
    async acquire() {
      const suite = suiteController.ensure();
      const environment = environmentForSuite(suite.environment);
      return {
        ownership: "borrowed",
        resource_ids: [resourceID],
        resource: { environment },
        detach() {},
      };
    },
  };
}

function browserSuiteEnvironment(environment) {
  const exact = new Set([
    "CARTULARY_TEST_SERVICES_CALL_MODE",
    "CARTULARY_TEST_SERVICES_LIFECYCLE_MODE",
    "CARTULARY_TEST_SUITE_ID",
  ]);
  return Object.fromEntries(
    Object.entries(environment).filter(([name]) =>
      exact.has(name) ||
      name.startsWith("CARTULARY_PGTEST_") ||
      name.startsWith("CARTULARY_S3TEST_"),
    ),
  );
}

export function productionFixtureProviders({
  root,
  selectionEnvironment = {},
  runtimeEnvironment = {},
  suiteController,
  suiteRuntime,
  signal,
  onOwnedResource = () => {},
  onChildProcess,
}) {
  const cloneOrdinals = new Map();
  const browserAllocationOrdinals = new Map();
  const sharedProviders = Object.fromEntries(
    [
      "postgres_transaction",
      "postgres_dedicated",
      "postgres_migration",
      "managed_process",
    ].map((capability) => [
      capability,
      borrowedProvider(suiteController, `suite:${capability}`),
    ]),
  );
  return {
    ...sharedProviders,
    object_store_namespace: objectStoreNamespaceProvider({ root, suiteController, suiteRuntime, onOwnedResource }),
    browser_stack: {
      async acquire({
        affinityKey,
        browserStage,
        runtimeProfileID = "default",
        fixtureProfileID,
        snapshotKey,
        builderUnitID,
        rowID,
        predicateID,
        leaseID,
      }) {
        const suite = suiteController.ensure();
        const suiteEnvironment = suite.environment;
        const safeAffinity = String(affinityKey).replaceAll(/[^A-Za-z0-9_.-]+/gu, "-");
        const allocationOrdinal = (browserAllocationOrdinals.get(safeAffinity) ?? 0) + 1;
        browserAllocationOrdinals.set(safeAffinity, allocationOrdinal);
        const browserSessionID = `${safeAffinity}-allocation-${String(allocationOrdinal).padStart(3, "0")}`;
        const sessionRoot = suiteRuntime.privatePath("browser-stack-leases", browserSessionID);
        mkdirSync(sessionRoot, { recursive: true, mode: 0o700 });
        const envFile = path.join(sessionRoot, "stack.env");
        const leaseFile = path.join(sessionRoot, "stack.lease");
        const lifecycle = path.join(root, "tools/harness/browser/start-web-e2e.sh");
        const profiled = Boolean(fixtureProfileID || snapshotKey || builderUnitID);
        if (
          profiled &&
          ![fixtureProfileID, snapshotKey, builderUnitID, rowID, predicateID, leaseID].every(Boolean)
        ) {
          throw new Error("profiled browser stack requires complete snapshot lease identity");
        }
        const ordinalKey = `${runtimeProfileID}:${fixtureProfileID}:${snapshotKey}`;
        const cloneOrdinal = profiled ? (cloneOrdinals.get(ordinalKey) ?? 0) + 1 : 0;
        if (profiled) cloneOrdinals.set(ordinalKey, cloneOrdinal);
        const environment = {
          ...suiteEnvironment,
          ...runtimeEnvironment,
          ...selectionEnvironment,
          CARTULARY_BROWSER_STAGE: browserStage,
          CARTULARY_BROWSER_RUNTIME_PROFILE_ID: runtimeProfileID,
          CARTULARY_BROWSER_SERVICE_REQUIREMENT: "test-services",
          CARTULARY_BROWSER_SESSION_GROUP: browserSessionID,
          CARTULARY_WEB_E2E_PRIVATE_SESSION_ROOT: sessionRoot,
          CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: suiteRuntime.root,
          CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: suiteRuntime.leaseID,
          CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: suiteRuntime.runID,
          CARTULARY_TEST_SUITE_ID: suiteEnvironment.CARTULARY_TEST_SUITE_ID || `work-graph-${browserSessionID}`,
          ...(profiled
            ? {
                CARTULARY_FIXTURE_PROFILE_ID: fixtureProfileID,
                CARTULARY_FIXTURE_SNAPSHOT_KEY: snapshotKey,
                CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID: builderUnitID,
                CARTULARY_FIXTURE_ROW_ID: rowID,
                CARTULARY_FIXTURE_PREDICATE_ID: predicateID,
                CARTULARY_FIXTURE_CLONE_LEASE_ID: leaseID,
                CARTULARY_FIXTURE_CLONE_ORDINAL: String(cloneOrdinal),
              }
            : {}),
        };
        onOwnedResource({ kind: "browser_stack", target: leaseFile, state: "pending" });
        const close = () => {
          terminateBrowserStackLease({ root, leaseFile, environment, retainProof: true });
          acknowledgeRelease(onOwnedResource, "browser_stack", leaseFile);
        };
        try {
          await acquireProcess(lifecycle, ["--session-start", "--env-file", envFile, "--lease-file", leaseFile], { cwd: root, environment, signal, onChildProcess });
        } catch (error) {
          if (existsSync(leaseFile)) {
            try { close(); }
            catch (cleanupError) { (error.cleanupFailures ??= []).push(cleanupError); }
          }
          throw error;
        }
        let stackEnvironment;
        try {
          stackEnvironment = readEnvironmentFile(envFile);
          const stackFile = stackEnvironment.CARTULARY_WEB_E2E_STACK_JSON_FILE;
          if (!stackFile || path.basename(stackFile) !== "stack-v7.json") {
            throw new Error("browser session did not publish its stack-v7 attachment path");
          }
          requireOwnerOnlyRegularFile(stackFile, "browser stack-v7 evidence");
          requireOwnerOnlyRegularFile(
            path.join(path.dirname(stackFile), "service-admission.json"),
            "browser service-admission evidence",
          );
          suiteRuntime.registerEnvironment(stackEnvironment);
          rmSync(envFile, { force: true });
          onOwnedResource({ kind: "browser_stack", target: leaseFile, state: "acquired" });
        } catch (error) {
          const failure = new CommandFailure("browser session publication failed", { failure_class: "artifact", failure_reason: "artifact_error" }, { cause: error });
          try { close(); }
          catch (cleanupError) { failure.cleanupFailures = [cleanupError]; }
          throw failure;
        }
        const unitEnvironment = {
          ...browserSuiteEnvironment(suiteEnvironment),
          ...selectionEnvironment,
          ...stackEnvironment,
          CARTULARY_BROWSER_STAGE: browserStage,
          CARTULARY_BROWSER_SESSION_GROUP: browserSessionID,
          CARTULARY_WEB_E2E_SESSION_LEASE_FILE: leaseFile,
        };
        return {
          ownership: "owned",
          resource_ids: [
            `browser-stack:${browserSessionID}`,
            ...(profiled ? [`fixture-clone:${snapshotKey}:${cloneOrdinal}`] : []),
          ],
          ...(profiled
            ? {
                fixture_profile_id: fixtureProfileID,
                snapshot_key: snapshotKey,
                builder_unit_id: builderUnitID,
                clone_ordinal: cloneOrdinal,
              }
            : {}),
          resource: { environment: unitEnvironment },
          environment: unitEnvironment,
          release: close,
        };
      },
      },
  };
}
