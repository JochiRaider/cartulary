import { describe, expect, it } from "vitest";
import { buildEvidenceLifecycleViewModel } from "../models/evidenceLifecycleViewModel";
import { evidenceAccessMessageLiveRegion } from "./evidenceAccessPresentation";

describe("evidenceAccessPresentation", () => {
  it("maps blocking evidence messages to assertive live regions", () => {
    const evidenceAccess = buildEvidenceLifecycleViewModel({
      evidenceLifecycleState: "available",
      objectBlobUploadState: "failed",
    });

    expect(
      evidenceAccessMessageLiveRegion("Failed: object blob upload failed.", {
        ...evidenceAccess,
        messageTone: "danger",
      }),
    ).toEqual({ ariaLive: "assertive", role: "alert" });
    expect(
      evidenceAccessMessageLiveRegion("Preview loaded inline.", {
        ...evidenceAccess,
        messageTone: "success",
        stateKey: "available",
      }),
    ).toEqual({ ariaLive: "polite", role: "status" });
  });
});
