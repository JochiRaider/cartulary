import type {
  CancelJobRequest,
  GetJobResponse,
  ImportIncidentBundleRequest,
} from "@cartulary/protocol-ts/http";
import {
  fetchHTTPOperation,
  fetchMultipartHTTPOperation,
  type HTTPOperationResult,
} from "../../services/browserApi";
import { jobResultSummaryCode } from "../../services/commonJobContract";
import { isUnappliedRefusal } from "../../shared/publicError";

export type IncidentImportJob = GetJobResponse["data"];
export type ImportAttempt = {
  readonly file: Blob;
  readonly filename: string;
  readonly metadata: Readonly<ImportIncidentBundleRequest["metadata"]>;
};
export type ImportCancelAttempt = Readonly<CancelJobRequest>;
export type ImportProblem =
  | "transport"
  | "contract"
  | "rejected"
  | "conflict"
  | "unavailable"
  | "temporarily_unavailable";
type AccessFailure = {
  readonly kind: "access_failed";
  readonly status: 401 | 403;
};
type ProblemOutcome<K extends string> = {
  readonly kind: K;
  readonly problem: ImportProblem;
};
export type ImportAdmissionOutcome =
  | { readonly kind: "accepted"; readonly job: IncidentImportJob }
  | ProblemOutcome<"rejected" | "uncertain">
  | AccessFailure;
export type ImportObservationOutcome =
  | { readonly kind: "observed"; readonly job: IncidentImportJob }
  | ProblemOutcome<"failed">
  | { readonly kind: "unavailable" }
  | AccessFailure;
export type ImportCancellationOutcome =
  | { readonly kind: "acknowledged"; readonly job: IncidentImportJob }
  | ProblemOutcome<"rejected" | "uncertain">
  | { readonly kind: "unavailable" }
  | AccessFailure;
type JobReceipt =
  | { kind: "job"; job: IncidentImportJob }
  | { kind: "failure"; status: number; problem: ImportProblem };

const mediaTypes = new Set([
  "application/zip",
  "application/x-tar",
  "application/gzip",
  "application/x-gzip",
  "application/octet-stream",
]);
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
export const terminalImportJob = (job: IncidentImportJob) =>
  job.status === "succeeded" ||
  job.status === "failed" ||
  job.status === "canceled";

/** Blob slicing changes only the envelope hint; no archive bytes are materialized. */
export function captureImport(
  file: File,
  transactionId: string,
): ImportAttempt {
  const hint = file.type.split(";")[0]?.trim().toLowerCase() ?? "";
  return Object.freeze({
    file: mediaTypes.has(hint)
      ? file
      : file.slice(0, file.size, "application/octet-stream"),
    filename: file.name,
    metadata: Object.freeze({ client_txn_id: transactionId }),
  });
}
export async function importIncidentBundle(
  attempt: ImportAttempt,
  signal: AbortSignal,
  replay = false,
): Promise<ImportAdmissionOutcome> {
  const result = await receiveJob(
    (onResponse) => {
      const body = new FormData();
      body.append(
        "metadata",
        new Blob([JSON.stringify(attempt.metadata)], {
          type: "application/json",
        }),
      );
      body.append("file", attempt.file, attempt.filename);
      return fetchMultipartHTTPOperation({
        operationID: "importIncidentBundle",
        body,
        init: { signal },
        onResponse,
      });
    },
    202,
    undefined,
    replay,
  );
  if (result.kind === "job") return { kind: "accepted", job: result.job };
  if (result.status === 401 || result.status === 403)
    return { kind: "access_failed", status: result.status };
  return {
    kind: writeSettlement(result, replay),
    problem: result.problem,
  };
}
export async function readImportJob(
  jobId: string,
  signal: AbortSignal,
): Promise<ImportObservationOutcome> {
  const result = await receiveJob(
    (onResponse) =>
      fetchHTTPOperation({
        operationID: "getJob",
        pathParameters: { job_id: jobId },
        init: { signal, cache: "no-store" },
        onResponse,
      }),
    200,
    jobId,
  );
  if (result.kind === "job") return { kind: "observed", job: result.job };
  if (result.status === 401 || result.status === 403)
    return { kind: "access_failed", status: result.status };
  if (result.status === 404) return { kind: "unavailable" };
  return { kind: "failed", problem: result.problem };
}
export async function cancelImportJob(
  jobId: string,
  attempt: ImportCancelAttempt,
  signal: AbortSignal,
  replay = false,
): Promise<ImportCancellationOutcome> {
  const result = await receiveJob(
    (onResponse) =>
      fetchHTTPOperation({
        operationID: "cancelJob",
        pathParameters: { job_id: jobId },
        init: { method: "POST", body: JSON.stringify(attempt), signal },
        onResponse,
      }),
    200,
    jobId,
  );
  if (result.kind === "job") return { kind: "acknowledged", job: result.job };
  if (result.status === 401 || result.status === 403)
    return { kind: "access_failed", status: result.status };
  if (result.status === 404) return { kind: "unavailable" };
  return {
    kind: writeSettlement(result, replay),
    problem: result.problem,
  };
}
const definitiveRejection = (status: number) =>
  status >= 400 && status < 500 && status !== 408;
