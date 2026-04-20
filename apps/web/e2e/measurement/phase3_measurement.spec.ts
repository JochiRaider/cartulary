import { expect, test } from "../fixtures";

import {
  createIncident,
  measureBlankRowCreate,
  measureTypingAck,
  percentile95,
  uniqueIncidentKey,
  uniqueTxn,
} from "../helpers";

test("E-3-02 measures user-visible typing_ack and blank-row-create completion within the Phase 3 envelope", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E302"),
    "Phase 3 E-3-02",
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByText("Timeline mutation substrate")).toBeVisible();

  const draftSummary = page.getByTestId("draft-row-summary");
  const typingSamples: number[] = [];
  for (let sampleIndex = 0; sampleIndex < 13; sampleIndex += 1) {
    const appendedCharacter = String.fromCharCode(97 + (sampleIndex % 26));
    typingSamples.push(
      await measureTypingAck(page, "draft-row-summary", appendedCharacter),
    );
  }
  const typingP95 = percentile95(typingSamples.slice(1));
  expect(typingP95).toBeLessThanOrEqual(100);

  await draftSummary.fill("");

  const blankRowCreateSamples: number[] = [];
  for (let sampleIndex = 0; sampleIndex < 13; sampleIndex += 1) {
    const summary = `Timing sample ${sampleIndex} ${uniqueTxn("blank-row")}`;
    await draftSummary.fill(summary);
    blankRowCreateSamples.push(await measureBlankRowCreate(page, summary));
    await expect(page.getByTestId("draft-row-summary")).toBeFocused();
  }
  const blankRowCreateP95 = percentile95(blankRowCreateSamples.slice(1));
  expect(blankRowCreateP95).toBeLessThanOrEqual(150);
});
