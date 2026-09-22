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
  draftTimelineCollectionInputTestId,
  gridRowTestId,
  gridRowVersionAttribute,
  rowCellTestId,
  rowHistoryActionTestId,
  rowHistoryOpenButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
  saveStateTestId,
  timelineCaptureActionTestId,
  timelineMutationSubstrateReadyTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowSupersedeButtonTestId,
  timelineScalarEditorTestId,
  workbookAddRowButtonTestId,
  workbookEditRecoveryDiscardButtonTestId,
  workbookEditRecoveryRetryButtonTestId,
  workbookInspectorCloseButtonTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Page, Route } from "@playwright/test";
import { expect, test } from "./fixtures";
import { waitForCommittedRowSummary } from "./measurement/timingSupport";
import { openIncidentAsTrackedUser } from "./pages/incidentDirectory";
import { gridDraftRows, gridSavedRows } from "./pages/workbookInspector";
import { csrfHeaders } from "./support/auth/browserSession";
import { collectionItems } from "./support/entities/mentions";
import { currentLifecycle, lifecycleAction } from "./support/incidentLifecycle";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase, webBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { holdBrowserRequest } from "./support/transport/requestInterception";
import { showTimelineCollectionColumns } from "./support/workbook/collections";
import {
  fetchFullRecordHistory,
  fetchRecordHistoryCount,
  openHistoryEventDetails,
} from "./support/workbook/history";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { openRecoveryItem, recoveryEntry } from "./support/workbook/recovery";
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

