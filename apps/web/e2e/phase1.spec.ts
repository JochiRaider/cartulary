import type { Page, StorageState } from "@playwright/test";
import {
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
  storageStateFromSetCookieHeaders,
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";

test("E-1-01 logs in as a local user and inspects the singleton session resource", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const email = uniqueEmail("phase1-e101");
  const password = "Phase1E101Pass!";
  const user = await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Phase 1 E101",
    initial_password: password,
    mfa_required: false,
  });

  await clearBrowserSession(page);
  await page.goto("/?debug=harness");
  const loginState = await signInThroughHarness(page, email, password);

  await expect(page.getByTestId("phase1-session-provider-type")).toHaveText(
    "local",
  );
  await expect(page.getByTestId("phase1-session-user-id")).not.toHaveText("");
  await expect(
    page.getByTestId("phase1-session-authenticated-at"),
  ).not.toHaveText("");
  await expect(
    page.getByTestId("phase1-session-idle-expires-at"),
  ).not.toHaveText("");
  await expect(
    page.getByTestId("phase1-session-absolute-expires-at"),
  ).not.toHaveText("");
  await expect(
    page.getByTestId("phase1-session-session-expires-at"),
  ).not.toHaveText("");
  await expect(
    page.locator('[data-testid="phase1-session-memberships"] li'),
  ).toHaveCount(0);
  await trackCurrentSession(sessionTracker, {
    email,
    storageState: loginState,
    purpose: "phase1 e101 successful harness login",
    userId: user.user_id,
  });
});

test("E-1-02 requires MFA when the account has an active factor, rejects wrong codes, and accepts a valid TOTP code", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
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
  await page.goto("/?debug=harness");

  await page.getByTestId("phase1-login-username").fill(email);
  await page.getByTestId("phase1-login-password").fill(password);
  await page.getByTestId("phase1-login-totp-code").fill("");
  await page.getByTestId("phase1-login").click();
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "mfa_required",
  );

  await page.getByTestId("phase1-login-totp-code").fill("000000");
  await page.getByTestId("phase1-login").click();
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "invalid_second_factor",
  );

  await page
    .getByTestId("phase1-login-totp-code")
    .fill(generateTotpCode(secretBase32));
  const loginState = await clickHarnessLogin(page);
  await expect(page.getByTestId("phase1-session-provider-type")).toHaveText(
    "local",
  );
  await trackCurrentSession(sessionTracker, {
    email,
    storageState: loginState,
    purpose: "phase1 e102 successful harness login",
    userId: user.user_id,
  });
});

test("E-1-03 rejects invalid credentials without issuing a session cookie", async ({
  page,
  workerAdminRequest,
}) => {
  const email = uniqueEmail("phase1-e103");
  await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Phase 1 E103",
    initial_password: "Phase1E103Pass!",
    mfa_required: false,
  });

  await clearBrowserSession(page);
  await page.goto("/?debug=harness");
  await signInThroughHarness(page, email, "WrongPassword1!");
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "invalid_credentials",
  );
  expect(await hasSessionCookie(page)).toBeFalsy();
});

test("E-1-04 forces the idle expiry boundary and requires a fresh login afterwards", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const email = uniqueEmail("phase1-e104");
  const password = "Phase1E104Pass!";
  const user = await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Phase 1 E104",
    initial_password: password,
    mfa_required: false,
  });

  await clearBrowserSession(page);
  await page.goto("/?debug=harness");
  const initialLoginState = await signInThroughHarness(page, email, password);
  await trackCurrentSession(sessionTracker, {
    email,
    storageState: initialLoginState,
    purpose: "phase1 e104 initial harness login",
    userId: user.user_id,
  });

  await page.getByTestId("phase1-expire-session").click();
  await page.getByTestId("phase1-refresh-session").click();
  await expect(page.getByTestId("phase1-session-error-code")).toHaveText(
    "session_required",
  );

  const reloginState = await signInThroughHarness(page, email, password);
  await expect(page.getByTestId("phase1-session-provider-type")).toHaveText(
    "local",
  );
  await trackCurrentSession(sessionTracker, {
    email,
    storageState: reloginState,
    purpose: "phase1 e104 re-login after expiry",
    userId: user.user_id,
  });
});

