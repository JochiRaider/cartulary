import {
  accountTestId,
  authTestId,
  currentIncidentRoleTestId,
  deploymentAdminTestId,
  incidentAdministrationTestId,
  incidentLandingTestId,
  incidentMembershipDeleteButtonTestId,
  incidentMembershipPatchButtonTestId,
  incidentMembershipRoleDisplayTestId,
  incidentMembershipRoleInputTestId,
  incidentMembershipRowTestId,
  landingAdminMenuItemTestId,
  landingIncidentCardTestId,
  publicErrorCodeTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import type { APIRequestContext, Page } from "@playwright/test";
import { expect, restoreTrackedStorageState, test } from "./fixtures";
import { AccountSettings } from "./pages/accountSettings";
import { AuthGateway } from "./pages/authGateway";
import {
  DeploymentAdministration,
  openIncidentControls,
} from "./pages/deploymentAdministration";
import { IncidentDirectory } from "./pages/incidentDirectory";
import { csrfHeaders } from "./support/auth/browserSession";
import {
  authenticatedRequestContextFromStorageState,
  createLocalUser,
  deploymentAdminMutationClient,
  loadUser,
  patchUser,
  readCurrentSession,
  resetUserTotp,
  revokeAllSessions,
  withOnlyActiveDeploymentAdmin,
} from "./support/auth/sessions";
import { sessionCookieName } from "./support/auth/storageState";
import {
  enrollTotpViaBootstrap,
  generateTotpCode,
} from "./support/auth/suiteAdmin";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";

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
  await new AccountSettings(page).open("account-security");
  await expect(page.getByTestId(accountTestId("refresh-state"))).toBeVisible();
  await expect(page.getByTestId(accountTestId("logout"))).toBeVisible();
  await expect(
    page.getByTestId(accountTestId("password-current")),
  ).toBeVisible();
}

test("signs in as a local user and inspects the ordinary session surface", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const email = uniqueEmail("authentication-e101");
  const password = "AuthenticationE101Pass!";
  const user = await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Authentication E101",
    initial_password: password,
    mfa_required: false,
  });

  await clearBrowserSession(page);
  await new AuthGateway(page).goto();
  await expect(page.getByTestId(authTestId("login-username"))).toBeVisible();
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
  await new AuthGateway(page).login(email, password);
  await Promise.all([
    loginResponse,
    sessionResponse,
    credentialStateResponse,
    incidentListResponse,
  ]);

  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication ordinary shell",
    email,
    purpose: "authentication e101 successful ordinary login",
    userId: user.user_id,
  });
});

test("requires MFA on the ordinary login surface, rejects wrong codes, and accepts a valid TOTP code", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const email = uniqueEmail("authentication-e102");
  const password = "AuthenticationE102Pass!";
  const user = await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Authentication E102",
    initial_password: password,
    mfa_required: true,
  });
  const secretBase32 = await enrollTotpViaBootstrap(email, password);

  await clearBrowserSession(page);
  await new AuthGateway(page).goto();
  const missingTotpResponse = waitForPublicAPIResponse(page, {
    method: "POST",
    path: "/api/v1/auth/login",
    status: 401,
  });

  await new AuthGateway(page).login(email, password);
  await missingTotpResponse;
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "mfa_required",
  );
  await expect(page.getByTestId(authTestId("login-totp-code"))).toBeVisible();
  await expect(page.getByTestId(publicErrorCodeTestId("auth"))).toHaveText("");
  expect(await hasSessionCookie(page)).toBeFalsy();

  const wrongTotpResponse = waitForPublicAPIResponse(page, {
    method: "POST",
    path: "/api/v1/auth/login",
    status: 401,
  });
  await new AuthGateway(page).login(email, password, "000000");
  await wrongTotpResponse;
  await expect(page.getByTestId(publicErrorCodeTestId("auth"))).toHaveText(
    "The verification code is incorrect or expired.",
  );
  expect(await hasSessionCookie(page)).toBeFalsy();

  const validTotpResponse = waitForPublicAPIResponse(page, {
    method: "POST",
    path: "/api/v1/auth/login",
    status: 200,
  });
  await new AuthGateway(page).login(
    email,
    password,
    generateTotpCode(secretBase32),
  );
  await validTotpResponse;
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication ordinary shell",
    email,
    purpose: "authentication e102 successful ordinary login",
    userId: user.user_id,
  });
});

