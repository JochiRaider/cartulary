import { listViewSchemaRegistryEntries } from "@cartulary/protocol-ts";

export type WorkbookSurface = string;

export type EntityType = "host" | "identity";

export const entityTypes = [
  "host",
  "identity",
] as const satisfies readonly EntityType[];

export type EntityMentionResolutionStatus =
  | "dismissed"
  | "resolved"
  | "unresolved";

export const entityMentionResolutionStatuses = [
  "unresolved",
  "resolved",
  "dismissed",
] as const satisfies readonly EntityMentionResolutionStatus[];

export type RowHistoryRollbackAction =
  | "change_set"
  | "history_entry"
  | "row_restore";

export const rowHistoryRollbackActions = [
  "change_set",
  "history_entry",
  "row_restore",
] as const satisfies readonly RowHistoryRollbackAction[];

export type TimelineScalarEditorSurface = "grid" | "inspector";

export const timelineScalarEditorSurfaces = [
  "grid",
  "inspector",
] as const satisfies readonly TimelineScalarEditorSurface[];

export type RowHistoryItemAnchor = {
  readonly historyItemRef: string;
};

export type RowHistoryActionAnchor = RowHistoryItemAnchor & {
  readonly action: RowHistoryRollbackAction;
};

const registeredViewSchemaIds = Object.freeze(
  new Set(listViewSchemaRegistryEntries().map((entry) => entry.view_schema_id)),
);

export function gridShellTestId(viewSchemaId: WorkbookSurface): string {
  return `${requireViewSchemaId(viewSchemaId)}-grid-shell`;
}

export function surfaceTabTestId(viewSchemaId: string): string {
  return `surface-tab-${requireViewSchemaId(viewSchemaId)}`;
}

export function systemViewSelectorTestId(): string {
  return "system-view-selector";
}

export function workbookShellReadyTestId(): string {
  return "workbook-shell-ready";
}

export function timelineMutationSubstrateReadyTestId(): string {
  return "timeline-mutation-substrate-ready";
}

export function currentIncidentRoleTestId(): string {
  return "current-incident-role";
}

export function landingIncidentCardTestId(incidentId: string): string {
  return `landing-incident-${encodeSelectorSegment(incidentId, "incident_id")}`;
}

export function landingIncidentOpenButtonTestId(incidentId: string): string {
  return `landing-open-${encodeSelectorSegment(incidentId, "incident_id")}`;
}

export function incidentMembershipEmailInputTestId(): string {
  return "incident-membership-email";
}

export function incidentMembershipRoleSelectTestId(): string {
  return "incident-membership-role";
}

export function incidentMembershipCreateButtonTestId(): string {
  return "incident-membership-create";
}

export function incidentMembershipAdminNoteTestId(): string {
  return "incident-membership-admin-note";
}

export function incidentMembershipListTestId(): string {
  return "incident-membership-list";
}

export function incidentMembershipRowTestId(userId: string): string {
  return `incident-membership-row-${encodeSelectorSegment(userId, "user_id")}`;
}

export function incidentMembershipVersionTestId(userId: string): string {
  return `incident-membership-version-${encodeSelectorSegment(userId, "user_id")}`;
}

export function incidentMembershipRoleInputTestId(userId: string): string {
  return `incident-membership-role-input-${encodeSelectorSegment(userId, "user_id")}`;
}

export function incidentMembershipPatchButtonTestId(userId: string): string {
  return `incident-membership-patch-${encodeSelectorSegment(userId, "user_id")}`;
}

export function incidentMembershipDeleteButtonTestId(userId: string): string {
  return `incident-membership-delete-${encodeSelectorSegment(userId, "user_id")}`;
}

export function incidentMembershipRoleDisplayTestId(userId: string): string {
  return `incident-membership-role-${encodeSelectorSegment(userId, "user_id")}`;
}

export function phase2IncidentRowTestId(incidentId: string): string {
  return `incident-row-${encodeSelectorSegment(incidentId, "incident_id")}`;
}

export function phase2SelectIncidentButtonTestId(incidentId: string): string {
  return `select-incident-${encodeSelectorSegment(incidentId, "incident_id")}`;
}

export function phase2MembershipRowTestId(userId: string): string {
  return `membership-row-${encodeSelectorSegment(userId, "user_id")}`;
}

export function phase2MembershipRoleInputTestId(userId: string): string {
  return `membership-role-input-${encodeSelectorSegment(userId, "user_id")}`;
}

export function phase2MembershipVersionTestId(userId: string): string {
  return `membership-version-${encodeSelectorSegment(userId, "user_id")}`;
}

