import { createHmac } from "node:crypto";
import { existsSync, readFileSync, unlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { expect, test, type APIResponse, type Page } from "@playwright/test";

const bootstrapEmail = "dev-admin@example.test";
const bootstrapPassword = "DevBootstrap1!";
const sessionCookieName = "cartulary_session";
const csrfCookieName = "cartulary_csrf";
const csrfHeaderName = "X-CSRF-Token";
const apiBase = "http://127.0.0.1:8080";
const adminTotpCachePath = join(tmpdir(), "cartulary-phase2-admin-totp.txt");

let adminTotpSecretBase32: string | null = null;
let adminCookies: { session: string; csrf: string } | null = null;

test("E-1-01 logs in as a local user and inspects the singleton session resource", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const email = uniqueEmail("phase1-e101");
  const password = "Phase1E101Pass!";
  await createLocalUser(page, {
    email,
    display_name: "Phase 1 E101",
    initial_password: password,
    mfa_required: false,
  });

  await clearBrowserSession(page);
  await page.goto("/");
  await signInThroughHarness(page, email, password);

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
});

test("E-1-02 requires MFA when the account has an active factor, rejects wrong codes, and accepts a valid TOTP code", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const email = uniqueEmail("phase1-e102");
  const password = "Phase1E102Pass!";
  await createLocalUser(page, {
    email,
    display_name: "Phase 1 E102",
    initial_password: password,
    mfa_required: true,
  });
  const secretBase32 = await enrollTotpViaBootstrap(page, email, password);

  await clearBrowserSession(page);
  await page.goto("/");

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
  await page.getByTestId("phase1-login").click();
  await expect(page.getByTestId("phase1-session-provider-type")).toHaveText(
    "local",
  );
});

test("E-1-03 rejects invalid credentials without issuing a session cookie", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const email = uniqueEmail("phase1-e103");
  await createLocalUser(page, {
    email,
    display_name: "Phase 1 E103",
    initial_password: "Phase1E103Pass!",
    mfa_required: false,
  });

  await clearBrowserSession(page);
  await page.goto("/");
  await signInThroughHarness(page, email, "WrongPassword1!");
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "invalid_credentials",
  );
  expect(await hasSessionCookie(page)).toBeFalsy();
});

test("E-1-04 forces the idle expiry boundary and requires a fresh login afterwards", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const email = uniqueEmail("phase1-e104");
  const password = "Phase1E104Pass!";
  await createLocalUser(page, {
    email,
    display_name: "Phase 1 E104",
    initial_password: password,
    mfa_required: false,
  });

  await clearBrowserSession(page);
  await page.goto("/");
  await signInThroughHarness(page, email, password);

  await page.getByTestId("phase1-expire-session").click();
  await page.getByTestId("phase1-refresh-session").click();
  await expect(page.getByTestId("phase1-session-error-code")).toHaveText(
    "session_required",
  );

  await signInThroughHarness(page, email, password);
  await expect(page.getByTestId("phase1-session-provider-type")).toHaveText(
    "local",
  );
});

test("E-1-05 lets deployment admins create and patch users, rejects stale versions, and preserves the last-admin guard", async ({
  page,
}) => {
  await ensureAdminSession(page);
  await page.goto("/");
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
  await page
    .getByTestId("phase1-admin-target-user-id-input")
    .fill(currentAdminID);
  await page.getByTestId("phase1-admin-load-user").click();
  await setCheckbox(page, "phase1-admin-patch-is-deployment-admin", false);
  await page.getByTestId("phase1-admin-patch-user").click();
  await expect(page.getByTestId("phase1-last-error-code")).toHaveText(
    "last_deployment_admin",
  );
});

test("E-1-06 follows the bootstrap-token enrollment sequence and proves first-time completion alone issues no session", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const email = uniqueEmail("phase1-e106");
  const password = "Phase1E106Pass!";
  await createLocalUser(page, {
    email,
    display_name: "Phase 1 E106",
    initial_password: password,
    mfa_required: true,
  });

  await clearBrowserSession(page);
  await page.goto("/");
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
  await page.getByTestId("phase1-login").click();
  await expect(page.getByTestId("phase1-session-provider-type")).toHaveText(
    "local",
  );
});

