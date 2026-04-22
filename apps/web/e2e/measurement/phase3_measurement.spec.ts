import type { Route } from "@playwright/test";

import { expect, test } from "../fixtures";

import {
  createIncident,
  measureBlankRowCreate,
  measureTypingAck,
  percentile95,
  uniqueIncidentKey,
  uniqueTxn,
} from "../helpers";

const timelineViewSchemaId = "cartulary.view.timeline.v1";

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
  await expect(page.getByTestId("save-state")).toHaveText("Saved");

  const draftSummary = page.getByTestId("draft-row-summary");
  const delayedCreateRoute = `**/api/v1/incidents/*/views/${timelineViewSchemaId}/rows`;
  let delayedCreateHitCount = 0;
  const routeHandler = async (route: Route) => {
    delayedCreateHitCount += 1;
    if (delayedCreateHitCount > 1) {
      await route.fallback();
      return;
    }
    await page.waitForTimeout(150);
    await route.fallback();
  };
  await page.route(delayedCreateRoute, routeHandler);

  const visibleSaveStateSummary = `Visible save state ${uniqueTxn("save-state")}`;
  await draftSummary.fill(visibleSaveStateSummary);
  await draftSummary.press("Enter");
  await expect(page.getByTestId("save-state")).toHaveText("Syncing");
  await expect(page.getByTestId("save-state")).toHaveText("Saved");
  expect(delayedCreateHitCount).toBe(1);
  await expect(page.getByTestId("draft-row-summary")).toBeFocused();
  await page.unroute(delayedCreateRoute, routeHandler);

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