test("rejects invalid credentials without issuing a session cookie", async ({
  page,
  workerAdminRequest,
}) => {
  const email = uniqueEmail("authentication-e103");
  await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Authentication E103",
    initial_password: "AuthenticationE103Pass!",
    mfa_required: false,
  });

  await clearBrowserSession(page);
  await new AuthGateway(page).goto();
  const invalidLoginResponse = waitForPublicAPIResponse(page, {
    method: "POST",
    path: "/api/v1/auth/login",
    status: 401,
  });
  await new AuthGateway(page).login(email, "WrongPassword1!");
  await invalidLoginResponse;
  await expect(page.getByTestId(publicErrorCodeTestId("auth"))).toHaveText(
    "Email or password is incorrect.",
  );
  expect(await hasSessionCookie(page)).toBeFalsy();
});

test("lets deployment admins create and patch users, rejects stale versions, and shows the last-admin guard on the ordinary shell", async ({
  page,
  workerAdminRequest,
}) => {
  await new AuthGateway(page).goto();
  const createEmail = uniqueEmail("authentication-e105");

  await new DeploymentAdministration(page).createUser({
    email: createEmail,
    displayName: "Authentication E105 User",
    password: "AuthenticationE105Pass!",
    mfaRequired: false,
  });

  await expect(page.getByTestId(deploymentAdminTestId("status"))).toHaveText(
    "Created local user",
  );
  await expect(
    page.getByTestId(deploymentAdminTestId("target-user-version")),
  ).toHaveText("1");
  const createdUserID = await page
    .getByTestId(deploymentAdminTestId("target-user-id"))
    .textContent();
  if (!createdUserID) {
    throw new Error("missing created user id");
  }

  await page
    .getByTestId(deploymentAdminTestId("patch-display-name"))
    .fill("Authentication E105 Patched");
  await new DeploymentAdministration(page).patchTargetUser();
  await expect(
    page.getByTestId(deploymentAdminTestId("target-user-version")),
  ).toHaveText("2");

  await patchUser(workerAdminRequest, createdUserID, {
    base_user_version: 2,
    display_name: "Authentication E105 Concurrent",
  });
  await new DeploymentAdministration(page).patchTargetUser();
  await expect(page.getByTestId(publicErrorCodeTestId("admin"))).toHaveText(
    "user_version_conflict",
  );

  const currentAdminID = (await readCurrentSession(page)).user_id;
  if (!currentAdminID) {
    throw new Error("missing current admin user id");
  }
  await withOnlyActiveDeploymentAdmin(
    deploymentAdminMutationClient(workerAdminRequest),
    currentAdminID,
    async () => {
      await new DeploymentAdministration(page).loadTargetUser(currentAdminID);
      await expect(
        page.getByTestId(deploymentAdminTestId("patch-is-deployment-admin")),
      ).toBeChecked();
      await expect(
        page.getByTestId(deploymentAdminTestId("patch-is-active")),
      ).toBeChecked();

      await new DeploymentAdministration(page).setCheckbox(
        deploymentAdminTestId("patch-is-deployment-admin"),
        false,
      );
      await new DeploymentAdministration(page).patchTargetUser();
      await expect(page.getByTestId(publicErrorCodeTestId("admin"))).toHaveText(
        "last_deployment_admin",
      );

      await new DeploymentAdministration(page).loadTargetUser(currentAdminID);
      await expect(
        page.getByTestId(deploymentAdminTestId("patch-is-deployment-admin")),
      ).toBeChecked();
      await new DeploymentAdministration(page).setCheckbox(
        deploymentAdminTestId("patch-is-active"),
        false,
      );
      await new DeploymentAdministration(page).patchTargetUser();
      await expect(page.getByTestId(publicErrorCodeTestId("admin"))).toHaveText(
        "last_deployment_admin",
      );
    },
  );
});

