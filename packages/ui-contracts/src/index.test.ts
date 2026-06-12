import { describe, expect, it } from "vitest";

import {
  autoResolutionNoticeFamilySelector,
  autoResolutionNoticeTestId,
  autoResolutionReviewButtonTestId,
  autoResolutionUndoButtonTestId,
  cellPresenceMarkerTestId,
  conflictMarkerTestId,
  currentIncidentRoleTestId,
  dataTestIdPrefixSelector,
  dataTestIdSelector,
  draftCellTestId,
  draftRelationshipItemsTestId,
  draftTimelineCollectionInputTestId,
  entityInspectButtonTestId,
  entityInspectorTestId,
  entityMentionResolutionStatuses,
  entityMentionResolutionStatusTestId,
  entityTypes,
  evidenceAccessMessageTestId,
  evidenceAttachFileInputTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  evidencePreviewPanelTestId,
  extensionProfileRowTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  genericEditActionSelectTestId,
  genericEditFieldSelectTestId,
  genericEditRecordSelectTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  gridActionsHeaderTestId,
  gridDraftRowSelector,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridGroupingSelectTestId,
  gridGroupRowsSelector,
  gridGroupRowTestId,
  gridGroupRowTestIdPrefix,
  gridRowGutterTestId,
  gridRowTestId,
  gridSavedRowSelector,
  gridSavedRowsSelector,
  gridScrollportClassName,
  gridScrollportSelector,
  gridShellTestId,
  gridSortHeaderTestId,
  incidentControlsPanelTestId,
  incidentControlsTriggerTestId,
  incidentMembershipAdminNoteTestId,
  incidentMembershipCreateButtonTestId,
  incidentMembershipDeleteButtonTestId,
  incidentMembershipEmailInputTestId,
  incidentMembershipListTestId,
  incidentMembershipPatchButtonTestId,
  incidentMembershipRoleDisplayTestId,
  incidentMembershipRoleInputTestId,
  incidentMembershipRoleSelectTestId,
  incidentMembershipRowTestId,
  incidentMembershipVersionTestId,
  landingIncidentCardTestId,
  landingIncidentOpenButtonTestId,
  mentionCreateEntityButtonTestId,
  mentionDismissButtonTestId,
  mentionItemTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  mentionRestoreUnresolvedButtonTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  phase1AccountTestId,
  phase1AdminTestId,
  phase1AuthTestId,
  phase1ErrorCodeTestId,
  phase1ErrorSummaryTestIds,
  phase1LandingTestId,
  phase1RouteTestId,
  phase2IncidentRowTestId,
  phase2MembershipDeleteButtonTestId,
  phase2MembershipPatchButtonTestId,
  phase2MembershipRoleInputTestId,
  phase2MembershipRowTestId,
  phase2MembershipVersionTestId,
  phase2SelectIncidentButtonTestId,
  referencePackAdminPanelTestId,
  referencePackCancelButtonTestId,
  referencePackErrorTestId,
  referencePackFileInputTestId,
  referencePackImportButtonTestId,
  referencePackJobStatusTestId,
  referencePackRefreshAllButtonTestId,
  referencePackRefreshSelectedButtonTestId,
  referencePackReloadButtonTestId,
  referencePackRowTestId,
  relationshipChipTestId,
  relationshipItemsTestId,
  rowCellTestId,
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryDestructiveCancelButtonTestId,
  rowHistoryDestructiveConfirmButtonTestId,
  rowHistoryDestructiveConfirmPanelTestId,
  rowHistoryDestructiveOperations,
  rowHistoryItemTestId,
  rowHistoryLoadingTestId,
  rowHistoryMessageTestId,
  rowHistoryOpenSelectedButtonTestId,
  rowHistoryPanelTestId,
  rowHistoryRestoreButtonTestId,
  rowHistoryRollbackActions,
  rowHistoryRollbackCancelButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
  rowHistoryRollbackPreviewTestId,
  rowInspectButtonTestId,
  rowInspectorFieldTestId,
  rowPresenceMarkerTestId,
  savedViewCreateButtonTestId,
  savedViewDeleteButtonTestId,
  savedViewDuplicateButtonTestId,
  savedViewFamilySelector,
  savedViewNameInputTestId,
  savedViewOptionTestId,
  savedViewScopeSelectTestId,
  savedViewSelectorTestId,
  savedViewSetDefaultButtonTestId,
  savedViewSetHomeButtonTestId,
  savedViewStatusTestId,
  savedViewUpdateButtonTestId,
  saveStateTestId,
  statusStripQueueCountTestId,
  surfaceTabTestId,
  systemViewSelectorTestId,
  systemViewSwitcherGroupTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  timelineCollectionInputTestId,
  timelineDraftEvidenceAttachSectionTestId,
  timelineDraftEvidenceFileInputTestId,
  timelineEvidenceAttachSectionTestId,
  timelineEvidenceFileInputTestId,
  timelineInspectorSections,
  timelineInspectorSectionTestId,
  timelineMutationSubstrateReadyTestId,
  timelinePreviewRowTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowReplacementInputTestId,
  timelineRowSupersedeButtonTestId,
  timelineRowVersionTestId,
  timelineScalarEditorSurfaces,
  timelineScalarEditorTestId,
  workbookShellReadyTestId,
  workbookShellSlotLabel,
  workbookShellSlots,
  workbookShellSlotTestId,
} from "./index";

const requireFixtureValue = <T>(value: T | undefined, label: string): T => {
  if (value === undefined) {
    throw new Error(`Missing fixture value: ${label}`);
  }
  return value;
};

