import type {
  CollectionActionsV1,
  RecordHistoryData,
  RecordHistoryItem,
} from "@cartulary/protocol-ts/http";
import {
  dataTestIdSelector,
  mentionItemTestId,
  relationshipChipTestId,
  relationshipItemsTestId,
  rowCellTestId,
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryDestructiveCancelButtonTestId,
  rowHistoryDestructiveConfirmButtonTestId,
  rowHistoryDestructiveConfirmPanelTestId,
  rowHistoryMessageTestId,
  rowHistoryOpenButtonTestId,
  rowHistoryPanelTestId,
  rowHistoryRollbackCancelButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
  rowHistoryRollbackPreviewTestId,
  timelineInspectorSectionTestId,
  timelineInspectorTestId,
  timelineScalarEditorTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorPanelTestId,
  workbookInspectorToggleTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import { expectVisibleSemanticGridCellFocus } from "./support/accessibility/grid";
import {
  expectNoFocusTrap,
  expectVisibleFocus,
} from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import { expectAndRecordContrast } from "./support/accessibility/records";
import {
  expectAlertRole,
  expectNoPrivateDiagnostics,
} from "./support/accessibility/status";
import { openIncidentAsTrackedUserReady } from "./support/collaboration/replay";
import {
  collectionActionsPayload,
  collectionItems,
  hostRefsFieldKey,
  requireItemByRawText,
} from "./support/entities/mentions";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { installOwnedRoute } from "./support/transport/requestInterception";
import { expectCollectionControlPainted } from "./support/workbook/collections";
import {
  fetchRecordHistory,
  openHistoryEventDetails,
} from "./support/workbook/history";
import { mountedGridCell } from "./support/workbook/keyboardGrid";
import { createViewRow, patchRecord } from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";

function A11yHistoryActionTestId(
  item: RecordHistoryItem,
  action: RecordHistoryItem["available_rollback_actions"][number],
) {
  return rowHistoryActionTestId({
    action,
    historyItemRef: item.history_item_ref,
  });
}

function A11yRollbackAnchor(
  item: RecordHistoryItem,
  action: RecordHistoryItem["available_rollback_actions"][number],
) {
  return {
    action,
    historyItemRef: item.history_item_ref,
  };
}

function requireA11yHistoryEntryAction(history: RecordHistoryData) {
  const item =
    history.items.find(
      (candidate) =>
        candidate.available_rollback_actions.includes("history_entry") &&
        typeof candidate.history_entry_ref === "string" &&
        candidate.history_entry_ref.length > 0,
    ) ?? null;
  if (item === null) {
    throw new Error(
      "missing a11y.inspector-history history_entry rollback item",
    );
  }
  return item;
}

function A11yAttachedEvidencePayload(recordId: string): CollectionActionsV1 {
  return {
    kind: "collection_actions_v1",
    actions: [
      {
        op: "add_record_ref",
        linked_record_id: recordId,
      },
    ],
  };
}

test("a11y.inspector selection editing and destructive cancellation restore focus", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YINSPECTORCONFIG"),
    "a11y.inspector-history.row-02 config-driven inspector",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.inspector-history-02-row"),
    "timeline.raw_activity_text":
      "a11y.inspector-history.row-02 inspector details",
    "timeline.activity_synopsis_text":
      "a11y.inspector-history.row-02 selected row",
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  const toggle = page.getByTestId(
    workbookInspectorToggleTestId(timelineViewSchemaId),
  );
  await toggle.focus();
  await expectVisibleFocus(toggle);
  await toggle.press("Enter");
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveAttribute(
    "data-inspector-state",
    "no_row_selected",
  );
  await expect(
    page.getByText("Select a saved row to inspect its details."),
  ).toBeVisible();
  await expect(page.getByTestId(timelineInspectorTestId())).not.toContainText(
    "no_row_selected",
  );
  await page.keyboard.press("Escape");
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await expectVisibleFocus(toggle);
  const summaryCell = await mountedGridCell(
    page,
    timelineViewSchemaId,
    row.record_id,
    "timeline.activity_synopsis_text",
  );
  await expectVisibleSemanticGridCellFocus(summaryCell);
  await openTimelineInspector(page, row.record_id);
  await page
    .locator('[data-inspector-edit-field="timeline.raw_activity_text"]')
    .click();
  const detailsEditor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.raw_activity_text",
      recordId: row.record_id,
      surface: "inspector",
    }),
  );
  await expectVisibleFocus(detailsEditor);
  await expect(detailsEditor).toHaveValue(
    "a11y.inspector-history.row-02 inspector details",
  );
  await expectAndRecordContrast(
    page,
    [workbookInspectorPanelTestId(timelineViewSchemaId, "details")],
    "inspector-details",
  );
  await detailsEditor.press("Escape");
  const editAction = page.locator(
    '[data-inspector-edit-field="timeline.raw_activity_text"]',
  );
  await expect(editAction).toBeFocused();
  await editAction.press("Escape");
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await expectVisibleSemanticGridCellFocus(summaryCell);
  await page.keyboard.press("Shift+F10");
  const openHistory = page.getByTestId(
    rowHistoryOpenButtonTestId(row.record_id),
  );
  await expectVisibleFocus(openHistory);
  await expectAndRecordContrast(
    page,
    [rowHistoryOpenButtonTestId(row.record_id)],
    "history-menu",
  );
  await openHistory.press("Enter");
  await expect(page.getByTestId(rowHistoryPanelTestId())).toBeVisible();
  const deleteButton = page.getByTestId(rowHistoryDeleteButtonTestId());
  await expectVisibleFocus(deleteButton);
  await deleteButton.press("Enter");
  const deletePanel = page.getByTestId(
    rowHistoryDestructiveConfirmPanelTestId({ operation: "delete" }),
  );
  await expect(deletePanel).toHaveAttribute("role", "alertdialog");
  await expect(deletePanel).not.toHaveAttribute("aria-modal");
  await expect(deletePanel).toContainText(row.record_id);
  const deleteConfirm = page.getByTestId(
    rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
  );
  const deleteCancel = page.getByTestId(
    rowHistoryDestructiveCancelButtonTestId({ operation: "delete" }),
  );
  await expect(deleteCancel).toBeFocused();
  await expectVisibleFocus(deleteConfirm);
  await expectVisibleFocus(deleteCancel);
  await expectAndRecordContrast(page, [
    workbookInspectorToggleTestId(timelineViewSchemaId),
    rowCellTestId(row.record_id, "timeline.activity_synopsis_text"),
    rowHistoryDeleteButtonTestId(),
    rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
    rowHistoryDestructiveCancelButtonTestId({ operation: "delete" }),
  ]);
  await deleteCancel.press("Escape");
  await expect(deletePanel).toHaveCount(0);
  await expect(deleteButton).toBeFocused();
});

