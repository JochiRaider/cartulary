import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { csrfHeaderName } from "./browserApi";
import { readHeader } from "./fetchMockTestSupport";
import {
  adminResetPassword,
  adminResetTotp,
  adminRevokeAllSessions,
  beginTotpEnrollment,
  changePassword,
  completeTotpEnrollment,
  createLocalUser,
  loadCredentialState,
  loadSession,
  loadUser,
  loginLocal,
  logoutCurrentSession,
  patchLocalUser,
} from "./phase1Client";

describe("Phase 1 API route boundaries", () => {
  let cookieValue = "";
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    cookieValue = "";
    fetchMock = vi.fn().mockImplementation(() =>
      Promise.resolve(jsonResponse({ data: { ok: true } })),
    );
    vi.stubGlobal("fetch", fetchMock);
    vi.spyOn(document, "cookie", "get").mockImplementation(() => cookieValue);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("FE-I-P1-01 route-boundary client helpers keep auth and account requests under /api/v1/ with cookie-backed session defaults", async () => {
    await loadSession();
    await loadCredentialState();
    await loginLocal({
      username: "operator@example.test",
      password: "OperatorPass1!",
    });
    await logoutCurrentSession();
    await changePassword({
      clientTxnId: "txn-password-change",
      currentPassword: "CurrentPass1!",
      newPassword: "ReplacementPass1!",
      secondFactorCode: "123456",
    });
    await createLocalUser({
      clientTxnId: "txn-user-create",
      displayName: "Phase 1 User",
      email: "phase1-user@example.test",
      initialPassword: "InitialPass1!",
      isDeploymentAdmin: false,
      mfaRequired: true,
    });
    await loadUser({ userId: "user-2" });
    await patchLocalUser({
      baseUserVersion: 7,
      displayName: "Updated User",
      isActive: true,
      isDeploymentAdmin: false,
      mfaRequired: true,
      userId: "user-2",
    });
    await adminResetPassword({
      baseUserVersion: 7,
      clientTxnId: "txn-admin-password-reset",
      newPassword: "AdminResetPass1!",
      reason: "routine reset",
      userId: "user-2",
    });
    await adminResetTotp({
      baseUserVersion: 7,
      clientTxnId: "txn-admin-totp-reset",
      reason: "routine reset",
      userId: "user-2",
    });
    await adminRevokeAllSessions({
      clientTxnId: "txn-admin-revoke-all",
      reason: "routine revoke",
      userId: "user-2",
    });

    const requests = capturedRequests(fetchMock);
    expect(
      requests.map((request) => `${request.method} ${request.url}`),
    ).toEqual([
      "GET /api/v1/auth/session",
      "GET /api/v1/auth/credential-state",
      "POST /api/v1/auth/login",
      "POST /api/v1/auth/logout",
      "POST /api/v1/auth/password/change",
      "POST /api/v1/users",
      "GET /api/v1/users/user-2",
      "PATCH /api/v1/users/user-2",
      "POST /api/v1/users/user-2/password/reset",
      "POST /api/v1/users/user-2/mfa/totp/reset",
      "POST /api/v1/users/user-2/sessions/revoke-all",
    ]);

    for (const request of requests) {
      expectCookieBackedAPIRequest(request);
    }

    expectClosedJSONBody(requests[0], null);
    expectClosedJSONBody(requests[1], null);
    expectClosedJSONBody(requests[2], {
      username: "operator@example.test",
      password: "OperatorPass1!",
    });
    expectClosedJSONBody(requests[3], {});
    expectClosedJSONBody(requests[4], {
      client_txn_id: "txn-password-change",
      current_password: "CurrentPass1!",
      new_password: "ReplacementPass1!",
      second_factor: {
        kind: "totp",
        assertion: {
          code: "123456",
        },
      },
    });
    expectClosedJSONBody(requests[5], {
      client_txn_id: "txn-user-create",
      auth_kind: "local",
      email: "phase1-user@example.test",
      display_name: "Phase 1 User",
      initial_password: "InitialPass1!",
      mfa_required: true,
      is_deployment_admin: false,
    });
    expectClosedJSONBody(requests[6], null);
    expectClosedJSONBody(requests[7], {
      base_user_version: 7,
      display_name: "Updated User",
      mfa_required: true,
      is_active: true,
      is_deployment_admin: false,
    });
    expectClosedJSONBody(requests[8], {
      base_user_version: 7,
      client_txn_id: "txn-admin-password-reset",
      new_password: "AdminResetPass1!",
      reason: "routine reset",
    });
    expectClosedJSONBody(requests[9], {
      base_user_version: 7,
      client_txn_id: "txn-admin-totp-reset",
      reason: "routine reset",
    });
    expectClosedJSONBody(requests[10], {
      client_txn_id: "txn-admin-revoke-all",
      reason: "routine revoke",
    });
  });

  it("FE-I-P1-01 route-boundary bootstrap token authorization is limited to TOTP begin and complete", async () => {
    cookieValue = "cartulary_csrf=session-csrf";

    await beginTotpEnrollment({
      authMode: "bootstrap",
      bootstrapToken: " bootstrap-token-1 ",
      clientTxnId: "txn-bootstrap-begin",
    });
    await completeTotpEnrollment({
      authMode: "bootstrap",
      bootstrapToken: "bootstrap-token-1",
      clientTxnId: "txn-bootstrap-complete",
      code: "123456",
      enrollmentId: "enrollment-1",
    });
    await beginTotpEnrollment({
      authMode: "session",
      clientTxnId: "txn-session-begin",
      currentFactorCode: "654321",
      currentPassword: "CurrentPass1!",
    });
    await completeTotpEnrollment({
      authMode: "session",
      clientTxnId: "txn-session-complete",
      code: "654321",
      enrollmentId: "enrollment-2",
    });

    const requests = capturedRequests(fetchMock);
    expect(
      requests.map((request) => `${request.method} ${request.url}`),
    ).toEqual([
      "POST /api/v1/auth/mfa/totp/begin",
      "POST /api/v1/auth/mfa/totp/complete",
      "POST /api/v1/auth/mfa/totp/begin",
      "POST /api/v1/auth/mfa/totp/complete",
    ]);

    expect(requests[0]?.init?.credentials).toBe("omit");
    expect(requests[1]?.init?.credentials).toBe("omit");
    expect(readHeader(requests[0]?.init, "Authorization")).toBe(
      "Bearer bootstrap-token-1",
    );
    expect(readHeader(requests[1]?.init, "Authorization")).toBe(
      "Bearer bootstrap-token-1",
    );
    expect(readHeader(requests[0]?.init, csrfHeaderName)).toBe("");
    expect(readHeader(requests[1]?.init, csrfHeaderName)).toBe("");

    expect(requests[2]?.init?.credentials).toBe("include");
    expect(requests[3]?.init?.credentials).toBe("include");
    expect(readHeader(requests[2]?.init, "Authorization")).toBe("");
    expect(readHeader(requests[3]?.init, "Authorization")).toBe("");
    expect(readHeader(requests[2]?.init, csrfHeaderName)).toBe("session-csrf");
    expect(readHeader(requests[3]?.init, csrfHeaderName)).toBe("session-csrf");

    expectClosedJSONBody(requests[0], {
      client_txn_id: "txn-bootstrap-begin",
    });
    expectClosedJSONBody(requests[1], {
      client_txn_id: "txn-bootstrap-complete",
      enrollment_id: "enrollment-1",
      code: "123456",
    });
    expectClosedJSONBody(requests[2], {
      client_txn_id: "txn-session-begin",
      current_password: "CurrentPass1!",
      second_factor: {
        kind: "totp",
        assertion: {
          code: "654321",
        },
      },
    });
    expectClosedJSONBody(requests[3], {
      client_txn_id: "txn-session-complete",
      enrollment_id: "enrollment-2",
      code: "654321",
    });
  });
});

