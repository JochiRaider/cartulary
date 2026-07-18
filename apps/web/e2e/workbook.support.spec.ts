import {
  applyFilterChip,
  assertActiveFilterChipVisible,
  assertGroupRowPresentationOnly,
  changeGrouping,
  collapseGridGroup,
  expandGridGroup,
  removeFilterChip,
  scrollGridCellIntoView,
  sortByHeader,
} from "@cartulary/test-utils/grid";
import {
  gridGroupRowTestId,
  rowCellTestId,
  timelineRowMarkReviewedButtonTestId,
  workbookShellReadyTestId,
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
import { createViewRow } from "./support/workbook/query";
import { clickTimelineRowAction } from "./support/workbook/rowMutations";
import {
  readSavedViewSelectionState,
  selectSavedView,
  selectSavedViewScope,
  setCurrentSavedViewAsDefaultAndWait,
  setCurrentSavedViewAsHomeAndWait,
  setSavedViewDraftName,
  updateSavedViewFromCurrentSurface,
} from "./support/workbook/savedViews";

const timelineViewSchemaId = "cartulary.view.timeline.v2";

test("Verify browser command helpers for sort, filter, group, active chips, layout persistence, group expand-collapse, and startup/default surface UI.", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("S8B01"),
    "Workbook query support browser.saved-view-query.row-01 helper coverage",
  );
  const alpha = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("s8b01-alpha"),
    "timeline.activity_synopsis_text": "Alpha support FE-B",
  });
  const beta = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("s8b01-beta"),
    "timeline.activity_synopsis_text": "Beta support FE-B",
  });
  const savedView = await createSavedView(page, incidentId, {
    display_name: "FE-B support source",
    scope: "private",
    view_schema_id: timelineViewSchemaId,
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
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
  ).toHaveText("Alpha support FE-B");

  await clickTimelineRowAction(
    page,
    beta.record_id,
    timelineRowMarkReviewedButtonTestId(beta.record_id),
  );
  await scrollGridCellIntoView({
    cellKey: "timeline.capture_state",
    page,
    recordId: beta.record_id,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(rowCellTestId(beta.record_id, "timeline.capture_state")),
  ).toHaveText("reviewed");

  const selectRequest = waitForTimelineQuery(page, incidentId);
  await selectSavedView(page, timelineViewSchemaId, savedView.saved_view_id);
  expect(Object.keys(readPostBody(await selectRequest)).sort()).toEqual([]);
  await expect
    .poll(() => readSavedViewSelectionState(page, timelineViewSchemaId))
    .toMatchObject({
      activeViewSchemaId: timelineViewSchemaId,
      selectedSavedViewId: savedView.saved_view_id,
      selectedSheetRefKind: "saved_view",
    });

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
  await assertActiveFilterChipVisible(
    page,
    timelineViewSchemaId,
    "timeline.capture_state",
  );

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
  await assertGroupRowPresentationOnly({
    groupTestId: reviewedGroupTestId,
    page,
    surface: timelineViewSchemaId,
  });
  await collapseGridGroup({
    groupTestId: reviewedGroupTestId,
    page,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(beta.record_id, "timeline.activity_synopsis_text"),
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
    recordId: beta.record_id,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(beta.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toBeVisible();

  await setSavedViewDraftName(
    page,
    timelineViewSchemaId,
    "FE-B support persisted",
  );
  await selectSavedViewScope(page, timelineViewSchemaId, "shared");
  const patchRequest = page.waitForRequest(
    (request) =>
      request.method() === "PATCH" &&
      request
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/saved-views/${savedView.saved_view_id}`,
        ),
  );
  await updateSavedViewFromCurrentSurface(
    page,
    timelineViewSchemaId,
    savedView.saved_view_id,
  );
  expect(readPostBody(await patchRequest)).toMatchObject({
    base_saved_view_version: savedView.saved_view_version,
    display_name: "FE-B support persisted",
    layout_json: {
      layout_schema_id: "cartulary.layout.v1",
    },
    query_json: {
      filters: [
        {
          arg: { value: "reviewed" },
          field_key: "timeline.capture_state",
          op: "eq",
        },
      ],
      group_by: "timeline.capture_state",
      sort: [
        { direction: "asc", field_key: "timeline.activity_synopsis_text" },
      ],
    },
    scope: "shared",
  });

  const savedViewRef = {
    kind: "saved_view",
    id: savedView.saved_view_id,
  } as const;
  const homeAction = await setCurrentSavedViewAsHomeAndWait(
    page,
    timelineViewSchemaId,
    {
      expectedSheetRef: savedViewRef,
      incidentId,
    },
  );
  expect(homeAction.requestBody).toEqual({
    home_sheet_ref: {
      kind: "saved_view",
      id: savedView.saved_view_id,
    },
  });

  const defaultAction = await setCurrentSavedViewAsDefaultAndWait(
    page,
    timelineViewSchemaId,
    {
      expectedSheetRef: savedViewRef,
      incidentId,
    },
  );
  expect(defaultAction.requestBody).toEqual({
    default_sheet_ref: {
      kind: "saved_view",
      id: savedView.saved_view_id,
    },
  });

  const removeFilterRequest = waitForTimelineQuery(page, incidentId);
  await removeFilterChip(page, timelineViewSchemaId, "timeline.capture_state");
  expect(readPostBody(await removeFilterRequest)).toMatchObject({
    group_by: "timeline.capture_state",
  });
});

type SavedViewResource = {
  display_name: string;
  saved_view_id: string;
  saved_view_version?: number;
  scope?: string;
  view_schema_id: string;
};

async function createSavedView(
  page: Page,
  incidentId: string,
  data: {
    display_name: string;
    layout_json?: Record<string, unknown>;
    query_json?: Record<string, unknown>;
    scope?: "private" | "shared";
    view_schema_id: string;
  },
) {
  const response = await page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/saved-views`,
    {
      headers: await csrfHeaders(page),
      data: {
        display_name: data.display_name,
        layout_json: data.layout_json ?? {},
        query_json: data.query_json ?? {},
        ...(data.scope === undefined ? {} : { scope: data.scope }),
        view_schema_id: data.view_schema_id,
      },
    },
  );
  expect(response.status()).toBe(201);
  return ((await response.json()) as { data: SavedViewResource }).data;
}

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
  const raw = request.postData();
  if (raw === null || raw.trim() === "") {
    return {};
  }
  return JSON.parse(raw) as Record<string, unknown>;
}
