import { describe, expect, it } from "vitest";

import {
  assessmentCreateControlTestId,
  assessmentCreatePanelTestId,
  autoResolutionNoticeFamilySelector,
  autoResolutionNoticeTestId,
  autoResolutionReviewButtonTestId,
  autoResolutionUndoButtonTestId,
  cellPresenceMarkerTestId,
  conflictMarkerTestId,
  coordinationWorkflowTestId,
  draftCellTestId,
  draftRelationshipItemsTestId,
  draftRowCreateButtonTestId,
  draftTimelineCollectionInputTestId,
  entityInspectButtonTestId,
  entityInspectorTestId,
  entityMergeControlTestId,
  entityMergePreconditionDetailsTestId,
  entityReusableIdentifierItemTestId,
  entityReusableIdentifiersSectionTestId,
  evidenceAccessMessageTestId,
  evidenceAccessStateTestId,
  evidenceAttachFileInputTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  evidencePreviewPanelTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  genericEditActionSelectTestId,
  genericEditFieldSelectTestId,
  genericEditRecordSelectTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  genericWorkbookTestId,
  gridRowTestId,
  mentionCreateEntityButtonTestId,
  mentionDismissButtonTestId,
  mentionItemTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  mentionRestoreUnresolvedButtonTestId,
  pasteConflictItemTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  referencePackAdminPanelTestId,
  referencePackCancelButtonTestId,
  referencePackErrorTestId,
  referencePackFileInputTestId,
  referencePackImportButtonTestId,
  referencePackJobStatusTestId,
  referencePackListStatusTestId,
  referencePackRefreshAllButtonTestId,
  referencePackRefreshSelectedButtonTestId,
  referencePackReloadButtonTestId,
  referencePackRowTestId,
  relationshipChipTestId,
  relationshipItemsTestId,
  relationshipOverflowButtonTestId,
  rowCellTestId,
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryDestructiveCancelButtonTestId,
  rowHistoryDestructiveConfirmButtonTestId,
  rowHistoryDestructiveConfirmPanelTestId,
  rowHistoryItemTestId,
  rowHistoryLoadingTestId,
  rowHistoryMessageTestId,
  rowHistoryOpenButtonTestId,
  rowHistoryOpenInspectorButtonTestId,
  rowHistoryOpenSelectedButtonTestId,
  rowHistoryPanelTestId,
  rowHistoryRestoreButtonTestId,
  rowHistoryRollbackCancelButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
  rowHistoryRollbackPreviewTestId,
  rowInspectorFieldTestId,
  rowPresenceMarkerTestId,
  savedViewActionMenuTestId,
  savedViewActionMenuTriggerTestId,
  savedViewCreateButtonTestId,
  savedViewDeleteButtonTestId,
  savedViewDuplicateButtonTestId,
  savedViewFamilySelector,
  savedViewModifiedTestId,
  savedViewNameInputTestId,
  savedViewOptionTestId,
  savedViewResetButtonTestId,
  savedViewScopeSelectTestId,
  savedViewSelectorTestId,
  savedViewSetDefaultButtonTestId,
  savedViewSetHomeButtonTestId,
  savedViewStatusTestId,
  savedViewUpdateButtonTestId,
  saveStateActionButtonTestId,
  saveStateTestId,
  timelineCollectionInputTestId,
  timelineDraftEvidenceAttachSectionTestId,
  timelineDraftEvidenceFileInputTestId,
  timelineEvidenceAttachSectionTestId,
  timelineEvidenceFileInputTestId,
  timelinePreviewRowTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowReplacementInputTestId,
  timelineRowSupersedeButtonTestId,
  timelineRowVersionTestId,
  workbookActiveSurfaceFocusTargetTestId,
  workbookConflictControlTestId,
  workbookConflictLocalValueTestId,
  workbookConflictResolverTestId,
  workbookConflictSavedValueTestId,
  workbookConflictSummaryTestId,
  workbookEditRecoveryDiscardButtonTestId,
  workbookEditRecoveryRetryButtonTestId,
  workbookEditRecoveryTestId,
  workbookFocusAnchorTestId,
  workbookPresenceSummaryTestId,
} from "./index";

const requireFixtureValue = <T>(value: T | undefined, label: string): T => {
  if (value === undefined) {
    throw new Error(`Missing fixture value: ${label}`);
  }
  return value;
};

