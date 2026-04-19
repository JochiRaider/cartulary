import { expect, test, type Page } from "@playwright/test";

import {
  apiBase,
  applyCookies,
  csrfCookieName,
  csrfHeaders,
  ensureAdminSession,
  requireCookie,
  sessionCookieName,
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";

test("E-2-01 creates an incident, bootstraps the creator as admin, and lands on the workbook surface", async ({
  page,
}) => {
  await ensureAdminSession(page);
  await page.goto("/");

  const incidentKey = uniqueIncidentKey("E201");
  await page.getByTestId("create-incident-key").fill(incidentKey);
  await page.getByTestId("create-incident-title").fill("Phase 2 E-2-01");
  await page.getByTestId("create-incident").click();

  await expect(page.getByTestId("phase2-status")).toHaveText("Created incident");
  await expect(page.getByTestId("current-incident-key")).toHaveText(incidentKey);
  await expect(page.getByTestId("current-incident-version")).toHaveText("1");
  await expect(page.getByTestId("session-memberships")).toContainText("admin");
  await expect(page.getByTestId("default-workbook-pref")).toHaveText("null");
  await expect(page.getByTestId("user-workbook-pref")).toHaveText("null");
});

test("E-2-02 shows incident discovery, direct retrieval, and promoted-field-only patching", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const incidentKey = uniqueIncidentKey("E202");
  const incidentId = await createIncident(page, incidentKey, "Phase 2 E-2-02");

  await openIncidentFromList(page, incidentId, incidentKey);
  await expect(page.getByTestId("current-incident-title")).toHaveText("Phase 2 E-2-02");

  await page.getByTestId("patch-tlp").fill("amber");
  await page.getByTestId("patch-current-phase").fill("containment");
  await page
    .getByTestId("patch-primary-external-case-ref")
    .fill("CASE-E202");
  await page.getByTestId("patch-incident").click();

  await expect(page.getByTestId("phase2-status")).toHaveText("Patched incident");
  await expect(page.getByTestId("current-incident-version")).toHaveText("2");
  await expect(page.getByTestId("patch-tlp")).toHaveValue("amber");
  await expect(page.getByTestId("patch-current-phase")).toHaveValue("containment");
  await expect(page.getByTestId("patch-primary-external-case-ref")).toHaveValue(
    "CASE-E202",
  );
  await expect(page.getByTestId("current-incident-title")).toHaveText("Phase 2 E-2-02");
});

test("E-2-03 lets admins manage memberships and denies the same actions to non-admin members", async ({
  browser,
  page,
}) => {
  await ensureAdminSession(page);
  const targetEmail = uniqueEmail("phase2-e203-member");
  const targetPassword = "Phase2E203Pass!";
  const incidentKey = uniqueIncidentKey("E203");
  const targetUser = await createLocalUser(page, {
    email: targetEmail,
    display_name: "Phase 2 E203 Member",
    initial_password: targetPassword,
  });
  const incidentId = await createIncident(page, incidentKey, "Phase 2 E-2-03");

  await openIncidentFromList(page, incidentId, incidentKey);
  await page.getByTestId("membership-email").fill(targetEmail);
  await page.getByTestId("membership-role").selectOption("viewer");
  await page.getByTestId("create-membership").click();

  await expect(page.getByTestId("phase2-status")).toHaveText("Created membership");
  await expect(page.getByTestId(`membership-row-${targetUser.user_id}`)).toContainText(
    "Phase 2 E203 Member",
  );

  await page
    .getByTestId(`membership-role-input-${targetUser.user_id}`)
    .selectOption("reviewer");
  await page.getByTestId(`patch-membership-${targetUser.user_id}`).click();
  await expect(page.getByTestId("phase2-status")).toHaveText("Patched membership");
  await expect(page.getByTestId(`membership-version-${targetUser.user_id}`)).toHaveText(
    "2",
  );

  const memberContext = await browser.newContext();
  const memberPage = await memberContext.newPage();
  await switchToLocalSession(memberPage, targetEmail, targetPassword);
  await openIncidentFromList(memberPage, incidentId, incidentKey);

  await memberPage.getByTestId("membership-email").fill(uniqueEmail("phase2-e203-denied"));
  await memberPage.getByTestId("membership-role").selectOption("viewer");
  await memberPage.getByTestId("create-membership").click();
  await expect(memberPage.getByTestId("last-error-code")).toHaveText(
    "authorization_denied",
  );

  await memberPage
    .getByTestId(`membership-role-input-${targetUser.user_id}`)
    .selectOption("admin");
  await memberPage.getByTestId(`patch-membership-${targetUser.user_id}`).click();
  await expect(memberPage.getByTestId("last-error-code")).toHaveText(
    "authorization_denied",
  );

  await memberPage.getByTestId(`delete-membership-${targetUser.user_id}`).click();
  await expect(memberPage.getByTestId("last-error-code")).toHaveText(
    "authorization_denied",
  );
  await memberContext.close();

  await page.getByTestId(`delete-membership-${targetUser.user_id}`).click();
  await expect(page.getByTestId("phase2-status")).toHaveText("Deleted membership");
  await expect(page.getByTestId(`membership-row-${targetUser.user_id}`)).toHaveCount(0);
});