export function phase2MembershipPatchButtonTestId(userId: string): string {
  return `patch-membership-${encodeSelectorSegment(userId, "user_id")}`;
}

export function phase2MembershipDeleteButtonTestId(userId: string): string {
  return `delete-membership-${encodeSelectorSegment(userId, "user_id")}`;
}

export function extensionProfileRowTestId(profileId: string): string {
  return `extension-${encodeSelectorSegment(profileId, "extension_profile_id")}`;
}

export function gridScrollportClassName(): string {
  return "cartulary-grid-scrollport";
}

export function gridScrollportSelector(): string {
  return `.${gridScrollportClassName()}`;
}

export function gridActionsHeaderTestId(viewSchemaId: WorkbookSurface): string {
  return `${requireViewSchemaId(viewSchemaId)}-actions-header`;
}

/**
 * Scope this selector through `gridShellTestId(surface)` when targeting
 * workbook rows. Do not rely on raw table markup or renderer classes.
 */
export function gridSavedRowsSelector(): string {
  return '[role="row"][data-grid-record-id]:not([data-grid-record-id=""])';
}

export function gridSavedRowSelector(recordId: string): string {
  return `[role="row"][data-grid-record-id="${cssAttributeValue(
    requireNonEmptySelectorValue(recordId, "record_id"),
  )}"]`;
}

/**
 * Scope this selector through `gridShellTestId(surface)` when targeting the
 * workbook draft row. Do not rely on raw table markup or renderer classes.
 */
export function gridDraftRowSelector(): string {
  return '[role="row"][data-grid-record-id=""]';
}

export function conflictMarkerTestId(
  recordId: string,
  fieldKey: string,
): string {
  return `conflict-marker-${requireRecordId(recordId)}-${requireFieldKey(fieldKey)}`;
}

export function rowPresenceMarkerTestId(recordId: string): string {
  return `presence-row-${requireRecordId(recordId)}`;
}

export function cellPresenceMarkerTestId(
  recordId: string,
  fieldKey: string,
): string {
  return `presence-cell-${requireRecordId(recordId)}-${requireFieldKey(fieldKey)}`;
}

export function saveStateTestId(): string {
  return "save-state";
}

export function pendingQueueNoticeTestId(): string {
  return "pending-queue-notice";
}

export function pendingQueueCountTestId(): string {
  return "pending-queue-count";
}

export function referencePackAdminPanelTestId(): string {
  return "reference-pack-admin-panel";
}

export function referencePackFileInputTestId(): string {
  return "reference-pack-file";
}

export function referencePackImportButtonTestId(): string {
  return "reference-pack-import";
}

export function referencePackJobStatusTestId(): string {
  return "reference-pack-job-status";
}

export function referencePackReloadButtonTestId(): string {
  return "reference-pack-reload";
}

export function referencePackCancelButtonTestId(): string {
  return "reference-pack-cancel";
}

export function referencePackRefreshAllButtonTestId(): string {
  return "reference-pack-refresh-all";
}

export function referencePackRefreshSelectedButtonTestId(): string {
  return "reference-pack-refresh-selected";
}

export function referencePackRowTestId(
  packKey: string,
  packVersion: string,
): string {
  return `reference-pack-row-${encodeSelectorSegment(
    packKey,
    "pack_key",
  )}-${encodeSelectorSegment(packVersion, "pack_version")}`;
}

export function referencePackErrorTestId(): string {
  return "reference-pack-error";
}

export function gridSortHeaderTestId(
  viewSchemaId: WorkbookSurface,
  fieldKey: string,
): string {
  return `${requireViewSchemaId(viewSchemaId)}-sort-${requireFieldKey(fieldKey)}`;
}

export function gridFilterChipTestId(
  viewSchemaId: WorkbookSurface,
  fieldKey: string,
): string {
  return `${requireViewSchemaId(viewSchemaId)}-filter-chip-${requireFieldKey(fieldKey)}`;
}

export function gridFilterFieldTestId(viewSchemaId: WorkbookSurface): string {
  return `${requireViewSchemaId(viewSchemaId)}-filter-field`;
}

export function gridFilterValueTestId(viewSchemaId: WorkbookSurface): string {
  return `${requireViewSchemaId(viewSchemaId)}-filter-value`;
}

export function gridFilterApplyTestId(viewSchemaId: WorkbookSurface): string {
  return `${requireViewSchemaId(viewSchemaId)}-filter-apply`;
}

export function gridGroupingSelectTestId(
  viewSchemaId: WorkbookSurface,
): string {
  return `${requireViewSchemaId(viewSchemaId)}-group-by`;
}

