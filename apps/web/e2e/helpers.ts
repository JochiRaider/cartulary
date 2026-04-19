import { createHmac } from "node:crypto";
import { existsSync, readFileSync, unlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import {
  type APIRequestContext,
  type APIResponse,
  expect,
  type Page,
  request,
} from "@playwright/test";

export const bootstrapEmail = "dev-admin@example.test";
export const bootstrapPassword = "DevBootstrap1!";
export const sessionCookieName = "cartulary_session";
export const csrfCookieName = "cartulary_csrf";
export const csrfHeaderName = "X-CSRF-Token";
export const apiBase = "http://127.0.0.1:8080";

const suiteAdminTotpStatePath = join(
  tmpdir(),
  "cartulary-playwright-admin-totp.txt",
);

let rememberedAdminTotpSecretBase32: string | null = null;

export async function ensureAdminSession(page: Page) {
  if (rememberedAdminTotpSecretBase32 === null) {
    rememberedAdminTotpSecretBase32 = loadSuiteAdminTotpSecret();
  }
  if (rememberedAdminTotpSecretBase32 === null) {
    throw new Error("missing suite admin TOTP state; global setup did not run");
  }

  await waitForPageRequestAPIReady(page);
  const loginResponse = await page.request.post(`${apiBase}/api/v1/auth/login`, {
    data: {
      username: bootstrapEmail,
      password: bootstrapPassword,
      second_factor: {
        kind: "totp",
        assertion: {
          code: generateTotpCode(rememberedAdminTotpSecretBase32),
        },
      },
    },
  });
  if (!loginResponse.ok()) {
    throw new Error(`admin login failed: ${await loginResponse.text()}`);
  }

  await applyCookies(
    page,
    requireCookie(loginResponse, sessionCookieName),
    requireCookie(loginResponse, csrfCookieName),
  );
}

export async function prepareSuiteAdminState() {
  rememberedAdminTotpSecretBase32 = null;
  clearSuiteAdminTotpSecret();

  const authRequests = await request.newContext({ baseURL: apiBase });
  try {
    await waitForAPIReady(authRequests);
    const secretBase32 = await provisionBootstrapAdminTotp(authRequests);
    writeSuiteAdminTotpSecret(secretBase32);
  } finally {
    await authRequests.dispose();
  }
}

export async function enrollTotpViaBootstrap(email: string, password: string) {
  const authRequests = await request.newContext({ baseURL: apiBase });
  try {
    await waitForAPIReady(authRequests);
    const secretBase32 = await provisionUserTotp(authRequests, email, password);
    return secretBase32;
  } finally {
    await authRequests.dispose();
  }
}

export function resetRememberedAdminSession() {
  // Each test performs its own login; no worker-shared session should persist.
}

export async function applyCookies(page: Page, session: string, csrf: string) {
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

export async function csrfHeaders(page: Page) {
  const cookies = await page.context().cookies(apiBase);
  const csrfCookie = cookies.find((cookie) => cookie.name === csrfCookieName);
  if (!csrfCookie) {
    throw new Error("missing CSRF cookie");
  }
  return {
    [csrfHeaderName]: csrfCookie.value,
  };
}

export async function loginLocalSession(
  page: Page,
  email: string,
  password: string,
) {
  await page.context().clearCookies();
  await waitForPageRequestAPIReady(page);
  const response = await page.request.post(`${apiBase}/api/v1/auth/login`, {
    data: {
      username: email,
      password,
    },
  });
  expect(response.ok()).toBeTruthy();
  await applyCookies(
    page,
    requireCookie(response, sessionCookieName),
    requireCookie(response, csrfCookieName),
  );
}

export async function createIncident(
  page: Page,
  incidentKey: string,
  title: string,
) {
  const response = await page.request.post(`${apiBase}/api/v1/incidents`, {
    headers: await csrfHeaders(page),
    data: {
      client_txn_id: uniqueTxn("incident"),
      incident_key: incidentKey,
      title,
    },
  });
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as { data: { incident_id: string } };
  return body.data.incident_id;
}

export async function createLocalUser(
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
      client_txn_id: uniqueTxn("user"),
      auth_kind: "local",
      email: options.email,
      display_name: options.display_name,
      initial_password: options.initial_password,
      mfa_required: options.mfa_required ?? false,
      is_deployment_admin: options.is_deployment_admin ?? false,
    },
  });
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: { user_id: string } }).data;
}

export async function createIncidentMembership(
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
        client_txn_id: uniqueTxn("membership"),
        email,
        role,
      },
    },
  );
  expect(response.ok()).toBeTruthy();
}

