import {
  applyFilterChip,
  changeGrouping,
  gridGroupRowTestId,
  sortByHeader,
} from "@cartulary/test-utils";
import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  apiBase,
  createIncident,
  createViewRow,
  csrfHeaders,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";

const timelineViewSchemaId = "cartulary.view.timeline.v1";

test("support Phase 3 grid controls submit sort, filter, and group query members through the shared route", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("S301"),
    "Phase 3 support query controls",
  );
  const alphaRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("s301-alpha"),
    "timeline.summary": "Alpha summary",
  });
  const betaRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("s301-beta"),
    "timeline.summary": "Beta summary",
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(`row-${alphaRow.record_id}-summary`),
  ).toHaveValue("Alpha summary");

  await page.getByTestId(`row-${betaRow.record_id}-mark-reviewed`).click();
  await expect(
    page.getByTestId(`row-${betaRow.record_id}-capture-state`),
  ).toHaveText("reviewed");

  const sortRequest = waitForTimelineQuery(page, incidentId);
  await sortByHeader(page, "timeline", "timeline.summary");
  expect(readPostBody(await sortRequest)).toEqual({
    sort: [{ direction: "asc", field_key: "timeline.summary" }],
  });

  const filterRequest = waitForTimelineQuery(page, incidentId);
  await applyFilterChip(page, "timeline", "timeline.capture_state", "reviewed");
  expect(readPostBody(await filterRequest)).toEqual({
    filters: [
      {
        arg: { value: "reviewed" },
        field_key: "timeline.capture_state",
        op: "eq",
      },
    ],
    sort: [{ direction: "asc", field_key: "timeline.summary" }],
  });
  await expect(
    page.getByTestId(`row-${betaRow.record_id}-summary`),
  ).toHaveValue("Beta summary");

  const groupRequest = waitForTimelineQuery(page, incidentId);
  await changeGrouping(page, "timeline", "timeline.capture_state");
  expect(readPostBody(await groupRequest)).toEqual({
    filters: [
      {
        arg: { value: "reviewed" },
        field_key: "timeline.capture_state",
        op: "eq",
      },
    ],
    group_by: "timeline.capture_state",
    sort: [
      { direction: "asc", field_key: "timeline.capture_state" },
      { direction: "asc", field_key: "timeline.summary" },
    ],
  });
  await expect(
    page.getByTestId(
      gridGroupRowTestId("timeline", "timeline.capture_state", "reviewed"),
    ),
  ).toBeVisible();
});

test("support Phase 3 keeps a pending edit anchored to its record under sort, filter, group, and live invalidation", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("S302"),
    "Phase 3 support anchor stability",
  );
  const alphaRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("s302-alpha"),
    "timeline.summary": "Alpha summary",
  });
  const betaRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("s302-beta"),
    "timeline.summary": "Beta summary",
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await page.getByTestId(`row-${betaRow.record_id}-mark-reviewed`).click();
  await expect(
    page.getByTestId(`row-${betaRow.record_id}-capture-state`),
  ).toHaveText("reviewed");

  await sortByHeader(page, "timeline", "timeline.summary");
  await applyFilterChip(page, "timeline", "timeline.has_evidence", "false");
  await changeGrouping(page, "timeline", "timeline.capture_state");

  const delayedPatchRoute = `${apiBase}/api/v1/records/${alphaRow.record_id}`;
  await page.route(delayedPatchRoute, async (route) => {
    await page.waitForTimeout(300);
    await route.continue();
  });

  const alphaSummary = page.getByTestId(`row-${alphaRow.record_id}-summary`);
  await alphaSummary.fill("Zulu anchored");
  await alphaSummary.press("Enter");
  await expect(page.getByTestId("save-state")).toHaveText("Syncing");

  const betaVersion = Number.parseInt(
    (await page
      .getByTestId(`row-${betaRow.record_id}-row-version`)
      .textContent()) ?? "0",
    10,
  );
  const invalidationResponse = await page.request.patch(
    `${apiBase}/api/v1/records/${betaRow.record_id}`,
    {
      headers: await csrfHeaders(page),
      data: {
        view_schema_id: timelineViewSchemaId,
        base_row_version: betaVersion,
        client_txn_id: uniqueTxn("s302-invalidation"),
        changes: [
          {
            field_key: "timeline.details",
            value: "Support invalidation",
          },
        ],
      },
    },
  );
  expect(invalidationResponse.ok()).toBeTruthy();

  await expect(page.getByTestId("save-state")).toHaveText("Saved");
  await page.unroute(delayedPatchRoute);
  await expect(
    page.getByTestId(`row-${alphaRow.record_id}-row-version`),
  ).toHaveText("2");
  await expect(
    page.getByTestId(`row-${alphaRow.record_id}-summary`),
  ).toHaveValue("Zulu anchored");
  const alphaCaptureState = (
    (await page
      .getByTestId(`row-${alphaRow.record_id}-capture-state`)
      .textContent()) ?? ""
  ).trim();
  expect(alphaCaptureState).not.toBe("");
  await expect(
    page.getByTestId(
      gridGroupRowTestId(
        "timeline",
        "timeline.capture_state",
        alphaCaptureState,
      ),
    ),
  ).toBeVisible();
});

function waitForTimelineQuery(page: Page, incidentId: string) {
  return page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
        ),
  );
}

function readPostBody(request: { postData: () => string | null }) {
  return JSON.parse(request.postData() ?? "{}") as Record<string, unknown>;
}
