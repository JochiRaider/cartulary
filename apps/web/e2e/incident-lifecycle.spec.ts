import type { CloseIncidentRequest } from "@cartulary/protocol-ts/http";
import {
  incidentAdministrationTestId,
  incidentLandingTestId,
  rowCellTestId,
  saveStateTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import {
  openIncidentAsTrackedUser,
  openIncidentFromLanding,
} from "./pages/incidentDirectory";
import { auditBrowserBarrier } from "./support/administrativeAudit";
import { csrfHeaders } from "./support/auth/browserSession";
import {
  confirmLifecycle,
  currentLifecycle,
  lifecycleAction,
  openLifecycle,
} from "./support/incidentLifecycle";
import { openMembershipManagement } from "./support/incidentMembershipManagement";
import { openMetadata } from "./support/incidentMetadata";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import { createViewRow } from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";
import { ensureTimelineGridTargetVisible } from "./support/workbook/rowMutations";
import { createSavedView } from "./support/workbook/savedViews";

test("lifecycle recovers a committed lost Close after another actor reopens without publishing its historical receipt", async ({
  workerAdminPage: page,
  browser,
  sessionTracker,
}, testInfo) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("LC-REPLAY"),
    "Lifecycle receipt recovery",
  );
  const second = await createIncidentMemberUser(page, incidentId, {
    email: uniqueEmail("lc-admin"),
    display_name: "Second lifecycle admin",
    initial_password: "LifecycleAdmin1!",
    role: "admin",
    mfa_required: false,
    is_deployment_admin: false,
  });
  const actorB = await openIncidentAsTrackedUser(browser, sessionTracker, {
    createdBy: "lifecycle replay second actor",
    email: second.email,
    incidentId,
    password: second.initial_password,
    purpose: "lifecycle current state divergence",
    userId: second.user_id,
  });
  try {
    await openIncidentFromLanding(page, incidentId);
    const panel = await openLifecycle(page);
    const requests: CloseIncidentRequest[] = [];
    const responses: unknown[] = [];
    await page.route(
      `**/api/v1/incidents/${incidentId}/close`,
      async (route) => {
        requests.push(route.request().postDataJSON());
        const response = await route.fetch();
        expect(response.status()).toBe(200);
        responses.push(await response.json());
        if (requests.length === 1) await route.abort("failed");
        else await route.fulfill({ response });
      },
    );
    await confirmLifecycle(page, "Close", "  e\u0301\nOriginal close reason  ");
    await expect(panel.getByText(/Close result is uncertain/u)).toBeVisible();
    expect((await currentLifecycle(page, incidentId)).status).toBe("closed");
    expect(
      (
        await lifecycleAction(actorB, incidentId, "reopenIncident", {
          base_incident_version: 2,
          client_txn_id: uniqueTxn("lc-other-reopen"),
          reason: "Another actor resumes response",
        })
      ).ok,
    ).toBe(true);
    await panel
      .getByRole("button", { name: "Refresh current incident", exact: true })
      .click();
    await expect(
      panel.getByText(/Current accepted state: Active · Version 3/u),
    ).toBeVisible();
    await panel
      .getByRole("textbox", { name: "Reason", exact: true })
      .fill("Newer reason must remain local");
    await panel
      .getByRole("button", { name: "Replay original action", exact: true })
      .click();
    await expect(
      panel.getByText("Close confirmed.", { exact: true }),
    ).toBeVisible();
    expect(requests).toHaveLength(2);
    expect(requests[1]).toEqual(requests[0]);
    expect(responses[1]).toMatchObject({
      data: { status: "closed", incident_version: 2 },
    });
    await expect(
      panel.getByText(/Current accepted state: Active · Version 3/u),
    ).toBeVisible();
    await expect(
      page.getByTestId(incidentAdministrationTestId("summary-status")),
    ).toHaveText("active");
    await expect(
      panel.getByRole("textbox", { name: "Reason", exact: true }),
    ).toHaveValue("Newer reason must remain local");
    const original = requests[0];
    if (!original) throw new Error("Expected original lifecycle request");
    const conflict = await lifecycleAction(page, incidentId, "closeIncident", {
      ...original,
      reason: "Divergent key payload",
    });
    expect(conflict).toMatchObject({
      ok: false,
      status: 409,
      payload: { error: { code: "client_txn_conflict" } },
    });
    expect((await currentLifecycle(page, incidentId)).incident_version).toBe(3);
    await testInfo.attach("real-lifecycle-replay", {
      body: JSON.stringify({
        requests,
        receipts: responses,
        current: await currentLifecycle(page, incidentId),
        divergent: conflict,
      }),
      contentType: "application/json",
    });
  } finally {
    await actorB.context().close();
  }
});

