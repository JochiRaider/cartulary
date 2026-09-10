import { Buffer } from "node:buffer";
import {
  networkAnalysisRowCellTestId,
  networkAnalysisTestId,
} from "@cartulary/ui-contracts";
import type { Page, Response } from "@playwright/test";
import { expect, test } from "./fixtures";
import {
  importNetworkFlowCSV,
  openClaimedNetworkAnalysis,
} from "./support/extensions/network_flow_activity/workspace";
import { TestClock } from "./support/runtime/testClock";

// This isolated scenario owns process-global time for actual cursor expiry.
test.beforeEach(async ({ page }) => {
  await new TestClock(page).reset();
});
test.afterEach(async ({ page }) => {
  await new TestClock(page).reset();
});

test("Network Analysis preserves committed pages and recovers real expired cursors across all exploration surfaces", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NFPAGES");
  const csv = [
    "Start Time,End Time,Source IP,Destination IP,Source Port,Destination Port,Protocol,Bytes,Packets",
  ];
  for (let i = 0; i < 1402; i++)
    csv.push(
      `2026-07-10T12:00:00Z,2026-07-10T12:01:00Z,${i < 1001 ? "192.0.2.1" : "invalid"},192.0.2.2,443,80,6,${i + 1},1`,
    );
  await importNetworkFlowCSV(page, {
    displayName: "paginated-flows",
    file: {
      name: "pages.csv",
      mimeType: "text/csv",
      buffer: Buffer.from(`${csv.join("\n")}\n`),
    },
  });
  const clock = new TestClock(page);
  let now = new Date();
  await clock.setFixed(now);
  for (const surface of ["rows", "diagnostics", "contributors"] as const) {
    now = new Date();
    await clock.setFixed(now);
    const route =
      surface === "rows"
        ? /\/tables\/nft_[^/]+\/query$/
        : surface === "diagnostics"
          ? /\/rejected-rows\/query$/
          : /\/graphs\/contributors\/query$/;
    if (surface === "diagnostics")
      await page.getByTestId(networkAnalysisTestId("mode-rejected")).click();
    if (surface === "contributors") {
      await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
      await page
        .getByTestId(/^network-flow-vertex-/)
        .first()
        .getByRole("button", { name: /^Select vertex/u })
        .click();
    }
    const nav = page.getByRole("navigation", {
      name: "Network Flow result pages",
    });
    const status = nav.getByTestId(networkAnalysisTestId("page-status"));
    const grid = page.getByTestId(
      networkAnalysisTestId(
        surface === "rows"
          ? "accepted-grid"
          : surface === "diagnostics"
            ? "rejected-grid"
            : "contributor-grid",
      ),
    );
    const next = nav.getByRole("button", { name: "Next", exact: true });
    const previous = nav.getByRole("button", { name: "Previous", exact: true });
    const refresh = nav.getByRole("button", { name: "Refresh page" });
    await expect(status).toHaveText("Page 1");
    const first = await readPage(page, route, () => refresh.click(), surface);
    expect(first.length).toBe(surface === "contributors" ? 500 : 200);
    expect(first.map((row) => row.source_row_number)).toEqual(
      Array.from(
        { length: first.length },
        (_, i) => i + (surface === "diagnostics" ? 1003 : 2),
      ),
    );
    const firstRow = first[0],
      lastRow = first.at(-1);
    if (firstRow === undefined || lastRow === undefined)
      throw new Error("Expected a nonempty first result page.");
    if (surface === "rows") {
      await page
        .getByTestId(
          networkAnalysisRowCellTestId(
            firstRow.network_flow_row_id,
            "network_flow.src_ip",
          ),
        )
        .click();
      await expect(
        page.getByTestId(networkAnalysisTestId("inspector")),
      ).toBeVisible();
    }
    let failTransition!: () => void;
    const transition = new Promise<void>((resolve) => {
      failTransition = resolve;
    });
    await page.route(
      route,
      async (route) => {
        await transition;
        await route.abort("failed");
      },
      { times: 1 },
    );
    await next.click();
    await expect(status).toHaveText("Page 1 · Loading page 2");
    failTransition();
    await expect(grid.getByRole("alert")).toContainText(
      "Could not load page 2. Showing page 1.",
    );
    await expect(status).toHaveText("Page 1");
    await expect(next).toBeFocused();
    const retry = nav.getByRole("button", { name: "Retry page 2" });
    const second = await readPage(page, route, () => retry.click(), surface);
    await expect(status).toHaveText("Page 2");
    expect(second[0]?.source_row_number).toBe(lastRow.source_row_number + 1);
    if (surface === "rows")
      await expect(
        page.getByTestId(networkAnalysisTestId("inspector")),
      ).toHaveCount(0);
    await expect(
      nav.getByRole("button", { name: "Restart query" }),
    ).toBeFocused();
    await page.route(route, (route) => route.abort("failed"), { times: 1 });
    await previous.click();
    await expect(grid.getByRole("alert")).toContainText(
      "Could not load page 1. Showing page 2.",
    );
    expect(await readPage(page, route, () => refresh.click(), surface)).toEqual(
      second,
    );
    await expect(status).toHaveText("Page 2");
    await page.route(route, (route) => route.abort("failed"), { times: 1 });
    await previous.click();
    await expect(
      nav.getByRole("button", { name: "Retry page 1" }),
    ).toBeEnabled();
    expect(
      await readPage(
        page,
        route,
        () => nav.getByRole("button", { name: "Retry page 1" }).click(),
        surface,
      ),
    ).toEqual(first);
    // Same-turn commands admit exactly one read, even before React disables controls.
    let reads = 0;
    const count = (response: Response) => {
      if (route.test(new URL(response.url()).pathname)) reads++;
    };
    page.on("response", count);
    expect(
      await readPage(
        page,
        route,
        () =>
          next.evaluate((button: HTMLButtonElement) => {
            button.click();
            button.click();
          }),
        surface,
      ),
    ).toEqual(second);
    await expect(status).toHaveText("Page 2");
    expect(reads).toBe(1);
    page.off("response", count);
    now = new Date(now.getTime() + 900_000);
    await clock.setFixed(now);
    const expired = page.waitForResponse(
      (response) =>
        route.test(new URL(response.url()).pathname) &&
        response.status() === 400,
    );
    const recovered = readPage(page, route, () => next.click(), surface);
    expect((await (await expired).json()).error.details).toMatchObject({
      reason_code: "expired",
      retry_action: "restart_query",
    });
    expect(await recovered).toEqual(first);
    await expect(status).toHaveText(
      "Page 1 · The page cursor expired. Results restarted at page one.",
    );
    // Reissue the backward-navigation chain within the session lifetime.
    now = new Date();
    await clock.setFixed(now);
    await readPage(page, route, () => refresh.click(), surface);
    await readPage(page, route, () => next.click(), surface);
    await readPage(page, route, () => next.click(), surface);
    await expect(status).toHaveText("Page 3");
    now = new Date(now.getTime() + 900_000);
    await clock.setFixed(now);
    expect(
      await readPage(page, route, () => previous.click(), surface),
    ).toEqual(first);
    await expect(status).toHaveText(
      "Page 1 · The page cursor expired. Results restarted at page one.",
    );
    await expect(previous).toBeFocused();
    if (surface === "contributors") {
      await expect(
        page.getByTestId(networkAnalysisTestId("contributor-drawer")),
      ).toBeVisible();
      await expect(
        page.getByText("Graph ready", { exact: true }),
      ).toBeVisible();
    }
  }
});

async function readPage(
  page: Page,
  route: RegExp,
  action: () => Promise<unknown>,
  surface: "rows" | "diagnostics" | "contributors",
) {
  const response = page.waitForResponse(
    (response) => route.test(new URL(response.url()).pathname) && response.ok(),
  );
  await action();
  const result = (await (await response).json()).data;
  return (
    surface === "contributors"
      ? result.contributors.map((item: { row: unknown }) => item.row)
      : result[surface]
  ) as { network_flow_row_id: string; source_row_number: number }[];
}