export function gridGroupRowTestId(
  viewSchemaId: WorkbookSurface,
  fieldKey: string,
  value: string,
): string {
  return `${requireViewSchemaId(viewSchemaId)}-group-${requireFieldKey(fieldKey)}-${encodeSelectorSegment(value, "group value")}`;
}

export function gridRowTestId(
  viewSchemaId: WorkbookSurface,
  recordId: string,
): string {
  return `grid-row-${requireViewSchemaId(viewSchemaId)}-${requireRecordId(recordId)}`;
}

export function rowCellTestId(recordId: string, fieldKey: string): string {
  return `row-${requireRecordId(recordId)}-${requireFieldKey(fieldKey)}`;
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
  return options.surface === "inspector" ? `${base}-inspector` : base;
}

export function rowInspectorFieldTestId(
  recordId: string,
  fieldKey: string,
): string {
  return `${rowCellTestId(recordId, fieldKey)}-inspector`;
}

export function rowInspectButtonTestId(recordId: string): string {
  return `row-${requireRecordId(recordId)}-inspect`;
}

export function rowHistoryOpenButtonTestId(recordId: string): string {
  return `row-history-open-${requireRecordId(recordId)}`;
}

export function rowHistoryOpenInspectorButtonTestId(recordId: string): string {
  return `${rowHistoryOpenButtonTestId(recordId)}-inspector`;
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

export function draftCellTestId(fieldKey: string): string {
  return `draft-row-${requireFieldKey(fieldKey)}`;
}

export function draftRowCreateButtonTestId(): string {
  return "draft-row-create";
}

export function relationshipItemsTestId(
  recordId: string,
  fieldKey: string,
): string {
  return `row-${requireRecordId(recordId)}-${requireFieldKey(fieldKey)}-items`;
}

export function draftRelationshipItemsTestId(fieldKey: string): string {
  return `draft-row-${requireFieldKey(fieldKey)}-items`;
}

export function timelineCollectionInputTestId(
  recordId: string,
  fieldKey: string,
): string {
  return `row-${requireRecordId(recordId)}-${requireFieldKey(fieldKey)}-input`;
}

export function draftTimelineCollectionInputTestId(fieldKey: string): string {
  return `draft-row-${requireFieldKey(fieldKey)}-input`;
}

export function timelineRowVersionTestId(recordId: string): string {
  return rowCellTestId(recordId, "row_version");
}

export function timelineRowMarkReviewedButtonTestId(recordId: string): string {
  return `row-${requireRecordId(recordId)}-mark-reviewed`;
}

export function timelineRowReplacementInputTestId(recordId: string): string {
  return `row-${requireRecordId(recordId)}-replacement-id`;
}

export function timelineRowSupersedeButtonTestId(recordId: string): string {
  return `row-${requireRecordId(recordId)}-supersede`;
}

export function timelineEvidenceFileInputTestId(recordId: string): string {
  return `timeline-evidence-file-${requireRecordId(recordId)}`;
}

export function timelineDraftEvidenceFileInputTestId(): string {
  return "timeline-evidence-file-draft";
}

export function timelineEvidenceAttachSectionTestId(recordId: string): string {
  return `timeline-evidence-attach-${requireRecordId(recordId)}`;
}

export function timelineDraftEvidenceAttachSectionTestId(): string {
  return "timeline-evidence-attach-draft";
}

export function timelinePreviewRowTestId(recordId: string): string {
  return `timeline-preview-row-${requireRecordId(recordId)}`;
}

export function relationshipChipTestId(itemRef: string): string {
  return `chip-${requireItemRef(itemRef)}`;
}

export function mentionItemTestId(itemRef: string): string {
  return `mention-${requireItemRef(itemRef)}`;
}

export function autoResolutionNoticeTestId(itemRef: string): string {
  return `auto-resolution-notice-${requireItemRef(itemRef)}`;
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

export function entityMentionResolutionStatusTestId(value: string): string {
  return `entity-mention-resolution-status-${requireEntityMentionResolutionStatus(value)}`;
}

export function entityInspectButtonTestId(
  entityType: EntityType,
  recordId: string,
): string {
  return `inspect-${requireEntityType(entityType)}-${requireRecordId(recordId)}`;
}

export function entityInspectorTestId(entityType: EntityType): string {
  return `${requireEntityType(entityType)}-inspector`;
}

export function assessmentCreatePanelTestId(): string {
  return "assessment-create-panel";
}

export function evidencePreviewButtonTestId(recordId: string): string {
  return `evidence-preview-${requireRecordId(recordId)}`;
}

export function evidenceDownloadButtonTestId(recordId: string): string {
  return `evidence-download-${requireRecordId(recordId)}`;
}

export function evidenceAttachFileInputTestId(recordId: string): string {
  return `evidence-attach-file-${requireRecordId(recordId)}`;
}

export function evidenceAccessMessageTestId(recordId: string): string {
  return `evidence-access-message-${requireRecordId(recordId)}`;
}

export function evidencePreviewFrameTestId(recordId: string): string {
  return `evidence-preview-frame-${requireRecordId(recordId)}`;
}

export function genericCreateFieldTestId(fieldKey: string): string {
  return `generic-create-field-${requireFieldKey(fieldKey)}`;
}

export function genericCreateSubmitTestId(viewSchemaId: string): string {
  return `generic-create-submit-${requireViewSchemaId(viewSchemaId)}`;
}

export function genericEditRecordSelectTestId(viewSchemaId: string): string {
  return `generic-edit-record-${requireViewSchemaId(viewSchemaId)}`;
}

export function genericEditFieldSelectTestId(viewSchemaId: string): string {
  return `generic-edit-field-${requireViewSchemaId(viewSchemaId)}`;
}

export function genericEditActionSelectTestId(viewSchemaId: string): string {
  return `generic-edit-action-${requireViewSchemaId(viewSchemaId)}`;
}

export function genericEditValueTestId(viewSchemaId: string): string {
  return `generic-edit-value-${requireViewSchemaId(viewSchemaId)}`;
}

export function genericEditSubmitTestId(viewSchemaId: string): string {
  return `generic-edit-submit-${requireViewSchemaId(viewSchemaId)}`;
}

export function mentionResolveTargetSelectTestId(): string {
  return "inspector-resolve-target";
}

export function mentionResolveExistingButtonTestId(): string {
  return "inspector-resolve-existing";
}

export function mentionCreateEntityButtonTestId(
  entityType: EntityType,
): string {
  return `inspector-create-${requireEntityType(entityType)}`;
}

export function mentionDismissButtonTestId(): string {
  return "inspector-dismiss-mention";
}

export function mentionRestoreUnresolvedButtonTestId(): string {
  return "inspector-restore-unresolved";
}

export function dataTestIdSelector(testId: string): string {
  return `[data-testid="${cssAttributeValue(
    requireNonEmptySelectorValue(testId, "data-testid"),
  )}"]`;
}

function requireViewSchemaId(value: string): string {
  const token = requireNonEmptySelectorValue(value, "view_schema_id");
  if (
    !/^cartulary\.view\.[a-z][a-z0-9_]*(?:\.[a-z0-9_]+)*\.v[1-9][0-9]*$/u.test(
      token,
    )
  ) {
    throw new Error(`Invalid view_schema_id selector token: ${value}`);
  }
  if (!registeredViewSchemaIds.has(token)) {
    throw new Error(`Unknown view_schema_id selector token: ${value}`);
  }
  const encoded = encodeSelectorSegment(token, "view_schema_id");
  return encoded;
}

function requireFieldKey(value: string): string {
  const encoded = encodeSelectorSegment(value, "field_key");
  if (!/^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u.test(value)) {
    throw new Error(`Invalid field_key selector token: ${value}`);
  }
  return encoded;
}

function requireRecordId(value: string): string {
  return encodeSelectorSegment(value, "record_id");
}

function requireItemRef(value: string): string {
  return encodeSelectorSegment(value, "item_ref");
}

function encodeSelectorSegment(value: string, label: string): string {
  return encodeURIComponent(requireNonEmptySelectorValue(value, label));
}

function requireNonEmptySelectorValue(value: string, label: string): string {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`Invalid ${label} selector token: ${value}`);
  }
  return value;
}

function cssAttributeValue(value: string): string {
  return value
    .replace(/\\/gu, "\\\\")
    .replace(/\n/gu, "\\a ")
    .replace(/\r/gu, "\\d ")
    .replace(/\f/gu, "\\c ")
    .replace(/"/gu, '\\"');
}

function requireEntityType(value: EntityType): EntityType {
  if (value === "host" || value === "identity") {
    return value;
  }
  throw new Error(`Invalid entity type selector token: ${value}`);
}

function requireEntityMentionResolutionStatus(
  value: string,
): EntityMentionResolutionStatus {
  if (value === "unresolved" || value === "resolved" || value === "dismissed") {
    return value;
  }
  throw new Error(`Invalid entity_mentions.resolution_status token: ${value}`);
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
