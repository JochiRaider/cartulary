import {
  assertGroupRowPresentationOnly,
  assertMountedGridRowCountAtMost,
  changeGrouping,
  pasteGridMatrix,
  scrollGridCellIntoView,
  scrollGridToOffset,
  sortByHeader,
} from "@cartulary/test-utils/grid";
import {
  dataTestIdSelector,
  gridGroupingSelectTestId,
  gridGroupRowTestId,
  gridRowGutterTestId,
  gridScrollportSelector,
  gridShellTestId,
  rowCellTestId,
  rowHistoryPanelTestId,
  rowInspectorFieldTestId,
  saveStateTestId,
  surfaceTabTestId,
  systemViewSwitcherTriggerTestId,
  timelineInspectorTestId,
  timelineMutationSubstrateReadyTestId,
  timelineScalarEditorTestId,
  workbookFilterPopoverTriggerTestId,
  workbookInspectorCloseButtonTestId,
  workbookResponsiveBandTestId,
  workbookSortMenuTriggerTestId,
  workbookTopBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
import type { Locator, Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import {
  assessmentsViewSchemaId,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  handoffViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  indicatorsViewSchemaId,
  lessonViewSchemaId,
  notesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
} from "./support/contracts/workbookSurfaces";
import {
  collectionActionsPayload,
  hostRefsFieldKey,
} from "./support/entities/mentions";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createTimelineFillers } from "./support/timeline/fixtures";
import {
  createViewRow,
  queryViewRows,
  type ViewApiRow,
  waitForViewRow,
} from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";

const timelineViewSchemaId = "cartulary.view.timeline.v2";

function stringCell(
  row: { readonly cells?: Record<string, { readonly value?: unknown }> },
  fieldKey: string,
): string {
  return String(row.cells?.[fieldKey]?.value ?? "");
}

function readPostBody(request: { postData: () => string | null }) {
  return JSON.parse(request.postData() ?? "{}") as Record<string, unknown>;
}

type SharedGridAnchorOptions = {
  page: Page;
  incidentId: string;
  viewSchemaId: string;
  row: ViewApiRow;
  fieldKey: string;
  expectedText: string;
  textMode?: "text" | "value";
  rightFieldKey?: string;
  rightFocusTestId?: string;
};

function semanticGridCell(content: Locator): Locator {
  return content.locator("xpath=ancestor::*[@role='gridcell'][1]");
}

async function activateSemanticGridCell(content: Locator): Promise<Locator> {
  const cell = semanticGridCell(content);
  await cell.dispatchEvent("mousedown", { button: 0 });
  await cell.focus();
  return cell;
}

async function waitForBrowserQueryWithRow(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  recordId: string,
) {
  let lastQueryStatus = "no matching query response observed";
  try {
    await page.waitForResponse(
      async (response) => {
        if (
          response.request().method() !== "POST" ||
          !response
            .url()
            .includes(
              `/api/v1/incidents/${incidentId}/views/${viewSchemaId}/query`,
            )
        ) {
          return false;
        }
        lastQueryStatus = `HTTP ${response.status()}`;
        if (!response.ok()) {
          return false;
        }
        const body = (await response.json().catch(() => null)) as {
          data?: { rows?: ViewApiRow[]; view_schema_id?: string };
        } | null;
        const rows = body?.data?.rows ?? [];
        lastQueryStatus = `HTTP ${response.status()}; view_schema_id=${
          body?.data?.view_schema_id ?? "missing"
        }; rows=${rows.map((row) => row.record_id).join(",")}`;
        return rows.some((row) => row.record_id === recordId);
      },
      { timeout: 10_000 },
    );
  } catch (error) {
    throw new Error(
      `${viewSchemaId} browser query did not include row ${recordId}: ${lastQueryStatus}; ${String(
        error,
      )}`,
    );
  }
}

async function readWorkbookDocumentLayout(page: Page) {
  return page.evaluate(
    ({ inspectorSelector, scrollportSelector, shellSelector }) => {
      const shell = document.querySelector<HTMLElement>(shellSelector);
      if (shell === null) {
        throw new Error(`Expected ${shellSelector} to exist`);
      }
      const scrollport = shell.querySelector<HTMLElement>(scrollportSelector);
      if (scrollport === null) {
        throw new Error(`Expected ${scrollportSelector} to exist`);
      }
      const shellRect = shell.getBoundingClientRect();
      const scrollportRect = scrollport.getBoundingClientRect();
      const inspector = document.querySelector<HTMLElement>(inspectorSelector);
      const inspectorRect = inspector?.getBoundingClientRect() ?? null;

      return {
        body: {
          clientHeight: document.body.clientHeight,
          scrollHeight: document.body.scrollHeight,
        },
        documentElement: {
          clientHeight: document.documentElement.clientHeight,
          scrollHeight: document.documentElement.scrollHeight,
        },
        grid: {
          clientHeight: scrollport.clientHeight,
          scrollHeight: scrollport.scrollHeight,
          maxTop: Math.max(
            0,
            scrollport.scrollHeight - scrollport.clientHeight,
          ),
          rect: {
            bottom: Math.round(scrollportRect.bottom),
            height: Math.round(scrollportRect.height),
            top: Math.round(scrollportRect.top),
          },
        },
        inspector:
          inspectorRect === null
            ? null
            : {
                bottom: Math.round(inspectorRect.bottom),
                height: Math.round(inspectorRect.height),
                scrollHeight: inspector?.scrollHeight ?? 0,
                top: Math.round(inspectorRect.top),
              },
        shell: {
          clientHeight: shell.clientHeight,
          scrollHeight: shell.scrollHeight,
          maxTop: Math.max(0, shell.scrollHeight - shell.clientHeight),
          rect: {
            bottom: Math.round(shellRect.bottom),
            height: Math.round(shellRect.height),
            top: Math.round(shellRect.top),
          },
        },
        viewport: {
          innerHeight: window.innerHeight,
          innerWidth: window.innerWidth,
          scrollY: window.scrollY,
        },
      };
    },
    {
      inspectorSelector: dataTestIdSelector(timelineInspectorTestId()),
      scrollportSelector: gridScrollportSelector(),
      shellSelector: dataTestIdSelector(gridShellTestId(timelineViewSchemaId)),
    },
  );
}

function expectWorkbookDocumentBounded(
  layout: Awaited<ReturnType<typeof readWorkbookDocumentLayout>>,
  label: string,
) {
  expect(layout.viewport.scrollY, `${label}: window scrollY`).toBe(0);
  expect(
    layout.documentElement.scrollHeight,
    `${label}: document element scroll height`,
  ).toBeLessThanOrEqual(layout.viewport.innerHeight + 1);
  expect(
    layout.body.scrollHeight,
    `${label}: body scroll height`,
  ).toBeLessThanOrEqual(layout.viewport.innerHeight + 1);
  expect(layout.shell.maxTop, `${label}: grid shell vertical scroll`).toBe(0);
  expect(
    layout.shell.rect.bottom,
    `${label}: shell bottom`,
  ).toBeLessThanOrEqual(layout.viewport.innerHeight + 1);
  expect(
    layout.grid.maxTop,
    `${label}: grid scrollport overflow`,
  ).toBeGreaterThan(0);
  expect(
    layout.grid.rect.bottom,
    `${label}: grid scrollport bottom`,
  ).toBeLessThanOrEqual(layout.viewport.innerHeight + 1);
}

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
    page.getByTestId(workbookTopBarQueryControlsTestId(timelineViewSchemaId)),
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

async function openSharedGridAnchorCell({
  page,
  incidentId,
  viewSchemaId,
  row,
  fieldKey,
  expectedText,
  textMode = "text",
}: SharedGridAnchorOptions): Promise<Locator> {
  const recordId = row.record_id;
  const queryRow = await waitForViewRow(
    page,
    incidentId,
    viewSchemaId,
    recordId,
  );
  expect(
    stringCell(queryRow, fieldKey),
    `${viewSchemaId} default query row ${recordId} should include ${fieldKey}`,
  ).toBe(expectedText);

  const renderedQuery = waitForBrowserQueryWithRow(
    page,
    incidentId,
    viewSchemaId,
    recordId,
  );
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      viewSchemaId,
    )}`,
  );
  await renderedQuery;
  if (viewSchemaId === timelineViewSchemaId) {
    await expect(
      page.getByTestId(timelineMutationSubstrateReadyTestId()),
    ).toBeVisible();
  }

  await scrollGridCellIntoView({
    cellKey: fieldKey,
    page,
    recordId,
    surface: viewSchemaId,
  });
  const cell = page.getByTestId(rowCellTestId(recordId, fieldKey));
  if (textMode === "value") {
    await expect(
      cell,
      `${viewSchemaId} rendered cell ${recordId}:${fieldKey}`,
    ).toHaveValue(expectedText);
  } else {
    await expect(
      cell,
      `${viewSchemaId} rendered cell ${recordId}:${fieldKey}`,
    ).toHaveText(expectedText);
  }
  return cell;
}

async function expectSharedGridAnchorSurface(options: SharedGridAnchorOptions) {
  const { page, viewSchemaId, row, fieldKey, rightFieldKey, rightFocusTestId } =
    options;
  const recordId = row.record_id;
  const content = await openSharedGridAnchorCell(options);
  const cell = await activateSemanticGridCell(content);
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${viewSchemaId}:${recordId}:${fieldKey}`,
  );

  await page.keyboard.press("ArrowRight");
  if (rightFieldKey) {
    await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
      `${viewSchemaId}:${recordId}:${rightFieldKey}`,
    );
    if (rightFocusTestId) {
      await expect(
        semanticGridCell(page.getByTestId(rightFocusTestId)),
      ).toBeFocused();
    } else {
      await expect(
        semanticGridCell(
          page.getByTestId(rowCellTestId(recordId, rightFieldKey)),
        ),
      ).toBeFocused();
    }
    return cell;
  }
  await expect(page.getByTestId("workbook-focus-anchor")).toContainText(
    `${viewSchemaId}:${recordId}:`,
  );
  await expect(page.getByTestId("workbook-focus-anchor")).not.toHaveText(
    `${viewSchemaId}:${recordId}:${fieldKey}`,
  );
  return cell;
}

