import type { StorageState } from "../../playwrightTypes";

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
