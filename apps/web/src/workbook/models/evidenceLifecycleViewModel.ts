export type EvidenceRecordLifecycleState =
  | "requested"
  | "pending_receipt"
  | "received"
  | "available"
  | "quarantined"
  | "released";
export type ObjectBlobUploadState =
  | "pending"
  | "available"
  | "failed"
  | "quarantined";
export type EvidenceUploadOverlay =
  | "none"
  | "pending_upload_slot"
  | "failed_upload_slot"
  | "quarantined_upload_slot";

export type EvidenceLifecycleViewModel = {
  readonly recordState: EvidenceRecordLifecycleState | null;
  readonly blobState: ObjectBlobUploadState | null;
  readonly uploadOverlay: EvidenceUploadOverlay;
  readonly consistency: "consistent" | "inconsistent" | "unverified";
  /** Local eligibility is only a precondition; each invocation rechecks access on the server. */
  readonly accessEligible: boolean;
};

export function buildEvidenceLifecycleViewModel(input: {
  /** The workbook projection reports pending even when no blob is linked. */
  readonly uploadStateSource?: "object_blob" | "evidence_projection";
  readonly evidenceLifecycleState?: unknown;
  readonly objectBlobUploadState?: unknown;
}): EvidenceLifecycleViewModel {
  const recordState = normalizeEvidenceRecordLifecycleState(
    input.evidenceLifecycleState,
  );
  const blobState = normalizeObjectBlobUploadState(input.objectBlobUploadState);
  const absentBlob =
    input.objectBlobUploadState === null ||
    input.objectBlobUploadState === undefined;
  const unreportedPending =
    input.uploadStateSource === "evidence_projection" &&
    blobState === "pending";
  const inconsistent =
    recordState === null ||
    (!absentBlob && blobState === null) ||
    ((recordState === "available" || recordState === "released") &&
      blobState !== "available") ||
    (recordState === "quarantined" &&
      blobState !== null &&
      blobState !== "quarantined" &&
      !unreportedPending);
  return {
    recordState,
    blobState,
    uploadOverlay:
      blobState === "pending" && !unreportedPending
        ? "pending_upload_slot"
        : blobState === "failed"
          ? "failed_upload_slot"
          : blobState === "quarantined"
            ? "quarantined_upload_slot"
            : "none",
    consistency: inconsistent
      ? "inconsistent"
      : unreportedPending
        ? "unverified"
        : "consistent",
    accessEligible:
      !inconsistent &&
      blobState === "available" &&
      (recordState === "available" || recordState === "released"),
  };
}

export type EvidenceCountDisplayStateKey =
  | "empty"
  | "available"
  | "inconsistent";
export type EvidenceCountDisplayViewModel = {
  readonly displayCount: string;
  readonly hasEvidence: boolean;
  readonly projectedCount: number;
  readonly stateKey: EvidenceCountDisplayStateKey;
};

/** Counts are record projections. Upload slots and access outcomes never contribute. */
export function buildEvidenceCountDisplayViewModel(input: {
  readonly projectedCount?: unknown;
  readonly projectedHasEvidence?: unknown;
}): EvidenceCountDisplayViewModel {
  const count = normalizeProjectedEvidenceCount(input.projectedCount);
  const flag = normalizeProjectedHasEvidence(input.projectedHasEvidence);
  const hasEvidence = flag.value ?? count.value > 0;
  const inconsistent =
    !count.valid || !flag.valid || hasEvidence !== count.value > 0;
  return {
    displayCount: String(count.value),
    hasEvidence: inconsistent ? true : hasEvidence,
    projectedCount: count.value,
    stateKey: inconsistent
      ? "inconsistent"
      : count.value > 0
        ? "available"
        : "empty",
  };
}

function normalizeEvidenceRecordLifecycleState(
  value: unknown,
): EvidenceRecordLifecycleState | null {
  return value === "requested" ||
    value === "pending_receipt" ||
    value === "received" ||
    value === "available" ||
    value === "quarantined" ||
    value === "released"
    ? value
    : null;
}

function normalizeObjectBlobUploadState(
  value: unknown,
): ObjectBlobUploadState | null {
  return value === "pending" ||
    value === "available" ||
    value === "failed" ||
    value === "quarantined"
    ? value
    : null;
}

function normalizeProjectedEvidenceCount(value: unknown): {
  readonly valid: boolean;
  readonly value: number;
} {
  if (value === null || value === undefined || value === "") {
    return { valid: true, value: 0 };
  }
  if (typeof value === "number") {
    return Number.isInteger(value) && value >= 0
      ? { valid: true, value }
      : { valid: false, value: 0 };
  }
  if (typeof value === "string" && /^[0-9]+$/u.test(value.trim())) {
    return { valid: true, value: Number.parseInt(value.trim(), 10) };
  }
  return { valid: false, value: 0 };
}

function normalizeProjectedHasEvidence(value: unknown): {
  readonly valid: boolean;
  readonly value: boolean | null;
} {
  if (value === null || value === undefined || value === "") {
    return { valid: true, value: null };
  }
  if (typeof value === "boolean") {
    return { valid: true, value };
  }
  if (value === "true") {
    return { valid: true, value: true };
  }
  if (value === "false") {
    return { valid: true, value: false };
  }
  return { valid: false, value: null };
}
