import {
  draftCellTestId,
  gridSortHeaderTestId,
  pendingQueueNoticeTestId,
  rowCellTestId,
  saveStateTestId,
  timelineRowVersionTestId,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  apiBase,
  createIncident,
  createViewRow,
  gridSavedRows,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import { readTimelineMutation, waitForTimelinePatch } from "./phase4Helpers";

const timelineViewSchemaId = "cartulary.view.timeline.v1";
const exactScenarioTitle =
  "FE-I-P4-01 Verify Timeline query response rows render full view_row_v1 cells and preserve row identity through create, patch, validation error, and refresh.";
const timelineSchemaFieldKeys = [
  "timeline.occurred_at",
  "timeline.summary",
  "timeline.details",
  "timeline.source_text",
  "timeline.host_refs",
  "timeline.identity_refs",
  "timeline.evidence_count",
  "timeline.tags",
  "timeline.attached_evidence_ids",
  "timeline.edited_at",
  "timeline.recorded_at",
  "timeline.sort_ts",
  "timeline.capture_state",
  "timeline.replacement_record_id",
  "timeline.occurred_day",
  "timeline.recorded_day",
  "timeline.has_evidence",
  "timeline.has_unresolved_mentions",
] as const;

type ViewApiCell = {
  value: unknown;
  [key: string]: unknown;
};

type ViewApiRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, ViewApiCell>;
  group_values?: Record<string, unknown>;
  [key: string]: unknown;
};

type QueryEnvelope = {
  data: {
    incident_id: string;
    rows: ViewApiRow[];
    view_schema_id: string;
  };
  meta?: {
    query?: {
      filters?: unknown;
      group_by?: unknown;
      sort?: unknown;
    };
  };
};

function expectFullTimelineRow(row: ViewApiRow) {
  expect(Object.keys(row.cells).sort()).toEqual(
    [...timelineSchemaFieldKeys].sort(),
  );
  expect(row.cells).not.toHaveProperty("record_id");
  expect(row.cells).not.toHaveProperty("row_version");
  expect(row.cells["timeline.replacement_record_id"]).toEqual({
    value: null,
  });
  expect(row).toHaveProperty("record_id");
  expect(row).toHaveProperty("row_version");
}

async function queryTimelineEnvelope(
  page: Page,
  incidentId: string,
  body: Record<string, unknown>,
) {
  const response = await page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
    {
      data: body,
    },
  );
  return {
    body: (await response.json()) as QueryEnvelope,
    ok: response.ok(),
    status: response.status(),
  };
}

function visibleRecordIds(page: Page) {
  return gridSavedRows(page, timelineViewSchemaId).evaluateAll((rows) =>
    rows.map((row) => row.getAttribute("data-grid-record-id") ?? ""),
  );
}

