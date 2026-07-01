import {
  currentIncidentRoleTestId,
  type IncidentControlsSection,
  incidentControlsMenuItemTestId,
  incidentControlsMenuTestId,
  incidentControlsPanelTestId,
  incidentControlsTriggerTestId,
  incidentMembershipDeleteButtonTestId,
  incidentMembershipPatchButtonTestId,
  incidentMembershipRoleDisplayTestId,
  incidentMembershipRoleInputTestId,
  incidentMembershipRowTestId,
  landingAdminMenuItemTestId,
  landingIncidentCardTestId,
  phase1AccountTestId,
  phase1AdminTestId,
  phase1AuthTestId,
  phase1ErrorCodeTestId,
  phase1LandingTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import type { APIRequestContext, Page } from "@playwright/test";

import {
  authenticatedRequestContextFromStorageState,
  createLocalUser,
  deploymentAdminMutationClient,
  loadUser,
  patchUser,
  resetUserTotp,
  revokeAllSessions,
  withOnlyActiveDeploymentAdmin,
} from "./authRuntime";
import { expect, restoreTrackedStorageState, test } from "./fixtures";
import {
  apiBase,
  csrfHeaders,
  enrollTotpViaBootstrap,
  generateTotpCode,
  sessionCookieName,
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import { Phase1Page } from "./phase1Page";

async function openIncidentControls(
  page: Page,
  section: IncidentControlsSection = "summary",
) {
  await page.getByLabel("Account and application navigation").click();
  const trigger = page.getByTestId(incidentControlsTriggerTestId());
  await expect(trigger).toHaveAttribute("aria-haspopup", "menu");
  await trigger.click();
  await expect(page.getByTestId(incidentControlsMenuTestId())).toBeVisible();
  const menuItem = page.getByTestId(incidentControlsMenuItemTestId(section));
  await expect(menuItem).toHaveAttribute("role", "menuitem");
  await menuItem.click();
  await expect(page.getByTestId(incidentControlsPanelTestId())).toBeVisible();
}

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

async function expectLandingAccountSession(page: Page) {
  const phase1 = new Phase1Page(page);
  await phase1.openAccountSettings("account-security");
  await expect(
    page.getByTestId(phase1AccountTestId("refresh-state")),
  ).toBeVisible();
  await expect(page.getByTestId(phase1AccountTestId("logout"))).toBeVisible();
  await expect(
    page.getByTestId(phase1AccountTestId("password-current")),
  ).toBeVisible();
}

test("E-1-01 signs in as a local user and inspects the ordinary session surface", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const phase1 = new Phase1Page(page);
  const email = uniqueEmail("phase1-e101");
  const password = "Phase1E101Pass!";
  const user = await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Phase 1 E101",
    initial_password: password,
    mfa_required: false,
  });

  await clearBrowserSession(page);
  await phase1.goto();
  await expect(
    page.getByTestId(phase1AuthTestId("login-username")),
  ).toBeVisible();
  const loginResponse = waitForPublicAPIResponse(page, {
    method: "POST",
    path: "/api/v1/auth/login",
    status: 200,
  });
  const sessionResponse = waitForPublicAPIResponse(page, {
    method: "GET",
    path: "/api/v1/auth/session",
    status: 200,
  });
  const credentialStateResponse = waitForPublicAPIResponse(page, {
    method: "GET",
    path: "/api/v1/auth/credential-state",
    status: 200,
  });
  const incidentListResponse = waitForPublicAPIResponse(page, {
    method: "GET",
    path: "/api/v1/incidents",
    status: 200,
  });
  await phase1.login(email, password);
  await Promise.all([
    loginResponse,
    sessionResponse,
    credentialStateResponse,
    incidentListResponse,
  ]);

  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email,
    purpose: "phase1 e101 successful ordinary login",
    userId: user.user_id,
  });
});

