import {
  accountTestId,
  appRouteTestId,
  authTestId,
  currentIncidentRoleTestId,
  deploymentAdminTestId,
  deploymentUserRowTestId,
  incidentControlsMenuItemTestId,
  incidentControlsMenuTestId,
  incidentControlsTriggerTestId,
  incidentLandingTestId,
  landingAdminMenuItemTestId,
  landingAdminPanelTestId,
  landingAdminShellTestId,
  landingIncidentCardTestId,
  landingIncidentOpenButtonTestId,
  referencePackAdminPanelTestId,
  referencePackRowTestId,
} from "@cartulary/ui-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../workbook/WorkbookShell", async () => {
  const React = await import("react");

  return {
    WorkbookShell: ({
      accountDensityMode,
      accountApplicationMenu,
      incidentId,
      onIncidentAccessLost,
    }: {
      accountDensityMode?: string | null;
      accountApplicationMenu?: (props: {
        currentIncidentRole: string;
        incidentControls?: {
          activeSection: "summary";
          items: readonly [
            {
              readonly description: "Incident summary and workbook defaults";
              readonly label: "Summary and preferences";
              readonly section: "summary";
            },
            {
              readonly description: "Incident access and roles";
              readonly label: "Memberships";
              readonly section: "memberships";
            },
          ];
          onSelectSection: (
            section: "summary" | "memberships",
            returnFocusTarget?: HTMLElement | null,
          ) => void;
        };
      }) => ReactNode;
      incidentId: string;
      onIncidentAccessLost?: () => void;
    }) => {
      const [currentRole, setCurrentRole] = React.useState("");

      React.useEffect(() => {
        let active = true;
        void fetch("/api/v1/auth/session")
          .then(async (response) => {
            const payload = (await response.json()) as {
              data?: {
                memberships?: Array<{
                  incident_id: string;
                  role: string;
                }>;
              };
            };
            if (!active) {
              return;
            }
            const membership =
              payload.data?.memberships?.find(
                (entry) => entry.incident_id === incidentId,
              ) ?? null;
            setCurrentRole(membership?.role ?? "");
          })
          .catch(() => {
            if (active) {
              setCurrentRole("");
            }
          });

        return () => {
          active = false;
        };
      }, [incidentId]);

      return (
        <section data-testid="mock-workbook">
          <p data-testid="mock-workbook-incident">{incidentId}</p>
          <p data-testid="mock-workbook-density">
            {accountDensityMode ?? "surface-default"}
          </p>
          <p data-testid="mock-workbook-role">{currentRole}</p>
          {accountApplicationMenu?.({
            currentIncidentRole: currentRole,
            incidentControls: {
              activeSection: "summary",
              items: [
                {
                  section: "summary",
                  label: "Summary and preferences",
                  description: "Incident summary and workbook defaults",
                },
                {
                  section: "memberships",
                  label: "Memberships",
                  description: "Incident access and roles",
                },
              ],
              onSelectSection: (section, returnFocusTarget) => {
                const target =
                  returnFocusTarget instanceof HTMLElement
                    ? returnFocusTarget.tagName
                    : "none";
                window.dispatchEvent(
                  new CustomEvent("mock-workbook-incident-controls", {
                    detail: { section, target },
                  }),
                );
              },
            },
          })}
          <button
            data-testid="mock-access-lost"
            type="button"
            onClick={() => {
              onIncidentAccessLost?.();
            }}
          >
            Lose access
          </button>
        </section>
      );
    },
    buildCreatePayload: vi.fn(),
    createDraftRow: vi.fn(),
    TimelineWorkbook: vi.fn(),
  };
});

vi.mock("./debug/AuthenticationDebugHarness", () => ({
  AuthenticationDebugHarness: () => (
    <section data-testid="mock-authentication-harness" />
  ),
}));

vi.mock("./debug/IncidentDirectoryDebugHarness", () => ({
  IncidentDirectoryDebugHarness: () => (
    <section data-testid="mock-incident-directory-harness" />
  ),
}));

import {
  type IncidentResource,
  incidentResource,
  installLandingShellFetch,
  sessionResource,
} from "../testing/appShellTestSupport";
import {
  abortablePendingResponse,
  deferred,
  expectStableFetchCount,
  findFetchCalls,
  findFetchCallsByPath,
  jsonResponse,
} from "../testing/fetchMockTestSupport";
import { AccountApplicationMenu } from "./AccountApplicationMenu";
import { AppRoot } from "./AppRoot";

