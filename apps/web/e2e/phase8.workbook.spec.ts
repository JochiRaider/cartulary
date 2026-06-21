import {
  applyFilterChip,
  assertActiveFilterChipVisible,
  assertGroupRowPresentationOnly,
  changeGrouping,
  collapseGridGroup,
  createSavedViewFromCurrentSurface,
  deleteSavedViewFromCurrentSurface,
  duplicateSavedViewFromCurrentSurface,
  expandGridGroup,
  openSavedViewActionMenu,
  readSavedViewSelectionState,
  removeFilterChip,
  selectSavedView,
  selectSavedViewScope,
  setCurrentSavedViewAsDefaultAndWait,
  setCurrentSavedViewAsHomeAndWait,
  setSavedViewDraftName,
  sortByHeader,
  updateSavedViewFromCurrentSurface,
} from "@cartulary/test-utils";
import {
  gridGroupRowsSelector,
  gridGroupRowTestId,
  gridShellTestId,
  rowCellTestId,
  savedViewDeleteButtonTestId,
  savedViewOptionTestId,
  savedViewSelectorTestId,
  savedViewUpdateButtonTestId,
  surfaceTabTestId,
  timelineRowMarkReviewedButtonTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  apiBase,
  createIncident,
  createIncidentMemberUser,
  createViewRow,
  csrfHeaders,
  loginLocalSession,
  seedSystemSavedView,
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import { clickTimelineRowAction } from "./phase4Helpers";

const evidenceViewSchemaId = "cartulary.view.evidence.v1";
const hostsViewSchemaId = "cartulary.view.hosts.v1";
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
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();

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

  await expect(
    page.getByRole("button", {
      name: /duplicate saved view|delete saved view/i,
    }),
  ).toHaveCount(0);
});

