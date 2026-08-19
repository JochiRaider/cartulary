#!/usr/bin/env node

import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { existsSync, lstatSync, mkdirSync, readFileSync, readdirSync, renameSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  canonicalJSONString,
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
import { reduceCanonicalUnitIntervals } from "../../evidence-accounting/canonical-unit-events.mjs";
import { buildSourceSnapshot } from "../../test-catalog/source-snapshot.mjs";
import { FixtureBroker } from "../fixture-broker/index.mjs";
import {
  productionFixtureProviders,
  startManagedSuite,
} from "../fixture-broker/providers.mjs";
import { WorkGraphCache, workGraphCacheRootRelative } from "./cache.mjs";
import { createAtomicNDJSONWriter } from "./atomic-ndjson.mjs";
import {
  captureCapabilitySnapshot,
  resourceCapacities,
} from "./capability.mjs";
import { WorkGraphCompiler } from "./compiler.mjs";
import { executeUnitProcess } from "./executor.mjs";
import { runWorkGraph } from "./scheduler.mjs";
import { resolveVulnerabilityDatabaseRevision } from "./vulnerability.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../../..");
const cacheModes = new Set(["normal", "cold", "off"]);

function usage() {
  return "usage: runner-cli.mjs --selection target|aggregate|owner|rows --target <target> [--owner <owner-id>] [--rows <row-id,...>] [--service-backed-only]";
}

function parseArgs(argv) {
  const options = {
    selection: "",
    target: "",
    owner: "",
    rows: undefined,
    serviceBackedOnly: false,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--service-backed-only") options.serviceBackedOnly = true;
    else if (arg === "--selection") options.selection = argv[++index] ?? "";
    else if (arg === "--target") options.target = argv[++index] ?? "";
    else if (arg === "--owner") options.owner = argv[++index] ?? "";
    else if (arg === "--rows") options.rows = (argv[++index] ?? "").split(",").filter(Boolean);
    else throw new Error(usage());
  }
  if (!new Set(["target", "aggregate", "owner", "rows"]).has(options.selection)) {
    throw new Error(usage());
  }
  if (!options.target) throw new Error(usage());
  if (options.selection === "owner" && !options.owner) throw new Error(usage());
  if (options.selection === "rows" && (!options.rows || options.rows.length === 0)) {
    throw new Error(usage());
  }
  if (options.serviceBackedOnly && options.selection !== "owner") throw new Error(usage());
  return options;
}

