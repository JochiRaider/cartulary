import type {
  RecordHistoryData,
  RecordHistoryItem,
} from "@cartulary/protocol-ts/http";
import {
  relationshipChipTestId,
  relationshipItemsTestId,
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryDestructiveCancelButtonTestId,
  rowHistoryDestructiveConfirmPanelTestId,
  rowHistoryMessageTestId,
  rowHistoryOpenButtonTestId,
  rowHistoryPanelTestId,
  rowHistoryRollbackCancelButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
  rowHistoryRollbackPreviewTestId,
  timelineInspectorSectionTestId,
  timelineInspectorTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorPanelTestId,
  workbookShellReadyTestId,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect } from "./fixtures";
import { openIncidentAsTrackedUserReady } from "./support/collaboration/replay";
import {
  collectionItems,
  hostRefsFieldKey,
  requireItemByRawText,
} from "./support/entities/mentions";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { navigateVisualApplication } from "./support/runtime/applicationReadiness";
import { uniqueEmail, uniqueTxn } from "./support/runtime/fixtureIdentity";
import { createInspectorReadingFixture } from "./support/timeline/inspectorReadingFixture";
import { installOwnedRoute } from "./support/transport/requestInterception";
import {
  assertViewportVisualRegression,
  assertVisualRegression,
} from "./support/visual/capture";
import { initializeVisualPage, test } from "./support/visual/fixtures";
import { clickTimelineRowAction } from "./support/visual/rowActions";
import { blurActiveElement } from "./support/visual/workbookCaptures";
import { normalizeWorkbookGridVisualState } from "./support/visual/workbookGeometry";
import {
  fetchRecordHistory,
  openHistoryEventDetails,
} from "./support/workbook/history";
import { patchRecord, queryViewRows } from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";

function visualHistoryActionTestId(
  item: RecordHistoryItem,
  action: RecordHistoryItem["available_rollback_actions"][number],
) {
  return rowHistoryActionTestId({
    action,
    historyItemRef: item.history_item_ref,
  });
}

function visualRollbackPreviewAnchor(
  item: RecordHistoryItem,
  action: RecordHistoryItem["available_rollback_actions"][number],
) {
  return {
    action,
    historyItemRef: item.history_item_ref,
  };
}

function requireVisualSourceHistoryEntryAction(history: RecordHistoryData) {
  const item =
    history.items.find(
      (candidate) =>
        candidate.available_rollback_actions.includes("history_entry") &&
        candidate.diff_summary.units.some((unit) =>
          unit.changes.some(
            (change) =>
              change.field_key === "timeline.raw_activity_text" &&
              change.before.state === "present",
          ),
        ) &&
        typeof candidate.history_entry_ref === "string" &&
        candidate.history_entry_ref.length > 0,
    ) ?? null;
  if (item === null) {
    throw new Error(
      "missing browser.inspector-history visual history_entry rollback item",
    );
  }
  return item;
}

