import { createHmac } from "node:crypto";
import { existsSync, readFileSync, unlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { expect, test, type APIResponse, type Page } from "@playwright/test";

const bootstrapEmail = "dev-admin@example.test";
const bootstrapPassword = "DevBootstrap1!";
const timelineViewSchemaId = "cartulary.view.timeline.v1";
const sessionCookieName = "cartulary_session";
const csrfCookieName = "cartulary_csrf";
const csrfHeaderName = "X-CSRF-Token";
const apiBase = "http://127.0.0.1:8080";
const adminTotpCachePath = join(tmpdir(), "cartulary-phase3-admin-totp.txt");

let adminTotpSecretBase32: string | null = null;
let adminCookies: { session: string; csrf: string } | null = null;

test("E-3-01 creates a Timeline row in-grid and continues editing on the draft row", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const incidentId = await createIncident(page, uniqueIncidentKey("E301"), "Phase 3 E-3-01");

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByText("Timeline mutation substrate")).toBeVisible();

  const draftSummary = page.getByTestId("draft-row-summary");
  await draftSummary.fill("First browser fact");
  await draftSummary.press("Enter");

  await expect(page.getByTestId("save-state")).toHaveText("Saved");
  await expect(page.locator("tbody tr")).toHaveCount(2);
  await expect(page.locator('[data-testid^="row-"][data-testid$="-summary"]').first()).toHaveValue("First browser fact");
  await expect(page.getByTestId("draft-row-summary")).toBeFocused();
});

test("E-3-02 shows Syncing, Saved, and Conflict across Enter, Tab, blur, and paste completion", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const incidentId = await createIncident(page, uniqueIncidentKey("E302"), "Phase 3 E-3-02");
  const row = await createTimelineRow(page, incidentId, {
    client_txn_id: uniqueTxn("row"),
    "timeline.summary": "Alpha",
  });
  const recordId = row.record_id as string;

  await page.goto(`/?incident_id=${incidentId}`);
  const summaryInput = page.getByTestId(`row-${recordId}-summary`);
  const detailsInput = page.getByTestId(`row-${recordId}-details`);
  const sourceTextInput = page.getByTestId(`row-${recordId}-sourceText`);

  let delayed = false;
  await page.route(`**/api/v1/records/${recordId}`, async (route) => {
    if (!delayed) {
      delayed = true;
      await page.waitForTimeout(350);
    }
    await route.continue();
  });

  await summaryInput.fill("Alpha enter");
  await summaryInput.press("Enter");
  await expect(page.getByTestId("save-state")).toHaveText("Syncing");
  await expect(page.getByTestId(`row-${recordId}-row-version`)).toHaveText("2");
  await expect(page.getByTestId("save-state")).toHaveText("Saved");

  await summaryInput.fill("Alpha tab");
  await summaryInput.press("Tab");
  await expect(page.getByTestId(`row-${recordId}-row-version`)).toHaveText("3");
  await expect(page.getByTestId("save-state")).toHaveText("Saved");

  await detailsInput.fill("Blur details");
  await page.getByText("Phase 3 Workbook").click();
  await expect(page.getByTestId(`row-${recordId}-row-version`)).toHaveText("4");
  await expect(page.getByTestId("save-state")).toHaveText("Saved");

  await sourceTextInput.fill("Pasted transcript");
  await sourceTextInput.dispatchEvent("paste");
  await expect(page.getByTestId(`row-${recordId}-row-version`)).toHaveText("5");
  await expect(page.getByTestId("save-state")).toHaveText("Saved");

  let conflictInjected = false;
  await page.route(`**/api/v1/records/${recordId}`, async (route) => {
    if (!conflictInjected) {
      conflictInjected = true;
      const body = JSON.parse(route.request().postData() ?? "{}") as {
        base_row_version: number;
      };
      await page.request.patch(`${apiBase}/api/v1/records/${recordId}`, {
        headers: await csrfHeaders(page),
        data: {
          view_schema_id: timelineViewSchemaId,
          base_row_version: body.base_row_version,
          client_txn_id: uniqueTxn("conflict"),
          changes: [{ field_key: "timeline.summary", value: "Server moved first" }],
        },
      });
    }
    await route.continue();
  });

  await summaryInput.fill("Conflict value");
  await page.getByText("Phase 3 Workbook").click();
  await expect(page.getByTestId("save-state")).toHaveText("Conflict");
});

