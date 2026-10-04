import {
  type ActivateReferencePackVersionResponse,
  buildHTTPOperationPath,
  type CancelJobRequest,
  type CancelJobResponse,
  type GetJobResponse,
  type GetReferencePackValidationSummaryResponse,
  type GetReferencePackVersionResponse,
  type HTTPOperationID,
  type ImportReferencePackResponse,
  type ListReferencePacksResponse,
  validateHTTPOperationResponse,
} from "@cartulary/protocol-ts/http";
import {
  fetchHTTPOperation,
  fetchMultipartHTTPOperation,
  type HTTPOperationResult,
} from "./browserApi";
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
  readonly code: string;
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
function problem(status: number, payload: unknown): ReferencePackProblem {
  const error = record(record(payload)?.error);
  const code =
    typeof error?.code === "string" ? error.code : "request_rejected";
  const reasonValue = record(error?.details)?.reason_code;
  const reason =
    typeof reasonValue === "string"
      ? validatedPublicErrorReason(code, reasonValue)
      : undefined;
  const kind =
    code === "reference_pack_operation_rejected"
      ? "operation_rejected"
      : code === "reference_pack_state_conflict"
        ? "state_conflict"
        : code === "reference_pack_activation_rejected"
          ? "activation_rejected"
          : code === "reference_pack_verification_failed"
            ? "verification_failed"
            : code === "client_txn_conflict"
              ? "transaction_conflict"
              : code === "job_not_found" || code === "reference_pack_not_found"
                ? "unavailable"
                : "rejected";
  return { kind, status, code, ...(reason === undefined ? {} : { reason }) };
}
export function referencePackJobProblem(
  job: ReferencePackJobResource,
): ReferencePackProblem {
  return problem(200, { error: job.error_summary });
}
type Received<T> =
  | { readonly kind: "value"; readonly value: T; readonly status: number }
  | { readonly kind: "failure"; readonly problem: ReferencePackProblem };
async function receive<T>(
  operation: HTTPOperationID,
  execute: (
    onResponse: (response: Response) => void,
  ) => Promise<HTTPOperationResult<T>>,
): Promise<Received<T>> {
  let receivedStatus = 0;
  try {
    const result = await execute((response) => {
      receivedStatus = response.status;
    });
    const status = receivedStatus || result.status;
    if (status < 200 || status >= 300)
      return { kind: "failure", problem: problem(status, result.payload) };
    if (
      !result.ok ||
      !validateHTTPOperationResponse(operation, result.payload, status).ok
    )
      return {
        kind: "failure",
        problem: { ...referencePackContractProblem, status },
      };
    return { kind: "value", value: result.payload, status };
  } catch {
    return {
      kind: "failure",
      problem:
        receivedStatus >= 400
          ? problem(receivedStatus, null)
          : receivedStatus
            ? { ...referencePackContractProblem, status: receivedStatus }
            : referencePackTransportProblem,
    };
  }
}
function failure<T>(
  received: Extract<Received<T>, { kind: "failure" }>,
): { kind: "failed"; problem: ReferencePackProblem } | AccessFailure {
  const status = received.problem.status;
  return status === 401 || status === 403
    ? { kind: "access_failed", status }
    : { kind: "failed", problem: received.problem };
}
function rejectedOrUncertain(
  problem: ReferencePackProblem,
):
  | { kind: "rejected" | "uncertain"; problem: ReferencePackProblem }
  | AccessFailure {
  if (problem.status === 401 || problem.status === 403)
    return { kind: "access_failed", status: problem.status };
  return {
    kind:
      problem.status >= 400 && problem.status < 500 && problem.status !== 408
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
  const result = await receive("listReferencePacks", (onResponse) =>
    fetchHTTPOperation<ListReferencePacksResponse>({
      operationID: "listReferencePacks",
      init: { signal: options.signal, cache: "no-store" },
      onResponse,
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
  const result = await receive("getReferencePackVersion", (onResponse) =>
    fetchHTTPOperation<GetReferencePackVersionResponse>({
      operationID: "getReferencePackVersion",
      pathParameters: target,
      init: { signal, cache: "no-store" },
      onResponse,
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
  const result = await receive(
    "getReferencePackValidationSummary",
    (onResponse) =>
      fetchHTTPOperation<GetReferencePackValidationSummaryResponse>({
        operationID: "getReferencePackValidationSummary",
        pathParameters: { summary_id: reference },
        init: { signal, cache: "no-store" },
        onResponse,
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
    return job.result_summary.code === "job_canceled";
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
  const result = await receive("getJob", (onResponse) =>
    fetchHTTPOperation<GetJobResponse>({
      operationID: "getJob",
      pathParameters: { job_id: jobID },
      init: { signal, cache: "no-store" },
      onResponse,
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
  >(operation, (onResponse) => {
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
        body,
        init: { signal },
        onResponse,
      });
    }
    return fetchHTTPOperation({
      operationID: operation,
      ...("target" in command ? { pathParameters: command.target } : {}),
      onResponse,
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
  if (result.kind === "failure") return rejectedOrUncertain(result.problem);
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
): Promise<ReferencePackCancellation> {
  const result = await receive("cancelJob", (onResponse) =>
    fetchHTTPOperation<CancelJobResponse>({
      operationID: "cancelJob",
      pathParameters: { job_id: jobID },
      onResponse,
      init: { signal, method: "POST", body: JSON.stringify(request) },
    }),
  );
  if (result.kind === "failure") return rejectedOrUncertain(result.problem);
  return validReferencePackJob(result.value.data, jobID)
    ? { kind: "acknowledged", job: result.value.data }
    : { kind: "uncertain", problem: referencePackContractProblem };
}