function sha256(bytes) {
  return `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
}

function readJSON(relative) {
  return JSON.parse(readFileSync(path.join(root, relative), "utf8"));
}

function resolvedRunRoot() {
  const resultsDir = process.env.CARTULARY_TEST_RESULTS_DIR || ".cartulary/test-results";
  const runID =
    process.env.CARTULARY_TEST_RUN_ID ||
    `${new Date().toISOString().replaceAll(/[-:.]/gu, "").replace("Z", "Z")}-p${process.pid}`;
  if (!/^[A-Za-z0-9_.-]+$/u.test(runID)) throw new Error("invalid graph run ID");
  return { resultsDir, runID, runRoot: path.resolve(root, resultsDir, runID) };
}

function commandEntry(target) {
  const taskSurface = readJSON("tools/task_surface_owner.json");
  const entry = taskSurface.targets.find((candidate) => candidate.name === target);
  if (!entry?.command_id) throw new Error(`public graph target ${target} has no command ID`);
  return entry;
}

function declaredInputs(entry) {
  return Object.fromEntries(
    (entry.input_contract?.inputs ?? [])
      .filter((input) => process.env[input.name] !== undefined && process.env[input.name] !== "")
      .map((input) => [
        input.name,
        input.summary_emission === "redacted_value"
          ? "<redacted>"
          : String(process.env[input.name]),
      ])
      .sort(([left], [right]) => left.localeCompare(right)),
  );
}

function selectionInputs(options) {
  return {
    ...(options.owner ? { OWNER: options.owner } : {}),
    ...(options.rows ? { ROWS: options.rows.join(",") } : {}),
    ...(options.serviceBackedOnly ? { SERVICE_BACKED_ONLY: "1" } : {}),
  };
}

function fixtureSelectionEnvironment(options) {
  const values = selectionInputs(options);
  const publicSelectionNames = ["OWNER", "ROWS"].filter(
    (name) => values[name] !== undefined,
  );
  const retainedSources = String(
    process.env.CARTULARY_MAKE_INPUT_SOURCES ?? "",
  )
    .split(/\s+/u)
    .filter(Boolean)
    .filter(
      (token) =>
        !publicSelectionNames.some((name) => token.startsWith(`${name}=`)),
    );
  return {
    ...values,
    CARTULARY_MAKE_INPUT_SOURCES: [
      ...retainedSources,
      ...publicSelectionNames.map((name) => `${name}=cli`),
    ].join(" "),
  };
}

function graphChildEnvironment(options) {
  const environment = { ...process.env };
  for (const name of [
    "CARTULARY_TEST_SERVICES_MODE",
    "CARTULARY_TEST_SERVICES_PERSISTENT_BORROWER",
    "CARTULARY_TEST_SERVICES_SESSION_FILE",
    "MAKEFLAGS",
    "MAKEOVERRIDES",
    "MFLAGS",
  ]) {
    delete environment[name];
  }
  if (new Set(["test-slice", "service-backed-test-slice"]).has(options.target)) {
    for (const name of ["JSON", "OWNER", "ROWS", "SERVICE_BACKED_ONLY"]) {
      delete environment[name];
    }
  }
  return environment;
}

function resolvedRuntimeEnvironment(compiler) {
  const environment = Object.fromEntries(
    compiler.topology.runtime_binaries
      .map((binary) => [
        binary.consumer_env,
        path.resolve(
          root,
          process.env[binary.output_make_variable] || binary.default_output_path,
        ),
      ])
      .sort(([left], [right]) => left.localeCompare(right)),
  );
  environment.CARTULARY_TEST_SERVICES_BIN = path.resolve(
    root,
    process.env.CARTULARY_TEST_SERVICES_BIN ||
      process.env.TEST_SERVICES_BIN ||
      "tmp/toolbin/cartulary-test-services",
  );
  environment.NODE_RUNTIME_DIR = path.resolve(
    root,
    process.env.NODE_RUNTIME_DIR || "tmp/node-runtime",
  );
  environment.NODE_BIN = process.env.NODE_BIN || path.join(environment.NODE_RUNTIME_DIR, "bin/node");
  environment.PNPM = process.env.PNPM || path.join(environment.NODE_RUNTIME_DIR, "bin/pnpm");
  return environment;
}

function selectionFor(options, compiler) {
  if (options.selection === "aggregate") {
    const plan = compiler.compileAggregatePlan(options.target);
    return { graph: plan.graph, projections: plan.projections };
  }
  if (options.selection === "target") {
    const graph = compiler.compile({ kind: "target", target: options.target });
    return { graph, projections: { [options.target]: graph.units.map((unit) => unit.unit_id) } };
  }
  if (options.selection === "rows") {
    const graph = compiler.compile({ kind: "rows", row_ids: options.rows });
    return { graph, projections: { [options.target]: graph.units.map((unit) => unit.unit_id) } };
  }
  let rowIDs = options.rows;
  if (options.serviceBackedOnly) {
    const serviceDependencies = new Map(
      compiler.catalog.rows.map((row) => [row.row_id, row.service_dependencies]),
    );
    const ownerRows = compiler.catalog.rows.filter((row) => row.owner_id === options.owner);
    rowIDs = (rowIDs ?? ownerRows.map((row) => row.row_id)).filter(
      (rowID) => (serviceDependencies.get(rowID)?.length ?? 0) > 0,
    );
    if (rowIDs.length === 0) throw new Error(`${options.owner} has no service-backed rows`);
  }
  const graph = compiler.compile({ kind: "owner", owner_id: options.owner, row_ids: rowIDs });
  return { graph, projections: { [options.target]: graph.units.map((unit) => unit.unit_id) } };
}

function repositoryProvenance() {
  const sourceCommit = execFileSync("git", ["rev-parse", "HEAD"], {
    cwd: root,
    encoding: "utf8",
  }).trim();
  const status = execFileSync(
    "git",
    ["status", "--porcelain=v1", "--untracked-files=all"],
    { cwd: root, encoding: "utf8", maxBuffer: 32 * 1024 * 1024 },
  );
  return {
    source_commit: sourceCommit,
    source_state: status === "" ? "clean" : "dirty",
  };
}

function intervalUnion(intervals) {
  const sorted = intervals
    .filter((interval) => interval.end >= interval.start)
    .sort((left, right) => left.start - right.start || left.end - right.end);
  let total = 0;
  let active = null;
  for (const interval of sorted) {
    if (!active || interval.start > active.end) {
      if (active) total += active.end - active.start;
      active = { ...interval };
    } else {
      active.end = Math.max(active.end, interval.end);
    }
  }
  if (active) total += active.end - active.start;
  return total;
}

function unitIntervals(events) {
  const starts = new Map();
  const intervals = new Map();
  for (const event of events) {
    if (event.event === "started") starts.set(event.unit_id, event.monotonic_ms);
    if (["completed", "failed", "cancelled"].includes(event.event) && starts.has(event.unit_id)) {
      intervals.set(event.unit_id, {
        start: starts.get(event.unit_id),
        end: event.monotonic_ms,
      });
    }
  }
  return intervals;
}

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

function actualCriticalPath(graph, intervals, selectedUnitIDs = null) {
  const selected = selectedUnitIDs ? new Set(selectedUnitIDs) : null;
  const units = selected
    ? graph.units.filter((unit) => selected.has(unit.unit_id))
    : graph.units;
  const byID = new Map(units.map((unit) => [unit.unit_id, unit]));
  const memo = new Map();
  function pathTo(unitID) {
    if (memo.has(unitID)) return memo.get(unitID);
    const unit = byID.get(unitID);
    const predecessors = unit.needs.filter((dependency) => byID.has(dependency)).map(pathTo);
    const prior = predecessors.sort(
      (left, right) => right.duration - left.duration || left.ids.join("\0").localeCompare(right.ids.join("\0")),
    )[0] ?? { duration: 0, ids: [] };
    const interval = intervals.get(unitID);
    const result = {
      duration: prior.duration + (interval ? interval.end - interval.start : 0),
      ids: [...prior.ids, unitID],
    };
    memo.set(unitID, result);
    return result;
  }
  return units
    .map((unit) => pathTo(unit.unit_id))
    .sort(
      (left, right) => right.duration - left.duration || left.ids.join("\0").localeCompare(right.ids.join("\0")),
    )[0]?.ids ?? [];
}

function canonicalTimingAccounting(
  events,
  graph,
  durationMs,
  { includeRunEnvelope = true, selectedUnitIDs = null } = {},
) {
  const selected = selectedUnitIDs ? new Set(selectedUnitIDs) : null;
  const byID = new Map(
    graph.units
      .filter((unit) => !selected || selected.has(unit.unit_id))
      .map((unit) => [unit.unit_id, unit]),
  );
  const started = new Map();
  const fixtureAcquired = new Map();
  const intervals = [];
  const resourceWaitStarted = new Map();
  const resourceWaitIntervals = [];
  let cleanupStarted = null;
  let firstUnitStarted = null;
  let processCount = 0;
  for (const event of events) {
    if (event.event === "started" && byID.has(event.unit_id)) {
      started.set(event.unit_id, event.monotonic_ms);
      firstUnitStarted ??= event.monotonic_ms;
      processCount += 1;
    }
    if (event.event === "fixture_acquired" && byID.has(event.unit_id)) {
      fixtureAcquired.set(event.unit_id, event.monotonic_ms);
    }
    if (
      event.event === "wait_started" &&
      byID.has(event.unit_id) &&
      !resourceWaitStarted.has(event.unit_id)
    ) {
      resourceWaitStarted.set(event.unit_id, event.monotonic_ms);
    }
    if (event.event === "wait_ended" && resourceWaitStarted.has(event.unit_id)) {
      resourceWaitIntervals.push({
        start: resourceWaitStarted.get(event.unit_id),
        end: event.monotonic_ms,
      });
      resourceWaitStarted.delete(event.unit_id);
    }
    if (event.event === "cleanup_started") cleanupStarted = event.monotonic_ms;
    if (
      ["completed", "failed", "cancelled"].includes(event.event) &&
      started.has(event.unit_id)
    ) {
      const unit = byID.get(event.unit_id);
      const start = started.get(event.unit_id);
      const acquired = fixtureAcquired.get(event.unit_id);
      if (acquired !== undefined && acquired > start) {
        intervals.push({ start, end: acquired, bucket: "fixture" });
      }
      const executionStart = acquired ?? start;
      const bucket = ["finalizer", "projection"].includes(unit.kind)
        ? "collation"
        : ["artifact", "readiness"].includes(unit.kind)
          ? "setup"
          : "execution";
      if (event.monotonic_ms > executionStart) {
        intervals.push({ start: executionStart, end: event.monotonic_ms, bucket });
      }
    }
  }
  if (includeRunEnvelope && (firstUnitStarted ?? 0) > 0) {
    intervals.push({ start: 0, end: firstUnitStarted, bucket: "setup" });
  }
  if (includeRunEnvelope && cleanupStarted !== null && durationMs > cleanupStarted) {
    intervals.push({ start: cleanupStarted, end: durationMs, bucket: "wrapper" });
  }
  const boundaries = [...new Set([
    ...(includeRunEnvelope ? [0, durationMs] : []),
    ...intervals.flatMap((interval) => [interval.start, interval.end]),
  ])].sort((left, right) => left - right);
  const totals = {
    setup_ms: 0,
    fixture_ms: 0,
    execution_ms: 0,
    collation_ms: 0,
    wrapper_ms: 0,
    unattributed_ms: 0,
  };
  const precedence = ["collation", "fixture", "execution", "setup", "wrapper"];
  for (let index = 1; index < boundaries.length; index += 1) {
    const start = boundaries[index - 1];
    const end = boundaries[index];
    if (end <= start) continue;
    const active = new Set(
      intervals
        .filter((interval) => interval.start < end && interval.end > start)
        .map((interval) => interval.bucket),
    );
    const bucket = precedence.find((candidate) => active.has(candidate));
    if (bucket) totals[`${bucket}_ms`] += end - start;
    else if (includeRunEnvelope) totals.unattributed_ms += end - start;
  }
  return {
    ...totals,
    resource_blocking_ms: intervalUnion(resourceWaitIntervals),
    process_count: processCount,
  };
}

function projectResourcePressure(events, graph, snapshot) {
  const byID = new Map(graph.units.map((unit) => [unit.unit_id, unit]));
  const capacities = Object.fromEntries(resourceCapacities(snapshot));
  const active = new Map();
  const saturationStart = new Map();
  const saturationMs = new Map();
  const peak = new Map();
  const admitted = new Set();
  const add = (resource, amount, now) => {
    const before = active.get(resource) ?? 0;
    const after = before + amount;
    active.set(resource, after);
    peak.set(resource, Math.max(peak.get(resource) ?? 0, after));
    const capacity = capacities[resource];
    if (before < capacity && after >= capacity) saturationStart.set(resource, now);
    if (before >= capacity && after < capacity && saturationStart.has(resource)) {
      saturationMs.set(resource, (saturationMs.get(resource) ?? 0) + now - saturationStart.get(resource));
      saturationStart.delete(resource);
    }
  };
  for (const event of events) {
    const unit = byID.get(event.unit_id);
    if (!unit) continue;
    if (event.event === "admitted") {
      admitted.add(event.unit_id);
      for (const [resource, amount] of Object.entries(unit.resource_claims)) add(resource, amount, event.monotonic_ms);
    }
    if (
      ["completed", "failed", "cancelled"].includes(event.event) &&
      admitted.delete(event.unit_id)
    ) {
      for (const [resource, amount] of Object.entries(unit.resource_claims)) add(resource, -amount, event.monotonic_ms);
    }
  }
  const completedAt = events.at(-1)?.monotonic_ms ?? 0;
  for (const [resource, startedAt] of saturationStart) {
    saturationMs.set(resource, (saturationMs.get(resource) ?? 0) + completedAt - startedAt);
  }
  const waits = events.filter((event) => event.event === "wait_started");
  return {
    requested_capacity: Object.fromEntries(
      Object.keys(capacities).sort().map((resource) => [resource, Math.max(...graph.units.map((unit) => unit.resource_claims[resource] ?? 0))]),
    ),
    resolved_capacity: capacities,
    peak_use: Object.fromEntries([...peak.entries()].sort(([left], [right]) => left.localeCompare(right))),
    saturation_ms: Object.fromEntries([...saturationMs.entries()].sort(([left], [right]) => left.localeCompare(right))),
    wait_events: waits.length,
    blocked_units: [...new Set(waits.map((event) => event.unit_id))].sort(),
    resource_holders: [...new Set(waits.flatMap((event) => event.blocking_unit_ids ?? []))].sort(),
  };
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

function writeUnitResult(runRoot, unit, result, missingOutputs) {
  const relative = unit.current_run_evidence_outputs.find((output) =>
    output.startsWith("unit-results/"),
  );
  if (!relative) throw new Error(`${unit.unit_id} has no canonical unit-result output`);
  const evidenceOutputs = [
    ...unit.current_run_evidence_outputs,
    ...serviceScopeArtifactRefs(result),
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
    { retainProjection: true },
  );
  if (canonical.finalMonotonicMs !== result.duration_ms) {
    throw new Error(
      `canonical event duration ${canonical.finalMonotonicMs} does not match scheduler duration ${result.duration_ms}`,
    );
  }
  result = { ...result, events: canonical.projectedEvents };
  const intervals = unitIntervals(result.events);
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
      failure_class: failureForUnits(unitIDs, result.events)?.failure_class ?? null,
      failure_reason: failureForUnits(unitIDs, result.events)?.failure_reason ?? null,
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
          return total + (interval ? interval.end - interval.start : 0);
        }, 0);
      })(),
      timing_accounting: canonicalTimingAccounting(
        result.events,
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
      return total + (interval ? interval.end - interval.start : 0);
    },
    0,
  );
  const timingAccounting = canonicalTimingAccounting(
    result.events,
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
  const cache = { hit: 0, miss: 0, bypass: 0 };
  for (const event of result.events) {
    if (event.event === "cache_hit") cache.hit += 1;
    if (event.event === "cache_miss") cache.miss += 1;
    if (event.event === "cache_bypass") cache.bypass += 1;
  }
  const failedEvent = result.events.find((event) => event.event === "failed") ??
    result.events.find((event) => event.failure_class);
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
    resource_pressure: projectResourcePressure(result.events, graph, snapshot),
    cache,
    artifact_refs: [
      "run-manifest.json",
      "unit-events.ndjson",
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
  const options = parseArgs(process.argv.slice(2));
  const serviceSession = resolveServiceSessionMode({
    target: options.target,
    environment: process.env,
  });
  const entry = commandEntry(options.target);
  const { resultsDir, runID, runRoot } = resolvedRunRoot();
  mkdirSync(runRoot, { recursive: true, mode: 0o700 });
  const compiler = new WorkGraphCompiler(root);
  const snapshot = captureCapabilitySnapshot({
    root,
    override: process.env.CARTULARY_HARNESS_CAPACITY_OVERRIDE,
    services: { browser: true, object_store: true, postgres: true, service_stack: true },
  });
  compiler.availableGoLanes = snapshot.cpu_tokens;
  compiler.availablePostgresLanes = snapshot.postgres_lanes;
  const { graph, projections } = selectionFor(options, compiler);
  const cacheMode = process.env.CARTULARY_HARNESS_CACHE_MODE || "normal";
  if (!cacheModes.has(cacheMode)) throw new Error(`invalid graph cache mode ${cacheMode}`);
  const source = buildSourceSnapshot(root);
  const provenance = repositoryProvenance();
  const toolchainDigest = sha256(readFileSync(path.join(root, "tools/toolchain_pins.json")));
  const helperDigest = sha256(readFileSync(path.join(root, "tools/harness_helper_ownership.json")));
  const startedAt = new Date().toISOString();
  const manifest = {
    schema_id: "cartulary.harness_run_manifest.v1",
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
    started_at: startedAt,
  };
  validateSchemaSync(manifest.schema_id, manifest);
  writeJSON(path.join(runRoot, "run-manifest.json"), manifest);

  const suiteRuntime = createSuiteRuntime({ repoRoot: root, runRoot, runID });
  const runtimeEnvironment = resolvedRuntimeEnvironment(compiler);
  const baseEnvironment = {
    ...graphChildEnvironment(options),
    ...runtimeEnvironment,
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
  };
  let suite = null;
  let suiteClosed = false;
  let suiteCloseError = null;
  const suiteController = {
    ensure() {
      if (suiteClosed) throw new Error("managed suite is closed");
      suite ??= serviceSession.mode === "attach"
        ? attachLocalSession({
            root,
            binary: runtimeEnvironment.CARTULARY_TEST_SERVICES_BIN,
            sessionFile: serviceSession.sessionFile,
            target: options.target,
            runID,
            suiteRuntime,
          })
        : startManagedSuite({
            root,
            target: options.target,
            suiteRuntime,
            environment: baseEnvironment,
          });
      return suite;
    },
    close() {
      if (suiteClosed) {
        if (suiteCloseError) throw suiteCloseError;
        return;
      }
      suiteClosed = true;
      try {
        suite?.close();
      } catch (error) {
        suiteCloseError = error;
        throw error;
      }
    },
  };
  let retainedScanAttempted = false;
  let primaryError = null;
  const publishRetainedScan = async () => {
    retainedScanAttempted = true;
    const retainedScan = await scanRetainedRoot(runRoot, {
      forbiddenValues: suiteRuntime.forbiddenValues(),
      removeUnsafe: true,
    });
    validateSchemaSync(retainedScan.schema_id, retainedScan);
    writeJSON(path.join(runRoot, "retained-secret-scan.json"), retainedScan);
  };
  try {
  const broker = new FixtureBroker({
    providers: productionFixtureProviders({
      root,
      selectionEnvironment: fixtureSelectionEnvironment(options),
      runtimeEnvironment,
      suiteController,
      suiteRuntime,
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
      inheritProcessEnvironment: false,
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
    writeUnitResult(runRoot, unit, result, missingOutputs);
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
      fixtureBroker: broker,
      cache,
      signal: controller.signal,
      agingQuantumMs: compiler.owner.aging_quantum_ms,
      cleanup: async () => suiteController.close(),
      onEvent: (event) => eventWriter.write(event),
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
  return {
    config: 2,
    infra: 3,
    product: 10,
    artifact: 11,
    harness: 11,
    security: 11,
    timing: 13,
    interrupted: 130,
  }[summary.failure_class] ?? 11;
  } catch (error) {
    primaryError = error;
    throw error;
  } finally {
    let boundaryError = null;
    if (!retainedScanAttempted) {
      try {
        await publishRetainedScan();
      } catch (error) {
        boundaryError = error;
      }
    }
    try {
      suiteController.close();
    } catch (error) {
      boundaryError ??= error;
    }
    if (!suiteCloseError) {
      try {
        suiteRuntime.close();
      } catch (error) {
        boundaryError ??= error;
      }
    }
    if (!primaryError && boundaryError) throw boundaryError;
  }
}

try {
  process.exitCode = await main();
} catch (error) {
  const configurationFailure =
    error.message === usage() ||
    error.message.includes("capacity override") ||
    error.message.includes("harness_capacity_override") ||
    error.message.includes("impossible resource claim") ||
    error.message.includes("dependency cycle");
  process.stderr.write(
    `[GRAPH-FAIL] failure_class=${configurationFailure ? "config" : "artifact"} failure_reason=${configurationFailure ? "configuration_error" : "artifact_error"} ${error.message}\n`,
  );
  process.exitCode = configurationFailure ? 2 : 11;
}
