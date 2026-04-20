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

import { App } from "./App";

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

  it("shows the ordinary login surface when the browser has no session", async () => {
    fetchMock.mockImplementation((input) => {
      if (String(input) === "/api/v1/auth/session") {
        return Promise.resolve(
          errorResponse("session_required", 401, {
            reason_code: "no_session",
          }),
        );
      }
      throw new Error(`unexpected fetch: ${String(input)}`);
    });

    render(<App />);

    expect(await screen.findByTestId("auth-login-username")).toBeTruthy();
    expect(screen.getByTestId("auth-shell-message").textContent).toContain(
      "Sign in with your local account",
    );
  });

  it("signs in through the ordinary shell and renders account plus denied admin panels", async () => {
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

    render(<App />);

    await screen.findByTestId("auth-login-username");
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
  });

  it("shows the bootstrap enrollment flow on the ordinary login shell", async () => {
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

    render(<App />);

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
  });

  it("renders the ordinary deployment-admin user management surface and shows patch conflicts", async () => {
    const createdUser = userResource({
      user_id: "user-2",
      email: "phase1-admin@example.test",
      display_name: "Phase 1 Admin Target",
      user_version: 1,
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
      throw new Error(`unexpected fetch: ${method} ${url}`);
    });

    render(<App />);

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

    fireEvent.change(screen.getByTestId("admin-patch-base-version"), {
      target: { value: "0" },
    });
    fireEvent.click(screen.getByTestId("admin-patch-user"));

    await waitFor(() => {
      expect(screen.getByTestId("admin-error-code").textContent).toBe(
        "user_version_conflict",
      );
    });
  });
});

function sessionResource(
  overrides?: Partial<{
    absolute_expires_at: string;
    authenticated_at: string;
    display_name: string;
    idle_expires_at: string;
    is_deployment_admin: boolean;
    memberships: Array<{ incident_id: string; role: string }>;
    mfa_state: string;
    provider_type: string;
    session_expires_at: string;
    user_id: string;
  }>,
) {
  return {
    user_id: "user-1",
    display_name: "Operator",
    provider_type: "local",
    mfa_state: "single_factor",
    is_deployment_admin: false,
    authenticated_at: "2026-04-20T12:00:00Z",
    idle_expires_at: "2026-04-20T12:30:00Z",
    absolute_expires_at: "2026-04-20T20:00:00Z",
    session_expires_at: "2026-04-20T12:30:00Z",
    memberships: [],
    ...overrides,
  };
}

function credentialStateResource() {
  return {
    user_id: "user-1",
    auth_kind: "local",
    recovery_model: "deployment_admin_reset",
    password_changed_at: "2026-04-20T12:00:00Z",
    totp: {
      state: "not_enrolled",
      enrolled_at: null,
      pending_expires_at: null,
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