test("Capture inspector Details, Relationships, Evidence, History, rollback preview, destructive confirmation, and public error fixtures.", async ({
  browserScope,
  page,
  sessionTracker,
}) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  const { incidentId, target, linkedTarget } =
    await createInspectorReadingFixture(page);
  const hostItem = requireItemByRawText(
    collectionItems(linkedTarget, hostRefsFieldKey),
    "browser.inspector-history visual host",
  );
  const narrowViewerPassword = "VisualInspectorNarrowViewer1!";
  const narrowViewer = await createIncidentMemberUser(page, incidentId, {
    display_name: "Visual Inspector Viewer",
    email: uniqueEmail("visual-inspector-narrow-viewer"),
    initial_password: narrowViewerPassword,
    role: "viewer",
    is_deployment_admin: false,
    mfa_required: false,
  });

  await navigateVisualApplication(
    page,
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      timelineViewSchemaId,
    )}`,
  );

  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await openTimelineInspector(page, target.record_id);
  for (const section of ["relationships", "evidence", "history"] as const) {
    await expect(
      page.getByTestId(timelineInspectorSectionTestId(section)),
    ).toBeVisible();
  }
  await expect(
    page.locator('[data-inspector-saved-field="timeline.raw_activity_text"]'),
  ).toContainText("browser.inspector-history visual inspector details");
  await expect(
    page
      .getByTestId(relationshipItemsTestId(target.record_id, hostRefsFieldKey))
      .getByTestId(relationshipChipTestId(String(hostItem.item_ref))),
  ).toBeVisible();
  await expect(
    page.getByTestId(timelineInspectorSectionTestId("evidence")),
  ).toContainText("Attached evidence count: 0");

  await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
    scroll: { top: 0, left: "left" },
  });
  const inspector = page.getByTestId(timelineInspectorTestId());
  const sectionsChooser = inspector.getByRole("button", {
    name: /^Sections:/,
    exact: true,
  });
  if (await sectionsChooser.isVisible()) await sectionsChooser.click();
  await inspector.getByRole("button", { name: "Details", exact: true }).click();
  await blurActiveElement(page);
  await assertViewportVisualRegression(page, "workbook-inspector-details");

  const field = page.locator(
    '[data-inspector-saved-field="timeline.activity_synopsis_text"]',
  );
  await field.locator("[data-inspector-edit-field]").click();
  const editor = field.getByRole("textbox");
  await editor.fill("Unfinished analyst summary");
  await field.scrollIntoViewIfNeeded();
  await assertViewportVisualRegression(
    page,
    "workbook-inspector-attached-edit",
  );
  await inspector
    .getByRole("button", { name: "Close editor", exact: true })
    .click();
  await expect(
    field.getByRole("button", {
      name: "Resume draft for Activity Synopsis",
      exact: true,
    }),
  ).toBeVisible();
  await field.scrollIntoViewIfNeeded();
  await blurActiveElement(page);
  await assertViewportVisualRegression(
    page,
    "workbook-inspector-retained-draft",
  );

  await blurActiveElement(page);
  await page
    .getByTestId(timelineInspectorSectionTestId("relationships"))
    .scrollIntoViewIfNeeded();
  await assertViewportVisualRegression(
    page,
    "workbook-inspector-relationships",
  );

  const narrowSession = await openIncidentAsTrackedUserReady(
    browserScope,
    sessionTracker,
    {
      preparePage: async (visualPage) => {
        await initializeVisualPage(visualPage, narrowViewer.user_id);
      },
      createdBy: "visual.inspector.narrow",
      email: narrowViewer.email,
      incidentId,
      password: narrowViewerPassword,
      purpose: "narrow inspector disabled-reason visual fixture",
      readyRecordId: target.record_id,
      userId: narrowViewer.user_id,
    },
  );
  {
    const narrowPage = narrowSession.page;
    await narrowPage.setViewportSize({ width: 1024, height: 720 });

    await openTimelineInspector(narrowPage, target.record_id);
    await expect(narrowPage.locator("[data-inspector-layout]")).toHaveAttribute(
      "data-inspector-layout",
      "right_overlay",
    );
    await expect(
      narrowPage.getByTestId(workbookShellSlotTestId("primary-grid")),
    ).toHaveAttribute("inert", "");
    const disabledAction = narrowPage.getByTestId(
      workbookInspectorFeatureActionTestId(
        timelineViewSchemaId,
        "create_related.lesson",
      ),
    );
    await expect(disabledAction).toBeDisabled();
    await expect(disabledAction).toHaveAttribute("aria-describedby", /.+/u);
    const disabledReason = narrowPage
      .getByText("Requires the editor incident role.")
      .last();
    await expect(disabledReason).toBeVisible();
    const recordMetadata = narrowPage.getByRole("region", {
      name: "Record technical metadata",
    });
    await recordMetadata.getByText("Technical details").click();
    const technicalRecordId = recordMetadata.getByText(target.record_id);
    await expect(technicalRecordId).toBeVisible();
    await technicalRecordId.scrollIntoViewIfNeeded();
    await expect(disabledReason).toBeInViewport();
    await expect(technicalRecordId).toBeInViewport();
    await assertViewportVisualRegression(
      narrowPage,
      "workbook-inspector-narrow-technical-details",
    );

    await browserScope.closeContext(narrowSession.page.context());
  }

  await page.setViewportSize({ width: 768, height: 640 });
  await expect(page.locator("[data-inspector-layout]")).toHaveAttribute(
    "data-inspector-layout",
    "full_overlay",
  );
  await expect(
    page.getByTestId(workbookShellSlotTestId("primary-grid")),
  ).toHaveAttribute("inert", "");
  const workflowPanel = page.getByTestId(
    workbookInspectorPanelTestId(timelineViewSchemaId, "workflow"),
  );
  const workflowActions = workflowPanel.locator(
    'button[data-route-kind="view_row_create"], button[data-feature-group-key="create_related.note"]',
  );
  await expect(workflowActions).toHaveCount(8);
  for (const action of await workflowActions.all()) {
    await action.scrollIntoViewIfNeeded();
    await expect(action).toBeVisible();
  }
  await assertViewportVisualRegression(
    page,
    "workbook-inspector-compact-actions",
  );
  await page.setViewportSize({ width: 1280, height: 720 });

  // The sparse reading/geometry fixture above stays unchanged. Give the
  // History/rollback captures an explicit source-edit event, rather than
  // selecting an arbitrary association event with runtime reference values.
  const current = (
    await queryViewRows(page, incidentId, timelineViewSchemaId)
  ).find((row) => row.record_id === target.record_id);
  if (!current) throw new Error("Missing sparse History fixture row");
  await patchRecord(page, target.record_id, {
    client_txn_id: uniqueTxn("VISUALINSPECTORHISTORY-SOURCE-EDIT"),
    view_schema_id: timelineViewSchemaId,
    base_row_version: current.row_version,
    changes: [
      {
        field_key: "timeline.raw_activity_text",
        value:
          "browser.inspector-history reviewed source excerpt\nObserved at 2026-04-18T14:12:34Z; reference 123e4567-e89b-42d3-a456-426614174000.",
      },
    ],
  });
  const history = await fetchRecordHistory(page, target.record_id);
  const rollbackItem = requireVisualSourceHistoryEntryAction(history);
  const rollbackAnchor = visualRollbackPreviewAnchor(
    rollbackItem,
    "history_entry",
  );

  await page
    .getByTestId(timelineInspectorSectionTestId("history"))
    .scrollIntoViewIfNeeded();
  await clickTimelineRowAction(
    page,
    target.record_id,
    rowHistoryOpenButtonTestId(target.record_id),
  );
  await expect(page.getByTestId(rowHistoryPanelTestId())).toBeVisible();
  await openHistoryEventDetails(page, rollbackItem.history_item_ref);
  await expect(
    page.getByTestId(visualHistoryActionTestId(rollbackItem, "history_entry")),
  ).toBeVisible();
  await assertViewportVisualRegression(page, "workbook-inspector-history", {
    anchor: {
      locator: page.getByTestId(rowHistoryPanelTestId()),
      align: "start",
      scrollportSelector: `aside[data-view-schema-id="${timelineViewSchemaId}"] [data-inspector-scroll-body]`,
    },
  });

  await openHistoryEventDetails(page, rollbackItem.history_item_ref);
  await page
    .getByTestId(visualHistoryActionTestId(rollbackItem, "history_entry"))
    .click();
  await expect(
    page.getByTestId(rowHistoryRollbackPreviewTestId(rollbackAnchor)),
  ).toContainText(rollbackItem.history_item_ref);
  await page
    .getByTestId(rowHistoryRollbackPreviewTestId(rollbackAnchor))
    .evaluate((element) => {
      element.classList.add("visual-row-history-rollback-preview");
    });
  await assertViewportVisualRegression(
    page,
    "workbook-inspector-rollback-preview",
    {
      anchor: {
        locator: page.getByTestId(
          rowHistoryRollbackPreviewTestId(rollbackAnchor),
        ),
        align: "center",
        scrollportSelector: `aside[data-view-schema-id="${timelineViewSchemaId}"] [data-inspector-scroll-body]`,
      },
      verifyFraming: async () => {
        await expect(
          page.getByTestId(
            rowHistoryRollbackCancelButtonTestId(rollbackAnchor),
          ),
        ).toBeFocused();
      },
    },
  );

  await page
    .getByTestId(rowHistoryRollbackCancelButtonTestId(rollbackAnchor))
    .click();
  await page.getByTestId(rowHistoryDeleteButtonTestId()).click();
  await expect(
    page.getByTestId(
      rowHistoryDestructiveConfirmPanelTestId({ operation: "delete" }),
    ),
  ).toContainText(target.record_id);
  const destructiveActions = page.getByTestId(
    rowHistoryDestructiveConfirmPanelTestId({ operation: "delete" }),
  );
  await destructiveActions.evaluate((element) => {
    element.setAttribute("data-design-fixture", "destructive-actions");
  });
  await destructiveActions.scrollIntoViewIfNeeded();
  await assertVisualRegression(
    page,
    "workbook-inspector-destructive-confirmation",
    destructiveActions,
  );

  await page
    .getByTestId(
      rowHistoryDestructiveCancelButtonTestId({ operation: "delete" }),
    )
    .click();
  await openHistoryEventDetails(page, rollbackItem.history_item_ref);
  await page
    .getByTestId(visualHistoryActionTestId(rollbackItem, "history_entry"))
    .click();
  await installOwnedRoute(
    page,
    `**/api/v1/records/${target.record_id}/rollback`,
    (route) =>
      route.fulfill({
        contentType: "application/json",
        status: 409,
        body: JSON.stringify({
          error: {
            status: 409,
            code: "row_version_conflict",
            message: "row_version_conflict",
            request_id: "visual-inspector-history-conflict",
            retryable: false,
            details: {
              reason_code: "stale_row_version",
            },
          },
        }),
      }),
  );
  await page
    .getByTestId(rowHistoryRollbackConfirmButtonTestId(rollbackAnchor))
    .click();
  await expect(page.getByTestId(rowHistoryMessageTestId())).toContainText(
    "This row changed; refresh it before retrying.",
  );
  await assertViewportVisualRegression(
    page,
    "workbook-inspector-public-error",
    {
      anchor: {
        locator: page.getByTestId(rowHistoryMessageTestId()),
        align: "start",
        scrollportSelector: `aside[data-view-schema-id="${timelineViewSchemaId}"] [data-inspector-scroll-body]`,
      },
    },
  );
});
