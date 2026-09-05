import { cartularyErrorPresentation } from "@cartulary/ui-contracts";
import type { EvidenceAccessReasonCode } from "../../services/publicErrorIdentity";
import type {
  EvidenceLifecycleViewModel,
  EvidenceRecordLifecycleState,
} from "../models/evidenceLifecycleViewModel";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";

export type EvidenceOperationKind = "preview" | "download" | "attach";
export type EvidenceOperationState =
  | { readonly kind: "pending"; readonly operation: EvidenceOperationKind }
  | { readonly kind: "accepted"; readonly operation: EvidenceOperationKind }
  | {
      readonly kind: "rejected";
      readonly operation: EvidenceOperationKind;
      readonly failure: WorkbookOperationFailure;
    };
export type EvidenceAccessStateKey =
  | EvidenceRecordLifecycleState
  | "inconsistent"
  | "pending_upload"
  | "failed"
  | "blocked"
  | "preview_blocked"
  | "public_error";
export type EvidenceFeedback = {
  readonly label: string;
  readonly message: string;
  readonly tone: "neutral" | "success" | "warning" | "danger";
  readonly announcement: "none" | "polite" | "assertive";
};
export type EvidenceAccessPresentation = EvidenceFeedback & {
  readonly stateKey: EvidenceAccessStateKey;
  readonly lifecycleLabel: string;
  readonly uploadLabel: string;
  readonly canPreview: boolean;
  readonly canDownload: boolean;
};

const lifecycleLabels = {
  requested: "Requested",
  pending_receipt: "Pending receipt",
  received: "Received",
  available: "Available",
  quarantined: "Quarantined",
  released: "Released",
} satisfies Record<EvidenceRecordLifecycleState, string>;

const accessBlockers = {
  no_visible_blob: {
    label: "No file",
    message: "No file is available for this evidence.",
    stateKey: "blocked",
  },
  blob_pending: {
    label: "Upload pending",
    message: "The evidence upload is not complete.",
    stateKey: "pending_upload",
  },
  blob_failed: {
    label: "Upload failed",
    message: "The evidence file upload failed.",
    stateKey: "failed",
  },
  blob_missing: {
    label: "File unavailable",
    message: "The evidence file is unavailable.",
    stateKey: "blocked",
  },
  evidence_quarantined: {
    label: "Quarantined",
    message: "This evidence is quarantined. Preview and download are blocked.",
    stateKey: "quarantined",
  },
  evidence_inconsistent: {
    label: "Inconsistent",
    message:
      "This evidence has inconsistent file information. Preview and download are blocked.",
    stateKey: "inconsistent",
  },
  unsupported_preview: {
    label: "No preview",
    message: "This file type cannot be previewed.",
    stateKey: "preview_blocked",
  },
  preview_payload_too_large: {
    label: "Preview too large",
    message: "This file is too large to preview.",
    stateKey: "preview_blocked",
  },
} as const satisfies Record<
  EvidenceAccessReasonCode,
  { label: string; message: string; stateKey: EvidenceAccessStateKey }
>;

function accessBlocker(failure: WorkbookOperationFailure) {
  if (
    failure.kind === "authentication_required" ||
    failure.kind === "authorization_lost" ||
    failure.presentation?.family === "permission_or_incident_access_loss" ||
    failure.publicCode !== "evidence_access_unavailable" ||
    failure.publicReason === undefined
  )
    return null;
  return (
    Object.entries(accessBlockers).find(
      ([reason]) => reason === failure.publicReason,
    )?.[1] ?? null
  );
}

