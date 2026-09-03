const evidenceRecordLifecycleStates = [
  "requested",
  "pending_receipt",
  "received",
  "available",
  "quarantined",
  "released",
] as const;

export type EvidenceRecordLifecycleState =
  (typeof evidenceRecordLifecycleStates)[number];

const objectBlobUploadStates = [
  "pending",
  "available",
  "failed",
  "quarantined",
] as const;

export type ObjectBlobUploadState = (typeof objectBlobUploadStates)[number];

const evidenceLifecycleViewStateKeys = [
  "requested",
  "pending_upload",
  "available",
  "preview_blocked",
  "failed",
  "inconsistent",
  "blocked",
  "public_error",
] as const;

export type EvidenceLifecycleViewStateKey =
  (typeof evidenceLifecycleViewStateKeys)[number];

const evidenceCountDisplayStateKeys = [
  "empty",
  "available",
  "pending_upload",
  "blocked",
  "failed",
  "inconsistent",
] as const;

export type EvidenceCountDisplayStateKey =
  (typeof evidenceCountDisplayStateKeys)[number];

export type EvidenceLifecycleMessageTone =
  | "neutral"
  | "success"
  | "warning"
  | "danger";

export type EvidencePublicErrorInput = {
  readonly code?: unknown;
  readonly details?: unknown;
  readonly message?: unknown;
};

export type EvidenceLifecycleViewModelInput = {
  readonly evidenceLifecycleState?: unknown;
  readonly objectBlobUploadState?: unknown;
  readonly publicError?: EvidencePublicErrorInput | null;
};

export type EvidenceLifecycleViewModel = {
  readonly blobState: ObjectBlobUploadState | null;
  readonly canDownload: boolean;
  readonly canPreview: boolean;
  readonly countContribution: 0 | 1;
  readonly message: string | null;
  readonly messageTone: EvidenceLifecycleMessageTone;
  readonly publicErrorCode: string | null;
  readonly reasonCode: string | null;
  readonly recordState: EvidenceRecordLifecycleState | null;
  readonly stateKey: EvidenceLifecycleViewStateKey;
};

export type EvidenceLifecycleCountSummary = {
  readonly availableCount: number;
  readonly blockedCount: number;
  readonly countContribution: number;
  readonly failedCount: number;
  readonly inconsistentCount: number;
  readonly pendingUploadCount: number;
  readonly publicErrorCount: number;
  readonly requestedCount: number;
};

export type EvidenceCountDisplayViewModelInput = {
  readonly lifecycleViewModels?: readonly EvidenceLifecycleViewModel[];
  readonly projectedCount?: unknown;
  readonly projectedHasEvidence?: unknown;
};

export type EvidenceCountDisplayViewModel = {
  readonly displayCount: string;
  readonly hasEvidence: boolean;
  readonly projectedCount: number;
  readonly stateKey: EvidenceCountDisplayStateKey;
};

export function buildEvidenceLifecycleViewModel(
  input: EvidenceLifecycleViewModelInput,
): EvidenceLifecycleViewModel {
  const recordState = normalizeEvidenceRecordLifecycleState(
    input.evidenceLifecycleState,
  );
  const blobState = normalizeObjectBlobUploadState(input.objectBlobUploadState);
  const publicError = normalizeEvidencePublicError(input.publicError);
  const baseAvailable =
    blobState === "available" &&
    (recordState === "available" || recordState === "released");

  if (publicError !== null) {
    return buildPublicErrorViewModel({
      baseAvailable,
      blobState,
      publicError,
      recordState,
    });
  }

  if (recordState === null) {
    return evidenceLifecycleViewModel({
      blobState,
      message: "Inconsistent: evidence lifecycle state is not recognized.",
      messageTone: "danger",
      recordState,
      stateKey: "inconsistent",
    });
  }

  if (blobState === "failed") {
    return evidenceLifecycleViewModel({
      blobState,
      message: "Failed: object blob upload failed.",
      messageTone: "danger",
      recordState,
      stateKey: "failed",
    });
  }

  if (recordState === "quarantined" || blobState === "quarantined") {
    return evidenceLifecycleViewModel({
      blobState,
      message: "Blocked: evidence is quarantined.",
      messageTone: "warning",
      recordState,
      stateKey: "blocked",
    });
  }

  if (recordState === "requested") {
    if (blobState === "available") {
      return inconsistentViewModel(recordState, blobState);
    }
    return evidenceLifecycleViewModel({
      blobState,
      message: "Requested: evidence has not been received.",
      messageTone: "neutral",
      recordState,
      stateKey: "requested",
    });
  }

  if (recordState === "pending_receipt" || recordState === "received") {
    return evidenceLifecycleViewModel({
      blobState,
      message: "Blocked: evidence receipt or upload is still pending.",
      messageTone: "neutral",
      recordState,
      stateKey: "pending_upload",
    });
  }

  if (recordState === "available" || recordState === "released") {
    if (blobState === "available") {
      return evidenceLifecycleViewModel({
        blobState,
        canDownload: true,
        canPreview: true,
        countContribution: 1,
        message: null,
        messageTone: "success",
        recordState,
        stateKey: "available",
      });
    }
    if (blobState === "pending") {
      return evidenceLifecycleViewModel({
        blobState,
        message: "Blocked: object blob upload is still pending.",
        messageTone: "neutral",
        recordState,
        stateKey: "pending_upload",
      });
    }
    return inconsistentViewModel(recordState, blobState);
  }

  return inconsistentViewModel(recordState, blobState);
}

