import {
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  currentIncidentRoleTestId,
  draftCellTestId,
  draftRowCreateButtonTestId,
  rowCellTestId,
  saveStateTestId,
  timelineMutationSubstrateReadyTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowReplacementInputTestId,
  timelineRowSupersedeButtonTestId,
  timelineRowVersionTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import type { Page, Route } from "@playwright/test";

import { expect, test } from "./fixtures";
import { waitForCommittedRowSummary } from "./measurement/timingSupport";
import { openIncidentAsTrackedUser } from "./pages/incidentDirectory";
import { gridDraftRows, gridSavedRows } from "./pages/workbookInspector";
import { csrfHeaders } from "./support/auth/browserSession";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase, webBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createViewRow } from "./support/workbook/query";
import {
  fetchTimelineRecordChangeCount,
  fetchTimelineRecordSubstrate,
} from "./support/workbook/refresh";
import {
  clickTimelineRowAction,
  commitInspectorScalarEdit,
  openTimelineInspector,
  openTimelineRowActions,
} from "./support/workbook/rowMutations";

const timelineViewSchemaId = "cartulary.view.timeline.v2";

async function expectCurrentIncidentRole(page: Page, roleText: string) {
  const accountMenuTrigger = page.getByLabel(
    "Account and application navigation",
  );
  await accountMenuTrigger.click();
  await expect(page.getByTestId(currentIncidentRoleTestId())).toHaveText(
    roleText,
  );
  await accountMenuTrigger.click();
}

test("creates a Timeline row in-grid and continues editing on the draft row", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E301"),
    "Timeline E-3-01",
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();

  const draftSummaryTestId = draftCellTestId("timeline.activity_synopsis_text");
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: draftSummaryTestId,
  });
  const draftSummary = page.getByTestId(draftSummaryTestId);
  await draftSummary.fill("First browser fact");
  await draftSummary.press("Enter");

  const committedRow = await waitForCommittedRowSummary(page, {
    expectedSummary: "First browser fact",
    surface: timelineViewSchemaId,
    timeoutMs: 5_000,
  });
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await expect(gridSavedRows(page, timelineViewSchemaId)).toHaveCount(1);
  await expect(gridDraftRows(page, timelineViewSchemaId)).toHaveCount(1);
  await expect(
    page.getByTestId(
      rowCellTestId(committedRow.recordId, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("First browser fact");
  await expect(
    page.getByTestId(draftCellTestId("timeline.activity_synopsis_text")),
  ).toBeFocused();
});

test("supports explicit blank Timeline row creation with only client_txn_id", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E301BLANK"),
    "Timeline E-3-01 blank create",
  );

  const createBodies: Record<string, unknown>[] = [];
  const createRoute = `**/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`;
  const routeHandler = async (route: Route) => {
    createBodies.push(
      route.request().postDataJSON() as Record<string, unknown>,
    );
    await route.fallback();
  };
  await page.route(createRoute, routeHandler);

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

  await page.getByTestId(draftRowCreateButtonTestId()).click();
  const committedRow = await waitForCommittedRowSummary(page, {
    expectedSummary: "",
    surface: timelineViewSchemaId,
    timeoutMs: 5_000,
  });

  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await expect(gridSavedRows(page, timelineViewSchemaId)).toHaveCount(1);
  await expect(gridDraftRows(page, timelineViewSchemaId)).toHaveCount(1);
  await expect(
    page.getByTestId(
      rowCellTestId(committedRow.recordId, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("—");
  await expect(
    page.getByTestId(
      rowCellTestId(committedRow.recordId, "timeline.capture_state"),
    ),
  ).toHaveText("rough");
  expect(createBodies).toHaveLength(1);
  expect(Object.keys(createBodies[0] ?? {})).toEqual(["client_txn_id"]);
  expect(typeof createBodies[0]?.client_txn_id).toBe("string");

  await page.unroute(createRoute, routeHandler);
});

test("drives review, demotion, and supersede through the visible workbook surface", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const reviewerEmail = uniqueEmail("phase3-e303-reviewer");
  const reviewerPassword = "Phase3E303Reviewer!";
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E303"),
    "Timeline E-3-03",
  );
  const reviewerUser = await createIncidentMemberUser(page, incidentId, {
    email: reviewerEmail,
    display_name: "Timeline E303 Reviewer",
    initial_password: reviewerPassword,
    role: "reviewer",
  });
  const primaryRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("primary"),
      "timeline.activity_synopsis_text": "Primary row",
    },
  );
  const replacementRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("replacement"),
      "timeline.activity_synopsis_text": "Replacement row",
    },
  );
  const recordId = primaryRow.record_id as string;
  const replacementId = replacementRow.record_id as string;

  const reviewerPage = await openIncidentAsTrackedUser(
    browser,
    sessionTracker,
    {
      createdBy: "phase3 reviewer lifecycle flow",
      email: reviewerEmail,
      incidentId,
      password: reviewerPassword,
      purpose: "phase3 e303 reviewer workbook lifecycle",
      userId: reviewerUser.user_id,
    },
  );

  await expectCurrentIncidentRole(
    reviewerPage,
    "Current incident role: reviewer",
  );

  await clickTimelineRowAction(
    reviewerPage,
    recordId,
    timelineRowMarkReviewedButtonTestId(recordId),
  );
  await expect(
    reviewerPage.getByTestId(rowCellTestId(recordId, "timeline.capture_state")),
  ).toHaveText("reviewed");
  await expect(
    reviewerPage.getByTestId(timelineRowVersionTestId(recordId)),
  ).toHaveText("2");

  await openTimelineInspector(reviewerPage, recordId);
  await commitInspectorScalarEdit(
    reviewerPage,
    recordId,
    "timeline.raw_activity_text",
    "Material edit after review",
  );
  await expect(
    reviewerPage.getByTestId(rowCellTestId(recordId, "timeline.capture_state")),
  ).toHaveText("enriched");
  await expect(
    reviewerPage.getByTestId(timelineRowVersionTestId(recordId)),
  ).toHaveText("3");

  await openTimelineRowActions(reviewerPage, recordId);
  await reviewerPage
    .getByTestId(timelineRowReplacementInputTestId(recordId))
    .fill(replacementId);
  const supersedeRequest = reviewerPage.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request.url().endsWith(`/api/v1/records/${recordId}/supersede`),
  );
  await reviewerPage
    .getByTestId(timelineRowSupersedeButtonTestId(recordId))
    .click();
  const supersedeBody = (await supersedeRequest).postDataJSON() as Record<
    string,
    unknown
  >;
  expect(supersedeBody.base_row_version).toBe(3);
  await expect(
    reviewerPage.getByTestId(rowCellTestId(recordId, "timeline.capture_state")),
  ).toHaveText("superseded");
  await expect(
    reviewerPage.getByTestId(timelineRowVersionTestId(recordId)),
  ).toHaveText("4");
  await openTimelineRowActions(reviewerPage, recordId);
  await expect(
    reviewerPage.getByTestId(timelineRowMarkReviewedButtonTestId(recordId)),
  ).toBeDisabled();
  await reviewerPage.context().close();
});

