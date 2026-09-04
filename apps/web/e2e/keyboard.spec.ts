import type { ViewRow } from "@cartulary/protocol-ts/http";
import {
  assertGroupRowPresentationOnly,
  assertMountedGridRowCountAtMost,
  changeGrouping,
  isTestIdVisibleWithinGridViewport,
  pasteGridMatrix,
  scrollGridCellIntoView,
  scrollGridToOffset,
  sortByHeader,
} from "@cartulary/test-utils/grid";
import {
  dataTestIdSelector,
  gridFillHandleSelector,
  gridGroupingSelectTestId,
  gridGroupRowTestId,
  gridRowGutterTestId,
  gridScrollportSelector,
  gridShellTestId,
  rowCellTestId,
  rowHistoryPanelTestId,
  rowInspectorFieldTestId,
  savedViewActionMenuTriggerTestId,
  savedViewSelectorTestId,
  saveStateTestId,
  surfaceTabTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  timelineInspectorSectionTestId,
  timelineInspectorTestId,
  timelineMutationSubstrateReadyTestId,
  timelineScalarEditorTestId,
  workbookAddRowButtonTestId,
  workbookFilterPopoverTriggerTestId,
  workbookFocusAnchorTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorToggleTestId,
  workbookPresenceSummaryTestId,
  workbookResponsiveBandTestId,
  workbookShellSlotTestId,
  workbookSortMenuTriggerTestId,
  workbookSurfacesMenuOptionTestId,
  workbookSurfacesMenuTestId,
  workbookSurfacesMenuTriggerTestId,
  workbookViewBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
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
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Locator, Page, Request, Response } from "@playwright/test";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
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
  waitForViewRow,
} from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";
import {
  createSavedView,
  selectSavedView,
} from "./support/workbook/savedViews";

function stringCell(
  row: { readonly cells?: Record<string, { readonly value?: unknown }> },
  fieldKey: string,
): string {
  return String(row.cells?.[fieldKey]?.value ?? "");
}

function readPostBody(request: { postData: () => string | null }) {
  return JSON.parse(request.postData() ?? "{}") as Record<string, unknown>;
}

const browserMutationActionTimeoutMs = 10_000;

async function observeBoundedPost(options: {
  action: () => Promise<void>;
  operation: string;
  page: Page;
  pathSuffix: string;
}): Promise<{
  request: Request;
  requestCount: number;
  response: Response;
}> {
  let requestCount = 0;
  let resolveResponse: (response: Response) => void = () => {};
  let rejectResponse: (error: Error) => void = () => {};
  const responsePromise = new Promise<Response>((resolve, reject) => {
    resolveResponse = resolve;
    rejectResponse = reject;
  });
  const matches = (url: string, method: string) =>
    method === "POST" && url.endsWith(options.pathSuffix);
  const onRequest = (request: Request) => {
    if (matches(request.url(), request.method())) requestCount += 1;
  };
  const onResponse = (response: Response) => {
    if (matches(response.url(), response.request().method())) {
      resolveResponse(response);
    }
  };
  options.page.on("request", onRequest);
  options.page.on("response", onResponse);
  const timeout = setTimeout(
    () =>
      rejectResponse(
        new Error(
          `${options.operation} response was not observed within ${browserMutationActionTimeoutMs}ms`,
        ),
      ),
    browserMutationActionTimeoutMs,
  );
  try {
    const [, response] = await Promise.all([options.action(), responsePromise]);
    return {
      request: response.request(),
      requestCount,
      response,
    };
  } catch (error) {
    const focusAnchor = await options.page
      .getByTestId(workbookFocusAnchorTestId())
      .textContent()
      .catch(() => null);
    const saveState = await options.page
      .getByTestId(saveStateTestId())
      .textContent()
      .catch(() => null);
    throw new Error(
      `${options.operation} failed: request_seen=${requestCount > 0}; request_count=${requestCount}; focus_anchor=${focusAnchor ?? "unavailable"}; save_state=${saveState ?? "unavailable"}; ${error instanceof Error ? error.message : String(error)}`,
    );
  } finally {
    clearTimeout(timeout);
    options.page.off("request", onRequest);
    options.page.off("response", onResponse);
  }
}

async function countPostRequestsDuring(options: {
  action: () => Promise<void>;
  page: Page;
  pathSuffix: string;
}): Promise<number> {
  let requestCount = 0;
  const onRequest = (request: Request) => {
    if (
      request.method() === "POST" &&
      request.url().endsWith(options.pathSuffix)
    ) {
      requestCount += 1;
    }
  };
  options.page.on("request", onRequest);
  try {
    await options.action();
    return requestCount;
  } finally {
    options.page.off("request", onRequest);
  }
}

