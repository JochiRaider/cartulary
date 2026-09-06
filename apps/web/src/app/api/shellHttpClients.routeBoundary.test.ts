import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { HTTPOperationResult } from "../../services/browserApi";
import { csrfHeaderName } from "../../services/browserApi";
import {
  credentialStateResource,
  incidentResource,
  sessionResource,
} from "../../testing/appShellTestSupport";
import { readHeader } from "../../testing/fetchMockTestSupport";
import {
  beginEnterpriseAuth,
  beginTotpEnrollment,
  changePassword,
  completeTotpEnrollment,
  listEnterpriseAuthProviders,
  loadAccountPreferences,
  loadAccountProfile,
  loadCredentialState,
  loadExtensions,
  loadSession,
  loginLocal,
  logoutCurrentSession,
  patchAccountProfile,
  putAccountPreferences,
} from "./authAccountClient";
import {
  adminResetPassword,
  adminResetTotp,
  adminRevokeAllSessions,
  createEnterpriseAuthBinding,
  createLocalUser,
  listUsers,
  loadUser,
  patchLocalUser,
  retireEnterpriseAuthBinding,
  rotateEnterpriseAuthBinding,
} from "./deploymentUserClient";
import { createIncident, listVisibleIncidents } from "./incidentClient";
import type { UserResource } from "./publicHttpTypes";

