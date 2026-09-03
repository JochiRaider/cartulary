import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookRow } from "./timelineRowModel";

export type TimelineEvidenceActionContext = {
  readonly authorized: boolean;
  readonly capabilityAvailable: boolean;
  readonly selectedRowKey: string | null;
  readonly surfaceKey: string;
};

export type TimelineEvidenceTargetIdentity = {
  readonly originalRecordId: string | null;
  readonly rowKey: string;
  readonly surfaceKey: string;
};

export type TimelineEvidenceTargetPlan =
  | { readonly kind: "dispatch"; readonly target: WorkbookRow }
  | {
      readonly kind: "reject";
      readonly reason:
        | "authorization_lost"
        | "capability_unavailable"
        | "surface_changed"
        | "selection_changed"
        | "target_missing"
        | "target_identity_changed"
        | "target_not_dispatchable";
    };

export function timelineEvidenceTargetIdentity(
  target: WorkbookRow,
  surfaceKey: string,
): TimelineEvidenceTargetIdentity {
  return {
    originalRecordId: target.recordId,
    rowKey: target.key,
    surfaceKey,
  };
}

export function planTimelineEvidenceTarget(options: {
  readonly context: TimelineEvidenceActionContext;
  readonly identity: TimelineEvidenceTargetIdentity;
  readonly rows: readonly WorkbookRow[];
}): TimelineEvidenceTargetPlan {
  if (!options.context.authorized) {
    return { kind: "reject", reason: "authorization_lost" };
  }
  if (!options.context.capabilityAvailable) {
    return { kind: "reject", reason: "capability_unavailable" };
  }
  if (options.context.surfaceKey !== options.identity.surfaceKey) {
    return { kind: "reject", reason: "surface_changed" };
  }
  const selectionMatches =
    options.context.selectedRowKey === options.identity.rowKey ||
    (options.identity.originalRecordId === null &&
      options.context.selectedRowKey === null);
  if (!selectionMatches) {
    return { kind: "reject", reason: "selection_changed" };
  }
  const target = options.rows.find(
    (candidate) => candidate.key === options.identity.rowKey,
  );
  if (target === undefined) {
    return { kind: "reject", reason: "target_missing" };
  }
  if (
    options.identity.originalRecordId !== null &&
    target.recordId !== options.identity.originalRecordId
  ) {
    return { kind: "reject", reason: "target_identity_changed" };
  }
  if (
    target.viewSchemaId !== timelineViewSchemaId ||
    target.pendingSignature !== null ||
    (target.recordId === null) !== (target.rowVersion === null)
  ) {
    return { kind: "reject", reason: "target_not_dispatchable" };
  }
  return { kind: "dispatch", target };
}
