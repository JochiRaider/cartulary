import {
  deploymentUserRowTestId,
  landingAdminMenuItemTestId,
  phase1AccountTestId,
  phase1AdminTestId,
  phase1AuthTestId,
  phase1ErrorCodeTestId,
  phase1ErrorSummaryTestIds,
  phase1LandingTestId,
  phase1RouteTestId,
} from "@cartulary/ui-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../workbook/WorkbookShell", () => ({
  WorkbookShell: vi.fn(),
  buildCreatePayload: vi.fn(),
  createDraftRow: vi.fn(),
  ensureDraftRow: vi.fn(),
  TimelineWorkbook: vi.fn(),
}));

import { csrfHeaderName } from "../services/browserApi";
import {
  credentialStateResource,
  installLandingShellFetch,
  sessionResource,
} from "../testing/appShellTestSupport";
import {
  deferred,
  errorResponse,
  expectStableFetchCount,
  findFetchCalls,
  jsonResponse,
  readHeader,
  requireJSONRequest,
} from "../testing/fetchMockTestSupport";
import { AppRoot } from "./AppRoot";
import { setEnterpriseAuthNavigateForTesting } from "./AuthGateway";

describe("ordinary app shell", () => {
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

  it("ordinary shell keeps the anonymous login surface, sends the login request, refreshes session and credential state after success, and keeps deployment-user controls denied for non-admin sessions", async () => {
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
      (await screen.findByTestId(phase1AuthTestId("shell"))).getAttribute(
        "data-bootstrap-state",
      ),
    ).toBe("loading");
    expect(screen.getByTestId(phase1AuthTestId("status")).textContent).toBe(
      "Checking current session...",
    );
    pendingInitialSession.resolve(
      errorResponse("session_required", 401, {
        reason_code: "no_session",
      }),
    );

    expect(
      await screen.findByTestId(phase1AuthTestId("login-username")),
    ).toBeTruthy();
    expect(
      screen
        .getByTestId(phase1AuthTestId("shell"))
        .getAttribute("data-bootstrap-state"),
    ).toBe("anonymous");
    expect(
      screen.getByTestId(phase1AuthTestId("shell-message")).textContent,
    ).toContain("Use your deployment account.");
    expect(
      screen.queryByTestId(phase1AuthTestId("login-totp-code")),
    ).toBeNull();
    fireEvent.change(screen.getByTestId(phase1AuthTestId("login-username")), {
      target: { value: "operator@example.test" },
    });
    fireEvent.change(screen.getByTestId(phase1AuthTestId("login-password")), {
      target: { value: "OperatorPass1!" },
    });
    fireEvent.click(screen.getByTestId(phase1AuthTestId("login-submit")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1LandingTestId("current-user")).textContent,
      ).toContain("Phase 1 Operator");
    });
    expect(
      screen
        .getByTestId(phase1RouteTestId("app-shell"))
        .getAttribute("data-bootstrap-state"),
    ).toBe("authenticated");
    await openAccountSecurity();
    expect(screen.getByTestId(phase1AccountTestId("logout")).textContent).toBe(
      "Sign out",
    );
    expect(screen.queryByText("Incident memberships")).toBe(null);
    expect(
      screen.queryByTestId(landingAdminMenuItemTestId("deployment-users")),
    ).toBe(null);
    expect(screen.queryByTestId(phase1AdminTestId("access-note"))).toBe(null);
    const loginRequest = requireJSONRequest(
      fetchMock,
      "/api/v1/auth/login",
      "POST",
    );
    expect(loginRequest.body).toEqual({
      username: "operator@example.test",
      password: "OperatorPass1!",
    });
    await expectStableFetchCount(fetchMock, 11);
  });

  it("enterprise auth discovery renders provider sign-in and begins with a relative return_to", async () => {
    const navigateSpy = vi.fn();
    const restoreNavigate = setEnterpriseAuthNavigateForTesting(navigateSpy);
    try {
      installLandingShellFetch(fetchMock, {
        session: errorResponse("session_required", 401),
        enterpriseProviders: {
          providers: [
            {
              provider_key: "corp-oidc",
              provider_type: "oidc",
              display_name: "Corporate OIDC",
            },
          ],
        },
        extraRoutes: [
          {
            method: "POST",
            url: "/api/v1/auth/providers/corp-oidc/begin",
            handler: () =>
              jsonResponse({
                data: {
                  provider_key: "corp-oidc",
                  provider_type: "oidc",
                  redirect_url: "https://idp.example.test/start",
                  expires_at: "2026-06-13T22:30:00Z",
                },
              }),
          },
        ],
      });

      renderApp();

      const providerButton = await screen.findByTestId(
        phase1AuthTestId("enterprise-provider-button"),
      );
      expect(providerButton.textContent).toBe("Corporate OIDC");
      fireEvent.click(providerButton);

      await waitFor(() => {
        expect(navigateSpy).toHaveBeenCalledWith(
          "https://idp.example.test/start",
        );
      });
      const beginRequest = requireJSONRequest(
        fetchMock,
        "/api/v1/auth/providers/corp-oidc/begin",
        "POST",
      );
      expect(beginRequest.body).toEqual({
        return_to: "/",
      });
      await expectStableFetchCount(fetchMock, 3);
    } finally {
      restoreNavigate();
    }
  });

  it("ordinary shell blocks authenticated bootstrap until credential state loads and renders credential public errors without private details", async () => {
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

    await screen.findByTestId(phase1LandingTestId("current-user"));
    expect(
      screen
        .getByTestId(phase1RouteTestId("app-shell"))
        .getAttribute("data-bootstrap-state"),
    ).toBe("public_error_envelope");
    expect(
      screen.getByTestId(phase1LandingTestId("current-user")).textContent,
    ).toContain("Phase 1 Operator");
    expect(
      screen.getByTestId(phase1ErrorCodeTestId("landing")).textContent,
    ).toBe("credential_bootstrap_rejected");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("landing").details)
        .textContent,
    ).toContain("Reason: not_allowed_for_route");
    await openAccountSecurity();
    expect(
      screen.getByTestId(phase1ErrorCodeTestId("account")).textContent,
    ).toBe("credential_bootstrap_rejected");
    const credentialErrorText = document.body.textContent ?? "";
    expect(credentialErrorText).not.toContain(
      "credential-bootstrap-token-must-not-render",
    );
    expect(credentialErrorText).not.toContain("req-private-credential-detail");
    expect(credentialErrorText).not.toContain("/var/lib/cartulary");
    await expectStableFetchCount(fetchMock, 8);
  });

  it("route-boundary auth login errors render public envelopes without private details", async () => {
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

    await screen.findByTestId(phase1AuthTestId("login-username"));
    fireEvent.change(screen.getByTestId(phase1AuthTestId("login-username")), {
      target: { value: "operator@example.test" },
    });
    fireEvent.change(screen.getByTestId(phase1AuthTestId("login-password")), {
      target: { value: "OperatorPass1!" },
    });
    fireEvent.click(screen.getByTestId(phase1AuthTestId("login-submit")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("auth")).textContent,
      ).toBe("Sign-in request could not be completed.");
    });
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("auth").message).textContent,
    ).toBe("Sign-in request could not be completed.");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("auth").details).textContent,
    ).toBe("");
    expect(
      screen
        .getByTestId(phase1ErrorCodeTestId("auth"))
        .getAttribute("data-error-code"),
    ).toBe("invalid_auth_request");
    expectPrivateErrorProbeNotRendered();
    await expectStableFetchCount(fetchMock, 3);
  });

  it("route-boundary credential-state errors render public envelopes on landing and account surfaces without private details", async () => {
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

    await screen.findByTestId(phase1LandingTestId("current-user"));
    expect(
      screen
        .getByTestId(phase1RouteTestId("app-shell"))
        .getAttribute("data-bootstrap-state"),
    ).toBe("public_error_envelope");
    expect(
      screen.getByTestId(phase1ErrorCodeTestId("landing")).textContent,
    ).toBe("credential_bootstrap_rejected");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("landing").message)
        .textContent,
    ).toBe("Credential bootstrap rejected.");
    await openAccountSecurity();
    expect(
      screen.getByTestId(phase1ErrorCodeTestId("account")).textContent,
    ).toBe("credential_bootstrap_rejected");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("account").details)
        .textContent,
    ).toContain("Reason: not_allowed_for_route");
    expectPrivateErrorProbeNotRendered();
    await expectStableFetchCount(fetchMock, 8);
  });

  it("ordinary shell follows mfa_setup_required through totp begin and complete, sends bootstrap-token requests, and proves completion alone does not issue a session", async () => {
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

    await screen.findByTestId(phase1AuthTestId("login-username"));
    fireEvent.change(screen.getByTestId(phase1AuthTestId("login-username")), {
      target: { value: "bootstrap@example.test" },
    });
    fireEvent.change(screen.getByTestId(phase1AuthTestId("login-password")), {
      target: { value: "BootstrapPass1!" },
    });
    fireEvent.click(screen.getByTestId(phase1AuthTestId("login-submit")));

    await waitFor(() => {
      expect(
        screen
          .getByTestId(phase1AuthTestId("shell"))
          .getAttribute("data-bootstrap-state"),
      ).toBe("mfa_required");
    });
    expect(
      screen.getByTestId(phase1AuthTestId("login-totp-code")),
    ).toBeTruthy();
    expect(screen.getByTestId(phase1ErrorCodeTestId("auth")).textContent).toBe(
      "",
    );
    expect(
      screen
        .getByTestId(phase1AuthTestId("shell"))
        .getAttribute("data-bootstrap-state"),
    ).toBe("mfa_required");
    expect(document.body.textContent ?? "").not.toContain(
      "mfa-required-token-must-not-render",
    );
    expect(document.body.textContent ?? "").not.toContain(
      "otpauth://mfa-required-private",
    );

    fireEvent.change(screen.getByTestId(phase1AuthTestId("login-totp-code")), {
      target: { value: "111111" },
    });
    fireEvent.click(screen.getByTestId(phase1AuthTestId("login-submit")));

    await waitFor(() => {
      expect(
        screen
          .getByTestId(phase1AuthTestId("shell"))
          .getAttribute("data-bootstrap-state"),
      ).toBe("mfa_setup_required");
    });
    expect(
      screen
        .getByTestId(phase1AuthTestId("shell"))
        .getAttribute("data-bootstrap-state"),
    ).toBe("mfa_setup_required");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("auth").message).textContent,
    ).toBe("Authenticator setup is required before sign-in.");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("auth").details).textContent,
    ).toBe("");
    expect(
      screen.getByTestId(phase1AuthTestId("bootstrap-token")).textContent,
    ).toBe("Stored for TOTP setup requests.");
    const preBeginText = document.body.textContent ?? "";
    expect(preBeginText).not.toContain("bootstrap-token-123");
    expect(preBeginText).not.toContain("ERRORSECRETBASE32");
    expect(preBeginText).not.toContain("otpauth://private-error");
    expect(preBeginText).not.toContain("req-private-setup");
    expect(preBeginText).not.toContain("/var/lib/cartulary");
    expect(preBeginText).not.toContain("private stack");

    fireEvent.click(screen.getByTestId(phase1AuthTestId("bootstrap-begin")));
    await waitFor(() => {
      expect(
        screen.getByTestId(phase1AuthTestId("bootstrap-secret-base32"))
          .textContent,
      ).toBe("JBSWY3DPEHPK3PXP");
    });

    fireEvent.change(
      screen.getByTestId(phase1AuthTestId("bootstrap-complete-code")),
      {
        target: { value: "123456" },
      },
    );
    fireEvent.click(screen.getByTestId(phase1AuthTestId("bootstrap-complete")));

    await waitFor(() => {
      expect(screen.getByTestId(phase1AuthTestId("status")).textContent).toBe(
        "",
      );
    });
    expect(document.body.textContent ?? "").toContain(
      "Authenticator setup is complete. Sign in again.",
    );
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
    expect(
      screen.queryByTestId(phase1LandingTestId("current-user")),
    ).toBeNull();
    await expectStableFetchCount(fetchMock, 6);
  });

  it("ordinary account-security controls issue password-change and totp-enrollment requests, surface failures on the shell, and refresh back to anonymous state after success", async () => {
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

    await openAccountSecurity();

    fireEvent.change(
      screen.getByTestId(phase1AccountTestId("totp-current-password")),
      {
        target: { value: "Current Phase 1 Password!" },
      },
    );
    fireEvent.change(
      screen.getByTestId(phase1AccountTestId("totp-current-factor")),
      {
        target: { value: "111111" },
      },
    );
    fireEvent.click(screen.getByTestId(phase1AccountTestId("totp-begin")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("account")).textContent,
      ).toBe("invalid_second_factor");
    });
    expect(screen.getByTestId(phase1AccountTestId("status")).textContent).toBe(
      "TOTP begin failed",
    );

    fireEvent.change(
      screen.getByTestId(phase1AccountTestId("totp-current-factor")),
      {
        target: { value: "222222" },
      },
    );
    fireEvent.click(screen.getByTestId(phase1AccountTestId("totp-begin")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1AccountTestId("totp-secret-base32"))
          .textContent,
      ).toBe("JBSWY3DPEHPK3PXP");
    });

    fireEvent.change(
      screen.getByTestId(phase1AccountTestId("password-current")),
      {
        target: { value: "Current Phase 1 Password!" },
      },
    );
    fireEvent.change(screen.getByTestId(phase1AccountTestId("password-next")), {
      target: { value: "Replacement Phase 1 Password!" },
    });
    fireEvent.change(
      screen.getByTestId(phase1AccountTestId("password-factor-code")),
      {
        target: { value: "654321" },
      },
    );
    fireEvent.click(screen.getByTestId(phase1AccountTestId("password-change")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1AuthTestId("login-username")),
      ).toBeTruthy();
    });
    expect(
      screen
        .getByTestId(phase1AuthTestId("shell"))
        .getAttribute("data-bootstrap-state"),
    ).toBe("revoked");
    expect(screen.getByTestId(phase1ErrorCodeTestId("auth")).textContent).toBe(
      "Sign in again to continue.",
    );
    expect(
      screen.getByTestId(phase1AuthTestId("shell-message")).textContent,
    ).toContain("Password changed. Sign in again.");

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
    await expectStableFetchCount(fetchMock, 14);
  });

  it("route-boundary account password and TOTP errors render public envelopes without private details", async () => {
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

    await openAccountSecurity();

    fireEvent.change(
      screen.getByTestId(phase1AccountTestId("totp-current-password")),
      {
        target: { value: "Current Phase 1 Password!" },
      },
    );
    fireEvent.change(
      screen.getByTestId(phase1AccountTestId("totp-current-factor")),
      {
        target: { value: "111111" },
      },
    );
    fireEvent.click(screen.getByTestId(phase1AccountTestId("totp-begin")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("account")).textContent,
      ).toBe("invalid_second_factor");
    });
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("account").message)
        .textContent,
    ).toBe("Second factor rejected.");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("account").details)
        .textContent,
    ).toContain("Reason: totp_code_invalid");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("account").details)
        .textContent,
    ).toContain("Field: second_factor.assertion.code");
    expectPrivateErrorProbeNotRendered();

    fireEvent.change(
      screen.getByTestId(phase1AccountTestId("password-current")),
      {
        target: { value: "Wrong Current Phase 1 Password!" },
      },
    );
    fireEvent.change(screen.getByTestId(phase1AccountTestId("password-next")), {
      target: { value: "Replacement Phase 1 Password!" },
    });
    fireEvent.change(
      screen.getByTestId(phase1AccountTestId("password-factor-code")),
      {
        target: { value: "654321" },
      },
    );
    fireEvent.click(screen.getByTestId(phase1AccountTestId("password-change")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("account")).textContent,
      ).toBe("invalid_current_password");
    });
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("account").message)
        .textContent,
    ).toBe("Current password rejected.");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("account").details)
        .textContent,
    ).toContain("Reason: password_mismatch");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("account").details)
        .textContent,
    ).toContain("Field: current_password");
    expectPrivateErrorProbeNotRendered();
    await expectStableFetchCount(fetchMock, 10);
  });

  it("route-boundary logout failures render public envelopes without ending the visible session", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Phase 1 Operator",
      }),
      extraRoutes: [
        {
          method: "POST",
          url: "/api/v1/auth/logout",
          handler: () =>
            publicErrorResponse("csrf_verification_failed", 403, {
              message: "Sign out request failed.",
              publicDetails: {
                reason_code: "csrf_token_missing",
              },
            }),
        },
      ],
    });

    renderApp();

    await openAccountSecurity();
    fireEvent.click(screen.getByTestId(phase1AccountTestId("logout")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("account")).textContent,
      ).toBe("csrf_verification_failed");
    });
    expect(screen.getByTestId(phase1AccountTestId("status")).textContent).toBe(
      "Sign out failed",
    );
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("account").message)
        .textContent,
    ).toBe("Sign out request failed.");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("account").details)
        .textContent,
    ).toContain("Reason: csrf_token_missing");
    expect(screen.getByTestId(phase1AccountTestId("logout")).textContent).toBe(
      "Sign out",
    );
    expect(screen.queryByTestId(phase1AuthTestId("login-username"))).toBeNull();
    expectPrivateErrorProbeNotRendered();
    await expectStableFetchCount(fetchMock, 9);
  });

  it("route-boundary bootstrap TOTP complete errors render public envelopes without private details", async () => {
    installLandingShellFetch(fetchMock, {
      session: errorResponse("session_required", 401),
      extraRoutes: [
        {
          method: "POST",
          url: "/api/v1/auth/login",
          handler: () =>
            jsonResponse(
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
            ),
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
          handler: () =>
            publicErrorResponse("invalid_second_factor", 401, {
              message: "TOTP completion failed.",
              publicDetails: {
                reason_code: "totp_code_invalid",
                field: "code",
              },
            }),
        },
      ],
    });

    renderApp();

    await screen.findByTestId(phase1AuthTestId("login-username"));
    fireEvent.change(screen.getByTestId(phase1AuthTestId("login-username")), {
      target: { value: "bootstrap@example.test" },
    });
    fireEvent.change(screen.getByTestId(phase1AuthTestId("login-password")), {
      target: { value: "BootstrapPass1!" },
    });
    fireEvent.click(screen.getByTestId(phase1AuthTestId("login-submit")));

    await waitFor(() => {
      expect(
        screen
          .getByTestId(phase1AuthTestId("shell"))
          .getAttribute("data-bootstrap-state"),
      ).toBe("mfa_setup_required");
    });
    fireEvent.click(screen.getByTestId(phase1AuthTestId("bootstrap-begin")));
    await waitFor(() => {
      expect(
        screen.getByTestId(phase1AuthTestId("bootstrap-secret-base32"))
          .textContent,
      ).toBe("JBSWY3DPEHPK3PXP");
    });

    fireEvent.change(
      screen.getByTestId(phase1AuthTestId("bootstrap-complete-code")),
      {
        target: { value: "000000" },
      },
    );
    fireEvent.click(screen.getByTestId(phase1AuthTestId("bootstrap-complete")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("auth")).textContent,
      ).toBe("The verification code is incorrect or expired.");
    });
    expect(screen.getByTestId(phase1AuthTestId("status")).textContent).toBe(
      "Authenticator setup required.",
    );
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("auth").message).textContent,
    ).toBe("The verification code is incorrect or expired.");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("auth").details).textContent,
    ).toBe("");
    expect(
      screen
        .getByTestId(phase1ErrorCodeTestId("auth"))
        .getAttribute("data-error-code"),
    ).toBe("invalid_second_factor");
    expect(
      screen.queryByTestId(phase1LandingTestId("current-user")),
    ).toBeNull();
    expectPrivateErrorProbeNotRendered();
    await expectStableFetchCount(fetchMock, 5);
  });

  it("route-boundary session TOTP complete errors render public envelopes without private details", async () => {
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
            jsonResponse({
              data: {
                enrollment_id: "replacement-enrollment-1",
                totp_setup: {
                  secret_base32: "JBSWY3DPEHPK3PXP",
                },
              },
            }),
        },
        {
          method: "POST",
          url: "/api/v1/auth/mfa/totp/complete",
          handler: () =>
            publicErrorResponse("invalid_second_factor", 401, {
              message: "Replacement TOTP completion failed.",
              publicDetails: {
                reason_code: "totp_code_invalid",
                field: "code",
              },
            }),
        },
      ],
    });

    renderApp();

    await openAccountSecurity();
    fireEvent.change(
      screen.getByTestId(phase1AccountTestId("totp-current-password")),
      {
        target: { value: "Current Phase 1 Password!" },
      },
    );
    fireEvent.change(
      screen.getByTestId(phase1AccountTestId("totp-current-factor")),
      {
        target: { value: "111111" },
      },
    );
    fireEvent.click(screen.getByTestId(phase1AccountTestId("totp-begin")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1AccountTestId("totp-enrollment-id"))
          .textContent,
      ).toBe("replacement-enrollment-1");
    });
    fireEvent.change(
      screen.getByTestId(phase1AccountTestId("totp-complete-code")),
      {
        target: { value: "000000" },
      },
    );
    fireEvent.click(screen.getByTestId(phase1AccountTestId("totp-complete")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("account")).textContent,
      ).toBe("invalid_second_factor");
    });
    expect(screen.getByTestId(phase1AccountTestId("status")).textContent).toBe(
      "TOTP complete failed",
    );
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("account").message)
        .textContent,
    ).toBe("Replacement TOTP completion failed.");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("account").details)
        .textContent,
    ).toContain("Reason: totp_code_invalid");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("account").details)
        .textContent,
    ).toContain("Field: code");
    expectPrivateErrorProbeNotRendered();
    await expectStableFetchCount(fetchMock, 10);
  });

  it("ordinary deployment-admin controls create and load users, send versioned patch requests, and surface user_version_conflict plus last_deployment_admin on the shell", async () => {
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
          method: "GET",
          url: "/api/v1/users?limit=100",
          handler: () =>
            jsonResponse({
              data: {
                users: [adminTarget],
              },
              meta: {
                paging: {
                  limit: 100,
                  has_more: false,
                  next_cursor: null,
                },
              },
            }),
        },
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

    await screen.findByTestId(phase1LandingTestId("current-user"));
    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("landing")).textContent,
      ).toBe("authorization_denied");
    });
    expect(
      screen
        .getByTestId(phase1LandingTestId("shell"))
        .getAttribute("data-bootstrap-state"),
    ).toBe("forbidden");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("landing").message)
        .textContent,
    ).toBe("Membership required.");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("landing").details)
        .textContent,
    ).toContain("Reason: incident_membership_required");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("landing").details)
        .textContent,
    ).toContain("Required role: viewer");
    const forbiddenText = document.body.textContent ?? "";
    expect(forbiddenText).not.toContain("req-private-landing");
    expect(forbiddenText).not.toContain("/var/lib/cartulary");
    expect(forbiddenText).not.toContain("select * from sessions");
    expect(forbiddenText).not.toContain("forbidden-bootstrap-token");
    await openDeploymentAdministration();
    fireEvent.click(screen.getByTestId(phase1AdminTestId("create-user")));
    fireEvent.change(screen.getByTestId(phase1AdminTestId("create-email")), {
      target: { value: createdUser.email },
    });
    fireEvent.change(
      screen.getByTestId(phase1AdminTestId("create-display-name")),
      {
        target: { value: createdUser.display_name },
      },
    );
    fireEvent.change(screen.getByTestId(phase1AdminTestId("create-password")), {
      target: { value: "CreatedPass1!" },
    });
    fireEvent.click(screen.getByTestId(phase1AdminTestId("create-user")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1AdminTestId("target-user-version"))
          .textContent,
      ).toBe("1");
    });
    expect(screen.getByTestId(phase1AdminTestId("status")).textContent).toBe(
      "Created local user",
    );
    await waitFor(() => {
      expect(
        (
          screen.getByTestId(
            phase1AdminTestId("patch-user"),
          ) as HTMLButtonElement
        ).disabled,
      ).toBe(false);
    });
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

    fireEvent.click(screen.getByTestId(phase1AdminTestId("patch-user")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("admin")).textContent,
      ).toBe("user_version_conflict");
    });
    const patchConflictRequest = requireJSONRequest(
      fetchMock,
      "/api/v1/users/user-2",
      "PATCH",
    );
    expect(patchConflictRequest.body).toEqual({
      base_user_version: 1,
      display_name: "Phase 1 Admin Target",
      email: "phase1-admin@example.test",
      mfa_required: true,
      is_active: true,
      is_deployment_admin: false,
    });

    fireEvent.click(
      await screen.findByTestId(deploymentUserRowTestId("user-1")),
    );
    await waitFor(() => {
      expect(
        screen.getByTestId(phase1AdminTestId("target-user-version"))
          .textContent,
      ).toBe("9");
    });
    await waitFor(() => {
      expect(
        (
          screen.getByTestId(
            phase1AdminTestId("patch-user"),
          ) as HTMLButtonElement
        ).disabled,
      ).toBe(false);
    });
    fireEvent.click(
      screen.getByTestId(phase1AdminTestId("patch-is-deployment-admin")),
    );
    fireEvent.click(screen.getByTestId(phase1AdminTestId("patch-user")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("admin")).textContent,
      ).toBe("last_deployment_admin");
    });
    const lastAdminPatchRequest = requireJSONRequest(
      fetchMock,
      "/api/v1/users/user-1",
      "PATCH",
    );
    expect(lastAdminPatchRequest.body).toEqual({
      base_user_version: 9,
      display_name: "Deployment Admin",
      email: "deployment-admin@example.test",
      mfa_required: true,
      is_active: true,
      is_deployment_admin: false,
    });
    expect(screen.getByTestId(phase1AdminTestId("status")).textContent).toBe(
      "Patch local user failed",
    );
    expect(findFetchCalls(fetchMock, "/api/v1/users", "POST")).toHaveLength(1);
    expect(
      findFetchCalls(fetchMock, "/api/v1/users/user-2", "PATCH"),
    ).toHaveLength(1);
    expect(
      findFetchCalls(fetchMock, "/api/v1/users/user-1", "PATCH"),
    ).toHaveLength(1);
    expect(
      findFetchCalls(fetchMock, "/api/v1/users/user-2/password/reset", "POST"),
    ).toHaveLength(0);
    expect(
      findFetchCalls(fetchMock, "/api/v1/users/user-2/mfa/totp/reset", "POST"),
    ).toHaveLength(0);
    expect(
      findFetchCalls(
        fetchMock,
        "/api/v1/users/user-2/sessions/revoke-all",
        "POST",
      ),
    ).toHaveLength(0);
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
          url: "/api/v1/users?limit=100",
          handler: () =>
            jsonResponse({
              data: {
                users: [loadedUser],
              },
              meta: {
                paging: {
                  limit: 100,
                  has_more: false,
                  next_cursor: null,
                },
              },
            }),
        },
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

    await screen.findByTestId(phase1LandingTestId("current-user"));
    await openDeploymentAdministration();
    expect(screen.queryByTestId(phase1AdminTestId("patch-user"))).toBe(null);
    expect(screen.queryByTestId(phase1AdminTestId("password-reset"))).toBe(
      null,
    );

    fireEvent.click(
      await screen.findByTestId(deploymentUserRowTestId("user-2")),
    );
    await waitFor(() => {
      expect(screen.getByTestId(phase1AdminTestId("status")).textContent).toBe(
        "Loading target user",
      );
    });

    expect(screen.queryByTestId(phase1AdminTestId("patch-user"))).toBe(null);
    expect(screen.queryByTestId(phase1AdminTestId("password-reset"))).toBe(
      null,
    );
    expect(
      findFetchCalls(fetchMock, "/api/v1/users/user-2/password/reset", "POST"),
    ).toHaveLength(0);
    expect(
      findFetchCalls(fetchMock, "/api/v1/users/user-2", "PATCH"),
    ).toHaveLength(0);

    pendingLoad.resolve(jsonResponse({ data: loadedUser }));
    await waitFor(() => {
      expect(
        screen.getByTestId(phase1AdminTestId("target-user-version"))
          .textContent,
      ).toBe("7");
    });
    expect(screen.getByTestId(phase1AdminTestId("status")).textContent).toBe(
      "Loaded target user",
    );
    expect(
      (screen.getByTestId(phase1AdminTestId("patch-user")) as HTMLButtonElement)
        .disabled,
    ).toBe(false);

    fireEvent.click(screen.getByTestId(phase1AdminTestId("patch-user")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("admin")).textContent,
      ).toBe("user_version_conflict");
    });
    expect(screen.getByTestId(phase1AdminTestId("status")).textContent).toBe(
      "Patch local user failed",
    );
    const patchRequest = requireJSONRequest(
      fetchMock,
      "/api/v1/users/user-2",
      "PATCH",
    );
    expect(patchRequest.body.base_user_version).toBe(7);
    expect(
      findFetchCalls(fetchMock, "/api/v1/users/user-2", "PATCH"),
    ).toHaveLength(1);
    expect(
      findFetchCalls(fetchMock, "/api/v1/users/user-2", "GET"),
    ).toHaveLength(1);
    expect(screen.getByTestId(phase1AdminTestId("status")).textContent).toBe(
      "Patch local user failed",
    );
  });

  it("route-boundary deployment-admin load and action errors render public envelopes without private details", async () => {
    const loadedUser = userResource({
      user_id: "user-2",
      email: "loaded-target@example.test",
      display_name: "Loaded Target",
      user_version: 7,
    });
    let loadAttempt = 0;
    let resetAttempt = 0;

    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Deployment Admin",
        is_deployment_admin: true,
      }),
      extraRoutes: [
        {
          method: "GET",
          url: "/api/v1/users?limit=100",
          handler: () =>
            jsonResponse({
              data: {
                users: [loadedUser],
              },
              meta: {
                paging: {
                  limit: 100,
                  has_more: false,
                  next_cursor: null,
                },
              },
            }),
        },
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
          handler: () => {
            loadAttempt += 1;
            if (loadAttempt === 1) {
              return publicErrorResponse("user_not_found", 404, {
                message: "Target user was not found.",
                publicDetails: {
                  reason_code: "target_not_visible",
                  field: "user_id",
                },
              });
            }
            return jsonResponse({
              data: loadedUser,
            });
          },
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
        {
          method: "POST",
          url: "/api/v1/users/user-2/sessions/revoke-all",
          handler: () =>
            publicErrorResponse("authorization_denied", 403, {
              message: "Revoke-all request is denied.",
              publicDetails: {
                reason_code: "deployment_admin_required",
                required_role: "deployment_admin",
              },
            }),
        },
      ],
    });

    renderApp();

    await screen.findByTestId(phase1LandingTestId("current-user"));
    await openDeploymentAdministration();
    fireEvent.click(screen.getByTestId(phase1AdminTestId("create-user")));
    fireEvent.change(screen.getByTestId(phase1AdminTestId("create-email")), {
      target: { value: "invalid-admin-target@example.test" },
    });
    fireEvent.change(
      screen.getByTestId(phase1AdminTestId("create-display-name")),
      {
        target: { value: "Invalid Admin Target" },
      },
    );
    fireEvent.change(screen.getByTestId(phase1AdminTestId("create-password")), {
      target: { value: "AdminCreatePass1!" },
    });
    fireEvent.click(screen.getByTestId(phase1AdminTestId("create-user")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("admin")).textContent,
      ).toBe("invalid_user_create");
    });
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").message)
        .textContent,
    ).toBe("User create request is invalid.");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").details)
        .textContent,
    ).toContain("Reason: email_not_allowed");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").details)
        .textContent,
    ).toContain("Field: email");
    expectPrivateErrorProbeNotRendered();
    fireEvent.click(screen.getByLabelText("Close create user"));

    fireEvent.click(
      await screen.findByTestId(deploymentUserRowTestId("user-2")),
    );

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("admin")).textContent,
      ).toBe("user_not_found");
    });
    expect(screen.getByTestId(phase1AdminTestId("status")).textContent).toBe(
      "Load target user failed",
    );
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").message)
        .textContent,
    ).toBe("Target user was not found.");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").details)
        .textContent,
    ).toContain("Reason: target_not_visible");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").details)
        .textContent,
    ).toContain("Field: user_id");
    expect(screen.queryByTestId(phase1AdminTestId("target-user-id"))).toBe(
      null,
    );
    expect(screen.queryByTestId(phase1AdminTestId("target-user-version"))).toBe(
      null,
    );
    expect(screen.queryByTestId(phase1AdminTestId("patch-user"))).toBe(null);
    expectPrivateErrorProbeNotRendered();

    fireEvent.click(screen.getByTestId(deploymentUserRowTestId("user-2")));
    await waitFor(() => {
      expect(
        screen.getByTestId(phase1AdminTestId("target-user-version"))
          .textContent,
      ).toBe("7");
    });
    fireEvent.click(screen.getByTestId(phase1AdminTestId("patch-user")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("admin")).textContent,
      ).toBe("invalid_user_patch");
    });
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").message)
        .textContent,
    ).toBe("User patch request is invalid.");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").details)
        .textContent,
    ).toContain("Reason: display_name_invalid");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").details)
        .textContent,
    ).toContain("Field: display_name");
    expectPrivateErrorProbeNotRendered();

    fireEvent.click(screen.getByTestId(phase1AdminTestId("password-reset")));
    fireEvent.change(screen.getByTestId(phase1AdminTestId("new-password")), {
      target: { value: "AdminResetPass1!" },
    });
    fireEvent.change(screen.getByTestId(phase1AdminTestId("reason")), {
      target: { value: "row-owned public error check" },
    });
    fireEvent.click(screen.getByTestId(phase1AdminTestId("password-reset")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("admin")).textContent,
      ).toBe("user_version_conflict");
    });
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").message)
        .textContent,
    ).toBe("User version conflict.");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").details)
        .textContent,
    ).toContain("Reason: stale_user_version");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").details)
        .textContent,
    ).toContain("Field: base_user_version");
    expectPrivateErrorProbeNotRendered();

    fireEvent.click(screen.getByLabelText("Close credential action"));
    fireEvent.click(screen.getByTestId(phase1AdminTestId("totp-reset")));
    fireEvent.change(screen.getByTestId(phase1AdminTestId("reason")), {
      target: { value: "row-owned public error check" },
    });
    fireEvent.click(screen.getByTestId(phase1AdminTestId("totp-reset")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("admin")).textContent,
      ).toBe("invalid_mutation_payload");
    });
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").message)
        .textContent,
    ).toBe("TOTP reset request is invalid.");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").details)
        .textContent,
    ).toContain("Reason: totp_reset_invalid_1");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").details)
        .textContent,
    ).toContain("Field: reason");
    expectPrivateErrorProbeNotRendered();

    fireEvent.click(screen.getByLabelText("Close credential action"));
    fireEvent.click(screen.getByTestId(phase1AdminTestId("revoke-all")));
    fireEvent.change(screen.getByTestId(phase1AdminTestId("reason")), {
      target: { value: "row-owned public error check" },
    });
    fireEvent.click(screen.getByTestId(phase1AdminTestId("revoke-all")));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("admin")).textContent,
      ).toBe("authorization_denied");
    });
    expect(screen.getByTestId(phase1AdminTestId("status")).textContent).toBe(
      "Revoke-all failed",
    );
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").message)
        .textContent,
    ).toBe("Revoke-all request is denied.");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").details)
        .textContent,
    ).toContain("Reason: deployment_admin_required");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("admin").details)
        .textContent,
    ).toContain("Required role: deployment_admin");
    expectPrivateErrorProbeNotRendered();
    expect(findFetchCalls(fetchMock, "/api/v1/users", "POST")).toHaveLength(1);
    expect(
      findFetchCalls(fetchMock, "/api/v1/users/user-2", "GET"),
    ).toHaveLength(2);
    expect(
      findFetchCalls(fetchMock, "/api/v1/users/user-2", "PATCH"),
    ).toHaveLength(1);
    expect(
      findFetchCalls(fetchMock, "/api/v1/users/user-2/password/reset", "POST"),
    ).toHaveLength(1);
    expect(
      findFetchCalls(fetchMock, "/api/v1/users/user-2/mfa/totp/reset", "POST"),
    ).toHaveLength(1);
    expect(
      findFetchCalls(
        fetchMock,
        "/api/v1/users/user-2/sessions/revoke-all",
        "POST",
      ),
    ).toHaveLength(1);
    for (const call of fetchMock.mock.calls) {
      expect(String(call[0]).startsWith("/api/v1/")).toBe(true);
    }
  });

  it("route-boundary incident landing uses /api/v1/incidents with closed create bodies and public errors", async () => {
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

    await screen.findByTestId(phase1LandingTestId("current-user"));
    expect(
      findFetchCalls(fetchMock, "/api/v1/auth/session", "GET"),
    ).toHaveLength(1);
    expect(
      findFetchCalls(fetchMock, "/api/v1/auth/credential-state", "GET"),
    ).toHaveLength(1);
    const listIncidentRequests = fetchMock.mock.calls
      .filter(([input, init]) => {
        const method = ((init as RequestInit | undefined)?.method ?? "GET")
          .toString()
          .toUpperCase();
        return (
          method === "GET" && String(input).startsWith("/api/v1/incidents?")
        );
      })
      .map(([, init]) => init as RequestInit | undefined);
    expect(listIncidentRequests).toHaveLength(2);
    for (const listIncidentRequest of listIncidentRequests) {
      expect(listIncidentRequest?.credentials).toBe("include");
      expect(readHeader(listIncidentRequest, "Authorization")).toBe("");
      expect(readHeader(listIncidentRequest, csrfHeaderName)).toBe("");
    }

    fireEvent.click(
      screen.getByTestId(phase1LandingTestId("create-open-button")),
    );
    fireEvent.change(screen.getByTestId(phase1LandingTestId("incident-key")), {
      target: { value: "IR-2026-003" },
    });
    fireEvent.change(
      screen.getByTestId(phase1LandingTestId("incident-title")),
      {
        target: { value: "Phase 1 boundary incident" },
      },
    );
    fireEvent.click(
      screen.getByTestId(phase1LandingTestId("create-submit-button")),
    );

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1ErrorCodeTestId("landing")).textContent,
      ).toBe("invalid_incident_create");
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

    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("landing").message)
        .textContent,
    ).toBe("Incident create request is invalid.");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("landing").details)
        .textContent,
    ).toContain("Reason: unknown_field");
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("landing").details)
        .textContent,
    ).toContain("Field: debug_field");
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
    await expectStableFetchCount(fetchMock, 7);
  });
});

function renderApp() {
  return render(<AppRoot />);
}

async function openAccountSecurity() {
  fireEvent.click(
    await screen.findByLabelText("Account and application navigation"),
  );
  fireEvent.click(screen.getByRole("menuitem", { name: "Account settings" }));
  fireEvent.click(screen.getByRole("tab", { name: "Security" }));
  await screen.findByTestId(phase1AccountTestId("password-current"));
}

async function openDeploymentAdministration() {
  fireEvent.click(
    await screen.findByLabelText("Account and application navigation"),
  );
  fireEvent.click(
    screen.getByRole("menuitem", { name: "Deployment administration" }),
  );
  await screen.findByTestId(landingAdminMenuItemTestId("deployment-users"));
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