describe("App-shell API route boundaries", () => {
  let cookieValue = "";
  let malformed = false;
  let operations: (() => Promise<HTTPOperationResult<unknown>>)[] = [];
  let expectedPayload: unknown;
  async function expectOperation(
    operation: () => Promise<HTTPOperationResult<unknown>>,
  ) {
    operations.push(operation);
    const result = await operation();
    expect(result.ok).toBe(true);
    expect(result.payload).toEqual(expectedPayload);
  }
  async function rejectMalformedSuccesses() {
    malformed = true;
    for (const operation of operations) {
      const result = await operation();
      expect(result.ok).toBe(false);
      expect(result.payload).toMatchObject({
        error: { code: "invalid_public_contract_response" },
      });
    }
  }
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    cookieValue = "";
    malformed = false;
    operations = [];
    fetchMock = vi.fn().mockImplementation((input, init) => {
      const fixture = shellFixture(String(input), init?.method ?? "GET");
      expectedPayload = fixture.payload;
      return Promise.resolve(
        jsonResponse(
          malformed
            ? { data: { ok: true }, meta: { request_id: "req-invalid" } }
            : fixture.payload,
          fixture.status,
        ),
      );
    });
    vi.stubGlobal("fetch", fetchMock);
    vi.spyOn(document, "cookie", "get").mockImplementation(() => cookieValue);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("route-boundary client helpers keep auth and account requests under /api/v1/ with route-specific cookie-backed CSRF and closed bodies", async () => {
    cookieValue = "cartulary_csrf=session-csrf";

    await expectOperation(() => loadSession());
    await expectOperation(() => loadCredentialState());
    await expectOperation(() =>
      loginLocal({
        username: "operator@example.test",
        password: "OperatorPass1!",
      }),
    );
    await expectOperation(() =>
      loginLocal({
        username: "operator@example.test",
        password: "OperatorPass1!",
        secondFactorCode: "123456",
      }),
    );
    await expectOperation(() => logoutCurrentSession());
    await expectOperation(() =>
      changePassword({
        clientTxnId: "txn-password-change",
        currentPassword: "CurrentPass1!",
        newPassword: "ReplacementPass1!",
        secondFactorCode: "654321",
      }),
    );
    await expectOperation(() =>
      createLocalUser({
        clientTxnId: "txn-user-create",
        displayName: "Authentication User",
        email: "deployment-user@example.test",
        initialPassword: "InitialPass1!",
        isDeploymentAdmin: false,
        mfaRequired: true,
      }),
    );
    await expectOperation(() => listUsers());
    await expectOperation(() =>
      listUsers({ cursorToken: " cursor-2 ", limit: 50 }),
    );
    await expectOperation(() =>
      loadUser({ userId: "00000000-0000-4000-8000-000000000005" }),
    );
    await expectOperation(() =>
      patchLocalUser({
        baseUserVersion: 7,
        displayName: "Updated User",
        email: "00000000-0000-4000-8000-000000000005@example.test",
        isActive: true,
        isDeploymentAdmin: false,
        mfaRequired: true,
        userId: "00000000-0000-4000-8000-000000000005",
      }),
    );
    await expectOperation(() =>
      createEnterpriseAuthBinding({
        baseUserVersion: 8,
        clientTxnId: "txn-auth-binding-create",
        providerKey: "corp-oidc",
        providerSubject: "subject-1",
        reason: "",
        userId: "00000000-0000-4000-8000-000000000005",
      }),
    );
    await expectOperation(() =>
      rotateEnterpriseAuthBinding({
        authBindingId: "00000000-0000-4000-8000-000000000003",
        baseUserVersion: 9,
        clientTxnId: "txn-auth-binding-rotate",
        newProviderSubject: "subject-2",
        reason: "subject rotation",
        userId: "00000000-0000-4000-8000-000000000005",
      }),
    );
    await expectOperation(() =>
      retireEnterpriseAuthBinding({
        authBindingId: "00000000-0000-4000-8000-000000000003",
        baseUserVersion: 10,
        clientTxnId: "txn-auth-binding-retire",
        reason: "",
        userId: "00000000-0000-4000-8000-000000000005",
      }),
    );
    await expectOperation(() =>
      adminResetPassword({
        baseUserVersion: 7,
        clientTxnId: "txn-admin-password-reset",
        newPassword: "AdminResetPass1!",
        reason: "routine reset",
        userId: "00000000-0000-4000-8000-000000000005",
      }),
    );
    await expectOperation(() =>
      adminResetTotp({
        baseUserVersion: 7,
        clientTxnId: "txn-admin-totp-reset",
        reason: "routine reset",
        userId: "00000000-0000-4000-8000-000000000005",
      }),
    );
    await expectOperation(() =>
      adminRevokeAllSessions({
        clientTxnId: "txn-admin-revoke-all",
        reason: "routine revoke",
        userId: "00000000-0000-4000-8000-000000000005",
      }),
    );

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
        body: null,
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
        url: "/api/v1/users/00000000-0000-4000-8000-000000000005",
      },
      {
        body: {
          base_user_version: 7,
          display_name: "Updated User",
          email: "00000000-0000-4000-8000-000000000005@example.test",
          mfa_required: true,
          is_active: true,
          is_deployment_admin: false,
        },
        csrfHeader: "session-csrf",
        method: "PATCH",
        url: "/api/v1/users/00000000-0000-4000-8000-000000000005",
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
        url: "/api/v1/users/00000000-0000-4000-8000-000000000005/auth-bindings",
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
        url: "/api/v1/users/00000000-0000-4000-8000-000000000005/auth-bindings/00000000-0000-4000-8000-000000000003/rotate",
      },
      {
        body: {
          base_user_version: 10,
          client_txn_id: "txn-auth-binding-retire",
          reason: "",
        },
        csrfHeader: "session-csrf",
        method: "DELETE",
        url: "/api/v1/users/00000000-0000-4000-8000-000000000005/auth-bindings/00000000-0000-4000-8000-000000000003",
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
        url: "/api/v1/users/00000000-0000-4000-8000-000000000005/password/reset",
      },
      {
        body: {
          base_user_version: 7,
          client_txn_id: "txn-admin-totp-reset",
          reason: "routine reset",
        },
        csrfHeader: "session-csrf",
        method: "POST",
        url: "/api/v1/users/00000000-0000-4000-8000-000000000005/mfa/totp/reset",
      },
      {
        body: {
          client_txn_id: "txn-admin-revoke-all",
          reason: "routine revoke",
        },
        csrfHeader: "session-csrf",
        method: "POST",
        url: "/api/v1/users/00000000-0000-4000-8000-000000000005/sessions/revoke-all",
      },
    ];

    expect(requests).toHaveLength(expectedRequests.length);
    for (const [index, expected] of expectedRequests.entries()) {
      expectCookieBackedAPIRequest(requests[index], expected);
      expectClosedJSONBody(requests[index], expected.body);
    }
    await rejectMalformedSuccesses();
  });

  it("route-boundary bootstrap token authorization is limited to TOTP begin and complete", async () => {
    cookieValue = "cartulary_csrf=session-csrf";

    await expectOperation(() =>
      beginTotpEnrollment({
        authMode: "bootstrap",
        bootstrapToken: " bootstrap-token-1 ",
        clientTxnId: "txn-bootstrap-begin",
      }),
    );
    await expectOperation(() =>
      completeTotpEnrollment({
        authMode: "bootstrap",
        bootstrapToken: "bootstrap-token-1",
        clientTxnId: "txn-bootstrap-complete",
        code: "123456",
        enrollmentId: "00000000-0000-4000-8000-000000000002",
      }),
    );
    await expectOperation(() =>
      beginTotpEnrollment({
        authMode: "session",
        clientTxnId: "txn-session-begin",
        currentFactorCode: "654321",
        currentPassword: "CurrentPass1!",
      }),
    );
    await expectOperation(() =>
      completeTotpEnrollment({
        authMode: "session",
        clientTxnId: "txn-session-complete",
        code: "654321",
        enrollmentId: "enrollment-2",
      }),
    );

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
      enrollment_id: "00000000-0000-4000-8000-000000000002",
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
    await rejectMalformedSuccesses();
  });

  it("validates every remaining shell operation and rejects malformed successes", async () => {
    await expectOperation(() => listEnterpriseAuthProviders());
    await expectOperation(() =>
      beginEnterpriseAuth({
        providerKey: "corporate/key",
        returnTo: "/?incident_id=00000000-0000-4000-8000-000000000004",
      }),
    );
    expect(String(fetchMock.mock.calls.at(-1)?.[0])).toContain(
      "corporate%2Fkey/begin",
    );
    expect(JSON.parse(String(fetchMock.mock.calls.at(-1)?.[1]?.body))).toEqual({
      return_to: "/?incident_id=00000000-0000-4000-8000-000000000004",
    });
    await expectOperation(() => loadAccountProfile());
    await expectOperation(() =>
      patchAccountProfile({
        clientTxnId: "profile-test",
        baseUserVersion: 1,
        displayName: "Updated",
      }),
    );
    await expectOperation(() => loadAccountPreferences());
    await expectOperation(() =>
      putAccountPreferences({
        clientTxnId: "preferences-test",
        basePreferencesVersion: 1,
        densityMode: null,
      }),
    );
    await expectOperation(() => loadExtensions());
    await expectOperation(() => listVisibleIncidents());
    await expectOperation(() =>
      createIncident({
        request: {
          client_txn_id: "txn-create",
          incident_key: "INC-1",
          title: "Incident",
        },
        signal: new AbortController().signal,
      }),
    );
    await rejectMalformedSuccesses();
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

function shellFixture(url: string, method: string) {
  const path = new URL(url, "https://cartulary.test").pathname;
  const timestamp = "2026-04-20T12:00:00Z";
  const user = {
    user_id: "00000000-0000-4000-8000-000000000005",
    email: "operator@example.test",
    display_name: "Operator",
    is_active: true,
    is_deployment_admin: false,
    mfa_required: false,
    created_at: timestamp,
    updated_at: timestamp,
    updated_by_user_id: null,
    last_login_at: null,
    user_version: 1,
    auth_bindings: [
      {
        provider_key: "local",
        provider_type: "local",
        username: "operator@example.test",
        created_at: timestamp,
      },
    ],
  } satisfies UserResource;
  let data: unknown;
  let status = 200;
  switch (path) {
    case "/api/v1/auth/session":
    case "/api/v1/auth/login":
      data = sessionResource();
      break;
    case "/api/v1/auth/credential-state":
      data = credentialStateResource();
      break;
    case "/api/v1/auth/providers":
      data = {
        providers: [
          {
            provider_key: "corporate/key",
            provider_type: "oidc",
            display_name: "Corporate",
          },
        ],
      };
      break;
    case "/api/v1/auth/providers/corporate%2Fkey/begin":
      data = {
        provider_key: "corporate/key",
        provider_type: "oidc",
        redirect_url: "https://idp.example.test/authorize",
        expires_at: timestamp,
      };
      break;
    case "/api/v1/auth/logout":
      data = {
        user_id: "00000000-0000-4000-8000-000000000006",
        logged_out: true,
        sessions_revoked: false,
      };
      break;
    case "/api/v1/auth/password/change":
      data = {
        user_id: "00000000-0000-4000-8000-000000000006",
        password: { changed_at: timestamp },
        sessions_revoked: true,
      };
      break;
    case "/api/v1/auth/mfa/totp/begin":
      data = {
        enrollment_id: "00000000-0000-4000-8000-000000000002",
        expires_at: timestamp,
        totp_setup: {
          secret_base32: "JBSWY3DPEHPK3PXP",
          otpauth_uri: "otpauth://totp/Cartulary?secret=JBSWY3DPEHPK3PXP",
          algorithm: "SHA1",
          digits: 6,
          period_seconds: 30,
        },
      };
      break;
    case "/api/v1/auth/mfa/totp/complete":
      data = {
        user_id: "00000000-0000-4000-8000-000000000006",
        totp: { enrolled_at: timestamp },
        sessions_revoked: false,
      };
      break;
    case "/api/v1/account/profile":
      data = {
        user_id: "00000000-0000-4000-8000-000000000006",
        email: "operator@example.test",
        display_name: "Operator",
        user_version: 1,
        created_at: timestamp,
        updated_at: timestamp,
      };
      break;
    case "/api/v1/account/preferences":
      data = {
        user_id: "00000000-0000-4000-8000-000000000006",
        density_mode: null,
        preferences_version: 1,
        created_at: timestamp,
        updated_at: timestamp,
      };
      break;
    case "/api/v1/extensions":
      data = { extensions: [] };
      break;
    case "/api/v1/incidents":
      data =
        method === "GET"
          ? {
              incidents: [
                incidentResource(
                  "00000000-0000-4000-8000-000000000004",
                  "INC-1",
                  "Incident",
                ),
              ],
            }
          : incidentResource(
              "00000000-0000-4000-8000-000000000004",
              "INC-1",
              "Incident",
            );
      status = method === "GET" ? 200 : 201;
      break;
    case "/api/v1/users":
      data = method === "GET" ? { users: [user] } : user;
      status = method === "GET" ? 200 : 201;
      break;
    case "/api/v1/users/00000000-0000-4000-8000-000000000005/sessions/revoke-all":
      data = {
        user_id: "00000000-0000-4000-8000-000000000005",
        sessions_revoked: true,
        revoked_at: timestamp,
      };
      break;
    default:
      if (
        !path.startsWith("/api/v1/users/00000000-0000-4000-8000-000000000005")
      )
        throw new Error(`missing fixture: ${method} ${path}`);
      data = user;
      if (path.endsWith("/auth-bindings") && method === "POST") status = 201;
  }
  return {
    status,
    payload: {
      data,
      meta: {
        request_id: "req-shell",
        ...(method === "GET" &&
        ["/api/v1/users", "/api/v1/incidents"].includes(path)
          ? { paging: { limit: 100, has_more: false, next_cursor: null } }
          : {}),
      },
    },
  };
}
