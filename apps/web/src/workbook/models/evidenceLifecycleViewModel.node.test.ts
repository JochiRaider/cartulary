import { describe, expect, it } from "vitest";
import {
  buildEvidenceCountDisplayViewModel,
  buildEvidenceLifecycleViewModel,
} from "./evidenceLifecycleViewModel";

describe("evidence lifecycle view models", () => {
  it("keeps owner-valid lifecycle and upload combinations separate", () => {
    for (const recordState of [
      "requested",
      "pending_receipt",
      "received",
    ] as const) {
      for (const blobState of [
        null,
        "pending",
        "available",
        "failed",
        "quarantined",
      ] as const) {
        expect(
          buildEvidenceLifecycleViewModel({
            evidenceLifecycleState: recordState,
            objectBlobUploadState: blobState,
          }),
        ).toMatchObject({
          recordState,
          blobState,
          consistency: "consistent",
          accessEligible: false,
        });
      }
    }
    for (const recordState of ["available", "released"] as const) {
      expect(
        buildEvidenceLifecycleViewModel({
          evidenceLifecycleState: recordState,
          objectBlobUploadState: "available",
        }),
      ).toMatchObject({
        recordState,
        blobState: "available",
        uploadOverlay: "none",
        consistency: "consistent",
        accessEligible: true,
      });
    }
    for (const blobState of [null, "quarantined"] as const) {
      expect(
        buildEvidenceLifecycleViewModel({
          evidenceLifecycleState: "quarantined",
          objectBlobUploadState: blobState,
        }),
      ).toMatchObject({
        recordState: "quarantined",
        consistency: "consistent",
        accessEligible: false,
      });
    }
    expect(
      buildEvidenceLifecycleViewModel({
        evidenceLifecycleState: "requested",
        objectBlobUploadState: "pending",
      }).uploadOverlay,
    ).toBe("pending_upload_slot");
    expect(
      buildEvidenceLifecycleViewModel({
        evidenceLifecycleState: "received",
        objectBlobUploadState: "failed",
      }).uploadOverlay,
    ).toBe("failed_upload_slot");
    for (const evidenceLifecycleState of [
      "requested",
      "pending_receipt",
      "quarantined",
    ]) {
      expect(
        buildEvidenceLifecycleViewModel({
          evidenceLifecycleState,
          objectBlobUploadState: "pending",
          uploadStateSource: "evidence_projection",
        }),
      ).toMatchObject({
        recordState: evidenceLifecycleState,
        consistency: "unverified",
        uploadOverlay: "none",
        accessEligible: false,
      });
    }
  });

  it("fails inconsistent and unknown owner states closed without inventing lifecycle values", () => {
    for (const recordState of ["available", "released"] as const) {
      for (const blobState of [null, "pending", "failed", "quarantined"]) {
        expect(
          buildEvidenceLifecycleViewModel({
            evidenceLifecycleState: recordState,
            objectBlobUploadState: blobState,
          }),
        ).toMatchObject({
          recordState,
          consistency: "inconsistent",
          accessEligible: false,
        });
      }
    }
    for (const blobState of ["available", "pending", "failed"]) {
      expect(
        buildEvidenceLifecycleViewModel({
          evidenceLifecycleState: "quarantined",
          objectBlobUploadState: blobState,
        }),
      ).toMatchObject({
        recordState: "quarantined",
        consistency: "inconsistent",
        accessEligible: false,
      });
    }
    for (const value of ["future", "inconsistent", 12, {}, null]) {
      expect(
        buildEvidenceLifecycleViewModel({
          evidenceLifecycleState: value,
          objectBlobUploadState: "available",
        }),
      ).toMatchObject({
        recordState: null,
        consistency: "inconsistent",
        accessEligible: false,
      });
    }
    expect(
      buildEvidenceLifecycleViewModel({
        evidenceLifecycleState: "requested",
        objectBlobUploadState: "future",
      }),
    ).toMatchObject({ consistency: "inconsistent", accessEligible: false });
  });

  it("uses projected record counts independently of access and pending upload slots", () => {
    expect(
      buildEvidenceCountDisplayViewModel({
        projectedCount: 3,
        projectedHasEvidence: true,
      }),
    ).toEqual({
      displayCount: "3",
      projectedCount: 3,
      hasEvidence: true,
      stateKey: "available",
    });
    expect(
      buildEvidenceCountDisplayViewModel({
        projectedCount: 0,
        projectedHasEvidence: false,
      }),
    ).toEqual({
      displayCount: "0",
      projectedCount: 0,
      hasEvidence: false,
      stateKey: "empty",
    });
  });

  it("derives count-display states and detects inconsistent projections", () => {
    expect(
      buildEvidenceCountDisplayViewModel({
        projectedCount: null,
        projectedHasEvidence: false,
      }),
    ).toMatchObject({ displayCount: "0", stateKey: "empty" });
    expect(
      buildEvidenceCountDisplayViewModel({
        projectedCount: "2",
        projectedHasEvidence: true,
      }),
    ).toMatchObject({ displayCount: "2", stateKey: "available" });
    for (const input of [
      { projectedCount: 0, projectedHasEvidence: true },
      { projectedCount: "2", projectedHasEvidence: false },
      { projectedCount: -1 },
      { projectedCount: "invalid" },
      { projectedHasEvidence: "invalid" },
    ])
      expect(buildEvidenceCountDisplayViewModel(input).stateKey).toBe(
        "inconsistent",
      );
  });
});
