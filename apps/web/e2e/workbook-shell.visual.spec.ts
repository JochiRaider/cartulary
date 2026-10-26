import type { ViewRow } from "@cartulary/protocol-ts/http";
import {
  dataTestIdSelector,
  gridGroupingSelectTestId,
  gridRowGutterTestId,
  gridSavedRowsSelector,
  gridScrollportSelector,
  gridShellTestId,
  gridSortHeaderTestId,
  incidentAdministrationTestId,
  incidentControlsPanelTestId,
  incidentMembershipListTestId,
  savedViewSelectorTestId,
  saveStateTestId,
  surfaceTabTestId,
  systemViewSwitcherTriggerTestId,
  timelineInspectorSectionTestId,
  timelineInspectorTestId,
  workbookFilterPopoverTriggerTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorToggleTestId,
  workbookPresenceSummaryTestId,
  workbookResponsiveBandTestId,
  workbookShellReadyTestId,
  workbookShellSlots,
  workbookShellSlotTestId,
  workbookSortMenuTriggerTestId,
  workbookViewBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect } from "./fixtures";
import { gridSavedRows } from "./pages/workbookInspector";
import { navigateVisualApplication } from "./support/runtime/applicationReadiness";
import { seedVisualTimelineInvestigation } from "./support/timeline/timelineInvestigation";
import { assertViewportVisualRegression } from "./support/visual/capture";
import { test } from "./support/visual/fixtures";
import { captureRichTimelineInspector } from "./support/visual/richTimelineInspector";
import {
  focusTimelineShellOrigin,
  verifyTimelineShellOrigin,
} from "./support/visual/timelineShellAnchor";
import { registerTimelineVisualFixture } from "./support/visual/timelineVisualFixture";
import {
  gridShellSelector,
  normalizeWorkbookGridVisualState,
} from "./support/visual/workbookGeometry";
import { captureWorkbookLayoutStudies } from "./support/visual/workbookLayoutStudies";
import {
  mountedGridCell,
  mountedGridTarget,
} from "./support/workbook/keyboardGrid";
import { queryViewRows } from "./support/workbook/query";
import { activateCommittedGridCell } from "./support/workbook/rowMutations";
import {
  createSavedView,
  selectSavedView,
} from "./support/workbook/savedViews";

