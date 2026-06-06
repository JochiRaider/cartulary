import { describe, expect, it } from "vitest";

import {
  buildEvidenceCountDisplayViewModel,
  buildEvidenceLifecycleViewModel,
  summarizeEvidenceLifecycleCounts,
} from "./evidenceLifecycleViewModel";

describe("FE-U-P6-01 evidence lifecycle view models", () => {
  it("FE-U-P6-01 distinguishes requested, pending upload, available, and preview-blocked states", () => {
    const requested = buildEvidenceLifecycleViewModel({
      evidenceLifecycleState: "requested",
      objectBlobUploadState: "pending",
    });
    expect(requested).toMatchObject({
      recordState: "requested",
      blobState: "pending",
      stateKey: "requested",
      canPreview: false,
      canDownload: false,
      countContribution: 0,
    });

    const pendingReceipt = buildEvidenceLifecycleViewModel({
      evidenceLifecycleState: "pending_receipt",
      objectBlobUploadState: "available",
    });
    expect(pendingReceipt).toMatchObject({
      recordState: "pending_receipt",
      blobState: "available",
      stateKey: "pending_upload",
      countContribution: 0,
    });

    const available = buildEvidenceLifecycleViewModel({
      evidenceLifecycleState: "released",
      objectBlobUploadState: "available",
    });
    expect(available).toMatchObject({
      recordState: "released",
      blobState: "available",
      stateKey: "available",
      canPreview: true,
      canDownload: true,
      countContribution: 1,
      message: null,
    });

    const previewBlocked = buildEvidenceLifecycleViewModel({
      evidenceLifecycleState: "available",
      objectBlobUploadState: "available",
      publicError: {
        code: "evidence_access_unavailable",
        message: "Preview is not available for this evidence.",
        details: { reason_code: "unsupported_preview" },
      },
    });
    expect(previewBlocked).toMatchObject({
      recordState: "available",
      blobState: "available",
      stateKey: "preview_blocked",
      canPreview: false,
      canDownload: true,
      countContribution: 1,
      messageTone: "warning",
      publicErrorCode: "evidence_access_unavailable",
      reasonCode: "unsupported_preview",
    });
  });

  it("FE-U-P6-01 keeps failed, blocked, inconsistent, and public-error rendering inputs distinct", () => {
    const failed = buildEvidenceLifecycleViewModel({
      evidenceLifecycleState: "available",
      objectBlobUploadState: "failed",
    });
    expect(failed).toMatchObject({
      stateKey: "failed",
      canPreview: false,
      canDownload: false,
      countContribution: 0,
      messageTone: "danger",
    });
    expect(failed.message).toContain("Failed:");

    const blocked = buildEvidenceLifecycleViewModel({
      evidenceLifecycleState: "quarantined",
      objectBlobUploadState: "available",
    });
    expect(blocked).toMatchObject({
      stateKey: "blocked",
      canPreview: false,
      canDownload: false,
      countContribution: 0,
      messageTone: "warning",
    });
    expect(blocked.message).toContain("Blocked:");

    const inconsistent = buildEvidenceLifecycleViewModel({
      evidenceLifecycleState: "available",
    });
    expect(inconsistent).toMatchObject({
      recordState: "available",
      blobState: null,
      stateKey: "inconsistent",
      canPreview: false,
      canDownload: false,
      countContribution: 0,
      messageTone: "danger",
    });
    expect(inconsistent.message).toContain("Inconsistent:");

    const publicError = buildEvidenceLifecycleViewModel({
      evidenceLifecycleState: "available",
      objectBlobUploadState: "available",
      publicError: {
        code: "object_store_unavailable",
        message: "Evidence storage is unavailable.",
        details: { reason_code: "dependency_unavailable" },
      },
    });
    expect(publicError).toMatchObject({
      stateKey: "public_error",
      canPreview: false,
      canDownload: false,
      countContribution: 0,
      message: "Evidence storage is unavailable.",
      messageTone: "danger",
      publicErrorCode: "object_store_unavailable",
      reasonCode: "dependency_unavailable",
    });

    const publicInconsistent = buildEvidenceLifecycleViewModel({
      evidenceLifecycleState: "available",
      objectBlobUploadState: "available",
      publicError: {
        code: "evidence_access_unavailable",
        message: "Evidence state is inconsistent.",
        details: { reason_code: "evidence_inconsistent" },
      },
    });
    expect(publicInconsistent).toMatchObject({
      stateKey: "inconsistent",
      publicErrorCode: "evidence_access_unavailable",
      reasonCode: "evidence_inconsistent",
    });
  });

  it("FE-U-P6-01 counts only available and preview-blocked evidence as attached", () => {
    const viewModels = [
      buildEvidenceLifecycleViewModel({
        evidenceLifecycleState: "available",
        objectBlobUploadState: "available",
      }),
      buildEvidenceLifecycleViewModel({
        evidenceLifecycleState: "available",
        objectBlobUploadState: "available",
        publicError: {
          code: "evidence_access_unavailable",
          message: "Preview is not available for this evidence.",
          details: { reason_code: "preview_payload_too_large" },
        },
      }),
      buildEvidenceLifecycleViewModel({
        evidenceLifecycleState: "requested",
        objectBlobUploadState: "pending",
      }),
      buildEvidenceLifecycleViewModel({
        evidenceLifecycleState: "received",
        objectBlobUploadState: "pending",
      }),
      buildEvidenceLifecycleViewModel({
        evidenceLifecycleState: "available",
        objectBlobUploadState: "failed",
      }),
      buildEvidenceLifecycleViewModel({
        evidenceLifecycleState: "quarantined",
        objectBlobUploadState: "available",
      }),
      buildEvidenceLifecycleViewModel({
        evidenceLifecycleState: "available",
      }),
      buildEvidenceLifecycleViewModel({
        evidenceLifecycleState: "available",
        objectBlobUploadState: "available",
        publicError: {
          code: "object_store_unavailable",
          message: "Evidence storage is unavailable.",
        },
      }),
    ];

    expect(summarizeEvidenceLifecycleCounts(viewModels)).toEqual({
      availableCount: 2,
      blockedCount: 1,
      countContribution: 2,
      failedCount: 1,
      inconsistentCount: 1,
      pendingUploadCount: 1,
      publicErrorCount: 1,
      requestedCount: 1,
    });
    expect(
      buildEvidenceCountDisplayViewModel({ lifecycleViewModels: viewModels }),
    ).toMatchObject({
      displayCount: "2",
      hasEvidence: true,
      projectedCount: 2,
      stateKey: "inconsistent",
    });
  });

  it("FE-U-P6-01 derives count-display states and detects inconsistent projections", () => {
    expect(
      buildEvidenceCountDisplayViewModel({
        projectedCount: null,
        projectedHasEvidence: false,
      }),
    ).toEqual({
      displayCount: "0",
      hasEvidence: false,
      projectedCount: 0,
      stateKey: "empty",
    });
    expect(
      buildEvidenceCountDisplayViewModel({
        projectedCount: "2",
        projectedHasEvidence: true,
      }),
    ).toEqual({
      displayCount: "2",
      hasEvidence: true,
      projectedCount: 2,
      stateKey: "available",
    });
    expect(
      buildEvidenceCountDisplayViewModel({
        lifecycleViewModels: [
          buildEvidenceLifecycleViewModel({
            evidenceLifecycleState: "requested",
            objectBlobUploadState: "pending",
          }),
        ],
      }),
    ).toMatchObject({
      displayCount: "0",
      hasEvidence: false,
      stateKey: "pending_upload",
    });
    expect(
      buildEvidenceCountDisplayViewModel({
        lifecycleViewModels: [
          buildEvidenceLifecycleViewModel({
            evidenceLifecycleState: "available",
            objectBlobUploadState: "failed",
          }),
        ],
      }),
    ).toMatchObject({
      displayCount: "0",
      hasEvidence: false,
      stateKey: "failed",
    });
    expect(
      buildEvidenceCountDisplayViewModel({
        projectedCount: 0,
        projectedHasEvidence: true,
      }),
    ).toMatchObject({
      stateKey: "inconsistent",
      hasEvidence: true,
    });
    expect(
      buildEvidenceCountDisplayViewModel({
        projectedCount: "2",
        projectedHasEvidence: false,
      }),
    ).toMatchObject({
      displayCount: "2",
      stateKey: "inconsistent",
      hasEvidence: true,
    });
  });
});