describe("@cartulary/ui-contracts workbook interaction selectors", () => {
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
    for (const entityType of ["host", "identity"] as const) {
      expect(mentionCreateEntityButtonTestId(entityType)).toBe(
        `inspector-create-${entityType}`,
      );
    }
    for (const action of [
      "change_set",
      "history_entry",
      "row_restore",
    ] as const) {
      expect(
        rowHistoryActionTestId({
          action,
          historyItemRef: "hitem_change_set_1",
        }),
      ).toContain(`-${action}`);
    }
    for (const operation of ["delete", "restore"] as const) {
      expect(rowHistoryDestructiveConfirmPanelTestId({ operation })).toBe(
        `row-history-destructive-confirm-${operation}`,
      );
    }
  });

  it("derives stable collaboration collaboration and status selectors", () => {
    expect(
      conflictMarkerTestId("record-1", "timeline.activity_synopsis_text"),
    ).toBe("conflict-marker-record-1-timeline.activity_synopsis_text");
    expect(workbookConflictResolverTestId()).toBe("workbook-conflict-resolver");
    expect(workbookConflictSummaryTestId()).toBe("workbook-conflict-summary");
    expect(workbookConflictSavedValueTestId()).toBe(
      "workbook-conflict-saved-value",
    );
    expect(workbookConflictLocalValueTestId()).toBe(
      "workbook-conflict-local-value",
    );
    expect(
      (
        [
          "activate-origin",
          "apply-collection",
          "close",
          "keep-saved",
          "merged-value",
          "paste-navigator",
          "paste-next",
          "paste-position",
          "paste-previous",
          "use-merged",
          "use-server-suggestion",
          "use-unsaved",
        ] as const
      ).map((control) => workbookConflictControlTestId(control)),
    ).toEqual([
      "conflict-activate-origin",
      "conflict-apply-collection",
      "conflict-close",
      "conflict-keep-saved",
      "conflict-merged-value",
      "paste-conflict-navigator",
      "paste-conflict-next",
      "paste-conflict-position",
      "paste-conflict-previous",
      "conflict-use-merged",
      "conflict-use-server-suggestion",
      "conflict-use-unsaved",
    ]);
    expect(() =>
      workbookConflictControlTestId("accept-server" as never),
    ).toThrow("Invalid workbook conflict control token: accept-server");
    expect(workbookEditRecoveryTestId()).toBe("workbook-edit-recovery");
    expect(workbookEditRecoveryRetryButtonTestId()).toBe(
      "workbook-edit-recovery-retry",
    );
    expect(workbookEditRecoveryDiscardButtonTestId()).toBe(
      "workbook-edit-recovery-discard",
    );
    expect(workbookActiveSurfaceFocusTargetTestId()).toBe(
      "workbook-active-surface-focus-target",
    );
    expect(rowPresenceMarkerTestId("record-1")).toBe("presence-row-record-1");
    expect(
      cellPresenceMarkerTestId("record-1", "timeline.activity_synopsis_text"),
    ).toBe("presence-cell-record-1-timeline.activity_synopsis_text");
    expect(saveStateTestId()).toBe("save-state");
    expect(saveStateActionButtonTestId()).toBe("save-state-action");
    expect(referencePackAdminPanelTestId()).toBe("reference-pack-admin-panel");
    expect(referencePackFileInputTestId()).toBe("reference-pack-file");
    expect(referencePackImportButtonTestId()).toBe("reference-pack-import");
    expect(referencePackJobStatusTestId()).toBe("reference-pack-job-status");
    expect(referencePackListStatusTestId()).toBe("reference-pack-list-status");
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
    expect(workbookFocusAnchorTestId()).toBe("workbook-focus-anchor");
    expect(workbookPresenceSummaryTestId()).toBe("presence-header");
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
    expect(savedViewActionMenuTriggerTestId("cartulary.view.hosts.v1")).toBe(
      "saved-view-action-menu-trigger-cartulary.view.hosts.v1",
    );
    expect(savedViewActionMenuTestId("cartulary.view.hosts.v1")).toBe(
      "saved-view-action-menu-cartulary.view.hosts.v1",
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
    expect(savedViewModifiedTestId("cartulary.view.hosts.v1")).toBe(
      "saved-view-modified-cartulary.view.hosts.v1",
    );
    expect(
      savedViewResetButtonTestId("cartulary.view.hosts.v1", "saved/view 1"),
    ).toBe("saved-view-reset-cartulary.view.hosts.v1-saved%2Fview%201");
    expect(savedViewStatusTestId("cartulary.view.hosts.v1")).toBe(
      "saved-view-status-cartulary.view.hosts.v1",
    );
  });

  it("provides shared builders for workbook action families", () => {
    expect(gridRowTestId("cartulary.view.hosts.v1", "host-1")).toBe(
      "grid-row-cartulary.view.hosts.v1-host-1",
    );
    expect(rowCellTestId("record-1", "timeline.activity_synopsis_text")).toBe(
      "row-record-1-timeline.activity_synopsis_text",
    );
    expect(
      rowInspectorFieldTestId("record-1", "timeline.raw_activity_text"),
    ).toBe("row-record-1-timeline.raw_activity_text-inspector");
    expect(draftCellTestId("timeline.activity_synopsis_text")).toBe(
      "draft-row-timeline.activity_synopsis_text",
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
    expect(
      relationshipOverflowButtonTestId("record-1", "timeline.host_refs"),
    ).toBe("row-record-1-timeline.host_refs-overflow");
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
    expect(
      (
        [
          "confirm",
          "loser-record",
          "message",
          "plan",
          "reason",
          "start",
        ] as const
      ).map((control) => entityMergeControlTestId(control)),
    ).toEqual([
      "merge-confirm",
      "merge-loser-record",
      "merge-message",
      "merge-plan",
      "merge-reason",
      "merge-start",
    ]);
    expect(() => entityMergeControlTestId("undo" as never)).toThrow(
      "Invalid entity merge control token: undo",
    );
    expect(assessmentCreatePanelTestId()).toBe("assessment-create-panel");
    expect(
      (
        [
          "assessed-at",
          "confidence-band",
          "message",
          "rationale",
          "state",
          "subject",
          "subject-type",
          "submit",
          "support-refs",
        ] as const
      ).map((control) => assessmentCreateControlTestId(control)),
    ).toEqual([
      "assessment-create-assessed-at",
      "assessment-create-confidence-band",
      "assessment-create-message",
      "assessment-create-rationale",
      "assessment-create-state",
      "assessment-create-subject",
      "assessment-create-subject-type",
      "assessment-create-submit",
      "assessment-create-support-refs",
    ]);
    expect(() => assessmentCreateControlTestId("delete" as never)).toThrow(
      "Invalid assessment create control token: delete",
    );
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
    for (const selector of [
      evidencePreviewButtonTestId,
      evidenceDownloadButtonTestId,
      evidenceAttachFileInputTestId,
      evidenceAccessMessageTestId,
      evidenceAccessStateTestId,
    ]) {
      expect(selector("evidence-1", "inspector")).not.toBe(
        selector("evidence-1", "row"),
      );
      expect(selector("evidence-1", "row")).toBe(selector("evidence-1"));
    }
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
    expect(
      (
        [
          "mutation-error",
          "note-source-record",
          "reference-load-error",
        ] as const
      ).map((selector) => genericWorkbookTestId(selector)),
    ).toEqual([
      "generic-mutation-error",
      "generic-create-note-source-record",
      "generic-reference-load-error",
    ]);
    expect(() => genericWorkbookTestId("delete-error" as never)).toThrow(
      "Invalid generic workbook selector token: delete-error",
    );
    expect(
      (
        [
          "decision-reason",
          "decision-replacement",
          "decision-submit",
          "decision-target",
          "party-clear-both",
          "party-clear-link",
          "party-clear-text",
          "party-create-from-text",
          "party-existing",
          "party-link-existing",
          "party-pair",
          "party-partial-completion",
          "party-retry-created-link",
          "task-blocked-reason",
          "task-status",
          "task-submit",
          "task-target",
        ] as const
      ).map((selector) => coordinationWorkflowTestId(selector)),
    ).toEqual([
      "decision-supersede-reason",
      "decision-supersede-replacement",
      "decision-supersede-submit",
      "decision-supersede-target",
      "party-link-clear-both",
      "party-link-clear-link",
      "party-link-clear-text",
      "party-link-create-from-text",
      "party-link-existing-party",
      "party-link-link-existing",
      "party-link-pair",
      "party-link-partial-completion",
      "party-link-retry-created",
      "task-lifecycle-blocked-reason",
      "task-lifecycle-status",
      "task-lifecycle-submit",
      "task-lifecycle-target",
    ]);
    expect(() => coordinationWorkflowTestId("party-delete" as never)).toThrow(
      "Invalid coordination workflow selector token: party-delete",
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

  it("builds exact history-open selectors from stable record identity", () => {
    expect(rowHistoryOpenButtonTestId("record/1")).toBe(
      "row-history-open-record%2F1",
    );
    expect(rowHistoryOpenInspectorButtonTestId("record/1")).toBe(
      "row-history-open-record%2F1-inspector",
    );
    expect(() => rowHistoryOpenButtonTestId("")).toThrow(
      "Invalid record_id selector token: ",
    );
  });

  it("characterizes draft and paste-conflict selectors", () => {
    expect(draftRowCreateButtonTestId()).toBe("draft-row-create");
    expect(draftCellTestId("timeline.activity_synopsis_text")).toBe(
      "draft-row-timeline.activity_synopsis_text",
    );
    expect(pasteConflictItemTestId("record/1:field key")).toBe(
      "paste-conflict-item-record%2F1%3Afield%20key",
    );
    expect(() => pasteConflictItemTestId(" ")).toThrow(
      "Invalid paste conflict key selector token:  ",
    );
  });

  it("characterizes reusable-entity and merge-precondition selectors", () => {
    expect(entityReusableIdentifiersSectionTestId("host", "host/1")).toBe(
      "host-reusable-identifiers-host%2F1",
    );
    expect(
      entityReusableIdentifierItemTestId("host", "host/1", "dns:name/1"),
    ).toBe("host-reusable-identifiers-host%2F1-dns%3Aname%2F1");
    expect(entityMergePreconditionDetailsTestId("identity", "identity/1")).toBe(
      "identity-merge-precondition-details-identity%2F1",
    );
    expect(() =>
      entityReusableIdentifiersSectionTestId("account" as never, "account-1"),
    ).toThrow("Invalid entity type selector token: account");
    expect(() =>
      entityReusableIdentifierItemTestId("host", "host-1", ""),
    ).toThrow("Invalid item_ref selector token: ");
  });
});