test("Phase 9 E-9-01 keyboard shortcuts keep workbook grid anchors without module switching", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E901"),
    "Phase 9 E-9-01 keyboard contract",
  );
  const alpha = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e901-alpha"),
    "timeline.date_entered_text": "Phase 9 alpha",
    [hostRefsFieldKey]: collectionActionsPayload(["Phase9Host?"]),
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect(
    page.getByTestId(
      rowCellTestId(alpha.record_id as string, "timeline.date_entered_text"),
    ),
  ).toHaveText("Phase 9 alpha");
  const initialURL = page.url();

  const alphaSummary = page.getByTestId(
    rowCellTestId(alpha.record_id as string, "timeline.date_entered_text"),
  );
  const alphaSummaryCell = await activateSemanticGridCell(alphaSummary);
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.date_entered_text`,
  );

  await page.keyboard.press("Tab");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.date_entered_text`,
  );
  await expect(alphaSummaryCell).not.toBeFocused();

  await openTimelineInspector(page, alpha.record_id as string);
  await expect(page.getByTestId(timelineInspectorTestId())).toContainText(
    "Phase9Host?",
  );
  expect(page.url()).toBe(initialURL);
  await scrollGridCellIntoView({
    cellKey: "timeline.analyst_text",
    page,
    recordId: alpha.record_id as string,
    surface: timelineViewSchemaId,
  });
  const alphaAnalyst = page.getByTestId(
    rowCellTestId(alpha.record_id as string, "timeline.analyst_text"),
  );
  await activateSemanticGridCell(alphaAnalyst);
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.analyst_text`,
  );

  await page.keyboard.press("Alt+H");
  await expect(page.getByTestId(rowHistoryPanelTestId())).toContainText(
    String(alpha.record_id),
  );
  expect(page.url()).toBe(initialURL);

  await page.keyboard.press("Control+V");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.analyst_text`,
  );
  await page.keyboard.press("Escape");
  if (await page.getByTestId(rowHistoryPanelTestId()).isVisible()) {
    await page.keyboard.press("Escape");
  }
  await expect(page.getByTestId(rowHistoryPanelTestId())).toHaveCount(0);
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.analyst_text`,
  );
  expect(page.url()).toBe(initialURL);
});

test("FE-B-P9-LAYOUT-01 keeps the incident workbook inside the browser viewport and delegates overflow to workbook panels", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E9LAYOUT"),
    "Phase 9 bounded workbook layout",
  );
  await createTimelineFillers(page, incidentId, "Phase 9 layout row", 72);

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();

  await expect
    .poll(async () => {
      const layout = await readWorkbookDocumentLayout(page);
      return layout.grid.maxTop > 0;
    })
    .toBe(true);

  await page.evaluate(() => window.scrollTo(0, 10_000));
  expectWorkbookDocumentBounded(
    await readWorkbookDocumentLayout(page),
    "initial 1280x720",
  );
  await expectWideWorkbookTopBarChrome(page);

  await page.setViewportSize({ width: 1280, height: 560 });
  await expectWideWorkbookTopBarChrome(page);
  await expect
    .poll(async () => {
      const layout = await readWorkbookDocumentLayout(page);
      return (
        layout.viewport.innerHeight === 560 &&
        layout.documentElement.scrollHeight <=
          layout.viewport.innerHeight + 1 &&
        layout.body.scrollHeight <= layout.viewport.innerHeight + 1 &&
        layout.grid.maxTop > 0
      );
    })
    .toBe(true);
  await page.evaluate(() => window.scrollTo(0, 10_000));
  expectWorkbookDocumentBounded(
    await readWorkbookDocumentLayout(page),
    "vertical resized 1280x560",
  );
  await page.setViewportSize({ width: 1280, height: 720 });
  await expectWideWorkbookTopBarChrome(page);

  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();
  await expect(page.getByTestId(timelineInspectorTestId())).toBeVisible();
  await page.evaluate(() => window.scrollTo(0, 10_000));
  const inspectorLayout = await readWorkbookDocumentLayout(page);
  expectWorkbookDocumentBounded(inspectorLayout, "inspector open");
  expect(inspectorLayout.inspector?.bottom ?? 0).toBeLessThanOrEqual(
    inspectorLayout.viewport.innerHeight + 1,
  );

  await page.setViewportSize({ width: 1024, height: 640 });
  await expect
    .poll(async () => {
      const layout = await readWorkbookDocumentLayout(page);
      return (
        layout.viewport.innerHeight === 640 &&
        layout.documentElement.scrollHeight <=
          layout.viewport.innerHeight + 1 &&
        layout.body.scrollHeight <= layout.viewport.innerHeight + 1 &&
        layout.grid.maxTop > 0
      );
    })
    .toBe(true);
  await page.evaluate(() => window.scrollTo(0, 10_000));
  expectWorkbookDocumentBounded(
    await readWorkbookDocumentLayout(page),
    "resized 1024x640",
  );
});

test("FE-B-P10-02 Verify full keyboard/clipboard contract: one-click edit, copy, paste, exact-range fill-down, frozen columns, virtual scroll, group rows, focus restoration, and Esc priority ladder.", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEBP1002"),
    "FE-B-P10-02 workbook keyboard contract",
  );
  const alpha = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("fe-b-p10-02-alpha"),
    "timeline.activity_synopsis_text": "FE-B-P10-02 Alpha",
    "timeline.raw_activity_text": "FE-B-P10-02 Alpha details",
  });
  const beta = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("fe-b-p10-02-beta"),
    "timeline.activity_synopsis_text": "FE-B-P10-02 Beta",
    "timeline.raw_activity_text": "FE-B-P10-02 Beta details",
  });
  await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("fe-b-p10-02-gamma"),
    "timeline.activity_synopsis_text": "FE-B-P10-02 Gamma",
    "timeline.raw_activity_text": "FE-B-P10-02 Gamma details",
  });
  await createTimelineFillers(page, incidentId, "FE-B-P10-02 filler", 24);
  const virtualTarget = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("fe-b-p10-02-virtual"),
      "timeline.activity_synopsis_text": "ZZZ FE-B-P10-02 virtual target",
      "timeline.raw_activity_text": "FE-B-P10-02 virtual details",
    },
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();

  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: alpha.record_id,
    surface: timelineViewSchemaId,
  });
  const alphaSummary = page.getByTestId(
    rowCellTestId(alpha.record_id, "timeline.activity_synopsis_text"),
  );
  await expect(alphaSummary).toHaveText("FE-B-P10-02 Alpha");
  const alphaSummaryCell = await activateSemanticGridCell(alphaSummary);
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.activity_synopsis_text`,
  );
  const copiedText = await alphaSummaryCell.evaluate((element) => {
    const data = new DataTransfer();
    const event = new ClipboardEvent("copy", {
      bubbles: true,
      cancelable: true,
      clipboardData: data,
    });
    element.dispatchEvent(event);
    return data.getData("text/plain");
  });
  expect(copiedText).toBe("FE-B-P10-02 Alpha");

  await page.keyboard.press("ArrowRight");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.data_source_text`,
  );
  await page.keyboard.press("ArrowLeft");
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.activity_synopsis_text`,
  );
  await expect(alphaSummaryCell).toBeFocused();

  await alphaSummary.click();
  const alphaSummaryEditor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.activity_synopsis_text",
      recordId: alpha.record_id,
      surface: "grid",
    }),
  );
  await alphaSummaryEditor.fill("FE-B-P10-02 dirty draft");
  await alphaSummaryEditor.press("Escape");
  await expect(alphaSummary).toHaveText("FE-B-P10-02 Alpha");
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

  await openTimelineInspector(page, alpha.record_id);
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: alpha.record_id,
    surface: timelineViewSchemaId,
  });
  await alphaSummary.focus();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.activity_synopsis_text`,
  );
  const inspectorDetails = page.getByTestId(
    rowInspectorFieldTestId(alpha.record_id, "timeline.raw_activity_text"),
  );
  await expect(inspectorDetails).toHaveValue("FE-B-P10-02 Alpha details");
  await inspectorDetails.focus();
  await inspectorDetails.fill("FE-B-P10-02 inspector dirty draft");
  await inspectorDetails.press("Escape");
  await expect(inspectorDetails).toHaveValue("FE-B-P10-02 Alpha details");
  await inspectorDetails.press("Escape");
  await expect(alphaSummaryCell).toBeFocused();
  await alphaSummaryCell.press("Escape");
  await expect(inspectorDetails).toHaveCount(0);
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await expect(alphaSummaryCell).toBeFocused();
  await alphaSummaryCell.press("Escape");
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await expect(alphaSummaryCell).toBeFocused();

  const pasteRequest = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
        ),
  );
  const pasteResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
        ),
  );
  await pasteGridMatrix({
    fieldKey: "timeline.activity_synopsis_text",
    matrix: [["FE-B-P10-02 pasted Beta", "fe-b-p10-02-host-token"]],
    page,
    recordId: beta.record_id,
    surface: timelineViewSchemaId,
  });
  expect(readPostBody(await pasteRequest)).toMatchObject({
    columns: ["timeline.activity_synopsis_text", "timeline.data_source_text"],
    start_field_key: "timeline.activity_synopsis_text",
    targets: [
      {
        kind: "record",
        record_id: beta.record_id,
      },
    ],
    view_schema_id: timelineViewSchemaId,
  });
  await expect((await pasteResponse).ok()).toBeTruthy();
  await expect(
    page.getByTestId(
      rowCellTestId(beta.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("FE-B-P10-02 pasted Beta");
  const betaAfterPaste = await waitForViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    beta.record_id,
  );
  expect(stringCell(betaAfterPaste, "timeline.activity_synopsis_text")).toBe(
    "FE-B-P10-02 pasted Beta",
  );

  await scrollGridCellIntoView({
    cellKey: "timeline.raw_activity_text",
    page,
    recordId: beta.record_id,
    surface: timelineViewSchemaId,
  });
  const [fillSourceRecordId, fillTargetRecordId] = await page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .evaluate(
      (grid, knownSourceRecordIds) => {
        const recordIds = Array.from(
          grid.querySelectorAll<HTMLElement>(
            '[role="row"][data-grid-record-id]',
          ),
          (row) => row.dataset.gridRecordId,
        ).filter((recordId): recordId is string => recordId !== undefined);
        for (let index = 0; index < recordIds.length - 1; index += 1) {
          const sourceRecordId = recordIds[index];
          const targetRecordId = recordIds[index + 1];
          if (
            sourceRecordId !== undefined &&
            targetRecordId !== undefined &&
            knownSourceRecordIds.includes(sourceRecordId)
          ) {
            return [sourceRecordId, targetRecordId] as [string, string];
          }
        }
        throw new Error(
          "Expected a visible non-empty source with a committed fill target",
        );
      },
      [alpha.record_id, beta.record_id],
    );
  const fillSource = await waitForViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    fillSourceRecordId,
  );
  const fillTarget = await waitForViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    fillTargetRecordId,
  );
  const fillSourceValue = stringCell(fillSource, "timeline.raw_activity_text");
  const fillSourceDisplay = page.getByTestId(
    rowCellTestId(fillSourceRecordId, "timeline.raw_activity_text"),
  );
  const fillSourceCell = await activateSemanticGridCell(fillSourceDisplay);
  const fillRequests: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      request
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/bulk-mutations`,
        )
    ) {
      fillRequests.push(request.postData() ?? "");
    }
  });
  const fillHandle = page.locator(".rdg-cell-drag-handle");
  await expect(fillHandle).toHaveAttribute(
    "aria-label",
    "Drag to fill this value",
  );
  await fillHandle.evaluate((handle) => {
    handle.dispatchEvent(new MouseEvent("dblclick", { bubbles: true }));
  });
  await page.evaluate(
    () =>
      new Promise<void>((resolve) =>
        requestAnimationFrame(() => requestAnimationFrame(() => resolve())),
      ),
  );
  expect(fillRequests).toHaveLength(0);

  await fillSourceCell.press("Shift+ArrowDown");
  const fillRequest = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/bulk-mutations`,
        ),
  );
  const fillResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/bulk-mutations`,
        ),
  );
  await fillSourceCell.press("Control+d");
  expect((await fillResponse).ok()).toBeTruthy();
  expect(readPostBody(await fillRequest)).toMatchObject({
    field_key: "timeline.raw_activity_text",
    kind: "fill_down_v1",
    targets: [
      {
        base_row_version: fillTarget.row_version,
        record_id: fillTarget.record_id,
      },
    ],
    value: fillSourceValue,
    view_schema_id: timelineViewSchemaId,
  });
  const betaBeforeReview = await waitForViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    beta.record_id,
  );

  const reviewResponse = await page.request.post(
    `${apiBase}/api/v1/records/${beta.record_id}/mark-reviewed`,
    {
      headers: await csrfHeaders(page),
      data: {
        base_row_version: betaBeforeReview.row_version,
        client_txn_id: uniqueTxn("fe-b-p10-02-review-beta"),
        reason: "FE-B-P10-02 grouping setup",
      },
    },
  );
  expect(reviewResponse.ok()).toBeTruthy();
  const reviewedBeta = await waitForViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    beta.record_id,
  );
  expect(stringCell(reviewedBeta, "timeline.capture_state")).toBe("reviewed");
  await page.reload();
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await scrollGridCellIntoView({
    cellKey: "timeline.capture_state",
    page,
    recordId: beta.record_id,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(rowCellTestId(beta.record_id, "timeline.capture_state")),
  ).toHaveText("reviewed");
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
  const reviewedGroupTestId = gridGroupRowTestId(
    timelineViewSchemaId,
    "timeline.capture_state",
    "reviewed",
  );
  await expect(page.getByTestId(reviewedGroupTestId)).toBeVisible();
  await assertGroupRowPresentationOnly({
    groupTestId: reviewedGroupTestId,
    page,
    surface: timelineViewSchemaId,
  });

  await sortByHeader(
    page,
    timelineViewSchemaId,
    "timeline.activity_synopsis_text",
  );
  await scrollGridToOffset(page, timelineViewSchemaId, 0);
  await assertMountedGridRowCountAtMost({
    maxRows: 48,
    page,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(virtualTarget.record_id, "timeline.activity_synopsis_text"),
    ),
  ).not.toBeVisible();
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: virtualTarget.record_id,
    surface: timelineViewSchemaId,
    timeoutMs: 6_000,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(virtualTarget.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toBeVisible();
  await page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator(gridScrollportSelector())
    .evaluate((element) => {
      element.scrollLeft = 480;
      element.dispatchEvent(new Event("scroll", { bubbles: true }));
    });
  const frozenGutter = page.getByTestId(
    gridRowGutterTestId(timelineViewSchemaId, virtualTarget.record_id),
  );
  await expect(frozenGutter).toHaveCSS("position", "sticky");
  await expect(frozenGutter).toHaveCSS("left", "0px");
});