test("E-2-04 rejects unknown or forbidden top-level members with route-owned errors", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const incidentKey = uniqueIncidentKey("E204");
  const incidentId = await createIncident(page, incidentKey, "Phase 2 E-2-04");

  await page.goto("/");
  await page.getByTestId("probe-invalid-create-initial-memberships").click();
  await expect(page.getByTestId("last-error-code")).toHaveText("invalid_incident_create");
  await expectLastErrorDetails(page, [
    '"field": "initial_memberships"',
    '"reason_code": "initial_memberships_not_supported"',
  ]);

  await page.getByTestId("probe-invalid-create-unknown").click();
  await expect(page.getByTestId("last-error-code")).toHaveText("invalid_incident_create");
  await expectLastErrorDetails(page, [
    '"field": "unexpected"',
    '"reason_code": "unknown_top_level_member"',
  ]);

  await openIncidentFromList(page, incidentId, incidentKey);

  await page.getByTestId("probe-invalid-patch-title").click();
  await expect(page.getByTestId("last-error-code")).toHaveText("invalid_incident_patch");
  await expectLastErrorDetails(page, [
    '"field": "title"',
    '"reason_code": "forbidden_field"',
  ]);

  await page.getByTestId("probe-invalid-patch-unknown").click();
  await expect(page.getByTestId("last-error-code")).toHaveText("invalid_incident_patch");
  await expectLastErrorDetails(page, [
    '"field": "unknown"',
    '"reason_code": "unknown_top_level_member"',
  ]);
});

test("E-2-05 allows zero-membership extension discovery and rejects singleton pagination semantics", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const zeroMemberEmail = uniqueEmail("phase2-e205");
  const zeroMemberPassword = "Phase2E205Pass!";
  await createLocalUser(page, {
    email: zeroMemberEmail,
    display_name: "Phase 2 E205 User",
    initial_password: zeroMemberPassword,
  });

  await switchToLocalSession(page, zeroMemberEmail, zeroMemberPassword);
  await page.goto("/");

  const extensionRows = page.locator('[data-testid^="extension-"]');
  await expect(extensionRows).toHaveCount(5);
  await expect(extensionRows.nth(0)).toHaveText("enterprise_authentication");
  await expect(extensionRows.nth(1)).toHaveText("import");
  await expect(extensionRows.nth(4)).toHaveText("snapshot_reporting");

  await page.getByTestId("probe-extensions-pagination").click();
  await expect(page.getByTestId("last-error-code")).toHaveText(
    "invalid_pagination_request",
  );
  await expect(page.getByTestId("last-probe-status")).toHaveText("400");
});

