import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  cartularyAc043PerformanceContract,
  draftCellTestId,
  gridSortHeaderTestId,
  rowCellTestId,
  saveStateTestId,
  timelineMutationSubstrateReadyTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import type { Locator, Page, TestInfo } from "@playwright/test";

import { expect, test } from "../fixtures";
import { createIncident } from "../support/incidents/fixtures";
import {
  type Ac043Fixture,
  type Ac043TrafficDriver,
  assembleAc043Fixture,
  startAc043Traffic,
} from "../support/performance/ac043Fixture";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "../support/runtime/fixtureIdentity";
import {
  ac043FixtureDigest,
  attachMeasurementSummary,
  interactiveMeasurementSamplePolicy,
  type MeasurementSample,
  measureBlankRowCreate,
  measureFocusEdit,
  measureSelectionDown,
  measureTypingAck,
  performancePredicate,
} from "./timingSupport";

test.describe.configure({ mode: "serial", timeout: 30 * 60 * 1_000 });

type PreparedMeasurement = {
  fixture: Ac043Fixture;
  fixtureDigest: string;
  incidentId: string;
  visibleRecordIds: readonly [string, string];
};

async function prepareMeasurement(
  page: Page,
  label: string,
): Promise<PreparedMeasurement> {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey(`AC043-${label}`),
    `AC-043 supported envelope ${label}`,
  );
  const fixture = await assembleAc043Fixture(page, incidentId);
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await scrollGridTargetIntoView({
    page,
    surface: "cartulary.view.timeline.v2",
    targetTestId: gridSortHeaderTestId(
      "cartulary.view.timeline.v2",
      "timeline.activity_synopsis_text",
    ),
    timeoutMs: 10_000,
  });
  const visibleRecordIds = await page
    .locator("[data-grid-record-id]")
    .evaluateAll((rows) =>
      Array.from(
        new Set(
          rows
            .map((row) => row.getAttribute("data-grid-record-id") ?? "")
            .filter((recordId) => recordId !== ""),
        ),
      ).slice(0, 2),
    );
  if (visibleRecordIds.length !== 2) {
    throw new Error(
      "AC-043 fixture did not expose two deterministic Timeline rows",
    );
  }
  return {
    fixture,
    fixtureDigest: ac043FixtureDigest(),
    incidentId,
    visibleRecordIds: [
      visibleRecordIds[0] as string,
      visibleRecordIds[1] as string,
    ],
  };
}

function semanticGridCell(content: Locator) {
  return content.locator("xpath=ancestor::*[@role='gridcell'][1]");
}

async function activateSemanticGridCell(content: Locator) {
  await expect(content).toBeVisible({ timeout: 10_000 });
  const cell = semanticGridCell(content);
  await cell.dispatchEvent("mousedown", { button: 0 });
  await cell.focus();
}

async function attachSummary(
  testInfo: TestInfo,
  input: {
    fixtureDigest: string;
    predicateId: Parameters<typeof performancePredicate>[0];
    samples: MeasurementSample[];
  },
) {
  return attachMeasurementSummary(testInfo, input);
}

function assertSummary(
  summary: Awaited<ReturnType<typeof attachMeasurementSummary>>,
  predicateId: Parameters<typeof performancePredicate>[0],
) {
  const predicate = performancePredicate(predicateId);
  expect(summary.outcome).toBe("passed");
  expect(summary.p95_ms).not.toBeNull();
  expect(summary.p95_ms as number).toBeLessThanOrEqual(predicate.thresholdMs);
}

