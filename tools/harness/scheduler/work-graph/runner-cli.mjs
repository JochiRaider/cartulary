#!/usr/bin/env node
import { parseArgs, sha256, resolvedRunRoot, commandEntry, declaredInputs, selectionInputs, fixtureSelectionEnvironment, graphChildEnvironment, resolvedRuntimeEnvironment, selectionFor } from "./preparation.mjs";

import { intervalUnion, actualCriticalPath, canonicalTimingAccounting, projectResourcePressure } from "../../evidence-accounting/index.mjs";
import { createDiagnosticSession } from "../../observability/diagnostic-session.mjs";
import { instrumentationPolicy } from "../../observability/resource-collector.mjs";
import { execFileSync } from "../../workspace/child-process.mjs";
import { existsSync, lstatSync, mkdirSync, readFileSync, readdirSync, renameSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  canonicalJSONString,
  primaryPublicFailure,
  publicExitCodeForFailure,
  semanticJSONDigest,
  validateSchemaSync,
} from "../../contract/index.mjs";
import {
  createSuiteRuntime,
  scanRetainedRoot,
} from "../../runtime/suite-runtime.mjs";
import {
  attachLocalSession,
  resolveServiceSessionMode,
} from "../../services/local-session.mjs";
import { reduceCanonicalUnitIntervals } from "../../evidence-accounting/index.mjs";
import { buildSourceSnapshot, executableProvenance } from "../../test-catalog/index.mjs";
import { FixtureBroker, createSuiteController, CleanupResults, aggregateCleanup } from "../fixture-broker/index.mjs";
import {
  productionFixtureProviders,
  startManagedSuite,
} from "../fixture-broker/providers.mjs";
import { WorkGraphCache, workGraphCacheRootRelative } from "./cache.mjs";
import { acquireHostAdmission, inheritedHostLease } from "../../runtime/host-admission.mjs";
import { recordRuntimeResource, runtimeRecoveryResources } from "../../runtime/resource-recovery.mjs";
import { createAtomicNDJSONWriter } from "./atomic-ndjson.mjs";
import {
  captureCapabilitySnapshot,
  resourceCapacities,
} from "./capability.mjs";
import { WorkGraphCompiler } from "./compiler.mjs";
import {
  executeUnitProcess,
  withGraphNodeRuntime,
} from "./executor.mjs";
import { runWorkGraph } from "./scheduler.mjs";
import { createLiveStatusPublisher } from "./live-status.mjs";
import { resolveVulnerabilityDatabaseRevision } from "./vulnerability.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../../..");
const cacheModes = new Set(["normal", "cold", "off"]);
const readJSON = (relative) => JSON.parse(readFileSync(path.join(root, relative), "utf8"));



function statusForUnits(unitIDs, states) {
  const values = unitIDs.map((unitID) => states[unitID]);
  if (values.some((value) => value === "cancelled")) return "cancelled";
  if (values.some((value) => value === "failed")) return "fail";
  if (values.every((value) => value === "passed")) return "pass";
  return "skipped";
}

function failureForUnits(unitIDs, events) {
  return events.find((event) =>
    unitIDs.includes(event.unit_id) && event.event === "failed",
  ) ?? null;
}


let atomicWriteCounter = 0;

function writeAtomicText(file, value) {
  mkdirSync(path.dirname(file), { recursive: true, mode: 0o700 });
  const temporary = `${file}.tmp-${process.pid}-${atomicWriteCounter++}`;
  writeFileSync(temporary, value, { mode: 0o600 });
  renameSync(temporary, file);
}

function writeJSON(file, value) {
  writeAtomicText(file, `${JSON.stringify(value, null, 2)}\n`);
}

