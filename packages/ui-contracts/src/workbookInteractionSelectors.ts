import type { StableTestId } from "./selectorCore";
import {
  dataTestIdPrefixSelector,
  encodeSelectorSegment,
  itemRefTestId,
  recordFieldTestId,
  recordTestId,
  requireClosedToken,
  requireFieldKey,
  requireNonEmptySelectorValue,
  requireRecordId,
  rowCellTestId,
  semanticSelectorTestId,
  stableTestId,
  tokenScopedTestId,
} from "./selectorCore";

export type RowHistoryRollbackAction =
  | "change_set"
  | "history_entry"
  | "row_restore";

export const rowHistoryRollbackActions = [
  "change_set",
  "history_entry",
  "row_restore",
] as const satisfies readonly RowHistoryRollbackAction[];

export type RowHistoryDestructiveOperation = "delete" | "restore";

export const rowHistoryDestructiveOperations = [
  "delete",
  "restore",
] as const satisfies readonly RowHistoryDestructiveOperation[];

export type TimelineScalarEditorSurface = "grid" | "inspector";

export const timelineScalarEditorSurfaces = [
  "grid",
  "inspector",
] as const satisfies readonly TimelineScalarEditorSurface[];

export type TimelineInspectorSection =
  | "operational-text"
  | "evidence"
  | "history"
  | "relationships";

export const timelineInspectorSections = [
  "operational-text",
  "relationships",
  "evidence",
  "history",
] as const satisfies readonly TimelineInspectorSection[];

export type WorkbookConflictControl =
  | "activate-origin"
  | "apply-collection"
  | "close"
  | "keep-saved"
  | "merged-value"
  | "paste-navigator"
  | "paste-next"
  | "paste-position"
  | "paste-previous"
  | "use-merged"
  | "use-server-suggestion"
  | "use-unsaved";

export type RowHistoryItemAnchor = {
  readonly historyItemRef: string;
};

export type RowHistoryActionAnchor = RowHistoryItemAnchor & {
  readonly action: RowHistoryRollbackAction;
};

export type RowHistoryDestructiveAnchor = {
  readonly operation: RowHistoryDestructiveOperation;
};

const workbookConflictControlTestIds = Object.freeze({
  "activate-origin": "conflict-activate-origin",
  "apply-collection": "conflict-apply-collection",
  close: "conflict-close",
  "keep-saved": "conflict-keep-saved",
  "merged-value": "conflict-merged-value",
  "paste-navigator": "paste-conflict-navigator",
  "paste-next": "paste-conflict-next",
  "paste-position": "paste-conflict-position",
  "paste-previous": "paste-conflict-previous",
  "use-merged": "conflict-use-merged",
  "use-server-suggestion": "conflict-use-server-suggestion",
  "use-unsaved": "conflict-use-unsaved",
} satisfies Record<WorkbookConflictControl, string>);

export function workbookConflictControlTestId(
  control: WorkbookConflictControl,
): StableTestId {
  return semanticSelectorTestId(
    workbookConflictControlTestIds,
    control,
    "workbook conflict control",
  );
}

export function timelineMutationSubstrateReadyTestId(): string {
  return "timeline-mutation-substrate-ready";
}

export function timelineInspectorSectionTestId(
  section: TimelineInspectorSection,
): StableTestId {
  return stableTestId(
    `timeline-inspector-section-${requireTimelineInspectorSection(section)}`,
  );
}

export function timelineInspectorMessageTestId(): StableTestId {
  return stableTestId("timeline-inspector-message");
}

export function timelineInspectorTestId(): StableTestId {
  return stableTestId("timeline-inspector");
}

export function conflictMarkerTestId(
  recordId: string,
  fieldKey: string,
): string {
  return recordFieldTestId("conflict-marker", recordId, fieldKey);
}

export function workbookConflictResolverTestId(): string {
  return "workbook-conflict-resolver";
}

export function workbookConflictSummaryTestId(): string {
  return "workbook-conflict-summary";
}

export function workbookConflictSavedValueTestId(): string {
  return "workbook-conflict-saved-value";
}

export function workbookConflictLocalValueTestId(): string {
  return "workbook-conflict-local-value";
}

export function workbookEditRecoveryTestId(): string {
  return "workbook-edit-recovery";
}

export function workbookEditRecoveryRetryButtonTestId(): string {
  return "workbook-edit-recovery-retry";
}

export function workbookEditRecoveryDiscardButtonTestId(): string {
  return "workbook-edit-recovery-discard";
}

export function workbookActiveSurfaceFocusTargetTestId(): string {
  return "workbook-active-surface-focus-target";
}

