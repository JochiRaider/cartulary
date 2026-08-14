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
import { AppRoot } from "./AppRoot";
import { AccountApplicationMenu } from "./LandingAdminLayout";

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
        incidentResource(
          "00000000-0000-4000-8000-000000001010",
          "IR-200",
          "Only visible incident",
        ),
      ],
    });

    renderApp();

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
    await waitFor(() => {
      expect(
        screen.getByTestId(accountTestId("profile-email")).textContent,
      ).toBe("operator@example.test");
    });
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
    const search = screen.getByTestId(incidentLandingTestId("search"));
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
      screen.getByTestId(incidentLandingTestId("incidents-count")).textContent,
    ).toBe("100 loaded +");

    fireEvent.click(
      screen.getByRole("button", { name: "Load more incidents" }),
    );

    expect(
      await screen.findByTestId(landingIncidentCardTestId("incident-101")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(incidentLandingTestId("incidents-count")).textContent,
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
    const search = screen.getByTestId(incidentLandingTestId("search"));
    fireEvent.change(search, { target: { value: "Closed" } });
    fireEvent.change(
      screen.getByTestId(incidentLandingTestId("status-filter")),
      {
        target: { value: "closed" },
      },
    );
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

  it("ordinary landing shell creates an incident, refreshes session-visible membership, routes to the workbook by incident_id, and falls back when a stale incident selection is no longer visible", async () => {
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
      incidents: [incidentResource("incident-live", "IR-204", "Live Incident")],
    });

    renderApp();

    expect(
      await screen.findByTestId(incidentLandingTestId("incident-list")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(landingIncidentOpenButtonTestId("incident-live")),
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
    await expectStableFetchCount(fetchMock, 8);
  });

  it("preserves incident route and manual directory route state across popstate navigation", async () => {
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
    await expectStableFetchCount(fetchMock, 8);
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
    await expectStableFetchCount(fetchMock, 20);
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
    await expectStableFetchCount(fetchMock, 7);
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
