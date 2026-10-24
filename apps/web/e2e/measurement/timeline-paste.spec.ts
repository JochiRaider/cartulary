import {
  cartularyAc003PerformanceContract as contract,
  draftCellTestId,
  saveStateTestId,
  timelineMutationSubstrateReadyTestId,
  workbookInspectorCloseButtonTestId,
} from "@cartulary/ui-contracts";
import {
  type Ac043TrafficDriver,
  prepareAc043Snapshot,
  startAc043SnapshotTraffic,
} from "../support/performance/ac043Snapshot";
import { expect, test } from "./fixtures";
import {
  attachPasteObservation,
  measureTimelinePaste,
  type PastePaintCompletion,
} from "./pasteTimingSupport";

test.use({ viewport: { width: 1920, height: 1080 }, deviceScaleFactor: 1 });
test.describe.configure({ mode: "serial", timeout: 30 * 60 * 1_000 });

test("measures 20 new Timeline rows and 100 visible TSV cells within AC-003", async ({
  browserScope,
  page,
  sessionTracker,
  testActor,
}, testInfo) => {
  let traffic: Ac043TrafficDriver | null = null;
  let completion: PastePaintCompletion | undefined;
  let attached = false;
  try {
    expect(testInfo.retry).toBe(0);
    const snapshot = await prepareAc043Snapshot(page, testActor);
    await page.goto(`/?incident_id=${snapshot.incidentId}`);
    await expect(
      page.getByTestId(timelineMutationSubstrateReadyTestId()),
    ).toBeVisible();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(
      page.getByTestId(
        workbookInspectorCloseButtonTestId(contract.view_schema_id),
      ),
    ).toHaveCount(0);
    const target = page.getByTestId(draftCellTestId(contract.field_keys[0]));
    // All preparation, including focusing the insertion target, precedes timing.
    await target.scrollIntoViewIfNeeded();
    await target.focus();
    await expect(target).toBeVisible();
    await target.press("Escape");
    await expect(
      target.locator("xpath=ancestor::*[@role='gridcell'][1]"),
    ).toBeFocused();
    expect(
      await page.evaluate(() => ({
        width: innerWidth,
        height: innerHeight,
        scale: devicePixelRatio,
      })),
    ).toEqual({ width: 1920, height: 1080, scale: 1 });
    traffic = await startAc043SnapshotTraffic(
      browserScope,
      sessionTracker,
      snapshot.incidentId,
      snapshot.timelineRows,
      snapshot.runtime.backgroundAccounts,
      contract.predicate_id,
    );
    expect(traffic.qualifiedBackgroundSessions).toBe(24);
    expect(traffic.sessionCount).toBe(25);
    expect(traffic.updatesPerSecond).toBe(4.8);
    completion = await measureTimelinePaste(
      page,
      snapshot.timelineRows.map((row) => row.record_id),
    );
    await traffic.stop();
    traffic = null;
    const observation = await attachPasteObservation(testInfo, completion);
    attached = true;
    expect(observation.outcome).toBe("passed");
    expect(observation.statistics.elapsed_ms).toBeLessThan(2_000);
  } catch (error) {
    if (!attached)
      await attachPasteObservation(
        testInfo,
        completion,
        error instanceof Error ? error.message : String(error),
      );
    throw error;
  } finally {
    await traffic?.stop();
  }
});
