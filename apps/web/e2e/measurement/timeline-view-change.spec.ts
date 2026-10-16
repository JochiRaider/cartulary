import type { ViewRow } from "@cartulary/protocol-ts/http";
import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  type CartularyAc044PredicateId,
  cartularyAc044PerformanceContract as contract,
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridSortHeaderTestId,
  rowCellTestId,
  timelineMutationSubstrateReadyTestId,
  workbookFilterClearButtonTestId,
  workbookFilterOperatorTestId,
  workbookFilterPopoverTriggerTestId,
  workbookPresenceSummaryTestId,
  workbookSortMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import type { Browser, Page, TestInfo } from "@playwright/test";
import type { WorkerAdminEntry } from "../support/auth/workerAdmin";
import {
  type Ac043TrafficDriver,
  prepareAc043Snapshot,
  startAc043SnapshotTraffic,
} from "../support/performance/ac043Snapshot";
import { activateCommittedGridCell } from "../support/workbook/rowMutations";
import { expect, test } from "./fixtures";
import {
  ac043FixtureDigest,
  attachMeasurementObservation,
  interactiveMeasurementSamplePolicy,
  type MeasurementSample,
} from "./timingSupport";
import {
  measureViewChange,
  viewChangePredicate,
} from "./viewChangeTimingSupport";

test.use({
  viewport: { width: 1920, height: 1080 },
  deviceScaleFactor: 1,
  actionTimeout: 10_000,
  navigationTimeout: 30_000,
});
test.describe.configure({ mode: "serial", timeout: 30 * 60 * 1_000 });

const view = contract.view_schema_id;
const synopsis = (row: ViewRow) =>
  String(row.cells["timeline.activity_synopsis_text"]?.value ?? "");
const byId = (left: ViewRow, right: ViewRow) =>
  left.record_id < right.record_id
    ? -1
    : left.record_id > right.record_id
      ? 1
      : 0;

async function prepareAction(
  page: Page,
  operation: "sort" | "filter" | "group",
) {
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  const frame = page.locator(`[data-workbook-view-schema-id="${view}"]`);
  await expect(frame).toHaveAttribute("data-workbook-query-pending", "false");
  await expect(frame).toHaveAttribute("data-workbook-query-filter-count", "0");
  await expect(frame).toHaveAttribute("data-workbook-query-sort-count", "0");
  await expect(frame).toHaveAttribute("data-workbook-query-group-field", "");
  await expect(frame).toHaveAttribute("data-inspector-layout", "closed");
  expect(
    await page.evaluate(() => ({
      width: innerWidth,
      height: innerHeight,
      scale: devicePixelRatio,
    })),
  ).toEqual({ width: 1920, height: 1080, scale: 1 });
  if (operation === "sort") {
    const testId = gridSortHeaderTestId(
      view,
      contract.workloads.sort.field_key,
    );
    await scrollGridTargetIntoView({
      page,
      surface: view,
      targetTestId: testId,
    });
    return () => page.getByTestId(testId).click();
  }
  if (operation === "filter") {
    await page.getByTestId(workbookFilterPopoverTriggerTestId(view)).click();
    await page
      .getByTestId(gridFilterFieldTestId(view))
      .selectOption(contract.workloads.filter.field_key);
    await page
      .getByTestId(workbookFilterOperatorTestId(view))
      .selectOption(contract.workloads.filter.operator);
    await page
      .getByTestId(gridFilterValueTestId(view))
      .fill(contract.workloads.filter.value);
    await expect(page.getByTestId(gridFilterApplyTestId(view))).toBeEnabled();
    return () => page.getByTestId(gridFilterApplyTestId(view)).click();
  }
  return async () => {
    await page
      .getByTestId(gridGroupingSelectTestId(view))
      .selectOption(contract.workloads.group.field_key);
  };
}