test("E-1-07 requires the current password and current TOTP code, revokes the session immediately, and requires re-login with the new password", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const email = uniqueEmail("phase1-e107");
  const password = "Phase1E107Pass!";
  await createLocalUser(page, {
    email,
    display_name: "Phase 1 E107",
    initial_password: password,
    mfa_required: true,
  });
  const secretBase32 = await enrollTotpViaBootstrap(page, email, password);

  await clearBrowserSession(page);
  await page.goto("/");
  await signInThroughHarness(
    page,
    email,
    password,
    generateTotpCode(secretBase32),
  );

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

  await signInThroughHarness(
    page,
    email,
    "Phase1E107Changed!",
    generateTotpCode(secretBase32),
  );
  await expect(page.getByTestId("phase1-session-provider-type")).toHaveText(
    "local",
  );
});

test("E-1-08 keeps credential actions deployment-admin only and denies the same routes to a non-deployment-admin incident admin", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const targetEmail = uniqueEmail("phase1-e108-target");
  const targetPassword = "Phase1E108Pass!";
  const targetUser = await createLocalUser(page, {
    email: targetEmail,
    display_name: "Phase 1 E108 Target",
    initial_password: targetPassword,
    mfa_required: true,
  });
  await enrollTotpViaBootstrap(page, targetEmail, targetPassword);

  const incidentAdminEmail = uniqueEmail("phase1-e108-incident-admin");
  const incidentAdminPassword = "Phase1E108Incident!";
  const incidentAdminUser = await createLocalUser(page, {
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
  await page.goto("/");
  await signInThroughHarness(page, incidentAdminEmail, incidentAdminPassword);
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

  await ensureAdminSession(page);
  await page.goto("/");
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
  await page.goto("/");
  await signInThroughHarness(page, incidentAdminEmail, incidentAdminPassword);
  await expect(page.getByTestId("phase1-session-user-id")).toHaveText(
    incidentAdminUser.user_id,
  );
});

async function ensureAdminSession(page: Page) {
  if (adminCookies !== null) {
    await applyCookies(page, adminCookies.session, adminCookies.csrf);
    return;
  }

  if (adminTotpSecretBase32 === null) {
    adminTotpSecretBase32 = loadCachedAdminTotpSecret();
  }

  const loginResponse = await page.request.post(
    `${apiBase}/api/v1/auth/login`,
    {
      data: {
        username: bootstrapEmail,
        password: bootstrapPassword,
      },
    },
  );
  if (loginResponse.ok()) {
    adminCookies = {
      session: requireCookie(loginResponse, sessionCookieName),
      csrf: requireCookie(loginResponse, csrfCookieName),
    };
    await applyCookies(page, adminCookies.session, adminCookies.csrf);
    return;
  }

  if (loginResponse.status() === 401) {
    const loginBody = (await loginResponse.json()) as {
      error: { code: string; details: { bootstrap_token?: string } };
    };
    if (loginBody.error.code === "mfa_setup_required") {
      clearCachedAdminTotpSecret();
      const bootstrapToken = loginBody.error.details.bootstrap_token;
      if (!bootstrapToken) {
        throw new Error("missing bootstrap_token");
      }

      const beginResponse = await page.request.post(
        `${apiBase}/api/v1/auth/mfa/totp/begin`,
        {
          headers: { Authorization: `Bearer ${bootstrapToken}` },
          data: { client_txn_id: uniqueTxn("phase1-admin-totp-begin") },
        },
      );
      expect(beginResponse.ok()).toBeTruthy();
      const beginBody = (await beginResponse.json()) as {
        data: { enrollment_id: string; totp_setup: { secret_base32: string } };
      };
      adminTotpSecretBase32 = beginBody.data.totp_setup.secret_base32;
      cacheAdminTotpSecret(adminTotpSecretBase32);

      const completeResponse = await page.request.post(
        `${apiBase}/api/v1/auth/mfa/totp/complete`,
        {
          headers: { Authorization: `Bearer ${bootstrapToken}` },
          data: {
            client_txn_id: uniqueTxn("phase1-admin-totp-complete"),
            enrollment_id: beginBody.data.enrollment_id,
            code: generateTotpCode(adminTotpSecretBase32),
          },
        },
      );
      expect(completeResponse.ok()).toBeTruthy();
    }
  }

  if (adminTotpSecretBase32 === null) {
    throw new Error("missing cached admin TOTP secret");
  }

  const secondFactorLogin = await page.request.post(
    `${apiBase}/api/v1/auth/login`,
    {
      data: {
        username: bootstrapEmail,
        password: bootstrapPassword,
        second_factor: {
          kind: "totp",
          assertion: {
            code: generateTotpCode(adminTotpSecretBase32),
          },
        },
      },
    },
  );
  if (!secondFactorLogin.ok()) {
    clearCachedAdminTotpSecret();
    throw new Error(
      `second factor login failed: ${await secondFactorLogin.text()}`,
    );
  }

  adminCookies = {
    session: requireCookie(secondFactorLogin, sessionCookieName),
    csrf: requireCookie(secondFactorLogin, csrfCookieName),
  };
  await applyCookies(page, adminCookies.session, adminCookies.csrf);
}

async function signInThroughHarness(
  page: Page,
  email: string,
  password: string,
  totpCode = "",
) {
  await page.getByTestId("phase1-login-username").fill(email);
  await page.getByTestId("phase1-login-password").fill(password);
  await page.getByTestId("phase1-login-totp-code").fill(totpCode);
  await page.getByTestId("phase1-login").click();
}

async function createLocalUser(
  page: Page,
  options: {
    email: string;
    display_name: string;
    initial_password: string;
    mfa_required?: boolean;
    is_deployment_admin?: boolean;
  },
) {
  const response = await page.request.post(`${apiBase}/api/v1/users`, {
    headers: await csrfHeaders(page),
    data: {
      client_txn_id: uniqueTxn("phase1-user"),
      auth_kind: "local",
      email: options.email,
      display_name: options.display_name,
      initial_password: options.initial_password,
      mfa_required: options.mfa_required ?? true,
      is_deployment_admin: options.is_deployment_admin ?? false,
    },
  });
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: { user_id: string } }).data;
}

