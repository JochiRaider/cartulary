import type { HTTPOperationErrorCode } from "@cartulary/protocol-ts/http";
import { describe, expect, it } from "vitest";
import type {
  HTTPOperationError,
  HTTPOperationPublicError,
} from "../../services/browserApi";
import { operationContractFailure } from "../../testing/fetchMockTestSupport";
import {
  isWorkbookTemporaryUnavailability,
  workbookFailureLifecycle,
} from "../ports/WorkbookPortResult";
import type { WorkbookOperationID } from "./workbookOperationContract";
import {
  classifyWorkbookOperationFailure,
  explicitWorkbookFailureOutcome,
  explicitWorkbookWriteUncertain,
} from "./workbookOperationErrorPolicy";

/** A decoded public error as fetchHTTPOperation surfaces it to the classifier. */
function publicError(
  code: HTTPOperationErrorCode<WorkbookOperationID>,
  status: number,
  options: {
    readonly conflict?: unknown;
    readonly details?: Readonly<Record<string, unknown>>;
    readonly retryable?: boolean;
  } = {},
) {
  return {
    code,
    details: options.details ?? {},
    message: `${code} message`,
    request_id: "request-1",
    retryable: options.retryable ?? false,
    status,
    ...(options.conflict === undefined ? {} : { conflict: options.conflict }),
  } as HTTPOperationPublicError<WorkbookOperationID>;
}
const contractFailure = (operation: WorkbookOperationID, status: number) =>
  operationContractFailure(operation, status).payload.error;
const classify = (
  error: HTTPOperationError<WorkbookOperationID>,
  operation: WorkbookOperationID,
) => classifyWorkbookOperationFailure<WorkbookOperationID>(error, operation);
const settleWrite = (
  status: number | null,
  error: HTTPOperationError<WorkbookOperationID>,
  requestId: string | null = null,
) =>
  explicitWorkbookFailureOutcome<WorkbookOperationID>(
    status,
    error,
    "patchRecord",
    requestId,
  );