test("Phase 9 E-9-GRIDANCHORS-01 shared grid keyboard anchors stay stable across workbook cells", async ({
  page,
  workerAdmin,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E9GRID01"),
    "Phase 9 E-9-GRIDANCHORS-01 keyboard anchor semantics",
  );

  let host: ViewApiRow | undefined;

  await test.step("Timeline anchor", async () => {
    const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("e9grid01"),
      "timeline.activity_synopsis_text": "Phase 9 grid anchor",
    });
    const summary = await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: timelineViewSchemaId,
      row,
      fieldKey: "timeline.activity_synopsis_text",
      expectedText: "Phase 9 grid anchor",
      rightFieldKey: "timeline.data_source_text",
      rightFocusTestId: rowCellTestId(
        row.record_id,
        "timeline.data_source_text",
      ),
    });
    await page.keyboard.press("ArrowLeft");
    await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
      `${timelineViewSchemaId}:${row.record_id}:timeline.activity_synopsis_text`,
    );
    await expect(summary).toBeFocused();
  });

  await test.step("Hosts anchor", async () => {
    host = await createViewRow(page, incidentId, hostsViewSchemaId, {
      client_txn_id: uniqueTxn("e9grid01-host"),
      "host.display_name": "Phase 9 host anchor",
      "host.hostname": "phase9-host.example.test",
    });
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: hostsViewSchemaId,
      row: host,
      fieldKey: "host.display_name",
      expectedText: "Phase 9 host anchor",
      rightFieldKey: "host.hostname",
    });
  });

  await test.step("Identities anchor", async () => {
    const identity = await createViewRow(
      page,
      incidentId,
      identitiesViewSchemaId,
      {
        client_txn_id: uniqueTxn("e9grid01-identity"),
        "identity.display_name": "Phase 9 identity anchor",
        "identity.upn": "phase9.identity@example.test",
      },
    );
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: identitiesViewSchemaId,
      row: identity,
      fieldKey: "identity.display_name",
      expectedText: "Phase 9 identity anchor",
    });
  });

  await test.step("Assessments anchor", async () => {
    if (!host) {
      throw new Error(
        "Host row must be created before assessment anchor step.",
      );
    }
    const assessment = await createViewRow(
      page,
      incidentId,
      assessmentsViewSchemaId,
      {
        client_txn_id: uniqueTxn("e9grid01-assessment"),
        "assessment.subject_ref": host.record_id,
        "assessment.subject_type": "host",
        "assessment.assessment_state": "confirmed",
        "assessment.confidence_score": 55,
        "assessment.rationale": "Phase 9 assessment anchor",
      },
    );
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: assessmentsViewSchemaId,
      row: assessment,
      fieldKey: "assessment.assessment_state",
      expectedText: "confirmed",
    });
  });

  await test.step("Task Requests anchor", async () => {
    const task = await createViewRow(
      page,
      incidentId,
      taskRequestsViewSchemaId,
      {
        client_txn_id: uniqueTxn("e9grid01-task"),
        "task.title": "Phase 9 task request anchor",
        "task.task_kind": "collection",
      },
    );
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: taskRequestsViewSchemaId,
      row: task,
      fieldKey: "task.title",
      expectedText: "Phase 9 task request anchor",
      rightFieldKey: "task.status",
    });
  });

  await test.step("Decisions anchor", async () => {
    const decision = await createViewRow(
      page,
      incidentId,
      decisionsViewSchemaId,
      {
        client_txn_id: uniqueTxn("e9grid01-decision"),
        "decision.summary": "Phase 9 decision anchor",
        "decision.decision_type": "containment",
        "decision.rationale": "Phase 9 decision grid anchor rationale",
      },
    );
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: decisionsViewSchemaId,
      row: decision,
      fieldKey: "decision.summary",
      expectedText: "Phase 9 decision anchor",
      rightFieldKey: "decision.status",
    });
  });

  await test.step("Notes anchor", async () => {
    const note = await createViewRow(page, incidentId, notesViewSchemaId, {
      client_txn_id: uniqueTxn("e9grid01-note"),
      "note.title": "Phase 9 note anchor",
      "note.body": "Phase 9 generic surface anchor body",
    });
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: notesViewSchemaId,
      row: note,
      fieldKey: "note.title",
      expectedText: "Phase 9 note anchor",
    });
  });

  await test.step("Comm Log anchor", async () => {
    const comm = await createViewRow(page, incidentId, commLogViewSchemaId, {
      client_txn_id: uniqueTxn("e9grid01-comm"),
      "comm_log.comm_type": "briefing",
      "comm_log.audience": "Phase 9 grid audience",
      "comm_log.channel_or_meeting": "Grid bridge",
      "comm_log.summary": "Phase 9 comm log anchor",
    });
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: commLogViewSchemaId,
      row: comm,
      fieldKey: "comm_log.summary",
      expectedText: "Phase 9 comm log anchor",
    });
  });

  await test.step("Handoff anchor", async () => {
    const handoff = await createViewRow(page, incidentId, handoffViewSchemaId, {
      client_txn_id: uniqueTxn("e9grid01-handoff"),
      "handoff.incoming_owner_user_id": workerAdmin.user_id,
      "handoff.current_state_summary": "Phase 9 handoff anchor",
    });
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: handoffViewSchemaId,
      row: handoff,
      fieldKey: "handoff.current_state_summary",
      expectedText: "Phase 9 handoff anchor",
    });
  });

  await test.step("Status Review anchor", async () => {
    const statusReview = await createViewRow(
      page,
      incidentId,
      statusReviewViewSchemaId,
      {
        client_txn_id: uniqueTxn("e9grid01-status"),
        "status_review.current_state_summary": "Phase 9 status review anchor",
      },
    );
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: statusReviewViewSchemaId,
      row: statusReview,
      fieldKey: "status_review.current_state_summary",
      expectedText: "Phase 9 status review anchor",
    });
  });

  await test.step("Lesson anchor", async () => {
    const lesson = await createViewRow(page, incidentId, lessonViewSchemaId, {
      client_txn_id: uniqueTxn("e9grid01-lesson"),
      "lesson.summary": "Phase 9 lesson anchor",
    });
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: lessonViewSchemaId,
      row: lesson,
      fieldKey: "lesson.summary",
      expectedText: "Phase 9 lesson anchor",
    });
  });

  await test.step("Indicators anchor", async () => {
    const indicator = await createViewRow(
      page,
      incidentId,
      indicatorsViewSchemaId,
      {
        client_txn_id: uniqueTxn("e9grid01-indicator"),
        "indicator.indicator_type": "ipv4_addr",
        "indicator.value_kind": "atomic",
        "indicator.display_value": "203.0.113.91",
      },
    );
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: indicatorsViewSchemaId,
      row: indicator,
      fieldKey: "indicator.indicator_type",
      expectedText: "ipv4_addr",
      rightFieldKey: "indicator.value_kind",
    });
  });
});

