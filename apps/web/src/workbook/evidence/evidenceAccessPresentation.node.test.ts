import { describe, expect, it } from "vitest";
import { classifyWorkbookOperationFailure } from "../adapters/workbookOperationErrorPolicy";
import { buildEvidenceLifecycleViewModel } from "../models/evidenceLifecycleViewModel";
import {
  buildEvidenceAccessPresentation,
  evidenceOperationFeedback,
} from "./evidenceAccessPresentation";

const lifecycle = buildEvidenceLifecycleViewModel({
  evidenceLifecycleState: "available",
  objectBlobUploadState: "available",
});
function failure(
  reason: unknown,
  operation: "preview" | "download" = "preview",
  code = "evidence_access_unavailable",
  message = "Untrusted private token",
) {
  return classifyWorkbookOperationFailure(
    409,
    {
      error: {
        code,
        status: 409,
        retryable: false,
        request_id: "request-1",
        message,
        details: { reason_code: reason },
      },
    },
    operation === "preview"
      ? "issueEvidencePreviewHandle"
      : "issueEvidenceDownloadHandle",
  );
}

describe("evidenceAccessPresentation", () => {
  it("keeps transient recovery separate from lifecycle and action blockers", () => {
    const transient = classifyWorkbookOperationFailure(
      503,
      {
        error: {
          code: "object_store_unavailable",
          status: 503,
          retryable: true,
          request_id: "request-2",
          message: "private",
          details: { reason_code: "endpoint_unreachable" },
        },
      },
      "issueEvidencePreviewHandle",
    );
    const retry = buildEvidenceAccessPresentation(lifecycle, {
      kind: "rejected",
      operation: "preview",
      failure: transient,
    });
    expect(retry).toMatchObject({
      stateKey: "available",
      canPreview: true,
      canDownload: true,
      retryKind: "preview",
      busyKind: null,
    });
    expect(retry.message).toContain("Retry Preview");
    expect(retry.message).not.toContain("private");
    expect(
      buildEvidenceAccessPresentation(lifecycle, {
        kind: "deadline",
        operation: "download",
      }),
    ).toMatchObject({
      stateKey: "available",
      canPreview: true,
      canDownload: true,
      retryKind: "download",
    });
    expect(
      buildEvidenceAccessPresentation(lifecycle, {
        kind: "rejected",
        operation: "preview",
        failure: failure("unsupported_preview"),
      }),
    ).toMatchObject({ canPreview: false, canDownload: true, retryKind: null });
    expect(
      buildEvidenceAccessPresentation(lifecycle, {
        kind: "rejected",
        operation: "preview",
        failure: failure("evidence_quarantined"),
      }),
    ).toMatchObject({ canPreview: false, canDownload: false, retryKind: null });
    const unknown = classifyWorkbookOperationFailure(
      503,
      {
        error: {
          code: "future_error",
          status: 503,
          retryable: true,
          request_id: "request-3",
          message: "private",
          details: {},
        },
      },
      "issueEvidencePreviewHandle",
    );
    expect(
      buildEvidenceAccessPresentation(lifecycle, {
        kind: "rejected",
        operation: "preview",
        failure: unknown,
      }),
    ).toMatchObject({ canPreview: false, canDownload: true, retryKind: null });
  });

  it("derives polite preview blockers and distinct operation announcements from typed outcomes", () => {
    for (const reason of [
      "unsupported_preview",
      "preview_payload_too_large",
      "blob_failed",
      "evidence_inconsistent",
      "evidence_quarantined",
      "blob_missing",
      "blob_pending",
      "no_visible_blob",
    ]) {
      const access = buildEvidenceAccessPresentation(lifecycle, {
        kind: "rejected",
        operation: "preview",
        failure: failure(reason),
      });
      expect(access.announcement).toBe("polite");
      expect(access.canPreview).toBe(false);
      expect(access.canDownload).toBe(
        reason === "unsupported_preview" ||
          reason === "preview_payload_too_large",
      );
      expect(access.message).not.toContain(reason);
    }
    expect(
      evidenceOperationFeedback({
        kind: "rejected",
        operation: "download",
        failure: failure("blob_failed", "download"),
      }).announcement,
    ).toBe("assertive");
    expect(
      evidenceOperationFeedback({
        kind: "rejected",
        operation: "attach",
        failure: {
          kind: "terminal",
          message: "anything",
          uploadFailure: { cause: "network" },
        },
      }),
    ).toMatchObject({
      announcement: "assertive",
      message: "The evidence file could not be uploaded.",
    });
    expect(
      evidenceOperationFeedback({ kind: "accepted", operation: "preview" })
        .message,
    ).toBe("Preview opened.");
    expect(
      evidenceOperationFeedback({ kind: "accepted", operation: "download" })
        .message,
    ).toBe("Download requested.");
    expect(buildEvidenceAccessPresentation(lifecycle).announcement).toBe(
      "none",
    );
  });

  it("ignores arbitrary public prose and unknown reasons", () => {
    for (const message of [
      "unsupported_preview",
      "blob_failed",
      "private_implementation_id",
      "s3://private/key",
      "SELECT secret FROM storage",
      "",
    ]) {
      const unknown = failure(
        "future_reason",
        "preview",
        "future_code",
        message,
      );
      expect(unknown.publicReason).toBeUndefined();
      expect(unknown.presentation).toMatchObject({
        family: "unknown_future_error",
        actions: [],
      });
      expect(
        evidenceOperationFeedback({
          kind: "rejected",
          operation: "preview",
          failure: unknown,
        }),
      ).toMatchObject({
        message: "The preview could not be opened.",
        announcement: "assertive",
      });
      expect(
        buildEvidenceAccessPresentation(lifecycle, {
          kind: "rejected",
          operation: "preview",
          failure: failure(
            "unsupported_preview",
            "preview",
            "evidence_access_unavailable",
            message,
          ),
        }).message,
      ).toBe("This file type cannot be previewed.");
    }
    expect(
      failure("unsupported_preview", "preview", "object_store_unavailable")
        .publicReason,
    ).toBeUndefined();
    expect(
      buildEvidenceAccessPresentation(
        buildEvidenceLifecycleViewModel({
          evidenceLifecycleState: "requested",
        }),
        {
          kind: "rejected",
          operation: "preview",
          failure: failure("unsupported_preview"),
        },
      ).canDownload,
    ).toBe(false);
  });
});
