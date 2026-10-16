import {
  type AttachBlobToEvidenceRecordResponse,
  buildHTTPOperationPath,
  type CreateObjectBlobSlotResponse,
  type GetJobResponse,
} from "@cartulary/protocol-ts/http";
import { evidenceViewSchemaId } from "@cartulary/view-contracts";
import { fetchHTTPOperation } from "../../services/browserApi";
import { validCommonJob } from "../../services/commonJobContract";
import {
  uploadEvidenceObjectBlobTarget,
  validEvidenceObjectUploadTarget,
} from "../../services/workbookEvidence";
import type {
  EvidenceFileAttempt,
  EvidenceFileTransport,
} from "../features/evidence/evidenceFileOperation";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import { freezeWorkbookValue } from "../utils/freezeWorkbookValue";
import { sendWorkbookRecordMutation } from "./sendWorkbookRecordMutation";
import { classifyWorkbookOperationFailure } from "./workbookOperationErrorPolicy";
import { acceptedRecordMutation } from "./workbookRecordPatchTransport";

export function createEvidenceFileTransport(
  apiBase: string | undefined,
  readScope?: WorkbookReadScopeSource,
): EvidenceFileTransport {
  async function verificationRequest(
    attempt: EvidenceFileAttempt,
    signal: AbortSignal,
    operationID: "verifyObjectBlob" | "getJob" | "cancelJob",
    jobId?: string,
    clientTxnId?: string,
  ): ReturnType<EvidenceFileTransport["verify"]> {
    if (
      attempt.stage !== "verify" ||
      !attempt.objectBlobId ||
      (operationID !== "verifyObjectBlob" && !jobId) ||
      attempt.path !==
        buildHTTPOperationPath("verifyObjectBlob", {
          object_blob_id: attempt.objectBlobId,
        })
    )
      return { kind: "uncertain" };
    let status: number | null = null,
      requestId: string | null = null;
    try {
      const result = await fetchHTTPOperation<GetJobResponse>({
        apiBase: attempt.apiBase,
        operationID,
        pathParameters:
          operationID === "verifyObjectBlob"
            ? { object_blob_id: attempt.objectBlobId }
            : { job_id: jobId ?? "" },
        init: {
          method: operationID === "getJob" ? "GET" : "POST",
          ...(operationID === "getJob"
            ? {}
            : {
                body:
                  operationID === "verifyObjectBlob"
                    ? attempt.body
                    : JSON.stringify({ client_txn_id: clientTxnId }),
              }),
          signal,
          cache: "no-store",
        },
        onResponse: (response) => {
          status = response.status;
          requestId = response.headers.get("X-Request-ID");
        },
      });
      if (!result.ok)
        return rejectedOrUncertain(
          status,
          requestId,
          result.payload,
          operationID,
        );
      if (
        (operationID === "verifyObjectBlob"
          ? status !== 202
          : status !== 200 && status !== 202) ||
        !validRequestId(result.payload.meta.request_id, requestId) ||
        (result.payload.data.status === "succeeded" &&
          result.payload.data.result_summary?.code !==
            "evidence_blob_verified") ||
        !validCommonJob(
          result.payload.data,
          attempt.authority.incidentId,
          jobId,
        ) ||
        result.payload.data.submitted_by_user_id !== attempt.authority.actorId
      )
        return { kind: "uncertain" };
      return {
        kind: "accepted",
        receipt: freezeWorkbookValue(structuredClone(result.payload.data)),
      };
    } catch {
      return { kind: "uncertain" };
    }
  }
  return {
    verify: (attempt, signal) =>
      verificationRequest(attempt, signal, "verifyObjectBlob"),
    readVerification: (attempt, jobId, signal) =>
      verificationRequest(attempt, signal, "getJob", jobId),
    cancelVerification: (attempt, jobId, clientTxnId, signal) =>
      verificationRequest(attempt, signal, "cancelJob", jobId, clientTxnId),
    capture(input) {
      const operationID =
        input.stage === "slot"
          ? "createObjectBlobSlot"
          : input.stage === "verify"
            ? "verifyObjectBlob"
            : input.stage === "attach"
              ? "attachBlobToEvidenceRecord"
              : "createViewRow";
      const body =
        input.stage === "slot"
          ? {
              incident_id: input.authority.incidentId,
              client_txn_id: input.clientTxnId,
              byte_size: input.file?.size,
              filename_hint: input.file?.name || null,
              content_type_hint: input.file?.type || null,
            }
          : input.stage === "verify"
            ? { client_txn_id: input.clientTxnId }
            : input.stage === "attach"
              ? {
                  client_txn_id: input.clientTxnId,
                  object_blob_id: input.objectBlobId,
                  base_row_version: input.baseRowVersion,
                }
              : {
                  ...input.fields,
                  client_txn_id: input.clientTxnId,
                  "evidence.initial_object_blob_id": input.objectBlobId,
                };
      return freezeWorkbookValue({
        stage: input.stage,
        authority: { ...input.authority },
        clientTxnId: input.clientTxnId,
        apiBase,
        body: JSON.stringify(body),
        recordId: input.recordId ?? null,
        baseRowVersion: input.baseRowVersion ?? null,
        objectBlobId: input.objectBlobId ?? null,
        path: buildHTTPOperationPath(
          operationID,
          input.stage === "slot"
            ? undefined
            : input.stage === "verify"
              ? { object_blob_id: input.objectBlobId ?? "" }
              : input.stage === "attach"
                ? { record_id: input.recordId ?? "" }
                : {
                    incident_id: input.authority.incidentId,
                    view_schema_id: evidenceViewSchemaId,
                  },
        ),
      });
    },
    async slot(attempt, signal) {
      let status: number | null = null;
      let requestId: string | null = null;
      try {
        if (
          attempt.stage !== "slot" ||
          attempt.path !== buildHTTPOperationPath("createObjectBlobSlot")
        )
          return { kind: "uncertain" };
        const result = await fetchHTTPOperation<CreateObjectBlobSlotResponse>({
          apiBase: attempt.apiBase,
          operationID: "createObjectBlobSlot",
          init: { method: "POST", body: attempt.body, signal },
          onResponse: (response) => {
            status = response.status;
            requestId = response.headers.get("X-Request-ID");
          },
        });
        if (!result.ok)
          return rejectedOrUncertain(
            status,
            requestId,
            result.payload,
            "createObjectBlobSlot",
          );
        const receipt = result.payload,
          data = receipt.data,
          request = JSON.parse(attempt.body);
        // Hints are advisory, schema-validated server normalization. Browser
        // trimming is not the accepted contract (for example Unicode whitespace).
        const accepted = data.accepted_contract;
        if (
          (status !== 200 && status !== 201) ||
          !validRequestId(receipt.meta?.request_id, requestId) ||
          request.client_txn_id !== attempt.clientTxnId ||
          request.incident_id !== attempt.authority.incidentId ||
          data.incident_id !== attempt.authority.incidentId ||
          accepted.incident_id !== data.incident_id ||
          accepted.byte_size !== request.byte_size ||
          accepted.sha256_hex !== (request.sha256_hex ?? null) ||
          !Number.isFinite(Date.parse(data.target_expires_at)) ||
          !Number.isFinite(Date.parse(data.pending_expires_at)) ||
          Date.parse(data.pending_expires_at) <=
            Date.parse(data.target_expires_at) ||
          data.upload_target.expires_at !== data.target_expires_at ||
          !validEvidenceObjectUploadTarget(attempt.apiBase, data.upload_target)
        )
          return { kind: "uncertain" };
        return {
          kind: "accepted",
          receipt: freezeWorkbookValue(structuredClone(receipt)),
        };
      } catch {
        return { kind: "uncertain" };
      }
    },
    async finalize(attempt, signal) {
      if (attempt.stage === "create") {
        const result = await sendWorkbookRecordMutation(
          {
            apiBase: attempt.apiBase,
            path: attempt.path,
            body: attempt.body,
            clientTxnId: attempt.clientTxnId,
            operationID: "createViewRow",
            viewSchemaId: evidenceViewSchemaId,
            pathParameters: {
              incident_id: attempt.authority.incidentId,
              view_schema_id: evidenceViewSchemaId,
            },
          },
          signal,
          readScope?.() ?? null,
        );
        if (
          result.kind === "accepted" &&
          result.receipt.data.row.cells["evidence.storage_ref"]?.value !==
            `object://${attempt.objectBlobId}`
        )
          return { kind: "uncertain" };
        return result;
      }
      let status: number | null = null,
        requestId: string | null = null;
      try {
        if (
          attempt.stage !== "attach" ||
          !attempt.recordId ||
          !attempt.baseRowVersion ||
          attempt.path !==
            buildHTTPOperationPath("attachBlobToEvidenceRecord", {
              record_id: attempt.recordId,
            })
        )
          return { kind: "uncertain" };
        const result =
          await fetchHTTPOperation<AttachBlobToEvidenceRecordResponse>({
            apiBase: attempt.apiBase,
            operationID: "attachBlobToEvidenceRecord",
            pathParameters: { record_id: attempt.recordId },
            init: { method: "POST", body: attempt.body, signal },
            onResponse: (response) => {
              status = response.status;
              requestId = response.headers.get("X-Request-ID");
            },
          });
        if (!result.ok)
          return rejectedOrUncertain(
            status,
            requestId,
            result.payload,
            "attachBlobToEvidenceRecord",
          );
        const receipt = result.payload;
        if (
          status !== 200 ||
          !validRequestId(receipt.meta?.request_id, requestId) ||
          receipt.data.object_blob_id !== attempt.objectBlobId ||
          !acceptedRecordMutation(
            receipt.data,
            evidenceViewSchemaId,
            attempt.recordId,
          ) ||
          receipt.data.row.row_version <= attempt.baseRowVersion
        )
          return { kind: "uncertain" };
        return {
          kind: "accepted",
          receipt: freezeWorkbookValue(structuredClone(receipt)),
        };
      } catch {
        return { kind: "uncertain" };
      }
    },
    async transfer(attempt, file, signal) {
      const result = await uploadEvidenceObjectBlobTarget(
        apiBase,
        attempt.target,
        file,
        signal,
      );
      if (result.kind === "accepted") return result;
      if (
        result.failure.cause === "csrf_missing" ||
        result.failure.cause === "invalid_target"
      )
        return { kind: "not_dispatched" };
      if (result.failure.cause === "http" && result.failure.status === 401)
        return { kind: "authentication_required" };
      return { kind: "uncertain" };
    },
  };
}

function validRequestId(actual: unknown, header: string | null) {
  return (
    typeof actual === "string" &&
    !!actual.trim() &&
    (header === null || header === actual)
  );
}
function rejectedOrUncertain(
  status: number | null,
  requestId: string | null,
  payload: Parameters<typeof classifyWorkbookOperationFailure>[1],
  operation:
    | "createObjectBlobSlot"
    | "attachBlobToEvidenceRecord"
    | "verifyObjectBlob"
    | "getJob"
    | "cancelJob",
) {
  const error =
    payload && typeof payload === "object" && "error" in payload
      ? payload.error
      : null;
  if (
    status === null ||
    status < 400 ||
    status >= 500 ||
    !error ||
    typeof error !== "object" ||
    !("code" in error) ||
    typeof error.code !== "string" ||
    !error.code ||
    (requestId !== null &&
      (!("request_id" in error) || requestId !== error.request_id))
  )
    return { kind: "uncertain" as const };
  const failure = classifyWorkbookOperationFailure(status, payload, operation);
  return failure.kind === "invalid_contract"
    ? { kind: "uncertain" as const }
    : { kind: "rejected" as const, failure };
}