async function enrollTotpViaBootstrap(
  page: Page,
  email: string,
  password: string,
) {
  const loginResponse = await page.request.post(
    `${apiBase}/api/v1/auth/login`,
    {
      data: {
        username: email,
        password,
      },
    },
  );
  expect(loginResponse.status()).toBe(401);
  const loginBody = (await loginResponse.json()) as {
    error: { code: string; details: { bootstrap_token: string } };
  };
  expect(loginBody.error.code).toBe("mfa_setup_required");
  const bootstrapToken = loginBody.error.details.bootstrap_token;

  const beginResponse = await page.request.post(
    `${apiBase}/api/v1/auth/mfa/totp/begin`,
    {
      headers: { Authorization: `Bearer ${bootstrapToken}` },
      data: { client_txn_id: uniqueTxn("phase1-bootstrap-begin") },
    },
  );
  expect(beginResponse.ok()).toBeTruthy();
  const beginBody = (await beginResponse.json()) as {
    data: { enrollment_id: string; totp_setup: { secret_base32: string } };
  };
  const secretBase32 = beginBody.data.totp_setup.secret_base32;

  const completeResponse = await page.request.post(
    `${apiBase}/api/v1/auth/mfa/totp/complete`,
    {
      headers: { Authorization: `Bearer ${bootstrapToken}` },
      data: {
        client_txn_id: uniqueTxn("phase1-bootstrap-complete"),
        enrollment_id: beginBody.data.enrollment_id,
        code: generateTotpCode(secretBase32),
      },
    },
  );
  expect(completeResponse.ok()).toBeTruthy();
  return secretBase32;
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

async function csrfHeaders(page: Page) {
  const cookies = await page.context().cookies(apiBase);
  const csrfCookie = cookies.find((cookie) => cookie.name === csrfCookieName);
  if (!csrfCookie) {
    throw new Error("missing CSRF cookie");
  }
  return {
    [csrfHeaderName]: csrfCookie.value,
  };
}

async function applyCookies(page: Page, session: string, csrf: string) {
  await page.context().addCookies([
    {
      name: sessionCookieName,
      value: session,
      domain: "127.0.0.1",
      path: "/",
      httpOnly: true,
      sameSite: "Lax",
    },
    {
      name: csrfCookieName,
      value: csrf,
      domain: "127.0.0.1",
      path: "/",
      sameSite: "Lax",
    },
  ]);
}

async function clearBrowserSession(page: Page) {
  adminCookies = null;
  await page.context().clearCookies();
}

async function hasSessionCookie(page: Page) {
  const cookies = await page.context().cookies(apiBase);
  return cookies.some((cookie) => cookie.name === sessionCookieName);
}

async function requireText(page: Page, testId: string) {
  const value = (await page.getByTestId(testId).textContent())?.trim() ?? "";
  if (value === "") {
    throw new Error(`missing text for ${testId}`);
  }
  return value;
}

async function setCheckbox(page: Page, testId: string, checked: boolean) {
  const checkbox = page.getByTestId(testId);
  if ((await checkbox.isChecked()) !== checked) {
    await checkbox.click();
  }
}

function requireCookie(response: APIResponse, name: string) {
  for (const header of response.headersArray()) {
    if (header.name.toLowerCase() !== "set-cookie") {
      continue;
    }
    const [cookiePair] = header.value.split(";", 1);
    if (!cookiePair) {
      continue;
    }
    const [cookieName, cookieValue] = cookiePair.split("=", 2);
    if (cookieName === name && cookieValue) {
      return cookieValue;
    }
  }
  throw new Error(`missing ${name} cookie on response`);
}

function loadCachedAdminTotpSecret() {
  if (!existsSync(adminTotpCachePath)) {
    return null;
  }

  const secret = readFileSync(adminTotpCachePath, "utf8").trim();
  return secret === "" ? null : secret;
}

function cacheAdminTotpSecret(secretBase32: string) {
  writeFileSync(adminTotpCachePath, `${secretBase32}\n`, "utf8");
}

function clearCachedAdminTotpSecret() {
  adminTotpSecretBase32 = null;
  if (!existsSync(adminTotpCachePath)) {
    return;
  }
  unlinkSync(adminTotpCachePath);
}

function uniqueTxn(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`;
}

function uniqueEmail(prefix: string) {
  return `${prefix}-${Date.now().toString(36)}@example.test`;
}

function uniqueIncidentKey(prefix: string) {
  return `IR-${prefix}-${Date.now().toString(36).toUpperCase()}`;
}

function generateTotpCode(secretBase32: string) {
  const secret = decodeBase32(secretBase32);
  const counter = Math.floor(Date.now() / 1000 / 30);
  const counterBuffer = Buffer.alloc(8);
  counterBuffer.writeBigUInt64BE(BigInt(counter));
  const digest = createHmac("sha1", secret).update(counterBuffer).digest();
  const offset = digest[digest.length - 1] & 0x0f;
  const code =
    ((digest[offset] & 0x7f) << 24) |
    ((digest[offset + 1] & 0xff) << 16) |
    ((digest[offset + 2] & 0xff) << 8) |
    (digest[offset + 3] & 0xff);
  return String(code % 1_000_000).padStart(6, "0");
}

function decodeBase32(input: string) {
  const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567";
  const normalized = input.replace(/=+$/u, "").toUpperCase();
  let bits = "";
  for (const character of normalized) {
    const index = alphabet.indexOf(character);
    if (index < 0) {
      throw new Error(`invalid base32 character: ${character}`);
    }
    bits += index.toString(2).padStart(5, "0");
  }

  const bytes: number[] = [];
  for (let index = 0; index + 8 <= bits.length; index += 8) {
    bytes.push(Number.parseInt(bits.slice(index, index + 8), 2));
  }
  return Buffer.from(bytes);
}
