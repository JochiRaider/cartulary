import {
  applyFilterChip,
  assertActiveFilterChipVisible,
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  currentIncidentRoleTestId,
  draftCellTestId,
  draftRowCreateButtonTestId,
  rowCellTestId,
  rowHistoryActionTestId,
  rowHistoryOpenButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
  saveStateTestId,
  timelineCaptureActionTestId,
  timelineMutationSubstrateReadyTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowSupersedeButtonTestId,
  timelineRowVersionTestId,
  timelineScalarEditorTestId,
  workbookInspectorCloseButtonTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
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
import {
  fetchFullRecordHistory,
  fetchRecordHistoryCount,
} from "./support/workbook/history";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import {
  clickTimelineRowAction,
  commitInspectorScalarEdit,
  openTimelineInspector,
  openTimelineRowActions,
} from "./support/workbook/rowMutations";
import { authorTimelineSupersession } from "./support/workbook/timelineCaptureActions";

async function expectCurrentIncidentRole(page: Page, roleText: string) {
  const accountMenuTrigger = page.getByRole("button", {
    name: "Account and application navigation",
  });
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
    uniqueIncidentKey("TIMELINE-WORKBOOK"),
    "Timeline timeline-workbook",
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
    uniqueIncidentKey("TIMELINE-BLANK-CREATE"),
    "Timeline timeline-workbook blank create",
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
  const reviewerEmail = uniqueEmail("timeline_mutation-e303-reviewer");
  const reviewerPassword = "TimelineMutationE303Reviewer!";
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TIMELINE-PUBLIC-ROUTE"),
    "Timeline timeline-workbook",
  );
  const reviewerUser = await createIncidentMemberUser(page, incidentId, {
    email: reviewerEmail,
    display_name: "Timeline E303 Reviewer",
    initial_password: reviewerPassword,
    role: "reviewer",
    is_deployment_admin: false,
    mfa_required: false,
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
      createdBy: "timeline_mutation reviewer lifecycle flow",
      email: reviewerEmail,
      incidentId,
      password: reviewerPassword,
      purpose: "timeline_mutation e303 reviewer workbook lifecycle",
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
    .getByTestId(timelineRowSupersedeButtonTestId(recordId))
    .click();
  await authorTimelineSupersession(
    reviewerPage,
    recordId,
    "Duplicate source; the replacement contains the verified chronology.",
    replacementId,
  );
  const supersedeRequest = reviewerPage.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request.url().endsWith(`/api/v1/records/${recordId}/supersede`),
  );
  await reviewerPage
    .getByTestId(timelineCaptureActionTestId("confirm", recordId))
    .press("Enter");
  const supersedeBody = (await supersedeRequest).postDataJSON() as Record<
    string,
    unknown
  >;
  expect(supersedeBody.base_row_version).toBe(3);
  expect(supersedeBody.reason).toBe(
    "Duplicate source; the replacement contains the verified chronology.",
  );
  expect(supersedeBody.replacement_record_id).toBe(replacementId);
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

test("uses public history and visible state to prove replay avoids duplicate mutation effects", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TIMELINE-QUERY"),
    "Timeline timeline-workbook",
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
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page: observer,
    recordId,
    surface: timelineViewSchemaId,
  });
  const baselineObserverQueries = observerQueryCount;
  const baselineHistoryCount = await fetchRecordHistoryCount(page, recordId);

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
  await expect(
    observer.getByTestId(timelineRowVersionTestId(recordId)),
  ).toHaveText("2");
  await expect(
    observer.getByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("Replay row patched");
  expect(observerQueryCount).toBe(baselineObserverQueries);
  const historyCountAfterFirstPatch = await fetchRecordHistoryCount(
    page,
    recordId,
  );
  expect(historyCountAfterFirstPatch).toBe(baselineHistoryCount + 1);

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
  expect(await fetchRecordHistoryCount(page, recordId)).toBe(
    historyCountAfterFirstPatch,
  );
  await expect(
    observer.getByTestId(timelineRowVersionTestId(recordId)),
  ).toHaveText("2");
  await observerContext.close();
});

