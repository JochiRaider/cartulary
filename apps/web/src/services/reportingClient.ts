import type {
  CancelJobRequest,
  CreateReleaseRequest,
  CreateSnapshotRequest,
  GetJobResponse,
  GetRetainedReportArtifactsResponse,
  GetSnapshotResponse,
} from "@cartulary/protocol-ts/http";
import {
  apiPath,
  fetchHTTPOperation,
  type HTTPOperationResult,
} from "./browserApi";

/** Exact built-in template projection; internal redaction is resolved by admission. */
export const defaultReportTemplate = {
  id: "cartulary.report.default",
  version: "4",
} as const;
export type ReportingJob = GetJobResponse["data"];
export type ReportingCancellation = Readonly<CancelJobRequest>;
export type ReportingReceipt = HTTPOperationResult<GetJobResponse>;
export type RetainedReportArtifact =
  GetRetainedReportArtifactsResponse["data"]["artifacts"][number];
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
  retainedArtifacts(
    incidentId: string,
    signal: AbortSignal,
  ): Promise<HTTPOperationResult<GetRetainedReportArtifactsResponse>>;
  retainedArtifactHref(incidentId: string, artifactId: string): string;
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
    retainedArtifacts: (incidentId, signal) =>
      fetchHTTPOperation<GetRetainedReportArtifactsResponse>({
        apiBase,
        operationID: "getRetainedReportArtifacts",
        pathParameters: { incident_id: incidentId },
        init: { signal, cache: "no-store" },
      }),
    retainedArtifactHref: (incidentId, artifactId) =>
      apiPath(
        apiBase,
        `/api/v1/incidents/${encodeURIComponent(incidentId)}/retained-report-artifacts/${encodeURIComponent(artifactId)}/content`,
      ),
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
