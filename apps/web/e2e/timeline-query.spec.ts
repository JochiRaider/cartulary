import {
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  draftCellTestId,
  gridSortHeaderTestId,
  pendingQueueNoticeTestId,
  rowCellTestId,
  saveStateTestId,
  timelineRowVersionTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import { gridSavedRows } from "./pages/workbookInspector";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createViewRow } from "./support/workbook/query";
import {
  readTimelineMutation,
  waitForTimelinePatch,
} from "./support/workbook/rowMutations";

const timelineViewSchemaId = "cartulary.view.timeline.v2";
const exactScenarioTitle =
  "Verify Timeline query response rows render full view_row_v1 cells and preserve row identity through create, patch, validation error, and refresh.";
const timelineSchemaFieldKeys = [
  "timeline.date_entered_text",
  "timeline.analyst_text",
  "timeline.mitre_stage_text",
  "timeline.device_object_text",
  "timeline.ip_address_text",
  "timeline.activity_utc_text",
  "timeline.activity_local_text",
  "timeline.raw_activity_text",
  "timeline.activity_synopsis_text",
  "timeline.data_source_text",
  "timeline.host_refs",
  "timeline.identity_refs",
  "timeline.evidence_count",
  "timeline.tags",
  "timeline.attached_evidence_ids",
  "timeline.edited_at",
  "timeline.recorded_at",
  "timeline.activity_sort_ts",
  "timeline.activity_time_pair_state",
  "timeline.capture_state",
  "timeline.replacement_record_id",
  "timeline.date_entered_sort_day",
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
    uniqueIncidentKey("TIMELINEQUERY"),
    "integration.mutation-lifecycle.row-01 Timeline query identity",
  );
  const alpha = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("timeline-query-alpha"),
    "timeline.activity_utc_text": "2026-04-10T10:00:00.000Z",
    "timeline.activity_synopsis_text":
      "integration.mutation-lifecycle.row-01 Alpha",
  });
  const beta = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("timeline-query-beta"),
    "timeline.activity_utc_text": "2026-04-10T10:05:00.000Z",
    "timeline.activity_synopsis_text":
      "integration.mutation-lifecycle.row-01 Beta",
  });

  const omittedQuery = await queryTimelineEnvelope(page, incidentId, {});
  expect(omittedQuery.ok).toBeTruthy();
  expect(omittedQuery.body.data.view_schema_id).toBe(timelineViewSchemaId);
  expect(omittedQuery.body.meta?.query).toBeTruthy();
  expect(omittedQuery.body.meta?.query?.filters).toEqual([]);
  expect(omittedQuery.body.meta?.query).not.toHaveProperty("group_by");
  expect(omittedQuery.body.meta?.query?.sort).toEqual([
    { field_key: "timeline.activity_sort_ts", direction: "asc" },
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
    { field_key: "timeline.activity_sort_ts", direction: "asc" },
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
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: alpha.record_id,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(alpha.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("integration.mutation-lifecycle.row-01 Alpha");

  const betaPatchResponse = waitForTimelinePatch(page, beta.record_id);
  const betaSummaryCell = page.getByTestId(
    rowCellTestId(beta.record_id, "timeline.activity_synopsis_text"),
  );
  await betaSummaryCell.click();
  const betaSummaryEditor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.activity_synopsis_text",
      recordId: beta.record_id,
      surface: "grid",
    }),
  );
  await betaSummaryEditor.fill(
    "integration.mutation-lifecycle.row-01 Beta patched",
  );
  await betaSummaryEditor.press("Enter");
  const betaPatchEnvelope = await readTimelineMutation(await betaPatchResponse);
  await expect(
    page.getByTestId(timelineRowVersionTestId(beta.record_id)),
  ).toHaveText(String(betaPatchEnvelope.data.row.row_version));
  await expect(
    page.getByTestId(
      rowCellTestId(beta.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("integration.mutation-lifecycle.row-01 Beta patched");

  const createResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
        ),
  );
  const draftSummaryTestId = draftCellTestId("timeline.activity_synopsis_text");
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: draftSummaryTestId,
  });
  await page
    .getByTestId(draftSummaryTestId)
    .fill("integration.mutation-lifecycle.row-01 Created");
  await page.getByTestId(draftSummaryTestId).press("Enter");
  const createEnvelope = await readTimelineMutation(await createResponse);
  const createdRecordId = createEnvelope.data.row.record_id;
  await expect(
    page.getByTestId(
      rowCellTestId(createdRecordId, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("integration.mutation-lifecycle.row-01 Created");
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

  await scrollGridCellIntoView({
    cellKey: "timeline.activity_utc_text",
    page,
    recordId: beta.record_id,
    surface: timelineViewSchemaId,
  });
  const betaOccurredAtCell = page.getByTestId(
    rowCellTestId(beta.record_id, "timeline.activity_utc_text"),
  );
  await expect(betaOccurredAtCell).toHaveCount(1);
  await betaOccurredAtCell.scrollIntoViewIfNeeded();
  await betaOccurredAtCell.click();
  const betaOccurredAtEditor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.activity_utc_text",
      recordId: beta.record_id,
      surface: "grid",
    }),
  );
  await betaOccurredAtEditor.fill("not-a-timestamp");
  await expect(betaOccurredAtEditor).toHaveValue("not-a-timestamp");
  const validationResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/v1/records/${beta.record_id}`),
  );
  await betaOccurredAtEditor.press("Enter");
  const validation = await validationResponse;
  const validationEnvelope = await readTimelineMutation(validation);
  expect(
    validationEnvelope.data.row.cells["timeline.activity_utc_text"]?.value,
  ).toBe("not-a-timestamp");
  await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);

  await page
    .getByTestId(
      gridSortHeaderTestId(
        timelineViewSchemaId,
        "timeline.activity_synopsis_text",
      ),
    )
    .click();
  await expect
    .poll(async () => visibleRecordIds(page))
    .toContain(beta.record_id);
  await expect(
    page.getByTestId(
      rowCellTestId(beta.record_id, "timeline.activity_utc_text"),
    ),
  ).toHaveText("not-a-timestamp");
  await expect(
    page.getByTestId(
      rowCellTestId(alpha.record_id, "timeline.activity_utc_text"),
    ),
  ).toHaveText("2026-04-10T10:00:00.000Z");
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

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
                  { field_key: "timeline.activity_sort_ts", direction: "asc" },
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
