import {
  applyFilterChip,
  changeGrouping,
  collapseGridGroup,
  expandGridGroup,
  pasteGridMatrix,
  removeFilterChip,
  scrollGridCellIntoView,
  sortByHeader,
} from "@cartulary/test-utils/grid";
import {
  gridGroupRowTestId,
  rowCellTestId,
  saveStateTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowVersionTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { holdBrowserRequest as holdBrowserApiRequest } from "./support/transport/requestInterception";
import {
  assertRecordFieldMutationAnchor,
  fillDownGridCells,
} from "./support/workbook/mutationAnchors";
import { createViewRow } from "./support/workbook/query";
import { clickTimelineRowAction } from "./support/workbook/rowMutations";

const timelineViewSchemaId = "cartulary.view.timeline.v2";

test("Verify one-click focused editing, reviewed-edit demotion, sort, filter, group, paste, exact-range fill-down, scroll-to-cell, group expand/collapse, and anchor assertions through browser command helpers.", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("S301"),
    "Timeline support query controls",
  );
  const alphaRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("s301-alpha"),
    "timeline.activity_synopsis_text": "Alpha summary",
  });
  const betaRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("s301-beta"),
    "timeline.activity_synopsis_text": "Beta summary",
  });
  const clickEditRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("s301-click-edit"),
      "timeline.device_object_text": "Host Alpha",
    },
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: alphaRow.record_id,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(alphaRow.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("Alpha summary");

  await scrollGridCellIntoView({
    cellKey: "timeline.device_object_text",
    page,
    recordId: clickEditRow.record_id,
    surface: timelineViewSchemaId,
  });
  const deviceCell = page.getByTestId(
    rowCellTestId(clickEditRow.record_id, "timeline.device_object_text"),
  );
  await expect(deviceCell).toHaveText("Host Alpha");
  const clickEditPatches: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/api/v1/records/${clickEditRow.record_id}`)
    ) {
      clickEditPatches.push(request.postData() ?? "");
    }
  });
  await deviceCell.click();
  const deviceEditor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.device_object_text",
      recordId: clickEditRow.record_id,
      surface: "grid",
    }),
  );
  await expect(deviceEditor).toBeFocused();
  await expect(deviceEditor).toHaveValue("Host Alpha");
  expect(
    await deviceEditor.evaluate((element) => ({
      end: (element as HTMLInputElement).selectionEnd,
      start: (element as HTMLInputElement).selectionStart,
    })),
  ).toEqual({ end: 10, start: 10 });
  await page.keyboard.insertText(" edited");
  await expect(deviceEditor).toHaveValue("Host Alpha edited");
  await deviceEditor.press("Enter");
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  expect(clickEditPatches).toHaveLength(1);
  expect(JSON.parse(clickEditPatches[0] ?? "{}")).toMatchObject({
    base_row_version: clickEditRow.row_version,
    changes: [
      {
        field_key: "timeline.device_object_text",
        value: "Host Alpha edited",
      },
    ],
  });

  await clickTimelineRowAction(
    page,
    betaRow.record_id,
    timelineRowMarkReviewedButtonTestId(betaRow.record_id),
  );
  await scrollGridCellIntoView({
    cellKey: "timeline.capture_state",
    page,
    recordId: betaRow.record_id,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(betaRow.record_id, "timeline.capture_state"),
    ),
  ).toHaveText("reviewed");

  const sortRequest = waitForTimelineQuery(page, incidentId);
  await sortByHeader(
    page,
    timelineViewSchemaId,
    "timeline.activity_synopsis_text",
  );
  expect(readPostBody(await sortRequest)).toEqual({
    sort: [{ direction: "asc", field_key: "timeline.activity_synopsis_text" }],
  });

  const filterRequest = waitForTimelineQuery(page, incidentId);
  await applyFilterChip(
    page,
    timelineViewSchemaId,
    "timeline.capture_state",
    "reviewed",
  );
  expect(readPostBody(await filterRequest)).toEqual({
    filters: [
      {
        arg: { value: "reviewed" },
        field_key: "timeline.capture_state",
        op: "eq",
      },
    ],
    sort: [{ direction: "asc", field_key: "timeline.activity_synopsis_text" }],
  });
  await expect(
    page.getByTestId(
      rowCellTestId(betaRow.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("Beta summary");

  const groupRequest = waitForTimelineQuery(page, incidentId);
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
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
      { direction: "asc", field_key: "timeline.activity_synopsis_text" },
    ],
  });
  const reviewedGroupTestId = gridGroupRowTestId(
    timelineViewSchemaId,
    "timeline.capture_state",
    "reviewed",
  );
  await expect(page.getByTestId(reviewedGroupTestId)).toBeVisible();
  await collapseGridGroup({
    groupTestId: reviewedGroupTestId,
    page,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(betaRow.record_id, "timeline.activity_synopsis_text"),
    ),
  ).not.toBeVisible();
  await expandGridGroup({
    groupTestId: reviewedGroupTestId,
    page,
    surface: timelineViewSchemaId,
  });
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: betaRow.record_id,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(betaRow.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toBeVisible();

  const bulkMutationRequest = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/bulk-mutations`,
        ),
  );
  const bulkMutationResponse = await fillDownGridCells({
    apiBase,
    csrfHeaders: await csrfHeaders(page),
    fieldKey: "timeline.raw_activity_text",
    incidentId,
    page,
    surface: timelineViewSchemaId,
    targetRecords: [
      {
        baseRowVersion: alphaRow.row_version,
        recordId: alphaRow.record_id,
      },
    ],
    value: "Filled details through helper",
  });
  expect(bulkMutationResponse.ok()).toBeTruthy();
  expect(readPostBody(await bulkMutationRequest)).toMatchObject({
    view_schema_id: timelineViewSchemaId,
    kind: "fill_down_v1",
    field_key: "timeline.raw_activity_text",
    value: "Filled details through helper",
    targets: [
      {
        base_row_version: alphaRow.row_version,
        record_id: alphaRow.record_id,
      },
    ],
  });

  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: betaRow.record_id,
    surface: timelineViewSchemaId,
  });
  const summaryPatch = page.waitForRequest(
    (request) =>
      request.method() === "PATCH" &&
      request.url().endsWith(`/api/v1/records/${betaRow.record_id}`),
  );
  const summaryPatchResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/v1/records/${betaRow.record_id}`),
  );
  const postEditQuery = waitForTimelineQuery(page, incidentId);
  const betaSummary = page.getByTestId(
    rowCellTestId(betaRow.record_id, "timeline.activity_synopsis_text"),
  );
  await betaSummary.click();
  const betaSummaryEditor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.activity_synopsis_text",
      recordId: betaRow.record_id,
      surface: "grid",
    }),
  );
  await betaSummaryEditor.fill("Beta summary anchored");
  await betaSummaryEditor.press("Enter");
  assertRecordFieldMutationAnchor({
    actualRecordId: betaRow.record_id,
    body: readPostBody(await summaryPatch),
    expectedRecordId: betaRow.record_id,
    expectedValue: "Beta summary anchored",
    fieldKey: "timeline.activity_synopsis_text",
  });
  const summaryPatchPayload = (await (await summaryPatchResponse).json()) as {
    data: {
      row: {
        cells: Record<string, { value: unknown }>;
      };
    };
  };
  expect(
    summaryPatchPayload.data.row.cells["timeline.capture_state"]?.value,
  ).toBe("enriched");
  expect(readPostBody(await postEditQuery)).toEqual({
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
      { direction: "asc", field_key: "timeline.activity_synopsis_text" },
    ],
  });
  await expect(
    page.getByText("No rows match the current filters."),
  ).toBeVisible();

  const removeReviewedFilterRequest = waitForTimelineQuery(page, incidentId);
  await removeFilterChip(page, timelineViewSchemaId, "timeline.capture_state");
  expect(readPostBody(await removeReviewedFilterRequest)).toEqual({
    group_by: "timeline.capture_state",
    sort: [
      { direction: "asc", field_key: "timeline.capture_state" },
      { direction: "asc", field_key: "timeline.activity_synopsis_text" },
    ],
  });
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: betaRow.record_id,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(betaRow.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("Beta summary anchored");

  const pasteRequest = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
        ),
  );
  const pasteResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
        ),
  );
  await pasteGridMatrix({
    fieldKey: "timeline.activity_synopsis_text",
    matrix: [["Beta pasted through helper", "host-token"]],
    page,
    recordId: betaRow.record_id,
    surface: timelineViewSchemaId,
  });
  expect(readPostBody(await pasteRequest)).toMatchObject({
    view_schema_id: timelineViewSchemaId,
    clipboard_text: "Beta pasted through helper\thost-token",
    format: "tsv",
    start_field_key: "timeline.activity_synopsis_text",
    columns: ["timeline.activity_synopsis_text", "timeline.data_source_text"],
    targets: [
      {
        kind: "record",
        record_id: betaRow.record_id,
      },
    ],
  });
  expect((await pasteResponse).ok()).toBeTruthy();
});

test("support keeps a pending edit anchored to its record under sort, filter, group, and live invalidation", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("S302"),
    "Timeline support anchor stability",
  );
  const alphaRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("s302-alpha"),
    "timeline.activity_synopsis_text": "Alpha summary",
  });
  const betaRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("s302-beta"),
    "timeline.activity_synopsis_text": "Beta summary",
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await clickTimelineRowAction(
    page,
    betaRow.record_id,
    timelineRowMarkReviewedButtonTestId(betaRow.record_id),
  );
  await expect(
    page.getByTestId(
      rowCellTestId(betaRow.record_id, "timeline.capture_state"),
    ),
  ).toHaveText("reviewed");

  await sortByHeader(
    page,
    timelineViewSchemaId,
    "timeline.activity_synopsis_text",
  );
  await applyFilterChip(
    page,
    timelineViewSchemaId,
    "timeline.has_evidence",
    "false",
  );
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");

  const heldPatch = await holdBrowserApiRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${alphaRow.record_id}`,
  });

  try {
    const alphaSummary = page.getByTestId(
      rowCellTestId(alphaRow.record_id, "timeline.activity_synopsis_text"),
    );
    await scrollGridCellIntoView({
      cellKey: "timeline.activity_synopsis_text",
      page,
      recordId: alphaRow.record_id,
      surface: timelineViewSchemaId,
    });
    await alphaSummary.click();
    const alphaSummaryEditor = page.getByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.activity_synopsis_text",
        recordId: alphaRow.record_id,
        surface: "grid",
      }),
    );
    await alphaSummaryEditor.fill("Zulu anchored");
    await alphaSummaryEditor.press("Enter");
    await heldPatch.waitForHit;
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");

    const betaVersion = Number.parseInt(
      (await page
        .getByTestId(timelineRowVersionTestId(betaRow.record_id))
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
              field_key: "timeline.raw_activity_text",
              value: "Support invalidation",
            },
          ],
        },
      },
    );
    expect(invalidationResponse.ok()).toBeTruthy();

    expect(heldPatch.hitCount()).toBe(1);
    heldPatch.release();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  } finally {
    await heldPatch.dispose();
  }
  await expect(
    page.getByTestId(timelineRowVersionTestId(alphaRow.record_id)),
  ).toHaveText("2");
  await expect(
    page.getByTestId(
      rowCellTestId(alphaRow.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("Zulu anchored");
  const alphaCaptureState = (
    (await page
      .getByTestId(rowCellTestId(alphaRow.record_id, "timeline.capture_state"))
      .textContent()) ?? ""
  ).trim();
  expect(alphaCaptureState).not.toBe("");
  await expect(
    page.getByTestId(
      gridGroupRowTestId(
        timelineViewSchemaId,
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
