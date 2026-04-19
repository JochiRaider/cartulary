import { expect, test } from "@playwright/test";

import {
  apiBase,
  createIncident,
  createViewRow,
  csrfHeaders,
  ensureAdminSession,
  fetchTimelineRecordChangeCount,
  fetchTimelineRecordSubstrate,
  measureBlankRowCreate,
  measureTypingAck,
  percentile95,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";

const timelineViewSchemaId = "cartulary.view.timeline.v1";

test("E-3-01 creates a Timeline row in-grid and continues editing on the draft row", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E301"),
    "Phase 3 E-3-01",
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByText("Timeline mutation substrate")).toBeVisible();

  const draftSummary = page.getByTestId("draft-row-summary");
  await draftSummary.fill("First browser fact");
  await draftSummary.press("Enter");

  await expect(page.getByTestId("save-state")).toHaveText("Saved");
  await expect(page.locator("tbody tr")).toHaveCount(2);
  await expect(
    page.locator('[data-testid^="row-"][data-testid$="-summary"]').first(),
  ).toHaveValue("First browser fact");
  await expect(page.getByTestId("draft-row-summary")).toBeFocused();
});

test("E-3-02 measures user-visible typing_ack and blank-row-create completion within the Phase 3 envelope", async ({
  page,
}) => {
  await ensureAdminSession(page);
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

test("E-3-03 drives review, demotion, and supersede through the visible workbook surface", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E303"),
    "Phase 3 E-3-03",
  );
  const primaryRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("primary"),
      "timeline.summary": "Primary row",
    },
  );
  const replacementRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("replacement"),
      "timeline.summary": "Replacement row",
    },
  );
  const recordId = primaryRow.record_id as string;
  const replacementId = replacementRow.record_id as string;

  await page.goto(`/?incident_id=${incidentId}`);

  await page.getByTestId(`row-${recordId}-mark-reviewed`).click();
  await expect(page.getByTestId(`row-${recordId}-capture-state`)).toHaveText(
    "reviewed",
  );
  await expect(page.getByTestId(`row-${recordId}-row-version`)).toHaveText("2");

  const detailsInput = page.getByTestId(`row-${recordId}-details`);
  await detailsInput.fill("Material edit after review");
  await page.getByTestId("timeline-blur-surface").click();
  await expect(page.getByTestId(`row-${recordId}-capture-state`)).toHaveText(
    "enriched",
  );
  await expect(page.getByTestId(`row-${recordId}-row-version`)).toHaveText("3");

  await page.getByTestId(`row-${recordId}-replacement-id`).fill(replacementId);
  await page.getByTestId(`row-${recordId}-supersede`).click();
  await expect(page.getByTestId(`row-${recordId}-capture-state`)).toHaveText(
    "superseded",
  );
  await expect(page.getByTestId(`row-${recordId}-row-version`)).toHaveText("4");
  await expect(
    page.getByTestId(`row-${recordId}-mark-reviewed`),
  ).toBeDisabled();
});

test("E-3-04 uses a disclosed hybrid replay harness to prove replay avoids duplicate history and visible invalidation", async ({
  browser,
  page,
}) => {
  await ensureAdminSession(page);
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E304"),
    "Phase 3 E-3-04",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("seed"),
    "timeline.summary": "Replay row",
  });
  const recordId = row.record_id as string;

  const observer = await browser.newPage({
    storageState: await page.context().storageState(),
  });
  let observerQueryCount = 0;
  observer.on("requestfinished", (request) => {
    if (
      request.method() === "POST" &&
      request
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
        )
    ) {
      observerQueryCount += 1;
    }
  });

  await observer.goto(`http://127.0.0.1:4173/?incident_id=${incidentId}`);
  await expect(observer.getByTestId(`row-${recordId}-row-version`)).toHaveText(
    "1",
  );
  const baselineObserverQueries = observerQueryCount;
  const baselineRecordChangeCount = await fetchTimelineRecordChangeCount(
    page,
    recordId,
  );

  await page.goto(`/?incident_id=${incidentId}`);
  const summaryInput = page.getByTestId(`row-${recordId}-summary`);
  const firstPatchResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/v1/records/${recordId}`),
  );

  await summaryInput.fill("Replay row patched");
  await summaryInput.press("Enter");
  const patchResponse = await firstPatchResponse;
  const firstPatchBody = JSON.parse(
    patchResponse.request().postData() ?? "{}",
  ) as Record<string, unknown>;
  const firstPatchData = (
    (await patchResponse.json()) as { data: { change_set_id: string } }
  ).data;

  await expect(page.getByTestId(`row-${recordId}-row-version`)).toHaveText("2");
  await expect
    .poll(() => observerQueryCount, { timeout: 5_000 })
    .toBeGreaterThan(baselineObserverQueries);
  await expect(observer.getByTestId(`row-${recordId}-row-version`)).toHaveText(
    "2",
  );
  const substrateAfterFirstPatch = await fetchTimelineRecordSubstrate(
    page,
    recordId,
  );
  expect(substrateAfterFirstPatch.row_version).toBe(2);
  expect(substrateAfterFirstPatch.record_revision_count).toBe(2);
  const recordChangeCountAfterFirstPatch = await fetchTimelineRecordChangeCount(
    page,
    recordId,
  );
  expect(recordChangeCountAfterFirstPatch).toBe(baselineRecordChangeCount + 1);

  const queriesAfterFirstPatch = observerQueryCount;
  const replayResponse = await page.request.patch(
    `${apiBase}/api/v1/records/${recordId}`,
    {
      headers: await csrfHeaders(page),
      data: firstPatchBody,
    },
  );
  expect(replayResponse.status()).toBe(200);
  const replayData = (
    (await replayResponse.json()) as { data: { change_set_id: string } }
  ).data;
  expect(replayData.change_set_id).toBe(firstPatchData.change_set_id);

  await page.waitForTimeout(500);
  expect(observerQueryCount).toBe(queriesAfterFirstPatch);
  const substrateAfterReplay = await fetchTimelineRecordSubstrate(
    page,
    recordId,
  );
  expect(substrateAfterReplay.row_version).toBe(
    substrateAfterFirstPatch.row_version,
  );
  expect(substrateAfterReplay.record_revision_count).toBe(
    substrateAfterFirstPatch.record_revision_count,
  );
  expect(await fetchTimelineRecordChangeCount(page, recordId)).toBe(
    recordChangeCountAfterFirstPatch,
  );
  await expect(observer.getByTestId(`row-${recordId}-row-version`)).toHaveText(
    "2",
  );
  await observer.close();
});