test("measures paint-qualified Timeline summary ArrowDown selection within AC-043", async ({
  browser,
  page,
  sessionTracker,
}, testInfo) => {
  const samples: MeasurementSample[] = [];
  const fixtureDigest = ac043FixtureDigest();
  let attached = false;
  let traffic: Ac043TrafficDriver | null = null;
  try {
    const prepared =
      await test.step("assemble and validate the supported-envelope fixture", () =>
        prepareMeasurement(page, "selection"));
    const [fromRecordId, toRecordId] = prepared.visibleRecordIds;
    const excluded = new Set(prepared.visibleRecordIds);
    traffic =
      await test.step("establish and qualify 25-session background traffic", () =>
        startAc043Traffic(
          browser,
          sessionTracker,
          page,
          prepared.incidentId,
          prepared.fixture.timelineRows.filter(
            (row) => !excluded.has(row.record_id),
          ),
          "perf.timeline_summary_selection_down.v1",
        ));
    expect(traffic.sessionCount).toBe(
      cartularyAc043PerformanceContract.fixture.analystSessions,
    );
    expect(traffic.updatesPerSecond).toBe(
      cartularyAc043PerformanceContract.fixture.backgroundUpdatesPerSecond,
    );
    expect(traffic.qualifiedBackgroundSessions).toBe(
      cartularyAc043PerformanceContract.fixture.backgroundSessions,
    );
    const source = page.getByTestId(
      rowCellTestId(fromRecordId, "timeline.activity_synopsis_text"),
    );
    await test.step("collect one warm-up and 100 selection samples", async () => {
      await scrollGridTargetIntoView({
        page,
        surface: "cartulary.view.timeline.v2",
        targetTestId: rowCellTestId(
          fromRecordId,
          "timeline.activity_synopsis_text",
        ),
        timeoutMs: 10_000,
      });
      for (
        let sampleIndex = 0;
        sampleIndex < interactiveMeasurementSamplePolicy.totalSamples;
        sampleIndex += 1
      ) {
        await activateSemanticGridCell(source);
        samples.push(
          await measureSelectionDown(page, { fromRecordId, toRecordId }),
        );
      }
    });
    await test.step("stop background traffic", () => traffic?.stop());
    traffic = null;
    const summary = await attachSummary(testInfo, {
      fixtureDigest: prepared.fixtureDigest,
      predicateId: "perf.timeline_summary_selection_down.v1",
      samples,
    });
    attached = true;
    assertSummary(summary, "perf.timeline_summary_selection_down.v1");
  } catch (error) {
    if (!attached) {
      await attachMeasurementSummary(testInfo, {
        failureReason: error instanceof Error ? error.message : String(error),
        fixtureDigest,
        predicateId: "perf.timeline_summary_selection_down.v1",
        samples,
      });
    }
    throw error;
  } finally {
    await traffic?.stop();
  }
});

test("measures paint-qualified Timeline summary Enter focus within AC-043", async ({
  browser,
  page,
  sessionTracker,
}, testInfo) => {
  const samples: MeasurementSample[] = [];
  const fixtureDigest = ac043FixtureDigest();
  let attached = false;
  let traffic: Ac043TrafficDriver | null = null;
  try {
    const prepared =
      await test.step("assemble and validate the supported-envelope fixture", () =>
        prepareMeasurement(page, "focus"));
    const recordId = prepared.visibleRecordIds[0];
    const excluded = new Set(prepared.visibleRecordIds);
    traffic =
      await test.step("establish and qualify 25-session background traffic", () =>
        startAc043Traffic(
          browser,
          sessionTracker,
          page,
          prepared.incidentId,
          prepared.fixture.timelineRows.filter(
            (row) => !excluded.has(row.record_id),
          ),
          "perf.timeline_summary_focus_edit.v1",
        ));
    const content = page.getByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    );
    const editor = page.getByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.activity_synopsis_text",
        recordId,
        surface: "grid",
      }),
    );
    await scrollGridTargetIntoView({
      page,
      surface: "cartulary.view.timeline.v2",
      targetTestId: rowCellTestId(recordId, "timeline.activity_synopsis_text"),
      timeoutMs: 10_000,
    });
    for (
      let sampleIndex = 0;
      sampleIndex < interactiveMeasurementSamplePolicy.totalSamples;
      sampleIndex += 1
    ) {
      await activateSemanticGridCell(content);
      samples.push(await measureFocusEdit(page, { recordId }));
      await editor.press("Escape");
    }
    await traffic.stop();
    traffic = null;
    const summary = await attachSummary(testInfo, {
      fixtureDigest: prepared.fixtureDigest,
      predicateId: "perf.timeline_summary_focus_edit.v1",
      samples,
    });
    attached = true;
    assertSummary(summary, "perf.timeline_summary_focus_edit.v1");
  } catch (error) {
    if (!attached) {
      await attachMeasurementSummary(testInfo, {
        failureReason: error instanceof Error ? error.message : String(error),
        fixtureDigest,
        predicateId: "perf.timeline_summary_focus_edit.v1",
        samples,
      });
    }
    throw error;
  } finally {
    await traffic?.stop();
  }
});

