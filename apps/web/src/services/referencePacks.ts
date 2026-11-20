import {
  type ActivateReferencePackVersionResponse,
  buildHTTPOperationPath,
  type CancelJobRequest,
  type GetJobResponse,
  type GetReferencePackValidationSummaryResponse,
  type ImportReferencePackResponse,
  type ListReferencePacksResponse,
} from "@cartulary/protocol-ts/http";
import {
  type ClientErrorCode,
  isUnappliedRefusal,
} from "../shared/publicError";
import {
  fetchHTTPOperation,
  fetchMultipartHTTPOperation,
  type HTTPOperationError,
  type HTTPOperationFailureCode,
  type HTTPOperationResult,
} from "./browserApi";
import { jobResultSummaryCode } from "./commonJobContract";
import { validatedPublicErrorReason } from "./publicErrorIdentity";

export type ReferencePackAction =
  | "activate"
  | "disable"
  | "reverify"
  | "remove";
export type ReferencePackVersion =
  ListReferencePacksResponse["data"]["pack_versions"][number];
export type ReferencePackJobResource = GetJobResponse["data"];
export type ReferencePackValidationSummary =
  GetReferencePackValidationSummaryResponse["data"];
export type ReferencePackPaging = NonNullable<
  ListReferencePacksResponse["meta"]["paging"]
>;
export type ReferencePackQuery = {
  active: string;
  packVersionState: string;
  search: string;
};
export type ReferencePackTarget = Readonly<{
  pack_key: string;
  pack_version: string;
}>;
export type ReferencePackCommand =
  | {
      readonly kind: Exclude<ReferencePackAction, "remove">;
      readonly target: ReferencePackTarget;
    }
  | {
      readonly kind: "remove";
      readonly target: ReferencePackTarget;
      readonly reason: string;
    }
  | { readonly kind: "refresh_all" }
  | { readonly kind: "refresh_selected"; readonly packKeys: readonly string[] }
  | { readonly kind: "import"; readonly filename: string };
export type ReferencePackAttempt = {
  readonly command: ReferencePackCommand;
  readonly request: Readonly<{ client_txn_id: string; reason?: string | null }>;
  readonly file?: Blob;
};
/** Operations whose failures Reference Pack administration decodes. */
type ReferencePackHTTPOperation =
  | "listReferencePacks"
  | "getReferencePackVersion"
  | "getReferencePackValidationSummary"
  | "getJob"
  | "cancelJob"
  | "activateReferencePackVersion"
  | "disableReferencePackVersion"
  | "reverifyReferencePackVersion"
  | "removeReferencePackVersion"
  | "refreshReferencePacks"
  | "importReferencePack";
/**
 * Declared operation codes, plus client-originated transport loss and local
 * rejection. Failed-job problems keep the job's own error summary code.
 */
export type ReferencePackProblemCode =
  | HTTPOperationFailureCode<ReferencePackHTTPOperation>
  | Extract<ClientErrorCode, "transport_unavailable" | "request_rejected">;
export type ReferencePackProblem = {
  readonly kind:
    | "transport"
    | "contract"
    | "rejected"
    | "state_conflict"
    | "activation_rejected"
    | "operation_rejected"
    | "verification_failed"
    | "transaction_conflict"
    | "unavailable";
  readonly status: number;
  readonly code: ReferencePackProblemCode;
  readonly reason?: string;
};
type AccessFailure = {
  readonly kind: "access_failed";
  readonly status: 401 | 403;
};
export type ReferencePackRead<T> =
  | { readonly kind: "read"; readonly value: T }
  | { readonly kind: "failed"; readonly problem: ReferencePackProblem }
  | AccessFailure;
export type ReferencePackMutation =
  | { readonly kind: "committed"; readonly pack: ReferencePackVersion }
  | { readonly kind: "accepted"; readonly job: ReferencePackJobResource }
  | {
      readonly kind: "rejected" | "uncertain";
      readonly problem: ReferencePackProblem;
    }
  | AccessFailure;