test("E-1-02 requires MFA on the ordinary login surface, rejects wrong codes, and accepts a valid TOTP code", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const phase1 = new Phase1Page(page);
  const email = uniqueEmail("phase1-e102");
  const password = "Phase1E102Pass!";
  const user = await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Phase 1 E102",
    initial_password: password,
    mfa_required: true,
  });
  const secretBase32 = await enrollTotpViaBootstrap(email, password);

  await clearBrowserSession(page);
  await phase1.goto();
  const missingTotpResponse = waitForPublicAPIResponse(page, {
    method: "POST",
    path: "/api/v1/auth/login",
    status: 401,
  });

  await phase1.login(email, password);
  await missingTotpResponse;
  await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "mfa_required",
  );
  await expect(
    page.getByTestId(phase1AuthTestId("login-totp-code")),
  ).toBeVisible();
  await expect(page.getByTestId(phase1ErrorCodeTestId("auth"))).toHaveText("");
  expect(await hasSessionCookie(page)).toBeFalsy();

  const wrongTotpResponse = waitForPublicAPIResponse(page, {
    method: "POST",
    path: "/api/v1/auth/login",
    status: 401,
  });
  await phase1.login(email, password, "000000");
  await wrongTotpResponse;
  await expect(page.getByTestId(phase1ErrorCodeTestId("auth"))).toHaveText(
    "The verification code is incorrect or expired.",
  );
  expect(await hasSessionCookie(page)).toBeFalsy();

  const validTotpResponse = waitForPublicAPIResponse(page, {
    method: "POST",
    path: "/api/v1/auth/login",
    status: 200,
  });
  await phase1.login(email, password, generateTotpCode(secretBase32));
  await validTotpResponse;
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email,
    purpose: "phase1 e102 successful ordinary login",
    userId: user.user_id,
  });
});

test("E-1-03 rejects invalid credentials without issuing a session cookie", async ({
  page,
  workerAdminRequest,
}) => {
  const phase1 = new Phase1Page(page);
  const email = uniqueEmail("phase1-e103");
  await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Phase 1 E103",
    initial_password: "Phase1E103Pass!",
    mfa_required: false,
  });

  await clearBrowserSession(page);
  await phase1.goto();
  const invalidLoginResponse = waitForPublicAPIResponse(page, {
    method: "POST",
    path: "/api/v1/auth/login",
    status: 401,
  });
  await phase1.login(email, "WrongPassword1!");
  await invalidLoginResponse;
  await expect(page.getByTestId(phase1ErrorCodeTestId("auth"))).toHaveText(
    "Email or password is incorrect.",
  );
  expect(await hasSessionCookie(page)).toBeFalsy();
});

test("E-1-05 lets deployment admins create and patch users, rejects stale versions, and shows the last-admin guard on the ordinary shell", async ({
  page,
  workerAdminRequest,
}) => {
  const phase1 = new Phase1Page(page);
  await phase1.goto();
  const createEmail = uniqueEmail("phase1-e105");

  await phase1.createUser({
    email: createEmail,
    displayName: "Phase 1 E105 User",
    password: "Phase1E105Pass!",
    mfaRequired: false,
  });

  await expect(page.getByTestId(phase1AdminTestId("status"))).toHaveText(
    "Created local user",
  );
  await expect(
    page.getByTestId(phase1AdminTestId("target-user-version")),
  ).toHaveText("1");
  const createdUserID = await page
    .getByTestId(phase1AdminTestId("target-user-id"))
    .textContent();
  if (!createdUserID) {
    throw new Error("missing created user id");
  }

  await page
    .getByTestId(phase1AdminTestId("patch-display-name"))
    .fill("Phase 1 E105 Patched");
  await phase1.patchTargetUser();
  await expect(
    page.getByTestId(phase1AdminTestId("target-user-version")),
  ).toHaveText("2");

  await patchUser(workerAdminRequest, createdUserID, {
    base_user_version: 2,
    display_name: "Phase 1 E105 Concurrent",
  });
  await phase1.patchTargetUser();
  await expect(page.getByTestId(phase1ErrorCodeTestId("admin"))).toHaveText(
    "user_version_conflict",
  );

  const currentAdminID = (await phase1.currentSession()).user_id;
  if (!currentAdminID) {
    throw new Error("missing current admin user id");
  }
  await withOnlyActiveDeploymentAdmin(
    deploymentAdminMutationClient(workerAdminRequest),
    currentAdminID,
    async () => {
      await phase1.loadTargetUser(currentAdminID);
      await expect(
        page.getByTestId(phase1AdminTestId("patch-is-deployment-admin")),
      ).toBeChecked();
      await expect(
        page.getByTestId(phase1AdminTestId("patch-is-active")),
      ).toBeChecked();

      await phase1.setCheckbox(
        phase1AdminTestId("patch-is-deployment-admin"),
        false,
      );
      await phase1.patchTargetUser();
      await expect(page.getByTestId(phase1ErrorCodeTestId("admin"))).toHaveText(
        "last_deployment_admin",
      );

      await phase1.loadTargetUser(currentAdminID);
      await expect(
        page.getByTestId(phase1AdminTestId("patch-is-deployment-admin")),
      ).toBeChecked();
      await phase1.setCheckbox(phase1AdminTestId("patch-is-active"), false);
      await phase1.patchTargetUser();
      await expect(page.getByTestId(phase1ErrorCodeTestId("admin"))).toHaveText(
        "last_deployment_admin",
      );
    },
  );
});

