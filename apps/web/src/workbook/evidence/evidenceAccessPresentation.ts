import type { EvidenceLifecycleViewModel } from "../models/evidenceLifecycleViewModel";

export type EvidenceAccessLiveRegion = {
  readonly ariaLive: "assertive" | "polite";
  readonly role: "alert" | "status";
};

export function evidenceAccessMessageLiveRegion(
  message: string,
  evidenceAccess: EvidenceLifecycleViewModel,
): EvidenceAccessLiveRegion {
  const normalized = message.toLowerCase();
  const isBlockingMessage =
    evidenceAccess.messageTone === "danger" ||
    evidenceAccess.stateKey === "blocked" ||
    evidenceAccess.stateKey === "failed" ||
    evidenceAccess.stateKey === "inconsistent" ||
    evidenceAccess.stateKey === "preview_blocked" ||
    evidenceAccess.stateKey === "public_error" ||
    /\b(?:blob_failed|evidence_access_unavailable|failed|inconsistent|quarantined|unavailable|unsupported_preview)\b/u.test(
      normalized,
    );
  return isBlockingMessage
    ? { ariaLive: "assertive", role: "alert" }
    : { ariaLive: "polite", role: "status" };
}
