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
import { csrfHeaderName } from "./browserApi";
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
    vi.restoreAllMocks();
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

  it("FE-I-P1-01 route-boundary auth login errors render public envelopes without private details", async () => {
    installLandingShellFetch(fetchMock, {
      session: errorResponse("session_required", 401),
      extraRoutes: [
        {
          method: "POST",
          url: "/api/v1/auth/login",
          handler: () =>
            publicErrorResponse("invalid_auth_request", 400, {
              message: "Login request is invalid.",
              publicDetails: {
                reason_code: "malformed_second_factor",
                field: "second_factor.assertion.code",
              },
            }),
        },
      ],
    });

    renderApp();

    await screen.findByTestId("auth-login-username");
    fireEvent.change(screen.getByTestId("auth-login-username"), {
      target: { value: "operator@example.test" },
    });
    fireEvent.change(screen.getByTestId("auth-login-password"), {
      target: { value: "OperatorPass1!" },
    });
    fireEvent.click(screen.getByTestId("auth-login-submit"));

    await waitFor(() => {
      expect(screen.getByTestId("auth-error-code").textContent).toBe(
        "invalid_auth_request",
      );
    });
    expect(screen.getByTestId("auth-error-message").textContent).toBe(
      "Login request is invalid.",
    );
    expect(screen.getByTestId("auth-error-details").textContent).toContain(
      "Reason: malformed_second_factor",
    );
    expect(screen.getByTestId("auth-error-details").textContent).toContain(
      "Field: second_factor.assertion.code",
    );
    expectPrivateErrorProbeNotRendered();
    await expectStableFetchCount(fetchMock, 2);
  });

  it("FE-I-P1-01 route-boundary credential-state errors render public envelopes on landing and account surfaces without private details", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Phase 1 Operator",
      }),
      credentialState: () =>
        publicErrorResponse("credential_bootstrap_rejected", 409, {
          message: "Credential bootstrap rejected.",
          publicDetails: {
            reason_code: "not_allowed_for_route",
          },
        }),
    });

    renderApp();

    await screen.findByTestId("landing-current-user");
    expect(
      screen.getByTestId("app-shell").getAttribute("data-bootstrap-state"),
    ).toBe("public_error_envelope");
    expect(screen.getByTestId("landing-error-code").textContent).toBe(
      "credential_bootstrap_rejected",
    );
    expect(screen.getByTestId("landing-error-message").textContent).toBe(
      "Credential bootstrap rejected.",
    );
    expect(screen.getByTestId("account-error-code").textContent).toBe(
      "credential_bootstrap_rejected",
    );
    expect(screen.getByTestId("account-error-details").textContent).toContain(
      "Reason: not_allowed_for_route",
    );
    expectPrivateErrorProbeNotRendered();
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

  it("FE-I-P1-01 route-boundary account password and TOTP errors render public envelopes without private details", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Phase 1 Operator",
        mfa_state: "satisfied",
      }),
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
          handler: () =>
            publicErrorResponse("invalid_second_factor", 401, {
              message: "Second factor rejected.",
              publicDetails: {
                reason_code: "totp_code_invalid",
                field: "second_factor.assertion.code",
              },
            }),
        },
        {
          method: "POST",
          url: "/api/v1/auth/password/change",
          handler: () =>
            publicErrorResponse("invalid_current_password", 409, {
              message: "Current password rejected.",
              publicDetails: {
                reason_code: "password_mismatch",
                field: "current_password",
              },
            }),
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
    expect(screen.getByTestId("account-error-message").textContent).toBe(
      "Second factor rejected.",
    );
    expect(screen.getByTestId("account-error-details").textContent).toContain(
      "Reason: totp_code_invalid",
    );
    expect(screen.getByTestId("account-error-details").textContent).toContain(
      "Field: second_factor.assertion.code",
    );
    expectPrivateErrorProbeNotRendered();

    fireEvent.change(screen.getByTestId("account-password-current"), {
      target: { value: "Wrong Current Phase 1 Password!" },
    });
    fireEvent.change(screen.getByTestId("account-password-next"), {
      target: { value: "Replacement Phase 1 Password!" },
    });
    fireEvent.change(screen.getByTestId("account-password-factor-code"), {
      target: { value: "654321" },
    });
    fireEvent.click(screen.getByTestId("account-password-change"));

    await waitFor(() => {
      expect(screen.getByTestId("account-error-code").textContent).toBe(
        "invalid_current_password",
      );
    });
    expect(screen.getByTestId("account-error-message").textContent).toBe(
      "Current password rejected.",
    );
    expect(screen.getByTestId("account-error-details").textContent).toContain(
      "Reason: password_mismatch",
    );
    expect(screen.getByTestId("account-error-details").textContent).toContain(
      "Field: current_password",
    );
    expectPrivateErrorProbeNotRendered();
    await expectStableFetchCount(fetchMock, 5);
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

  it("FE-I-P1-01 route-boundary deployment-admin action errors render public envelopes without private details", async () => {
    const loadedUser = userResource({
      user_id: "user-2",
      email: "loaded-target@example.test",
      display_name: "Loaded Target",
      user_version: 7,
    });
    let resetAttempt = 0;

    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Deployment Admin",
        is_deployment_admin: true,
      }),
      extraRoutes: [
        {
          method: "POST",
          url: "/api/v1/users",
          handler: () =>
            publicErrorResponse("invalid_user_create", 400, {
              message: "User create request is invalid.",
              publicDetails: {
                reason_code: "email_not_allowed",
                field: "email",
              },
            }),
        },
        {
          method: "GET",
          url: "/api/v1/users/user-2",
          handler: () =>
            jsonResponse({
              data: loadedUser,
            }),
        },
        {
          method: "PATCH",
          url: "/api/v1/users/user-2",
          handler: () =>
            publicErrorResponse("invalid_user_patch", 400, {
              message: "User patch request is invalid.",
              publicDetails: {
                reason_code: "display_name_invalid",
                field: "display_name",
              },
            }),
        },
        {
          method: "POST",
          url: "/api/v1/users/user-2/password/reset",
          handler: () =>
            publicErrorResponse("user_version_conflict", 409, {
              message: "User version conflict.",
              publicDetails: {
                reason_code: "stale_user_version",
                field: "base_user_version",
              },
            }),
        },
        {
          method: "POST",
          url: "/api/v1/users/user-2/mfa/totp/reset",
          handler: () => {
            resetAttempt += 1;
            return publicErrorResponse("invalid_mutation_payload", 400, {
              message: "TOTP reset request is invalid.",
              publicDetails: {
                reason_code: `totp_reset_invalid_${resetAttempt}`,
                field: "reason",
              },
            });
          },
        },
      ],
    });

    renderApp();

    await screen.findByTestId("landing-current-user");
    fireEvent.change(screen.getByTestId("admin-create-email"), {
      target: { value: "invalid-admin-target@example.test" },
    });
    fireEvent.change(screen.getByTestId("admin-create-display-name"), {
      target: { value: "Invalid Admin Target" },
    });
    fireEvent.change(screen.getByTestId("admin-create-password"), {
      target: { value: "AdminCreatePass1!" },
    });
    fireEvent.click(screen.getByTestId("admin-create-user"));

    await waitFor(() => {
      expect(screen.getByTestId("admin-error-code").textContent).toBe(
        "invalid_user_create",
      );
    });
    expect(screen.getByTestId("admin-error-message").textContent).toBe(
      "User create request is invalid.",
    );
    expect(screen.getByTestId("admin-error-details").textContent).toContain(
      "Reason: email_not_allowed",
    );
    expect(screen.getByTestId("admin-error-details").textContent).toContain(
      "Field: email",
    );
    expectPrivateErrorProbeNotRendered();

    fireEvent.change(screen.getByTestId("admin-target-user-id-input"), {
      target: { value: "user-2" },
    });
    fireEvent.click(screen.getByTestId("admin-load-user"));
    await waitFor(() => {
      expect(screen.getByTestId("admin-target-user-version").textContent).toBe(
        "7",
      );
    });
    fireEvent.click(screen.getByTestId("admin-patch-user"));

    await waitFor(() => {
      expect(screen.getByTestId("admin-error-code").textContent).toBe(
        "invalid_user_patch",
      );
    });
    expect(screen.getByTestId("admin-error-message").textContent).toBe(
      "User patch request is invalid.",
    );
    expect(screen.getByTestId("admin-error-details").textContent).toContain(
      "Reason: display_name_invalid",
    );
    expect(screen.getByTestId("admin-error-details").textContent).toContain(
      "Field: display_name",
    );
    expectPrivateErrorProbeNotRendered();

    fireEvent.change(screen.getByTestId("admin-new-password"), {
      target: { value: "AdminResetPass1!" },
    });
    fireEvent.change(screen.getByTestId("admin-reason"), {
      target: { value: "row-owned public error check" },
    });
    fireEvent.click(screen.getByTestId("admin-password-reset"));

    await waitFor(() => {
      expect(screen.getByTestId("admin-error-code").textContent).toBe(
        "user_version_conflict",
      );
    });
    expect(screen.getByTestId("admin-error-message").textContent).toBe(
      "User version conflict.",
    );
    expect(screen.getByTestId("admin-error-details").textContent).toContain(
      "Reason: stale_user_version",
    );
    expect(screen.getByTestId("admin-error-details").textContent).toContain(
      "Field: base_user_version",
    );
    expectPrivateErrorProbeNotRendered();

    fireEvent.click(screen.getByTestId("admin-totp-reset"));

    await waitFor(() => {
      expect(screen.getByTestId("admin-error-code").textContent).toBe(
        "invalid_mutation_payload",
      );
    });
    expect(screen.getByTestId("admin-error-message").textContent).toBe(
      "TOTP reset request is invalid.",
    );
    expect(screen.getByTestId("admin-error-details").textContent).toContain(
      "Reason: totp_reset_invalid_1",
    );
    expect(screen.getByTestId("admin-error-details").textContent).toContain(
      "Field: reason",
    );
    expectPrivateErrorProbeNotRendered();
    await expectStableFetchCount(fetchMock, 8);
  });

  it("FE-I-P1-01 route-boundary incident landing uses /api/v1/incidents with closed create bodies and public errors", async () => {
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=incident-csrf",
    );
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Phase 1 Operator",
      }),
      onCreateIncident: () =>
        jsonResponse(
          {
            error: {
              code: "invalid_incident_create",
              status: 400,
              message: "Incident create request is invalid.",
              request_id: "req-private-create",
              details: {
                reason_code: "unknown_field",
                field: "debug_field",
                request_id: "req-private-detail",
                internal_path: "/var/lib/cartulary/incidents.go",
                sql: "select * from incidents",
                bootstrap_token: "create-bootstrap-token",
                secret_base32: "CREATESECRETBASE32",
                otpauth_uri: "otpauth://create-private",
                stack:
                  "Error: private stack at /var/lib/cartulary/incidents.go:14",
                unknown_private_detail: "private-create-detail",
              },
            },
          },
          400,
        ),
    });

    renderApp();

    await screen.findByTestId("landing-current-user");
    expect(
      findFetchCalls(fetchMock, "/api/v1/auth/session", "GET"),
    ).toHaveLength(1);
    expect(
      findFetchCalls(fetchMock, "/api/v1/auth/credential-state", "GET"),
    ).toHaveLength(1);
    expect(findFetchCalls(fetchMock, "/api/v1/incidents", "GET")).toHaveLength(
      1,
    );
    const listIncidentRequest = findFetchCalls(
      fetchMock,
      "/api/v1/incidents",
      "GET",
    )[0]?.[1];
    expect(listIncidentRequest?.credentials).toBe("include");
    expect(readHeader(listIncidentRequest, "Authorization")).toBe("");
    expect(readHeader(listIncidentRequest, csrfHeaderName)).toBe("");

    fireEvent.change(screen.getByTestId("landing-incident-key"), {
      target: { value: "IR-2026-003" },
    });
    fireEvent.change(screen.getByTestId("landing-incident-title"), {
      target: { value: "Phase 1 boundary incident" },
    });
    fireEvent.click(screen.getByTestId("landing-create-button"));

    await waitFor(() => {
      expect(screen.getByTestId("landing-error-code").textContent).toBe(
        "invalid_incident_create",
      );
    });

    const createRequest = requireJSONRequest(
      fetchMock,
      "/api/v1/incidents",
      "POST",
    );
    expect(Object.keys(createRequest.body)).toEqual([
      "client_txn_id",
      "incident_key",
      "title",
    ]);
    expect(createRequest.body).toEqual({
      client_txn_id: expect.any(String),
      incident_key: "IR-2026-003",
      title: "Phase 1 boundary incident",
    });
    expect(createRequest.init?.credentials).toBe("include");
    expect(readHeader(createRequest.init, "Authorization")).toBe("");
    expect(readHeader(createRequest.init, csrfHeaderName)).toBe(
      "incident-csrf",
    );
    for (const call of fetchMock.mock.calls) {
      expect(String(call[0]).startsWith("/api/v1/")).toBe(true);
    }

    expect(screen.getByTestId("landing-error-message").textContent).toBe(
      "Incident create request is invalid.",
    );
    expect(screen.getByTestId("landing-error-details").textContent).toContain(
      "Reason: unknown_field",
    );
    expect(screen.getByTestId("landing-error-details").textContent).toContain(
      "Field: debug_field",
    );
    const renderedText = document.body.textContent ?? "";
    expect(renderedText).not.toContain("req-private-create");
    expect(renderedText).not.toContain("req-private-detail");
    expect(renderedText).not.toContain("/var/lib/cartulary");
    expect(renderedText).not.toContain("select * from incidents");
    expect(renderedText).not.toContain("create-bootstrap-token");
    expect(renderedText).not.toContain("CREATESECRETBASE32");
    expect(renderedText).not.toContain("otpauth://create-private");
    expect(renderedText).not.toContain("private stack");
    expect(renderedText).not.toContain("private-create-detail");
    await expectStableFetchCount(fetchMock, 4);
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

