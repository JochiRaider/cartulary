import type { GetJobResponse } from "@cartulary/protocol-ts/http";
export type CommonJobResource = GetJobResponse["data"];

/**
 * The terminal `result_summary.code` values the client recognizes: every Job
 * kind's declared success code and the common `job_canceled` (Core 01
 * §3.3.9.1). The list is closed, so a comparison with any other literal is a
 * type error; a test checks it against the Job-kind contracts.
 */
export const jobResultSummaryCodes = [
  "composition_preview_rendered",
  "evidence_blob_verified",
  "import_session_applied",
  "import_session_discovered",
  "import_session_partially_applied",
  "incident_bundle_exported",
  "incident_bundle_imported",
  "job_canceled",
  "network_flow_graph_view_materialized",
  "reference_pack_imported",
  "reference_pack_reverified",
  "reference_packs_refreshed",
  "release_created",
  "snapshot_created",
] as const;

export type JobResultSummaryCode = (typeof jobResultSummaryCodes)[number];

/** A job's terminal result code, or null when it has none or an unrecognized one. */
export function jobResultSummaryCode(job: {
  readonly result_summary: { readonly code: string } | null;
}): JobResultSummaryCode | null {
  const code = job.result_summary?.code;
  return code !== undefined &&
    (jobResultSummaryCodes as readonly string[]).includes(code)
    ? (code as JobResultSummaryCode)
    : null;
}

export const terminalCommonJob = (job: CommonJobResource) =>
  job.status === "succeeded" ||
  job.status === "failed" ||
  job.status === "canceled";

/** Compare JSON resource values independently of transport property order. */
export function equalJSONResource(left: unknown, right: unknown): boolean {
  if (left === right) return true;
  if (!left || !right || typeof left !== "object" || typeof right !== "object")
    return false;
  if (Array.isArray(left) !== Array.isArray(right)) return false;
  const a = Object.entries(left),
    b = Object.entries(right);
  return (
    a.length === b.length &&
    a.every(
      ([key, value]) =>
        Object.hasOwn(right, key) &&
        equalJSONResource(value, (right as Record<string, unknown>)[key]),
    )
  );
}

export function validCommonJob(
  job: CommonJobResource,
  incidentId: string,
  jobId?: string,
): boolean {
  if (
    job.scope.kind !== "incident" ||
    job.scope.incident_id !== incidentId ||
    (jobId !== undefined && job.job_id !== jobId) ||
    job.status_route !== `/api/v1/jobs/${job.job_id}`
  )
    return false;
  const terminal = terminalCommonJob(job);
  if ((terminal || job.status === "cancel_requested") && job.cancelable)
    return false;
  if (
    job.progress.completed < 0 ||
    !Number.isSafeInteger(job.progress.completed) ||
    (job.progress.total !== null &&
      (!Number.isSafeInteger(job.progress.total) ||
        job.progress.total <= 0 ||
        job.progress.completed > job.progress.total))
  )
    return false;
  if (
    job.status === "succeeded" &&
    job.progress.total !== null &&
    job.progress.completed !== job.progress.total
  )
    return false;
  const submitted = Date.parse(job.submitted_at),
    updated = Date.parse(job.updated_at);
  if (
    !Number.isFinite(submitted) ||
    !Number.isFinite(updated) ||
    updated < submitted
  )
    return false;
  if (
    job.started_at !== null &&
    (!Number.isFinite(Date.parse(job.started_at)) ||
      Date.parse(job.started_at) < submitted ||
      Date.parse(job.started_at) > updated)
  )
    return false;
  if (terminal) {
    if (
      job.finished_at === null ||
      job.retained_until === null ||
      !Number.isFinite(Date.parse(job.finished_at)) ||
      Date.parse(job.finished_at) > updated ||
      Date.parse(job.finished_at) < submitted ||
      !Number.isFinite(Date.parse(job.retained_until)) ||
      Date.parse(job.retained_until) <
        Date.parse(job.finished_at) + 7 * 86_400_000
    )
      return false;
  } else if (
    job.finished_at !== null ||
    job.retained_until !== null ||
    job.result_summary !== null ||
    job.error_summary !== null
  )
    return false;
  if (job.status === "queued" && job.started_at !== null) return false;
  if (
    ["running", "succeeded", "failed"].includes(job.status) &&
    job.started_at === null
  )
    return false;
  if (
    job.status === "failed" &&
    (job.error_summary === null || job.result_summary !== null)
  )
    return false;
  if (
    (job.status === "succeeded" || job.status === "canceled") &&
    (job.result_summary === null || job.error_summary !== null)
  )
    return false;
  if (job.status === "canceled" && jobResultSummaryCode(job) !== "job_canceled")
    return false;
  const refs = job.result_summary?.resource_refs ?? [];
  return (
    new Set(refs.map((ref) => `${ref.kind}:${ref.id}`)).size === refs.length
  );
}

export function commonJobDoesNotRegress(
  previous: CommonJobResource,
  next: CommonJobResource,
): boolean {
  if (
    previous.job_id !== next.job_id ||
    !equalJSONResource(previous.scope, next.scope) ||
    previous.status_route !== next.status_route ||
    (previous.started_at !== null && previous.started_at !== next.started_at) ||
    previous.submitted_by_user_id !== next.submitted_by_user_id ||
    previous.submitted_at !== next.submitted_at ||
    Date.parse(next.updated_at) < Date.parse(previous.updated_at) ||
    next.progress.completed < previous.progress.completed ||
    (previous.progress.total !== null &&
      (next.progress.total === null ||
        next.progress.total < previous.progress.total))
  )
    return false;
  if (terminalCommonJob(previous)) return equalJSONResource(previous, next);
  if (
    previous.status === "cancel_requested" &&
    ["queued", "running"].includes(next.status)
  )
    return false;
  return !(previous.status === "running" && next.status === "queued");
}
