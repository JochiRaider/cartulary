export type APIError = {
  code: string;
  details?: Record<string, unknown>;
  message?: string;
  request_id?: string;
  status?: number;
};

export type APIResult<T> = {
  ok: boolean;
  status: number;
  payload: T | { error?: APIError };
};

export type PublicErrorDetail = {
  key: string;
  label: string;
  value: string;
};

export type PublicErrorView = {
  code: string;
  details: PublicErrorDetail[];
  status: number | null;
  statusText: string;
};

export const csrfCookieName = "cartulary_csrf";
export const csrfHeaderName = "X-CSRF-Token";

const publicDetailLabels: Record<string, string> = {
  reason_code: "Reason",
  field: "Field",
  required_role: "Required role",
  required_second_factor_kinds: "Required second factor kinds",
  required_setup_kinds: "Required setup kinds",
  bootstrap_expires_at: "Bootstrap expires at",
};

const publicDetailKeys = Object.keys(publicDetailLabels);

const unsafePublicTextPatterns = [
  /bootstrap[_ -]?token/i,
  /secret[_ -]?base32/i,
  /otpauth/i,
  /request[_ -]?id/i,
  /\bstack\b/i,
  /\btraceback\b/i,
  /\bat\s+\S+\s*\(/,
  /\/(?:home|var|tmp|usr|app|workspace)\//i,
  /\bselect\b[\s\S]{0,80}\bfrom\b/i,
  /\binsert\b[\s\S]{0,80}\binto\b/i,
  /\bupdate\b[\s\S]{0,80}\bset\b/i,
];

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

export function publicErrorView(
  error: APIError | null | undefined,
  fallbackStatus?: number,
): PublicErrorView | null {
  if (!error) {
    return null;
  }
  const status =
    typeof error.status === "number"
      ? error.status
      : typeof fallbackStatus === "number"
        ? fallbackStatus
        : null;
  return {
    code: error.code,
    status,
    statusText: publicStatusText(error, status),
    details: publicErrorDetails(error),
  };
}

function publicStatusText(error: APIError, status: number | null): string {
  const message = safePublicText(error.message);
  if (message !== null) {
    return message;
  }
  switch (status) {
    case 400:
      return "Invalid request.";
    case 401:
      return "Authentication required.";
    case 403:
      return "Access denied.";
    case 404:
      return "Not found.";
    case 409:
      return "Conflict.";
    case 413:
      return "Request too large.";
    default:
      return "Request failed.";
  }
}

function publicErrorDetails(error: APIError): PublicErrorDetail[] {
  const details = error.details ?? {};
  return publicDetailKeys.flatMap((key) => {
    const value = publicDetailValue(details[key]);
    if (value === null) {
      return [];
    }
    return [
      {
        key,
        label: publicDetailLabels[key] ?? key,
        value,
      },
    ];
  });
}

function publicDetailValue(value: unknown): string | null {
  if (
    typeof value === "string" ||
    typeof value === "number" ||
    typeof value === "boolean"
  ) {
    return safePublicText(String(value));
  }
  if (Array.isArray(value)) {
    const parts = value.flatMap((item) => {
      if (
        typeof item === "string" ||
        typeof item === "number" ||
        typeof item === "boolean"
      ) {
        const text = safePublicText(String(item));
        return text === null ? [] : [text];
      }
      return [];
    });
    return parts.length === 0 ? null : parts.join(", ");
  }
  return null;
}

function safePublicText(value: string | undefined): string | null {
  const text = value?.trim() ?? "";
  if (text === "" || text.length > 240) {
    return null;
  }
  if (unsafePublicTextPatterns.some((pattern) => pattern.test(text))) {
    return null;
  }
  return text;
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