const privateErrorProbeDetails = {
  request_id: "req-private-detail",
  internal_path: "/var/lib/cartulary/private.go",
  sql: "select * from sessions",
  bootstrap_token: "private-bootstrap-token",
  secret_base32: "PRIVATESECRETBASE32",
  otpauth_uri: "otpauth://private-secret",
  stack: "Error: private stack at /var/lib/cartulary/private.go:12",
  unknown_private_detail: "private-rendering-detail",
};

const privateErrorProbeTokens = [
  "req-private-envelope",
  "req-private-detail",
  "/var/lib/cartulary",
  "select * from sessions",
  "private-bootstrap-token",
  "PRIVATESECRETBASE32",
  "otpauth://private-secret",
  "private stack",
  "private-rendering-detail",
];

function publicErrorResponse(
  code: string,
  status: number,
  options: {
    message: string;
    publicDetails: Record<string, unknown>;
  },
) {
  return jsonResponse(
    {
      error: {
        code,
        status,
        message: options.message,
        request_id: "req-private-envelope",
        details: {
          ...options.publicDetails,
          ...privateErrorProbeDetails,
        },
      },
    },
    status,
  );
}

function expectPrivateErrorProbeNotRendered() {
  const renderedText = document.body.textContent ?? "";
  for (const token of privateErrorProbeTokens) {
    expect(renderedText).not.toContain(token);
  }
}
