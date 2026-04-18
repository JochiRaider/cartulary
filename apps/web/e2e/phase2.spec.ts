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

  await page.goto("/");
  await expect(page.getByTestId(`incident-row-${incidentId}`)).toHaveText(incidentKey);
  await page.getByTestId(`select-incident-${incidentId}`).click();
  await expect(page.getByTestId("current-incident-id")).toHaveText(incidentId);
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
  const targetUser = await createLocalUser(page, {
    email: targetEmail,
    display_name: "Phase 2 E203 Member",
    initial_password: targetPassword,
  });
  const incidentId = await createIncident(page, uniqueIncidentKey("E203"), "Phase 2 E-2-03");

  await page.goto("/");
  await page.getByTestId(`select-incident-${incidentId}`).click();
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
  await memberPage.goto("/");
  await memberPage.getByTestId(`select-incident-${incidentId}`).click();

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
  const incidentId = await createIncident(page, uniqueIncidentKey("E204"), "Phase 2 E-2-04");

  await page.goto("/");
  await page.getByTestId("probe-invalid-create").click();
  await expect(page.getByTestId("last-error-code")).toHaveText("invalid_incident_create");
  await expect(page.getByTestId("last-error-details")).toContainText(
    "initial_memberships",
  );

  await page.getByTestId(`select-incident-${incidentId}`).click();
  await page.getByTestId("probe-invalid-patch").click();
  await expect(page.getByTestId("last-error-code")).toHaveText("invalid_incident_patch");
  await expect(page.getByTestId("last-error-details")).toContainText("title");
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

async function ensureAdminSession(page: Page) {
  if (adminCookies !== null) {
    await applyCookies(page, adminCookies.session, adminCookies.csrf);
    return;
  }

  if (adminTotpSecretBase32 === null) {
    adminTotpSecretBase32 = loadCachedAdminTotpSecret();
  }

  const loginResponse = await page.request.post(`${apiBase}/api/v1/auth/login`, {
    data: {
      username: bootstrapEmail,
      password: bootstrapPassword,
    },
  });
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
          data: { client_txn_id: uniqueTxn("phase2-totp-begin") },
        },
      );
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
            client_txn_id: uniqueTxn("phase2-totp-complete"),
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

  const secondFactorLogin = await page.request.post(`${apiBase}/api/v1/auth/login`, {
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
  });
  if (!secondFactorLogin.ok()) {
    clearCachedAdminTotpSecret();
    throw new Error(`second factor login failed: ${await secondFactorLogin.text()}`);
  }

  adminCookies = {
    session: requireCookie(secondFactorLogin, sessionCookieName),
    csrf: requireCookie(secondFactorLogin, csrfCookieName),
  };
  await applyCookies(page, adminCookies.session, adminCookies.csrf);
}

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
  throw new Error(`missing ${name} cookie on login response`);
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

function uniqueIncidentKey(prefix: string) {
  return `IR-${prefix}-${Date.now().toString(36).toUpperCase()}`;
}

function uniqueEmail(prefix: string) {
  return `${prefix}-${Date.now().toString(36)}@example.test`;
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