function boundedRedactedDiagnostic(value, forbiddenValues) {
  let text = String(value ?? "").slice(-(64 * 1024));
  for (const secret of forbiddenValues) {
    text = text.replaceAll(secret, "<redacted>");
  }
  return text
    .replaceAll(/((?:PASSWORD|SECRET|TOKEN|COOKIE|SESSION|DSN|ACCESS_KEY|PRIVATE_KEY)=)[^\s]+/giu, "$1<redacted>")
    .replaceAll(/((?:postgres|postgresql):\/\/[^\s/:]+:)[^\s/@]+(@)/giu, "$1<redacted>$2");
}

function containedRunFile(runRoot, relative) {
  if (path.isAbsolute(relative) || relative.split(/[\\/]/u).includes("..")) {
    throw new Error(`unit evidence path escapes the run root: ${relative}`);
  }
  const resolved = path.resolve(runRoot, relative);
  if (resolved !== runRoot && !resolved.startsWith(`${runRoot}${path.sep}`)) {
    throw new Error(`unit evidence path escapes the run root: ${relative}`);
  }
  return resolved;
}

function fixtureLeaseArtifactRefs(runRoot) {
  const directory = path.join(runRoot, "_shared", "fixture-leases");
  if (!existsSync(directory)) return [];
  return readdirSync(directory, { withFileTypes: true })
    .filter((entry) => entry.isFile() && !entry.isSymbolicLink() && entry.name.endsWith(".json"))
    .map((entry) => `_shared/fixture-leases/${entry.name}`)
    .sort();
}

function serviceScopeArtifactRefs(result) {
  return (result?.artifact_refs ?? []).filter((value) =>
    /^_shared\/test-services\/[A-Za-z0-9][A-Za-z0-9._-]*\/service-scope\.json$/u.test(value),
  );
}

function vitestDiagnosticArtifactRefs(runRoot, unit) {
  if (!unit.unit_id.startsWith("row:")) return [];
  const directory = `unit-logs/${unit.unit_id.replaceAll(":", "-")}`;
  const files = ["runner.json", "vitest-failure-details.json"];
  return files.map((name) => `${directory}/${name}`).filter((relative) => {
    const info = lstatSync(containedRunFile(runRoot, relative), { throwIfNoEntry: false });
    return info?.isFile() && !info.isSymbolicLink();
  });
}

function writeUnitResult(runRoot, unit, result, missingOutputs) {
  const relative = unit.current_run_evidence_outputs.find((output) =>
    output.startsWith("unit-results/"),
  );
  if (!relative) throw new Error(`${unit.unit_id} has no canonical unit-result output`);
  const evidenceOutputs = [
    ...unit.current_run_evidence_outputs,
    ...serviceScopeArtifactRefs(result),
    ...vitestDiagnosticArtifactRefs(runRoot, unit),
  ].filter((value, index, values) => values.indexOf(value) === index).sort();
  const payload = {
    schema_id: "cartulary.harness_unit_result.v1",
    unit_id: unit.unit_id,
    semantic_digest: unit.semantic_digest,
    status: result.status,
    exit_code: Number.isInteger(result.exit_code) ? result.exit_code : null,
    signal: result.signal ?? null,
    failure_class: result.failure_class ?? null,
    failure_reason: result.failure_reason ?? null,
    evidence_outputs: evidenceOutputs,
    missing_outputs: missingOutputs,
  };
  validateSchemaSync(payload.schema_id, payload);
  writeJSON(containedRunFile(runRoot, relative), payload);
}

function writeTerminalUnitArtifacts(runRoot, graph, result) {
  for (const unit of graph.units) {
    const terminal = result.unit_results[unit.unit_id];
    if (!terminal) {
      throw new Error(`${unit.unit_id} has no terminal scheduler result`);
    }
    const safeID = unit.unit_id.replaceAll(/[^A-Za-z0-9_.-]+/gu, "-");
    const logRoot = path.join(runRoot, "unit-logs", safeID);
    mkdirSync(logRoot, { recursive: true, mode: 0o700 });
    const errorText = terminal.error
      ? String(terminal.error.message ?? terminal.error)
      : "";
    for (const [name, value] of [
      ["stdout.log", terminal.stdout ?? ""],
      ["stderr.log", terminal.stderr ?? (errorText ? `${errorText}\n` : "")],
    ]) {
      const file = path.join(logRoot, name);
      if (!existsSync(file)) writeFileSync(file, value, { mode: 0o600 });
    }
    writeUnitResult(runRoot, unit, terminal, terminal.missing_outputs ?? []);
  }
}

