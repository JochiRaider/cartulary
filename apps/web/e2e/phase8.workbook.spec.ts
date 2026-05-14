import {
  applyFilterChip,
  changeGrouping,
  sortByHeader,
} from "@cartulary/test-utils";
import { gridGroupRowTestId, rowCellTestId } from "@cartulary/ui-contracts";
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

const notesViewSchemaId = "cartulary.view.notes.v1";
const timelineViewSchemaId = "cartulary.view.timeline.v1";

test("E-8-01 saved-view route foundation persists canonical state while browser lifecycle affordances remain pending", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E801"),
    "Phase 8 E-8-01 saved-view route foundation",
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByText("Timeline workbook shell")).toBeVisible();

  const listBefore = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}/saved-views`,
  );
  expect(listBefore.status()).toBe(200);
  const beforeBody = (await listBefore.json()) as {
    data: { saved_views: unknown[] };
  };
  expect(beforeBody.data.saved_views).toEqual([]);

  const createResponse = await page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/saved-views`,
    {
      headers: await csrfHeaders(page),
      data: {
        view_schema_id: timelineViewSchemaId,
        display_name: "  Phase 8 saved view  ",
        query_json: {},
        layout_json: {},
      },
    },
  );
  expect(createResponse.status()).toBe(201);
  const createBody = (await createResponse.json()) as {
    data: {
      display_name: string;
      layout_json: {
        column_widths: unknown[];
        layout_schema_id: string;
      };
      query_json: {
        filters: unknown[];
        sort: unknown[];
      };
      scope: string;
      view_schema_id: string;
    };
  };
  expect(createBody.data.scope).toBe("private");
  expect(createBody.data.display_name).toBe("Phase 8 saved view");
  expect(createBody.data.view_schema_id).toBe(timelineViewSchemaId);
  expect(createBody.data.query_json.sort).toEqual([]);
  expect(createBody.data.query_json.filters).toEqual([]);
  expect(createBody.data.layout_json.layout_schema_id).toBe(
    "cartulary.layout.v1",
  );
  expect(createBody.data.layout_json.column_widths).toEqual([]);

  const listAfter = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}/saved-views`,
  );
  expect(listAfter.status()).toBe(200);
  const afterBody = (await listAfter.json()) as {
    data: { saved_views: Array<{ display_name: string }> };
  };
  expect(afterBody.data.saved_views.map((view) => view.display_name)).toEqual([
    "Phase 8 saved view",
  ]);
});

test("E-8-02 workbook startup falls back to Timeline for an unsupported explicit sheet", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E802"),
    "Phase 8 E-8-02 startup fallback",
  );

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=cartulary.view.unknown.v1`,
  );

  await expect(page.getByText("Timeline workbook shell")).toBeVisible();
  await expect(page.getByTestId("surface-tab-timeline")).toBeVisible();
  await expect(page).toHaveURL(/view_schema_id=cartulary\.view\.timeline\.v1/);
});

test("E-8-03 browser Timeline sort, filter, and group controls submit stable query keys", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E803"),
    "Phase 8 E-8-03 Timeline query controls",
  );
  const alpha = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e803-alpha"),
    "timeline.summary": "Alpha Phase 8",
  });
  const beta = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e803-beta"),
    "timeline.summary": "Beta Phase 8",
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(rowCellTestId(alpha.record_id as string, "summary")),
  ).toHaveValue("Alpha Phase 8");

  await page.getByTestId(`row-${beta.record_id}-mark-reviewed`).click();
  await expect(
    page.getByTestId(`row-${beta.record_id}-capture-state`),
  ).toHaveText("reviewed");

  const sortRequest = waitForViewQuery(page, incidentId, timelineViewSchemaId);
  await sortByHeader(page, "timeline", "timeline.summary");
  expect(readPostBody(await sortRequest)).toEqual({
    sort: [{ direction: "asc", field_key: "timeline.summary" }],
  });

  const filterRequest = waitForViewQuery(
    page,
    incidentId,
    timelineViewSchemaId,
  );
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

  const groupRequest = waitForViewQuery(page, incidentId, timelineViewSchemaId);
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

