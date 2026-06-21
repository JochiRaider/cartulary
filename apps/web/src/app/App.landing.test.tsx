import {
  currentIncidentRoleTestId,
  deploymentUserRowTestId,
  incidentControlsMenuItemTestId,
  incidentControlsMenuTestId,
  incidentControlsTriggerTestId,
  landingAdminMenuItemTestId,
  landingAdminPanelTestId,
  landingAdminShellTestId,
  landingIncidentCardTestId,
  landingIncidentOpenButtonTestId,
  phase1AccountTestId,
  phase1AdminTestId,
  phase1AuthTestId,
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
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../workbook/WorkbookShell", async () => {
  const React = await import("react");

  return {
    WorkbookShell: ({
      accountApplicationMenu,
      incidentId,
      onIncidentAccessLost,
    }: {
      accountApplicationMenu?: (props: {
        currentIncidentRole: string;
        incidentControls?: undefined;
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
          <p data-testid="mock-workbook-role">{currentRole}</p>
          {accountApplicationMenu?.({
            currentIncidentRole: currentRole,
            incidentControls: undefined,
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
    ensureDraftRow: vi.fn(),
    TimelineWorkbook: vi.fn(),
  };
});

vi.mock("./Phase1Harness", () => ({
  Phase1Harness: () => <section data-testid="mock-phase1-harness" />,
}));

vi.mock("./Phase2Harness", () => ({
  Phase2Harness: () => <section data-testid="mock-phase2-harness" />,
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
import { AppRoot } from "./AppRoot";
import { AccountApplicationMenu } from "./LandingAdminSurface";

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
        triggerTestId={phase1RouteTestId("workbook-current-user")}
      />,
    );

    const trigger = screen.getByTestId(
      phase1RouteTestId("workbook-current-user"),
    );
    fireEvent.click(trigger);

    expect(screen.getByTestId(currentIncidentRoleTestId()).textContent).toBe(
      "Current incident role: admin",
    );
    expect(screen.getByTestId(phase1LandingTestId("return")).textContent).toBe(
      "Incidents",
    );
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
    fireEvent.click(screen.getByTestId(phase1LandingTestId("return")));
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
      await screen.findByTestId(phase1LandingTestId("empty-state")),
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
      screen.getByTestId(phase1LandingTestId("incidents-count")).textContent,
    ).toBe("0 loaded");
    expect(
      screen.getByTestId(phase1LandingTestId("current-user")).textContent,
    ).toBe("Bootstrap Admin · deployment admin");
    expect(
      findFetchCalls(fetchMock, "/api/v1/auth/session", "GET"),
    ).toHaveLength(1);
    expect(
      findFetchCalls(fetchMock, "/api/v1/auth/credential-state", "GET"),
    ).toHaveLength(1);
    expect(
      findFetchCallsByPath(fetchMock, "/api/v1/incidents", "GET"),
    ).toHaveLength(2);
    expect(
      fetchMock.mock.calls.filter(([, init]) => {
        const method = ((init as RequestInit | undefined)?.method ?? "GET")
          .toString()
          .toUpperCase();
        return method !== "GET";
      }),
    ).toHaveLength(0);
  });

  it("opens the workbook automatically when exactly one incident is visible", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Deployment Admin",
        is_deployment_admin: true,
      }),
      incidents: [
        incidentResource("incident-one", "IR-200", "Only visible incident"),
      ],
    });

    renderApp();

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    expect(screen.getByTestId("mock-workbook-incident").textContent).toBe(
      "incident-one",
    );
    expect(window.location.pathname).toBe("/");
    expect(window.location.search).toContain("incident_id=incident-one");
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
            incident_id: "incident-one",
            role: "admin",
          },
        ],
      }),
      incidents: [
        incidentResource("incident-one", "IR-200", "Only visible incident"),
      ],
    });

    renderApp();

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    fireEvent.click(
      screen.getByLabelText("Account and application navigation"),
    );
    fireEvent.click(screen.getByTestId(phase1LandingTestId("return")));

    expect(
      await screen.findByTestId(phase1LandingTestId("incident-list")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(landingIncidentOpenButtonTestId("incident-one")),
    ).toBeTruthy();
    expect(screen.queryByTestId("mock-workbook")).toBe(null);
    expect(window.location.search).not.toContain("incident_id=");
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
    expect(screen.queryByTestId(phase1LandingTestId("shell"))).toBe(null);
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
            claimed: true,
            route_families: ["/api/v1/incident-bundles"],
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
                  code: "incident_imported",
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
                  code: "incident_imported",
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

  it("returns non-admin direct deployment administration navigation to the incident directory", async () => {
    window.history.replaceState({}, "", "/deployment-administration");
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
      }),
    });

    renderApp();

    expect(
      await screen.findByTestId(phase1LandingTestId("empty-state")),
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
                user_id: "user-1",
                email: "operator@example.test",
                display_name: "Operator",
                user_version: 1,
              },
            }),
        },
        {
          method: "GET",
          url: "/api/v1/account/preferences",
          handler: () =>
            jsonResponse({
              data: {
                user_id: "user-1",
                density_mode: null,
                preferences_version: 1,
              },
            }),
        },
      ],
    });

    renderApp();

    await screen.findByTestId(phase1LandingTestId("empty-state"));
    fireEvent.click(
      screen.getByLabelText("Account and application navigation"),
    );
    expect(screen.queryByText("Deployment administration")).toBe(null);
    fireEvent.click(screen.getByRole("menuitem", { name: "Account settings" }));
    expect(
      await screen.findByTestId(phase1AccountTestId("profile-email")),
    ).toBeTruthy();
    await waitFor(() => {
      expect(
        screen.getByTestId(phase1AccountTestId("profile-email")).textContent,
      ).toBe("operator@example.test");
    });
    expect(
      screen.getByTestId(phase1AccountTestId("profile-display-name")),
    ).toBeTruthy();
    expect(screen.queryByText("MFA required")).toBe(null);

    fireEvent.click(screen.getByRole("tab", { name: "Appearance" }));
    const densityMode = await screen.findByTestId(
      phase1AccountTestId("appearance-density-mode"),
    );
    expect(
      Array.from((densityMode as HTMLSelectElement).options).map(
        (option) => option.textContent,
      ),
    ).toEqual(["Use surface default", "Compact", "Default", "Comfortable"]);

    fireEvent.click(screen.getByRole("tab", { name: "Security" }));
    expect(
      screen.getByTestId(phase1AccountTestId("refresh-state")),
    ).toBeTruthy();
    expect(screen.getByTestId(phase1AccountTestId("logout"))).toBeTruthy();
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
    expect(screen.queryByTestId(phase1AdminTestId("patch-user"))).toBe(null);
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
                    user_id: "user-2",
                    email: "target@example.test",
                    display_name: "Target User",
                    user_version: 3,
                    is_active: true,
                    mfa_required: true,
                    is_deployment_admin: false,
                  },
                ],
              },
              meta: {
                paging: {
                  limit: 100,
                  has_more: false,
                  next_cursor: null,
                },
                request_id: "req-users",
              },
            }),
        },
        {
          method: "GET",
          url: "/api/v1/users/user-2",
          handler: () =>
            jsonResponse({
              data: {
                user_id: "user-2",
                email: "target@example.test",
                display_name: "Target User",
                user_version: 3,
                is_active: true,
                mfa_required: true,
                is_deployment_admin: false,
              },
            }),
        },
      ],
    });

    renderApp();

    await screen.findByTestId(phase1LandingTestId("empty-state"));
    await openDeploymentAdministration();
    expect(screen.queryByTestId(phase1AdminTestId("patch-user"))).toBe(null);
    const userRow = await screen.findByTestId(
      deploymentUserRowTestId("user-2"),
    );
    fireEvent.click(userRow);
    await waitFor(() => {
      expect(
        screen.getByTestId(phase1AdminTestId("target-user-id")).textContent,
      ).toBe("user-2");
    });
    expect(
      (screen.getByTestId(phase1AdminTestId("patch-user")) as HTMLButtonElement)
        .disabled,
    ).toBe(false);
    expect(
      (
        screen.getByTestId(
          phase1AdminTestId("password-reset"),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(false);
    expect(
      (screen.getByTestId(phase1AdminTestId("revoke-all")) as HTMLButtonElement)
        .disabled,
    ).toBe(false);
  });

  it("renders one or many visible incidents on the landing screen", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
      }),
      incidents: [
        incidentResource("incident-1", "IR-201", "First Incident"),
        incidentResource("incident-2", "IR-202", "Second Incident"),
      ],
    });

    renderApp();

    expect(
      await screen.findByTestId(phase1LandingTestId("incident-list")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(phase1LandingTestId("incidents-count")).textContent,
    ).toBe("2 loaded");
    expect(
      screen.getByTestId(landingIncidentCardTestId("incident-1")).textContent,
    ).toContain("First Incident");
    expect(
      screen.getByTestId(landingIncidentCardTestId("incident-2")).textContent,
    ).toContain("Second Incident");
    expect(
      screen
        .getByTestId(landingIncidentCardTestId("incident-1"))
        .getAttribute("data-selected"),
    ).toBe(null);
    fireEvent.click(
      screen.getByTestId(landingIncidentOpenButtonTestId("incident-2")),
    );

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    expect(screen.getByTestId("mock-workbook-incident").textContent).toBe(
      "incident-2",
    );
  });

  it("searches visible incidents through the owner-backed list query and discards stale directory responses", async () => {
    const staleSearch = deferred<Response>();
    const acceptedSearch = deferred<Response>();
    const incidentRequestURLs: string[] = [];
    const alpha = incidentResource("incident-alpha", "IR-301", "Alpha Case");
    const beta = incidentResource("incident-beta", "IR-302", "Beta Case");
    const malware = incidentResource(
      "incident-malware",
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
      "incident-phish",
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

    await screen.findByTestId(landingIncidentCardTestId("incident-alpha"));
    const search = screen.getByTestId(phase1LandingTestId("search"));
    fireEvent.change(search, { target: { value: "phish" } });
    fireEvent.keyDown(search, { key: "Enter" });

    await waitFor(() => {
      expect(
        incidentRequestURLs.some((url) => url.includes("search=phish")),
      ).toBe(true);
    });
    expect(
      screen.getByTestId(landingIncidentCardTestId("incident-alpha")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(phase1LandingTestId("loading")).textContent,
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
      await screen.findByTestId(landingIncidentCardTestId("incident-malware")),
    ).toBeTruthy();
    expect(
      screen.queryByTestId(landingIncidentCardTestId("incident-phish")),
    ).toBe(null);

    staleSearch.resolve(incidentListResponse([phish]));
    await Promise.resolve();
    await Promise.resolve();
    expect(
      screen.queryByTestId(landingIncidentCardTestId("incident-phish")),
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
        `incident-${index + 1}`,
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

    await screen.findByTestId(landingIncidentCardTestId("incident-100"));
    expect(
      screen.queryByTestId(landingIncidentCardTestId("incident-101")),
    ).toBe(null);
    expect(
      screen.getByTestId(phase1LandingTestId("incidents-count")).textContent,
    ).toBe("100 loaded +");

    fireEvent.click(
      screen.getByRole("button", { name: "Load more incidents" }),
    );

    expect(
      await screen.findByTestId(landingIncidentCardTestId("incident-101")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(phase1LandingTestId("incidents-count")).textContent,
    ).toBe("101 loaded");
  });

  it("loads more incidents against the accepted search and status scope", async () => {
    const closedIncidents = Array.from({ length: 101 }, (_value, index) =>
      incidentResource(
        `incident-closed-${index + 1}`,
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
        incidentResource("incident-active-1", "IR-A-001", "Active Incident"),
        ...closedIncidents,
      ],
    });

    renderApp();

    await screen.findByTestId(landingIncidentCardTestId("incident-active-1"));
    const search = screen.getByTestId(phase1LandingTestId("search"));
    fireEvent.change(search, { target: { value: "Closed" } });
    fireEvent.change(screen.getByTestId(phase1LandingTestId("status-filter")), {
      target: { value: "closed" },
    });
    fireEvent.keyDown(search, { key: "Enter" });

    await screen.findByTestId(landingIncidentCardTestId("incident-closed-100"));
    expect(
      screen.queryByTestId(landingIncidentCardTestId("incident-active-1")),
    ).toBe(null);

    fireEvent.change(search, { target: { value: "Draft" } });
    fireEvent.click(
      screen.getByRole("button", { name: "Load more incidents" }),
    );

    expect(
      await screen.findByTestId(
        landingIncidentCardTestId("incident-closed-101"),
        {},
        { timeout: 5000 },
      ),
    ).toBeTruthy();
    const cursorRequest = findFetchCallsByPath(
      fetchMock,
      "/api/v1/incidents",
      "GET",
    ).find(([input]) =>
      new URL(String(input), "http://cartulary.test").searchParams.has(
        "cursor_token",
      ),
    );
    expect(cursorRequest).toBeTruthy();
    const cursorParams = new URL(
      String(cursorRequest?.[0] ?? ""),
      "http://cartulary.test",
    ).searchParams;
    expect(cursorParams.get("search")).toBe("Closed");
    expect(cursorParams.get("status")).toBe("closed");
    expect(cursorParams.has("group_by")).toBe(false);
  });

  it("sends only declared optional metadata when creating an incident from the directory", async () => {
    let createBody: Record<string, unknown> | null = null;
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
        memberships: [
          {
            incident_id: "incident-created",
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
          "incident-created",
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

    await screen.findByTestId(phase1LandingTestId("empty-state"));
    fireEvent.click(
      screen.getByTestId(phase1LandingTestId("create-open-button")),
    );
    fireEvent.change(screen.getByTestId(phase1LandingTestId("incident-key")), {
      target: { value: "IR-401" },
    });
    fireEvent.change(
      screen.getByTestId(phase1LandingTestId("incident-title")),
      {
        target: { value: "Created with metadata" },
      },
    );
    fireEvent.click(screen.getByText("More details"));
    fireEvent.change(
      screen.getByTestId(phase1LandingTestId("create-description")),
      {
        target: { value: "Initial notes" },
      },
    );
    fireEvent.change(
      screen.getByTestId(phase1LandingTestId("create-severity")),
      {
        target: { value: "high" },
      },
    );
    fireEvent.change(screen.getByTestId(phase1LandingTestId("create-tlp")), {
      target: { value: "TLP:AMBER" },
    });
    fireEvent.change(
      screen.getByTestId(phase1LandingTestId("create-current-phase")),
      {
        target: { value: "triage" },
      },
    );
    fireEvent.change(
      screen.getByTestId(phase1LandingTestId("create-external-case")),
      {
        target: { value: "CASE-401" },
      },
    );
    fireEvent.click(
      screen.getByTestId(phase1LandingTestId("create-submit-button")),
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

  it("Phase 2 U-2-11 ordinary landing shell creates an incident, refreshes session-visible membership, routes to the workbook by incident_id, and falls back when a stale incident selection is no longer visible", async () => {
    let created = false;
    installLandingShellFetch(fetchMock, {
      session: () =>
        sessionResource({
          display_name: "Operator",
          memberships: created
            ? [
                {
                  incident_id: "incident-created",
                  role: "admin",
                },
              ]
            : [],
        }),
      incidents: () =>
        created
          ? [incidentResource("incident-created", "IR-203", "Created Incident")]
          : [],
      onCreateIncident: () => {
        created = true;
        return incidentResource(
          "incident-created",
          "IR-203",
          "Created Incident",
        );
      },
    });

    renderApp();

    await screen.findByTestId(phase1LandingTestId("empty-state"));
    fireEvent.click(
      screen.getByTestId(phase1LandingTestId("create-open-button")),
    );
    fireEvent.change(screen.getByTestId(phase1LandingTestId("incident-key")), {
      target: { value: "IR-203" },
    });
    fireEvent.change(
      screen.getByTestId(phase1LandingTestId("incident-title")),
      {
        target: { value: "Created Incident" },
      },
    );
    fireEvent.click(
      screen.getByTestId(phase1LandingTestId("create-submit-button")),
    );

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    expect(screen.getByTestId("mock-workbook-incident").textContent).toBe(
      "incident-created",
    );
    await waitFor(() => {
      expect(screen.getByTestId("mock-workbook-role").textContent).toBe(
        "admin",
      );
    });
    expect(window.location.search).toContain("incident_id=incident-created");

    cleanup();
    window.history.replaceState({}, "", "/?incident_id=incident-stale");
    fetchMock.mockReset();
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
      }),
      incidents: [incidentResource("incident-live", "IR-204", "Live Incident")],
    });

    renderApp();

    expect(
      await screen.findByTestId(phase1LandingTestId("incident-list")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(landingIncidentOpenButtonTestId("incident-live")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(phase1LandingTestId("status")).textContent?.trim(),
    ).not.toBe("");
    expect(window.location.search).not.toContain("incident_id=");
  });

  it("renders the workbook directly from an authenticated incident route under StrictMode", async () => {
    window.history.replaceState({}, "", "/?incident_id=incident-5");
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
        memberships: [
          {
            incident_id: "incident-5",
            role: "admin",
          },
        ],
      }),
      incidents: [incidentResource("incident-5", "IR-205", "Visible Incident")],
    });

    renderApp();

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    expect(screen.getByTestId("mock-workbook-incident").textContent).toBe(
      "incident-5",
    );
    const appShell = screen.getByTestId(phase1RouteTestId("app-shell"));
    expect(appShell.style.blockSize).toBe("var(--ct-app-viewport-block-size)");
    expect(appShell.style.overflow).toBe("hidden");
    expect(["0", "0px"]).toContain(appShell.style.minBlockSize);
    expect(["0", "0px"]).toContain(appShell.style.minHeight);
    const workbookFrame = screen.getByTestId("mock-workbook").parentElement;
    expect(workbookFrame).toBeInstanceOf(HTMLElement);
    expect((workbookFrame as HTMLElement).style.display).toBe("grid");
    expect((workbookFrame as HTMLElement).style.blockSize).toBe("100%");
    expect((workbookFrame as HTMLElement).style.overflow).toBe("hidden");
    await expectStableFetchCount(fetchMock, 7);
  });

  it("returns to the landing screen when workbook access is lost", async () => {
    window.history.replaceState({}, "", "/?incident_id=incident-5");
    let accessLost = false;
    installLandingShellFetch(fetchMock, {
      session: () =>
        sessionResource({
          display_name: "Operator",
          memberships: accessLost
            ? []
            : [
                {
                  incident_id: "incident-5",
                  role: "admin",
                },
              ],
        }),
      incidents: () =>
        accessLost
          ? []
          : [incidentResource("incident-5", "IR-205", "Visible Incident")],
    });

    renderApp();

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    await expectStableFetchCount(fetchMock, 7);
    accessLost = true;
    fireEvent.click(screen.getByTestId("mock-access-lost"));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1LandingTestId("empty-state")),
      ).toBeTruthy();
    });
    expect(
      screen.getByTestId(phase1LandingTestId("status")).textContent?.trim(),
    ).not.toBe("");
    expect(window.location.search).not.toContain("incident_id=");
    await expectStableFetchCount(fetchMock, 17);
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
    expect(screen.getByTestId(phase1AuthTestId("status")).textContent).toBe(
      "Checking current session…",
    );

    window.history.pushState({}, "", "/?debug=harness");
    fireEvent.popState(window);

    expect(await screen.findByText("Debug harness shell")).toBeTruthy();
    await waitFor(() => {
      expect(abortedSignals).toHaveLength(1);
    });

    readyToLoad = true;
    window.history.pushState({}, "", "/");
    fireEvent.popState(window);

    expect(
      await screen.findByTestId(phase1LandingTestId("empty-state")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(phase1LandingTestId("current-user")).textContent,
    ).toBe("Operator");
    await expectStableFetchCount(fetchMock, 6);
  });
});

function renderApp() {
  return render(<AppRoot />);
}

function incidentListResponse(incidents: IncidentResource[]) {
  return jsonResponse({
    data: { incidents },
    meta: {
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