test("E-1-05 lets deployment admins create and patch users, rejects stale versions, and proves the last-admin guard through the harness", async ({
  page,
  workerAdminRequest,
}) => {
  await page.goto("/?debug=harness");
  const createEmail = uniqueEmail("phase1-e105");

  await page.getByTestId("phase1-admin-create-email").fill(createEmail);
  await page
    .getByTestId("phase1-admin-create-display-name")
    .fill("Phase 1 E105 User");
  await page
    .getByTestId("phase1-admin-create-password")
    .fill("Phase1E105Pass!");
  await setCheckbox(page, "phase1-admin-create-mfa-required", false);
  await page.getByTestId("phase1-admin-create-user").click();

  await expect(page.getByTestId("phase1-status")).toHaveText(
    "Created local user",
  );
  await expect(page.getByTestId("phase1-admin-target-user-version")).toHaveText(
    "1",
  );

  await page
    .getByTestId("phase1-admin-patch-display-name")
    .fill("Phase 1 E105 Patched");
  await page.getByTestId("phase1-admin-patch-user").click();
  await expect(page.getByTestId("phase1-admin-target-user-version")).toHaveText(
    "2",
  );

  await page.getByTestId("phase1-admin-patch-base-version").fill("1");
  await page.getByTestId("phase1-admin-patch-user").click();
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "user_version_conflict",
  );

  const currentAdminID = await page
    .getByTestId("phase1-session-user-id")
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
  await page
    .getByTestId("phase1-admin-target-user-id-input")
    .fill(currentAdminID);
  await page.getByTestId("phase1-admin-load-user").click();
  await expect(page.getByTestId("phase1-status")).toHaveText(
    "Loaded target user",
  );
  await expect(
    page.getByTestId("phase1-admin-patch-is-deployment-admin"),
  ).toBeChecked();
  await expect(page.getByTestId("phase1-admin-patch-is-active")).toBeChecked();

  await setCheckbox(page, "phase1-admin-patch-is-deployment-admin", false);
  await page.getByTestId("phase1-admin-patch-user").click();
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "last_deployment_admin",
  );

  await page.getByTestId("phase1-admin-load-user").click();
  await expect(
    page.getByTestId("phase1-admin-patch-is-deployment-admin"),
  ).toBeChecked();
  await setCheckbox(page, "phase1-admin-patch-is-active", false);
  await page.getByTestId("phase1-admin-patch-user").click();
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
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

test("E-1-06 follows the bootstrap-token enrollment sequence for first enrollment and post-reset re-enrollment, and proves completion alone issues no session", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const email = uniqueEmail("phase1-e106");
  const password = "Phase1E106Pass!";
  const user = await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Phase 1 E106",
    initial_password: password,
    mfa_required: true,
  });

  await clearBrowserSession(page);
  await page.goto("/?debug=harness");
  await page.getByTestId("phase1-login-username").fill(email);
  await page.getByTestId("phase1-login-password").fill(password);
  await page.getByTestId("phase1-login-totp-code").fill("");
  await page.getByTestId("phase1-login").click();
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "mfa_setup_required",
  );
  await expect(page.getByTestId("phase1-bootstrap-token")).not.toHaveText("");
  expect(await hasSessionCookie(page)).toBeFalsy();

  await page.getByTestId("phase1-totp-auth-mode").selectOption("bootstrap");
  await page.getByTestId("phase1-totp-begin").click();
  await expect(page.getByTestId("phase1-status")).toHaveText(
    "Began TOTP enrollment",
  );
  const secretBase32 = await requireText(page, "phase1-totp-secret-base32");

  await page
    .getByTestId("phase1-totp-complete-code")
    .fill(generateTotpCode(secretBase32));
  await page.getByTestId("phase1-totp-complete").click();
  await expect(page.getByTestId("phase1-status")).toHaveText(
    "Completed TOTP enrollment",
  );
  expect(await hasSessionCookie(page)).toBeFalsy();

  await page.getByTestId("phase1-refresh-session").click();
  await expect(page.getByTestId("phase1-session-error-code")).toHaveText(
    "session_required",
  );

  await page.getByTestId("phase1-login-totp-code").fill("");
  await page.getByTestId("phase1-login").click();
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "mfa_required",
  );

  await page
    .getByTestId("phase1-login-totp-code")
    .fill(generateTotpCode(secretBase32));
  const loginState = await clickHarnessLogin(page);
  await expect(page.getByTestId("phase1-session-provider-type")).toHaveText(
    "local",
  );
  await trackCurrentSession(sessionTracker, {
    email,
    storageState: loginState,
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
  await page.goto("/?debug=harness");
  await page.getByTestId("phase1-login-username").fill(email);
  await page.getByTestId("phase1-login-password").fill(password);
  await page.getByTestId("phase1-login-totp-code").fill("");
  await page.getByTestId("phase1-login").click();
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "mfa_setup_required",
  );
  await expect(page.getByTestId("phase1-bootstrap-token")).not.toHaveText("");
  expect(await hasSessionCookie(page)).toBeFalsy();

  await page.getByTestId("phase1-totp-auth-mode").selectOption("bootstrap");
  await page.getByTestId("phase1-totp-begin").click();
  await expect(page.getByTestId("phase1-status")).toHaveText(
    "Began TOTP enrollment",
  );
  const replacementSecretBase32 = await requireText(
    page,
    "phase1-totp-secret-base32",
  );

  await page
    .getByTestId("phase1-totp-complete-code")
    .fill(generateTotpCode(replacementSecretBase32));
  await page.getByTestId("phase1-totp-complete").click();
  await expect(page.getByTestId("phase1-status")).toHaveText(
    "Completed TOTP enrollment",
  );
  expect(await hasSessionCookie(page)).toBeFalsy();

  await page.getByTestId("phase1-login-totp-code").fill("");
  await page.getByTestId("phase1-login").click();
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "mfa_required",
  );

  const reloginState = await signInThroughHarness(
    page,
    email,
    password,
    generateTotpCode(replacementSecretBase32),
  );
  await expect(page.getByTestId("phase1-session-provider-type")).toHaveText(
    "local",
  );
  await trackCurrentSession(sessionTracker, {
    email,
    storageState: reloginState,
    purpose: "phase1 e106 post-reset bootstrap login",
    userId: user.user_id,
  });
});

