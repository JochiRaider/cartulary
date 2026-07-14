import { type PublicAPIError, publicErrorView } from "../shared/publicError";
import {
  csrfCookieName,
  csrfHeaderName,
  readCookie,
  requestJSON,
} from "./httpTransport";

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

export { csrfCookieName, csrfHeaderName, publicErrorView, readCookie };

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
  return requestJSON<T>(input, init, {
    contentType: "mutations",
    responseParsing: "content-aware",
  }) as Promise<APIResult<T>>;
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
