import { createHmac } from "node:crypto";
import { existsSync, readFileSync, unlinkSync, writeFileSync } from "node:fs";
import {
  gridDraftRowSelector,
  gridSavedRowsSelector,
  gridShellTestId,
  type WorkbookSurface,
} from "@cartulary/ui-contracts";
import {
  type APIRequestContext,
  type APIResponse,
  type Browser,
  expect,
  type Page,
  type Route,
  request,
} from "@playwright/test";

import {
  isExternalServerHarnessMode,
  resolvePlaywrightStateFile,
  sharedPlaywrightStateDir,
} from "./harnessState";
import type { StorageState } from "./playwrightTypes";

export const bootstrapEmail = "dev-admin@example.test";
export const bootstrapPassword = "DevBootstrap1!";
export const sessionCookieName = "cartulary_session";
export const csrfCookieName = "cartulary_csrf";
export const csrfHeaderName = "X-CSRF-Token";

function originFromEnv(name: string, fallback: string) {
  return (process.env[name] ?? fallback).replace(/\/+$/, "");
}

export const apiBase = originFromEnv(
  "CARTULARY_WEB_E2E_API_ORIGIN",
  "http://127.0.0.1:8080",
);
export const webBase = originFromEnv(
  "CARTULARY_WEB_E2E_PUBLIC_ORIGIN",
  "http://127.0.0.1:4173",
);

type HeldBrowserAPIRequest = {
  waitForHit: Promise<void>;
  release: () => void;
  dispose: () => Promise<void>;
  hitCount: () => number;
};

export const ordinaryMeasurementSamplePolicy = {
  warmupSamples: 1,
  measuredSamples: 25,
  totalSamples: 26,
} as const;

export type ServerTimingMetric = {
  attributes: Record<string, string | true>;
  durationMs: number | null;
  name: string;
  raw: string;
};

const suiteAdminTotpStatePath = resolvePlaywrightStateFile(
  "cartulary-playwright-admin-totp.txt",
);

export type LocalLoginResult =
  | {
      kind: "success";
    }
  | {
      kind: "error";
      status: number;
      code: string;
      details: Record<string, unknown>;
    };

export type SuiteAdminAuthClient = {
  loginLocal: (secondFactorCode?: string | null) => Promise<LocalLoginResult>;
  provisionTotpFromBootstrap: (bootstrapToken: string) => Promise<string>;
};

