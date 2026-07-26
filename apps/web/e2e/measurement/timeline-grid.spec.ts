import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  draftCellTestId,
  saveStateTestId,
  timelineMutationSubstrateReadyTestId,
} from "@cartulary/ui-contracts";

import { expect, test } from "../fixtures";
import { createIncident } from "../support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "../support/runtime/fixtureIdentity";
import { holdBrowserRequest as holdBrowserApiRequest } from "../support/transport/requestInterception";
import {
  measureBlankRowCreate,
  measureTypingAck,
  ordinaryMeasurementSamplePolicy,
  percentile95,
} from "./timingSupport";

const timelineViewSchemaId = "cartulary.view.timeline.v2";
const draftSummaryTestId = draftCellTestId("timeline.activity_synopsis_text");

test("measures user-visible typing_ack and blank-row-create completion within the Timeline envelope", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TIMELINE-MEASUREMENT"),
    "Timeline timeline-performance",
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: draftSummaryTestId,
  });

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
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
    expect(heldCreate.hitCount()).toBe(1);
    heldCreate.release();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
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
    clientTimingEvents: Array<{
      at: number;
      name: string;
      [key: string]: unknown;
    }>;
    committedDurationMs: number;
    networkDurationMs: number;
    recordId: string;
    rowVersion: number;
    sampleIndex: number;
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
  // Core 04 AC-043 owns this implementation threshold. This ordinary
  // measurement target is not Core 05 claim-bearing publication evidence.
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
