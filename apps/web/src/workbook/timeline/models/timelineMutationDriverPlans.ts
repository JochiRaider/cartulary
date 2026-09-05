import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type {
  PendingQueueSnapshot,
  PendingReplaySettlement,
  PendingReplayUnitInput,
  PendingReplayUnitState,
} from "../../utils/workbookPendingQueue";
import type { TimelineReplayContext } from "./timelineControllerPorts";

export type TimelinePendingReplayAdmission = PendingReplayUnitInput &
  Omit<TimelineReplayContext, "sheetRef">;

export type TimelineReplayAdmissionPlan =
  | { readonly kind: "dispatch"; readonly committedRowVersion: number | null }
  | { readonly kind: "idle" }
  | { readonly kind: "reject_missing_metadata" }
  | { readonly kind: "retry_missing_committed_row" }
  | {
      readonly kind: "pause";
      readonly reason:
        | "authentication"
        | "conflict"
        | "halted"
        | "owner_mismatch"
        | "refresh";
    };

export function planTimelineReplayAdmission({
  candidate,
  currentRowVersion,
  envelopeViewSchemaId,
  expectedUnitId,
  hasLocalConflict,
  hasMetadata,
  refreshBlocked,
  snapshot,
}: {
  readonly candidate: PendingReplayUnitState | null;
  readonly currentRowVersion: number | null | undefined;
  readonly envelopeViewSchemaId: string;
  readonly expectedUnitId: string;
  readonly hasLocalConflict: boolean;
  readonly hasMetadata: boolean;
  readonly refreshBlocked: boolean;
  readonly snapshot: PendingQueueSnapshot;
}): TimelineReplayAdmissionPlan {
  if (snapshot.authPaused) return { kind: "pause", reason: "authentication" };
  if (snapshot.halted !== null) return { kind: "pause", reason: "halted" };
  if (snapshot.sameFieldConflicts.length > 0 || hasLocalConflict) {
    return { kind: "pause", reason: "conflict" };
  }
  if (candidate === null) return { kind: "idle" };
  if (
    candidate.id !== expectedUnitId ||
    candidate.viewSchemaId !== envelopeViewSchemaId
  ) {
    return { kind: "pause", reason: "owner_mismatch" };
  }
  if (refreshBlocked) return { kind: "pause", reason: "refresh" };
  if (!hasMetadata) return { kind: "reject_missing_metadata" };
  if (
    candidate.kind === "patch" &&
    (currentRowVersion === null || currentRowVersion === undefined)
  ) {
    return { kind: "retry_missing_committed_row" };
  }
  return {
    kind: "dispatch",
    committedRowVersion: currentRowVersion ?? null,
  };
}

export type TimelineRejectedSettlementPlan =
  | { readonly kind: "request_authorization" }
  | {
      readonly kind: "register_conflict";
      readonly conflict: Extract<
        WorkbookOperationFailure,
        { readonly kind: "same_field_conflict" }
      >["conflict"];
    }
  | { readonly kind: "retry" }
  | { readonly kind: "halt"; readonly message: string }
  | { readonly kind: "invalid_settlement"; readonly message: string };

export function planTimelineRejectedSettlement(
  settlement: PendingReplaySettlement,
  failure: WorkbookOperationFailure,
): TimelineRejectedSettlementPlan {
  switch (settlement.outcome) {
    case "auth_paused":
      return { kind: "request_authorization" };
    case "same_field_conflict":
      return failure.kind === "same_field_conflict"
        ? { kind: "register_conflict", conflict: failure.conflict }
        : {
            kind: "invalid_settlement",
            message: "Conflict settlement did not include an exact conflict.",
          };
    case "retryable_failure":
      return { kind: "retry" };
    case "halted":
      return { kind: "halt", message: settlement.halt.message };
    case "no_dispatched_unit":
      return {
        kind: "invalid_settlement",
        message: "Mutation settlement lost its dispatched unit.",
      };
    case "success":
      return {
        kind: "invalid_settlement",
        message: "A rejected mutation produced a success settlement.",
      };
  }
}

export type TimelineAcceptedProjectionPlan = {
  readonly kind: "apply_committed_result";
  readonly preserveKnownCommittedRow: boolean;
  readonly refreshAfterApply: boolean;
};

export function planTimelineAcceptedProjection({
  currentRowVersion,
  postMutationQueryRefreshRequired,
  responseRowVersion,
}: {
  readonly currentRowVersion: number | null | undefined;
  readonly postMutationQueryRefreshRequired: boolean;
  readonly responseRowVersion: number;
}): TimelineAcceptedProjectionPlan {
  return {
    kind: "apply_committed_result",
    preserveKnownCommittedRow:
      currentRowVersion !== null &&
      currentRowVersion !== undefined &&
      currentRowVersion > responseRowVersion,
    refreshAfterApply: postMutationQueryRefreshRequired,
  };
}

export type TimelineDiscardPlan =
  | { readonly kind: "refused" }
  | {
      readonly kind: "reconcile";
      readonly clearViewportContinuity: boolean;
    };

export function planTimelineDiscard({
  hasMetadata,
  recovered,
}: {
  readonly hasMetadata: boolean;
  readonly recovered: boolean;
}): TimelineDiscardPlan {
  return recovered
    ? { kind: "reconcile", clearViewportContinuity: hasMetadata }
    : { kind: "refused" };
}
