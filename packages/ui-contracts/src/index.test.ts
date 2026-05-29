import { describe, expect, it } from "vitest";

import {
  autoResolutionNoticeTestId,
  autoResolutionReviewButtonTestId,
  autoResolutionUndoButtonTestId,
  cellPresenceMarkerTestId,
  conflictMarkerTestId,
  currentIncidentRoleTestId,
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
  gridGroupRowTestId,
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
  rowHistoryItemTestId,
  rowHistoryRollbackActions,
  rowInspectButtonTestId,
  rowInspectorFieldTestId,
  rowPresenceMarkerTestId,
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

  it("builds escaped data-testid selectors and timeline editor test ids", () => {
    expect(dataTestIdSelector('row-record"1')).toBe(
      '[data-testid="row-record\\"1"]',
    );
    expect(dataTestIdSelector("row\\record")).toBe(
      '[data-testid="row\\\\record"]',
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