test("E-1-06 follows the bootstrap-token enrollment sequence on the ordinary login shell and proves completion alone issues no session", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const phase1 = new Phase1Page(page);
  const email = uniqueEmail("phase1-e106");
  const password = "Phase1E106Pass!";
  const user = await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Phase 1 E106",
    initial_password: password,
    mfa_required: true,
  });

  await clearBrowserSession(page);
  await phase1.goto();
  await phase1.login(email, password);
  await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "mfa_setup_required",
  );
  await expect(
    page.getByTestId(phase1AuthTestId("bootstrap-token")),
  ).not.toHaveText("");
  expect(await hasSessionCookie(page)).toBeFalsy();

  await phase1.beginBootstrapEnrollment();
  const secretBase32 = await phase1.requireText(
    phase1AuthTestId("bootstrap-secret-base32"),
  );

  await phase1.completeBootstrapEnrollment(generateTotpCode(secretBase32));
  await expect(
    page.getByText("Authenticator setup is complete. Sign in again.").first(),
  ).toBeVisible();
  expect(await hasSessionCookie(page)).toBeFalsy();

  await phase1.login(email, password);
  await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "mfa_required",
  );

  await phase1.login(email, password, generateTotpCode(secretBase32));
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email,
    purpose: "phase1 e106 post-bootstrap login",
    userId: user.user_id,
  });

  const loadedUser = await loadUser(workerAdminRequest, user.user_id);
  await resetUserTotp(
    workerAdminRequest,
    user.user_id,
    loadedUser.user_version,
    "phase1 e106 admin totp reset",
  );

  await clearBrowserSession(page);
  await phase1.goto();
  await phase1.login(email, password);
  await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "mfa_setup_required",
  );
  await expect(
    page.getByTestId(phase1AuthTestId("bootstrap-token")),
  ).not.toHaveText("");
  expect(await hasSessionCookie(page)).toBeFalsy();

  await phase1.beginBootstrapEnrollment();
  const replacementSecretBase32 = await phase1.requireText(
    phase1AuthTestId("bootstrap-secret-base32"),
  );

  await phase1.completeBootstrapEnrollment(
    generateTotpCode(replacementSecretBase32),
  );
  await expect(
    page.getByText("Authenticator setup is complete. Sign in again.").first(),
  ).toBeVisible();
  expect(await hasSessionCookie(page)).toBeFalsy();

  await phase1.login(email, password);
  await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "mfa_required",
  );

  await phase1.login(
    email,
    password,
    generateTotpCode(replacementSecretBase32),
  );
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email,
    purpose: "phase1 e106 post-reset bootstrap login",
    userId: user.user_id,
  });
});

