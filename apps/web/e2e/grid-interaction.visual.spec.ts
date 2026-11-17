import { changeGrouping } from "@cartulary/test-utils/grid";
import {
  gridFillHandleSelector,
  gridGroupRowSelector,
  gridGroupRowTestId,
  gridScrollportSelector,
  gridShellTestId,
  timelineInspectorTestId,
  timelineRowMarkReviewedButtonTestId,
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
  workbookInspectorCloseButtonTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect } from "./fixtures";
import {
  activateTimelineGridEditor,
  expectVisibleSemanticGridCellFocus,
} from "./support/accessibility/grid";
import { navigateVisualApplication } from "./support/runtime/applicationReadiness";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { test } from "./support/visual/fixtures";
import { clickTimelineRowAction } from "./support/visual/rowActions";
import {
  createRichTimelineIncident,
  createVisualTimelineRow,
  visualTimelineRecord,
} from "./support/visual/timelineVisualFixture";
import { assertWorkbookGridVisualRegression } from "./support/visual/workbookCaptures";
import { mountedGridCell } from "./support/workbook/keyboardGrid";

test("captures Timeline grouped rows and currently exposed grid chrome", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUALTIMELINEGROUPED"),
    "Timeline mutation visual grouped rows",
  );
  const firstRow = await createVisualTimelineRow(
    page,
    incidentId,
    "visualtimelinegrouped-rowa",
    {
      client_txn_id: uniqueTxn("VISUALTIMELINEGROUPED-ROWA"),
      "timeline.activity_utc_text": "2025-02-17T11:00:00Z",
      "timeline.activity_synopsis_text": "Alpha grouped row",
    },
  );
  await createVisualTimelineRow(
    page,
    incidentId,
    "visualtimelinegrouped-rowb",
    {
      client_txn_id: uniqueTxn("VISUALTIMELINEGROUPED-ROWB"),
      "timeline.activity_utc_text": "2025-02-17T11:05:00Z",
      "timeline.activity_synopsis_text": "Beta grouped row",
    },
  );

  await navigateVisualApplication(page, `/?incident_id=${incidentId}`);

  await clickTimelineRowAction(
    page,
    firstRow.record_id,
    timelineRowMarkReviewedButtonTestId(firstRow.record_id),
  );
  await expect(
    await mountedGridCell(
      page,
      timelineViewSchemaId,
      firstRow.record_id,
      "timeline.capture_state",
    ),
  ).toHaveText("reviewed");

  // The shortcut opens canonical History; retain this fixture's grid framing.
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
  await expect(
    page.getByTestId(
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "reviewed",
      ),
    ),
  ).toBeVisible();
  await expect(
    page.getByTestId(
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "rough",
      ),
    ),
  ).toBeVisible();
  await expect(
    page
      .getByTestId(gridShellTestId(timelineViewSchemaId))
      .getByText("Unassigned", { exact: true }),
  ).toHaveCount(0);

  await assertWorkbookGridVisualRegression(
    page,
    "timeline-grid-grouped-grid",
    timelineViewSchemaId,
    { scroll: { top: 0, left: "left" } },
  );
});

test("Capture frozen column, resize handle, fill-down handle, edit cell, and group outline row grid-adapter fixtures.", async ({
  page,
}) => {
  // The real grid adapter, not injected markup: each feature is asserted
  // semantically before its capture. The adapter removes the fill handle
  // while editing or grouped, so the features need three captures.
  await page.setViewportSize({ width: 1440, height: 900 });
  const view = timelineViewSchemaId;
  // The first default column is frozen and its neighbour is edited, so every
  // feature stays inside the first viewport.
  const frozenField = "timeline.date_entered_text";
  const cellField = "timeline.analyst_text";
  // The shared rich investigation: 12 realistic rows across the rough,
  // enriched, reviewed and superseded capture states.
  const incidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUALGRIDADAPTER"),
    "browser.grid-interaction grid adapter visual fixture",
  );
  const row = visualTimelineRecord(incidentId, "script-execution");

  await navigateVisualApplication(page, `/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  const grid = page.getByTestId(gridShellTestId(view));

  // Frozen data column through the real Columns control.
  const columns = page.getByTestId(workbookColumnsMenuTestId(view));
  if (!(await columns.isVisible())) {
    const trigger = page.getByTestId(workbookColumnsMenuTriggerTestId(view));
    if (!(await trigger.isVisible()))
      await page
        .getByRole("button", { name: "View options controls", exact: true })
        .click();
    await trigger.click();
  }
  await columns
    .getByRole("button", {
      name: `Freeze through ${requireViewContract(view).fieldMap[frozenField]?.label}`,
      exact: true,
    })
    .click();
  await columns
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
  await expect(page.locator(gridScrollportSelector())).toHaveAttribute(
    "data-grid-freeze-state",
    "active",
  );
  await expect(
    grid.locator('.cartulary-grid-frozen-boundary[role="columnheader"]'),
  ).toHaveCount(1);
  // Resize handles are transparent by design, so they are asserted in the DOM.
  expect(
    await grid.locator("[data-grid-column-resize]").count(),
  ).toBeGreaterThan(1);

  // Selected cell with the real fill handle.
  await expectVisibleSemanticGridCellFocus(
    await mountedGridCell(page, view, row.record_id, cellField),
  );
  await expect(page.locator(gridFillHandleSelector())).toHaveCount(1);
  await assertWorkbookGridVisualRegression(
    page,
    "timeline-grid-adapter-fill-handle",
    view,
    { scroll: { top: 0, left: "left" } },
  );

  // Open in-cell editor on the selected cell; the adapter removes the fill
  // handle while editing.
  const editor = await activateTimelineGridEditor(
    page,
    row.record_id,
    cellField,
  );
  await expect(grid.locator('[data-grid-editing="true"]')).toHaveCount(1);
  await expect(page.locator(gridFillHandleSelector())).toHaveCount(0);
  await assertWorkbookGridVisualRegression(
    page,
    "timeline-grid-adapter-editor",
    view,
    { scroll: { top: 0, left: "left" } },
  );
  // The capture normalizes scroll, so the editor must still be in view.
  await expect(editor).toBeInViewport();
  await expect(grid.locator('[data-grid-editing="true"]')).toHaveCount(1);
  await editor.press("Escape");
  await expect(grid.locator('[data-grid-editing="true"]')).toHaveCount(0);

  // Grouped outline rows beside the frozen boundary.
  await changeGrouping(page, view, "timeline.capture_state");
  for (const state of ["rough", "enriched", "reviewed", "superseded"]) {
    await expect(
      page.getByTestId(
        gridGroupRowTestId(view, "timeline.capture_state", state),
      ),
    ).toBeVisible();
  }
  await expect(grid.locator(gridGroupRowSelector())).toHaveCount(4);
  await expect(page.locator(gridFillHandleSelector())).toHaveCount(0);
  await assertWorkbookGridVisualRegression(
    page,
    "timeline-grid-adapter-grouped-frozen",
    view,
    { scroll: { top: 0, left: "left" } },
  );
});
