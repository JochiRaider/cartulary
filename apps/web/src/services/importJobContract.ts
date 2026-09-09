import { validCommonJob } from "./commonJobContract";
import type { ImportJobResource } from "./importContractAdapter";

/** Cross-resource rules supplement generated envelope validation. */
export function validWorkbookImportJob(
  job: ImportJobResource,
  incidentId: string,
  jobId?: string,
  sessionId?: string,
  purpose?: "discovery" | "apply",
): boolean {
  if (!validCommonJob(job, incidentId, jobId)) return false;
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

export function importSessionIdFromReceipt(
  job: ImportJobResource,
): string | null {
  return (
    job.result_summary?.resource_refs?.find(
      (ref) => ref.kind === "import_session",
    )?.id ?? null
  );
}
