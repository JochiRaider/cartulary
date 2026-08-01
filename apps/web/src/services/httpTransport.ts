import type { Decoder } from "@cartulary/protocol-ts";

const csrfCookieName = "cartulary_csrf";
export const csrfHeaderName = "X-CSRF-Token";

export type HTTPTransportError = {
  readonly code?: string;
  readonly details?: unknown;
  readonly message?: string;
  readonly request_id?: string;
  readonly retryable?: boolean;
};

export type HTTPTransportResult<T> = {
  readonly ok: boolean;
  readonly status: number;
  readonly payload: T | { error?: HTTPTransportError };
};

export function readCookie(name: string): string | null {
  if (typeof document === "undefined") {
    return null;
  }
  const prefix = `${name}=`;
  for (const segment of document.cookie.split(";")) {
    const trimmed = segment.trim();
    if (trimmed.startsWith(prefix)) {
      return decodeURIComponent(trimmed.slice(prefix.length));
    }
  }
  return null;
}

function requestHeaders(
  init: RequestInit | undefined,
  contentType: "always" | "mutations",
): Headers {
  const method = (init?.method ?? "GET").toUpperCase();
  const credentials = init?.credentials ?? "include";
  const headers = new Headers(init?.headers);
  const stateChanging =
    method !== "GET" && method !== "HEAD" && method !== "OPTIONS";
  if (contentType === "always" || stateChanging) {
    headers.set("Content-Type", "application/json");
  }
  if (stateChanging && credentials === "include") {
    const csrfToken = readCookie(csrfCookieName);
    if (csrfToken) {
      headers.set(csrfHeaderName, csrfToken);
    }
  }
  return headers;
}

export async function requestJSON<T>(
  input: RequestInfo | URL,
  init?: RequestInit,
  options: {
    readonly contentType?: "always" | "mutations";
    readonly decoder?: Decoder<T>;
    readonly onJSONParsed?: () => void;
    readonly onResponse?: (response: Response) => void;
    readonly responseParsing?: "content-aware" | "json";
  } = {},
): Promise<HTTPTransportResult<T>> {
  const response = await fetch(input, {
    credentials: "include",
    ...init,
    headers: requestHeaders(init, options.contentType ?? "mutations"),
  });
  options.onResponse?.(response);
  const contentType = response.headers.get("Content-Type") ?? "";
  const raw =
    options.responseParsing === "json" ||
    contentType.includes("application/json")
      ? await response.json()
      : await response.text();
  options.onJSONParsed?.();
  if (response.ok && options.decoder) {
    const decoded = options.decoder.decode(raw);
    if (!decoded.ok) {
      return {
        ok: false,
        status: 502,
        payload: {
          error: {
            code: "invalid_upstream_contract",
            details: decoded.error,
            retryable: false,
          },
        },
      };
    }
    return { ok: true, status: response.status, payload: decoded.value };
  }
  return {
    ok: response.ok,
    status: response.status,
    payload: raw as T | { error?: HTTPTransportError },
  };
}

export async function requestMultipartJSON<T>(
  input: RequestInfo | URL,
  body: FormData,
  init: Omit<RequestInit, "body" | "credentials" | "headers"> = {},
): Promise<HTTPTransportResult<T>> {
  const headers = new Headers();
  const csrfToken = readCookie(csrfCookieName);
  if (csrfToken) {
    headers.set(csrfHeaderName, csrfToken);
  }
  const response = await fetch(input, {
    ...init,
    method: init.method ?? "POST",
    credentials: "include",
    headers,
    body,
  });
  const payload = (await response.json()) as
    | T
    | {
        error?: HTTPTransportError;
      };
  return { ok: response.ok, status: response.status, payload };
}