export function rowPresenceMarkerTestId(recordId: string): string {
  return `presence-row-${requireRecordId(recordId)}`;
}

export function cellPresenceMarkerTestId(
  recordId: string,
  fieldKey: string,
): string {
  return recordFieldTestId("presence-cell", recordId, fieldKey);
}

export function saveStateTestId(): string {
  return "save-state";
}

export function saveStateActionButtonTestId(): string {
  return "save-state-action";
}

export function workbookFocusAnchorTestId(): string {
  return "workbook-focus-anchor";
}

export function workbookPresenceSummaryTestId(): string {
  return "presence-header";
}

export function pendingQueueNoticeTestId(): string {
  return "pending-queue-notice";
}

export function pendingQueueCountTestId(): string {
  return "pending-queue-count";
}

export function timelineScalarEditorTestId(options: {
  readonly fieldKey: string;
  readonly recordId: string | null;
  readonly surface?: TimelineScalarEditorSurface | undefined;
}): string {
  const base =
    options.recordId === null
      ? draftCellTestId(options.fieldKey)
      : rowCellTestId(options.recordId, options.fieldKey);
  if (options.surface === "inspector") return `${base}-inspector`;
  return options.recordId === null ? base : `${base}-grid-editor`;
}

export function rowHistoryOpenButtonTestId(recordId: string): string {
  return recordTestId("row-history-open", recordId);
}

export function rowHistoryOpenInspectorButtonTestId(recordId: string): string {
  return `${rowHistoryOpenButtonTestId(recordId)}-inspector`;
}

export function rowHistoryPanelTestId(): string {
  return "row-history-panel";
}

export function rowHistoryOpenSelectedButtonTestId(): string {
  return "row-history-open-selected";
}

export function rowHistoryLoadingTestId(): string {
  return "row-history-loading";
}

export function rowHistoryMessageTestId(): string {
  return "row-history-message";
}

export function rowHistoryDeleteButtonTestId(): string {
  return "row-history-delete";
}

export function rowHistoryRestoreButtonTestId(): string {
  return "row-history-restore";
}

export function rowHistoryDestructiveConfirmPanelTestId(
  anchor: RowHistoryDestructiveAnchor,
): string {
  return `row-history-destructive-confirm-${requireRowHistoryDestructiveOperation(anchor.operation)}`;
}

export function rowHistoryDestructiveConfirmButtonTestId(
  anchor: RowHistoryDestructiveAnchor,
): string {
  return `${rowHistoryDestructiveConfirmPanelTestId(anchor)}-confirm`;
}

export function rowHistoryDestructiveCancelButtonTestId(
  anchor: RowHistoryDestructiveAnchor,
): string {
  return `${rowHistoryDestructiveConfirmPanelTestId(anchor)}-cancel`;
}

export function rowHistoryItemTestId(anchor: RowHistoryItemAnchor): string {
  return `row-history-item-${encodeSelectorSegment(
    rowHistoryItemIdentity(anchor),
    "row history item identity",
  )}`;
}

export function rowHistoryActionTestId(anchor: RowHistoryActionAnchor): string {
  const action = requireRowHistoryRollbackAction(anchor.action);
  return `row-history-action-${encodeSelectorSegment(
    rowHistoryActionIdentity(anchor),
    "row history action identity",
  )}-${action}`;
}

export function rowHistoryRollbackPreviewTestId(
  anchor: RowHistoryActionAnchor,
): string {
  const action = requireRowHistoryRollbackAction(anchor.action);
  return `row-history-rollback-preview-${encodeSelectorSegment(
    rowHistoryActionIdentity(anchor),
    "row history action identity",
  )}-${action}`;
}

export function rowHistoryRollbackConfirmButtonTestId(
  anchor: RowHistoryActionAnchor,
): string {
  return `${rowHistoryRollbackPreviewTestId(anchor)}-confirm`;
}

export function rowHistoryRollbackCancelButtonTestId(
  anchor: RowHistoryActionAnchor,
): string {
  return `${rowHistoryRollbackPreviewTestId(anchor)}-cancel`;
}

export function draftCellTestId(fieldKey: string): string {
  return draftFieldTestId(fieldKey);
}

export function draftRowCreateButtonTestId(): string {
  return "draft-row-create";
}

export function relationshipItemsTestId(
  recordId: string,
  fieldKey: string,
  surface: "grid" | "inspector" = "inspector",
): string {
  const identity = timelineCollectionFieldControlTestId(
    recordId,
    fieldKey,
    "items",
  );
  return surface === "grid" ? `${identity}-grid` : identity;
}