test("Phase 9 E-9-GRIDHOST-01 Host entity-origin clipboard paste reuses exact matches and creates stubs", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E9GRIDHOSTPASTE"),
    "Phase 9 E-9-GRIDHOST-01 host paste",
  );
  const existing = await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e9grid-host-existing"),
    "host.display_name": "Phase 9 reusable host",
    "host.hostname": "phase9-host-reuse.example.test",
  });
  const postURLs: string[] = [];
  page.on("request", (request) => {
    if (request.method() === "POST") {
      postURLs.push(request.url());
    }
  });

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      hostsViewSchemaId,
    )}`,
  );
  const displayName = page.getByTestId(
    rowCellTestId(existing.record_id as string, "host.display_name"),
  );
  await expect(displayName).toHaveText("Phase 9 reusable host");
  await activateSemanticGridCell(displayName);
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${hostsViewSchemaId}:${existing.record_id}:host.display_name`,
  );

  const pasteResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .includes(
          `/api/v1/incidents/${incidentId}/views/${hostsViewSchemaId}/clipboard-paste`,
        ),
  );
  await displayName.evaluate((element) => {
    const data = new DataTransfer();
    data.setData(
      "text/plain",
      [
        "Phase 9 pasted host reuse\tphase9-host-reuse.example.test",
        "Phase 9 pasted host create\tphase9-host-create.example.test",
      ].join("\n"),
    );
    element.dispatchEvent(
      new ClipboardEvent("paste", {
        bubbles: true,
        cancelable: true,
        clipboardData: data,
      }),
    );
  });
  await expect((await pasteResponse).ok()).toBeTruthy();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${hostsViewSchemaId}:${existing.record_id}:host.display_name`,
  );
  await expect(
    page.getByTestId(
      rowCellTestId(existing.record_id as string, "host.display_name"),
    ),
  ).toHaveText("Phase 9 pasted host reuse");

  const rows = await queryViewRows(page, incidentId, hostsViewSchemaId);
  const reused = rows.find((row) => row.record_id === existing.record_id);
  expect(stringCell(reused ?? {}, "host.display_name")).toBe(
    "Phase 9 pasted host reuse",
  );
  const created = rows.find(
    (row) =>
      row.record_id !== existing.record_id &&
      stringCell(row, "host.hostname") === "phase9-host-create.example.test",
  );
  expect(created).toBeTruthy();
  if (created) {
    await expect(
      page.getByTestId(rowCellTestId(created.record_id, "host.display_name")),
    ).toHaveText("Phase 9 pasted host create");
  }
  expect(postURLs.some((url) => url.includes("/imports"))).toBeFalsy();
});

test("Phase 9 E-9-GRIDIDENTITY-01 Identity entity-origin clipboard paste reuses exact matches and creates stubs", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E9GRIDIDENTITYPASTE"),
    "Phase 9 E-9-GRIDIDENTITY-01 identity paste",
  );
  const existing = await createViewRow(
    page,
    incidentId,
    identitiesViewSchemaId,
    {
      client_txn_id: uniqueTxn("e9grid-identity-existing"),
      "identity.display_name": "Phase 9 reusable identity",
      "identity.upn": "phase9.identity.reuse@example.test",
    },
  );
  const postURLs: string[] = [];
  page.on("request", (request) => {
    if (request.method() === "POST") {
      postURLs.push(request.url());
    }
  });

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      identitiesViewSchemaId,
    )}`,
  );
  const displayName = page.getByTestId(
    rowCellTestId(existing.record_id as string, "identity.display_name"),
  );
  await expect(displayName).toHaveText("Phase 9 reusable identity");
  await activateSemanticGridCell(displayName);
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${identitiesViewSchemaId}:${existing.record_id}:identity.display_name`,
  );

  const pasteResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .includes(
          `/api/v1/incidents/${incidentId}/views/${identitiesViewSchemaId}/clipboard-paste`,
        ),
  );
  await displayName.evaluate((element) => {
    const data = new DataTransfer();
    data.setData(
      "text/plain",
      [
        "Phase 9 pasted identity reuse\tphase9.identity.reuse@example.test",
        "Phase 9 pasted identity create\tphase9.identity.create@example.test",
      ].join("\n"),
    );
    element.dispatchEvent(
      new ClipboardEvent("paste", {
        bubbles: true,
        cancelable: true,
        clipboardData: data,
      }),
    );
  });
  await expect((await pasteResponse).ok()).toBeTruthy();
  await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
    `${identitiesViewSchemaId}:${existing.record_id}:identity.display_name`,
  );
  await expect(
    page.getByTestId(
      rowCellTestId(existing.record_id as string, "identity.display_name"),
    ),
  ).toHaveText("Phase 9 pasted identity reuse");

  const rows = await queryViewRows(page, incidentId, identitiesViewSchemaId);
  const reused = rows.find((row) => row.record_id === existing.record_id);
  expect(stringCell(reused ?? {}, "identity.display_name")).toBe(
    "Phase 9 pasted identity reuse",
  );
  const created = rows.find(
    (row) =>
      row.record_id !== existing.record_id &&
      stringCell(row, "identity.upn") === "phase9.identity.create@example.test",
  );
  expect(created).toBeTruthy();
  if (created) {
    await expect(
      page.getByTestId(
        rowCellTestId(created.record_id, "identity.display_name"),
      ),
    ).toHaveText("Phase 9 pasted identity create");
  }
  expect(postURLs.some((url) => url.includes("/imports"))).toBeFalsy();
});