describe("@cartulary/ui-contracts workbook row selectors", () => {
  it("returns deterministic selectors for identical stable identifiers", () => {
    const first = {
      cell: rowCellTestId("record-1", "timeline.summary"),
      shell: gridShellTestId("cartulary.view.timeline.v1"),
      tab: surfaceTabTestId("cartulary.view.timeline.v1"),
    };
    const second = {
      cell: rowCellTestId("record-1", "timeline.summary"),
      shell: gridShellTestId("cartulary.view.timeline.v1"),
      tab: surfaceTabTestId("cartulary.view.timeline.v1"),
    };

    expect(second).toEqual(first);
  });

  it("derives surface selectors from view_schema_id instead of visible titles", () => {
    const originalSurface = {
      title: "Timeline",
      viewSchemaId: "cartulary.view.timeline.v1",
    };
    const renamedSurface = {
      title: "Activity",
      viewSchemaId: "cartulary.view.timeline.v1",
    };

    expect(surfaceTabTestId(originalSurface.viewSchemaId)).toBe(
      "surface-tab-cartulary.view.timeline.v1",
    );
    expect(surfaceTabTestId(renamedSurface.viewSchemaId)).toBe(
      surfaceTabTestId(originalSurface.viewSchemaId),
    );
    expect(gridShellTestId(renamedSurface.viewSchemaId)).toBe(
      gridShellTestId(originalSurface.viewSchemaId),
    );
    expect(systemViewSelectorTestId()).toBe("system-view-selector");
  });

  it("builds stable System views switcher selectors from closed groups and view_schema_id", () => {
    const originalSurface = {
      title: "Indicators",
      viewSchemaId: "cartulary.view.indicators.v1",
    };
    const renamedSurface = {
      title: "Observable Signals",
      viewSchemaId: "cartulary.view.indicators.v1",
    };

    expect(systemViewSelectorTestId()).toBe(systemViewSwitcherTriggerTestId());
    expect(systemViewSwitcherTriggerTestId()).toBe("system-view-selector");
    expect(systemViewSwitcherMenuTestId()).toBe("system-view-switcher-menu");
    expect(systemViewSwitcherGroupTestId("scope-assessment")).toBe(
      "system-view-switcher-group-scope-assessment",
    );
    expect(
      systemViewSwitcherOptionTestId(
        "scope-assessment",
        originalSurface.viewSchemaId,
      ),
    ).toBe(
      "system-view-switcher-option-scope-assessment-cartulary.view.indicators.v1",
    );
    expect(
      systemViewSwitcherOptionTestId(
        "scope-assessment",
        renamedSurface.viewSchemaId,
      ),
    ).toBe(
      systemViewSwitcherOptionTestId(
        "scope-assessment",
        originalSurface.viewSchemaId,
      ),
    );
    expect(systemViewSwitcherGroupTestId("coordination")).toBe(
      "system-view-switcher-group-coordination",
    );
    expect(systemViewSwitcherGroupTestId("review-learning")).toBe(
      "system-view-switcher-group-review-learning",
    );
    expect(systemViewSwitcherGroupTestId("optional-artifact-surfaces")).toBe(
      "system-view-switcher-group-optional-artifact-surfaces",
    );
  });

  it("fails closed for System views switcher group and option selector inputs", () => {
    expect(() => systemViewSwitcherGroupTestId("future" as never)).toThrow(
      "Invalid system view switcher group token: future",
    );
    expect(() =>
      systemViewSwitcherOptionTestId("scope-assessment", "timeline"),
    ).toThrow("Invalid view_schema_id selector token: timeline");
    expect(() =>
      systemViewSwitcherOptionTestId(
        "scope-assessment",
        "cartulary.view.future.v1",
      ),
    ).toThrow(
      "Unknown view_schema_id selector token: cartulary.view.future.v1",
    );
  });

  it("keeps tab, field, column, row, and fixture reordering out of selector identity", () => {
    const surfaces = [
      { title: "Timeline", viewSchemaId: "cartulary.view.timeline.v1" },
      { title: "Evidence", viewSchemaId: "cartulary.view.evidence.v1" },
    ];
    const fields = [
      { fieldKey: "timeline.summary", label: "Summary" },
      { fieldKey: "timeline.details", label: "Details" },
    ];
    const rows = [
      { displayName: "Alpha", recordId: "record-alpha" },
      { displayName: "Beta", recordId: "record-beta" },
    ];
    const timelineSurface = requireFixtureValue(surfaces[0], "timeline");
    const evidenceSurface = requireFixtureValue(surfaces[1], "evidence");
    const summaryField = requireFixtureValue(fields[0], "summary");
    const detailsField = requireFixtureValue(fields[1], "details");
    const betaRow = requireFixtureValue(rows[1], "beta row");

    const selectors = {
      betaCell: rowCellTestId(betaRow.recordId, summaryField.fieldKey),
      detailsHeader: gridSortHeaderTestId(
        timelineSurface.viewSchemaId,
        detailsField.fieldKey,
      ),
      evidenceTab: surfaceTabTestId(evidenceSurface.viewSchemaId),
    };

    const reorderedSurfaces = [...surfaces].reverse();
    const reorderedFields = [...fields].reverse();
    const reorderedRows = [...rows].reverse();

    expect(
      surfaceTabTestId(
        reorderedSurfaces.find((surface) => surface.title === "Evidence")
          ?.viewSchemaId ?? "",
      ),
    ).toBe(selectors.evidenceTab);
    expect(
      gridSortHeaderTestId(
        timelineSurface.viewSchemaId,
        reorderedFields.find((field) => field.label === "Details")?.fieldKey ??
          "",
      ),
    ).toBe(selectors.detailsHeader);
    expect(
      rowCellTestId(
        reorderedRows.find((row) => row.displayName === "Beta")?.recordId ?? "",
        summaryField.fieldKey,
      ),
    ).toBe(selectors.betaCell);
  });

  it("keeps visible field labels and display names out of cell selectors", () => {
    const originalField = {
      fieldKey: "timeline.summary",
      label: "Summary",
    };
    const renamedField = {
      fieldKey: "timeline.summary",
      label: "Executive summary",
    };
    const originalRow = {
      displayName: "Alpha workstation",
      recordId: "record-alpha",
    };
    const renamedRow = {
      displayName: "Renamed workstation",
      recordId: "record-alpha",
    };

    expect(rowCellTestId(originalRow.recordId, originalField.fieldKey)).toBe(
      rowCellTestId(renamedRow.recordId, renamedField.fieldKey),
    );
    expect(gridFilterChipTestId("cartulary.view.timeline.v1", "status")).toBe(
      gridFilterChipTestId("cartulary.view.timeline.v1", "status"),
    );
    expect(gridFilterFieldTestId("cartulary.view.timeline.v1")).toBe(
      "cartulary.view.timeline.v1-filter-field",
    );
    expect(gridGroupingSelectTestId("cartulary.view.timeline.v1")).toBe(
      "cartulary.view.timeline.v1-group-by",
    );
    expect(
      gridGroupRowTestId(
        "cartulary.view.timeline.v1",
        "timeline.capture_state",
        "rough",
      ),
    ).toBe("cartulary.view.timeline.v1-group-timeline.capture_state-rough");
    expect(
      gridGroupRowTestIdPrefix(
        "cartulary.view.timeline.v1",
        "timeline.capture_state",
      ),
    ).toBe("cartulary.view.timeline.v1-group-timeline.capture_state-");
    expect(
      gridGroupRowsSelector(
        "cartulary.view.timeline.v1",
        "timeline.capture_state",
      ),
    ).toBe(
      '[data-testid^="cartulary.view.timeline.v1-group-timeline.capture_state-"]',
    );
  });

  it("derives row and row-action selectors from record_id", () => {
    expect(gridSavedRowSelector("record-alpha")).toBe(
      '[role="row"][data-grid-record-id="record-alpha"]',
    );
    expect(rowInspectButtonTestId("record-alpha")).toBe(
      "row-record-alpha-inspect",
    );
  });

  it("derives item-ref selectors without using item display text or fixture order", () => {
    const item = {
      displayText: "WS-023?",
      itemRef: "mention:host:1",
    };
    const renamedItem = {
      displayText: "Workstation 023",
      itemRef: "mention:host:1",
    };

    expect(relationshipChipTestId(item.itemRef)).toBe(
      "chip-mention%3Ahost%3A1",
    );
    expect(relationshipChipTestId(renamedItem.itemRef)).toBe(
      relationshipChipTestId(item.itemRef),
    );
    expect(mentionItemTestId(renamedItem.itemRef)).toBe(
      mentionItemTestId(item.itemRef),
    );
    expect(autoResolutionNoticeTestId(renamedItem.itemRef)).toBe(
      autoResolutionNoticeTestId(item.itemRef),
    );
    expect(autoResolutionNoticeFamilySelector()).toBe(
      '[data-testid^="auto-resolution-notice-"]',
    );
  });

  it("derives row-history items and actions from stable history identity rather than render index", () => {
    const historyItems = [
      {
        action: "change_set" as const,
        historyItemRef: "hitem_change_set_1",
      },
      {
        action: "row_restore" as const,
        historyItemRef: "hitem_change_set_2",
      },
    ];
    const restoreItem = requireFixtureValue(historyItems[1], "restore item");
    const restoreItemId = rowHistoryItemTestId(restoreItem);
    const restoreActionId = rowHistoryActionTestId(restoreItem);
    const restorePreviewId = rowHistoryRollbackPreviewTestId(restoreItem);

    expect(restoreItemId).toBe("row-history-item-hitem_change_set_2");
    expect(restoreActionId).toBe(
      "row-history-action-hitem_change_set_2-row_restore",
    );
    expect(restorePreviewId).toBe(
      "row-history-rollback-preview-hitem_change_set_2-row_restore",
    );
    expect(rowHistoryRollbackConfirmButtonTestId(restoreItem)).toBe(
      "row-history-rollback-preview-hitem_change_set_2-row_restore-confirm",
    );
    expect(rowHistoryRollbackCancelButtonTestId(restoreItem)).toBe(
      "row-history-rollback-preview-hitem_change_set_2-row_restore-cancel",
    );
    const reversedRestoreItem = requireFixtureValue(
      [...historyItems].reverse()[0],
      "reversed restore item",
    );
    expect(rowHistoryItemTestId(reversedRestoreItem)).toBe(restoreItemId);
    expect(rowHistoryActionTestId(reversedRestoreItem)).toBe(restoreActionId);
    expect(rowHistoryRollbackPreviewTestId(reversedRestoreItem)).toBe(
      restorePreviewId,
    );
    expect(rowHistoryPanelTestId()).toBe("row-history-panel");
    expect(rowHistoryOpenSelectedButtonTestId()).toBe(
      "row-history-open-selected",
    );
    expect(rowHistoryLoadingTestId()).toBe("row-history-loading");
    expect(rowHistoryMessageTestId()).toBe("row-history-message");
    expect(rowHistoryDeleteButtonTestId()).toBe("row-history-delete");
    expect(rowHistoryRestoreButtonTestId()).toBe("row-history-restore");
    expect(
      rowHistoryDestructiveConfirmPanelTestId({ operation: "delete" }),
    ).toBe("row-history-destructive-confirm-delete");
    expect(
      rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
    ).toBe("row-history-destructive-confirm-delete-confirm");
    expect(
      rowHistoryDestructiveCancelButtonTestId({ operation: "restore" }),
    ).toBe("row-history-destructive-confirm-restore-cancel");
  });

  it("validates closed-vocabulary-derived selector tokens", () => {
    for (const status of entityMentionResolutionStatuses) {
      expect(entityMentionResolutionStatusTestId(status)).toBe(
        `entity-mention-resolution-status-${status}`,
      );
    }
    for (const entityType of entityTypes) {
      expect(mentionCreateEntityButtonTestId(entityType)).toBe(
        `inspector-create-${entityType}`,
      );
    }
    for (const action of rowHistoryRollbackActions) {
      expect(
        rowHistoryActionTestId({
          action,
          historyItemRef: "hitem_change_set_1",
        }),
      ).toContain(`-${action}`);
    }
    for (const operation of rowHistoryDestructiveOperations) {
      expect(rowHistoryDestructiveConfirmPanelTestId({ operation })).toBe(
        `row-history-destructive-confirm-${operation}`,
      );
    }
    expect(() =>
      entityMentionResolutionStatusTestId("Resolved" as never),
    ).toThrow("Invalid entity_mentions.resolution_status token: Resolved");
  });

  it("derives inspector field ids from the stable row cell id", () => {
    expect(rowInspectorFieldTestId("record-1", "timeline.details")).toBe(
      "row-record-1-timeline.details-inspector",
    );
  });

  it("targets saved and draft workbook rows when scoped through the grid shell", () => {
    expect(gridShellTestId("cartulary.view.timeline.v1")).toBe(
      "cartulary.view.timeline.v1-grid-shell",
    );
    expect(gridScrollportClassName()).toBe("cartulary-grid-scrollport");
    expect(gridScrollportSelector()).toBe(".cartulary-grid-scrollport");
    expect(gridActionsHeaderTestId("cartulary.view.timeline.v1")).toBe(
      "cartulary.view.timeline.v1-actions-header",
    );
    expect(gridRowGutterTestId("cartulary.view.timeline.v1", "record-1")).toBe(
      "cartulary.view.timeline.v1-row-gutter-record-1",
    );
    expect(gridSavedRowsSelector()).toBe(
      '[role="row"][data-grid-record-id]:not([data-grid-record-id=""])',
    );
    expect(gridDraftRowSelector()).toBe('[role="row"][data-grid-record-id=""]');
  });

  it("derives stable Phase 6 collaboration and status selectors", () => {
    expect(conflictMarkerTestId("record-1", "timeline.summary")).toBe(
      "conflict-marker-record-1-timeline.summary",
    );
    expect(rowPresenceMarkerTestId("record-1")).toBe("presence-row-record-1");
    expect(cellPresenceMarkerTestId("record-1", "timeline.summary")).toBe(
      "presence-cell-record-1-timeline.summary",
    );
    expect(saveStateTestId()).toBe("save-state");
    expect(statusStripQueueCountTestId()).toBe("status-strip-queue-count");
    expect(referencePackAdminPanelTestId()).toBe("reference-pack-admin-panel");
    expect(referencePackFileInputTestId()).toBe("reference-pack-file");
    expect(referencePackImportButtonTestId()).toBe("reference-pack-import");
    expect(referencePackJobStatusTestId()).toBe("reference-pack-job-status");
    expect(referencePackReloadButtonTestId()).toBe("reference-pack-reload");
    expect(referencePackCancelButtonTestId()).toBe("reference-pack-cancel");
    expect(referencePackRefreshAllButtonTestId()).toBe(
      "reference-pack-refresh-all",
    );
    expect(referencePackRefreshSelectedButtonTestId()).toBe(
      "reference-pack-refresh-selected",
    );
    expect(referencePackErrorTestId()).toBe("reference-pack-error");
    expect(referencePackRowTestId("type_registry.host", "1")).toBe(
      "reference-pack-row-type_registry.host-1",
    );
    expect(referencePackRowTestId("pack/key", "v 1")).toBe(
      "reference-pack-row-pack%2Fkey-v%201",
    );
    expect(pendingQueueNoticeTestId()).toBe("pending-queue-notice");
    expect(pendingQueueCountTestId()).toBe("pending-queue-count");
    expect(savedViewFamilySelector()).toBe('[data-testid^="saved-view-"]');
    expect(savedViewSelectorTestId("cartulary.view.hosts.v1")).toBe(
      "saved-view-selector-cartulary.view.hosts.v1",
    );
    expect(
      savedViewOptionTestId(
        "cartulary.view.hosts.v1",
        "11111111-1111-4111-8111-111111111111",
      ),
    ).toBe(
      "saved-view-option-cartulary.view.hosts.v1-11111111-1111-4111-8111-111111111111",
    );
    expect(
      savedViewOptionTestId("cartulary.view.hosts.v1", "saved/view 1"),
    ).toBe("saved-view-option-cartulary.view.hosts.v1-saved%2Fview%201");
    expect(savedViewNameInputTestId("cartulary.view.hosts.v1")).toBe(
      "saved-view-name-cartulary.view.hosts.v1",
    );
    expect(savedViewScopeSelectTestId("cartulary.view.hosts.v1")).toBe(
      "saved-view-scope-cartulary.view.hosts.v1",
    );
    expect(savedViewCreateButtonTestId("cartulary.view.hosts.v1")).toBe(
      "saved-view-create-cartulary.view.hosts.v1",
    );
    expect(
      savedViewDuplicateButtonTestId("cartulary.view.hosts.v1", "saved/view 1"),
    ).toBe("saved-view-duplicate-cartulary.view.hosts.v1-saved%2Fview%201");
    expect(
      savedViewUpdateButtonTestId("cartulary.view.hosts.v1", "saved/view 1"),
    ).toBe("saved-view-update-cartulary.view.hosts.v1-saved%2Fview%201");
    expect(
      savedViewDeleteButtonTestId("cartulary.view.hosts.v1", "saved/view 1"),
    ).toBe("saved-view-delete-cartulary.view.hosts.v1-saved%2Fview%201");
    expect(savedViewSetHomeButtonTestId("cartulary.view.hosts.v1")).toBe(
      "saved-view-set-home-cartulary.view.hosts.v1",
    );
    expect(savedViewSetDefaultButtonTestId("cartulary.view.hosts.v1")).toBe(
      "saved-view-set-default-cartulary.view.hosts.v1",
    );
    expect(savedViewStatusTestId("cartulary.view.hosts.v1")).toBe(
      "saved-view-status-cartulary.view.hosts.v1",
    );
  });

  it("provides shared builders for workbook action families", () => {
    expect(gridRowTestId("cartulary.view.hosts.v1", "host-1")).toBe(
      "grid-row-cartulary.view.hosts.v1-host-1",
    );
    expect(rowCellTestId("record-1", "timeline.summary")).toBe(
      "row-record-1-timeline.summary",
    );
    expect(rowInspectorFieldTestId("record-1", "timeline.details")).toBe(
      "row-record-1-timeline.details-inspector",
    );
    expect(draftCellTestId("timeline.summary")).toBe(
      "draft-row-timeline.summary",
    );
    expect(timelineRowVersionTestId("record-1")).toBe(
      "row-record-1-row_version",
    );
    expect(timelineRowMarkReviewedButtonTestId("record-1")).toBe(
      "row-record-1-mark-reviewed",
    );
    expect(timelineRowReplacementInputTestId("record-1")).toBe(
      "row-record-1-replacement-id",
    );
    expect(timelineRowSupersedeButtonTestId("record-1")).toBe(
      "row-record-1-supersede",
    );
    expect(timelineEvidenceFileInputTestId("record-1")).toBe(
      "timeline-evidence-file-record-1",
    );
    expect(timelineDraftEvidenceFileInputTestId()).toBe(
      "timeline-evidence-file-draft",
    );
    expect(timelineEvidenceAttachSectionTestId("record-1")).toBe(
      "timeline-evidence-attach-record-1",
    );
    expect(timelineDraftEvidenceAttachSectionTestId()).toBe(
      "timeline-evidence-attach-draft",
    );
    expect(timelinePreviewRowTestId("record-1")).toBe(
      "timeline-preview-row-record-1",
    );
    expect(relationshipItemsTestId("record-1", "timeline.host_refs")).toBe(
      "row-record-1-timeline.host_refs-items",
    );
    expect(draftRelationshipItemsTestId("timeline.host_refs")).toBe(
      "draft-row-timeline.host_refs-items",
    );
    expect(
      timelineCollectionInputTestId("record-1", "timeline.host_refs"),
    ).toBe("row-record-1-timeline.host_refs-input");
    expect(draftTimelineCollectionInputTestId("timeline.host_refs")).toBe(
      "draft-row-timeline.host_refs-input",
    );
    expect(entityInspectButtonTestId("host", "host-1")).toBe(
      "inspect-host-host-1",
    );
    expect(entityInspectorTestId("identity")).toBe("identity-inspector");
    expect(evidencePreviewButtonTestId("evidence-1")).toBe(
      "evidence-preview-evidence-1",
    );
    expect(evidenceDownloadButtonTestId("evidence-1")).toBe(
      "evidence-download-evidence-1",
    );
    expect(evidenceAttachFileInputTestId("evidence-1")).toBe(
      "evidence-attach-file-evidence-1",
    );
    expect(evidenceAccessMessageTestId("evidence-1")).toBe(
      "evidence-access-message-evidence-1",
    );
    expect(evidencePreviewFrameTestId("evidence-1")).toBe(
      "evidence-preview-frame-evidence-1",
    );
    expect(evidencePreviewPanelTestId()).toBe("evidence-preview-panel");
    expect(genericCreateFieldTestId("note.title")).toBe(
      "generic-create-field-note.title",
    );
    expect(genericCreateSubmitTestId("cartulary.view.notes.v1")).toBe(
      "generic-create-submit-cartulary.view.notes.v1",
    );
    expect(genericEditRecordSelectTestId("cartulary.view.notes.v1")).toBe(
      "generic-edit-record-cartulary.view.notes.v1",
    );
    expect(genericEditFieldSelectTestId("cartulary.view.notes.v1")).toBe(
      "generic-edit-field-cartulary.view.notes.v1",
    );
    expect(genericEditActionSelectTestId("cartulary.view.notes.v1")).toBe(
      "generic-edit-action-cartulary.view.notes.v1",
    );
    expect(genericEditValueTestId("cartulary.view.notes.v1")).toBe(
      "generic-edit-value-cartulary.view.notes.v1",
    );
    expect(genericEditSubmitTestId("cartulary.view.notes.v1")).toBe(
      "generic-edit-submit-cartulary.view.notes.v1",
    );
    expect(mentionResolveTargetSelectTestId()).toBe("inspector-resolve-target");
    expect(mentionResolveExistingButtonTestId()).toBe(
      "inspector-resolve-existing",
    );
    expect(mentionCreateEntityButtonTestId("identity")).toBe(
      "inspector-create-identity",
    );
    expect(mentionDismissButtonTestId()).toBe("inspector-dismiss-mention");
    expect(mentionRestoreUnresolvedButtonTestId()).toBe(
      "inspector-restore-unresolved",
    );
    expect(autoResolutionUndoButtonTestId("mention:host:1")).toBe(
      "auto-resolution-notice-mention%3Ahost%3A1-undo",
    );
    expect(autoResolutionReviewButtonTestId("mention:host:1")).toBe(
      "auto-resolution-notice-mention%3Ahost%3A1-review",
    );
  });

  it("rejects malformed stable IDs and keeps selector segments collision-resistant", () => {
    expect(() => gridShellTestId("timeline")).toThrow(
      "Invalid view_schema_id selector token: timeline",
    );
    expect(() => gridShellTestId("")).toThrow(
      "Invalid view_schema_id selector token: ",
    );
    expect(() => rowCellTestId("record-1", "timeline summary")).toThrow(
      "Invalid field_key selector token: timeline summary",
    );
    expect(() => rowCellTestId("record-1", "timeline..summary")).toThrow(
      "Invalid field_key selector token: timeline..summary",
    );
    expect(() => rowCellTestId("", "timeline.summary")).toThrow(
      "Invalid record_id selector token: ",
    );
    expect(() => rowCellTestId("   ", "timeline.summary")).toThrow(
      "Invalid record_id selector token:    ",
    );
    expect(() =>
      rowHistoryActionTestId({
        action: "restore" as never,
        historyItemRef: "hitem_change_set_1",
      }),
    ).toThrow("Invalid row history rollback action token: restore");
    expect(() => mentionCreateEntityButtonTestId("account" as never)).toThrow(
      "Invalid entity type selector token: account",
    );
    expect(() =>
      rowHistoryActionTestId({
        action: "rollback" as never,
        historyItemRef: "hitem_change_set_1",
      }),
    ).toThrow("Invalid row history rollback action token: rollback");
    expect(() =>
      rowHistoryDestructiveConfirmPanelTestId({
        operation: "remove" as never,
      }),
    ).toThrow("Invalid row history destructive operation token: remove");
    expect(() => rowHistoryItemTestId({ historyItemRef: "" })).toThrow(
      "Invalid history_item_ref selector token: ",
    );
    expect(() => rowHistoryItemTestId({ historyItemRef: "   " })).toThrow(
      "Invalid history_item_ref selector token:    ",
    );

    expect(rowHistoryItemTestId({ historyItemRef: "h:item/1?x=y#z" })).toBe(
      "row-history-item-h%3Aitem%2F1%3Fx%3Dy%23z",
    );
    expect(
      rowHistoryActionTestId({
        action: "history_entry",
        historyItemRef: "h:item/1?x=y#z",
      }),
    ).toBe("row-history-action-h%3Aitem%2F1%3Fx%3Dy%23z-history_entry");
    expect(
      rowHistoryRollbackPreviewTestId({
        action: "history_entry",
        historyItemRef: "h:item/1?x=y#z",
      }),
    ).toBe(
      "row-history-rollback-preview-h%3Aitem%2F1%3Fx%3Dy%23z-history_entry",
    );
    expect(rowHistoryItemTestId({ historyItemRef: "a:b" })).not.toBe(
      rowHistoryItemTestId({ historyItemRef: "a-b" }),
    );
    expect(rowCellTestId("record/1?x=y#z", "timeline.summary")).toBe(
      "row-record%2F1%3Fx%3Dy%23z-timeline.summary",
    );
    expect(conflictMarkerTestId("record/1?x=y#z", "timeline.summary")).toBe(
      "conflict-marker-record%2F1%3Fx%3Dy%23z-timeline.summary",
    );
    expect(cellPresenceMarkerTestId("record/1?x=y#z", "timeline.summary")).toBe(
      "presence-cell-record%2F1%3Fx%3Dy%23z-timeline.summary",
    );

    expect(relationshipChipTestId("a:b")).not.toBe(
      relationshipChipTestId("a-b"),
    );
    expect(relationshipChipTestId("a b")).not.toBe(
      relationshipChipTestId("a-b"),
    );
  });

  it("validates view_schema_id selectors against the generated registry", () => {
    for (const viewSchemaId of [
      "cartulary.view.timeline.v1",
      "cartulary.view.comm_log.v1",
      "cartulary.view.findings.v1",
      "cartulary.view.forensic_keywords.v1",
    ]) {
      expect(gridShellTestId(viewSchemaId)).toBe(`${viewSchemaId}-grid-shell`);
    }

    expect(() => gridShellTestId("cartulary.view.future.v1")).toThrow(
      "Unknown view_schema_id selector token: cartulary.view.future.v1",
    );
  });

  it("provides shared builders for app shell and incident membership selectors", () => {
    expect(workbookShellReadyTestId()).toBe("workbook-shell-ready");
    expect(workbookShellSlots).toEqual([
      "top-bar",
      "view-bar",
      "primary-grid",
      "inspector",
      "status-strip",
    ]);
    expect(workbookShellSlotTestId("top-bar")).toBe(
      "workbook-shell-slot-top-bar",
    );
    expect(workbookShellSlotTestId("view-bar")).toBe(
      "workbook-shell-slot-view-bar",
    );
    expect(workbookShellSlotLabel("top-bar")).toBe("Workbook top bar");
    expect(workbookShellSlotLabel("primary-grid")).toBe("Primary grid");
    expect(() => workbookShellSlotTestId("toolbar" as never)).toThrow(
      "Invalid workbook shell slot token: toolbar",
    );
    expect(() => workbookShellSlotLabel("toolbar" as never)).toThrow(
      "Invalid workbook shell slot token: toolbar",
    );
    expect(timelineMutationSubstrateReadyTestId()).toBe(
      "timeline-mutation-substrate-ready",
    );
    expect(currentIncidentRoleTestId()).toBe("current-incident-role");
    expect(incidentControlsTriggerTestId()).toBe("incident-controls-trigger");
    expect(incidentControlsPanelTestId()).toBe("incident-controls-panel");
    expect(timelineInspectorSections).toEqual([
      "details",
      "relationships",
      "evidence",
      "history",
    ]);
    expect(timelineInspectorSectionTestId("details")).toBe(
      "timeline-inspector-section-details",
    );
    expect(timelineInspectorSectionTestId("relationships")).toBe(
      "timeline-inspector-section-relationships",
    );
    expect(timelineInspectorSectionTestId("evidence")).toBe(
      "timeline-inspector-section-evidence",
    );
    expect(timelineInspectorSectionTestId("history")).toBe(
      "timeline-inspector-section-history",
    );
    expect(timelineInspectorSectionTestId("details")).toBe(
      timelineInspectorSectionTestId("details"),
    );
    expect(() => timelineInspectorSectionTestId("summary" as never)).toThrow(
      "Invalid timeline inspector section token: summary",
    );
    expect(landingIncidentCardTestId("incident-1")).toBe(
      "landing-incident-incident-1",
    );
    expect(landingIncidentOpenButtonTestId("incident-1")).toBe(
      "landing-open-incident-1",
    );
    expect(landingIncidentCardTestId("incident:1")).toBe(
      "landing-incident-incident%3A1",
    );
    expect(incidentMembershipEmailInputTestId()).toBe(
      "incident-membership-email",
    );
    expect(incidentMembershipRoleSelectTestId()).toBe(
      "incident-membership-role",
    );
    expect(incidentMembershipCreateButtonTestId()).toBe(
      "incident-membership-create",
    );
    expect(incidentMembershipAdminNoteTestId()).toBe(
      "incident-membership-admin-note",
    );
    expect(incidentMembershipListTestId()).toBe("incident-membership-list");
    expectSelectorCases<
      "delete" | "patch" | "roleDisplay" | "roleInput" | "row" | "version"
    >(
      (control) => {
        const testIdFor = {
          row: incidentMembershipRowTestId,
          version: incidentMembershipVersionTestId,
          roleInput: incidentMembershipRoleInputTestId,
          patch: incidentMembershipPatchButtonTestId,
          delete: incidentMembershipDeleteButtonTestId,
          roleDisplay: incidentMembershipRoleDisplayTestId,
        }[control];
        return testIdFor("user-2");
      },
      [
        ["row", "incident-membership-row-user-2"],
        ["version", "incident-membership-version-user-2"],
        ["roleInput", "incident-membership-role-input-user-2"],
        ["patch", "incident-membership-patch-user-2"],
        ["delete", "incident-membership-delete-user-2"],
        ["roleDisplay", "incident-membership-role-user-2"],
      ],
    );
    expect(phase2IncidentRowTestId("incident-2")).toBe(
      "incident-row-incident-2",
    );
    expect(phase2SelectIncidentButtonTestId("incident-2")).toBe(
      "select-incident-incident-2",
    );
    expectSelectorCases<"delete" | "patch" | "roleInput" | "row" | "version">(
      (control) => {
        const testIdFor = {
          row: phase2MembershipRowTestId,
          roleInput: phase2MembershipRoleInputTestId,
          version: phase2MembershipVersionTestId,
          patch: phase2MembershipPatchButtonTestId,
          delete: phase2MembershipDeleteButtonTestId,
        }[control];
        return testIdFor("user-2");
      },
      [
        ["row", "membership-row-user-2"],
        ["roleInput", "membership-role-input-user-2"],
        ["version", "membership-version-user-2"],
        ["patch", "patch-membership-user-2"],
        ["delete", "delete-membership-user-2"],
      ],
    );
    expect(extensionProfileRowTestId("profile:core")).toBe(
      "extension-profile%3Acore",
    );
  });

  it("provides stable Phase 1 bootstrap, landing, session, admin, and error selectors", () => {
    expectSelectorCases(phase1AuthTestId, [
      ["shell", "auth-shell"],
      ["shell-message", "auth-shell-message"],
      ["status", "auth-status"],
      ["login-username", "auth-login-username"],
      ["login-password", "auth-login-password"],
      ["login-totp-code", "auth-login-totp-code"],
      ["login-submit", "auth-login-submit"],
      ["bootstrap-token", "auth-bootstrap-token"],
      ["bootstrap-enrollment-id", "auth-bootstrap-enrollment-id"],
      ["bootstrap-secret-base32", "auth-bootstrap-secret-base32"],
      ["bootstrap-begin", "auth-bootstrap-begin"],
      ["bootstrap-complete-code", "auth-bootstrap-complete-code"],
      ["bootstrap-complete", "auth-bootstrap-complete"],
    ]);

    expectSelectorCases(phase1LandingTestId, [
      ["shell", "incident-landing"],
      ["current-user", "landing-current-user"],
      ["refresh", "landing-refresh"],
      ["incident-key", "landing-incident-key"],
      ["incident-title", "landing-incident-title"],
      ["create-button", "landing-create-button"],
      ["incidents-count", "landing-incidents-count"],
      ["loading", "landing-loading"],
      ["empty-state", "landing-empty-state"],
      ["incident-list", "landing-incident-list"],
      ["status", "landing-status"],
      ["return", "landing-return"],
    ]);

    expectSelectorCases(phase1RouteTestId, [
      ["app-shell", "app-shell"],
      ["workbook-current-user", "workbook-current-user"],
      ["workbook-loading", "workbook-loading"],
      ["debug-harness-loading", "debug-harness-loading"],
    ]);

    expectSelectorCases(phase1AccountTestId, [
      ["refresh-state", "account-refresh-state"],
      ["logout", "account-logout"],
      ["session-user-id", "account-session-user-id"],
      ["session-provider-type", "account-session-provider-type"],
      ["session-mfa-state", "account-session-mfa-state"],
      ["session-is-deployment-admin", "account-session-is-deployment-admin"],
      ["session-authenticated-at", "account-session-authenticated-at"],
      ["session-idle-expires-at", "account-session-idle-expires-at"],
      ["session-absolute-expires-at", "account-session-absolute-expires-at"],
      ["session-session-expires-at", "account-session-session-expires-at"],
      ["session-memberships", "account-session-memberships"],
      ["credential-auth-kind", "account-credential-auth-kind"],
      ["credential-recovery-model", "account-credential-recovery-model"],
      [
        "credential-password-changed-at",
        "account-credential-password-changed-at",
      ],
      ["credential-totp-state", "account-credential-totp-state"],
      [
        "credential-pending-expires-at",
        "account-credential-pending-expires-at",
      ],
      ["password-current", "account-password-current"],
      ["password-next", "account-password-next"],
      ["password-factor-code", "account-password-factor-code"],
      ["password-change", "account-password-change"],
      ["totp-current-password", "account-totp-current-password"],
      ["totp-current-factor", "account-totp-current-factor"],
      ["totp-begin", "account-totp-begin"],
      ["totp-enrollment-id", "account-totp-enrollment-id"],
      ["totp-secret-base32", "account-totp-secret-base32"],
      ["totp-complete-code", "account-totp-complete-code"],
      ["totp-complete", "account-totp-complete"],
      ["status", "account-status"],
    ]);

    expectSelectorCases(phase1AdminTestId, [
      ["access-note", "admin-access-note"],
      ["create-email", "admin-create-email"],
      ["create-display-name", "admin-create-display-name"],
      ["create-password", "admin-create-password"],
      ["create-mfa-required", "admin-create-mfa-required"],
      ["create-is-deployment-admin", "admin-create-is-deployment-admin"],
      ["create-user", "admin-create-user"],
      ["target-user-id-input", "admin-target-user-id-input"],
      ["load-user", "admin-load-user"],
      ["target-user-id", "admin-target-user-id"],
      ["target-user-version", "admin-target-user-version"],
      ["target-is-active", "admin-target-is-active"],
      ["target-is-deployment-admin", "admin-target-is-deployment-admin"],
      ["patch-base-version", "admin-patch-base-version"],
      ["patch-display-name", "admin-patch-display-name"],
      ["patch-mfa-required", "admin-patch-mfa-required"],
      ["patch-is-active", "admin-patch-is-active"],
      ["patch-is-deployment-admin", "admin-patch-is-deployment-admin"],
      ["patch-user", "admin-patch-user"],
      ["new-password", "admin-new-password"],
      ["reason", "admin-reason"],
      ["password-reset", "admin-password-reset"],
      ["totp-reset", "admin-totp-reset"],
      ["revoke-all", "admin-revoke-all"],
      ["status", "admin-status"],
    ]);

    expect(phase1ErrorCodeTestId("auth")).toBe("auth-error-code");
    expect(phase1ErrorSummaryTestIds("auth")).toEqual({
      container: "auth-error-public",
      details: "auth-error-details",
      message: "auth-error-message",
    });
    expect(phase1ErrorCodeTestId("account")).toBe("account-error-code");
    expect(phase1ErrorSummaryTestIds("account")).toEqual({
      container: "account-error-public",
      details: "account-error-details",
      message: "account-error-message",
    });
    expect(phase1ErrorCodeTestId("admin")).toBe("admin-error-code");
    expect(phase1ErrorSummaryTestIds("admin")).toEqual({
      container: "admin-error-public",
      details: "admin-error-details",
      message: "admin-error-message",
    });
    expect(phase1ErrorCodeTestId("landing")).toBe("landing-error-code");
    expect(phase1ErrorSummaryTestIds("landing")).toEqual({
      container: "landing-error-public",
      details: "landing-error-details",
      message: "landing-error-message",
    });
  });

  it("keeps Phase 1 selector identity on semantic state and stable field identifiers", () => {
    const renamedSession = {
      displayLabel: "Current operator",
      field: "session-user-id" as const,
    };
    const relabeledSession = {
      displayLabel: "Signed-in user",
      field: "session-user-id" as const,
    };
    const authControls = [
      { field: "login-password" as const, label: "Password" },
      { field: "login-username" as const, label: "Email address" },
    ];

    expect(phase1AccountTestId(renamedSession.field)).toBe(
      phase1AccountTestId(relabeledSession.field),
    );
    expect(
      phase1AuthTestId(
        authControls.find((control) => control.label === "Email address")
          ?.field ?? "login-password",
      ),
    ).toBe("auth-login-username");
    expect(
      phase1AuthTestId(
        [...authControls]
          .reverse()
          .find((control) => control.field === "login-username")?.field ??
          "login-password",
      ),
    ).toBe("auth-login-username");
    expect(landingIncidentCardTestId("incident:stable")).toBe(
      "landing-incident-incident%3Astable",
    );
  });

  it("rejects invalid Phase 1 selector vocabularies", () => {
    expect(() => phase1AuthTestId("username" as never)).toThrow(
      "Invalid phase1 auth selector token: username",
    );
    expect(() => phase1AccountTestId("user-id" as never)).toThrow(
      "Invalid phase1 account selector token: user-id",
    );
    expect(() => phase1AdminTestId("target-user" as never)).toThrow(
      "Invalid phase1 admin selector token: target-user",
    );
    expect(() => phase1LandingTestId("incident-card" as never)).toThrow(
      "Invalid phase1 landing selector token: incident-card",
    );
    expect(() => phase1RouteTestId("shell" as never)).toThrow(
      "Invalid phase1 route selector token: shell",
    );
    expect(() => phase1ErrorCodeTestId("session" as never)).toThrow(
      "Invalid phase1 error surface selector token: session",
    );
  });

  it("builds escaped data-testid selectors and timeline editor test ids", () => {
    expect(dataTestIdSelector('row-record"1')).toBe(
      '[data-testid="row-record\\"1"]',
    );
    expect(dataTestIdSelector("row\\record")).toBe(
      '[data-testid="row\\\\record"]',
    );
    expect(
      dataTestIdSelector(rowHistoryItemTestId({ historyItemRef: 'h"1\\2' })),
    ).toBe('[data-testid="row-history-item-h%221%5C2"]');
    expect(dataTestIdPrefixSelector("saved-view-")).toBe(
      '[data-testid^="saved-view-"]',
    );
    expect(() => dataTestIdSelector("")).toThrow(
      "Invalid data-testid selector token: ",
    );

    for (const surface of timelineScalarEditorSurfaces) {
      expect(
        timelineScalarEditorTestId({
          fieldKey: "timeline.summary",
          recordId: "record-1",
          surface,
        }),
      ).toBe(
        surface === "inspector"
          ? "row-record-1-timeline.summary-inspector"
          : "row-record-1-timeline.summary",
      );
    }
    expect(
      timelineScalarEditorTestId({
        fieldKey: "timeline.summary",
        recordId: null,
      }),
    ).toBe("draft-row-timeline.summary");
  });
});

function expectSelectorCases<T extends string>(
  testIdFor: (selector: T) => string,
  cases: ReadonlyArray<readonly [T, string]>,
): void {
  for (const [selector, expected] of cases) {
    expect(testIdFor(selector)).toBe(expected);
  }
}