export type ReferencePackCancellation =
  | { readonly kind: "acknowledged"; readonly job: ReferencePackJobResource }
  | {
      readonly kind: "rejected" | "uncertain";
      readonly problem: ReferencePackProblem;
    }
  | AccessFailure;
export const referencePackContractProblem: ReferencePackProblem = {
  kind: "contract",
  status: 502,
  code: "invalid_public_contract_response",
};
export const referencePackTransportProblem: ReferencePackProblem = {
  kind: "transport",
  status: 0,
  code: "transport_unavailable",
};
const mediaTypes = new Set([
  "application/zip",
  "application/x-tar",
  "application/gzip",
  "application/x-gzip",
  "application/octet-stream",
]);
export function captureReferencePackAttempt(
  command: ReferencePackCommand,
  transactionId: string,
  file?: File,
): ReferencePackAttempt {
  const capturedCommand =
    command.kind === "refresh_selected"
      ? Object.freeze({
          ...command,
          packKeys: Object.freeze([...new Set(command.packKeys)].sort()),
        })
      : "target" in command
        ? Object.freeze({
            ...command,
            target: Object.freeze({ ...command.target }),
          })
        : Object.freeze({ ...command });
  if (command.kind === "refresh_selected" && command.packKeys.length === 0)
    throw new Error("An explicit refresh requires pack keys");
  if (command.kind === "import" && !file)
    throw new Error("An import requires a file");
  if (command.kind === "remove" && !command.reason.trim())
    throw new Error("Removal requires a reason");
  const hint = file?.type.split(";")[0]?.trim().toLowerCase() ?? "";
  return Object.freeze({
    command: capturedCommand,
    request: Object.freeze({
      client_txn_id: transactionId,
      ...(command.kind === "remove" ? { reason: command.reason } : {}),
    }),
    ...(file
      ? {
          file: mediaTypes.has(hint)
            ? file
            : file.slice(0, file.size, "application/octet-stream"),
        }
      : {}),
  });
}
function record(value: unknown): Record<string, unknown> | null {
  return value !== null && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : null;
}
/** A public error decoded by fetchHTTPOperation; contract failures never select a feature outcome. */
function problem(
  error: HTTPOperationError<ReferencePackHTTPOperation>,
): ReferencePackProblem {
  if (error.code === "invalid_public_contract_response")
    return {
      ...referencePackContractProblem,
      status: error.details.response_status,
    };
  const reason = validatedPublicErrorReason(
    error.code,
    error.details.reason_code,
  );
  const kind =
    error.code === "reference_pack_operation_rejected"
      ? "operation_rejected"
      : error.code === "reference_pack_activation_rejected"
        ? "activation_rejected"
        : error.code === "reference_pack_verification_failed"
          ? "verification_failed"
          : error.code === "client_txn_conflict"
            ? "transaction_conflict"
            : error.code === "job_not_found" ||
                error.code === "reference_pack_not_found"
              ? "unavailable"
              : "rejected";
  return {
    kind,
    status: error.status,
    code: error.code,
    ...(reason === undefined ? {} : { reason }),
  };
}
/** A failed job's error summary is job payload, not an operation error response. */
export function referencePackJobProblem(
  job: ReferencePackJobResource,
): ReferencePackProblem {
  const summary = job.error_summary;
  const reasonValue = record(summary?.details)?.reason_code;
  const kind =
    summary?.code === "reference_pack_operation_rejected"
      ? "operation_rejected"
      : summary?.code === "reference_pack_activation_rejected"
        ? "activation_rejected"
        : summary?.code === "reference_pack_verification_failed"
          ? "verification_failed"
          : "rejected";
  const reason =
    kind === "rejected" || typeof reasonValue !== "string"
      ? undefined
      : validatedPublicErrorReason(summary?.code ?? "", reasonValue);
  return {
    kind,
    status: 200,
    code:
      kind === "operation_rejected"
        ? "reference_pack_operation_rejected"
        : kind === "activation_rejected"
          ? "reference_pack_activation_rejected"
          : kind === "verification_failed"
            ? "reference_pack_verification_failed"
            : "request_rejected",
    ...(reason === undefined ? {} : { reason }),
  };
}
type Received<T> =
  | { readonly kind: "value"; readonly value: T; readonly status: number }
  | { readonly kind: "failure"; readonly problem: ReferencePackProblem };
