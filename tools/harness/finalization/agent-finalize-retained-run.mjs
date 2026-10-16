import { closeSync, readdirSync } from "node:fs";
import os from "node:os";
import path from "node:path";

import { assertMachineEvidencePath, canonicalJSONString, openDirectory, parseStrictJSON, readRetainedArtifactBytes, validateSchemaSync } from "../contract/index.mjs";
import { buildSourceSnapshot, loadTestCatalog } from "../test-catalog/index.mjs";
import { captureCapabilitySnapshot, compileRetainedRun } from "../scheduler/work-graph/index.mjs";
import { validateCanonicalRun } from "../observability/canonical-evidence.mjs";
import { validateExecutionContext } from "../observability/canonical-context.mjs";
import { validateRetainedCleanup, validateRetainedGraph, validateRetainedSelection } from "../observability/retained-selection.mjs";

const read = (root, relative) => parseStrictJSON(readRetainedArtifactBytes(root, relative, 16 * 1024 ** 2).toString("utf8"));
const same = (a, b) => canonicalJSONString(a) === canonicalJSONString(b);

function preflightFailure(actionID, reason, failureClass = "artifact") {
  return {
    action_id: actionID,
    substep_id: "retained-run-preflight",
    target: null,
    failure_class: failureClass,
    failure_reason: failureClass === "config" ? "configuration_error" : "artifact_error",
    headline: reason,
    summary_json: null,
  };
}