test("E-1-07 requires the current password and current TOTP code, revokes the session immediately, and requires re-login with the new password", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
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
  await page.goto("/?debug=harness");
  const initialLoginState = await signInThroughHarness(
    page,
    email,
    password,
    generateTotpCode(secretBase32),
  );
  await trackCurrentSession(sessionTracker, {
    email,
    storageState: initialLoginState,
    purpose: "phase1 e107 initial login before password change",
    userId: user.user_id,
  });

  await page.getByTestId("phase1-password-current").fill("WrongCurrent1!");
  await page.getByTestId("phase1-password-next").fill("Phase1E107Changed!");
  await page
    .getByTestId("phase1-password-factor-code")
    .fill(generateTotpCode(secretBase32));
  await page.getByTestId("phase1-password-change").click();
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "invalid_current_password",
  );

  await page.getByTestId("phase1-password-current").fill(password);
  await page.getByTestId("phase1-password-factor-code").fill("");
  await page.getByTestId("phase1-password-change").click();
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "invalid_second_factor",
  );

  await page.getByTestId("phase1-password-current").fill(password);
  await page.getByTestId("phase1-password-next").fill("Phase1E107Changed!");
  await page
    .getByTestId("phase1-password-factor-code")
    .fill(generateTotpCode(secretBase32));
  await page.getByTestId("phase1-password-change").click();
  await expect(page.getByTestId("phase1-session-error-code")).toHaveText(
    "session_required",
  );

  await page.getByTestId("phase1-login-password").fill(password);
  await page
    .getByTestId("phase1-login-totp-code")
    .fill(generateTotpCode(secretBase32));
  await page.getByTestId("phase1-login").click();
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "invalid_credentials",
  );

  const changedPasswordLoginState = await signInThroughHarness(
    page,
    email,
    "Phase1E107Changed!",
    generateTotpCode(secretBase32),
  );
  await expect(page.getByTestId("phase1-session-provider-type")).toHaveText(
    "local",
  );
  await trackCurrentSession(sessionTracker, {
    email,
    storageState: changedPasswordLoginState,
    purpose: "phase1 e107 login after password change",
    userId: user.user_id,
  });
});