test("FE-I-P8-01 Verify saved-view create/update/select/default UI uses active surface scope and public saved-view/workbook-preference contracts.", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEIP801"),
    "Phase 8 FE-I-P8-01 saved-view UI integration",
  );
  const timelineRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("feip801-row"),
      "timeline.summary": "FE-I-P8 saved view row",
    },
  );
  const privateSavedView = await createSavedView(page, incidentId, {
    display_name: "FE-I private timeline",
    scope: "private",
    view_schema_id: timelineViewSchemaId,
    query_json: {
      filters: [
        {
          field_key: "timeline.capture_state",
          op: "eq",
          arg: { value: "rough" },
        },
      ],
      group_by: "timeline.capture_state",
      sort: [{ field_key: "timeline.summary", direction: "asc" }],
    },
    layout_json: {},
  });
  const systemSavedView = await seedSystemSavedView(page, incidentId, {
    display_name: "FE-I system timeline",
    view_schema_id: timelineViewSchemaId,
    query_json: {
      filters: [
        {
          field_key: "timeline.tags",
          op: "contains_any",
          arg: { values: ["alpha", "zeta"] },
        },
      ],
      sort: [],
    },
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  const selector = page.getByTestId(
    savedViewSelectorTestId(timelineViewSchemaId),
  );
  await expect(
    selector.getByTestId(
      savedViewOptionTestId(
        timelineViewSchemaId,
        privateSavedView.saved_view_id,
      ),
    ),
  ).toHaveCount(1);

  const selectRequest = waitForViewQuery(
    page,
    incidentId,
    timelineViewSchemaId,
  );
  await selector.selectOption(privateSavedView.saved_view_id);
  expect(readPostBody(await selectRequest)).toEqual({
    filters: [
      {
        arg: { value: "rough" },
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
  await expect(page).toHaveURL(/sheet_ref_kind=saved_view/);
  await expect(page).toHaveURL(
    new RegExp(`sheet_ref_id=${privateSavedView.saved_view_id}`),
  );

  await setSavedViewDraftName(
    page,
    timelineViewSchemaId,
    "FE-I updated shared",
  );
  await selectSavedViewScope(page, timelineViewSchemaId, "shared");
  const patchRequest = page.waitForRequest(
    (request) =>
      request.method() === "PATCH" &&
      request
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/saved-views/${privateSavedView.saved_view_id}`,
        ),
  );
  await updateSavedViewFromCurrentSurface(
    page,
    timelineViewSchemaId,
    privateSavedView.saved_view_id,
  );
  const patchBody = readPostBody(await patchRequest);
  expect(Object.keys(patchBody).sort()).toEqual([
    "base_saved_view_version",
    "display_name",
    "layout_json",
    "query_json",
    "scope",
  ]);
  expect(patchBody).toMatchObject({
    base_saved_view_version: privateSavedView.saved_view_version,
    display_name: "FE-I updated shared",
    scope: "shared",
    query_json: {
      filters: [
        {
          arg: { value: "rough" },
          field_key: "timeline.capture_state",
          op: "eq",
        },
      ],
      group_by: "timeline.capture_state",
      sort: [{ direction: "asc", field_key: "timeline.summary" }],
    },
  });
  expect(patchBody).not.toHaveProperty("saved_view_id");
  expect(patchBody).not.toHaveProperty("view_schema_id");
  expect(patchBody).not.toHaveProperty("owner_user_id");

  const privateSavedViewRef = {
    kind: "saved_view",
    id: privateSavedView.saved_view_id,
  } as const;
  const homeAction = await setCurrentSavedViewAsHomeAndWait(
    page,
    timelineViewSchemaId,
    {
      expectedSheetRef: privateSavedViewRef,
      incidentId,
    },
  );
  expect(homeAction.requestBody).toEqual({
    home_sheet_ref: {
      kind: "saved_view",
      id: privateSavedView.saved_view_id,
    },
  });

  const defaultAction = await setCurrentSavedViewAsDefaultAndWait(
    page,
    timelineViewSchemaId,
    {
      expectedSheetRef: privateSavedViewRef,
      incidentId,
    },
  );
  expect(defaultAction.requestBody).toEqual({
    default_sheet_ref: {
      kind: "saved_view",
      id: privateSavedView.saved_view_id,
    },
  });

  await selector.selectOption(systemSavedView.saved_view_id);
  await openSavedViewActionMenu(page, timelineViewSchemaId);
  await expect(
    page.getByTestId(
      savedViewUpdateButtonTestId(
        timelineViewSchemaId,
        systemSavedView.saved_view_id,
      ),
    ),
  ).toBeDisabled();
  await expect(
    page.getByTestId(
      savedViewDeleteButtonTestId(
        timelineViewSchemaId,
        systemSavedView.saved_view_id,
      ),
    ),
  ).toBeDisabled();
  const duplicateRequest = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request.url().endsWith(`/api/v1/incidents/${incidentId}/saved-views`),
  );
  const duplicateResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/api/v1/incidents/${incidentId}/saved-views`),
  );
  await duplicateSavedViewFromCurrentSurface(
    page,
    timelineViewSchemaId,
    systemSavedView.saved_view_id,
  );
  expect(readPostBody(await duplicateRequest)).toMatchObject({
    display_name: "FE-I system timeline Copy",
    scope: "private",
    view_schema_id: timelineViewSchemaId,
    query_json: {
      filters: [
        {
          arg: { values: ["alpha", "zeta"] },
          field_key: "timeline.tags",
          op: "contains_any",
        },
      ],
      sort: [],
    },
  });
  const duplicateSavedView = (
    (await (await duplicateResponse).json()) as {
      data: SavedViewResource;
    }
  ).data;

  const deleteRequest = page.waitForRequest(
    (request) =>
      request.method() === "DELETE" &&
      request
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/saved-views/${duplicateSavedView.saved_view_id}`,
        ),
  );
  await deleteSavedViewFromCurrentSurface(
    page,
    timelineViewSchemaId,
    duplicateSavedView.saved_view_id,
  );
  await deleteRequest;
  await expect(page).toHaveURL(/view_schema_id=cartulary\.view\.timeline\.v1/);
  expect(
    rowIDs(await queryViewRows(page, incidentId, timelineViewSchemaId, {})),
  ).toContain(String(timelineRow.record_id));

  await setSavedViewDraftName(
    page,
    timelineViewSchemaId,
    "FE-I created current",
  );
  const createRequest = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request.url().endsWith(`/api/v1/incidents/${incidentId}/saved-views`),
  );
  await createSavedViewFromCurrentSurface(page, timelineViewSchemaId);
  expect(readPostBody(await createRequest)).toMatchObject({
    display_name: "FE-I created current",
    scope: "private",
    view_schema_id: timelineViewSchemaId,
  });
});

test("FE-B-P8-01 Verify browser command helpers for sort, filter, group, active chips, layout persistence, group expand-collapse, and startup/default surface UI.", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEBP801"),
    "Phase 8 FE-B-P8-01 browser helper coverage",
  );
  const alpha = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("febp801-alpha"),
    "timeline.summary": "Alpha FE-B",
  });
  const beta = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("febp801-beta"),
    "timeline.summary": "Beta FE-B",
  });
  const gamma = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("febp801-gamma"),
    "timeline.summary": "Gamma FE-B",
  });
  const savedView = await createSavedView(page, incidentId, {
    display_name: "FE-B helper source",
    scope: "private",
    view_schema_id: timelineViewSchemaId,
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(
    page.getByTestId(rowCellTestId(alpha.record_id, "timeline.summary")),
  ).toHaveValue("Alpha FE-B");

  await clickTimelineRowAction(
    page,
    beta.record_id,
    timelineRowMarkReviewedButtonTestId(beta.record_id),
  );
  await expect(
    page.getByTestId(rowCellTestId(beta.record_id, "timeline.capture_state")),
  ).toHaveText("reviewed");

  const selectRequest = waitForViewQuery(
    page,
    incidentId,
    timelineViewSchemaId,
  );
  await selectSavedView(page, timelineViewSchemaId, savedView.saved_view_id);
  expect(Object.keys(readPostBody(await selectRequest)).sort()).toEqual([]);
  await expect
    .poll(() => readSavedViewSelectionState(page, timelineViewSchemaId))
    .toEqual({
      activeViewSchemaId: timelineViewSchemaId,
      selectedSavedViewId: savedView.saved_view_id,
      selectedSheetRefKind: "saved_view",
    });

  const sortRequest = waitForViewQuery(page, incidentId, timelineViewSchemaId);
  await sortByHeader(page, timelineViewSchemaId, "timeline.summary");
  expect(readPostBody(await sortRequest)).toEqual({
    sort: [{ direction: "asc", field_key: "timeline.summary" }],
  });
  await expectFirstDataRow(page, String(alpha.record_id));
  expect(await visibleRecordIds(page)).toEqual([
    String(alpha.record_id),
    String(beta.record_id),
    String(gamma.record_id),
  ]);

  const filterRequest = waitForViewQuery(
    page,
    incidentId,
    timelineViewSchemaId,
  );
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
    sort: [{ direction: "asc", field_key: "timeline.summary" }],
  });
  await assertActiveFilterChipVisible(
    page,
    timelineViewSchemaId,
    "timeline.capture_state",
  );
  await expectFirstDataRow(page, String(beta.record_id));

  const groupRequest = waitForViewQuery(page, incidentId, timelineViewSchemaId);
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
      { direction: "asc", field_key: "timeline.summary" },
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
    page.getByTestId(rowCellTestId(beta.record_id, "timeline.summary")),
  ).not.toBeVisible();
  await expandGridGroup({
    groupTestId: reviewedGroupTestId,
    page,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(rowCellTestId(beta.record_id, "timeline.summary")),
  ).toBeVisible();

  await setSavedViewDraftName(
    page,
    timelineViewSchemaId,
    "FE-B helper persisted",
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
  const patchBody = readPostBody(await patchRequest);
  expect(patchBody).toMatchObject({
    base_saved_view_version: savedView.saved_view_version,
    display_name: "FE-B helper persisted",
    scope: "shared",
    query_json: {
      filters: [
        {
          arg: { value: "reviewed" },
          field_key: "timeline.capture_state",
          op: "eq",
        },
      ],
      group_by: "timeline.capture_state",
      sort: [{ direction: "asc", field_key: "timeline.summary" }],
    },
    layout_json: {
      layout_schema_id: "cartulary.layout.v1",
    },
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

  const removeFilterRequest = waitForViewQuery(
    page,
    incidentId,
    timelineViewSchemaId,
  );
  await removeFilterChip(page, timelineViewSchemaId, "timeline.capture_state");
  expect(readPostBody(await removeFilterRequest)).toEqual({
    group_by: "timeline.capture_state",
    sort: [
      { direction: "asc", field_key: "timeline.capture_state" },
      { direction: "asc", field_key: "timeline.summary" },
    ],
  });
});

test("FE-E-P8-01 Verify saved-view persistence, default/startup surface persistence, and query replay through /api/v1/ after reload.", async ({
  page,
}) => {
  await verifySavedViewPersistenceReplay(
    page,
    "FEEP801",
    "Phase 8 FE-E-P8-01 persisted query replay",
  );
});

test("E-8-05 verifies saved-view persistence, default/startup surface persistence, and query replay through /api/v1/ after reload", async ({
  page,
}) => {
  await verifySavedViewPersistenceReplay(
    page,
    "E805",
    "Phase 8 E-8-05 persisted query replay",
  );
});

async function verifySavedViewPersistenceReplay(
  page: Page,
  incidentKeyPrefix: string,
  incidentTitle: string,
) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey(incidentKeyPrefix),
    incidentTitle,
  );
  const alpha = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("feep801-alpha"),
    "timeline.summary": "Alpha FE-E",
  });
  const beta = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("feep801-beta"),
    "timeline.summary": "Beta FE-E",
  });
  await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("feep801-gamma"),
    "timeline.summary": "Gamma FE-E",
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(
    page.getByTestId(rowCellTestId(alpha.record_id, "timeline.summary")),
  ).toHaveValue("Alpha FE-E");

  await clickTimelineRowAction(
    page,
    beta.record_id,
    timelineRowMarkReviewedButtonTestId(beta.record_id),
  );
  await expect(
    page.getByTestId(rowCellTestId(beta.record_id, "timeline.capture_state")),
  ).toHaveText("reviewed");

  const sortRequest = waitForViewQuery(page, incidentId, timelineViewSchemaId);
  await sortByHeader(page, timelineViewSchemaId, "timeline.summary");
  expect(readPostBody(await sortRequest)).toEqual({
    sort: [{ direction: "asc", field_key: "timeline.summary" }],
  });

  const filterRequest = waitForViewQuery(
    page,
    incidentId,
    timelineViewSchemaId,
  );
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
    sort: [{ direction: "asc", field_key: "timeline.summary" }],
  });

  const groupRequest = waitForViewQuery(page, incidentId, timelineViewSchemaId);
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
  const replayedQuery = {
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
  };
  expect(readPostBody(await groupRequest)).toEqual(replayedQuery);
  await expectFirstDataRow(page, String(beta.record_id));

  await setSavedViewDraftName(
    page,
    timelineViewSchemaId,
    "FE-E persisted replay",
  );
  const createRequest = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request.url().endsWith(`/api/v1/incidents/${incidentId}/saved-views`),
  );
  const createResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/api/v1/incidents/${incidentId}/saved-views`),
  );
  await createSavedViewFromCurrentSurface(page, timelineViewSchemaId);
  const createBody = readPostBody(await createRequest);
  expect(createBody).toMatchObject({
    display_name: "FE-E persisted replay",
    scope: "private",
    view_schema_id: timelineViewSchemaId,
    query_json: {
      filters: [
        {
          arg: { value: "reviewed" },
          field_key: "timeline.capture_state",
          op: "eq",
        },
      ],
      group_by: "timeline.capture_state",
      sort: [{ direction: "asc", field_key: "timeline.summary" }],
    },
    layout_json: {
      layout_schema_id: "cartulary.layout.v1",
    },
  });
  const savedView = (
    (await (await createResponse).json()) as {
      data: SavedViewResource;
    }
  ).data;
  await expect
    .poll(() => readSavedViewSelectionState(page, timelineViewSchemaId))
    .toMatchObject({
      activeViewSchemaId: timelineViewSchemaId,
      selectedSavedViewId: savedView.saved_view_id,
      selectedSheetRefKind: "saved_view",
    });

  const savedViewsResponse = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}/saved-views`,
  );
  expect(savedViewsResponse.ok()).toBeTruthy();
  const savedViews = (
    (await savedViewsResponse.json()) as {
      data: { saved_views: SavedViewResource[] };
    }
  ).data.saved_views;
  const persistedSavedView = savedViews.find(
    (candidate) => candidate.saved_view_id === savedView.saved_view_id,
  );
  expect(persistedSavedView).toMatchObject({
    display_name: "FE-E persisted replay",
    view_schema_id: timelineViewSchemaId,
    query_json: {
      filters: [
        {
          arg: { value: "reviewed" },
          field_key: "timeline.capture_state",
          op: "eq",
        },
      ],
      group_by: "timeline.capture_state",
      sort: [{ direction: "asc", field_key: "timeline.summary" }],
    },
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
  expect(
    (await getUserWorkbookPreferences(page, incidentId)).home_sheet_ref,
  ).toEqual({
    kind: "saved_view",
    id: savedView.saved_view_id,
  });
  expect(await getWorkbookStartup(page, incidentId)).toMatchObject({
    selected_sheet_ref: {
      kind: "saved_view",
      id: savedView.saved_view_id,
    },
    selected_view_schema_id: timelineViewSchemaId,
    source: "home",
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
  expect(
    (await getDefaultWorkbookPreferences(page, incidentId)).default_sheet_ref,
  ).toEqual({
    kind: "saved_view",
    id: savedView.saved_view_id,
  });

  await putUserHomeSheetRef(page, incidentId, null);
  expect(await getWorkbookStartup(page, incidentId)).toMatchObject({
    selected_sheet_ref: {
      kind: "saved_view",
      id: savedView.saved_view_id,
    },
    selected_view_schema_id: timelineViewSchemaId,
    source: "default",
  });

  const startupQueryRequest = waitForViewQueryBody(
    page,
    incidentId,
    timelineViewSchemaId,
    replayedQuery,
  );
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  expect(readPostBody(await startupQueryRequest)).toEqual(replayedQuery);
  await expect(page).toHaveURL(/sheet_ref_kind=saved_view/);
  await expect(page).toHaveURL(
    new RegExp(`sheet_ref_id=${savedView.saved_view_id}`),
  );
  expect(await visibleRecordIds(page)).toEqual([String(beta.record_id)]);
  await expect(
    page.getByTestId(
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "reviewed",
      ),
    ),
  ).toBeVisible();

  const reloadQueryRequest = waitForViewQueryBody(
    page,
    incidentId,
    timelineViewSchemaId,
    replayedQuery,
  );
  await page.reload();
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  expect(readPostBody(await reloadQueryRequest)).toEqual(replayedQuery);
  expect(
    rowIDs(
      await queryViewRows(
        page,
        incidentId,
        timelineViewSchemaId,
        replayedQuery,
      ),
    ),
  ).toEqual([String(beta.record_id)]);
}

test("E-8-02 workbook startup falls back to Timeline for an unsupported explicit sheet", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E802"),
    "Phase 8 E-8-02 startup fallback",
  );

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      hostsViewSchemaId,
    )}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(hostsViewSchemaId)),
  ).toBeVisible();
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(hostsViewSchemaId)}`),
  );

  const homeSavedView = await createSavedView(page, incidentId, {
    display_name: "Phase 8 home hosts",
    scope: "shared",
    view_schema_id: hostsViewSchemaId,
  });
  await putUserHomeSheetRef(page, incidentId, {
    kind: "saved_view",
    id: homeSavedView.saved_view_id,
  });
  const homeStartup = await getWorkbookStartup(page, incidentId);
  expect(homeStartup.source).toBe("home");
  expect(homeStartup.selected_view_schema_id).toBe(hostsViewSchemaId);
  expect(homeStartup.selected_sheet_ref).toEqual({
    kind: "saved_view",
    id: homeSavedView.saved_view_id,
  });
  const explicitSavedViewStartup = await getWorkbookStartup(
    page,
    incidentId,
    `?sheet_ref_kind=saved_view&sheet_ref_id=${homeSavedView.saved_view_id}`,
  );
  expect(explicitSavedViewStartup.source).toBe("explicit");
  expect(explicitSavedViewStartup.selected_view_schema_id).toBe(
    hostsViewSchemaId,
  );
  expect(explicitSavedViewStartup.selected_sheet_ref).toEqual({
    kind: "saved_view",
    id: homeSavedView.saved_view_id,
  });

  await putUserHomeSheetRef(page, incidentId, null);
  await putDefaultSheetRef(page, incidentId, {
    kind: "view_schema",
    id: evidenceViewSchemaId,
  });
  const defaultStartup = await getWorkbookStartup(page, incidentId);
  expect(defaultStartup.source).toBe("default");
  expect(defaultStartup.selected_view_schema_id).toBe(evidenceViewSchemaId);
  expect(defaultStartup.selected_sheet_ref).toEqual({
    kind: "view_schema",
    id: evidenceViewSchemaId,
  });

  const invalidExplicitStartup = await getWorkbookStartup(
    page,
    incidentId,
    "?view_schema_id=cartulary.view.unknown.v1",
  );
  expect(invalidExplicitStartup.source).toBe("default");
  expect(invalidExplicitStartup.selected_view_schema_id).toBe(
    evidenceViewSchemaId,
  );

  const deletedDefault = await createSavedView(page, incidentId, {
    display_name: "Phase 8 deleted default",
    view_schema_id: evidenceViewSchemaId,
  });
  await putDefaultSheetRef(page, incidentId, {
    kind: "saved_view",
    id: deletedDefault.saved_view_id,
  });
  await deleteSavedView(page, incidentId, deletedDefault.saved_view_id);
  const deletedDefaultStartup = await getWorkbookStartup(page, incidentId);
  expect(deletedDefaultStartup.source).toBe("timeline");
  expect(deletedDefaultStartup.selected_sheet_ref).toEqual({
    kind: "view_schema",
    id: timelineViewSchemaId,
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(
    page.getByTestId(surfaceTabTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(page).toHaveURL(/view_schema_id=cartulary\.view\.timeline\.v1/);
  expect(
    (await getDefaultWorkbookPreferences(page, incidentId)).default_sheet_ref,
  ).toBeNull();

  const hiddenSavedView = await createSavedView(page, incidentId, {
    display_name: "Phase 8 hidden home",
    view_schema_id: hostsViewSchemaId,
  });
  const viewerPassword = "Phase8E802Viewer!";
  const viewer = await createIncidentMemberUser(page, incidentId, {
    email: uniqueEmail("phase8-e802-viewer"),
    display_name: "Phase 8 E-8-02 Viewer",
    initial_password: viewerPassword,
    role: "viewer",
  });
  await loginLocalSession(page, viewer.email, viewerPassword);
  await putUserHomeSheetRef(page, incidentId, {
    kind: "saved_view",
    id: hiddenSavedView.saved_view_id,
  });
  const hiddenHomeStartup = await getWorkbookStartup(page, incidentId);
  expect(hiddenHomeStartup.source).toBe("timeline");
  expect(hiddenHomeStartup.selected_sheet_ref).toEqual({
    kind: "view_schema",
    id: timelineViewSchemaId,
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(page).toHaveURL(/view_schema_id=cartulary\.view\.timeline\.v1/);
  expect(
    (await getUserWorkbookPreferences(page, incidentId)).home_sheet_ref,
  ).toBeNull();
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
  const gamma = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e803-gamma"),
    "timeline.summary": "Gamma Phase 8",
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(
      rowCellTestId(alpha.record_id as string, "timeline.summary"),
    ),
  ).toHaveValue("Alpha Phase 8");

  await clickTimelineRowAction(
    page,
    beta.record_id as string,
    timelineRowMarkReviewedButtonTestId(beta.record_id as string),
  );
  await expect(
    page.getByTestId(rowCellTestId(beta.record_id, "timeline.capture_state")),
  ).toHaveText("reviewed");

  const sortRequest = waitForViewQuery(page, incidentId, timelineViewSchemaId);
  await sortByHeader(page, timelineViewSchemaId, "timeline.summary");
  expect(readPostBody(await sortRequest)).toEqual({
    sort: [{ direction: "asc", field_key: "timeline.summary" }],
  });
  await expectFirstDataRow(page, String(alpha.record_id));
  expect(await visibleRecordIds(page)).toEqual([
    String(alpha.record_id),
    String(beta.record_id),
    String(gamma.record_id),
  ]);

  const filterRequest = waitForViewQuery(
    page,
    incidentId,
    timelineViewSchemaId,
  );
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
    sort: [{ direction: "asc", field_key: "timeline.summary" }],
  });
  await expectFirstDataRow(page, String(beta.record_id));
  expect(await visibleRecordIds(page)).toEqual([String(beta.record_id)]);

  const removeFilterRequest = waitForViewQuery(
    page,
    incidentId,
    timelineViewSchemaId,
  );
  await removeFilterChip(page, timelineViewSchemaId, "timeline.capture_state");
  expect(readPostBody(await removeFilterRequest)).toEqual({
    sort: [{ direction: "asc", field_key: "timeline.summary" }],
  });

  const groupRequest = waitForViewQuery(page, incidentId, timelineViewSchemaId);
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
  expect(readPostBody(await groupRequest)).toEqual({
    group_by: "timeline.capture_state",
    sort: [
      { direction: "asc", field_key: "timeline.capture_state" },
      { direction: "asc", field_key: "timeline.summary" },
    ],
  });
  const timelineGrid = page.getByTestId(gridShellTestId(timelineViewSchemaId));
  await expect
    .poll(async () => visibleGroupLabels(page))
    .toEqual(["reviewed", "rough"]);
  await expect(
    timelineGrid.getByTestId(
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "reviewed",
      ),
    ),
  ).toHaveCount(1);
  await expect(
    timelineGrid.getByTestId(
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "rough",
      ),
    ),
  ).toHaveCount(1);
  const groupTestIds = await visibleGroupTestIds(page);
  expect(groupTestIds).toEqual([
    gridGroupRowTestId(
      timelineViewSchemaId,
      "timeline.capture_state",
      "reviewed",
    ),
    gridGroupRowTestId(timelineViewSchemaId, "timeline.capture_state", "rough"),
  ]);
  expect(new Set(groupTestIds).size).toBe(groupTestIds.length);
  expect(await visibleRecordIds(page)).toEqual([
    String(beta.record_id),
    String(alpha.record_id),
    String(gamma.record_id),
  ]);
  for (const state of ["reviewed", "rough"]) {
    const groupTestId = gridGroupRowTestId(
      timelineViewSchemaId,
      "timeline.capture_state",
      state,
    );
    const groupToggle = timelineGrid.getByTestId(groupTestId);
    const groupRow = timelineGrid
      .getByTestId(groupTestId)
      .locator("xpath=ancestor::*[@role='row'][1]");
    await expect(groupRow).toHaveCount(1);
    await expect(groupRow).not.toHaveAttribute("data-grid-record-id", /.+/);
    await expect(groupToggle).toHaveAttribute("aria-expanded", "true");
    await expect(groupRow.locator("input, textarea, select")).toHaveCount(0);
    await expect(groupRow.locator("button")).toHaveCount(1);
  }
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
  const bravo = await createViewRow(page, incidentId, notesViewSchemaId, {
    client_txn_id: uniqueTxn("e804-bravo"),
    "note.title": "Bravo Alpha",
    "note.body": "Shell token appears earlier alphabetically",
  });
  const powershell = await createViewRow(page, incidentId, notesViewSchemaId, {
    client_txn_id: uniqueTxn("e804-powershell"),
    "note.title": "Powershell only",
  });
  const phrase = await createViewRow(page, incidentId, notesViewSchemaId, {
    client_txn_id: uniqueTxn("e804-phrase"),
    "note.title": "Phrase note",
    "note.body": "alpha middle shell",
  });
  await createViewRow(page, incidentId, notesViewSchemaId, {
    client_txn_id: uniqueTxn("e804-wildcard"),
    "note.title": "Wildcard note",
    "note.body": "shells alpha",
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
  const evidencePrefix = await createViewRow(
    page,
    incidentId,
    evidenceViewSchemaId,
    {
      client_txn_id: uniqueTxn("e804-evidence-alpha"),
      "evidence.title": "Prefix Alpha",
      "evidence.storage_ref": "Alpha",
    },
  );
  const evidenceWildcardLiteral = await createViewRow(
    page,
    incidentId,
    evidenceViewSchemaId,
    {
      client_txn_id: uniqueTxn("e804-evidence-wildcard"),
      "evidence.title": "Prefix Wildcard",
      "evidence.storage_ref": "Al%ha",
    },
  );
  await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("e804-evidence-infix"),
    "evidence.title": "Prefix Infix",
    "evidence.storage_ref": "XAlpha",
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
  await expect(
    page.getByTestId(gridShellTestId(notesViewSchemaId)),
  ).toBeVisible();

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
        sort: [{ direction: "asc", field_key: "note.title" }],
        filters: [
          {
            arg: { query: "shell alpha shell" },
            field_key: "note.full_text",
            op: "full_text",
          },
        ],
      }),
    ),
  ).toEqual([
    String(alpha.record_id),
    String(bravo.record_id),
    String(phrase.record_id),
  ]);
  expect(
    rowIDs(
      await queryViewRows(page, incidentId, notesViewSchemaId, {
        sort: [{ direction: "asc", field_key: "note.title" }],
        filters: [
          {
            arg: { query: "shell" },
            field_key: "note.full_text",
            op: "full_text",
          },
        ],
      }),
    ),
  ).toEqual([
    String(alpha.record_id),
    String(bravo.record_id),
    String(phrase.record_id),
  ]);
  expect(
    rowIDs(
      await queryViewRows(page, incidentId, notesViewSchemaId, {
        filters: [
          {
            arg: { query: "respond" },
            field_key: "note.full_text",
            op: "full_text",
          },
        ],
      }),
    ),
  ).toEqual([]);
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
      await queryViewRows(page, incidentId, evidenceViewSchemaId, {
        sort: [{ direction: "asc", field_key: "evidence.title" }],
        filters: [
          {
            arg: { value: "al" },
            field_key: "evidence.storage_ref",
            op: "prefix",
          },
        ],
      }),
    ),
  ).toEqual([
    String(evidencePrefix.record_id),
    String(evidenceWildcardLiteral.record_id),
  ]);
  expect(
    rowIDs(
      await queryViewRows(page, incidentId, evidenceViewSchemaId, {
        filters: [
          {
            arg: { value: "lph" },
            field_key: "evidence.storage_ref",
            op: "prefix",
          },
        ],
      }),
    ),
  ).toEqual([]);
  expect(
    rowIDs(
      await queryViewRows(page, incidentId, evidenceViewSchemaId, {
        filters: [
          {
            arg: { value: "al%" },
            field_key: "evidence.storage_ref",
            op: "prefix",
          },
        ],
      }),
    ),
  ).toEqual([String(evidenceWildcardLiteral.record_id)]);
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
  return page.waitForRequest((request) =>
    isViewQueryRequest(request, incidentId, viewSchemaId),
  );
}