/**
 * Classifies one operation result from its decoded contract outcome only. A
 * contract failure is handled like an invalid success response and never by
 * its raw HTTP status (Core 03 REQ-03-086a).
 */
async function receive<T>(
  execute: (
    onResponse: (response: Response) => void,
  ) => Promise<HTTPOperationResult<T, ReferencePackHTTPOperation>>,
): Promise<Received<T>> {
  let receivedStatus = 0;
  try {
    const result = await execute((response) => {
      receivedStatus = response.status;
    });
    return result.ok
      ? { kind: "value", value: result.payload, status: result.status }
      : { kind: "failure", problem: problem(result.payload.error) };
  } catch {
    // A received but unreadable body fails the contract; only a missing
    // response is transport loss.
    return {
      kind: "failure",
      problem: receivedStatus
        ? { ...referencePackContractProblem, status: receivedStatus }
        : referencePackTransportProblem,
    };
  }
}
const indeterminate = (problem: ReferencePackProblem) =>
  problem.kind === "contract" || problem.kind === "transport";
function failure<T>(
  received: Extract<Received<T>, { kind: "failure" }>,
): { kind: "failed"; problem: ReferencePackProblem } | AccessFailure {
  const status = received.problem.status;
  return !indeterminate(received.problem) && (status === 401 || status === 403)
    ? { kind: "access_failed", status }
    : { kind: "failed", problem: received.problem };
}
/**
 * A decoded `503` `service_unavailable` was refused before it was applied: a
 * definite rejection on a first attempt, while a refused replay leaves the
 * earlier outcome uncertain (Core 03 REQ-03-086a).
 */