test("E-1-07 requires the current password and current TOTP code, revokes the session immediately, and requires re-login with the new password", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const phase1 = new Phase1Page(page);
  const email = uniqueEmail("phase1-e107");
  const password = "Phase1E107Pass!";
  const user = await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Phase 1 E107",
    initial_password: password,
    mfa_required: true,
  });
  const secretBase32 = await enrollTotpViaBootstrap(email, password);

  await clearBrowserSession(page);
  await phase1.goto();
  await phase1.login(email, password, generateTotpCode(secretBase32));
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email,
    purpose: "phase1 e107 initial login before password change",
    userId: user.user_id,
  });

  await phase1.changePassword(
    "WrongCurrent1!",
    "Phase1E107Changed!",
    generateTotpCode(secretBase32),
  );
  await expect(page.getByTestId(phase1ErrorCodeTestId("account"))).toHaveText(
    "invalid_current_password",
  );

  await phase1.changePassword(password, "Phase1E107Changed!", "");
  await expect(page.getByTestId(phase1ErrorCodeTestId("account"))).toHaveText(
    "invalid_second_factor",
  );

  await phase1.changePassword(
    password,
    "Phase1E107Changed!",
    generateTotpCode(secretBase32),
  );
  await expect(
    page.getByTestId(phase1AuthTestId("login-username")),
  ).toBeVisible();

  await phase1.login(email, password);
  await expect(page.getByTestId(phase1ErrorCodeTestId("auth"))).toHaveText(
    "Email or password is incorrect.",
  );

  await phase1.login(
    email,
    "Phase1E107Changed!",
    generateTotpCode(secretBase32),
  );
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email,
    purpose: "phase1 e107 login after password change",
    userId: user.user_id,
  });
});

test("E-1-08 keeps deployment-user administration on deployment-admin sessions and hides it from an incident admin on the ordinary shell", async ({
  page,
  sessionTracker,
  workerAdmin,
  workerAdminRequest,
}) => {
  const phase1 = new Phase1Page(page);
  const targetEmail = uniqueEmail("phase1-e108-target");
  const targetPassword = "Phase1E108Pass!";
  const targetUser = await createLocalUser(workerAdminRequest, {
    email: targetEmail,
    display_name: "Phase 1 E108 Target",
    initial_password: targetPassword,
    mfa_required: true,
  });
  await enrollTotpViaBootstrap(targetEmail, targetPassword);

  const incidentAdminEmail = uniqueEmail("phase1-e108-incident-admin");
  const incidentAdminPassword = "Phase1E108Incident!";
  const incidentAdminUser = await createLocalUser(workerAdminRequest, {
    email: incidentAdminEmail,
    display_name: "Phase 1 E108 Incident Admin",
    initial_password: incidentAdminPassword,
    mfa_required: false,
  });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E108"),
    "Phase 1 E-1-08",
  );
  await createIncidentMembership(page, incidentId, incidentAdminEmail, "admin");

  await clearBrowserSession(page);
  await phase1.goto();
  await phase1.login(incidentAdminEmail, incidentAdminPassword);
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email: incidentAdminEmail,
    purpose: "phase1 e108 incident-admin login",
    userId: incidentAdminUser.user_id,
  });
  await expect(
    page.getByTestId(landingAdminMenuItemTestId("deployment-users")),
  ).toHaveCount(0);
  await expect(page.getByTestId(phase1AdminTestId("access-note"))).toHaveCount(
    0,
  );
  await expect(
    page.getByTestId(phase1AdminTestId("password-reset")),
  ).toHaveCount(0);

  const targetUserVersion = (
    await loadUser(workerAdminRequest, targetUser.user_id)
  ).user_version;
  const incidentAdminRequests =
    await authenticatedRequestContextFromStorageState(
      await page.context().storageState(),
    );
  try {
    await expectUnauthorizedCredentialAction(
      incidentAdminRequests,
      `/api/v1/users/${targetUser.user_id}/password/reset`,
      {
        base_user_version: targetUserVersion,
        client_txn_id: uniqueTxn("phase1-e108-incident-password-reset"),
        new_password: "Phase1E108Changed!",
        reason: "incident admin denial probe",
      },
    );
    await expectUnauthorizedCredentialAction(
      incidentAdminRequests,
      `/api/v1/users/${targetUser.user_id}/mfa/totp/reset`,
      {
        base_user_version: targetUserVersion,
        client_txn_id: uniqueTxn("phase1-e108-incident-totp-reset"),
        reason: "incident admin denial probe",
      },
    );
    await expectUnauthorizedCredentialAction(
      incidentAdminRequests,
      `/api/v1/users/${targetUser.user_id}/sessions/revoke-all`,
      {
        client_txn_id: uniqueTxn("phase1-e108-incident-revoke-all"),
        reason: "incident admin denial probe",
      },
    );
  } finally {
    await incidentAdminRequests.dispose();
  }

  await restoreTrackedStorageState(page, workerAdmin.storageState);
  await phase1.goto();
  await phase1.loadTargetUser(targetUser.user_id);
  await expect(
    page.getByTestId(phase1AdminTestId("target-user-version")),
  ).toHaveText(String(targetUserVersion));
  await page.getByTestId(phase1AdminTestId("password-reset")).click();
  await page
    .getByTestId(phase1AdminTestId("new-password"))
    .fill("Phase1E108Changed!");
  await page
    .getByTestId(phase1AdminTestId("reason"))
    .fill("deployment admin action");

  await page.getByTestId(phase1AdminTestId("password-reset")).click();
  await expect(page.getByTestId(phase1AdminTestId("status"))).toHaveText(
    "Reset user password",
  );
  await expect(
    page.getByTestId(phase1AdminTestId("target-user-version")),
  ).toHaveText(String(targetUserVersion + 1));

  await page.getByTestId(phase1AdminTestId("totp-reset")).click();
  await page
    .getByTestId(phase1AdminTestId("reason"))
    .fill("deployment admin action");
  await page.getByTestId(phase1AdminTestId("totp-reset")).click();
  await expect(page.getByTestId(phase1AdminTestId("status"))).toHaveText(
    "Reset user TOTP",
  );
  await expect(
    page.getByTestId(phase1AdminTestId("target-user-version")),
  ).toHaveText(String(targetUserVersion + 2));

  await page.getByTestId(phase1AdminTestId("revoke-all")).click();
  await page
    .getByTestId(phase1AdminTestId("reason"))
    .fill("deployment admin action");
  await page.getByTestId(phase1AdminTestId("revoke-all")).click();
  await expect(page.getByTestId(phase1AdminTestId("status"))).toHaveText(
    "Revoked every user session",
  );
  await expect(
    page.getByTestId(phase1AdminTestId("target-user-version")),
  ).toHaveText(String(targetUserVersion + 2));

  await clearBrowserSession(page);
  await phase1.goto();
  await phase1.login(incidentAdminEmail, incidentAdminPassword);
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email: incidentAdminEmail,
    purpose: "phase1 e108 incident-admin login after admin actions",
    userId: incidentAdminUser.user_id,
  });
});

