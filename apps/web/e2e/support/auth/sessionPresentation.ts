import { errorRegistry } from "@cartulary/protocol-ts/errors";
import {
  type GetCurrentSessionResponse,
  validateHTTPOperationResponse,
} from "@cartulary/protocol-ts/http";
import type { APIResponse, Route } from "@playwright/test";

type SessionResponse = Pick<APIResponse, "status" | "ok" | "json">;
type SessionDiagnostic = Readonly<{
  operation: "getCurrentSession";
  status: number | null;
  reason:
    | "upstream_failure"
    | "invalid_json"
    | "invalid_contract"
    | "transport_failure";
  code?: string;
  request_id?: string;
}>;

function object(value: unknown): Record<string, unknown> | null {
  return typeof value === "object" && value !== null && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : null;
}

/** Only bounded public identities cross the diagnostic boundary, never response text. */
function diagnostic(
  status: number | null,
  reason: SessionDiagnostic["reason"],
  body?: unknown,
): SessionDiagnostic {
  const envelope = object(body);
  const error = object(envelope?.error);
  const meta = object(envelope?.meta);
  const code = errorRegistry.errors.find(
    (entry) => entry.code === error?.code && entry.http_status === status,
  )?.code;
  const requestId = meta?.request_id ?? error?.request_id;
  return {
    operation: "getCurrentSession",
    status,
    reason,
    ...(code === undefined ? {} : { code }),
    ...(typeof requestId === "string" &&
    /^[A-Za-z0-9._-]{1,128}$/.test(requestId)
      ? { request_id: requestId }
      : {}),
  };
}

export async function readSessionPresentationResponse(
  response: SessionResponse,
) {
  let body: unknown;
  try {
    body = await response.json();
  } catch {
    return {
      kind: "rejected",
      diagnostic: diagnostic(response.status(), "invalid_json"),
    } as const;
  }
  if (!response.ok())
    return {
      kind: "rejected",
      diagnostic: diagnostic(response.status(), "upstream_failure", body),
    } as const;
  if (!validateHTTPOperationResponse("getCurrentSession", body).ok)
    return {
      kind: "rejected",
      diagnostic: diagnostic(response.status(), "invalid_contract"),
    } as const;
  return {
    kind: "accepted",
    value: body as GetCurrentSessionResponse,
  } as const;
}

/** Capture mutable presentation choices before calling; upstream failures remain upstream failures. */
export async function rewriteSessionPresentation(
  route: Pick<Route, "fetch" | "fulfill" | "abort">,
  rewrite: (
    session: GetCurrentSessionResponse["data"],
  ) => GetCurrentSessionResponse["data"],
  report: (failure: SessionDiagnostic) => void = (failure) =>
    console.warn("Session presentation fixture:", JSON.stringify(failure)),
): Promise<void> {
  let response: APIResponse;
  try {
    response = await route.fetch();
  } catch {
    report(diagnostic(null, "transport_failure"));
    await route.abort("failed");
    return;
  }
  const result = await readSessionPresentationResponse(response);
  if (result.kind === "rejected") {
    report(result.diagnostic);
    await route.fulfill({ response });
    return;
  }
  await route.fulfill({
    response,
    json: { ...result.value, data: rewrite(result.value.data) },
  });
}
