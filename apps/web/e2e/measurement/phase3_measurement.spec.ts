import {
  draftCellTestId,
  timelineMutationSubstrateReadyTestId,
} from "@cartulary/ui-contracts";

import { expect, test } from "../fixtures";

import {
  createIncident,
  holdBrowserApiRequest,
  measureBlankRowCreate,
  measureTypingAck,
  ordinaryMeasurementSamplePolicy,
  percentile95,
  type ServerTimingMetric,
  uniqueIncidentKey,
  uniqueTxn,
} from "../helpers";

const timelineViewSchemaId = "cartulary.view.timeline.v1";
const draftSummaryTestId = draftCellTestId("timeline.summary");

test("E-3-02 measures user-visible typing_ack and blank-row-create completion within the Phase 3 envelope", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E302"),
    "Phase 3 E-3-02",
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect(page.getByTestId("save-state")).toHaveText("Saved");

  const draftSummary = page.getByTestId(draftSummaryTestId);
  const visibleSaveStateSummary = `Visible save state ${uniqueTxn("save-state")}`;
  const heldCreate = await holdBrowserApiRequest(page, {
    method: "POST",
    path: `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
  });
  try {
    await draftSummary.fill(visibleSaveStateSummary);
    await draftSummary.press("Enter");
    await heldCreate.waitForHit;
    await expect(page.getByTestId("save-state")).toHaveText("Syncing");
    expect(heldCreate.hitCount()).toBe(1);
    heldCreate.release();
    await expect(page.getByTestId("save-state")).toHaveText("Saved");
    await expect(page.getByTestId(draftSummaryTestId)).toBeFocused();
  } finally {
    heldCreate.release();
    await heldCreate.dispose();
  }

  const typingSamples: number[] = [];
  for (
    let sampleIndex = 0;
    sampleIndex < ordinaryMeasurementSamplePolicy.totalSamples;
    sampleIndex += 1
  ) {
    const appendedCharacter = String.fromCharCode(97 + (sampleIndex % 26));
    typingSamples.push(
      await measureTypingAck(page, draftSummaryTestId, appendedCharacter),
    );
  }
  const typingP95 = percentile95(
    typingSamples.slice(ordinaryMeasurementSamplePolicy.warmupSamples),
    { sampleLabel: "typing_ack" },
  );
  expect(typingP95).toBeLessThanOrEqual(100);

  await draftSummary.fill("");

  const blankRowCreateSamples: Array<{
    committedDurationMs: number;
    networkDurationMs: number;
    recordId: string;
    rowVersion: number;
    sampleIndex: number;
    serverTiming: string;
    serverTimingMetrics: ServerTimingMetric[];
    status: number;
    summary: string;
  }> = [];
  for (
    let sampleIndex = 0;
    sampleIndex < ordinaryMeasurementSamplePolicy.totalSamples;
    sampleIndex += 1
  ) {
    const summary = `Timing sample ${sampleIndex} ${uniqueTxn("blank-row")}`;
    await draftSummary.fill(summary);
    blankRowCreateSamples.push({
      ...(await measureBlankRowCreate(page, summary)),
      sampleIndex,
      summary,
    });
    await expect(page.getByTestId(draftSummaryTestId)).toBeFocused();
  }
  const blankRowCreateP95 = percentile95(
    blankRowCreateSamples
      .slice(ordinaryMeasurementSamplePolicy.warmupSamples)
      .map((sample) => sample.committedDurationMs),
    { sampleLabel: "timeline_blank_row_create" },
  );
  expect(
    blankRowCreateP95,
    JSON.stringify(
      {
        blankRowCreateP95,
        samples: blankRowCreateSamples,
      },
      null,
      2,
    ),
  ).toBeLessThanOrEqual(150);
});
