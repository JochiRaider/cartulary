import { csrfCookieName, csrfHeaderName, readCookie } from "./browserApi";

export type LatestQueryRuntime = {
  controller: AbortController | null;
  sequence: number;
};

export type LatestQueryRequest = {
  isCurrent: () => boolean;
  signal: AbortSignal;
};

type WorkbookTimingEvent = {
  at: number;
  name: string;
  [key: string]: unknown;
};

declare global {
  interface Window {
    __cartularyWorkbookTimingProbe?:
      | {
          events: WorkbookTimingEvent[];
          mark?: (event: WorkbookTimingEvent) => void;
        }
      | undefined;
  }
}

export function readEnvelope<T>(payload: unknown): T {
  return payload as T;
}

export function beginLatestQuery(runtime: {
  current: LatestQueryRuntime;
}): LatestQueryRequest {
  const previousController = runtime.current.controller;
  const controller = new AbortController();
  const sequence = runtime.current.sequence + 1;
  runtime.current = { controller, sequence };
  previousController?.abort();

  return {
    signal: controller.signal,
    isCurrent: () =>
      runtime.current.sequence === sequence &&
      runtime.current.controller === controller &&
      !controller.signal.aborted,
  };
}

export function abortLatestQuery(runtime: { current: LatestQueryRuntime }) {
  runtime.current.controller?.abort();
  runtime.current = {
    controller: null,
    sequence: runtime.current.sequence + 1,
  };
}

export function isAbortError(error: unknown): boolean {
  return error instanceof DOMException
    ? error.name === "AbortError"
    : error instanceof Error && error.name === "AbortError";
}

export async function fetchJSON<T>(
  input: RequestInfo | URL,
  init?: RequestInit,
  options: {
    onJSONParsed?: () => void;
    onResponse?: (response: Response) => void;
  } = {},
): Promise<{
  ok: boolean;
  status: number;
  payload:
    | T
    | { error?: { code?: string; message?: string; details?: unknown } };
}> {
  const method = (init?.method ?? "GET").toUpperCase();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init?.headers as Record<string, string> | undefined),
  };
  if (method !== "GET" && method !== "HEAD" && method !== "OPTIONS") {
    const csrfToken = readCookie(csrfCookieName);
    if (csrfToken !== null && csrfToken !== "") {
      headers[csrfHeaderName] = csrfToken;
    }
  }
  const requestURL = input instanceof Request ? input.url : String(input);
  if (
    window.__cartularyWorkbookTimingProbe !== undefined &&
    method === "POST" &&
    requestURL.includes("/views/cartulary.view.timeline.v2/rows")
  ) {
    headers["X-Cartulary-Timing-Debug"] = "1";
  }

  const response = await fetch(input, {
    credentials: "include",
    ...init,
    headers,
  });
  options.onResponse?.(response);
  const payload = (await response.json()) as
    | T
    | {
        error?: {
          code?: string;
          message?: string;
          retryable?: unknown;
          conflict?: unknown;
          details?: unknown;
        };
      };
  options.onJSONParsed?.();
  return { ok: response.ok, status: response.status, payload };
}

export function parseErrorMessage(payload: unknown) {
  if (!payload || typeof payload !== "object" || !("error" in payload)) {
    return "Request failed.";
  }
  const error = payload.error;
  if (!error || typeof error !== "object") {
    return "Request failed.";
  }
  if ("code" in error && typeof error.code === "string") {
    if (
      "details" in error &&
      error.details &&
      typeof error.details === "object" &&
      "reason_code" in error.details &&
      typeof error.details.reason_code === "string"
    ) {
      return `${error.code}: ${error.details.reason_code}`;
    }
    return error.code;
  }
  if ("message" in error && typeof error.message === "string") {
    return error.message;
  }
  return "Request failed.";
}

export function handleWorkbookLoadFailure(
  error: unknown,
  fallbackMessage: string,
  onIncidentAccessLost: (() => void) | undefined,
): string {
  const message =
    typeof error === "string"
      ? error
      : error instanceof Error
        ? error.message
        : fallbackMessage;
  if (
    message.includes("incident_not_found") ||
    message.includes("authorization_denied")
  ) {
    onIncidentAccessLost?.();
  }
  return message;
}