/** All Evidence copy is authored here; no server prose or exception text is rendered. */
export function evidenceOperationFeedback(
  state: EvidenceOperationState,
): EvidenceFeedback {
  if (state.kind === "pending") {
    const message =
      state.operation === "preview"
        ? "Opening preview…"
        : state.operation === "download"
          ? "Requesting download…"
          : "Uploading evidence…";
    return {
      label: state.operation === "attach" ? "Uploading" : "Pending",
      message,
      tone: "neutral",
      announcement: "polite",
    };
  }
  if (state.kind === "accepted") {
    const message =
      state.operation === "preview"
        ? "Preview opened."
        : state.operation === "download"
          ? "Download requested."
          : "Evidence attached.";
    return {
      label:
        state.operation === "preview"
          ? "Preview open"
          : state.operation === "download"
            ? "Download requested"
            : "Attached",
      message,
      tone: "success",
      announcement: "polite",
    };
  }
  const failure = state.failure;
  const presentation =
    failure.presentation ?? cartularyErrorPresentation("unknown_future_error");
  const announcement = presentation.live === "polite" ? "polite" : "assertive";
  const blocked = accessBlocker(failure);
  if (blocked !== null) return { ...blocked, tone: "warning", announcement };
  let message: string;
  if (failure.kind === "authentication_required")
    message = "Sign in again to access evidence.";
  else if (
    failure.kind === "authorization_lost" ||
    presentation.family === "permission_or_incident_access_loss"
  )
    message = "You no longer have access to this evidence.";
  else if (failure.publicCode === "incident_closed")
    message = "This incident is closed. Evidence attachment is unavailable.";
  else if (
    failure.kind === "same_field_conflict" ||
    failure.publicCode === "row_version_conflict"
  )
    message = "The evidence record changed before the attachment completed.";
  else if (failure.kind === "client_txn_conflict")
    message = "The evidence change needs transaction recovery.";
  else if (failure.uploadFailure !== undefined)
    message = "The evidence file could not be uploaded.";
  else
    message =
      state.operation === "attach"
        ? "Evidence attachment failed."
        : state.operation === "preview"
          ? "The preview could not be opened."
          : "The download could not be requested.";
  return {
    label:
      state.operation === "attach"
        ? "Attach failed"
        : state.operation === "preview"
          ? "Preview unavailable"
          : "Download unavailable",
    message,
    tone: "danger",
    announcement,
  };
}

export function buildEvidenceAccessPresentation(
  lifecycle: EvidenceLifecycleViewModel,
  operation: EvidenceOperationState | null = null,
): EvidenceAccessPresentation {
  const lifecycleLabel =
    lifecycle.recordState === null
      ? "Unavailable"
      : lifecycleLabels[lifecycle.recordState];
  const uploadLabel =
    lifecycle.uploadOverlay === "pending_upload_slot"
      ? "Upload pending"
      : lifecycle.uploadOverlay === "failed_upload_slot"
        ? "Upload failed"
        : lifecycle.uploadOverlay === "quarantined_upload_slot"
          ? "File quarantined"
          : lifecycle.blobState === "available"
            ? "File available"
            : lifecycle.blobState === "pending"
              ? "File pending"
              : "No file attached";
  let stateKey: EvidenceAccessStateKey =
    lifecycle.recordState ?? "inconsistent";
  let message = lifecycle.accessEligible
    ? "Preview and download are available."
    : lifecycle.recordState === "requested"
      ? "Evidence has been requested and is not yet available for access."
      : lifecycle.recordState === "pending_receipt"
        ? "Evidence receipt is pending."
        : "Evidence is not yet available for access.";
  let tone: EvidenceFeedback["tone"] = lifecycle.accessEligible
    ? "success"
    : "neutral";
  let label = lifecycleLabel;
  if (lifecycle.consistency === "inconsistent") {
    stateKey = "inconsistent";
    label = "Inconsistent";
    message = accessBlockers.evidence_inconsistent.message;
    tone = "danger";
  } else if (
    lifecycle.recordState === "quarantined" ||
    lifecycle.blobState === "quarantined"
  ) {
    stateKey = "quarantined";
    label = "Quarantined";
    message = accessBlockers.evidence_quarantined.message;
    tone = "warning";
  } else if (lifecycle.blobState === "failed") {
    stateKey = "failed";
    label = "Upload failed";
    message = accessBlockers.blob_failed.message;
    tone = "danger";
  } else if (lifecycle.uploadOverlay === "pending_upload_slot") {
    stateKey = "pending_upload";
    label = "Upload pending";
    message = accessBlockers.blob_pending.message;
  }
  let canPreview = lifecycle.accessEligible;
  let canDownload = lifecycle.accessEligible;
  let feedback: EvidenceFeedback = {
    label,
    message,
    tone,
    announcement: "none",
  };
  if (operation !== null) {
    feedback = evidenceOperationFeedback(operation);
    if (operation.kind === "rejected" && operation.operation !== "attach") {
      const blocked = accessBlocker(operation.failure);
      stateKey = blocked?.stateKey ?? "public_error";
      canPreview = false;
      // Preview limitations do not grant download permission; retain only independent eligibility.
      canDownload =
        lifecycle.accessEligible &&
        operation.operation === "preview" &&
        blocked?.stateKey === "preview_blocked";
    }
  }
  return {
    ...feedback,
    stateKey,
    lifecycleLabel,
    uploadLabel,
    canPreview,
    canDownload,
  };
}
