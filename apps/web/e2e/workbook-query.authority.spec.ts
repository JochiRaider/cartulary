import {
  authTestId,
  gridShellTestId,
  incidentLandingTestId,
} from "@cartulary/ui-contracts";
import { notesViewSchemaId } from "@cartulary/view-contracts";

import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { currentLifecycle, lifecycleAction } from "./support/incidentLifecycle";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { responseBarrier } from "./support/runtime/cleanup";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import {
  awaitResponseBarrier,
  installOwnedRoute,
  installPublicErrorResponse,
} from "./support/transport/requestInterception";

import { observeQuery, seed } from "./support/workbook/queryBrowsing";

test("Workbook continuation revalidates role and membership and fences a late authorized page", async ({
  page,
  testActor,
  testActorRequest,
  sessionTracker,
}) => {
  test.setTimeout(180_000);
  const view = notesViewSchemaId;
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-AUTH"),
    "Workbook continuation authority",
  );
  await seed(page, incident, view, 405, testActor.user_id);
  const member = await createIncidentMemberUser(page, incident, {
    email: uniqueEmail("wqc-authority"),
    display_name: "Workbook paging analyst",
    initial_password: "WorkbookPaging1!",
    role: "editor",
    is_deployment_admin: false,
    mfa_required: false,
  });
  await sessionTracker.loginTrackedUser(page, {
    createdBy: "wqc-authority",
    email: member.email,
    password: member.initial_password,
    purpose: "live query authority",
    userId: member.user_id,
  });
  const reads = await observeQuery(page, incident, view);
  const sockets = installIncidentSocketMonitor(page, incident);
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`,
  );
  await sockets.waitForAcceptedSocket();
  const controls = page.getByRole("group", { name: "Workbook browsing" });
  const more = controls.getByRole("button", { name: "Load more", exact: true });
  await expect(controls).toContainText("100 records loaded; more available.");
  await more.click();
  await expect(controls).toContainText("200 records loaded; more available.");
  const membershipPath = `/api/v1/incidents/${incident}/memberships/${member.user_id}`;
  expect(
    (
      await testActorRequest.patch(membershipPath, {
        data: { base_membership_version: 1, role: "viewer" },
      })
    ).ok(),
  ).toBe(true);
  await more.click();
  await expect(controls).toContainText("300 records loaded; more available.");
  const later = reads.at(-1)?.response.data.rows[0];
  if (!later) throw new Error("Missing live-authorized later row");
  const deniedWrite = await publicHttpOperation({
    operationID: "patchRecord",
    pathParameters: { record_id: later.record_id },
    headers: await csrfHeaders(page),
    request: atJsonOrigin(page.request, apiBase),
    body: {
      view_schema_id: view,
      base_row_version: later.row_version,
      client_txn_id: uniqueTxn("wqc-denied"),
      changes: [{ field_key: "note.title", value: "Must not change" }],
    },
  });
  expect(deniedWrite.ok).toBe(false);
  expect(deniedWrite.status).toBe(403);
  let release: () => void = () => {};
  let held = false;
  const gateBarrier = responseBarrier(page);
  release = gateBarrier.release;
  const gate = gateBarrier.promise;
  await installOwnedRoute(
    page,
    `**/incidents/${incident}/views/${view}/query`,
    async (route) => {
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      held = true;
      if (!(await awaitResponseBarrier(route, gate))) return;
      await route.fulfill({ response });
    },
    { times: 1 },
  );
  await more.click();
  await expect.poll(() => held).toBe(true);
  expect(
    (
      await testActorRequest.delete(membershipPath, {
        data: { base_membership_version: 2 },
      })
    ).status(),
  ).toBe(204);
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  release();
  await expect(controls).toHaveCount(0);
  await expect(page.getByTestId(gridShellTestId(view))).toHaveCount(0);
  await expect(page).not.toHaveURL(/incident_id=/u);
});

test("Workbook expiry clears continuation and account replacement starts a readable closed incident without an old cursor", async ({
  page,
  testActor,
  sessionTracker,
}) => {
  test.setTimeout(180_000);
  const view = notesViewSchemaId;
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-SESSION"),
    "Workbook paging session lifetime",
  );
  await seed(page, incident, view, 305, testActor.user_id);
  const members = [];
  for (const label of ["first", "replacement"])
    members.push(
      await createIncidentMemberUser(page, incident, {
        email: uniqueEmail(`wqc-${label}`),
        display_name: `Workbook ${label}`,
        initial_password: "WorkbookPaging1!",
        role: "viewer",
        is_deployment_admin: false,
        mfa_required: false,
      }),
    );
  const before = await currentLifecycle(page, incident);
  expect(
    (
      await lifecycleAction(page, incident, "closeIncident", {
        client_txn_id: uniqueTxn("wqc-close"),
        base_incident_version: before.incident_version,
        reason: "Closed readable continuation evidence",
      })
    ).ok,
  ).toBe(true);
  const first = members[0],
    second = members[1];
  if (!first || !second) throw new Error("Missing account fixtures");
  await sessionTracker.loginTrackedUser(page, {
    createdBy: "wqc-session",
    email: first.email,
    password: first.initial_password,
    purpose: "closed workbook browsing",
    userId: first.user_id,
  });
  const reads = await observeQuery(page, incident, view);
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`,
  );
  const controls = page.getByRole("group", { name: "Workbook browsing" });
  await expect(controls).toContainText("100 records loaded; more available.");
  await controls
    .getByRole("button", { name: "Load more", exact: true })
    .click();
  await expect(controls).toContainText("200 records loaded; more available.");
  const sessionExpiry = await installPublicErrorResponse(page, {
    method: "GET",
    path: "/api/v1/auth/session",
    code: "session_required",
    status: 401,
  });
  const expiry = await installPublicErrorResponse(page, {
    method: "POST",
    path: `/api/v1/incidents/${incident}/views/${view}/query`,
    code: "session_required",
    status: 401,
  });
  // The local response exercises retirement; actual expiry has its own HTTP case.
  if (await controls.isVisible())
    await controls
      .getByRole("button", { name: "Refresh", exact: true })
      .click();
  await expect(page.getByTestId(authTestId("shell"))).toBeVisible({
    timeout: 30_000,
  });
  await expect(controls).toHaveCount(0);
  await expect(page.getByTestId(gridShellTestId(view))).toHaveCount(0);
  expect(expiry.hitCount()).toBe(1);
  expect(sessionExpiry.hitCount()).toBe(1);
  await expiry.dispose();
  await sessionExpiry.dispose();
  const beforeReplacement = reads.length;
  await sessionTracker.loginTrackedUser(page, {
    createdBy: "wqc-session",
    email: second.email,
    password: second.initial_password,
    purpose: "replacement account browsing",
    userId: second.user_id,
    recovery: true,
  });
  // Account replacement may return to landing; enter the same readable incident explicitly.
  if (!new URL(page.url()).searchParams.has("incident_id"))
    await page.goto(
      `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`,
    );
  await expect(controls).toContainText("100 records loaded; more available.");
  expect(
    reads
      .slice(beforeReplacement)
      .every((read) => read.request.cursor_token === undefined),
  ).toBe(true);
  await controls
    .getByRole("button", { name: "Load more", exact: true })
    .click();
  await expect(controls).toContainText("200 records loaded; more available.");
});