test("measures paint-qualified committed Timeline summary typing acknowledgment within AC-043", async ({
  browser,
  page,
  sessionTracker,
}, testInfo) => {
  const samples: MeasurementSample[] = [];
  const fixtureDigest = ac043FixtureDigest();
  let attached = false;
  let traffic: Ac043TrafficDriver | null = null;
  try {
    const prepared =
      await test.step("assemble and validate the supported-envelope fixture", () =>
        prepareMeasurement(page, "typing"));
    const recordId = prepared.visibleRecordIds[0];
    const excluded = new Set(prepared.visibleRecordIds);
    traffic =
      await test.step("establish and qualify 25-session background traffic", () =>
        startAc043Traffic(
          browser,
          sessionTracker,
          page,
          prepared.incidentId,
          prepared.fixture.timelineRows.filter(
            (row) => !excluded.has(row.record_id),
          ),
          "perf.typing_ack.v1",
        ));
    const content = page.getByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    );
    await scrollGridTargetIntoView({
      page,
      surface: "cartulary.view.timeline.v2",
      targetTestId: rowCellTestId(recordId, "timeline.activity_synopsis_text"),
      timeoutMs: 10_000,
    });
    await activateSemanticGridCell(content);
    await content.press("Enter");
    for (
      let sampleIndex = 0;
      sampleIndex < interactiveMeasurementSamplePolicy.totalSamples;
      sampleIndex += 1
    ) {
      samples.push(await measureTypingAck(page, { recordId }));
    }
    await traffic.stop();
    traffic = null;
    const summary = await attachSummary(testInfo, {
      fixtureDigest: prepared.fixtureDigest,
      predicateId: "perf.typing_ack.v1",
      samples,
    });
    attached = true;
    assertSummary(summary, "perf.typing_ack.v1");
  } catch (error) {
    if (!attached) {
      await attachMeasurementSummary(testInfo, {
        failureReason: error instanceof Error ? error.message : String(error),
        fixtureDigest,
        predicateId: "perf.typing_ack.v1",
        samples,
      });
    }
    throw error;
  } finally {
    await traffic?.stop();
  }
});

test("measures paint-qualified Timeline blank-row creation within AC-043", async ({
  browser,
  page,
  sessionTracker,
}, testInfo) => {
  const samples: MeasurementSample[] = [];
  const fixtureDigest = ac043FixtureDigest();
  let attached = false;
  let traffic: Ac043TrafficDriver | null = null;
  try {
    const prepared =
      await test.step("assemble and validate the supported-envelope fixture", () =>
        prepareMeasurement(page, "blank-row"));
    const excluded = new Set(prepared.visibleRecordIds);
    traffic =
      await test.step("establish and qualify 25-session background traffic", () =>
        startAc043Traffic(
          browser,
          sessionTracker,
          page,
          prepared.incidentId,
          prepared.fixture.timelineRows.filter(
            (row) => !excluded.has(row.record_id),
          ),
          "perf.timeline_blank_row_create.v1",
        ));
    const draft = page.getByTestId(
      draftCellTestId("timeline.activity_synopsis_text"),
    );
    for (
      let sampleIndex = 0;
      sampleIndex < interactiveMeasurementSamplePolicy.totalSamples;
      sampleIndex += 1
    ) {
      const summary = `AC-043 blank sample ${sampleIndex} ${uniqueTxn("blank-row")}`;
      await draft.fill(summary);
      const sample = await measureBlankRowCreate(page, summary);
      expect(sample.status).toBe(201);
      samples.push({ stagesMs: sample.stagesMs, totalMs: sample.totalMs });
    }
    await traffic.stop();
    traffic = null;
    const summary = await attachSummary(testInfo, {
      fixtureDigest: prepared.fixtureDigest,
      predicateId: "perf.timeline_blank_row_create.v1",
      samples,
    });
    attached = true;
    assertSummary(summary, "perf.timeline_blank_row_create.v1");
  } catch (error) {
    if (!attached) {
      await attachMeasurementSummary(testInfo, {
        failureReason: error instanceof Error ? error.message : String(error),
        fixtureDigest,
        predicateId: "perf.timeline_blank_row_create.v1",
        samples,
      });
    }
    throw error;
  } finally {
    await traffic?.stop();
  }
});
