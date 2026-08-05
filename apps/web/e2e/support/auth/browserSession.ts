import type { LoginLocalUserRequest } from "@cartulary/protocol-ts/http";
import type { Page } from "@playwright/test";

import { apiBase } from "../runtime/configuration";
import { waitForPageRequestAPIReady } from "../runtime/readiness";
import { publicHttpOperationObserved } from "../transport/publicHttpOperationClient";
import {
  atJsonOrigin,
  type JsonRequestContextLike,
  type RepeatedHeaderJsonResponse,
} from "../transport/publicJsonClient";
import type { StorageState } from "./storageState";
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
  authRequests: JsonRequestContextLike,
  options: {
    email: string;
    password: string;
    secondFactorCode?: string | null;
  },
) {
  return publicHttpOperationObserved({
    body: localLoginRequest(options),
    operationID: "loginLocalUser",
    request: authRequests,
  });
}

export async function loginLocalSession(
  page: Page,
  email: string,
  password: string,
) {
  await page.context().clearCookies();
  await waitForPageRequestAPIReady(page);
  const result = await publicHttpOperationObserved({
    body: localLoginRequest({ email, password }),
    operationID: "loginLocalUser",
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!result.ok) {
    throw new Error(`local login failed with HTTP ${result.status}`);
  }
  await applyCookies(
    page,
    requireCookie(result.response, sessionCookieName),
    requireCookie(result.response, csrfCookieName),
  );
}

export function requireCookie(
  response: RepeatedHeaderJsonResponse,
  name: string,
) {
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

function localLoginRequest(options: {
  email: string;
  password: string;
  secondFactorCode?: string | null;
}): LoginLocalUserRequest {
  const secondFactorCode = options.secondFactorCode?.trim() ?? "";
  return {
    username: options.email,
    password: options.password,
    ...(secondFactorCode === ""
      ? {}
      : {
          second_factor: {
            kind: "totp" as const,
            assertion: { code: secondFactorCode },
          },
        }),
  };
}
