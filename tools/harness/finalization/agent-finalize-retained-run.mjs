import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";
import { reduceCanonicalUnitIntervals } from "../evidence-accounting/canonical-unit-events.mjs";
import { buildSourceSnapshot } from "../test-catalog/source-snapshot.mjs";

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

function canonicalFiles(root) {
  return {
    manifest: path.join(root, "run-manifest.json"),
    events: path.join(root, "unit-events.ndjson"),
    summary: path.join(root, "run-summary.json"),
    target: path.join(root, "target-summaries", "check.json"),
  };
}

export function createRetainedRunPreflight({
  allowOlderResultsDir,
  readJSON,
  relToRepo,
  repoRoot,
  resultsDirInput,
}) {
  const sourceSnapshot = buildSourceSnapshot(repoRoot);
  const currentIdentity = { source_snapshot_digest: sourceSnapshot.digest };

  async function validateRetainedRunArtifacts(resolved, resultsDir, actionID) {
    if (!existsSync(resolved)) {
      return {
        ok: false,
        failure: preflightFailure(actionID, `RESULTS_DIR does not exist: ${resultsDir}`, "config"),
      };
    }
    if (!statSync(resolved).isDirectory()) {
      return {
        ok: false,
        failure: preflightFailure(actionID, `RESULTS_DIR is not a directory: ${resultsDir}`, "config"),
      };
    }
    const files = canonicalFiles(resolved);
    for (const file of Object.values(files)) {
      if (!existsSync(file)) {
        return {
          ok: false,
          failure: preflightFailure(
            actionID,
            `${relToRepo(file)} is required; RESULTS_DIR must be a successful canonical full warm make check run root`,
            "config",
          ),
        };
      }
    }
    try {
      const manifest = readJSON(files.manifest);
      const summary = readJSON(files.summary);
      const target = readJSON(files.target);
      validateSchemaSync("cartulary.harness_run_manifest.v1", manifest);
      validateSchemaSync("cartulary.harness_run_summary.v1", summary);
      validateSchemaSync("cartulary.harness_target_summary.v1", target);
      const eventState = await reduceCanonicalUnitIntervals(files.events);
      if (manifest.target !== "check" || summary.target !== "check" || target.target !== "check") {
        throw new Error("canonical retained root must identify check in manifest, run summary, and target summary");
      }
      if (manifest.run_id !== summary.run_id || manifest.run_id !== path.basename(resolved)) {
        throw new Error("canonical retained root run identity does not close");
      }
      if (manifest.source_digest !== sourceSnapshot.digest) {
        throw new Error("canonical retained root is incompatible with the current source digest");
      }
      if (summary.status !== "pass" || target.status !== "pass") {
        throw new Error("canonical retained root does not record a passing check run");
      }
      const counts = summary.unit_counts;
      if (counts.total !== counts.passed + counts.failed + counts.skipped + counts.cancelled) {
        throw new Error("canonical retained root unit counts do not close");
      }
      if (counts.failed !== 0 || counts.cancelled !== 0 || counts.passed !== counts.total) {
        throw new Error("canonical retained root contains non-passing units");
      }
      if (new Set(target.unit_ids).size !== counts.total) {
        throw new Error("canonical retained check projection does not cover the complete run roster");
      }
      if (
        eventState.terminals.size !== counts.total ||
        [...eventState.terminals.values()].some((event) => event.event !== "completed")
      ) {
        throw new Error("canonical retained event roster does not close as successful");
      }
      for (const artifact of ["run-manifest.json", "unit-events.ndjson"]) {
        if (!summary.artifact_refs.includes(artifact)) {
          throw new Error(`canonical retained run summary omits ${artifact}`);
        }
      }
    } catch (error) {
      return {
        ok: false,
        failure: preflightFailure(
          actionID,
          `canonical retained evidence is invalid at ${relToRepo(resolved)}: ${error.message}`,
        ),
      };
    }
    return { ok: true, resolved };
  }

  function completedAt(runRoot) {
    return readJSON(canonicalFiles(runRoot).manifest).started_at ?? path.basename(runRoot);
  }

  async function latestSuccessfulCheckRun(parentDir, actionID) {
    if (!existsSync(parentDir) || !statSync(parentDir).isDirectory()) return null;
    const candidates = [];
    for (const entry of readdirSync(parentDir, { withFileTypes: true })) {
      if (!entry.isDirectory()) continue;
      const candidate = path.join(parentDir, entry.name);
      if (!(await validateRetainedRunArtifacts(candidate, candidate, actionID)).ok) continue;
      candidates.push({ resolved: candidate, completed_at: completedAt(candidate) });
    }
    candidates.sort((left, right) =>
      String(left.completed_at).localeCompare(String(right.completed_at)) ||
      left.resolved.localeCompare(right.resolved),
    );
    return candidates.at(-1) ?? null;
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
    const validation = await validateRetainedRunArtifacts(resolved, resultsDir, actionID);
    if (!validation.ok) return { ...validation, selection: baseSelection() };
    const latest = await latestSuccessfulCheckRun(path.dirname(resolved), actionID);
    const latestResolved = latest?.resolved ?? resolved;
    const suppliedIsLatest = path.resolve(latestResolved) === resolved;
    const selection = {
      status: suppliedIsLatest ? "latest" : allowOlderResultsDir ? "older_with_override" : "older_rejected",
      supplied_results_dir: relToRepo(resolved),
      latest_results_dir: relToRepo(latestResolved),
      supplied_is_latest: suppliedIsLatest,
      allow_older_results_dir: allowOlderResultsDir,
    };
    if (!suppliedIsLatest && !allowOlderResultsDir) {
      return {
        ok: false,
        selection,
        failure: preflightFailure(
          actionID,
          `RESULTS_DIR is older than the latest successful canonical full warm check root ${relToRepo(latestResolved)}; set ALLOW_OLDER_RESULTS_DIR=1 to use ${relToRepo(resolved)}`,
          "config",
        ),
      };
    }
    return { ok: true, resolved, selection };
  }

  return {
    baseSelection,
    currentIdentity() { return { ...currentIdentity }; },
    latestSuccessfulCheckRun,
    validate,
    validateRetainedRunArtifacts,
  };
}