test("follows the bootstrap-token enrollment sequence on the ordinary login shell and proves completion alone issues no session", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const email = uniqueEmail("authentication-e106");
  const password = "AuthenticationE106Pass!";
  const user = await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Authentication E106",
    initial_password: password,
    mfa_required: true,
  });

  await clearBrowserSession(page);
  await new AuthGateway(page).goto();
  await new AuthGateway(page).login(email, password);
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "mfa_setup_required",
  );
  await expect(page.getByTestId(authTestId("bootstrap-token"))).not.toHaveText(
    "",
  );
  expect(await hasSessionCookie(page)).toBeFalsy();

  await new AuthGateway(page).beginBootstrapEnrollment();
  const secretBase32 = await new AuthGateway(page).requireText(
    authTestId("bootstrap-secret-base32"),
  );

  await new AuthGateway(page).completeBootstrapEnrollment(
    generateTotpCode(secretBase32),
  );
  await expect(
    page.getByText("Authenticator setup is complete. Sign in again.").first(),
  ).toBeVisible();
  expect(await hasSessionCookie(page)).toBeFalsy();

  await new AuthGateway(page).login(email, password);
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "mfa_required",
  );

  await new AuthGateway(page).login(
    email,
    password,
    generateTotpCode(secretBase32),
  );
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication ordinary shell",
    email,
    purpose: "authentication e106 post-bootstrap login",
    userId: user.user_id,
  });

  const loadedUser = await loadUser(workerAdminRequest, user.user_id);
  await resetUserTotp(
    workerAdminRequest,
    user.user_id,
    loadedUser.user_version,
    "authentication e106 admin totp reset",
  );

  await clearBrowserSession(page);
  await new AuthGateway(page).goto();
  await new AuthGateway(page).login(email, password);
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "mfa_setup_required",
  );
  await expect(page.getByTestId(authTestId("bootstrap-token"))).not.toHaveText(
    "",
  );
  expect(await hasSessionCookie(page)).toBeFalsy();

  await new AuthGateway(page).beginBootstrapEnrollment();
  const replacementSecretBase32 = await new AuthGateway(page).requireText(
    authTestId("bootstrap-secret-base32"),
  );

  await new AuthGateway(page).completeBootstrapEnrollment(
    generateTotpCode(replacementSecretBase32),
  );
  await expect(
    page.getByText("Authenticator setup is complete. Sign in again.").first(),
  ).toBeVisible();
  expect(await hasSessionCookie(page)).toBeFalsy();

  await new AuthGateway(page).login(email, password);
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "mfa_required",
  );

  await new AuthGateway(page).login(
    email,
    password,
    generateTotpCode(replacementSecretBase32),
  );
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication ordinary shell",
    email,
    purpose: "authentication e106 post-reset bootstrap login",
    userId: user.user_id,
  });
});

test("requires the current password and current TOTP code, revokes the session immediately, and requires re-login with the new password", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const email = uniqueEmail("authentication-e107");
  const password = "AuthenticationE107Pass!";
  const user = await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Authentication E107",
    initial_password: password,
    mfa_required: true,
  });
  const secretBase32 = await enrollTotpViaBootstrap(email, password);

  await clearBrowserSession(page);
  await new AuthGateway(page).goto();
  await new AuthGateway(page).login(
    email,
    password,
    generateTotpCode(secretBase32),
  );
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication ordinary shell",
    email,
    purpose: "authentication e107 initial login before password change",
    userId: user.user_id,
  });

  await new AccountSettings(page).changePassword(
    "WrongCurrent1!",
    "AuthenticationE107Changed!",
    generateTotpCode(secretBase32),
  );
  await expect(page.getByTestId(publicErrorCodeTestId("account"))).toHaveText(
    "invalid_current_password",
  );

  await new AccountSettings(page).changePassword(
    password,
    "AuthenticationE107Changed!",
    "",
  );
  await expect(page.getByTestId(publicErrorCodeTestId("account"))).toHaveText(
    "invalid_second_factor",
  );

  await new AccountSettings(page).changePassword(
    password,
    "AuthenticationE107Changed!",
    generateTotpCode(secretBase32),
  );
  await expect(page.getByTestId(authTestId("login-username"))).toBeVisible();

  await new AuthGateway(page).login(email, password);
  await expect(page.getByTestId(publicErrorCodeTestId("auth"))).toHaveText(
    "Email or password is incorrect.",
  );

  await new AuthGateway(page).login(
    email,
    "AuthenticationE107Changed!",
    generateTotpCode(secretBase32),
  );
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication ordinary shell",
    email,
    purpose: "authentication e107 login after password change",
    userId: user.user_id,
  });
});