/**
 * A decoded `503` `service_unavailable` was refused before it was applied: a
 * definite rejection on a first attempt, while a refused replay leaves the
 * earlier outcome uncertain (Core 03 REQ-03-086a).
 */
const writeSettlement = (
  failure: { readonly status: number; readonly problem: ImportProblem },
  replay: boolean,
): "rejected" | "uncertain" =>
  definitiveRejection(failure.status) ||
  (!replay && failure.problem === "temporarily_unavailable")
    ? "rejected"
    : "uncertain";
/** Operations whose decoded results drive incident bundle import. */
type IncidentImportHTTPOperation =
  | "importIncidentBundle"
  | "getJob"
  | "cancelJob";
/**
 * Classifies one job operation from its decoded contract outcome. A contract
 * failure or an unreadable received body never selects an outcome by its raw
 * HTTP status (Core 03 REQ-03-086a).
 */
async function receiveJob(
  run: (
    onResponse: (response: Response) => void,
  ) => Promise<
    HTTPOperationResult<GetJobResponse, IncidentImportHTTPOperation>
  >,
  expectedStatus: number,
  jobId?: string,
  replay = false,
): Promise<JobReceipt> {
  let received = 0;
  try {
    const result = await run((response) => {
      received = response.status;
    });
    if (result.ok) {
      if (
        result.status === expectedStatus &&
        validImportJob(result.payload.data, jobId) &&
        (expectedStatus !== 202 ||
          replay ||
          ["queued", "running"].includes(result.payload.data.status))
      )
        return { kind: "job", job: result.payload.data };
      return { kind: "failure", status: 502, problem: "contract" };
    }
    const error = result.payload.error;
    if (error.code === "invalid_public_contract_response")
      return { kind: "failure", status: error.status, problem: "contract" };
    return {
      kind: "failure",
      status: error.status,
      problem:
        error.code === "client_txn_conflict"
          ? "conflict"
          : error.status === 404
            ? "unavailable"
            : definitiveRejection(error.status)
              ? "rejected"
              : isUnappliedRefusal(error.status, error)
                ? "temporarily_unavailable"
                : "transport",
    };
  } catch {
    // A received but unreadable body fails the contract; only a missing
    // response is transport loss.
    return received === 0
      ? { kind: "failure", status: 0, problem: "transport" }
      : { kind: "failure", status: 502, problem: "contract" };
  }
}
/** Cross-field owner rules supplement the generated closed envelope validator. */
function validImportJob(job: IncidentImportJob, expectedId?: string): boolean {
  if (
    !uuid.test(job.job_id) ||
    (expectedId !== undefined && job.job_id !== expectedId) ||
    job.scope.kind !== "deployment" ||
    job.status_route !== `/api/v1/jobs/${job.job_id}`
  )
    return false;
  const { completed, total } = job.progress;
  if (
    !Number.isSafeInteger(completed) ||
    completed < 0 ||
    (total !== null &&
      (!Number.isSafeInteger(total) || total <= 0 || completed > total))
  )
    return false;
  const terminal = terminalImportJob(job);
  const submitted = Date.parse(job.submitted_at);
  const updated = Date.parse(job.updated_at);
  const started =
    job.started_at === null ? submitted : Date.parse(job.started_at);
  const finished =
    job.finished_at === null ? updated : Date.parse(job.finished_at);
  if (
    ![submitted, updated, started, finished].every(Number.isFinite) ||
    started < submitted ||
    updated < started ||
    finished < started ||
    updated < finished
  )
    return false;
  if (job.cancelable && !["queued", "running"].includes(job.status))
    return false;
  if (job.status === "queued" && job.started_at !== null) return false;
  if (job.status === "running" && job.started_at === null) return false;
  if (!terminal)
    return (
      job.finished_at === null &&
      job.retained_until === null &&
      job.result_summary === null &&
      job.error_summary === null
    );
  if (
    job.finished_at === null ||
    job.retained_until === null ||
    Date.parse(job.retained_until) <
      Date.parse(job.finished_at) + 7 * 86_400_000
  )
    return false;
  if (job.status === "failed")
    return job.error_summary !== null && job.result_summary === null;
  if (job.error_summary !== null || job.result_summary === null) return false;
  if (job.status === "canceled")
    return jobResultSummaryCode(job) === "job_canceled";
  return total === null || completed === total;
}
export function importJobAdvances(
  previous: IncidentImportJob,
  next: IncidentImportJob,
): boolean {
  if (
    !validImportJob(next, previous.job_id) ||
    next.submitted_by_user_id !== previous.submitted_by_user_id ||
    next.submitted_at !== previous.submitted_at ||
    Date.parse(next.updated_at) < Date.parse(previous.updated_at) ||
    next.progress.completed < previous.progress.completed ||
    (previous.progress.total !== null &&
      (next.progress.total === null ||
        next.progress.total < previous.progress.total))
  )
    return false;
  if (previous.started_at !== null && next.started_at !== previous.started_at)
    return false;
  if (
    previous.finished_at !== null &&
    next.finished_at !== previous.finished_at
  )
    return false;
  if (terminalImportJob(previous)) {
    const target = importedIncidentTarget(previous);
    return (
      next.status === previous.status &&
      next.result_summary?.code === previous.result_summary?.code &&
      next.error_summary?.code === previous.error_summary?.code &&
      (target === null || importedIncidentTarget(next) === target)
    );
  }
  if (previous.status === "cancel_requested")
    return next.status !== "queued" && next.status !== "running";
  return previous.status !== "running" || next.status !== "queued";
}
const supportedKinds = new Set([
  "incident",
  "import_session",
  "snapshot",
  "release",
  "reference_pack_version",
  "incident_bundle",
  "network_flow_table",
]);
export function importedIncidentTarget(job: IncidentImportJob): string | null {
  if (
    job.status !== "succeeded" ||
    jobResultSummaryCode(job) !== "incident_bundle_imported" ||
    job.result_summary === null
  )
    return null;
  const refs =
    job.result_summary.resource_refs?.filter((ref) =>
      supportedKinds.has(ref.kind),
    ) ?? [];
  const ref = refs[0];
  if (
    refs.length !== 1 ||
    ref?.kind !== "incident" ||
    !uuid.test(ref.id) ||
    ref.route === undefined ||
    !/^\/api\/v1\/[^?#]+$/.test(ref.route)
  )
    return null;
  return ref.id;
}
