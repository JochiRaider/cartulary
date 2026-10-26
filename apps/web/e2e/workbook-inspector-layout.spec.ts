import type { ViewRow } from "@cartulary/protocol-ts/http";

import {
  dataTestIdSelector,
  gridScrollportSelector,
  gridShellTestId,
  surfaceTabTestId,
  workbookInspectorToggleTestId,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  hostsViewSchemaId,
  notesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { gridSavedRows } from "./pages/workbookInspector";

import { createIncident } from "./support/incidents/fixtures";

import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";

import { createViewRow } from "./support/workbook/query";

async function openWorkbookInspectorSlot(page: Page, viewSchemaId: string) {
  const inspectorSlot = activeWorkbookSlot(page, "inspector", viewSchemaId);
  if ((await inspectorSlot.count()) === 0) {
    await page.getByTestId(workbookInspectorToggleTestId(viewSchemaId)).click();
  }
  await expect(inspectorSlot).toBeVisible();
}

function activeWorkbookSlot(
  page: Page,
  slot: "inspector" | "primary-grid" | "status-strip",
  viewSchemaId: string,
) {
  return page.locator(activeWorkbookSlotSelector(slot, viewSchemaId));
}

function activeWorkbookSlotSelector(
  slot: "inspector" | "primary-grid" | "status-strip",
  viewSchemaId: string,
) {
  return `${dataTestIdSelector(
    workbookShellSlotTestId(slot),
  )}[data-view-schema-id="${viewSchemaId}"]`;
}

async function expectWorkbookInspectorWorkAreaGeometry(
  page: Page,
  viewSchemaId: string,
) {
  const snapshot = await readWorkbookInspectorGeometry(page, viewSchemaId);
  expectWithinOneCssPixel(snapshot.inspectorSlot.top, snapshot.workArea.top);
  expectWithinOneCssPixel(
    snapshot.inspectorSlot.bottom,
    snapshot.workArea.bottom,
  );
  expectWithinOneCssPixel(snapshot.primaryGridSlot.top, snapshot.workArea.top);
  expectWithinOneCssPixel(
    snapshot.primaryGridSlot.bottom,
    snapshot.workArea.bottom,
  );
  expectWithinOneCssPixel(snapshot.gridShell.top, snapshot.workArea.top);
  const browsing = activeWorkbookSlot(
    page,
    "primary-grid",
    viewSchemaId,
  ).getByRole("group", { name: "Workbook browsing", exact: true });
  await expect(browsing).toBeVisible();
  const browsingBounds = await browsing.boundingBox();
  if (browsingBounds === null)
    throw new Error("Workbook browsing has no layout bounds");
  expectWithinOneCssPixel(snapshot.gridShell.bottom, browsingBounds.y);
  expectWithinOneCssPixel(
    browsingBounds.y + browsingBounds.height,
    snapshot.workArea.bottom,
  );
  expectWithinOneCssPixel(snapshot.statusStrip.top, snapshot.workArea.bottom);
  expectWithinOneCssPixel(
    snapshot.statusStrip.bottom,
    snapshot.documentElement.clientHeight,
  );
  expect(snapshot.documentElement.windowScrollY).toBe(0);
  expect(snapshot.documentElement.scrollHeight).toBeLessThanOrEqual(
    snapshot.documentElement.clientHeight + 1,
  );
}

async function readWorkbookInspectorGeometry(
  page: Page,
  viewSchemaId: string,
): Promise<WorkbookInspectorGeometry> {
  return page.evaluate(
    ({
      gridScrollportSelectorValue,
      gridSelector,
      inspectorSlotSelector,
      primaryGridSlotSelector,
      statusStripSelector,
    }) => {
      const roundRect = (element: HTMLElement) => {
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
      const requireElement = (selector: string) => {
        const element = document.querySelector<HTMLElement>(selector);
        if (element === null) {
          throw new Error(`Expected ${selector} to exist`);
        }
        return element;
      };
      const gridShell = requireElement(gridSelector);
      const primaryGridSlot = requireElement(primaryGridSlotSelector);
      const inspectorSlot = requireElement(inspectorSlotSelector);
      const statusStrip = requireElement(statusStripSelector);
      const workArea = primaryGridSlot.parentElement;
      if (workArea === null) {
        throw new Error(
          "Expected primary grid slot to have a work area parent",
        );
      }
      const gridScrollport = gridShell.querySelector<HTMLElement>(
        gridScrollportSelectorValue,
      );
      if (gridScrollport === null) {
        throw new Error("Expected grid scrollport to exist");
      }
      return {
        documentElement: {
          clientHeight: document.documentElement.clientHeight,
          scrollHeight: document.documentElement.scrollHeight,
          windowScrollY: Math.round(window.scrollY),
        },
        gridScrollport: {
          clientHeight: gridScrollport.clientHeight,
          scrollHeight: gridScrollport.scrollHeight,
          scrollTop: Math.round(gridScrollport.scrollTop),
        },
        gridShell: roundRect(gridShell),
        inspectorSlot: roundRect(inspectorSlot),
        primaryGridSlot: roundRect(primaryGridSlot),
        statusStrip: roundRect(statusStrip),
        workArea: roundRect(workArea),
      };
    },
    {
      gridScrollportSelectorValue: gridScrollportSelector(),
      gridSelector: dataTestIdSelector(gridShellTestId(viewSchemaId)),
      inspectorSlotSelector: activeWorkbookSlotSelector(
        "inspector",
        viewSchemaId,
      ),
      primaryGridSlotSelector: activeWorkbookSlotSelector(
        "primary-grid",
        viewSchemaId,
      ),
      statusStripSelector: activeWorkbookSlotSelector(
        "status-strip",
        viewSchemaId,
      ),
    },
  );
}

type WorkbookInspectorGeometry = {
  documentElement: {
    clientHeight: number;
    scrollHeight: number;
    windowScrollY: number;
  };
  gridScrollport: {
    clientHeight: number;
    scrollHeight: number;
    scrollTop: number;
  };
  gridShell: WorkbookLayoutRect;
  inspectorSlot: WorkbookLayoutRect;
  primaryGridSlot: WorkbookLayoutRect;
  statusStrip: WorkbookLayoutRect;
  workArea: WorkbookLayoutRect;
};

type WorkbookLayoutRect = {
  bottom: number;
  height: number;
  left: number;
  right: number;
  top: number;
  width: number;
};

function expectWithinOneCssPixel(actual: number, expected: number) {
  expect(Math.abs(actual - expected)).toBeLessThanOrEqual(1);
}

async function expectWorkbookIndependentScroll(
  page: Page,
  viewSchemaId: string,
) {
  const scrollState = await page.evaluate(
    ({ gridScrollportSelectorValue, gridSelector, inspectorSlotSelector }) => {
      const gridShell = document.querySelector<HTMLElement>(gridSelector);
      const inspectorSlot = document.querySelector<HTMLElement>(
        inspectorSlotSelector,
      );
      const inspectorPanel =
        inspectorSlot?.querySelector<HTMLElement>(
          "[data-inspector-scroll-body]",
        ) ?? null;
      if (
        gridShell === null ||
        inspectorSlot === null ||
        inspectorPanel === null
      ) {
        throw new Error("Expected grid and inspector scrollports to exist");
      }
      const gridScrollport = gridShell.querySelector<HTMLElement>(
        gridScrollportSelectorValue,
      );
      if (gridScrollport === null) {
        throw new Error("Expected grid and inspector scrollports to exist");
      }
      const gridMaxScroll = Math.max(
        0,
        gridScrollport.scrollHeight - gridScrollport.clientHeight,
      );
      const inspectorMaxScroll = Math.max(
        0,
        inspectorPanel.scrollHeight - inspectorPanel.clientHeight,
      );
      gridScrollport.scrollTop = Math.min(48, gridMaxScroll);
      inspectorPanel.scrollTop = Math.min(32, inspectorMaxScroll);
      return {
        gridMaxScroll,
        gridOverflowY: getComputedStyle(gridScrollport).overflowY,
        gridScrollTop: Math.round(gridScrollport.scrollTop),
        inspectorMaxScroll,
        inspectorOverflowY: getComputedStyle(inspectorPanel).overflowY,
        inspectorScrollTop: Math.round(inspectorPanel.scrollTop),
      };
    },
    {
      gridScrollportSelectorValue: gridScrollportSelector(),
      gridSelector: dataTestIdSelector(gridShellTestId(viewSchemaId)),
      inspectorSlotSelector: activeWorkbookSlotSelector(
        "inspector",
        viewSchemaId,
      ),
    },
  );
  expect(scrollState.gridOverflowY).toBe("auto");
  expect(scrollState.inspectorOverflowY).toBe("auto");
  if (scrollState.gridMaxScroll > 0) {
    expect(scrollState.gridScrollTop).toBeGreaterThan(0);
  }
  if (scrollState.inspectorMaxScroll > 0) {
    expect(scrollState.inspectorScrollTop).toBeGreaterThan(0);
  }
}

test("Workbook inspector fills the shell work area across row counts", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("WORKBOOK-LAYOUT"),
    "Workbook inspector inspector shell layout",
  );
  const timelineRows: ViewRow[] = [];
  for (let index = 0; index < 16; index += 1) {
    timelineRows.push(
      await createViewRow(page, incidentId, timelineViewSchemaId, {
        client_txn_id: uniqueTxn(`e9layout-timeline-${index}`),
        "timeline.activity_synopsis_text": `Workbook inspector layout timeline row ${index + 1}`,
      }),
    );
  }
  for (let index = 0; index < 3; index += 1) {
    await createViewRow(page, incidentId, hostsViewSchemaId, {
      client_txn_id: uniqueTxn(`e9layout-host-${index}`),
      "host.display_name": `Workbook inspector layout host ${index + 1}`,
      "host.hostname": `workbook_interaction-layout-host-${index + 1}.example.test`,
    });
  }
  await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("e9layout-evidence"),
    "evidence.title": "Workbook inspector layout evidence",
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  expect(timelineRows.length).toBeGreaterThan(0);
  await openWorkbookInspectorSlot(page, timelineViewSchemaId);
  await expectWorkbookInspectorWorkAreaGeometry(page, timelineViewSchemaId);
  await expectWorkbookIndependentScroll(page, timelineViewSchemaId);

  await page.getByTestId(surfaceTabTestId(hostsViewSchemaId)).click();
  await expect(
    page.getByTestId(gridShellTestId(hostsViewSchemaId)),
  ).toBeVisible();
  await openWorkbookInspectorSlot(page, hostsViewSchemaId);
  await expectWorkbookInspectorWorkAreaGeometry(page, hostsViewSchemaId);

  await page.getByTestId(surfaceTabTestId(evidenceViewSchemaId)).click();
  await expect(
    page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
  ).toBeVisible();
  await openWorkbookInspectorSlot(page, evidenceViewSchemaId);
  await expectWorkbookInspectorWorkAreaGeometry(page, evidenceViewSchemaId);
  await expectWorkbookIndependentScroll(page, evidenceViewSchemaId);

  await page.getByTestId(surfaceTabTestId(notesViewSchemaId)).click();
  await expect(
    page.getByTestId(gridShellTestId(notesViewSchemaId)),
  ).toBeVisible();
  await expect(gridSavedRows(page, notesViewSchemaId)).toHaveCount(0);
  await openWorkbookInspectorSlot(page, notesViewSchemaId);
  await expectWorkbookInspectorWorkAreaGeometry(page, notesViewSchemaId);
});
