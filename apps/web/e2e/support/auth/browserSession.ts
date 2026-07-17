import type { APIRequestContext, APIResponse, Page } from "@playwright/test";

import { expect } from "@playwright/test";
import type { StorageState } from "../../playwrightTypes";
import { apiBase } from "../runtime/configuration";
import { waitForPageRequestAPIReady } from "../runtime/readiness";
import {
  csrfCookieName,
  csrfHeaderName,
  sessionCookieName,
} from "./storageState";

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
  if (storageState.cookies.length > 0) {
    await page.context().addCookies(storageState.cookies);
  }
}

export async function csrfHeaders(page: Page) {
  const cookies = await page.context().cookies();
  const csrfCookie = cookies.find((cookie) => cookie.name === csrfCookieName);
  if (!csrfCookie) {
    throw new Error("missing CSRF cookie");
  }
  return { [csrfHeaderName]: csrfCookie.value };
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
              assertion: { code: secondFactorCode },
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
    data: { username: email, password },
  });
  expect(response.ok()).toBeTruthy();
  await applyCookies(
    page,
    requireCookie(response, sessionCookieName),
    requireCookie(response, csrfCookieName),
  );
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
