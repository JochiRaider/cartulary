import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("./WorkbookShell", () => ({
  WorkbookShell: vi.fn(),
  buildCreatePayload: vi.fn(),
  createDraftRow: vi.fn(),
  ensureDraftRow: vi.fn(),
  TimelineWorkbook: vi.fn(),
}));

import { AppRoot } from "./AppRoot";
import {
  findFetchCalls,
  readHeader,
  requireJSONRequest,
} from "./fetchMockTestSupport";
import type { CredentialState, SessionData } from "./phase1Client";

describe("Phase 1 ordinary app shell", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    window.history.replaceState({}, "", "/");
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("Phase 1 U-1-14 ordinary shell keeps the anonymous login surface, sends the login request, refreshes session and credential state after success, and keeps deployment-user controls denied for non-admin sessions", async () => {
    let authenticated = false;
    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();

      if (url === "/api/v1/auth/session") {
        if (!authenticated) {
          return Promise.resolve(errorResponse("session_required", 401));
        }
        return Promise.resolve(
          jsonResponse({
            data: sessionResource({
              display_name: "Phase 1 Operator",
            }),
          }),
        );
      }
      if (url === "/api/v1/auth/login" && method === "POST") {
        authenticated = true;
        return Promise.resolve(
          jsonResponse({
            data: sessionResource({
              display_name: "Phase 1 Operator",
            }),
          }),
        );
      }
      if (url === "/api/v1/auth/credential-state") {
        return Promise.resolve(
          jsonResponse({
            data: credentialStateResource(),
          }),
        );
      }
      if (url === "/api/v1/incidents") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incidents: [],
            },
          }),
        );
      }
      throw new Error(`unexpected fetch: ${method} ${url}`);
    });

    renderApp();

    expect(await screen.findByTestId("auth-login-username")).toBeTruthy();
    expect(screen.getByTestId("auth-shell-message").textContent).toContain(
      "Sign in with your local account",
    );
    fireEvent.change(screen.getByTestId("auth-login-username"), {
      target: { value: "operator@example.test" },
    });
    fireEvent.change(screen.getByTestId("auth-login-password"), {
      target: { value: "OperatorPass1!" },
    });
    fireEvent.click(screen.getByTestId("auth-login-submit"));

    await waitFor(() => {
      expect(screen.getByTestId("landing-current-user").textContent).toContain(
        "Phase 1 Operator",
      );
    });
    expect(screen.getByTestId("account-session-user-id").textContent).toBe(
      "user-1",
    );
    expect(screen.getByTestId("admin-access-note").textContent).toContain(
      "Deployment admin access is required",
    );
    const loginRequest = requireJSONRequest(
      fetchMock,
      "/api/v1/auth/login",
      "POST",
    );
    expect(loginRequest.body).toEqual({
      username: "operator@example.test",
      password: "OperatorPass1!",
    });
    await expectStableFetchCount(fetchMock, 5);
  });

  it("Phase 1 U-1-15 ordinary shell follows mfa_setup_required through totp begin and complete, sends bootstrap-token requests, and proves completion alone does not issue a session", async () => {
    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();

      if (url === "/api/v1/auth/session") {
        return Promise.resolve(errorResponse("session_required", 401));
      }
      if (url === "/api/v1/auth/login" && method === "POST") {
        return Promise.resolve(
          errorResponse("mfa_setup_required", 409, {
            bootstrap_token: "bootstrap-token-123",
          }),
        );
      }
      if (url === "/api/v1/auth/mfa/totp/begin" && method === "POST") {
        return Promise.resolve(
          jsonResponse({
            data: {
              enrollment_id: "enrollment-1",
              totp_setup: {
                secret_base32: "JBSWY3DPEHPK3PXP",
              },
            },
          }),
        );
      }
      if (url === "/api/v1/auth/mfa/totp/complete" && method === "POST") {
        return Promise.resolve(jsonResponse({ data: {} }));
      }
      throw new Error(`unexpected fetch: ${method} ${url}`);
    });

    renderApp();

    await screen.findByTestId("auth-login-username");
    fireEvent.change(screen.getByTestId("auth-login-username"), {
      target: { value: "bootstrap@example.test" },
    });
    fireEvent.change(screen.getByTestId("auth-login-password"), {
      target: { value: "BootstrapPass1!" },
    });
    fireEvent.click(screen.getByTestId("auth-login-submit"));

    await waitFor(() => {
      expect(screen.getByTestId("auth-error-code").textContent).toBe(
        "mfa_setup_required",
      );
    });
    expect(screen.getByTestId("auth-bootstrap-token").textContent).toBe(
      "bootstrap-token-123",
    );

    fireEvent.click(screen.getByTestId("auth-bootstrap-begin"));
    await waitFor(() => {
      expect(
        screen.getByTestId("auth-bootstrap-secret-base32").textContent,
      ).toBe("JBSWY3DPEHPK3PXP");
    });

    fireEvent.change(screen.getByTestId("auth-bootstrap-complete-code"), {
      target: { value: "123456" },
    });
    fireEvent.click(screen.getByTestId("auth-bootstrap-complete"));

    await waitFor(() => {
      expect(screen.getByTestId("auth-status").textContent).toContain(
        "TOTP enrollment completed",
      );
    });
    const beginRequest = requireJSONRequest(
      fetchMock,
      "/api/v1/auth/mfa/totp/begin",
      "POST",
    );
    expect(typeof beginRequest.body.client_txn_id).toBe("string");
    expect(beginRequest.init?.credentials).toBe("omit");
    expect(readHeader(beginRequest.init, "Authorization")).toBe(
      "Bearer bootstrap-token-123",
    );

    const completeRequest = requireJSONRequest(
      fetchMock,
      "/api/v1/auth/mfa/totp/complete",
      "POST",
    );
    expect(typeof completeRequest.body.client_txn_id).toBe("string");
    expect(completeRequest.body.enrollment_id).toBe("enrollment-1");
    expect(completeRequest.body.code).toBe("123456");
    expect(completeRequest.init?.credentials).toBe("omit");
    expect(readHeader(completeRequest.init, "Authorization")).toBe(
      "Bearer bootstrap-token-123",
    );

    expect(
      findFetchCalls(fetchMock, "/api/v1/auth/session", "GET"),
    ).toHaveLength(1);
    expect(screen.queryByTestId("landing-current-user")).toBeNull();
    await expectStableFetchCount(fetchMock, 4);
  });

  it("Phase 1 U-1-16 ordinary account-security controls issue password-change and totp-enrollment requests, surface failures on the shell, and refresh back to anonymous state after success", async () => {
    let sessionActive = true;
    let totpBeginAttempts = 0;

    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();

      if (url === "/api/v1/auth/session") {
        if (!sessionActive) {
          return Promise.resolve(errorResponse("session_required", 401));
        }
        return Promise.resolve(
          jsonResponse({
            data: sessionResource({
              display_name: "Phase 1 Operator",
              mfa_state: "satisfied",
            }),
          }),
        );
      }
      if (url === "/api/v1/auth/credential-state") {
        return Promise.resolve(
          jsonResponse({
            data: credentialStateResource({
              totp: {
                state: "active",
                enrolled_at: "2026-04-20T11:00:00Z",
                pending_expires_at: null,
              },
            }),
          }),
        );
      }
      if (url === "/api/v1/incidents") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incidents: [],
            },
          }),
        );
      }
      if (url === "/api/v1/auth/mfa/totp/begin" && method === "POST") {
        totpBeginAttempts += 1;
        if (totpBeginAttempts === 1) {
          return Promise.resolve(errorResponse("invalid_second_factor", 401));
        }
        return Promise.resolve(
          jsonResponse({
            data: {
              enrollment_id: "replacement-enrollment-1",
              totp_setup: {
                secret_base32: "JBSWY3DPEHPK3PXP",
              },
            },
          }),
        );
      }
      if (url === "/api/v1/auth/password/change" && method === "POST") {
        sessionActive = false;
        return Promise.resolve(jsonResponse({ data: {} }));
      }
      throw new Error(`unexpected fetch: ${method} ${url}`);
    });

    renderApp();

    await screen.findByTestId("account-session-user-id");

    fireEvent.change(screen.getByTestId("account-totp-current-password"), {
      target: { value: "Current Phase 1 Password!" },
    });
    fireEvent.change(screen.getByTestId("account-totp-current-factor"), {
      target: { value: "111111" },
    });
    fireEvent.click(screen.getByTestId("account-totp-begin"));

    await waitFor(() => {
      expect(screen.getByTestId("account-error-code").textContent).toBe(
        "invalid_second_factor",
      );
    });
    expect(screen.getByTestId("account-status").textContent).toBe(
      "TOTP begin failed",
    );

    fireEvent.change(screen.getByTestId("account-totp-current-factor"), {
      target: { value: "222222" },
    });
    fireEvent.click(screen.getByTestId("account-totp-begin"));

    await waitFor(() => {
      expect(screen.getByTestId("account-totp-secret-base32").textContent).toBe(
        "JBSWY3DPEHPK3PXP",
      );
    });

    fireEvent.change(screen.getByTestId("account-password-current"), {
      target: { value: "Current Phase 1 Password!" },
    });
    fireEvent.change(screen.getByTestId("account-password-next"), {
      target: { value: "Replacement Phase 1 Password!" },
    });
    fireEvent.change(screen.getByTestId("account-password-factor-code"), {
      target: { value: "654321" },
    });
    fireEvent.click(screen.getByTestId("account-password-change"));

    await waitFor(() => {
      expect(screen.getByTestId("auth-login-username")).toBeTruthy();
    });
    expect(screen.getByTestId("auth-shell-message").textContent).toContain(
      "Password changed. Sign in again.",
    );

    const totpBeginRequests = findFetchCalls(
      fetchMock,
      "/api/v1/auth/mfa/totp/begin",
      "POST",
    ).map((_call, index) =>
      requireJSONRequest(
        fetchMock,
        "/api/v1/auth/mfa/totp/begin",
        "POST",
        index,
      ),
    );
    expect(totpBeginRequests).toHaveLength(2);
    for (const request of totpBeginRequests) {
      expect(typeof request.body.client_txn_id).toBe("string");
      expect(request.body.current_password).toBe("Current Phase 1 Password!");
      expect(request.body.second_factor).toEqual({
        kind: "totp",
        assertion: {
          code: expect.any(String),
        },
      });
    }
    expect(
      (
        totpBeginRequests[0]?.body.second_factor as {
          assertion?: { code?: string };
        }
      )?.assertion?.code,
    ).toBe("111111");
    expect(
      (
        totpBeginRequests[1]?.body.second_factor as {
          assertion?: { code?: string };
        }
      )?.assertion?.code,
    ).toBe("222222");

    const passwordChangeRequest = requireJSONRequest(
      fetchMock,
      "/api/v1/auth/password/change",
      "POST",
    );
    expect(typeof passwordChangeRequest.body.client_txn_id).toBe("string");
    expect(passwordChangeRequest.body).toEqual({
      client_txn_id: expect.any(String),
      current_password: "Current Phase 1 Password!",
      new_password: "Replacement Phase 1 Password!",
      second_factor: {
        kind: "totp",
        assertion: {
          code: "654321",
        },
      },
    });
    await expectStableFetchCount(fetchMock, 7);
  });

  it("Phase 1 U-1-17 ordinary deployment-admin controls create and load users, send versioned patch requests, and surface user_version_conflict plus last_deployment_admin on the shell", async () => {
    const createdUser = userResource({
      user_id: "user-2",
      email: "phase1-admin@example.test",
      display_name: "Phase 1 Admin Target",
      user_version: 1,
    });
    const adminTarget = userResource({
      user_id: "user-1",
      email: "deployment-admin@example.test",
      display_name: "Deployment Admin",
      user_version: 9,
      is_deployment_admin: true,
    });

    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();

      if (url === "/api/v1/auth/session") {
        return Promise.resolve(
          jsonResponse({
            data: sessionResource({
              display_name: "Deployment Admin",
              is_deployment_admin: true,
            }),
          }),
        );
      }
      if (url === "/api/v1/auth/credential-state") {
        return Promise.resolve(
          jsonResponse({
            data: credentialStateResource(),
          }),
        );
      }
      if (url === "/api/v1/incidents") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incidents: [],
            },
          }),
        );
      }
      if (url === "/api/v1/users" && method === "POST") {
        return Promise.resolve(
          jsonResponse({
            data: createdUser,
          }),
        );
      }
      if (url === "/api/v1/users/user-2" && method === "GET") {
        return Promise.resolve(
          jsonResponse({
            data: createdUser,
          }),
        );
      }
      if (url === "/api/v1/users/user-2" && method === "PATCH") {
        return Promise.resolve(errorResponse("user_version_conflict", 409));
      }
      if (url === "/api/v1/users/user-1" && method === "GET") {
        return Promise.resolve(
          jsonResponse({
            data: adminTarget,
          }),
        );
      }
      if (url === "/api/v1/users/user-1" && method === "PATCH") {
        return Promise.resolve(errorResponse("last_deployment_admin", 409));
      }
      throw new Error(`unexpected fetch: ${method} ${url}`);
    });

    renderApp();

    await screen.findByTestId("landing-current-user");
    fireEvent.change(screen.getByTestId("admin-create-email"), {
      target: { value: createdUser.email },
    });
    fireEvent.change(screen.getByTestId("admin-create-display-name"), {
      target: { value: createdUser.display_name },
    });
    fireEvent.change(screen.getByTestId("admin-create-password"), {
      target: { value: "CreatedPass1!" },
    });
    fireEvent.click(screen.getByTestId("admin-create-user"));

    await waitFor(() => {
      expect(screen.getByTestId("admin-target-user-version").textContent).toBe(
        "1",
      );
    });
    expect(screen.getByTestId("admin-status").textContent).toBe(
      "Created local user",
    );
    const createRequest = requireJSONRequest(
      fetchMock,
      "/api/v1/users",
      "POST",
    );
    expect(typeof createRequest.body.client_txn_id).toBe("string");
    expect(createRequest.body).toEqual({
      client_txn_id: expect.any(String),
      auth_kind: "local",
      email: "phase1-admin@example.test",
      display_name: "Phase 1 Admin Target",
      initial_password: "CreatedPass1!",
      mfa_required: true,
      is_deployment_admin: false,
    });

    fireEvent.click(screen.getByTestId("admin-load-user"));
    await waitFor(() => {
      expect(screen.getByTestId("admin-status").textContent).toBe(
        "Loaded target user",
      );
    });
    fireEvent.click(screen.getByTestId("admin-patch-user"));

    await waitFor(() => {
      expect(screen.getByTestId("admin-error-code").textContent).toBe(
        "user_version_conflict",
      );
    });
    const patchConflictRequest = requireJSONRequest(
      fetchMock,
      "/api/v1/users/user-2",
      "PATCH",
    );
    expect(patchConflictRequest.body).toEqual({
      base_user_version: 1,
      display_name: "Phase 1 Admin Target",
      mfa_required: true,
      is_active: true,
      is_deployment_admin: false,
    });

    fireEvent.change(screen.getByTestId("admin-target-user-id-input"), {
      target: { value: "user-1" },
    });
    fireEvent.click(screen.getByTestId("admin-load-user"));
    await waitFor(() => {
      expect(screen.getByTestId("admin-target-user-version").textContent).toBe(
        "9",
      );
    });
    fireEvent.click(screen.getByTestId("admin-patch-is-deployment-admin"));
    fireEvent.click(screen.getByTestId("admin-patch-user"));

    await waitFor(() => {
      expect(screen.getByTestId("admin-error-code").textContent).toBe(
        "last_deployment_admin",
      );
    });
    const lastAdminPatchRequest = requireJSONRequest(
      fetchMock,
      "/api/v1/users/user-1",
      "PATCH",
    );
    expect(lastAdminPatchRequest.body).toEqual({
      base_user_version: 9,
      display_name: "Deployment Admin",
      mfa_required: true,
      is_active: true,
      is_deployment_admin: false,
    });
    expect(screen.getByTestId("admin-status").textContent).toBe(
      "Patch local user failed",
    );
    await expectStableFetchCount(fetchMock, 11);
  });

  it("keeps deployment-admin target actions disabled until a target load completes and leaves version-conflict status stable", async () => {
    const loadedUser = userResource({
      user_id: "user-2",
      email: "loaded-target@example.test",
      display_name: "Loaded Target",
      user_version: 7,
    });
    const pendingLoad = deferred<Response>();

    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();

      if (url === "/api/v1/auth/session") {
        return Promise.resolve(
          jsonResponse({
            data: sessionResource({
              display_name: "Deployment Admin",
              is_deployment_admin: true,
            }),
          }),
        );
      }
      if (url === "/api/v1/auth/credential-state") {
        return Promise.resolve(
          jsonResponse({
            data: credentialStateResource(),
          }),
        );
      }
      if (url === "/api/v1/incidents") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incidents: [],
            },
          }),
        );
      }
      if (url === "/api/v1/users/user-2" && method === "GET") {
        return pendingLoad.promise;
      }
      if (url === "/api/v1/users/user-2" && method === "PATCH") {
        return Promise.resolve(errorResponse("user_version_conflict", 409));
      }
      throw new Error(`unexpected fetch: ${method} ${url}`);
    });

    renderApp();

    await screen.findByTestId("landing-current-user");
    expect(
      (screen.getByTestId("admin-patch-user") as HTMLButtonElement).disabled,
    ).toBe(true);
    expect(
      (screen.getByTestId("admin-password-reset") as HTMLButtonElement)
        .disabled,
    ).toBe(true);

    fireEvent.change(screen.getByTestId("admin-target-user-id-input"), {
      target: { value: "user-2" },
    });
    fireEvent.click(screen.getByTestId("admin-load-user"));
    await waitFor(() => {
      expect(screen.getByTestId("admin-status").textContent).toBe(
        "Loading target user",
      );
    });

    expect(
      (screen.getByTestId("admin-patch-user") as HTMLButtonElement).disabled,
    ).toBe(true);
    expect(
      (screen.getByTestId("admin-password-reset") as HTMLButtonElement)
        .disabled,
    ).toBe(true);
    fireEvent.click(screen.getByTestId("admin-password-reset"));
    fireEvent.click(screen.getByTestId("admin-patch-user"));
    expect(
      findFetchCalls(fetchMock, "/api/v1/users/user-2/password/reset", "POST"),
    ).toHaveLength(0);
    expect(
      findFetchCalls(fetchMock, "/api/v1/users/user-2", "PATCH"),
    ).toHaveLength(0);

    pendingLoad.resolve(jsonResponse({ data: loadedUser }));
    await waitFor(() => {
      expect(screen.getByTestId("admin-target-user-version").textContent).toBe(
        "7",
      );
    });
    expect(screen.getByTestId("admin-status").textContent).toBe(
      "Loaded target user",
    );
    expect(
      (screen.getByTestId("admin-patch-user") as HTMLButtonElement).disabled,
    ).toBe(false);

    fireEvent.change(screen.getByTestId("admin-patch-base-version"), {
      target: { value: "1" },
    });
    fireEvent.click(screen.getByTestId("admin-patch-user"));

    await waitFor(() => {
      expect(screen.getByTestId("admin-error-code").textContent).toBe(
        "user_version_conflict",
      );
    });
    expect(screen.getByTestId("admin-status").textContent).toBe(
      "Patch local user failed",
    );
    const patchRequest = requireJSONRequest(
      fetchMock,
      "/api/v1/users/user-2",
      "PATCH",
    );
    expect(patchRequest.body.base_user_version).toBe(1);
    await expectStableFetchCount(fetchMock, 5);
    expect(
      findFetchCalls(fetchMock, "/api/v1/users/user-2", "GET"),
    ).toHaveLength(1);
    expect(screen.getByTestId("admin-status").textContent).toBe(
      "Patch local user failed",
    );
  });
});