test("E-1-09 creates an incident from the landing screen, lists it, and opens the workbook as incident admin", async ({
  page,
}) => {
  const phase1 = new Phase1Page(page);
  const incidentKey = uniqueIncidentKey("E109");
  const incidentTitle = "Phase 1 E-1-09";
  const secondIncidentKey = uniqueIncidentKey("E109B");
  const secondIncidentTitle = "Phase 1 E-1-09 companion";

  await phase1.gotoIncidentDirectory();
  await expect(page.getByTestId(phase1LandingTestId("shell"))).toBeVisible();

  const createResponsePromise = waitForPublicAPIResponse(page, {
    method: "POST",
    path: "/api/v1/incidents",
    status: [200, 201],
  });
  await phase1.createAndOpenIncident(incidentKey, incidentTitle);
  const createResponse = await createResponsePromise;
  const createBody = (await createResponse.json()) as {
    data: { incident_id: string };
  };
  const incidentId = createBody.data.incident_id;

  await expect(page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectCurrentIncidentRole(page, "Current incident role: admin");
  await openIncidentControls(page);
  await expect(page.getByTestId("incident-summary-key")).toHaveText(
    incidentKey,
  );

  const secondIncidentId = await createIncident(
    page,
    secondIncidentKey,
    secondIncidentTitle,
  );
  await phase1.returnToLanding();
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toContainText(incidentTitle);
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toContainText(incidentKey);
  await expect(
    page.getByTestId(landingIncidentCardTestId(secondIncidentId)),
  ).toContainText(secondIncidentTitle);

  await phase1.openIncident(incidentId);
  await expect(page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
});

test("E-1-10 clears a stale selected incident after membership removal while preserving the account session", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const phase1 = new Phase1Page(page);
  const selectedIncidentId = await createIncident(
    page,
    uniqueIncidentKey("E110A"),
    "Phase 1 E-1-10 selected",
  );
  const alternateIncidentId = await createIncident(
    page,
    uniqueIncidentKey("E110B"),
    "Phase 1 E-1-10 alternate",
  );
  const targetEmail = uniqueEmail("phase1-e110-target");
  const targetPassword = "Phase1E110Pass!";
  const targetUser = await createLocalUser(workerAdminRequest, {
    email: targetEmail,
    display_name: "Phase 1 E110 Target",
    initial_password: targetPassword,
    mfa_required: false,
  });
  await createIncidentMembership(
    page,
    selectedIncidentId,
    targetEmail,
    "admin",
  );
  await createIncidentMembership(
    page,
    alternateIncidentId,
    targetEmail,
    "viewer",
  );

  await clearBrowserSession(page);
  await phase1.goto();
  await phase1.login(targetEmail, targetPassword);
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email: targetEmail,
    purpose: "phase1 e110 stale incident selection target",
    userId: targetUser.user_id,
  });

  await phase1.openIncident(selectedIncidentId);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectCurrentIncidentRole(page, "Current incident role: admin");
  await openIncidentControls(page, "memberships");
  await expect(
    page.getByTestId(incidentMembershipRowTestId(targetUser.user_id)),
  ).toBeVisible();
  await expect(
    page.getByTestId(incidentMembershipRoleInputTestId(targetUser.user_id)),
  ).toHaveValue("admin");
  await expect(
    page.getByTestId(incidentMembershipPatchButtonTestId(targetUser.user_id)),
  ).toBeVisible();
  await expect(
    page.getByTestId(incidentMembershipDeleteButtonTestId(targetUser.user_id)),
  ).toBeVisible();

  const selectedMembership = await loadIncidentMembership(
    workerAdminRequest,
    selectedIncidentId,
    targetUser.user_id,
  );
  await deleteIncidentMembership(
    workerAdminRequest,
    selectedIncidentId,
    targetUser.user_id,
    selectedMembership.membership_version,
  );

  await page.reload();
  await expect(page).not.toHaveURL(
    new RegExp(`incident_id=${selectedIncidentId}`),
  );
  await expect(page).toHaveURL(
    new RegExp(`incident_id=${alternateIncidentId}`),
  );
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectCurrentIncidentRole(page, "Current incident role: viewer");
  await expectLandingAccountSession(page);

  await phase1.openIncident(alternateIncidentId);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectCurrentIncidentRole(page, "Current incident role: viewer");
});

