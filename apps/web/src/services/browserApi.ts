import { type PublicAPIError, publicErrorView } from "../shared/publicError";

export type APIError = {
  code: string;
  details?: Record<string, unknown>;
  message?: string;
  request_id?: string;
  retryable?: boolean;
  status?: number;
} & PublicAPIError;

export type APIResult<T> = {
  ok: boolean;
  status: number;
  payload: T | { error?: APIError };
};

export { publicErrorView };

export const csrfCookieName = "cartulary_csrf";
export const csrfHeaderName = "X-CSRF-Token";

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

export function apiPath(base: string | undefined, path: string): string {
  const trimmedBase = (base ?? "").trim();
  if (trimmedBase === "") {
    return path;
  }
  return `${trimmedBase.replace(/\/$/, "")}${path}`;
}

export async function fetchJSON<T>(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<APIResult<T>> {
  const method = (init?.method ?? "GET").toUpperCase();
  const credentials = init?.credentials ?? "include";
  const headers = new Headers(init?.headers);

  if (method !== "GET" && method !== "HEAD" && method !== "OPTIONS") {
    headers.set("Content-Type", "application/json");
    if (credentials === "include") {
      const csrfToken = readCookie(csrfCookieName);
      if (csrfToken !== null && csrfToken !== "") {
        headers.set(csrfHeaderName, csrfToken);
      }
    }
  }

  const response = await fetch(input, {
    credentials,
    ...init,
    headers,
  });
  const contentType = response.headers.get("Content-Type") ?? "";
  const payload = contentType.includes("application/json")
    ? ((await response.json()) as T | { error?: APIError })
    : ((await response.text()) as unknown as T | { error?: APIError });
  return { ok: response.ok, status: response.status, payload };
}

export function extractError(payload: unknown): APIError | null {
  if (!payload || typeof payload !== "object") {
    return null;
  }
  return ((payload as { error?: APIError }).error ?? null) as APIError | null;
}

export function clientTxnID(prefix: string): string {
  if (
    typeof crypto !== "undefined" &&
    typeof crypto.randomUUID === "function"
  ) {
    return `${prefix}-${crypto.randomUUID()}`;
  }
  return `${prefix}-${Date.now()}`;
}