export function relationshipOverflowButtonTestId(
  recordId: string,
  fieldKey: string,
): string {
  return timelineCollectionFieldControlTestId(recordId, fieldKey, "overflow");
}

export function draftRelationshipItemsTestId(fieldKey: string): string {
  return timelineCollectionFieldControlTestId(null, fieldKey, "items");
}

export function timelineCollectionInputTestId(
  recordId: string,
  fieldKey: string,
  surface: "grid" | "inspector" = "inspector",
): string {
  const identity = timelineCollectionFieldControlTestId(
    recordId,
    fieldKey,
    "input",
  );
  return surface === "grid" ? `${identity}-grid` : identity;
}

export function draftTimelineCollectionInputTestId(fieldKey: string): string {
  return timelineCollectionFieldControlTestId(null, fieldKey, "input");
}

export function timelineRowVersionTestId(recordId: string): string {
  return rowCellTestId(recordId, "row_version");
}

export function timelineRowMarkReviewedButtonTestId(recordId: string): string {
  return recordTestId("row", recordId, "mark-reviewed");
}

export function timelineRowReplacementInputTestId(recordId: string): string {
  return recordTestId("row", recordId, "replacement-id");
}

export function timelineRowSupersedeButtonTestId(recordId: string): string {
  return recordTestId("row", recordId, "supersede");
}

export function timelineEvidenceFileInputTestId(recordId: string): string {
  return recordTestId("timeline-evidence-file", recordId);
}

export function timelineDraftEvidenceFileInputTestId(): string {
  return "timeline-evidence-file-draft";
}

export function timelineEvidenceAttachSectionTestId(recordId: string): string {
  return recordTestId("timeline-evidence-attach", recordId);
}

export function timelineDraftEvidenceAttachSectionTestId(): string {
  return "timeline-evidence-attach-draft";
}

export function timelinePreviewRowTestId(recordId: string): string {
  return recordTestId("timeline-preview-row", recordId);
}

export function relationshipChipTestId(itemRef: string): string {
  return itemRefTestId("chip", itemRef);
}

export function mentionItemTestId(itemRef: string): string {
  return itemRefTestId("mention", itemRef);
}

export function autoResolutionNoticeTestId(itemRef: string): string {
  return itemRefTestId("auto-resolution-notice", itemRef);
}

export function autoResolutionNoticeFamilySelector(): string {
  return dataTestIdPrefixSelector("auto-resolution-notice-");
}

export function autoResolutionUndoButtonTestId(itemRef: string): string {
  return `${autoResolutionNoticeTestId(itemRef)}-undo`;
}

export function autoResolutionReviewButtonTestId(itemRef: string): string {
  return `${autoResolutionNoticeTestId(itemRef)}-review`;
}

export function pasteConflictItemTestId(itemKey: string): string {
  return `paste-conflict-item-${encodeSelectorSegment(itemKey, "paste conflict key")}`;
}

function draftFieldTestId(fieldKey: string, suffix?: string): string {
  return tokenScopedTestId("draft-row", requireFieldKey(fieldKey), suffix);
}

function timelineCollectionFieldControlTestId(
  recordId: string | null,
  fieldKey: string,
  suffix: "items" | "input" | "overflow",
): string {
  return recordId === null
    ? draftFieldTestId(fieldKey, suffix)
    : recordFieldTestId("row", recordId, fieldKey, suffix);
}

function requireTimelineInspectorSection(
  section: TimelineInspectorSection,
): TimelineInspectorSection {
  return requireClosedToken(
    timelineInspectorSections,
    section,
    "timeline inspector section",
  );
}

function requireRowHistoryRollbackAction(
  value: string,
): RowHistoryRollbackAction {
  if (
    value === "change_set" ||
    value === "history_entry" ||
    value === "row_restore"
  ) {
    return value;
  }
  throw new Error(`Invalid row history rollback action token: ${value}`);
}

function requireRowHistoryDestructiveOperation(
  value: string,
): RowHistoryDestructiveOperation {
  if (value === "delete" || value === "restore") {
    return value;
  }
  throw new Error(`Invalid row history destructive operation token: ${value}`);
}

function rowHistoryItemIdentity(anchor: RowHistoryItemAnchor): string {
  return requireNonEmptySelectorValue(
    anchor.historyItemRef,
    "history_item_ref",
  );
}

function rowHistoryActionIdentity(anchor: RowHistoryActionAnchor): string {
  return requireNonEmptySelectorValue(
    anchor.historyItemRef,
    "history_item_ref",
  );
}
