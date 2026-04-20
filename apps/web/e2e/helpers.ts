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
  type StorageState,
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
  const loginResponse = await page.request.post(
    `${apiBase}/api/v1/auth/login`,
    {
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
    },
  );
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

export async function applyStorageState(
  page: Page,
  storageState: StorageState,
) {
  await page.context().clearCookies();
  if (storageState.cookies.length === 0) {
    return;
  }
  await page.context().addCookies(storageState.cookies);
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

export function cookieValueFromStorageState(
  storageState: StorageState,
  name: string,
) {
  const match = storageState.cookies.find((cookie) => cookie.name === name);
  return match?.value ?? null;
}

export function requireCookieValueFromStorageState(
  storageState: StorageState,
  name: string,
) {
  const value = cookieValueFromStorageState(storageState, name);
  if (!value) {
    throw new Error(`missing ${name} cookie in storage state`);
  }
  return value;
}

export function csrfHeadersForStorageState(storageState: StorageState) {
  return {
    [csrfHeaderName]: requireCookieValueFromStorageState(
      storageState,
      csrfCookieName,
    ),
  };
}

export function cookieHeaderForStorageState(storageState: StorageState) {
  return storageState.cookies
    .map((cookie) => `${cookie.name}=${cookie.value}`)
    .join("; ");
}

export function authHeadersForStorageState(storageState: StorageState) {
  return {
    Cookie: cookieHeaderForStorageState(storageState),
    ...csrfHeadersForStorageState(storageState),
  };
}

export function storageStateFromCookieValues(
  session: string,
  csrf: string,
): StorageState {
  return {
    cookies: [
      {
        name: sessionCookieName,
        value: session,
        domain: "127.0.0.1",
        path: "/",
        expires: -1,
        httpOnly: true,
        sameSite: "Lax",
        secure: false,
      },
      {
        name: csrfCookieName,
        value: csrf,
        domain: "127.0.0.1",
        path: "/",
        expires: -1,
        httpOnly: false,
        sameSite: "Lax",
        secure: false,
      },
    ],
    origins: [],
  };
}

export function requireCookieValueFromSetCookieHeaders(
  headers: string[],
  name: string,
) {
  for (const header of headers) {
    const [cookiePair] = header.split(";", 1);
    if (!cookiePair) {
      continue;
    }
    const [cookieName, cookieValue] = cookiePair.split("=", 2);
    if (cookieName === name && cookieValue) {
      return cookieValue;
    }
  }
  throw new Error(`missing ${name} cookie in Set-Cookie headers`);
}

export function storageStateFromSetCookieHeaders(headers: string[]) {
  return storageStateFromCookieValues(
    requireCookieValueFromSetCookieHeaders(headers, sessionCookieName),
    requireCookieValueFromSetCookieHeaders(headers, csrfCookieName),
  );
}

export async function loginLocalAPIContext(
  authRequests: APIRequestContext,
  options: {
    email: string;
    password: string;
    secondFactorCode?: string | null;
  },
) {
  const secondFactorCode = options.secondFactorCode?.trim() ?? "";
  return authRequests.post("/api/v1/auth/login", {
    data: {
      username: options.email,
      password: options.password,
      ...(secondFactorCode === ""
        ? {}
        : {
            second_factor: {
              kind: "totp",
              assertion: {
                code: secondFactorCode,
              },
            },
          }),
    },
  });
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

export async function fetchTimelineRecordSubstrate(
  page: Page,
  recordId: string,
) {
  const response = await page.request.get(
    `${apiBase}/api/v1/test/timeline/records/${recordId}/substrate`,
  );
  expect(response.ok()).toBeTruthy();
  return (
    (await response.json()) as {
      data: {
        record_id: string;
        row_version: number;
        capture_state: string;
        replacement_record_id: string | null;
        record_revision_count: number;
      };
    }
  ).data;
}

export async function fetchTimelineRecordChangeCount(
  page: Page,
  recordId: string,
) {
  const response = await page.request.get(
    `${apiBase}/api/v1/test/timeline/record-changes`,
  );
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as {
    data: {
      record_changes: Array<{ record_id: string }>;
    };
  };
  return body.data.record_changes.filter(
    (change) => change.record_id === recordId,
  ).length;
}

export async function measureTypingAck(
  page: Page,
  testId: string,
  appendedCharacter: string,
) {
  const input = page.getByTestId(testId);
  const currentValue = await input.inputValue();
  const completion = waitForInputValue(page, {
    testId,
    expectedValue: `${currentValue}${appendedCharacter}`,
    requireFocus: true,
    timeoutMs: 5_000,
  });
  await input.press(appendedCharacter);
  return completion;
}

export async function measureBlankRowCreate(
  page: Page,
  expectedSummary: string,
) {
  const completion = waitForCommittedRowSummary(page, {
    expectedSummary,
    timeoutMs: 5_000,
  });
  await page.getByTestId("draft-row-summary").press("Enter");
  return completion;
}

export function percentile95(samples: number[]) {
  if (samples.length === 0) {
    throw new Error("cannot compute percentile95 for an empty sample set");
  }
  const sorted = [...samples].sort((left, right) => left - right);
  const index = Math.max(0, Math.ceil(sorted.length * 0.95) - 1);
  return sorted[index] ?? sorted[sorted.length - 1];
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

export async function waitForAPIReady(authRequests: APIRequestContext) {
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

export async function waitForPageRequestAPIReady(page: Page) {
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

async function waitForInputValue(
  page: Page,
  options: {
    testId: string;
    expectedValue: string;
    requireFocus: boolean;
    timeoutMs: number;
  },
) {
  const start = await page.evaluate(() => performance.now());
  return page.evaluate(
    ({ expectedValue, requireFocus, startMark, testId, timeoutMs }) =>
      new Promise<number>((resolve, reject) => {
        const deadline = startMark + timeoutMs;
        const selector = `[data-testid="${CSS.escape(testId)}"]`;
        const tick = () => {
          const element = document.querySelector(selector);
          const isTextInput =
            element instanceof HTMLInputElement ||
            element instanceof HTMLTextAreaElement;
          if (
            isTextInput &&
            element.value === expectedValue &&
            (!requireFocus || document.activeElement === element)
          ) {
            resolve(performance.now() - startMark);
            return;
          }
          if (performance.now() > deadline) {
            reject(
              new Error(
                `timed out waiting for ${testId} to reach ${expectedValue}`,
              ),
            );
            return;
          }
          requestAnimationFrame(tick);
        };
        requestAnimationFrame(tick);
      }),
    {
      expectedValue: options.expectedValue,
      requireFocus: options.requireFocus,
      startMark: start,
      testId: options.testId,
      timeoutMs: options.timeoutMs,
    },
  );
}

async function waitForCommittedRowSummary(
  page: Page,
  options: {
    expectedSummary: string;
    timeoutMs: number;
  },
) {
  const start = await page.evaluate(() => performance.now());
  return page.evaluate(
    ({ expectedSummary, startMark, timeoutMs }) =>
      new Promise<number>((resolve, reject) => {
        const deadline = startMark + timeoutMs;
        const tick = () => {
          const candidates = document.querySelectorAll(
            'input[data-testid$="-summary"], textarea[data-testid$="-summary"]',
          );
          for (const candidate of candidates) {
            if (
              !(
                candidate instanceof HTMLInputElement ||
                candidate instanceof HTMLTextAreaElement
              )
            ) {
              continue;
            }
            const testId = candidate.getAttribute("data-testid") ?? "";
            if (
              !testId.startsWith("row-") ||
              !testId.endsWith("-summary") ||
              candidate.value !== expectedSummary
            ) {
              continue;
            }
            const recordId = testId.slice(4, -"-summary".length);
            const versionSelector = `[data-testid="${CSS.escape(`row-${recordId}-row-version`)}"]`;
            const versionElement = document.querySelector(versionSelector);
            if (
              versionElement instanceof HTMLElement &&
              versionElement.textContent?.trim() &&
              versionElement.textContent.trim() !== "new"
            ) {
              resolve(performance.now() - startMark);
              return;
            }
          }
          if (performance.now() > deadline) {
            reject(
              new Error(
                `timed out waiting for committed row summary ${expectedSummary}`,
              ),
            );
            return;
          }
          requestAnimationFrame(tick);
        };
        requestAnimationFrame(tick);
      }),
    {
      expectedSummary: options.expectedSummary,
      startMark: start,
      timeoutMs: options.timeoutMs,
    },
  );
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

export function loadSuiteAdminTotpSecret() {
  if (!existsSync(suiteAdminTotpStatePath)) {
    return null;
  }

  const secret = readFileSync(suiteAdminTotpStatePath, "utf8").trim();
  return secret === "" ? null : secret;
}

export function writeSuiteAdminTotpSecret(secretBase32: string) {
  writeFileSync(suiteAdminTotpStatePath, `${secretBase32}\n`, "utf8");
}

export function clearSuiteAdminTotpSecret() {
  if (!existsSync(suiteAdminTotpStatePath)) {
    return;
  }
  unlinkSync(suiteAdminTotpStatePath);
}
