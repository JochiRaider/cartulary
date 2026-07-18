import {
  gridDataCellsSelector,
  gridDataRowsSelector,
  gridScrollportSelector,
  networkAnalysisDiagnosticCellTestId,
  networkAnalysisRowCellTestId,
  networkAnalysisTestId,
} from "@cartulary/ui-contracts";
import type { Locator, Page } from "@playwright/test";

import { expect, test } from "../fixtures";

type FixtureSurface = {
  readonly buttonLabel: string;
  readonly expectedColumns: number;
  readonly gridSelector: "accepted-grid" | "contributor-grid" | "rejected-grid";
  readonly horizontalScrollExpected: boolean;
};

const fixtureSurfaces = [
  {
    buttonLabel: "Accepted rows",
    expectedColumns: 14,
    gridSelector: "accepted-grid",
    horizontalScrollExpected: true,
  },
  {
    buttonLabel: "Rejected diagnostics",
    expectedColumns: 6,
    gridSelector: "rejected-grid",
    horizontalScrollExpected: false,
  },
  {
    buttonLabel: "Graph contributors",
    expectedColumns: 14,
    gridSelector: "contributor-grid",
    horizontalScrollExpected: true,
  },
] as const satisfies readonly FixtureSurface[];

test("measures bounded Network Flow production-grid virtualization without a timing claim", async ({
  page,
}, testInfo) => {
  const control = await measureFixture(page, 100);
  const supported = await measureFixture(page, 1_000);

  for (const surface of fixtureSurfaces) {
    const controlResult = control[surface.gridSelector];
    const supportedResult = supported[surface.gridSelector];
    expect(supportedResult.mountedRows).toBeLessThan(1_000);
    expect(supportedResult.mountedRows).toBeLessThanOrEqual(
      controlResult.mountedRows + 8,
    );
    expect(supportedResult.mountedCells).toBeLessThanOrEqual(
      controlResult.mountedCells + 8 * (surface.expectedColumns + 1),
    );
    expect(supportedResult.clientHeight).toBe(controlResult.clientHeight);
    expect(supportedResult.scrollHeight).toBeGreaterThan(
      supportedResult.clientHeight,
    );
    if (surface.horizontalScrollExpected) {
      expect(supportedResult.scrollWidth).toBeGreaterThan(
        supportedResult.clientWidth,
      );
    }
  }

  await proveSupportedEnvelopeReachabilityAndContinuity(page);
  process.stdout.write(
    `[NETWORK_FLOW_GRID_SUPPORTED_ENVELOPE]\n${JSON.stringify({ control, supported }, null, 2)}\n`,
  );
  await testInfo.attach("network-flow-grid-supported-envelope.json", {
    body: Buffer.from(JSON.stringify({ control, supported }, null, 2)),
    contentType: "application/json",
  });
});

test("Measure the 1000-resource all-column production Network Flow grid envelope as informative implementation evidence.", async ({
  page,
}) => {
  const supported = await measureFixture(page, 1_000);
  for (const surface of fixtureSurfaces) {
    const result = supported[surface.gridSelector];
    expect(result.mountedRows).toBeLessThan(1_000);
    expect(result.scrollHeight).toBeGreaterThan(result.clientHeight);
  }
});

async function measureFixture(page: Page, logicalRows: 100 | 1_000) {
  await page.goto(
    `/?debug=harness&fixture=network-flow-grid-load&fixture_rows=${logicalRows}`,
  );
  const fixture = page.getByTestId(networkAnalysisTestId("load-fixture"));
  await expect(fixture).toHaveAttribute(
    "data-logical-row-count",
    String(logicalRows),
  );

  const results = {} as Record<
    FixtureSurface["gridSelector"],
    GridDOMMeasurement
  >;
  for (const surface of fixtureSurfaces) {
    await page.getByRole("button", { name: surface.buttonLabel }).click();
    const grid = page.getByTestId(networkAnalysisTestId(surface.gridSelector));
    await expect(grid.locator(gridDataRowsSelector()).first()).toBeVisible();
    await showEveryDeclaredColumn(page, surface.expectedColumns);
    results[surface.gridSelector] = await readGridDOMMeasurement(grid);
  }
  return results;
}

async function showEveryDeclaredColumn(page: Page, expectedColumns: number) {
  const menu = page.getByTestId(networkAnalysisTestId("column-menu"));
  if ((await menu.getAttribute("open")) === null) {
    await menu.locator("summary").click();
  }
  const toggles = menu.locator('input[type="checkbox"]');
  await expect(toggles).toHaveCount(expectedColumns);
  for (let index = 0; index < expectedColumns; index += 1) {
    const toggle = toggles.nth(index);
    if (!(await toggle.isChecked())) {
      await toggle.check();
    }
  }
  for (let index = 0; index < expectedColumns; index += 1) {
    await expect(toggles.nth(index)).toBeChecked();
  }
}