test("E-1-11 observes current-role authorization on a stale reviewer edit through the public incident error envelope", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const phase1 = new Phase1Page(page);
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E111"),
    "Phase 1 E-1-11",
  );
  const targetEmail = uniqueEmail("phase1-e111-reviewer");
  const targetPassword = "Phase1E111Pass!";
  const targetUser = await createLocalUser(workerAdminRequest, {
    email: targetEmail,
    display_name: "Phase 1 E111 Reviewer",
    initial_password: targetPassword,
    mfa_required: false,
  });
  await createIncidentMembership(page, incidentId, targetEmail, "reviewer");

  await clearBrowserSession(page);
  await phase1.goto();
  await phase1.login(targetEmail, targetPassword);
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email: targetEmail,
    purpose: "phase1 e111 reviewer stale edit",
    userId: targetUser.user_id,
  });

  await phase1.openIncident(incidentId);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectCurrentIncidentRole(page, "Current incident role: reviewer");
  await openIncidentControls(page, "memberships");
  await expect(
    page.getByTestId(incidentMembershipRoleDisplayTestId(targetUser.user_id)),
  ).toHaveText("reviewer");
  await openIncidentControls(page, "incident-fields");
  await expect(page.getByTestId("incident-patch-button")).toBeVisible();

  const reviewerMembership = await loadIncidentMembership(
    workerAdminRequest,
    incidentId,
    targetUser.user_id,
  );
  await patchIncidentMembership(workerAdminRequest, incidentId, {
    baseMembershipVersion: reviewerMembership.membership_version,
    role: "editor",
    userId: targetUser.user_id,
  });

  const forbiddenPatchResponse = waitForPublicAPIResponse(page, {
    method: "PATCH",
    path: `/api/v1/incidents/${incidentId}`,
    status: 403,
  });
  await phase1.patchIncidentFields({
    currentPhase: "containment",
    externalCase: "CASE-E111",
    tlp: "TLP:AMBER",
  });
  const patchResponse = await forbiddenPatchResponse;
  await expect(patchResponse.json()).resolves.toMatchObject({
    error: {
      code: "authorization_denied",
    },
  });
  await expect(page.getByTestId("incident-admin-error-code")).toHaveText(
    "authorization_denied",
  );
  await expect(page.locator("body")).not.toContainText("request_id");
  await expect(page.locator("body")).not.toContainText("traceback");

  await page.reload();
  await expectCurrentIncidentRole(page, "Current incident role: editor");
  await openIncidentControls(page, "incident-fields");
  await expect(page.getByTestId("incident-patch-readonly-note")).toBeVisible();
});