test("lifecycle confirms a real action independently of failed reads and preserves newer input during delayed acknowledgement", async ({
  workerAdminPage: page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("LC-ACK"),
    "Confirmed lifecycle action",
  );
  await openIncidentFromLanding(page, incidentId);
  const panel = await openLifecycle(page);
  const gate = auditBrowserBarrier();
  let writes = 0;
  let failRead = false;
  let failAuthorization = false;
  await page.route("**/api/v1/auth/session", async (route) => {
    if (failAuthorization)
      await route.fulfill({
        status: 503,
        json: {
          error: { code: "internal_error" },
          meta: { request_id: "lc-auth-observation" },
        },
      });
    else await route.continue();
  });
  await page.route(
    `**/api/v1/incidents/${incidentId}/reopen`,
    async (route) => {
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      failAuthorization = true;
      await route.fulfill({ response });
    },
  );
  await page.route(`**/api/v1/incidents/${incidentId}/close`, async (route) => {
    ++writes;
    const response = await route.fetch();
    expect(response.status()).toBe(200);
    await gate.promise;
    await route.fulfill({ response });
  });
  await page.route(`**/api/v1/incidents/${incidentId}`, async (route) => {
    if (failRead && route.request().method() === "GET")
      await route.fulfill({
        status: 503,
        json: {
          error: { code: "internal_error" },
          meta: { request_id: "lc-read-failed" },
        },
      });
    else await route.continue();
  });
  await confirmLifecycle(page, "Close", "Close captured reason");
  await expect.poll(() => writes).toBe(1);
  await panel
    .getByRole("textbox", { name: "Reason", exact: true })
    .fill("  Newer\nlocal work  ");
  failRead = true;
  gate.release();
  await expect(
    panel.getByText(
      /The action is confirmed, but current state could not be refreshed/u,
    ),
  ).toBeVisible();
  await expect(
    panel.getByRole("textbox", { name: "Reason", exact: true }),
  ).toHaveValue("  Newer\nlocal work  ");
  await expect(
    panel.getByRole("button", { name: "Replay original action", exact: true }),
  ).toHaveCount(0);
  failRead = false;
  await panel
    .getByRole("button", { name: "Refresh current incident", exact: true })
    .click();
  await expect(
    panel.getByText(/Current accepted state: Closed, read-only · Version 2/u),
  ).toBeVisible();
  expect(writes).toBe(1);
  await confirmLifecycle(page, "Reopen", "Fresh explicit reopen");
  await expect(
    panel.getByText("Reopen confirmed.", { exact: true }),
  ).toBeVisible();
  await expect(
    panel.getByText(
      /The action is confirmed, but current state could not be refreshed/u,
    ),
  ).toBeVisible();
  failAuthorization = false;
  await panel
    .getByRole("button", { name: "Refresh current incident", exact: true })
    .click();
  await expect(
    panel.getByText(/Current accepted state: Active · Version 3/u),
  ).toBeVisible();
});

test("lifecycle validates normalized reason boundaries and requires fresh review after a real version conflict", async ({
  workerAdminPage: page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("LC-REASON"),
    "Lifecycle reason boundary",
  );
  await openIncidentFromLanding(page, incidentId);
  const panel = await openLifecycle(page);
  await confirmLifecycle(page, "Close", "😀".repeat(4097));
  await expect(
    panel.getByRole("textbox", { name: "Reason", exact: true }),
  ).toHaveAttribute("aria-invalid", "true");
  await panel
    .getByRole("textbox", { name: "Reason", exact: true })
    .fill("  " + "e\u0301".repeat(4096) + "  ");
  await panel
    .getByRole("button", { name: "Close incident", exact: true })
    .click();
  const patch = await publicHttpOperation({
    request: atJsonOrigin(page.request, apiBase),
    headers: await csrfHeaders(page),
    operationID: "patchIncident",
    pathParameters: { incident_id: incidentId },
    body: { base_incident_version: 1, severity: "Concurrent metadata" },
  });
  expect(patch.ok).toBe(true);
  await panel
    .getByRole("button", { name: "Confirm Close incident", exact: true })
    .click();
  await expect(panel.getByText(/The incident version changed/u)).toBeVisible();
  await panel
    .getByRole("button", { name: "Review current incident", exact: true })
    .click();
  await expect(panel.getByText(/Reviewed active, version 2/u)).toBeVisible();
  await panel
    .getByRole("button", { name: "Confirm Close incident", exact: true })
    .click();
  await expect(
    panel.getByText("Close confirmed.", { exact: true }),
  ).toBeVisible();
  const illegal = await lifecycleAction(page, incidentId, "closeIncident", {
    base_incident_version: 3,
    client_txn_id: uniqueTxn("lc-illegal"),
    reason: "Already closed",
  });
  expect(illegal).toMatchObject({
    ok: false,
    status: 409,
    payload: {
      error: {
        code: "illegal_transition",
        details: { reason_code: "incident_already_closed" },
      },
    },
  });
  await confirmLifecycle(page, "Reopen", "😀".repeat(4096));
  await expect(
    panel.getByText("Reopen confirmed.", { exact: true }),
  ).toBeVisible();
});