test("a11y.viewer inspector describes disabled actions and preserves responsive access", async ({
  browserScope,
  page,
  sessionTracker,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YINSPECTORCONFIG"),
    "a11y.inspector-history.row-02 config-driven inspector",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.inspector-history-02-row"),
    "timeline.raw_activity_text":
      "a11y.inspector-history.row-02 inspector details",
    "timeline.activity_synopsis_text":
      "a11y.inspector-history.row-02 selected row",
  });
  const viewerPassword = "A11yInspectorViewer1!";
  const viewer = await createIncidentMemberUser(page, incidentId, {
    display_name: "A11y Inspector Viewer",
    email: uniqueEmail("a11y-inspector-viewer"),
    initial_password: viewerPassword,
    role: "viewer",
    is_deployment_admin: false,
    mfa_required: false,
  });
  const viewerSession = await openIncidentAsTrackedUserReady(
    browserScope,
    sessionTracker,
    {
      createdBy: "a11y.inspector-history.row-02",
      email: viewer.email,
      incidentId,
      password: viewerPassword,
      purpose: "a11y inspector disabled reason viewer",
      readyRecordId: row.record_id,
      userId: viewer.user_id,
    },
  );
  {
    const viewerPage = viewerSession.page;
    await viewerPage.setViewportSize({ width: 1024, height: 720 });
    await openTimelineInspector(viewerPage, row.record_id);
    const disabledAction = viewerPage.getByTestId(
      workbookInspectorFeatureActionTestId(
        timelineViewSchemaId,
        "indicator.observations.manage",
      ),
    );
    await expect(disabledAction).toBeDisabled();
    const descriptionId = (
      await disabledAction.getAttribute("aria-describedby")
    )
      ?.split(" ")
      .at(-1);
    expect(descriptionId).not.toBeNull();
    await expect(viewerPage.locator(`[id="${descriptionId}"]`)).toHaveText(
      "Requires the editor incident role.",
    );

    await withPresentation(
      viewerPage,
      {
        textSpacing: `
              ${dataTestIdSelector(timelineInspectorTestId())} * {
                letter-spacing: 0.12em !important;
                line-height: 1.5 !important;
                word-spacing: 0.16em !important;
              }
              ${dataTestIdSelector(timelineInspectorTestId())} p {
                margin-block-end: 2em !important;
              }
            `,
      },
      async () => {
        const textSpacingGeometry = await viewerPage
          .getByTestId(timelineInspectorTestId())
          .evaluate((element) => ({
            clientWidth: element.clientWidth,
            scrollWidth: element.scrollWidth,
          }));
        expect(textSpacingGeometry.scrollWidth).toBeLessThanOrEqual(
          textSpacingGeometry.clientWidth + 1,
        );
      },
    );
    await withPresentation(viewerPage, { zoom: "200%" }, async () => {
      await expect(
        viewerPage.getByTestId(
          workbookInspectorCloseButtonTestId(timelineViewSchemaId),
        ),
      ).toBeVisible();
      const zoomGeometry = await viewerPage
        .getByTestId(timelineInspectorTestId())
        .evaluate((element) => ({
          clientWidth: element.clientWidth,
          scrollWidth: element.scrollWidth,
        }));
      expect(zoomGeometry.scrollWidth).toBeLessThanOrEqual(
        zoomGeometry.clientWidth + 1,
      );

      await browserScope.closeContext(viewerSession.page.context());
    });
  }
});