test("keeps deployment-user administration on deployment-admin sessions and hides it from an incident admin on the ordinary shell", async ({
  page,
  sessionTracker,
  workerAdmin,
  workerAdminRequest,
}) => {
  const targetEmail = uniqueEmail("authentication-e108-target");
  const targetPassword = "AuthenticationE108Pass!";
  const targetUser = await createLocalUser(workerAdminRequest, {
    email: targetEmail,
    display_name: "Authentication E108 Target",
    initial_password: targetPassword,
    mfa_required: true,
  });
  await enrollTotpViaBootstrap(targetEmail, targetPassword);

  const incidentAdminEmail = uniqueEmail("authentication-e108-incident-admin");
  const incidentAdminPassword = "AuthenticationE108Incident!";
  const incidentAdminUser = await createLocalUser(workerAdminRequest, {
    email: incidentAdminEmail,
    display_name: "Authentication E108 Incident Admin",
    initial_password: incidentAdminPassword,
    mfa_required: false,
  });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("ACCESS-SESSION"),
    "Authentication access-control",
  );
  await createIncidentMembership(page, incidentId, incidentAdminEmail, "admin");

  await clearBrowserSession(page);
  await new AuthGateway(page).goto();
  await new AuthGateway(page).login(incidentAdminEmail, incidentAdminPassword);
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication ordinary shell",
    email: incidentAdminEmail,
    purpose: "authentication e108 incident-admin login",
    userId: incidentAdminUser.user_id,
  });
  await expect(
    page.getByTestId(landingAdminMenuItemTestId("deployment-users")),
  ).toHaveCount(0);
  await expect(
    page.getByTestId(deploymentAdminTestId("access-note")),
  ).toHaveCount(0);
  await expect(
    page.getByTestId(deploymentAdminTestId("password-reset")),
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
        client_txn_id: uniqueTxn("authentication-e108-incident-password-reset"),
        new_password: "AuthenticationE108Changed!",
        reason: "incident admin denial probe",
      },
    );
    await expectUnauthorizedCredentialAction(
      incidentAdminRequests,
      `/api/v1/users/${targetUser.user_id}/mfa/totp/reset`,
      {
        base_user_version: targetUserVersion,
        client_txn_id: uniqueTxn("authentication-e108-incident-totp-reset"),
        reason: "incident admin denial probe",
      },
    );
    await expectUnauthorizedCredentialAction(
      incidentAdminRequests,
      `/api/v1/users/${targetUser.user_id}/sessions/revoke-all`,
      {
        client_txn_id: uniqueTxn("authentication-e108-incident-revoke-all"),
        reason: "incident admin denial probe",
      },
    );
  } finally {
    await incidentAdminRequests.dispose();
  }

  await restoreTrackedStorageState(page, workerAdmin.storageState);
  await new AuthGateway(page).goto();
  await new DeploymentAdministration(page).loadTargetUser(targetUser.user_id);
  await expect(
    page.getByTestId(deploymentAdminTestId("target-user-version")),
  ).toHaveText(String(targetUserVersion));
  await page.getByTestId(deploymentAdminTestId("password-reset")).click();
  await page
    .getByTestId(deploymentAdminTestId("new-password"))
    .fill("AuthenticationE108Changed!");
  await page
    .getByTestId(deploymentAdminTestId("reason"))
    .fill("deployment admin action");

  await page.getByTestId(deploymentAdminTestId("password-reset")).click();
  await expect(page.getByTestId(deploymentAdminTestId("status"))).toHaveText(
    "Reset user password",
  );
  await expect(
    page.getByTestId(deploymentAdminTestId("target-user-version")),
  ).toHaveText(String(targetUserVersion + 1));

  await page.getByTestId(deploymentAdminTestId("totp-reset")).click();
  await page
    .getByTestId(deploymentAdminTestId("reason"))
    .fill("deployment admin action");
  await page.getByTestId(deploymentAdminTestId("totp-reset")).click();
  await expect(page.getByTestId(deploymentAdminTestId("status"))).toHaveText(
    "Reset user TOTP",
  );
  await expect(
    page.getByTestId(deploymentAdminTestId("target-user-version")),
  ).toHaveText(String(targetUserVersion + 2));

  await page.getByTestId(deploymentAdminTestId("revoke-all")).click();
  await page
    .getByTestId(deploymentAdminTestId("reason"))
    .fill("deployment admin action");
  await page.getByTestId(deploymentAdminTestId("revoke-all")).click();
  await expect(page.getByTestId(deploymentAdminTestId("status"))).toHaveText(
    "Revoked every user session",
  );
  await expect(
    page.getByTestId(deploymentAdminTestId("target-user-version")),
  ).toHaveText(String(targetUserVersion + 2));

  await clearBrowserSession(page);
  await new AuthGateway(page).goto();
  await new AuthGateway(page).login(incidentAdminEmail, incidentAdminPassword);
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication ordinary shell",
    email: incidentAdminEmail,
    purpose: "authentication e108 incident-admin login after admin actions",
    userId: incidentAdminUser.user_id,
  });
});

