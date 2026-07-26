import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { csrfHeaderName } from "../../services/browserApi";
import { readHeader } from "../../testing/fetchMockTestSupport";
import {
  adminResetPassword,
  adminResetTotp,
  adminRevokeAllSessions,
  beginTotpEnrollment,
  changePassword,
  completeTotpEnrollment,
  createEnterpriseAuthBinding,
  createLocalUser,
  listUsers,
  loadCredentialState,
  loadSession,
  loadUser,
  loginLocal,
  logoutCurrentSession,
  patchLocalUser,
  retireEnterpriseAuthBinding,
  rotateEnterpriseAuthBinding,
} from "./appShellClient";

describe("App-shell API route boundaries", () => {
  let cookieValue = "";
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    cookieValue = "";
    fetchMock = vi
      .fn()
      .mockImplementation(() =>
        Promise.resolve(jsonResponse({ data: { ok: true } })),
      );
    vi.stubGlobal("fetch", fetchMock);
    vi.spyOn(document, "cookie", "get").mockImplementation(() => cookieValue);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("route-boundary client helpers keep auth and account requests under /api/v1/ with route-specific cookie-backed CSRF and closed bodies", async () => {
    cookieValue = "cartulary_csrf=session-csrf";

    await loadSession();
    await loadCredentialState();
    await loginLocal({
      username: "operator@example.test",
      password: "OperatorPass1!",
    });
    await loginLocal({
      username: "operator@example.test",
      password: "OperatorPass1!",
      secondFactorCode: "123456",
    });
    await logoutCurrentSession();
    await changePassword({
      clientTxnId: "txn-password-change",
      currentPassword: "CurrentPass1!",
      newPassword: "ReplacementPass1!",
      secondFactorCode: "654321",
    });
    await createLocalUser({
      clientTxnId: "txn-user-create",
      displayName: "Authentication User",
      email: "deployment-user@example.test",
      initialPassword: "InitialPass1!",
      isDeploymentAdmin: false,
      mfaRequired: true,
    });
    await listUsers();
    await listUsers({ cursorToken: " cursor-2 ", limit: 50 });
    await loadUser({ userId: "user-2" });
    await patchLocalUser({
      baseUserVersion: 7,
      displayName: "Updated User",
      email: "user-2@example.test",
      isActive: true,
      isDeploymentAdmin: false,
      mfaRequired: true,
      userId: "user-2",
    });
    await createEnterpriseAuthBinding({
      baseUserVersion: 8,
      clientTxnId: "txn-auth-binding-create",
      providerKey: "corp-oidc",
      providerSubject: "subject-1",
      reason: "",
      userId: "user-2",
    });
    await rotateEnterpriseAuthBinding({
      authBindingId: "binding-1",
      baseUserVersion: 9,
      clientTxnId: "txn-auth-binding-rotate",
      newProviderSubject: "subject-2",
      reason: "subject rotation",
      userId: "user-2",
    });
    await retireEnterpriseAuthBinding({
      authBindingId: "binding-1",
      baseUserVersion: 10,
      clientTxnId: "txn-auth-binding-retire",
      reason: "",
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
    const expectedRequests: ExpectedRouteRequest[] = [
      {
        body: null,
        csrfHeader: "",
        method: "GET",
        url: "/api/v1/auth/session",
      },
      {
        body: null,
        csrfHeader: "",
        method: "GET",
        url: "/api/v1/auth/credential-state",
      },
      {
        body: {
          username: "operator@example.test",
          password: "OperatorPass1!",
        },
        csrfHeader: "session-csrf",
        method: "POST",
        url: "/api/v1/auth/login",
      },
      {
        body: {
          username: "operator@example.test",
          password: "OperatorPass1!",
          second_factor: {
            kind: "totp",
            assertion: {
              code: "123456",
            },
          },
        },
        csrfHeader: "session-csrf",
        method: "POST",
        url: "/api/v1/auth/login",
      },
      {
        body: {},
        csrfHeader: "session-csrf",
        method: "POST",
        url: "/api/v1/auth/logout",
      },
      {
        body: {
          client_txn_id: "txn-password-change",
          current_password: "CurrentPass1!",
          new_password: "ReplacementPass1!",
          second_factor: {
            kind: "totp",
            assertion: {
              code: "654321",
            },
          },
        },
        csrfHeader: "session-csrf",
        method: "POST",
        url: "/api/v1/auth/password/change",
      },
      {
        body: {
          client_txn_id: "txn-user-create",
          auth_kind: "local",
          email: "deployment-user@example.test",
          display_name: "Authentication User",
          initial_password: "InitialPass1!",
          mfa_required: true,
          is_deployment_admin: false,
        },
        csrfHeader: "session-csrf",
        method: "POST",
        url: "/api/v1/users",
      },
      {
        body: null,
        csrfHeader: "",
        method: "GET",
        url: "/api/v1/users?limit=100",
      },
      {
        body: null,
        csrfHeader: "",
        method: "GET",
        url: "/api/v1/users?cursor_token=cursor-2&limit=50",
      },
      {
        body: null,
        csrfHeader: "",
        method: "GET",
        url: "/api/v1/users/user-2",
      },
      {
        body: {
          base_user_version: 7,
          display_name: "Updated User",
          email: "user-2@example.test",
          mfa_required: true,
          is_active: true,
          is_deployment_admin: false,
        },
        csrfHeader: "session-csrf",
        method: "PATCH",
        url: "/api/v1/users/user-2",
      },
      {
        body: {
          base_user_version: 8,
          client_txn_id: "txn-auth-binding-create",
          provider_key: "corp-oidc",
          provider_subject: "subject-1",
          reason: "",
        },
        csrfHeader: "session-csrf",
        method: "POST",
        url: "/api/v1/users/user-2/auth-bindings",
      },
      {
        body: {
          base_user_version: 9,
          client_txn_id: "txn-auth-binding-rotate",
          new_provider_subject: "subject-2",
          reason: "subject rotation",
        },
        csrfHeader: "session-csrf",
        method: "POST",
        url: "/api/v1/users/user-2/auth-bindings/binding-1/rotate",
      },
      {
        body: {
          base_user_version: 10,
          client_txn_id: "txn-auth-binding-retire",
          reason: "",
        },
        csrfHeader: "session-csrf",
        method: "DELETE",
        url: "/api/v1/users/user-2/auth-bindings/binding-1",
      },
      {
        body: {
          base_user_version: 7,
          client_txn_id: "txn-admin-password-reset",
          new_password: "AdminResetPass1!",
          reason: "routine reset",
        },
        csrfHeader: "session-csrf",
        method: "POST",
        url: "/api/v1/users/user-2/password/reset",
      },
      {
        body: {
          base_user_version: 7,
          client_txn_id: "txn-admin-totp-reset",
          reason: "routine reset",
        },
        csrfHeader: "session-csrf",
        method: "POST",
        url: "/api/v1/users/user-2/mfa/totp/reset",
      },
      {
        body: {
          client_txn_id: "txn-admin-revoke-all",
          reason: "routine revoke",
        },
        csrfHeader: "session-csrf",
        method: "POST",
        url: "/api/v1/users/user-2/sessions/revoke-all",
      },
    ];

    expect(requests).toHaveLength(expectedRequests.length);
    for (const [index, expected] of expectedRequests.entries()) {
      expectCookieBackedAPIRequest(requests[index], expected);
      expectClosedJSONBody(requests[index], expected.body);
    }
  });

  it("route-boundary bootstrap token authorization is limited to TOTP begin and complete", async () => {
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

type ExpectedRouteRequest = {
  body: Record<string, unknown> | null;
  csrfHeader: string;
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

function expectCookieBackedAPIRequest(
  request: CapturedRequest | undefined,
  expected: ExpectedRouteRequest,
) {
  expect(request).toBeDefined();
  expect(request?.method).toBe(expected.method);
  expect(request?.url).toBe(expected.url);
  expect(request?.url.startsWith("/api/v1/")).toBe(true);
  expect(request?.init?.credentials).toBe("include");
  expect(readHeader(request?.init, "Authorization")).toBe("");
  expect(readHeader(request?.init, csrfHeaderName)).toBe(expected.csrfHeader);
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