test("E-1-12 returns a revoked target browser to login and allows re-authentication with unchanged incident membership", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const phase1 = new Phase1Page(page);
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E112"),
    "Phase 1 E-1-12",
  );
  const alternateIncidentId = await createIncident(
    page,
    uniqueIncidentKey("E112B"),
    "Phase 1 E-1-12 alternate",
  );
  const targetEmail = uniqueEmail("phase1-e112-target");
  const targetPassword = "Phase1E112Pass!";
  const targetUser = await createLocalUser(workerAdminRequest, {
    email: targetEmail,
    display_name: "Phase 1 E112 Target",
    initial_password: targetPassword,
    mfa_required: false,
  });
  await createIncidentMembership(page, incidentId, targetEmail, "viewer");
  await createIncidentMembership(
    page,
    alternateIncidentId,
    targetEmail,
    "viewer",
  );

  await clearBrowserSession(page);
  await phase1.goto();
  await phase1.login(targetEmail, targetPassword);
  await expectLandingAccountSession(page);
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toBeVisible();
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email: targetEmail,
    purpose: "phase1 e112 target before revoke-all",
    userId: targetUser.user_id,
  });

  await revokeAllSessions(
    workerAdminRequest,
    targetUser.user_id,
    "phase1 e112 explicit revoke-all",
  );

  const revokedSessionResponse = waitForPublicAPIResponse(page, {
    method: "GET",
    path: "/api/v1/auth/session",
    status: 401,
  });
  await phase1.refreshLanding();
  const sessionResponse = await revokedSessionResponse;
  await expect(sessionResponse.json()).resolves.toMatchObject({
    error: {
      code: "session_required",
    },
  });
  await expect(
    page.getByTestId(phase1AuthTestId("login-username")),
  ).toBeVisible();
  await expect(
    page.getByTestId(phase1AuthTestId("shell-message")),
  ).toContainText("Sign in again");

  await phase1.login(targetEmail, targetPassword);
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email: targetEmail,
    purpose: "phase1 e112 target after revoke-all re-auth",
    userId: targetUser.user_id,
  });
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toBeVisible();

  await phase1.openIncident(incidentId);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectCurrentIncidentRole(page, "Current incident role: viewer");
});

test("FE-E-P1-01 Verify ordinary login, incident entry, and current-role refresh stay on public browser routes.", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const phase1 = new Phase1Page(page);
  const email = uniqueEmail("phase1-feep101");
  const password = "Phase1FEEP101Pass!";
  const user = await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Phase 1 FE-E-P1-01",
    initial_password: password,
    mfa_required: false,
  });
  const incidentKey = uniqueIncidentKey("FEEP101");
  const incidentTitle = "FE-E-P1-01 incident entry";
  const incidentId = await createIncident(page, incidentKey, incidentTitle);
  await createIncidentMembership(page, incidentId, email, "admin");

  await clearBrowserSession(page);
  await phase1.goto();
  await expect(
    page.getByTestId(phase1AuthTestId("login-username")),
  ).toBeVisible();
  const loginResponse = waitForPublicAPIResponse(page, {
    method: "POST",
    path: "/api/v1/auth/login",
    status: 200,
  });
  await phase1.login(email, password);
  await loginResponse;
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email,
    purpose: "FE-E-P1-01 ordinary login",
    userId: user.user_id,
  });

  await phase1.gotoIncidentDirectory();
  await expect(page.getByTestId(phase1LandingTestId("shell"))).toBeVisible();
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toContainText(incidentTitle);
  await phase1.openIncident(incidentId);
  await expect(page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectCurrentIncidentRole(page, "Current incident role: admin");
  await openIncidentControls(page);
  await expect(page.getByTestId("incident-summary-key")).toHaveText(
    incidentKey,
  );

  const membership = await loadIncidentMembership(
    workerAdminRequest,
    incidentId,
    user.user_id,
  );
  await patchIncidentMembership(workerAdminRequest, incidentId, {
    baseMembershipVersion: membership.membership_version,
    role: "viewer",
    userId: user.user_id,
  });
  await page.reload();
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectCurrentIncidentRole(page, "Current incident role: viewer");
  await expectLandingAccountSession(page);
});