export type SuiteAdminStateContext = {
  externalServerMode: boolean;
  sharedStateDir: string | null;
  stateFilePath: string;
};

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

  const authRequests = await request.newContext({ baseURL: apiBase });
  try {
    await waitForAPIReady(authRequests);
    const secretBase32 = await reconcileSuiteAdminTotpState(
      suiteAdminAuthClient(authRequests),
      loadSuiteAdminTotpSecret(),
    );
    rememberedAdminTotpSecretBase32 = secretBase32;
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

export function browserApiRoute(path: string): string {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `**${normalizedPath}`;
}

export async function holdBrowserApiRequest(
  page: Page,
  options: {
    method: string;
    path: string;
  },
): Promise<HeldBrowserAPIRequest> {
  const routePattern = browserApiRoute(options.path);
  const expectedMethod = options.method.toUpperCase();
  let matchingHitCount = 0;
  let hitResolved = false;
  let releaseHold: (() => void) | null = null;
  let resolveHit: (() => void) | null = null;
  const waitForHit = new Promise<void>((resolve) => {
    resolveHit = resolve;
  });
  const hold = new Promise<void>((resolve) => {
    releaseHold = resolve;
  });

  const routeHandler = async (route: Route) => {
    if (route.request().method().toUpperCase() !== expectedMethod) {
      await route.fallback();
      return;
    }

    matchingHitCount += 1;
    if (!hitResolved) {
      hitResolved = true;
      resolveHit?.();
    }
    await hold;
    await route.continue();
  };

  await page.route(routePattern, routeHandler);

  return {
    waitForHit,
    release: () => {
      releaseHold?.();
    },
    dispose: async () => {
      releaseHold?.();
      if (page.isClosed()) {
        return;
      }
      try {
        await page.unroute(routePattern, routeHandler);
      } catch (error: unknown) {
        if (!isClosedPageRouteCleanupError(error)) {
          throw error;
        }
      }
    },
    hitCount: () => matchingHitCount,
  };
}

function isClosedPageRouteCleanupError(error: unknown) {
  const message = error instanceof Error ? error.message : String(error);
  return (
    message.includes("Target page, context or browser has been closed") ||
    message.includes("Page closed") ||
    message.includes("BrowserContext closed")
  );
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
  const cookies = await page.context().cookies();
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

export async function createIncidentMemberUser(
  page: Page,
  incidentId: string,
  options: {
    email: string;
    display_name: string;
    initial_password: string;
    role: string;
    mfa_required?: boolean;
    is_deployment_admin?: boolean;
  },
) {
  const userOptions: {
    email: string;
    display_name: string;
    initial_password: string;
    mfa_required?: boolean;
    is_deployment_admin?: boolean;
  } = {
    email: options.email,
    display_name: options.display_name,
    initial_password: options.initial_password,
  };
  if (options.mfa_required !== undefined) {
    userOptions.mfa_required = options.mfa_required;
  }
  if (options.is_deployment_admin !== undefined) {
    userOptions.is_deployment_admin = options.is_deployment_admin;
  }
  const user = await createLocalUser(page, userOptions);
  await createIncidentMembership(page, incidentId, options.email, options.role);
  return {
    ...user,
    email: options.email,
    initial_password: options.initial_password,
    role: options.role,
  };
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
  const createRoute =
    "**/api/v1/incidents/*/views/cartulary.view.timeline.v1/rows";
  const routeHandler = async (route: Route) => {
    await route.fallback({
      headers: {
        ...route.request().headers(),
        "X-Cartulary-Timing-Debug": "1",
      },
    });
  };
  await page.route(createRoute, routeHandler);
  const completion = waitForCommittedRowSummary(page, {
    expectedSummary,
    surface: "timeline",
    timeoutMs: 5_000,
  });
  const responseCompletion = page.waitForResponse(
    (response) => {
      const request = response.request();
      if (
        request.method() !== "POST" ||
        !response.url().includes("/views/cartulary.view.timeline.v1/rows")
      ) {
        return false;
      }
      const postData = request.postData();
      return postData?.includes(expectedSummary) ?? false;
    },
    { timeout: 5_000 },
  );
  const networkStart = Date.now();
  await page.getByTestId("draft-row-summary").press("Enter");
  try {
    const [committed, response] = await Promise.all([
      completion,
      responseCompletion,
    ]);
    const status = response.status();
    expect(status).toBe(201);
    const serverTiming = response.headers()["server-timing"] ?? "";
    return {
      committedDurationMs: committed.durationMs,
      networkDurationMs: Date.now() - networkStart,
      recordId: committed.recordId,
      rowVersion: committed.rowVersion,
      serverTiming,
      serverTimingMetrics: parseServerTiming(serverTiming),
      status,
    };
  } finally {
    await page.unroute(createRoute, routeHandler);
  }
}

export function percentile95(
  samples: number[],
  options: {
    minimumSampleCount?: number;
    sampleLabel?: string;
  } = {},
) {
  const minimumSampleCount =
    options.minimumSampleCount ??
    ordinaryMeasurementSamplePolicy.measuredSamples;
  const sampleLabel = options.sampleLabel ?? "samples";
  if (samples.length === 0) {
    throw new Error("cannot compute percentile95 for an empty sample set");
  }
  if (samples.length < minimumSampleCount) {
    throw new Error(
      `cannot compute percentile95 for ${sampleLabel}: expected at least ${minimumSampleCount} samples, got ${samples.length}`,
    );
  }
  const sorted = [...samples].sort((left, right) => left - right);
  const index = Math.max(0, Math.ceil(sorted.length * 0.95) - 1);
  return sorted[index] ?? sorted[sorted.length - 1];
}

export function parseServerTiming(header: string): ServerTimingMetric[] {
  return splitServerTimingHeader(header)
    .map((rawPart) => rawPart.trim())
    .filter((rawPart) => rawPart !== "")
    .map((raw) => {
      const [rawName = "", ...rawAttributes] = raw.split(";");
      const attributes: Record<string, string | true> = {};
      let durationMs: number | null = null;
      for (const rawAttribute of rawAttributes) {
        const attribute = rawAttribute.trim();
        if (attribute === "") {
          continue;
        }
        const separatorIndex = attribute.indexOf("=");
        if (separatorIndex < 0) {
          attributes[attribute] = true;
          continue;
        }
        const key = attribute.slice(0, separatorIndex).trim();
        const value = unquoteServerTimingValue(
          attribute.slice(separatorIndex + 1).trim(),
        );
        attributes[key] = value;
        if (key.toLowerCase() === "dur") {
          const parsed = Number.parseFloat(value);
          durationMs = Number.isFinite(parsed) ? parsed : null;
        }
      }
      return {
        attributes,
        durationMs,
        name: rawName.trim(),
        raw,
      };
    });
}

function splitServerTimingHeader(header: string) {
  const parts: string[] = [];
  let current = "";
  let quoted = false;
  for (const character of header) {
    if (character === '"') {
      quoted = !quoted;
      current += character;
      continue;
    }
    if (character === "," && !quoted) {
      parts.push(current);
      current = "";
      continue;
    }
    current += character;
  }
  parts.push(current);
  return parts;
}

function unquoteServerTimingValue(value: string) {
  if (value.length >= 2 && value.startsWith('"') && value.endsWith('"')) {
    return value.slice(1, -1).replace(/\\"/gu, '"');
  }
  return value;
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

export function gridSavedRows(page: Page, surface: WorkbookSurface) {
  return page
    .getByTestId(gridShellTestId(surface))
    .locator(gridSavedRowsSelector());
}

export function gridDraftRows(page: Page, surface: WorkbookSurface) {
  return page
    .getByTestId(gridShellTestId(surface))
    .locator(gridDraftRowSelector());
}

export function generateTotpCode(secretBase32: string) {
  const secret = decodeBase32(secretBase32);
  const counter = Math.floor(Date.now() / 1000 / 30);
  const counterBuffer = Buffer.alloc(8);
  counterBuffer.writeBigUInt64BE(BigInt(counter));
  const digest = createHmac("sha1", secret).update(counterBuffer).digest();
  const offsetSource = digest.at(-1);
  if (offsetSource === undefined) {
    throw new Error("empty TOTP digest");
  }
  const offset = offsetSource & 0x0f;
  const codeBytes = digest.subarray(offset, offset + 4);
  if (codeBytes.length !== 4) {
    throw new Error("short TOTP digest window");
  }
  const [byte0, byte1, byte2, byte3] = codeBytes;
  if (
    byte0 === undefined ||
    byte1 === undefined ||
    byte2 === undefined ||
    byte3 === undefined
  ) {
    throw new Error("short TOTP digest window");
  }
  const code =
    ((byte0 & 0x7f) << 24) |
    ((byte1 & 0xff) << 16) |
    ((byte2 & 0xff) << 8) |
    (byte3 & 0xff);
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

export function currentSuiteAdminStateContext(): SuiteAdminStateContext {
  return {
    externalServerMode: isExternalServerHarnessMode(),
    sharedStateDir: sharedPlaywrightStateDir(),
    stateFilePath: suiteAdminTotpStatePath,
  };
}

export async function reconcileSuiteAdminTotpState(
  client: SuiteAdminAuthClient,
  storedSecretBase32: string | null,
  context: SuiteAdminStateContext = currentSuiteAdminStateContext(),
) {
  const normalizedStoredSecret = storedSecretBase32?.trim() ?? "";
  if (normalizedStoredSecret !== "") {
    const loginWithStoredSecret = await client.loginLocal(
      generateTotpCode(normalizedStoredSecret),
    );
    if (loginWithStoredSecret.kind === "success") {
      return normalizedStoredSecret;
    }
    if (loginWithStoredSecret.code === "mfa_setup_required") {
      return provisionSuiteAdminTotp(
        client,
        loginWithStoredSecret.details,
        context,
      );
    }
    throw suiteAdminStateError(
      `stored suite admin TOTP secret no longer matches the current backend state (login code ${loginWithStoredSecret.code})`,
      context,
    );
  }

  const loginWithoutSecondFactor = await client.loginLocal();
  if (loginWithoutSecondFactor.kind === "success") {
    throw new Error(
      "suite admin login unexpectedly succeeded without MFA during harness setup",
    );
  }
  if (loginWithoutSecondFactor.code === "mfa_setup_required") {
    return provisionSuiteAdminTotp(
      client,
      loginWithoutSecondFactor.details,
      context,
    );
  }
  if (loginWithoutSecondFactor.code === "mfa_required") {
    throw suiteAdminStateError(
      "suite admin MFA is already active but no stored harness TOTP secret is available",
      context,
    );
  }
  throw new Error(
    `suite admin harness login failed with ${loginWithoutSecondFactor.code}`,
  );
}

export async function openIncidentFromLanding(page: Page, incidentId: string) {
  await page.goto("/");
  await expect(
    page.getByTestId(`landing-incident-${incidentId}`),
  ).toBeVisible();
  await page.getByTestId(`landing-open-${incidentId}`).click();
  await expect(page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
}

export async function openIncidentAsTrackedUser(
  browser: Browser,
  sessionTracker: {
    loginTrackedUser: (
      page: Page,
      details: {
        createdBy: string;
        email: string;
        password: string;
        purpose: string;
        userId: string;
      },
    ) => Promise<void>;
  },
  options: {
    createdBy: string;
    email: string;
    incidentId: string;
    password: string;
    purpose: string;
    userId: string;
  },
) {
  const context = await browser.newContext();
  const page = await context.newPage();
  await sessionTracker.loginTrackedUser(page, {
    createdBy: options.createdBy,
    email: options.email,
    password: options.password,
    purpose: options.purpose,
    userId: options.userId,
  });
  await openIncidentFromLanding(page, options.incidentId);
  return page;
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

export async function waitForCommittedRowSummary(
  page: Page,
  options: {
    expectedSummary: string;
    surface: WorkbookSurface;
    timeoutMs: number;
  },
) {
  const start = await page.evaluate(() => performance.now());
  return page.evaluate(
    ({
      draftRowSelector,
      expectedSummary,
      gridShell,
      savedRowsSelector,
      startMark,
      timeoutMs,
    }) =>
      new Promise<{ durationMs: number; recordId: string; rowVersion: number }>(
        (resolve, reject) => {
          const deadline = startMark + timeoutMs;
          const tick = () => {
            const gridRoot = document.querySelector(
              `[data-testid="${CSS.escape(gridShell)}"]`,
            );
            if (!(gridRoot instanceof HTMLElement)) {
              if (performance.now() > deadline) {
                reject(
                  new Error(`timed out waiting for workbook grid ${gridShell}`),
                );
                return;
              }
              requestAnimationFrame(tick);
              return;
            }
            const candidates = gridRoot.querySelectorAll(
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
              if (candidate.value !== expectedSummary) {
                continue;
              }
              if (candidate.closest(draftRowSelector) !== null) {
                continue;
              }
              const savedRow = candidate.closest(savedRowsSelector);
              if (!(savedRow instanceof HTMLElement)) {
                continue;
              }
              const recordId = savedRow
                .getAttribute("data-grid-record-id")
                ?.trim();
              if (!recordId) {
                continue;
              }
              const rowVersionTestId = `row-${recordId}-row-version`;
              const rowVersionElement = savedRow.querySelector(
                `[data-testid="${CSS.escape(rowVersionTestId)}"]`,
              );
              const rowVersion =
                rowVersionElement instanceof HTMLElement
                  ? Number.parseInt(rowVersionElement.textContent ?? "", 10)
                  : Number.NaN;
              if (!Number.isInteger(rowVersion) || rowVersion < 1) {
                continue;
              }
              resolve({
                durationMs: performance.now() - startMark,
                recordId,
                rowVersion,
              });
              return;
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
        },
      ),
    {
      draftRowSelector: gridDraftRowSelector(),
      expectedSummary: options.expectedSummary,
      gridShell: gridShellTestId(options.surface),
      savedRowsSelector: gridSavedRowsSelector(),
      startMark: start,
      timeoutMs: options.timeoutMs,
    },
  );
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
  const loginResult = await readLocalLoginResult(loginResponse);
  if (
    loginResult.kind !== "error" ||
    loginResult.status !== 401 ||
    loginResult.code !== "mfa_setup_required"
  ) {
    throw new Error(
      `expected mfa_setup_required while provisioning TOTP for ${email}, got ${formatLocalLoginResult(loginResult)}`,
    );
  }
  return provisionTotpFromBootstrap(
    authRequests,
    requireBootstrapToken(loginResult.details, currentSuiteAdminStateContext()),
  );
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

function suiteAdminAuthClient(
  authRequests: APIRequestContext,
): SuiteAdminAuthClient {
  return {
    loginLocal: async (secondFactorCode) => {
      const response = await loginLocalAPIContext(authRequests, {
        email: bootstrapEmail,
        password: bootstrapPassword,
        ...(secondFactorCode === undefined ? {} : { secondFactorCode }),
      });
      return readLocalLoginResult(response);
    },
    provisionTotpFromBootstrap: async (bootstrapToken) =>
      provisionTotpFromBootstrap(authRequests, bootstrapToken),
  };
}

async function readLocalLoginResult(
  response: APIResponse,
): Promise<LocalLoginResult> {
  if (response.ok()) {
    return { kind: "success" };
  }
  const body = await readErrorEnvelope(response);
  return {
    kind: "error",
    status: response.status(),
    code: body.error?.code ?? "unknown_error",
    details: toErrorDetails(body.error?.details),
  };
}

async function readErrorEnvelope(response: APIResponse) {
  return (await response.json()) as {
    error?: { code?: string; details?: unknown };
  };
}

function formatLocalLoginResult(result: LocalLoginResult) {
  if (result.kind === "success") {
    return "success";
  }
  return `${result.status} ${result.code}`;
}

function toErrorDetails(value: unknown) {
  if (value && typeof value === "object" && !Array.isArray(value)) {
    return value as Record<string, unknown>;
  }
  return {};
}

async function provisionSuiteAdminTotp(
  client: SuiteAdminAuthClient,
  details: Record<string, unknown>,
  context: SuiteAdminStateContext,
) {
  return client.provisionTotpFromBootstrap(
    requireBootstrapToken(details, context),
  );
}

async function provisionTotpFromBootstrap(
  authRequests: APIRequestContext,
  bootstrapToken: string,
) {
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

function requireBootstrapToken(
  details: Record<string, unknown>,
  context: SuiteAdminStateContext,
) {
  const bootstrapToken = details.bootstrap_token;
  if (typeof bootstrapToken === "string" && bootstrapToken.trim() !== "") {
    return bootstrapToken;
  }
  throw suiteAdminStateError(
    "suite admin login did not return a bootstrap_token for TOTP enrollment",
    context,
  );
}

function suiteAdminStateError(
  message: string,
  context: SuiteAdminStateContext,
) {
  const stateLocation =
    context.sharedStateDir === null
      ? context.stateFilePath
      : `${context.stateFilePath} (CARTULARY_PLAYWRIGHT_STATE_DIR=${context.sharedStateDir})`;
  const ownership = context.externalServerMode
    ? "The reused external-server stack owns this state for its full lifetime."
    : "The harness expected to provision this state during Playwright global setup.";
  return new Error(
    `${message}. Expected harness state at ${stateLocation}. ${ownership}`,
  );
}
