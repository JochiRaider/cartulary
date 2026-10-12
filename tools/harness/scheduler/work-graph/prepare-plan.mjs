import path from "node:path";
import { readFileSync } from "node:fs";
import { resolveServiceSessionMode } from "../../services/local-session.mjs";
import { WorkGraphCompiler } from "./compiler.mjs";
import { captureCapabilitySnapshot, resourceCapacities } from "./capability.mjs";
import { CachePlan } from "./cache-plan.mjs";
import { resolveVulnerabilityDatabaseRevision } from "./vulnerability.mjs";
import { parseArgs, selectionFor, commandEntry, resolvedRunRoot, declaredInputs, selectionInputs, resolvedRuntimeEnvironment, sha256 } from "./preparation.mjs";
import { buildSourceSnapshot, executableProvenance } from "../../test-catalog/index.mjs";
import { instrumentationPolicy } from "../../observability/instrumentation-policy.mjs";
import { readLocalClock, captureClockAnchor } from "../../runtime/boot-clock.mjs";
import { validateSchemaSync } from "../../contract/index.mjs";

// This process owns source-aware preparation only. No graph unit, fixture,
// collector, retained writer or cache transaction starts here.
export function prepareRuntimePlan(argv) {
  const root = path.resolve(import.meta.dirname, "../../../..");
  const options = parseArgs(argv), mode = process.env.HARNESS_DIAGNOSTICS || "off";
  if (!["off", "basic"].includes(mode)) throw Error("invalid HARNESS_DIAGNOSTICS");
  const cacheMode = process.env.CARTULARY_HARNESS_CACHE_MODE || "normal";
  if (!["normal", "cold", "off"].includes(cacheMode)) throw Error("invalid cache mode");
  const clock = readLocalClock(), first = mode === "basic" ? captureClockAnchor() : null;
  const intervals = []; let previous = 0;
  const mark = (phase) => {
    const end_ms = readLocalClock().monotonic_ms - clock.monotonic_ms;
    if (mode === "basic") intervals.push({ phase, start_ms: previous, end_ms });
    previous = end_ms;
  };
  const serviceSession = resolveServiceSessionMode({ target: options.target, environment: process.env });
  const compiler = new WorkGraphCompiler(root), entry = commandEntry(options.target);
  const { runID, runRoot } = resolvedRunRoot();
  mark("configuration_catalog");
  const snapshot = captureCapabilitySnapshot({ root, override: process.env.CARTULARY_HARNESS_CAPACITY_OVERRIDE,
    services: { browser: true, object_store: true, postgres: true, service_stack: true } });
  compiler.availableGoLanes = snapshot.cpu_tokens; compiler.availablePostgresLanes = snapshot.postgres_lanes;
  mark("capability_capture");
  const { graph, projections } = selectionFor(options, compiler);
  mark("graph_construction");
  const source = buildSourceSnapshot(root);
  mark("source_hashing");
  const toolchain_digest = sha256(readFileSync(path.join(root, "tools/toolchain_pins.json")));
  const helper_digest = sha256(readFileSync(path.join(root, "tools/harness_helper_ownership.json")));
  const manifest = { schema_id: "cartulary.harness_run_manifest.v2", run_id: runID, command_id: entry.command_id,
    target: options.target, declared_inputs: { ...declaredInputs(entry), ...selectionInputs(options) },
    source_digest: source.digest, ...executableProvenance(root), toolchain_digest, system_digest: snapshot.snapshot_digest,
    capability_snapshot: snapshot, graph_digest: graph.graph_digest, cache_mode: cacheMode,
    instrumentation: { mode, policy_digest: instrumentationPolicy().digest }, started_at: new Date().toISOString() };
  mark("provenance");
  const vulnerability = resolveVulnerabilityDatabaseRevision({ root, database: process.env.GOVULNCHECK_DB || "",
    declaredRevision: process.env.CARTULARY_VULNERABILITY_DATABASE_REVISION || "" });
  const cache = new CachePlan({ root, mode: cacheMode,
    toolchainDigest: toolchain_digest, helperDigest: helper_digest, sourceEntries: source.entries,
    getCatalog: () => compiler.catalog, vulnerabilityDatabaseRevision: vulnerability.revision });
  cache.validateGraph(graph);
  const source_inputs = cacheMode === "off" ? [] : source.entries;
  const sourceIndex = new Map(source_inputs.map((entry, index) => [entry.path, index]));
  const cache_inputs = {};
  const facts = graph.units.map((unit) => {
    const context = cacheMode === "off" ? { eligible: false, reason: "mode_off" } : cache.context(unit);
    if (context.eligible) cache_inputs[context.dependencyClosure.digest] ??= context.dependencyClosure.entries.map((entry) => sourceIndex.get(entry.path));
    return { unit_id: unit.unit_id, eligible: context.eligible, reason: context.reason ?? null,
      profile_id: context.profile?.profile_id ?? null, policy: context.profile?.policy ?? null,
      input_digest: context.inputDigest ?? null, dependency_digest: context.dependencyClosure?.digest ?? null, dependency_strategy: context.dependencyClosure?.strategy ?? null,
      vulnerability_database_revision: context.profile?.requires_vulnerability_database_revision ? vulnerability.revision || null : null };
  });
  mark("cache_preparation");
  const bindings = { ...resolvedRuntimeEnvironment(compiler), CARTULARY_TEST_SERVICES_MODE: serviceSession.mode,
    ...(serviceSession.sessionFile ? { CARTULARY_TEST_SERVICES_SESSION_FILE: serviceSession.sessionFile } : {}) };
  mark("launch_bindings");
  const commandIDs = new Map(compiler.taskSurface.targets.map((target) => [target.name, target.command_id]));
  const value = { schema_id: "cartulary.harness_runtime_plan.v1", graph, manifest,
    projections: Object.fromEntries(Object.entries(projections).map(([target, unit_ids]) => [target, { command_id: commandIDs.get(target) ?? entry.command_id, unit_ids }])),
    capacities: Object.fromEntries(resourceCapacities(snapshot)), aging_quantum_ms: compiler.owner.aging_quantum_ms,
    bindings, cache: facts, source_inputs, cache_inputs, helper_digest, preparation: { clock: mode === "basic" ? { clock_identity: clock.clock_identity,
      epoch_ms: clock.monotonic_ms, anchors: [first, captureClockAnchor()].filter(Boolean) } : null, intervals } };
  validateSchemaSync(value.schema_id, value);
  return value;
}