function renderApp() {
  return render(<AppRoot />);
}

function sessionResource(overrides?: Partial<SessionData>): SessionData {
  return {
    user_id: "user-1",
    display_name: "Operator",
    provider_type: "local",
    mfa_state: "not_required",
    is_deployment_admin: false,
    authenticated_at: "2026-04-20T12:00:00Z",
    idle_expires_at: "2026-04-20T12:30:00Z",
    absolute_expires_at: "2026-04-20T20:00:00Z",
    session_expires_at: "2026-04-20T12:30:00Z",
    memberships: [],
    ...overrides,
  };
}

function credentialStateResource(
  overrides?: Partial<CredentialState>,
): CredentialState {
  const baseTotp = {
    state: "not_enrolled" as const,
    enrolled_at: null,
    pending_expires_at: null,
  };
  return {
    user_id: "user-1",
    auth_kind: "local",
    recovery_model: "admin_assisted",
    password_changed_at: "2026-04-20T12:00:00Z",
    ...overrides,
    totp: {
      ...baseTotp,
      ...(overrides?.totp ?? {}),
    },
  };
}

function userResource(
  overrides?: Partial<{
    display_name: string;
    email: string;
    is_active: boolean;
    is_deployment_admin: boolean;
    mfa_required: boolean;
    user_id: string;
    user_version: number;
  }>,
) {
  return {
    user_id: "user-2",
    email: "user-2@example.test",
    display_name: "User Two",
    user_version: 1,
    is_active: true,
    mfa_required: true,
    is_deployment_admin: false,
    ...overrides,
  };
}

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}

async function expectStableFetchCount(
  fetchMock: ReturnType<typeof vi.fn>,
  expectedCount: number,
) {
  await waitFor(() => {
    expect(fetchMock).toHaveBeenCalledTimes(expectedCount);
  });
  await flushMicrotasks();
  expect(fetchMock).toHaveBeenCalledTimes(expectedCount);
}

async function flushMicrotasks() {
  await Promise.resolve();
  await Promise.resolve();
}

function deferred<T>() {
  let resolve: (value: T) => void = () => {};
  let reject: (reason?: unknown) => void = () => {};
  const promise = new Promise<T>((nextResolve, nextReject) => {
    resolve = nextResolve;
    reject = nextReject;
  });
  return { promise, reject, resolve };
}

function errorResponse(
  code: string,
  status: number,
  details?: Record<string, unknown>,
) {
  return jsonResponse(
    {
      error: {
        code,
        status,
        ...(details ? { details } : {}),
      },
    },
    status,
  );
}