function missingUnitOutputs(runRoot, unit) {
  return unit.current_run_evidence_outputs
    .filter((output) => !output.startsWith("unit-results/"))
    .filter((output) => {
      const file = containedRunFile(runRoot, output);
      return !existsSync(file) || !lstatSync(file).isFile() || lstatSync(file).isSymbolicLink();
    });
}

async function writeCanonicalArtifacts({
  target,
  entry,
  graph,
  projections,
  result,
  runRoot,
  runID,
  snapshot,
}) {
  const canonical = await reduceCanonicalUnitIntervals(
    path.join(runRoot, "unit-events.ndjson"),
  );
  if (canonical.finalMonotonicMs !== result.duration_ms) {
    throw new Error(
      `canonical event duration ${canonical.finalMonotonicMs} does not match scheduler duration ${result.duration_ms}`,
    );
  }
  const terminals = [...canonical.terminals.values()];
  const intervals = canonical.intervals;
  const taskSurface = readJSON("tools/task_surface_owner.json");
  const commandIDs = new Map(
    taskSurface.targets
      .filter((candidate) => candidate.command_id)
      .map((candidate) => [candidate.name, candidate.command_id]),
  );
  const projectionNames = Object.keys(projections).sort();
  for (const projection of projectionNames) {
    const commandID = commandIDs.get(projection) ?? entry.command_id;
    const unitIDs = [...new Set(projections[projection])].sort();
    const children = projectionNames.filter(
      (candidate) =>
        candidate !== projection &&
        projections[candidate].every((unitID) => unitIDs.includes(unitID)),
    );
    const childUnits = new Set(children.flatMap((child) => projections[child]));
    const inclusiveIntervals = unitIDs.map((unitID) => intervals.get(unitID)).filter(Boolean);
    const exclusiveIntervals = unitIDs
      .filter((unitID) => !childUnits.has(unitID))
      .map((unitID) => intervals.get(unitID))
      .filter(Boolean);
    const summary = {
      schema_id: "cartulary.harness_target_summary.v1",
      target: projection,
      command_id: commandID,
      status: statusForUnits(unitIDs, result.states),
      failure_class: failureForUnits(unitIDs, terminals)?.failure_class ?? null,
      failure_reason: failureForUnits(unitIDs, terminals)?.failure_reason ?? null,
      workload_digest: semanticJSONDigest({
        target: projection,
        evidence_outputs: graph.units
          .filter((unit) => unitIDs.includes(unit.unit_id))
          .flatMap((unit) => unit.current_run_evidence_outputs)
          .filter((output) => !output.startsWith("unit-results/"))
          .filter((output, index, outputs) => outputs.indexOf(output) === index)
          .sort(),
      }),
      unit_ids: unitIDs,
      inclusive_wall_ms: intervalUnion(inclusiveIntervals),
      exclusive_wall_ms: intervalUnion(exclusiveIntervals),
      actual_dependency_critical_path_ms: (() => {
        const pathUnits = actualCriticalPath(graph, intervals, unitIDs);
        return pathUnits.reduce((total, unitID) => {
          const interval = intervals.get(unitID);
          return total + (interval ? interval.end - interval.start + interval.queue_ms : 0);
        }, 0);
      })(),
      timing_accounting: canonicalTimingAccounting(
        canonical,
        graph,
        intervalUnion(inclusiveIntervals),
        { includeRunEnvelope: false, selectedUnitIDs: unitIDs },
      ),
      children,
      evidence_refs: graph.units
        .filter((unit) => unitIDs.includes(unit.unit_id))
        .flatMap((unit) => [
          ...unit.current_run_evidence_outputs,
          ...serviceScopeArtifactRefs(result.unit_results[unit.unit_id]),
        ])
        .filter((value, index, values) => values.indexOf(value) === index)
        .sort(),
    };
    validateSchemaSync(summary.schema_id, summary);
    writeJSON(path.join(runRoot, "target-summaries", `${projection}.json`), summary);
  }
  const counts = { total: graph.units.length, passed: 0, failed: 0, skipped: 0, cancelled: 0 };
  for (const state of Object.values(result.states)) {
    if (state === "passed") counts.passed += 1;
    else if (state === "cancelled") counts.cancelled += 1;
    else if (state === "failed") counts.failed += 1;
    else counts.skipped += 1;
  }
  const criticalPath = actualCriticalPath(graph, intervals);
  const criticalPathDurationMs = criticalPath.reduce(
    (total, unitID) => {
      const interval = intervals.get(unitID);
      return total + (interval ? interval.end - interval.start + interval.queue_ms : 0);
    },
    0,
  );
  const timingAccounting = canonicalTimingAccounting(
    canonical,
    graph,
    result.duration_ms,
  );
  const accounted = [
    "setup_ms",
    "fixture_ms",
    "execution_ms",
    "collation_ms",
    "wrapper_ms",
    "unattributed_ms",
  ].reduce((total, field) => total + timingAccounting[field], 0);
  if (accounted !== result.duration_ms) {
    throw new Error(
      `canonical timing buckets do not close: accounted=${accounted} wall=${result.duration_ms}`,
    );
  }
  const cache = canonical.cache;
  const failedEvent = primaryPublicFailure(Object.values(result.unit_results)
    .filter((terminal) => terminal.failure_class).concat(result.cleanup_error ? [result.cleanup_error] : []));
  const runSummary = {
    schema_id: "cartulary.harness_run_summary.v1",
    run_id: runID,
    target,
    status: result.status === "passed" ? "pass" : counts.cancelled > 0 ? "cancelled" : "fail",
    failure_class: failedEvent?.failure_class ?? null,
    failure_reason: failedEvent?.failure_reason ?? null,
    unit_counts: counts,
    wall_duration_ms: result.duration_ms,
    critical_path: criticalPath,
    actual_dependency_critical_path_ms: criticalPathDurationMs,
    timing_accounting: timingAccounting,
    resource_pressure: projectResourcePressure(canonical, graph, Object.fromEntries(resourceCapacities(snapshot))),
    cache,
    artifact_refs: [
      "run-manifest.json",
      "unit-events.ndjson",
      "cleanup-results.json",
      ...projectionNames.map((projection) => `target-summaries/${projection}.json`),
      ...fixtureLeaseArtifactRefs(runRoot),
      ...Object.values(result.unit_results).flatMap(serviceScopeArtifactRefs),
    ].filter((value, index, values) => values.indexOf(value) === index).sort(),
  };
  validateSchemaSync(runSummary.schema_id, runSummary);
  // The run summary is the terminal completion marker and is published last.
  writeJSON(path.join(runRoot, "run-summary.json"), runSummary);
  return runSummary;
}