test("lifecycle metadata and membership retained work receive sequential departure review with cancellation and history", async ({
  workerAdminPage: page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("LC-LEAVE"),
    "Lifecycle departure owners",
  );
  await openIncidentFromLanding(page, incidentId);
  const members = await openMembershipManagement(page);
  await members
    .getByRole("button", { name: "Add existing account", exact: true })
    .click();
  await members
    .getByLabel("User email", { exact: true })
    .fill("retained@example.test");
  const metadata = await openMetadata(page);
  await metadata
    .getByLabel("Severity", { exact: true })
    .fill("Retained severity");
  const panel = await openLifecycle(page);
  await panel
    .getByRole("textbox", { name: "Reason", exact: true })
    .fill("Retained lifecycle reason");
  await page
    .getByRole("button", { name: "Close incident controls", exact: true })
    .click();
  await openLifecycle(page);
  await expect(
    panel.getByRole("textbox", { name: "Reason", exact: true }),
  ).toHaveValue("Retained lifecycle reason");
  await page.getByLabel("Account and application navigation").click();
  await page.getByRole("menuitem", { name: "Incidents", exact: true }).click();
  const membershipDeparture = page.getByRole("dialog", {
    name: /Leave membership work/u,
  });
  await expect(membershipDeparture).toBeVisible();
  await expect(
    page.getByRole("dialog", { name: /Leave .* work/u }),
  ).toHaveCount(1);
  await membershipDeparture
    .getByRole("button", { name: "Leave and forget recovery", exact: true })
    .click();
  const metadataDeparture = page.getByRole("dialog", {
    name: "Leave metadata work?",
    exact: true,
  });
  await expect(metadataDeparture).toBeVisible();
  await expect(
    page.getByRole("dialog", { name: /Leave .* work/u }),
  ).toHaveCount(1);
  await metadataDeparture
    .getByRole("button", { name: "Leave and forget recovery", exact: true })
    .click();
  const lifecycleDeparture = page.getByRole("dialog", {
    name: "Leave lifecycle work?",
    exact: true,
  });
  await expect(lifecycleDeparture).toBeVisible();
  await expect(
    page.getByRole("dialog", { name: /Leave .* work/u }),
  ).toHaveCount(1);
  await page.keyboard.press("Escape");
  await expect(page).toHaveURL(new RegExp(incidentId));
  await page.goBack();
  await expect(lifecycleDeparture).toBeVisible();
  await lifecycleDeparture
    .getByRole("button", { name: "Leave and forget recovery", exact: true })
    .click();
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
});

