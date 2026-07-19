import { existsSync, readdirSync, statSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";
import {
  collectServiceTimingContamination,
  formatContaminationReasons,
} from "../duration-accounting/duration-drift.mjs";
import { buildSourceSnapshot } from "../owner-slice/source-snapshot.mjs";
import { loadTestCatalog } from "../test-catalog/index.mjs";

const successfulTerminalStates = new Set(["passed", "skipped_authorized"]);
const partitionIdentityFields = [
  "command_id",
  "run_id",
  "owner_id",
  "selected_rows",
  "source_snapshot_digest",
  "catalog_semantic_digest",
  "verification_semantic_digest",
  "runtime_profile_digest",
  "resource_profile_digest",
  "fixture_profile_digest",
  "target_id",
];

function preflightFailure(actionID, reason, failureClass = "artifact") {
  return {
    action_id: actionID,
    substep_id: "retained-run-preflight",
    target: null,
    failure_class: failureClass,
    failure_reason:
      failureClass === "config" ? "configuration_error" : "artifact_error",
    headline: reason,
    summary_json: null,
  };
}

export function createRetainedRunPreflight({
  allowOlderResultsDir,
  filesNamed,
  readJSON,
  relToRepo,
  repoRoot,
  resultsDirInput,
}) {
  const catalog = loadTestCatalog(repoRoot);
  const currentIdentity = {
    source_snapshot_digest: buildSourceSnapshot(repoRoot).digest,
    catalog_semantic_digest: catalog.summary.catalog_semantic_digest,
    verification_semantic_digest: catalog.summary.verification_semantic_digest,
  };

  function sameValue(left, right) {
    return JSON.stringify(left) === JSON.stringify(right);
  }

  function validateOwnerEvidence(resolved, actionID) {
    const accountingFiles = filesNamed(resolved, "test-evidence-accounting.json");
    const ownerSummaryFiles = filesNamed(resolved, "test-owner-summary.json");
    if (accountingFiles.length === 0 || ownerSummaryFiles.length === 0) {
      return {
        ok: false,
        failure: preflightFailure(
          actionID,
          "RESULTS_DIR must contain test-evidence-accounting.json and test-owner-summary.json owner artifact families",
        ),
      };
    }
    const summaryByDirectory = new Map(
      ownerSummaryFiles.map((file) => [path.dirname(file), file]),
    );
    const accountingDirectories = new Set(accountingFiles.map(path.dirname));
    const unmatchedSummary = ownerSummaryFiles.find(
      (file) => !accountingDirectories.has(path.dirname(file)),
    );
    if (unmatchedSummary) {
      return {
        ok: false,
        failure: preflightFailure(
          actionID,
          `${relToRepo(unmatchedSummary)} has no matching test-evidence-accounting.json`,
        ),
      };
    }
    for (const accountingFile of accountingFiles) {
      const ownerSummaryFile = summaryByDirectory.get(path.dirname(accountingFile));
      if (!ownerSummaryFile) {
        return {
          ok: false,
          failure: preflightFailure(
            actionID,
            `${relToRepo(accountingFile)} has no matching test-owner-summary.json`,
          ),
        };
      }
      const accounting = readJSON(accountingFile);
      const ownerSummary = readJSON(ownerSummaryFile);
      try {
        validateSchemaSync("cartulary.test_evidence_accounting.v1", accounting);
        validateSchemaSync("cartulary.test_owner_summary.v1", ownerSummary);
      } catch (error) {
        return {
          ok: false,
          failure: preflightFailure(
            actionID,
            `owner evidence is schema-invalid at ${relToRepo(path.dirname(accountingFile))}: ${error.message}`,
          ),
        };
      }
      for (const [field, expected] of Object.entries(currentIdentity)) {
        if (accounting[field] !== expected) {
          return {
            ok: false,
            failure: preflightFailure(
              actionID,
              `${relToRepo(accountingFile)} is incompatible with current ${field}`,
            ),
          };
        }
      }
      for (const field of partitionIdentityFields) {
        if (!sameValue(accounting[field], ownerSummary[field])) {
          return {
            ok: false,
            failure: preflightFailure(
              actionID,
              `${relToRepo(ownerSummaryFile)} disagrees with its accounting partition on ${field}`,
            ),
          };
        }
      }
      if (
        accounting.status !== "pass" ||
        ownerSummary.status !== "pass" ||
        accounting.observed_rows.some(
          (row) => !successfulTerminalStates.has(row.terminal_state),
        )
      ) {
        return {
          ok: false,
          failure: preflightFailure(
            actionID,
            `${relToRepo(accountingFile)} does not record successful owner-row closure`,
          ),
        };
      }
    }
    return { ok: true };
  }

  function validateRetainedRunArtifacts(resolved, resultsDir, actionID) {
    if (!existsSync(resolved)) {
      return {
        ok: false,
        failure: preflightFailure(
          actionID,
          `RESULTS_DIR does not exist: ${resultsDir}`,
          "config",
        ),
      };
    }
    if (!statSync(resolved).isDirectory()) {
      return {
        ok: false,
        failure: preflightFailure(
          actionID,
          `RESULTS_DIR is not a directory: ${resultsDir}`,
          "config",
        ),
      };
    }

    const checkToolSummary = path.join(
      resolved,
      "check",
      "tool-run-summary.json",
    );
    const serviceBackedMarkers = [
      path.join(resolved, "check-service-backed", "tool-run-summary.json"),
      path.join(resolved, "check-service-backed", "target-summary.json"),
      path.join(resolved, "check-service-backed", "scheduler-summary.json"),
    ];
    const checkSchedulerSummary = path.join(
      resolved,
      "check",
      "scheduler-summary.json",
    );
    const checkEvents = path.join(resolved, "check", "scheduler-events.jsonl");
    if (!existsSync(checkToolSummary)) {
      if (serviceBackedMarkers.some((file) => existsSync(file))) {
        return {
          ok: false,
          failure: preflightFailure(
            actionID,
            `${relToRepo(resolved)} contains check-service-backed artifacts but no ${relToRepo(checkToolSummary)}; RESULTS_DIR is a partial service-backed run root and must be a successful full warm make check retained run root`,
            "config",
          ),
        };
      }
      return {
        ok: false,
        failure: preflightFailure(
          actionID,
          `${relToRepo(checkToolSummary)} is required; RESULTS_DIR must be a successful full warm make check retained run root`,
          "config",
        ),
      };
    }
    const checkSummary = readJSON(checkToolSummary);
    if (!checkSummary || checkSummary.status !== "pass") {
      return {
        ok: false,
        failure: preflightFailure(
          actionID,
          `${relToRepo(checkToolSummary)} must identify a passing warm check run`,
        ),
      };
    }
    for (const file of [checkSchedulerSummary, checkEvents]) {
      if (!existsSync(file)) {
        return {
          ok: false,
          failure: preflightFailure(
            actionID,
            `${relToRepo(file)} is required for warm scheduler checks`,
          ),
        };
      }
    }

    const schedulerSummaries = filesNamed(resolved, "scheduler-summary.json");
    const targetSummaries = filesNamed(resolved, "target-summary.json");
    if (
      schedulerSummaries.length === 0 ||
      targetSummaries.length === 0
    ) {
      return {
        ok: false,
        failure: preflightFailure(
          actionID,
          "RESULTS_DIR must contain scheduler and target summary artifact families",
        ),
      };
    }

    const ownerEvidence = validateOwnerEvidence(resolved, actionID);
    if (!ownerEvidence.ok) {
      return ownerEvidence;
    }

    const failedSummary = filesNamed(resolved, "tool-run-summary.json")
      .map((file) => ({ file, summary: readJSON(file) }))
      .find((entry) => entry.summary?.status === "fail");
    if (failedSummary) {
      return {
        ok: false,
        failure: preflightFailure(
          actionID,
          `${relToRepo(failedSummary.file)} records a failed retained target`,
        ),
      };
    }

    const contamination = collectServiceTimingContamination(repoRoot, resolved);
    if (contamination.contaminated) {
      return {
        ok: false,
        failure: preflightFailure(
          actionID,
          `RESULTS_DIR contains contaminated timing evidence: ${formatContaminationReasons(contamination)}`,
        ),
      };
    }
    return { ok: true, resolved };
  }

  function checkCompletedAt(runRoot) {
    const summary = readJSON(path.join(runRoot, "check", "tool-run-summary.json"));
    return summary?.completed_at || summary?.started_at || path.basename(runRoot);
  }

  function latestSuccessfulCheckRun(parentDir, actionID) {
    if (!existsSync(parentDir) || !statSync(parentDir).isDirectory()) {
      return null;
    }
    const candidates = [];
    for (const entry of readdirSync(parentDir, { withFileTypes: true })) {
      if (!entry.isDirectory()) {
        continue;
      }
      const candidate = path.join(parentDir, entry.name);
      const validation = validateRetainedRunArtifacts(candidate, candidate, actionID);
      if (!validation.ok) {
        continue;
      }
      candidates.push({
        resolved: candidate,
        completed_at: checkCompletedAt(candidate),
      });
    }
    candidates.sort((left, right) => {
      const byTime = String(left.completed_at).localeCompare(String(right.completed_at));
      return byTime || left.resolved.localeCompare(right.resolved);
    });
    return candidates.at(-1) ?? null;
  }

  function baseSelection() {
    return {
      status: resultsDirInput ? "not_evaluated" : "skipped",
      supplied_results_dir: resultsDirInput
        ? relToRepo(path.resolve(resultsDirInput))
        : null,
      latest_results_dir: null,
      supplied_is_latest: null,
      allow_older_results_dir: allowOlderResultsDir,
    };
  }

  function validate(resultsDir, actionID) {
    const resolved = path.resolve(resultsDir);
    const validation = validateRetainedRunArtifacts(resolved, resultsDir, actionID);
    if (!validation.ok) {
      return {
        ...validation,
        selection: {
          ...baseSelection(),
          status: "not_evaluated",
        },
      };
    }

    const latest = latestSuccessfulCheckRun(path.dirname(resolved), actionID);
    const latestResolved = latest?.resolved ?? resolved;
    const suppliedIsLatest = path.resolve(latestResolved) === resolved;
    const selection = {
      status: suppliedIsLatest
        ? "latest"
        : allowOlderResultsDir
          ? "older_with_override"
          : "older_rejected",
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
          `RESULTS_DIR is older than the latest successful full warm check retained root ${relToRepo(latestResolved)}; set ALLOW_OLDER_RESULTS_DIR=1 to intentionally use ${relToRepo(resolved)}`,
          "config",
        ),
      };
    }

    return { ok: true, resolved, selection };
  }

  return {
    baseSelection,
    currentIdentity() {
      return { ...currentIdentity };
    },
    latestSuccessfulCheckRun,
    validate,
    validateRetainedRunArtifacts,
  };
}