export function buildEvidenceCountDisplayViewModel(
  input: EvidenceCountDisplayViewModelInput,
): EvidenceCountDisplayViewModel {
  const normalizedCount = normalizeProjectedEvidenceCount(input.projectedCount);
  const normalizedHasEvidence = normalizeProjectedHasEvidence(
    input.projectedHasEvidence,
  );
  const summary =
    input.lifecycleViewModels === undefined
      ? null
      : summarizeEvidenceLifecycleCounts(input.lifecycleViewModels);
  const projectedCount =
    summary === null ? normalizedCount.value : summary.countContribution;
  const hasEvidence =
    normalizedHasEvidence.value === null
      ? projectedCount > 0
      : normalizedHasEvidence.value;

  if (
    !normalizedCount.valid ||
    !normalizedHasEvidence.valid ||
    hasEvidence !== projectedCount > 0 ||
    summary?.inconsistentCount
  ) {
    return evidenceCountDisplayViewModel("inconsistent", projectedCount, true);
  }
  if (projectedCount > 0) {
    return evidenceCountDisplayViewModel("available", projectedCount);
  }
  if (summary !== null) {
    if (summary.failedCount > 0) {
      return evidenceCountDisplayViewModel("failed", 0);
    }
    if (summary.blockedCount > 0 || summary.publicErrorCount > 0) {
      return evidenceCountDisplayViewModel("blocked", 0);
    }
    if (summary.pendingUploadCount > 0 || summary.requestedCount > 0) {
      return evidenceCountDisplayViewModel("pending_upload", 0);
    }
  }
  return evidenceCountDisplayViewModel("empty", 0);
}

export function summarizeEvidenceLifecycleCounts(
  viewModels: readonly EvidenceLifecycleViewModel[],
): EvidenceLifecycleCountSummary {
  const summary = {
    availableCount: 0,
    blockedCount: 0,
    countContribution: 0,
    failedCount: 0,
    inconsistentCount: 0,
    pendingUploadCount: 0,
    publicErrorCount: 0,
    requestedCount: 0,
  };

  for (const viewModel of viewModels) {
    summary.countContribution += viewModel.countContribution;
    switch (viewModel.stateKey) {
      case "available":
      case "preview_blocked":
        summary.availableCount += 1;
        break;
      case "blocked":
        summary.blockedCount += 1;
        break;
      case "failed":
        summary.failedCount += 1;
        break;
      case "inconsistent":
        summary.inconsistentCount += 1;
        break;
      case "pending_upload":
        summary.pendingUploadCount += 1;
        break;
      case "public_error":
        summary.publicErrorCount += 1;
        break;
      case "requested":
        summary.requestedCount += 1;
        break;
    }
  }

  return summary;
}