function waitForViewQueryBody(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  expectedBody: Record<string, unknown>,
) {
  return page.waitForRequest(
    (request) =>
      isViewQueryRequest(request, incidentId, viewSchemaId) &&
      postBodyEquals(request, expectedBody),
  );
}

function isViewQueryRequest(
  request: { method: () => string; url: () => string },
  incidentId: string,
  viewSchemaId: string,
) {
  return (
    request.method() === "POST" &&
    request
      .url()
      .endsWith(`/api/v1/incidents/${incidentId}/views/${viewSchemaId}/query`)
  );
}

function postBodyEquals(
  request: { postData: () => string | null },
  expectedBody: Record<string, unknown>,
) {
  return stableJSON(readPostBody(request)) === stableJSON(expectedBody);
}

function readPostBody(request: { postData: () => string | null }) {
  return JSON.parse(request.postData() ?? "{}") as Record<string, unknown>;
}

function stableJSON(value: unknown): string {
  if (Array.isArray(value)) {
    return `[${value.map((item) => stableJSON(item)).join(",")}]`;
  }
  if (value !== null && typeof value === "object") {
    return `{${Object.entries(value as Record<string, unknown>)
      .sort(([left], [right]) => left.localeCompare(right))
      .map(([key, item]) => `${JSON.stringify(key)}:${stableJSON(item)}`)
      .join(",")}}`;
  }
  return JSON.stringify(value) ?? "undefined";
}