function rejectedOrUncertain(
  problem: ReferencePackProblem,
  replay: boolean,
):
  | { kind: "rejected" | "uncertain"; problem: ReferencePackProblem }
  | AccessFailure {
  if (indeterminate(problem)) return { kind: "uncertain", problem };
  if (problem.status === 401 || problem.status === 403)
    return { kind: "access_failed", status: problem.status };
  return {
    kind:
      (problem.status >= 400 &&
        problem.status < 500 &&
        problem.status !== 408) ||
      (!replay && isUnappliedRefusal(problem.status, problem))
        ? "rejected"
        : "uncertain",
    problem,
  };
}
export async function listReferencePacks(options: {
  query: ReferencePackQuery;
  cursorToken?: string | null;
  signal: AbortSignal;
}): Promise<ReferencePackRead<ListReferencePacksResponse>> {
  const result = await receive((onResponse) =>
    fetchHTTPOperation({
      operationID: "listReferencePacks",
      onResponse,
      init: { signal: options.signal, cache: "no-store" },
      query: {
        limit: 100,
        ...(options.cursorToken ? { cursor_token: options.cursorToken } : {}),
        ...(options.query.search === ""
          ? {}
          : { search: options.query.search }),
        ...(options.query.active === ""
          ? {}
          : { active: options.query.active }),
        ...(options.query.packVersionState === ""
          ? {}
          : { pack_version_state: options.query.packVersionState }),
      },
    }),
  );
  if (result.kind === "failure") return failure(result);
  return { kind: "read", value: result.value };
}
export function referencePackVersionRoute(target: ReferencePackTarget) {
  return buildHTTPOperationPath("getReferencePackVersion", target);
}
export function referencePackResultTarget(ref: {
  kind: string;
  id: string;
  route?: string | null;
}): ReferencePackTarget | null {
  if (ref.kind !== "reference_pack_version" || ref.id !== ref.route)
    return null;
  const match = /^\/api\/v1\/reference-packs\/([^/?#]+)\/([^/?#]+)$/u.exec(
    ref.id,
  );
  if (!match) return null;
  try {
    const target = {
      pack_key: decodeURIComponent(match[1] ?? ""),
      pack_version: decodeURIComponent(match[2] ?? ""),
    };
    return referencePackVersionRoute(target) === ref.id ? target : null;
  } catch {
    return null;
  }
}
export async function readReferencePackVersion(
  target: ReferencePackTarget,
  signal: AbortSignal,
): Promise<ReferencePackRead<ReferencePackVersion>> {
  const result = await receive((onResponse) =>
    fetchHTTPOperation({
      operationID: "getReferencePackVersion",
      onResponse,
      pathParameters: target,
      init: { signal, cache: "no-store" },
    }),
  );
  if (result.kind === "failure") return failure(result);
  const pack = result.value.data;
  return pack.pack_key === target.pack_key &&
    pack.pack_version === target.pack_version
    ? { kind: "read", value: pack }
    : { kind: "failed", problem: referencePackContractProblem };
}
export function referencePackValidationRef(
  job: ReferencePackJobResource,
): string | null {
  if (
    job.status !== "failed" ||
    job.error_summary?.code !== "reference_pack_verification_failed"
  )
    return null;
  const ref = record(job.error_summary.details)?.validation_summary_ref;
  return typeof ref === "string" && /^rpvs_[0-9a-f]{64}$/u.test(ref)
    ? ref
    : null;
}
export function referencePackValidationMatches(
  job: ReferencePackJobResource,
  summary: ReferencePackValidationSummary,
) {
  const details = record(job.error_summary?.details);
  return (
    referencePackValidationRef(job) !== null &&
    summary.result === "failed" &&
    summary.primary_issue_id === details?.primary_issue_id &&
    summary.total_issue_count === details?.total_issue_count &&
    summary.retained_issue_count === details?.retained_issue_count &&
    summary.issues_truncated === details?.issues_truncated &&
    summary.issues.every((issue) => issue.check_id === details?.check_id)
  );
}
export async function readReferencePackValidationSummary(
  reference: string,
  signal: AbortSignal,
): Promise<ReferencePackRead<ReferencePackValidationSummary>> {
  if (!/^rpvs_[0-9a-f]{64}$/u.test(reference))
    return { kind: "failed", problem: referencePackContractProblem };
  const result = await receive((onResponse) =>
    fetchHTTPOperation({
      operationID: "getReferencePackValidationSummary",
      onResponse,
      pathParameters: { summary_id: reference },
      init: { signal, cache: "no-store" },
    }),
  );
  if (result.kind === "failure") return failure(result);
  const summary = result.value.data;
  const count = summary.issues.length;
  if (
    summary.result !== "failed" ||
    count === 0 ||
    summary.primary_issue_id !== summary.issues[0]?.issue_id ||
    summary.retained_issue_count !== count ||
    count !== Math.min(1000, summary.total_issue_count) ||
    summary.issues_truncated !== summary.total_issue_count > count ||
    new Set(summary.issues.map((issue) => issue.issue_id)).size !== count
  )
    return { kind: "failed", problem: referencePackContractProblem };
  return { kind: "read", value: summary };
}
const terminal = (job: ReferencePackJobResource) =>
  ["succeeded", "failed", "canceled"].includes(job.status);
export function validReferencePackJob(
  job: ReferencePackJobResource,
  expectedId?: string,
) {
  if (
    job.scope.kind !== "deployment" ||
    (expectedId !== undefined && job.job_id !== expectedId) ||
    job.status_route !==
      buildHTTPOperationPath("getJob", { job_id: job.job_id })
  )
    return false;
  if (
    job.progress.total !== null &&
    job.progress.completed > job.progress.total
  )
    return false;
  if (job.status === "queued" && job.started_at !== null) return false;
  if (job.status === "running" && job.started_at === null) return false;
  if (
    ["cancel_requested", "succeeded", "failed", "canceled"].includes(
      job.status,
    ) &&
    job.cancelable
  )
    return false;
  if (!terminal(job))
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
      Date.parse(job.finished_at) + 7 * 24 * 60 * 60 * 1000
  )
    return false;
  if (job.status === "failed")
    return job.error_summary !== null && job.result_summary === null;
  if (job.result_summary === null || job.error_summary !== null) return false;
  if (job.status === "canceled")
    return jobResultSummaryCode(job) === "job_canceled";
  if (
    job.progress.total !== null &&
    job.progress.completed !== job.progress.total
  )
    return false;
  return (job.result_summary.resource_refs ?? []).every(
    (ref) =>
      ref.kind !== "reference_pack_version" ||
      referencePackResultTarget(ref) !== null,
  );
}
export async function loadReferencePackJob(
  jobID: string,
  signal: AbortSignal,
): Promise<ReferencePackRead<ReferencePackJobResource>> {
  const result = await receive((onResponse) =>
    fetchHTTPOperation({
      operationID: "getJob",
      onResponse,
      pathParameters: { job_id: jobID },
      init: { signal, cache: "no-store" },
    }),
  );
  if (result.kind === "failure") return failure(result);
  return validReferencePackJob(result.value.data, jobID)
    ? { kind: "read", value: result.value.data }
    : { kind: "failed", problem: referencePackContractProblem };
}
const operations = {
  activate: "activateReferencePackVersion",
  disable: "disableReferencePackVersion",
  reverify: "reverifyReferencePackVersion",
  remove: "removeReferencePackVersion",
  refresh_all: "refreshReferencePacks",
  refresh_selected: "refreshReferencePacks",
  import: "importReferencePack",
} as const;
export async function submitReferencePackAttempt(
  attempt: ReferencePackAttempt,
  signal: AbortSignal,
  replay = false,
): Promise<ReferencePackMutation> {
  const command = attempt.command;
  const operation = operations[command.kind];
  const result = await receive<
    ActivateReferencePackVersionResponse | ImportReferencePackResponse
  >((onResponse) => {
    if (command.kind === "import") {
      if (!attempt.file) throw new Error("Missing captured file");
      const body = new FormData();
      body.append(
        "metadata",
        new Blob(
          [
            JSON.stringify({
              ...attempt.request,
              activation_policy: "staged_only",
            }),
          ],
          { type: "application/json" },
        ),
      );
      body.append("file", attempt.file, command.filename);
      return fetchMultipartHTTPOperation({
        operationID: operation,
        onResponse,
        body,
        init: { signal },
      });
    }
    return fetchHTTPOperation({
      operationID: operation,
      onResponse,
      ...("target" in command ? { pathParameters: command.target } : {}),
      init: {
        signal,
        method: "POST",
        body: JSON.stringify({
          ...attempt.request,
          ...(command.kind === "refresh_selected"
            ? { pack_keys: command.packKeys }
            : {}),
        }),
      },
    });
  });
  if (result.kind === "failure")
    return rejectedOrUncertain(result.problem, replay);
  if (result.status === 200) {
    if (!("pack_version" in result.value.data))
      return {
        kind: "uncertain",
        problem: { ...referencePackContractProblem, status: result.status },
      };
    const pack = result.value.data.pack_version;
    if (
      !("target" in command) ||
      pack.pack_key !== command.target.pack_key ||
      pack.pack_version !== command.target.pack_version
    )
      return { kind: "uncertain", problem: referencePackContractProblem };
    return { kind: "committed", pack };
  }
  // The status-specific generated validator above establishes the 202 job variant.
  const job = result.value.data as ReferencePackJobResource;
  if (
    !validReferencePackJob(job) ||
    (!replay && job.status !== "queued" && job.status !== "running")
  )
    return { kind: "uncertain", problem: referencePackContractProblem };
  return { kind: "accepted", job };
}
export async function cancelReferencePackJob(
  jobID: string,
  request: Readonly<CancelJobRequest>,
  signal: AbortSignal,
  replay = false,
): Promise<ReferencePackCancellation> {
  const result = await receive((onResponse) =>
    fetchHTTPOperation({
      operationID: "cancelJob",
      onResponse,
      pathParameters: { job_id: jobID },
      init: { signal, method: "POST", body: JSON.stringify(request) },
    }),
  );
  if (result.kind === "failure")
    return rejectedOrUncertain(result.problem, replay);
  return validReferencePackJob(result.value.data, jobID)
    ? { kind: "acknowledged", job: result.value.data }
    : { kind: "uncertain", problem: referencePackContractProblem };
}