type SharedGridAnchorOptions = {
  page: Page;
  incidentId: string;
  viewSchemaId: string;
  row: ViewRow;
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
          data?: { rows?: ViewRow[]; view_schema_id?: string };
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

async function readWorkbookViewBarGeometry(page: Page) {
  return page.evaluate(
    ({
      actionMenuSelector,
      addRowSelector,
      filterSelector,
      groupingSelector,
      inspectorSelector,
      queryControlsSelector,
      savedViewSelector,
      sortSelector,
    }) => {
      const select = (selector: string) =>
        document.querySelector<HTMLElement>(selector);
      const requireElement = (element: HTMLElement | null, label: string) => {
        if (element === null) throw new Error(`Expected ${label} to exist`);
        return element;
      };
      const queryControls = requireElement(
        select(queryControlsSelector),
        "query controls",
      );
      const viewBar = requireElement(
        queryControls.closest<HTMLElement>(
          'section[aria-label="Workbook query and action controls"]',
        ),
        "workbook view bar",
      );
      const columns = requireElement(
        Array.from(
          queryControls.querySelectorAll<HTMLButtonElement>("button"),
        ).find((button) => button.textContent?.trim() === "Columns") ?? null,
        "Columns button",
      );
      const chipButtons = Array.from(
        queryControls.querySelectorAll<HTMLButtonElement>(
          '[role="toolbar"][aria-label="Active query chips"] button[data-query-entry-key]',
        ),
      );
      const controls: ReadonlyArray<readonly [string, HTMLElement]> = [
        ["saved-view", requireElement(select(savedViewSelector), "saved view")],
        [
          "saved-view-actions",
          requireElement(select(actionMenuSelector), "saved view actions"),
        ],
        ["sort", requireElement(select(sortSelector), "Sort button")],
        ["group", requireElement(select(groupingSelector), "Group select")],
        ["filters", requireElement(select(filterSelector), "Filters button")],
        ["columns", columns],
        ...chipButtons.map(
          (button, index) => [`chip-${index}`, button] as const,
        ),
        [
          "inspector",
          requireElement(select(inspectorSelector), "Inspector button"),
        ],
        ["add-row", requireElement(select(addRowSelector), "Add row button")],
      ];
      const viewBarRect = viewBar.getBoundingClientRect();
      const filter = requireElement(select(filterSelector), "Filters button");

      return {
        capacity: Number(queryControls.dataset.queryChipCapacity ?? "-1"),
        controls: controls.map(([name, element]) => {
          const rect = element.getBoundingClientRect();
          return {
            bottom: Math.round(rect.bottom * 100) / 100,
            left: Math.round(rect.left * 100) / 100,
            name,
            right: Math.round(rect.right * 100) / 100,
            top: Math.round(rect.top * 100) / 100,
            width: Math.round(rect.width * 100) / 100,
          };
        }),
        document: {
          clientWidth: document.documentElement.clientWidth,
          scrollWidth: document.documentElement.scrollWidth,
        },
        filter: {
          accessibleName: filter.getAttribute("aria-label"),
          clientWidth: filter.clientWidth,
          scrollWidth: filter.scrollWidth,
          visibleText: filter.innerText.trim(),
        },
        hiddenCount: Number(queryControls.dataset.hiddenQueryChipCount ?? "-1"),
        viewBar: {
          bottom: Math.round(viewBarRect.bottom * 100) / 100,
          left: Math.round(viewBarRect.left * 100) / 100,
          right: Math.round(viewBarRect.right * 100) / 100,
          top: Math.round(viewBarRect.top * 100) / 100,
        },
      };
    },
    {
      actionMenuSelector: dataTestIdSelector(
        savedViewActionMenuTriggerTestId(timelineViewSchemaId),
      ),
      addRowSelector: dataTestIdSelector(
        workbookAddRowButtonTestId(timelineViewSchemaId),
      ),
      filterSelector: dataTestIdSelector(
        workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
      ),
      groupingSelector: dataTestIdSelector(
        gridGroupingSelectTestId(timelineViewSchemaId),
      ),
      inspectorSelector: dataTestIdSelector(
        workbookInspectorToggleTestId(timelineViewSchemaId),
      ),
      queryControlsSelector: dataTestIdSelector(
        workbookViewBarQueryControlsTestId(timelineViewSchemaId),
      ),
      savedViewSelector: dataTestIdSelector(
        savedViewSelectorTestId(timelineViewSchemaId),
      ),
      sortSelector: dataTestIdSelector(
        workbookSortMenuTriggerTestId(timelineViewSchemaId),
      ),
    },
  );
}

async function expectWorkbookViewBarGeometry(
  page: Page,
  options: {
    readonly capacity: number;
    readonly hiddenCount: number;
    readonly label: string;
  },
) {
  await expect(
    page.getByTestId(workbookViewBarQueryControlsTestId(timelineViewSchemaId)),
  ).toHaveAttribute("data-query-chip-capacity", String(options.capacity));
  await expect(
    page.getByTestId(workbookViewBarQueryControlsTestId(timelineViewSchemaId)),
  ).toHaveAttribute(
    "data-hidden-query-chip-count",
    String(options.hiddenCount),
  );
  const geometry = await readWorkbookViewBarGeometry(page);
  expect(geometry.capacity, `${options.label}: chip capacity`).toBe(
    options.capacity,
  );
  expect(geometry.hiddenCount, `${options.label}: hidden chip count`).toBe(
    options.hiddenCount,
  );
  expect(
    geometry.filter.scrollWidth,
    `${options.label}: Filters content clipping`,
  ).toBeLessThanOrEqual(geometry.filter.clientWidth + 1);
  expect(
    geometry.filter.visibleText,
    `${options.label}: Filters visible text`,
  ).toContain("Filters");
  expect(
    geometry.document.scrollWidth,
    `${options.label}: document inline overflow`,
  ).toBeLessThanOrEqual(geometry.document.clientWidth + 1);
  for (const control of geometry.controls) {
    expect(
      control.width,
      `${options.label}: ${control.name} width`,
    ).toBeGreaterThan(0);
    expect(
      control.left,
      `${options.label}: ${control.name} left containment`,
    ).toBeGreaterThanOrEqual(geometry.viewBar.left - 1);
    expect(
      control.right,
      `${options.label}: ${control.name} right containment`,
    ).toBeLessThanOrEqual(geometry.viewBar.right + 1);
  }
  geometry.controls.slice(1).forEach((control, index) => {
    const previous = geometry.controls[index];
    expect(
      control.left,
      `${options.label}: ${previous?.name ?? "previous"} before ${control.name}`,
    ).toBeGreaterThanOrEqual((previous?.right ?? 0) - 1);
  });
  return geometry;
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
  expect(
    layout.shell.maxTop,
    `${label}: grid shell vertical scroll`,
  ).toBeLessThanOrEqual(1);
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
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${viewSchemaId}:${recordId}:${fieldKey}`,
  );

  await page.keyboard.press("ArrowRight");
  if (rightFieldKey) {
    await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
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
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toContainText(
    `${viewSchemaId}:${recordId}:`,
  );
  await expect(page.getByTestId(workbookFocusAnchorTestId())).not.toHaveText(
    `${viewSchemaId}:${recordId}:${fieldKey}`,
  );
  return cell;
}

test("keyboard shortcuts keep workbook grid anchors without module switching", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("KEYBOARD-CONTRACT"),
    "Workbook inspector workbook-interaction keyboard contract",
  );
  const alpha = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e901-alpha"),
    "timeline.date_entered_text": "Workbook inspector alpha",
    [hostRefsFieldKey]: collectionActionsPayload(["WorkbookInteractionHost?"]),
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect(
    page.getByTestId(
      rowCellTestId(alpha.record_id as string, "timeline.date_entered_text"),
    ),
  ).toHaveText("Workbook inspector alpha");
  const initialURL = page.url();

  const alphaSummary = page.getByTestId(
    rowCellTestId(alpha.record_id as string, "timeline.date_entered_text"),
  );
  const alphaSummaryCell = await activateSemanticGridCell(alphaSummary);
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.date_entered_text`,
  );

  await page.keyboard.press("Tab");
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.date_entered_text`,
  );
  await expect(alphaSummaryCell).not.toBeFocused();
  await page.keyboard.press("Escape");

  await openTimelineInspector(page, alpha.record_id as string);
  await expect(page.getByTestId(timelineInspectorTestId())).toContainText(
    "WorkbookInteractionHost?",
  );
  expect(page.url()).toBe(initialURL);
  await alphaSummaryCell.focus();
  await expect(alphaSummaryCell).toBeFocused();
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.date_entered_text`,
  );

  await page.keyboard.press("Alt+H");
  await expect(page.getByTestId(rowHistoryPanelTestId())).toContainText(
    String(alpha.record_id),
  );
  await expect(
    page.getByTestId(timelineInspectorSectionTestId("history")),
  ).toBeFocused();
  expect(page.url()).toBe(initialURL);

  await alphaSummaryCell.focus();
  await page.keyboard.press("Space");
  await expect(
    page.getByTestId(timelineInspectorSectionTestId("evidence")),
  ).toBeFocused();

  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();
  await scrollGridCellIntoView({
    cellKey: "timeline.analyst_text",
    page,
    recordId: alpha.record_id as string,
    surface: timelineViewSchemaId,
  });
  const alphaAnalyst = page.getByTestId(
    rowCellTestId(alpha.record_id as string, "timeline.analyst_text"),
  );
  await alphaAnalyst.click();
  const alphaAnalystEditor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.analyst_text",
      recordId: alpha.record_id as string,
      surface: "grid",
    }),
  );
  await page.evaluate(() => {
    Reflect.set(window, "__cartularyShortcutDefaultPrevented", null);
    const recordAltH = (event: KeyboardEvent) => {
      if (event.altKey && event.key.toLowerCase() === "h") {
        Reflect.set(
          window,
          "__cartularyShortcutDefaultPrevented",
          event.defaultPrevented,
        );
        window.removeEventListener("keydown", recordAltH);
      }
    };
    window.addEventListener("keydown", recordAltH);
  });
  await alphaAnalystEditor.press("Alt+H");
  expect(
    await page.evaluate(() =>
      Reflect.get(window, "__cartularyShortcutDefaultPrevented"),
    ),
  ).toBe(false);
  await expect(page.getByTestId(rowHistoryPanelTestId())).toHaveCount(0);

  await page.keyboard.press("Control+V");
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.analyst_text`,
  );
  await page.keyboard.press("Escape");
  if (await page.getByTestId(rowHistoryPanelTestId()).isVisible()) {
    await page.keyboard.press("Escape");
  }
  await expect(page.getByTestId(rowHistoryPanelTestId())).toHaveCount(0);
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.analyst_text`,
  );
  expect(page.url()).toBe(initialURL);
});