// The explicit ports permit adversarial evidence tests without service startup
// or making a synthetic checkout into an executable input.
export function createRetainedRunPreflight({ allowOlderResultsDir, relToRepo, repoRoot, resultsDirInput }, ports = {}) {
  const sourceSnapshot = ports.sourceSnapshot ?? buildSourceSnapshot(repoRoot);
  const compileRun = ports.compileRun ?? ((manifest) => compileRetainedRun(repoRoot, manifest));
  const currentIdentity = { source_snapshot_digest: sourceSnapshot.digest };
  const validations = new Map();
  let current;
  function currentContract() {
    const override = process.env.CARTULARY_HARNESS_CAPACITY_OVERRIDE;
    if (override) assertMachineEvidencePath(path.resolve(repoRoot, override));
    current ??= {
      catalog: ports.catalog ?? loadTestCatalog(repoRoot),
      surface: ports.surface ?? read(repoRoot, "tools/task_surface_owner.json"),
      capability: ports.capability ?? captureCapabilitySnapshot({ root: repoRoot,
        override,
        services: { browser: true, object_store: true, postgres: true, service_stack: true } }),
      host: ports.host ?? { platform: os.platform(), architecture: os.arch(), kernel_release: os.release(), cpu_models: [...new Set(os.cpus().map((cpu) => cpu.model))].sort() },
    };
    return current;
  }

  function eligibleManifest(resolved) {
    let manifest;
    try { manifest = read(resolved, "run-manifest.json"); }
    catch (error) {
      if (error.code === "ENOENT") error.callerInput = true;
      throw error;
    }
    validateSchemaSync("cartulary.harness_run_manifest.v2", manifest);
    if (manifest.target !== "check" || manifest.cache_mode !== "normal") {
      throw Object.assign(new Error("RESULTS_DIR must be a full warm make check root with cache_mode=normal"), { callerInput: true });
    }
    if (manifest.run_id !== path.basename(resolved)) throw new Error("canonical retained root run identity does not close");
    if (manifest.source_digest !== sourceSnapshot.digest) throw new Error("canonical retained root is incompatible with the current source digest");
    return manifest;
  }

  async function validateRetainedRunArtifacts(resolved, _resultsDir, actionID) {
    try {
      assertMachineEvidencePath(resolved);
      const manifest = eligibleManifest(resolved);
      const run = await validateCanonicalRun(resolved, "check");
      if (!same(manifest, run.manifest)) throw new Error("retained manifest changed during preflight");
      if (run.summary.status !== "pass" || run.summary.failure_class !== null || run.summary.failure_reason !== null ||
          [...run.terminal.values()].some((event) => event.status !== "passed")) throw new Error("canonical retained root does not record a completely passing check run");
      const context = read(resolved, "execution-context.json"), cleanup = read(resolved, "cleanup-results.json");
      validateExecutionContext(context, run, cleanup);
      validateRetainedCleanup(run, cleanup);
      const scan = read(resolved, "retained-secret-scan.json");
      validateSchemaSync("cartulary.harness_retained_secret_scan.v1", scan);
      // The scanner owns this policy identity; instrumentation has a separate policy.
      const contract = currentContract();
      if (!same(context.host_profile, contract.host) || !same(manifest.capability_snapshot, contract.capability)) throw new Error("retained system and capacity differ from the current host");
      const plan = compileRun(manifest);
      validateRetainedGraph(run, plan);
      if (!same([...run.targetSummaries.keys()].sort(), Object.keys(plan.projections).sort())) throw new Error("retained target roster differs from the complete current check");
      const commands = new Map(contract.surface.targets.map((entry) => [entry.name, entry.command_id]));
      for (const [target, ids] of Object.entries(plan.projections)) {
        const projection = run.targetSummaries.get(target);
        if (projection.status !== "pass" || projection.failure_class !== null || projection.failure_reason !== null ||
            projection.command_id !== (commands.get(target) ?? manifest.command_id) || !same(projection.unit_ids, [...ids].sort())) throw new Error("retained target projection differs from the complete current check");
      }
      validateRetainedSelection(resolved, run, new Map(plan.graph.units.map((unit) => [unit.unit_id, unit])), run.targetSummaries.get("check"), contract.catalog.rows);
      for (const reference of ["run-manifest.json", "unit-events.ndjson"]) {
        if (!run.summary.artifact_refs.includes(reference)) throw new Error(`canonical retained run summary omits ${reference}`);
      }
      return { ok: true, resolved, started_at: manifest.started_at };
    } catch (error) {
      return { ok: false, failure: preflightFailure(actionID,
        `canonical retained evidence is invalid at ${relToRepo(resolved)}: ${error.message}`,
        error.callerInput ? "config" : "artifact") };
    }
  }

  async function validated(resolved, actionID) {
    if (!validations.has(resolved)) validations.set(resolved, validateRetainedRunArtifacts(resolved, resolved, actionID));
    return validations.get(resolved);
  }

  async function latestSuccessfulCheckRun(parentDir, actionID) {
    assertMachineEvidencePath(parentDir);
    const descriptor = openDirectory(parentDir);
    let entries;
    try { entries = readdirSync(`/proc/self/fd/${descriptor}`, { withFileTypes: true }); }
    finally { closeSync(descriptor); }
    const candidates = [];
    for (const entry of entries) {
      if (!entry.isDirectory()) continue;
      const resolved = path.join(parentDir, entry.name);
      try {
        const manifest = eligibleManifest(resolved);
        candidates.push({ resolved, started_at: manifest.started_at });
      } catch { /* An incompatible sibling cannot qualify as the latest check. */ }
    }
    candidates.sort((left, right) => String(right.started_at).localeCompare(String(left.started_at)) || right.resolved.localeCompare(left.resolved));
    for (const candidate of candidates) if ((await validated(candidate.resolved, actionID)).ok) return candidate;
    return null;
  }

  function baseSelection() {
    return {
      status: resultsDirInput ? "not_evaluated" : "skipped",
      supplied_results_dir: resultsDirInput ? relToRepo(path.resolve(resultsDirInput)) : null,
      latest_results_dir: null,
      supplied_is_latest: null,
      allow_older_results_dir: allowOlderResultsDir,
    };
  }

  async function validate(resultsDir, actionID) {
    const resolved = path.resolve(resultsDir);
    const validation = await validated(resolved, actionID);
    if (!validation.ok) return { ...validation, selection: baseSelection() };
    const latest = await latestSuccessfulCheckRun(path.dirname(resolved), actionID);
    const latestResolved = latest?.resolved ?? resolved;
    const suppliedIsLatest = latestResolved === resolved;
    const selection = {
      status: suppliedIsLatest ? "latest" : allowOlderResultsDir ? "older_with_override" : "older_rejected",
      supplied_results_dir: relToRepo(resolved),
      latest_results_dir: relToRepo(latestResolved),
      supplied_is_latest: suppliedIsLatest,
      allow_older_results_dir: allowOlderResultsDir,
    };
    if (!suppliedIsLatest && !allowOlderResultsDir) return { ok: false, selection, failure: preflightFailure(actionID,
      `RESULTS_DIR is older than the latest successful canonical full warm check root ${relToRepo(latestResolved)}; set ALLOW_OLDER_RESULTS_DIR=1 to use ${relToRepo(resolved)}`, "config") };
    return { ok: true, resolved, selection };
  }

  return { baseSelection, currentIdentity() { return { ...currentIdentity }; }, latestSuccessfulCheckRun, validate, validateRetainedRunArtifacts };
}
