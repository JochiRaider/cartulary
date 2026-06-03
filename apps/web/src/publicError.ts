export type PublicAPIError = {
  code?: string | undefined;
  details?: Record<string, unknown> | undefined;
  message?: string | undefined;
  request_id?: string | undefined;
  retryable?: boolean | undefined;
  status?: number | undefined;
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

export function publicErrorView(
  error: PublicAPIError | null | undefined,
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
    code: publicErrorCode(error),
    status,
    statusText: publicErrorStatusText(error, status),
    details: publicErrorDetails(error),
  };
}

export function publicErrorStatusText(
  error: PublicAPIError | null | undefined,
  fallbackStatus?: number | null,
): string {
  const message = safePublicText(error?.message);
  if (message !== null) {
    return message;
  }
  const status =
    typeof error?.status === "number" ? error.status : fallbackStatus ?? null;
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

export function publicErrorCode(
  error: PublicAPIError | null | undefined,
): string {
  const code = error?.code?.trim() ?? "";
  return code === "" ? "unknown_public_error" : code;
}

function publicErrorDetails(error: PublicAPIError): PublicErrorDetail[] {
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