test("keeps the incident workbook inside the browser viewport and delegates overflow to workbook panels", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("WORKBOOK-LAYOUT"),
    "Workbook inspector bounded workbook layout",
  );
  await createTimelineFillers(
    page,
    incidentId,
    "Workbook inspector layout row",
    72,
  );
  const longSavedViewName =
    "Workbook view-bar layout resilience with a deliberately long saved-view name";
  const longQuerySavedView = await createSavedView(page, incidentId, {
    display_name: longSavedViewName,
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

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await selectSavedView(
    page,
    timelineViewSchemaId,
    longQuerySavedView.saved_view_id,
  );
  await expect(
    page.getByTestId(savedViewSelectorTestId(timelineViewSchemaId)),
  ).toHaveValue(longQuerySavedView.saved_view_id);
  await expect(
    page.getByTestId(workbookViewBarQueryControlsTestId(timelineViewSchemaId)),
  ).toHaveAttribute("data-hidden-query-chip-count", "6");
  await expect(page.locator('[data-grid-data-state="refreshing"]')).toHaveCount(
    0,
  );
  await expect(
    page.locator(
      '[data-grid-data-state="stale_error"], [data-grid-data-state="unavailable"]',
    ),
  ).toHaveCount(0);

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
  await expect(
    page
      .getByTestId(workbookShellSlotTestId("top-bar"))
      .getByTestId(workbookPresenceSummaryTestId()),
  ).toBeVisible();
  await expect(
    page
      .getByTestId(workbookShellSlotTestId("status-strip"))
      .getByTestId(workbookPresenceSummaryTestId()),
  ).toHaveCount(0);

  const baseGeometry = await expectWorkbookViewBarGeometry(page, {
    capacity: 3,
    hiddenCount: 6,
    label: "base 1440x900",
  });
  await page.setViewportSize({ width: 1440, height: 720 });
  const baseShortGeometry = await expectWorkbookViewBarGeometry(page, {
    capacity: 3,
    hiddenCount: 6,
    label: "base 1440x720",
  });
  expect(baseShortGeometry.filter).toEqual(baseGeometry.filter);
  expect(baseShortGeometry.viewBar).toEqual(baseGeometry.viewBar);

  await page.setViewportSize({ width: 1024, height: 720 });
  const narrowGeometry = await expectWorkbookViewBarGeometry(page, {
    capacity: 2,
    hiddenCount: 7,
    label: "narrow 1024x720",
  });
  await page.setViewportSize({ width: 1024, height: 640 });
  const narrowShortGeometry = await expectWorkbookViewBarGeometry(page, {
    capacity: 2,
    hiddenCount: 7,
    label: "narrow 1024x640",
  });
  expect(narrowShortGeometry.filter).toEqual(narrowGeometry.filter);
  expect(narrowShortGeometry.viewBar).toEqual(narrowGeometry.viewBar);

  await page.setViewportSize({ width: 768, height: 640 });
  const compactGeometry = await expectWorkbookViewBarGeometry(page, {
    capacity: 0,
    hiddenCount: 9,
    label: "compact 768x640",
  });
  await page.setViewportSize({ width: 768, height: 560 });
  const compactShortGeometry = await expectWorkbookViewBarGeometry(page, {
    capacity: 0,
    hiddenCount: 9,
    label: "compact 768x560",
  });
  expect(compactShortGeometry.filter).toEqual(compactGeometry.filter);
  expect(compactShortGeometry.viewBar).toEqual(compactGeometry.viewBar);

  await page.setViewportSize({ width: 1280, height: 720 });

  await page.setViewportSize({ width: 1280, height: 639 });
  await expectWideWorkbookTopBarChrome(page);
  await expect(
    page.getByTestId(workbookResponsiveBandTestId()),
  ).toHaveAttribute("data-workbook-block-mode", "short_height");
  await expect
    .poll(async () => {
      const layout = await readWorkbookDocumentLayout(page);
      return (
        layout.viewport.innerHeight === 639 &&
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
    "vertical resized 1280x639",
  );
  await page.setViewportSize({ width: 1280, height: 640 });
  await expectWideWorkbookTopBarChrome(page);
  await expect(
    page.getByTestId(workbookResponsiveBandTestId()),
  ).toHaveAttribute("data-workbook-block-mode", "compact_height");
  await page.setViewportSize({ width: 1280, height: 720 });
  await expectWideWorkbookTopBarChrome(page);
  await expect(
    page.getByTestId(workbookResponsiveBandTestId()),
  ).toHaveAttribute("data-workbook-block-mode", "base_height");

  await page
    .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
    .click();
  await expect(page.getByTestId(timelineInspectorTestId())).toBeVisible();
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toHaveAttribute("data-inspector-layout", "adjacent");
  await expect(
    page.getByRole("separator", { name: "Resize inspector" }),
  ).toBeVisible();
  await page.evaluate(() => window.scrollTo(0, 10_000));
  const inspectorLayout = await readWorkbookDocumentLayout(page);
  expectWorkbookDocumentBounded(inspectorLayout, "inspector open");
  expect(inspectorLayout.inspector?.bottom ?? 0).toBeLessThanOrEqual(
    inspectorLayout.viewport.innerHeight + 1,
  );

  await page.setViewportSize({ width: 1024, height: 640 });
  await expect(
    page.getByTestId(workbookResponsiveBandTestId()),
  ).toHaveAttribute("data-workbook-responsive-band", "narrow_desktop");
  await expect(
    page.getByTestId(workbookViewBarQueryControlsTestId(timelineViewSchemaId)),
  ).toHaveAttribute("data-query-chip-capacity", "2");
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toHaveAttribute("data-inspector-layout", "right_overlay");
  await expect(
    page.getByTestId(workbookShellSlotTestId("primary-grid")),
  ).toHaveAttribute("inert", "");
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

  await page.setViewportSize({ width: 768, height: 640 });
  await expect(
    page.getByTestId(workbookResponsiveBandTestId()),
  ).toHaveAttribute("data-workbook-responsive-band", "compact_desktop");
  await expect(
    page.getByTestId(workbookViewBarQueryControlsTestId(timelineViewSchemaId)),
  ).toHaveAttribute("data-query-chip-capacity", "0");
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toHaveAttribute("data-inspector-layout", "full_overlay");
  await expect(
    page.getByTestId(workbookShellSlotTestId("primary-grid")),
  ).toHaveAttribute("inert", "");
  await expect(
    page
      .getByTestId(workbookShellSlotTestId("top-bar"))
      .getByTestId(workbookPresenceSummaryTestId()),
  ).toHaveCount(0);
  await expect(
    page
      .getByTestId(workbookShellSlotTestId("status-strip"))
      .getByTestId(workbookPresenceSummaryTestId()),
  ).toBeVisible();

  await page.keyboard.press("Escape");
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  for (const viewSchemaId of [hostsViewSchemaId, notesViewSchemaId]) {
    await page.getByTestId(workbookSurfacesMenuTriggerTestId()).click();
    await page
      .getByTestId(workbookSurfacesMenuOptionTestId(viewSchemaId))
      .click();
    await expect(page.getByTestId(gridShellTestId(viewSchemaId))).toBeVisible();
    const inspectorToggle = page.getByTestId(
      workbookInspectorToggleTestId(viewSchemaId),
    );
    await inspectorToggle.click();
    await expect(page.locator("[data-inspector-layout]")).toHaveAttribute(
      "data-inspector-layout",
      "full_overlay",
    );
    await expect(
      page.getByTestId(workbookShellSlotTestId("primary-grid")),
    ).toHaveAttribute("inert", "");
    await page.keyboard.press("Escape");
    await expect(inspectorToggle).toBeFocused();
  }
  await page.getByTestId(systemViewSwitcherTriggerTestId()).click();
  await page
    .getByTestId(
      systemViewSwitcherOptionTestId(
        "scope-indicators",
        assessmentsViewSchemaId,
      ),
    )
    .click();
  await expect(
    page.getByTestId(gridShellTestId(assessmentsViewSchemaId)),
  ).toBeVisible();
  const assessmentInspectorToggle = page.getByTestId(
    workbookInspectorToggleTestId(assessmentsViewSchemaId),
  );
  await assessmentInspectorToggle.click();
  await expect(page.locator("[data-inspector-layout]")).toHaveAttribute(
    "data-inspector-layout",
    "full_overlay",
  );
  await expect(
    page.getByTestId(workbookShellSlotTestId("primary-grid")),
  ).toHaveAttribute("inert", "");
  await page.keyboard.press("Escape");
  await expect(assessmentInspectorToggle).toBeFocused();
  await page.getByTestId(workbookSurfacesMenuTriggerTestId()).click();
  await page
    .getByTestId(workbookSurfacesMenuOptionTestId(timelineViewSchemaId))
    .click();
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await page
    .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
    .click();

  await page.setViewportSize({ width: 767, height: 639 });
  await expect(
    page.getByTestId(workbookResponsiveBandTestId()),
  ).toHaveAttribute("data-workbook-responsive-band", "below_supported_minimum");
  await expect(
    page.getByTestId(workbookSurfacesMenuTriggerTestId()),
  ).toBeVisible();
  await page
    .getByTestId(workbookSurfacesMenuTriggerTestId())
    .press("ArrowDown");
  await expect(page.getByTestId(workbookSurfacesMenuTestId())).toBeVisible();
  await expect(
    page
      .getByTestId(workbookSurfacesMenuTestId())
      .getByRole("menuitemradio")
      .first(),
  ).toBeFocused();
  await page.keyboard.press("End");
  await expect(
    page
      .getByTestId(workbookSurfacesMenuTestId())
      .getByRole("menuitemradio")
      .last(),
  ).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(page.getByTestId(workbookSurfacesMenuTestId())).toHaveCount(0);
  await expect(
    page.getByTestId(workbookSurfacesMenuTriggerTestId()),
  ).toBeFocused();
  await expect(page.getByTestId(saveStateTestId())).toBeVisible();
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toHaveAttribute("data-inspector-layout", "full_overlay");

  await page.keyboard.press("Escape");
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await expect(
    page.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
  ).toBeFocused();
});

test("Timeline grid keyboard navigation, edit cancellation, and Esc restore semantic focus", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("WORKBOOKKEYBOARD"),
    "browser.coordination-review.row-02 workbook keyboard contract",
  );
  const alpha = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("browser.coordination-review.row-02-alpha"),
    "timeline.activity_synopsis_text":
      "browser.coordination-review.row-02 Alpha",
    "timeline.raw_activity_text":
      "browser.coordination-review.row-02 Alpha details",
  });
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
  await expect(alphaSummary).toHaveText(
    "browser.coordination-review.row-02 Alpha",
  );
  const alphaSummaryCell = await activateSemanticGridCell(alphaSummary);
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
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
  expect(copiedText).toBe("browser.coordination-review.row-02 Alpha");

  await page.keyboard.press("ArrowRight");
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.data_source_text`,
  );
  await page.keyboard.press("ArrowLeft");
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
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
  await alphaSummaryEditor.fill(
    "browser.coordination-review.row-02 dirty draft",
  );
  await alphaSummaryEditor.press("Escape");
  await expect(alphaSummary).toHaveText(
    "browser.coordination-review.row-02 Alpha",
  );
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

  await openTimelineInspector(page, alpha.record_id);
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: alpha.record_id,
    surface: timelineViewSchemaId,
  });
  await alphaSummary.focus();
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${timelineViewSchemaId}:${alpha.record_id}:timeline.activity_synopsis_text`,
  );
  const inspectorDetails = page.getByTestId(
    rowInspectorFieldTestId(alpha.record_id, "timeline.raw_activity_text"),
  );
  await expect(inspectorDetails).toHaveValue(
    "browser.coordination-review.row-02 Alpha details",
  );
  await inspectorDetails.focus();
  await inspectorDetails.fill(
    "browser.coordination-review.row-02 inspector dirty draft",
  );
  await inspectorDetails.press("Escape");
  await expect(inspectorDetails).toHaveValue(
    "browser.coordination-review.row-02 Alpha details",
  );
  await inspectorDetails.press("Escape");
  await expect(alphaSummaryCell).toBeFocused();
  await alphaSummaryCell.press("Escape");
  await expect(inspectorDetails).toHaveCount(0);
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await expect(alphaSummaryCell).toBeFocused();
  await alphaSummaryCell.press("Escape");
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await expect(alphaSummaryCell).toBeFocused();
});