async function createIncident(page: Page, incidentKey: string, title: string) {
  const response = await page.request.post(`${apiBase}/api/v1/incidents`, {
    headers: await csrfHeaders(page),
    data: {
      client_txn_id: uniqueTxn("phase1-incident"),
      incident_key: incidentKey,
      title,
    },
  });
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as { data: { incident_id: string } };
  return body.data.incident_id;
}

async function createIncidentMembership(
  page: Page,
  incidentId: string,
  email: string,
  role: string,
) {
  const response = await page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships`,
    {
      headers: await csrfHeaders(page),
      data: {
        client_txn_id: uniqueTxn("phase1-membership"),
        email,
        role,
      },
    },
  );
  expect(response.ok()).toBeTruthy();
}

type IncidentMembershipRecord = {
  membership_version: number;
  role: string;
  user_id: string;
};

async function loadIncidentMembership(
  authRequests: APIRequestContext,
  incidentId: string,
  userId: string,
) {
  const response = await authRequests.get(
    `/api/v1/incidents/${incidentId}/memberships`,
  );
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as {
    data: { memberships: IncidentMembershipRecord[] };
  };
  const membership =
    body.data.memberships.find((candidate) => candidate.user_id === userId) ??
    null;
  if (membership === null) {
    throw new Error(`missing incident membership for ${userId}`);
  }
  return membership;
}

async function patchIncidentMembership(
  authRequests: APIRequestContext,
  incidentId: string,
  options: {
    baseMembershipVersion: number;
    role: string;
    userId: string;
  },
) {
  const response = await authRequests.patch(
    `/api/v1/incidents/${incidentId}/memberships/${options.userId}`,
    {
      data: {
        base_membership_version: options.baseMembershipVersion,
        role: options.role,
      },
    },
  );
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: IncidentMembershipRecord }).data;
}

async function deleteIncidentMembership(
  authRequests: APIRequestContext,
  incidentId: string,
  userId: string,
  baseMembershipVersion: number,
) {
  const response = await authRequests.delete(
    `/api/v1/incidents/${incidentId}/memberships/${userId}`,
    {
      data: {
        base_membership_version: baseMembershipVersion,
      },
    },
  );
  expect(response.status()).toBe(204);
}

function waitForPublicAPIResponse(
  page: Page,
  options: {
    method: string;
    path: string;
    status?: number | number[];
  },
) {
  const expectedMethod = options.method.toUpperCase();
  const expectedStatuses =
    options.status === undefined
      ? null
      : new Set(
          Array.isArray(options.status) ? options.status : [options.status],
        );
  return page.waitForResponse((candidate) => {
    const request = candidate.request();
    if (request.method().toUpperCase() !== expectedMethod) {
      return false;
    }
    if (new URL(candidate.url()).pathname !== options.path) {
      return false;
    }
    return (
      expectedStatuses === null || expectedStatuses.has(candidate.status())
    );
  });
}

async function expectUnauthorizedCredentialAction(
  authRequests: APIRequestContext,
  path: string,
  data: Record<string, unknown>,
) {
  const response = await authRequests.post(path, { data });
  expect(response.status()).toBe(403);
  await expect(response.json()).resolves.toMatchObject({
    error: {
      code: "authorization_denied",
      details: {
        required_capability: "deployment_admin",
      },
    },
  });
}

async function clearBrowserSession(page: Page) {
  await page.context().clearCookies();
}

async function hasSessionCookie(page: Page) {
  const cookies = await page.context().cookies(apiBase);
  return cookies.some((cookie) => cookie.name === sessionCookieName);
}