type GridDOMMeasurement = {
  readonly clientHeight: number;
  readonly clientWidth: number;
  readonly mountedCells: number;
  readonly mountedRows: number;
  readonly scrollHeight: number;
  readonly scrollWidth: number;
};

async function readGridDOMMeasurement(
  grid: Locator,
): Promise<GridDOMMeasurement> {
  const scrollport = grid.locator(gridScrollportSelector());
  await expect(scrollport).toHaveCount(1);
  const dimensions = await scrollport.evaluate((element) => ({
    clientHeight: element.clientHeight,
    clientWidth: element.clientWidth,
    scrollHeight: element.scrollHeight,
    scrollWidth: element.scrollWidth,
  }));
  return {
    ...dimensions,
    mountedCells: await grid.locator(gridDataCellsSelector()).count(),
    mountedRows: await grid.locator(gridDataRowsSelector()).count(),
  };
}

async function proveSupportedEnvelopeReachabilityAndContinuity(page: Page) {
  await page.getByRole("button", { name: "Accepted rows" }).click();
  await showEveryDeclaredColumn(page, 14);
  const acceptedGrid = page.getByTestId(networkAnalysisTestId("accepted-grid"));
  const acceptedScrollport = acceptedGrid.locator(gridScrollportSelector());
  await scrollToBoundary(acceptedScrollport, "end");
  await expect(
    page.getByTestId(
      networkAnalysisRowCellTestId(
        "nfr_load_1000",
        "network_flow.application_label",
      ),
    ),
  ).toBeVisible();

  await scrollToBoundary(acceptedScrollport, "start");
  const firstContent = page.getByTestId(
    networkAnalysisRowCellTestId(
      "nfr_load_0001",
      "network_flow.flow_start_utc",
    ),
  );
  await expect(firstContent).toBeVisible();
  const firstCell = acceptedGrid.getByRole("gridcell").filter({
    has: firstContent,
  });
  await firstCell.click();
  await firstCell.press("Shift+ArrowDown");
  const secondContent = page.getByTestId(
    networkAnalysisRowCellTestId(
      "nfr_load_0002",
      "network_flow.flow_start_utc",
    ),
  );
  const secondCell = acceptedGrid.getByRole("gridcell").filter({
    has: secondContent,
  });
  await expect(secondCell).toBeFocused();
  await expect(
    page.getByTestId(networkAnalysisTestId("inspector")),
  ).toBeVisible();
  const selection = page.getByLabel("Fixture semantic selection");
  const selectionBeforeRefresh = await selection.textContent();
  expect(selectionBeforeRefresh).toContain("nfr_load_0001");
  expect(selectionBeforeRefresh).toContain("nfr_load_0002");

  await page
    .getByRole("button", { name: "Refresh equivalent resources" })
    .evaluate((button: HTMLButtonElement) => button.click());
  await expect(secondCell).toBeFocused();
  await expect(selection).toHaveText(selectionBeforeRefresh ?? "");
  await expect(
    page.getByTestId(networkAnalysisTestId("inspector")),
  ).toBeVisible();

  await page.getByRole("button", { name: "Rejected diagnostics" }).click();
  await showEveryDeclaredColumn(page, 6);
  const rejectedGrid = page.getByTestId(networkAnalysisTestId("rejected-grid"));
  await scrollToBoundary(rejectedGrid.locator(gridScrollportSelector()), "end");
  await expect(
    page.getByTestId(
      networkAnalysisDiagnosticCellTestId("nfd_load_1000", "message"),
    ),
  ).toBeVisible();

  await page.getByRole("button", { name: "Graph contributors" }).click();
  await showEveryDeclaredColumn(page, 14);
  const contributorGrid = page.getByTestId(
    networkAnalysisTestId("contributor-grid"),
  );
  await scrollToBoundary(
    contributorGrid.locator(gridScrollportSelector()),
    "end",
  );
  await expect(
    page.getByTestId(
      networkAnalysisRowCellTestId(
        "nfr_load_1000",
        "network_flow.application_label",
      ),
    ),
  ).toBeVisible();
}

async function scrollToBoundary(
  scrollport: Locator,
  boundary: "end" | "start",
) {
  await scrollport.evaluate((element, target) => {
    element.scrollTop = target === "end" ? element.scrollHeight : 0;
    element.scrollLeft = target === "end" ? element.scrollWidth : 0;
    element.dispatchEvent(new Event("scroll"));
  }, boundary);
}
