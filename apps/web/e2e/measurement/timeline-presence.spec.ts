import { scrollGridCellIntoView } from "@cartulary/test-utils/grid";
import {
  type CartularyPresencePredicateId,
  cellPresenceMarkerTestId,
  cartularyPresencePerformanceContract as contract,
  dataTestIdSelector,
  rowCellTestId,
  rowPresenceMarkerTestId,
  timelineMutationSubstrateReadyTestId,
  workbookPresenceSummaryTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { hostsViewSchemaId } from "@cartulary/view-contracts";
import type { Browser, Page, TestInfo } from "@playwright/test";
import type { FixtureActorCredentials } from "../support/auth/sessions";
import { readCurrentSession } from "../support/auth/sessions";
import {
  type Ac043TrafficDriver,
  prepareAc043Snapshot,
  startAc043SnapshotTraffic,
} from "../support/performance/ac043Snapshot";
import { activateCommittedGridCell } from "../support/workbook/rowMutations";
import { expect, test } from "./fixtures";
import {
  attachPresenceObservation,
  measurePresence,
  type PresenceKind,
  type PresenceTarget,
} from "./presenceTimingSupport";

test.use({
  viewport: { width: 1920, height: 1080 },
  deviceScaleFactor: 1,
  actionTimeout: 10_000,
  navigationTimeout: 30_000,
});
test.describe.configure({ mode: "serial", timeout: 10 * 60 * 1_000 });

async function runPresenceMeasurement(
  browser: Browser,
  page: Page,
  testActor: FixtureActorCredentials,
  sessionTracker: Parameters<typeof startAc043SnapshotTraffic>[1],
  testInfo: TestInfo,
  kind: PresenceKind,
) {
  const predicate = contract.predicates.find((value) => value.kind === kind);
  if (predicate === undefined)
    throw new Error("Presence predicate is not admitted");
  const predicateId = predicate.predicate_id as CartularyPresencePredicateId;
  let traffic: Ac043TrafficDriver | null = null;
  let target: PresenceTarget | undefined;
  let completion: Awaited<ReturnType<typeof measurePresence>> | undefined;
  let attached = false;
  try {
    expect(testInfo.retry).toBe(0);
    const snapshot = await prepareAc043Snapshot(page, testActor);
    const actor = await readCurrentSession(page);
    const firstPage = [...snapshot.timelineRows]
      .sort((a, b) =>
        a.record_id < b.record_id ? -1 : a.record_id > b.record_id ? 1 : 0,
      )
      .slice(0, 100);
    const excluded = new Set(firstPage.map((row) => row.record_id));
    const [first, second] = firstPage;
    if (first === undefined || second === undefined)
      throw new Error("Presence fixture requires two target rows");
    const firstId = first.record_id,
      secondId = second.record_id;
    await page.goto(`/?incident_id=${snapshot.incidentId}`);
    await expect(
      page.getByTestId(timelineMutationSubstrateReadyTestId()),
    ).toBeVisible();
    if (kind === "surface") {
      await page.getByRole("tab", { name: "Hosts", exact: true }).click();
      await expect(
        page.getByTestId(workbookShellReadyTestId()),
      ).toHaveAttribute("data-active-view-schema-id", hostsViewSchemaId);
    }
    traffic = await startAc043SnapshotTraffic(
      browser,
      sessionTracker,
      snapshot.incidentId,
      snapshot.timelineRows.filter((row) => !excluded.has(row.record_id)),
      snapshot.runtime.backgroundAccounts,
      predicateId,
      { workbookObserver: true },
    );
    expect(traffic.qualifiedBackgroundSessions).toBe(24);
    expect(traffic.sessionCount).toBe(25);
    expect(traffic.updatesPerSecond).toBe(4.8);
    const observer = traffic.observerPage;
    if (observer === null)
      throw new Error("Presence observer was not prepared");
    await expect(
      observer.getByTestId(timelineMutationSubstrateReadyTestId()),
    ).toBeVisible();
    await expect(
      observer.getByTestId(workbookPresenceSummaryTestId()),
    ).toHaveAttribute(
      "aria-label",
      new RegExp(
        `^${kind === "surface" ? 23 : 24} collaborators present on this sheet`,
      ),
    );
    const actorGlyph = `[data-presence-user-id="${actor.user_id}"]`;
    let action: () => Promise<void>;
    if (kind === "surface") {
      const option = page.getByRole("tab", { name: "Timeline", exact: true });
      await expect(option).toBeVisible();
      target = {
        kind,
        actorUserId: actor.user_id,
        markerSelector: dataTestIdSelector(workbookPresenceSummaryTestId()),
        recordId: null,
        fieldKey: null,
      };
      action = () => option.click();
    } else {
      // Both independently rendered workbooks must expose the target before timing.
      for (const workbook of [observer, page]) {
        await scrollGridCellIntoView({
          page: workbook,
          surface: contract.view_schema_id,
          recordId: firstId,
          cellKey: contract.field_key,
        });
        const cell = workbook
          .getByTestId(rowCellTestId(firstId, contract.field_key))
          .locator("xpath=ancestor::*[@role='gridcell'][1]");
        await cell.scrollIntoViewIfNeeded();
        await expect
          .poll(() =>
            cell.evaluate((element) => {
              const grid = element.closest('[role="grid"], [role="treegrid"]');
              if (grid === null) return false;
              const viewport = grid.getBoundingClientRect();
              const box = element.getBoundingClientRect();
              return (
                box.left >= Math.max(0, viewport.left) &&
                box.right <= Math.min(innerWidth, viewport.right) &&
                box.top >= Math.max(0, viewport.top) &&
                box.bottom <= Math.min(innerHeight, viewport.bottom)
              );
            }),
          )
          .toBe(true);
      }
      const firstContent = page.getByTestId(
        rowCellTestId(firstId, contract.field_key),
      );
      const firstCell = firstContent.locator(
        "xpath=ancestor::*[@role='gridcell'][1]",
      );
      await activateCommittedGridCell(firstCell);
      await expect(
        observer
          .getByTestId(rowPresenceMarkerTestId(firstId))
          .locator(`${actorGlyph}[data-presence-mode="focused"]`),
      ).toBeVisible();
      const recordId = kind === "row" ? secondId : firstId;
      const marker =
        kind === "row"
          ? rowPresenceMarkerTestId(recordId)
          : cellPresenceMarkerTestId(recordId, contract.field_key);
      target = {
        kind,
        actorUserId: actor.user_id,
        markerSelector: dataTestIdSelector(marker),
        recordId,
        fieldKey: kind === "cell" ? contract.field_key : null,
      };
      action =
        kind === "row"
          ? () => firstCell.press("ArrowDown")
          : () => firstContent.click();
    }
    completion = await measurePresence(page, observer, target, action);
    await expect(
      observer.getByTestId(workbookPresenceSummaryTestId()),
    ).toHaveAttribute("aria-label", /^24 collaborators present on this sheet/);
    await traffic.stop();
    traffic = null;
    const observation = await attachPresenceObservation(
      testInfo,
      predicateId,
      target,
      completion,
    );
    attached = true;
    expect(observation.outcome).toBe("passed");
    expect(completion.elapsed).toBeLessThanOrEqual(contract.threshold_ms);
  } catch (error) {
    if (traffic?.observerPage) {
      await testInfo.attach("presence-observer-failure", {
        body: await traffic.observerPage.screenshot(),
        contentType: "image/png",
      });
    }
    if (!attached)
      await attachPresenceObservation(
        testInfo,
        predicateId,
        target,
        completion,
        "incomplete",
      );
    throw error;
  } finally {
    await traffic?.stop();
  }
}

test("measures Timeline surface presence acceptance to remote paint within AC-008", async ({
  browser,
  page,
  testActor,
  sessionTracker,
}, info) => {
  await runPresenceMeasurement(
    browser,
    page,
    testActor,
    sessionTracker,
    info,
    "surface",
  );
});
test("measures Timeline row presence acceptance to remote paint within AC-008", async ({
  browser,
  page,
  testActor,
  sessionTracker,
}, info) => {
  await runPresenceMeasurement(
    browser,
    page,
    testActor,
    sessionTracker,
    info,
    "row",
  );
});
test("measures Timeline cell presence acceptance to remote paint within AC-008", async ({
  browser,
  page,
  testActor,
  sessionTracker,
}, info) => {
  await runPresenceMeasurement(
    browser,
    page,
    testActor,
    sessionTracker,
    info,
    "cell",
  );
});
