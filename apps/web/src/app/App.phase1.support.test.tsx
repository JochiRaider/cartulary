import {
  deploymentUserRowTestId,
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

import {
  credentialStateResource,
  installLandingShellFetch,
  sessionResource,
} from "../testing/appShellTestSupport";
import { AppRoot } from "./AppRoot";
import { Phase1AccountPanel, Phase1AdminPanel } from "./Phase1Surface";
import type { UserResource } from "./phase1Client";

describe("Phase 1 ordinary shell support", () => {
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

  it("shows the anonymous login surface when the browser has no session", async () => {
    fetchMock.mockImplementation((input) => {
      if (String(input) === "/api/v1/auth/session") {
        return Promise.resolve(
          jsonResponse(
            {
              error: {
                code: "session_required",
                status: 401,
                details: {
                  reason_code: "no_session",
                },
              },
            },
            401,
          ),
        );
      }
      if (String(input) === "/api/v1/auth/providers") {
        return Promise.resolve(jsonResponse({ data: { providers: [] } }));
      }
      throw new Error(`unexpected fetch: ${String(input)}`);
    });

    render(<AppRoot />);

    expect(
      await screen.findByTestId(phase1AuthTestId("login-username")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(phase1AuthTestId("shell-message")).textContent,
    ).toContain("Sign in with your local account");
    await waitFor(() => {
      expect(
        screen.getByTestId(phase1AuthTestId("status")).textContent?.trim(),
      ).not.toBe("");
    });
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("FE-S-P1-01 Verify bootstrap route selectors and error-state selectors use stable test-id builders.", async () => {
    installLandingShellFetch(fetchMock, {
      credentialState: credentialStateResource(),
      incidents: [],
      session: sessionResource({
        display_name: "Phase 1 Support Operator",
      }),
    });

    render(<AppRoot />);

    expect(
      await screen.findByTestId(phase1RouteTestId("app-shell")),
    ).toBeTruthy();
    expect(screen.getByTestId(phase1LandingTestId("shell"))).toBeTruthy();
    expect(screen.getByTestId(phase1LandingTestId("status"))).toBeTruthy();
    await openAccountSecurity();
    expect(
      screen.getByTestId(phase1AccountTestId("logout")).textContent,
    ).not.toBe("");
    expect(screen.queryByText("Incident memberships")).toBe(null);
    expect(screen.queryByTestId(phase1AdminTestId("access-note"))).toBe(null);
    expect(screen.getByTestId(phase1ErrorCodeTestId("landing"))).toBeTruthy();
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("landing").container),
    ).toBeTruthy();
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("account").message),
    ).toBeTruthy();

    cleanup();
    fetchMock.mockReset();
    installAnonymousSessionRequiredFetch(fetchMock);
    window.history.replaceState({}, "", "/");
    render(<AppRoot />);

    expect(
      await screen.findByTestId(phase1AuthTestId("login-username")),
    ).toBeTruthy();
    expect(screen.getByTestId(phase1AuthTestId("shell-message"))).toBeTruthy();
    expect(screen.getByTestId(phase1AuthTestId("status"))).toBeTruthy();
    expect(screen.getByTestId(phase1ErrorCodeTestId("auth"))).toBeTruthy();
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("auth").container),
    ).toBeTruthy();
  });

  it("loads, filters, selects, and pages deployment users", async () => {
    const usersPageOne = [
      userResource({
        user_id: "user-alpha",
        email: "alpha@example.test",
        display_name: "Alpha Admin",
        is_deployment_admin: true,
      }),
      userResource({
        user_id: "user-bravo",
        email: "bravo@example.test",
        display_name: "Bravo Analyst",
      }),
    ];
    const bravoUser = usersPageOne[1] ?? userResource({});
    const usersPageTwo = [
      userResource({
        user_id: "user-charlie",
        email: "charlie@example.test",
        display_name: "Charlie Reviewer",
      }),
    ];
    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (url === "/api/v1/users?limit=100" && method === "GET") {
        return Promise.resolve(
          userListResponse(usersPageOne, "cursor-2", true),
        );
      }
      if (url === "/api/v1/users?limit=100&search=bravo" && method === "GET") {
        return Promise.resolve(userListResponse([bravoUser], null, false));
      }
      if (url === "/api/v1/users/user-bravo" && method === "GET") {
        return Promise.resolve(jsonResponse({ data: bravoUser }));
      }
      if (
        url === "/api/v1/users?limit=100&cursor_token=cursor-2" &&
        method === "GET"
      ) {
        return Promise.resolve(userListResponse(usersPageTwo, null, false));
      }
      throw new Error(`unexpected fetch: ${method} ${url}`);
    });

    render(
      <Phase1AdminPanel
        autoLoadUsers
        onRefreshShell={() => undefined}
        session={sessionResource({ is_deployment_admin: true })}
      />,
    );

    expect(
      await screen.findByTestId(deploymentUserRowTestId("user-alpha")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(phase1AdminTestId("status")).textContent?.trim(),
    ).not.toBe("");
    const userFilter = screen.getByTestId(phase1AdminTestId("user-filter"));
    fireEvent.change(userFilter, {
      target: { value: "bravo" },
    });
    fireEvent.keyDown(userFilter, { key: "Enter" });
    await waitFor(() => {
      expect(
        screen.queryByTestId(deploymentUserRowTestId("user-alpha")),
      ).toBeNull();
    });
    const bravoRow = await screen.findByTestId(
      deploymentUserRowTestId("user-bravo"),
    );
    expect(bravoRow.textContent).toContain("Bravo Analyst");
    fireEvent.click(bravoRow);
    await waitFor(() => {
      expect(
        screen.getByTestId(phase1AdminTestId("target-user-id")).textContent,
      ).toBe("user-bravo");
    });

    fireEvent.change(userFilter, {
      target: { value: "" },
    });
    fireEvent.keyDown(userFilter, { key: "Enter" });
    await screen.findByTestId(deploymentUserRowTestId("user-alpha"));
    fireEvent.click(screen.getByTestId(phase1AdminTestId("load-more-users")));
    expect(
      await screen.findByTestId(deploymentUserRowTestId("user-charlie")),
    ).toBeTruthy();
    expect(fetchMock).toHaveBeenCalledTimes(5);
  });

  it("keeps account and deployment-admin panels reachable through stable selectors", () => {
    const session = sessionResource({ is_deployment_admin: true });

    render(
      <>
        <Phase1AccountPanel
          credentialStateError={null}
          onRefreshShell={() => undefined}
        />
        <Phase1AdminPanel onRefreshShell={() => undefined} session={session} />
      </>,
    );

    for (const testId of [
      phase1AccountTestId("refresh-state"),
      phase1AccountTestId("logout"),
      phase1AccountTestId("password-current"),
      phase1AccountTestId("password-next"),
      phase1AccountTestId("password-change"),
      phase1AccountTestId("totp-begin"),
      phase1AccountTestId("status"),
      phase1AdminTestId("create-user"),
      phase1AdminTestId("user-filter"),
      phase1AdminTestId("user-list"),
      phase1AdminTestId("status"),
    ]) {
      expect(screen.getByTestId(testId)).toBeTruthy();
    }
    expect(screen.queryByTestId(phase1AdminTestId("patch-email"))).toBe(null);
    expect(screen.queryByTestId(phase1AdminTestId("reason"))).toBe(null);

    fireEvent.click(screen.getByTestId(phase1AdminTestId("create-user")));
    for (const testId of [
      phase1AdminTestId("create-email"),
      phase1AdminTestId("create-display-name"),
      phase1AdminTestId("create-user"),
    ]) {
      expect(screen.getByTestId(testId)).toBeTruthy();
    }
  });

  it("renders claimed enterprise binding controls only inside the selected-user inspector and calls binding routes", async () => {
    const loadedUser = userResource({
      user_id: "user-enterprise",
      email: "enterprise@example.test",
      display_name: "Enterprise User",
      user_version: 5,
      auth_bindings: [
        {
          provider_type: "local",
          provider_key: "local",
          username: "enterprise@example.test",
          created_at: "2026-06-18T12:00:00Z",
        },
        {
          auth_binding_id: "binding-1",
          provider_type: "oidc",
          provider_key: "corp-oidc",
          provider_subject: "subject-1",
          created_at: "2026-06-18T13:00:00Z",
          last_auth_at: null,
        },
      ],
    });
    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (url === "/api/v1/users?limit=100" && method === "GET") {
        return Promise.resolve(userListResponse([loadedUser], null, false));
      }
      if (url === "/api/v1/auth/providers" && method === "GET") {
        return Promise.resolve(
          jsonResponse({
            data: {
              providers: [
                {
                  provider_key: "corp-oidc",
                  provider_type: "oidc",
                  display_name: "Corporate OIDC",
                },
              ],
            },
          }),
        );
      }
      if (url === "/api/v1/users/user-enterprise" && method === "GET") {
        return Promise.resolve(jsonResponse({ data: loadedUser }));
      }
      if (
        url === "/api/v1/users/user-enterprise/auth-bindings" &&
        method === "POST"
      ) {
        return Promise.resolve(
          jsonResponse({ data: { ...loadedUser, user_version: 6 } }, 201),
        );
      }
      if (
        url ===
          "/api/v1/users/user-enterprise/auth-bindings/binding-1/rotate" &&
        method === "POST"
      ) {
        return Promise.resolve(
          jsonResponse({ data: { ...loadedUser, user_version: 7 } }),
        );
      }
      if (
        url === "/api/v1/users/user-enterprise/auth-bindings/binding-1" &&
        method === "DELETE"
      ) {
        return Promise.resolve(
          jsonResponse({
            data: {
              ...loadedUser,
              user_version: 8,
              auth_bindings: loadedUser.auth_bindings?.slice(0, 1),
            },
          }),
        );
      }
      throw new Error(`unexpected fetch: ${method} ${url}`);
    });

    render(
      <Phase1AdminPanel
        autoLoadUsers
        enterpriseAuthClaimed
        onRefreshShell={() => undefined}
        session={sessionResource({ is_deployment_admin: true })}
      />,
    );

    const row = await screen.findByTestId(
      deploymentUserRowTestId("user-enterprise"),
    );
    fireEvent.click(row);
    await screen.findByText("Subject: subject-1");
    expect(screen.queryByText(/provider configuration/i)).toBeNull();
    expect(screen.queryByText(/client secret/i)).toBeNull();

    fireEvent.change(screen.getByLabelText("Provider key"), {
      target: { value: "corp-oidc" },
    });
    fireEvent.change(screen.getByLabelText("Provider subject"), {
      target: { value: "subject-2" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Create binding" }));
    await waitFor(() => {
      expect(screen.getByTestId(phase1AdminTestId("status")).textContent).toBe(
        "Created enterprise auth binding",
      );
    });

    fireEvent.change(screen.getByLabelText("New provider subject"), {
      target: { value: "subject-rotated" },
    });
    const bindingReasonInput = screen.getAllByLabelText("Reason").at(-1);
    if (typeof bindingReasonInput === "undefined") {
      throw new Error("missing enterprise binding reason input");
    }
    fireEvent.change(bindingReasonInput, {
      target: { value: "identity correction" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Rotate subject" }));
    await waitFor(() => {
      expect(screen.getByTestId(phase1AdminTestId("status")).textContent).toBe(
        "Rotated enterprise auth binding",
      );
    });

    fireEvent.click(screen.getByRole("button", { name: "Retire binding" }));
    await waitFor(() => {
      expect(screen.getByTestId(phase1AdminTestId("status")).textContent).toBe(
        "Retired enterprise auth binding",
      );
    });

    expect(
      jsonRequestBody(
        fetchMock,
        "/api/v1/users/user-enterprise/auth-bindings",
        "POST",
      ),
    ).toMatchObject({
      base_user_version: 5,
      provider_key: "corp-oidc",
      provider_subject: "subject-2",
      reason: "",
    });
    expect(
      jsonRequestBody(
        fetchMock,
        "/api/v1/users/user-enterprise/auth-bindings/binding-1/rotate",
        "POST",
      ),
    ).toMatchObject({
      base_user_version: 6,
      new_provider_subject: "subject-rotated",
      reason: "identity correction",
    });
    expect(
      jsonRequestBody(
        fetchMock,
        "/api/v1/users/user-enterprise/auth-bindings/binding-1",
        "DELETE",
      ),
    ).toMatchObject({
      base_user_version: 7,
      reason: "identity correction",
    });
  });
});

async function openAccountSecurity() {
  fireEvent.click(
    await screen.findByLabelText("Account and application navigation"),
  );
  fireEvent.click(screen.getByRole("menuitem", { name: "Account settings" }));
  fireEvent.click(screen.getByRole("tab", { name: "Security" }));
  await screen.findByTestId(phase1AccountTestId("password-current"));
}

function installAnonymousSessionRequiredFetch(
  fetchMock: ReturnType<typeof vi.fn>,
) {
  fetchMock.mockImplementation((input) => {
    if (String(input) === "/api/v1/auth/session") {
      return Promise.resolve(
        jsonResponse(
          {
            error: {
              code: "session_required",
              status: 401,
              details: {
                reason_code: "no_session",
              },
            },
          },
          401,
        ),
      );
    }
    throw new Error(`unexpected fetch: ${String(input)}`);
  });
}

function jsonRequestBody(
  fetchMock: ReturnType<typeof vi.fn>,
  url: string,
  method: string,
) {
  const call = fetchMock.mock.calls.find(([input, init]) => {
    const requestMethod = (init?.method ?? "GET").toUpperCase();
    return String(input) === url && requestMethod === method;
  });
  if (typeof call === "undefined") {
    throw new Error(`missing request: ${method} ${url}`);
  }
  const body = call[1]?.body;
  if (typeof body !== "string") {
    throw new Error(`missing JSON body: ${method} ${url}`);
  }
  return JSON.parse(body) as Record<string, unknown>;
}

function userResource(overrides: Partial<UserResource>): UserResource {
  return {
    user_id: "user-default",
    email: "default@example.test",
    display_name: "Default User",
    user_version: 1,
    is_active: true,
    mfa_required: true,
    is_deployment_admin: false,
    ...overrides,
  };
}

function userListResponse(
  users: UserResource[],
  nextCursor: string | null,
  hasMore: boolean,
) {
  return jsonResponse({
    data: {
      users,
    },
    meta: {
      paging: {
        limit: 100,
        has_more: hasMore,
        next_cursor: nextCursor,
      },
      request_id: "req-users",
    },
  });
}

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
