import { primaryPublicFailure, secureWriteFile, validateSchemaSync } from "../../contract/index.mjs";
import { CommandFailure } from "../../runtime/command-failure.mjs";
import path from "node:path";

export function cleanupErrors(error, seen = new Set()) {
  if (seen.has(error)) return [];
  seen.add(error);
  const children = [...(error instanceof AggregateError ? error.errors : []), ...(error?.cleanupFailures ?? [])];
  return [...(error instanceof AggregateError ? [] : [error]), ...children.flatMap((child) => cleanupErrors(child, seen))];
}

export function cleanupFailure(error) {
  if (["cleanup_error", "timeout_failure", "artifact_error"].includes(error?.failure_reason)) return error;
  if (error?.code === "ETIMEDOUT") return new CommandFailure("fixture cleanup deadline expired", {
    failure_class: "timing", failure_reason: "timeout_failure",
  }, { cause: error });
  return new CommandFailure("fixture cleanup failed", {
    failure_class: "harness", failure_reason: "cleanup_error",
  }, { cause: error });
}

export function aggregateCleanup(errors) {
  const unique = [...new Set(errors.flatMap((error) => cleanupErrors(error)))];
  if (!unique.length) return null;
  const selected = primaryPublicFailure(unique.map((error, index) => ({ ...cleanupFailure(error), lifecycle_step: "cleanup_finalizers", scheduler_event_sequence: index })));
  const failure = { failure_class: selected.failure_class, failure_reason: selected.failure_reason,
    lifecycle_step: selected.lifecycle_step, scheduler_event_sequence: selected.scheduler_event_sequence };
  return Object.assign(new AggregateError(unique, "fixture finalization failed"), failure);
}

/** Composition roots own the suite; providers borrow its environment only. */
export function createSuiteController(acquire) {
  let suite, closed = false, outcome;
  return {
    ensure() {
      if (closed) throw new Error("managed suite is closed");
      suite ??= acquire();
      return suite;
    },
    close() {
      if (!outcome) {
        closed = true;
        outcome = Promise.resolve().then(() => suite?.close());
      }
      return outcome;
    },
  };
}

export class CleanupResults {
  constructor() { this.results = []; }
  record(operation, { unitID = null, leaseID = null, artifactRefs = [], error, blocked = false } = {}) {
    const failures = error ? cleanupErrors(error).map(cleanupFailure).map((failure) => ({
      failure_class: failure.failure_class, failure_reason: failure.failure_reason,
    })) : [];
    const primary = primaryPublicFailure(failures.map((failure, index) => ({
      ...failure, lifecycle_step: "cleanup_finalizers", scheduler_event_sequence: index,
    })));
    const record = {
      sequence: this.results.length + 1, operation, unit_id: unitID,
      fixture_lease_id: leaseID, outcome: blocked ? "blocked" : error ? "failed" : "completed",
      failure_class: primary?.failure_class ?? null, failure_reason: primary?.failure_reason ?? null,
      failures, artifact_refs: artifactRefs,
    };
    this.results.push(record);
    return record;
  }
  async attempt(operation, release, options) {
    try { await release(); this.record(operation, options); return null; }
    catch (error) {
      const failure = cleanupFailure(error);
      this.record(operation, { ...options, error: failure });
      return failure;
    }
  }
  publish(runRoot, runID) {
    try {
      const payload = { schema_id: "cartulary.harness_cleanup_results.v1", run_id: runID, results: this.results };
      validateSchemaSync(payload.schema_id, payload);
      secureWriteFile(path.join(runRoot, "cleanup-results.json"), `${JSON.stringify(payload, null, 2)}\n`, { allowedRoot: runRoot });
    } catch (error) {
      throw new CommandFailure("cleanup evidence publication failed", {
        failure_class: "artifact", failure_reason: "artifact_error",
      }, { cause: error });
    }
  }
}