async function restoreDefaultQuery(
  page: Page,
  operation: "sort" | "filter" | "group",
  firstRecordId: string,
) {
  if (operation === "sort") {
    const trigger = page.getByTestId(workbookSortMenuTriggerTestId(view));
    await trigger.click();
    await page
      .getByRole("menuitem", {
        name: "Remove Activity Synopsis sort",
        exact: true,
      })
      .click();
    if ((await trigger.getAttribute("aria-expanded")) === "true")
      await trigger.click();
  } else if (operation === "filter") {
    const trigger = page.getByTestId(workbookFilterPopoverTriggerTestId(view));
    await trigger.click();
    await page.getByTestId(workbookFilterClearButtonTestId(view)).click();
    if ((await trigger.getAttribute("aria-expanded")) === "true")
      await trigger.click();
  } else {
    await page.getByTestId(gridGroupingSelectTestId(view)).selectOption("");
  }
  const frame = page.locator(`[data-workbook-view-schema-id="${view}"]`);
  await expect(frame).toHaveAttribute("data-workbook-query-pending", "false");
  await expect(frame).toHaveAttribute("data-workbook-query-filter-count", "0");
  await expect(frame).toHaveAttribute("data-workbook-query-sort-count", "0");
  await expect(frame).toHaveAttribute("data-workbook-query-group-field", "");
  await frame
    .locator('[role="grid"], [role="treegrid"]')
    .evaluate((element) => {
      element.scrollTop = 0;
    });
  const firstRow = frame.locator('[role="row"][data-grid-record-id]').first();
  await expect(firstRow).toHaveAttribute("data-grid-record-id", firstRecordId);
  await expect(firstRow.locator('[role="gridcell"]').first()).toBeVisible();
}

async function runMeasurement(
  browser: Browser,
  page: Page,
  workerAdmin: WorkerAdminEntry,
  sessionTracker: Parameters<typeof startAc043SnapshotTraffic>[1],
  testInfo: TestInfo,
  predicateId: CartularyAc044PredicateId,
) {
  const predicate = viewChangePredicate(predicateId);
  const samples: MeasurementSample[] = [];
  let traffic: Ac043TrafficDriver | null = null;
  let attached = false;
  try {
    expect(testInfo.retry).toBe(0);
    const snapshot = await prepareAc043Snapshot(page, workerAdmin);
    await page.goto(`/?incident_id=${snapshot.incidentId}`);
    const defaultRows = [...snapshot.timelineRows].sort(byId).slice(0, 100);
    const destination =
      predicate.operation === "sort"
        ? [...snapshot.timelineRows]
            .sort((a, b) =>
              synopsis(a) < synopsis(b)
                ? -1
                : synopsis(a) > synopsis(b)
                  ? 1
                  : byId(a, b),
            )
            .slice(0, 100)
        : predicate.operation === "filter"
          ? snapshot.timelineRows
              .filter(
                (row) =>
                  row.cells[contract.workloads.filter.field_key]?.value ===
                  contract.workloads.filter.value,
              )
              .sort(byId)
              .slice(0, 100)
          : defaultRows;
    expect(destination).toHaveLength(100);
    const firstDefaultRow = defaultRows[0];
    if (firstDefaultRow === undefined)
      throw new Error("View-change fixture has no default row");
    const excluded = new Set(
      [...defaultRows, ...destination].map((row) => row.record_id),
    );
    traffic = await startAc043SnapshotTraffic(
      browser,
      sessionTracker,
      snapshot.incidentId,
      snapshot.timelineRows.filter((row) => !excluded.has(row.record_id)),
      snapshot.runtime.backgroundAccounts,
      predicateId,
    );
    expect(traffic.qualifiedBackgroundSessions).toBe(24);
    expect(traffic.sessionCount).toBe(25);
    expect(traffic.updatesPerSecond).toBe(4.8);
    const presence = page.getByTestId(workbookPresenceSummaryTestId());
    await expect(presence).toHaveAttribute(
      "aria-label",
      /^24 collaborators present on this sheet/,
      { timeout: 10_000 },
    );
    for (
      let index = 0;
      index < interactiveMeasurementSamplePolicy.totalSamples;
      index++
    ) {
      if (index > 0)
        await restoreDefaultQuery(
          page,
          predicate.operation,
          firstDefaultRow.record_id,
        );
      await expect(presence).toHaveAttribute(
        "aria-label",
        /^24 collaborators present on this sheet/,
      );
      const action = await prepareAction(page, predicate.operation);
      samples.push(
        await measureViewChange(
          page,
          predicateId,
          destination.map((row) => ({
            recordId: row.record_id,
            minimumVersion: row.row_version,
          })),
          action,
        ),
      );
      // Exercise the keyboard entry point after recording completion; it cannot make an earlier stop pass.
      const grid = page
        .locator(`[data-workbook-view-schema-id="${view}"]`)
        .locator('[role="grid"], [role="treegrid"]');
      const renderedRows = grid.locator('[role="row"][data-grid-record-id]');
      const firstRecordId = await renderedRows
        .first()
        .getAttribute("data-grid-record-id");
      const secondRecordId = await renderedRows
        .nth(1)
        .getAttribute("data-grid-record-id");
      expect(firstRecordId).toBeTruthy();
      expect(secondRecordId).toBeTruthy();
      if (!firstRecordId || !secondRecordId)
        throw new Error("View-change keyboard targets are missing");
      const entry = page
        .getByTestId(
          rowCellTestId(firstRecordId, "timeline.activity_synopsis_text"),
        )
        .locator("xpath=ancestor::*[@role='gridcell'][1]");
      await activateCommittedGridCell(entry);
      const keyboardAnchor = () =>
        grid.evaluate((element) => {
          const active = document.activeElement;
          if (!active || !element.contains(active)) return null;
          return {
            descendant: active.getAttribute("aria-activedescendant"),
            row: active.closest('[role="row"]')?.getAttribute("aria-rowindex"),
            recordId: active
              .closest('[role="row"]')
              ?.getAttribute("data-grid-record-id"),
            column: active
              .closest('[role="gridcell"], [role="columnheader"]')
              ?.getAttribute("aria-colindex"),
            group: active.getAttribute("data-cartulary-grid-group-id"),
          };
        });
      const beforeNavigation = await keyboardAnchor();
      if (beforeNavigation === null)
        throw new Error("View-change keyboard anchor is missing");
      expect(beforeNavigation?.recordId).toBe(firstRecordId);
      await entry.press("ArrowDown");
      await expect.poll(keyboardAnchor).toEqual({
        ...beforeNavigation,
        recordId: secondRecordId,
        row: String(Number(beforeNavigation.row) + 1),
      });
      await expect
        .poll(() =>
          grid.evaluate((element) => element.contains(document.activeElement)),
        )
        .toBe(true);
      await expect(presence).toHaveAttribute(
        "aria-label",
        /^24 collaborators present on this sheet/,
      );
    }
    await traffic.stop();
    traffic = null;
    const observation = await attachMeasurementObservation(testInfo, {
      fixtureDigest: ac043FixtureDigest(),
      predicateId,
      samples,
    });
    attached = true;
    expect(observation.outcome).toBe("passed");
    expect(observation.statistics.p95_ms).toBeLessThanOrEqual(
      predicate.threshold_ms,
    );
  } catch (error) {
    if (!attached)
      await attachMeasurementObservation(testInfo, {
        fixtureDigest: ac043FixtureDigest(),
        predicateId,
        samples,
        failureReason: error instanceof Error ? error.message : String(error),
      });
    throw error;
  } finally {
    await traffic?.stop();
  }
}