test("Timeline clipboard paste maps an exact rectangle and persists the target row", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("WORKBOOKPASTE"),
    "Timeline clipboard paste contract",
  );
  const beta = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("timeline-clipboard-paste-beta"),
    "timeline.activity_synopsis_text": "Timeline clipboard paste Beta",
    "timeline.raw_activity_text": "Timeline clipboard paste Beta details",
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

  const pastePath = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`;
  const paste = await observeBoundedPost({
    action: () =>
      pasteGridMatrix({
        fieldKey: "timeline.activity_synopsis_text",
        matrix: [
          ["Timeline clipboard pasted Beta", "timeline-paste-host-token"],
        ],
        page,
        recordId: beta.record_id,
        surface: timelineViewSchemaId,
      }),
    operation: "Timeline clipboard paste",
    page,
    pathSuffix: pastePath,
  });
  expect(paste.requestCount).toBe(1);
  expect(readPostBody(paste.request)).toMatchObject({
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
  expect(paste.response.ok()).toBeTruthy();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await expect(
    page.getByTestId(
      rowCellTestId(beta.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("Timeline clipboard pasted Beta");
  const betaAfterPaste = await waitForViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    beta.record_id,
  );
  expect(stringCell(betaAfterPaste, "timeline.activity_synopsis_text")).toBe(
    "Timeline clipboard pasted Beta",
  );
});

test("Timeline keyboard fill-down preserves the selected range and restores source focus", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("WORKBOOKFILL"),
    "Timeline keyboard fill-down contract",
  );
  await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("timeline-keyboard-fill-one"),
    "timeline.activity_synopsis_text": "Timeline fill row one",
    "timeline.raw_activity_text": "Timeline fill details one",
  });
  await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("timeline-keyboard-fill-two"),
    "timeline.activity_synopsis_text": "Timeline fill row two",
    "timeline.raw_activity_text": "Timeline fill details two",
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

  const committedRows = page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator('[role="row"][data-grid-record-id]');
  await expect(committedRows).toHaveCount(2);
  const [fillSourceRecordId, fillTargetRecordId] =
    await committedRows.evaluateAll((rows) =>
      rows.map((row) => (row as HTMLElement).dataset.gridRecordId ?? ""),
    );
  if (!fillSourceRecordId || !fillTargetRecordId) {
    throw new Error("Expected two adjacent committed Timeline rows");
  }
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
  await scrollGridCellIntoView({
    cellKey: "timeline.raw_activity_text",
    page,
    recordId: fillSourceRecordId,
    surface: timelineViewSchemaId,
  });
  const fillSourceDisplay = page.getByTestId(
    rowCellTestId(fillSourceRecordId, "timeline.raw_activity_text"),
  );
  await expect(fillSourceDisplay).toBeVisible();
  const fillSourceCell = await activateSemanticGridCell(fillSourceDisplay);
  const fillPath = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/bulk-mutations`;
  const fillHandle = page.locator(gridFillHandleSelector());
  await expect(fillHandle).toHaveAttribute(
    "aria-label",
    "Drag to fill this value",
  );
  const doubleClickRequests = await countPostRequestsDuring({
    action: async () => {
      await fillHandle.evaluate((handle) => {
        handle.dispatchEvent(new MouseEvent("dblclick", { bubbles: true }));
      });
      await page.evaluate(
        () =>
          new Promise<void>((resolve) =>
            requestAnimationFrame(() => requestAnimationFrame(() => resolve())),
          ),
      );
    },
    page,
    pathSuffix: fillPath,
  });
  expect(doubleClickRequests).toBe(0);

  await fillSourceCell.press("Shift+ArrowDown");
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${timelineViewSchemaId}:${fillTargetRecordId}:timeline.raw_activity_text`,
  );
  const fillTargetDisplay = page.getByTestId(
    rowCellTestId(fillTargetRecordId, "timeline.raw_activity_text"),
  );
  await expect(semanticGridCell(fillTargetDisplay)).toBeFocused();

  const fill = await observeBoundedPost({
    action: () => page.keyboard.press("Control+d"),
    operation: "Timeline keyboard fill-down",
    page,
    pathSuffix: fillPath,
  });
  expect(fill.requestCount).toBe(1);
  expect(fill.response.ok()).toBeTruthy();
  expect(readPostBody(fill.request)).toMatchObject({
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
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await expect(fillTargetDisplay).toHaveText(fillSourceValue);
  await expect
    .poll(async () => {
      const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
      const target = rows.find((row) => row.record_id === fillTarget.record_id);
      return target === undefined
        ? "missing"
        : stringCell(target, "timeline.raw_activity_text");
    })
    .toBe(fillSourceValue);
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${timelineViewSchemaId}:${fillSourceRecordId}:timeline.raw_activity_text`,
  );
  await expect(fillSourceCell).toBeFocused();
  await expect(
    page.getByRole("alert").filter({
      hasText: "Select a writable one-column range before using fill down.",
    }),
  ).toHaveCount(0);
});

