import {
  gridShellTestId,
  pendingReplayCountAttribute,
  saveStateTestId,
  timelineScalarEditorTestId,
  workbookEditRecoveryDiscardButtonTestId,
  workbookEditRecoveryTestId,
  workbookInlineDraftRowTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  handoffViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Route } from "@playwright/test";
import { expect } from "./fixtures";
import { gridSavedRows } from "./pages/workbookInspector";
import {
  installPatchController,
  installPatchTransportFailureController,
  successfulPatchCalls,
} from "./support/collaboration/replay";
import { currentLifecycle, lifecycleAction } from "./support/incidentLifecycle";
import { createIncident } from "./support/incidents/fixtures";
import { navigateVisualApplication } from "./support/runtime/applicationReadiness";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  holdBrowserRequest as holdBrowserApiRequest,
  installOwnedRoute,
  removeOwnedRoute,
} from "./support/transport/requestInterception";
import {
  assertViewportVisualRegression,
  assertVisualRegression,
} from "./support/visual/capture";
import { test } from "./support/visual/fixtures";
import {
  createRichTimelineIncident,
  createVisualTimelineRow,
} from "./support/visual/timelineVisualFixture";
import {
  assertStatusStripVisualRegression,
  assertTimelineEditorVisualRegression,
  assertWorkbookGridVisualRegression,
  hideVisualTechnicalColumns,
} from "./support/visual/workbookCaptures";
import { normalizeWorkbookGridVisualState } from "./support/visual/workbookGeometry";
import { mountedGridCell } from "./support/workbook/keyboardGrid";
import {
  fillOrdinaryField,
  openOrdinaryFixture,
  ordinaryField,
  retainOrdinaryUncertainty,
  switchOrdinarySheet,
} from "./support/workbook/ordinaryCreate";
import { queryViewRows } from "./support/workbook/query";