async function expectWideWorkbookTopBarChrome(page: Page) {
  await expect(
    page.getByTestId(workbookResponsiveBandTestId()),
  ).toHaveAttribute("data-workbook-responsive-band", "base");
  await expect(
    page.getByTestId(surfaceTabTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(systemViewSwitcherTriggerTestId()),
  ).toBeVisible();
  await expect(
    page.getByTestId(workbookViewBarQueryControlsTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(workbookSortMenuTriggerTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(gridGroupingSelectTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(workbookFilterPopoverTriggerTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByLabel("Account and application navigation"),
  ).toBeVisible();
}

async function readTimelineGridFirstLayout(page: Page) {
  return page.evaluate(
    ({
      gridSelector,
      inspectorSelector,
      scrollportSelector,
      topBarSelector,
      viewBarSelector,
    }) => {
      const rectFor = (element: HTMLElement) => {
        const rect = element.getBoundingClientRect();
        return {
          bottom: Math.round(rect.bottom),
          height: Math.round(rect.height),
          left: Math.round(rect.left),
          right: Math.round(rect.right),
          top: Math.round(rect.top),
          width: Math.round(rect.width),
        };
      };
      const roundedRect = (selector: string) => {
        const element = document.querySelector<HTMLElement>(selector);
        if (element === null) {
          throw new Error(`Expected ${selector} to exist`);
        }
        return rectFor(element);
      };
      const optionalRoundedRect = (selector: string) => {
        const element = document.querySelector<HTMLElement>(selector);
        if (element === null) {
          return null;
        }
        return rectFor(element);
      };
      const scrollport =
        document.querySelector<HTMLElement>(scrollportSelector);
      if (scrollport === null) {
        throw new Error(`Expected ${scrollportSelector} to exist`);
      }
      const columnHeaders = Array.from(
        scrollport.querySelectorAll<HTMLElement>('[role="columnheader"]'),
      );
      const lastColumnElement =
        columnHeaders.length === 0
          ? null
          : columnHeaders[columnHeaders.length - 1];
      const lastColumn =
        lastColumnElement === null || lastColumnElement === undefined
          ? null
          : rectFor(lastColumnElement);
      return {
        grid: roundedRect(gridSelector),
        innerGrid: roundedRect(scrollportSelector),
        inspector: optionalRoundedRect(inspectorSelector),
        lastColumn,
        scrollport: {
          clientHeight: scrollport.clientHeight,
          clientWidth: scrollport.clientWidth,
          scrollTop: Math.round(scrollport.scrollTop),
        },
        topBar: roundedRect(topBarSelector),
        viewBar: roundedRect(viewBarSelector),
        windowY: Math.round(window.scrollY),
      };
    },
    {
      gridSelector: gridShellSelector(timelineViewSchemaId),
      inspectorSelector: dataTestIdSelector(timelineInspectorTestId()),
      scrollportSelector: `${gridShellSelector(
        timelineViewSchemaId,
      )} ${gridScrollportSelector()}`,
      topBarSelector: dataTestIdSelector(workbookShellSlotTestId("top-bar")),
      viewBarSelector: dataTestIdSelector(workbookShellSlotTestId("view-bar")),
    },
  );
}

test("Capture Default Timeline workbook shell with view-bar query controls, compact sheet toolbar, dense Timeline grid, collapsed inspector default, explicit inspector opener, bottom draft row, and status strip.", async ({
  page,
}) => {
  test.setTimeout(120_000);
  await page.setViewportSize({ width: 1440, height: 900 });
  const investigation = registerTimelineVisualFixture(
    await seedVisualTimelineInvestigation(page, { continuationCount: 36 }),
  );
  const incidentId = investigation.incidentId;
  const rows: ViewRow[] = investigation.rows;
  const rowSummariesById = new Map(
    rows.map((row) => [
      row.record_id,
      String(row.cells["timeline.activity_synopsis_text"]?.value ?? ""),
    ]),
  );
  await test.info().attach("investigation-receipt", {
    body: JSON.stringify(investigation.outcomes),
    contentType: "application/json",
  });
  const longQuerySavedView = await createSavedView(page, incidentId, {
    display_name:
      "Workbook view-bar visual resilience with a deliberately long selected saved-view name",
    query_json: {
      group_by: "timeline.capture_state",
      sort: [
        { direction: "asc", field_key: "timeline.activity_sort_ts" },
        { direction: "desc", field_key: "timeline.date_entered_sort_day" },
        { direction: "asc", field_key: "timeline.activity_synopsis_text" },
        { direction: "desc", field_key: "timeline.analyst_text" },
        { direction: "asc", field_key: "timeline.mitre_stage_text" },
        { direction: "desc", field_key: "timeline.device_object_text" },
        { direction: "asc", field_key: "timeline.ip_address_text" },
        { direction: "desc", field_key: "timeline.capture_state" },
      ],
    },
    view_schema_id: timelineViewSchemaId,
  });

  await navigateVisualApplication(page, `/?incident_id=${incidentId}`);

  const shell = page.getByTestId(workbookShellReadyTestId());
  await expect(shell).toBeVisible();
  await expect(shell).toHaveAttribute(
    "data-active-view-schema-id",
    timelineViewSchemaId,
  );
  for (const slot of workbookShellSlots) {
    if (slot === "inspector") {
      await expect(
        shell.locator(dataTestIdSelector(workbookShellSlotTestId(slot))),
      ).toHaveCount(0);
      continue;
    }
    await expect(
      shell.locator(dataTestIdSelector(workbookShellSlotTestId(slot))),
    ).toBeVisible();
  }
  await expect(page.getByTestId(incidentControlsPanelTestId())).toHaveCount(0);
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-key")),
  ).toHaveCount(0);
  await expect(
    page.getByTestId(incidentAdministrationTestId("patch-button")),
  ).toHaveCount(0);
  await expect(page.getByTestId(incidentMembershipListTestId())).toHaveCount(0);
  await expect(page.getByText("Timeline mutation workbook")).toHaveCount(0);
  await expect(page.getByText(/Timeline mutation substrate/u)).toHaveCount(0);
  await expect(
    page.getByTestId(surfaceTabTestId(timelineViewSchemaId)),
  ).toHaveAttribute("aria-selected", "true");
  await expect(
    page.getByTestId(systemViewSwitcherTriggerTestId()),
  ).toBeVisible();
  await expect(
    page.getByTestId(savedViewSelectorTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

  const grid = page.getByTestId(gridShellTestId(timelineViewSchemaId));
  await expect(grid).toBeVisible();
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, incidentId, timelineViewSchemaId)).length,
    )
    .toBe(rows.length);
  await test.info().attach("initial-workbook-geometry", {
    body: JSON.stringify(await readTimelineGridFirstLayout(page)),
    contentType: "application/json",
  });
  await test.info().attach("initial-workbook-study", {
    body: await page.screenshot({ animations: "disabled" }),
    contentType: "image/png",
  });
  await expect
    .poll(async () => gridSavedRows(page, timelineViewSchemaId).count())
    .toBeGreaterThanOrEqual(12);
  const renderedRecordIds = await grid
    .locator(gridSavedRowsSelector())
    .evaluateAll((rowElements) =>
      rowElements.map(
        (rowElement) => rowElement.getAttribute("data-grid-record-id") ?? "",
      ),
    );
  expect(rows.map((row) => row.record_id)).toEqual(
    expect.arrayContaining(renderedRecordIds),
  );
  const selectedRowId = investigation.recordId("authentication-anomaly");
  const selectedRow = rows.find((row) => row.record_id === selectedRowId);
  if (selectedRow === undefined) {
    throw new Error(
      `visual.workbook-shell fixture selected unknown row ${selectedRowId}`,
    );
  }
  const selectedGridRow = grid.locator(
    `[data-grid-record-id="${selectedRow.record_id}"]`,
  );
  await activateCommittedGridCell(
    (
      await mountedGridCell(
        page,
        timelineViewSchemaId,
        selectedRow.record_id,
        "timeline.date_entered_text",
      )
    ).locator("xpath=ancestor::*[@role='gridcell'][1]"),
  );
  await expect(selectedGridRow).toHaveAttribute(
    "data-inspector-active",
    "true",
  );
  const selectedGridRowGutter = selectedGridRow.getByTestId(
    gridRowGutterTestId(timelineViewSchemaId, selectedRow.record_id),
  );
  await expect(selectedGridRowGutter).toBeVisible();
  await expect(selectedGridRowGutter).toHaveAttribute(
    "data-grid-field-key",
    "__cartulary_row_gutter__",
  );

  const defaultTimelineFields = [
    "timeline.date_entered_text",
    "timeline.analyst_text",
    "timeline.mitre_stage_text",
    "timeline.device_object_text",
    "timeline.ip_address_text",
    "timeline.activity_utc_text",
    "timeline.activity_local_text",
    "timeline.raw_activity_text",
    "timeline.activity_synopsis_text",
    "timeline.data_source_text",
  ];
  for (const fieldKey of defaultTimelineFields) {
    const header = await mountedGridTarget(
      page,
      timelineViewSchemaId,
      gridSortHeaderTestId(timelineViewSchemaId, fieldKey),
    );
    await expect(header).toHaveAttribute("data-grid-field-key", fieldKey);
  }

  await expect(
    await mountedGridCell(
      page,
      timelineViewSchemaId,
      selectedRow.record_id,
      "timeline.activity_synopsis_text",
    ),
  ).toHaveText(rowSummariesById.get(selectedRow.record_id) ?? "");

  const fixtureViewport = page.viewportSize() ?? { width: 1440, height: 900 };
  await expectWideWorkbookTopBarChrome(page);
  for (const shortHeight of [640, 560]) {
    await page.setViewportSize({
      width: fixtureViewport.width,
      height: shortHeight,
    });
    await expectWideWorkbookTopBarChrome(page);
    await expect
      .poll(async () => (await readTimelineGridFirstLayout(page)).windowY)
      .toBe(0);
  }
  await page.setViewportSize(fixtureViewport);
  await expectWideWorkbookTopBarChrome(page);

  await page.setViewportSize({ width: 2048, height: fixtureViewport.height });
  await expect
    .poll(async () => {
      const layout = await readTimelineGridFirstLayout(page);
      return (
        layout.innerGrid.right >= layout.grid.right - 2 &&
        layout.lastColumn !== null &&
        layout.lastColumn.right >= layout.grid.right - 2
      );
    })
    .toBe(true);
  const wideLayout = await readTimelineGridFirstLayout(page);
  expect(wideLayout.grid.width).toBeGreaterThan(1600);
  expect(wideLayout.innerGrid.right).toBeGreaterThanOrEqual(
    wideLayout.grid.right - 2,
  );
  const wideLastColumn = wideLayout.lastColumn;
  expect(wideLastColumn).not.toBeNull();
  if (wideLastColumn === null) {
    throw new Error("Expected wide Timeline grid to render a last column");
  }
  expect(wideLastColumn.right).toBeGreaterThanOrEqual(
    wideLayout.grid.right - 2,
  );
  await page
    .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
    .click();
  await expect(page.getByTestId(timelineInspectorTestId())).toBeVisible();
  const wideDrawerOpenLayout = await readTimelineGridFirstLayout(page);
  expect(wideDrawerOpenLayout.inspector).not.toBeNull();
  expect(wideDrawerOpenLayout.grid.left).toBe(wideLayout.grid.left);
  expect(wideDrawerOpenLayout.grid.right).toBeLessThan(wideLayout.grid.right);
  expect(wideDrawerOpenLayout.grid.width).toBeGreaterThan(0);
  expect(wideDrawerOpenLayout.inspector?.left).toBeGreaterThanOrEqual(
    wideDrawerOpenLayout.grid.right,
  );
  const wideDrawerLastColumn = wideDrawerOpenLayout.lastColumn;
  expect(wideDrawerLastColumn).not.toBeNull();
  if (wideDrawerLastColumn === null) {
    throw new Error(
      "Expected wide Timeline grid with inspector to render a last column",
    );
  }
  expect(wideDrawerLastColumn.right).toBeGreaterThanOrEqual(
    wideDrawerOpenLayout.grid.right - 2,
  );
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await page.setViewportSize(fixtureViewport);
  await expect
    .poll(async () => {
      const layout = await readTimelineGridFirstLayout(page);
      return layout.innerGrid.right >= layout.grid.right - 2;
    })
    .toBe(true);
  const closedLayout = await readTimelineGridFirstLayout(page);
  expect(closedLayout.windowY).toBe(0);
  expect(closedLayout.inspector).toBeNull();
  expect(closedLayout.innerGrid.right).toBeGreaterThanOrEqual(
    closedLayout.grid.right - 2,
  );
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await expect(
    page.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await page
    .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
    .click();
  await expect(page.getByTestId(timelineInspectorTestId())).toBeVisible();
  const drawerOpenLayout = await readTimelineGridFirstLayout(page);
  expect(drawerOpenLayout.grid.left).toBe(closedLayout.grid.left);
  expect(drawerOpenLayout.grid.right).toBeLessThan(closedLayout.grid.right);
  expect(drawerOpenLayout.grid.width).toBeGreaterThan(0);
  expect(drawerOpenLayout.viewBar).toEqual(closedLayout.viewBar);
  expect(drawerOpenLayout.topBar).toEqual(closedLayout.topBar);
  expect(drawerOpenLayout.inspector).not.toBeNull();
  expect(drawerOpenLayout.inspector?.left).toBeGreaterThanOrEqual(
    drawerOpenLayout.grid.right,
  );
  expect(drawerOpenLayout.inspector?.top).toBeGreaterThanOrEqual(
    drawerOpenLayout.viewBar.bottom - 1,
  );
  expect(drawerOpenLayout.inspector?.top).toBeLessThanOrEqual(
    drawerOpenLayout.grid.top + 1,
  );
  const gridRegion = await page
    .getByTestId(workbookShellSlotTestId("primary-grid"))
    .boundingBox();
  if (!gridRegion) throw new Error("Missing primary grid region");
  expect(drawerOpenLayout.inspector?.bottom).toBeLessThanOrEqual(
    gridRegion.y + gridRegion.height,
  );
  expect(drawerOpenLayout.windowY).toBe(0);
  await expect(
    page.getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(page.getByTestId(timelineInspectorTestId())).toContainText(
    rowSummariesById.get(selectedRow.record_id) ?? "Selected timeline row",
  );
  for (const section of ["relationships", "evidence", "history"] as const) {
    await expect(
      page.getByTestId(timelineInspectorSectionTestId(section)),
    ).toBeVisible();
  }
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);

  const timelineScrollportSelector = `${dataTestIdSelector(
    gridShellTestId(timelineViewSchemaId),
  )} ${gridScrollportSelector()}`;
  await expect(
    page
      .getByTestId(gridShellTestId(timelineViewSchemaId))
      .locator(gridScrollportSelector()),
  ).toBeVisible();
  await page.evaluate((selector) => {
    document.querySelector<HTMLElement>(selector)?.scrollTo({ top: 240 });
  }, timelineScrollportSelector);
  await expect
    .poll(async () => (await readTimelineGridFirstLayout(page)).scrollport)
    .toMatchObject({ scrollTop: 240 });
  const scrolledLayout = await readTimelineGridFirstLayout(page);
  expect(scrolledLayout.topBar).toEqual(closedLayout.topBar);
  expect(scrolledLayout.viewBar).toEqual(closedLayout.viewBar);
  expect(scrolledLayout.windowY).toBe(0);

  await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
    scroll: { top: 0, left: "left" },
  });
  await investigation.verify();
  await page.mouse.move(0, 0);
  await assertViewportVisualRegression(
    page,
    "incident-directory-default-timeline-workbook-shell",
    {
      prepareAnchor: () =>
        focusTimelineShellOrigin(page, selectedRow.record_id),
      verifyFraming: () =>
        verifyTimelineShellOrigin(page, selectedRow.record_id),
    },
  );

  await selectSavedView(
    page,
    timelineViewSchemaId,
    longQuerySavedView.saved_view_id,
  );
  await expect(
    page.getByTestId(savedViewSelectorTestId(timelineViewSchemaId)),
  ).toHaveAttribute(
    "data-selected-saved-view-id",
    longQuerySavedView.saved_view_id,
  );
  await expect(page.locator('[data-grid-data-state="refreshing"]')).toHaveCount(
    0,
  );
  await expect(
    page.locator(
      '[data-grid-data-state="stale_error"], [data-grid-data-state="unavailable"]',
    ),
  ).toHaveCount(0);
  await expect(
    page.getByTestId(workbookViewBarQueryControlsTestId(timelineViewSchemaId)),
  ).toHaveAttribute("data-hidden-query-chip-count", "6");

  await page.setViewportSize({ width: 1024, height: 720 });
  await expect(
    page.getByTestId(workbookResponsiveBandTestId()),
  ).toHaveAttribute("data-workbook-responsive-band", "narrow_desktop");
  await expect(
    page.getByTestId(workbookViewBarQueryControlsTestId(timelineViewSchemaId)),
  ).toHaveAttribute("data-query-chip-capacity", "2");
  await expect(
    page.getByTestId(workbookViewBarQueryControlsTestId(timelineViewSchemaId)),
  ).toHaveAttribute("data-hidden-query-chip-count", "7");
  await expect(
    page.getByTestId(workbookFilterPopoverTriggerTestId(timelineViewSchemaId)),
  ).toHaveAttribute(
    "aria-label",
    "Filters, 0 active filters, 7 hidden query entries",
  );
  await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
    scroll: { top: 0, left: "left" },
  });
  await assertViewportVisualRegression(
    page,
    "incident-directory-narrow-desktop-workbook-shell",
  );

  await page.setViewportSize({ width: 768, height: 640 });
  await expect(
    page.getByTestId(workbookResponsiveBandTestId()),
  ).toHaveAttribute("data-workbook-responsive-band", "compact_desktop");
  await expect(
    page.getByTestId(workbookViewBarQueryControlsTestId(timelineViewSchemaId)),
  ).toHaveAttribute("data-query-chip-capacity", "0");
  await expect(
    page.getByTestId(workbookViewBarQueryControlsTestId(timelineViewSchemaId)),
  ).toHaveAttribute("data-hidden-query-chip-count", "9");
  await expect(
    page.getByTestId(workbookFilterPopoverTriggerTestId(timelineViewSchemaId)),
  ).toHaveAttribute(
    "aria-label",
    "Filters, 0 active filters, 9 hidden query entries",
  );
  await expect(
    page
      .getByTestId(workbookShellSlotTestId("status-strip"))
      .getByTestId(workbookPresenceSummaryTestId()),
  ).toBeVisible();
  await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
    scroll: { top: 0, left: "left" },
  });
  await assertViewportVisualRegression(
    page,
    "incident-directory-compact-desktop-workbook-shell",
  );
  await captureWorkbookLayoutStudies(page, test.info(), () =>
    normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "left" },
    }),
  );
  if (
    !(await page
      .getByTestId(savedViewSelectorTestId(timelineViewSchemaId))
      .isVisible())
  ) {
    await page
      .getByRole("button", { name: "View options controls", exact: true })
      .click();
  }
  await selectSavedView(page, timelineViewSchemaId, "");
  await captureRichTimelineInspector(
    page,
    test.info(),
    investigation.recordId("evidence-collected"),
  );
  await investigation.verify();
});
