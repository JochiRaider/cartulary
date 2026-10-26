import { changeGrouping } from "@cartulary/test-utils/grid";
import {
  gridGroupRowTestId,
  gridShellTestId,
  timelineInspectorTestId,
  timelineRowMarkReviewedButtonTestId,
  workbookInspectorCloseButtonTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect } from "./fixtures";
import { createIncident } from "./support/incidents/fixtures";
import { navigateVisualApplication } from "./support/runtime/applicationReadiness";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { assertVisualRegression } from "./support/visual/capture";
import { injectDesignFixture, test } from "./support/visual/fixtures";
import { clickTimelineRowAction } from "./support/visual/rowActions";
import {
  createRichTimelineIncident,
  createVisualTimelineRow,
} from "./support/visual/timelineVisualFixture";
import { assertWorkbookGridVisualRegression } from "./support/visual/workbookCaptures";
import { mountedGridCell } from "./support/workbook/keyboardGrid";
import { createViewRow } from "./support/workbook/query";

async function injectFeP3GridAdapterVisualFixture(page: Page) {
  await injectDesignFixture(page, {
    ariaLabel: "browser.grid-interaction grid adapter visual fixtures",
    fixtureName: "grid-interaction-grid-adapter",
    missingMainMessage:
      "Expected workbook shell main before browser.grid-interaction fixture",
    styleText: `
      [data-design-fixture='grid-interaction-grid-adapter'] {
        position: fixed;
        inset: var(--ct-spacing-xl);
        box-sizing: border-box;
        display: grid;
        grid-template-columns: minmax(0, 1fr);
        gap: var(--ct-spacing-md);
        overflow: hidden;
        background: var(--ct-colors-canvas);
        color: var(--ct-colors-ink);
        border: var(--ct-border-strong);
        border-radius: var(--ct-rounded-lg);
        padding: var(--ct-spacing-lg);
        box-shadow: var(--ct-elevation-panel);
        font-family: var(--ct-typography-ui-fontFamily);
        font-size: var(--ct-typography-ui-fontSize);
        font-weight: var(--ct-typography-ui-fontWeight);
        letter-spacing: var(--ct-typography-ui-letterSpacing);
        line-height: var(--ct-typography-ui-lineHeight);
        z-index: 1000;
      }

      [data-design-fixture='grid-interaction-grid-adapter'] * {
        box-sizing: border-box;
      }

      .grid-interaction-grid-fixture-table {
        min-width: 0;
        overflow: hidden;
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-md);
        background: var(--ct-colors-surface-1);
      }

      .grid-interaction-grid-fixture-row {
        display: grid;
        grid-template-columns: 10rem 16rem 12rem 14rem 14rem;
        min-width: 66rem;
      }

      .grid-interaction-grid-fixture-head,
      .grid-interaction-grid-fixture-cell {
        position: relative;
        min-width: 0;
        min-height: 3.75rem;
        display: flex;
        align-items: center;
        gap: var(--ct-spacing-xs);
        overflow: hidden;
        padding: var(--ct-density-default-cellPadding);
        border-inline-end: var(--ct-border-hairline);
        border-block-end: var(--ct-border-hairline);
        background: var(--ct-colors-surface-1);
        color: var(--ct-colors-ink);
      }

      .grid-interaction-grid-fixture-head {
        min-height: 3rem;
        background: var(--ct-colors-surface-2);
        color: var(--ct-colors-ink-muted);
        font-family: var(--ct-typography-metadata-fontFamily);
        font-size: var(--ct-typography-metadata-fontSize);
        font-weight: var(--ct-typography-metadata-fontWeight);
      }

      .grid-interaction-grid-fixture-frozen {
        position: sticky;
        left: 0;
        z-index: 2;
        background: var(--ct-colors-surface-2);
        box-shadow: 0.75rem 0 1rem rgba(0, 0, 0, 0.28);
      }

      .grid-interaction-grid-fixture-resize-handle {
        position: absolute;
        inset-block: 0.45rem;
        inset-inline-end: 0.2rem;
        width: 0.25rem;
        border-radius: var(--ct-rounded-sm);
        background: var(--ct-colors-hairline-focus);
      }

      .grid-interaction-grid-fixture-active {
        outline: var(--ct-component-focus-ring-border);
        outline-offset: -0.2rem;
        background: var(--ct-colors-surface-3);
      }

      .grid-interaction-grid-fixture-editor {
        width: 100%;
        min-width: 0;
        border: var(--ct-border-strong);
        border-radius: var(--ct-rounded-sm);
        padding: 0.45rem 0.55rem;
        background: var(--ct-colors-surface-1);
        color: var(--ct-colors-ink);
        font: inherit;
      }

      .grid-interaction-grid-fixture-fill {
        position: absolute;
        right: 0.2rem;
        bottom: 0.2rem;
        width: 0.65rem;
        height: 0.65rem;
        border: 0.15rem solid var(--ct-colors-hairline-focus);
        border-radius: var(--ct-rounded-sm);
        background: var(--ct-colors-surface-1);
      }

      .grid-interaction-grid-fixture-group {
        grid-column: 1 / -1;
        min-height: 3.5rem;
        background: var(--ct-colors-surface-2);
        color: var(--ct-colors-ink-muted);
        font-weight: 600;
      }

      .grid-interaction-grid-fixture-tree-toggle {
        display: inline-grid;
        place-items: center;
        width: 1.35rem;
        height: 1.35rem;
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-sm);
        background: var(--ct-colors-surface-1);
        color: var(--ct-colors-ink);
      }

    `,
    html: `
      <div class="grid-interaction-grid-fixture-table" role="grid" aria-label="Adapter fixture grid">
        <div class="grid-interaction-grid-fixture-row" role="row">
          <div class="grid-interaction-grid-fixture-head grid-interaction-grid-fixture-frozen" role="columnheader" data-fixture-id="visual.fixture.frozen_column">Record</div>
          <div class="grid-interaction-grid-fixture-head" role="columnheader" data-fixture-id="visual.fixture.resize_handle">Summary<span class="grid-interaction-grid-fixture-resize-handle" aria-hidden="true"></span></div>
          <div class="grid-interaction-grid-fixture-head" role="columnheader">State</div>
          <div class="grid-interaction-grid-fixture-head" role="columnheader">Assignee</div>
          <div class="grid-interaction-grid-fixture-head" role="columnheader">Last edit</div>
        </div>
        <div class="grid-interaction-grid-fixture-row" role="row">
          <div class="grid-interaction-grid-fixture-cell grid-interaction-grid-fixture-group" role="rowheader" data-fixture-id="visual.fixture.tree_group_row"><span class="grid-interaction-grid-fixture-tree-toggle" aria-hidden="true">v</span> reviewed group, 2 rows</div>
        </div>
        <div class="grid-interaction-grid-fixture-row" role="row">
          <div class="grid-interaction-grid-fixture-cell grid-interaction-grid-fixture-frozen" role="rowheader">record-1</div>
          <div class="grid-interaction-grid-fixture-cell grid-interaction-grid-fixture-active" role="gridcell" data-fixture-id="visual.fixture.edit_cell"><input class="grid-interaction-grid-fixture-editor" value="Edit cell adapter" aria-label="Summary editor" readonly><span class="grid-interaction-grid-fixture-fill" data-fixture-id="visual.fixture.drag_fill_handle" aria-hidden="true"></span></div>
          <div class="grid-interaction-grid-fixture-cell" role="gridcell">reviewed</div>
          <div class="grid-interaction-grid-fixture-cell" role="gridcell">Analyst</div>
          <div class="grid-interaction-grid-fixture-cell" role="gridcell">saved</div>
        </div>
        <div class="grid-interaction-grid-fixture-row" role="row">
          <div class="grid-interaction-grid-fixture-cell grid-interaction-grid-fixture-frozen" role="rowheader">record-2</div>
          <div class="grid-interaction-grid-fixture-cell" role="gridcell">Frozen column remains pinned</div>
          <div class="grid-interaction-grid-fixture-cell" role="gridcell">rough</div>
          <div class="grid-interaction-grid-fixture-cell" role="gridcell">Unassigned</div>
          <div class="grid-interaction-grid-fixture-cell" role="gridcell">clean</div>
        </div>
      </div>
    `,
  });
}

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
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("VISUALGRIDADAPTER"),
    "browser.grid-interaction grid adapter visual fixture",
  );
  await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("VISUALGRIDADAPTER-ROW"),
    "timeline.activity_utc_text": "2026-05-31T10:00:00Z",
    "timeline.activity_synopsis_text":
      "browser.grid-interaction visual adapter row",
  });

  await navigateVisualApplication(page, `/?incident_id=${incidentId}`);

  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await injectFeP3GridAdapterVisualFixture(page);

  const fixture = page.locator(
    "[data-design-fixture='grid-interaction-grid-adapter']",
  );
  await expect(fixture).toBeVisible();
  for (const fixtureId of [
    "visual.fixture.frozen_column",
    "visual.fixture.resize_handle",
    "visual.fixture.drag_fill_handle",
    "visual.fixture.edit_cell",
    "visual.fixture.tree_group_row",
  ]) {
    await expect(
      fixture.locator(`[data-fixture-id='${fixtureId}']`),
    ).toBeVisible();
  }
  await assertVisualRegression(page, "timeline-grid-adapter-fixtures", fixture);
});
