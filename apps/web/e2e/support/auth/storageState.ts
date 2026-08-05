import type { BrowserContext } from "@playwright/test";

export type StorageState = Awaited<ReturnType<BrowserContext["storageState"]>>;

export const sessionCookieName = "cartulary_session";
export const csrfCookieName = "cartulary_csrf";
export const csrfHeaderName = "X-CSRF-Token";

export function cookieValueFromStorageState(
  storageState: StorageState,
  name: string,
) {
  const match = storageState.cookies.find((cookie) => cookie.name === name);
  return match?.value ?? null;
}

function requireCookieValueFromStorageState(
  storageState: StorageState,
  name: string,
) {
  const value = cookieValueFromStorageState(storageState, name);
  if (!value) {
    throw new Error(`missing ${name} cookie in storage state`);
  }
  return value;
}

function csrfHeadersForStorageState(storageState: StorageState) {
  return {
    [csrfHeaderName]: requireCookieValueFromStorageState(
      storageState,
      csrfCookieName,
    ),
  };
}

function cookieHeaderForStorageState(storageState: StorageState) {
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
