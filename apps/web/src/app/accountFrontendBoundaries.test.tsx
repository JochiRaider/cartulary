import {
  accountTestId,
  authTestId,
  deploymentAdminTestId,
  deploymentUserRowTestId,
} from "@cartulary/ui-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  credentialStateResource,
  sessionResource,
} from "../testing/appShellTestSupport";
import { deferred, jsonResponse } from "../testing/fetchMockTestSupport";
import {
  AccountSecurityPanel,
  DeploymentUsersPanel,
} from "./AccountAdministrationPanels";
import { AuthGateway } from "./AuthGateway";
import { AccountSecurityController } from "./accountSecurityModel";
import type { UserResource } from "./api/publicHttpTypes";
import { AuthenticationController } from "./authenticationModel";
import { DeploymentUsersController } from "./deploymentUsersModel";

const user = (name = "Alpha", version = 1): UserResource => ({
  user_id: "00000000-0000-4000-8000-000000000100",
  email: "alpha@example.test",
  display_name: name,
  user_version: version,
  is_active: true,
  is_deployment_admin: false,
  mfa_required: true,
  created_at: "2026-09-06T12:00:00Z",
  updated_at: "2026-09-06T12:00:00Z",
  updated_by_user_id: null,
  last_login_at: null,
  auth_bindings: [],
});
const response = (data: unknown) =>
  jsonResponse({ data, meta: { request_id: "account-boundary" } });
const page = (users: UserResource[], next: string | null = null) =>
  jsonResponse({
    data: { users },
    meta: {
      request_id: "account-boundary",
      paging: { limit: 100, next_cursor: next, has_more: next !== null },
    },
  });
let fetchMock: ReturnType<typeof vi.fn>;
beforeEach(() => {
  fetchMock = vi.fn();
  vi.stubGlobal("fetch", fetchMock);
  window.history.replaceState({}, "", "/");
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});
const input = (id: string, value: string) =>
  fireEvent.change(screen.getByTestId(id), { target: { value } });
const writes = (path: string) =>
  fetchMock.mock.calls.filter(
    ([url, init]) => String(url) === path && init?.method === "POST",
  );