test("measures Timeline sort first useful viewport within AC-044", async ({
  browser,
  page,
  workerAdmin,
  sessionTracker,
}, info) => {
  await runMeasurement(
    browser,
    page,
    workerAdmin,
    sessionTracker,
    info,
    "perf.timeline_sort.first_useful_viewport.v1",
  );
});
test("measures Timeline sort stable viewport within AC-044", async ({
  browser,
  page,
  workerAdmin,
  sessionTracker,
}, info) => {
  await runMeasurement(
    browser,
    page,
    workerAdmin,
    sessionTracker,
    info,
    "perf.timeline_sort.stable_viewport.v1",
  );
});
test("measures Timeline filter first useful viewport within AC-044", async ({
  browser,
  page,
  workerAdmin,
  sessionTracker,
}, info) => {
  await runMeasurement(
    browser,
    page,
    workerAdmin,
    sessionTracker,
    info,
    "perf.timeline_filter.first_useful_viewport.v1",
  );
});
test("measures Timeline filter stable viewport within AC-044", async ({
  browser,
  page,
  workerAdmin,
  sessionTracker,
}, info) => {
  await runMeasurement(
    browser,
    page,
    workerAdmin,
    sessionTracker,
    info,
    "perf.timeline_filter.stable_viewport.v1",
  );
});
test("measures Timeline group first useful viewport within AC-044", async ({
  browser,
  page,
  workerAdmin,
  sessionTracker,
}, info) => {
  await runMeasurement(
    browser,
    page,
    workerAdmin,
    sessionTracker,
    info,
    "perf.timeline_group.first_useful_viewport.v1",
  );
});
test("measures Timeline group stable viewport within AC-044", async ({
  browser,
  page,
  workerAdmin,
  sessionTracker,
}, info) => {
  await runMeasurement(
    browser,
    page,
    workerAdmin,
    sessionTracker,
    info,
    "perf.timeline_group.stable_viewport.v1",
  );
});