test("a11y.inspector-history.row-01 Verify inspector tabs, relationship links, evidence controls, history controls, rollback, destructive actions, and errors are keyboard reachable and announced.", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YINSPECTORHISTORY"),
    "a11y.inspector-history inspector actions",
  );
  const evidence = await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.inspector-history-evidence"),
    "evidence.collector_party_text": "a11y.inspector-history collector",
    "evidence.title": "a11y.inspector-history evidence",
  });
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    [hostRefsFieldKey]: collectionActionsPayload([
      "a11y.inspector-history host",
    ]),
    client_txn_id: uniqueTxn("a11y.inspector-history-row"),
    "timeline.raw_activity_text": "a11y.inspector-history inspector details",
    "timeline.activity_synopsis_text": "a11y.inspector-history selected row",
  });
  const linkedRow = await patchRecord(page, row.record_id, {
    base_row_version: row.row_version,
    changes: [
      {
        action_payload: A11yAttachedEvidencePayload(evidence.record_id),
        field_key: "timeline.attached_evidence_ids",
      },
    ],
    client_txn_id: uniqueTxn("a11y.inspector-history-link"),
    view_schema_id: timelineViewSchemaId,
  });
  const hostItem = requireItemByRawText(
    collectionItems(linkedRow, hostRefsFieldKey),
    "a11y.inspector-history host",
  );
  const history = await fetchRecordHistory(page, row.record_id);
  const rollbackItem = requireA11yHistoryEntryAction(history);
  const rollbackAnchor = A11yRollbackAnchor(rollbackItem, "history_entry");

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  const summaryCell = await mountedGridCell(
    page,
    timelineViewSchemaId,
    row.record_id,
    "timeline.activity_synopsis_text",
  );
  await expectVisibleSemanticGridCellFocus(summaryCell);
  await openTimelineInspector(page, row.record_id);

  for (const section of ["relationships", "evidence", "history"] as const) {
    await expect(
      page.getByTestId(timelineInspectorSectionTestId(section)),
    ).toBeVisible();
  }
  await page
    .locator('[data-inspector-edit-field="timeline.raw_activity_text"]')
    .click();
  const detailsEditor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.raw_activity_text",
      recordId: row.record_id,
      surface: "inspector",
    }),
  );
  await expectVisibleFocus(detailsEditor);
  await expect(detailsEditor).toHaveValue(
    "a11y.inspector-history inspector details",
  );

  const relationshipChip = page
    .getByTestId(relationshipItemsTestId(row.record_id, hostRefsFieldKey))
    .getByTestId(mentionItemTestId(String(hostItem.item_ref)));
  await expect(relationshipChip).toHaveAccessibleName(
    "Unresolved host mention: a11y.inspector-history host",
  );
  await relationshipChip.scrollIntoViewIfNeeded();
  await expectCollectionControlPainted(
    relationshipChip.getByText("?", { exact: true }),
  );
  await expectVisibleFocus(relationshipChip);
  await expectAndRecordContrast(
    page,
    [
      timelineScalarEditorTestId({
        fieldKey: "timeline.raw_activity_text",
        recordId: row.record_id,
        surface: "inspector",
      }),
      {
        id: "inspector-relationship-chip",
        selector: `${dataTestIdSelector(timelineInspectorTestId())} ${dataTestIdSelector(mentionItemTestId(String(hostItem.item_ref)))} ${dataTestIdSelector(relationshipChipTestId(String(hostItem.item_ref)))}`,
      },
    ],
    "inspector-details",
  );

  await expectVisibleFocus(detailsEditor);
  await detailsEditor.press("Escape");
  const editAction = page.locator(
    '[data-inspector-edit-field="timeline.raw_activity_text"]',
  );
  await expect(editAction).toBeFocused();
  await editAction.press("Escape");
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await expectVisibleSemanticGridCellFocus(summaryCell);
  await page.keyboard.press("Shift+F10");
  const openHistory = page.getByTestId(
    rowHistoryOpenButtonTestId(row.record_id),
  );
  await expectVisibleFocus(openHistory);
  await expectAndRecordContrast(
    page,
    [rowHistoryOpenButtonTestId(row.record_id)],
    "history-menu",
  );
  await openHistory.press("Enter");
  await expect(page.getByTestId(rowHistoryPanelTestId())).toBeVisible();
  await openHistoryEventDetails(page, rollbackItem.history_item_ref);
  const rollbackAction = page.getByTestId(
    A11yHistoryActionTestId(rollbackItem, "history_entry"),
  );
  await expectVisibleFocus(rollbackAction);
  await expectAndRecordContrast(
    page,
    [A11yHistoryActionTestId(rollbackItem, "history_entry")],
    "history-entry",
  );
  await rollbackAction.press("Enter");

  const rollbackPreview = page.getByTestId(
    rowHistoryRollbackPreviewTestId(rollbackAnchor),
  );
  await expect(rollbackPreview).toHaveAttribute("role", "alertdialog");
  await expect(rollbackPreview).not.toHaveAttribute("aria-modal");
  await expect(rollbackPreview).toContainText(/rollback/i);
  const rollbackCancel = page.getByTestId(
    rowHistoryRollbackCancelButtonTestId(rollbackAnchor),
  );
  const rollbackConfirm = page.getByTestId(
    rowHistoryRollbackConfirmButtonTestId(rollbackAnchor),
  );
  await expectVisibleFocus(rollbackCancel);
  await expectVisibleFocus(rollbackConfirm);
  await expectAndRecordContrast(
    page,
    [
      rowHistoryRollbackConfirmButtonTestId(rollbackAnchor),
      rowHistoryRollbackCancelButtonTestId(rollbackAnchor),
    ],
    "rollback-preview",
  );

  await installOwnedRoute(
    page,
    `**/api/v1/records/${row.record_id}/rollback`,
    (route) =>
      route.fulfill({
        contentType: "application/json",
        status: 409,
        body: JSON.stringify({
          error: {
            status: 409,
            code: "row_version_conflict",
            message: "row_version_conflict",
            request_id: "a11y-inspector-history-conflict",
            retryable: false,
            details: {
              reason_code: "stale_row_version",
            },
          },
        }),
      }),
  );
  await rollbackConfirm.press("Enter");
  const historyMessage = page.getByTestId(rowHistoryMessageTestId());
  await expectAlertRole(historyMessage);
  await expect(historyMessage).toHaveAttribute("aria-live", "assertive");
  await expect(historyMessage).toContainText(
    "This row changed; refresh it before retrying.",
  );
  await expect(historyMessage).toContainText("row_version_conflict");
  await expectNoPrivateDiagnostics(historyMessage);
  await expectAndRecordContrast(
    page,
    [rowHistoryMessageTestId()],
    "rollback-conflict",
  );

  const deleteButton = page.getByTestId(rowHistoryDeleteButtonTestId());
  await expectVisibleFocus(deleteButton);
  await deleteButton.press("Enter");
  const deletePanel = page.getByTestId(
    rowHistoryDestructiveConfirmPanelTestId({ operation: "delete" }),
  );
  await expect(deletePanel).toHaveAttribute("role", "alertdialog");
  await expect(deletePanel).not.toHaveAttribute("aria-modal");
  await expect(deletePanel).toContainText(row.record_id);
  const deleteConfirm = page.getByTestId(
    rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
  );
  const deleteCancel = page.getByTestId(
    rowHistoryDestructiveCancelButtonTestId({ operation: "delete" }),
  );
  await expectVisibleFocus(deleteConfirm);
  await expectVisibleFocus(deleteCancel);
  await expectAndRecordContrast(page, [
    rowCellTestId(row.record_id, "timeline.activity_synopsis_text"),
    rowHistoryDeleteButtonTestId(),
    rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
    rowHistoryDestructiveCancelButtonTestId({ operation: "delete" }),
  ]);
  await deleteCancel.press("Enter");
  await expect(deletePanel).toHaveCount(0);

  await expectAllInteractiveControlsNamed(page);
  await expectNoFocusTrap(page);
});