type WorkbookSheetRef = {
  kind: "saved_view" | "view_schema";
  id: string;
};

type SavedViewResource = {
  display_name: string;
  layout_json?: Record<string, unknown>;
  owner_user_id?: string | null;
  query_json?: Record<string, unknown>;
  saved_view_id: string;
  saved_view_version?: number;
  scope?: string;
  view_schema_id: string;
};

type WorkbookPreferencesResource = {
  default_sheet_ref?: WorkbookSheetRef | null;
  home_sheet_ref?: WorkbookSheetRef | null;
};

type WorkbookStartupResource = {
  selected_sheet_ref: WorkbookSheetRef;
  selected_view_schema_id: string;
  source: "default" | "explicit" | "home" | "timeline";
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

async function deleteSavedView(
  page: Page,
  incidentId: string,
  savedViewId: string,
) {
  const response = await page.request.delete(
    `${apiBase}/api/v1/incidents/${incidentId}/saved-views/${savedViewId}`,
    {
      headers: await csrfHeaders(page),
    },
  );
  expect(response.ok()).toBeTruthy();
}

async function putDefaultSheetRef(
  page: Page,
  incidentId: string,
  ref: WorkbookSheetRef | null,
) {
  const response = await page.request.put(
    `${apiBase}/api/v1/incidents/${incidentId}/workbook-preferences/default`,
    {
      headers: await csrfHeaders(page),
      data: { default_sheet_ref: ref },
    },
  );
  expect(response.ok()).toBeTruthy();
}

async function putUserHomeSheetRef(
  page: Page,
  incidentId: string,
  ref: WorkbookSheetRef | null,
) {
  const response = await page.request.put(
    `${apiBase}/api/v1/incidents/${incidentId}/workbook-preferences/me`,
    {
      headers: await csrfHeaders(page),
      data: { home_sheet_ref: ref },
    },
  );
  expect(response.ok()).toBeTruthy();
}

async function getDefaultWorkbookPreferences(page: Page, incidentId: string) {
  const response = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}/workbook-preferences/default`,
  );
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: WorkbookPreferencesResource })
    .data;
}

async function getUserWorkbookPreferences(page: Page, incidentId: string) {
  const response = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}/workbook-preferences/me`,
  );
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: WorkbookPreferencesResource })
    .data;
}

async function getWorkbookStartup(page: Page, incidentId: string, query = "") {
  const response = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}/workbook-startup${query}`,
  );
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: WorkbookStartupResource }).data;
}

async function visibleRecordIds(page: Page) {
  return page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator('[role="row"][data-grid-record-id]:not([data-grid-record-id=""])')
    .evaluateAll((rows) =>
      rows.map((row) => row.getAttribute("data-grid-record-id") ?? ""),
    );
}

async function visibleGroupLabels(page: Page) {
  return page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator(
      gridGroupRowsSelector(timelineViewSchemaId, "timeline.capture_state"),
    )
    .evaluateAll((nodes) =>
      nodes.map((node) => (node.textContent ?? "").trim()),
    );
}

async function visibleGroupTestIds(page: Page) {
  return page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator(
      gridGroupRowsSelector(timelineViewSchemaId, "timeline.capture_state"),
    )
    .evaluateAll((nodes) =>
      nodes.map((node) => node.getAttribute("data-testid") ?? ""),
    );
}

async function expectFirstDataRow(page: Page, recordId: string) {
  await expect
    .poll(async () => (await visibleRecordIds(page))[0] ?? null)
    .toBe(recordId);
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