async function openTimelineCreateFixture(page: Page, name: string) {
  const incidentId = await createIncident(page, uniqueIncidentKey(name), name);
  const createPath = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`;
  const creates: string[] = [];
  page.on("request", (request) => {
    if (request.method() === "POST" && request.url().endsWith(createPath))
      creates.push(request.postData() ?? "");
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  return { incidentId, createPath, creates };
}

test("Timeline Create uses native click Enter Space and click-only activation once while pending", async ({
  page,
}) => {
  const f = await openTimelineCreateFixture(page, "CREATE-ACTIVATION");
  const create = page.getByRole("button", {
    name: "Create timeline row",
    exact: true,
  });
  const draft = page.getByTestId(
    draftCellTestId("timeline.activity_synopsis_text"),
  );
  for (const activation of [
    "pointer",
    "Enter",
    "Space",
    "click-only",
  ] as const) {
    const before = f.creates.length;
    const held = await holdBrowserRequest(page, {
      method: "POST",
      path: f.createPath,
    });
    try {
      if (activation === "pointer") {
        await create.hover();
        await page.mouse.down();
        expect(f.creates).toHaveLength(before);
        await page.mouse.up();
      } else if (activation === "click-only") {
        // One browser task also exercises admission before disabled is painted.
        await create.evaluate((button: HTMLButtonElement) => {
          button.click();
          button.click();
        });
      } else {
        await create.focus();
        await expect(create).toBeFocused();
        await page.keyboard.down(activation);
        if (activation === "Space") expect(f.creates).toHaveLength(before);
        await page.keyboard.up(activation);
      }
      await held.waitForHit;
      await expect(create).toBeDisabled();
      await create.evaluate((button: HTMLButtonElement) => button.click());
      expect(f.creates).toHaveLength(before + 1);
      expect(JSON.parse(f.creates[before] ?? "null")).toEqual({
        client_txn_id: expect.any(String),
      });
      held.release();
      await expect(gridSavedRows(page, timelineViewSchemaId)).toHaveCount(
        before + 1,
      );
      await expect(gridDraftRows(page, timelineViewSchemaId)).toHaveCount(1);
      await expect(draft).toBeFocused();
      await expect(draft).toHaveValue("");
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      expect(f.creates).toHaveLength(before + 1);
    } finally {
      await held.dispose();
    }
  }
  expect(
    new Set(f.creates.map((body) => JSON.parse(body).client_txn_id)).size,
  ).toBe(4);
  expect(
    await queryViewRows(page, f.incidentId, timelineViewSchemaId),
  ).toHaveLength(4);
});

test("Timeline Create ignores cancelled and non-primary presses without blur-submitting raw drafts", async ({
  page,
}) => {
  const f = await openTimelineCreateFixture(page, "CREATE-CANCEL");
  await showTimelineCollectionColumns(page, ["Hosts"]);
  const raw = "Unresolved HOST Ω";
  const input = page.getByTestId(
    draftTimelineCollectionInputTestId("timeline.host_refs"),
  );
  const create = page.getByRole("button", {
    name: "Create timeline row",
    exact: true,
  });
  await input.fill(raw);
  // This observation records the native sequence without replacing any handler.
  await input.evaluate((element) =>
    element.addEventListener("blur", () =>
      element.setAttribute("data-create-test-blurred", "true"),
    ),
  );
  await create.hover();
  await page.mouse.down();
  expect(f.creates).toHaveLength(0);
  await expect(input).toBeFocused();
  await page.mouse.move(0, 0);
  await page.mouse.up();
  expect(f.creates).toHaveLength(0);
  for (const button of ["right", "middle"] as const) {
    await create.click({ button });
    expect(f.creates).toHaveLength(0);
    await expect(input).toBeFocused();
  }
  // A browser-native cancelled contact must not synthesize an activation click.
  const cdp = await page.context().newCDPSession(page);
  const bounds = await create.boundingBox();
  if (!bounds) throw new Error("Missing Create button bounds");
  await create.evaluate((element) =>
    element.addEventListener("pointercancel", () =>
      element.setAttribute("data-create-test-cancelled", "true"),
    ),
  );
  try {
    await cdp.send("Input.dispatchTouchEvent", {
      type: "touchStart",
      touchPoints: [
        { x: bounds.x + bounds.width / 2, y: bounds.y + bounds.height / 2 },
      ],
    });
    await cdp.send("Input.dispatchTouchEvent", {
      type: "touchCancel",
      touchPoints: [],
    });
    await expect(create).toHaveAttribute("data-create-test-cancelled", "true");
  } finally {
    await cdp.detach();
  }
  expect(f.creates).toHaveLength(0);
  await expect(input).toHaveValue(raw);
  await expect(input).toBeFocused();
  await expect(input).not.toHaveAttribute("data-create-test-blurred", "true");
  await create.click();
  await expect(gridSavedRows(page, timelineViewSchemaId)).toHaveCount(1);
  await expect(gridDraftRows(page, timelineViewSchemaId)).toHaveCount(1);
  await expect(
    page.getByTestId(draftCellTestId("timeline.activity_synopsis_text")),
  ).toBeFocused();
  expect(f.creates).toHaveLength(1);
  expect(JSON.parse(f.creates[0] ?? "null")).toEqual({
    client_txn_id: expect.any(String),
    "timeline.host_refs": {
      kind: "collection_actions_v1",
      actions: [{ op: "add_token", raw_text: raw }],
    },
  });
  const rows = await queryViewRows(page, f.incidentId, timelineViewSchemaId);
  expect(rows).toHaveLength(1);
  const saved = rows[0];
  if (!saved) throw new Error("Missing created row");
  expect(
    collectionItems(saved, "timeline.host_refs").map((item) => item.raw_text),
  ).toEqual([raw]);
});

test("Timeline Create preserves pending fast capture and yields fresh draft focus to newer intent", async ({
  page,
}) => {
  const f = await openTimelineCreateFixture(page, "CREATE-CONTINUITY");
  const create = page.getByRole("button", {
    name: "Create timeline row",
    exact: true,
  });
  const synopsis = "timeline.activity_synopsis_text";
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: draftCellTestId(synopsis),
  });
  const draft = page.getByTestId(draftCellTestId(synopsis));
  const heldCapture = await holdBrowserRequest(page, {
    method: "POST",
    path: f.createPath,
  });
  try {
    await draft.fill("Incomplete fact Ω");
    await heldCapture.waitForHit;
    await expect(create).toBeDisabled();
    await create.evaluate((button: HTMLButtonElement) => button.click());
    await create.hover();
    await page.mouse.down();
    await page.mouse.up();
    expect(f.creates).toHaveLength(1);
    await expect(draft).toHaveValue("Incomplete fact Ω");
    heldCapture.release();
    await expect
      .poll(async () =>
        (await queryViewRows(page, f.incidentId, timelineViewSchemaId)).map(
          (row) => row.cells[synopsis]?.value,
        ),
      )
      .toEqual(["Incomplete fact Ω"]);
    await expect(gridSavedRows(page, timelineViewSchemaId)).toHaveCount(1);
    await expect(gridDraftRows(page, timelineViewSchemaId)).toHaveCount(1);
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    expect(f.creates).toHaveLength(1);
  } finally {
    await heldCapture.dispose();
  }

  const heldExplicit = await holdBrowserRequest(page, {
    method: "POST",
    path: f.createPath,
  });
  const destination = page.getByRole("button", {
    name: "Account and application navigation",
    exact: true,
  });
  try {
    await create.click();
    await heldExplicit.waitForHit;
    await destination.focus();
    heldExplicit.release();
    await expect(gridSavedRows(page, timelineViewSchemaId)).toHaveCount(2);
    await expect(gridDraftRows(page, timelineViewSchemaId)).toHaveCount(1);
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(destination).toBeFocused();
    expect(f.creates).toHaveLength(2);
  } finally {
    await heldExplicit.dispose();
  }
});

test("Timeline Create rejection retains raw authoring and existing discard recovery", async ({
  page,
}) => {
  const f = await openTimelineCreateFixture(page, "CREATE-REJECT");
  await showTimelineCollectionColumns(page, ["Hosts"]);
  const input = page.getByTestId(
    draftTimelineCollectionInputTestId("timeline.host_refs"),
  );
  const raw = "Unresolved\u0001HOST";
  await input.fill(raw);
  await page
    .getByRole("button", { name: "Create timeline row", exact: true })
    .click();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
  await expect(input).toHaveValue(raw);
  expect(f.creates).toHaveLength(1);
  expect(
    await queryViewRows(page, f.incidentId, timelineViewSchemaId),
  ).toHaveLength(0);
  await openRecoveryItem(page, /^Queued edit recovery ·/);
  await expect(
    page.getByTestId(workbookEditRecoveryDiscardButtonTestId()),
  ).toBeVisible();
  await expect(
    page.getByTestId(workbookEditRecoveryRetryButtonTestId()),
  ).toHaveCount(0);
  expect(f.creates).toHaveLength(1);
});

test("Timeline Create uncertain acceptance replays captured bytes without another logical row", async ({
  page,
}) => {
  const f = await openTimelineCreateFixture(page, "CREATE-UNCERTAIN");
  const receipts: { recordId: string; changeSetId: string }[] = [];
  let releaseReplay = () => {};
  const replayGate = new Promise<void>((resolve) => {
    releaseReplay = resolve;
  });
  let attempts = 0;
  await page.route(`**${f.createPath}`, async (route) => {
    const first = ++attempts === 1;
    if (!first) await replayGate;
    const response = await route.fetch();
    expect(response.status()).toBe(first ? 201 : 200);
    const body = await response.json();
    receipts.push({
      recordId: body.data.row.record_id,
      changeSetId: body.data.change_set_id,
    });
    if (first) await route.abort("connectionfailed");
    else await route.fulfill({ response });
  });
  try {
    const create = page.getByRole("button", {
      name: "Create timeline row",
      exact: true,
    });
    await create.click();
    await expect.poll(() => attempts).toBe(2);
    await expect(create).toBeDisabled();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
    await create.evaluate((button: HTMLButtonElement) => button.click());
    expect(f.creates).toHaveLength(2);
    expect(new Set(f.creates).size).toBe(1);
    releaseReplay();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    expect(receipts).toHaveLength(2);
    expect(receipts[1]).toEqual(receipts[0]);
    expect(
      await queryViewRows(page, f.incidentId, timelineViewSchemaId),
    ).toHaveLength(1);
    await expect(gridDraftRows(page, timelineViewSchemaId)).toHaveCount(1);
  } finally {
    releaseReplay();
    await page.unroute(`**${f.createPath}`);
  }
});

test("Timeline Add row and Evidence stay independent of Create and closed or viewer admission", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const f = await openTimelineCreateFixture(page, "CREATE-CONTROLS");
  const draft = page.getByTestId(draftCellTestId("timeline.date_entered_text"));
  await page
    .getByRole("button", {
      name: "Account and application navigation",
      exact: true,
    })
    .focus();
  await page
    .getByTestId(workbookAddRowButtonTestId(timelineViewSchemaId))
    .click();
  await expect(draft).toBeFocused();
  expect(f.creates).toHaveLength(0);
  const chooser = page.waitForEvent("filechooser");
  await page
    .getByRole("button", {
      name: "Attach evidence to draft timeline row",
      exact: true,
    })
    .click();
  await (await chooser).setFiles([]);
  expect(f.creates).toHaveLength(0);

  const email = uniqueEmail("create-viewer");
  const password = "TimelineCreateViewer!2026";
  const viewer = await createIncidentMemberUser(page, f.incidentId, {
    email,
    display_name: "Create viewer",
    initial_password: password,
    role: "viewer",
    is_deployment_admin: false,
    mfa_required: false,
  });
  const viewerPage = await openIncidentAsTrackedUser(browser, sessionTracker, {
    createdBy: "Timeline Create admission",
    email,
    incidentId: f.incidentId,
    password,
    purpose: "Verify viewer create admission",
    userId: viewer.user_id,
  });
  try {
    await expect(viewerPage.getByRole("grid")).toHaveAttribute(
      "aria-readonly",
      "true",
    );
    await expect(
      viewerPage.getByRole("button", {
        name: "Create timeline row",
        exact: true,
      }),
    ).toHaveCount(0);
    await expect(
      viewerPage.getByRole("button", {
        name: "Attach evidence to draft timeline row",
        exact: true,
      }),
    ).toHaveCount(0);
  } finally {
    await viewerPage.context().close();
  }
  const lifecycle = await currentLifecycle(page, f.incidentId);
  expect(
    (
      await lifecycleAction(page, f.incidentId, "closeIncident", {
        client_txn_id: uniqueTxn("create-close"),
        base_incident_version: lifecycle.incident_version,
        reason: "Verify closed create admission",
      })
    ).ok,
  ).toBe(true);
  await expect(page.getByRole("grid")).toHaveAttribute("aria-readonly", "true");
  await expect(
    page.getByRole("button", { name: "Create timeline row", exact: true }),
  ).toHaveCount(0);
  await expect(
    page.getByRole("button", {
      name: "Attach evidence to draft timeline row",
      exact: true,
    }),
  ).toHaveCount(0);
  expect(f.creates).toHaveLength(0);
});

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
  await page.keyboard.press("Enter");

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
    expectedSummary: "—",
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
    reviewerPage.getByTestId(gridRowTestId(timelineViewSchemaId, recordId)),
  ).toHaveAttribute(gridRowVersionAttribute, "2");

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
    reviewerPage.getByTestId(gridRowTestId(timelineViewSchemaId, recordId)),
  ).toHaveAttribute(gridRowVersionAttribute, "3");

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
    reviewerPage.getByTestId(gridRowTestId(timelineViewSchemaId, recordId)),
  ).toHaveAttribute(gridRowVersionAttribute, "4");
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
    observer.getByTestId(gridRowTestId(timelineViewSchemaId, recordId)),
  ).toHaveAttribute(gridRowVersionAttribute, "1");
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

  await expect(
    page.getByTestId(gridRowTestId(timelineViewSchemaId, recordId)),
  ).toHaveAttribute(gridRowVersionAttribute, "2");
  await expect(
    observer.getByTestId(gridRowTestId(timelineViewSchemaId, recordId)),
  ).toHaveAttribute(gridRowVersionAttribute, "2");
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
    observer.getByTestId(gridRowTestId(timelineViewSchemaId, recordId)),
  ).toHaveAttribute(gridRowVersionAttribute, "2");
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
    // Background settlement publishes recovery without opening its panel.
    await expect(recoveryEntry(page)).toHaveText("Recovery (1)");
    await openRecoveryItem(page, /^Timeline action ·/);
    const recovery = page.getByRole("region", {
      name: "Timeline action recovery",
      exact: true,
    });
    await expect(recovery).toContainText("Timeline action outcome unknown.");
    const before = await fetchFullRecordHistory(page, target.record_id);
    await expect(
      page
        .getByRole("region", { name: "Recovery navigation", exact: true })
        .locator(":scope > h2"),
    ).toBeFocused();
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
      .filter((unit) => unit.kind === "link");
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
    await expect(recoveryEntry(page)).toBeFocused();
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
      await openHistoryEventDetails(page, rollbackItem.history_item_ref);
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