test("E-1-08 keeps credential actions deployment-admin only and denies the same routes to a non-deployment-admin incident admin", async ({
  page,
  sessionTracker,
  workerAdmin,
  workerAdminRequest,
}) => {
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
  await page.goto("/?debug=harness");
  const deniedLoginState = await signInThroughHarness(
    page,
    incidentAdminEmail,
    incidentAdminPassword,
  );
  await trackCurrentSession(sessionTracker, {
    email: incidentAdminEmail,
    storageState: deniedLoginState,
    purpose: "phase1 e108 incident-admin denied login",
    userId: incidentAdminUser.user_id,
  });
  await page
    .getByTestId("phase1-admin-target-user-id-input")
    .fill(targetUser.user_id);
  await page.getByTestId("phase1-admin-load-user").click();
  await page
    .getByTestId("phase1-admin-new-password")
    .fill("Phase1E108Changed!");
  await page.getByTestId("phase1-admin-reason").fill("incident admin denied");

  await page.getByTestId("phase1-admin-password-reset").click();
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "session_required",
  );
  await page.getByTestId("phase1-admin-totp-reset").click();
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "session_required",
  );
  await page.getByTestId("phase1-admin-revoke-all").click();
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "session_required",
  );

  await restoreTrackedStorageState(page, workerAdmin.storageState);
  await page.goto("/?debug=harness");
  await page
    .getByTestId("phase1-admin-target-user-id-input")
    .fill(targetUser.user_id);
  await page.getByTestId("phase1-admin-load-user").click();
  await page
    .getByTestId("phase1-admin-new-password")
    .fill("Phase1E108Changed!");
  await page.getByTestId("phase1-admin-reason").fill("deployment admin action");

  await page.getByTestId("phase1-admin-password-reset").click();
  await expect(page.getByTestId("phase1-status")).toHaveText(
    "Reset user password",
  );

  await page.getByTestId("phase1-admin-totp-reset").click();
  await expect(page.getByTestId("phase1-status")).toHaveText("Reset user TOTP");

  await page.getByTestId("phase1-admin-revoke-all").click();
  await expect(page.getByTestId("phase1-status")).toHaveText(
    "Revoked every user session",
  );

  await clearBrowserSession(page);
  await page.goto("/?debug=harness");
  const postAdminActionLoginState = await signInThroughHarness(
    page,
    incidentAdminEmail,
    incidentAdminPassword,
  );
  await trackCurrentSession(sessionTracker, {
    email: incidentAdminEmail,
    storageState: postAdminActionLoginState,
    purpose: "phase1 e108 incident-admin login after admin actions",
    userId: incidentAdminUser.user_id,
  });
  await expect(page.getByTestId("phase1-session-user-id")).toHaveText(
    incidentAdminUser.user_id,
  );
});

async function signInThroughHarness(
  page: Page,
  email: string,
  password: string,
  totpCode = "",
) {
  await page.getByTestId("phase1-login-username").fill(email);
  await page.getByTestId("phase1-login-password").fill(password);
  await page.getByTestId("phase1-login-totp-code").fill(totpCode);
  return clickHarnessLogin(page);
}

async function clickHarnessLogin(page: Page) {
  const loginResponsePromise = page.waitForResponse((response) => {
    return (
      response.request().method() === "POST" &&
      new URL(response.url()).pathname === "/api/v1/auth/login"
    );
  });
  await page.getByTestId("phase1-login").click();
  const loginResponse = await loginResponsePromise;
  const setCookieHeaders = await loginResponse.headerValues("set-cookie");
  if (setCookieHeaders.length === 0) {
    return null;
  }
  return storageStateFromSetCookieHeaders(setCookieHeaders);
}

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

async function clearBrowserSession(page: Page) {
  await page.context().clearCookies();
}

async function trackCurrentSession(
  sessionTracker: {
    captureStorageState: (
      storageState: StorageState,
      details: {
        createdBy: string;
        email: string;
        purpose: string;
        userId: string;
      },
    ) => Promise<void>;
  },
  details: {
    email: string;
    purpose: string;
    storageState: StorageState | null;
    userId: string;
  },
) {
  if (details.storageState === null) {
    throw new Error(
      `missing tracked storage state for ${details.userId} (${details.email})`,
    );
  }
  await sessionTracker.captureStorageState(details.storageState, {
    createdBy: "phase1 harness",
    email: details.email,
    purpose: details.purpose,
    userId: details.userId,
  });
}

async function hasSessionCookie(page: Page) {
  const cookies = await page.context().cookies(apiBase);
  return cookies.some((cookie) => cookie.name === sessionCookieName);
}

async function requireText(page: Page, testId: string) {
  const locator = page.getByTestId(testId);
  await expect
    .poll(async () => (await locator.textContent())?.trim() ?? "")
    .not.toBe("");
  const value = (await locator.textContent())?.trim() ?? "";
  if (value === "") {
    throw new Error(`missing text for ${testId}, got "${value}"`);
  }
  return value;
}

async function setCheckbox(page: Page, testId: string, checked: boolean) {
  const checkbox = page.getByTestId(testId);
  await checkbox.setChecked(checked);
  if (checked) {
    await expect(checkbox).toBeChecked();
    return;
  }
  await expect(checkbox).not.toBeChecked();
}
