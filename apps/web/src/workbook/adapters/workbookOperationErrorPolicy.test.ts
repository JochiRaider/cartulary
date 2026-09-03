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
    expect(
      classifyWorkbookOperationFailure(
        422,
        publicError("invalid_mutation_payload", 422, {
          details: { field: "task.title", reason_code: "required" },
        }),
        "patchRecord",
      ),
    ).toMatchObject({
      kind: "validation",
      fields: [{ field: "task.title", message: "required" }],
    });
    expect(
      classifyWorkbookOperationFailure(
        409,
        publicError("merge_precondition_failed", 409, {
          details: {
            reason_code: "identifier_collision",
            blocking_record_id: "host-2",
          },
        }),
        "mergeEntityRecord",
      ),
    ).toMatchObject({
      kind: "validation",
      fields: expect.arrayContaining([
        { field: "reason_code", message: "identifier_collision" },
        { field: "blocking_record_id", message: "host-2" },
      ]),
    });
    expect(
      classifyWorkbookOperationFailure(
        409,
        publicError("future_retryable_error", 409, { retryable: true }),
        "patchRecord",
      ),
    ).toMatchObject({ kind: "retryable" });
    expect(
      classifyWorkbookOperationFailure(
        503,
        publicError("future_terminal_error", 503),
        "patchRecord",
      ),
    ).toMatchObject({ kind: "retryable" });
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
      message: "evidence_access_unavailable: unsupported_preview",
      presentation: { family: "evidence_preview_blocked" },
    });
  });
});
