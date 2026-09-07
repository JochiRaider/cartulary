import type { GetJobResponse } from "@cartulary/protocol-ts/http";

export const importActorID = "00000000-0000-4000-8000-000000005001";
export const importedIncidentID = "00000000-0000-4000-8000-000000005002";
export const importJobID = "00000000-0000-4000-8000-000000005003";
export type ImportTestJob = GetJobResponse["data"];
export function importJob(
  status: ImportTestJob["status"] = "queued",
  overrides: Partial<ImportTestJob> = {},
): ImportTestJob {
  const terminal = ["succeeded", "failed", "canceled"].includes(status);
  const job = {
    job_id: importJobID,
    scope: { kind: "deployment" as const },
    status_route: `/api/v1/jobs/${importJobID}`,
    status,
    cancelable: status === "queued" || status === "running",
    submitted_by_user_id: importActorID,
    submitted_at: "2026-09-07T12:00:00Z",
    updated_at: "2026-09-07T12:00:01Z",
    started_at: status === "queued" ? null : "2026-09-07T12:00:01Z",
    finished_at: terminal ? "2026-09-07T12:00:01Z" : null,
    retained_until: terminal ? "2099-09-14T12:00:01Z" : null,
    progress: { completed: status === "succeeded" ? 1 : 0, total: null },
    error_summary:
      status === "failed"
        ? {
            code: "incident_bundle_import_rejected",
            message: "Import rejected",
            retryable: false,
          }
        : null,
    result_summary:
      status === "succeeded"
        ? {
            code: "incident_bundle_imported",
            message: "Imported",
            resource_refs: [
              {
                kind: "incident",
                id: importedIncidentID,
                route: `/api/v1/incidents/${importedIncidentID}`,
              },
            ],
          }
        : status === "canceled"
          ? { code: "job_canceled", message: "Canceled" }
          : null,
    ...overrides,
  };
  return job;
}
export function jobEnvelope(job = importJob()): GetJobResponse {
  return { data: job, meta: { request_id: "request-import-test" } };
}
export function readTestBlob(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result));
    reader.onerror = reject;
    reader.readAsText(blob);
  });
}