describe("Incident landing", () => {
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

  describe("directory lifecycle", () => {
    const alpha = incidentResource(
      "00000000-0000-4000-8000-000000003001",
      "IR-D1",
      "Alpha",
    );
    const beta = incidentResource(
      "00000000-0000-4000-8000-000000003002",
      "IR-D2",
      "Beta",
    );
    const later = incidentResource(
      "00000000-0000-4000-8000-000000003003",
      "IR-D3",
      "Later",
    );
    function firstPage() {
      return jsonResponse({
        data: { incidents: [alpha, beta] },
        meta: {
          request_id: "request-directory",
          paging: { limit: 100, has_more: true, next_cursor: "next-page" },
        },
      });
    }
    function pageCalls() {
      return findFetchCallsByPath(fetchMock, "/api/v1/incidents", "GET").filter(
        ([input]) =>
          new URL(String(input), "http://cartulary.test").searchParams.has(
            "cursor_token",
          ),
      );
    }
    it("preserves explicit workbook startup through local sign-in without directory authorization probes", async () => {
      let authenticated = false;
      const session = sessionResource({
        memberships: [{ incident_id: alpha.incident_id, role: "viewer" }],
      });
      window.history.replaceState(
        {},
        "",
        `/?incident_id=${alpha.incident_id}&view_schema_id=cartulary.view.hosts.v1`,
      );
      installLandingShellFetch(fetchMock, {
        session: () =>
          authenticated
            ? session
            : jsonResponse(
                { error: { code: "session_required", status: 401 } },
                401,
              ),
        extraRoutes: [
          {
            method: "POST",
            url: "/api/v1/auth/login",
            handler: () => {
              authenticated = true;
              return jsonResponse({ data: session });
            },
          },
        ],
      });
      renderApp();
      fireEvent.change(
        await screen.findByTestId(authTestId("login-username")),
        { target: { value: "operator@example.test" } },
      );
      fireEvent.change(screen.getByTestId(authTestId("login-password")), {
        target: { value: "OperatorPass1!" },
      });
      fireEvent.click(screen.getByTestId(authTestId("login-submit")));
      await screen.findByTestId("mock-workbook");
      expect(
        new URLSearchParams(window.location.search).get("incident_id"),
      ).toBe(alpha.incident_id);
      expect(
        new URLSearchParams(window.location.search).get("view_schema_id"),
      ).toBe("cartulary.view.hosts.v1");
      expect(
        findFetchCallsByPath(fetchMock, "/api/v1/incidents", "GET"),
      ).toHaveLength(0);
    });

    it("keeps local and provider root entry membership-scoped for every directory size", async () => {
      for (const provider_type of ["local", "oidc", "saml"] as const) {
        for (const visible of [[], [alpha], [alpha, beta]]) {
          installLandingShellFetch(fetchMock, {
            session: sessionResource({
              provider_type,
              is_deployment_admin: true,
              memberships: visible.map((incident) => ({
                incident_id: incident.incident_id,
                role: "viewer",
              })),
            }),
            incidents: visible,
          });
          renderApp();
          await waitFor(() =>
            expect(
              screen
                .getByTestId(incidentLandingTestId("shell"))
                .getAttribute("data-directory-state"),
            ).toBe("ready"),
          );
          expect(
            screen.getByTestId(incidentLandingTestId("incidents-count"))
              .textContent,
          ).toBe(`${visible.length} loaded`);
          expect(screen.queryByTestId("mock-workbook")).toBeNull();
          expect(window.location.search).toBe("");
          expect(
            findFetchCallsByPath(fetchMock, "/api/v1/incidents", "GET"),
          ).toHaveLength(1);
          cleanup();
          fetchMock.mockReset();
        }
      }
    });

    it("retains the query across workbook navigation and refreshes only authoritative directory results", async () => {
      let changed = false;
      installLandingShellFetch(fetchMock, {
        session: sessionResource(),
        incidents: ({ query }) =>
          query.get("search") === "Alpha"
            ? changed
              ? []
              : [alpha]
            : [alpha, beta],
      });
      renderApp();
      await screen.findByTestId(landingIncidentCardTestId(alpha.incident_id));
      const before = [
        "/api/v1/auth/session",
        "/api/v1/auth/credential-state",
        "/api/v1/account/preferences",
        "/api/v1/extensions",
      ].map((path) => findFetchCallsByPath(fetchMock, path, "GET").length);
      fireEvent.change(screen.getByTestId(incidentLandingTestId("search")), {
        target: { value: "Alpha" },
      });
      await waitFor(() =>
        expect(
          screen.queryByTestId(landingIncidentCardTestId(beta.incident_id)),
        ).toBeNull(),
      );
      fireEvent.click(screen.getByTestId(incidentLandingTestId("refresh")));
      await waitFor(() =>
        expect(
          screen.queryByTestId(incidentLandingTestId("loading")),
        ).toBeNull(),
      );
      expect(
        [
          "/api/v1/auth/session",
          "/api/v1/auth/credential-state",
          "/api/v1/account/preferences",
          "/api/v1/extensions",
        ].map((path) => findFetchCallsByPath(fetchMock, path, "GET").length),
      ).toEqual(before);
      fireEvent.click(
        screen.getByTestId(landingIncidentOpenButtonTestId(alpha.incident_id)),
      );
      await screen.findByTestId("mock-workbook");
      changed = true;
      window.history.replaceState({}, "", "/");
      fireEvent.popState(window);
      await screen.findByTestId(incidentLandingTestId("empty-state"));
      expect(
        (
          screen.getByTestId(
            incidentLandingTestId("search"),
          ) as HTMLInputElement
        ).value,
      ).toBe("Alpha");
      expect(
        screen.queryByTestId(landingIncidentCardTestId(alpha.incident_id)),
      ).toBeNull();
      expect(
        new URL(
          String(
            findFetchCallsByPath(fetchMock, "/api/v1/incidents", "GET").at(
              -1,
            )?.[0],
          ),
          "http://cartulary.test",
        ).searchParams.get("search"),
      ).toBe("Alpha");
    });

    it("keeps malformed and transport directory failures local and never presents successful empty results", async () => {
      let outcome: "malformed" | "transport" | "valid" | "denied" = "malformed";
      installLandingShellFetch(fetchMock, {
        session: sessionResource(),
        incidents: () => {
          if (outcome === "malformed")
            return jsonResponse({
              data: { incidents: [alpha] },
              meta: { request_id: "bad-paging" },
            });
          if (outcome === "transport")
            return Promise.reject(new TypeError("Network failed"));
          if (outcome === "denied")
            return jsonResponse(
              { error: { code: "authorization_denied", status: 403 } },
              403,
            );
          return [alpha];
        },
      });
      renderApp();
      await screen.findByRole("button", { name: "Retry loading incidents" });
      expect(
        screen
          .getByTestId(appRouteTestId("app-shell"))
          .getAttribute("data-bootstrap-state"),
      ).toBe("authenticated");
      expect(
        screen.queryByTestId(incidentLandingTestId("empty-state")),
      ).toBeNull();
      expect(document.body.textContent).toContain(
        "invalid_public_contract_response",
      );
      outcome = "transport";
      fireEvent.click(
        screen.getByRole("button", { name: "Retry loading incidents" }),
      );
      await screen.findByRole("button", { name: "Retry loading incidents" });
      expect(
        screen.queryByTestId(incidentLandingTestId("empty-state")),
      ).toBeNull();
      outcome = "valid";
      fireEvent.click(
        screen.getByRole("button", { name: "Retry loading incidents" }),
      );
      await screen.findByTestId(landingIncidentCardTestId(alpha.incident_id));
      outcome = "denied";
      fireEvent.click(screen.getByTestId(incidentLandingTestId("refresh")));
      await screen.findByRole("button", { name: "Refresh first page" });
      expect(
        screen.queryByTestId(landingIncidentCardTestId(alpha.incident_id)),
      ).toBeNull();
      expect(
        screen.queryByTestId(incidentLandingTestId("empty-state")),
      ).toBeNull();
    });

    it("rejects obsolete authorization errors without ending replacement authentication", async () => {
      const pending = deferred<Response>();
      let currentSession = sessionResource();
      installLandingShellFetch(fetchMock, {
        session: () => currentSession,
        incidents: ({ query }) =>
          query.has("cursor_token") ? pending.promise : firstPage(),
      });
      renderApp();
      fireEvent.click(
        await screen.findByRole("button", { name: "Load more incidents" }),
      );
      currentSession = sessionResource({
        user_id: "00000000-0000-4000-8000-000000000002",
        display_name: "Replacement",
      });
      window.history.replaceState({}, "", "/?debug=harness");
      fireEvent.popState(window);
      await screen.findByTestId("mock-authentication-harness");
      window.history.replaceState({}, "", "/");
      fireEvent.popState(window);
      await waitFor(() =>
        expect(
          screen.getByTestId(incidentLandingTestId("current-user")).textContent,
        ).toBe("Replacement"),
      );
      await act(async () =>
        pending.resolve(
          jsonResponse(
            { error: { code: "session_required", status: 401 } },
            401,
          ),
        ),
      );
      expect(
        screen.getByTestId(incidentLandingTestId("current-user")).textContent,
      ).toBe("Replacement");
      expect(screen.queryByTestId(authTestId("shell"))).toBeNull();
    });

    it("preserves debounced edits made while a refresh is settling", async () => {
      const pending = deferred<Response>();
      const queries: string[] = [];
      let refreshing = false;
      installLandingShellFetch(fetchMock, {
        session: sessionResource(),
        incidents: ({ query }) => {
          const search = query.get("search") ?? "";
          queries.push(search);
          return search !== ""
            ? [later]
            : refreshing
              ? pending.promise
              : [alpha, beta];
        },
      });
      renderApp();
      await screen.findByTestId(landingIncidentCardTestId(alpha.incident_id));
      refreshing = true;
      const beforeRefresh = queries.length;
      vi.useFakeTimers();
      try {
        fireEvent.click(screen.getByTestId(incidentLandingTestId("refresh")));
        await act(async () => {
          await vi.advanceTimersByTimeAsync(1);
        });
        expect(queries.length).toBeGreaterThan(beforeRefresh);
        fireEvent.change(screen.getByTestId(incidentLandingTestId("search")), {
          target: { value: "Later" },
        });
        await act(async () => {
          pending.resolve(incidentListResponse([alpha, beta]));
        });
        await act(async () => {
          await vi.advanceTimersByTimeAsync(181);
        });
        expect(queries).toContain("Later");
      } finally {
        vi.useRealTimers();
      }
    });
    it("locks duplicate pagination before the next render", async () => {
      const pending = deferred<Response>();
      installLandingShellFetch(fetchMock, {
        session: sessionResource(),
        incidents: ({ query }) =>
          query.has("cursor_token") ? pending.promise : firstPage(),
      });
      renderApp();
      const button = await screen.findByRole("button", {
        name: "Load more incidents",
      });
      act(() => {
        button.click();
        button.click();
      });
      expect(pageCalls()).toHaveLength(1);
      await act(async () => {
        pending.resolve(incidentListResponse([later]));
      });
    });
    it("rejects a late page after a replacement query", async () => {
      const pending = deferred<Response>();
      installLandingShellFetch(fetchMock, {
        session: sessionResource(),
        incidents: ({ query }) =>
          query.has("cursor_token")
            ? pending.promise
            : query.has("search")
              ? incidentListResponse([beta])
              : firstPage(),
      });
      renderApp();
      fireEvent.click(
        await screen.findByRole("button", { name: "Load more incidents" }),
      );
      const search = screen.getByTestId(incidentLandingTestId("search"));
      fireEvent.change(search, { target: { value: "Beta" } });
      fireEvent.keyDown(search, { key: "Enter" });
      await waitFor(() =>
        expect(
          screen.queryByTestId(landingIncidentCardTestId(alpha.incident_id)),
        ).toBeNull(),
      );
      await act(async () => {
        pending.resolve(incidentListResponse([later]));
      });
      expect(
        screen.queryByTestId(landingIncidentCardTestId(later.incident_id)),
      ).toBeNull();
    });
    it("rejects a late page after directory navigation and re-entry", async () => {
      const pending = deferred<Response>();
      let returned = false;
      installLandingShellFetch(fetchMock, {
        session: sessionResource(),
        incidents: ({ query }) =>
          query.has("cursor_token")
            ? pending.promise
            : returned
              ? incidentListResponse([alpha, beta])
              : firstPage(),
      });
      renderApp();
      fireEvent.click(
        await screen.findByRole("button", { name: "Load more incidents" }),
      );
      fireEvent.click(
        screen.getByTestId(landingIncidentOpenButtonTestId(alpha.incident_id)),
      );
      await screen.findByTestId("mock-workbook");
      returned = true;
      window.history.replaceState({}, "", "/");
      fireEvent.popState(window);
      await screen.findByTestId(landingIncidentCardTestId(alpha.incident_id));
      await waitFor(() =>
        expect(
          screen.queryByTestId(incidentLandingTestId("loading")),
        ).toBeNull(),
      );
      await act(async () => {
        pending.resolve(incidentListResponse([later]));
      });
      expect(
        screen.queryByTestId(landingIncidentCardTestId(later.incident_id)),
      ).toBeNull();
    });
    it("rejects a late page after authentication replacement", async () => {
      const pending = deferred<Response>();
      let replaced = false;
      installLandingShellFetch(fetchMock, {
        session: () =>
          sessionResource(
            replaced
              ? {
                  user_id: "00000000-0000-4000-8000-000000003099",
                  display_name: "Replacement",
                }
              : {},
          ),
        incidents: ({ query }) =>
          query.has("cursor_token")
            ? pending.promise
            : replaced
              ? incidentListResponse([alpha, beta])
              : firstPage(),
      });
      renderApp();
      fireEvent.click(
        await screen.findByRole("button", { name: "Load more incidents" }),
      );
      replaced = true;
      // Browser navigation requires session bootstrap independently of directory refresh.
      window.history.replaceState({}, "", "/?debug=harness");
      fireEvent.popState(window);
      await screen.findByText("Debug harness shell");
      window.history.replaceState({}, "", "/");
      fireEvent.popState(window);
      await waitFor(() =>
        expect(
          screen.getByTestId(incidentLandingTestId("current-user")).textContent,
        ).toBe("Replacement"),
      );
      await act(async () => {
        pending.resolve(incidentListResponse([later]));
      });
      expect(
        screen.queryByTestId(landingIncidentCardTestId(later.incident_id)),
      ).toBeNull();
    });
    it("opens an explicit incident outside the former count probe", async () => {
      window.history.replaceState({}, "", `/?incident_id=${later.incident_id}`);
      installLandingShellFetch(fetchMock, {
        session: sessionResource({
          memberships: [{ incident_id: later.incident_id, role: "viewer" }],
        }),
        incidents: [alpha, beta, later],
      });
      renderApp();
      expect(
        (await screen.findByTestId("mock-workbook-incident")).textContent,
      ).toBe(later.incident_id);
      expect(
        findFetchCallsByPath(fetchMock, "/api/v1/incidents", "GET"),
      ).toHaveLength(0);
    });
  });

  describe("incident creation lifecycle", () => {
    async function openCreation() {
      await screen.findByTestId(incidentLandingTestId("empty-state"));
      fireEvent.click(
        screen.getByTestId(incidentLandingTestId("create-open-button")),
      );
      fireEvent.change(screen.getByLabelText("Incident key"), {
        target: { value: "IR-CREATE" },
      });
      fireEvent.change(screen.getByLabelText("Title"), {
        target: { value: "Creation draft" },
      });
    }

    function createdResponse() {
      return jsonResponse(
        {
          data: {
            ...incidentResource(
              "00000000-0000-4000-8000-000000001401",
              "IR-CREATE",
              "Creation draft",
            ),
            status: "active",
            closed_at: null,
            created_at: "2026-04-20T12:00:00Z",
            updated_at: "2026-04-20T12:00:00Z",
            created_by_user_id: sessionResource().user_id,
            updated_by_user_id: sessionResource().user_id,
          },
          meta: { request_id: "request-test" },
        },
        201,
      );
    }

    it("submits a single-line field with native Enter", async () => {
      installLandingShellFetch(fetchMock, {
        session: sessionResource(),
        onCreateIncident: createdResponse,
      });
      renderApp();
      await openCreation();
      screen.getByLabelText("Title").focus();
      await userEvent.keyboard("{Enter}");
      await waitFor(() =>
        expect(
          findFetchCalls(fetchMock, "/api/v1/incidents", "POST"),
        ).toHaveLength(1),
      );
    });

    it("suppresses same-tick duplicate dispatch", async () => {
      const pending = deferred<Response>();
      installLandingShellFetch(fetchMock, {
        session: sessionResource(),
        onCreateIncident: () => pending.promise,
      });
      renderApp();
      await openCreation();
      const button = screen.getByRole("button", { name: "Create and open" });
      act(() => {
        button.click();
        button.click();
      });
      expect(
        findFetchCalls(fetchMock, "/api/v1/incidents", "POST"),
      ).toHaveLength(1);
      await act(async () => {
        pending.resolve(createdResponse());
      });
    });

    it("associates required errors locally without replacing directory state", async () => {
      installLandingShellFetch(fetchMock, { session: sessionResource() });
      renderApp();
      await openCreation();
      fireEvent.change(screen.getByLabelText("Incident key"), {
        target: { value: "" },
      });
      fireEvent.click(screen.getByRole("button", { name: "Create and open" }));
      const key = screen.getByLabelText("Incident key");
      expect(key.getAttribute("aria-invalid")).toBe("true");
      expect(
        document.getElementById(key.getAttribute("aria-describedby") ?? "")
          ?.textContent,
      ).toContain("required");
      expect(
        screen
          .getByTestId(incidentLandingTestId("shell"))
          .getAttribute("data-bootstrap-state"),
      ).toBe("authenticated");
      expect(
        findFetchCalls(fetchMock, "/api/v1/incidents", "POST"),
      ).toHaveLength(0);
    });

    it("replays an uncertain response with the same captured payload", async () => {
      installLandingShellFetch(fetchMock, {
        session: sessionResource(),
        onCreateIncident: () =>
          jsonResponse(
            {
              error: {
                code: "service_unavailable",
                status: 502,
                retryable: true,
              },
            },
            502,
          ),
      });
      renderApp();
      await openCreation();
      fireEvent.click(screen.getByRole("button", { name: "Create and open" }));
      const retry = await screen.findByRole("button", {
        name: "Retry creation",
      });
      expect(
        (screen.getByLabelText("Title") as HTMLInputElement).readOnly,
      ).toBe(true);
      fireEvent.click(retry);
      await waitFor(() =>
        expect(
          findFetchCalls(fetchMock, "/api/v1/incidents", "POST"),
        ).toHaveLength(2),
      );
      const requests = findFetchCalls(fetchMock, "/api/v1/incidents", "POST");
      expect(requests[1]?.[1]?.body).toBe(requests[0]?.[1]?.body);
    });

    it("keeps a draft through keyboard dismissal and explicit reopening", async () => {
      installLandingShellFetch(fetchMock, { session: sessionResource() });
      renderApp();
      await openCreation();
      const key = screen.getByLabelText("Incident key");
      expect(document.activeElement).toBe(key);
      await userEvent.keyboard("{Escape}");
      expect(screen.queryByRole("form", { name: "New incident" })).toBeNull();
      expect(document.activeElement).toBe(
        screen.getByTestId(incidentLandingTestId("create-open-button")),
      );
      await userEvent.keyboard("{Enter}");
      expect((screen.getByLabelText("Title") as HTMLInputElement).value).toBe(
        "Creation draft",
      );
      expect(document.activeElement).toBe(
        screen.getByLabelText("Incident key"),
      );
      expect(screen.queryByRole("dialog", { name: "New incident" })).toBeNull();
    });

    it("reveals an optional field rejection with an associated message", async () => {
      installLandingShellFetch(fetchMock, {
        session: sessionResource(),
        onCreateIncident: () =>
          jsonResponse(
            {
              error: {
                code: "invalid_incident_create",
                status: 400,
                details: { field: "severity", reason_code: "field_too_long" },
              },
            },
            400,
          ),
      });
      renderApp();
      await openCreation();
      fireEvent.click(screen.getByRole("button", { name: "Create and open" }));
      await screen.findByText("Shorten this value and try again.");
      const severity = screen.getByLabelText("Severity");
      expect(severity.getAttribute("aria-invalid")).toBe("true");
      expect(severity.closest("details")?.open).toBe(true);
      expect(
        document.getElementById(severity.getAttribute("aria-describedby") ?? "")
          ?.textContent,
      ).toContain("Shorten");
      expect(
        screen
          .getByTestId(incidentLandingTestId("shell"))
          .getAttribute("data-bootstrap-state"),
      ).toBe("authenticated");
    });

    it("retries a failed confirmed handoff without creating again", async () => {
      let created = false;
      let failRefresh = true;
      installLandingShellFetch(fetchMock, {
        session: () =>
          created && failRefresh
            ? jsonResponse(
                { error: { code: "service_unavailable", status: 502 } },
                502,
              )
            : sessionResource({
                memberships: created
                  ? [
                      {
                        incident_id: "00000000-0000-4000-8000-000000001401",
                        role: "admin",
                      },
                    ]
                  : [],
              }),
        onCreateIncident: () => {
          created = true;
          return createdResponse();
        },
      });
      renderApp();
      await openCreation();
      fireEvent.click(screen.getByRole("button", { name: "Create and open" }));
      await screen.findByText(
        "Incident created, but the workbook could not be opened. Try opening it again.",
      );
      failRefresh = false;
      fireEvent.click(
        screen.getByRole("button", { name: "Open created incident" }),
      );
      await screen.findByTestId("mock-workbook");
      expect(
        findFetchCalls(fetchMock, "/api/v1/incidents", "POST"),
      ).toHaveLength(1);
      expect(window.location.search).toContain(
        "incident_id=00000000-0000-4000-8000-000000001401",
      );
    });

    it("clears a replaced account draft before accepting a delayed response", async () => {
      let currentSession = sessionResource();
      const pending = deferred<Response>();
      installLandingShellFetch(fetchMock, {
        session: () => currentSession,
        onCreateIncident: () => pending.promise,
      });
      renderApp();
      await openCreation();
      fireEvent.click(screen.getByRole("button", { name: "Create and open" }));
      currentSession = sessionResource({
        user_id: "00000000-0000-4000-8000-000000000002",
        display_name: "Other operator",
      });
      window.history.replaceState({}, "", "/?debug=harness");
      fireEvent.popState(window);
      await screen.findByTestId("mock-authentication-harness");
      window.history.replaceState({}, "", "/");
      fireEvent.popState(window);
      await screen.findByText("Other operator", { selector: "dd" });
      const response = createdResponse();
      const parsed = vi.spyOn(response, "json");
      await act(async () => {
        pending.resolve(response);
      });
      await waitFor(() => expect(parsed).toHaveBeenCalled());
      await act(async () => {
        await parsed.mock.results[0]?.value;
      });
      expect(window.location.search).not.toContain("incident_id=");
      fireEvent.click(
        screen.getByTestId(incidentLandingTestId("create-open-button")),
      );
      expect((screen.getByLabelText("Title") as HTMLInputElement).value).toBe(
        "",
      );
    });

    it("keeps a chosen account destination when a delayed create succeeds", async () => {
      const pending = deferred<Response>();
      installLandingShellFetch(fetchMock, {
        session: sessionResource(),
        onCreateIncident: () => pending.promise,
      });
      renderApp();
      await openCreation();
      fireEvent.click(screen.getByRole("button", { name: "Create and open" }));
      fireEvent.click(
        screen.getByLabelText("Account and application navigation"),
      );
      fireEvent.click(
        screen.getByRole("menuitem", { name: "Account settings" }),
      );
      const response = createdResponse();
      const parsed = vi.spyOn(response, "json");
      await act(async () => {
        pending.resolve(response);
      });
      await waitFor(() => expect(parsed).toHaveBeenCalled());
      await act(async () => {
        await parsed.mock.results[0]?.value;
      });
      expect(window.location.search).not.toContain("incident_id=");
      expect(screen.queryByTestId("mock-workbook")).toBeNull();
      expect(
        screen.getByRole("dialog", { name: "Account settings" }),
      ).toBeTruthy();
    });
  });

  it("renders workbook role, incidents, and expandable controls inside the account menu", () => {
    const openAccountSettings = vi.fn();
    const openDeploymentAdministration = vi.fn();
    const openIncidentDirectory = vi.fn();
    const selectControlsSection = vi.fn();

    render(
      <AccountApplicationMenu
        canOpenDeploymentAdministration
        currentContext="workbook"
        currentIncidentRole="admin"
        currentUserLabel="Dev Admin"
        incidentControls={{
          activeSection: "summary",
          items: [
            {
              section: "summary",
              label: "Summary and preferences",
              description: "Incident summary and workbook defaults",
            },
            {
              section: "incident-fields",
              label: "Promoted fields",
              description: "TLP, phase, and external case",
            },
            {
              section: "memberships",
              label: "Memberships",
              description: "Incident access and roles",
            },
          ],
          onSelectSection: selectControlsSection,
        }}
        onOpenAccountSettings={openAccountSettings}
        onOpenDeploymentAdministration={openDeploymentAdministration}
        onOpenIncidentDirectory={openIncidentDirectory}
        triggerTestId={appRouteTestId("workbook-current-user")}
      />,
    );

    const trigger = screen.getByTestId(appRouteTestId("workbook-current-user"));
    fireEvent.click(trigger);

    expect(screen.getByTestId(currentIncidentRoleTestId()).textContent).toBe(
      "Current incident role: admin",
    );
    expect(
      screen.getByTestId(incidentLandingTestId("return")).textContent,
    ).toBe("Incidents");
    expect(
      screen.getByRole("menuitem", { name: "Deployment administration" }),
    ).toBeTruthy();

    fireEvent.click(screen.getByTestId(incidentControlsTriggerTestId()));
    expect(
      screen.getByTestId(incidentControlsMenuTestId()).getAttribute("role"),
    ).toBe("menu");
    fireEvent.click(
      screen.getByTestId(incidentControlsMenuItemTestId("memberships")),
    );

    expect(selectControlsSection).toHaveBeenCalledWith("memberships", trigger);
    expect(screen.queryByTestId(incidentControlsMenuTestId())).toBe(null);

    fireEvent.click(trigger);
    fireEvent.click(screen.getByTestId(incidentLandingTestId("return")));
    expect(openIncidentDirectory).toHaveBeenCalledTimes(1);
    expect(openAccountSettings).not.toHaveBeenCalled();
    expect(openDeploymentAdministration).not.toHaveBeenCalled();
  });

  it("shows the zero-incident landing state", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Bootstrap Admin",
        is_deployment_admin: true,
      }),
    });

    renderApp();

    expect(
      await screen.findByTestId(incidentLandingTestId("empty-state")),
    ).toBeTruthy();
    expect(screen.getByText("Visible incidents")).toBeTruthy();
    expect(screen.queryByText("Workbook access")).toBe(null);
    const adminShell = screen.getByTestId(landingAdminShellTestId("shell"));
    expect(adminShell).toBeTruthy();
    expect((adminShell as HTMLElement).style.width).toBe("100%");
    expect((adminShell as HTMLElement).style.margin).toBe("");
    expect((adminShell as HTMLElement).style.borderRadius).toBe("");
    expect(window.getComputedStyle(document.body).margin).toBe("0px");
    expect(
      Array.from(document.querySelectorAll("style")).some((style) =>
        Boolean(
          style.textContent?.includes("#root") &&
            style.textContent.includes("margin: 0") &&
            style.textContent.includes("--ct-app-viewport-block-size: 100vh") &&
            style.textContent.includes("@supports (height: 100dvh)") &&
            style.textContent.includes(
              "scroll-margin-block-start: var(--cartulary-grid-scroll-margin-block-start)",
            ),
        ),
      ),
    ).toBe(true);
    expect(screen.queryByTestId(landingAdminShellTestId("menu"))).toBe(null);
    expect(screen.queryByTestId(landingAdminMenuItemTestId("incidents"))).toBe(
      null,
    );
    expect(
      screen
        .getByTestId(landingAdminPanelTestId("incidents"))
        .getAttribute("role"),
    ).toBe(null);
    expect(screen.queryByText("Open selected")).toBe(null);
    expect(
      screen.getByTestId(incidentLandingTestId("incidents-count")).textContent,
    ).toBe("0 loaded");
    expect(
      screen.getByTestId(incidentLandingTestId("current-user")).textContent,
    ).toBe("Bootstrap Admin");
    expect(
      findFetchCalls(fetchMock, "/api/v1/auth/session", "GET"),
    ).toHaveLength(1);
    expect(
      findFetchCalls(fetchMock, "/api/v1/auth/credential-state", "GET"),
    ).toHaveLength(1);
    expect(
      findFetchCallsByPath(fetchMock, "/api/v1/incidents", "GET"),
    ).toHaveLength(1);
    expect(
      fetchMock.mock.calls.filter(([, init]) => {
        const method = ((init as RequestInit | undefined)?.method ?? "GET")
          .toString()
          .toUpperCase();
        return method !== "GET";
      }),
    ).toHaveLength(0);
  });

  it("keeps exactly one visible incident in the directory until explicit selection", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Deployment Admin",
        is_deployment_admin: true,
      }),
      incidents: [
        incidentResource(
          "00000000-0000-4000-8000-000000001010",
          "IR-200",
          "Only visible incident",
        ),
      ],
    });

    renderApp();

    expect(
      await screen.findByTestId(incidentLandingTestId("incident-list")),
    ).toBeTruthy();
    expect(window.location.search).toBe("");
    expect(screen.queryByTestId("mock-workbook")).toBeNull();
    fireEvent.click(
      screen.getByTestId(
        landingIncidentOpenButtonTestId("00000000-0000-4000-8000-000000001010"),
      ),
    );
    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    expect(screen.getByTestId("mock-workbook-incident").textContent).toBe(
      "00000000-0000-4000-8000-000000001010",
    );
    expect(window.location.pathname).toBe("/");
    expect(window.location.search).toContain(
      "incident_id=00000000-0000-4000-8000-000000001010",
    );
    expect(new URLSearchParams(window.location.search).has("sheet_ref")).toBe(
      false,
    );
  });

  it("keeps explicit incident-directory navigation on the landing screen when exactly one incident is visible", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Deployment Admin",
        is_deployment_admin: true,
        memberships: [
          {
            incident_id: "00000000-0000-4000-8000-000000001010",
            role: "admin",
          },
        ],
      }),
      incidents: [
        incidentResource(
          "00000000-0000-4000-8000-000000001010",
          "IR-200",
          "Only visible incident",
        ),
      ],
    });

    window.history.replaceState(
      {},
      "",
      "/?incident_id=00000000-0000-4000-8000-000000001010",
    );
    renderApp();

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    fireEvent.click(
      screen.getByLabelText("Account and application navigation"),
    );
    fireEvent.click(screen.getByTestId(incidentLandingTestId("return")));

    expect(
      await screen.findByTestId(incidentLandingTestId("incident-list")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(
        landingIncidentOpenButtonTestId("00000000-0000-4000-8000-000000001010"),
      ),
    ).toBeTruthy();
    expect(screen.queryByTestId("mock-workbook")).toBe(null);
    expect(window.location.search).not.toContain("incident_id=");
    expect(
      screen.getByTestId(landingAdminShellTestId("heading")).textContent,
    ).toBe("Incident directory");
    expect(document.activeElement).toBe(
      screen.getByTestId(landingAdminShellTestId("heading")),
    );
  });

  it("loads deployment administration as a separate route without incident-directory fetches", async () => {
    window.history.replaceState({}, "", "/deployment-administration");
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
              data: { users: [] },
              meta: {
                request_id: "request-test",
                paging: {
                  limit: 100,
                  has_more: false,
                  next_cursor: null,
                },
              },
            }),
        },
      ],
    });

    renderApp();

    expect(
      await screen.findByTestId(landingAdminMenuItemTestId("deployment-users")),
    ).toBeTruthy();
    expect(screen.queryByTestId(incidentLandingTestId("shell"))).toBe(null);
    expect(
      screen.queryByTestId(landingAdminMenuItemTestId("reference-packs")),
    ).toBe(null);
    expect(
      screen.queryByTestId(landingAdminMenuItemTestId("incident-import")),
    ).toBe(null);
    expect(
      findFetchCallsByPath(fetchMock, "/api/v1/incidents", "GET"),
    ).toHaveLength(0);
  });

  it("opens an imported incident from terminal incident import job refs", async () => {
    window.history.replaceState({}, "", "/deployment-administration");
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Deployment Admin",
        is_deployment_admin: true,
      }),
      extensions: {
        extensions: [
          {
            profile_id: "incident_portability",
            claimable: true,
            claimed: true,
            contract_major: 1,
            route_families: ["/api/v1/incident-bundles"],
            workspace_keys: [],
            capabilities: [],
          },
        ],
      },
      extraRoutes: [
        {
          method: "GET",
          url: "/api/v1/users?limit=100",
          handler: () =>
            jsonResponse({
              data: { users: [] },
              meta: {
                request_id: "request-test",
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
          url: "/api/v1/incident-bundles/import",
          handler: () =>
            jsonResponse({
              data: {
                job_id: "job-import-1",
                status: "succeeded",
                progress: { completed: 1, total: 1 },
                result_summary: {
                  code: "incident_bundle_imported",
                  resource_refs: [
                    {
                      kind: "incident",
                      id: "incident-imported",
                      route: "/api/v1/incidents/incident-imported",
                    },
                  ],
                },
              },
            }),
        },
        {
          method: "GET",
          url: "/api/v1/jobs/job-import-1",
          handler: () =>
            jsonResponse({
              data: {
                job_id: "job-import-1",
                status: "succeeded",
                progress: { completed: 1, total: 1 },
                result_summary: {
                  code: "incident_bundle_imported",
                  resource_refs: [
                    {
                      kind: "incident",
                      id: "incident-imported",
                      route: "/api/v1/incidents/incident-imported",
                    },
                  ],
                },
              },
            }),
        },
      ],
    });

    renderApp();

    fireEvent.click(
      await screen.findByTestId(landingAdminMenuItemTestId("incident-import")),
    );
    expect(document.body.textContent ?? "").not.toContain(
      "Imported incident navigation is intentionally withheld",
    );
    fireEvent.click(screen.getByRole("button", { name: "Import incident" }));
    fireEvent.change(screen.getByLabelText("Incident bundle file"), {
      target: {
        files: [
          new File(["bundle"], "incident.zip", { type: "application/zip" }),
        ],
      },
    });
    fireEvent.click(screen.getByRole("button", { name: "Start import" }));

    fireEvent.click(await screen.findByText("Open imported incident"));

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    expect(screen.getByTestId("mock-workbook-incident").textContent).toBe(
      "incident-imported",
    );
    expect(window.location.search).toBe("?incident_id=incident-imported");
  });

  it("loads administrative audit directly from the deployment administration panel", async () => {
    window.history.replaceState({}, "", "/deployment-administration");
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
              data: { users: [] },
              meta: {
                request_id: "request-test",
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
          url: "/api/v1/administrative-audit-events?limit=100",
          handler: () =>
            jsonResponse({
              data: {
                audit_events: [
                  {
                    audit_event_id: "00000000-0000-4000-8000-000000002001",
                    scope_kind: "deployment",
                    scope_id: null,
                    occurred_at: "2026-05-24T00:00:00Z",
                    actor_kind: "user",
                    actor_user_id: "00000000-0000-4000-8000-000000000001",
                    source: "ui",
                    action_code: "user_created",
                    target_kind: "user",
                    target_id: "00000000-0000-4000-8000-000000000002",
                    reason_code: null,
                    changes: [
                      {
                        field_path: "display_name",
                        value_state: "visible",
                        before: null,
                        after: "Target User",
                      },
                    ],
                  },
                ],
              },
              meta: {
                request_id: "request-test",
                paging: {
                  limit: 100,
                  has_more: false,
                  next_cursor: null,
                },
              },
            }),
        },
      ],
    });

    renderApp();

    fireEvent.click(
      await screen.findByTestId(
        landingAdminMenuItemTestId("administrative-audit"),
      ),
    );

    expect(await screen.findByText("User created")).toBeTruthy();
    expect(
      screen.getByText("User 00000000-0000-4000-8000-000000000002"),
    ).toBeTruthy();
    expect(
      findFetchCalls(
        fetchMock,
        "/api/v1/administrative-audit-events?limit=100",
        "GET",
      ),
    ).toHaveLength(1);
    expect(
      findFetchCallsByPath(fetchMock, "/api/v1/incidents", "GET"),
    ).toHaveLength(0);
  });

  it("gates reference-pack deployment administration on the claimed extension profile", async () => {
    window.history.replaceState({}, "", "/deployment-administration");
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Deployment Admin",
        is_deployment_admin: true,
      }),
      extensions: {
        extensions: [
          {
            profile_id: "reference_pack",
            claimable: true,
            claimed: true,
            contract_major: 1,
            route_families: ["/api/v1/reference-packs"],
            workspace_keys: [],
            capabilities: [],
          },
        ],
      },
      extraRoutes: [
        {
          method: "GET",
          url: "/api/v1/users?limit=100",
          handler: () =>
            jsonResponse({
              data: { users: [] },
              meta: {
                request_id: "request-test",
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
          url: "/api/v1/reference-packs?limit=100",
          handler: () =>
            jsonResponse({
              data: {
                pack_versions: [
                  {
                    activated_at: "2026-05-24T00:00:01Z",
                    activated_by_user_id: null,
                    active: true,
                    imported_at: "2026-05-24T00:00:00Z",
                    imported_by_user_id: null,
                    manifest_sha256: "a".repeat(64),
                    pack_contract_version: "1",
                    pack_key: "type_registry.host",
                    pack_kind: "type_registry",
                    pack_version: "1",
                    pack_version_state: "verified_available",
                    payload_sha256: "b".repeat(64),
                    previous_active_version: null,
                    signer_key_id: null,
                    source_identifier: null,
                    verification_method: "sha256",
                    verification_result: "passed",
                  },
                ],
              },
              meta: {
                request_id: "request-test",
                paging: {
                  limit: 100,
                  has_more: false,
                  next_cursor: null,
                },
              },
            }),
        },
      ],
    });

    renderApp();

    fireEvent.click(
      await screen.findByTestId(landingAdminMenuItemTestId("reference-packs")),
    );

    expect(
      await screen.findByTestId(referencePackAdminPanelTestId()),
    ).toBeTruthy();
    expect(
      await screen.findByTestId(
        referencePackRowTestId("type_registry.host", "1"),
      ),
    ).toBeTruthy();
    expect(
      findFetchCalls(fetchMock, "/api/v1/reference-packs?limit=100", "GET"),
    ).toHaveLength(1);
  });

  it.each([
    {
      caseName: "the profile is missing",
      extensions: { extensions: [] },
    },
    {
      caseName: "the profile is unclaimed",
      extensions: {
        extensions: [
          {
            profile_id: "reference_pack" as const,
            claimable: true,
            claimed: false,
            contract_major: 1,
            route_families: ["/api/v1/reference-packs" as const],
            workspace_keys: [],
            capabilities: [],
          },
        ],
      },
    },
  ])("keeps reference-pack deployment administration hidden when $caseName", async ({
    extensions,
  }) => {
    window.history.replaceState({}, "", "/deployment-administration");
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Deployment Admin",
        is_deployment_admin: true,
      }),
      extensions,
      extraRoutes: [
        {
          method: "GET",
          url: "/api/v1/users?limit=100",
          handler: () =>
            jsonResponse({
              data: { users: [] },
              meta: {
                request_id: "request-test",
                paging: {
                  limit: 100,
                  has_more: false,
                  next_cursor: null,
                },
              },
            }),
        },
      ],
    });

    renderApp();

    expect(
      await screen.findByTestId(landingAdminMenuItemTestId("deployment-users")),
    ).toBeTruthy();
    expect(
      screen.queryByTestId(landingAdminMenuItemTestId("reference-packs")),
    ).toBe(null);
    expect(
      screen.queryByTestId(landingAdminPanelTestId("reference-packs")),
    ).toBe(null);
    expect(screen.queryByTestId(referencePackAdminPanelTestId())).toBe(null);
    expect(
      findFetchCallsByPath(fetchMock, "/api/v1/reference-packs", "GET"),
    ).toHaveLength(0);
  });

  it("returns non-admin direct deployment administration navigation to the incident directory", async () => {
    window.history.replaceState({}, "", "/deployment-administration");
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
      }),
    });

    renderApp();

    expect(
      await screen.findByTestId(incidentLandingTestId("empty-state")),
    ).toBeTruthy();
    expect(window.location.pathname).toBe("/");
    expect(
      screen.queryByTestId(landingAdminMenuItemTestId("deployment-users")),
    ).toBe(null);
  });

  it("opens account settings from the account menu and keeps deployment panels off the root", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
      }),
      extraRoutes: [
        {
          method: "GET",
          url: "/api/v1/account/profile",
          handler: () =>
            jsonResponse({
              data: {
                user_id: "00000000-0000-4000-8000-000000000001",
                email: "operator@example.test",
                display_name: "Operator",
                user_version: 1,
                created_at: "2026-04-20T12:00:00Z",
                updated_at: "2026-04-20T12:00:00Z",
              },
              meta: { request_id: "request-test" },
            }),
        },
        {
          method: "GET",
          url: "/api/v1/account/preferences",
          handler: () =>
            jsonResponse({
              data: {
                user_id: "00000000-0000-4000-8000-000000000001",
                density_mode: null,
                preferences_version: 1,
                created_at: "2026-04-20T12:00:00Z",
                updated_at: "2026-04-20T12:00:00Z",
              },
              meta: { request_id: "request-test" },
            }),
        },
      ],
    });

    renderApp();

    await screen.findByTestId(incidentLandingTestId("empty-state"));
    fireEvent.click(
      screen.getByLabelText("Account and application navigation"),
    );
    expect(screen.queryByText("Deployment administration")).toBe(null);
    fireEvent.click(screen.getByRole("menuitem", { name: "Account settings" }));
    expect(
      await screen.findByTestId(accountTestId("profile-email")),
    ).toBeTruthy();
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Close" }),
    );
    await waitFor(() => {
      expect(
        screen.getByTestId(accountTestId("profile-email")).textContent,
      ).toBe("operator@example.test");
    });
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Close" }),
    );
    expect(
      screen.getByTestId(accountTestId("profile-display-name")),
    ).toBeTruthy();
    expect(screen.queryByText("MFA required")).toBe(null);

    fireEvent.click(screen.getByRole("tab", { name: "Appearance" }));
    const densityMode = await screen.findByTestId(
      accountTestId("appearance-density-mode"),
    );
    expect(
      Array.from((densityMode as HTMLSelectElement).options).map(
        (option) => option.textContent,
      ),
    ).toEqual(["Use surface default", "Compact", "Default", "Comfortable"]);

    fireEvent.click(screen.getByRole("tab", { name: "Security" }));
    expect(screen.getByTestId(accountTestId("refresh-state"))).toBeTruthy();
    expect(screen.getByTestId(accountTestId("logout"))).toBeTruthy();
    expect(document.body.textContent ?? "").not.toContain("MFA required");
    expect(document.body.textContent ?? "").not.toContain(
      "Incident memberships",
    );
    expect(document.body.textContent ?? "").not.toContain("Deployment admin");

    expect(
      screen.queryByTestId(landingAdminMenuItemTestId("reference-packs")),
    ).toBe(null);
    expect(
      screen.queryByTestId(landingAdminMenuItemTestId("deployment-users")),
    ).toBe(null);
    expect(screen.queryByTestId(deploymentAdminTestId("patch-user"))).toBe(
      null,
    );
  });

  it("updates the open workbook density after saving account appearance", async () => {
    window.history.replaceState(
      {},
      "",
      "/?incident_id=00000000-0000-4000-8000-000000001001",
    );
    const incident = incidentResource(
      "00000000-0000-4000-8000-000000001001",
      "IR-1",
      "Density test",
    );
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
        memberships: [
          {
            incident_id: "00000000-0000-4000-8000-000000001001",
            role: "admin",
          },
        ],
      }),
      accountPreferences: {
        user_id: "00000000-0000-4000-8000-000000000001",
        density_mode: null,
        preferences_version: 1,
        created_at: "2026-04-20T12:00:00Z",
        updated_at: "2026-04-20T12:00:00Z",
      },
      incidents: [incident],
      extraRoutes: [
        {
          method: "GET",
          url: "/api/v1/account/profile",
          handler: () =>
            jsonResponse({
              data: {
                user_id: "00000000-0000-4000-8000-000000000001",
                email: "operator@example.test",
                display_name: "Operator",
                user_version: 1,
                created_at: "2026-04-20T12:00:00Z",
                updated_at: "2026-04-20T12:00:00Z",
              },
              meta: { request_id: "request-test" },
            }),
        },
        {
          method: "PUT",
          url: "/api/v1/account/preferences",
          handler: (request) => {
            const body = JSON.parse(String(request.init?.body ?? "{}")) as {
              density_mode?: string | null;
            };
            expect(body.density_mode).toBe("compact");
            return jsonResponse({
              data: {
                user_id: "00000000-0000-4000-8000-000000000001",
                density_mode: "compact",
                preferences_version: 2,
                created_at: "2026-04-20T12:00:00Z",
                updated_at: "2026-04-20T12:05:00Z",
              },
              meta: { request_id: "request-test" },
            });
          },
        },
      ],
    });

    renderApp();

    expect(
      (await screen.findByTestId("mock-workbook-density")).textContent,
    ).toBe("surface-default");
    fireEvent.click(
      screen.getByLabelText("Account and application navigation"),
    );
    fireEvent.click(screen.getByRole("menuitem", { name: "Account settings" }));
    fireEvent.click(screen.getByRole("tab", { name: "Appearance" }));
    fireEvent.change(
      await screen.findByTestId(accountTestId("appearance-density-mode")),
      {
        target: { value: "compact" },
      },
    );
    fireEvent.click(screen.getByTestId(accountTestId("appearance-save")));

    await waitFor(() => {
      expect(screen.getByTestId("mock-workbook-density").textContent).toBe(
        "compact",
      );
    });
  });

  it("enables deployment-user panel target actions after selecting a loaded user", async () => {
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
                users: [
                  {
                    user_id: "00000000-0000-4000-8000-000000000002",
                    email: "target@example.test",
                    display_name: "Target User",
                    user_version: 3,
                    is_active: true,
                    mfa_required: true,
                    is_deployment_admin: false,
                    created_at: "2026-04-20T12:00:00Z",
                    updated_at: "2026-04-20T12:00:00Z",
                    updated_by_user_id: null,
                    last_login_at: null,
                    auth_bindings: [],
                  },
                ],
              },
              meta: {
                request_id: "request-test",
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
          url: "/api/v1/users/00000000-0000-4000-8000-000000000002",
          handler: () =>
            jsonResponse({
              data: {
                user_id: "00000000-0000-4000-8000-000000000002",
                email: "target@example.test",
                display_name: "Target User",
                user_version: 3,
                is_active: true,
                mfa_required: true,
                is_deployment_admin: false,
                created_at: "2026-04-20T12:00:00Z",
                updated_at: "2026-04-20T12:00:00Z",
                updated_by_user_id: null,
                last_login_at: null,
                auth_bindings: [],
              },
              meta: { request_id: "request-test" },
            }),
        },
      ],
    });

    renderApp();

    await screen.findByTestId(incidentLandingTestId("empty-state"));
    await openDeploymentAdministration();
    expect(screen.queryByTestId(deploymentAdminTestId("patch-user"))).toBe(
      null,
    );
    const userRow = await screen.findByTestId(
      deploymentUserRowTestId("00000000-0000-4000-8000-000000000002"),
    );
    fireEvent.click(userRow);
    await waitFor(() => {
      expect(
        screen.getByTestId(deploymentAdminTestId("target-user-id")).textContent,
      ).toBe("00000000-0000-4000-8000-000000000002");
    });
    expect(
      (
        screen.getByTestId(
          deploymentAdminTestId("patch-user"),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(false);
    expect(
      (
        screen.getByTestId(
          deploymentAdminTestId("password-reset"),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(false);
    expect(
      (
        screen.getByTestId(
          deploymentAdminTestId("revoke-all"),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(false);
  });

  it("renders one or many visible incidents on the landing screen", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
      }),
      incidents: [
        incidentResource(
          "00000000-0000-4000-8000-000000001001",
          "IR-201",
          "First Incident",
        ),
        incidentResource(
          "00000000-0000-4000-8000-000000001002",
          "IR-202",
          "Second Incident",
        ),
      ],
    });

    renderApp();

    expect(
      await screen.findByTestId(incidentLandingTestId("incident-list")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(incidentLandingTestId("incidents-count")).textContent,
    ).toBe("2 loaded");
    expect(
      screen.getByTestId(
        landingIncidentCardTestId("00000000-0000-4000-8000-000000001001"),
      ).textContent,
    ).toContain("First Incident");
    expect(
      screen.getByTestId(
        landingIncidentCardTestId("00000000-0000-4000-8000-000000001002"),
      ).textContent,
    ).toContain("Second Incident");
    expect(
      screen
        .getByTestId(
          landingIncidentCardTestId("00000000-0000-4000-8000-000000001001"),
        )
        .getAttribute("data-selected"),
    ).toBe(null);
    fireEvent.click(
      screen.getByTestId(
        landingIncidentOpenButtonTestId("00000000-0000-4000-8000-000000001002"),
      ),
    );

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    expect(screen.getByTestId("mock-workbook-incident").textContent).toBe(
      "00000000-0000-4000-8000-000000001002",
    );
  });

  it("searches visible incidents through the owner-backed list query and discards stale directory responses", async () => {
    const staleSearch = deferred<Response>();
    const acceptedSearch = deferred<Response>();
    const incidentRequestURLs: string[] = [];
    const alpha = incidentResource(
      "9e484e84-6b2a-557d-9d20-f32c43468ce6",
      "IR-301",
      "Alpha Case",
    );
    const beta = incidentResource(
      "4fee4752-bb6d-5405-8217-f9495530734d",
      "IR-302",
      "Beta Case",
    );
    const malware = incidentResource(
      "67ae9f47-e4cc-5677-ba7a-9a4d38f02d85",
      "IR-303",
      "Malware investigation",
      {
        current_phase: "containment",
        primary_external_case_ref: "CASE-303",
        severity: "high",
        tlp: "TLP:AMBER",
      },
    );
    const phish = incidentResource(
      "8a621150-9a71-5706-9f88-c1e43a3ea68d",
      "IR-304",
      "Phishing investigation",
    );
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
      }),
      incidents: ({ query, url }) => {
        incidentRequestURLs.push(url);
        const search = query.get("search") ?? "";
        if (search === "phish") {
          return staleSearch.promise;
        }
        if (search === "malware") {
          return acceptedSearch.promise;
        }
        return [alpha, beta];
      },
    });

    renderApp();

    await screen.findByTestId(
      landingIncidentCardTestId("9e484e84-6b2a-557d-9d20-f32c43468ce6"),
    );
    const search = screen.getByTestId(incidentLandingTestId("search"));
    fireEvent.change(search, { target: { value: "phish" } });
    fireEvent.keyDown(search, { key: "Enter" });

    await waitFor(() => {
      expect(
        incidentRequestURLs.some((url) => url.includes("search=phish")),
      ).toBe(true);
    });
    expect(
      screen.getByTestId(
        landingIncidentCardTestId("9e484e84-6b2a-557d-9d20-f32c43468ce6"),
      ),
    ).toBeTruthy();
    expect(
      screen.getByTestId(incidentLandingTestId("loading")).textContent,
    ).toContain("Searching visible incidents");

    fireEvent.change(search, { target: { value: "malware" } });
    fireEvent.keyDown(search, { key: "Enter" });
    await waitFor(() => {
      expect(
        incidentRequestURLs.some((url) => url.includes("search=malware")),
      ).toBe(true);
    });
    acceptedSearch.resolve(incidentListResponse([malware]));

    expect(
      await screen.findByTestId(
        landingIncidentCardTestId("67ae9f47-e4cc-5677-ba7a-9a4d38f02d85"),
      ),
    ).toBeTruthy();
    expect(
      screen.queryByTestId(
        landingIncidentCardTestId("8a621150-9a71-5706-9f88-c1e43a3ea68d"),
      ),
    ).toBe(null);

    staleSearch.resolve(incidentListResponse([phish]));
    await Promise.resolve();
    await Promise.resolve();
    expect(
      screen.queryByTestId(
        landingIncidentCardTestId("8a621150-9a71-5706-9f88-c1e43a3ea68d"),
      ),
    ).toBe(null);
    for (const [input] of findFetchCallsByPath(
      fetchMock,
      "/api/v1/incidents",
      "GET",
    )) {
      const params = new URL(String(input), "http://cartulary.test")
        .searchParams;
      expect(params.has("group_by")).toBe(false);
    }
  });

  it("loads additional visible incident rows from the server cursor", async () => {
    const visibleIncidents = Array.from({ length: 101 }, (_value, index) =>
      incidentResource(
        `00000000-0000-4000-8000-${String(index + 1).padStart(12, "0")}`,
        `IR-${String(index + 1).padStart(3, "0")}`,
        `Incident ${index + 1}`,
      ),
    );
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
      }),
      incidents: visibleIncidents,
    });

    renderApp();

    await screen.findByTestId(
      landingIncidentCardTestId("00000000-0000-4000-8000-000000000100"),
    );
    expect(
      screen.queryByTestId(
        landingIncidentCardTestId("00000000-0000-4000-8000-000000000101"),
      ),
    ).toBe(null);
    expect(
      screen.getByTestId(incidentLandingTestId("incidents-count")).textContent,
    ).toBe("100 loaded +");

    fireEvent.click(
      screen.getByRole("button", { name: "Load more incidents" }),
    );

    expect(
      await screen.findByTestId(
        landingIncidentCardTestId("00000000-0000-4000-8000-000000000101"),
      ),
    ).toBeTruthy();
    expect(
      screen.getByTestId(incidentLandingTestId("incidents-count")).textContent,
    ).toBe("101 loaded");
  });

  it("disables continuation while replacing the accepted search and status scope", async () => {
    const closedIncidents = Array.from({ length: 101 }, (_value, index) =>
      incidentResource(
        `00000000-0000-4000-8001-${String(index + 1).padStart(12, "0")}`,
        `IR-C-${String(index + 1).padStart(3, "0")}`,
        `Closed Incident ${index + 1}`,
        { status: "closed" },
      ),
    );
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
      }),
      incidents: [
        incidentResource(
          "956e61eb-28f0-5245-85a1-4a2aa32f01bd",
          "IR-A-001",
          "Active Incident",
        ),
        ...closedIncidents,
      ],
    });

    renderApp();

    await screen.findByTestId(
      landingIncidentCardTestId("956e61eb-28f0-5245-85a1-4a2aa32f01bd"),
    );
    const search = screen.getByTestId(incidentLandingTestId("search"));
    fireEvent.change(search, { target: { value: "Closed" } });
    fireEvent.change(
      screen.getByTestId(incidentLandingTestId("status-filter")),
      {
        target: { value: "closed" },
      },
    );
    fireEvent.keyDown(search, { key: "Enter" });

    await screen.findByTestId(
      landingIncidentCardTestId("00000000-0000-4000-8001-000000000100"),
    );
    expect(
      screen.queryByTestId(
        landingIncidentCardTestId("956e61eb-28f0-5245-85a1-4a2aa32f01bd"),
      ),
    ).toBe(null);

    fireEvent.change(search, { target: { value: "Draft" } });
    const more = screen.getByRole("button", { name: "Load more incidents" });
    expect((more as HTMLButtonElement).disabled).toBe(true);
    fireEvent.click(more);
    await screen.findByTestId(incidentLandingTestId("empty-state"));
    const requests = findFetchCallsByPath(
      fetchMock,
      "/api/v1/incidents",
      "GET",
    ).map(
      ([input]) => new URL(String(input), "http://cartulary.test").searchParams,
    );
    expect(requests.some((params) => params.has("cursor_token"))).toBe(false);
    expect(requests.at(-1)?.get("search")).toBe("Draft");
    expect(requests.at(-1)?.get("status")).toBe("closed");
  });

  it("sends only declared optional metadata when creating an incident from the directory", async () => {
    let createBody: Record<string, unknown> | null = null;
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
        memberships: [
          {
            incident_id: "00000000-0000-4000-8000-000000001401",
            role: "admin",
          },
        ],
      }),
      onCreateIncident: ({ init }) => {
        createBody = JSON.parse(String(init?.body ?? "{}")) as Record<
          string,
          unknown
        >;
        return incidentResource(
          "00000000-0000-4000-8000-000000001401",
          "IR-401",
          "Created with metadata",
          {
            current_phase: "triage",
            description: "Initial notes",
            primary_external_case_ref: "CASE-401",
            severity: "high",
            tlp: "TLP:AMBER",
          },
        );
      },
    });

    renderApp();

    await screen.findByTestId(incidentLandingTestId("empty-state"));
    fireEvent.click(
      screen.getByTestId(incidentLandingTestId("create-open-button")),
    );
    fireEvent.change(
      screen.getByTestId(incidentLandingTestId("incident-key")),
      {
        target: { value: "IR-401" },
      },
    );
    fireEvent.change(
      screen.getByTestId(incidentLandingTestId("incident-title")),
      {
        target: { value: "Created with metadata" },
      },
    );
    fireEvent.click(screen.getByText("More details"));
    fireEvent.change(
      screen.getByTestId(incidentLandingTestId("create-description")),
      {
        target: { value: "Initial notes" },
      },
    );
    fireEvent.change(
      screen.getByTestId(incidentLandingTestId("create-severity")),
      {
        target: { value: "high" },
      },
    );
    fireEvent.change(screen.getByTestId(incidentLandingTestId("create-tlp")), {
      target: { value: "TLP:AMBER" },
    });
    fireEvent.change(
      screen.getByTestId(incidentLandingTestId("create-current-phase")),
      {
        target: { value: "triage" },
      },
    );
    fireEvent.change(
      screen.getByTestId(incidentLandingTestId("create-external-case")),
      {
        target: { value: "CASE-401" },
      },
    );
    fireEvent.click(
      screen.getByTestId(incidentLandingTestId("create-submit-button")),
    );

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    expect(createBody).toEqual({
      client_txn_id: expect.any(String),
      current_phase: "triage",
      description: "Initial notes",
      incident_key: "IR-401",
      primary_external_case_ref: "CASE-401",
      severity: "high",
      title: "Created with metadata",
      tlp: "TLP:AMBER",
    });
    for (const forbidden of [
      "default_workbook_preferences",
      "home_sheet_ref",
      "initial_memberships",
      "policy_defaults",
      "saved_views",
    ]) {
      expect(createBody).not.toHaveProperty(forbidden);
    }
  });

  it("ordinary landing shell creates an incident, refreshes membership, opens the workbook, and delegates stale selection to workbook access recovery", async () => {
    let created = false;
    installLandingShellFetch(fetchMock, {
      session: () =>
        sessionResource({
          display_name: "Operator",
          memberships: created
            ? [
                {
                  incident_id: "00000000-0000-4000-8000-000000001401",
                  role: "admin",
                },
              ]
            : [],
        }),
      incidents: () =>
        created
          ? [
              incidentResource(
                "00000000-0000-4000-8000-000000001401",
                "IR-203",
                "Created Incident",
              ),
            ]
          : [],
      onCreateIncident: () => {
        created = true;
        return incidentResource(
          "00000000-0000-4000-8000-000000001401",
          "IR-203",
          "Created Incident",
        );
      },
    });

    renderApp();

    await screen.findByTestId(incidentLandingTestId("empty-state"));
    fireEvent.click(
      screen.getByTestId(incidentLandingTestId("create-open-button")),
    );
    fireEvent.change(
      screen.getByTestId(incidentLandingTestId("incident-key")),
      {
        target: { value: "IR-203" },
      },
    );
    fireEvent.change(
      screen.getByTestId(incidentLandingTestId("incident-title")),
      {
        target: { value: "Created Incident" },
      },
    );
    fireEvent.click(
      screen.getByTestId(incidentLandingTestId("create-submit-button")),
    );

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    expect(screen.getByTestId("mock-workbook-incident").textContent).toBe(
      "00000000-0000-4000-8000-000000001401",
    );
    await waitFor(() => {
      expect(screen.getByTestId("mock-workbook-role").textContent).toBe(
        "admin",
      );
    });
    expect(window.location.search).toContain(
      "incident_id=00000000-0000-4000-8000-000000001401",
    );

    cleanup();
    window.history.replaceState({}, "", "/?incident_id=incident-stale");
    fetchMock.mockReset();
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
      }),
      incidents: [
        incidentResource(
          "50bc02c4-969e-55ec-8a26-5a33b0987fcf",
          "IR-204",
          "Live Incident",
        ),
      ],
    });

    renderApp();
    await screen.findByTestId("mock-workbook");
    expect(
      findFetchCallsByPath(fetchMock, "/api/v1/incidents", "GET"),
    ).toHaveLength(0);
    fireEvent.click(screen.getByTestId("mock-access-lost"));

    expect(
      await screen.findByTestId(incidentLandingTestId("incident-list")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(
        landingIncidentOpenButtonTestId("50bc02c4-969e-55ec-8a26-5a33b0987fcf"),
      ),
    ).toBeTruthy();
    expect(
      screen.getByTestId(incidentLandingTestId("status")).textContent?.trim(),
    ).not.toBe("");
    expect(window.location.search).not.toContain("incident_id=");
  });

  it("renders the workbook directly from an authenticated incident route under StrictMode", async () => {
    window.history.replaceState(
      {},
      "",
      "/?incident_id=00000000-0000-4000-8000-000000001005",
    );
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
        memberships: [
          {
            incident_id: "00000000-0000-4000-8000-000000001005",
            role: "admin",
          },
        ],
      }),
      incidents: [
        incidentResource(
          "00000000-0000-4000-8000-000000001005",
          "IR-205",
          "Visible Incident",
        ),
      ],
    });

    renderApp();

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    expect(screen.getByTestId("mock-workbook-incident").textContent).toBe(
      "00000000-0000-4000-8000-000000001005",
    );
    const appShell = screen.getByTestId(appRouteTestId("app-shell"));
    expect(appShell.style.blockSize).toBe("var(--ct-app-viewport-block-size)");
    expect(appShell.style.overflow).toBe("hidden");
    expect(["0", "0px"]).toContain(appShell.style.minBlockSize);
    expect(["0", "0px"]).toContain(appShell.style.minHeight);
    const workbookFrame = screen.getByTestId("mock-workbook").parentElement;
    expect(workbookFrame).toBeInstanceOf(HTMLElement);
    expect((workbookFrame as HTMLElement).style.display).toBe("grid");
    expect((workbookFrame as HTMLElement).style.blockSize).toBe("100%");
    expect((workbookFrame as HTMLElement).style.overflow).toBe("hidden");
    await expectStableFetchCount(fetchMock, 6);
  });

  it("preserves incident and directory routes across popstate navigation", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
        memberships: [
          {
            incident_id: "00000000-0000-4000-8000-000000001001",
            role: "admin",
          },
          {
            incident_id: "00000000-0000-4000-8000-000000001002",
            role: "viewer",
          },
        ],
      }),
      incidents: [
        incidentResource(
          "00000000-0000-4000-8000-000000001001",
          "IR-501",
          "First Incident",
        ),
        incidentResource(
          "00000000-0000-4000-8000-000000001002",
          "IR-502",
          "Second Incident",
        ),
      ],
    });

    renderApp();

    await screen.findByTestId(
      landingIncidentCardTestId("00000000-0000-4000-8000-000000001001"),
    );
    fireEvent.click(
      screen.getByTestId(
        landingIncidentOpenButtonTestId("00000000-0000-4000-8000-000000001002"),
      ),
    );
    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    expect(screen.getByTestId("mock-workbook-incident").textContent).toBe(
      "00000000-0000-4000-8000-000000001002",
    );

    window.history.replaceState(
      {},
      "",
      "/?incident_id=00000000-0000-4000-8000-000000001001",
    );
    fireEvent.popState(window);
    await waitFor(() => {
      expect(screen.getByTestId("mock-workbook-incident").textContent).toBe(
        "00000000-0000-4000-8000-000000001001",
      );
    });

    window.history.replaceState({ cartularyIncidentDirectory: true }, "", "/");
    fireEvent.popState(window);
    expect(
      await screen.findByTestId(incidentLandingTestId("incident-list")),
    ).toBeTruthy();
    expect(screen.queryByTestId("mock-workbook")).toBe(null);
    expect(
      screen.getByTestId(
        landingIncidentOpenButtonTestId("00000000-0000-4000-8000-000000001001"),
      ),
    ).toBeTruthy();
  });

  it("passes workbook incident controls through the app account menu handoff", async () => {
    window.history.replaceState(
      {},
      "",
      "/?incident_id=00000000-0000-4000-8000-000000001001",
    );
    const controlsEvents: Array<{ section: string; target: string }> = [];
    window.addEventListener("mock-workbook-incident-controls", (event) => {
      controlsEvents.push(
        (event as CustomEvent<{ section: string; target: string }>).detail,
      );
    });
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
        memberships: [
          {
            incident_id: "00000000-0000-4000-8000-000000001001",
            role: "admin",
          },
        ],
      }),
      incidents: [
        incidentResource(
          "00000000-0000-4000-8000-000000001001",
          "IR-503",
          "Incident",
        ),
      ],
    });

    renderApp();

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    fireEvent.click(
      screen.getByLabelText("Account and application navigation"),
    );
    fireEvent.click(screen.getByTestId(incidentControlsTriggerTestId()));
    fireEvent.click(
      screen.getByTestId(incidentControlsMenuItemTestId("memberships")),
    );

    expect(controlsEvents).toEqual([
      {
        section: "memberships",
        target: "BUTTON",
      },
    ]);
  });

  it("returns to the landing screen when workbook access is lost", async () => {
    window.history.replaceState(
      {},
      "",
      "/?incident_id=00000000-0000-4000-8000-000000001005",
    );
    let accessLost = false;
    installLandingShellFetch(fetchMock, {
      session: () =>
        sessionResource({
          display_name: "Operator",
          memberships: accessLost
            ? []
            : [
                {
                  incident_id: "00000000-0000-4000-8000-000000001005",
                  role: "admin",
                },
              ],
        }),
      incidents: () =>
        accessLost
          ? []
          : [
              incidentResource(
                "00000000-0000-4000-8000-000000001005",
                "IR-205",
                "Visible Incident",
              ),
            ],
    });

    renderApp();

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    await expectStableFetchCount(fetchMock, 6);
    accessLost = true;
    fireEvent.click(screen.getByTestId("mock-access-lost"));

    await waitFor(() => {
      expect(
        screen.getByTestId(incidentLandingTestId("empty-state")),
      ).toBeTruthy();
    });
    expect(
      screen.getByTestId(incidentLandingTestId("status")).textContent?.trim(),
    ).not.toBe("");
    expect(window.location.search).not.toContain("incident_id=");
    await expectStableFetchCount(fetchMock, 7);
  });

  it("cancels an in-flight shell refresh when the app unmounts", async () => {
    const abortedSignals: AbortSignal[] = [];
    fetchMock.mockImplementation((input, init) => {
      if (String(input) === "/api/v1/auth/session") {
        return abortablePendingResponse(
          init?.signal as AbortSignal | undefined,
          (signal) => {
            abortedSignals.push(signal);
          },
        );
      }
      throw new Error(`unexpected fetch: ${String(input)}`);
    });

    const view = renderApp();

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(1);
    });
    view.unmount();

    await waitFor(() => {
      expect(abortedSignals).toHaveLength(1);
    });
    await expectStableFetchCount(fetchMock, 1);
  });

  it("aborts an in-flight refresh when entering the debug harness and loads the ordinary shell after leaving it", async () => {
    const abortedSignals: AbortSignal[] = [];
    let readyToLoad = false;
    installLandingShellFetch(fetchMock, {
      session: ({ init }) => {
        if (!readyToLoad) {
          return abortablePendingResponse(
            init?.signal as AbortSignal | undefined,
            (signal) => {
              abortedSignals.push(signal);
            },
          );
        }
        return sessionResource({
          display_name: "Operator",
        });
      },
    });

    renderApp();

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(1);
    });
    expect(screen.getByTestId(authTestId("status")).textContent).toBe(
      "Checking current session...",
    );

    window.history.pushState({}, "", "/?debug=harness");
    fireEvent.popState(window);

    expect(await screen.findByText("Debug harness shell")).toBeTruthy();
    expect(
      await screen.findByTestId(appRouteTestId("debug-harness-shell")),
    ).toBeTruthy();
    await waitFor(() => {
      expect(abortedSignals).toHaveLength(1);
    });

    readyToLoad = true;
    window.history.pushState({}, "", "/");
    fireEvent.popState(window);

    expect(
      await screen.findByTestId(incidentLandingTestId("empty-state")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(incidentLandingTestId("current-user")).textContent,
    ).toBe("Operator");
    await expectStableFetchCount(fetchMock, 6);
  });

  it("loads the debug harness directly without ordinary shell bootstrap requests", async () => {
    window.history.replaceState({}, "", "/?debug=harness");
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
      }),
    });

    renderApp();

    expect(await screen.findByText("Debug harness shell")).toBeTruthy();
    expect(
      await screen.findByTestId(appRouteTestId("debug-harness-shell")),
    ).toBeTruthy();
    expect(screen.getByTestId("mock-authentication-harness")).toBeTruthy();
    expect(screen.getByTestId("mock-incident-directory-harness")).toBeTruthy();
    await expectStableFetchCount(fetchMock, 0);
  });
});

function renderApp() {
  return render(<AppRoot />);
}

function incidentListResponse(incidents: IncidentResource[]) {
  return jsonResponse({
    data: { incidents },
    meta: {
      request_id: "request-test",
      paging: {
        has_more: false,
        limit: 100,
        next_cursor: null,
      },
    },
  });
}

async function openDeploymentAdministration() {
  fireEvent.click(screen.getByLabelText("Account and application navigation"));
  fireEvent.click(
    screen.getByRole("menuitem", { name: "Deployment administration" }),
  );
  await screen.findByTestId(landingAdminMenuItemTestId("deployment-users"));
}