test("uses a disclosed hybrid replay harness to prove replay avoids duplicate history and visible invalidation", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E304"),
    "Timeline E-3-04",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("seed"),
    "timeline.activity_synopsis_text": "Replay row",
  });
  const recordId = row.record_id as string;

  const observerContext = await sessionTracker.newTrackedContext(
    browser,
    await page.context().storageState(),
  );
  const observer = await observerContext.newPage();
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

  await observer.goto(`${webBase}/?incident_id=${incidentId}`);
  await expect(
    observer.getByTestId(timelineRowVersionTestId(recordId)),
  ).toHaveText("1");
  const baselineObserverQueries = observerQueryCount;
  const baselineRecordChangeCount = await fetchTimelineRecordChangeCount(
    page,
    recordId,
  );

  await page.goto(`/?incident_id=${incidentId}`);
  const summaryDisplay = page.getByTestId(
    rowCellTestId(recordId, "timeline.activity_synopsis_text"),
  );
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId,
    surface: timelineViewSchemaId,
  });
  const firstPatchResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/v1/records/${recordId}`),
  );

  await summaryDisplay.click();
  const summaryEditor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.activity_synopsis_text",
      recordId,
      surface: "grid",
    }),
  );
  await expect(summaryEditor).toBeFocused();
  await summaryEditor.fill("Replay row patched");
  await summaryEditor.press("Enter");
  const patchResponse = await firstPatchResponse;
  const firstPatchBody = JSON.parse(
    patchResponse.request().postData() ?? "{}",
  ) as Record<string, unknown>;
  const firstPatchData = (
    (await patchResponse.json()) as { data: { change_set_id: string } }
  ).data;

  await expect(page.getByTestId(timelineRowVersionTestId(recordId))).toHaveText(
    "2",
  );
  await expect
    .poll(() => observerQueryCount, { timeout: 5_000 })
    .toBeGreaterThan(baselineObserverQueries);
  await expect(
    observer.getByTestId(timelineRowVersionTestId(recordId)),
  ).toHaveText("2");
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
  await expect(
    observer.getByTestId(timelineRowVersionTestId(recordId)),
  ).toHaveText("2");
  await observerContext.close();
});
