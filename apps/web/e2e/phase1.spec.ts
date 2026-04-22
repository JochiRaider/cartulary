import type { APIRequestContext, Page } from "@playwright/test";

import {
  authenticatedRequestContextFromStorageState,
  createLocalUser,
  listUsers,
  loadUser,
  patchUser,
  resetUserTotp,
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
  await phase1.login(email, password);

  await expect(page.getByTestId("account-session-provider-type")).toHaveText(
    "local",
  );
  await expect(page.getByTestId("account-session-user-id")).not.toHaveText("");
  await expect(
    page.getByTestId("account-session-authenticated-at"),
  ).not.toHaveText("");
  await expect(
    page.getByTestId("account-session-idle-expires-at"),
  ).not.toHaveText("");
  await expect(
    page.getByTestId("account-session-absolute-expires-at"),
  ).not.toHaveText("");
  await expect(
    page.getByTestId("account-session-session-expires-at"),
  ).not.toHaveText("");
  await expect(
    page.locator('[data-testid="account-session-memberships"] li'),
  ).toHaveCount(0);
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

  await phase1.login(email, password);
  await expect(page.getByTestId("auth-error-code")).toHaveText("mfa_required");

  await phase1.login(email, password, "000000");
  await expect(page.getByTestId("auth-error-code")).toHaveText(
    "invalid_second_factor",
  );

  await phase1.login(email, password, generateTotpCode(secretBase32));
  await expect(page.getByTestId("account-session-provider-type")).toHaveText(
    "local",
  );
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
  await phase1.login(email, "WrongPassword1!");
  await expect(page.getByTestId("auth-error-code")).toHaveText(
    "invalid_credentials",
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

  await expect(page.getByTestId("admin-status")).toHaveText(
    "Created local user",
  );
  await expect(page.getByTestId("admin-target-user-version")).toHaveText("1");

  await page
    .getByTestId("admin-patch-display-name")
    .fill("Phase 1 E105 Patched");
  await phase1.patchTargetUser();
  await expect(page.getByTestId("admin-target-user-version")).toHaveText("2");

  await page.getByTestId("admin-patch-base-version").fill("1");
  await phase1.patchTargetUser();
  await expect(page.getByTestId("admin-error-code")).toHaveText(
    "user_version_conflict",
  );

  const currentAdminID = await page
    .getByTestId("account-session-user-id")
    .textContent();
  if (!currentAdminID) {
    throw new Error("missing current admin user id");
  }
  const remainingAdmins = (await listUsers(workerAdminRequest)).filter(
    (user) =>
      user.user_id !== currentAdminID &&
      user.is_active &&
      user.is_deployment_admin,
  );
  for (const user of remainingAdmins) {
    await patchUser(workerAdminRequest, user.user_id, {
      base_user_version: user.user_version,
      is_deployment_admin: false,
    });
  }
  await phase1.loadTargetUser(currentAdminID);
  await expect(
    page.getByTestId("admin-patch-is-deployment-admin"),
  ).toBeChecked();
  await expect(page.getByTestId("admin-patch-is-active")).toBeChecked();

  await phase1.setCheckbox("admin-patch-is-deployment-admin", false);
  await phase1.patchTargetUser();
  await expect(page.getByTestId("admin-error-code")).toHaveText(
    "last_deployment_admin",
  );

  await phase1.loadTargetUser(currentAdminID);
  await expect(
    page.getByTestId("admin-patch-is-deployment-admin"),
  ).toBeChecked();
  await phase1.setCheckbox("admin-patch-is-active", false);
  await phase1.patchTargetUser();
  await expect(page.getByTestId("admin-error-code")).toHaveText(
    "last_deployment_admin",
  );

  for (const user of remainingAdmins) {
    const reloaded = await loadUser(workerAdminRequest, user.user_id);
    if (reloaded.is_deployment_admin) {
      continue;
    }
    await patchUser(workerAdminRequest, user.user_id, {
      base_user_version: reloaded.user_version,
      is_deployment_admin: true,
    });
  }
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
  await expect(page.getByTestId("auth-error-code")).toHaveText(
    "mfa_setup_required",
  );
  await expect(page.getByTestId("auth-bootstrap-token")).not.toHaveText("");
  expect(await hasSessionCookie(page)).toBeFalsy();

  await phase1.beginBootstrapEnrollment();
  await expect(page.getByTestId("auth-status")).toHaveText(
    "Began TOTP enrollment",
  );
  const secretBase32 = await phase1.requireText("auth-bootstrap-secret-base32");

  await phase1.completeBootstrapEnrollment(generateTotpCode(secretBase32));
  await expect(page.getByTestId("auth-status")).toHaveText(
    "TOTP enrollment completed. Sign in with your TOTP code.",
  );
  expect(await hasSessionCookie(page)).toBeFalsy();

  await phase1.login(email, password);
  await expect(page.getByTestId("auth-error-code")).toHaveText("mfa_required");

  await phase1.login(email, password, generateTotpCode(secretBase32));
  await expect(page.getByTestId("account-session-provider-type")).toHaveText(
    "local",
  );
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
  await expect(page.getByTestId("auth-error-code")).toHaveText(
    "mfa_setup_required",
  );
  await expect(page.getByTestId("auth-bootstrap-token")).not.toHaveText("");
  expect(await hasSessionCookie(page)).toBeFalsy();

  await phase1.beginBootstrapEnrollment();
  await expect(page.getByTestId("auth-status")).toHaveText(
    "Began TOTP enrollment",
  );
  const replacementSecretBase32 = await phase1.requireText(
    "auth-bootstrap-secret-base32",
  );

  await phase1.completeBootstrapEnrollment(
    generateTotpCode(replacementSecretBase32),
  );
  await expect(page.getByTestId("auth-status")).toHaveText(
    "TOTP enrollment completed. Sign in with your TOTP code.",
  );
  expect(await hasSessionCookie(page)).toBeFalsy();

  await phase1.login(email, password);
  await expect(page.getByTestId("auth-error-code")).toHaveText("mfa_required");

  await phase1.login(
    email,
    password,
    generateTotpCode(replacementSecretBase32),
  );
  await expect(page.getByTestId("account-session-provider-type")).toHaveText(
    "local",
  );
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
  await expect(page.getByTestId("account-session-provider-type")).toHaveText(
    "local",
  );
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
  await expect(page.getByTestId("account-error-code")).toHaveText(
    "invalid_current_password",
  );

  await phase1.changePassword(password, "Phase1E107Changed!", "");
  await expect(page.getByTestId("account-error-code")).toHaveText(
    "invalid_second_factor",
  );

  await phase1.changePassword(
    password,
    "Phase1E107Changed!",
    generateTotpCode(secretBase32),
  );
  await expect(page.getByTestId("auth-login-username")).toBeVisible();

  await phase1.login(email, password, generateTotpCode(secretBase32));
  await expect(page.getByTestId("auth-error-code")).toHaveText(
    "invalid_credentials",
  );

  await phase1.login(
    email,
    "Phase1E107Changed!",
    generateTotpCode(secretBase32),
  );
  await expect(page.getByTestId("account-session-provider-type")).toHaveText(
    "local",
  );
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
  await expect(page.getByTestId("account-session-provider-type")).toHaveText(
    "local",
  );
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email: incidentAdminEmail,
    purpose: "phase1 e108 incident-admin login",
    userId: incidentAdminUser.user_id,
  });
  await expect(page.getByTestId("admin-access-note")).toContainText(
    "Incident-admin membership alone does not unlock these controls",
  );
  await expect(page.getByTestId("admin-password-reset")).toHaveCount(0);

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
  await page.getByTestId("admin-new-password").fill("Phase1E108Changed!");
  await page.getByTestId("admin-reason").fill("deployment admin action");

  await page.getByTestId("admin-password-reset").click();
  await expect(page.getByTestId("admin-status")).toHaveText(
    "Reset user password",
  );

  await page.getByTestId("admin-totp-reset").click();
  await expect(page.getByTestId("admin-status")).toHaveText("Reset user TOTP");

  await page.getByTestId("admin-revoke-all").click();
  await expect(page.getByTestId("admin-status")).toHaveText(
    "Revoked every user session",
  );

  await clearBrowserSession(page);
  await phase1.goto();
  await phase1.login(incidentAdminEmail, incidentAdminPassword);
  await expect(page.getByTestId("account-session-provider-type")).toHaveText(
    "local",
  );
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email: incidentAdminEmail,
    purpose: "phase1 e108 incident-admin login after admin actions",
    userId: incidentAdminUser.user_id,
  });
  await expect(page.getByTestId("account-session-user-id")).toHaveText(
    incidentAdminUser.user_id,
  );
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

async function expectUnauthorizedCredentialAction(
  authRequests: APIRequestContext,
  path: string,
  data: Record<string, unknown>,
) {
  const response = await authRequests.post(path, { data });
  expect(response.status()).toBe(401);
  await expect(response.json()).resolves.toMatchObject({
    error: {
      code: "session_required",
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