test("creates an incident from the landing screen, lists it, and opens the workbook as incident admin", async ({
  page,
}) => {
  const incidentKey = uniqueIncidentKey("INCIDENT-DIRECTORY");
  const incidentTitle = "Authentication incident-directory";
  const secondIncidentKey = uniqueIncidentKey("INCIDENT-DIRECTORY-COMPANION");
  const secondIncidentTitle = "Authentication incident-directory companion";

  await new IncidentDirectory(page).goto();
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();

  const createResponsePromise = waitForPublicAPIResponse(page, {
    method: "POST",
    path: "/api/v1/incidents",
    status: [200, 201],
  });
  await new IncidentDirectory(page).createAndOpenIncident(
    incidentKey,
    incidentTitle,
  );
  const createResponse = await createResponsePromise;
  const createBody = (await createResponse.json()) as {
    data: { incident_id: string };
  };
  const incidentId = createBody.data.incident_id;

  await expect(page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectCurrentIncidentRole(page, "Current incident role: admin");
  await openIncidentControls(page);
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-key")),
  ).toHaveText(incidentKey);

  const secondIncidentId = await createIncident(
    page,
    secondIncidentKey,
    secondIncidentTitle,
  );
  await new IncidentDirectory(page).open();
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toContainText(incidentTitle);
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toContainText(incidentKey);
  await expect(
    page.getByTestId(landingIncidentCardTestId(secondIncidentId)),
  ).toContainText(secondIncidentTitle);

  await new IncidentDirectory(page).openIncident(incidentId);
  await expect(page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
});

test("clears a stale selected incident after membership removal while preserving the account session", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const selectedIncidentId = await createIncident(
    page,
    uniqueIncidentKey("INCIDENT-SELECTION"),
    "Authentication incident-directory selected",
  );
  const alternateIncidentId = await createIncident(
    page,
    uniqueIncidentKey("INCIDENT-ALTERNATE"),
    "Authentication incident-directory alternate",
  );
  const targetEmail = uniqueEmail("authentication-e110-target");
  const targetPassword = "AuthenticationE110Pass!";
  const targetUser = await createLocalUser(workerAdminRequest, {
    email: targetEmail,
    display_name: "Authentication E110 Target",
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
  await new AuthGateway(page).goto();
  await new AuthGateway(page).login(targetEmail, targetPassword);
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication ordinary shell",
    email: targetEmail,
    purpose: "authentication e110 stale incident selection target",
    userId: targetUser.user_id,
  });

  await new IncidentDirectory(page).openIncident(selectedIncidentId);
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

  await new IncidentDirectory(page).openIncident(alternateIncidentId);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectCurrentIncidentRole(page, "Current incident role: viewer");
});

test("observes current-role authorization on a stale reviewer edit through the public incident error envelope", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("ACCESS-BOOTSTRAP"),
    "Authentication access-control",
  );
  const targetEmail = uniqueEmail("authentication-e111-reviewer");
  const targetPassword = "AuthenticationE111Pass!";
  const targetUser = await createLocalUser(workerAdminRequest, {
    email: targetEmail,
    display_name: "Authentication E111 Reviewer",
    initial_password: targetPassword,
    mfa_required: false,
  });
  await createIncidentMembership(page, incidentId, targetEmail, "reviewer");

  await clearBrowserSession(page);
  await new AuthGateway(page).goto();
  await new AuthGateway(page).login(targetEmail, targetPassword);
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication ordinary shell",
    email: targetEmail,
    purpose: "authentication e111 reviewer stale edit",
    userId: targetUser.user_id,
  });

  await new IncidentDirectory(page).openIncident(incidentId);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectCurrentIncidentRole(page, "Current incident role: reviewer");
  await openIncidentControls(page, "memberships");
  await expect(
    page.getByTestId(incidentMembershipRoleDisplayTestId(targetUser.user_id)),
  ).toHaveText("reviewer");
  await openIncidentControls(page, "incident-fields");
  await expect(
    page.getByTestId(incidentAdministrationTestId("patch-button")),
  ).toBeVisible();

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
  await new IncidentDirectory(page).patchIncidentFields({
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
  await expect(
    page.getByTestId(incidentAdministrationTestId("admin-error-code")),
  ).toHaveText("authorization_denied");
  await expect(page.locator("body")).not.toContainText("request_id");
  await expect(page.locator("body")).not.toContainText("traceback");

  await page.reload();
  await expectCurrentIncidentRole(page, "Current incident role: editor");
  await openIncidentControls(page, "incident-fields");
  await expect(
    page.getByTestId(incidentAdministrationTestId("patch-readonly-note")),
  ).toBeVisible();
});