test("captures Timeline edit save-state visuals for active cell syncing saved and conflict states", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUALTIMELINEEDIT"),
    "Timeline mutation visual edit state",
  );
  const timelineRow = await createVisualTimelineRow(
    page,
    incidentId,
    "visualtimelineedit-row",
    {
      client_txn_id: uniqueTxn("VISUALTIMELINEEDIT-ROW"),
      "timeline.activity_utc_text": "2025-01-01T00:00:00Z",
      "timeline.activity_synopsis_text": "Editable visual row",
    },
  );

  await navigateVisualApplication(page, `/?incident_id=${incidentId}`);

  const saveState = page.getByTestId(saveStateTestId());
  const summaryInput = await mountedGridCell(
    page,
    timelineViewSchemaId,
    timelineRow.record_id,
    "timeline.activity_synopsis_text",
  );

  await expect(saveState).toHaveText("Saved");
  await summaryInput.click();
  const summaryEditor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.activity_synopsis_text",
      recordId: timelineRow.record_id,
      surface: "grid",
    }),
  );
  await expect(summaryEditor).toBeFocused();
  await summaryEditor.fill("Active visual edit");
  await assertTimelineEditorVisualRegression(
    page,
    "timeline-grid-active-edit-cell",
    summaryEditor,
  );

  const patchUrl = `**/api/v1/records/${timelineRow.record_id}`;
  const hold = await holdBrowserApiRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${timelineRow.record_id}`,
  });
  await summaryEditor.press("Enter");
  await hold.waitForHit;
  await expect(saveState).toHaveText("Syncing");
  await assertStatusStripVisualRegression(page, "timeline-grid-syncing-strip");
  await hold.release();
  await expect(saveState).toHaveText("Saved");
  await assertStatusStripVisualRegression(page, "timeline-grid-saved-strip");

  await hold.dispose();

  const conflictHandler = async (route: Route) => {
    if (route.request().method().toUpperCase() !== "PATCH") {
      await route.fallback();
      return;
    }
    await route.fulfill({
      status: 409,
      contentType: "application/json",
      body: JSON.stringify({
        error: {
          status: 409,
          code: "same_field_conflict",
          message: "same-field conflict",
          request_id: "visual-conflict",
          retryable: false,
          details: {},
          conflict: {
            conflict_token: "visual-conflict-token",
            record_id: timelineRow.record_id,
            field_key: "timeline.activity_synopsis_text",
            conflict_resolution_class: "text_compare_merge",
            base_row_version: timelineRow.row_version,
            current_row_version: timelineRow.row_version + 1,
            base_value: "Active visual edit",
            server_value: "Server visual edit",
            client_value: "Conflict visual edit",
            server_updated_by: "visual-remote-user",
            server_updated_at: "2026-06-02T12:00:00Z",
          },
        },
      }),
    });
  };

  await installOwnedRoute(page, patchUrl, conflictHandler);
  {
    const conflictInput = await mountedGridCell(
      page,
      timelineViewSchemaId,
      timelineRow.record_id,
      "timeline.activity_synopsis_text",
    );
    await conflictInput.click();
    const conflictEditor = page.getByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.activity_synopsis_text",
        recordId: timelineRow.record_id,
        surface: "grid",
      }),
    );
    await expect(conflictEditor).toBeFocused();
    await conflictEditor.fill("Conflict visual edit");
    await conflictEditor.press("Enter");
    await expect(saveState).toHaveText("Conflict");
    await assertStatusStripVisualRegression(
      page,
      "timeline-grid-conflict-strip",
    );

    await removeOwnedRoute(page, patchUrl, conflictHandler);
  }
});

test("Capture save-state, pending replay, transaction recovery, inline edit, and empty Timeline fixtures.", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUALMUTATION"),
    "browser.mutation-lifecycle visual readiness",
  );
  const timelineRow = await createVisualTimelineRow(
    page,
    incidentId,
    "visualmutation-row",
    {
      client_txn_id: uniqueTxn("VISUALMUTATION-ROW"),
      "timeline.activity_utc_text": "2026-06-03T10:00:00Z",
      "timeline.activity_synopsis_text":
        "browser.mutation-lifecycle visual editable row",
    },
  );

  await navigateVisualApplication(page, `/?incident_id=${incidentId}`);

  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

  const summaryInput = await mountedGridCell(
    page,
    timelineViewSchemaId,
    timelineRow.record_id,
    "timeline.activity_synopsis_text",
  );
  await expect(summaryInput).toHaveText(
    "browser.mutation-lifecycle visual editable row",
  );
  await summaryInput.click();
  const summaryEditor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.activity_synopsis_text",
      recordId: timelineRow.record_id,
      surface: "grid",
    }),
  );
  await expect(summaryEditor).toBeFocused();
  await summaryEditor.fill("browser.mutation-lifecycle active visual edit");
  await assertTimelineEditorVisualRegression(
    page,
    "timeline-mutation-active-edit-cell",
    summaryEditor,
  );

  const patchController = await installPatchTransportFailureController(page);
  patchController.disconnect();
  await summaryEditor.press("Enter");
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
  await expect(page.getByTestId(saveStateTestId())).toHaveAttribute(
    pendingReplayCountAttribute,
    "1",
  );
  await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
    scroll: { top: 0, left: "left" },
  });
  await assertVisualRegression(page, "timeline-mutation-pending-replay-status");

  patchController.connect();
  await expect
    .poll(() => successfulPatchCalls(patchController.calls).length)
    .toBe(1);
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await expect(page.getByTestId(saveStateTestId())).toHaveAttribute(
    pendingReplayCountAttribute,
    "0",
  );

  await patchController.dispose();

  const conflictController = await installPatchController(page);
  {
    conflictController.failNextPatch(409, "client_txn_conflict", {
      recordId: timelineRow.record_id,
    });
    const blockedSummary = await mountedGridCell(
      page,
      timelineViewSchemaId,
      timelineRow.record_id,
      "timeline.activity_synopsis_text",
    );
    await blockedSummary.click();
    const blockedEditor = page.getByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.activity_synopsis_text",
        recordId: timelineRow.record_id,
        surface: "grid",
      }),
    );
    await blockedEditor.fill(
      "browser.mutation-lifecycle blocked transaction edit",
    );
    await blockedEditor.press("Enter");
    await expect.poll(() => conflictController.calls.length).toBe(1);
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
    await page
      .getByRole("button", { name: "Open conflict recovery", exact: true })
      .click();
    await expect(page.getByTestId(workbookEditRecoveryTestId())).toBeVisible();
    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "left" },
    });
    await assertVisualRegression(
      page,
      "timeline-mutation-transaction-recovery-panel",
    );

    await page.setViewportSize({ width: 1024, height: 720 });
    await expect(page.getByTestId(workbookEditRecoveryTestId())).toBeVisible();
    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "left" },
    });
    await assertViewportVisualRegression(
      page,
      "timeline-mutation-transaction-recovery-panel-narrow",
    );

    await page.setViewportSize({ width: 768, height: 640 });
    await expect(page.getByTestId(workbookEditRecoveryTestId())).toBeVisible();
    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "left" },
    });
    await assertViewportVisualRegression(
      page,
      "timeline-mutation-transaction-recovery-panel-compact",
    );
    await page.setViewportSize({ width: 1440, height: 900 });

    await page.getByTestId(workbookEditRecoveryDiscardButtonTestId()).click();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(page.getByTestId(workbookEditRecoveryTestId())).toHaveCount(0);

    await conflictController.dispose();
  }

  const emptyIncidentId = await createIncident(
    page,
    uniqueIncidentKey("VISUALEMPTYQUERY"),
    "browser.mutation-lifecycle empty Timeline query",
  );
  await navigateVisualApplication(page, `/?incident_id=${emptyIncidentId}`);

  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, emptyIncidentId, timelineViewSchemaId))
          .length,
    )
    .toBe(0);
  await expect(gridSavedRows(page, timelineViewSchemaId)).toHaveCount(0);
  await expect(
    page.getByTestId(workbookInlineDraftRowTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Create timeline row" }),
  ).toBeVisible();
  const successfulEmptyState = page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator('[data-grid-data-state="empty"]');
  await expect(successfulEmptyState).toBeVisible();
  await expect(successfulEmptyState).toContainText(
    "No Timeline records have been added.",
  );
  await expect(
    successfulEmptyState.getByRole("button", { name: "Add row" }),
  ).toBeVisible();
  await successfulEmptyState.getByRole("button", { name: "Add row" }).click();
  await expect
    .poll(() =>
      page
        .getByTestId(workbookInlineDraftRowTestId(timelineViewSchemaId))
        .evaluate((row) => row.contains(document.activeElement)),
    )
    .toBe(true);
  await assertWorkbookGridVisualRegression(
    page,
    "timeline-mutation-empty-timeline-query",
    timelineViewSchemaId,
    { scroll: { top: 0, left: "left" } },
  );
});

test("Capture ordinary grid reference authoring and retained recovery across workbook layouts", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  const handoff = handoffViewSchemaId;
  const { incident } = await openOrdinaryFixture(page, handoff, (url) =>
    navigateVisualApplication(page, url),
  );
  await fillOrdinaryField(
    page,
    handoff,
    "handoff.current_state_summary",
    "Retained ordinary handoff",
  );
  await hideVisualTechnicalColumns(page, handoff, ["Outgoing Owner"]);
  const referenceInput = await ordinaryField(
    page,
    handoff,
    "handoff.incoming_owner_user_id",
  );
  const reference = page.getByRole("gridcell").filter({ has: referenceInput });
  await reference
    .getByRole("button", { name: "Choose incoming owner", exact: true })
    .click();
  await expect(
    reference
      .getByRole("group", { name: "Incoming Owner", exact: true })
      .getByRole("radio")
      .first(),
  ).toBeEnabled();
  await assertViewportVisualRegression(page, "ordinary-reference-authoring", {
    ready: () =>
      normalizeWorkbookGridVisualState(page, handoff, {
        scroll: { top: 0, left: "left" },
      }),
  });
  await test.info().attach("ordinary-reference-authoring-review", {
    body: await page.screenshot({ animations: "disabled", caret: "hide" }),
    contentType: "image/png",
  });
  await reference
    .getByRole("button", { name: "Cancel references", exact: true })
    .click();
  await switchOrdinarySheet(page, evidenceViewSchemaId);
  await fillOrdinaryField(
    page,
    evidenceViewSchemaId,
    "evidence.title",
    "Retained ordinary evidence",
  );
  await fillOrdinaryField(
    page,
    evidenceViewSchemaId,
    "evidence.requested_at",
    "2025-03-01T10:00:00Z",
  );
  const { recovery } = await retainOrdinaryUncertainty(page, incident);
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    await assertViewportVisualRegression(
      page,
      `ordinary-recovery-${viewport.width}`,
      {
        ready: () =>
          normalizeWorkbookGridVisualState(page, evidenceViewSchemaId, {
            scroll: { top: 0, left: "left" },
          }),
      },
    );
    await test.info().attach(`ordinary-recovery-${viewport.width}-review`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
  }
  const lifecycle = await currentLifecycle(page, incident);
  expect(
    (
      await lifecycleAction(page, incident, "closeIncident", {
        client_txn_id: uniqueTxn("ordinary-visual-close"),
        base_incident_version: lifecycle.incident_version,
        reason: "Copy retained ordinary authoring",
      })
    ).ok,
  ).toBe(true);
  await expect(
    page.getByRole("textbox", {
      name: "Title retained authoring",
      exact: true,
    }),
  ).toBeVisible();
  await assertViewportVisualRegression(
    page,
    "ordinary-closed-retained-narrow",
    {
      ready: async () => {
        await normalizeWorkbookGridVisualState(page, evidenceViewSchemaId, {
          scroll: { top: 0, left: "left" },
        });
        const lifecycleStatus = page.getByRole("status", {
          name: "Incident lifecycle",
        });
        await expect(lifecycleStatus).toHaveText("Closed, read-only");
        expect(
          await lifecycleStatus.evaluate((element) => {
            const label = element.getBoundingClientRect();
            const band = element.closest("section")?.getBoundingClientRect();
            return (
              !!band && label.top >= band.top && label.bottom <= band.bottom
            );
          }),
        ).toBe(true);
      },
    },
  );
  await test.info().attach("ordinary-closed-retained-review", {
    body: await page.screenshot({ animations: "disabled", caret: "hide" }),
    contentType: "image/png",
  });
  await test.info().attach("ordinary-recovery-tree", {
    body: await recovery.ariaSnapshot(),
    contentType: "text/plain",
  });
});