test("E-8-04 browser Notes full_text and prefix queries remain exact", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E804"),
    "Phase 8 E-8-04 exact search",
  );
  const alpha = await createViewRow(page, incidentId, notesViewSchemaId, {
    client_txn_id: uniqueTxn("e804-alpha"),
    "note.title": "Alpha Delta",
    "note.body": "Responder contained shell",
  });
  const powershell = await createViewRow(page, incidentId, notesViewSchemaId, {
    client_txn_id: uniqueTxn("e804-powershell"),
    "note.title": "Powershell only",
  });
  const cafe = await createViewRow(page, incidentId, notesViewSchemaId, {
    client_txn_id: uniqueTxn("e804-cafe"),
    "note.title": "Cafe note",
    "note.body": "cafe token",
  });
  const accent = await createViewRow(page, incidentId, notesViewSchemaId, {
    client_txn_id: uniqueTxn("e804-accent"),
    "note.title": "Accent note",
    "note.body": "café token",
  });
  const timelinePrefix = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e804-prefix"),
      "timeline.summary": "Alpha prefix marker",
    },
  );
  const timelinePrefixPeer = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e804-prefix-peer"),
      "timeline.summary": "Beta prefix marker",
    },
  );

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      notesViewSchemaId,
    )}`,
  );
  await expect(page.getByRole("heading", { name: "Notes" })).toBeVisible();

  const uiFilterRequest = waitForViewQuery(page, incidentId, notesViewSchemaId);
  await applyFilterChip(
    page,
    notesViewSchemaId,
    "note.full_text",
    "shell alpha shell",
  );
  expect(readPostBody(await uiFilterRequest)).toEqual({
    filters: [
      {
        arg: { query: "shell alpha shell" },
        field_key: "note.full_text",
        op: "full_text",
      },
    ],
  });
  await expect(
    page.getByTestId(rowCellTestId(alpha.record_id as string, "note.title")),
  ).toHaveText("Alpha Delta");
  await expect(
    page.getByTestId(
      rowCellTestId(powershell.record_id as string, "note.title"),
    ),
  ).toHaveCount(0);

  expect(
    rowIDs(
      await queryViewRows(page, incidentId, notesViewSchemaId, {
        filters: [
          {
            arg: { query: "shell" },
            field_key: "note.full_text",
            op: "full_text",
          },
        ],
      }),
    ),
  ).toEqual([String(alpha.record_id)]);
  expect(
    rowIDs(
      await queryViewRows(page, incidentId, notesViewSchemaId, {
        filters: [
          {
            arg: { query: "cafe" },
            field_key: "note.full_text",
            op: "full_text",
          },
        ],
      }),
    ),
  ).toEqual([String(cafe.record_id)]);
  expect(
    rowIDs(
      await queryViewRows(page, incidentId, notesViewSchemaId, {
        filters: [
          {
            arg: { query: "café" },
            field_key: "note.full_text",
            op: "full_text",
          },
        ],
      }),
    ),
  ).toEqual([String(accent.record_id)]);
  expect(
    rowIDs(
      await queryViewRows(page, incidentId, timelineViewSchemaId, {
        filters: [
          {
            arg: { value: "rou" },
            field_key: "timeline.capture_state",
            op: "prefix",
          },
        ],
      }),
    ).sort(),
  ).toEqual(
    [
      String(timelinePrefix.record_id),
      String(timelinePrefixPeer.record_id),
    ].sort(),
  );
  expect(
    rowIDs(
      await queryViewRows(page, incidentId, timelineViewSchemaId, {
        filters: [
          {
            arg: { value: "ough" },
            field_key: "timeline.capture_state",
            op: "prefix",
          },
        ],
      }),
    ),
  ).toEqual([]);

  const emptyResponse = await page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/views/${notesViewSchemaId}/query`,
    {
      data: {
        filters: [
          {
            arg: { query: " -- " },
            field_key: "note.full_text",
            op: "full_text",
          },
        ],
      },
    },
  );
  expect(emptyResponse.status()).toBe(400);
  const emptyBody = (await emptyResponse.json()) as {
    error: { details?: { reason_code?: string } };
  };
  expect(emptyBody.error.details?.reason_code).toBe(
    "empty_full_text_after_tokenization",
  );
});

function waitForViewQuery(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
) {
  return page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${viewSchemaId}/query`,
        ),
  );
}

function readPostBody(request: { postData: () => string | null }) {
  return JSON.parse(request.postData() ?? "{}") as Record<string, unknown>;
}

async function queryViewRows(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  body: Record<string, unknown>,
) {
  const response = await page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/views/${viewSchemaId}/query`,
    {
      data: body,
    },
  );
  expect(response.ok()).toBeTruthy();
  return (
    (await response.json()) as {
      data: { rows: Array<Record<string, unknown>> };
    }
  ).data.rows;
}

function rowIDs(rows: Array<Record<string, unknown>>) {
  return rows.map((row) => String(row.record_id));
}
