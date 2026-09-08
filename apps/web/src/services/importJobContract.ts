import type { ImportJobResource } from "./importContractAdapter";

export const terminalImportJob = (job: ImportJobResource) =>
  job.status === "succeeded" ||
  job.status === "failed" ||
  job.status === "canceled";

/** Compare JSON resource values independently of transport property order. */
export function equalImportResource(left: unknown, right: unknown): boolean {
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
        equalImportResource(value, (right as Record<string, unknown>)[key]),
    )
  );
}

/** Cross-resource rules supplement generated envelope validation. */
export function validWorkbookImportJob(
  job: ImportJobResource,
  incidentId: string,
  jobId?: string,
  sessionId?: string,
  purpose?: "discovery" | "apply",
): boolean {
  if (
    job.scope.kind !== "incident" ||
    job.scope.incident_id !== incidentId ||
    (jobId !== undefined && job.job_id !== jobId) ||
    job.status_route !== `/api/v1/jobs/${job.job_id}`
  )
    return false;
  const terminal = terminalImportJob(job);
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
  if (job.status === "canceled" && job.result_summary?.code !== "job_canceled")
    return false;
  const refs = job.result_summary?.resource_refs ?? [];
  const sessions = refs.filter((ref) => ref.kind === "import_session");
  if (
    sessions.length > 1 ||
    sessions.some(
      (ref) =>
        ref.route !== `/api/v1/import-sessions/${ref.id}` ||
        (sessionId !== undefined && ref.id !== sessionId),
    )
  )
    return false;
  if (
    job.status === "succeeded" &&
    (![
      "import_session_discovered",
      "import_session_applied",
      "import_session_partially_applied",
    ].includes(job.result_summary?.code ?? "") ||
      sessions.length !== 1)
  )
    return false;
  if (
    job.status === "succeeded" &&
    purpose !== undefined &&
    (purpose === "discovery"
      ? job.result_summary?.code !== "import_session_discovered" ||
        refs.length !== 1
      : job.result_summary?.code === "import_session_discovered")
  )
    return false;
  const seen = new Set<string>();
  for (const ref of refs) {
    if (seen.has(`${ref.kind}:${ref.id}`)) return false;
    seen.add(`${ref.kind}:${ref.id}`);
    // Unknown additive kinds remain message-only; never become navigation authority.
    if (
      ref.kind === "network_flow_table" &&
      ref.route !==
        `/api/v1/incidents/${incidentId}/network-flow/tables/${ref.id}`
    )
      return false;
  }
  return true;
}

export function importJobDoesNotRegress(
  previous: ImportJobResource,
  next: ImportJobResource,
): boolean {
  if (
    previous.job_id !== next.job_id ||
    previous.submitted_by_user_id !== next.submitted_by_user_id ||
    previous.submitted_at !== next.submitted_at ||
    Date.parse(next.updated_at) < Date.parse(previous.updated_at) ||
    next.progress.completed < previous.progress.completed ||
    (previous.progress.total !== null &&
      (next.progress.total === null ||
        next.progress.total < previous.progress.total))
  )
    return false;
  if (terminalImportJob(previous)) return equalImportResource(previous, next);
  if (
    previous.status === "cancel_requested" &&
    ["queued", "running"].includes(next.status)
  )
    return false;
  return !(previous.status === "running" && next.status === "queued");
}

export function importSessionIdFromReceipt(
  job: ImportJobResource,
): string | null {
  return (
    job.result_summary?.resource_refs?.find(
      (ref) => ref.kind === "import_session",
    )?.id ?? null
  );
}