describe("Workbook operation error policy", () => {
  it("accepts only canonical transient Evidence handle error identity", () => {
    const operation = "issueEvidencePreviewHandle";
    expect(
      classify(
        publicError("object_store_unavailable", 503, {
          retryable: true,
          details: { reason_code: "endpoint_unreachable" },
        }),
        operation,
      ),
    ).toMatchObject({
      kind: "retryable",
      publicCode: "object_store_unavailable",
      publicReason: "endpoint_unreachable",
    });
    // A status mismatch never reaches the classifier: the central decoder
    // turns it into the client contract failure.
    for (const error of [
      publicError("object_store_unavailable", 503, {
        details: { reason_code: "endpoint_unreachable" },
      }),
      contractFailure(operation, 502),
      publicError("object_store_unavailable", 503, {
        retryable: true,
        details: { reason_code: "future_reason" },
      }),
      publicError("object_store_access_rejected", 503, {
        details: { reason_code: "credential_denied" },
      }),
      publicError("internal_error", 500, { retryable: true }),
    ]) {
      expect(classify(error, operation).kind).not.toBe("retryable");
    }
  });

  it("classifies a declared service_unavailable as retryable temporary unavailability before operation rules", () => {
    for (const [operation, family] of [
      ["patchRecord", "stale_refresh"],
      ["createViewRow", "stale_refresh"],
      ["applyWorkbookBulkMutation", "stale_refresh"],
      ["queryWorkbookView", "initial_load_failure"],
      ["getIncidentWorkbookStartup", "initial_load_failure"],
      ["issueEvidencePreviewHandle", "stale_refresh"],
      ["issueEvidenceDownloadHandle", "stale_refresh"],
    ] as const) {
      const failure = classify(
        publicError("service_unavailable", 503, {
          retryable: true,
          details: { reason_code: "extension_publication_pending" },
        }),
        operation,
      );
      expect(failure).toMatchObject({
        kind: "retryable",
        message: "Cartulary is temporarily unavailable. Try again shortly.",
        publicCode: "service_unavailable",
        publicReason: "extension_publication_pending",
        presentation: { family },
      });
      expect(isWorkbookTemporaryUnavailability(failure)).toBe(true);
      expect(workbookFailureLifecycle(failure).kind).toBe(
        "temporarily_unavailable",
      );
      expect(JSON.stringify(failure.message)).not.toContain(
        "extension_publication_pending",
      );
    }
    // The declared 500 is not temporary unavailability.
    const internal = classify(
      publicError("internal_error", 500),
      "patchRecord",
    );
    expect(isWorkbookTemporaryUnavailability(internal)).toBe(false);
    expect(workbookFailureLifecycle(internal).kind).toBe("operation_rejected");
  });

  it("rejects a decoded 503 explicit write as not applied and keeps other server failures uncertain", () => {
    // Core 03 REQ-03-086a: a decoded 503 service_unavailable was not applied,
    // so it is a definite, retryable rejection that keeps the draft.
    const refused = settleWrite(
      503,
      publicError("service_unavailable", 503, {
        details: { reason_code: "admission_paused" },
        retryable: true,
      }),
      "request-1",
    );
    expect(refused).toEqual({
      kind: "rejected",
      failure: expect.objectContaining({
        kind: "retryable",
        message: "Cartulary is temporarily unavailable. Try again shortly.",
        publicCode: "service_unavailable",
      }),
    });
    for (const [status, error, requestId] of [
      [500, publicError("internal_error", 500), null],
      [503, contractFailure("patchRecord", 503), null],
      [503, publicError("service_unavailable", 503), "other-request"],
      [null, publicError("service_unavailable", 503), null],
    ] as const)
      expect(settleWrite(status, error, requestId)).toEqual({
        kind: "uncertain",
      });
    expect(settleWrite(409, publicError("row_version_conflict", 409))).toEqual({
      kind: "rejected",
      failure: expect.objectContaining({ kind: "stale_target" }),
    });
    expect(
      explicitWorkbookWriteUncertain(503, {
        kind: "retryable",
        publicCode: "service_unavailable",
      }),
    ).toBe(false);
    expect(
      explicitWorkbookWriteUncertain(500, {
        kind: "retryable",
        publicCode: "internal_error",
      }),
    ).toBe(true);
  });

  it("fails malformed success, error, and conflict envelopes closed", () => {
    expect(
      classify(contractFailure("queryWorkbookView", 500), "queryWorkbookView"),
    ).toMatchObject({
      kind: "invalid_contract",
      message: "The server returned an invalid public contract response.",
      presentation: { family: "initial_load_failure" },
    });
    const malformed = classify(
      contractFailure("patchRecord", 500),
      "patchRecord",
    );
    expect(malformed).toMatchObject({
      kind: "invalid_contract",
      message: "The server returned an invalid public contract response.",
    });
    expect(malformed).not.toHaveProperty("publicCode");
    expect(
      classify(
        publicError("same_field_conflict", 409, { conflict: {} }),
        "patchRecord",
      ),
    ).toMatchObject({
      kind: "invalid_contract",
      message: "The server returned an invalid conflict response.",
      publicCode: "same_field_conflict",
    });
  });

  it("classifies exact conflicts, authentication, authorization, and stale targets", () => {
    const conflict = {
      base_row_version: 3,
      base_value: "before",
      client_value: "local",
      conflict_resolution_class: "text_compare_merge",
      conflict_token: "conflict-1",
      current_row_version: 4,
      field_key: "task.title",
      record_id: "task-1",
      server_value: "server",
    };
    expect(
      classify(
        publicError("same_field_conflict", 409, { conflict }),
        "patchRecord",
      ),
    ).toMatchObject({ kind: "same_field_conflict", conflict });
    expect(
      classify(publicError("session_required", 401), "queryWorkbookView"),
    ).toMatchObject({ kind: "authentication_required" });
    expect(
      classify(publicError("authorization_denied", 403), "patchRecord"),
    ).toMatchObject({ kind: "authorization_lost" });
    expect(
      classify(publicError("row_version_conflict", 409), "patchRecord"),
    ).toMatchObject({ kind: "stale_target" });
  });

  it("classifies validation, merge preconditions, and retryability exactly", () => {
    const blockingRecordId = "20000000-0000-4000-8000-000000000002";
    expect(
      classify(
        publicError("invalid_mutation_payload", 400, {
          details: {
            field: "task.title",
            reason_code: "missing_required_field",
          },
        }),
        "patchRecord",
      ),
    ).toEqual({
      kind: "validation",
      message: "invalid_mutation_payload",
      fields: [{ field: "task.title", message: "missing_required_field" }],
      publicCode: "invalid_mutation_payload",
      publicReason: "missing_required_field",
      presentation: expect.any(Object),
    });
    expect(
      classify(
        publicError("merge_precondition_failed", 409, {
          details: {
            reason_code: "carry_forward_identifier_collision",
            blocking_record_id: blockingRecordId,
          },
        }),
        "mergeEntityRecord",
      ),
    ).toEqual({
      kind: "validation",
      message:
        "merge_precondition_failed message: carry_forward_identifier_collision",
      fields: [
        { field: "reason_code", message: "carry_forward_identifier_collision" },
        { field: "blocking_record_id", message: blockingRecordId },
      ],
      publicCode: "merge_precondition_failed",
      publicReason: "carry_forward_identifier_collision",
      presentation: expect.any(Object),
    });
    expect(
      classify(
        publicError("credential_bootstrap_rejected", 409, { retryable: true }),
        "patchRecord",
      ),
    ).toEqual({
      kind: "retryable",
      message: "credential_bootstrap_rejected message",
      publicCode: "credential_bootstrap_rejected",
      presentation: expect.any(Object),
    });
    expect(
      classify(publicError("internal_error", 500), "patchRecord"),
    ).toMatchObject({ kind: "terminal", publicCode: "internal_error" });
  });

  it("projects Evidence access reasons without exposing server prose", () => {
    expect(
      classify(
        publicError("evidence_access_unavailable", 409, {
          details: { reason_code: "unsupported_preview" },
        }),
        "issueEvidencePreviewHandle",
      ),
    ).toMatchObject({
      kind: "terminal",
      message: "Evidence request failed.",
      publicReason: "unsupported_preview",
      presentation: { family: "evidence_preview_blocked" },
    });
  });
});