test("returns a revoked target browser to login and allows re-authentication with unchanged incident membership", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("ACCESS-SESSION-RECOVERY"),
    "Authentication access-control",
  );
  const alternateIncidentId = await createIncident(
    page,
    uniqueIncidentKey("ACCESS-SESSION-ALTERNATE"),
    "Authentication access-control alternate",
  );
  const targetEmail = uniqueEmail("authentication-e112-target");
  const targetPassword = "AuthenticationE112Pass!";
  const targetUser = await createLocalUser(workerAdminRequest, {
    email: targetEmail,
    display_name: "Authentication E112 Target",
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
  await new AuthGateway(page).goto();
  await new AuthGateway(page).login(targetEmail, targetPassword);
  await expectLandingAccountSession(page);
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toBeVisible();
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication ordinary shell",
    email: targetEmail,
    purpose: "authentication e112 target before revoke-all",
    userId: targetUser.user_id,
  });

  await revokeAllSessions(
    workerAdminRequest,
    targetUser.user_id,
    "authentication e112 explicit revoke-all",
  );

  const revokedSessionResponse = waitForPublicAPIResponse(page, {
    method: "GET",
    path: "/api/v1/auth/session",
    status: 401,
  });
  await new IncidentDirectory(page).refresh();
  const sessionResponse = await revokedSessionResponse;
  await expect(sessionResponse.json()).resolves.toMatchObject({
    error: {
      code: "session_required",
    },
  });
  await expect(page.getByTestId(authTestId("login-username"))).toBeVisible();
  await expect(page.getByTestId(authTestId("shell-message"))).toContainText(
    "Sign in again",
  );

  await new AuthGateway(page).login(targetEmail, targetPassword);
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication ordinary shell",
    email: targetEmail,
    purpose: "authentication e112 target after revoke-all re-auth",
    userId: targetUser.user_id,
  });
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toBeVisible();

  await new IncidentDirectory(page).openIncident(incidentId);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectCurrentIncidentRole(page, "Current incident role: viewer");
});

test("Verify ordinary login, incident entry, and current-role refresh stay on public browser routes.", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const email = uniqueEmail("authentication-auth-session");
  const password = "AuthenticationAUTHSESSIONPass!";
  const user = await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Authentication end-to-end.incident-selection.row-01",
    initial_password: password,
    mfa_required: false,
  });
  const incidentKey = uniqueIncidentKey("AUTHSESSION");
  const incidentTitle = "end-to-end.incident-selection.row-01 incident entry";
  const incidentId = await createIncident(page, incidentKey, incidentTitle);
  await createIncidentMembership(page, incidentId, email, "admin");

  await clearBrowserSession(page);
  await new AuthGateway(page).goto();
  await expect(page.getByTestId(authTestId("login-username"))).toBeVisible();
  const loginResponse = waitForPublicAPIResponse(page, {
    method: "POST",
    path: "/api/v1/auth/login",
    status: 200,
  });
  await new AuthGateway(page).login(email, password);
  await loginResponse;
  await expectLandingAccountSession(page);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication ordinary shell",
    email,
    purpose: "end-to-end.incident-selection.row-01 ordinary login",
    userId: user.user_id,
  });

  await new IncidentDirectory(page).goto();
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toContainText(incidentTitle);
  await new IncidentDirectory(page).openIncident(incidentId);
  await expect(page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectCurrentIncidentRole(page, "Current incident role: admin");
  await openIncidentControls(page);
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-key")),
  ).toHaveText(incidentKey);

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
      client_txn_id: uniqueTxn("authentication-incident"),
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
        client_txn_id: uniqueTxn("authentication-membership"),
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