describe("account frontend boundaries", () => {
  it("admits one same-turn login and accepts owner-valid deployment email", async () => {
    const pending = deferred<Response>();
    fetchMock.mockImplementation((url) =>
      String(url).endsWith("/providers")
        ? Promise.resolve(response({ providers: [] }))
        : pending.promise,
    );
    render(
      <AuthGateway
        bootstrapState="anonymous"
        message="Sign in"
        controller={
          new AuthenticationController(() => ({
            actor: sessionResource().user_id,
            current: () => true,
            authenticated: () => {},
            uncertain: async () => false,
          }))
        }
      />,
    );
    input(authTestId("login-username"), "operator@deployment");
    input(authTestId("login-password"), "Exact Password!");
    const form = screen
      .getByTestId(authTestId("login-username"))
      .closest("form");
    expect(form).not.toBeNull();
    act(() => {
      if (form) {
        fireEvent.submit(form);
        fireEvent.submit(form);
      }
    });
    await waitFor(() => expect(writes("/api/v1/auth/login")).toHaveLength(1));
  });

  it("suppresses duplicate login with ordinary email", async () => {
    const pending = deferred<Response>();
    fetchMock.mockImplementation((url) =>
      String(url).endsWith("/providers")
        ? Promise.resolve(response({ providers: [] }))
        : pending.promise,
    );
    render(
      <AuthGateway
        bootstrapState="anonymous"
        message="Sign in"
        controller={
          new AuthenticationController(() => ({
            actor: sessionResource().user_id,
            current: () => true,
            authenticated: () => {},
            uncertain: async () => false,
          }))
        }
      />,
    );
    input(authTestId("login-username"), "operator@example.test");
    input(authTestId("login-password"), "Exact Password!");
    const form = screen
      .getByTestId(authTestId("login-username"))
      .closest("form");
    expect(form).not.toBeNull();
    act(() => {
      if (form) {
        fireEvent.submit(form);
        fireEvent.submit(form);
      }
    });
    await waitFor(() => expect(writes("/api/v1/auth/login")).toHaveLength(1));
  });

  it("clears dispatched password inputs and excludes retired Security completion", async () => {
    const pending = deferred<Response>();
    const sessionEvent = vi.fn();
    fetchMock.mockImplementation((url) =>
      String(url).endsWith("/credential-state")
        ? Promise.resolve(response(credentialStateResource()))
        : pending.promise,
    );
    const view = render(
      <AccountSecurityPanel
        controller={
          new AccountSecurityController(() => ({
            actor: sessionResource().user_id,
            current: () => true,
            event: sessionEvent,
          }))
        }
      />,
    );
    input(accountTestId("password-current"), "Current Password!");
    input(accountTestId("password-next"), "Replacement Password!");
    act(() => {
      fireEvent.click(screen.getByTestId(accountTestId("password-change")));
      fireEvent.click(screen.getByTestId(accountTestId("password-change")));
    });
    expect(writes("/api/v1/auth/password/change")).toHaveLength(1);
    expect(
      (screen.getByTestId(accountTestId("password-next")) as HTMLInputElement)
        .value,
    ).toBe("");
    view.unmount();
    await act(async () =>
      pending.resolve(
        response({
          user_id: sessionResource().user_id,
          password: { changed_at: "2026-09-06T12:00:00Z" },
          sessions_revoked: true,
        }),
      ),
    );
    expect(sessionEvent).not.toHaveBeenCalled();
  });

  it("excludes Security completion after its form retires", async () => {
    const pending = deferred<Response>();
    const sessionEvent = vi.fn();
    fetchMock.mockImplementation((url) =>
      String(url).endsWith("/credential-state")
        ? Promise.resolve(response(credentialStateResource()))
        : pending.promise,
    );
    const view = render(
      <AccountSecurityPanel
        controller={
          new AccountSecurityController(() => ({
            actor: sessionResource().user_id,
            current: () => true,
            event: sessionEvent,
          }))
        }
      />,
    );
    input(accountTestId("password-current"), "Current Password!");
    input(accountTestId("password-next"), "Replacement Password!");
    act(() => {
      fireEvent.click(screen.getByTestId(accountTestId("password-change")));
    });
    expect(writes("/api/v1/auth/password/change")).toHaveLength(1);
    view.unmount();
    await act(async () =>
      pending.resolve(
        response({
          user_id: sessionResource().user_id,
          password: { changed_at: "2026-09-06T12:00:00Z" },
          sessions_revoked: true,
        }),
      ),
    );
    expect(sessionEvent).not.toHaveBeenCalled();
  });

  it("does not append a page from a replaced user query", async () => {
    const pending = deferred<Response>();
    fetchMock.mockImplementation((url) =>
      String(url).includes("cursor_token")
        ? pending.promise
        : Promise.resolve(
            String(url).includes("search=")
              ? page([])
              : page([user()], "page-two"),
          ),
    );
    render(
      <DeploymentUsersPanel
        autoLoadUsers
        controller={deploymentController(
          sessionResource({ is_deployment_admin: true }),
        )}
      />,
    );
    await screen.findByTestId(deploymentUserRowTestId(user().user_id));
    fireEvent.click(
      screen.getByTestId(deploymentAdminTestId("load-more-users")),
    );
    input(deploymentAdminTestId("user-filter"), "missing");
    fireEvent.keyDown(
      screen.getByTestId(deploymentAdminTestId("user-filter")),
      { key: "Enter" },
    );
    await waitFor(() =>
      expect(
        screen.queryByTestId(deploymentUserRowTestId(user().user_id)),
      ).toBeNull(),
    );
    await act(async () => pending.resolve(page([user()])));
    expect(
      screen.queryByTestId(deploymentUserRowTestId(user().user_id)),
    ).toBeNull();
  });

  it("sends a sparse user patch and preserves edits made while saving", async () => {
    const pending = deferred<Response>();
    fetchMock.mockImplementation((url, init) =>
      init?.method === "PATCH"
        ? pending.promise
        : Promise.resolve(
            String(url).includes("limit=") ? page([user()]) : response(user()),
          ),
    );
    render(
      <DeploymentUsersPanel
        autoLoadUsers
        controller={deploymentController(
          sessionResource({ is_deployment_admin: true }),
        )}
      />,
    );
    fireEvent.click(
      await screen.findByTestId(deploymentUserRowTestId(user().user_id)),
    );
    await screen.findByTestId(deploymentAdminTestId("patch-display-name"));
    input(deploymentAdminTestId("patch-display-name"), "Saved edit");
    fireEvent.click(screen.getByTestId(deploymentAdminTestId("patch-user")));
    const request = fetchMock.mock.calls.find(
      ([, init]) => init?.method === "PATCH",
    );
    expect(JSON.parse(String(request?.[1].body))).toEqual({
      base_user_version: 1,
      display_name: "Saved edit",
    });
    input(deploymentAdminTestId("patch-display-name"), "Newer edit");
    await act(async () => pending.resolve(response(user("Saved edit", 2))));
    expect(
      (
        screen.getByTestId(
          deploymentAdminTestId("patch-display-name"),
        ) as HTMLInputElement
      ).value,
    ).toBe("Newer edit");
  });

  it("retains newer user input across save completion", async () => {
    const pending = deferred<Response>();
    fetchMock.mockImplementation((url, init) =>
      init?.method === "PATCH"
        ? pending.promise
        : Promise.resolve(
            String(url).includes("limit=") ? page([user()]) : response(user()),
          ),
    );
    render(
      <DeploymentUsersPanel
        autoLoadUsers
        controller={deploymentController(
          sessionResource({ is_deployment_admin: true }),
        )}
      />,
    );
    fireEvent.click(
      await screen.findByTestId(deploymentUserRowTestId(user().user_id)),
    );
    await screen.findByTestId(deploymentAdminTestId("patch-display-name"));
    input(deploymentAdminTestId("patch-display-name"), "Saved edit");
    fireEvent.click(screen.getByTestId(deploymentAdminTestId("patch-user")));
    input(deploymentAdminTestId("patch-display-name"), "Newer edit");
    await act(async () => pending.resolve(response(user("Saved edit", 2))));
    expect(
      (
        screen.getByTestId(
          deploymentAdminTestId("patch-display-name"),
        ) as HTMLInputElement
      ).value,
    ).toBe("Newer edit");
  });

  it("clears a create password when its form closes", async () => {
    render(
      <DeploymentUsersPanel
        controller={deploymentController(
          sessionResource({ is_deployment_admin: true }),
        )}
      />,
    );
    fireEvent.click(screen.getByTestId(deploymentAdminTestId("create-user")));
    input(deploymentAdminTestId("create-password"), "Private Password!");
    fireEvent.click(screen.getByRole("button", { name: "Close create user" }));
    fireEvent.click(screen.getByTestId(deploymentAdminTestId("create-user")));
    expect(
      (
        screen.getByTestId(
          deploymentAdminTestId("create-password"),
        ) as HTMLInputElement
      ).value,
    ).toBe("");
  });

  it("keeps confirmed user mutation separate from a failed session refresh", async () => {
    fetchMock.mockImplementation((url, init) =>
      Promise.resolve(
        init?.method === "PATCH"
          ? response(user("Saved edit", 2))
          : String(url).includes("limit=")
            ? page([user()])
            : response(user()),
      ),
    );
    render(
      <DeploymentUsersPanel
        autoLoadUsers
        controller={deploymentController(
          sessionResource({ is_deployment_admin: true }),
          async () => {
            throw new Error("read unavailable");
          },
        )}
      />,
    );
    fireEvent.click(
      await screen.findByTestId(deploymentUserRowTestId(user().user_id)),
    );
    await screen.findByTestId(deploymentAdminTestId("patch-display-name"));
    input(deploymentAdminTestId("patch-display-name"), "Saved edit");
    fireEvent.click(screen.getByTestId(deploymentAdminTestId("patch-user")));
    await screen.findByRole("button", { name: "Retry session refresh" });
    expect(
      screen.getByTestId(deploymentAdminTestId("status")).textContent,
    ).toContain("Patched local user");
    expect(
      fetchMock.mock.calls.filter(([, init]) => init?.method === "PATCH"),
    ).toHaveLength(1);
  });
});

function deploymentController(
  session: ReturnType<typeof sessionResource>,
  refresh: () => Promise<void> | void = () => {},
) {
  return new DeploymentUsersController({
    identity: () => ({
      actor: session.user_id,
      lifetime: "test-lifetime",
      admin: session.is_deployment_admin,
    }),
    subscribe: () => () => {},
    refresh,
    revoked: () => {},
    lost: () => {},
  });
}