test(exactScenarioTitle, async ({ page }) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEIP401"),
    "FE-I-P4-01 Timeline query identity",
  );
  const alpha = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("feip401-alpha"),
    "timeline.occurred_at": "2026-04-10T10:00:00.000Z",
    "timeline.summary": "FE-I-P4-01 Alpha",
  });
  const beta = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("feip401-beta"),
    "timeline.occurred_at": "2026-04-10T10:05:00.000Z",
    "timeline.summary": "FE-I-P4-01 Beta",
  });

  const omittedQuery = await queryTimelineEnvelope(page, incidentId, {});
  expect(omittedQuery.ok).toBeTruthy();
  expect(omittedQuery.body.data.view_schema_id).toBe(timelineViewSchemaId);
  expect(omittedQuery.body.meta?.query).toBeTruthy();
  expect(omittedQuery.body.meta?.query?.filters).toEqual([]);
  expect(omittedQuery.body.meta?.query).not.toHaveProperty("group_by");
  expect(omittedQuery.body.meta?.query?.sort).toEqual([
    { field_key: "timeline.sort_ts", direction: "asc" },
    { field_key: "record_id", direction: "asc" },
  ]);
  for (const row of omittedQuery.body.data.rows) {
    expectFullTimelineRow(row);
  }

  const emptyArraysQuery = await queryTimelineEnvelope(page, incidentId, {
    filters: [],
    sort: [],
  });
  expect(emptyArraysQuery.ok).toBeTruthy();
  expect(emptyArraysQuery.body.meta?.query?.filters).toEqual([]);
  expect(emptyArraysQuery.body.meta?.query?.sort).toEqual([
    { field_key: "timeline.sort_ts", direction: "asc" },
    { field_key: "record_id", direction: "asc" },
  ]);

  const groupedQuery = await queryTimelineEnvelope(page, incidentId, {
    group_by: "timeline.capture_state",
  });
  expect(groupedQuery.ok).toBeTruthy();
  expect(groupedQuery.body.meta?.query?.group_by).toBe(
    "timeline.capture_state",
  );
  expect(
    groupedQuery.body.data.rows[0]?.group_values?.["timeline.capture_state"],
  ).toBe("rough");

  const nullGroupQuery = await queryTimelineEnvelope(page, incidentId, {
    group_by: null,
  });
  expect(nullGroupQuery.ok).toBe(false);
  expect(nullGroupQuery.status).toBeGreaterThanOrEqual(400);
  expect(nullGroupQuery.status).toBeLessThan(500);

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(rowCellTestId(alpha.record_id, "timeline.summary")),
  ).toHaveValue("FE-I-P4-01 Alpha");

  const betaPatchResponse = waitForTimelinePatch(page, beta.record_id);
  await page
    .getByTestId(rowCellTestId(beta.record_id, "timeline.summary"))
    .fill("FE-I-P4-01 Beta patched");
  await page
    .getByTestId(rowCellTestId(beta.record_id, "timeline.summary"))
    .press("Enter");
  const betaPatchEnvelope = await readTimelineMutation(await betaPatchResponse);
  await expect(
    page.getByTestId(timelineRowVersionTestId(beta.record_id)),
  ).toHaveText(String(betaPatchEnvelope.data.row.row_version));
  await expect(
    page.getByTestId(rowCellTestId(beta.record_id, "timeline.summary")),
  ).toHaveValue("FE-I-P4-01 Beta patched");

  const createResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
        ),
  );
  await page
    .getByTestId(draftCellTestId("timeline.summary"))
    .fill("FE-I-P4-01 Created");
  await page.getByTestId(draftCellTestId("timeline.summary")).press("Enter");
  const createEnvelope = await readTimelineMutation(await createResponse);
  const createdRecordId = createEnvelope.data.row.record_id;
  await expect(
    page.getByTestId(rowCellTestId(createdRecordId, "timeline.summary")),
  ).toHaveValue("FE-I-P4-01 Created");

  const validationResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/v1/records/${beta.record_id}`),
  );
  await page
    .getByTestId(rowCellTestId(beta.record_id, "timeline.occurred_at"))
    .fill("not-a-timestamp");
  await page
    .getByTestId(rowCellTestId(beta.record_id, "timeline.occurred_at"))
    .press("Enter");
  const validation = await validationResponse;
  expect(validation.ok()).toBe(false);
  await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeVisible();

  await page
    .getByTestId(gridSortHeaderTestId(timelineViewSchemaId, "timeline.summary"))
    .click();
  await expect
    .poll(async () => visibleRecordIds(page))
    .toContain(beta.record_id);
  await expect(
    page.getByTestId(rowCellTestId(beta.record_id, "timeline.occurred_at")),
  ).toHaveValue("not-a-timestamp");
  await expect(
    page.getByTestId(rowCellTestId(alpha.record_id, "timeline.occurred_at")),
  ).toHaveValue("2026-04-10T10:00:00Z");
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");

  const invalidRow = {
    ...alpha,
    cells: { ...alpha.cells },
  } as ViewApiRow;
  delete invalidRow.cells["timeline.attached_evidence_ids"];
  await page.route(
    `**/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
    async (route) => {
      await route.fulfill({
        body: JSON.stringify({
          data: {
            incident_id: incidentId,
            view_schema_id: timelineViewSchemaId,
            rows: [invalidRow],
            meta: {
              query: {
                filters: [],
                sort: [
                  { field_key: "timeline.sort_ts", direction: "asc" },
                  { field_key: "record_id", direction: "asc" },
                ],
              },
            },
          },
          meta: { request_id: "invalid-row-fixture" },
        }),
        contentType: "application/json",
        status: 200,
      });
    },
  );
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByText("Timeline projection load failed."),
  ).toBeVisible();
});