export async function createViewRow(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  payload: Record<string, unknown>,
) {
  const response = await page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/views/${viewSchemaId}/rows`,
    {
      headers: await csrfHeaders(page),
      data: payload,
    },
  );
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: { row: Record<string, unknown> } })
    .data.row;
}

export async function queryViewRows(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
) {
  const response = await page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/views/${viewSchemaId}/query`,
    {
      data: {},
    },
  );
  expect(response.ok()).toBeTruthy();
  return (
    (await response.json()) as {
      data: { rows: Array<Record<string, unknown>> };
    }
  ).data.rows;
}

export async function patchTimelineRecord(
  page: Page,
  recordId: string,
  payload: Record<string, unknown>,
) {
  const response = await page.request.patch(
    `${apiBase}/api/v1/records/${recordId}`,
    {
      headers: await csrfHeaders(page),
      data: payload,
    },
  );
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: { row: Record<string, unknown> } })
    .data.row;
}

export function requireCookie(response: APIResponse, name: string) {
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

export function uniqueTxn(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`;
}

export function uniqueEmail(prefix: string) {
  return `${prefix}-${Date.now().toString(36)}@example.test`;
}

export function uniqueIncidentKey(prefix: string) {
  return `IR-${prefix}-${Date.now().toString(36).toUpperCase()}`;
}

export function generateTotpCode(secretBase32: string) {
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

async function waitForAPIReady(authRequests: APIRequestContext) {
  for (let attempt = 0; attempt < 60; attempt += 1) {
    try {
      const response = await authRequests.get("/readyz");
      if (response.ok()) {
        return;
      }
    } catch (error) {
      if (!isConnectionRefused(error)) {
        throw error;
      }
    }
    await sleep(500);
  }
  throw new Error(`timed out waiting for API readiness at ${apiBase}/readyz`);
}

async function waitForPageRequestAPIReady(page: Page) {
  for (let attempt = 0; attempt < 60; attempt += 1) {
    try {
      const response = await page.request.get(`${apiBase}/readyz`);
      if (response.ok()) {
        return;
      }
    } catch (error) {
      if (!isConnectionRefused(error)) {
        throw error;
      }
    }
    await sleep(500);
  }
  throw new Error(`timed out waiting for API readiness at ${apiBase}/readyz`);
}

function isConnectionRefused(error: unknown) {
  return error instanceof Error && error.message.includes("ECONNREFUSED");
}

function sleep(milliseconds: number) {
  return new Promise((resolve) => {
    setTimeout(resolve, milliseconds);
  });
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

async function provisionBootstrapAdminTotp(authRequests: APIRequestContext) {
  const secretBase32 = await provisionUserTotp(
    authRequests,
    bootstrapEmail,
    bootstrapPassword,
  );
  rememberedAdminTotpSecretBase32 = secretBase32;
  return secretBase32;
}

async function provisionUserTotp(
  authRequests: APIRequestContext,
  email: string,
  password: string,
) {
  const loginResponse = await authRequests.post("/api/v1/auth/login", {
    data: {
      username: email,
      password,
    },
  });
  expect(loginResponse.status()).toBe(401);
  const loginBody = (await loginResponse.json()) as {
    error: { code: string; details: { bootstrap_token?: string } };
  };
  expect(loginBody.error.code).toBe("mfa_setup_required");

  const bootstrapToken = loginBody.error.details.bootstrap_token;
  if (!bootstrapToken) {
    throw new Error("missing bootstrap_token");
  }

  const beginResponse = await authRequests.post("/api/v1/auth/mfa/totp/begin", {
    headers: { Authorization: `Bearer ${bootstrapToken}` },
    data: { client_txn_id: uniqueTxn("bootstrap-begin") },
  });
  expect(beginResponse.ok()).toBeTruthy();
  const beginBody = (await beginResponse.json()) as {
    data: { enrollment_id: string; totp_setup: { secret_base32: string } };
  };
  const secretBase32 = beginBody.data.totp_setup.secret_base32;

  const completeResponse = await authRequests.post(
    "/api/v1/auth/mfa/totp/complete",
    {
      headers: { Authorization: `Bearer ${bootstrapToken}` },
      data: {
        client_txn_id: uniqueTxn("bootstrap-complete"),
        enrollment_id: beginBody.data.enrollment_id,
        code: generateTotpCode(secretBase32),
      },
    },
  );
  expect(completeResponse.ok()).toBeTruthy();
  return secretBase32;
}

function loadSuiteAdminTotpSecret() {
  if (!existsSync(suiteAdminTotpStatePath)) {
    return null;
  }

  const secret = readFileSync(suiteAdminTotpStatePath, "utf8").trim();
  return secret === "" ? null : secret;
}

function writeSuiteAdminTotpSecret(secretBase32: string) {
  writeFileSync(suiteAdminTotpStatePath, `${secretBase32}\n`, "utf8");
}

function clearSuiteAdminTotpSecret() {
  if (!existsSync(suiteAdminTotpStatePath)) {
    return;
  }
  unlinkSync(suiteAdminTotpStatePath);
}
