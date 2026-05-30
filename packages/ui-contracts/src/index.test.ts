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
  gridRowTestId,
  gridSavedRowSelector,
  gridSavedRowsSelector,
  gridScrollportClassName,
  gridScrollportSelector,
  gridShellTestId,
  gridSortHeaderTestId,
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
  rowHistoryItemTestId,
  rowHistoryLoadingTestId,
  rowHistoryMessageTestId,
  rowHistoryOpenSelectedButtonTestId,
  rowHistoryPanelTestId,
  rowHistoryRestoreButtonTestId,
  rowHistoryRollbackActions,
  rowInspectButtonTestId,
  rowInspectorFieldTestId,
  rowPresenceMarkerTestId,
  savedViewFamilySelector,
  saveStateTestId,
  surfaceTabTestId,
  systemViewSelectorTestId,
  timelineCollectionInputTestId,
  timelineDraftEvidenceAttachSectionTestId,
  timelineDraftEvidenceFileInputTestId,
  timelineEvidenceAttachSectionTestId,
  timelineEvidenceFileInputTestId,
  timelineMutationSubstrateReadyTestId,
  timelinePreviewRowTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowReplacementInputTestId,
  timelineRowSupersedeButtonTestId,
  timelineRowVersionTestId,
  timelineScalarEditorSurfaces,
  timelineScalarEditorTestId,
  workbookShellReadyTestId,
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

    expect(restoreItemId).toBe("row-history-item-hitem_change_set_2");
    expect(restoreActionId).toBe(
      "row-history-action-hitem_change_set_2-row_restore",
    );
    const reversedRestoreItem = requireFixtureValue(
      [...historyItems].reverse()[0],
      "reversed restore item",
    );
    expect(rowHistoryItemTestId(reversedRestoreItem)).toBe(restoreItemId);
    expect(rowHistoryActionTestId(reversedRestoreItem)).toBe(restoreActionId);
    expect(rowHistoryPanelTestId()).toBe("row-history-panel");
    expect(rowHistoryOpenSelectedButtonTestId()).toBe(
      "row-history-open-selected",
    );
    expect(rowHistoryLoadingTestId()).toBe("row-history-loading");
    expect(rowHistoryMessageTestId()).toBe("row-history-message");
    expect(rowHistoryDeleteButtonTestId()).toBe("row-history-delete");
    expect(rowHistoryRestoreButtonTestId()).toBe("row-history-restore");
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
  });

  it("provides shared builders for workbook action families", () => {
    expect(gridRowTestId("cartulary.view.hosts.v1", "host-1")).toBe(
      "grid-row-cartulary.view.hosts.v1-host-1",
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
    expect(rowHistoryItemTestId({ historyItemRef: "a:b" })).not.toBe(
      rowHistoryItemTestId({ historyItemRef: "a-b" }),
    );
    expect(rowCellTestId("record/1?x=y#z", "timeline.summary")).toBe(
      "row-record%2F1%3Fx%3Dy%23z-timeline.summary",
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
      "tab-bar",
      "system-views",
      "current-title",
      "view-bar",
      "primary-grid",
      "inspector",
      "status-strip",
      "presence",
    ]);
    expect(workbookShellSlotTestId("top-bar")).toBe(
      "workbook-shell-slot-top-bar",
    );
    expect(workbookShellSlotTestId("system-views")).toBe(
      "workbook-shell-slot-system-views",
    );
    expect(() => workbookShellSlotTestId("toolbar" as never)).toThrow(
      "Invalid workbook shell slot token: toolbar",
    );
    expect(timelineMutationSubstrateReadyTestId()).toBe(
      "timeline-mutation-substrate-ready",
    );
    expect(currentIncidentRoleTestId()).toBe("current-incident-role");
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
    expect(incidentMembershipRowTestId("user-2")).toBe(
      "incident-membership-row-user-2",
    );
    expect(incidentMembershipVersionTestId("user-2")).toBe(
      "incident-membership-version-user-2",
    );
    expect(incidentMembershipRoleInputTestId("user-2")).toBe(
      "incident-membership-role-input-user-2",
    );
    expect(incidentMembershipPatchButtonTestId("user-2")).toBe(
      "incident-membership-patch-user-2",
    );
    expect(incidentMembershipDeleteButtonTestId("user-2")).toBe(
      "incident-membership-delete-user-2",
    );
    expect(incidentMembershipRoleDisplayTestId("user-2")).toBe(
      "incident-membership-role-user-2",
    );
    expect(phase2IncidentRowTestId("incident-2")).toBe(
      "incident-row-incident-2",
    );
    expect(phase2SelectIncidentButtonTestId("incident-2")).toBe(
      "select-incident-incident-2",
    );
    expect(phase2MembershipRowTestId("user-2")).toBe("membership-row-user-2");
    expect(phase2MembershipRoleInputTestId("user-2")).toBe(
      "membership-role-input-user-2",
    );
    expect(phase2MembershipVersionTestId("user-2")).toBe(
      "membership-version-user-2",
    );
    expect(phase2MembershipPatchButtonTestId("user-2")).toBe(
      "patch-membership-user-2",
    );
    expect(phase2MembershipDeleteButtonTestId("user-2")).toBe(
      "delete-membership-user-2",
    );
    expect(extensionProfileRowTestId("profile:core")).toBe(
      "extension-profile%3Acore",
    );
  });

  it("provides stable Phase 1 bootstrap, landing, session, admin, and error selectors", () => {
    expect(phase1AuthTestId("shell")).toBe("auth-shell");
    expect(phase1AuthTestId("shell-message")).toBe("auth-shell-message");
    expect(phase1AuthTestId("status")).toBe("auth-status");
    expect(phase1AuthTestId("login-username")).toBe("auth-login-username");
    expect(phase1AuthTestId("login-password")).toBe("auth-login-password");
    expect(phase1AuthTestId("login-totp-code")).toBe("auth-login-totp-code");
    expect(phase1AuthTestId("login-submit")).toBe("auth-login-submit");
    expect(phase1AuthTestId("bootstrap-token")).toBe("auth-bootstrap-token");
    expect(phase1AuthTestId("bootstrap-enrollment-id")).toBe(
      "auth-bootstrap-enrollment-id",
    );
    expect(phase1AuthTestId("bootstrap-secret-base32")).toBe(
      "auth-bootstrap-secret-base32",
    );
    expect(phase1AuthTestId("bootstrap-begin")).toBe("auth-bootstrap-begin");
    expect(phase1AuthTestId("bootstrap-complete-code")).toBe(
      "auth-bootstrap-complete-code",
    );
    expect(phase1AuthTestId("bootstrap-complete")).toBe(
      "auth-bootstrap-complete",
    );

    expect(phase1LandingTestId("shell")).toBe("incident-landing");
    expect(phase1LandingTestId("current-user")).toBe("landing-current-user");
    expect(phase1LandingTestId("refresh")).toBe("landing-refresh");
    expect(phase1LandingTestId("incident-key")).toBe("landing-incident-key");
    expect(phase1LandingTestId("incident-title")).toBe(
      "landing-incident-title",
    );
    expect(phase1LandingTestId("create-button")).toBe("landing-create-button");
    expect(phase1LandingTestId("incidents-count")).toBe(
      "landing-incidents-count",
    );
    expect(phase1LandingTestId("loading")).toBe("landing-loading");
    expect(phase1LandingTestId("empty-state")).toBe("landing-empty-state");
    expect(phase1LandingTestId("incident-list")).toBe("landing-incident-list");
    expect(phase1LandingTestId("status")).toBe("landing-status");
    expect(phase1LandingTestId("return")).toBe("landing-return");

    expect(phase1RouteTestId("app-shell")).toBe("app-shell");
    expect(phase1RouteTestId("workbook-current-user")).toBe(
      "workbook-current-user",
    );
    expect(phase1RouteTestId("workbook-loading")).toBe("workbook-loading");
    expect(phase1RouteTestId("debug-harness-loading")).toBe(
      "debug-harness-loading",
    );

    expect(phase1AccountTestId("refresh-state")).toBe("account-refresh-state");
    expect(phase1AccountTestId("logout")).toBe("account-logout");
    expect(phase1AccountTestId("session-user-id")).toBe(
      "account-session-user-id",
    );
    expect(phase1AccountTestId("session-provider-type")).toBe(
      "account-session-provider-type",
    );
    expect(phase1AccountTestId("session-mfa-state")).toBe(
      "account-session-mfa-state",
    );
    expect(phase1AccountTestId("session-is-deployment-admin")).toBe(
      "account-session-is-deployment-admin",
    );
    expect(phase1AccountTestId("session-authenticated-at")).toBe(
      "account-session-authenticated-at",
    );
    expect(phase1AccountTestId("session-idle-expires-at")).toBe(
      "account-session-idle-expires-at",
    );
    expect(phase1AccountTestId("session-absolute-expires-at")).toBe(
      "account-session-absolute-expires-at",
    );
    expect(phase1AccountTestId("session-session-expires-at")).toBe(
      "account-session-session-expires-at",
    );
    expect(phase1AccountTestId("session-memberships")).toBe(
      "account-session-memberships",
    );
    expect(phase1AccountTestId("credential-auth-kind")).toBe(
      "account-credential-auth-kind",
    );
    expect(phase1AccountTestId("credential-recovery-model")).toBe(
      "account-credential-recovery-model",
    );
    expect(phase1AccountTestId("credential-password-changed-at")).toBe(
      "account-credential-password-changed-at",
    );
    expect(phase1AccountTestId("credential-totp-state")).toBe(
      "account-credential-totp-state",
    );
    expect(phase1AccountTestId("credential-pending-expires-at")).toBe(
      "account-credential-pending-expires-at",
    );
    expect(phase1AccountTestId("password-current")).toBe(
      "account-password-current",
    );
    expect(phase1AccountTestId("password-next")).toBe("account-password-next");
    expect(phase1AccountTestId("password-factor-code")).toBe(
      "account-password-factor-code",
    );
    expect(phase1AccountTestId("password-change")).toBe(
      "account-password-change",
    );
    expect(phase1AccountTestId("totp-current-password")).toBe(
      "account-totp-current-password",
    );
    expect(phase1AccountTestId("totp-current-factor")).toBe(
      "account-totp-current-factor",
    );
    expect(phase1AccountTestId("totp-begin")).toBe("account-totp-begin");
    expect(phase1AccountTestId("totp-enrollment-id")).toBe(
      "account-totp-enrollment-id",
    );
    expect(phase1AccountTestId("totp-secret-base32")).toBe(
      "account-totp-secret-base32",
    );
    expect(phase1AccountTestId("totp-complete-code")).toBe(
      "account-totp-complete-code",
    );
    expect(phase1AccountTestId("totp-complete")).toBe("account-totp-complete");
    expect(phase1AccountTestId("status")).toBe("account-status");

    expect(phase1AdminTestId("access-note")).toBe("admin-access-note");
    expect(phase1AdminTestId("create-email")).toBe("admin-create-email");
    expect(phase1AdminTestId("create-display-name")).toBe(
      "admin-create-display-name",
    );
    expect(phase1AdminTestId("create-password")).toBe("admin-create-password");
    expect(phase1AdminTestId("create-mfa-required")).toBe(
      "admin-create-mfa-required",
    );
    expect(phase1AdminTestId("create-is-deployment-admin")).toBe(
      "admin-create-is-deployment-admin",
    );
    expect(phase1AdminTestId("create-user")).toBe("admin-create-user");
    expect(phase1AdminTestId("target-user-id-input")).toBe(
      "admin-target-user-id-input",
    );
    expect(phase1AdminTestId("load-user")).toBe("admin-load-user");
    expect(phase1AdminTestId("target-user-id")).toBe("admin-target-user-id");
    expect(phase1AdminTestId("target-user-version")).toBe(
      "admin-target-user-version",
    );
    expect(phase1AdminTestId("target-is-active")).toBe(
      "admin-target-is-active",
    );
    expect(phase1AdminTestId("target-is-deployment-admin")).toBe(
      "admin-target-is-deployment-admin",
    );
    expect(phase1AdminTestId("patch-base-version")).toBe(
      "admin-patch-base-version",
    );
    expect(phase1AdminTestId("patch-display-name")).toBe(
      "admin-patch-display-name",
    );
    expect(phase1AdminTestId("patch-mfa-required")).toBe(
      "admin-patch-mfa-required",
    );
    expect(phase1AdminTestId("patch-is-active")).toBe("admin-patch-is-active");
    expect(phase1AdminTestId("patch-is-deployment-admin")).toBe(
      "admin-patch-is-deployment-admin",
    );
    expect(phase1AdminTestId("patch-user")).toBe("admin-patch-user");
    expect(phase1AdminTestId("new-password")).toBe("admin-new-password");
    expect(phase1AdminTestId("reason")).toBe("admin-reason");
    expect(phase1AdminTestId("password-reset")).toBe("admin-password-reset");
    expect(phase1AdminTestId("totp-reset")).toBe("admin-totp-reset");
    expect(phase1AdminTestId("revoke-all")).toBe("admin-revoke-all");
    expect(phase1AdminTestId("status")).toBe("admin-status");

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