test("E-3-03 drives review, demotion, and supersede through the visible workbook surface", async ({
  page,
}) => {
  await ensureAdminSession(page);
  const incidentId = await createIncident(page, uniqueIncidentKey("E303"), "Phase 3 E-3-03");
  const primaryRow = await createTimelineRow(page, incidentId, {
    client_txn_id: uniqueTxn("primary"),
    "timeline.summary": "Primary row",
  });
  const replacementRow = await createTimelineRow(page, incidentId, {
    client_txn_id: uniqueTxn("replacement"),
    "timeline.summary": "Replacement row",
  });
  const recordId = primaryRow.record_id as string;
  const replacementId = replacementRow.record_id as string;

  await page.goto(`/?incident_id=${incidentId}`);

  await page.getByTestId(`row-${recordId}-mark-reviewed`).click();
  await expect(page.getByTestId(`row-${recordId}-capture-state`)).toHaveText("reviewed");
  await expect(page.getByTestId(`row-${recordId}-row-version`)).toHaveText("2");

  const detailsInput = page.getByTestId(`row-${recordId}-details`);
  await detailsInput.fill("Material edit after review");
  await page.getByText("Phase 3 Workbook").click();
  await expect(page.getByTestId(`row-${recordId}-capture-state`)).toHaveText("enriched");
  await expect(page.getByTestId(`row-${recordId}-row-version`)).toHaveText("3");

  await page.getByTestId(`row-${recordId}-replacement-id`).fill(replacementId);
  await page.getByTestId(`row-${recordId}-supersede`).click();
  await expect(page.getByTestId(`row-${recordId}-capture-state`)).toHaveText("superseded");
  await expect(page.getByTestId(`row-${recordId}-row-version`)).toHaveText("4");
  await expect(page.getByTestId(`row-${recordId}-mark-reviewed`)).toBeDisabled();
});