test("Timeline grouping renders reviewed and unreviewed presentation-only rows", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("WORKBOOKGROUPS"),
    "Timeline grouping presentation contract",
  );
  const beta = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("timeline-grouping-reviewed"),
    "timeline.activity_synopsis_text": "Timeline grouping reviewed row",
  });
  await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("timeline-grouping-unreviewed"),
    "timeline.activity_synopsis_text": "Timeline grouping unreviewed row",
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

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
        client_txn_id: uniqueTxn(
          "browser.coordination-review.row-02-review-beta",
        ),
        reason: "browser.coordination-review.row-02 grouping setup",
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
  const unreviewedGroupTestId = gridGroupRowTestId(
    timelineViewSchemaId,
    "timeline.capture_state",
    "rough",
  );
  await expect(page.getByTestId(unreviewedGroupTestId)).toBeVisible();
  await assertGroupRowPresentationOnly({
    groupTestId: unreviewedGroupTestId,
    page,
    surface: timelineViewSchemaId,
  });
  await changeGrouping(page, timelineViewSchemaId, "");
  await expect(page.getByTestId(reviewedGroupTestId)).not.toBeVisible();
  await expect(page.getByTestId(unreviewedGroupTestId)).not.toBeVisible();
});

test("Timeline virtualization reaches off-screen rows with a frozen gutter", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("WORKBOOKVIRTUAL"),
    "Timeline virtualization and frozen gutter contract",
  );
  await createTimelineFillers(
    page,
    incidentId,
    "Timeline virtualization filler",
    48,
  );
  const virtualTarget = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("timeline-virtualization-target"),
      "timeline.activity_synopsis_text": "zzz Timeline virtualization target",
      "timeline.raw_activity_text": "Timeline virtualization details",
    },
  );
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();

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
  const virtualTargetTestId = rowCellTestId(
    virtualTarget.record_id,
    "timeline.activity_synopsis_text",
  );
  await expect
    .poll(() =>
      isTestIdVisibleWithinGridViewport(
        page,
        timelineViewSchemaId,
        virtualTargetTestId,
      ),
    )
    .toBe(false);
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: virtualTarget.record_id,
    surface: timelineViewSchemaId,
    timeoutMs: 6_000,
  });
  await expect(page.getByTestId(virtualTargetTestId)).toBeVisible();
  await expect
    .poll(() =>
      isTestIdVisibleWithinGridViewport(
        page,
        timelineViewSchemaId,
        virtualTargetTestId,
      ),
    )
    .toBe(true);
  const frozenGutter = page.getByTestId(
    gridRowGutterTestId(timelineViewSchemaId, virtualTarget.record_id),
  );
  await expect(frozenGutter).toHaveCSS("position", "sticky");
  const frozenGutterLeft = await frozenGutter.evaluate(
    (element) => getComputedStyle(element).left,
  );
  await page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator(gridScrollportSelector())
    .evaluate((element) => {
      element.scrollLeft = 480;
      element.dispatchEvent(new Event("scroll", { bubbles: true }));
    });
  await expect(frozenGutter).toHaveCSS("position", "sticky");
  await expect(frozenGutter).toHaveCSS("left", frozenGutterLeft);
});

