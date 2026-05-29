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
  credentialStateResource,
  installLandingShellFetch,
  sessionResource,
} from "./appShellTestSupport";
import {
  deferred,
  errorResponse,
  expectStableFetchCount,
  findFetchCalls,
  jsonResponse,
  readHeader,
  requireJSONRequest,
} from "./fetchMockTestSupport";

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
    let sessionProbeCount = 0;
    const pendingInitialSession = deferred<Response>();
    installLandingShellFetch(fetchMock, {
      session: () => {
        sessionProbeCount += 1;
        if (sessionProbeCount === 1) {
          return pendingInitialSession.promise;
        }
        if (!authenticated) {
          return errorResponse("session_required", 401, {
            reason_code: "no_session",
          });
        }
        return sessionResource({
          display_name: "Phase 1 Operator",
        });
      },
      extraRoutes: [
        {
          method: "POST",
          url: "/api/v1/auth/login",
          handler: () => {
            authenticated = true;
            return jsonResponse({
              data: sessionResource({
                display_name: "Phase 1 Operator",
              }),
            });
          },
        },
      ],
    });

    renderApp();

    expect(
      (await screen.findByTestId("auth-shell")).getAttribute(
        "data-bootstrap-state",
      ),
    ).toBe("loading");
    expect(screen.getByTestId("auth-status").textContent).toBe(
      "Checking current session…",
    );
    pendingInitialSession.resolve(
      errorResponse("session_required", 401, {
        reason_code: "no_session",
      }),
    );

    expect(await screen.findByTestId("auth-login-username")).toBeTruthy();
    expect(
      screen.getByTestId("auth-shell").getAttribute("data-bootstrap-state"),
    ).toBe("anonymous");
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
    expect(
      screen.getByTestId("app-shell").getAttribute("data-bootstrap-state"),
    ).toBe("authenticated");
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

  it("Phase 1 U-1-14 ordinary shell blocks authenticated bootstrap until credential state loads and renders credential public errors without private details", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Phase 1 Operator",
      }),
      credentialState: () =>
        errorResponse("credential_bootstrap_rejected", 409, {
          reason_code: "not_allowed_for_route",
          bootstrap_token: "credential-bootstrap-token-must-not-render",
          request_id: "req-private-credential-detail",
          internal_path: "/var/lib/cartulary/credential.go",
        }),
    });

    renderApp();

    await screen.findByTestId("landing-current-user");
    expect(
      screen.getByTestId("app-shell").getAttribute("data-bootstrap-state"),
    ).toBe("public_error_envelope");
    expect(screen.getByTestId("landing-current-user").textContent).toContain(
      "Phase 1 Operator",
    );
    expect(screen.getByTestId("landing-error-code").textContent).toBe(
      "credential_bootstrap_rejected",
    );
    expect(screen.getByTestId("landing-error-details").textContent).toContain(
      "Reason: not_allowed_for_route",
    );
    expect(screen.getByTestId("account-error-code").textContent).toBe(
      "credential_bootstrap_rejected",
    );
    const credentialErrorText = document.body.textContent ?? "";
    expect(credentialErrorText).not.toContain(
      "credential-bootstrap-token-must-not-render",
    );
    expect(credentialErrorText).not.toContain("req-private-credential-detail");
    expect(credentialErrorText).not.toContain("/var/lib/cartulary");
    await expectStableFetchCount(fetchMock, 3);
  });

  it("Phase 1 U-1-15 ordinary shell follows mfa_setup_required through totp begin and complete, sends bootstrap-token requests, and proves completion alone does not issue a session", async () => {
    let loginAttempts = 0;
    installLandingShellFetch(fetchMock, {
      session: errorResponse("session_required", 401),
      extraRoutes: [
        {
          method: "POST",
          url: "/api/v1/auth/login",
          handler: () => {
            loginAttempts += 1;
            if (loginAttempts === 1) {
              return errorResponse("mfa_required", 401, {
                required_second_factor_kinds: ["totp"],
                bootstrap_token: "mfa-required-token-must-not-render",
                secret_base32: "MFAREQUIREDSECRET",
                otpauth_uri: "otpauth://mfa-required-private",
              });
            }
            return jsonResponse(
              {
                error: {
                  code: "mfa_setup_required",
                  status: 401,
                  message: "TOTP setup is required.",
                  request_id: "req-private-setup",
                  details: {
                    required_setup_kinds: ["totp"],
                    bootstrap_token: "bootstrap-token-123",
                    bootstrap_expires_at: "2026-04-17T12:10:00Z",
                    secret_base32: "ERRORSECRETBASE32",
                    otpauth_uri: "otpauth://private-error",
                    request_id: "req-private-detail",
                    stack:
                      "Error: private stack at /var/lib/cartulary/auth.go:12",
                  },
                },
              },
              401,
            );
          },
        },
        {
          method: "POST",
          url: "/api/v1/auth/mfa/totp/begin",
          handler: () =>
            jsonResponse({
              data: {
                enrollment_id: "enrollment-1",
                totp_setup: {
                  secret_base32: "JBSWY3DPEHPK3PXP",
                },
              },
            }),
        },
        {
          method: "POST",
          url: "/api/v1/auth/mfa/totp/complete",
          handler: () => jsonResponse({ data: {} }),
        },
      ],
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
        "mfa_required",
      );
    });
    expect(
      screen.getByTestId("auth-shell").getAttribute("data-bootstrap-state"),
    ).toBe("mfa_required");
    expect(screen.getByTestId("auth-error-details").textContent).toContain(
      "Required second factor kinds: totp",
    );
    expect(document.body.textContent ?? "").not.toContain(
      "mfa-required-token-must-not-render",
    );
    expect(document.body.textContent ?? "").not.toContain(
      "otpauth://mfa-required-private",
    );

    fireEvent.click(screen.getByTestId("auth-login-submit"));

    await waitFor(() => {
      expect(screen.getByTestId("auth-error-code").textContent).toBe(
        "mfa_setup_required",
      );
    });
    expect(
      screen.getByTestId("auth-shell").getAttribute("data-bootstrap-state"),
    ).toBe("mfa_setup_required");
    expect(screen.getByTestId("auth-error-message").textContent).toBe(
      "TOTP setup is required.",
    );
    expect(screen.getByTestId("auth-error-details").textContent).toContain(
      "Required setup kinds: totp",
    );
    expect(screen.getByTestId("auth-error-details").textContent).toContain(
      "Bootstrap expires at: 2026-04-17T12:10:00Z",
    );
    expect(screen.getByTestId("auth-bootstrap-token").textContent).toBe(
      "Stored for TOTP setup requests.",
    );
    const preBeginText = document.body.textContent ?? "";
    expect(preBeginText).not.toContain("bootstrap-token-123");
    expect(preBeginText).not.toContain("ERRORSECRETBASE32");
    expect(preBeginText).not.toContain("otpauth://private-error");
    expect(preBeginText).not.toContain("req-private-setup");
    expect(preBeginText).not.toContain("/var/lib/cartulary");
    expect(preBeginText).not.toContain("private stack");

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
    await expectStableFetchCount(fetchMock, 5);
  });

  it("Phase 1 U-1-16 ordinary account-security controls issue password-change and totp-enrollment requests, surface failures on the shell, and refresh back to anonymous state after success", async () => {
    let sessionActive = true;
    let totpBeginAttempts = 0;

    installLandingShellFetch(fetchMock, {
      session: () => {
        if (!sessionActive) {
          return errorResponse("session_required", 401);
        }
        return sessionResource({
          display_name: "Phase 1 Operator",
          mfa_state: "satisfied",
        });
      },
      credentialState: credentialStateResource({
        totp: {
          state: "active",
          enrolled_at: "2026-04-20T11:00:00Z",
          pending_expires_at: null,
        },
      }),
      extraRoutes: [
        {
          method: "POST",
          url: "/api/v1/auth/mfa/totp/begin",
          handler: () => {
            totpBeginAttempts += 1;
            if (totpBeginAttempts === 1) {
              return errorResponse("invalid_second_factor", 401);
            }
            return jsonResponse({
              data: {
                enrollment_id: "replacement-enrollment-1",
                totp_setup: {
                  secret_base32: "JBSWY3DPEHPK3PXP",
                },
              },
            });
          },
        },
        {
          method: "POST",
          url: "/api/v1/auth/password/change",
          handler: () => {
            sessionActive = false;
            return jsonResponse({ data: {} });
          },
        },
      ],
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
    expect(
      screen.getByTestId("auth-shell").getAttribute("data-bootstrap-state"),
    ).toBe("revoked");
    expect(screen.getByTestId("auth-error-code").textContent).toBe(
      "session_required",
    );
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

    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Deployment Admin",
        is_deployment_admin: true,
      }),
      incidents: () =>
        jsonResponse(
          {
            error: {
              code: "authorization_denied",
              status: 403,
              message: "Membership required.",
              request_id: "req-private-landing",
              details: {
                reason_code: "incident_membership_required",
                required_role: "viewer",
                request_id: "req-private-detail",
                internal_path: "/var/lib/cartulary/server.go",
                sql: "select * from sessions",
                bootstrap_token: "forbidden-bootstrap-token",
              },
            },
          },
          403,
        ),
      extraRoutes: [
        {
          method: "POST",
          url: "/api/v1/users",
          handler: () =>
            jsonResponse({
              data: createdUser,
            }),
        },
        {
          method: "GET",
          url: "/api/v1/users/user-2",
          handler: () =>
            jsonResponse({
              data: createdUser,
            }),
        },
        {
          method: "PATCH",
          url: "/api/v1/users/user-2",
          handler: () => errorResponse("user_version_conflict", 409),
        },
        {
          method: "GET",
          url: "/api/v1/users/user-1",
          handler: () =>
            jsonResponse({
              data: adminTarget,
            }),
        },
        {
          method: "PATCH",
          url: "/api/v1/users/user-1",
          handler: () => errorResponse("last_deployment_admin", 409),
        },
      ],
    });

    renderApp();

    await screen.findByTestId("landing-current-user");
    await waitFor(() => {
      expect(screen.getByTestId("landing-error-code").textContent).toBe(
        "authorization_denied",
      );
    });
    expect(
      screen
        .getByTestId("incident-landing")
        .getAttribute("data-bootstrap-state"),
    ).toBe("forbidden");
    expect(screen.getByTestId("landing-error-message").textContent).toBe(
      "Membership required.",
    );
    expect(screen.getByTestId("landing-error-details").textContent).toContain(
      "Reason: incident_membership_required",
    );
    expect(screen.getByTestId("landing-error-details").textContent).toContain(
      "Required role: viewer",
    );
    const forbiddenText = document.body.textContent ?? "";
    expect(forbiddenText).not.toContain("req-private-landing");
    expect(forbiddenText).not.toContain("/var/lib/cartulary");
    expect(forbiddenText).not.toContain("select * from sessions");
    expect(forbiddenText).not.toContain("forbidden-bootstrap-token");
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

    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Deployment Admin",
        is_deployment_admin: true,
      }),
      extraRoutes: [
        {
          method: "GET",
          url: "/api/v1/users/user-2",
          handler: () => pendingLoad.promise,
        },
        {
          method: "PATCH",
          url: "/api/v1/users/user-2",
          handler: () => errorResponse("user_version_conflict", 409),
        },
      ],
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