test("E-2-06 shows reserved-family 404 precedence while base and outside-reserved paths keep their ordinary dispatch", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const zeroMemberEmail = uniqueEmail("phase2-e206");
  const zeroMemberPassword = "Phase2E206Pass!";
  await createLocalUser(page, {
    email: zeroMemberEmail,
    display_name: "Phase 2 E206 User",
    initial_password: zeroMemberPassword,
  });

  await switchToLocalSession(page, zeroMemberEmail, zeroMemberPassword);
  await page.goto("/");

  await page.getByTestId("probe-base-route").click();
  await expect(page.getByTestId("last-probe-status")).toHaveText("200");
  await expect(page.getByTestId("last-probe-payload")).toContainText("ready");

  await page.getByTestId("probe-reserved-root").click();
  await expect(page.getByTestId("last-error-code")).toHaveText(
    "extension_profile_not_claimed",
  );
  await expect(page.getByTestId("last-probe-payload")).toContainText("import");

  await page.getByTestId("probe-reserved-descendant").click();
  await expect(page.getByTestId("last-error-code")).toHaveText(
    "extension_profile_not_claimed",
  );
  await expect(page.getByTestId("last-probe-payload")).toContainText(
    "enterprise_authentication",
  );

  await page.getByTestId("probe-outside-reserved").click();
  await expect(page.getByTestId("last-probe-status")).toHaveText("404");
  await expect(page.getByTestId("last-error-code")).toHaveText("");
  await expect(page.getByTestId("last-probe-payload")).not.toContainText(
    "extension_profile_not_claimed",
  );
});

async function switchToLocalSession(page: Page, email: string, password: string) {
  await page.context().clearCookies();
  const loginResponse = await page.request.post(`${apiBase}/api/v1/auth/login`, {
    data: {
      username: email,
      password,
    },
  });
  expect(loginResponse.ok()).toBeTruthy();
  await applyCookies(
    page,
    requireCookie(loginResponse, sessionCookieName),
    requireCookie(loginResponse, csrfCookieName),
  );
}

async function createIncident(page: Page, incidentKey: string, title: string) {
  const response = await page.request.post(`${apiBase}/api/v1/incidents`, {
    headers: await csrfHeaders(page),
    data: {
      client_txn_id: uniqueTxn("phase2-incident"),
      incident_key: incidentKey,
      title,
    },
  });
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as { data: { incident_id: string } };
  return body.data.incident_id;
}

async function openIncidentFromList(
  page: Page,
  incidentId: string,
  incidentKey: string,
) {
  await expect
    .poll(async () => {
      const response = await page.request.get(`${apiBase}/api/v1/incidents`);
      expect(response.ok()).toBeTruthy();
      const body = (await response.json()) as {
        data: { incidents: Array<{ incident_id: string }> };
      };
      return body.data.incidents.some(
        (incident) => incident.incident_id === incidentId,
      );
    })
    .toBe(true);

  await page.goto("/");
  await expect(page.getByTestId(`incident-row-${incidentId}`)).toHaveText(
    incidentKey,
  );
  await page.getByTestId(`select-incident-${incidentId}`).click();
  await expect(page.getByTestId("current-incident-id")).toHaveText(incidentId);
}

async function createLocalUser(
  page: Page,
  options: {
    email: string;
    display_name: string;
    initial_password: string;
    is_deployment_admin?: boolean;
  },
) {
  const response = await page.request.post(`${apiBase}/api/v1/users`, {
    headers: await csrfHeaders(page),
    data: {
      client_txn_id: uniqueTxn("phase2-user"),
      auth_kind: "local",
      email: options.email,
      display_name: options.display_name,
      initial_password: options.initial_password,
      mfa_required: false,
      is_deployment_admin: options.is_deployment_admin ?? false,
    },
  });
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: { user_id: string } }).data;
}

async function expectLastErrorDetails(page: Page, fragments: string[]) {
  const details = page.getByTestId("last-error-details");
  for (const fragment of fragments) {
    await expect(details).toContainText(fragment);
  }
}