test("shared grid keyboard anchors stay stable across workbook cells", async ({
  page,
  workerAdmin,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("KEYBOARD-GRID"),
    "Workbook inspector workbook-interaction keyboard anchor semantics",
  );

  let host: ViewRow | undefined;

  await test.step("Timeline anchor", async () => {
    const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("e9grid01"),
      "timeline.activity_synopsis_text": "Workbook inspector grid anchor",
    });
    const summary = await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: timelineViewSchemaId,
      row,
      fieldKey: "timeline.activity_synopsis_text",
      expectedText: "Workbook inspector grid anchor",
      rightFieldKey: "timeline.data_source_text",
      rightFocusTestId: rowCellTestId(
        row.record_id,
        "timeline.data_source_text",
      ),
    });
    await page.keyboard.press("ArrowLeft");
    await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
      `${timelineViewSchemaId}:${row.record_id}:timeline.activity_synopsis_text`,
    );
    await expect(summary).toBeFocused();
  });

  await test.step("Hosts anchor", async () => {
    host = await createViewRow(page, incidentId, hostsViewSchemaId, {
      client_txn_id: uniqueTxn("e9grid01-host"),
      "host.display_name": "Workbook inspector host anchor",
      "host.hostname": "workbook_interaction-host.example.test",
    });
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: hostsViewSchemaId,
      row: host,
      fieldKey: "host.display_name",
      expectedText: "Workbook inspector host anchor",
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
        "identity.display_name": "Workbook inspector identity anchor",
        "identity.upn": "workbook_interaction.identity@example.test",
      },
    );
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: identitiesViewSchemaId,
      row: identity,
      fieldKey: "identity.display_name",
      expectedText: "Workbook inspector identity anchor",
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
        "assessment.rationale": "Workbook inspector assessment anchor",
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
        "task.title": "Workbook inspector task request anchor",
        "task.task_kind": "collection",
      },
    );
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: taskRequestsViewSchemaId,
      row: task,
      fieldKey: "task.title",
      expectedText: "Workbook inspector task request anchor",
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
        "decision.summary": "Workbook inspector decision anchor",
        "decision.decision_type": "containment",
        "decision.rationale":
          "Workbook inspector decision grid anchor rationale",
      },
    );
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: decisionsViewSchemaId,
      row: decision,
      fieldKey: "decision.summary",
      expectedText: "Workbook inspector decision anchor",
      rightFieldKey: "decision.status",
    });
  });

  await test.step("Notes anchor", async () => {
    const note = await createViewRow(page, incidentId, notesViewSchemaId, {
      client_txn_id: uniqueTxn("e9grid01-note"),
      "note.title": "Workbook inspector note anchor",
      "note.body": "Workbook inspector generic surface anchor body",
    });
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: notesViewSchemaId,
      row: note,
      fieldKey: "note.title",
      expectedText: "Workbook inspector note anchor",
    });
  });

  await test.step("Comm Log anchor", async () => {
    const comm = await createViewRow(page, incidentId, commLogViewSchemaId, {
      client_txn_id: uniqueTxn("e9grid01-comm"),
      "comm_log.comm_type": "briefing",
      "comm_log.audience": "Workbook inspector grid audience",
      "comm_log.channel_or_meeting": "Grid bridge",
      "comm_log.summary": "Workbook inspector comm log anchor",
    });
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: commLogViewSchemaId,
      row: comm,
      fieldKey: "comm_log.summary",
      expectedText: "Workbook inspector comm log anchor",
    });
  });

  await test.step("Handoff anchor", async () => {
    const handoff = await createViewRow(page, incidentId, handoffViewSchemaId, {
      client_txn_id: uniqueTxn("e9grid01-handoff"),
      "handoff.incoming_owner_user_id": workerAdmin.user_id,
      "handoff.current_state_summary": "Workbook inspector handoff anchor",
    });
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: handoffViewSchemaId,
      row: handoff,
      fieldKey: "handoff.current_state_summary",
      expectedText: "Workbook inspector handoff anchor",
    });
  });

  await test.step("Status Review anchor", async () => {
    const statusReview = await createViewRow(
      page,
      incidentId,
      statusReviewViewSchemaId,
      {
        client_txn_id: uniqueTxn("e9grid01-status"),
        "status_review.current_state_summary":
          "Workbook inspector status review anchor",
      },
    );
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: statusReviewViewSchemaId,
      row: statusReview,
      fieldKey: "status_review.current_state_summary",
      expectedText: "Workbook inspector status review anchor",
    });
  });

  await test.step("Lesson anchor", async () => {
    const lesson = await createViewRow(page, incidentId, lessonViewSchemaId, {
      client_txn_id: uniqueTxn("e9grid01-lesson"),
      "lesson.summary": "Workbook inspector lesson anchor",
    });
    await expectSharedGridAnchorSurface({
      page,
      incidentId,
      viewSchemaId: lessonViewSchemaId,
      row: lesson,
      fieldKey: "lesson.summary",
      expectedText: "Workbook inspector lesson anchor",
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

test("Host entity-origin clipboard paste reuses exact matches and creates stubs", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("KEYBOARD-HOST-PASTE"),
    "Workbook inspector workbook-interaction host paste",
  );
  const existing = await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e9grid-host-existing"),
    "host.display_name": "Workbook inspector reusable host",
    "host.hostname": "workbook_interaction-host-reuse.example.test",
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
  await expect(displayName).toHaveText("Workbook inspector reusable host");
  await activateSemanticGridCell(displayName);
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
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
        "Workbook inspector pasted host reuse\tworkbook_interaction-host-reuse.example.test",
        "Workbook inspector pasted host create\tworkbook_interaction-host-create.example.test",
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
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${hostsViewSchemaId}:${existing.record_id}:host.display_name`,
  );
  await expect(
    page.getByTestId(
      rowCellTestId(existing.record_id as string, "host.display_name"),
    ),
  ).toHaveText("Workbook inspector pasted host reuse");

  const rows = await queryViewRows(page, incidentId, hostsViewSchemaId);
  const reused = rows.find((row) => row.record_id === existing.record_id);
  expect(stringCell(reused ?? {}, "host.display_name")).toBe(
    "Workbook inspector pasted host reuse",
  );
  const created = rows.find(
    (row) =>
      row.record_id !== existing.record_id &&
      stringCell(row, "host.hostname") ===
        "workbook_interaction-host-create.example.test",
  );
  expect(created).toBeTruthy();
  if (created) {
    await expect(
      page.getByTestId(rowCellTestId(created.record_id, "host.display_name")),
    ).toHaveText("Workbook inspector pasted host create");
  }
  expect(postURLs.some((url) => url.includes("/imports"))).toBeFalsy();
});

test("Identity entity-origin clipboard paste reuses exact matches and creates stubs", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("KEYBOARD-IDENTITY-PASTE"),
    "Workbook inspector workbook-interaction identity paste",
  );
  const existing = await createViewRow(
    page,
    incidentId,
    identitiesViewSchemaId,
    {
      client_txn_id: uniqueTxn("e9grid-identity-existing"),
      "identity.display_name": "Workbook inspector reusable identity",
      "identity.upn": "workbook_interaction.identity.reuse@example.test",
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
  await expect(displayName).toHaveText("Workbook inspector reusable identity");
  await activateSemanticGridCell(displayName);
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
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
        "Workbook inspector pasted identity reuse\tworkbook_interaction.identity.reuse@example.test",
        "Workbook inspector pasted identity create\tworkbook_interaction.identity.create@example.test",
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
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${identitiesViewSchemaId}:${existing.record_id}:identity.display_name`,
  );
  await expect(
    page.getByTestId(
      rowCellTestId(existing.record_id as string, "identity.display_name"),
    ),
  ).toHaveText("Workbook inspector pasted identity reuse");

  const rows = await queryViewRows(page, incidentId, identitiesViewSchemaId);
  const reused = rows.find((row) => row.record_id === existing.record_id);
  expect(stringCell(reused ?? {}, "identity.display_name")).toBe(
    "Workbook inspector pasted identity reuse",
  );
  const created = rows.find(
    (row) =>
      row.record_id !== existing.record_id &&
      stringCell(row, "identity.upn") ===
        "workbook_interaction.identity.create@example.test",
  );
  expect(created).toBeTruthy();
  if (created) {
    await expect(
      page.getByTestId(
        rowCellTestId(created.record_id, "identity.display_name"),
      ),
    ).toHaveText("Workbook inspector pasted identity create");
  }
  expect(postURLs.some((url) => url.includes("/imports"))).toBeFalsy();
});
