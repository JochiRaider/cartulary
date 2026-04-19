import { expect, type Page, test } from "@playwright/test";

import {
  apiBase,
  csrfHeaders,
  ensureAdminSession,
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

test("E-3-02 shows Syncing, Saved, and Conflict across Enter, Tab, blur, and paste completion", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E302"),
    "Phase 3 E-3-02",
  );
  const row = await createTimelineRow(page, incidentId, {
    client_txn_id: uniqueTxn("row"),
    "timeline.summary": "Alpha",
  });
  const recordId = row.record_id as string;

  await page.goto(`/?incident_id=${incidentId}`);
  const summaryInput = page.getByTestId(`row-${recordId}-summary`);
  const detailsInput = page.getByTestId(`row-${recordId}-details`);
  const sourceTextInput = page.getByTestId(`row-${recordId}-sourceText`);

  let delayed = false;
  await page.route(`**/api/v1/records/${recordId}`, async (route) => {
    if (!delayed) {
      delayed = true;
      await page.waitForTimeout(350);
    }
    await route.continue();
  });

  await summaryInput.fill("Alpha enter");
  await summaryInput.press("Enter");
  await expect(page.getByTestId("save-state")).toHaveText("Syncing");
  await expect(page.getByTestId(`row-${recordId}-row-version`)).toHaveText("2");
  await expect(page.getByTestId("save-state")).toHaveText("Saved");

  await summaryInput.fill("Alpha tab");
  await summaryInput.press("Tab");
  await expect(page.getByTestId(`row-${recordId}-row-version`)).toHaveText("3");
  await expect(page.getByTestId("save-state")).toHaveText("Saved");

  await detailsInput.fill("Blur details");
  await detailsInput.blur();
  await expect(page.getByTestId("save-state")).toHaveText("Syncing");
  await expect(page.getByTestId(`row-${recordId}-row-version`)).toHaveText("4");
  await expect(page.getByTestId("save-state")).toHaveText("Saved");

  await sourceTextInput.fill("Pasted transcript");
  await sourceTextInput.dispatchEvent("paste");
  await expect(page.getByTestId(`row-${recordId}-row-version`)).toHaveText("5");
  await expect(page.getByTestId("save-state")).toHaveText("Saved");

  let conflictInjected = false;
  await page.route(`**/api/v1/records/${recordId}`, async (route) => {
    if (!conflictInjected) {
      conflictInjected = true;
      await route.fulfill({
        status: 409,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            code: "conflict",
            message: "record version conflict",
          },
        }),
      });
      return;
    }
    await route.continue();
  });

  await summaryInput.fill("Conflict value");
  await summaryInput.blur();
  await expect(page.getByTestId("save-state")).toHaveText("Conflict");
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
  const primaryRow = await createTimelineRow(page, incidentId, {
    client_txn_id: uniqueTxn("primary"),
    "timeline.summary": "Primary row",
  });
  const replacementRow = await createTimelineRow(page, incidentId, {
    client_txn_id: uniqueTxn("replacement"),
    "timeline.summary": "Replacement row",
  });
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

test("E-3-04 replays the same patch without duplicate visible collaboration updates", async ({
  browser,
  page,
}) => {
  await ensureAdminSession(page);
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E304"),
    "Phase 3 E-3-04",
  );
  const row = await createTimelineRow(page, incidentId, {
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
  await expect(observer.getByTestId(`row-${recordId}-row-version`)).toHaveText(
    "2",
  );
  await observer.close();
});

async function createIncident(page: Page, incidentKey: string, title: string) {
  const response = await page.request.post(`${apiBase}/api/v1/incidents`, {
    headers: await csrfHeaders(page),
    data: {
      client_txn_id: uniqueTxn("incident"),
      incident_key: incidentKey,
      title,
    },
  });
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as { data: { incident_id: string } };
  return body.data.incident_id;
}

async function createTimelineRow(
  page: Page,
  incidentId: string,
  payload: Record<string, unknown>,
) {
  const response = await page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
    {
      headers: await csrfHeaders(page),
      data: payload,
    },
  );
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as {
    data: {
      row: {
        record_id: string;
      };
    };
  };
  return body.data.row as Record<string, unknown>;
}