async function main() {
  const graphEpoch = process.hrtime.bigint();
  const mode = process.env.HARNESS_DIAGNOSTICS || "off";
  if (!["off", "basic"].includes(mode)) throw new Error("invalid HARNESS_DIAGNOSTICS");
  const instrumentation = instrumentationPolicy();
  const diagnostics = createDiagnosticSession({ mode, epoch: graphEpoch });
  const mark = diagnostics.mark;
  const options = parseArgs(process.argv.slice(2));
  const serviceSession = resolveServiceSessionMode({
    target: options.target,
    environment: process.env,
  });
  const entry = commandEntry(options.target);
  const { resultsDir, runID, runRoot } = resolvedRunRoot();
  mkdirSync(runRoot, { recursive: true, mode: 0o700 });
  const compiler = new WorkGraphCompiler(root);
  mark("configuration_catalog");
  const snapshot = captureCapabilitySnapshot({
    root,
    override: process.env.CARTULARY_HARNESS_CAPACITY_OVERRIDE,
    services: { browser: true, object_store: true, postgres: true, service_stack: true },
  });
  mark("capability_capture");
  compiler.availableGoLanes = snapshot.cpu_tokens;
  compiler.availablePostgresLanes = snapshot.postgres_lanes;
  const { graph, projections } = selectionFor(options, compiler);
  mark("graph_construction");
  const cacheMode = process.env.CARTULARY_HARNESS_CACHE_MODE || "normal";
  if (!cacheModes.has(cacheMode)) throw new Error(`invalid graph cache mode ${cacheMode}`);
  const source = buildSourceSnapshot(root);
  mark("source_hashing");
  const provenance = executableProvenance(root);
  const toolchainDigest = sha256(readFileSync(path.join(root, "tools/toolchain_pins.json")));
  const helperDigest = sha256(readFileSync(path.join(root, "tools/harness_helper_ownership.json")));
  mark("provenance");
  const startedAt = new Date().toISOString();
  const manifest = {
    schema_id: "cartulary.harness_run_manifest.v2",
    run_id: runID,
    command_id: entry.command_id,
    target: options.target,
    declared_inputs: {
      ...declaredInputs(entry),
      ...selectionInputs(options),
    },
    source_digest: source.digest,
    ...provenance,
    toolchain_digest: toolchainDigest,
    system_digest: snapshot.snapshot_digest,
    capability_snapshot: snapshot,
    graph_digest: graph.graph_digest,
    cache_mode: cacheMode,
    instrumentation: { mode, policy_digest: instrumentation.digest },
    started_at: startedAt,
  };
  validateSchemaSync(manifest.schema_id, manifest);
  writeJSON(path.join(runRoot, "run-manifest.json"), manifest);
  mark("manifest_publication");
  let livePublisher;
  let resources;

  const runtimeEnvironment = resolvedRuntimeEnvironment(compiler);
  const suiteRuntime = createSuiteRuntime({ repoRoot: root, runRoot, runID });
  resources = diagnostics.start({ runRoot, manifest, runtime: suiteRuntime, capacities: Object.fromEntries(resourceCapacities(snapshot)), epoch: graphEpoch });
  const baseEnvironment = withGraphNodeRuntime({
    ...graphChildEnvironment(options),
    ...runtimeEnvironment,
    ...diagnostics.executionEnvironment(),
    CARTULARY_HARNESS_GRAPH_CHILD: "1",
    CARTULARY_HARNESS_IDENTITY_PREPARED: "1",
    CARTULARY_HARNESS_SKIP_PREREQUISITES: "1",
    CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
    CARTULARY_TEST_RESULTS_DIR: resultsDir,
    CARTULARY_TEST_RUN_ID: runID,
    CARTULARY_UNIT_CPU_TOKENS: String(snapshot.cpu_tokens),
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: suiteRuntime.root,
    CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: suiteRuntime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runID,
  }, runtimeEnvironment.NODE_BIN);
  const cleanupResults = new CleanupResults();
  const onOwnedResource = (resource) => recordRuntimeResource(suiteRuntime, resource);
  const suiteController = createSuiteController(() => serviceSession.mode === "attach"
    ? attachLocalSession({ root, binary: runtimeEnvironment.CARTULARY_TEST_SERVICES_BIN,
        sessionFile: serviceSession.sessionFile, target: options.target, runID, suiteRuntime, onContainerProof: (proof) => resources?.registerContainers(proof) })
    : startManagedSuite({ root, target: options.target, suiteRuntime, environment: baseEnvironment, onOwnedResource, onContainerProof: (proof) => resources?.registerContainers(proof) }));
  let broker;
  let finalizationComplete = false;
  let resourcesUnresolved = false;
  let retainedScanAttempted = false;
  let primaryError = null;
  const publishRetainedScan = async () => {
    await diagnostics.close(() => livePublisher?.stop());
    retainedScanAttempted = true;
    diagnostics.scanStarted();
    try {
    const retainedScan = await scanRetainedRoot(runRoot, {
      forbiddenValues: suiteRuntime.forbiddenValues(),
      removeUnsafe: true,
    });
    validateSchemaSync(retainedScan.schema_id, retainedScan);
    writeJSON(path.join(runRoot, "retained-secret-scan.json"), retainedScan);
    diagnostics.scanFinished(true);
    } catch (error) { diagnostics.scanFinished(false); throw error; }
  };
  try {
  livePublisher = createLiveStatusPublisher({ runRoot, manifest, graph, projections, resources: () => resources?.snapshot() });
  broker = new FixtureBroker({
    cleanupResults,
    observeAllocation: (record) => diagnostics.observeExecution(record),
    observeLease: (record) => diagnostics.observeExecution({ kind: "lease", ref: `lease:${record.lease_ref}`, ...record }),
    providers: productionFixtureProviders({
      root,
      selectionEnvironment: fixtureSelectionEnvironment(options),
      // Fixture launch receives a complete environment. Review composition can
      // supply its own sanitized environment without implicit ambient merging.
      runtimeEnvironment: { ...graphChildEnvironment(options), ...runtimeEnvironment, ...diagnostics.executionEnvironment(), CARTULARY_TEST_RUN_ID: runID, CARTULARY_TEST_RESULTS_DIR: resultsDir },
      suiteController,
      suiteRuntime,
      onOwnedResource,
      onChildProcess: (pid, correlation) => { resources?.register(pid, null, correlation); return () => {}; },
    }),
    recordSink(record) {
      writeJSON(
        path.join(runRoot, "_shared", "fixture-leases", `${record.lease_id}.json`),
        record,
      );
    },
  });
  const vulnerability = resolveVulnerabilityDatabaseRevision({
    root,
    database: process.env.GOVULNCHECK_DB || "",
    declaredRevision: process.env.CARTULARY_VULNERABILITY_DATABASE_REVISION || "",
  });
  const cache = new WorkGraphCache({
    root,
    runRoot,
    cacheRoot: path.join(root, workGraphCacheRootRelative),
    mode: cacheMode,
    toolchainDigest,
    helperDigest,
    sourceEntries: source.entries,
    getCatalog: () => compiler.catalog,
    vulnerabilityDatabaseRevision: vulnerability.revision,
  });
  const controller = new AbortController();
  const abort = () => controller.abort();
  process.once("SIGINT", abort);
  process.once("SIGTERM", abort);
  const executeUnit = async (unit, context) => {
    const safeID = unit.unit_id.replaceAll(/[^A-Za-z0-9_.-]+/gu, "-");
    const unitArtifactRoot = path.join(runRoot, "unit-artifacts", safeID);
    mkdirSync(unitArtifactRoot, { recursive: true, mode: 0o700 });
    let result = await executeUnitProcess(unit, {
      ...context,
      onProcess: (pid, invocation) => resources?.register(pid, unit.unit_id, { invocationID: invocation.invocation_id }),
      inheritProcessEnvironment: false,
      nodeBinary: runtimeEnvironment.NODE_BIN,
      environment: {
        ...context.environment,
        CARTULARY_WORK_UNIT_ID: unit.unit_id,
        CARTULARY_STEP_ARTIFACT_DIR: unitArtifactRoot,
      },
    });
    const privateLogRoot = suiteRuntime.privatePath("unit-output", safeID);
    mkdirSync(privateLogRoot, { recursive: true, mode: 0o700 });
    writeFileSync(path.join(privateLogRoot, "stdout.log"), result.stdout ?? "", { mode: 0o600 });
    writeFileSync(path.join(privateLogRoot, "stderr.log"), result.stderr ?? "", { mode: 0o600 });
    const logRoot = path.join(runRoot, "unit-logs", safeID);
    mkdirSync(logRoot, { recursive: true, mode: 0o700 });
    writeFileSync(
      path.join(logRoot, "stdout.log"),
      boundedRedactedDiagnostic(result.stdout, suiteRuntime.forbiddenValues()),
      { mode: 0o600 },
    );
    writeFileSync(
      path.join(logRoot, "stderr.log"),
      boundedRedactedDiagnostic(result.stderr, suiteRuntime.forbiddenValues()),
      { mode: 0o600 },
    );
    const missingOutputs = result.status === "passed" ? missingUnitOutputs(runRoot, unit) : [];
    if (missingOutputs.length > 0) {
      result = {
        ...result,
        status: "failed",
        failure_class: "artifact",
        stderr: `${result.stderr ?? ""}${result.stderr ? "\n" : ""}missing declared unit evidence: ${missingOutputs.join(", ")}\n`,
      };
      writeFileSync(
        path.join(logRoot, "stderr.log"),
        boundedRedactedDiagnostic(result.stderr, suiteRuntime.forbiddenValues()),
        { mode: 0o600 },
      );
    }
    // Publish the canonical unit result at terminal-unit time. Downstream graph
    // units may consume exact producer evidence before whole-run projections
    // are rendered, and cache storage must include this result.
    return { ...result, missing_outputs: missingOutputs };
  };
  let result;
  const liveEventFile = path.join(runRoot, "unit-events.ndjson");
  const eventWriter = createAtomicNDJSONWriter(liveEventFile, canonicalJSONString);
  baseEnvironment.CARTULARY_HARNESS_LIVE_UNIT_EVENTS_FILE = eventWriter.stagingFile;
  try {
    result = await runWorkGraph({
      graph,
      capacities: resourceCapacities(snapshot),
      cwd: root,
      environment: baseEnvironment,
      executeUnit,
      hostAdmission: (unit, signal) => {
        const exclusive = unit.exclusive_locks.includes("host_activity");
        if (!exclusive && !unit.shared_locks.includes("host_activity")) return null;
        return acquireHostAdmission({ mode: exclusive ? "exclusive" : "shared", claims: unit.resource_claims, capacities: Object.fromEntries(resourceCapacities(snapshot)), parent: inheritedHostLease(), signal, timeoutMs: unit.timeout_ms });
      },
      fixtureBroker: broker,
      cache,
      signal: controller.signal,
      agingQuantumMs: compiler.owner.aging_quantum_ms,
      cleanupResults,
      cleanup: async () => {
        if (runtimeRecoveryResources(suiteRuntime, { maximum: 4096 }).some((resource) => resource.kind !== "managed_suite")) {
          resourcesUnresolved = true;
          throw Object.assign(new Error("fixture ownership remains unresolved"), { failure_class: "harness", failure_reason: "cleanup_error" });
        }
        await suiteController.close();
        if (runtimeRecoveryResources(suiteRuntime, { maximum: 4096 }).length) {
          resourcesUnresolved = true;
          throw Object.assign(new Error("suite ownership remains unresolved"), { failure_class: "harness", failure_reason: "cleanup_error" });
        }
      },
      onUnitTerminal: (unit, result) => writeUnitResult(runRoot, unit, result, result.missing_outputs ?? []),
      finalize: async ({ unresolved }) => {
        resourcesUnresolved ||= unresolved;
        diagnostics.captureExecution();
        const error = await cleanupResults.attempt(unresolved ? "recovery_preserve" : "runtime_close",
          () => resourcesUnresolved ? suiteRuntime.preserveRecovery() : suiteRuntime.close());
        const errors = error ? [error] : [];
        try { cleanupResults.publish(runRoot, runID); } catch (failure) { errors.push(failure); }
        finalizationComplete = true;
        if (errors.length) throw aggregateCleanup(errors);
      },
      onEvent: (event) => {
        if (event.event === "run_started") mark("runtime_prepare");
        if (event.event === "run_completed") mark("canonical_execution_reference");
        return eventWriter.write(event);
      },
      observation: livePublisher,
      diagnostics: resources,
      onClockEpoch: diagnostics.schedulerEpoch,
      retainEvents: false,
    });
    await eventWriter.close();
  } catch (error) {
    await eventWriter.abort();
    throw error;
  } finally {
    process.removeListener("SIGINT", abort);
    process.removeListener("SIGTERM", abort);
  }
  writeTerminalUnitArtifacts(runRoot, graph, result);
  const summary = await writeCanonicalArtifacts({
    target: options.target,
    entry,
    graph,
    projections,
    result,
    runRoot,
    runID,
    snapshot,
  });
  mark("evidence_publication");
  await publishRetainedScan();
  const line = `[GRAPH] target=${options.target} status=${summary.status} units=${summary.unit_counts.passed}/${summary.unit_counts.total} duration_ms=${summary.wall_duration_ms} run_root=${path.relative(root, runRoot)}\n`;
  const outputMode = process.env.CARTULARY_OUTPUT_MODE || "summary";
  if (outputMode === "machine") {
    process.stdout.write(`${JSON.stringify(summary)}\n`);
  } else if (outputMode !== "quiet" || summary.status !== "pass") {
    (summary.status === "pass" ? process.stdout : process.stderr).write(line);
  }
  if (summary.status === "pass") return 0;
  if (summary.status === "cancelled") return 130;
  return publicExitCodeForFailure(summary);
  } catch (error) {
    primaryError = error;
    throw error;
  } finally {
    livePublisher?.stop();
    const boundaryErrors = [];
    if (!finalizationComplete) {
      try { await broker?.close(); } catch (error) { boundaryErrors.push(error); }
      resourcesUnresolved ||= broker?.hasUnresolvedCleanup() ?? false;
      try { resourcesUnresolved ||= runtimeRecoveryResources(suiteRuntime, { maximum: 4096 }).some((resource) => resource.kind !== "managed_suite"); }
      catch (error) { boundaryErrors.push(error); resourcesUnresolved = true; }
      const error = resourcesUnresolved ? null : await cleanupResults.attempt("services_close", () => suiteController.close());
      if (resourcesUnresolved) cleanupResults.record("services_close", { blocked: true });
      if (error) { boundaryErrors.push(error); resourcesUnresolved = true; }
      try { resourcesUnresolved ||= runtimeRecoveryResources(suiteRuntime, { maximum: 4096 }).length > 0; }
      catch (error) { boundaryErrors.push(error); resourcesUnresolved = true; }
      diagnostics.captureExecution();
      const runtimeError = await cleanupResults.attempt(resourcesUnresolved ? "recovery_preserve" : "runtime_close",
        () => resourcesUnresolved ? suiteRuntime.preserveRecovery() : suiteRuntime.close());
      if (runtimeError) boundaryErrors.push(runtimeError);
      try { cleanupResults.publish(runRoot, runID); } catch (error) { boundaryErrors.push(error); }
    }
    if (!retainedScanAttempted) {
      try { await publishRetainedScan(); }
      catch (error) { boundaryErrors.push(error); }
    }
    diagnostics.invocationComplete();
    if (boundaryErrors.length) {
      if (primaryError) (primaryError.cleanupFailures ??= []).push(...boundaryErrors);
      else throw aggregateCleanup(boundaryErrors);
    }
  }
}

try {
  process.exitCode = await main();
} catch (error) {
  const configurationFailure =
    error.failure_class === "config" ||
    error.message === usage() ||
    error.message.includes("capacity override") ||
    error.message.includes("harness_capacity_override") ||
    error.message.includes("impossible resource claim") ||
    error.message.includes("dependency cycle");
  const failure = primaryPublicFailure([error.failure_reason ? error : {
    failure_class: configurationFailure ? "config" : "artifact",
    failure_reason: configurationFailure ? "configuration_error" : "artifact_error",
  }]);
  process.stderr.write(`[GRAPH-FAIL] failure_class=${failure.failure_class} failure_reason=${failure.failure_reason}\n`);
  process.exitCode = publicExitCodeForFailure(failure);
}