test("lifecycle closure retains rejected workbook work and allowed configuration while reopening requires fresh source action", async ({
  workerAdminPage: page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("LC-WORKBOOK"),
    "Closed workbook recovery",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("lc-source-seed"),
    "timeline.activity_synopsis_text": "Accepted source text",
  });
  await openIncidentFromLanding(page, incidentId);
  const retainedMetadata = await openMetadata(page);
  await retainedMetadata
    .getByLabel("Severity", { exact: true })
    .fill("Retained investigation severity");
  await page
    .getByRole("button", { name: "Close incident controls", exact: true })
    .click();
  const cell = page.getByTestId(
    rowCellTestId(row.record_id, "timeline.activity_synopsis_text"),
  );
  await ensureTimelineGridTargetVisible(
    page,
    rowCellTestId(row.record_id, "timeline.activity_synopsis_text"),
  );
  await expect(cell).toBeVisible();
  const patchGate = auditBrowserBarrier();
  await page.route(`**/api/v1/records/${row.record_id}`, async (route) => {
    if (route.request().method() === "PATCH") await patchGate.promise;
    await route.continue();
  });
  const patches: unknown[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/records/${row.record_id}`)
    )
      patches.push(request.postDataJSON());
  });
  await cell.click();
  const editor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.activity_synopsis_text",
      recordId: row.record_id,
      surface: "grid",
    }),
  );
  await editor.fill("Rejected local source draft");
  await editor.press("Tab");
  await expect.poll(() => patches.length).toBe(1);
  expect(
    (
      await lifecycleAction(page, incidentId, "closeIncident", {
        base_incident_version: 1,
        client_txn_id: uniqueTxn("lc-server-close"),
        reason: "Closure by another surface",
      })
    ).ok,
  ).toBe(true);
  patchGate.release();
  const discard = page.getByRole("button", {
    name: "Discard blocked edit",
    exact: true,
  });
  await openRecoveryItem(page, /^Queued edit recovery ·/);
  await expect(discard).toBeVisible();
  await page
    .getByRole("button", { name: "Close recovery", exact: true })
    .click();
  const panel = await openLifecycle(page);
  await expect(
    panel.getByText(/Current accepted state: Closed, read-only/u),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Close incident controls", exact: true })
    .click();
  await expect(page.getByRole("grid")).toHaveAttribute("aria-readonly", "true");
  await expect(
    page.getByText("Closed, read-only", { exact: true }),
  ).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).not.toHaveText(
    "Closed, read-only",
  );
  await expect(
    page.getByText("Authentication required", { exact: true }),
  ).toHaveCount(0);
  const saved = await createSavedView(page, incidentId, {
    display_name: "Closed incident saved view",
    view_schema_id: timelineViewSchemaId,
    scope: "shared",
  });
  expect(saved.saved_view_id).toBeTruthy();
  const request = atJsonOrigin(page.request, apiBase);
  const headers = await csrfHeaders(page);
  expect(
    (
      await publicHttpOperation({
        request,
        headers,
        operationID: "putIncidentDefaultWorkbookPreferences",
        pathParameters: { incident_id: incidentId },
        body: {
          default_sheet_ref: { kind: "view_schema", id: timelineViewSchemaId },
        },
      })
    ).ok,
  ).toBe(true);
  expect(
    (
      await publicHttpOperation({
        request,
        headers,
        operationID: "putCurrentUserWorkbookPreferences",
        pathParameters: { incident_id: incidentId },
        body: {
          home_sheet_ref: { kind: "view_schema", id: timelineViewSchemaId },
        },
      })
    ).ok,
  ).toBe(true);
  await createIncidentMemberUser(page, incidentId, {
    email: uniqueEmail("lc-closed-reader"),
    display_name: "Closed incident reader",
    initial_password: "LifecycleReader1!",
    role: "viewer",
    mfa_required: false,
    is_deployment_admin: false,
  });
  const metadata = await openMetadata(page);
  await expect(metadata.getByText(/This incident is closed/u)).toBeVisible();
  await expect(
    metadata.getByText("Retained investigation severity", { exact: true }),
  ).toBeVisible();
  await expect(
    metadata.getByText(/Your local changes are retained/u),
  ).toBeVisible();
  await openLifecycle(page);
  await confirmLifecycle(page, "Reopen", "Fresh user-authorized reopening");
  await expect(
    panel.getByText("Reopen confirmed.", { exact: true }),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Close incident controls", exact: true })
    .click();
  await expect(page.getByRole("grid")).toHaveAttribute(
    "aria-readonly",
    "false",
  );
  await openRecoveryItem(page, /^Queued edit recovery ·/);
  await expect(discard).toBeVisible();
  expect(patches).toHaveLength(1);
  await expect(
    page.getByText("Authentication required", { exact: true }),
  ).toHaveCount(0);
  await discard.click();
  await expect(discard).toHaveCount(0);
  await page
    .getByRole("button", { name: "Close recovery", exact: true })
    .click();
  await cell.click();
  await editor.fill("Fresh source action after reopening");
  await editor.press("Tab");
  await expect(cell).toHaveText("Fresh source action after reopening");
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  expect(patches).toHaveLength(2);
});