test("E-3-04 replays the same patch without duplicate visible collaboration updates", async ({
  browser,
  page,
}) => {
  await ensureAdminSession(page);
  const incidentId = await createIncident(page, uniqueIncidentKey("E304"), "Phase 3 E-3-04");
  const row = await createTimelineRow(page, incidentId, {
    client_txn_id: uniqueTxn("seed"),
    "timeline.summary": "Replay row",
  });
  const recordId = row.record_id as string;

  const observer = await browser.newPage({ storageState: await page.context().storageState() });
  let observerQueryCount = 0;
  observer.on("requestfinished", (request) => {
    if (
      request.method() === "POST" &&
      request.url().endsWith(
        `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
      )
    ) {
      observerQueryCount += 1;
    }
  });

  await observer.goto(`http://127.0.0.1:4173/?incident_id=${incidentId}`);
  await expect(observer.getByTestId(`row-${recordId}-row-version`)).toHaveText("1");
  const baselineObserverQueries = observerQueryCount;

  await page.goto(`/?incident_id=${incidentId}`);
  const summaryInput = page.getByTestId(`row-${recordId}-summary`);
  const firstPatchResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/v1/records/${recordId}`),
  );

  await summaryInput.fill("Replay row patched");
  await summaryInput.press("Enter");
  const patchResponse = await firstPatchResponse;
  const firstPatchBody = JSON.parse(
    patchResponse.request().postData() ?? "{}",
  ) as Record<string, unknown>;
  const firstPatchData = ((await patchResponse.json()) as { data: { change_set_id: string } }).data;

  await expect(page.getByTestId(`row-${recordId}-row-version`)).toHaveText("2");
  await expect
    .poll(() => observerQueryCount, { timeout: 5_000 })
    .toBeGreaterThan(baselineObserverQueries);
  await expect(observer.getByTestId(`row-${recordId}-row-version`)).toHaveText("2");

  const queriesAfterFirstPatch = observerQueryCount;
  const replayResponse = await page.request.patch(`${apiBase}/api/v1/records/${recordId}`, {
    headers: await csrfHeaders(page),
    data: firstPatchBody,
  });
  expect(replayResponse.status()).toBe(200);
  const replayData = ((await replayResponse.json()) as { data: { change_set_id: string } }).data;
  expect(replayData.change_set_id).toBe(firstPatchData.change_set_id);

  await page.waitForTimeout(500);
  expect(observerQueryCount).toBe(queriesAfterFirstPatch);
  await expect(observer.getByTestId(`row-${recordId}-row-version`)).toHaveText("2");
  await observer.close();
});

async function ensureAdminSession(page: Page) {
  if (adminCookies !== null) {
    await applyAdminCookies(page);
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
    await applyAdminCookies(page);
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
      const beginResponse = await page.request.post(`${apiBase}/api/v1/auth/mfa/totp/begin`, {
        headers: { Authorization: `Bearer ${bootstrapToken}` },
        data: { client_txn_id: uniqueTxn("totp-begin") },
      });
      const beginBody = (await beginResponse.json()) as {
        data: { enrollment_id: string; totp_setup: { secret_base32: string } };
      };
      adminTotpSecretBase32 = beginBody.data.totp_setup.secret_base32;
      cacheAdminTotpSecret(adminTotpSecretBase32);

      const completeResponse = await page.request.post(`${apiBase}/api/v1/auth/mfa/totp/complete`, {
        headers: { Authorization: `Bearer ${bootstrapToken}` },
        data: {
          client_txn_id: uniqueTxn("totp-complete"),
          enrollment_id: beginBody.data.enrollment_id,
          code: generateTotpCode(adminTotpSecretBase32),
        },
      });
      expect(completeResponse.ok()).toBeTruthy();
    } else if (loginBody.error.code !== "mfa_required") {
      throw new Error(`unexpected login error: ${JSON.stringify(loginBody)}`);
    }
  } else {
    throw new Error(`unexpected login status: ${loginResponse.status()}`);
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
          code: generateTotpCode(adminTotpSecretBase32 ?? ""),
        },
      },
    },
  });
  if (!secondFactorLogin.ok()) {
    clearCachedAdminTotpSecret();
    const body = await secondFactorLogin.json();
    throw new Error(`second factor login failed: ${JSON.stringify(body)}`);
  }
  adminCookies = {
    session: requireCookie(secondFactorLogin, sessionCookieName),
    csrf: requireCookie(secondFactorLogin, csrfCookieName),
  };
  await applyAdminCookies(page);
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

async function applyAdminCookies(page: Page) {
  if (adminCookies === null) {
    throw new Error("missing cached admin cookies");
  }

  await page.context().addCookies([
    {
      name: sessionCookieName,
      value: adminCookies.session,
      domain: "127.0.0.1",
      path: "/",
      httpOnly: true,
      sameSite: "Lax",
    },
    {
      name: csrfCookieName,
      value: adminCookies.csrf,
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

async function createIncident(page: Page, incidentKey: string, title: string) {
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

async function createTimelineRow(page: Page, incidentId: string, payload: Record<string, unknown>) {
  const response = await page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
    {
      headers: await csrfHeaders(page),
      data: payload,
    },
  );
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as {
    data: {
      row: {
        record_id: string;
      };
    };
  };
  return body.data.row as Record<string, unknown>;
}

function uniqueTxn(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`;
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