test("Timeline exact action recovery preserves committed transitions change sets and replacement links", async ({
  page,
  workerAdmin,
}, testInfo) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TIMELINE-RECOVERY"),
    "Timeline capture action recovery",
  );
  const replacement = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("replacement"),
      "timeline.activity_synopsis_text": "Verified replacement",
      "timeline.device_object_text": "Workstation B",
    },
  );
  const reviewedReplacement = await page.request.post(
    `${apiBase}/api/v1/records/${replacement.record_id}/mark-reviewed`,
    {
      headers: await csrfHeaders(page),
      data: {
        base_row_version: replacement.row_version,
        client_txn_id: uniqueTxn("review-replacement"),
      },
    },
  );
  expect(reviewedReplacement.ok()).toBeTruthy();
  const cases = [
    { action: "mark-reviewed", replacement: null, loss: "before" },
    { action: "mark-reviewed", replacement: null, loss: "committed" },
    { action: "supersede", replacement: null, loss: "malformed" },
    {
      action: "supersede",
      replacement: replacement.record_id,
      loss: "committed",
    },
  ] as const;
  const evidence: unknown[] = [];
  for (const scenario of cases) {
    const target = await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("target"),
      "timeline.activity_synopsis_text": `Recovery ${scenario.action} ${scenario.loss}`,
    });
    const endpoint = `**/api/v1/records/${target.record_id}/${scenario.action}`;
    const requests: { url: string; method: string; body: string | null }[] = [];
    const receipts: Record<string, unknown>[] = [];
    await page.route(endpoint, async (route) => {
      requests.push({
        url: route.request().url(),
        method: route.request().method(),
        body: route.request().postData(),
      });
      if (requests.length === 1 && scenario.loss === "before") {
        await route.abort("failed");
        return;
      }
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      receipts.push((await response.json()).data);
      if (requests.length === 1) {
        if (scenario.loss === "malformed")
          await route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              data: { target_record_id: target.record_id },
              meta: { request_id: "lost-envelope" },
            }),
          });
        else await route.abort("failed");
      } else await route.fulfill({ response });
    });
    await page.goto(`/?incident_id=${incidentId}`);
    if (scenario.replacement) {
      await applyFilterChip(
        page,
        timelineViewSchemaId,
        "timeline.capture_state",
        "rough",
      );
      await assertActiveFilterChipVisible(
        page,
        timelineViewSchemaId,
        "timeline.capture_state",
      );
      await expect(
        page.getByTestId(
          rowCellTestId(
            replacement.record_id,
            "timeline.activity_synopsis_text",
          ),
        ),
      ).toHaveCount(0);
    }
    await openTimelineRowActions(page, target.record_id);
    const authored =
      "  Later evidence corrects this entry.\nKeep the original attribution.  ";
    if (scenario.action === "supersede") {
      await page
        .getByTestId(timelineRowSupersedeButtonTestId(target.record_id))
        .press("Enter");
      await authorTimelineSupersession(
        page,
        target.record_id,
        authored,
        scenario.replacement,
      );
      await page
        .getByTestId(timelineCaptureActionTestId("confirm", target.record_id))
        .press("Enter");
    } else
      await page
        .getByTestId(timelineRowMarkReviewedButtonTestId(target.record_id))
        .press("Enter");
    await expect(
      page.getByText("Timeline action outcome unknown.", { exact: true }),
    ).toBeVisible();
    await page
      .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
      .click();
    await expect(
      page.getByTestId(
        workbookInspectorCloseButtonTestId(timelineViewSchemaId),
      ),
    ).toHaveCount(0);
    const before = await fetchFullRecordHistory(page, target.record_id);
    await page.getByRole("button", { name: /^Timeline actions \(/ }).click();
    const recovery = page.getByRole("region", {
      name: "Timeline action recovery",
      exact: true,
    });
    await expect(recovery).toBeFocused();
    const retry = page.getByTestId(
      timelineCaptureActionTestId("retry", target.record_id),
    );
    await retry.focus();
    await retry.press("Enter");
    await expect(
      page.getByTestId(timelineCaptureActionTestId("result", target.record_id)),
    ).toContainText("completed");
    await expect(
      page.getByTestId(
        timelineCaptureActionTestId("refresh", target.record_id),
      ),
    ).toHaveCount(0);
    expect(requests).toHaveLength(2);
    await expect(
      page.getByTestId(
        workbookInspectorCloseButtonTestId(timelineViewSchemaId),
      ),
    ).toHaveCount(0);
    expect(requests[1]).toEqual(requests[0]);
    const body = JSON.parse(requests[0]?.body ?? "{}");
    expect(body.base_row_version).toBe(1);
    expect(body.client_txn_id).toBeTruthy();
    if (scenario.replacement)
      expect(body.replacement_record_id).toBe(scenario.replacement);
    else expect(body).not.toHaveProperty("replacement_record_id");
    if (scenario.action === "supersede")
      expect(body.reason).toBe(authored.trim());
    else expect(body).not.toHaveProperty("reason");
    if (scenario.loss !== "before") expect(receipts[1]).toEqual(receipts[0]);
    const after = await fetchFullRecordHistory(page, target.record_id);
    if (scenario.loss !== "before") expect(after).toEqual(before);
    expect(after.row_version).toBe(2);
    const receipt = receipts[0];
    if (!receipt) throw new Error("Expected service receipt");
    const transition = after.items.filter(
      (item) => item.change_set_id === receipt.change_set_id,
    );
    expect(transition.length).toBeGreaterThan(0);
    expect(new Set(transition.map((item) => item.change_set_id)).size).toBe(1);
    expect(new Set(transition.map((item) => item.actor_user_id)).size).toBe(1);
    expect([...new Set(transition.map((item) => item.actor_user_id))]).toEqual([
      workerAdmin.user_id,
    ]);
    const linkUnits = transition
      .flatMap((item) => item.diff_summary.units)
      .filter((unit) => unit.target_kind === "record_link");
    expect(linkUnits).toHaveLength(scenario.replacement ? 1 : 0);
    const row = (
      await queryViewRows(page, incidentId, timelineViewSchemaId)
    ).find((row) => row.record_id === target.record_id);
    expect(row?.cells["timeline.capture_state"]?.value).toBe(
      scenario.action === "supersede" ? "superseded" : "reviewed",
    );
    expect(row?.cells["timeline.replacement_record_id"]?.value).toBe(
      scenario.replacement,
    );
    await expect(recovery).toContainText(
      scenario.replacement
        ? "Verified replacement"
        : scenario.action === "supersede"
          ? "No replacement"
          : "Mark reviewed",
    );
    evidence.push({
      scenario,
      requests,
      receipts,
      before,
      after,
      replacementLinkCount: linkUnits.length,
    });
    await page.keyboard.press("Escape");
    await expect(
      page.getByRole("button", { name: /^Timeline actions \(/ }),
    ).toBeFocused();
    if (scenario.replacement) {
      await assertActiveFilterChipVisible(
        page,
        timelineViewSchemaId,
        "timeline.capture_state",
      );
      await expect(
        page.getByTestId(
          rowCellTestId(target.record_id, "timeline.activity_synopsis_text"),
        ),
      ).toHaveCount(0);
    }
    await page.unroute(endpoint);
    if (scenario.action === "supersede") {
      const rollbackItem = transition.find((item) =>
        item.available_rollback_actions.includes("change_set"),
      );
      if (!rollbackItem)
        throw new Error("Expected existing change-set rollback");
      await page.goto(`/?incident_id=${incidentId}`);
      await clickTimelineRowAction(
        page,
        target.record_id,
        rowHistoryOpenButtonTestId(target.record_id),
      );
      const anchor = {
        action: "change_set" as const,
        historyItemRef: rollbackItem.history_item_ref,
      };
      await page.getByTestId(rowHistoryActionTestId(anchor)).click();
      const rollbackResponse = page.waitForResponse(
        (response) =>
          response.request().method() === "POST" &&
          response
            .url()
            .endsWith(`/api/v1/records/${target.record_id}/rollback`),
      );
      await page
        .getByTestId(rowHistoryRollbackConfirmButtonTestId(anchor))
        .click();
      expect((await rollbackResponse).ok()).toBeTruthy();
      await expect(
        page.getByTestId(
          rowCellTestId(target.record_id, "timeline.capture_state"),
        ),
      ).toHaveText(String(target.cells["timeline.capture_state"]?.value));
      const restored = (
        await queryViewRows(page, incidentId, timelineViewSchemaId)
      ).find((row) => row.record_id === target.record_id);
      expect(
        restored?.cells["timeline.replacement_record_id"]?.value,
      ).toBeNull();
      expect(restored?.row_version).toBeGreaterThan(2);
      evidence.push({
        rollbackOf: receipt.change_set_id,
        restored,
        history: await fetchFullRecordHistory(page, target.record_id),
      });
    }
  }
  await testInfo.attach("timeline-capture-exact-replay.json", {
    body: JSON.stringify(evidence, null, 2),
    contentType: "application/json",
  });
});
