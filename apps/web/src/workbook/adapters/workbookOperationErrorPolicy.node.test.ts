import { describe, expect, it } from "vitest";
import { classifyWorkbookOperationFailure } from "./workbookOperationErrorPolicy";

function publicError(
  code: string,
  status: number,
  options: {
    readonly conflict?: unknown;
    readonly details?: Readonly<Record<string, unknown>>;
    readonly retryable?: boolean;
  } = {},
) {
  return {
    error: {
      code,
      details: options.details ?? {},
      message: `${code} message`,
      request_id: "request-1",
      retryable: options.retryable ?? false,
      status,
      ...(options.conflict === undefined ? {} : { conflict: options.conflict }),
    },
  };
}

describe("Workbook operation error policy", () => {
  it("accepts only canonical transient Evidence handle error identity", () => {
    const operation = "issueEvidencePreviewHandle";
    expect(
      classifyWorkbookOperationFailure(
        503,
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
    for (const [status, payload] of [
      [
        503,
        publicError("object_store_unavailable", 503, {
          details: { reason_code: "endpoint_unreachable" },
        }),
      ],
      [
        502,
        publicError("object_store_unavailable", 503, {
          retryable: true,
          details: { reason_code: "endpoint_unreachable" },
        }),
      ],
      [
        503,
        publicError("object_store_unavailable", 503, {
          retryable: true,
          details: { reason_code: "future_reason" },
        }),
      ],
      [
        503,
        publicError("object_store_access_rejected", 503, {
          details: { reason_code: "credential_denied" },
        }),
      ],
      [503, publicError("future_retryable_error", 503, { retryable: true })],
    ] as const) {
      expect(
        classifyWorkbookOperationFailure(status, payload, operation).kind,
      ).not.toBe("retryable");
    }
  });

  it("fails malformed success, error, and conflict envelopes closed", () => {
    expect(
      classifyWorkbookOperationFailure(
        500,
        { error: { code: "invalid_public_contract_response" } },
        "queryWorkbookView",
      ),
    ).toMatchObject({
      kind: "invalid_contract",
      message: "The server returned an invalid public contract response.",
      presentation: { family: "initial_load_failure" },
    });
    expect(
      classifyWorkbookOperationFailure(
        500,
        { error: { message: "untyped" } },
        "patchRecord",
      ),
    ).toMatchObject({
      kind: "invalid_contract",
      message: "The server returned an invalid public error response.",
    });
    expect(
      classifyWorkbookOperationFailure(
        409,
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
      classifyWorkbookOperationFailure(
        409,
        publicError("same_field_conflict", 409, { conflict }),
        "patchRecord",
      ),
    ).toMatchObject({ kind: "same_field_conflict", conflict });
    expect(
      classifyWorkbookOperationFailure(
        401,
        publicError("session_required", 401),
        "queryWorkbookView",
      ),
    ).toMatchObject({ kind: "authentication_required" });
    expect(
      classifyWorkbookOperationFailure(
        409,
        publicError("authorization_denied", 409),
        "patchRecord",
      ),
    ).toMatchObject({ kind: "authorization_lost" });
    expect(
      classifyWorkbookOperationFailure(
        409,
        publicError("row_version_conflict", 409),
        "patchRecord",
      ),
    ).toMatchObject({ kind: "stale_target" });
  });

  it("classifies validation, merge preconditions, and retryability exactly", () => {
    const blockingRecordId = "20000000-0000-4000-8000-000000000002";
    expect(
      classifyWorkbookOperationFailure(
        400,
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
      classifyWorkbookOperationFailure(
        409,
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
      classifyWorkbookOperationFailure(
        409,
        publicError("future_retryable_error", 409, { retryable: true }),
        "patchRecord",
      ),
    ).toEqual({
      kind: "retryable",
      message: "future_retryable_error message",
      publicCode: "future_retryable_error",
      presentation: expect.any(Object),
    });
    expect(
      classifyWorkbookOperationFailure(
        503,
        publicError("future_terminal_error", 503),
        "patchRecord",
      ),
    ).toEqual({
      kind: "retryable",
      message: "future_terminal_error message",
      publicCode: "future_terminal_error",
      presentation: expect.any(Object),
    });
  });

  it("projects Evidence access reasons without exposing server prose", () => {
    expect(
      classifyWorkbookOperationFailure(
        409,
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
