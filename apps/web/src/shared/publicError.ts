/**
 * The closed set of public-error codes the client originates itself, such as
 * transport loss, timeouts, local unavailability and contract failures. It
 * never overlaps a registered server error code.
 */
export const clientErrorCodes = [
  "access_unavailable",
  "credential_state_unavailable",
  "deployment_user_unavailable",
  "deployment_users_unavailable",
  "incident_summary_unavailable",
  "invalid_deployment_user_response",
  "invalid_public_contract_response",
  "invalid_upstream_contract",
  "observation_interrupted",
  "pending_queue_capacity_exceeded",
  "pending_runtime_metadata_missing",
  "public_error",
  "request_rejected",
  "session_observation_unavailable",
  "timeout",
  "transport",
  "transport_failure",
  "transport_unavailable",
  "transport_uncertain",
  "unknown_public_error",
] as const;

export type ClientErrorCode = (typeof clientErrorCodes)[number];

/**
 * Tests a typed failure code against candidate codes. Candidates must belong
 * to the code's own union, so a code the handled operations never declare is
 * a compile error rather than a dead branch (Core 03 REQ-03-086a).
 */
export function isFailureCode<Code extends string, const Listed extends Code>(
  code: Code | null | undefined,
  candidates: readonly Listed[],
): code is Listed {
  return (
    code !== null &&
    code !== undefined &&
    (candidates as readonly string[]).includes(code)
  );
}

export type PublicAPIError = {
  code?: string | undefined;
  details?: Record<string, unknown> | undefined;
  message?: string | undefined;
  request_id?: string | undefined;
  retryable?: boolean | undefined;
  status?: number | undefined;
};

/**
 * The admission gate's refusal (Core 01 REQ-01-702, Core 04 AC-618). Every
 * operation declares `503` `service_unavailable`, and each registered reason
 * means nothing was applied, so the same request may be retried. The client
 * presents it as temporary unavailability with a retry path, never as a
 * contract failure (Core 03 REQ-03-086a). The reason selects no behaviour or
 * copy and is never shown.
 */
export function isTemporaryUnavailability(
  error: Pick<PublicAPIError, "code" | "status"> | null | undefined,
): boolean {
  return (
    error?.code === "service_unavailable" &&
    (error.status === undefined || error.status === 503)
  );
}

/** The one authored copy for temporary unavailability; `retryLabel` names a visible retry control. */
export function temporaryUnavailabilityText(retryLabel?: string): string {
  return retryLabel === undefined
    ? "Cartulary is temporarily unavailable. Try again shortly."
    : `Cartulary is temporarily unavailable. Choose ${retryLabel} to try again.`;
}

/**
 * A write refused before it was applied: a decoded `503`
 * `service_unavailable`. Every registered reason either refused admission or
 * proved the commit absent (Core 01 §3.3.6.1–3.3.6.2), so a first attempt is a
 * definite rejection that keeps the draft; a resubmission is a new logical
 * action. A refused recovery replay proves only that the replay was not
 * applied, so the earlier outcome stays uncertain (Core 03 REQ-03-086a,
 * REQ-03-301).
 */
export function isUnappliedRefusal(
  status: number | null | undefined,
  error: Pick<PublicAPIError, "code" | "status"> | null | undefined,
): boolean {
  return status === 503 && isTemporaryUnavailability(error);
}

/** Copy for a first attempt refused before it was applied. */
export function notAppliedText(retryLabel?: string): string {
  return retryLabel === undefined
    ? "Cartulary is temporarily unavailable. This action was not applied; try again shortly."
    : `Cartulary is temporarily unavailable. This action was not applied; choose ${retryLabel} to try again.`;
}

/** Copy for a recovery replay refused before it was applied. */
export function stillUnconfirmedText(): string {
  return "Cartulary is temporarily unavailable. The earlier outcome is still unconfirmed; try recovery again shortly.";
}

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
  if (isTemporaryUnavailability(error)) {
    return temporaryUnavailabilityText();
  }
  const message = safePublicText(error?.message);
  if (message !== null) {
    return message;
  }
  const status =
    typeof error?.status === "number" ? error.status : (fallbackStatus ?? null);
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
  return code === ""
    ? ("unknown_public_error" satisfies ClientErrorCode)
    : code;
}

function publicErrorDetails(error: PublicAPIError): PublicErrorDetail[] {
  const details = error.details ?? {};
  const unavailable = isTemporaryUnavailability(error);
  return publicDetailKeys.flatMap((key) => {
    if (unavailable && key === "reason_code") {
      return [];
    }
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