function buildPublicErrorViewModel({
  baseAvailable,
  blobState,
  publicError,
  recordState,
}: {
  readonly baseAvailable: boolean;
  readonly blobState: ObjectBlobUploadState | null;
  readonly publicError: NormalizedEvidencePublicError;
  readonly recordState: EvidenceRecordLifecycleState | null;
}): EvidenceLifecycleViewModel {
  if (publicError.reasonCode === "evidence_inconsistent") {
    return evidenceLifecycleViewModel({
      blobState,
      message: `Inconsistent: ${publicError.message}`,
      messageTone: "danger",
      publicErrorCode: publicError.code,
      reasonCode: publicError.reasonCode,
      recordState,
      stateKey: "inconsistent",
    });
  }
  if (publicError.reasonCode === "blob_failed") {
    return evidenceLifecycleViewModel({
      blobState,
      message: `Failed: ${publicError.message}`,
      messageTone: "danger",
      publicErrorCode: publicError.code,
      reasonCode: publicError.reasonCode,
      recordState,
      stateKey: "failed",
    });
  }
  if (
    baseAvailable &&
    (publicError.reasonCode === "unsupported_preview" ||
      publicError.reasonCode === "preview_payload_too_large")
  ) {
    return evidenceLifecycleViewModel({
      blobState,
      canDownload: true,
      countContribution: 1,
      message: `Preview blocked: ${publicError.message}`,
      messageTone: "warning",
      publicErrorCode: publicError.code,
      reasonCode: publicError.reasonCode,
      recordState,
      stateKey: "preview_blocked",
    });
  }
  return evidenceLifecycleViewModel({
    blobState,
    message: publicError.message,
    messageTone: "danger",
    publicErrorCode: publicError.code,
    reasonCode: publicError.reasonCode,
    recordState,
    stateKey: "public_error",
  });
}

function evidenceLifecycleViewModel({
  blobState,
  canDownload = false,
  canPreview = false,
  countContribution = 0,
  message,
  messageTone,
  publicErrorCode = null,
  reasonCode = null,
  recordState,
  stateKey,
}: {
  readonly blobState: ObjectBlobUploadState | null;
  readonly canDownload?: boolean;
  readonly canPreview?: boolean;
  readonly countContribution?: 0 | 1;
  readonly message: string | null;
  readonly messageTone: EvidenceLifecycleMessageTone;
  readonly publicErrorCode?: string | null;
  readonly reasonCode?: string | null;
  readonly recordState: EvidenceRecordLifecycleState | null;
  readonly stateKey: EvidenceLifecycleViewStateKey;
}): EvidenceLifecycleViewModel {
  return {
    blobState,
    canDownload,
    canPreview,
    countContribution,
    message,
    messageTone,
    publicErrorCode,
    reasonCode,
    recordState,
    stateKey,
  };
}

function evidenceCountDisplayViewModel(
  stateKey: EvidenceCountDisplayStateKey,
  projectedCount: number,
  hasEvidence = projectedCount > 0,
): EvidenceCountDisplayViewModel {
  return {
    displayCount: String(projectedCount),
    hasEvidence,
    projectedCount,
    stateKey,
  };
}

function inconsistentViewModel(
  recordState: EvidenceRecordLifecycleState | null,
  blobState: ObjectBlobUploadState | null,
): EvidenceLifecycleViewModel {
  return evidenceLifecycleViewModel({
    blobState,
    message: "Inconsistent: evidence lifecycle and object blob state disagree.",
    messageTone: "danger",
    recordState,
    stateKey: "inconsistent",
  });
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

type NormalizedEvidencePublicError = {
  readonly code: string | null;
  readonly message: string;
  readonly reasonCode: string | null;
};

function normalizeEvidencePublicError(
  value: EvidencePublicErrorInput | null | undefined,
): NormalizedEvidencePublicError | null {
  if (value === null || value === undefined) {
    return null;
  }
  const code = typeof value.code === "string" ? value.code : null;
  const message =
    typeof value.message === "string" && value.message.trim() !== ""
      ? value.message
      : (code ?? "Evidence access failed.");
  return {
    code,
    message,
    reasonCode: readReasonCode(value.details),
  };
}

function readReasonCode(details: unknown): string | null {
  if (
    typeof details !== "object" ||
    details === null ||
    !("reason_code" in details)
  ) {
    return null;
  }
  const reasonCode = details.reason_code;
  return typeof reasonCode === "string" && reasonCode.trim() !== ""
    ? reasonCode
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