type CapturedRequest = {
  body: Record<string, unknown> | null;
  init: RequestInit | undefined;
  method: string;
  url: string;
};

function capturedRequests(fetchMock: ReturnType<typeof vi.fn>) {
  return fetchMock.mock.calls.map((call): CapturedRequest => {
    const input = call[0] as RequestInfo | URL;
    const init = call[1] as RequestInit | undefined;
    const method = (init?.method ?? "GET").toUpperCase();
    return {
      body:
        typeof init?.body === "undefined"
          ? null
          : (JSON.parse(String(init.body)) as Record<string, unknown>),
      init,
      method,
      url: String(input),
    };
  });
}

function expectCookieBackedAPIRequest(request: CapturedRequest | undefined) {
  expect(request).toBeDefined();
  expect(request?.url.startsWith("/api/v1/")).toBe(true);
  expect(request?.init?.credentials).toBe("include");
  expect(readHeader(request?.init, "Authorization")).toBe("");
}

function expectClosedJSONBody(
  request: CapturedRequest | undefined,
  expected: Record<string, unknown> | null,
) {
  expect(request).toBeDefined();
  expect(request?.body).toEqual(expected);
  const actualKeys =
    request?.body === null ? null : Object.keys(request?.body ?? {}).sort();
  const expectedKeys = expected === null ? null : Object.keys(expected).sort();
  expect(actualKeys).toEqual(expectedKeys);
}

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
