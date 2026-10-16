import type {
  CancelJobRequest,
  CreateReleaseRequest,
  CreateSnapshotRequest,
  GetJobResponse,
  GetSnapshotResponse,
} from "@cartulary/protocol-ts/http";
import { fetchHTTPOperation, type HTTPOperationResult } from "./browserApi";

/** Exact built-in template projection; internal redaction is resolved by admission. */
export const defaultReportTemplate = {
  id: "cartulary.report.default",
  version: "4",
} as const;
export type ReportingJob = GetJobResponse["data"];
export type ReportingCancellation = Readonly<CancelJobRequest>;
export type ReportingReceipt = HTTPOperationResult<GetJobResponse>;
export type ReportingAttempt =
  | {
      readonly kind: "snapshot";
      readonly request: Readonly<CreateSnapshotRequest>;
    }
  | {
      readonly kind: "report";
      readonly request: Readonly<CreateReleaseRequest>;
    };
export interface ReportingClient {
  submit(
    attempt: ReportingAttempt,
    signal: AbortSignal,
  ): Promise<ReportingReceipt>;
  read(jobId: string, signal: AbortSignal): Promise<ReportingReceipt>;
  cancel(
    jobId: string,
    request: Readonly<CancelJobRequest>,
    signal: AbortSignal,
  ): Promise<ReportingReceipt>;
  snapshot(
    snapshotId: string,
    signal: AbortSignal,
  ): Promise<HTTPOperationResult<GetSnapshotResponse>>;
}
export function createReportingClient(apiBase?: string): ReportingClient {
  return {
    submit: (attempt, signal) =>
      fetchHTTPOperation<GetJobResponse>({
        apiBase,
        operationID:
          attempt.kind === "snapshot" ? "createSnapshot" : "createRelease",
        init: { method: "POST", body: JSON.stringify(attempt.request), signal },
      }),
    read: (jobId, signal) =>
      fetchHTTPOperation<GetJobResponse>({
        apiBase,
        operationID: "getJob",
        pathParameters: { job_id: jobId },
        init: { signal, cache: "no-store" },
      }),
    cancel: (jobId, request, signal) =>
      fetchHTTPOperation<GetJobResponse>({
        apiBase,
        operationID: "cancelJob",
        pathParameters: { job_id: jobId },
        init: { method: "POST", body: JSON.stringify(request), signal },
      }),
    snapshot: (snapshotId, signal) =>
      fetchHTTPOperation<GetSnapshotResponse>({
        apiBase,
        operationID: "getSnapshot",
        pathParameters: { snapshot_id: snapshotId },
        init: { signal, cache: "no-store" },
      }),
  };
}
